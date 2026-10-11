package fep

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeFEGDC is the front end: it reports everything it receives. The AUI is NOT
// sent from here — in the real topology the runtime (the job client on the
// server) sends `om …`, which is why those frames land in the rt2gdc recording.
func fakeFEGDC(t *testing.T) (net.Listener, <-chan []byte) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	got := make(chan []byte, 16)
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 4096)
				for {
					n, err := c.Read(buf)
					if n > 0 {
						got <- append([]byte(nil), buf[:n]...)
					}
					if err != nil {
						return
					}
				}
			}(c)
		}
	}()
	return ln, got
}

// sessionFixture 起一个真代理（连着假前端）+ 一条"运行期"连接（由它发 AUI），
// 返回可用的 Client。
func sessionFixture(t *testing.T) (*Client, *Proxy, net.Conn, <-chan []byte, string) {
	t.Helper()
	gdc, gdcGot := fakeFEGDC(t)
	dir := t.TempDir()
	p := NewProxy(ProxyConfig{
		Listen:  "127.0.0.1:0",
		Target:  gdc.Addr().String(),
		LogDir:  dir,
		Tag:     "t",
		Control: "127.0.0.1:0",
	})
	if err := p.Start(); err != nil {
		t.Fatalf("start proxy: %v", err)
	}
	t.Cleanup(func() { p.Close(); gdc.Close() })

	rt, err := net.Dial("tcp", p.ListenAddr())
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	t.Cleanup(func() { rt.Close() })
	waitConns(t, p, 1)

	// 运行期把这一屏 AUI 发过来（真拓扑就是这个方向）
	aui := `om 1 {{an 0 Window 4097 {{name "w_axmt500"}} {{` +
		`{an 4097 ToolBarItem 5160 {{text "录入"}{actionActive "1"}} {}} ` +
		`{an 4097 Label 270 {{comment "cmt_xmda004"}{text "客户编号"}} {}} ` +
		`{an 4097 ButtonEdit 272 {{comment "cmt_xmda004"}} {{an 272 FormField 271 {{value ""}} {}}} {}} ` +
		`{an 4097 ToolBarItem 5156 {{text "确定"}{actionActive "0"}} {}}}}}`
	if _, err := rt.Write(Encode(Frame{Type: FrameTypeAUI, Payload: []byte(aui)})); err != nil {
		t.Fatalf("write AUI: %v", err)
	}

	// 数据路径先确认：假前端必须看到这屏 AUI
	expectChunk(t, gdcGot, "ToolBarItem")
	waitRecording(t, dir, "-rt2gdc.log")
	return &Client{InjectAddr: p.ControlAddr(), RecordDir: dir, Tag: "t"}, p, rt, gdcGot, dir
}

// waitRecording waits until a recording with the suffix exists and is non-empty.
func waitRecording(t *testing.T, dir, suffix string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), suffix) {
				continue
			}
			// 注意：DirEntry.Info() 在 Windows 上可能是枚举那一刻的陈旧大小（实测：
			// entryInfo=0 而 os.Stat=356），判断"有没有写进去"必须用 os.Stat。
			if st, serr := os.Stat(filepath.Join(dir, e.Name())); serr == nil && st.Size() > 0 {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	entries, _ := os.ReadDir(dir)
	var listing []string
	for _, e := range entries {
		st, _ := os.Stat(filepath.Join(dir, e.Name()))
		listing = append(listing, fmt.Sprintf("%s:%d", e.Name(), st.Size()))
	}
	t.Fatalf("%s 里始终没有非空的 *%s（目录：%v）", dir, suffix, listing)
}

func TestDoClickByNameReachesTheFrontEnd(t *testing.T) {
	c, p, rt, _, _ := sessionFixture(t)

	res, err := Do(c, DoConfig{Kind: "click", Target: "录入", Wait: -1})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if res.Event != `ActionEvent 0{{idRef "5160"}}` {
		t.Fatalf("事件 = %q", res.Event)
	}
	if !res.Sent {
		t.Fatalf("没发出去：%+v", res)
	}
	if res.Target != "#5160" || res.Name != "录入" {
		t.Fatalf("目标解析 = %+v", res)
	}
	expectOnConn(t, rt, `event _om 1{}{{ActionEvent 0{{idRef "5160"}}}}`)
	// 序号必须由注入方推进，且下一发要用 2
	waitSeq(t, p, 1, 2)
}

func TestDoInputPrefersTheFormFieldAndDryRunSendsNothing(t *testing.T) {
	c, _, rt, _, _ := sessionFixture(t)

	// dry-run：事件照构造、照校验，但一个字节都不发
	res, err := Do(c, DoConfig{Kind: "input", Target: "客户编号", Value: "CNJ-A50", DryRun: true, Wait: -1})
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if res.Sent {
		t.Fatal("dry-run 不该发送")
	}
	if res.Target != "#271" || !strings.HasPrefix(res.Event, `ConfigureEvent 0{{idRef "271"}{ value "CNJ-A50"}`) {
		t.Fatalf("应发到字段的 FormField（#271）：%+v", res)
	}
	// dry-run：运行期这一侧一个字节都不该收到
	rt.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	scratch := make([]byte, 512)
	if n, rerr := rt.Read(scratch); rerr == nil && n > 0 {
		t.Fatalf("dry-run 竟然发了字节: %q", scratch[:n])
	}
	rt.SetReadDeadline(time.Time{})

	// 真发一次
	res2, err := Do(c, DoConfig{Kind: "input", Target: "客户编号", Value: "CNJ-A50", Wait: -1})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if !res2.Sent {
		t.Fatalf("没发出去：%+v", res2)
	}
	expectOnConn(t, rt, `{ value "CNJ-A50"}`)
}

