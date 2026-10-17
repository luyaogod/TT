package web

// 录制**明细页**的读侧（计划 C 的 D1）。
//
// 四个端点都用这里的同一份解析结果：
//
//	GET /api/recordings/{id}/overview   概览：各方向字节/帧数/索引状态 + 条带要的帧数组
//	GET /api/recordings/{id}/timeline   分页的帧列表（全文流）
//	GET /api/recordings/{id}/steps      三层模型：语义动作 > 原始事件 > 运行期回应
//	GET /api/recordings/{id}/raw        某一帧的原始字节（hex + 可打印形态）
//
// 三条纪律（与 internal/fep 那份计划里的 H 条一一对应）：
//
//  1. **raw 是唯一权威**：位置（off/len）**现场**从字节解出来，帧索引只提供**时刻**。
//     所以索引缺了、坏了、格式过时，明细页照样能列出每一帧 —— 只是时间一栏说"不知道"
//     （H5/H8）。绝不用索引去"修正"位置。
//  2. **语义层读时算**：动作/步的定义改过多次（回放白名单从"滤导航"变成"全放行"），
//     落盘必然漂移。每次请求现算，代价是几十毫秒。
//  3. **不编**：解析不出名字就显示 id（H4）；没有帧级时间就说"块级"，不插值、不猜。

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tt/internal/fep"
)

// recLoc 是"一份录制住在哪"。
type recLoc struct {
	ID    string
	Dir   string
	Where string // 长期库 / 缓存
	Meta  *fep.RecordingMeta
}

// locateRecording 按 ID 找一份录制：长期库优先，再缓存（与 `tt recording show` 同一优先级）。
func (s *Server) locateRecording(id string) (recLoc, error) {
	_, cacheDir, lib, err := s.recDirs()
	if err != nil {
		return recLoc{}, err
	}
	if it, found, lerr := lib.Find(id); lerr == nil && found {
		return recLoc{ID: it.ID, Dir: it.Dir, Where: "长期库", Meta: it.Meta}, nil
	}
	if files, ferr := fep.RecordingFilesFor(cacheDir, id); ferr == nil && len(files) > 0 {
		mt, _ := fep.ReadMeta(cacheDir, id) // 没有身份也能读（老件）：不是错误
		return recLoc{ID: id, Dir: cacheDir, Where: "缓存", Meta: mt}, nil
	}
	return recLoc{}, fmt.Errorf("找不到录制 %s（长期库与缓存都没有）", id)
}

// recIndexInfo 是一个方向的索引状态（界面要照实说，不许糊成"有/无"）。
type recIndexInfo struct {
	// State: complete（完整且格式对）/ stale（格式过时，要重建）/ incomplete（尾行缺失，
	// 多半是进程被硬杀）/ missing（没有）。
	State   string `json:"state"`
	Frames  int    `json:"frames"`
	Covered int64  `json:"covered"`
	Rebuilt bool   `json:"rebuilt"`
	Version int    `json:"version,omitempty"`
	Bad     int    `json:"bad,omitempty"`
}

// recFrame 是读侧的一帧。
//
// 这里**刻意不用 omitempty**：`t=0` 表示"时刻不知道"、`q=0` 是一条真实存在的序号
// （实测事件侧就是从 `event _om 0` 开始的）。省掉它们会让前端分不清"没有"与"是 0"，
// 而这两种情形在界面上该长得完全不同。
type recFrame struct {
	Off   int64  `json:"o"`
	Len   int    `json:"n"`
	T     int64  `json:"t"` // 0 = 不知道（没有索引 / 索引是事后重建的）
	Conn  int    `json:"c"`
	Dir   int    `json:"d"`
	Kind  string `json:"k"`
	Seq   int    `json:"q"`
	IDRef int    `json:"r"`
	// Text 是载荷（AUI 帧才有）。**原样**留着：截断是渲染层的事，
	// 而"这一步到底发了什么"必须能从这份数据里逐字查到。
	Text string `json:"-"`
}

