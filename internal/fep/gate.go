package fep

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// 指纹闸（R0.5）：回放之前，先回答"这份录制能不能用来驱动**现在这个界面**"。
//
// 为什么必须补这道闸：现有的 AUI 安全闸只判"这个 id 在不在当前界面里"，
// 而另一个会话的编号几乎总能"找到" —— 实测 **6/6 全部落到别的对象上**
// （#327 从「加工类型编号」变成「重测/返工否」的勾选框）**且不报错**。
// 它拦不住的东西，必须有另一样东西拦。
//
// 判据与阈值都是**量出来的**（真录制件，见 internal/fep/README 的「指纹闸」一节）：
//
//	同一路径两次独立会话（调试线，3976 条）      · 编号表逐条相同 → 100%
//	同一路径、录制件 4961 条 vs 起完就录 4046 条  · 4044/4046 = 99%
//	换一条启动路径（debug 的件 vs r.r 的会话）    · 0%
//	换一个作业（同一路径，两份不同作业的件）      · 36% / 79%
//
// 所以 99% 落在一条很宽的空档里：同路径不可能掉到 99% 以下，
// 跨路径/跨作业不可能爬到 99% 以上。
//
// **按连接比，不是把整个会话揉成一张表**：一个会话本来就有多条连接，而且它们可以是
// 完全不同的两棵树 —— 实测 `r.d` 会话里连接 1 是 fgldeb 调试器的界面、连接 2 是作业界面。
// 揉在一起比等于拿"A + B"去比"B"，同一条路径也会判不过（实测 93%）。

// MinLiveMatchPct 是"当前连接的编号表里，必须有多少条能在录制件的同一条连接里原样找到"。
const MinLiveMatchPct = 99

// Snapshot 是一条连接的编号表 + 它是谁。
//
// 录制件与当前会话**用同一个形状**，因为闸门做的事就是"把两面摆在一起对"。
type Snapshot struct {
	Label string `json:"label"`
	ID    string `json:"id,omitempty"`
	// Path 是启动路径。空/PathNone = 没起作业（或身份不详），闸门不拿它比对。
	Path LaunchPath     `json:"path,omitempty"`
	Prog string         `json:"prog,omitempty"`
	Meta *RecordingMeta `json:"meta,omitempty"`
	FP   Fingerprint    `json:"fingerprint"`
	// Note 是"这一面为什么给不出编号表"（读不动、一条 AUI 都没录到……）。
	// 有它才不会把"读失败"与"这份录制里本来就没有界面"混成一句话。
	Note    string    `json:"note,omitempty"`
	Entries []FPEntry `json:"-"`
}

// EmptySnapshot 是"有这一面、但它给不出编号表"的样子（作业还没连上来，或录制件里没有界面）。
func EmptySnapshot(label, note string) Snapshot {
	return Snapshot{Label: label, Note: note}
}

// ConnSnapshot 是某一条连接的编号表。
type ConnSnapshot struct {
	Conn int
	Snapshot
}

// MultiSnapshot 是一个会话（或一份录制）的全部连接。
type MultiSnapshot struct {
	Label string
	Path  LaunchPath
	Prog  string
	Meta  *RecordingMeta
	Conns []ConnSnapshot
}

// NonEmpty 报告这一面有没有能比的连接（有界面，而不是只有握手）。
func (m MultiSnapshot) NonEmpty() bool {
	for _, c := range m.Conns {
		if c.FP.Total > 0 {
			return true
		}
	}
	return false
}

// OfSnapshot 把一个"整会话一张表"的快照当成单连接面（老调用点与单连接场景用它）。
func OfSnapshot(s Snapshot) MultiSnapshot {
	return MultiSnapshot{Label: s.Label, Path: s.Path, Prog: s.Prog, Meta: s.Meta,
		Conns: []ConnSnapshot{{Conn: 0, Snapshot: s}}}
}

