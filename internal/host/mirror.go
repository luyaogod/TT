package host

// 本地源码镜像引擎:把某环境(T100 服务器)的源码拉到本地镜像目录。
//
// 两个 profile 的白名单：
//   - source（默认）：4gl 目录下后缀为 .4gl 的源码 + 42f（编译表单）+
//     42s/zh_CN（编译字符串）+ *.inc（4GL include）；
//     .4gl.src、.4gltest、.str、per、其它语言目录和备份都不取；
//   - interface：只取 42f 与 42s/zh_CN（供 tt interface 合成画面；只想省体积时
//     用它配一个独立目录，不要和 source 混在同一个目录里）。
//
// **默认 profile 必须是“所有读者的并集”**：一个环境只有一个镜像目录，而
// tt outline 只认 4gl、tt interface 只认 42f/42s，两者读的都是这个目录。少收一类
// 就等于让某个命令静默断粮（命令只会报“找不到文件”，看起来像文件不存在）。
// *.inc 没有命令直接读，它服务的是“人/AI 离线读码”（全局变量、共用参数），
// 缺了同样不报错 —— 这类静默缺口正是最该避免的。
//
// 4fd 是设计期字段描述，fieldId 会随保存重排，两个 profile 都不取。
// 备份/临时文件(x.4gl.bak、x.bck、x.bck1、*~ 等)不拉,并会清理本地残留。
//
// 传输:服务器侧 tar 打包(绝对路径,exec 通道不经登录 profile) → SFTP 流式
// 下载 → 本地标准库解压(gzip+tar),全程不进内存。**只有全量拉取这一种模式**:
// 每次都在服务器上重压一遍白名单内的全部文件,本地整目录替换(含删除服务器已删的文件)。
//
// 归档放**服务器临时目录**(默认 ${TMPDIR:-/tmp},可用 mirror.tmpDir 覆盖),
// 不进 $TOP —— 既不要求 ERP 顶级目录可写,也不在 ERP 目录里留任何东西。
// 文件名带 nonce,两台机器拉同一环境不会互相覆盖。
//
// 服务器上只可能出现两个文件(<base>.list、<base>.tar.gz),清理分四层:
//  1. 打包脚本自己删(.list 成功/失败都删;tar 失败连半截归档一起删);
//  2. 客户端 defer 无条件删(成功、失败、取消都走);
//  3. 取消时关闭 session,远端 tar 收 SIGHUP 立刻退出(OutputContext);
//  4. 客户端被强杀/断网时没人执行删除 —— 下次拉取前扫掉 3 小时前的遗留。
// 服务器重启后 /tmp 由系统清空,不需要我们管。

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type MirrorProfile string

const (
	MirrorSource    MirrorProfile = "source"
	MirrorInterface MirrorProfile = "interface"
)

// ParseMirrorProfile 只认两个公开值；空串按 source。
func ParseMirrorProfile(s string) (MirrorProfile, error) {
	switch MirrorProfile(strings.ToLower(strings.TrimSpace(s))) {
	case "", MirrorSource:
		return MirrorSource, nil
	case MirrorInterface:
		return MirrorInterface, nil
	default:
		return "", fmt.Errorf("未知镜像 profile %q（可用：source | interface）", s)
	}
}

// Label 是给人看的 profile 名称。
func (p MirrorProfile) Label() string {
	if p == MirrorInterface {
		return "42f/42s"
	}
	return "4gl/42f/42s/inc"
}

// MirrorStats 一次镜像拉取的统计结果(CLI 打印用)。
type MirrorStats struct {
	Env     string        // 环境名
	Profile MirrorProfile // source | interface
	TopDir  string        // 服务器 TOP(打包根)
	TmpDir  string        // 服务器临时目录(归档落点)
	Files   int           // 归档内文件数(白名单内一个都没有时为 0)
	Bytes   int64         // 下载字节数(压缩后)
	Pruned  int           // 本地清理掉的残留备份文件数(历史拉取留下)
	Elapsed string        // 耗时
	Note    string        // 附加说明(如 "白名单内没有文件")
}

// MirrorOptions 拉取的附加参数;零值即全部缺省。
type MirrorOptions struct {
	// TmpDir 服务器临时目录。空 = 服务器上的 ${TMPDIR:-/tmp}。
	TmpDir string
}

// localMarkName 本地镜像目录内的"上次成功"标记文件:存在表示该目录曾完整拉取成功。
// 全量模式下它不再参与任何拉取决策,只用于判断"这个目录有没有内容"(MirrorReadyDir)
// 以及记下这份目录拉过哪几类(见 localMark)。
const localMarkName = ".tdict-mirror.ok"

// mirrorTag 环境名 → 服务器 marker 文件名 ASCII 尾缀。
// 环境名可能是中文/含特殊字符(不能直接作文件名),且同一服务器可挂多个环境,
// 统一取 sha1 前 10 位十六进制:唯一、安全、与语言无关。
func mirrorTag(name string) string {
	sum := sha1.Sum([]byte(name))
	return fmt.Sprintf("%x", sum)[:10]
}

// mirrorTempBase 服务器侧归档/清单的基础路径(不带 .list/.tar.gz 后缀)。
//
// 落在**临时目录**而不是 $TOP。旧版本写在 `<TOP>/.tdict-mirror-*`：取消或失败时
// 265 MB 的归档就留在 ERP 顶级目录里，而且要求登录账号对 $TOP 有写权限（很多客户给不了）。
//
// 名字里的 nonce 不是装饰：以前文件名固定成「环境 × profile」，两台机器同时拉同一环境
// 会互相覆盖（一台正在下载的归档被另一台重写，gzip 校验必然失败）。nonce 每次拉取唯一，
// 陈旧遗留由下一次拉取的 3 小时清扫兜掉（见 mirrorTmpProbe）。
func mirrorTempBase(tmpDir, envName string, profile MirrorProfile, nonce string) string {
	if profile == "" {
		profile = MirrorSource
	}
	dir := strings.TrimSuffix(strings.TrimSpace(tmpDir), "/")
	return dir + "/.tt-mirror-" + mirrorTag(envName) + "-" + string(profile) + "-" + nonce
}

// mirrorNonce 生成 8 位十六进制随机串(归档名用)。随机源失效时退回时间戳 —— 宁可
// 重复也不能空，空会让名字变成 `.tt-mirror-x-source-` 这种无法区分的东西。
func mirrorNonce() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%08x", time.Now().UnixNano()&0xffffffff)
	}
	return fmt.Sprintf("%x", b)
}

