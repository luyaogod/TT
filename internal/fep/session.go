package fep

import (
	"bytes"
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

// 会话的两种用途：**录一轮**与**把一份录制回放出来**。
//
// 为什么登记里要写这一栏：两种会话落在登记里的形状一模一样（都是隔离 GDC + 代理 +
// 常驻 Runtime），但活法完全不同 —— 录制是"人还在录"，回放是"发完就完了"。
// 不区分的话，录制面板只能把两者都说成"正在录"：面板上写着"正在录:fep-x"，
// 而那一轮其实是回放留下的窗口（用户 2026-10-15 撞上过）。
const (
	SessionKindRec    = "rec"    // tt fep rec start：人在录
	SessionKindReplay = "replay" // tt fep rec reproduce：把一份录制重放出来
)

// Session is one running, isolated job: the GDC instance we launched, the local
// proxy the runtime dials, and the control endpoint for injections.
type Session struct {
	Env        string     `json:"env"`
	Zone       string     `json:"zone"`
	Prog       string     `json:"prog"`
	Path       LaunchPath `json:"path,omitempty"` // 这一次作业怎么起的（进录制身份；见 path.go）
	Kind       string     `json:"kind,omitempty"` // 这一轮是录还是回放（SessionKindRec / SessionKindReplay）
	Ent        string     `json:"ent,omitempty"`  // 本轮下发的企业编号（ENT）；空 = 沿用登录 shell 的 $TOPENT
	GDCPID     int        `json:"gdcPid"`
	GDCPort    int        `json:"gdcPort"`
	ProxyPort  int        `json:"proxyPort"`
	InjectPort int        `json:"injectPort"`
	Tag        string     `json:"tag"`       // recording prefix inside the cache dir（= 录制 ID）
	RecordDir  string     `json:"recordDir"` // where this session's recordings live
	OwnerPID   int        `json:"ownerPid"`  // the process that started it
	Started    time.Time  `json:"started"`
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

// Stale reports whether this registration outlived the process that started it.
//
// 为什么需要它：**登记文件比进程活得久**。`rec reproduce --no-wait`（面板那个回放
// 按钮就是这么用的）发完事件就退，留下的是一份 ownerPid 已经不在的登记 + 一个还开着
// 的 gdc 窗口（用户要看最后那一屏）。谁把这份登记当成"还在录"，下一轮起录就会被莫名
// 拒掉 —— 面板会说"已经在录了"，可其实谁都没在录。
//
// OwnerPID 为 0 表示"不知道是谁起的"（更早版本写的登记）：**不**当作过期，
// 宁可保守地拦住 —— 猜错的代价是一轮录制与一组端口。
func (s *Session) Stale() bool {
	return s != nil && s.OwnerPID != 0 && !winproc.Alive(s.OwnerPID)
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

// ErrAmbiguous 报告：登记里有**多条**会话，而调用方没有点名要哪一条。
//
// 为什么要这个错误而不是"取第一条"：多会话之后，一个不带目标的 `tt fep do` / `replay`
// 落到哪一条上是不可预测的 —— 而"落错了"的表现是动作打在别人的界面上、输出一切正常。
// 与调试线那条线（`SESSION_AMBIGUOUS`）同一条规矩：**多候选不自己挑**。
type ErrAmbiguous struct {
	Tags []string
}

func (e *ErrAmbiguous) Error() string {
	return "fep: 登记里有多条会话，请点名要哪一条（--tag）: " + strings.Join(e.Tags, "、")
}

// ErrSessionNotFound 报告：点名的 `Tag` 不在登记里（与"一份登记都没有"区分开）。
var ErrSessionNotFound = errors.New("fep: 没有这份登记")

// readAll 读出登记里的**全部**会话（新格式 `{"sessions":[…]}`）。
//
// 兼容旧格式：从前这个文件里是**单个** Session 对象（`{"tag":…,"gdcPid":…}`）。
// 两种都认，是因为升级是渐进的 —— 一份旧登记不该让用户先手工删文件才能用新版。
func (r *Registry) readAll() ([]*Session, error) {
	b, err := os.ReadFile(r.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	trimmed := bytes.TrimSpace(b)
	if len(trimmed) == 0 {
		return nil, nil
	}
	// 新格式（也可能是一个裸数组，两处都认）
	var wrapper struct {
		Sessions []*Session `json:"sessions"`
	}
	if err := json.Unmarshal(trimmed, &wrapper); err == nil && wrapper.Sessions != nil {
		return wrapper.Sessions, nil
	}
	var list []*Session
	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &list); err != nil {
			return nil, fmt.Errorf("fep: session file %s: %w", r.Path, err)
		}
		return list, nil
	}
	// 旧格式：单个对象
	var one Session
	if err := json.Unmarshal(trimmed, &one); err != nil {
		return nil, fmt.Errorf("fep: session file %s: %w", r.Path, err)
	}
	return []*Session{&one}, nil
}

// writeAll 原子写回全部会话（临时文件 + rename）。
func (r *Registry) writeAll(list []*Session) error {
	if list == nil {
		list = []*Session{}
	}
	b, err := json.MarshalIndent(struct {
		Sessions []*Session `json:"sessions"`
	}{Sessions: list}, "", " ")
	if err != nil {
		return err
	}
	tmp := r.Path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, r.Path)
}

// List 返回登记里的全部会话（空 = 空切片，不报错）。
func (r *Registry) List() ([]*Session, error) {
	return r.readAll()
}

// Tags 返回登记里的全部 Tag（给"多候选"的报错与界面列用）。
func (r *Registry) Tags() ([]string, error) {
	list, err := r.readAll()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(list))
	for _, s := range list {
		out = append(out, s.Tag)
	}
	return out, nil
}