// FilesMultiSnapshot 把"若干 AUI 录制文件"按连接各摊一张表（一份录制就是这条路）。
func FilesMultiSnapshot(files []string, label string) (MultiSnapshot, error) {
	out := MultiSnapshot{Label: label}
	for i, f := range files {
		conn := connIDOf(filepath.Base(f))
		if conn == 0 {
			conn = i + 1 // 不是代理落的成对文件（夹具/单文件）：按顺序编号，别都算 0
		}
		m, _, _, err := loadModelFiles([]string{f})
		if err != nil {
			// 某一条连接读不动不该把整份录制判死：照实给它一张空表 + 说明
			out.Conns = append(out.Conns, ConnSnapshot{Conn: conn,
				Snapshot: EmptySnapshot(fmt.Sprintf("%s #%d", label, conn), err.Error())})
			continue
		}
		es := EntriesOf(m)
		out.Conns = append(out.Conns, ConnSnapshot{Conn: conn,
			Snapshot: Snapshot{Label: fmt.Sprintf("%s #%d", label, conn), ID: label,
				Entries: es, FP: FingerprintOf(es)}})
	}
	sort.Slice(out.Conns, func(i, j int) bool { return out.Conns[i].Conn < out.Conns[j].Conn })
	return out, nil
}

// SessionMultiSnapshot 从"录制件目录 + tag"按连接摊表（当前会话就是这条路）。
func SessionMultiSnapshot(dir, tag, label string) (MultiSnapshot, error) {
	models, _, err := ConnectionModels(dir, tag)
	if err != nil {
		return MultiSnapshot{}, err
	}
	out := MultiSnapshot{Label: label}
	for id, m := range models {
		es := EntriesOf(m)
		out.Conns = append(out.Conns, ConnSnapshot{Conn: id,
			Snapshot: Snapshot{Label: fmt.Sprintf("%s #%d", label, id), ID: tag,
				Entries: es, FP: FingerprintOf(es)}})
	}
	sort.Slice(out.Conns, func(i, j int) bool { return out.Conns[i].Conn < out.Conns[j].Conn })
	return out, nil
}

// SnapshotOfFiles 从"若干 rt2gdc 录制文件"摊出**一张**表（整会话合起来看时用它，
// 例如 `rec show` 报一个总指纹）。判定闸门请用 FilesMultiSnapshot。
func SnapshotOfFiles(files []string, label string) (Snapshot, error) {
	m, _, _, err := loadModelFiles(files)
	if err != nil {
		return Snapshot{}, err
	}
	es := EntriesOf(m)
	return Snapshot{Label: label, Entries: es, FP: FingerprintOf(es)}, nil
}

// SnapshotOfSession 整会话一张表（同上，只给人读）。
func SnapshotOfSession(dir, tag, label string) (Snapshot, error) {
	m, _, _, err := LoadModelFrom(dir, tag)
	if err != nil {
		return Snapshot{}, err
	}
	es := EntriesOf(m)
	return Snapshot{Label: label, ID: tag, Entries: es, FP: FingerprintOf(es)}, nil
}

// GateCode 是闸门的判决类别（调用方据此决定退出码与文案）。
type GateCode string

const (
	GateOK          GateCode = "ok"
	GatePathDiff    GateCode = "path"         // 启动路径不符
	GateFPNoMatch   GateCode = "fingerprint"  // 编号表对不上
	GateNoRecording GateCode = "no-recording" // 录制件里没有界面
	GateNoLive      GateCode = "no-live"      // 当前会话还没有界面
)

// ConnPair 是比过的一对连接。
type ConnPair struct {
	Rec  int    `json:"recConn"`
	Live int    `json:"liveConn"`
	Diff FPDiff `json:"diff"`
}