// recDir 是一个方向（一个 raw 文件）读出来的东西。
type recDir struct {
	Raw    fep.RawInfo
	Bytes  int64
	Frames []recFrame
	Index  recIndexInfo
	Noise  []fep.ByteRange
	// model / res 只有**界面侧**（rt2gdc）才有：名字是从这一侧的 AUI 解出来的。
	model *fep.Model
	res   *fep.Resolver
}

// recAction 是三层模型的第一层：一个语义动作（点一下 / 输入一个值）。
type recAction struct {
	N     int    `json:"n"`
	Kind  string `json:"kind"` // click | input
	ID    int    `json:"id"`
	Value string `json:"value,omitempty"`
	Conn  int    `json:"conn"`
	// Frame 是它出自哪一帧。行 key 用 `连接:帧偏移`（H3：步号随 --stop-at/--skip 变，不是身份）。
	Frame recFrame `json:"frame"`
	// FrameText 是这一帧的完整载荷。事件帧都很小（几 KB），所以原样带着 ——
	// 表格要显示"这一步发出的原文"，而帧的原始字节走 /raw 另取。
	FrameText string         `json:"frameText,omitempty"`
	Target    fep.TargetInfo `json:"target"`
	// Events 是这一帧里的**全部**原始事件（一条 `event _om` 帧可以带好几个：
	// 实测 34 帧 → 63 个事件，所以"帧数"与"步数"不是一回事，H2）。
	Events []string `json:"events"`
	// Dropped 是这一帧里被语义层丢掉的原始事件数（导航/状态上报）：
	// 照实报出来，用户才不会以为"丢事件了"。
	Dropped int `json:"dropped"`
	// Responses 是运行期在这一步之后、下一步之前推回来的帧（时间窗归属，见 attributeResponses）。
	Responses []recResponse `json:"responses,omitempty"`
	// ResponsesTotal 是**一共**有多少帧落在这个窗口里（列表最多列 20 条，数字照实给）。
	ResponsesTotal int `json:"responsesTotal,omitempty"`
	// ResponsesState 是机器可读的那一半（界面据此选徽标，**不许**去匹配下面那句人话）：
	//
	//	attributed —— 窗口有效,有回应（看 ResponsesTotal）
	//	none       —— 窗口有效,运行期确实没回（"本来就没应答"或"没处理"）
	//	unknown    —— **归不了**:没有帧级时间、窗口退化、或这条连接没录到界面侧
	//
	// 为什么把 none 与 unknown 分开:它们的**下一步**完全不同 —— "没回包"要去看回放日志与作业
	// 状态,"归不了"是这一份录制的时间精度不够。把它们都画成"无回包"就是在冤枉运行期。
	ResponsesState string `json:"responsesState"`
	// ResponsesNote 是"回应归属不了"时的原因（没有帧级时间、窗口退化等）。**不许**留空装没事。
	ResponsesNote string `json:"responsesNote,omitempty"`
}

// recResponse 是一条"被这一步引出的"运行期帧（只带界面要看的几样）。
type recResponse struct {
	Off  int64  `json:"o"`
	Len  int    `json:"n"`
	T    int64  `json:"t"` // 0 = 不知道（同 recFrame）
	Conn int    `json:"c"`
	Kind string `json:"k"`
	Seq  int    `json:"q"`
	// Preview 是载荷开头的一段（全文用 /raw 取）。
	Preview string `json:"preview,omitempty"`
	// Bytes 是这一帧之后到下一帧之前运行期一共推了多少字节 ——
	// "这一步被运行期确认了"的量度（没有回包 ≠ 0 毫秒，见 H8/D7）。
	Bytes int `json:"bytes,omitempty"`
}

// recBlock 是块级时间的刻度（老件专用，来自 timeline.log）。
type recBlock struct {
	T     int64 `json:"t"`
	Dir   int   `json:"d"`
	Bytes int   `json:"n,omitempty"`
}

// recRead 是一份录制的读侧全貌。
type recRead struct {
	ID        string
	Dir       string
	Where     string
	Meta      *fep.RecordingMeta
	Dirs      []*recDir
	Actions   []*recAction
	EventN    int
	FrameN    int
	Precision string // frame | block
	Blocks    []recBlock
	Notes     []string
}

