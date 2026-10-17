package web

// 录制明细页的四个读侧端点（计划 C 的 D1 落地处）。读侧模型在 recordingread.go。

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"tt/internal/fep"
)

// registerRecordingDetail 挂上明细页的接口。
func (s *Server) registerRecordingDetail(m *http.ServeMux) {
	m.HandleFunc("GET /api/recordings/{id}/overview", s.hRecordingOverview)
	m.HandleFunc("GET /api/recordings/{id}/timeline", s.hRecordingTimeline)
	m.HandleFunc("GET /api/recordings/{id}/steps", s.hRecordingSteps)
	m.HandleFunc("GET /api/recordings/{id}/raw", s.hRecordingRaw)
}

// maxRawRead 是"一次最多取多少原始字节"。
//
// 为什么要有这个上限：一帧实测可以是 269 KB（首屏整棵树），把它一次塞进 JSON 会让
// 浏览器卡住、也让"取一帧看看"变成一次重活。超了就 400 —— 前端改成按段取。
const maxRawRead = 8 << 10

// dirInfo 是一个方向的元信息（概览条带与列表都要它）。
type dirInfo struct {
	Conn     int             `json:"conn"`
	Dir      int             `json:"dir"`
	DirLabel string          `json:"dirLabel"`
	Raw      string          `json:"raw"`
	Bytes    int64           `json:"bytes"`
	Frames   int             `json:"frames"`
	First    int64           `json:"first,omitempty"`
	Last     int64           `json:"last,omitempty"`
	Index    recIndexInfo    `json:"index"`
	Noise    []fep.ByteRange `json:"noise,omitempty"`
}

// recCounts 是概览里那几行数字。**分开列**，因为它们回答的是不同的问题：
//
//	eventFrames 34 —— "前端点了几次"的**帧**数（一个帧可以带好几个事件）
//	events      63 —— 原始事件数（含被语义层丢掉的导航/状态上报）
//	actions     17 —— 语义动作（= 用户眼里的"步"）
//
// 只报一个数就必然误导：H2 实测的 34/63/17 正是这三层。
type recCounts struct {
	Frames      int `json:"frames"`
	Messages    int `json:"messages"`
	Cache       int `json:"cache"`
	EventFrames int `json:"eventFrames"`
	Events      int `json:"events"`
	Actions     int `json:"actions"`
}

// recOverview 是 /overview 的响应体。
type recOverview struct {
	OK            bool       `json:"ok"`
	ID            string     `json:"id"`
	Where         string     `json:"where"`
	Dir           string     `json:"dir"`
	Prog          string     `json:"prog,omitempty"`
	Started       string     `json:"started,omitempty"`
	Finished      string     `json:"finished,omitempty"`
	TimePrecision string     `json:"timePrecision"`
	Counts        recCounts  `json:"counts"`
	Directions    []dirInfo  `json:"directions"`
	Items         []recFrame `json:"items"`
	Blocks        []recBlock `json:"blocks,omitempty"`
	Notes         []string   `json:"notes,omitempty"`
}