// MirrorProgress 拉取过程的进度(Web 进度条用;CLI 传 nil 回调即无进度)。
type MirrorProgress struct {
	Phase   string // connect|probe|pack|download|done
	Message string // 人类可读说明
	Bytes   int64  // 已处理字节:pack=服务器归档当前大小,download=已读压缩字节
	Total   int64  // 归档大小:download=真实值;pack=**预估**值(见 packEstimate);0=未知
	Files   int    // 文件数(已知时)
}

// ProgressFunc 进度回调(可能被多次调用,实现需自行加锁)。
type ProgressFunc func(MirrorProgress)

// MirrorPull 对指定环境做一次**全量拉取**(4gl/**/*.4gl + 42f + 42s/zh_CN + *.inc)
// 到 mirrorDir/<环境名>/。
//
// 只有全量：不再有服务器侧基线，也不再有"只拉变更"。成功结束后写入/刷新本地
// .tdict-mirror.ok 标记(只表示"这个目录拉成功过哪几类")。
// 镜像目录本身(mirrorDir)由调用方保证已配置;本函数只负责连服务器与落盘。
// ctx 见 MirrorPullProgress(它才是实现,这里是给不需要取消的调用方的薄封装)。
func MirrorPull(ctx context.Context, e *NamedSsh, mirrorDir string, opts MirrorOptions) (*MirrorStats, error) {
	return MirrorPullProgress(ctx, e, mirrorDir, opts, nil)
}

// MirrorPullToDir 把镜像直接拉到一个指定目录（不再拼 <环境名>）。
// 用于设置页里“每个环境单独指定镜像目录”的场景；默认 source profile。
func MirrorPullToDir(ctx context.Context, e *NamedSsh, envDir string, opts MirrorOptions) (*MirrorStats, error) {
	return MirrorPullToDirProgress(ctx, e, envDir, opts, nil)
}

// MirrorPullToDirProfile 按 profile 拉到指定目录。
func MirrorPullToDirProfile(ctx context.Context, e *NamedSsh, envDir string, profile MirrorProfile, opts MirrorOptions) (*MirrorStats, error) {
	return MirrorPullToDirProfileProgress(ctx, e, envDir, profile, opts, nil)
}

// MirrorPullProgress 同 MirrorPull,并在各阶段回调 onProgress(供 Web 显示拉取进度)。
//
// 兼容旧调用：mirrorDir 是根目录，实际落点是 mirrorDir/<环境名>。
func MirrorPullProgress(ctx context.Context, e *NamedSsh, mirrorDir string, opts MirrorOptions, onProgress ProgressFunc) (*MirrorStats, error) {
	return MirrorPullToDirProgress(ctx, e, MirrorEnvDir(mirrorDir, e.Name), opts, onProgress)
}

// MirrorPullToDirProgress 是 source profile 的薄封装。
func MirrorPullToDirProgress(ctx context.Context, e *NamedSsh, envDir string, opts MirrorOptions, onProgress ProgressFunc) (*MirrorStats, error) {
	return MirrorPullToDirProfileProgress(ctx, e, envDir, MirrorSource, opts, onProgress)
}

