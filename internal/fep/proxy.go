package fep

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ProxyConfig describes a local reverse proxy between the T100 runtime and the
// local GDC instance.
//
// Topology (validated on a live job): the runtime dials Listen, we dial Target
// (the GDC instance's port). One runtime connection maps to one GDC connection,
// and each screen of the job gets its own connection — so everything that acts
// on a window must pick a connection, never assume "the" one.
type ProxyConfig struct {
	Listen  string // e.g. "0.0.0.0:7600" — where the runtime connects
	Target  string // e.g. "127.0.0.1:7400" — the local GDC instance
	LogDir  string // recordings; empty disables logging
	Tag     string // file prefix, e.g. "fep"
	Control string // control endpoint, e.g. "127.0.0.1:7899"; empty disables it
}

// ConnInfo is one live proxied connection.
type ConnInfo struct {
	ID      int
	Remote  string
	Opened  time.Time
	Seq     int // next `_om` sequence number the runtime expects from this front end
	Alive   bool
	BytesRT int64 // runtime -> GDC
	BytesFE int64 // GDC -> runtime
}

// ErrNoConn reports an injection aimed at a connection that does not exist.
var ErrNoConn = errors.New("fep: no such connection")

// Proxy is a byte-faithful, multi-connection reverse proxy. It never rewrites
// traffic: recordings are raw bytes, which is what makes replay and the AUI
// model reproducible from a session afterwards.
type Proxy struct {
	cfg ProxyConfig

	mu       sync.Mutex
	conns    map[int]*conn
	nextID   int
	listener net.Listener
	control  net.Listener
	closed   bool
	quitOnce sync.Once
	quit     chan struct{}
}

type conn struct {
	id      int
	remote  string
	opened  time.Time
	toGDC   net.Conn
	toRT    net.Conn
	seq     int
	bytesRT int64
	bytesFE int64
	alive   bool
	files   []*os.File
	mu      sync.Mutex
}

// NewProxy prepares a proxy; call Start to bind the sockets.
func NewProxy(cfg ProxyConfig) *Proxy {
	return &Proxy{cfg: cfg, conns: make(map[int]*conn), quit: make(chan struct{})}
}

// Quit is closed when a control client asked the proxy to stop
// (`quit` on the control endpoint). The resident process waits on it, which is
// how `tt fep session close` stops a session that lives in another process.
func (p *Proxy) Quit() <-chan struct{} { return p.quit }

// Start binds the listen and control sockets.
func (p *Proxy) Start() error {
	if p.cfg.Tag == "" {
		p.cfg.Tag = "fep"
	}
	if p.cfg.LogDir != "" {
		if err := os.MkdirAll(p.cfg.LogDir, 0o755); err != nil {
			return fmt.Errorf("fep: create log dir: %w", err)
		}
	}
	ln, err := net.Listen("tcp", p.cfg.Listen)
	if err != nil {
		return fmt.Errorf("fep: listen %s: %w", p.cfg.Listen, err)
	}
	p.listener = ln
	if p.cfg.Control != "" {
		ctl, err := net.Listen("tcp", p.cfg.Control)
		if err != nil {
			ln.Close()
			return fmt.Errorf("fep: control %s: %w", p.cfg.Control, err)
		}
		p.control = ctl
		go p.acceptControl(ctl)
	}
	go p.acceptLoop(ln)
	return nil
}

// ListenAddr reports the bound listen address ("host:port").
func (p *Proxy) ListenAddr() string {
	if p.listener == nil {
		return ""
	}
	return p.listener.Addr().String()
}

// ControlAddr reports the bound control address ("" when disabled).
func (p *Proxy) ControlAddr() string {
	if p.control == nil {
		return ""
	}
	return p.control.Addr().String()
}

// Close stops both sockets and every live connection.
func (p *Proxy) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	ln, ctl := p.listener, p.control
	conns := make([]*conn, 0, len(p.conns))
	for _, c := range p.conns {
		conns = append(conns, c)
	}
	p.mu.Unlock()

	if ln != nil {
		ln.Close()
	}
	if ctl != nil {
		ctl.Close()
	}
	for _, c := range conns {
		c.close()
	}
	return nil
}

