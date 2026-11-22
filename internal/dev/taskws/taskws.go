package taskws

// 任务工作区：一个程序、一个任务、一个自包含目录。
//
// 为什么要有它（而不是继续用 devdata/<环境>/prog/<程序>）：材料目录回答的是"这个程序
// 的材料在哪"，而一次开发要的东西是"这个**任务**的全部输入与输出" —— 材料、索引、
// 可编辑视图、要交出去的包、变更记录。把它们放在一棵树里，才谈得上"断网可干活、
// 可回放、可交接"。
//
// 目录（全部相对工作区根）：
//
//	task.json          身份与历史（不存实时状态，实时状态一律现算）
//	materials/         材料（fetch 的落点，含 manifest.json 指纹）
//	mdl/slice/         模板切片（可选，切片脱版检测用；与 materials 分开放，
//	                   因为 fetch --force 会整目录替换 materials）
//	index/link.json    蓝图 ↔ 代码 索引
//	ws/                tzc 围栏工作区（唯一可编辑文件 ws/prog.full.4gl）
//	pkg/<程序>(<身份>).tzc  要交给设计器的产物
//	notes/             人的备注（工具不碰）
//
// 三条纪律（都是踩过的坑）：
//  1. task.json 只存身份/历史，**不存**"是否脏""包 sha"这类会变的量 ——
//     存了就会与 ws/manifest.json、pkg 实际内容打架，而打架时错的是"看起来新"的那个。
//  2. 有未落盘改动（ws 与基线不同）时拒绝 update：材料换了而视图没换，行号会静默错。
//  3. 材料刷新是整目录原子替换（fetch 的纪律），模板树放在 materials 外面才不会被连带清掉。

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/internal/dev/pkgfile"
	"github.com/luyaogod/TT/internal/dev/store"
	"github.com/luyaogod/TT/internal/devdata"
)

// TaskSchema 是 task.json 的契约版本（改结构要升它）。
const TaskSchema = "tt.dev.ws/v1"

// 目录/文件名（只有这里能拼工作区路径）。
const (
	TaskFileName = "task.json"

	DirMaterials = "materials"
	DirMdl       = "mdl"
	DirIndex     = "index"
	DirWs        = "ws"
	DirPkg       = "pkg"
	DirNotes     = "notes"

	linkFileName = "link.json"
)

// 领域错误（CLI 据此映射退出码）。
var (
	// ErrNoWorkspace 目录里没有 task.json。
	ErrNoWorkspace = errors.New("不是工作区")
	// ErrWorkspaceExists 目标目录已存在且不是空目录。
	ErrWorkspaceExists = errors.New("工作区已存在")
	// ErrDirtyWorkspace 有未落盘的改动，拒绝刷新材料。
	ErrDirtyWorkspace = errors.New("工作区有未落盘的改动")
	// ErrTaskName 任务名非法（会被拼进目录名）。
	ErrTaskName = errors.New("任务名非法")
	// ErrWorkspaceVersion task.json 的 schema 不是本版认的（多半是新版 tt 建的）。
	// 这是用户可处理的问题（换新版 tt 或重新 open），不是 IO 失败。
	ErrWorkspaceVersion = errors.New("工作区版本不兼容")
)

// Layout 记录各子目录名（写进 task.json，便于将来迁移目录结构）。
type Layout struct {
	Materials string `json:"materials"`
	Mdl       string `json:"mdl"`
	Index     string `json:"index"`
	Ws        string `json:"ws"`
	Pkg       string `json:"pkg"`
	Notes     string `json:"notes"`
}

// DefaultLayout 是当前版本的目录布局。
func DefaultLayout() Layout {
	return Layout{Materials: DirMaterials, Mdl: DirMdl, Index: DirIndex, Ws: DirWs, Pkg: DirPkg, Notes: DirNotes}
}

// Event 是工作区大事记（open/update/…）。只留最近若干条，避免变成日志。
type Event struct {
	At     string `json:"at"`
	Action string `json:"action"`
	Detail string `json:"detail,omitempty"`
}

