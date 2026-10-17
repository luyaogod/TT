package fep

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"tt/internal/fep"
)

// E2 接线：`tt fep replay` / `tt recording replay` 要把**这份录制自己的**界面模型
// 递进 `fep.PlanReplay`，否则身份链在产品路径上等于没生效（库层做完了却摸不到）。
//
// 三条纪律（架构师给定的约束）都落在这一个文件里：
//
//	1. 给了模型 ⇒ `ReplayOptions.Model` 非 nil，且其它字段一个不变；
//	2. **读不到模型 ⇒ nil** ⇒ 旧行为（不记身份、按 idRef 照旧发）一字不变，绝不报错；
//	3. 端到端：一条带 `ColName` 的步骤走产品那条组装路径发出去，**落在身份解析后的 id 上**
//	   （对标库层 `TestReplaySendsToTheIdentityResolvedTarget`）。
//
// 全部离线：录制件是临时目录里**编出来的帧**，注入口是本机回环上的假服务（不连真机、
// 不开会话、不发任何真实事件）。

// recFixture 在临时目录里编一份"录制"，返回它的 recTarget 与目录。
//
// 形态照真录制：`<ID>-<连接>-rt2gdc.log`（界面侧）+ `-gdc2rt.log`（事件侧），都是帧流
// （`fep.Encode`）。界面侧两帧**同一个 (colName, fieldId) 两个不同裸 id** —— 正是 H1/H2
// 的现场：录制时是 6000，重建后是 6008。
func recFixture(t *testing.T) (recTarget, string) {
	t.Helper()
	dir := t.TempDir()
	aui := filepath.Join(dir, "t-1-rt2gdc.log")
	ev := filepath.Join(dir, "t-1-gdc2rt.log")

	writeFrames(t, aui,
		`om 1 {{an 0 Window 4097 {{name "w_axmt500"}} {{`+
			`{an 4097 FormField 6000 {{name "formonly.b_target"}{colName "b_target"}{fieldId "5"}{text "目标字段"}} {}} `+
			`}}}}`,
		`om 2 {{an 0 Window 4097 {{name "w_axmt500"}} {{`+
			`{an 4097 FormField 6008 {{name "formonly.b_target"}{colName "b_target"}{fieldId "5"}{text "目标字段（重建后）"}} {}} `+
			`}}}}`,
	)
	writeFrames(t, ev, `event _om 1{}{{ConfigureEvent 0{{idRef "6000"}{ value "X"}}}}`)

	return recTarget{ID: "t-1", Dir: dir, Files: []string{aui}}, dir
}

// writeFrames 把若干条消息按真帧格式落盘（帧头 + type + 载荷 + 换行）。
func writeFrames(t *testing.T, path string, msgs ...string) {
	t.Helper()
	var b []byte
	for _, msg := range msgs {
		b = append(b, fep.Encode(fep.Frame{Type: fep.FrameTypeAUI, Payload: []byte(msg)})...)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("写夹具 %s: %v", filepath.Base(path), err)
	}
}

// TestReplayOptionsForWiresTheModel：纯接缝 —— 给模型就带上，不给就是 nil，其余字段不变。
func TestReplayOptionsForWiresTheModel(t *testing.T) {
	none := replayOptionsFor(3, []string{"5170", "tabIndex"}, nil)
	if none.Model != nil {
		t.Fatalf("没给模型却带上了：%+v", none.Model)
	}
	if none.StopAt != 3 {
		t.Fatalf("StopAt = %d，想要 3（接线不许动其它字段）", none.StopAt)
	}
	if len(none.Skip) != 2 || none.Skip[0] != "5170" || none.Skip[1] != "tabIndex" {
		t.Fatalf("Skip 被改了：%+v", none.Skip)
	}

	m := fep.NewModel()
	with := replayOptionsFor(0, nil, m)
	if with.Model != m {
		t.Fatalf("给了模型却没带上：%+v", with.Model)
	}
	if with.StopAt != 0 || len(with.Skip) != 0 {
		t.Fatalf("接线凭空改了别的字段：%+v", with)
	}
}