// MirrorPullToDirProfileProgress 是镜像拉取的实现：envDir 就是本环境的最终目录。
//
// profile 只影响服务器侧 find 的白名单：source=4gl/**/*.4gl + 42f + 42s/zh_CN + *.inc，
// interface=42f/42s。落点与取消语义一致 —— 同一目录切 profile 不会删除另一类的文件
// （全量解压也会把它们原样搬进新目录，见 extractFull）；想只要一类的干净目录，
// 给 interface profile 单独配一个目录。
//
// 服务器流程：探测 TOP → 探测临时目录(含陈旧清扫、可写与剩余空间预检) → 估算归档大小
// → tar 打包 → SFTP 下载 → 本地解压 → 字节数校验 → 写本地标记 → 删服务器归档。
// 归档只存在于临时目录，$TOP 一个字节都不写(见 mirrorTempBase)。
//
// ctx 用来**中途取消**:界面上的「停止」按钮走的就是它。取消点选在连接、预检、
// 服务器打包、下载解压循环(见 countingReader / extractTarGz)，并且取消时会关闭
// 远端 session，让服务器上的 tar 立刻退出(OutputContext)。
// 取消后返回 ctx.Err()(context.Canceled),调用方据此把任务标成"已停止"而不是"失败"。
//
// 取消不会留下"看着完整"的半个镜像:.tdict-mirror.ok 只在**解压全部成功之后**才写,
// 而 MirrorReady() 判的就是这个标记。服务器侧的归档由 defer 删除;客户端被强杀时
// 由下一次拉取的陈旧清扫兜底。
func MirrorPullToDirProfileProgress(ctx context.Context, e *NamedSsh, envDir string, profile MirrorProfile, opts MirrorOptions, onProgress ProgressFunc) (*MirrorStats, error) {
	if profile == "" {
		profile = MirrorSource
	}
	if e == nil {
		return nil, fmt.Errorf("环境为空")
	}
	if strings.TrimSpace(envDir) == "" {
		return nil, fmt.Errorf("镜像目录为空")
	}
	envDir = filepath.Clean(envDir)
	report := func(p MirrorProgress) {
		if onProgress != nil {
			onProgress(p)
		}
	}
	start := time.Now()
	st := &MirrorStats{Env: e.Name, Profile: profile}
	if e.Name == "" {
		return nil, fmt.Errorf("环境名为空")
	}

	report(MirrorProgress{Phase: "connect", Message: "连接 SSH " + e.SSHConfig.Addr()})
	conn, err := Dial(e.SSHConfig)
	if err != nil {
		return nil, fmt.Errorf("连接 %s 失败: %w", e.SSHConfig.Addr(), err)
	}
	defer conn.Close()
	if err := ctx.Err(); err != nil { // 连上就被取消:别再往下走
		return nil, err
	}

	report(MirrorProgress{Phase: "probe", Message: "探测 T100 目录(zone " + e.Zone + ")"})
	top, err := mirrorTopDir(conn, e)
	if err != nil {
		return nil, err
	}
	st.TopDir = top

	// 服务器临时目录：旧版本把归档写在 $TOP，取消/失败就把 265 MB 留在 ERP 目录里，
	// 还要求登录账号对 $TOP 可写。现在归档只在一处、只活一个下载周期。
	// 这一步同时做掉三件事：解析目录、删 3 小时前的遗留、报可写与剩余空间。
	tmp, err := mirrorTmpProbe(ctx, conn, opts.TmpDir)
	if err != nil {
		return nil, err
	}
	st.TmpDir = tmp.Dir

	nonce := mirrorNonce()
	base := mirrorTempBase(tmp.Dir, e.Name, profile, nonce)
	// 从这行往后服务器上就可能有文件了：清理必须发生在**所有**返回路径上。
	// 注册在 defer conn.Close() 之后 —— LIFO 保证先删文件、再断连接。
	defer cleanRemoteArchive(conn, base)

	// 本地上次被强杀留下的 .staging-* / .old-*（只清 1 小时前的，别碰另一台正在跑的）
	cleanLocalStaging(envDir)

	// 预估归档大小：既是打包阶段进度条的分母，也是空间预检的依据。
	est := packEstimate(ctx, conn, top, profile)
	if err := mirrorCheckSpace(tmp, est); err != nil {
		return nil, err
	}

	report(MirrorProgress{Phase: "pack", Message: "服务器打包中…", Total: est})
	n, size, err := mirrorPackProgress(ctx, conn, base, top, profile, est, report)
	if err != nil {
		return nil, fmt.Errorf("服务器打包失败: %w", err)
	}
	if n == 0 {
		// 白名单内一个文件都没找到(服务器那棵树是空的)。没有归档可下,但本地
		// 仍可能残留旧版本拉下来的备份文件。
		st.Elapsed = time.Since(start).Round(100 * time.Millisecond).String()
		st.Note = "白名单内没有可打包的文件(服务器 " + profile.Label() + " 是空的?)"
		st.Pruned = pruneBackupFiles(envDir)
		if st.Pruned > 0 {
			st.Note += fmt.Sprintf(";已清理本地备份 %d 个", st.Pruned)
		}
		report(MirrorProgress{Phase: "done", Message: st.Note})
		return st, nil
	}
	st.Bytes = size

	// 下载并解压(流式):countingReader 按压缩字节回报进度
	if err := os.MkdirAll(filepath.Dir(envDir), 0o755); err != nil {
		return nil, fmt.Errorf("创建镜像父目录失败: %w", err)
	}
	sc, err := conn.SFTP()
	if err != nil {
		return nil, fmt.Errorf("打开 SFTP 失败: %w", err)
	}
	defer sc.Close()
	rf, err := sc.Open(base + ".tar.gz")
	if err != nil {
		return nil, fmt.Errorf("打开服务器归档失败: %w", err)
	}
	defer rf.Close()

	report(MirrorProgress{Phase: "download", Message: "下载并解压…", Total: size})
	cr := &countingReader{ctx: ctx, r: rf, total: size, report: report}
	files, err := extractFull(cr, envDir, profile)
	if err != nil {
		// 用户中途取消:原样返回 context.Canceled(调用方据此标"已停止",不当失败)。
		// 此时**不写基线标记**，.staging 临时目录由 extractFull 自己清掉。
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("本地解压失败: %w", err)
	}
	// tar 读到两个 0 块就停,剩下的 gzip 尾部还在管道里;排空后再比字节数,
	// 否则正常归档也会被判成"下载不完整"。
	if _, err := io.Copy(io.Discard, cr); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("读取归档尾部失败: %w", err)
	}
	// 先压缩后下载的好处:字节数是确定的,能硬校验。断网/半截下载在这里拦住,
	// 不让它变成"看着完整、其实少了一半"的镜像。
	if cr.n != size {
		return nil, fmt.Errorf("下载不完整(实收 %d 字节,应为 %d);服务器归档会被自动清理,请重试", cr.n, size)
	}
	// 解压完成、还没落任何"完成"痕迹之前再确认一次取消:
	// 否则用户刚点停止、这边刚好解压完,会留下一个标记完整但用户以为已停止的镜像。
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	st.Files = files
	if err := writeLocalMark(envDir, profile); err != nil {
		return nil, fmt.Errorf("写入本地基线标记失败: %w", err)
	}
	// 历史版本可能留下备份文件(全量解压不会带进来,但旧目录里可能有)
	st.Pruned = pruneBackupFiles(envDir)

	// 服务器归档与清单由 defer cleanRemoteArchive 删除(成功/失败/取消都走)。
	st.Elapsed = time.Since(start).Round(100 * time.Millisecond).String()
	msg := fmt.Sprintf("完成:%d 个文件", files)
	if st.Pruned > 0 {
		msg += fmt.Sprintf(",清理本地备份 %d 个", st.Pruned)
	}
	report(MirrorProgress{Phase: "done", Message: msg, Bytes: size, Total: size, Files: files})
	return st, nil
}

// countingReader 统计已读字节并节流回调进度(下载/解压阶段共用;压缩字节即下载量)。
// 同时是**取消点**:带 ctx 时每次 Read 先看有没有被取消,取消了就返回 ctx.Err(),
// 让解压链路立刻断开(不必等整个归档下完)。
type countingReader struct {
	ctx    context.Context
	r      io.Reader
	n      int64
	total  int64
	last   time.Time
	report ProgressFunc
}

func (c *countingReader) Read(p []byte) (int, error) {
	if c.ctx != nil {
		if err := c.ctx.Err(); err != nil {
			return 0, err
		}
	}
	n, err := c.r.Read(p)
	if n > 0 {
		c.n += int64(n)
	}
	now := time.Now()
	// 节流:每 200ms 一次;流结束时(EOF)必报一次最终字节数。
	// report 可能为 nil —— MirrorPullProgress 的进度回调是可选参数,
	// 而 CLI(tt file mirror)就是不传的那个;这里不判空会直接空指针崩。
	if c.report != nil && ((n > 0 && now.Sub(c.last) >= 200*time.Millisecond) || err == io.EOF) {
		c.last = now
		c.report(MirrorProgress{Phase: "download", Message: "下载并解压…", Bytes: c.n, Total: c.total})
	}
	return n, err
}

// archiveSizeRatio 源码树压成 tar.gz 的经验压缩比（源文件字节 → 归档字节）。
//
// 它只用来给「服务器打包」阶段一个**预估**分母，让进度条能动起来；界面标签写的是
// "预估"。估偏了也只影响那根条走得快慢：打包完成时前端直接跳到 100%。
// 实测这套 T100 源码树约 0.15（265 MB 归档 / ~1.8 GB 源文件）。
const archiveSizeRatio = 0.15