// hRecordingOverview GET /api/recordings/{id}/overview
//
// 条带的全部输入：各方向的字节/帧/索引状态 + 一份**按时间排好的**帧数组。
// 只读索引里的时刻，位置一律现场解（见 recordingread.go 的纪律 1）。
func (s *Server) hRecordingOverview(w http.ResponseWriter, r *http.Request) {
	d, err := s.loadRecording(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	ov := recOverview{
		OK: true, ID: d.ID, Where: d.Where, Dir: d.Dir,
		TimePrecision: d.Precision, Blocks: d.Blocks, Notes: d.Notes,
	}
	if d.Meta != nil {
		ov.Prog = d.Meta.Prog
		if !d.Meta.Started.IsZero() {
			ov.Started = d.Meta.Started.Format(time.RFC3339)
		}
		if !d.Meta.Finished.IsZero() {
			ov.Finished = d.Meta.Finished.Format(time.RFC3339)
		}
	}
	for _, rd := range d.Dirs {
		info := dirInfo{
			Conn: rd.Raw.Conn, Dir: rd.Raw.Dir, DirLabel: fep.DirLabel(rd.Raw.Dir),
			Raw: rd.Raw.Name, Bytes: rd.Bytes, Frames: len(rd.Frames),
			Index: rd.Index, Noise: rd.Noise,
		}
		for _, fr := range rd.Frames {
			if fr.T == 0 {
				continue
			}
			if info.First == 0 || fr.T < info.First {
				info.First = fr.T
			}
			if fr.T > info.Last {
				info.Last = fr.T
			}
		}
		ov.Directions = append(ov.Directions, info)
	}
	ov.Items = orderedFrames(d)
	ov.Counts = countsOf(d)
	writeJSON(w, http.StatusOK, ov)
}

// hRecordingTimeline GET /api/recordings/{id}/timeline?conn=&dir=&kind=&from=&limit=
//
// 全文流：一页一页地看"这条线上到底过了什么"。`from` 是**页序号**（不是时间）：
// 时间在块级精度下是 0，用它分页会永远停在第一页。
//
// `kind` 是**服务端**过滤（逗号分隔，例如 `kind=om,un,rn,event`）：分页也在这一侧，
// 让前端自己滤会把"这一页只剩 2 条"变成常态。
func (s *Server) hRecordingTimeline(w http.ResponseWriter, r *http.Request) {
	d, err := s.loadRecording(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	q := r.URL.Query()
	conn := atoiDefault(q.Get("conn"), 0)
	dir := atoiDefault(q.Get("dir"), -1)
	from := atoiDefault(q.Get("from"), 0)
	limit := atoiDefault(q.Get("limit"), 500)
	kinds := kindFilter(q.Get("kind"))
	if limit <= 0 {
		limit = 500
	}
	if limit > 5000 {
		limit = 5000
	}
	if from < 0 {
		from = 0
	}
	var all []recFrame
	for _, fr := range orderedFrames(d) {
		if conn != 0 && fr.Conn != conn {
			continue
		}
		if dir >= 0 && fr.Dir != dir {
			continue
		}
		if kinds != nil && !kinds[fr.Kind] {
			continue
		}
		all = append(all, fr)
	}
	type row struct {
		recFrame
		Preview string `json:"preview,omitempty"`
	}
	var page []row
	for i := from; i < len(all) && len(page) < limit; i++ {
		page = append(page, row{recFrame: all[i], Preview: preview(all[i].Text, 300)})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "id": d.ID, "timePrecision": d.Precision,
		"total": len(all), "from": from, "limit": limit, "frames": page,
		"note": pageNote(d.Precision),
	})
}

// kindFilter 解析 `kind=a,b,c`（空 = 不过滤，返回 nil）。
func kindFilter(s string) map[string]bool {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	out := map[string]bool{}
	for _, k := range strings.Split(s, ",") {
		if k = strings.TrimSpace(k); k != "" {
			out[k] = true
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// pageNote 说清"这个列表是按什么排的"。块级精度下顺序不是时间，必须说出来。
func pageNote(precision string) string {
	if precision == "frame" {
		return "按帧时刻排序（同一时刻内按连接与字节序）"
	}
	return "**没有帧级时间**：顺序按各方向文件内的字节序（不是时间序）"
}

// hRecordingSteps GET /api/recordings/{id}/steps?conn=
//
// 三层模型：语义动作（用户眼里的"步"）> 原始事件 > 运行期回应。
func (s *Server) hRecordingSteps(w http.ResponseWriter, r *http.Request) {
	d, err := s.loadRecording(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	conn := atoiDefault(r.URL.Query().Get("conn"), 0)
	// 三列「身份 vs 句柄」挂在每一行上（旧键一个不动，identity 是新键；见 identitycols.go）。
	// 列的顺序/文案/提示也由服务端给一处，界面照它渲染，不在两处各写一套。
	idents := identityRows(d)
	type actionRow struct {
		*recAction
		Identity identityCols `json:"identity"`
	}
	var acts []actionRow
	for i, a := range d.Actions {
		if conn != 0 && a.Conn != conn {
			continue
		}
		acts = append(acts, actionRow{recAction: a, Identity: idents[i]})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "id": d.ID, "timePrecision": d.Precision,
		"counts": countsOf(d), "actions": acts, "notes": d.Notes,
		"identityColumns": identityColumnSpec, "identityHint": identityHint,
	})
}

// hRecordingRaw GET /api/recordings/{id}/raw?conn=&dir=&off=&len=
//
// 检查器的"原始帧"页签：把录制件里那一段字节**原样**拿出来。
//
// 三条边界：
//   - 越界（off<0、off+len 超过文件、len<=0）→ 400；
//   - 单次 > 8 KB → 400（前端改按段取）。**允许**取一帧中间的一段：一帧实测可以
//     269 KB，"不许分段"等于这一帧在界面上根本看不了；
//   - 位置对不上任何一帧时**照常返回**，但明说"这一段不是某一帧的整帧"——
//     握手的明文与图片缓存的二进制也在文件里，它们同样值得能看。
func (s *Server) hRecordingRaw(w http.ResponseWriter, r *http.Request) {
	d, err := s.loadRecording(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	q := r.URL.Query()
	conn := atoiDefault(q.Get("conn"), 0)
	dir := atoiDefault(q.Get("dir"), -1)
	off, oerr := strconv.ParseInt(q.Get("off"), 10, 64)
	length, lerr := strconv.Atoi(q.Get("len"))
	if oerr != nil || lerr != nil || off < 0 || length <= 0 {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("off/len 要给全且合法（off >= 0、len > 0）"))
		return
	}
	if length > maxRawRead {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("单次最多取 %d 字节（要了 %d）：改成按段取", maxRawRead, length))
		return
	}
	rd := d.dirOf(conn, dir)
	if rd == nil {
		// 只给了 conn 没给 dir 时，把这条连接存在的方向说清楚，别让人猜。
		var have []string
		for _, x := range d.Dirs {
			if x.Raw.Conn == conn {
				have = append(have, fmt.Sprintf("%d(%s)", x.Raw.Dir, fep.DirLabel(x.Raw.Dir)))
			}
		}
		writeErr(w, http.StatusNotFound, fmt.Errorf("这份录制里没有连接 #%d 方向 %d 的录制件（该连接有：%s）",
			conn, dir, strings.Join(have, "、")))
		return
	}
	blob, rerr := os.ReadFile(rd.Raw.Path)
	if rerr != nil {
		writeErr(w, http.StatusInternalServerError, rerr)
		return
	}
	end := off + int64(length)
	if end > int64(len(blob)) {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("越界：%d+%d 超过文件大小 %d", off, length, len(blob)))
		return
	}
	chunk := blob[off:end]
	resp := map[string]any{
		"ok": true, "id": d.ID, "conn": conn, "dir": dir,
		"raw": rd.Raw.Name, "fileBytes": len(blob),
		"off": off, "len": length, "inFrame": false,
		"hex":       hex.Dump(chunk),
		"hexFlat":   strings.ToUpper(hex.EncodeToString(chunk)),
		"printable": printable(chunk),
	}
	if fr, exact, within := frameAt(rd.Frames, off, length); within {
		resp["inFrame"] = true
		resp["frame"] = map[string]any{
			"off": fr.Off, "len": fr.Len, "kind": fr.Kind, "seq": fr.Seq,
			"idRef": fr.IDRef, "t": fr.T, "whole": exact,
		}
	} else {
		// 不在任何一帧里 = 握手的明文（`meta Connection{{…}}` / `meta Client{{…}}`，
		// 整条流里唯一不带帧头的一段）或还没认出帧头来的字节。**照常返回**并说清它是什么：
		// 这两样同样值得能看，但绝不能让人以为"这就是某条消息"。
		resp["note"] = "这一段不在任何一帧里：它多半是握手的明文（不带帧头）或没认出帧头的字节"
	}
	writeJSON(w, http.StatusOK, resp)
}

// frameAt 回答"这一段落在哪一帧里"：整帧（exact）还是帧的一部分（within）。
func frameAt(frames []recFrame, off int64, length int) (recFrame, bool, bool) {
	for _, fr := range frames {
		if off < fr.Off || off+int64(length) > fr.Off+int64(fr.Len) {
			continue
		}
		return fr, off == fr.Off && length == fr.Len, true
	}
	return recFrame{}, false, false
}

// printable 把一段字节压成"能贴进文本里"的形态：可打印字符照留（UTF-8 中文照留，
// 方便直接 grep），控制字符转义 —— 9 字节帧头那种 `\x00` 满天飞的东西才读得下去。
func printable(b []byte) string {
	var sb strings.Builder
	sb.Grow(len(b) + len(b)/4)
	for _, c := range b {
		switch {
		case c == '\n':
			sb.WriteString("\\n")
		case c == '\r':
			sb.WriteString("\\r")
		case c == '\t':
			sb.WriteString("\\t")
		case c >= 0x20 && c < 0x7f:
			sb.WriteByte(c)
		case c >= 0x80:
			sb.WriteByte(c)
		default:
			sb.WriteString(fmt.Sprintf("\\x%02x", c))
		}
	}
	return sb.String()
}

// countsOf 统计三层各自的规模。
func countsOf(d *recRead) recCounts {
	c := recCounts{Frames: d.FrameN, Events: d.EventN, Actions: len(d.Actions)}
	for _, rd := range d.Dirs {
		for _, fr := range rd.Frames {
			switch fr.Kind {
			case "cache":
				c.Cache++
			case "event":
				c.EventFrames++
				c.Messages++
			default:
				c.Messages++
			}
		}
	}
	return c
}

// orderedFrames 把各方向的帧合起来排成一条序列。
//
// 排序规则（条带与文本流都靠它，所以只能有一处定义）：
//   - 有帧级时间：按 `t`，同一时刻按（连接，方向，字节偏移）；
//   - 没有帧级时间：按（连接，方向，字节偏移）—— **不插值、不假装**，
//     界面会明说"这是字节序，不是时间序"。
func orderedFrames(d *recRead) []recFrame {
	var all []recFrame
	for _, rd := range d.Dirs {
		all = append(all, rd.Frames...)
	}
	sort.SliceStable(all, func(i, j int) bool {
		a, b := all[i], all[j]
		if a.T != b.T {
			// t == 0 表示"不知道"：它们排在最后，而不是被当成 1970 年排到最前。
			if a.T == 0 {
				return false
			}
			if b.T == 0 {
				return true
			}
			return a.T < b.T
		}
		if a.Conn != b.Conn {
			return a.Conn < b.Conn
		}
		if a.Dir != b.Dir {
			return a.Dir < b.Dir
		}
		return a.Off < b.Off
	})
	return all
}

// atoiDefault 解析一个查询参数（空或非法时用默认值，不报错：这些参数是可选过滤）。
func atoiDefault(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
