package fep

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	Tag        string    `json:"tag"`      // recording prefix inside the cache dir
	OwnerPID   int       `json:"ownerPid"` // the process that started it
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

// Stop kills the processes the session registered and clears the registry.
// Processes that are already gone are not an error.
func (r *Registry) Stop() (*Session, error) {
	s, err := r.Load()
	if err != nil {
		return nil, err
	}
	for _, pid := range s.PIDs() {
		if winproc.Alive(pid) {
			if err := winproc.Kill(pid); err != nil {
				return s, fmt.Errorf("fep: kill %d: %w", pid, err)
			}
		}
	}
	return s, r.Clear()
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
