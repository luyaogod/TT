package web

// 录制明细页的「身份 vs 句柄」三列 —— 契约 C3 与 A7/A9/A10/A11/A12/A15/A18 的**消费面**。
//
// 一行的三列各答一件事，顺序固定（服务端用 identityColumns 报出来，界面照它排）：
//
//	colName    身份·稳      节点的 Attr("colName")   —— 编译期分配，表单内稳定
//	fieldId    表单内编号·稳 节点的 Attr("fieldId")   —— 与 colName 成对出现
//	sessionId  句柄·会复用    事件里的 idRef          —— 单步有效，LIFO 复用，rn 后全量失效
//
// 这里钉四件事：
//  1. 三列都出现，且前两列**真的取自节点属性**（拿夹具原文对一遍，不是拿模型自己对自己）；
//  2. 节点上没有 colName 时显示"未解析" —— **不能**给一个空串冒充已解析（A15/A18：取不到就是不认识，
//     启发式不得参与判定，例如自己拼 `w_`+后缀 那种）；
//  3. 旧键一个都不许丢（既有前端不炸）；
//  4. 同一个会话 id 在 rn 之后被别的节点复用时：号码列不变，身份列跟着**当时那个节点**走。
//
// 夹具用仓库里现成的现场录制增量子流（testdata/fep-recordings/*.txt）：读它们走既有的
// fep.ReadRecording，落成一份录制走既有的 seedRawDir —— 测试里不另写解析器。

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tt/internal/fep"
	"tt/internal/testkit"
)

// fixturePath 是仓库夹具的绝对路径。
func fixturePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(testkit.RepoRoot(t), "testdata", "fep-recordings", name)
}

// fixtureMessages 把一份夹具读成消息序列（走既有的读侧入口）。
func fixtureMessages(t *testing.T, name string) []string {
	t.Helper()
	msgs, err := fep.ReadRecording(fixturePath(t, name))
	if err != nil {
		t.Fatalf("夹具读不出来（%s）: %v", name, err)
	}
	if len(msgs) < 5 {
		t.Fatalf("夹具 %s 只有 %d 条消息，这份用例要的增量流不在里面", name, len(msgs))
	}
	return msgs
}

// indexOfMessage 找"第一条包含 marker 的消息"的下标。
//
// 为什么不写死行号：夹具是现场录制剪出来的，加/换夹具时行号会变，而"在哪条消息之后"
// 才是用例真正依赖的东西。
func indexOfMessage(t *testing.T, msgs []string, marker string) int {
	t.Helper()
	for i, m := range msgs {
		if strings.Contains(m, marker) {
			return i
		}
	}
	t.Fatalf("夹具里没有一条消息包含 %q —— 夹具换了，用例的落点要重挑", marker)
	return -1
}

// fixtureProbe 是"在某条界面消息之后，让事件侧点一下这个 id"。
//
// After 空 = 全部消息之后（于是"这一步发生时"就是整棵树都到齐了）。
type fixtureProbe struct {
	After string
	ID    int
}

// seedFixtureRecording 把夹具消息当成界面侧（rt2gdc）录制件写进一份录制，
// 再按 probes 在事件侧（gdc2rt）放上点击。
//
// 时刻安排（这是"身份取自当时那个节点"能成立的前提）：
// 第 i 条界面消息 t = 1000*(i+1)；"在第 at 条之后"的那一步 t = 1000*(at+1)+500 ——
// 落在两条消息之间，于是帧级时间足以判出"这一步发生时界面到哪儿了"。
func seedFixtureRecording(t *testing.T, cfgPath, fixture string, probes []fixtureProbe) string {
	t.Helper()
	msgs := fixtureMessages(t, fixture)
	dir := filepath.Join(filepath.Dir(cfgPath), "fep")
	id := "fep-ident-" + strings.TrimSuffix(fixture, ".txt")
	const conn = 2

	rtFrames := make([][]byte, 0, len(msgs))
	rtTimes := make([]int64, 0, len(msgs))
	for i, m := range msgs {
		rtFrames = append(rtFrames, auiFrame(m))
		rtTimes = append(rtTimes, int64(1000*(i+1)))
	}
	var evFrames [][]byte
	var evTimes []int64
	for j, p := range probes {
		at := len(msgs) - 1
		if p.After != "" {
			at = indexOfMessage(t, msgs, p.After)
		}
		evFrames = append(evFrames,
			auiFrame(fmt.Sprintf(`event _om %d{}{{ActionEvent 0{{idRef "%d"}}}}`, j+1, p.ID)))
		evTimes = append(evTimes, int64(1000*(at+1))+int64(500+j*10))
	}
	seedRawDir(t, dir, id+"-2-rt2gdc.log", conn, fep.DirRuntime, rtFrames, rtTimes)
	seedRawDir(t, dir, id+"-2-gdc2rt.log", conn, fep.DirFront, evFrames, evTimes)
	return id
}

