package fep

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// 指纹闸（R0.5）：回放之前，先回答"这份录制能不能用来驱动**现在这个界面**"。
//
// 为什么必须补这道闸：AUI 安全闸只判"这个 id 在不在当前界面里"，
// 而另一个会话的编号几乎总能"找到" —— 实测 **6/6 全部落到别的对象上**
// （#327 从「加工类型编号」变成「重测/返工否」的勾选框）**且不报错**。
//
// **判据是契约 A21（指纹 v3，非对称覆盖）**，不是号码：
//
//	通过 ⟺ ① |live| ≥ max(10, 0.5×|rec|)（证据量）
//	      且 ② live→rec 身份覆盖 ≥ 0.98
//	      且 ③ live→rec 边覆盖 ≥ 0.98
//	      且 ④ live 多余身份 ≤ 2%
//	      且 ⑤ 类型直方图差异 ≤ 2%（按 live 的身份集合收窄后比）
//
// 阈值写死在 `fingerprint.go`（FPEvidenceMinNodes / FPEvidenceRecShare /
// FPIdentJaccardMin / FPEdgeIsoMin / FPTypeHistMaxDiff），既不读配置，也不留
// "仅凭 idDigest 相等就放行"的分支 —— 那正是 v2 的误拒来源：编号是启动历史的函数，
// 同一界面换一条启动路径，编号**一条都不相同**。
//
// 方向是**单向**的（live→rec）：录制件是整会话、当前只开半屏是**常态**（会话渐进），
// 不能因此误拒；"live 有、rec 没有"才是界面真的变了。
//
// v2 的"门槛 99%（Same/LiveTotal）"已被 A21 取代：逐条口径（FPDiff.Same/Differ/…）
// 保留下来**只作诊断**，判决只看上面五个判据 + 分两栏的差异清单。
//
// **按连接比，不是把整个会话揉成一张表**：一个会话本来就有多条连接，而且它们可以是
// 完全不同的两棵树 —— 实测 `r.d` 会话里连接 1 是 fgldeb 调试器的界面、连接 2 是作业界面。
// 揉在一起比等于拿"A + B"去比"B"，同一条界面也会判不过。所以判据是"存在一条连接对得上"。

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
		if c.FP.Total > 0 || len(c.Entries) > 0 {
			return true
		}
	}
	return false
}

// usable 报告这一条连接能不能进比对（有编号表或条目才有得比）。
func (c ConnSnapshot) usable() bool { return c.FP.Total > 0 || len(c.Entries) > 0 }

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
	GateFPNoMatch   GateCode = "fingerprint"  // 指纹 v3 判不过（身份/引用边/类型直方图）
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
	// 这时判决完全靠指纹 v3（**不是**"路径没比就放过"）。
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
	// Diagnostics 是"读到旧东西、已忽略"的记账（A19：旧 v2 指纹字段读到就忽略并记诊断）。
	// 它**不影响判决**，只是别让人以为那份指纹还算数。
	Diagnostics []string `json:"diagnostics,omitempty"`
}

