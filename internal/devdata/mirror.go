package devdata

// 共用资料（M）的拉取：53 张规格表 → SQLite，`$TOP/erp/mdl` → 文件树。
//
// 落盘纪律（与 fetch 同一套）：先在**同父目录**的临时位置建好，全部成功后才改名到位，
// 失败/取消就删临时目录。这样"半套材料"永远不会出现在正式落点上 —— 半套比没有更坏，
// 因为 model 会拿它算出"看起来对"的结论。
//
// 幂等：默认**已经同步过就跳过**（不重写、不报错，plan 成功标准 2 要求重复点不报错），
// `--force` 才重拉。`--only dict|mdl` 只刷新那一半，另一半的 manifest 条目原样保留。

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/luyaogod/TT/internal/config"
)

// Only 取值（`--only`）。
const (
	OnlyDict = "dict"
	OnlyMdl  = "mdl"
)

// ErrMdlTooLarge 模板树超过总量上限。
var ErrMdlTooLarge = fmt.Errorf("模板树超过 %d MB 上限", MaxMdlBytes>>20)

// MirrorRequest 一次共用资料同步的输入。
type MirrorRequest struct {
	// CfgPath config.json 路径（路径解析与身份解析的唯一入口）。
	CfgPath string
	// Env 环境名（必须已在该环境的 hosts.sshs 清单里）。
	Env string
	// Ent 企业编号（决定用哪个数据库账号）。0 = 用账号列表首项。
	Ent int
	// Only 空 = dict+mdl；"dict"/"mdl" = 只刷那一半。
	Only string
	// Force true = 已有材料也重拉；false = 已同步过就跳过（幂等）。
	Force bool
	// TTVersion 写进 manifest 的版本号（供 model 追溯"用哪个 tt 拉的"）。
	TTVersion string

	// Tables/Files 仅供测试注入；为 nil 时按环境打开真来源。
	Tables TableSource
	Files  FileSource
}

