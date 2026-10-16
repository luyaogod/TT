package fep

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/debug"
	"tt/internal/fep"
)

// tt fep rec —— 录制件的身份与清单。
//
// 一条录制的**名字就是它的 ID**（代理落盘用的 tag，形如 fep-20261015-170031）。
// 这个 ID 是给人复制流转的：前端面板录完把它显示出来，用户复制给 agent，
// agent 拿它跑 `tt fep rec show <ID>` / `tt fep rec replay --id <ID>`。
// 所以 ID 必须能从文件后缀里切出来（见 fep.RecordingIDFromFile），不能是另一套编号。

var recID string

// recTarget 是"用户说的那份录制"解析之後的样子。
type recTarget struct {
	ID    string
	Dir   string
	Files []string // AUI 侧（rt2gdc）文件，按时间从早到晚
	Meta  *fep.RecordingMeta
	// Path / Prog 取自身份（没有身份文件时为空）—— 闸门的路径比对要用它们
	Path fep.LaunchPath
	Prog string
}

// recRoot 是一个录制落点。一份录制可能在两处：**缓存**（这一轮刚录、收工刚搬走之前的）
// 与**长期库**（按 环境+ENT 分层、不会被清缓存删掉）。
type recRoot struct {
	Label string
	Dir   string
}

// recordingRoots 默认要找的两个落点。
func recordingRoots() ([]recRoot, error) {
	cache, err := cacheDir()
	if err != nil {
		return nil, err
	}
	lib, err := library()
	if err != nil {
		return nil, err
	}
	return []recRoot{{Label: "缓存", Dir: cache}, {Label: "长期库", Dir: lib.Root}}, nil
}

// resolveRecording 把「录制 ID」或「--from 文件/目录」解析成一份确定的录制。
//
// 优先级是有意的：**--from 优先**——它说的是"就用这份文件/这个目录"；
// 只给 ID 时先看缓存、再看长期库（长期库按 环境/ENT 分层，要遍历）。
// 两个都不给就报用法错，绝不悄悄取最新一份。
func resolveRecording(from, id string) (recTarget, error) {
	var t recTarget
	switch {
	case from != "":
		st, err := os.Stat(from)
		if err != nil {
			return t, fail(exitFailed, "读不到 %s: %v", from, err)
		}
		if st.IsDir() {
			t.Dir = from
			t.ID = id
			if t.ID == "" {
				newest, err := newestRecordingID(from)
				if err != nil {
					return t, err
				}
				t.ID = newest
			}
		} else {
			t.Dir = filepath.Dir(from)
			// --from 可以是任意一侧：界面侧（rt2gdc）不在时，用它的孪生文件 ——
			// 拿事件文件去配别的连接的界面，编号会对不上（见 fep.go 的 pairedRecording）。
			aui := siblingAUI(from)
			if bid, ok := fep.RecordingIDFromFile(filepath.Base(aui)); ok {
				t.ID = bid
			} else {
				t.ID = strings.TrimSuffix(filepath.Base(aui), filepath.Ext(aui))
			}
			t.Files = []string{aui}
		}
	case id != "":
		found, err := findByID(id)
		if err != nil {
			return t, err
		}
		t = found
	default:
		return t, fail(exitUsage, "要么给录制 ID（tt fep rec show <ID>），要么给 --from <文件或目录>")
	}
	if len(t.Files) == 0 {
		files, err := fep.RecordingFilesFor(t.Dir, t.ID)
		if err != nil {
			return t, fail(exitFailed, "读 %s: %v", t.Dir, err)
		}
		if len(files) == 0 {
			return t, fail(exitTarget, "%s 里没有录制 %s（看 tt fep rec list）", t.Dir, t.ID)
		}
		t.Files = files
	}
	if t.Meta == nil {
		if mt, err := fep.ReadMeta(t.Dir, t.ID); err == nil {
			t.Meta = mt
		} else if !errors.Is(err, fep.ErrNoMeta) {
			return t, fail(exitFailed, "%v", err)
		}
	}
	if t.Meta != nil {
		t.Path, t.Prog = t.Meta.Path, t.Meta.Prog
	}
	return t, nil
}

// findByID 按录制 ID 在两个默认落点里找一份。
//
// 找不到时的报文要把**找过哪里**说出来：录制件可能刚被清缓存删掉、也可能还在
// 服务器上没传回来，只说"没有"会让人以为 ID 写错了。
func findByID(id string) (recTarget, error) {
	var t recTarget
	cache, err := cacheDir()
	if err != nil {
		return t, err
	}
	if files, ferr := fep.RecordingFilesFor(cache, id); ferr == nil && len(files) > 0 {
		return recTarget{ID: id, Dir: cache, Files: files}, nil
	}
	lib, err := library()
	if err != nil {
		return t, err
	}
	if item, found, lerr := lib.Find(id); lerr != nil {
		return t, fail(exitFailed, "%v", lerr)
	} else if found {
		return recTarget{ID: item.ID, Dir: item.Dir, Files: item.Files, Meta: item.Meta}, nil
	}
	return t, fail(exitTarget, "找不到录制 %s（找过缓存 %s 与长期库 %s）—— 看 tt fep rec list",
		id, cache, lib.Root)
}