// packEstimate 估算这次打包的归档大小：按与 mirrorPack **完全相同**的选文件规则
// （白名单 + 备份排除）量出源文件总字节，再按经验压缩比折算。全量模式没有 `-newer`。
//
// 两个用途：打包阶段进度条的分母、以及空间预检的门槛。任何一步失败都返回 0 ——
// 拿不到预估时前端退化成"总量未知"，预检退化成 300 MB 兜底下限，都不阻断拉取。
func packEstimate(ctx context.Context, conn *SSHConn, top string, profile MirrorProfile) int64 {
	findExpr := mirrorWhitelistFindExpr(profile) + " " + mirrorBackupFindExpr()
	script := `cd ` + shdq(top) + ` 2>/dev/null || exit 0
R=""
for d in erp com; do
  [ -d "$d" ] && R="$R $d"
done
[ -z "$R" ] && exit 0
find $R -type f ` + findExpr + ` -printf '%s\n' 2>/dev/null | awk '{s+=$1} END{print s+0}'`
	if err := ctx.Err(); err != nil {
		return 0
	}
	out, err := conn.Output(script, 2*time.Minute)
	if err != nil {
		return 0
	}
	var src int64
	if _, err := fmt.Sscanf(strings.TrimSpace(lastLine(out)), "%d", &src); err != nil || src <= 0 {
		return 0
	}
	return int64(float64(src) * archiveSizeRatio)
}

// lastLine 取输出的最后一行（服务器脚本前面可能带着 find/tar 的告警）。
func lastLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

// mirrorPackProgress 在服务器打包期间轮询归档大小回报进度。
//
// 打包本身跑在**受 ctx 约束**的 goroutine 里：取消时用 OutputContext 关掉远端 session，
// 服务器上的 tar 收到 SIGHUP 立刻退出（旧实现只是本端不再等它，tar 会继续把
// 265 MB 写完 —— 在白拿一份 CPU 和磁盘 IO）。
func mirrorPackProgress(ctx context.Context, conn *SSHConn, base, top string, profile MirrorProfile, est int64, report ProgressFunc) (int, int64, error) {
	var stopped atomic.Bool
	done := make(chan struct{})
	defer func() {
		stopped.Store(true)
		close(done)
	}()
	go func() {
		t := time.NewTicker(2 * time.Second)
		defer t.Stop()
		cmd := `stat -c %s ` + shdq(base+".tar.gz") + ` 2>/dev/null || echo 0`
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-t.C:
				out, err := conn.Output(cmd, 8*time.Second)
				if err != nil {
					continue
				}
				var sz int64
				if _, err := fmt.Sscanf(strings.TrimSpace(out), "%d", &sz); err == nil && sz > 0 && !stopped.Load() && report != nil {
					report(MirrorProgress{Phase: "pack", Message: "服务器打包中…", Bytes: sz})
				}
			}
		}
	}()
	type packResult struct {
		n    int
		size int64
		err  error
	}
	ch := make(chan packResult, 1)
	go func() {
		n, size, err := mirrorPack(ctx, conn, base, top, profile)
		ch <- packResult{n, size, err}
	}()
	select {
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	case r := <-ch:
		return r.n, r.size, r.err
	}
}

// localMark 是本地基线标记(.tdict-mirror.ok)的内容。
//
//	v5 起: "v5 source,interface 2026-10-26T10:00:00Z"
//	更早:   "v4 2026-10-26T10:00:00Z"(无 profile 列表)或纯时间戳
//
// profiles 记的是“这份目录拉过哪些 profile”,它是**并集**:第二次拉另一个 profile
// 不能把前一个从记录里抹掉,否则日志/判断会把“两类都在”误报成“只有一类”。
type localMark struct {
	Version  int
	Profiles map[MirrorProfile]bool
}

func (m localMark) Has(p MirrorProfile) bool { return m.Profiles[p] }

// ProfilesLabel 按固定顺序列出已拉过的 profile(日志与测试需要稳定输出)。
func (m localMark) ProfilesLabel() string {
	var out []string
	for _, p := range []MirrorProfile{MirrorSource, MirrorInterface} {
		if m.Profiles[p] {
			out = append(out, string(p))
		}
	}
	if len(out) == 0 {
		return "(无记录)"
	}
	return strings.Join(out, ",")
}

// readLocalMark 读本地基线标记;无标记、旧格式、字段缺失都返回零值(版本 0、空集合),
// 调用方据此当作“无基线”处理,不会凭空认为镜像完整。
func readLocalMark(envDir string) localMark {
	m := localMark{Profiles: map[MirrorProfile]bool{}}
	b, err := os.ReadFile(filepath.Join(envDir, localMarkName))
	if err != nil {
		return m
	}
	fields := strings.Fields(strings.TrimSpace(string(b)))
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "v") {
		return m
	}
	if n, err := strconv.Atoi(strings.TrimPrefix(fields[0], "v")); err == nil {
		m.Version = n
	}
	for _, f := range fields[1:] {
		// profile 集合写成一个逗号分隔的字段（source,interface）；也容忍空格分隔的旧写法。
		for _, name := range strings.Split(f, ",") {
			switch p := MirrorProfile(strings.ToLower(strings.TrimSpace(name))); p {
			case MirrorSource, MirrorInterface:
				m.Profiles[p] = true
			}
		}
	}
	return m
}

// writeLocalMark 在镜像环境目录写入完整基线标记(白名单版本 + 已拉 profile 集合 + 时间)。
// profile 取并集:同目录先后拉两种 profile 后,记录里两种都在。
func writeLocalMark(envDir string, profile MirrorProfile) error {
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		return err
	}
	mark := readLocalMark(envDir)
	if mark.Version != mirrorWhitelistVersion {
		// 版本升级:旧的 profile 记录不再作数(白名单变了,那一类可能得重新拉),
		// 但保留本次这个 —— 它刚刚被完整解压过。
		mark.Profiles = map[MirrorProfile]bool{}
	}
	if mark.Profiles == nil {
		mark.Profiles = map[MirrorProfile]bool{}
	}
	if profile != "" {
		mark.Profiles[profile] = true
	}
	return os.WriteFile(filepath.Join(envDir, localMarkName),
		[]byte(fmt.Sprintf("v%d %s %s\n", mirrorWhitelistVersion, mark.ProfilesLabel(),
			time.Now().UTC().Format(time.RFC3339))), 0o644)
}