// dirOf 找一个方向（conn+dir）的读侧结果。
func (d *recRead) dirOf(conn, dir int) *recDir {
	for _, x := range d.Dirs {
		if x.Raw.Conn == conn && x.Raw.Dir == dir {
			return x
		}
	}
	return nil
}

// loadRecording 把一份录制读成读侧模型。
//
// 每次请求重新解析：M1 先把**正确性**钉住（哪个数字从哪来都必须说得清），缓存留给后面。
// 实测一份 448 KB / 383 帧的录制，从字节解到动作表约 10 ms 量级。
func (s *Server) loadRecording(id string) (*recRead, error) {
	loc, err := s.locateRecording(id)
	if err != nil {
		return nil, err
	}
	raws, err := fep.RawFilesFor(loc.Dir, loc.ID)
	if err != nil {
		return nil, err
	}
	if len(raws) == 0 {
		return nil, fmt.Errorf("%s 里没有 %s 的录制件", loc.Dir, loc.ID)
	}
	out := &recRead{ID: loc.ID, Dir: loc.Dir, Where: loc.Where, Meta: loc.Meta, Precision: "frame"}

	for _, raw := range raws {
		rd, rerr := readOneDir(raw)
		if rerr != nil {
			// 一个方向读不动不该让整页打不开：照实记一条说明，另一侧照常显示。
			out.Notes = append(out.Notes, fmt.Sprintf("%s 读不动：%v", raw.Name, rerr))
			continue
		}
		out.Dirs = append(out.Dirs, rd)
		out.FrameN += len(rd.Frames)
	}
	if len(out.Dirs) == 0 {
		return nil, fmt.Errorf("%s 的录制件一个方向都读不出来", loc.ID)
	}

	// 界面侧（rt2gdc）每个连接建一棵 AUI：名字只能从**这份录制自己的**树里解（H1）。
	for _, rd := range out.Dirs {
		if rd.Raw.Dir != fep.DirRuntime {
			continue
		}
		m := fep.NewModel()
		for _, fr := range rd.Frames {
			if fr.Kind == "cache" || fr.Text == "" {
				continue
			}
			m.ApplyMessage(fr.Text)
		}
		rd.model, rd.res = m, fep.NewResolver(m)
	}

	// 时间精度：任何一个方向有帧级时间就是 frame，否则退回块级。
	//
	// 三种"没有帧级时间"必须分开说，因为它们对用户的**下一步**完全不同：
	//   - 索引在、但是事后重建的 → 位置准确，时刻当时没记（重建补不回时刻）；
	//   - 索引不在（老件/写失败）→ 现在重建也只能补位置，想有时刻要**重新录一轮**；
	//   - 连 timeline.log 都没有 → 时间完全未知，只按字节序排。
	timed := false
	for _, rd := range out.Dirs {
		for _, fr := range rd.Frames {
			if fr.T > 0 {
				timed = true
				break
			}
		}
		if timed {
			break
		}
	}
	if !timed {
		out.Precision = "block"
		out.Blocks = s.readBlocks(loc)
		rebuilt, missing := 0, 0
		for _, rd := range out.Dirs {
			switch rd.Index.State {
			case "missing":
				missing++
			default:
				if rd.Index.Rebuilt {
					rebuilt++
				}
			}
		}
		switch {
		case rebuilt > 0:
			out.Notes = append(out.Notes,
				"帧索引是**事后重建**的：每一帧的位置准确，但时刻当时没记下来（重建补不回时刻）")
		case missing > 0:
			out.Notes = append(out.Notes,
				"这一份录制录的时候还没有帧索引（本功能是这一轮新加的）：可以重建出位置，"+
					"但时刻只有块级 —— 想要帧级时间要重新录一轮")
		}
		if len(out.Blocks) == 0 {
			out.Notes = append(out.Notes,
				"这一份连 timeline.log 都没有：**时间完全未知**，界面只按各方向的字节序排列")
		} else {
			out.Notes = append(out.Notes, fep.TimelineNote(len(out.Blocks)))
		}
	}

	// 三层模型的第一层：语义动作。只有**事件侧**（gdc2rt）才有动作。
	out.Actions = buildActions(out)
	attributeResponses(out)
	return out, nil
}

