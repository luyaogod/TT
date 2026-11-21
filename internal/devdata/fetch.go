package devdata

// 按需资料（D）的拉取：一个程序、一套材料（plan ADR-04）。
//
// 只有"程序"一个入口，一次拉齐 dzx/src/inc（可选 form/design）。为什么不做散装下载：
// model 的第一类结论（"这一行为什么长这样"）依赖 tab（蓝图）与 tgl（排序骨架）来自
// 同一时刻；散装会让两者错配，而错配的产物看起来完全正常，只是答案微妙地不对。
//
// 版本锚点（manifest.versions）来自**本环境的共用资料**（mirror 拉的那份）：
//   - templateVersions ← `<env>/mdl` 里各模板的 `(Version:N)`
//   - dzaa002        ← `<env>/dict/<env>.db` 里该程序的最大有效版次
// 共用资料没同步时这两个锚点为空，只写一条 warning —— 不因为"判不了新鲜度"就拒绝
// 给材料：材料本身仍然是完整的，只是过不过期这件事要等 mirror 之后才能回答。

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/luyaogod/TT/internal/config"
)

// FetchRequest 一次按程序拉取的输入。
type FetchRequest struct {
	CfgPath string
	Env     string
	Prog    string
	// Module 模块目录名（erp/<mod>/…）。空 = 从程序名前三位推断并验证。
	Module string
	Ent    int
	// WithDesign 把本地工作区里的设计稿 `<prog>*.tzs` 一并归档。
	WithDesign bool
	// Form 连同 `.42f` 一起拉（二进制、体积大，默认不拉）。
	Form bool
	// Force 目标程序目录已存在时是否覆盖。false → ErrExists。
	Force     bool
	TTVersion string

	// Files 仅供测试注入；nil 时按环境打开真 SFTP。
	Files FileSource
}