// GateVerdict 是闸门的判决（人读 + JSON 两形态都用它）。
type GateVerdict struct {
	OK     bool     `json:"ok"`
	Code   GateCode `json:"code"`
	Reason string   `json:"reason,omitempty"`

	RecPath  LaunchPath `json:"recPath,omitempty"`
	LivePath LaunchPath `json:"livePath,omitempty"`
	// PathChecked 报告这次到底比没比路径：两边有一边不知道路径时是 false，
	// 这时判决完全靠指纹（**不是**"路径没比就放过"）。
	PathChecked bool `json:"pathChecked"`

	// 被采用（通过）或最接近（不过）的那一对连接的编号表与连接号。
	RecFP    Fingerprint `json:"recFingerprint"`
	LiveFP   Fingerprint `json:"liveFingerprint"`
	RecConn  int         `json:"recConn,omitempty"`
	LiveConn int         `json:"liveConn,omitempty"`
	Diff     FPDiff      `json:"diff"`
	// Pairs 是比过的全部连接对（人读时列出来，省得只看到一个孤零零的百分比）。
	Pairs []ConnPair `json:"pairs,omitempty"`
	// Examples 是"坐标对得上、编号不同"的头几条（跨路径时它就是现场证据）。
	Examples []string `json:"examples,omitempty"`
}

// GateReplay 比对两面，给出"能不能用这份录制驱动当前界面"的判决。
//
// 顺序是有意的：**先判路径**（便宜、确定、报文最好懂），再按连接比指纹
// （贵、但能发现同路径跨编译/跨作业）。任一条不符就拒 —— 这道闸的错法必须是
// "拦住能用的"，不能是"放过不能用的"：放过的后果是安静地点错控件。
func GateReplay(rec, live MultiSnapshot) GateVerdict {
	v := GateVerdict{RecPath: rec.Path, LivePath: live.Path}
	if !rec.NonEmpty() {
		v.Code = GateNoRecording
		v.Reason = fmt.Sprintf("录制件 %s 里没有界面（一条带对象的 AUI 消息都没读到）—— 它多半是"+
			"作业还没连上来就断了的连接，用它回放只会打到空气上", rec.Label) + emptyHint(rec.firstNote())
		return v
	}
	if !live.NonEmpty() {
		v.Code = GateNoLive
		v.Reason = fmt.Sprintf("当前会话（%s）还没有界面：作业还没连上来，或连接已经断了。"+
			"等界面出来再回放（tt fep wait）", live.Label) + emptyHint(live.firstNote())
		return v
	}
	if rec.Path.Known() && live.Path.Known() {
		v.PathChecked = true
		if rec.Path != live.Path {
			v.Code = GatePathDiff
			v.Reason = fmt.Sprintf("启动路径不符：录制件是 %s 录的，当前会话是 %s。"+
				"编号是启动历史的函数（实测跨路径共同条目 0 条相同），跨路径回放会安静地点错控件 —— "+
				"请用同一条路径重录或重起", rec.Path.Label(), live.Path.Label())
			return v
		}
	}

	// 逐对比。同号优先：连接号本身就是"哪个窗口"的第一手线索。
	best := ConnPair{Rec: -1, Live: -1}
	bestMatched := false
	for _, rc := range rec.Conns {
		if rc.FP.Total == 0 {
			continue
		}
		for _, lc := range live.Conns {
			if lc.FP.Total == 0 {
				continue
			}
			p := ConnPair{Rec: rc.Conn, Live: lc.Conn, Diff: CompareFP(rc.Entries, lc.Entries)}
			v.Pairs = append(v.Pairs, p)
			matched := p.Diff.Same > 0 && p.Diff.MatchPct() >= MinLiveMatchPct
			if best.Rec < 0 || better(matched, p, bestMatched, best, rc, lc) {
				best, bestMatched = p, matched
				v.RecFP, v.LiveFP = rc.FP, lc.FP
				v.RecConn, v.LiveConn = rc.Conn, lc.Conn
			}
		}
	}
	v.Diff = best.Diff
	if bestMatched {
		v.OK, v.Code = true, GateOK
		v.Reason = fmt.Sprintf("闸门通过：录制件 #%d 与当前界面 #%d 的编号表对得上 %d/%d 条（%d%%），路径 %s",
			v.RecConn, v.LiveConn, v.Diff.Same, v.Diff.LiveTotal, v.Diff.MatchPct(), pathPhrase(v))
		return v
	}
	v.Code = GateFPNoMatch
	v.Examples = TranslateEntries(recEntriesOf(rec, best.Rec), recEntriesOf(live, best.Live), 5)
	v.Reason = fmt.Sprintf("编号表对不上：最接近的一对是录制件 #%d ↔ 当前界面 #%d，%d 条里只有 %d 条能原样找到（%d%%，要求 ≥%d%%）。"+
		"这份录制不是这个界面录的（换了路径、换了作业，或程序重编译过）%s",
		best.Rec, best.Live, v.Diff.LiveTotal, v.Diff.Same, v.Diff.MatchPct(), MinLiveMatchPct, connSummary(rec, live))
	return v
}