// localMarkVersion 读本地基线标记里的白名单版本;无标记或旧格式(纯时间戳)返回 0。
func localMarkVersion(envDir string) int {
	return readLocalMark(envDir).Version
}

// MirrorEnvDir 返回某环境的本地镜像目录(mirrorDir/<环境名>)。
func MirrorEnvDir(mirrorDir, envName string) string {
	if mirrorDir == "" || envName == "" {
		return ""
	}
	return filepath.Join(mirrorDir, envName)
}

// MirrorReady 本地是否已有完整镜像(.tdict-mirror.ok 标记存在)。
// 只用于界面显示"拉过没有";拉取本身不依赖它(每次都是全量)。
func MirrorReady(mirrorDir, envName string) bool {
	return MirrorReadyDir(MirrorEnvDir(mirrorDir, envName))
}

// MirrorReadyDir 判一个“最终镜像目录”有没有完整基线。
func MirrorReadyDir(envDir string) bool {
	if strings.TrimSpace(envDir) == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(envDir, localMarkName))
	return err == nil
}

func logfMirror(format string, args ...any) {
	// 轻量日志:镜像失败清理之类不致命信息走 stderr(无全局 logger 依赖)
	fmt.Fprintf(os.Stderr, "[mirror] "+format+"\n", args...)
}

// mirrorWhitelistFindExpr 服务器侧 find 的白名单。
//
// source 的四类各自限定：4gl 用后缀级限制 `-path '*/4gl/*' -iname '*.4gl'`（因此
// a.4gl.src、a.4gltest、a.4gl.20201110 都不会进来）；42f 与 42s 按目录限定（42s 只收
// zh_CN 语言目录）；*.inc 不限定目录（技能文档里的 inc 就在 4g1/、$ERP/cfg、$COM/cfg 多处）。
//
// **source 必须包含 4gl + 42f + 42s + inc 四类**：一个环境只有一个镜像目录，tt outline
// 只认 4gl、tt interface 只认 42f/42s，两者读同一个目录；少收一类就等于让某个命令
// 静默断粮（它只会报“找不到文件”，看起来像文件不存在）。
//
// 改这里必须同步 mirrorPathInProfile（本地分类器，全量解压时用它决定“哪些文件属于
// 本 profile、要以服务器为准”），mirror_test.go 里有交叉验证。
func mirrorWhitelistFindExpr(profile MirrorProfile) string {
	if profile == MirrorInterface {
		return `\( -path '*/42f/*' -o -path '*/42s/zh_CN/*' -o -path '*/42s/*/zh_CN/*' \)`
	}
	return `\( -path '*/4gl/*' -iname '*.4gl' ` +
		`-o -path '*/42f/*' ` +
		`-o -path '*/42s/zh_CN/*' -o -path '*/42s/*/zh_CN/*' ` +
		`-o -iname '*.inc' \)`
}

// mirrorPathInProfile 判断一个**本地**相对路径是否落在该 profile 的白名单内。
// 只给全量解压用：属于本 profile 的文件以服务器为准（新目录里没有就是服务器删了），
// 不属于本 profile 的文件原样保留（不因“给 A 做了全量”而删掉 B）。
//
// 语义必须与 mirrorWhitelistFindExpr 一致：/"/" 前缀归一化后，路径分段判断。
func mirrorPathInProfile(profile MirrorProfile, rel string) bool {
	p := strings.ToLower("/" + strings.TrimPrefix(filepath.ToSlash(rel), "./"))
	is42f := strings.Contains(p, "/42f/")
	// 42s 只收 zh_CN：浅层 <模块>/42s/zh_CN/* 与更深层 <模块>/42s/<...>/zh_CN/* 都算。
	is42s := strings.Contains(p, "/42s/") && strings.Contains(p, "/zh_cn/")
	if profile == MirrorInterface {
		return is42f || is42s
	}
	is4gl := strings.Contains(p, "/4gl/") && strings.HasSuffix(p, ".4gl")
	isInc := strings.HasSuffix(p, ".inc")
	return is4gl || is42f || is42s || isInc
}

// mirrorWhitelistVersion 白名单版本:新增文件类型时 +1。
//
// 只写进本地标记(.tdict-mirror.ok)供排查:全量模式下**不需要**靠版本判定补拉 ——
// 每次拉取本来就会重压全部白名单文件,新增的类自然就进来了。
//
//	v1: 4gl/4fd
//	v2: +42s/zh_CN、*.inc
//	v3: 4fd → 42f（前端画面改为 42f+42s；4fd 不再镜像）
//	v4: source 收紧为 4gl/**/*.4gl；interface 保留 42f/42s
//	v5: source 恢复为 4gl + 42f + 42s/zh_CN + *.inc（一个镜像目录必须装齐所有读者的并集）
//	v6: 删除增量,只剩全量拉取(服务器临时目录 + 四层清理)
const mirrorWhitelistVersion = 6

// mirrorBackupSuffixes 备份/临时文件名特征(不含点全名包含即算备份)。
// 服务器侧打包排除与本地残留清理共用同一份列表,避免两边规则漂移:
//   - 框架/人工备份:x.4gl.bak、x.42f.bak、x.bck、x.bck1、x.bck2 …
//   - 编辑器临时文件:x.4gl~、x.swp、x.tmp 等
var mirrorBackupSuffixes = []string{".bak", ".bck", ".old", ".orig", ".tmp", ".swp", ".swo"}

// isBackupFile 判断文件名是否备份/临时文件(与服务器侧 find 排除规则一致)。
func isBackupFile(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return false
	}
	if strings.HasSuffix(n, "~") {
		return true
	}
	for _, suf := range mirrorBackupSuffixes {
		if strings.Contains(n, suf) {
			return true
		}
	}
	return false
}

// mirrorBackupFindExpr 生成服务器侧 find 的备份排除条件(与 isBackupFile 同语义)。
func mirrorBackupFindExpr() string {
	parts := make([]string, 0, len(mirrorBackupSuffixes)+1)
	for _, suf := range mirrorBackupSuffixes {
		parts = append(parts, `-not -iname '*`+suf+`*'`)
	}
	parts = append(parts, `-not -iname '*~'`)
	return strings.Join(parts, " ")
}