// TestReplayIdentityModelReadsTheRecordingsAUISide：读的是**录制自己的**界面侧。
func TestReplayIdentityModelReadsTheRecordingsAUISide(t *testing.T) {
	tgt, _ := recFixture(t)
	m := replayIdentityModel(tgt)
	if m == nil {
		t.Fatal("录制里有界面侧，却读不出模型")
	}
	if idn := fep.IdentityOf(m, 6000); idn.ColName != "b_target" || idn.FieldID != 5 {
		t.Fatalf("模型里的身份 = (%s, %d)，想要 (b_target, 5)", idn.ColName, idn.FieldID)
	}
	// 第二帧（6008）也必须在：同一身份的两帧都要进模型，C3.5 的"取最后声明"才有对象可取。
	if _, ok := m.Get(6008); !ok {
		t.Fatalf("后一帧没进模型：len=%d", m.Len())
	}
}

// TestReplayIdentityModelIsNilWithoutAUISide：读不到 ⇒ **nil**（不是空模型、不是错误）。
//
// 这一条就是"旧行为不变"的闸门：`ReplayOptions.Model == nil` ⇒ 身份链自动不参与。
func TestReplayIdentityModelIsNilWithoutAUISide(t *testing.T) {
	if m := replayIdentityModel(recTarget{}); m != nil {
		t.Fatalf("没有文件却给出了模型：len=%d", m.Len())
	}
	// 只有事件侧：不许拿它充数
	dir := t.TempDir()
	ev := filepath.Join(dir, "t-1-gdc2rt.log")
	writeFrames(t, ev, `event _om 1{}{{ConfigureEvent 0{{idRef "6000"}{ value "X"}}}}`)
	if m := replayIdentityModel(recTarget{Files: []string{ev}}); m != nil {
		t.Fatalf("只有事件侧也给出了模型：len=%d", m.Len())
	}
	// 文件读不动 / 不是消息：同样 nil
	bad := filepath.Join(dir, "t-2-rt2gdc.log")
	if err := os.WriteFile(bad, []byte("not a frame at all"), 0o644); err != nil {
		t.Fatal(err)
	}
	if m := replayIdentityModel(recTarget{Files: []string{bad}}); m != nil {
		t.Fatalf("读不出消息却给出了模型：len=%d", m.Len())
	}
}

// TestReplayPlanKeepsLegacyShapeWithoutModel：**旧行为不变**的证据 ——
// 读不到模型时（老件 / 只有事件侧 / 没有 AUI），计划里没有身份、事件体一字不改，
// 与接线前完全一致；接上模型只**多记**身份，事件取舍一条都不变。
func TestReplayPlanKeepsLegacyShapeWithoutModel(t *testing.T) {
	tgt, _ := recFixture(t)
	events, err := tgt.events()
	if err != nil {
		t.Fatalf("读事件侧: %v", err)
	}
	// ① 接线前的构造（Model 为 nil）就是"旧行为"本身
	legacy := fep.PlanReplay(events, replayOptionsFor(replayStopAt, replaySkip, nil))
	if len(legacy.Events) != 1 {
		t.Fatalf("旧路径的计划条数 = %d，想要 1", len(legacy.Events))
	}
	if legacy.Events[0].ColName != "" || legacy.Events[0].FieldID != 0 {
		t.Fatalf("没有模型却记了身份：%+v", legacy.Events[0])
	}
	if legacy.Events[0].Event != `ConfigureEvent 0{{idRef "6000"}{ value "X"}}` {
		t.Fatalf("事件体被动了：%q", legacy.Events[0].Event)
	}
	if legacy.Scanned != 1 || len(legacy.Skips) != 0 {
		t.Fatalf("扫描/跳过计数变了：%+v", legacy)
	}

	// ② 产品路径在**读不到模型**的录制上（只有事件侧、没有界面侧）必须走回①的同一形状
	dir := t.TempDir()
	evOnly := filepath.Join(dir, "t-9-gdc2rt.log")
	writeFrames(t, evOnly, `event _om 1{}{{ConfigureEvent 0{{idRef "6000"}{ value "X"}}}}`)
	broken := planReplayFor(recTarget{ID: "t-9", Dir: dir, Files: []string{evOnly}}, events, replayStopAt, replaySkip)
	if len(broken.Events) != len(legacy.Events) {
		t.Fatalf("没有界面侧的录制改变了事件取舍：%d → %d", len(legacy.Events), len(broken.Events))
	}
	if broken.Events[0].ColName != "" || broken.Events[0].Event != legacy.Events[0].Event {
		t.Fatalf("没有界面侧却记了身份/改了事件：%+v", broken.Events[0])
	}

	// ③ 接上模型：同一批事件、同样条数，只**多记**身份
	wired := planReplayFor(tgt, events, replayStopAt, replaySkip)
	if len(wired.Events) != len(legacy.Events) {
		t.Fatalf("接线改变了事件取舍：%d → %d", len(legacy.Events), len(wired.Events))
	}
	if wired.Events[0].Event != legacy.Events[0].Event {
		t.Fatalf("接线改了事件体：%q → %q", legacy.Events[0].Event, wired.Events[0].Event)
	}
	if wired.Events[0].ColName != "b_target" || wired.Events[0].FieldID != 5 {
		t.Fatalf("身份没记上：%+v", wired.Events[0])
	}
}