// GateReplay 比对两面，给出"能不能用这份录制驱动当前界面"的判决。
//
// 顺序是有意的：**先判路径**（便宜、确定、报文最好懂），再按连接比指纹 v3
// （贵、但能发现同路径下的另一张界面/另一张表单）。任一条不符就拒 —— 这道闸的错法
// 必须是"拦住能用的"，不能是"放过不能用的"：放过的后果是安静地点错控件。
//
// **判决只看 JudgeSameInterface 的三条判据**（A19）。这里没有任何一条分支会
// 因为 `idDigest`（`Fingerprint.Hash`）相等而放行 —— 那等于把 v2 借尸还魂。
func GateReplay(rec, live MultiSnapshot) GateVerdict {
	v := GateVerdict{RecPath: rec.Path, LivePath: live.Path}
	v.Diagnostics = append(v.Diagnostics, legacyNotes(rec, "录制件")...)
	v.Diagnostics = append(v.Diagnostics, legacyNotes(live, "当前会话")...)
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
				"请用同一条路径重录或重起（指纹 v3 已经不拿编号判界面，路径仍是另一道独立的闸）",
				rec.Path.Label(), live.Path.Label())
			return v
		}
	}

	// 逐对比。同号优先：连接号本身就是"哪个窗口"的第一手线索。
	best := ConnPair{Rec: -1, Live: -1}
	bestMatched := false
	for _, rc := range rec.Conns {
		if !rc.usable() {
			continue
		}
		for _, lc := range live.Conns {
			if !lc.usable() {
				continue
			}
			diff, matched := JudgeSameInterface(rc.Entries, lc.Entries)
			p := ConnPair{Rec: rc.Conn, Live: lc.Conn, Diff: diff}
			v.Pairs = append(v.Pairs, p)
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
		v.Reason = fmt.Sprintf("闸门通过（指纹 v3 · 非对称覆盖）：录制件 #%d ↔ 当前界面 #%d —— "+
			"身份覆盖 %.4f（≥%.2f）／ 边覆盖 %.4f（≥%.2f）／ 多余身份 %.4f（≤%.2f）／ "+
			"类型直方图差异 %.4f（≤%.2f）；当前 %d 个身份、%d 条边（录制件 %d / %d），路径 %s",
			v.RecConn, v.LiveConn, v.Diff.LiveCoverage, FPIdentJaccardMin,
			v.Diff.EdgeCoverage, FPEdgeIsoMin, v.Diff.ExtraShare, FPTypeHistMaxDiff,
			v.Diff.TypeDiff, FPTypeHistMaxDiff,
			v.Diff.IdentLive, v.Diff.EdgeLive, v.Diff.IdentRec, v.Diff.EdgeRec, pathPhrase(v))
		return v
	}
	v.Code = GateFPNoMatch
	v.Examples = TranslateEntries(recEntriesOf(rec, best.Rec), recEntriesOf(live, best.Live), 5)
	v.Reason = fmt.Sprintf("指纹 v3 判不过（非对称覆盖）：最接近的一对是录制件 #%d ↔ 当前界面 #%d —— "+
		"身份覆盖 %.4f（要求 ≥%.2f）／ 边覆盖 %.4f（要求 ≥%.2f）／ 多余身份 %.4f（要求 ≤%.2f）／ "+
		"类型直方图差异 %.4f（要求 ≤%.2f）；证据量 %s。%s差异清单分两栏见下%s",
		best.Rec, best.Live, v.Diff.LiveCoverage, FPIdentJaccardMin,
		v.Diff.EdgeCoverage, FPEdgeIsoMin, v.Diff.ExtraShare, FPTypeHistMaxDiff,
		v.Diff.TypeDiff, FPTypeHistMaxDiff, evidencePhrase(v.Diff),
		failPhrase(v.Diff), connSummary(rec, live))
	return v
}

// evidencePhrase 说清①（证据量）这一条过没过。
func evidencePhrase(d FPDiff) string {
	if d.EvidenceOK {
		return fmt.Sprintf("够（当前 %d 条／录制件 %d 条）", d.LiveTotal, d.RecTotal)
	}
	return fmt.Sprintf("不足（当前只有 %d 条，至少要 max(%d, %.0f%%×录制件 %d 条)）",
		d.LiveTotal, FPEvidenceMinNodes, FPEvidenceRecShare*100, d.RecTotal)
}

// failPhrase 一句话说清"差在哪几级、差多少"（拒绝报文的第一段）。
// 两栏口径：**意外**（live 有 rec 无）才是拒的理由；**尚未加载**只报数量，不作理由。
func failPhrase(d FPDiff) string {
	var bad []string
	if !d.EvidenceOK {
		bad = append(bad, "证据不足（界面太小，百分比没有意义）")
	}
	if share := d.ExtraShare; share > FPTypeHistMaxDiff {
		bad = append(bad, fmt.Sprintf("界面多出 %d 个录制件里没有的身份（%.2f%%）", len(d.IdentUnexpected), share*100))
	}
	if d.EdgeCoverage < FPEdgeIsoMin {
		bad = append(bad, fmt.Sprintf("有 %d 条边录制件里没有", len(d.EdgeUnexpected)))
	}
	if d.TypeDiff > FPTypeHistMaxDiff {
		bad = append(bad, fmt.Sprintf("%d 类节点数变了", len(d.TypeChanged)))
	}
	if len(bad) == 0 {
		return "（五个判据的数值都在门槛内，但不该走到这里）"
	}
	return strings.Join(bad, "；") + "。"
}

