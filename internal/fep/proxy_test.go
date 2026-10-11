package fep

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeGDC behaves like a Genero front end: on accept it immediately emits one
// `event _om 3` (its own event numbering) and then reports every chunk it
// receives from the runtime side on the returned channel.
func fakeGDC(t *testing.T) (net.Listener, <-chan []byte) {
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
				c.Write(Encode(Frame{Type: FrameTypeAUI, Payload: []byte(`event _om 3{}{{ActionEvent 0{{idRef "42"}}}}`)}))
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

// expectChunk waits for one received chunk that contains want.
func expectChunk(t *testing.T, got <-chan []byte, want string) {
	t.Helper()
	deadline := time.After(3 * time.Second)
	for {
		select {
		case b := <-got:
			if strings.Contains(string(b), want) {
				return
			}
		case <-deadline:
			t.Fatalf("never saw %q on the GDC side", want)
		}
	}
}

func TestProxyForwardsBothWaysAndRecords(t *testing.T) {
	gdc, gdcGot := fakeGDC(t)
	defer gdc.Close()
	dir := t.TempDir()

	p := NewProxy(ProxyConfig{Listen: "127.0.0.1:0", Target: gdc.Addr().String(), LogDir: dir, Tag: "t", Control: "127.0.0.1:0"})
	if err := p.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer p.Close()

	rt, err := net.Dial("tcp", p.ListenAddr())
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer rt.Close()

	const hello = "RUNTIME-HELLO"
	if _, err := rt.Write([]byte(hello)); err != nil {
		t.Fatalf("write: %v", err)
	}
	expectChunk(t, gdcGot, hello)

	buf := make([]byte, 512)
	rt.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, err := rt.Read(buf)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if !strings.Contains(string(buf[:n]), "event _om 3") {
		t.Fatalf("front-end frame not forwarded: %q", string(buf[:n]))
	}

	// one connection, and it must have learned the front end's next number
	waitConns(t, p, 1)
	infos := p.Conns()
	if len(infos) != 1 {
		t.Fatalf("conns = %d", len(infos))
	}
	if infos[0].Seq != 4 {
		t.Fatalf("next sequence = %d, want 4 (the front end numbered its event 3)", infos[0].Seq)
	}

	// recordings: raw bytes, both directions, one file each
	deadline := time.Now().Add(2 * time.Second)
	var entries []os.DirEntry
	for time.Now().Before(deadline) {
		if entries, _ = os.ReadDir(dir); len(entries) == 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if len(entries) != 2 {
		t.Fatalf("expected two recording files, got %d", len(entries))
	}
	var sawHello, sawEvent bool
	for _, e := range entries {
		b, _ := os.ReadFile(filepath.Join(dir, e.Name()))
		if strings.Contains(string(b), hello) {
			sawHello = true
		}
		if strings.Contains(string(b), "event _om 3") {
			sawEvent = true
		}
	}
	if !sawHello || !sawEvent {
		t.Fatalf("recordings incomplete: runtime bytes=%v front-end bytes=%v", sawHello, sawEvent)
	}
}

func TestProxyInjectionIsNumberedAndAddressed(t *testing.T) {
	gdc, gdcGot := fakeGDC(t)
	defer gdc.Close()
	p := NewProxy(ProxyConfig{Listen: "127.0.0.1:0", Target: gdc.Addr().String(), Control: "127.0.0.1:0"})
	if err := p.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer p.Close()

	rt, err := net.Dial("tcp", p.ListenAddr())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer rt.Close()
	waitConns(t, p, 1)
	waitSeq(t, p, 1, 4) // the front end numbered its own event 3

	if err := p.Inject(1, `ActionEvent 0{{idRef "5160"}}`); err != nil {
		t.Fatalf("inject: %v", err)
	}
	expectChunk(t, gdcGot, `event _om 4{}{{ActionEvent 0{{idRef "5160"}}}}`)

	if err := p.Inject(99, `ActionEvent 0{{idRef "1"}}`); err == nil {
		t.Fatal("injecting into a missing connection must fail")
	}
}

func TestProxyControlEndpointAddressesConnections(t *testing.T) {
	gdc, gdcGot := fakeGDC(t)
	defer gdc.Close()
	p := NewProxy(ProxyConfig{Listen: "127.0.0.1:0", Target: gdc.Addr().String(), Control: "127.0.0.1:0"})
	if err := p.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer p.Close()

	rt, err := net.Dial("tcp", p.ListenAddr())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer rt.Close()
	waitConns(t, p, 1)
	waitSeq(t, p, 1, 4)

	ctl, err := net.Dial("tcp", p.ControlAddr())
	if err != nil {
		t.Fatalf("dial control: %v", err)
	}
	defer ctl.Close()
	if _, err := ctl.Write([]byte("list\n")); err != nil {
		t.Fatal(err)
	}
	ctl.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 256)
	n, err := ctl.Read(buf)
	if err != nil {
		t.Fatalf("read control: %v", err)
	}
	if !strings.Contains(string(buf[:n]), "#1 seq=") {
		t.Fatalf("control list = %q", string(buf[:n]))
	}

	if _, err := ctl.Write([]byte("@1 " + `KeyEvent 0{{keyName "Tab"}}` + "\n")); err != nil {
		t.Fatal(err)
	}
	expectChunk(t, gdcGot, `event _om 4{}{{KeyEvent 0{{keyName "Tab"}}}}`)
}

func TestProxySurvivesAClientReset(t *testing.T) {
	gdc, _ := fakeGDC(t)
	defer gdc.Close()
	p := NewProxy(ProxyConfig{Listen: "127.0.0.1:0", Target: gdc.Addr().String()})
	if err := p.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer p.Close()

	// first client disappears without a goodbye
	first, err := net.Dial("tcp", p.ListenAddr())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	waitConns(t, p, 1)
	first.Close()
	waitConns(t, p, 0)

	// a second connection must still work
	second, err := net.Dial("tcp", p.ListenAddr())
	if err != nil {
		t.Fatalf("second dial: %v", err)
	}
	defer second.Close()
	waitConns(t, p, 1)
}

func waitConns(t *testing.T, p *Proxy, want int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if len(p.Conns()) == want {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("connections = %d, want %d", len(p.Conns()), want)
}

// waitSeq waits until the proxy learned a connection's next `_om` number.
func waitSeq(t *testing.T, p *Proxy, id, want int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		for _, ci := range p.Conns() {
			if ci.ID == id && ci.Seq == want {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("connection %d never reached next-sequence %d: %+v", id, want, p.Conns())
}