// fixtureModel 把夹具整条流灌进模型（就是录制末尾那棵树）。
func fixtureModel(t *testing.T, name string) *fep.Model {
	t.Helper()
	m := fep.NewModel()
	for _, s := range fixtureMessages(t, name) {
		m.ApplyMessage(s)
	}
	return m
}

// stepsOf 取明细页三层模型里的一层（顺手断言状态码）。
func stepsOf(t *testing.T, s *Server, id string) (map[string]any, []map[string]any) {
	t.Helper()
	code, out := getJSON(t, s, "/api/recordings/"+id+"/steps")
	if code != 200 {
		t.Fatalf("步骤端点该成功：%d %v", code, out)
	}
	raw, _ := out["actions"].([]any)
	acts := make([]map[string]any, 0, len(raw))
	for _, x := range raw {
		m, _ := x.(map[string]any)
		acts = append(acts, m)
	}
	return out, acts
}

// identityOf 取一行的三列（缺了就是用例要抓的错）。
func identityOf(t *testing.T, act map[string]any) map[string]any {
	t.Helper()
	v, ok := act["identity"]
	if !ok {
		t.Fatalf("这一行没有 identity（三列没出来）：%v", act)
	}
	m, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("identity 该是一个对象：%#v", v)
	}
	return m
}

// ① 三列都出现；colName/fieldId 取自节点属性（与夹具原文对得上）；顺序与提示写在响应上。
func TestIdentityColumnsComeFromNodeAttrs(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	const fixture = "openwin-with-dialogs.txt"

	// 末尾那棵树里挑一个**两把钥匙都在**的节点，让事件点它。
	//
	// 条件里要连 fieldId 一起对：同一个裸 id 会在不同作用域里各有一份
	// （实测 #172 既做过 `index#3` 又做过 `index#6`），只比 colName 会挑到"不是 Get 那一个"。
	m := fixtureModel(t, fixture)
	var want *fep.Object
	for _, o := range m.Objects() {
		if o.Attr("colName") == "" || o.Attr("fieldId") == "" {
			continue
		}
		if g, ok := m.Get(o.ID); ok && g.Type == o.Type &&
			g.Attr("colName") == o.Attr("colName") && g.Attr("fieldId") == o.Attr("fieldId") {
			want = o
			break
		}
	}
	if want == nil {
		t.Fatalf("夹具 %s 里找不到带 (colName, fieldId) 的节点 —— 用例的前提不在了", fixture)
	}
	id := seedFixtureRecording(t, cfgPath, fixture, []fixtureProbe{{ID: want.ID}})

	out, acts := stepsOf(t, s, id)
	if len(acts) != 1 {
		t.Fatalf("该有 1 个语义动作：%v", acts)
	}
	cols, _ := out["identityColumns"].([]any)
	if len(cols) != 3 {
		t.Fatalf("三列都要报出来（顺序由服务端定）：%v", out["identityColumns"])
	}
	wantKeys := []string{"colName", "fieldId", "sessionId"}
	wantStable := []bool{true, true, false}
	for i, x := range cols {
		c, _ := x.(map[string]any)
		if c["key"] != wantKeys[i] {
			t.Errorf("第 %d 列该是 %s（身份在前、句柄在后）：%v", i+1, wantKeys[i], c)
		}
		if st, _ := c["stable"].(bool); st != wantStable[i] {
			t.Errorf("%s 的 stable 该是 %v：%v", c["key"], wantStable[i], c)
		}
		if c["label"] == "" {
			t.Errorf("%s 要有列名（UI 直接显示它）：%v", c["key"], c)
		}
	}
	// 身份两列缺失时的文案由服务端给（界面不许自己编一个）
	for _, x := range cols[:2] {
		c, _ := x.(map[string]any)
		if c["missing"] != "未解析" {
			t.Errorf("%s 的缺失文案该是「未解析」：%v", c["key"], c)
		}
	}
	hint, _ := out["identityHint"].(string)
	if !strings.Contains(hint, "复用") || !strings.Contains(hint, "句柄") {
		t.Errorf("要有一句提示说清「号码会复用、身份才稳」：%q", hint)
	}

	ident := identityOf(t, acts[0])
	if ident["colName"] != want.Attr("colName") || ident["fieldId"] != want.Attr("fieldId") {
		t.Fatalf("身份两列该取自节点属性：得到 %v，节点上是 %q/%q",
			ident, want.Attr("colName"), want.Attr("fieldId"))
	}
	if ident["colNameKnown"] != true || ident["fieldIdKnown"] != true {
		t.Errorf("取到了就该说取到了：%v", ident)
	}
	if got, _ := ident["sessionId"].(float64); int(got) != want.ID {
		t.Errorf("第三列是会话 id（这件事里的 idRef）：%v", ident)
	}
	// 非自证：这两个值在夹具**原文**里就是这么写的
	blob, err := os.ReadFile(fixturePath(t, fixture))
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{
		{"colName", want.Attr("colName")},
		{"fieldId", want.Attr("fieldId")},
	} {
		needle := "{" + pair[0] + " \"" + pair[1] + "\"}"
		if !strings.Contains(string(blob), needle) {
			t.Errorf("夹具原文里没有 %s —— 说明这个值不是从节点属性来的", needle)
		}
	}
}