// readOneDir 读一个方向：现场解出每一帧的位置，再用索引补上时刻。
func readOneDir(raw fep.RawInfo) (*recDir, error) {
	blob, err := os.ReadFile(raw.Path)
	if err != nil {
		return nil, err
	}
	rd := &recDir{Raw: raw, Bytes: int64(len(blob)), Index: recIndexInfo{State: "missing"}}

	// 时刻表：索引给的。**只用来补时刻**，位置一律来自下面的现场解析。
	times := map[int64]int64{}
	ix, ierr := fep.ReadFrameIndex(fep.IndexPathFor(raw.Path))
	switch {
	case ierr != nil:
		rd.Index.State = "missing"
	case ix.StaleFormat():
		rd.Index = recIndexInfo{State: "stale", Frames: len(ix.Recs), Covered: ix.Covered(), Rebuilt: ix.Rebuilt, Version: ix.Version, Bad: ix.Bad}
	case !ix.Complete():
		rd.Index = recIndexInfo{State: "incomplete", Frames: len(ix.Recs), Covered: ix.Covered(), Rebuilt: ix.Rebuilt, Version: ix.Version, Bad: ix.Bad}
	default:
		rd.Index = recIndexInfo{State: "complete", Frames: len(ix.Recs), Covered: ix.Covered(), Rebuilt: ix.Rebuilt, Version: ix.Version, Bad: ix.Bad}
		for _, r := range ix.Recs {
			if r.T > 0 {
				times[r.Off] = r.T
			}
		}
		rd.Noise = ix.Noise()
	}

	for _, sp := range fep.DecodeOffsets(blob) {
		kind, seq, idRef := fep.FrameKind(sp)
		fr := recFrame{Off: sp.Off, Len: sp.Len, Conn: raw.Conn, Dir: raw.Dir, Kind: kind, Seq: seq, IDRef: idRef, T: times[sp.Off]}
		if sp.Type == fep.FrameTypeAUI {
			fr.Text = strings.TrimSpace(string(sp.Payload))
		}
		rd.Frames = append(rd.Frames, fr)
	}
	return rd, nil
}