// better 挑"更好的那一对"：先看能不能过闸，再看百分比，最后同号优先。
func better(matched bool, p ConnPair, bestMatched bool, best ConnPair, rc, lc ConnSnapshot) bool {
	if matched != bestMatched {
		return matched
	}
	if p.Diff.MatchPct() != best.Diff.MatchPct() {
		return p.Diff.MatchPct() > best.Diff.MatchPct()
	}
	// 百分比一样时，同号的那一对更像"同一个窗口"
	return rc.Conn == lc.Conn && best.Rec != best.Live
}

func recEntriesOf(m MultiSnapshot, conn int) []FPEntry {
	for _, c := range m.Conns {
		if c.Conn == conn {
			return c.Entries
		}
	}
	return nil
}

// connSummary 把"两边各有哪些连接、各多少条"写成一句话 —— 拒绝时报文里最有用的一段。
func connSummary(rec, live MultiSnapshot) string {
	side := func(m MultiSnapshot) string {
		var parts []string
		for _, c := range m.Conns {
			parts = append(parts, fmt.Sprintf("#%d:%d 条", c.Conn, c.FP.Total))
		}
		return strings.Join(parts, " ")
	}
	return fmt.Sprintf("（录制件连接 %s ／ 当前连接 %s）", side(rec), side(live))
}

// firstNote 取第一条带说明的连接的说明（"为什么这一面是空的"）。
func (m MultiSnapshot) firstNote() string {
	for _, c := range m.Conns {
		if c.Note != "" {
			return c.Note
		}
	}
	return ""
}

func pathPhrase(v GateVerdict) string {
	if !v.PathChecked {
		return "无从比对（有一边没记路径，判决完全靠指纹）"
	}
	return string(v.LivePath)
}

// emptyHint 把"这一面为什么是空的"接在判决后面（没有就不接）。
func emptyHint(note string) string {
	if strings.TrimSpace(note) == "" {
		return ""
	}
	return "（" + note + "）"
}

// Report 给人读的一段（回放命令在动手之前先打它）。
func (v GateVerdict) Report() string {
	var b strings.Builder
	mark := "✓"
	if !v.OK {
		mark = "✗"
	}
	fmt.Fprintf(&b, "%s 指纹闸：%s\n", mark, v.Reason)
	if len(v.Pairs) > 0 {
		fmt.Fprintf(&b, "  录制件编号表 %s ／ 当前界面编号表 %s\n", v.RecFP, v.LiveFP)
		fmt.Fprintf(&b, "  逐条：对得上 %d ／ 编号不同 %d ／ 只在当前 %d ／ 只在录制件 %d\n",
			v.Diff.Same, v.Diff.Differ, v.Diff.OnlyNew, v.Diff.OnlyRec)
		if len(v.Pairs) > 1 {
			var parts []string
			for _, p := range v.Pairs {
				parts = append(parts, fmt.Sprintf("#%d↔#%d %d%%", p.Rec, p.Live, p.Diff.MatchPct()))
			}
			fmt.Fprintf(&b, "  比过的连接对：%s\n", strings.Join(parts, " ／ "))
		}
	}
	for _, e := range v.Examples {
		fmt.Fprintf(&b, "  · %s\n", e)
	}
	return b.String()
}