// TestReplayCommandPathSendsToResolvedIdentity 是**端到端**：
// 走 `rec replay` 那条组装路径（`t.events()` → `planReplayFor` → `RunReplayWith`），
// 断言发出去的是**身份解析后的 id**（6008），而不是录制时那个（6000）。
// 注入口是本机回环上的假服务。
func TestReplayCommandPathSendsToResolvedIdentity(t *testing.T) {
	tgt, dir := recFixture(t)
	events, err := tgt.events()
	if err != nil {
		t.Fatalf("读事件侧: %v", err)
	}
	plan := planReplayFor(tgt, events, replayStopAt, replaySkip)
	if len(plan.Events) != 1 || plan.Events[0].ColName != "b_target" {
		t.Fatalf("产品路径没把身份写进计划：%+v", plan.Events)
	}

	client, sent := fakeInject(t, dir)
	res, rerr := client.RunReplayWith(plan, fep.ReplayRunOptions{Conn: 1, Settle: 10 * time.Millisecond})
	if rerr != nil {
		t.Fatalf("回放失败：%v", rerr)
	}
	if res.Sent != 1 || res.Skipped != 0 || res.Failed != 0 || res.Refused != 0 {
		t.Fatalf("结果 = %+v", res)
	}
	got := sent()
	if len(got) != 1 {
		t.Fatalf("注入口收到 %d 条，想要 1：%v", len(got), got)
	}
	if !strings.Contains(got[0], `idRef "6008"`) {
		t.Fatalf("发出去的不是身份解析后的 id：%q（还是 6000 的话，接线就没生效）", got[0])
	}
	if strings.Contains(got[0], `idRef "6000"`) {
		t.Fatalf("发的还是录制时那个裸 id：%q", got[0])
	}
}

// fakeInject 起一个**本机回环**上的注入口假服务（协议照 Client.roundTrip：
// 一问一答、行分隔）。它只做两件事：回 `status`、收事件。
//
// 收到事件后**延迟**推一次字节数（模拟运行期回包）：`waitResponse` 是"发完之后
// 等运行期的字节数变化"，所以涨字节必须发生在它开始等之后。
func fakeInject(t *testing.T, recordDir string) (*fep.Client, func() []string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	var mu sync.Mutex
	var got []string
	bytesRT := 0

	go func() {
		for {
			c, aerr := ln.Accept()
			if aerr != nil {
				return
			}
			func(c net.Conn) {
				defer c.Close()
				line, rerr := bufio.NewReader(c).ReadString('\n')
				line = strings.TrimSpace(line)
				reply := "OK"
				switch {
				case line == "status":
					mu.Lock()
					b := bytesRT
					mu.Unlock()
					reply = fmt.Sprintf(`{"conns":[{"id":1,"seq":1,"alive":true,"bytesRT":%d}]}`, b)
				case strings.HasPrefix(line, "@1 "):
					mu.Lock()
					got = append(got, strings.TrimPrefix(line, "@1 "))
					mu.Unlock()
					// 回包晚一点到：waitResponse 先读基线，下一次轮询才看得见变化
					time.AfterFunc(60*time.Millisecond, func() {
						mu.Lock()
						bytesRT += 100
						mu.Unlock()
					})
				}
				if rerr != nil && line == "" {
					return
				}
				_, _ = c.Write([]byte(reply + "\n"))
			}(c)
		}
	}()

	return &fep.Client{InjectAddr: ln.Addr().String(), RecordDir: recordDir, Tag: "t"},
		func() []string {
			mu.Lock()
			defer mu.Unlock()
			return append([]string(nil), got...)
		}
}