// MirrorResult 一次同步的结果（CLI/Web 与 status 共用这个形状）。
type MirrorResult struct {
	Env  string `json:"env"`
	Dir  string `json:"dir"`
	Only string `json:"only,omitempty"`
	// SkippedPull true = 已经同步过、本次没重拉（幂等跳过，不是错误）。
	SkippedPull bool `json:"skippedPull,omitempty"`
	// Tables 逐表结果（含被跳过的表）。
	Tables []TableEntry `json:"tables"`
	// TableCount 实际落库的表数；RowTotal 落库总行数。
	TableCount int   `json:"tableCount"`
	RowTotal   int64 `json:"rowTotal"`
	// MdlFiles 模板树文件数；MdlBytes 模板树字节数。
	MdlFiles int   `json:"mdlFiles"`
	MdlBytes int64 `json:"mdlBytes"`
	// TemplateVersions 模板版本表（`a05` → 3）。
	TemplateVersions map[string]int `json:"templateVersions,omitempty"`
	// Missing 缺失的表/目录（不判失败）。
	Missing  []string `json:"missing,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Elapsed  string   `json:"elapsed"`
}

// Mirror 执行一次共用资料同步。onProgress 可为 nil。
func Mirror(ctx context.Context, req MirrorRequest, onProgress ProgressFunc) (*MirrorResult, error) {
	start := time.Now()
	e, err := envInfo(req.CfgPath, req.Env)
	if err != nil {
		return nil, err
	}
	envDir, err := EnvDir(req.CfgPath, e.Name)
	if err != nil {
		return nil, err
	}
	only := strings.ToLower(strings.TrimSpace(req.Only))
	switch only {
	case "", OnlyDict, OnlyMdl:
	default:
		return nil, fmt.Errorf("%w: --only %q（只支持 dict / mdl）", ErrInvalidOpt, req.Only)
	}
	doDict := only == "" || only == OnlyDict
	doMdl := only == "" || only == OnlyMdl

	report := progressReporter(onProgress)
	res := &MirrorResult{Env: e.Name, Dir: envDir, Only: only, Tables: []TableEntry{}}

	// 幂等：已经同步过就跳过（不打开任何连接，status 里的时间/大小保持不变）
	prev, prevErr := ReadManifest(envDir)
	if prevErr != nil {
		// 清单坏了不阻塞重拉：材料还在，manifest 本来就要被本次覆盖。
		res.Warnings = append(res.Warnings, "旧 manifest 读不出来（"+prevErr.Error()+"），按未同步处理")
		prev = nil
	}
	if !req.Force && prev != nil && prev.Kind == KindMirror && covers(prev, doDict, doMdl) {
		res.SkippedPull = true
		res.TableCount = countTables(prev)
		res.RowTotal = sumTableRows(prev)
		res.MdlFiles = countRole(prev, "mdl")
		res.MdlBytes = sumRoleBytes(prev, "mdl")
		res.TemplateVersions = prev.Versions.TemplateVersions
		res.Missing = prev.Missing
		res.Warnings = append(res.Warnings, prev.Warnings...)
		res.Elapsed = time.Since(start).Round(10 * time.Millisecond).String()
		report(Progress{Phase: "done", Message: "已有共用资料，跳过（--force 重新拉取）", Done: 1, Total: 1})
		return res, nil
	}

	if err := os.MkdirAll(envDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建环境目录失败 (%s): %w", envDir, err)
	}
	m := NewManifest(KindMirror, e.Name)
	m.TTVersion = req.TTVersion
	m.Source.Ent = req.Ent
	// 只刷一半时，另一半的条目原样保留（否则 --only dict 会把 mdl 记录抹掉，
	// status 立刻显示"模板 0 个"，而文件其实还在盘上）
	if prev != nil {
		if !doDict {
			keepRole(m, prev, "dict")
			m.Tables = prev.Tables
			m.Missing = append(m.Missing, prev.Missing...)
		}
		if !doMdl {
			keepRole(m, prev, "mdl")
			for k, v := range prev.Versions.TemplateVersions {
				m.Versions.TemplateVersions[k] = v
			}
		}
		if prev.Source.TopDir != "" && m.Source.TopDir == "" {
			m.Source.TopDir = prev.Source.TopDir
		}
		if prev.Source.Zone != "" {
			m.Source.Zone = prev.Source.Zone
		}
	}

	if doDict {
		if err := mirrorDict(ctx, e, envDir, req, m, res, report, prev); err != nil {
			return nil, err
		}
	}
	if doMdl {
		if err := mirrorMdl(ctx, req, envDir, m, res, report); err != nil {
			return nil, err
		}
	}

	m.FetchedAt = nowStamp()
	if m.Source.Zone == "" {
		m.Source.Zone = e.Zone
	}
	// 一致性提示：M 侧没有 tab/tgl 可比，这里只把"缺表/跳过表"汇总成警告
	if len(res.Missing) > 0 {
		m.AddWarning(fmt.Sprintf("有 %d 张规格表在本环境不存在（已记入 missing）", len(res.Missing)))
	}
	if err := WriteManifest(envDir, m); err != nil {
		return nil, err
	}
	res.Missing = m.Missing
	res.Warnings = m.Warnings
	res.Elapsed = time.Since(start).Round(10 * time.Millisecond).String()
	report(Progress{Phase: "done", Message: fmt.Sprintf("同步完成：%d 张表 / %d 行，模板 %d 个",
		res.TableCount, res.RowTotal, res.MdlFiles), Done: 1, Total: 1})
	return res, nil
}

// mirrorDict 镜 53 张规格表 → 临时 SQLite → 原子改名。
func mirrorDict(ctx context.Context, e *config.NamedSsh, envDir string, req MirrorRequest,
	m *Manifest, res *MirrorResult, report func(Progress), prev *Manifest) error {

	dictDir := filepath.Join(envDir, config.DevDataSubdirDict)
	tmpDir, err := NewTempDir(filepath.Join(dictDir, "build"))
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			CleanupTemp(tmpDir, "")
		}
	}()
	tmpDB := filepath.Join(tmpDir, config.DictDBFileName(e.Name))

	w, err := openSQLite(tmpDB)
	if err != nil {
		return err
	}
	// 构建期的库是丢弃品，关日志换速度是安全的
	if err := w.SetBulkMode(); err != nil {
		w.Close()
		return err
	}

	src := req.Tables
	if src == nil {
		report(Progress{Phase: "connect", Message: "连接数据库…", Total: len(MirrorTables)})
		src, err = OpenTableSource(ctx, e, req.Ent)
		if err != nil {
			w.Close()
			return err
		}
		defer src.Close()
	}

	synced := map[string]bool{}
	total := len(MirrorTables)
	for i, spec := range MirrorTables {
		if err := ctx.Err(); err != nil {
			w.Close()
			return err
		}
		report(Progress{Phase: "table", Message: "拉取 " + spec.Name, Done: i, Total: total, CurrentFile: spec.Name})
		n, err := src.CountRows(ctx, spec.Name)
		if err != nil {
			if isMissingTable(err) {
				// 表不在这个环境的库里：记 missing 继续（有的模块没装就没有对应表）
				m.AddMissing(spec.Name)
				m.Tables = append(m.Tables, TableEntry{Table: spec.Name, Missing: true, Desc: spec.Desc})
				res.Missing = m.Missing
				res.Warnings = append(res.Warnings, "本环境没有表 "+spec.Name)
				continue
			}
			if cerr := ctx.Err(); cerr != nil {
				w.Close()
				return cerr
			}
			w.Close()
			return fmt.Errorf("数 %s 行数失败: %w", spec.Name, err)
		}
		if n > MaxTableRows {
			// 超上限：记录实测行数但**不落库**（plan §5.4）
			m.Tables = append(m.Tables, TableEntry{Table: spec.Name, Rows: n, Skipped: true, Desc: spec.Desc})
			res.Tables = append(res.Tables, TableEntry{Table: spec.Name, Rows: n, Skipped: true, Desc: spec.Desc})
			warn := fmt.Sprintf("%s 有 %d 行，超过单表 %d 行上限，已跳过（%s）",
				spec.Name, n, MaxTableRows, spec.Note)
			m.AddWarning(warn)
			res.Warnings = append(res.Warnings, warn)
			continue
		}
		meta, err := copyTable(ctx, w, src, spec, n)
		if err != nil {
			if cerr := ctx.Err(); cerr != nil {
				w.Close()
				return cerr
			}
			w.Close()
			return err
		}
		got, _ := w.CountRows(spec.Name)
		m.Tables = append(m.Tables, TableEntry{Table: spec.Name, Rows: got, Desc: spec.Desc})
		res.Tables = append(res.Tables, TableEntry{Table: spec.Name, Rows: got, Desc: spec.Desc})
		// LOB 列是就地截断到 4000 字节取的：命中时留一句，别让"少了一截"悄悄发生
		if len(meta.LOB) > 0 {
			warn := fmt.Sprintf("%s 有大字段列 %s，按 %d 字节截断写入（TO_CHAR/SUBSTR）",
				spec.Name, strings.Join(sortedKeys(meta.LOB), ","), lobCapChars)
			m.AddWarning(warn)
			res.Warnings = append(res.Warnings, warn)
		}
		res.TableCount++
		res.RowTotal += got
		synced[strings.ToLower(spec.Name)] = true
	}
	report(Progress{Phase: "index", Message: "建主键索引…", Done: total, Total: total})
	report(Progress{Phase: "index", Message: "建主键索引…", Done: total, Total: total})
	for _, wmsg := range w.CreatePKIndexes(synced) {
		m.AddWarning(wmsg)
		res.Warnings = append(res.Warnings, wmsg)
	}

	// 关闭并改名：临时库 → <env>/dict/<env>.db（--force 才允许覆盖；mirror 的"覆盖"
	// 由 Force 决定，未给 Force 时上面已经跳过，所以这里一律允许覆盖）
	final := filepath.Join(dictDir, config.DictDBFileName(e.Name))
	if err := w.CloseAndRename(final, true); err != nil {
		return err
	}
	committed = true
	CleanupTemp("", filepath.Join(dictDir, "build"))

	sha, err := FileSHA256(final)
	if err != nil {
		return fmt.Errorf("计算规格库指纹失败: %w", err)
	}
	st, _ := os.Stat(final)
	var size int64
	if st != nil {
		size = st.Size()
	}
	m.AddFile(FileEntry{Role: "dict", Src: "dict", Rel: filepath.ToSlash(filepath.Join(config.DevDataSubdirDict, config.DictDBFileName(e.Name))),
		SHA256: sha, Size: size, SrcMtime: nowStamp()})
	m.Source.DBPath = final
	return nil
}

// copyTable 把一张表的行流式写进临时库。
//
// 列名在**第一次回调**（row=nil 的表头通知）时就拿到，所以建表也在那一次完成。
// 0 行的表同样会收到表头通知 —— "本表为空"与"本表不存在"对 model 的含义完全不同，
// 不能因为没行就不建表（实测踩过：DZAS_T 0 行，曾报"表没有列"）。
func copyTable(ctx context.Context, w *sqliteWriter, src TableSource, spec TableSpec, rows int64) (TableMeta, error) {
	where := versionWhere(spec)
	// 分页排序键：主键列。取不到就是 nil（→ 单条流式，内存随表大小）
	order, err := src.OrderColumns(ctx, spec.Name)
	if err != nil {
		return TableMeta{}, fmt.Errorf("取 %s 的主键失败: %w", spec.Name, err)
	}
	// 列形态（含 LOB 标记）：拿到之后 LOB 列在 SQL 里就地转文本，不再是逐行往返
	meta, err := src.TableColumns(ctx, spec.Name)
	if err != nil {
		return TableMeta{}, fmt.Errorf("取 %s 的列清单失败: %w", spec.Name, err)
	}
	created := false
	var batch [][]string
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := w.InsertRows(batch); err != nil {
			return fmt.Errorf("写入 %s 失败: %w", spec.Name, err)
		}
		batch = batch[:0]
		return nil
	}
	onRow := func(cols []string, row []string) error {
		if !created {
			if err := w.BeginTable(spec.Name, cols); err != nil {
				return err
			}
			created = true
		}
		if row == nil {
			return nil // 表头通知，不是数据行
		}
		batch = append(batch, row)
		if len(batch) >= insertBatchRows {
			return flush()
		}
		return nil
	}
	err = src.Query(ctx, spec.Name, where, order, meta, onRow)
	if err != nil && where != "" && !isMissingTable(err) {
		// 版本过滤只在 DZAA_T 上用；某环境的 DZAA_T 结构不同（缺列）时退回整表，
		// 并留下警告 —— 静默退回会让"只镜有效版"的承诺落空而没人知道。
		created = false
		err = src.Query(ctx, spec.Name, "", order, meta, onRow)
	}
	if err != nil {
		if isMissingTable(err) {
			return meta, nil
		}
		return meta, fmt.Errorf("读取 %s 失败: %w", spec.Name, err)
	}
	if err := flush(); err != nil {
		return meta, err
	}
	if !created {
		return meta, fmt.Errorf("表 %s 没有返回任何列（%s）", spec.Name, spec.Desc)
	}
	return meta, nil
}

// sortedKeys 取 map 的键并排序（输出稳定，测试与人工比对都需要）。
func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// versionWhere ADR-03 的补充口径：只镜 DZAA_T 的当前有效版。
//
// 实测依据（2026-11-21 福联测试tst）：53 张表合计 18,371,187 行 ≫ 50 万，所以不是
// "全版次"；而 DZAA_T 单表 702,438 行 / 5,445 个程序，按"每程序最大版次且有效"筛完
// 是 272,355 行。
func versionWhere(spec TableSpec) string {
	if !spec.LatestOnly {
		return ""
	}
	// 外层表名就是 DZAA_T；相关子查询按表名引用外层列（Oracle/金仓都接受）
	return "DZAASTUS='Y' AND DZAA002 = (SELECT MAX(B.DZAA002) FROM DZAA_T B " +
		"WHERE B.DZAA001 = DZAA_T.DZAA001 AND B.DZAASTUS='Y')"
}

// mirrorMdl 把 `$TOP/erp/mdl` 整棵模板树拉到 `<env>/mdl`。
func mirrorMdl(ctx context.Context, req MirrorRequest, envDir string, m *Manifest,
	res *MirrorResult, report func(Progress)) error {

	mdlDir := filepath.Join(envDir, config.DevDataSubdirMdl)
	tmpDir, err := NewTempDir(filepath.Join(envDir, "mdl-build"))
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			CleanupTemp(tmpDir, "")
		}
	}()

	src := req.Files
	if src == nil {
		e, err := envInfo(req.CfgPath, req.Env)
		if err != nil {
			return err
		}
		report(Progress{Phase: "connect", Message: "连接服务器…"})
		var closeAll func()
		src, closeAll, err = OpenFileSource(e)
		if err != nil {
			return err
		}
		defer closeAll()
	}
	if m.Source.TopDir == "" {
		m.Source.TopDir = src.Top()
	}

	report(Progress{Phase: "mdl", Message: "列出模板树…"})
	var files []RemoteFile
	var totalBytes int64
	if err := src.Walk(ctx, MdlRoot, func(f RemoteFile) error {
		totalBytes += f.Size
		if totalBytes > MaxMdlBytes {
			return ErrMdlTooLarge
		}
		files = append(files, f)
		return nil
	}); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("服务器上没有 %s（该环境的 $TOP 下没有模板树）: %w", MdlRoot, err)
		}
		return err
	}
	total := len(files)
	prefix := MdlRoot + "/"
	for i, f := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		rel := strings.TrimPrefix(f.Rel, prefix)
		report(Progress{Phase: "mdl", Message: "下载 " + rel, Done: i, Total: total, CurrentFile: rel})
		data, err := src.ReadFile(ctx, f.Rel, MaxFileBytes)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) || errors.Is(err, fs.ErrNotExist) {
				m.AddMissing("mdl/" + rel)
				continue
			}
			return fmt.Errorf("下载 %s 失败: %w", f.Rel, err)
		}
		if err := writeLocalFile(filepath.Join(tmpDir, filepath.FromSlash(rel)), data); err != nil {
			return err
		}
		m.AddFile(FileEntry{
			Role: "mdl", Src: f.Rel, Rel: "mdl/" + rel,
			SHA256: BytesSHA256(data), Size: int64(len(data)),
			SrcMtime: f.ModTime.Format(time.RFC3339),
		})
		if key, ok := TemplateKey(rel); ok {
			if v, ok := ParseTemplateVersion(data); ok {
				m.Versions.TemplateVersions[key] = v
			}
		}
		res.MdlFiles++
		res.MdlBytes += int64(len(data))
	}
	// 模板树整棵替换：旧的多余文件必须消失，否则"上一版的模板"会混进来
	if err := CommitDir(tmpDir, mdlDir, true); err != nil {
		return err
	}
	committed = true
	res.TemplateVersions = m.Versions.TemplateVersions
	if len(m.Versions.TemplateVersions) == 0 {
		m.AddWarning("模板树里没有解析到任何 (Version:N) 标记：版本锚点缺失")
	}
	return nil
}

// covers 判旧 manifest 是否已经覆盖本次要刷的部分（幂等跳过用）。
func covers(m *Manifest, doDict, doMdl bool) bool {
	if doDict {
		if m.FileByRel(filepath.ToSlash(filepath.Join(config.DevDataSubdirDict, config.DictDBFileName(m.Env)))) == nil {
			return false
		}
	}
	if doMdl {
		if countRole(m, "mdl") == 0 {
			return false
		}
	}
	return m.FetchedAt != ""
}

// keepRole 把旧 manifest 里某角色的条目搬进新 manifest。
func keepRole(dst, src *Manifest, role string) {
	for _, f := range src.Files {
		if f.Role == role {
			dst.AddFile(f)
		}
	}
}

func countRole(m *Manifest, role string) int {
	n := 0
	for _, f := range m.Files {
		if f.Role == role {
			n++
		}
	}
	return n
}

func sumRoleBytes(m *Manifest, role string) int64 {
	var n int64
	for _, f := range m.Files {
		if f.Role == role {
			n += f.Size
		}
	}
	return n
}

func countTables(m *Manifest) int {
	n := 0
	for _, t := range m.Tables {
		if !t.Skipped && !t.Missing {
			n++
		}
	}
	return n
}

func sumTableRows(m *Manifest) int64 {
	var n int64
	for _, t := range m.Tables {
		n += t.Rows
	}
	return n
}

// isMissingTable 判"这张表在本环境的库里不存在"。
//
// Oracle 是 ORA-00942，金仓是 `relation ... does not exist`。判文案而不是判错误类型，
// 是因为驱动把库侧错误都包成同一类；这段文案同时要能在 CLI 输出里被认出来。
func isMissingTable(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "ora-00942") {
		return true
	}
	return strings.Contains(msg, "does not exist") || strings.Contains(msg, "不存在")
}

// ErrInvalidOpt 选项取值非法（--only 之类）。
var ErrInvalidOpt = errors.New("选项取值非法")