// pruneBackupFiles 删除本地镜像目录内残留的备份文件(旧版本拉取过),返回删除数。
// 只删除文件,不动目录;root 不存在时静默返回 0。
func pruneBackupFiles(root string) int {
	removed := 0
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		if isBackupFile(d.Name()) {
			if os.Remove(path) == nil {
				removed++
			}
		}
		return nil
	})
	return removed
}

// mirrorTopDir 确定打包根:仅 SSH 登录探测(Runtime.TOP)——T100 路径无静态配置,
// 一律按登录区域动态获取;探测失败即报错。
func mirrorTopDir(conn *SSHConn, e *NamedSsh) (string, error) {
	if conn == nil || e.Zone == "" {
		return "", fmt.Errorf("环境 %q 缺少登录区域(zone),无法探测 T100 目录", e.Name)
	}
	env, err := ProbeTEnv(conn, e.Zone)
	if err != nil {
		return "", fmt.Errorf("环境 %s(zone %s)登录探测 T100 目录失败: %w", e.Name, e.Zone, err)
	}
	return strings.TrimSuffix(env.TOP, "/"), nil
}

// shdq 把值包成双引号 shell 字面量(路径含特殊字符安全)。
func shdq(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "$", `\$`)
	s = strings.ReplaceAll(s, "`", "\\`")
	return `"` + s + `"`
}

// mirrorPack 在服务器临时目录里打一个**全量** tar.gz。返回 (文件数, 归档字节);
// 白名单内一个文件都没有时返回 (0, 0) 且不产生归档。
//
// 脚本里没有 `set -e`：每一步自己收尾。这是"打包失败也不留垃圾"的第一层 ——
// 清单在成功/失败两条路上都删，tar 失败连半截归档一起删。
// 归档与清单都在调用方给的 base 上（临时目录，见 mirrorTempBase），任务结束后
// 由 cleanRemoteArchive 删除；这里只负责别把半成品留在服务器上。
//
// 输出标记（客户端按最后一个标记判定）：
//
//	TDICT_MIRROR_OK <文件数> <字节>
//	TDICT_MIRROR_NONE          白名单内没有文件
//	TDICT_MIRROR_CD            进不去 TOP
//	TDICT_MIRROR_FINDFAIL      find 失败（清单已删）
//	TDICT_MIRROR_TARFAIL       tar 失败（半截归档与清单已删）
func mirrorPack(ctx context.Context, conn *SSHConn, base, top string, profile MirrorProfile) (int, int64, error) {
	out, err := conn.OutputContext(ctx, mirrorPackScript(base, top, profile), 20*time.Minute)
	if err != nil {
		if ctx.Err() != nil {
			return 0, 0, ctx.Err()
		}
		return 0, 0, fmt.Errorf("%s(输出: %s)", err, strings.TrimSpace(out))
	}
	return parseMirrorPackOutput(out, top)
}

// mirrorPackScript 生成服务器侧的打包脚本。单独成函数是为了能被单测钉住：
// 这里少一句 `rm -f`，服务器上就会多一个 265 MB 的文件，而且很难在生产环境复现。
func mirrorPackScript(base, top string, profile MirrorProfile) string {
	o := shdq(base + ".tar.gz")
	l := shdq(base + ".list")
	t := shdq(top)
	// exec 通道不经过登录 profile，这里只依赖绝对路径与标准工具。
	// 白名单树(按 profile) + 排除备份/临时文件(如 x.4gl.bak、x.bck、x.bck1)。
	findExpr := mirrorWhitelistFindExpr(profile) + " " + mirrorBackupFindExpr()
	return `cd ` + t + ` || { echo TDICT_MIRROR_CD; exit 9; }
R=""
for d in erp com; do
  [ -d "$d" ] && R="$R $d"
done
if [ -z "$R" ]; then
  echo TDICT_MIRROR_NONE
  exit 0
fi
if ! find $R -type f ` + findExpr + ` > ` + l + `; then
  rm -f ` + l + `
  echo TDICT_MIRROR_FINDFAIL
  exit 8
fi
if [ ! -s ` + l + ` ]; then
  rm -f ` + l + `
  echo TDICT_MIRROR_NONE
  exit 0
fi
n=$(wc -l < ` + l + `)
if ! tar -czf ` + o + ` -T ` + l + `; then
  rm -f ` + o + ` ` + l + `
  echo TDICT_MIRROR_TARFAIL
  exit 0
fi
rm -f ` + l + `
echo "TDICT_MIRROR_OK $n $(wc -c < ` + o + `)"`
}

// parseMirrorPackOutput 解析打包脚本的输出，取**最后**一个标记行。
//
// 取最后一行的原因：tar 会把警告打到同一个通道里，前面的行不可信。
func parseMirrorPackOutput(out, top string) (int, int64, error) {
	mark := ""
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "TDICT_MIRROR_OK"),
			line == "TDICT_MIRROR_NONE",
			line == "TDICT_MIRROR_CD",
			line == "TDICT_MIRROR_FINDFAIL",
			line == "TDICT_MIRROR_TARFAIL":
			mark = line
		}
	}
	switch {
	case strings.HasPrefix(mark, "TDICT_MIRROR_OK"):
		var n int
		var s int64
		if _, err := fmt.Sscanf(mark, "TDICT_MIRROR_OK %d %d", &n, &s); err != nil {
			return 0, 0, fmt.Errorf("解析打包标记失败: %q", mark)
		}
		return n, s, nil
	case mark == "TDICT_MIRROR_NONE":
		return 0, 0, nil
	case mark == "TDICT_MIRROR_CD":
		return 0, 0, fmt.Errorf("进不去服务器 TOP: %s", top)
	case mark == "TDICT_MIRROR_FINDFAIL":
		return 0, 0, fmt.Errorf("服务器 find 失败(清单已清理)")
	case mark == "TDICT_MIRROR_TARFAIL":
		return 0, 0, fmt.Errorf("服务器 tar 打包失败(半截归档已清理,通常是磁盘满或权限)")
	default:
		return 0, 0, fmt.Errorf("服务器脚本未输出预期标记(输出: %s)", strings.TrimSpace(out))
	}
}

// serverTmp 服务器临时目录的探测结果。
type serverTmp struct {
	Dir    string // 解析后的绝对路径
	FreeKB int64  // 剩余空间(df -Pk 的可用列,KB;0=没量到)
}