// siblingAUI 把"同一连接的另一侧"换成界面侧：给了 `-gdc2rt.log` 就换成它的
// `-rt2gdc.log`（存在才换）。界面状态只在 rt2gdc 那侧。
func siblingAUI(file string) string {
	if !strings.HasSuffix(file, "-gdc2rt.log") {
		return file
	}
	p := strings.TrimSuffix(file, "-gdc2rt.log") + "-rt2gdc.log"
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return file
}

// events 把这份录制里"前端→运行期"那一侧读成事件串。
//
// 事件必须来自**同一份**录制：同一个连接的事件配另一个连接的界面，编号对不上，
// 动作会全变"名字未解析"（这正是 pairedRecording 存在的理由）。
func (t recTarget) events() ([]string, error) {
	var out []string
	for _, f := range t.Files {
		src := f
		if strings.HasSuffix(f, "-rt2gdc.log") {
			src = strings.TrimSuffix(f, "-rt2gdc.log") + "-gdc2rt.log"
			if _, err := os.Stat(src); err != nil {
				// 没有事件侧：这一轮只录到了界面（作业刚打开就断了）。照实说，别静默当成 0 条。
				return nil, fail(exitTarget, "%s 没有事件侧（%s）：这份录制里没有可回放的动作",
					filepath.Base(f), filepath.Base(src))
			}
		}
		bodies, err := readRecording(src)
		if err != nil {
			return nil, err
		}
		for _, body := range bodies {
			out = append(out, fep.SplitEvents(body)...)
		}
	}
	return out, nil
}

// newestRecordingID 取目录里最新的一份录制 ID（按 AUI 文件的修改时间）。
func newestRecordingID(dir string) (string, error) {
	files, err := fep.RecordingFiles(dir, "-rt2gdc.log")
	if err != nil {
		return "", fail(exitFailed, "读 %s: %v", dir, err)
	}
	for _, f := range files { // RecordingFiles 已按时间倒序
		if id, ok := fep.RecordingIDFromFile(filepath.Base(f)); ok {
			return id, nil
		}
	}
	return "", fail(exitTarget, "%s 里没有录制件（*%s）", dir, "-rt2gdc.log")
}

// multi 把一份录制摊成闸门要的那一面：**按连接各一张表**。
//
// 为什么不合成一张：一个会话本来就可能有多条连接、而且可以是完全不同的两棵树
// （实测 `r.d` 会话：连接 1 是 fgldeb 调试器的界面、连接 2 是作业界面）。
// 合成一张去比单连接的表，同一条路径也会判不过（实测 93%）。
func (t recTarget) multi() fep.MultiSnapshot {
	snap, err := fep.FilesMultiSnapshot(t.Files, t.ID)
	if err != nil {
		snap = fep.MultiSnapshot{Label: t.ID,
			Conns: []fep.ConnSnapshot{{Snapshot: fep.EmptySnapshot(t.ID, err.Error())}}}
	}
	snap.Path, snap.Prog, snap.Meta = t.Path, t.Prog, t.Meta
	return snap
}

// snapshot 把一份录制摊成**整会话一张表**（只给 `rec show` 报一个总指纹用；
// 判闸门请用 multi）。
//
// 读不出编号表时**不报错**，而是给一面"空表 + 为什么空"的快照：闸门要能说出
// "这份录制里没有界面"，那与"文件读不动"是两件事。
func (t recTarget) snapshot() fep.Snapshot {
	snap, err := fep.SnapshotOfFiles(t.Files, t.ID)
	if err != nil {
		snap = fep.EmptySnapshot(t.ID, err.Error())
	}
	snap.ID = t.ID
	snap.Path, snap.Prog = t.Path, t.Prog
	snap.Meta = t.Meta
	if snap.FP.Hash == "" && t.Meta != nil {
		snap.FP = t.Meta.Fingerprint // 身份里那份是收工时算的；录制件读不动时才用它
	}
	return snap
}

// liveMulti 取"当前会话"那一面（没有会话就返回 ok=false，不报错），**按连接摊表**。
//
// 路径从**登记**里读而不是从录制件里推：`--no-job` 起的代理自己没有启动历史，
// 而它后面挂的作业是哪条路径起的，只有起作业的那一方知道。
func liveMulti(client *fep.Client) (fep.MultiSnapshot, bool, error) {
	if client == nil {
		return fep.MultiSnapshot{}, false, nil
	}
	label := "当前会话 " + client.Tag
	snap, err := fep.SessionMultiSnapshot(client.RecordDir, client.Tag, label)
	if err != nil {
		// 会话在，但界面还没出来（作业还没连上来）—— 照样是一面，只是空表
		snap = fep.MultiSnapshot{Label: label,
			Conns: []fep.ConnSnapshot{{Snapshot: fep.EmptySnapshot(label, err.Error())}}}
	}
	reg, _, rerr := registry()
	if rerr != nil {
		return fep.MultiSnapshot{}, false, rerr
	}
	if s, lerr := reg.Load(); lerr == nil {
		snap.Path, snap.Prog = s.Path, s.Prog
	}
	return snap, true, nil
}

var recListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出录制件（ID / 启动路径 / 作业 / 环境·区域·ENT / 指纹），缓存与长期库都看",
	Long: `列出录制件：默认把**缓存**（这一轮刚录的）与**长期库**（<数据目录>/records/<环境>/<ENT>/）
两处都列出来，每行标出它住哪。--from 只列指到的那一处。

每份录制都会**当场复算指纹**（拿录制件的字节重建界面再摊成编号表），
所以没有身份文件（.meta.json）的老录制件也能列出指纹 —— 只是路径与作业名一栏空着。`,
	Args: noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var roots []recRoot
		if uiFrom != "" {
			roots = []recRoot{{Label: "指定", Dir: uiFrom}}
		} else {
			r, err := recordingRoots()
			if err != nil {
				return err
			}
			roots = r
		}
		var rows []recRow
		for _, root := range roots {
			// 长期库是 环境/ENT 两层 + 录制一层；缓存是平的。库走它自己的遍历。
			entries, err := listRoot(root)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return fail(exitFailed, "读 %s: %v", root.Dir, err)
			}
			rows = append(rows, entries...)
		}
		sort.SliceStable(rows, func(i, j int) bool { return rows[i].Started > rows[j].Started })
		dirs := make([]string, 0, len(roots))
		for _, r := range roots {
			dirs = append(dirs, r.Label+" "+r.Dir)
		}
		return printJSONOrText(map[string]any{"roots": dirs, "count": len(rows), "recordings": rows}, func() {
			if len(rows) == 0 {
				fmt.Printf("没有录制件（看过的落点：%s）\n", strings.Join(dirs, "；"))
				return
			}
			for _, d := range dirs {
				fmt.Printf("落点 %s\n", d)
			}
			for _, r := range rows {
				if r.Err != "" {
					fmt.Printf("  %s  %s  %s\n", pad(r.ID, 24), padDisp(r.Where, 8), "读不了："+r.Err)
					continue
				}
				path := string(r.Path)
				if !r.HasMeta {
					path = "（无身份）"
				}
				where := r.Env
				if r.Zone != "" {
					where += "/" + r.Zone
				}
				if r.Ent != "" {
					where += "/ENT " + r.Ent
				}
				fmt.Printf("  %s  %s  %s  %s  %s  %s\n", pad(r.ID, 24), padDisp(r.Where, 8),
					padDisp(path, 12), padDisp(r.Prog, 14), padDisp(where, 22), r.Fingerprint)
			}
		})
	},
}

// recRow 是清单里的一行。
type recRow struct {
	ID          string          `json:"id"`
	Where       string          `json:"where"` // 缓存 / 长期库 / 指定
	Dir         string          `json:"dir"`
	Path        fep.LaunchPath  `json:"path,omitempty"`
	Prog        string          `json:"prog,omitempty"`
	Env         string          `json:"env,omitempty"`
	Zone        string          `json:"zone,omitempty"`
	Ent         string          `json:"ent,omitempty"`
	Origin      fep.Origin      `json:"origin,omitempty"`
	Retention   fep.Retention   `json:"retention,omitempty"`
	Started     string          `json:"started,omitempty"`
	Conns       int             `json:"conns"`
	BytesRT     int64           `json:"bytesRT,omitempty"`
	Fingerprint fep.Fingerprint `json:"fingerprint"`
	HasMeta     bool            `json:"hasMeta"`
	Err         string          `json:"error,omitempty"`

	Meta *fep.RecordingMeta `json:"-"`
}

// describeRecording 描述一份录制：有身份就用身份里记的（收工时算的那份指纹），
// 没有就从字节**当场复算** —— 老录制件照样能列出来，只是路径与作业名一栏空着。
func describeRecording(where, dir, id string) recRow {
	r := recRow{ID: id, Where: where, Dir: dir}
	if mt, err := fep.ReadMeta(dir, id); err == nil {
		r.Meta = mt
		fillFromMeta(&r)
	}
	if r.Fingerprint.Hash == "" {
		files, ferr := fep.RecordingFilesFor(dir, id)
		if ferr != nil {
			r.Err = ferr.Error()
		} else if snap, serr := fep.SnapshotOfFiles(files, id); serr != nil {
			r.Err = serr.Error()
		} else {
			r.Fingerprint = snap.FP
		}
		if r.Conns == 0 {
			r.Conns = len(files)
		}
	}
	return r
}

func fillFromMeta(r *recRow) {
	mt := r.Meta
	if mt == nil {
		return
	}
	r.HasMeta = true
	r.Path, r.Prog, r.Env, r.Zone, r.Ent = mt.Path, mt.Prog, mt.Env, mt.Zone, mt.Ent
	r.Origin, r.Retention = mt.Origin, mt.Retention
	r.Started = mt.Started.Format(time.RFC3339)
	r.Conns = len(mt.Conns)
	for _, c := range mt.Conns {
		r.BytesRT += c.BytesRT
	}
	r.Fingerprint = mt.Fingerprint
}

