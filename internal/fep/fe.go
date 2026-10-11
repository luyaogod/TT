package fep

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FEServer 是"我们自己当 GDC"的**最小前端**。
//
// 为什么要它（用户提的方案，也是卡点最多的那条路的正解）：
// 运行期只会把界面画给"一个前端"看；我们一直借 GDC 当前端 ✗，于是
// 焦点/状态/事件形状都得从 GDC 的转述里反推（今天大半时间都花在这上面 ✗）。
// 换成我们自己当那一个前端之后：**收什么、发什么、什么时候发**全在我们手里。
//
// 关键事实（实测）：运行期按 `FGLSERVER 端口 + 6400` 回连前端。
// 所以 FGLSERVER 设成 `<我们的IP>:1300`，它就会连到我们监听的 7700。
// `internal/debug` 的启动器会自动导出 FGLSERVER（session.go:550），接线是现成的。
//
// 它做的三件事：
//  1. 监听 → 收运行期的帧 → 喂给 Model（AUI 折叠）；
//  2. **把收到的字节原样写进录制文件**（与既有 `*-rt2gdc.log` 同格式）——
//     这样 `tt fep ui state/find/roles --from <该文件>` 等既有命令**全部照用**；
//  3. `Send(ev)`：按运行期的 `_om` 序号推进，把事件发回去（与注入口同一套规矩）。
type FEServer struct {
	ln  net.Listener
	ctl net.Listener

	mu      sync.Mutex
	conn    net.Conn
	dec     *Decoder
	model   *Model
	seq     int
	bytesRT int64
	bytesFE int64
	msgs    int
	rec     *os.File
	recPath string
	err     error
	// handshook：已经回过 `meta Client`（只回一次）。
	handshook bool
}

// ListenFE 在 addr 上等运行期连入，并把收到的 AUI 写进 dir 下的录制文件。
//
// ctlAddr 是**控制口**（本地回环）：驱动用它读状态、发事件 —— 与注入口同一形态，
// 目的就是让上层那些已有的命令不用改。
func ListenFE(addr, ctlAddr, dir string) (*FEServer, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("前端监听 %s 失败：%w", addr, err)
	}
	s := &FEServer{ln: ln, dec: NewDecoder(), model: NewModel()}
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err == nil {
			p := filepath.Join(dir, "fe-"+time.Now().Format("20060102-150405")+"-rt2fe.log")
			if f, ferr := os.Create(p); ferr == nil {
				s.rec, s.recPath = f, p
			}
		}
	}
	if ctlAddr != "" {
		cl, cerr := net.Listen("tcp", ctlAddr)
		if cerr != nil {
			_ = ln.Close()
			return nil, fmt.Errorf("控制口监听 %s 失败：%w", ctlAddr, cerr)
		}
		s.ctl = cl
		go s.ctlLoop()
	}
	return s, nil
}

// RecPath 是录制文件路径（喂给既有的 ui 命令用）。
func (s *FEServer) RecPath() string { return s.recPath }

// Addr / CtlAddr 便于调用方回显"该把 FGLSERVER 指到哪"。
func (s *FEServer) Addr() string {
	if s == nil || s.ln == nil {
		return ""
	}
	return s.ln.Addr().String()
}

func (s *FEServer) CtlAddr() string {
	if s == nil || s.ctl == nil {
		return ""
	}
	return s.ctl.Addr().String()
}

// Accept 阻塞等运行期连进来（连上就开始读）。
func (s *FEServer) Accept() error {
	c, err := s.ln.Accept()
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.conn = c
	s.mu.Unlock()
	go s.readLoop(c)
	return nil
}

func (s *FEServer) readLoop(c net.Conn) {
	buf := make([]byte, 64*1024)
	for {
		n, err := c.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			s.mu.Lock()
			s.bytesRT += int64(n)
			if s.rec != nil {
				_, _ = s.rec.Write(chunk)
				_ = s.rec.Sync()
			}
			frames, derr := s.dec.Feed(chunk)
			// **握手要在原始字节上认**，不能等"切出整帧"：
			// 实测运行期先发 206 字节的首帧，比整帧少几个字节，此时解码器还在等剩下的；
			// 而它在等我们回 `meta Client` 才肯继续推 —— 两边互等就死住了。
			if !s.handshook && strings.Contains(string(chunk), "meta Connection") {
				s.handshook = true
				host, _ := os.Hostname()
				reply := fmt.Sprintf(`meta Client{{name "tt-fe"}{version "0.1"}{host %q}{connections "1"}{port %q}{frontEndID2 "{%s}"}{encapsulation "1"}{filetransfer "1"}}`,
					host, s.portStr(), newFEID())
				if _, werr := c.Write(Encode(Frame{Type: FrameTypeAUI, Payload: []byte(reply)})); werr != nil {
					s.err = werr
				}
			}
			for _, f := range frames {
				s.msgs++
				// 只折叠看起来是 AUI 的帧（握手帧 meta … 不进去）。
				p := string(f.Payload)
				if len(p) >= 3 && (p[:3] == "om " || p[:3] == "an " || p[:3] == "un " || p[:3] == "rn ") {
					_ = s.model.ApplyFrame(f)
				}
			}
			if derr != nil {
				s.err = derr
			}
			s.mu.Unlock()
		}
		if err != nil {
			s.mu.Lock()
			s.err = err
			s.mu.Unlock()
			return
		}
	}
}