// ② 节点上没有 colName：显示"未解析"，**不是**空串冒充已解析。
func TestIdentityUnresolvedWhenNodeHasNoColName(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	const fixture = "openwin-with-dialogs.txt"

	// 挑一个"是真控件、但不是字段"的节点（没有 colName）。优先挑动作/工具栏/树项：
	// 它们正是运行期大量下发、又最容易被误当成字段的那一类。
	m := fixtureModel(t, fixture)
	prefer := map[string]bool{"Action": true, "ToolBarItem": true, "TreeItem": true, "Item": true, "Dialog": true}
	var want *fep.Object
	for _, o := range m.Objects() {
		if o.Attr("colName") != "" || !prefer[o.Type] {
			continue
		}
		if g, ok := m.Get(o.ID); ok && g.Attr("colName") == "" && g.Type == o.Type {
			want = o
			break
		}
	}
	if want == nil {
		t.Fatalf("夹具 %s 里找不到一个「没有 colName 的真控件」", fixture)
	}
	id := seedFixtureRecording(t, cfgPath, fixture, []fixtureProbe{{ID: want.ID}})

	_, acts := stepsOf(t, s, id)
	if len(acts) != 1 {
		t.Fatalf("该有 1 个动作：%v", acts)
	}
	ident := identityOf(t, acts[0])
	if ident["colName"] != "未解析" {
		t.Fatalf("取不到 colName 就要写「未解析」（空串等于假装解出来了）：%v", ident)
	}
	if ident["fieldId"] != "未解析" {
		t.Fatalf("fieldId 同样：%v", ident)
	}
	if ident["colNameKnown"] != false || ident["fieldIdKnown"] != false {
		t.Errorf("没解出来就得说没解出来（界面据此选样式，不许去匹配文案）：%v", ident)
	}
	if note, _ := ident["note"].(string); note == "" {
		t.Error("「未解析」要给一句为什么（节点上没有这个属性），不能留空")
	}
	// 号码仍然照实给：它是句柄，不是身份
	if got, _ := ident["sessionId"].(float64); int(got) != want.ID {
		t.Errorf("会话 id 照样要照实给：%v", ident)
	}
}