// better 挑"更好的那一对"：先看能不能过闸，再看三条判据里最弱的那条（Fit），最后同号优先。
func better(matched bool, p ConnPair, bestMatched bool, best ConnPair, rc, lc ConnSnapshot) bool {
	if matched != bestMatched {
		return matched
	}
	if p.Diff.Fit() != best.Diff.Fit() {
		return p.Diff.Fit() > best.Diff.Fit()
	}
	// 相似度一样时，同号的那一对更像"同一个窗口"
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

// legacyNotes 把"这一面带着旧 v2 指纹"记成诊断（A19：读到就忽略，不许因此崩/拒/放）。
func legacyNotes(m MultiSnapshot, side string) []string {
	var out []string
	for _, c := range m.Conns {
		if n := LegacyFingerprintNotice(c.FP); n != "" {
			out = append(out, fmt.Sprintf("%s #%d：%s", side, c.Conn, n))
			continue
		}
		if c.Meta != nil {
			if n := LegacyFingerprintNotice(c.Meta.Fingerprint); n != "" {
				out = append(out, fmt.Sprintf("%s #%d（身份文件）：%s", side, c.Conn, n))
			}
		}
	}
	return out
}

// connSummary 把"两边各有哪些连接、各多少条"写成一句话 —— 拒绝时报文里最有用的一段。
func connSummary(rec, live MultiSnapshot) string {
	side := func(m MultiSnapshot) string {
		var parts []string
		for _, c := range m.Conns {
			parts = append(parts, fmt.Sprintf("#%d:%d 条/%d 身份", c.Conn, c.FP.Total, c.FP.IdentN))
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
		return "无从比对（有一边没记路径，判决完全靠指纹 v3）"
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
//
// A21 要求：不通过时**必须**打差异清单，且**分两栏**——
// **意外**（live 有、rec 无）是拒的理由；**尚未加载**（rec 有、live 无）只作信息。
// 两栏混在一起，"半屏"就永远读不懂（这正是 A19 被改掉的原因）。
func (v GateVerdict) Report() string {
	var b strings.Builder
	mark := "✓"
	if !v.OK {
		mark = "✗"
	}
	fmt.Fprintf(&b, "%s 指纹闸：%s\n", mark, v.Reason)
	if len(v.Pairs) > 0 {
		fmt.Fprintf(&b, "  判据（A21 冻结，单向 live→rec）：身份覆盖 %.4f（≥%.2f）／ 边覆盖 %.4f（≥%.2f）／ 多余身份 %.4f（≤%.2f）／ 类型直方图差异 %.4f（≤%.2f）／ 证据量 %s\n",
			v.Diff.LiveCoverage, FPIdentJaccardMin, v.Diff.EdgeCoverage, FPEdgeIsoMin,
			v.Diff.ExtraShare, FPTypeHistMaxDiff, v.Diff.TypeDiff, FPTypeHistMaxDiff,
			evidencePhrase(v.Diff))
		fmt.Fprintf(&b, "  录制件指纹 %s ／ 当前界面指纹 %s\n", v.RecFP, v.LiveFP)
		fmt.Fprintf(&b, "  逐条（诊断）：对得上 %d ／ 编号不同 %d ／ 只在当前 %d ／ 只在录制件 %d\n",
			v.Diff.Same, v.Diff.Differ, v.Diff.OnlyNew, v.Diff.OnlyRec)
		if len(v.Pairs) > 1 {
			var parts []string
			for _, p := range v.Pairs {
				parts = append(parts, fmt.Sprintf("#%d↔#%d 身份覆盖%.2f/边覆盖%.2f/类型差%.2f",
					p.Rec, p.Live, p.Diff.LiveCoverage, p.Diff.EdgeCoverage, p.Diff.TypeDiff))
			}
			fmt.Fprintf(&b, "  比过的连接对：%s\n", strings.Join(parts, " ／ "))
		}
		fmt.Fprintf(&b, "  差异清单（两栏）：意外 %d 个身份 / %d 条边（有它才拒）；尚未加载 %d 个身份 / %d 条边（只作信息）\n",
			listTrueCount(v.Diff.IdentUnexpected), listTrueCount(v.Diff.EdgeUnexpected),
			listTrueCount(v.Diff.IdentPending), listTrueCount(v.Diff.EdgePending))
		printList(&b, "  ［意外］身份（当前有、录制件没有 —— 拒的理由）", v.Diff.IdentUnexpected)
		printList(&b, "  ［意外］边（当前有、录制件没有 —— 拒的理由）", v.Diff.EdgeUnexpected)
		printList(&b, "  ［尚未加载］身份（录制件有、当前还没走到 —— 正常）", v.Diff.IdentPending)
		printList(&b, "  ［尚未加载］边（录制件有、当前还没走到 —— 正常）", v.Diff.EdgePending)
		printList(&b, "  ［类型变了］（按当前的身份集合比）", v.Diff.TypeChanged)
	}
	for _, d := range v.Diagnostics {
		fmt.Fprintf(&b, "  · 诊断：%s\n", d)
	}
	for _, e := range v.Examples {
		fmt.Fprintf(&b, "  · %s\n", e)
	}
	return b.String()
}

// printList 打印差异清单的一行（空列表也要说"没有"，免得人以为漏打了）。
func printList(b *strings.Builder, label string, items []string) {
	if len(items) == 0 {
		fmt.Fprintf(b, "  %s：无\n", label)
		return
	}
	fmt.Fprintf(b, "  %s：%s\n", label, strings.Join(items, " "))
}