// readBlocks 读块级时间刻度（老件降级用）。
func (s *Server) readBlocks(loc recLoc) []recBlock {
	var out []recBlock
	day := time.Now()
	if loc.Meta != nil && !loc.Meta.Started.IsZero() {
		day = loc.Meta.Started
	}
	files, err := os.ReadDir(loc.Dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range files {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "-timeline.log") {
			continue
		}
		if got, ok := fep.RecordingIDOfAny(e.Name()); !ok || got != loc.ID {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, n := range names {
		blocks, berr := fep.ReadTimeline(filepath.Join(loc.Dir, n), day)
		if berr != nil {
			continue
		}
		for _, b := range blocks {
			if b.Bad || b.T.IsZero() {
				continue
			}
			out = append(out, recBlock{T: b.T.UnixMilli(), Dir: b.Dir, Bytes: len(b.Raw)})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].T < out[j].T })
	return out
}

// buildActions 从事件侧解出语义动作（H2：`event _om` 的序号 ≠ 步号；一个帧可含多个事件）。
func buildActions(d *recRead) []*recAction {
	var out []*recAction
	for _, rd := range d.Dirs {
		if rd.Raw.Dir != fep.DirFront {
			continue
		}
		for _, fr := range rd.Frames {
			if fr.Kind != "event" || fr.Text == "" {
				continue
			}
			events := fep.SplitEvents(fr.Text)
			if len(events) == 0 {
				continue
			}
			d.EventN += len(events)
			acts := fep.SemanticActions(events)
			for _, a := range acts {
				n := len(out) + 1
				out = append(out, &recAction{
					N: n, Kind: string(a.Kind), ID: a.ID, Value: a.Value, Conn: fr.Conn,
					Frame: fr, FrameText: fr.Text, Target: targetOf(d, fr.Conn, a.ID),
					Events: events, Dropped: len(events) - len(acts),
				})
			}
		}
	}
	return out
}

// targetOf 问**这份录制自己的**那棵树要名字（H1：id 不是身份，跨界面会撞）。
func targetOf(d *recRead, conn, id int) fep.TargetInfo {
	rd := d.dirOf(conn, fep.DirRuntime)
	if rd == nil {
		return fep.TargetInfo{
			ID: id, Confidence: fep.ConfNone, Source: "unresolved",
			Why: fmt.Sprintf("这条连接（#%d）没有录到界面侧：动作指向的对象叫什么无从得知", conn),
		}
	}
	return fep.DescribeTarget(rd.model, rd.res, id)
}

// attributeResponses 把"这一步引出的运行期回应"按**时间窗**归到动作上。
//
// 归属规则（写在这里，因为它是唯一的定义处）：同一条连接上，时刻落在
// `[本动作的帧时刻, 下一个动作的帧时刻)` 里的**界面侧**帧，算这一步引出的；
// 最后一个动作用 `[t, +∞)`。每一帧只属于一个动作（相邻窗口不重叠）。
//
// 为什么必须按时间而不是按序号：两个方向是**两条独立的字节流**（各自一个文件），
// 序号又分属两侧（`_om` 是前端的，`om` 是运行期的），没有任何一个数字能把它们排到一起。
// 唯一的共同坐标是时间 —— 而没有帧级时间时，就照实说"归属不了"，**不猜**（H8）。
//
// 窗口会**退化**：打点精度是 1 ms（块级），两次点击落在同一毫秒、或落在同一块里时，
// `from == to` —— 这时"哪些回包属于它"是真的分不出来，于是照实说，而不是把回包
// 硬塞给后一步（那会让人以为第一步没被处理）。
func attributeResponses(d *recRead) {
	if d.Precision != "frame" {
		for _, a := range d.Actions {
			a.ResponsesState = "unknown"
			a.ResponsesNote = "这一份没有帧级时间（索引缺失或由老件重建）：回答归不到具体某一步，不猜"
		}
		return
	}
	for i, a := range d.Actions {
		if a.Frame.T == 0 {
			a.ResponsesState = "unknown"
			a.ResponsesNote = "这一帧没有时刻（索引里缺这一条）：回答归不到它，不猜"
			continue
		}
		from := a.Frame.T
		to := int64(1) << 62
		for _, b := range d.Actions[i+1:] {
			if b.Conn == a.Conn && b.Frame.T > 0 {
				to = b.Frame.T
				break
			}
		}
		if to <= from {
			a.ResponsesState = "unknown"
			a.ResponsesNote = "这一步与下一步的时刻相同（帧时间是块级打点，精度 1 ms）：" +
				"这段时间里的回包属于哪一步分不出来，不猜"
			continue
		}
		rd := d.dirOf(a.Conn, fep.DirRuntime)
		if rd == nil {
			a.ResponsesState = "unknown"
			a.ResponsesNote = fmt.Sprintf("这条连接（#%d）没有录到界面侧：没有可归属的回应", a.Conn)
			continue
		}
		const maxList = 20
		for _, fr := range rd.Frames {
			if fr.T < from || fr.T >= to {
				continue
			}
			if fr.Kind == "cache" {
				continue // 图片缓存不算"运行期的回答"
			}
			a.ResponsesTotal++
			if len(a.Responses) >= maxList {
				continue // 数字照实给（ResponsesTotal），列表不刷屏
			}
			a.Responses = append(a.Responses, recResponse{
				Off: fr.Off, Len: fr.Len, T: fr.T, Conn: fr.Conn, Kind: fr.Kind, Seq: fr.Seq,
				Preview: preview(fr.Text, 200), Bytes: fr.Len,
			})
		}
		if a.ResponsesTotal == 0 {
			a.ResponsesState = "none"
			a.ResponsesNote = "这一步之后运行期没有回包（本来就没应答，还是没处理 —— 看回放日志）"
		} else {
			a.ResponsesState = "attributed"
		}
	}
}

// preview 把一段载荷压成一行预览（空白折叠 + 截断）。**只用于预览**：
// 全文一律走 /raw 取原始字节，预览不许被当成"这一步发了什么"的证据。
func preview(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}