// Load 读**唯一**那条登记。
//
// 一条 → 它；多条 → `ErrAmbiguous`（不挑）；一份都没有 → `ErrNoSession`。
// 这条与今天的行为逐字相同（单条时代只有一个分支成立）。
func (r *Registry) Load() (*Session, error) {
	list, err := r.readAll()
	if err != nil {
		return nil, err
	}
	switch len(list) {
	case 0:
		return nil, ErrNoSession
	case 1:
		return list[0], nil
	default:
		tags := make([]string, 0, len(list))
		for _, s := range list {
			tags = append(tags, s.Tag)
		}
		return nil, &ErrAmbiguous{Tags: tags}
	}
}

// LoadTag 读点名的那条登记（tag 空 = 等价 Load）。
func (r *Registry) LoadTag(tag string) (*Session, error) {
	if strings.TrimSpace(tag) == "" {
		return r.Load()
	}
	list, err := r.readAll()
	if err != nil {
		return nil, err
	}
	for _, s := range list {
		if s.Tag == tag {
			return s, nil
		}
	}
	return nil, ErrSessionNotFound
}

// Save 写回一条会话：**按 Tag upsert**（同名替换，否则追加）。
//
// 原子写（临时文件 + rename）：写了一半的登记比没有更糟 —— 收尾与"已经在录"的判断都信它。
func (r *Registry) Save(s *Session) error {
	if s == nil {
		return errors.New("fep: Save(nil)")
	}
	if strings.TrimSpace(s.Tag) == "" {
		return errors.New("fep: 登记缺少 Tag（无法定位是哪一条）")
	}
	list, err := r.readAll()
	if err != nil {
		return err
	}
	replaced := false
	for i, old := range list {
		if old.Tag == s.Tag {
			list[i] = s
			replaced = true
			break
		}
	}
	if !replaced {
		list = append(list, s)
	}
	return r.writeAll(list)
}

// Remove 从登记里摘掉一条（不存在也算成功：收尾必须幂等）。
func (r *Registry) Remove(tag string) error {
	list, err := r.readAll()
	if err != nil {
		return err
	}
	out := make([]*Session, 0, len(list))
	for _, s := range list {
		if s.Tag != tag {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return r.Clear()
	}
	return r.writeAll(out)
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
func (r *Registry) Stop() (*Session, error) { return r.StopTag("") }

// StopTag 收掉点名的那一条（tag 空 = 唯一那条，多条则 ErrAmbiguous）。
//
// 只摘**这一条**：多会话之后，"收掉一条"不许把别人的登记一起清掉。
func (r *Registry) StopTag(tag string) (*Session, error) {
	s, err := r.LoadTag(tag)
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
	if rerr := r.Remove(s.Tag); rerr != nil {
		errs = append(errs, rerr.Error())
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