// mirrorTmpProbe 解析并检查服务器临时目录，顺便扫掉 3 小时前的遗留归档。
//
// 为什么必须走一趟服务器：
//   - `${TMPDIR:-/tmp}` 是 shell 的展开，客户端拿不到最终路径；
//   - 不可写要立刻报清楚，而不是让 tar 跑一半报 Permission denied；
//   - 剩余空间装不下时直接拒绝，避免"压到一半 ENOSPC"留下半截归档；
//   - **3 小时清扫是唯一能救"客户端被强杀/断网"的机制** —— 那种情况没人执行删除。
//
// 清扫只认 `.tt-mirror-*` 且 `-mmin +180`：不会碰另一台机器正在下载的归档。
func mirrorTmpProbe(ctx context.Context, conn *SSHConn, want string) (serverTmp, error) {
	out, err := conn.Output(mirrorTmpProbeScript(want), 2*time.Minute)
	line := ""
	for _, ln := range strings.Split(out, "\n") {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "TDICT_MIRROR_TMP") {
			line = ln
		}
	}
	switch {
	case strings.HasPrefix(line, "TDICT_MIRROR_TMPMISSING"):
		dir := strings.TrimPrefix(line, "TDICT_MIRROR_TMPMISSING ")
		return serverTmp{}, fmt.Errorf("服务器临时目录不存在: %s；请清理/创建它,或在设置里用 mirror.tmpDir 指定别的目录", dir)
	case strings.HasPrefix(line, "TDICT_MIRROR_TMPNOWRITE"):
		dir := strings.TrimPrefix(line, "TDICT_MIRROR_TMPNOWRITE ")
		return serverTmp{}, fmt.Errorf("服务器临时目录不可写: %s；可用 mirror.tmpDir 换一个落点", dir)
	case strings.HasPrefix(line, "TDICT_MIRROR_TMP "):
		rest := strings.TrimPrefix(line, "TDICT_MIRROR_TMP ")
		parts := strings.SplitN(rest, " ", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
			return serverTmp{}, fmt.Errorf("服务器返回的临时目录格式不认识: %q", line)
		}
		free, _ := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		return serverTmp{Dir: strings.TrimSpace(parts[1]), FreeKB: free}, nil
	}
	if err != nil {
		return serverTmp{}, fmt.Errorf("探测服务器临时目录失败: %w(输出: %s)", err, strings.TrimSpace(out))
	}
	return serverTmp{}, fmt.Errorf("服务器未返回临时目录信息(输出: %s)", strings.TrimSpace(out))
}

// mirrorTmpProbeScript 生成临时目录探测脚本（单独成函数，便于单测钉住清扫规则）。
func mirrorTmpProbeScript(want string) string {
	// want 为空时让服务器自己展开 ${TMPDIR:-/tmp}：exec 通道不读登录 profile，
	// TMPDIR 通常是空的，于是落到 /tmp。
	spec := `T=${TMPDIR:-/tmp}`
	if w := strings.TrimSpace(want); w != "" {
		spec = `T=` + shdq(w)
	}
	return spec + `
[ -d "$T" ] || { echo "TDICT_MIRROR_TMPMISSING $T"; exit 3; }
[ -w "$T" ] || { echo "TDICT_MIRROR_TMPNOWRITE $T"; exit 4; }
find "$T"/.tt-mirror-* -type f -mmin +180 -exec rm -f {} + 2>/dev/null
free=$(df -Pk "$T" 2>/dev/null | awk 'NR==2{print $4}')
[ -n "$free" ] || free=0
echo "TDICT_MIRROR_TMP $free $T"`
}

// mirrorCheckSpace 空间预检：剩余空间 < 预估×1.2 就拒绝（拿不到预估时按 300 MB 兜底）。
//
// 宁可提前失败，也不要在服务器上压到 ENOSPC：那时留下的是半截归档 + 一个失败任务，
// 还得靠 3 小时清扫收尾。`df` 没量到（FreeKB=0）时不拦，交给 tar 自己报错。
func mirrorCheckSpace(tmp serverTmp, est int64) error {
	need := est
	if need <= 0 {
		need = 300 << 20
	} else {
		need += need / 5 // ×1.2：给 tar 的块缓冲和清单留点余量
	}
	free := tmp.FreeKB << 10
	if free > 0 && free < need {
		return fmt.Errorf("服务器临时目录空间不足: %s 可用 %d MB,本次归档约需 %d MB；请清理该目录或用 mirror.tmpDir 换一个落点",
			tmp.Dir, tmp.FreeKB/1024, need/(1<<20))
	}
	return nil
}

// cleanRemoteArchive 删除本次拉取在服务器临时目录里留下的归档与清单。
//
// 挂在 defer 上：成功、失败、取消都走。必须幂等 —— 文件可能压根没生成（白名单空、
// 预检失败），也可能被脚本自己删过了，`rm -f` 两种情况都不报错。
// 发不出去（连接已断）时只能靠下一次拉取的 3 小时清扫。
func cleanRemoteArchive(conn *SSHConn, base string) {
	if _, err := conn.Output(mirrorCleanScript(base), time.Minute); err != nil {
		logfMirror("清理服务器临时归档失败(3 小时后的下一次拉取会清扫): %v", err)
	}
}

// mirrorCleanScript 生成"杀掉本次拉起的服务器进程 + 删掉本次两个临时文件"的命令。
//
// 为什么还要 kill：取消时关掉 SSH 会话**不足以**让远端停下 —— 实测 OpenSSH 对非 pty
// 的 exec 通道，会话关掉后 ksh 与 tar 仍会继续跑（PPID 变 1），把 265 MB 写完。
// 于是"取消"反而留下一个完整归档，比不取消还糟。
//
// 按 base 里的随机 nonce 反查进程：唯一、不会误伤别人的进程；只认 tar/sh 两类
// （ps/awk/grep 自己的命令行里也带着这个串，必须排除）。先 TERM 再 KILL，最后删文件。
func mirrorCleanScript(base string) string {
	return `B=` + shdq(base) + `
me=$$
for p in $(ps -ef 2>/dev/null | awk -v me="$me" -v b="$B" 'index($0,b)>0 && $2!=me && ($8=="tar" || $8=="ksh" || $8=="sh" || $8=="bash") {print $2}'); do
  kill -TERM "$p" 2>/dev/null
done
sleep 1
for p in $(ps -ef 2>/dev/null | awk -v me="$me" -v b="$B" 'index($0,b)>0 && $2!=me && ($8=="tar" || $8=="ksh" || $8=="sh" || $8=="bash") {print $2}'); do
  kill -KILL "$p" 2>/dev/null
done
rm -f "$B.tar.gz" "$B.list"`
}