func (p *Proxy) acceptLoop(ln net.Listener) {
	for {
		rt, err := ln.Accept()
		if err != nil {
			return // listener closed
		}
		go p.serve(rt)
	}
}

// serve handles one runtime connection: dial the GDC instance, pump both
// directions, and keep the failure local to this connection.
func (p *Proxy) serve(rt net.Conn) {
	gdc, err := net.Dial("tcp", p.cfg.Target)
	if err != nil {
		rt.Close()
		return
	}
	c := &conn{remote: rt.RemoteAddr().String(), opened: time.Now(), toGDC: gdc, toRT: rt, alive: true}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		c.close()
		return
	}
	p.nextID++
	c.id = p.nextID
	p.conns[c.id] = c
	p.mu.Unlock()

	c.openLogs(p.cfg.LogDir, p.cfg.Tag)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); p.pump(c, rt, gdc, true) }()  // runtime -> GDC
	go func() { defer wg.Done(); p.pump(c, gdc, rt, false) }() // GDC -> runtime
	wg.Wait()
	c.close()
	p.mu.Lock()
	delete(p.conns, c.id)
	p.mu.Unlock()
}

// pump copies one direction, tees it into the recording and tracks the front
// end's `_om` sequence so injections can be numbered correctly.
func (p *Proxy) pump(c *conn, src, dst net.Conn, fromRuntime bool) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			c.write(chunk, fromRuntime)
			if _, werr := dst.Write(chunk); werr != nil {
				dst.Close()
				src.Close()
				return
			}
		}
		if err != nil {
			dst.Close()
			src.Close()
			return
		}
	}
}

// Conns snapshots the live connections.
func (p *Proxy) Conns() []ConnInfo {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]ConnInfo, 0, len(p.conns))
	for _, c := range p.conns {
		c.mu.Lock()
		out = append(out, ConnInfo{ID: c.id, Remote: c.remote, Opened: c.opened, Seq: c.seq, Alive: c.alive, BytesRT: c.bytesRT, BytesFE: c.bytesFE})
		c.mu.Unlock()
	}
	return out
}

// Inject sends one front-end event (`ActionEvent 0{{…}}`) to a connection,
// wrapped the way the client wraps its own events:
//
//	event _om <next-seq>{}{{<event>}}
//
// The sequence number must be the next one the runtime expects for that
// connection; guessing it resets the connection (measured).
func (p *Proxy) Inject(connID int, eventText string) error {
	p.mu.Lock()
	c, ok := p.conns[connID]
	p.mu.Unlock()
	if !ok {
		return fmt.Errorf("%w: %d", ErrNoConn, connID)
	}
	// `seq` already holds the next number the runtime expects, so use it as is
	// and only then advance it.
	c.mu.Lock()
	if c.seq == 0 {
		c.seq = 1
	}
	seq := c.seq
	c.seq++
	c.mu.Unlock()

	text := fmt.Sprintf("event _om %d{}{{%s}}\n", seq, strings.TrimSpace(eventText))
	frame := Encode(Frame{Type: FrameTypeAUI, Payload: []byte(strings.TrimRight(text, "\n"))})
	c.write(frame, false) // injected frames belong on the "front end" side of the recording
	_, err := c.toGDC.Write(frame)
	return err
}

// acceptControl serves the injection endpoint:
//
//	list                 -> one line per connection
//	@<id> <event text>   -> send to that connection
//	<event text>         -> send to the most recent connection
func (p *Proxy) acceptControl(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 8192)
			var acc []byte
			for {
				n, err := c.Read(buf)
				if n > 0 {
					acc = append(acc, buf[:n]...)
					for {
						i := indexByte(acc, '\n')
						if i < 0 {
							break
						}
						line := strings.TrimRight(string(acc[:i]), "\r")
						acc = acc[i+1:]
						p.handleControlLine(c, line)
					}
				}
				if err != nil {
					return
				}
			}
		}(conn)
	}
}