// Task 是 task.json 的内容：身份 + 历史，**不含实时状态**。
type Task struct {
	Schema   string `json:"schema"`
	Prog     string `json:"prog"`
	Env      string `json:"env"`
	Module   string `json:"module,omitempty"`
	Identity string `json:"identity,omitempty"`
	Task     string `json:"task,omitempty"`
	// Ent 企业编号（--at <环境>@<企业> 才记；交接时"这份料是哪个企业的"必须能答）。
	Ent       int     `json:"ent,omitempty"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	TTVersion string  `json:"ttVersion,omitempty"`
	Layout    Layout  `json:"layout"`
	Events    []Event `json:"events,omitempty"`
}

// Paths 是工作区里各落点的计算入口（**唯一**拼路径的地方）。
type Paths struct{ Root string }

// Root 的绝对路径。
func (p Paths) Abs() string { return p.Root }

func (p Paths) Materials() string { return filepath.Join(p.Root, DirMaterials) }
func (p Paths) Mdl() string       { return filepath.Join(p.Root, DirMdl) }
func (p Paths) Index() string     { return filepath.Join(p.Root, DirIndex, linkFileName) }
func (p Paths) Ws() string        { return filepath.Join(p.Root, DirWs) }
func (p Paths) Notes() string     { return filepath.Join(p.Root, DirNotes) }
func (p Paths) TaskFile() string  { return filepath.Join(p.Root, TaskFileName) }

// Pkg 目标包路径：`pkg/<程序>(<身份>).tzc`。
func (p Paths) Pkg(prog, identity string) string {
	return filepath.Join(p.Root, DirPkg, devdata.PackFileName(prog, identity))
}

// PkgDir 包目录（探测已有包时用）。
func (p Paths) PkgDir() string { return filepath.Join(p.Root, DirPkg) }

// ---------- 读写 task.json ----------

// Load 读工作区的 task.json；目录不对时给出能照做的提示。
func Load(dir string) (*Task, Paths, error) {
	root, err := filepath.Abs(strings.TrimSpace(dir))
	if err != nil {
		return nil, Paths{}, fmt.Errorf("工作区路径不可用 (%s): %w", dir, err)
	}
	p := Paths{Root: root}
	b, err := os.ReadFile(p.TaskFile())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// 路径本身不存在 / 存在但不是工作区：这两件事对用户的下一步完全不同
			// （一个是"路径打错了"，一个是"这目录不是工作区"），文案必须分开。
			if _, serr := os.Stat(root); serr != nil {
				return nil, p, fmt.Errorf("%w: 路径不存在：%s", ErrNoWorkspace, root)
			}
			return nil, p, fmt.Errorf("%w: %s 里没有 %s（用 tt dev open 建一个）",
				ErrNoWorkspace, root, TaskFileName)
		}
		return nil, p, fmt.Errorf("读 %s 失败: %w", p.TaskFile(), err)
	}
	var t Task
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, p, fmt.Errorf("%s 不是合法 JSON: %w", p.TaskFile(), err)
	}
	if t.Schema != TaskSchema {
		return nil, p, fmt.Errorf("%w: %s 的 schema 是 %q，本版只认 %q（用新版 tt，或重新 open）",
			ErrWorkspaceVersion, p.TaskFile(), t.Schema, TaskSchema)
	}
	return &t, p, nil
}

// Save 写 task.json（原子替换）。
func Save(t *Task, p Paths) error {
	t.UpdatedAt = nowStamp()
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 task.json 失败: %w", err)
	}
	b = append(b, '\n')
	if err := store.AtomicWrite(p.TaskFile(), b); err != nil {
		return fmt.Errorf("写 %s 失败: %w", p.TaskFile(), err)
	}
	return nil
}

// AddEvent 记一条大事（最多留 20 条）。
func (t *Task) AddEvent(action, detail string) {
	t.Events = append(t.Events, Event{At: nowStamp(), Action: action, Detail: detail})
	if len(t.Events) > 20 {
		t.Events = t.Events[len(t.Events)-20:]
	}
}

// ---------- open ----------

// CreateRequest 是一次 open 的输入。
type CreateRequest struct {
	CfgPath string
	Env     string
	Prog    string
	// Module 模块目录（空 = 从前三位推断，客制优先）。
	Module string
	// Name 任务名（拼进默认目录名 `<程序>-<任务名>`，并记进 task.json）。
	Name string
	// Dir 目标目录（空 = 当前目录下的 `<程序>[-<任务名>]`）。
	Dir string
	// Identity 包身份 c/s（空 = 从 .tap 推断）。
	Identity string
	// Only 传给 tzc export 的部分导出。
	Only []string
	// NoMdl 不复制模板切片（切片脱版检测会退化成"没镜像"）。
	NoMdl bool
	// Form 连 .42f 一起拉。
	Form bool
	Ent  int
	// TTVersion 记进 task.json 与材料 manifest。
	TTVersion string
	// Files 仅供测试注入（离线跑通"拉料→建索引→打包→导出"整条 open 链）；
	// nil 时按环境打开真 SFTP（与 devdata.FetchRequest.Files 同一条缝）。
	Files devdata.FileSource
	// Progress 进度回调（可为 nil）。
	Progress devdata.ProgressFunc
}

// CreateResult 是 open 的结果。
type CreateResult struct {
	Dir      string
	Task     *Task
	Fetch    *devdata.FetchResult
	Pack     *devdata.PackResult
	Export   *ExportResult
	Index    *LinkSummary
	Warnings []string
}

// LinkSummary 是索引的摘要（不把 800 KB 的 doc 带出去）。
type LinkSummary struct {
	Path        string
	Schema      string
	GeneratedAt string
	Points      int
	Custom      int
	Orphans     int
	Sections    int
	Slices      int
	Drift       []devdata.SliceDrift
}

// SummarizeLink 从索引 doc 取摘要。
func SummarizeLink(doc *devdata.LinkDoc) *LinkSummary {
	if doc == nil {
		return nil
	}
	return &LinkSummary{
		Path: doc.Path, Schema: doc.Schema, GeneratedAt: doc.GeneratedAt,
		Points: doc.Stats.Points, Custom: doc.Stats.PointsCustom,
		Orphans: doc.Stats.DeclOrphans, Sections: len(doc.Sections),
		Slices: doc.Slices.Distinct, Drift: doc.Slices.Drift,
	}
}

// Create 建一个任务工作区：拉材料 → 建索引 → 打包 → 导出视图 → 写 task.json。
//
// 失败不留半套：中途出错就把刚建的目录删掉（只删**本次创建**的那个，不删用户已有的）。
func Create(ctx context.Context, req CreateRequest) (*CreateResult, error) {
	if err := devdata.ValidProgName(req.Prog); err != nil {
		return nil, err
	}
	prog := strings.TrimSpace(req.Prog)
	// 先按原值校验、再修剪：否则 `--name " epi "` 会被静默改成 `epi`，
	// validTaskName 里"首尾不要空白"那条就永远不可达（契约与实现打架）。
	if err := validTaskName(req.Name); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	root, err := resolveCreateDir(req.Dir, prog, name)
	if err != nil {
		return nil, err
	}
	p := Paths{Root: root}

	createdHere := false
	if st, serr := os.Stat(root); serr == nil {
		if !st.IsDir() {
			return nil, fmt.Errorf("%w: %s 是个文件", ErrWorkspaceExists, root)
		}
		entries, _ := os.ReadDir(root)
		if len(entries) > 0 {
			if _, terr := os.Stat(p.TaskFile()); terr == nil {
				return nil, fmt.Errorf("%w: %s（tt dev status 看它，或换个 --dir）", ErrWorkspaceExists, root)
			}
			return nil, fmt.Errorf("%w: %s 不是空目录（换个 --dir）", ErrWorkspaceExists, root)
		}
	} else if os.IsNotExist(serr) {
		if err := os.MkdirAll(root, 0o755); err != nil {
			return nil, fmt.Errorf("创建目录失败 (%s): %w", root, err)
		}
		createdHere = true
	} else {
		return nil, fmt.Errorf("检查目标目录失败 (%s): %w", root, serr)
	}
	// 任何中途失败都清掉本次建的目录 —— 半套工作区比没有更坏。
	ok := false
	defer func() {
		if !ok && createdHere {
			_ = os.RemoveAll(root)
		}
	}()

	res := &CreateResult{Dir: root}

	// 1) 材料
	fetch, err := devdata.Fetch(ctx, devdata.FetchRequest{
		CfgPath: req.CfgPath, Env: req.Env, Prog: prog, Module: req.Module, Ent: req.Ent,
		Form: req.Form, Force: true, TTVersion: req.TTVersion, Dir: p.Materials(),
		Files: req.Files,
	}, req.Progress)
	if err != nil {
		return nil, err
	}
	res.Fetch = fetch
	res.Warnings = append(res.Warnings, fetch.Warnings...)
	// 没有 .tap/.tgl 的程序（没被设计器接管过的标准程序）：建不了工作区 —— 视图和包
	// 都是从 .tzc 来的，而服务器上根本没有这个包。早期拦下来，比在打包那一步报
	// "材料不成套"更容易读懂。
	if !hasRole(fetch.Files, "tap") || !hasRole(fetch.Files, "tgl") {
		return nil, fmt.Errorf("%w: %s 没有被设计器接管（服务器上没有 .tap/.tgl），建不了工作区；\n"+
			"  只想读它的代码：tt gen fetch %s && tt gen map %s",
			devdata.ErrMaterialIncomplete, prog, prog, prog)
	}

	// 2) 模板切片（可选，失败不阻断：只是少一条脱版检查）
	mdlDir, mwarn := copyMdl(req.CfgPath, req.Env, p.Mdl(), req.NoMdl)
	res.Warnings = append(res.Warnings, mwarn...)

	// 3) 索引
	doc, err := devdata.Map(devdata.LinkRequest{
		CfgPath: req.CfgPath, Env: req.Env, Prog: prog,
		Dir: p.Materials(), Out: p.Index(), Mdl: mdlDir,
	})
	if err != nil {
		return nil, err
	}
	res.Index = SummarizeLink(doc)

	// 4) 包（与设计器下载同构：材料原样，不含本地改动）
	// 先把 pkg/ 建出来：devdata.Pack 把 --out 当目录用要在 Stat 时它已经是目录，
	// 否则会把目录路径本身当成包文件名（结果是写出一个没有扩展名的 zip）。
	if err := os.MkdirAll(p.PkgDir(), 0o755); err != nil {
		return nil, fmt.Errorf("建包目录失败 (%s): %w", p.PkgDir(), err)
	}
	pack, err := devdata.Pack(devdata.PackRequest{
		CfgPath: req.CfgPath, Env: req.Env, Prog: prog, Identity: req.Identity,
		Out: p.PkgDir(), Force: true, Dir: p.Materials(),
	})
	if err != nil {
		return nil, err
	}
	res.Pack = pack
	res.Warnings = append(res.Warnings, pack.Warnings...)
	// Pack 的 Out 是目录时按身份定名；这里把它记成实际落点，后面 export 用同一份。
	pkgPath := pack.Path

	// 5) 视图
	exp, err := Export(ExportRequest{Dir: p.Ws(), Pkg: pkgPath, Only: req.Only})
	if err != nil {
		return nil, err
	}
	res.Export = exp
	res.Warnings = append(res.Warnings, ExportWarnings(exp)...)

	// 6) task.json
	task := &Task{
		Schema: TaskSchema, Prog: prog, Env: fetch.Env, Module: fetch.Module,
		Identity: pack.Identity, Task: name, Ent: req.Ent, CreatedAt: nowStamp(),
		TTVersion: req.TTVersion, Layout: DefaultLayout(),
	}
	task.AddEvent("open", fmt.Sprintf("材料 %d 个文件 / 索引 %d 点 / 包 %s",
		len(fetch.Files), doc.Stats.Points, filepath.Base(pkgPath)))
	if err := Save(task, p); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(p.Notes(), 0o755); err != nil {
		res.Warnings = append(res.Warnings, "建 notes/ 失败："+err.Error())
	}
	res.Task = task
	ok = true
	return res, nil
}

// ErrPackageMismatch 工作区里的包与视图基线记的 sha 对不上（被换过/复制来的）。
var ErrPackageMismatch = errors.New("包与视图基线不一致")

// RebindPackage 把视图清单里的"源包路径"指到**当前工作区**的包，并返回该路径。
//
// 为什么必须做：`ws/manifest.json` 里 pkg.path 是绝对路径，apply 默认照它写回。
// 工作区一旦被复制/移动/发给同事，apply 就会静默写进**原目录**的包（实测踩过，
// 而且 status 还会说"可交付"）。所以工作区命令在 apply 之前先把路径绑回本地，
// 并且要求本地包与基线 sha 一致 —— 不一致就拒绝，别拿别人/旧的包当基线。
func RebindPackage(dir string) (string, bool, error) {
	t, p, err := Load(dir)
	if err != nil {
		return "", false, err
	}
	ws, err := store.Open(p.Ws())
	if err != nil {
		return "", false, err
	}
	mf, merr := ws.Manifest()
	if merr != nil {
		return "", false, merr
	}
	local := p.Pkg(t.Prog, identityOf(t))
	if mf.Pkg.Path == local {
		return local, false, nil
	}
	sha, _ := pkgSha(local)
	if sha == "" {
		return "", false, fmt.Errorf("%w: 当前工作区没有包 %s（用 tt dev update --full 重建，或把包放回去）",
			ErrPackageMismatch, local)
	}
	if mf.Pkg.Sha256 != "" && sha != mf.Pkg.Sha256 {
		return "", false, fmt.Errorf("%w: 本地包 %s 的 sha256 是 %s，视图基线记的是 %s。\n"+
			"  这通常是「工作区被复制/移动」或「包被替换过」；\n"+
			"  出路：tt dev relink（以当前包重新导出视图；未落盘的改动会丢），或把基线那份包放回来",
			ErrPackageMismatch, local, short(sha), short(mf.Pkg.Sha256))
	}
	if err := ws.RebindPackagePath(local); err != nil {
		return "", false, err
	}
	return local, true, nil
}

// Short 取 sha 前 12 位（提示文案用）。导出是因为 CLI 的拒绝文案要用同一口径。
func Short(sha string) string { return short(sha) }

// short 取 sha 前 12 位（提示文案用）。
func short(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

// hasRole 判断材料里有没有某个角色（tap/tgl/4gl…）。
func hasRole(files []devdata.FileEntry, role string) bool {
	for _, f := range files {
		if f.Role == role {
			return true
		}
	}
	return false
}

// resolveCreateDir 算出工作区根：显式 --dir，否则 `<当前目录>/<程序>[-<任务名>]`。
func resolveCreateDir(dir, prog, name string) (string, error) {
	if d := strings.TrimSpace(dir); d != "" {
		return filepath.Abs(d)
	}
	base := prog
	if name != "" {
		base = prog + "-" + name
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("取当前目录失败（用 --dir 指定）: %w", err)
	}
	return filepath.Join(wd, base), nil
}

// validTaskName 任务名会进目录名，所以不许路径分隔符/上跳/空白头尾。
func validTaskName(name string) error {
	if name == "" {
		return nil
	}
	if len(name) > 40 {
		return fmt.Errorf("%w: %q 太长（最多 40 字节）", ErrTaskName, name)
	}
	if strings.ContainsAny(name, `/\:*?"<>|`) || strings.Contains(name, "..") {
		return fmt.Errorf("%w: %q（不能含路径符号或 ..）", ErrTaskName, name)
	}
	if strings.TrimSpace(name) != name {
		return fmt.Errorf("%w: %q（首尾不要空白）", ErrTaskName, name)
	}
	return nil
}

