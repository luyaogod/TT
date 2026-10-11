package fep

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"tt/internal/winproc"
)

// SessionFileName is the cross-process session registry inside the data
// directory. `tt fep` commands are separate processes, so the session has to be
// discoverable (and collectable) from outside the process that started it.
const SessionFileName = ".tt-fep.json"

// ErrNoSession reports that no session is registered.
var ErrNoSession = errors.New("fep: no session registered")

// Session is one running, isolated job: the GDC instance we launched, the local
// proxy the runtime dials, and the control endpoint for injections.
type Session struct {
	Env        string    `json:"env"`
	Zone       string    `json:"zone"`
	Prog       string    `json:"prog"`
	GDCPID     int       `json:"gdcPid"`
	GDCPort    int       `json:"gdcPort"`
	ProxyPort  int       `json:"proxyPort"`
	InjectPort int       `json:"injectPort"`
	Tag        string    `json:"tag"`       // recording prefix inside the cache dir
	RecordDir  string    `json:"recordDir"` // where this session's recordings live
	OwnerPID   int       `json:"ownerPid"`  // the process that started it
	Started    time.Time `json:"started"`
}

// PIDs lists the processes this session owns. Cleanup must only ever touch
// these: the rule that kept a stray sweep from killing unrelated processes.
func (s *Session) PIDs() []int {
	var out []int
	if s.GDCPID != 0 {
		out = append(out, s.GDCPID)
	}
	return out
}

// Registry reads and writes the session file. The registration is written
// before anything else happens (register first, clean up later) so an
// interrupted run can always be collected.
type Registry struct {
	Path string
}

// NewRegistry points a registry at the data directory.
func NewRegistry(dataDir string) *Registry {
	return &Registry{Path: filepath.Join(dataDir, SessionFileName)}
}

// Load reads the session file.
func (r *Registry) Load() (*Session, error) {
	b, err := os.ReadFile(r.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoSession
		}
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("fep: session file %s: %w", r.Path, err)
	}
	return &s, nil
}

// Save writes the session file atomically (temp file + rename): a partially
// written registry would be worse than none, because cleanup trusts it.
func (r *Registry) Save(s *Session) error {
	b, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}
	tmp := r.Path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, r.Path)
}

// Clear removes the session file (used after a successful cleanup).
func (r *Registry) Clear() error {
	err := os.Remove(r.Path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Stop quits the session and cleans up after it.
//
// 顺序很重要：先请**常驻进程**自己收工（它知道怎么关远端作业、停代理、收实例），
// 再它没退干净时兵底杀登记的 pid。反过来的话，远端 `r.r` 作业会被留在服务器上。
func (r *Registry) Stop() (*Session, error) {
	s, err := r.Load()
	if err != nil {
		return nil, err
	}
	self := os.Getpid()
	// 1) 请常驻进程自己退（走完整收尾）
	if s.InjectPort != 0 {
		c := &Client{InjectAddr: net.JoinHostPort("127.0.0.1", strconv.Itoa(s.InjectPort))}
		if qerr := c.Quit(); qerr == nil {
			deadline := time.Now().Add(8 * time.Second)
			for time.Now().Before(deadline) {
				if s.OwnerPID == 0 || !winproc.Alive(s.OwnerPID) {
					break
				}
				time.Sleep(200 * time.Millisecond)
			}
		}
	}
	// 2) 还活着的：只杀登记的 pid
	var errs []string
	for _, pid := range s.PIDs() {
		if winproc.Alive(pid) {
			if kerr := winproc.Kill(pid); kerr != nil {
				errs = append(errs, kerr.Error())
			}
		}
	}
	// 3) 常驻进程自己（如果它没退干净）：它也是本会话登记的进程
	if s.OwnerPID != 0 && s.OwnerPID != self && winproc.Alive(s.OwnerPID) {
		if kerr := winproc.Kill(s.OwnerPID); kerr != nil {
			errs = append(errs, kerr.Error())
		}
	}
	if cerr := r.Clear(); cerr != nil {
		errs = append(errs, cerr.Error())
	}
	if len(errs) > 0 {
		return s, fmt.Errorf("fep: 收尾有失败: %s", strings.Join(errs, "; "))
	}
	return s, nil
}

// Elevation is the privilege relationship the whole approach depends on: an
// elevated GDC cannot be driven from a normal-privilege process (window
// messages are refused with ACCESS_DENIED and its new windows steal the
// foreground), so the check is a pre-flight, not a warning.
type Elevation struct {
	SelfElevated   bool
	TargetElevated bool
	OK             bool
	Advice         string
}

// CheckElevation compares our token elevation with the GDC instance's.
// A zero pid means "this process", so the comparison is trivially OK — it exists
// so callers can ask "am I elevated?" through the same door.
func CheckElevation(gdcPID int) (Elevation, error) {
	self, err := winproc.TokenElevated(0)
	if err != nil {
		return Elevation{}, err
	}
	target := self
	if gdcPID != 0 {
		target, err = winproc.TokenElevated(gdcPID)
		if err != nil {
			return Elevation{SelfElevated: self}, err
		}
	}
	e := Elevation{SelfElevated: self, TargetElevated: target, OK: self == target}
	if !e.OK {
		if target && !self {
			e.Advice = "GDC 以管理员权限运行，TT 是普通权限：无法驱动它的窗口。修复：结束 GDC，" +
				"清掉 HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\AppCompatFlags\\Layers 里 gdc.exe 的 RUNASADMIN，" +
				"再以普通权限重开（GDC 清单本身是 asInvoker，不需要提权）。"
		} else {
			e.Advice = "TT 以管理员权限运行而 GDC 是普通权限：请用普通权限运行 TT。"
		}
	}
	return e, nil
}