func (p *Proxy) handleControlLine(w io.Writer, line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	if line == "status" {
		io.WriteString(w, p.statusJSON()+"\n")
		return
	}
	if line == "list" {
		infos := p.Conns()
		var b strings.Builder
		for _, ci := range infos {
			fmt.Fprintf(&b, "#%d seq=%d alive=%v remote=%s\n", ci.ID, ci.Seq, ci.Alive, ci.Remote)
		}
		if b.Len() == 0 {
			b.WriteString("(no connections)\n")
		}
		io.WriteString(w, b.String())
		return
	}
	if line == "quit" || line == "stop" {
		io.WriteString(w, "OK quitting\n")
		p.quitOnce.Do(func() { close(p.quit) })
		return
	}
	target := 0
	text := line
	if strings.HasPrefix(line, "@") {
		sp := strings.IndexByte(line, ' ')
		if sp < 0 {
			fmt.Fprintln(w, "ERR bad command")
			return
		}
		id, err := strconv.Atoi(line[1:sp])
		if err != nil {
			fmt.Fprintln(w, "ERR bad connection id")
			return
		}
		target, text = id, line[sp+1:]
	} else {
		p.mu.Lock()
		best := 0
		for id := range p.conns {
			if id > best {
				best = id
			}
		}
		p.mu.Unlock()
		if best == 0 {
			fmt.Fprintln(w, "ERR no connection")
			return
		}
		target = best
	}
	if err := p.Inject(target, text); err != nil {
		fmt.Fprintf(w, "ERR %v\n", err)
		return
	}
	fmt.Fprintf(w, "OK #%d %s\n", target, strings.TrimSpace(text))
}

func indexByte(b []byte, c byte) int {
	for i := range b {
		if b[i] == c {
			return i
		}
	}
	return -1
}

// write tees a chunk into the connection's recording.
func (c *conn) write(chunk []byte, fromRuntime bool) {
	c.mu.Lock()
	// The `_om` numbers belong to the front end, so they travel on the
	// GDC -> runtime direction (fromRuntime == false), not the other way.
	if fromRuntime {
		c.bytesRT += int64(len(chunk))
	} else {
		c.bytesFE += int64(len(chunk))
		c.trackSeq(chunk)
	}
	files := c.files
	c.mu.Unlock()
	if len(files) == 0 {
		return
	}
	var f *os.File
	if fromRuntime {
		f = files[0]
	} else {
		f = files[1]
	}
	if f != nil {
		f.Write(chunk)
	}
}

// trackSeq keeps the next `_om` number for this connection. The front end
// numbers its own events, so the next number is max(seen)+1.
func (c *conn) trackSeq(chunk []byte) {
	s := string(chunk)
	for {
		i := strings.Index(s, "event _om ")
		if i < 0 {
			return
		}
		s = s[i+len("event _om "):]
		j := 0
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		if j > 0 {
			if n, err := strconv.Atoi(s[:j]); err == nil && n+1 > c.seq {
				c.seq = n + 1
			}
		}
	}
}

// openLogs creates the per-connection recording files: index 0 is runtime->GDC,
// index 1 is GDC->runtime (the front end's own events).
func (c *conn) openLogs(dir, tag string) {
	if dir == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	names := []string{
		filepath.Join(dir, fmt.Sprintf("%s-%d-rt2gdc.log", tag, c.id)),
		filepath.Join(dir, fmt.Sprintf("%s-%d-gdc2rt.log", tag, c.id)),
	}
	for _, n := range names {
		f, err := os.Create(n)
		if err == nil {
			c.files = append(c.files, f)
		}
	}
}

func (c *conn) close() {
	c.mu.Lock()
	c.alive = false
	conns := c.toGDC
	rt := c.toRT
	files := c.files
	c.files = nil
	c.mu.Unlock()
	if conns != nil {
		conns.Close()
	}
	if rt != nil {
		rt.Close()
	}
	for _, f := range files {
		f.Close()
	}
}