// ③ 旧键一个都不许丢（既有前端不炸）。
func TestIdentityKeepsLegacyKeys(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	id := seedDetailRecording(t, s, cfgPath) // 既有的合成录制（不依赖大夹具）

	out, acts := stepsOf(t, s, id)
	for _, k := range []string{"ok", "id", "timePrecision", "counts", "actions", "notes"} {
		if _, ok := out[k]; !ok {
			t.Errorf("顶层旧键 %q 不该丢：%v", k, out)
		}
	}
	if len(acts) == 0 {
		t.Fatal("这条录制该有动作")
	}
	for _, k := range []string{"n", "kind", "id", "conn", "frame", "target", "events", "dropped", "responsesState"} {
		if _, ok := acts[0][k]; !ok {
			t.Errorf("行上的旧键 %q 不该丢：%v", k, acts[0])
		}
	}
	fr, _ := acts[0]["frame"].(map[string]any)
	if fr == nil {
		t.Fatal("frame 该仍是对象（行 key 与检查器都靠它）")
	}
	for _, k := range []string{"o", "n", "t", "c", "d", "k", "q"} {
		if _, ok := fr[k]; !ok {
			t.Errorf("frame 上的旧键 %q 不该丢：%v", k, fr)
		}
	}
	tgt, _ := acts[0]["target"].(map[string]any)
	if tgt == nil || tgt["name"] != "审核" {
		t.Fatalf("target（名字/置信度/父链）不该被动过：%v", acts[0]["target"])
	}
	// 新旧共存：同一行里既有 target 又有 identity
	if _, ok := acts[0]["identity"]; !ok {
		t.Fatal("新键 identity 该与旧键并存")
	}
}

// ④ 同一个 id 在 rn 之后被复用：号码列不变，身份列跟着**当时那个节点**走。
//
// id-reuse.txt 里的现成场景（现场录制剪出来的）：
//
//	om 6  … {rn 3541} …                                   ← 先把 3541 删掉
//	om 7  … {ToolBarItem 3541 {{name "selectpageall"} …}   ← 3541 的新一生：工具栏项，**没有** colName
//	om 12 … {{TableColumn 3541 {{name "formonly.check"} {colName "check"} {fieldId "0"} …  ← 又是 3541，这回是字段
//
// 所以同一条 id 上前后两步：第一步的身份必须还是"未解析"（那时它是工具栏项），
// 第二步才是 check#0。若身份取自录制末尾那棵树，第一步会被写成 check#0 —— 那就是在撒谎。
func TestIdentityFollowsCurrentNodeAcrossIDReuse(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	const fixture = "id-reuse.txt"
	const reuseID = 3541 // 就是上面那三行里的 3541

	id := seedFixtureRecording(t, cfgPath, fixture, []fixtureProbe{
		{After: "ToolBarItem 3541 {", ID: reuseID},
		{After: "TableColumn 3541 {", ID: reuseID},
	})
	_, acts := stepsOf(t, s, id)
	if len(acts) != 2 {
		t.Fatalf("该有 2 个动作（同一个 id 上的两次点击）：%v", acts)
	}
	early, late := identityOf(t, acts[0]), identityOf(t, acts[1])

	// 号码列：一模一样（这就是"句柄会复用"的样子）
	for _, x := range []map[string]any{early, late} {
		if got, _ := x["sessionId"].(float64); int(got) != reuseID {
			t.Fatalf("两次点击的号码该都是 %d：%v", reuseID, x)
		}
	}
	if got, _ := acts[0]["id"].(float64); int(got) != reuseID {
		t.Fatalf("动作自己带的 id 也该是 %d：%v", reuseID, acts[0]["id"])
	}

	// 身份列：早期那一步是工具栏项 → 没有身份（未解析），后面的才是这个字段
	if early["colName"] != "未解析" || early["colNameKnown"] != false {
		t.Errorf("rn 之后 3541 的新一生是 ToolBarItem（没有 colName）：身份该是「未解析」，"+
			"拿末尾那棵树回填就是撒谎：%v", early)
	}
	if early["nodeType"] != "ToolBarItem" {
		t.Errorf("顺便要能看出「当时它是什么类型」：%v", early)
	}
	if late["colName"] != "check" || late["fieldId"] != "0" {
		t.Errorf("第二次声明之后，3541 是 formonly.check（colName=check / fieldId=0）：%v", late)
	}
	if late["nodeType"] != "TableColumn" {
		t.Errorf("第二次的身份该来自新节点：%v", late)
	}
}
