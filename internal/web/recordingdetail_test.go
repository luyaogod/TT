package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tt/internal/fep"
)

// 录制明细页的四个读侧端点（计划 C 的 D1）。
//
// 这里钉三类事：
//  1. **三层数字**（帧 / 事件 / 语义动作）不许混成一个 —— H2 实测 34 帧 / 63 事件 / 17 动作；
//  2. **没有帧索引时**照实降级为"块级时间"，而不是假装有、也不是报错；
//  3. `/raw` 的边界：越界与超限一律 400，但**允许**取一帧中间的一段（一帧可以 269 KB）。

// seedRawDir 在 dir 里写一个方向的 raw 文件（真帧流），并按给定时刻落帧索引 ——
// 走的就是代理那条路：一边喂块、一边落 `<raw>.frames.jsonl`。
func seedRawDir(t *testing.T, dir, name string, conn, dirCode int, frames [][]byte, times []int64) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := filepath.Join(dir, name)
	var blob []byte
	w, err := fep.NewFrameIndexWriter(fep.IndexPathFor(raw), conn, dirCode)
	if err != nil {
		t.Fatal(err)
	}
	for i, fr := range frames {
		blob = append(blob, fr...)
		at := time.UnixMilli(0)
		if i < len(times) && times[i] > 0 {
			at = time.UnixMilli(times[i])
		}
		w.Write(fr, at)
	}
	if err := os.WriteFile(raw, blob, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

// auiFrame 造一条界面侧消息（AUI 帧）。
func auiFrame(body string) []byte {
	return fep.Encode(fep.Frame{Type: fep.FrameTypeAUI, Payload: []byte(body)})
}

// cacheFrame 造一条图片缓存帧（type=5）—— 它不是消息，但占字节、有时刻。
func cacheFrame(body string) []byte {
	return fep.Encode(fep.Frame{Type: fep.FrameTypeCache, Payload: []byte(body)})
}

// seedDetailRecording 造一份"像真录制"的录制：一条连接、两个方向、帧索引带时刻。
//
// 界面侧（rt2gdc）声明了 #42 这个 Edit 与它的标签；事件侧（gdc2rt）有一条点击
// （ActionEvent #42）、一条输入（ConfigureEvent #42 value）、外加被语义层丢掉的 KeyEvent。
func seedDetailRecording(t *testing.T, s *Server, cfgPath string) string {
	t.Helper()
	dir := filepath.Join(filepath.Dir(cfgPath), "fep")
	const id = "fep-detail"
	// 界面侧：先声明窗口与字段，再推一条带值的更新（名字要能解出来才有意义）
	seedRawDir(t, dir, id+"-1-rt2gdc.log", 1, fep.DirRuntime, [][]byte{
		auiFrame(`om 1 {{an 0 Window 100 {{name "w_main"} {text "录入"}}}}`),
		auiFrame(`om 2 {{an 100 Label 101 {{comment "cmt_a"} {text "审核"}}}}`),
		auiFrame(`om 3 {{an 100 Edit 42 {{comment "cmt_a"}}}}`),
		cacheFrame(`!!__cached__!!`),
		auiFrame(`om 4 {{un 42 {{value "X"}}}}`),
	}, []int64{1000, 1100, 1200, 1300, 5000})
	// 事件侧：两条事件帧（3 个原始事件 → 2 个语义动作，1 个被丢）
	seedRawDir(t, dir, id+"-1-gdc2rt.log", 1, fep.DirFront, [][]byte{
		auiFrame(`event _om 1{}{{ActionEvent 0{{idRef "42"}}}{{KeyEvent 0{{keyName "Tab"}}}}}`),
		auiFrame(`event _om 2{}{{ConfigureEvent 0{{idRef "42"}}{{value "LUYAOWF"}}}}`),
	}, []int64{2000, 3000})
	return id
}

func TestRecordingOverviewTellsTheThreeLayers(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	id := seedDetailRecording(t, s, cfgPath)

	code, out := getJSON(t, s, "/api/recordings/"+id+"/overview")
	if code != http.StatusOK {
		t.Fatalf("概览该成功：%d %v", code, out)
	}
	if out["timePrecision"] != "frame" {
		t.Fatalf("带了帧索引就该说 frame：%v", out["timePrecision"])
	}
	counts, _ := out["counts"].(map[string]any)
	num := func(k string) int { v, _ := counts[k].(float64); return int(v) }
	if num("actions") != 2 {
		t.Errorf("语义动作该是 2（一次点击 + 一次输入）：%v", counts)
	}
	if num("events") != 3 {
		t.Errorf("原始事件该是 3（含被丢掉的 KeyEvent）：%v", counts)
	}
	if num("eventFrames") != 2 {
		t.Errorf("事件帧该是 2（一帧可以带两个事件）：%v", counts)
	}
	if num("cache") != 1 {
		t.Errorf("图片缓存帧该单独计数：%v", counts)
	}
	if num("frames") != 7 {
		t.Errorf("帧总数该是 5 + 2 = 7：%v", counts)
	}
	dirs, _ := out["directions"].([]any)
	if len(dirs) != 2 {
		t.Fatalf("两个方向都要报出来：%v", out["directions"])
	}
	for _, x := range dirs {
		d, _ := x.(map[string]any)
		idx, _ := d["index"].(map[string]any)
		if idx["state"] != "complete" {
			t.Errorf("方向 %v 的索引该是完整的：%v", d["dirLabel"], idx)
		}
	}
	// 条带用的帧数组：按时刻排好，缓存帧也在里面（它有自己的时刻）
	items, _ := out["items"].([]any)
	if len(items) != 7 {
		t.Fatalf("条带该拿到全部 7 帧：%d", len(items))
	}
	first, _ := items[0].(map[string]any)
	if got, _ := first["t"].(float64); int64(got) != 1000 {
		t.Errorf("第一帧该是最早的那一帧（t=1000）：%v", first)
	}
	sawCache := false
	for _, x := range items {
		m, _ := x.(map[string]any)
		if m["k"] == "cache" {
			sawCache = true
			if got, _ := m["t"].(float64); int64(got) != 1300 {
				t.Errorf("缓存帧要带上它的时刻（否则「缓存」那条泳道没法画在时间轴上）：%v", m)
			}
		}
	}
	if !sawCache {
		t.Error("图片缓存帧也要进 items（kind=cache），否则那条泳道没有时间")
	}
}

func TestRecordingStepsResolveNamesAndResponses(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	id := seedDetailRecording(t, s, cfgPath)

	code, out := getJSON(t, s, "/api/recordings/"+id+"/steps")
	if code != http.StatusOK {
		t.Fatalf("步骤该成功：%d %v", code, out)
	}
	acts, _ := out["actions"].([]any)
	if len(acts) != 2 {
		t.Fatalf("该有 2 个语义动作：%v", out["actions"])
	}
	a0, _ := acts[0].(map[string]any)
	if a0["kind"] != "click" || int(a0["id"].(float64)) != 42 {
		t.Errorf("第一个动作该是点 #42：%v", a0)
	}
	tgt, _ := a0["target"].(map[string]any)
	// 名字来自 R1（comment-key → Label 的 text）
	if tgt["name"] != "审核" {
		t.Errorf("名字该从 comment 键解出「审核」：%v", tgt)
	}
	if tgt["confidence"] != "high" {
		t.Errorf("置信度该是 high：%v", tgt)
	}
	if tgt["type"] != "Edit" || int(tgt["id"].(float64)) != 42 {
		t.Errorf("类型与 id 要照实给：%v", tgt)
	}
	// 行 key 用 连接:帧偏移（步号不是身份）
	fr, _ := a0["frame"].(map[string]any)
	if fr == nil || int(fr["c"].(float64)) != 1 {
		t.Errorf("动作要带出自哪一帧：%v", a0["frame"])
	}
	if _, ok := fr["o"]; !ok {
		t.Error("帧要带字节偏移（行 key 与检查器取原字节都靠它）")
	}
	// 丢掉的原始事件数要照实说（这里是 KeyEvent）
	if int(a0["dropped"].(float64)) != 1 {
		t.Errorf("这一帧里被语义层丢掉 1 个事件，要照实报：%v", a0["dropped"])
	}
	// 回应归属：动作在 t=2000，窗口到下一个动作（t=3000）为止 —— 界面侧 t=5000 那条
	// 属于**最后一个**动作（输入），这也是这条规则要表达的意思。
	if resp, _ := a0["responses"].([]any); len(resp) != 0 {
		t.Errorf("第一个动作的窗口里不该有回应（t=5000 属于下一个动作）：%v", resp)
	}
	a1, _ := acts[1].(map[string]any)
	if a1["kind"] != "input" || a1["value"] != "LUYAOWF" {
		t.Fatalf("第二个动作该是输入：%v", a1)
	}
	resp, _ := a1["responses"].([]any)
	if len(resp) != 1 {
		t.Fatalf("最后一个动作该归属 1 条回应（t=5000 的 un）：%v note=%v", resp, a1["responsesNote"])
	}
	r0, _ := resp[0].(map[string]any)
	// 运行期的顶层消息几乎都是 `om`（`un`/`rn` 是它**里面**的节点类型）——
	// 所以 kind=om，而"改的是哪个对象"看载荷文本，不从 om 里猜 id。
	if int64(r0["t"].(float64)) != 5000 || r0["k"] != "om" {
		t.Errorf("回应该是 t=5000 的那条 om：%v", r0)
	}
	if pv, _ := r0["preview"].(string); !strings.Contains(pv, `un 42`) {
		t.Errorf("预览里该看得见它改的是 #42：%q", pv)
	}
	if int(r0["bytes"].(float64)) != 38 {
		t.Errorf("回应的字节数要照实给（它是「这一步被确认」的量度）：%v", r0["bytes"])
	}
}

// 老件没有帧索引：时间精度降级为块级、每一步都说清"归属不了"，**不报错**。
func TestRecordingStepsWithoutIndexDegradeToBlockTime(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	dir := filepath.Join(filepath.Dir(cfgPath), "fep")
	const id = "fep-legacy"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// 手写：raw 帧流 + timeline.log（**没有**帧索引）
	blob := append(auiFrame(`om 1 {{an 0 Window 100 {{name "w"} {text "录入"}}}}`),
		auiFrame(`om 2 {{an 100 Edit 42 {{text "审核"}}}}`)...)
	if err := os.WriteFile(filepath.Join(dir, id+"-1-rt2gdc.log"), blob, 0o644); err != nil {
		t.Fatal(err)
	}
	ev := auiFrame(`event _om 1{}{{ActionEvent 0{{idRef "42"}}}}`)
	if err := os.WriteFile(filepath.Join(dir, id+"-1-gdc2rt.log"), ev, 0o644); err != nil {
		t.Fatal(err)
	}
	tl := "11:24:13.221 >> om 1 {{…}}\n11:24:14.500 << event _om 1{}{{…}}\n"
	if err := os.WriteFile(filepath.Join(dir, id+"-1-timeline.log"), []byte(tl), 0o644); err != nil {
		t.Fatal(err)
	}

	code, out := getJSON(t, s, "/api/recordings/"+id+"/overview")
	if code != http.StatusOK {
		t.Fatalf("没有索引也要能打开（降级，不是报错）：%d %v", code, out)
	}
	if out["timePrecision"] != "block" {
		t.Fatalf("没有帧索引就该说块级：%v", out["timePrecision"])
	}
	if blocks, _ := out["blocks"].([]any); len(blocks) != 2 {
		t.Fatalf("块级刻度该来自 timeline.log（2 块）：%v", out["blocks"])
	}
	notes, _ := out["notes"].([]any)
	joined := fmt.Sprint(notes...)
	if !strings.Contains(joined, "块级") {
		t.Errorf("降级原因要说出来（notes）：%v", notes)
	}
	items, _ := out["items"].([]any)
	for _, x := range items {
		m, _ := x.(map[string]any)
		if tv, ok := m["t"]; !ok || tv.(float64) != 0 {
			t.Errorf("没有帧级时间时 t 要**显式**给 0（缺省会让前端分不清「没有」与「是 0」）：%v", m)
		}
	}
	// 行上照实说"时间不知道"，并保留"为什么"
	_, steps := getJSON(t, s, "/api/recordings/"+id+"/steps")
	acts, _ := steps["actions"].([]any)
	if len(acts) != 1 {
		t.Fatalf("该有 1 个动作：%v", steps["actions"])
	}
	a0, _ := acts[0].(map[string]any)
	if note, _ := a0["responsesNote"].(string); !strings.Contains(note, "不猜") {
		t.Errorf("归属不了要说清：%v", a0["responsesNote"])
	}
}

// /raw 的边界：越界与超限 400；帧中间的一段**允许**取（一帧可以 269 KB）。
func TestRecordingRawBounds(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	id := seedDetailRecording(t, s, cfgPath)
	base := "/api/recordings/" + id + "/raw?conn=1&dir=1"

	if code, out := getJSON(t, s, base+"&off=0&len=100000"); code != http.StatusBadRequest {
		t.Fatalf("越界该 400：%d %v", code, out)
	}
	if code, out := getJSON(t, s, base+"&off=0&len=999999"); code != http.StatusBadRequest {
		t.Fatalf("超限该 400：%d %v", code, out)
	}
	if code, _ := getJSON(t, s, base+"&off=-1&len=10"); code != http.StatusBadRequest {
		t.Fatalf("负偏移该 400：%d", code)
	}
	if code, _ := getJSON(t, s, base+"&off=0"); code != http.StatusBadRequest {
		t.Fatalf("缺 len 该 400：%d", code)
	}
	// 方向要显式给：给了不存在的方向要 404（并说清这条连接有哪些方向）
	if code, out := getJSON(t, s, "/api/recordings/"+id+"/raw?conn=1&dir=7&off=0&len=10"); code != http.StatusNotFound {
		t.Fatalf("方向不存在该 404：%d %v", code, out)
	}
	if code, out := getJSON(t, s, "/api/recordings/"+id+"/raw?conn=1&off=0&len=10"); code != http.StatusNotFound {
		t.Fatalf("没给方向该 404（不猜哪一侧）：%d %v", code, out)
	}
	// 第一帧（off=0，是那一帧的整帧）
	code, out := getJSON(t, s, base+"&off=0&len="+itoa(len(auiFrame(`om 1 {{an 0 Window 100 {{name "w_main"} {text "录入"}}}}`))))
	if code != http.StatusOK {
		t.Fatalf("整帧该取得到：%d %v", code, out)
	}
	fr, _ := out["frame"].(map[string]any)
	if fr == nil || fr["whole"] != true || fr["kind"] != "om" {
		t.Fatalf("要报出「这一段是整帧」：%v", out["frame"])
	}
	if s, _ := out["hexFlat"].(string); !strings.HasPrefix(s, "00000") {
		t.Errorf("hexFlat 该以帧头开始：%q", s)
	}
	if s, _ := out["printable"].(string); !strings.Contains(s, "an 0 Window 100") {
		t.Errorf("可打印形态里该看见消息文本：%q", s)
	}
	// 帧中间的一段：允许，但要标出 whole=false
	code, out = getJSON(t, s, base+"&off=9&len=8")
	if code != http.StatusOK {
		t.Fatalf("帧中间的一段该允许取（否则大帧根本看不了）：%d %v", code, out)
	}
	if fr, _ := out["frame"].(map[string]any); fr == nil || fr["whole"] != false {
		t.Errorf("要标出这不是整帧：%v", out["frame"])
	}
}

// 全文流：分页按**页序号**（块级时间下 t 全是 0，用 t 分页会永远停在第一页）。
func TestRecordingTimelinePaging(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	id := seedDetailRecording(t, s, cfgPath)
	code, out := getJSON(t, s, "/api/recordings/"+id+"/timeline?limit=3")
	if code != http.StatusOK {
		t.Fatalf("全文流该成功：%d %v", code, out)
	}
	if int(out["total"].(float64)) != 7 {
		t.Fatalf("总数该是 7：%v", out["total"])
	}
	page, _ := out["frames"].([]any)
	if len(page) != 3 {
		t.Fatalf("该返回 3 条：%d", len(page))
	}
	code, out2 := getJSON(t, s, "/api/recordings/"+id+"/timeline?limit=3&from=3")
	if code != http.StatusOK {
		t.Fatal(code)
	}
	page2, _ := out2["frames"].([]any)
	if len(page2) != 3 {
		t.Fatalf("第二页该有 3 条：%d", len(page2))
	}
	p0, _ := page[0].(map[string]any)
	p2, _ := page2[0].(map[string]any)
	if p0["o"] == p2["o"] {
		t.Error("第二页不该与第一页重复")
	}
	// 只取事件侧
	_, out3 := getJSON(t, s, "/api/recordings/"+id+"/timeline?dir=0")
	if int(out3["total"].(float64)) != 2 {
		t.Errorf("事件侧该只有 2 帧：%v", out3["total"])
	}
}

// 找不到的 ID：404（不是 500，也不是空页）。
func TestRecordingDetailUnknownID(t *testing.T) {
	s, _ := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	for _, ep := range []string{"overview", "timeline", "steps"} {
		if code, _ := getJSON(t, s, "/api/recordings/没有这一份/"+ep); code != http.StatusNotFound {
			t.Errorf("%s 该 404：%d", ep, code)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/api/recordings/没有这一份/raw?conn=1&dir=1&off=0&len=8", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("raw 该 404：%d", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if msg, _ := body["error"].(string); !strings.Contains(msg, "找不到录制") {
		t.Errorf("要说清是找不到那一份：%v", body)
	}
}

// itoa 是测试里的小工具（避免为一个数字引入 strconv 的噪音）。
func itoa(n int) string { return fmt.Sprintf("%d", n) }