// copyMdl 把环境的模板切片复制到工作区（`<mdl>/slice/**`）。
//
// 只复制 slice/：脱版检测只读 `<mdl>/slice/code_<代号>.template`，整棵模板树 3.2 MB
// 里其余部分（表单样板、UITemplateSetting）对"改哪些区块"没有直接作用，放进来只会
// 让工作区更难带走。环境里没镜过模板就跳过并给提示（不是错误）。
func copyMdl(cfgPath, env, dst string, noMdl bool) (string, []string) {
	if noMdl {
		return "", nil
	}
	srcRoot := config.DevDataMdlDirForEnv(cfgPath, env)
	if srcRoot == "" {
		return "", nil
	}
	src := filepath.Join(srcRoot, "slice")
	if st, err := os.Stat(src); err != nil || !st.IsDir() {
		return "", []string{"模板切片没镜像（tt gen mirror --only mdl 之后重跑 update 才有脱版检查）"}
	}
	tmp, err := devdata.NewTempDir(dst)
	if err != nil {
		return "", []string{"复制模板切片失败（跳过脱版检查）：" + err.Error()}
	}
	if err := copyTree(src, filepath.Join(tmp, "slice")); err != nil {
		devdata.CleanupTemp(tmp, "")
		return "", []string{"复制模板切片失败（跳过脱版检查）：" + err.Error()}
	}
	// CommitDir 只认 final 的父目录；这里 final=dst，tmp 是它的兄弟。
	if err := devdata.CommitDir(tmp, dst, true); err != nil {
		devdata.CleanupTemp(tmp, "")
		return "", []string{"替换模板切片失败（跳过脱版检查）：" + err.Error()}
	}
	return dst, nil
}

// copyTree 递归复制目录（保留文件权限位；不复制符号链接指向的内容）。
func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(src, path)
		if rerr != nil {
			return rerr
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		return store.AtomicWrite(target, b)
	})
}

// nowStamp 统一时间戳形态（RFC3339，带本地时区）。
func nowStamp() string { return time.Now().Format(time.RFC3339) }

// sortedKeys 给提示文案用（稳定输出）。
func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// pkgSha 读包算 sha256（读不到返回空）。
func pkgSha(path string) (string, int64) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", 0
	}
	return pkgfile.Sha256Hex(b), int64(len(b))
}

// sameFile 比较两份字节内容（用于脏检测）。
func sameFile(a, b []byte) bool { return bytes.Equal(a, b) }