var recKeepCmd = &cobra.Command{
	Use:   "keep <录制 ID>",
	Short: "把一份录制搬进长期库（按 环境+ENT 分层，不会被清缓存删掉）",
	Long: `把一份录制搬进长期库：<数据目录>/records/<环境>/<ENT>/<录制 ID>/。

收工时其实已经自动搬了（session open 默认搬，--no-keep 才留在缓存目录）。
这条命令是给"当时没搬、或从别处拿回来的一份"用的 —— 它要求这份录制有身份文件
（.meta.json）：进哪个环境、哪个 ENT 的库只能从身份里读，猜不得。`,
	Args: exactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := resolveRecording("", args[0])
		if err != nil {
			return err
		}
		lib, err := library()
		if err != nil {
			return err
		}
		if rel, rerr := filepath.Rel(lib.Root, t.Dir); rerr == nil && !strings.HasPrefix(rel, "..") {
			return fail(exitState, "这份录制已经在长期库里了：%s", t.Dir)
		}
		if t.Meta == nil {
			return fail(exitTarget,
				"这份录制没有身份文件（%s）：长期库按 环境+ENT 分层，进哪一层只能从身份里读，不猜",
				fep.MetaPath(t.Dir, t.ID))
		}
		item, err := lib.Promote(t.Dir, t.ID)
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		return printJSONOrText(map[string]any{
			"id": item.ID, "env": item.Env, "ent": item.Ent, "dir": item.Dir, "files": len(item.Files),
		}, func() {
			fmt.Printf("已入长期库：%s\n", item.Dir)
			fmt.Printf("  环境 %s ／ ENT %s ／ 录制 ID %s（%d 份文件）\n", item.Env, item.Ent, item.ID, len(item.Files))
		})
	},
}

// recRmCmd / recPruneCmd 是"手动清除"与"级联兜底"两条路（策略见 internal/fep/retention.go）：
// 用户主动录的进长期库、**不参与自动清理**，所以要有人能删 —— 前端点一下走 rm，
// 环境被删掉时走级联（设置页那条）或 prune（手改配置那条的兜底）。

var (
	recRmDryRun    bool
	recRmYes       bool
	recPruneDryRun bool
	recPruneYes    bool
)

var recRmCmd = &cobra.Command{
	Use:   "rm <录制 ID>",
	Short: "删掉一份录制（前端手动清除落地的那一下）",
	Long: `删掉一份录制：整份删（目录 + 里面所有文件），并报出删了多少、释放了多少。

只删**长期库**里的那一份；传进来的不在库里就拒（不猜着删）。
缓存目录里的那些（系统自动录的）不用这条 —— 它们走统一清理框架：
tt serve 启动清超过 7 天的、tt cache clear 立刻清。

先看一眼要删什么：加 --dry-run。`,
	Args: exactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		lib, err := library()
		if err != nil {
			return err
		}
		item, found, lerr := lib.Find(id)
		if lerr != nil {
			return fail(exitFailed, "%v", lerr)
		}
		if !found {
			// 在缓存里？如实说"它归统一清理框架管"，别在这里删
			if cache, cerr := cacheDir(); cerr == nil {
				if files, ferr := fep.RecordingFilesFor(cache, id); ferr == nil && len(files) > 0 {
					return fail(exitState,
						"%s 在缓存目录（%s），归统一清理框架管：tt cache clear 清全部，或等启动清理按年龄删",
						id, cache)
				}
			}
			return fail(exitTarget, "长期库里没有 %s（看 tt fep rec list）", id)
		}
		if recRmDryRun {
			return printJSONOrText(map[string]any{"id": id, "dir": item.Dir, "dryRun": true}, func() {
				fmt.Printf("会删：%s（%d 份文件 / %s）\n", item.Dir, len(item.Files), humanBytes(dirBytes(item.Dir)))
			})
		}
		// 破坏性动作要 --yes（与 tt dev tzc unlock / tt update 同一条规矩）：
		// 录制是**不可再生**的 —— 用户再报一次 bug、再操作一遍才有这一份。
		if !recRmYes {
			return fail(exitUsage,
				"删这份录制要显式确认（不可再生）：%s（%d 份文件 / %s）\n  "+
					"确认后加 --yes 重新执行（AI 不得自主加 --yes）",
				item.Dir, len(item.Files), humanBytes(dirBytes(item.Dir)))
		}
		files, bytes, rerr := lib.Remove(item)
		if rerr != nil {
			return fail(exitFailed, "删 %s 失败: %v", item.Dir, rerr)
		}
		return printJSONOrText(map[string]any{"id": id, "dir": item.Dir, "files": files, "bytes": bytes}, func() {
			fmt.Printf("已删 %s：%d 份文件 / %s\n", id, files, humanBytes(bytes))
		})
	},
}