// cleanLocalStaging 删掉本地上一次被强杀留下的 `.staging-*` / `.old-*` 目录。
//
// extractFull 的正常与失败路径都会自己 RemoveAll，只有客户端被强杀才留得下来。
// 只删 1 小时前的：镜像根目录下可能正有另一台机器（或另一个进程）在写 staging，
// 见一个删一个会把人家正在解压的目录端掉。
func cleanLocalStaging(envDir string) {
	parent := filepath.Dir(envDir)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-time.Hour)
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || (!strings.HasPrefix(name, ".staging-") && !strings.HasPrefix(name, ".old-")) {
			continue
		}
		info, ierr := e.Info()
		if ierr != nil || !info.ModTime().Before(cutoff) {
			continue
		}
		p := filepath.Join(parent, name)
		if err := os.RemoveAll(p); err != nil {
			logfMirror("清理本地残留目录失败(忽略): %s: %v", p, err)
			continue
		}
		logfMirror("已清理本地残留目录: %s", p)
	}
}

// extractFull 全量解压:先写 .staging 临时目录,再把**本 profile 之外**的旧文件搬进去,
// 最后整目录替换。
//
// 为什么不能“旧目录整个删掉”:一个环境只有一个镜像目录,source(4gl/42f/42s/inc)与
// interface(42f/42s)可能先后拉进同一个目录。旧实现直接删旧目录,于是“给 profile A
// 做一次全量”会把 profile B 的文件一起删掉 —— 用户看到的是“我只拉了一类,另一类没了”。
// 全量的语义应当是“本 profile 的白名单内以服务器为准”,不是“清空别人的东西”。
func extractFull(r io.Reader, envDir string, profile MirrorProfile) (int, error) {
	ts := time.Now().Format("20060102-150405.000")
	staging := filepath.Join(filepath.Dir(envDir), ".staging-"+ts)
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return 0, err
	}
	n, err := extractTarGz(r, staging)
	if err != nil {
		os.RemoveAll(staging)
		return 0, err
	}
	// 另一类的既有文件先搬进新目录:全量只对本 profile 的白名单负责。
	if kept, kerr := keepOtherProfileFiles(envDir, staging, profile); kerr != nil {
		os.RemoveAll(staging)
		return 0, kerr
	} else if kept > 0 {
		logfMirror("全量拉取保留另一类文件 %d 个(不属于 %s 的白名单)", kept, profile)
	}
	old := envDir + ".old-" + ts
	if _, err := os.Stat(envDir); err == nil {
		if err := os.Rename(envDir, old); err != nil {
			os.RemoveAll(staging)
			return 0, fmt.Errorf("备份旧镜像目录失败: %w", err)
		}
	}
	if err := os.Rename(staging, envDir); err != nil {
		// 回滚:把旧目录放回去
		if _, serr := os.Stat(old); serr == nil {
			os.Rename(old, envDir)
		}
		os.RemoveAll(staging)
		return 0, fmt.Errorf("启用新镜像目录失败: %w", err)
	}
	if _, err := os.Stat(old); err == nil {
		if err := os.RemoveAll(old); err != nil {
			logfMirror("删除旧镜像备份失败(可手动清理 %s): %v", old, err)
		}
	}
	return n, nil
}

// keepOtherProfileFiles 把 oldDir 里**不属于当前 profile** 的普通文件搬进 staging。
//
// 几条规矩：
//   - staging 已有同名文件 = 服务器也给了它 → 以服务器为准，不覆盖；
//   - 备份/临时文件不搬（它们由 pruneBackupFiles 清理）；
//   - 基线标记 .tdict-mirror.ok 不搬（新目录会在解压完成后重新写，且要记上本 profile）；
//   - .staging-* / .old-* 这些临时目录跳过（历史失败留下时不当内容）。
func keepOtherProfileFiles(oldDir, staging string, profile MirrorProfile) (int, error) {
	if _, err := os.Stat(oldDir); err != nil {
		return 0, nil // 首次拉取:没有旧目录
	}
	kept := 0
	err := filepath.WalkDir(oldDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // 单个文件读不到不该打断整个重建
		}
		if d.IsDir() {
			name := d.Name()
			if p != oldDir && (strings.HasPrefix(name, ".staging-") || strings.HasPrefix(name, ".old-")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() || isBackupFile(d.Name()) {
			return nil
		}
		rel, rerr := filepath.Rel(oldDir, p)
		if rerr != nil || rel == localMarkName {
			return nil
		}
		if mirrorPathInProfile(profile, rel) {
			return nil // 属于本 profile:以服务器为准(新目录里没有就是服务器删了)
		}
		dst := filepath.Join(staging, rel)
		if _, serr := os.Stat(dst); serr == nil {
			return nil
		}
		if err := copyRegularFile(p, dst); err != nil {
			return err
		}
		kept++
		return nil
	})
	return kept, err
}

// copyRegularFile 复制单个普通文件（保留 mtime，与原解压行为一致）。
func copyRegularFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if st, err := in.Stat(); err == nil {
		_ = os.Chtimes(dst, st.ModTime(), st.ModTime())
	}
	return nil
}

// extractTarGz 从 tar.gz 流解压到 dest(标准库;不做任何 shell)。返回文件数。
// tar 头路径做穿越防护;回写 mtime 便于与服务器比对;文件 0644。
func extractTarGz(r io.Reader, dest string) (int, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return 0, fmt.Errorf("解压 gzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	files := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return files, err
		}
		name := filepath.Clean(filepath.FromSlash(hdr.Name))
		if name == "." || filepath.IsAbs(name) || name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return files, fmt.Errorf("归档含非法路径: %q", hdr.Name)
		}
		target := filepath.Join(dest, name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return files, err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return files, err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return files, err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return files, err
			}
			f.Close()
			_ = os.Chtimes(target, hdr.ModTime, hdr.ModTime)
			files++
		default:
			// 符号链接/设备等不落盘(白名单内不会出现)
		}
	}
	return files, nil
}
