package fep

import (
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
	c, p, _, gdcGot, _ := sessionFixture(t)

	res, err := Do(c, DoConfig{Kind: "click", Target: "录入"})
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
	expectChunk(t, gdcGot, `event _om 1{}{{ActionEvent 0{{idRef "5160"}}}}`)
	// 序号必须由注入方推进，且下一发要用 2
	waitSeq(t, p, 1, 2)
}

func TestDoInputPrefersTheFormFieldAndDryRunSendsNothing(t *testing.T) {
	c, _, _, gdcGot, _ := sessionFixture(t)

	// dry-run：事件照构造、照校验，但一个字节都不发
	res, err := Do(c, DoConfig{Kind: "input", Target: "客户编号", Value: "CNJ-A50", DryRun: true})
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if res.Sent {
		t.Fatal("dry-run 不该发送")
	}
	if res.Target != "#271" || !strings.HasPrefix(res.Event, `ConfigureEvent 0{{idRef "271"}{ value "CNJ-A50"}`) {
		t.Fatalf("应发到字段的 FormField（#271）：%+v", res)
	}
	select {
	case b := <-gdcGot:
		t.Fatalf("dry-run 竟然发了字节: %q", b)
	case <-time.After(300 * time.Millisecond):
	}

	// 真发一次
	res2, err := Do(c, DoConfig{Kind: "input", Target: "客户编号", Value: "CNJ-A50"})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if !res2.Sent {
		t.Fatalf("没发出去：%+v", res2)
	}
	expectChunk(t, gdcGot, `{ value "CNJ-A50"}`)
}

func TestDoRefusesUnavailableAction(t *testing.T) {
	c, _, _, _, _ := sessionFixture(t)
	// 5156 的 actionActive 为假：必须被拦，且不发包
	_, err := Do(c, DoConfig{Kind: "click", Target: "#5156"})
	var ref *Refusal
	if !asRefusal(err, &ref) || ref.Exit != 4 {
		t.Fatalf("actionActive 为假必须被拒: %v", err)
	}
}

func TestDoRefusesUnknownName(t *testing.T) {
	c, _, _, _, _ := sessionFixture(t)
	_, err := Do(c, DoConfig{Kind: "click", Target: "根本没有这个栏位"})
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