func TestDoRefusesUnavailableAction(t *testing.T) {
	c, _, _, _, _ := sessionFixture(t)
	// 5156 的 actionActive 为假：必须被拦，且不发包
	_, err := Do(c, DoConfig{Kind: "click", Target: "#5156", Wait: -1})
	var ref *Refusal
	if !asRefusal(err, &ref) || ref.Exit != 4 {
		t.Fatalf("actionActive 为假必须被拒: %v", err)
	}
}

func TestDoRefusesUnknownName(t *testing.T) {
	c, _, _, _, _ := sessionFixture(t)
	_, err := Do(c, DoConfig{Kind: "click", Target: "根本没有这个栏位", Wait: -1})
	var ref *Refusal
	if !asRefusal(err, &ref) || ref.Exit != 5 {
		t.Fatalf("不存在的名字必须被拒: %v", err)
	}
}

func TestAuditTrailIsAppendableJSONLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	if err := AppendAudit(path, AuditEntry{Verb: "click", Conn: 1, Target: "#5160", Sent: true}); err != nil {
		t.Fatal(err)
	}
	if err := AppendAudit(path, AuditEntry{Verb: "key", Conn: 1, Event: `KeyEvent 0{{keyName "Tab"}}`, Sent: true}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "{") || !strings.Contains(lines[1], "KeyEvent") {
		t.Fatalf("审计日志形状不对: %q", string(b))
	}
}

// 审计的每一行都要有时间：事后追"这一步是什么时候做的"是审计存在的理由。
func TestAuditEntryGetsATimestamp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	if err := AppendAudit(path, AuditEntry{Verb: "click", Sent: true}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var e AuditEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(b))), &e); err != nil {
		t.Fatal(err)
	}
	if e.Time.IsZero() || time.Since(e.Time) > time.Minute {
		t.Fatalf("时间戳不对: %v", e.Time)
	}
}

func TestDoWaitsForTheRuntimeAndReportsChanges(t *testing.T) {
	gdc, _ := fakeFEGDC(t)
	defer gdc.Close()
	dir := t.TempDir()
	p := NewProxy(ProxyConfig{Listen: "127.0.0.1:0", Target: gdc.Addr().String(),
		LogDir: dir, Tag: "t", Control: "127.0.0.1:0"})
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	rt, err := net.Dial("tcp", p.ListenAddr())
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	waitConns(t, p, 1)
	aui := `om 4 {{an 2858 Window 4097 {{name "w"}} {{an 4097 FormField 271 {{name "xmda_t.xmda004"}{value ""}} {}}}}}`
	if _, err := rt.Write(Encode(Frame{Type: FrameTypeAUI, Payload: []byte(aui)})); err != nil {
		t.Fatal(err)
	}
	waitRecording(t, dir, "-rt2gdc.log")

	// 运行期一侧在看到事件后推回一条更新 —— 这就是"作业真的做了事"在线上长什么样
	go func() {
		buf := make([]byte, 4096)
		for {
			rt.SetReadDeadline(time.Now().Add(3 * time.Second))
			n, rerr := rt.Read(buf)
			if n > 0 && strings.Contains(string(buf[:n]), "ConfigureEvent") {
				rt.Write(Encode(Frame{Type: FrameTypeAUI,
					Payload: []byte(`om 9 {{un 271 {{value "REPLY"}}}}`)}))
				return
			}
			if rerr != nil {
				return
			}
		}
	}()

	c := &Client{InjectAddr: p.ControlAddr(), RecordDir: dir, Tag: "t"}
	res, err := Do(c, DoConfig{Kind: "input", Target: "#271", Value: "X", Wait: 5 * time.Second})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if !res.Sent {
		t.Fatal("没发出去")
	}
	if res.Changes == nil {
		t.Fatal("应有 changes")
	}
	if res.Changes.RuntimeBytes == 0 || !res.Changes.Acked {
		t.Fatalf("运行期的回包没被等出来: %+v", res.Changes)
	}
	found := false
	for _, vc := range res.Changes.Values {
		if vc.ID == 271 && vc.To == "REPLY" {
			found = true
		}
	}
	if !found {
		t.Fatalf("值的回填没被汇总: %+v", res.Changes.Values)
	}
}

// 等条件：轮询录制件直到条件成立。
func TestWaitForValue(t *testing.T) {
	c, _, rt, _, _ := sessionFixture(t)
	done := make(chan error, 1)
	go func() {
		time.Sleep(400 * time.Millisecond)
		_, err := rt.Write(Encode(Frame{Type: FrameTypeAUI,
			Payload: []byte(`om 9 {{un 271 {{value "LATER"}}}}`)}))
		done <- err
	}()
	if _, err := c.WaitFor(WaitCond{Value: "客户编号=LATER"}, 5*time.Second); err != nil {
		t.Fatalf("WaitFor: %v", err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