// Send 把一条事件体发给运行期（自动推进 _om 序号）。
//
// 序号规矩与注入口一致：**永远由发送方推进**，重号会打断连接（实测）。
func (s *FEServer) Send(ev string) error {
	s.mu.Lock()
	s.seq++
	seq := s.seq
	c := s.conn
	s.mu.Unlock()
	if c == nil {
		return fmt.Errorf("运行期还没连上来")
	}
	b := EncodeEvent(seq, ev)
	s.mu.Lock()
	s.bytesFE += int64(len(b))
	s.mu.Unlock()
	_, err := c.Write(b)
	time.Sleep(300 * time.Millisecond) // 让运行期把这一屏推完（fold 用）
	return err
}

// Model 给出当前折叠出来的 AUI（调用方自己加锁读）。
func (s *FEServer) Model() *Model {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.model
}

// FEStatus 是控制口 status 的返回（人读 + 机读都要够）。
type FEStatus struct {
	Listening string `json:"listening"`
	Control   string `json:"control"`
	Recording string `json:"recording,omitempty"`
	Connected bool   `json:"connected"`
	Seq       int    `json:"seq"`
	BytesRT   int64  `json:"bytesRT"`
	BytesFE   int64  `json:"bytesFE"`
	Messages  int    `json:"messages"`
	Objects   int    `json:"objects"`
	Err       string `json:"err,omitempty"`
}

func (s *FEServer) Status() FEStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := FEStatus{
		Listening: s.Addr(), Control: s.CtlAddr(), Recording: s.recPath,
		Connected: s.conn != nil, Seq: s.seq,
		BytesRT: s.bytesRT, BytesFE: s.bytesFE, Messages: s.msgs,
	}
	if s.model != nil {
		st.Objects = s.model.Len()
	}
	if s.err != nil {
		st.Err = s.err.Error()
	}
	return st
}

// ctlLoop 是控制口：与注入口同一形态的**行协议**，好让上层驱动不用改。
//
//	status            → 一行 JSON（FEStatus）
//	send <事件体>      → 发一条事件，回 "OK"
//	quit              → 关闭
func (s *FEServer) ctlLoop() {
	for {
		c, err := s.ctl.Accept()
		if err != nil {
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 64*1024)
			n, _ := c.Read(buf)
			line := string(buf[:n])
			line = trimNewline(line)
			switch {
			case line == "status":
				st := s.Status()
				fmt.Fprintf(c, "{\"connected\":%v,\"seq\":%d,\"bytesRT\":%d,\"bytesFE\":%d,\"messages\":%d,\"objects\":%d,\"recording\":%q}\n",
					st.Connected, st.Seq, st.BytesRT, st.BytesFE, st.Messages, st.Objects, st.Recording)
			case line == "quit":
				fmt.Fprintln(c, "OK")
				os.Exit(0)
			case len(line) > 5 && line[:5] == "send ":
				if err := s.Send(line[5:]); err != nil {
					fmt.Fprintf(c, "ERR %v\n", err)
					return
				}
				fmt.Fprintln(c, "OK")
			default:
				fmt.Fprintln(c, "ERR 未知命令（status / send <事件体> / quit）")
			}
		}(c)
	}
}

func trimNewline(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

// portStr 是我们**自己在 FGLSERVER 里报的那个端口**（GDC 报的是它自己监听的 7400；
// 我们回给运行期的就是本进程实际监听的那个）。
func (s *FEServer) portStr() string {
	if s == nil || s.ln == nil {
		return "0"
	}
	if a, ok := s.ln.Addr().(*net.TCPAddr); ok {
		return fmt.Sprintf("%d", a.Port)
	}
	return "0"
}

// newFEID 给一个前端实例标识（真机形态是带花括号的 UUID）。
func newFEID() string {
	b := make([]byte, 16)
	now := time.Now().UnixNano()
	for i := range b {
		b[i] = byte(now >> (uint(i%8) * 8))
	}
	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15])
}