// FetchResult 一次拉取的结果。
type FetchResult struct {
	Env    string `json:"env"`
	Prog   string `json:"prog"`
	Module string `json:"module"`
	Dir    string `json:"dir"`
	// Files 落盘文件清单（与 manifest.files 同形）。
	Files []FileEntry `json:"files"`
	// Missing 可选素材的缺失项（不进 missing 的就是完整）。
	Missing  []string `json:"missing,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	// TemplateVersions / DZAA002 版本锚点（空 = 共用资料未同步）。
	TemplateVersions map[string]int `json:"templateVersions,omitempty"`
	DZAA002          int            `json:"dzaa002,omitempty"`
	Bytes            int64          `json:"bytes"`
	Elapsed          string         `json:"elapsed"`
}

// reModule 模块目录名白名单（3 位小写字母为主，允许数字/下划线防意外）。
var reModule = regexp.MustCompile(`^[a-z][a-z0-9_]{1,15}$`)

// GuessModule 从程序名前三位推断模块目录（`axmt500_wf` → `axm`）。
//
// T100 的模块目录就是程序编号的前三位。这条推断**必须被验证**（见 resolveModule）：
// 推断错了会去一个不存在的目录里找文件，最后表现成"这个程序没有蓝图"。
func GuessModule(prog string) string {
	p := strings.TrimSpace(prog)
	if len(p) < 3 {
		return ""
	}
	return p[:3]
}

// Fetch 执行一次按程序拉取。onProgress 可为 nil。
func Fetch(ctx context.Context, req FetchRequest, onProgress ProgressFunc) (*FetchResult, error) {
	start := time.Now()
	if err := ValidProgName(req.Prog); err != nil {
		return nil, err
	}
	prog := strings.TrimSpace(req.Prog)
	e, err := envInfo(req.CfgPath, req.Env)
	if err != nil {
		return nil, err
	}
	progDir, err := ProgDir(req.CfgPath, e.Name, prog)
	if err != nil {
		return nil, err
	}
	report := progressReporter(onProgress)

	// 已存在：默认拒绝（"整包替换要显式"），force 才覆盖。
	if _, statErr := os.Stat(progDir); statErr == nil && !req.Force {
		return nil, fmt.Errorf("%w: %s（加 --force 重新拉取）", ErrExists, progDir)
	}

	tmpDir, err := NewTempDir(progDir)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			CleanupTemp(tmpDir, "")
		}
	}()

	src := req.Files
	if src == nil {
		report(Progress{Phase: "connect", Message: "连接服务器…"})
		var closeAll func()
		src, closeAll, err = OpenFileSource(e)
		if err != nil {
			return nil, err
		}
		defer closeAll()
	}

	mod, err := resolveModule(ctx, src, prog, req.Module)
	if err != nil {
		return nil, err
	}

	m := NewManifest(KindFetch, e.Name)
	m.Prog = prog
	m.Module = mod
	m.TTVersion = req.TTVersion
	m.Source.TopDir = src.Top()
	m.Source.Zone = e.Zone
	m.Source.Ent = req.Ent

	res := &FetchResult{Env: e.Name, Prog: prog, Module: mod, Dir: progDir}
	roles := FetchRoles[:0:0]
	for _, r := range FetchRoles {
		if r.Form && !req.Form {
			continue
		}
		roles = append(roles, r)
	}
	total := len(roles)
	for i, r := range roles {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		srcRel := r.SrcPath(mod, prog)
		rel := r.RelPath(prog)
		report(Progress{Phase: "fetch", Message: "下载 " + srcRel, Done: i, Total: total, CurrentFile: rel})
		st, err := src.Stat(ctx, srcRel)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) || errors.Is(err, os.ErrNotExist) {
				if r.Required {
					// 必需项缺失：先分清"这个程序根本不存在"与"程序在、但那件材料没产"
					// —— 前者是 404（调用方要改程序名），后者是 404 加一句具体缺什么。
					// 半套材料一律不落盘：让 model 拿到一半会得出"看起来对"的错结论。
					if !programExists(ctx, src, mod, prog) {
						return nil, fmt.Errorf("%w: 服务器模块 %s 下找不到程序 %s（tab/tgl/4gl 都没有）",
							ErrProgNotFound, mod, prog)
					}
					return nil, fmt.Errorf("%w: 程序 %s 缺必需素材 %s（服务器上没有 %s）",
						ErrProgNotFound, prog, rel, srcRel)
				}
				m.AddMissing(rel)
				continue
			}
			return nil, err
		}
		if st.IsDir {
			// 同名目录（设计器偶尔把 xxx.tab 建成目录）：按缺失处理，不要往下钻
			if r.Required {
				return nil, fmt.Errorf("必需素材 %s 在服务器上是个目录，不是文件", srcRel)
			}
			m.AddMissing(rel)
			continue
		}
		if st.Size > MaxFileBytes {
			m.AddMissing(rel)
			warn := fmt.Sprintf("%s 有 %s，超过单文件 %d MB 上限，已跳过", rel, humanBytes(st.Size), MaxFileBytes>>20)
			m.AddWarning(warn)
			continue
		}
		data, err := src.ReadFile(ctx, srcRel, MaxFileBytes)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) || errors.Is(err, os.ErrNotExist) {
				if r.Required {
					return nil, fmt.Errorf("必需素材缺失：%s（读取时已不存在）", srcRel)
				}
				m.AddMissing(rel)
				continue
			}
			return nil, fmt.Errorf("下载 %s 失败: %w", srcRel, err)
		}
		if err := writeLocalFile(filepath.Join(tmpDir, filepath.FromSlash(rel)), data); err != nil {
			return nil, err
		}
		entry := FileEntry{
			Role: r.Role, Src: srcRel, Rel: rel,
			SHA256: BytesSHA256(data), Size: int64(len(data)),
			SrcMtime: st.ModTime.Format(time.RFC3339),
		}
		m.AddFile(entry)
		res.Files = append(res.Files, entry)
		res.Bytes += int64(len(data))
		if r.Role == "tgl" {
			m.Versions.TglHeader = ParseTglHeader(data)
		}
	}

	// 设计稿（只做本地归档，不解析）：plan ADR-05 的同一条纪律
	if req.WithDesign {
		report(Progress{Phase: "write", Message: "归档设计稿…", Done: total, Total: total})
		designFiles, designMissing, designWarn, err := archiveDesign(tmpDir, e.Workspace, prog)
		if err != nil {
			return nil, err
		}
		m.Files = append(m.Files, designFiles...)
		res.Files = append(res.Files, designFiles...)
		for _, x := range designMissing {
			m.AddMissing(x)
		}
		for _, w := range designWarn {
			m.AddWarning(w)
		}
		if len(designFiles) == 0 && len(designMissing) == 0 {
			m.AddWarning("--with-design 没有找到设计稿（" + describeWorkspace(e.Workspace) + "）")
		}
	}

	// 版本锚点：版次**在线查**（取不到才看本地镜像），模板版本来自本地共用资料
	tv, dzaa := versionAnchors(ctx, req.CfgPath, e.Name, prog, req.Ent)
	m.Versions.TemplateVersions = tv
	m.Versions.DZAA002 = dzaa
	if len(tv) == 0 {
		m.AddWarning("没有共用资料：模板版本锚点缺失（要看模板版本需先 tt gen mirror --only mdl）")
	}
	if dzaa == 0 {
		m.AddWarning("取不到该程序的版次（在线库不可用、本地也没有共用资料）：无法判断规格是否换版")
	}
	ConsistencyWarnings(m)

	m.FetchedAt = nowStamp()
	if err := WriteManifest(tmpDir, m); err != nil {
		return nil, err
	}
	if err := CommitDir(tmpDir, progDir, req.Force); err != nil {
		return nil, err
	}
	committed = true
	CleanupTemp("", progDir)

	res.Missing = m.Missing
	res.Warnings = m.Warnings
	res.TemplateVersions = tv
	res.DZAA002 = dzaa
	res.Elapsed = time.Since(start).Round(10 * time.Millisecond).String()
	report(Progress{Phase: "done", Message: fmt.Sprintf("拉取完成：%d 个文件 / %s", len(res.Files), humanBytes(res.Bytes)), Done: total, Total: total})
	return res, nil
}

// programExists 判"服务器上到底有没有这个程序"：三个必需角色任意一个存在就算有。
//
// 判据只用 Status（元数据），不读内容：这里是在决定"回 404 还是回'材料不齐'"。
func programExists(ctx context.Context, src FileSource, mod, prog string) bool {
	for _, r := range FetchRoles {
		if !r.Required {
			continue
		}
		if _, err := src.Stat(ctx, r.SrcPath(mod, prog)); err == nil {
			return true
		}
	}
	return false
}

// resolveModule 决定模块目录：显式给的必须存在，没给的按前三位推断并验证。
//
// 验证有两条判据，从宽到严：
//  1. `erp/<mod>/dzx` 目录在 —— 模块装了就有；
//  2. 蓝图文件本身在 —— 少数站点没有 dzx 目录（早期版本）。
//
// 两条都不成立才判"推断不出来"。**不要**只判蓝图文件：蓝图缺失本身就是一种要报
// "必需素材缺失"的状态，提前把它误报成"模块推断不出来"会把用户引向错误的动作
// （去改 --module，而不是去查为什么没蓝图）。
func resolveModule(ctx context.Context, src FileSource, prog, explicit string) (string, error) {
	mod := strings.ToLower(strings.TrimSpace(explicit))
	if mod != "" {
		if !reModule.MatchString(mod) {
			return "", fmt.Errorf("%w: 模块名 %q（应为小写字母开头、仅含小写字母/数字/下划线）", ErrInvalidOpt, explicit)
		}
		return mod, nil
	}
	guess := GuessModule(prog)
	if guess == "" {
		return "", fmt.Errorf("%w: 无法从 %q 推断模块，请用 --module 指定", ErrInvalidOpt, prog)
	}
	if st, err := src.Stat(ctx, "erp/"+guess+"/dzx"); err == nil && st.IsDir {
		return guess, nil
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	rel := fmt.Sprintf(FetchRoles[0].Src, guess, prog)
	if _, err := src.Stat(ctx, rel); err == nil {
		return guess, nil
	} else if !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	// 推断失败：给一条可操作的提示，而不是让调用方自己猜
	return "", fmt.Errorf("%w: 按程序名前三位推断的模块 %q 下既没有 dzx/ 也没有 %s；请用 --module <模块> 指定",
		ErrModuleUnknown, guess, rel)
}

// ErrModuleUnknown 推断不出模块（需要的程序极可能不在标准模块目录下）。
var ErrModuleUnknown = errors.New("无法确定程序所属模块")

// archiveDesign 把本地工作区里的设计稿复制进 design/。
//
// "设计稿"是用户本机工作区里的 `<prog>(c).tzs` / `<prog>(s).tzs` 这类文件（实测命名），
// 也可能是一个同名目录。**只归档、不解析**（ADR-05）：解析会把"解析缺陷"混进
// "材料缺陷"，model 排障时无法区分。
func archiveDesign(tmpDir, workspace, prog string) ([]FileEntry, []string, []string, error) {
	if strings.TrimSpace(workspace) == "" {
		return nil, []string{"design/"}, nil, nil
	}
	matches, err := filepath.Glob(filepath.Join(workspace, prog+"*.tzs"))
	if err != nil {
		return nil, nil, []string{"设计稿路径通配失败（" + err.Error() + "）"}, nil
	}
	sort.Strings(matches)
	var files []FileEntry
	var warnings []string
	for _, p := range matches {
		st, err := os.Stat(p)
		if err != nil {
			continue
		}
		base := filepath.Base(p)
		if st.IsDir() {
			warnings = append(warnings, "设计稿目录 "+base+" 未归档（本期只归档 .tzs 文件）")
			continue
		}
		if st.Size() > MaxFileBytes {
			warnings = append(warnings, fmt.Sprintf("%s 有 %s，超过单文件上限，已跳过", base, humanBytes(st.Size())))
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			warnings = append(warnings, "读设计稿 "+base+" 失败（"+err.Error()+"）")
			continue
		}
		rel := path.Join("design", base)
		if err := writeLocalFile(filepath.Join(tmpDir, filepath.FromSlash(rel)), data); err != nil {
			return nil, nil, nil, err
		}
		files = append(files, FileEntry{
			Role: "design", Src: "local:" + base, Rel: rel,
			SHA256: BytesSHA256(data), Size: int64(len(data)),
		})
	}
	if len(files) == 0 {
		return files, []string{"design/"}, warnings, nil
	}
	return files, nil, warnings, nil
}

// describeWorkspace 给警告文案用的工作区描述。
func describeWorkspace(ws string) string {
	if strings.TrimSpace(ws) == "" {
		return "该环境未配置工作区（hosts.sshs[].workspace）"
	}
	return "工作区 " + ws
}

// versionAnchors 取版本锚点。
//
// **版次默认在线查**（`DZAA_T` 的 `MAX(DZAA002)`），在线取不到才看本地共用资料 ——
// 镜像暂不启用（材料范围等 model 定型后再定），而"这个程序现在第几版"应以库为准。
// 模板版本只能来自本地材料（它是 `$TOP/erp/mdl` 里的文件内容），共用资料没同步就是空，
// 这种情况只记 warning，不影响材料本身。
func versionAnchors(ctx context.Context, cfgPath, env, prog string, ent int) (map[string]int, int) {
	tv := map[string]int{}
	envDir := config.DevDataEnvDirForEnv(cfgPath, env)
	if envDir == "" {
		return tv, 0
	}
	if m, merr := ReadManifest(envDir); merr == nil && m != nil {
		for k, v := range m.Versions.TemplateVersions {
			tv[k] = v
		}
	}
	if len(tv) == 0 {
		// 没有 manifest（或旧契约）：直接扫本地模板树，尽力给出锚点
		tv = scanTemplateVersions(filepath.Join(envDir, config.DevDataSubdirMdl))
	}
	return tv, programVersionAnchor(ctx, cfgPath, env, prog, ent)
}

// programVersionAnchor 版次锚点：**在线优先**，本地镜像只是兜底。
//
// 在线失败（环境没配库 / 连不上 / 账号不对）不判 fetch 失败：版本锚点缺失只是
// "判断不了有没有换版"，材料本身仍然是完整的 —— 所以降级 + 记 warning。
func programVersionAnchor(ctx context.Context, cfgPath, env, prog string, ent int) int {
	if e, err := envInfo(cfgPath, env); err == nil && e.DB != nil {
		if v, err := OnlineProgramVersion(ctx, e, ent, prog); err == nil && v > 0 {
			return v
		}
	}
	return localDZAAVersion(cfgPath, env, prog)
}

// scanTemplateVersions 扫一棵模板树，解析每个 `code_*.template` 的版本。
func scanTemplateVersions(mdlDir string) map[string]int {
	out := map[string]int{}
	if mdlDir == "" {
		return out
	}
	root, err := os.OpenRoot(mdlDir)
	if err != nil {
		return out
	}
	defer root.Close()
	_ = fs.WalkDir(root.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		key, ok := TemplateKey(p)
		if !ok {
			return nil
		}
		f, oerr := root.Open(p)
		if oerr != nil {
			return nil
		}
		head := make([]byte, 4096)
		n, _ := f.Read(head)
		f.Close()
		if v, ok := ParseTemplateVersion(head[:n]); ok {
			out[key] = v
		}
		return nil
	})
	return out
}

// localDZAAVersion 从本环境的规格库里读该程序的最大有效版次；读不到返回 0。
//
// 只查本地 SQLite：fetch 的来源是 SFTP，不该为了一个版本号再连一次数据库
// （那会让"拉文件"变成"要库密码"的动作）。
func localDZAAVersion(cfgPath, env, prog string) int {
	p := config.DevDataDictPathForEnv(cfgPath, env)
	if p == "" {
		return 0
	}
	if _, err := os.Stat(p); err != nil {
		return 0
	}
	conn, err := sql.Open("sqlite", p)
	if err != nil {
		return 0
	}
	defer conn.Close()
	var v sql.NullInt64
	err = conn.QueryRow(`SELECT MAX(CAST(dzaa002 AS INTEGER)) FROM dzaa_t WHERE dzaa001 = ? AND dzaastus = 'Y'`, prog).Scan(&v)
	if err != nil || !v.Valid {
		return 0
	}
	return int(v.Int64)
}

// humanBytes 人类可读的体积（CLI/Web 的"大小"列都用它，单位只有一份）。
func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGT"[exp])
}