var recPruneCmd = &cobra.Command{
	Use:   "prune [环境名...]",
	Short: "清掉长期库里「环境已不在配置里」的录制（环境被删时的级联兜底）",
	Long: `清掉长期库里那些**环境已经不在 config.json 里**的录制。

为什么要它：设置页删环境时会顺手级联删它的录制库，但环境也可能是**手改配置**删掉的
（那条路没有任何代码跑得到）。所以清理要有一个兜底：拿库里的环境目录名与配置里的
环境列表对一遍，多出来的就是没人认领的那些。

不给环境名就清所有多出来的；给了就只清指名的那些。
先看一眼要删什么：加 --dry-run。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := common.ResolveConfig(true)
		if err != nil {
			return err
		}
		dcfg, err := debug.LoadConfig(cfgPath)
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		lib, err := library()
		if err != nil {
			return err
		}
		inLib, err := lib.Envs()
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		// 配置里的环境名 → 库里那一段目录名（两边必须用同一个函数，否则永远对不上）
		configured := map[string]bool{}
		for _, e := range dcfg.SSHs {
			configured[fep.EnvSegment(e.Name)] = true
		}
		want := map[string]bool{}
		for _, a := range args {
			want[fep.EnvSegment(a)] = true
		}
		type doomed struct {
			Seg    string `json:"segment"`
			Reason string `json:"reason"`
		}
		var targets []doomed
		for _, seg := range inLib {
			switch {
			case len(want) > 0:
				if want[seg] {
					targets = append(targets, doomed{seg, "指名清理"})
				}
			case !configured[seg]:
				targets = append(targets, doomed{seg, "配置里已经没有这个环境"})
			}
		}
		if len(targets) == 0 {
			return printJSONOrText(map[string]any{"removed": 0, "envsInLibrary": len(inLib)}, func() {
				fmt.Printf("长期库里没有该清的环境（库里 %d 个环境）\n", len(inLib))
			})
		}
		type gone struct {
			Seg    string `json:"segment"`
			Reason string `json:"reason"`
			Files  int    `json:"files"`
			Bytes  int64  `json:"bytes"`
		}
		if !recPruneDryRun && !recPruneYes {
			for _, t := range targets {
				fmt.Fprintf(cmd.ErrOrStderr(), "会清 %s（%s）\n", t.Seg, t.Reason)
			}
			return fail(exitUsage, "清这些环境的录制库要显式确认：共 %d 个环境\n  确认后加 --yes 重新执行（AI 不得自主加 --yes）", len(targets))
		}
		var done []gone
		for _, t := range targets {
			if recPruneDryRun {
				done = append(done, gone{t.Seg, t.Reason, 0, dirBytes(filepath.Join(lib.Root, t.Seg))})
				continue
			}
			files, bytes, rerr := lib.RemoveEnv(t.Seg)
			if rerr != nil {
				return fail(exitFailed, "删 %s 的录制库失败: %v", t.Seg, rerr)
			}
			done = append(done, gone{t.Seg, t.Reason, files, bytes})
		}
		return printJSONOrText(map[string]any{"dryRun": recPruneDryRun, "removed": done}, func() {
			for _, d := range done {
				verdict := "已删"
				if recPruneDryRun {
					verdict = "会删"
				}
				fmt.Printf("%s %s（%s）：%d 份文件 / %s\n", verdict, d.Seg, d.Reason, d.Files, humanBytes(d.Bytes))
			}
		})
	},
}

// dirBytes 量一棵目录树有多少字节（给人看的估算；量不到就当 0）。
func dirBytes(dir string) int64 {
	var n int64
	_ = filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if fi, ierr := d.Info(); ierr == nil {
			n += fi.Size()
		}
		return nil
	})
	return n
}

// humanBytes 给人读的体积。
func humanBytes(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	}
	return fmt.Sprintf("%d 字节", n)
}

// listRoot 列出一个落点里的全部录制。
//
// 长期库要按 环境/ENT/ID 三层走（`Library.List`），缓存是平的一层。
// 判据是"这个目录下有没有直接可读的录制件"，而不是拿路径名字去猜。
func listRoot(root recRoot) ([]recRow, error) {
	if ids, err := fep.ListIDs(root.Dir); err == nil && len(ids) > 0 {
		out := make([]recRow, 0, len(ids))
		for _, id := range ids {
			out = append(out, describeRecording(root.Label, root.Dir, id))
		}
		return out, nil
	}
	lib := fep.Library{Root: root.Dir}
	items, err := lib.List()
	if err != nil {
		return nil, err
	}
	out := make([]recRow, 0, len(items))
	for _, it := range items {
		// 身份只填一次：`describeRecording` 自己已经从 `.meta.json` 填过一遍（它就在那个目录里），
		// 这里再调一次 `fillFromMeta` 会把 bytesRT 加**两遍** —— 长期库行体积翻倍、而缓存行正常，
		// 表现就是"界面与 rec list 的体积对不上"（实测 545030 = 2 × 272515）。
		// 库这份 items 已经带了 Meta，用它只是为了省一次读盘，所以只在"它真的没读到自己的身份"时补。
		r := describeRecording(root.Label, it.Dir, it.ID)
		if !r.HasMeta && it.Meta != nil {
			r.Meta = it.Meta
			fillFromMeta(&r)
		}
		out = append(out, r)
	}
	return out, nil
}

// pad / padDisp 按**显示宽度**补齐：中日韩字符占两格，用 %-Ns 补出来的表格
// 在中文列上会歪（实测「（无身份）」那一列一歪，整列就对不上行）。
func pad(s string, n int) string { return s + strings.Repeat(" ", max(0, n-dispWidth(s))) }

func padDisp(s string, n int) string {
	if s == "" {
		s = "-"
	}
	return pad(s, n)
}

// dispWidth 估一段字符串的显示宽度（宽字符算 2）。
func dispWidth(s string) int {
	w := 0
	for _, r := range s {
		switch {
		case r >= 0x1100 && (r <= 0x115f || // 谚文字母
			r >= 0x2e80 && r <= 0xa4cf || // 中日韩部首/汉字/假名
			r >= 0xac00 && r <= 0xd7a3 || // 谚文音节
			r >= 0xf900 && r <= 0xfaff || // 兼容汉字
			r >= 0xfe30 && r <= 0xfe6f || // 兼容形式
			r >= 0xff00 && r <= 0xff60 || // 全角
			r >= 0xffe0 && r <= 0xffe6):
			w += 2
		default:
			w++
		}
	}
	return w
}

// recIndexCmd / recIndexAllCmd 是"索引"这一条（D0）。
//
// 帧索引（`<raw>.frames.jsonl`）记的是**帧级时间 + 字节偏移**，录制时由代理顺手落。
// 它必须能重建，因为两条路都会走到"没有索引"：
//   - 老录制件（那时候还没记）；
//   - 进程被硬杀（索引带缓冲，尾行没落盘 → 读侧报"不完整"）。
//
// 两条路都**不动 raw 一个字节**：索引是派生视图，raw 才是证据。
var (
	recIndexForce bool
	recIndexAll   bool
)

var recIndexCmd = &cobra.Command{
	Use:   "index [录制ID]",
	Short: "给录制件建帧索引（帧级时间 + 字节偏移）：老件补建、索引损坏或不全时重建",
	Long: `现场解析录制件并落一份帧索引（` + "`<raw>.frames.jsonl`" + `，与 raw 同一个干）。

为什么要它：录制件里只有字节、没有时间。timeline.log 有时间，但那是**块级**的、
而且每块截断 1500 字节 —— 拿它当索引源只能得到一个安静的错答案。明细页要的
"这一帧在第几个字节、什么时候来的"，只有在录制那一刻记下来才准确。

索引是**派生**的：raw 是唯一权威，索引随时能删、能重建。重建出来的索引
**不带帧时间**（老件本来就没有），界面会照实标"时间精度：块级"。

  --force  已有索引也重建（默认：已经是完整的就跳过）`,
	Args: atMostOneArg,
	RunE: func(cmd *cobra.Command, args []string) error {
		if recIndexAll {
			return runIndexAll(cmd)
		}
		id := ""
		if len(args) == 1 {
			id = args[0]
		} else if recID != "" {
			id = recID
		}
		t, err := resolveRecording(uiFrom, id)
		if err != nil {
			return err
		}
		rows, err := indexOne(t.Dir, t.ID)
		if err != nil {
			return err
		}
		return printJSONOrText(map[string]any{"id": t.ID, "dir": t.Dir, "indexes": rows}, func() {
			fmt.Printf("录制 %s（%s）\n", t.ID, t.Dir)
			for _, r := range rows {
				fmt.Printf("  %s  %s  %d 帧 / 覆盖 %s%s\n",
					pad(r.Name, 34), padDisp(r.Dir, 14), r.Frames, humanBytes(r.Covered), r.Note())
			}
		})
	},
}

// indexRow 是"建完索引"给用户看的一行。
type indexRow struct {
	Name    string `json:"raw"`
	Dir     string `json:"dir"`
	Frames  int    `json:"frames"`
	Bytes   int64  `json:"bytes"`
	Covered int64  `json:"covered"`
	Rebuilt bool   `json:"rebuilt"`
	Err     string `json:"error,omitempty"`
}

// Note 是这一行后面那句人话（重建了 / 已经有了 / 不是帧流）。
func (r indexRow) Note() string {
	switch {
	case r.Err != "":
		return "　" + r.Err
	case r.Frames == 0:
		return "　（没有帧：这份可能是早期的文本形态录制件）"
	case r.Rebuilt:
		return "　已重建（无帧级时间）"
	default:
		return "　已有索引，跳过（要重建加 --force）"
	}
}

// indexOne 给一份录制的每个方向建索引。
func indexOne(dir, id string) ([]indexRow, error) {
	raws, err := fep.RawFilesFor(dir, id)
	if err != nil {
		return nil, fail(exitFailed, "读 %s: %v", dir, err)
	}
	if len(raws) == 0 {
		return nil, fail(exitTarget, "%s 里没有 %s 的录制件", dir, id)
	}
	rows := make([]indexRow, 0, len(raws))
	for _, raw := range raws {
		b, berr := fep.BuildFrameIndex(raw.Path, raw.Conn, raw.Dir, recIndexForce)
		row := indexRow{Name: raw.Name, Dir: fep.DirLabel(raw.Dir),
			Frames: b.Frames, Bytes: b.Bytes, Covered: b.Covered, Rebuilt: b.Rebuilt}
		if berr != nil {
			row.Err = "建索引失败：" + berr.Error()
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// runIndexAll 给两个默认落点里的**每一份**录制建索引。
//
// 存在的理由：索引是这一轮才加的，库里已经攒下的那些一份都没有 —— 逐个敲 ID
// 既慢又容易漏。它只写派生文件，不动任何录制件。
func runIndexAll(cmd *cobra.Command) error {
	roots, err := recordingRoots()
	if err != nil {
		return err
	}
	type one struct {
		Where string     `json:"where"`
		ID    string     `json:"id"`
		Rows  []indexRow `json:"indexes"`
	}
	var all []one
	for _, root := range roots {
		ids, lerr := listAllIDs(root.Dir)
		if lerr != nil {
			if os.IsNotExist(lerr) {
				continue
			}
			return fail(exitFailed, "读 %s: %v", root.Dir, lerr)
		}
		for _, entry := range ids {
			rows, rerr := indexOne(entry.Dir, entry.ID)
			if rerr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "跳过 %s：%v\n", entry.ID, rerr)
				continue
			}
			all = append(all, one{Where: root.Label, ID: entry.ID, Rows: rows})
		}
	}
	return printJSONOrText(map[string]any{"count": len(all), "recordings": all}, func() {
		if len(all) == 0 {
			fmt.Println("没有录制件要建索引")
			return
		}
		for _, a := range all {
			fmt.Printf("%s %s\n", padDisp(a.Where, 6), a.ID)
			for _, r := range a.Rows {
				fmt.Printf("  %s  %s  %d 帧%s\n", pad(r.Name, 34), padDisp(r.Dir, 14), r.Frames, r.Note())
			}
		}
	})
}

// recRef 是"一份录制住在哪个目录"（建索引的 `--all` 要遍历两处落点）。
type recRef struct{ ID, Dir string }

// listAllIDs 列出一个落点里的全部录制（缓存是平的；长期库是 环境/ENT 两层，
// 所以要走库自己的遍历 —— 与 `rec list` 同一套判据，别在这里再写一遍）。
func listAllIDs(root string) ([]recRef, error) {
	if ids, err := fep.ListIDs(root); err == nil && len(ids) > 0 {
		out := make([]recRef, 0, len(ids))
		for _, id := range ids {
			out = append(out, recRef{ID: id, Dir: root})
		}
		return out, nil
	}
	lib := fep.Library{Root: root}
	items, err := lib.List()
	if err != nil {
		return nil, err
	}
	out := make([]recRef, 0, len(items))
	for _, it := range items {
		out = append(out, recRef{ID: it.ID, Dir: it.Dir})
	}
	return out, nil
}

// indexStatus 说清"这份录制的帧索引是什么状态"，每个方向一句。
//
// 四种状态都要能说出来，因为它们对明细页的意义完全不同：
//
//	34 帧（帧级时间）        → 时间轴能画到帧
//	34 帧（无帧级时间）      → 老件或重建过的：只能画块级
//	34 帧（不完整，可重建）  → 多半是进程被硬杀，尾行没落盘
//	无（可建）              → 老件，还没建过
//
// 一个都不许合并成"有/无"：那会让人以为"有索引"就等于"有时间精度"。
func indexStatus(dir, id string) (string, []indexRow) {
	raws, err := fep.RawFilesFor(dir, id)
	if err != nil || len(raws) == 0 {
		return "（读不到录制件）", nil
	}
	var parts []string
	var rows []indexRow
	for _, raw := range raws {
		ix, rerr := fep.ReadFrameIndex(fep.IndexPathFor(raw.Path))
		row := indexRow{Name: raw.Name, Dir: fep.DirLabel(raw.Dir)}
		var text string
		switch {
		case rerr != nil:
			text = "无（用 tt fep rec index 建）"
			row.Err = "无索引"
		case !ix.Complete() && ix.StaleFormat():
			text = fmt.Sprintf("%d 帧（格式版本 %d 已过时，可重建）", len(ix.Recs), ix.Version)
			row.Frames, row.Bytes, row.Covered = len(ix.Recs), ix.End, ix.Covered()
			row.Err = "索引格式过时"
		case !ix.Complete():
			text = fmt.Sprintf("%d 帧（不完整，可重建）", len(ix.Recs))
			row.Frames, row.Bytes, row.Covered = len(ix.Recs), ix.End, ix.Covered()
			row.Err = "索引不完整"
		case !ix.Timed():
			text = fmt.Sprintf("%d 帧（无帧级时间：老件重建）", len(ix.Recs))
			row.Frames, row.Bytes, row.Covered, row.Rebuilt = len(ix.Recs), ix.End, ix.Covered(), true
		default:
			text = fmt.Sprintf("%d 帧（帧级时间）", len(ix.Recs))
			row.Frames, row.Bytes, row.Covered = len(ix.Recs), ix.End, ix.Covered()
		}
		parts = append(parts, fmt.Sprintf("#%d %s %s", raw.Conn, fep.DirLabel(raw.Dir), text))
		rows = append(rows, row)
	}
	return strings.Join(parts, " ／ "), rows
}

var recShowCmd = &cobra.Command{
	Use:   "show [录制ID]",
	Short: "看一份录制是什么：启动路径、作业、环境·区域·ENT、编译件标识、指纹，以及现在能不能用",
	Long: `看一份录制的身份与指纹。

给录制 ID 时去默认落点找；--from 可以直接指一个 .log 或一个目录（长期库也用这个）。

指纹是**当场从录制件的字节复算**的（编号表：什么种类、在哪个作用域、叫什么名字、编号多少）。
业务数据不进指纹 —— 它能随便复制流转，也不会因为"这次单号不一样"就失效。

有会话在跑时还会**顺手比一次闸门**：这份录制能不能驱动现在这个界面。
不能的话，这里就会看到差在哪、差多少。`,
	Args: atMostOneArg,
	RunE: func(cmd *cobra.Command, args []string) error {
		from, id := uiFrom, ""
		if len(args) == 1 {
			// --from 指文件时以文件为准（下面 resolveRecording 的优先级），指目录时用它选份
			id = args[0]
		} else if recID != "" {
			id = recID
		}
		t, err := resolveRecording(from, id)
		if err != nil {
			return err
		}
		snap := t.snapshot()
		idxText, idxRows := indexStatus(t.Dir, t.ID)
		view := map[string]any{
			"id": t.ID, "dir": t.Dir, "files": t.Files,
			"path": snap.Path, "prog": snap.Prog,
			"fingerprint": snap.FP,
			"indexes":     idxRows,
		}
		if t.Meta != nil {
			view["meta"] = t.Meta
		} else {
			view["meta"] = nil
			view["note"] = "这份录制没有身份文件（.meta.json）：启动路径与编译件标识无从得知，" +
				"指纹是从录制件的字节当场算的"
		}
		// 顺手比一次闸门：看一眼"现在能不能用它"。比的是**按连接**的表。
		recMulti := t.multi()
		view["connections"] = connTables(recMulti)
		var verdict *fep.GateVerdict
		if client, cerr := newClient(); cerr == nil {
			if live, ok, lerr := liveMulti(client); lerr == nil && ok {
				v := fep.GateReplay(recMulti, live)
				verdict = &v
				view["gate"] = v
			}
		}
		return printJSONOrText(view, func() {
			fmt.Printf("录制 ID  : %s\n", t.ID)
			fmt.Printf("落点     : %s\n", t.Dir)
			if t.Meta == nil {
				fmt.Println("身份     : 没有 .meta.json（老录制件，或收工被硬杀）—— 路径与编译件无从得知")
			} else {
				fmt.Printf("启动路径 : %s\n", t.Meta.Path.Label())
				fmt.Printf("作业     : %s\n", dashIfBlank(t.Meta.Prog))
				fmt.Printf("生效身份 : 环境 %s / 区域 %s / %s\n", dashIfBlank(t.Meta.Env), dashIfBlank(t.Meta.Zone), entText(t.Meta.Ent))
				fmt.Printf("寄存主机 : %s\n", dashIfBlank(t.Meta.Host))
				fmt.Printf("画面接线 : %s\n", dashIfBlank(t.Meta.FGLServer))
				fmt.Printf("时间     : %s → %s（%s）\n",
					t.Meta.Started.Format(time.RFC3339),
					t.Meta.Finished.Format(time.RFC3339),
					t.Meta.Duration().Round(time.Second))
				fmt.Printf("编译件   : %s\n", t.Meta.Binary.FingerprintString())
				fmt.Printf("归属     : %s ／ %s\n", t.Meta.Origin.Label(), t.Meta.Retention.Label())
			}
			// 整体（各连接合起来）的指纹：回答"这份录制一共录到了多少界面"。
			// 判闸门用的是**按连接**的表（下面那一行），两者不是一回事。
			fmt.Printf("指纹     : %s（各连接合起来）\n", snap.FP)
			if snap.Note != "" {
				fmt.Printf("读不出表 : %s\n", snap.Note)
			}
			fmt.Printf("连接表   : %s\n", connTables(recMulti))
			fmt.Printf("录制件   : %d 份 AUI 文件\n", len(t.Files))
			fmt.Printf("帧索引   : %s\n", idxText)
			if t.Meta != nil && len(t.Meta.Conns) > 0 {
				fmt.Println("连接     :")
				for _, c := range t.Meta.Conns {
					// 身份里存的是文件名，这里还原成绝对路径 —— 录制件可能已经进了长期库
					// （收工时自动搬），照抄身份里的字符串会得到一个搬到之前的位置。
					fmt.Printf("  #%d  AUI %d 字节 ／ 事件 %d 字节  (%s)\n",
						c.ID, c.BytesRT, c.BytesFE, dashIfBlank(c.Remote))
					fmt.Printf("      %s\n", c.Path(t.Dir, c.RT2GDC))
				}
			}
			if t.Meta != nil && t.Meta.Note != "" {
				fmt.Printf("备注     : %s\n", t.Meta.Note)
			}
			if verdict != nil {
				fmt.Print(verdict.Report())
			} else {
				fmt.Println("闸门     : 没有会话在跑，这一栏等有会话时再看（tt fep rec replay 会自己先比）")
			}
		})
	},
}

// connTables 把"每条连接各多少条"写成一行（`rec show` 与拒绝报文都要它：
// 一个会话可能有好几条连接，而它们可以是完全不同的两棵树）。
func connTables(m fep.MultiSnapshot) string {
	if len(m.Conns) == 0 {
		return "（没有连接）"
	}
	var parts []string
	for _, c := range m.Conns {
		parts = append(parts, fmt.Sprintf("#%d %s", c.Conn, c.FP))
	}
	return strings.Join(parts, " ／ ")
}

func dashIfBlank(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func init() {
	recCmd.AddCommand(recListCmd, recShowCmd, recKeepCmd, recRmCmd, recPruneCmd, recIndexCmd)
	for _, c := range []*cobra.Command{recListCmd, recShowCmd, recIndexCmd} {
		c.Flags().StringVar(&uiFrom, "from", "", "录制件落点：单个 .log，或存放录制件的目录（默认缓存 + 长期库都看）")
	}
	recShowCmd.Flags().StringVar(&recID, "id", "", "录制 ID（与位置参数二选一）")
	recIndexCmd.Flags().StringVar(&recID, "id", "", "录制 ID（与位置参数二选一）")
	recIndexCmd.Flags().BoolVar(&recIndexForce, "force", false, "已有索引也重建（默认跳过完整的那些）")
	recIndexCmd.Flags().BoolVar(&recIndexAll, "all", false, "给两个默认落点里的每一份录制都建索引")
	recRmCmd.Flags().BoolVar(&recRmDryRun, "dry-run", false, "只说要删什么，一个字节都不删")
	recRmCmd.Flags().BoolVar(&recRmYes, "yes", false, "确认删除（录制不可再生；AI 不得自主加）")
	recPruneCmd.Flags().BoolVar(&recPruneDryRun, "dry-run", false, "只列该清的环境，一个字节都不删")
	recPruneCmd.Flags().BoolVar(&recPruneYes, "yes", false, "确认删除（AI 不得自主加）")
}
