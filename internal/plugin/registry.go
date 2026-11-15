package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tt/internal/config"
)

// RegistrySchemaVersion 注册表结构版本。更高的主版本一律拒绝读（不猜测字段含义）。
const RegistrySchemaVersion = 1

// 来源种类。
const (
	SourceFile = "file" // 本地 zip 包
	SourceDir  = "dir"  // 本地目录（开发态）
	SourceURL  = "url"  // 远程 https（阶段 6）
)

// Source 插件从哪儿来的。升级/回滚与排障要靠它回答"这个版本是谁装的"。
type Source struct {
	Kind string `json:"kind"`
	Ref  string `json:"ref,omitempty"`
	// SHA256 安装时对包文件算的摘要（目录来源为空）。
	SHA256 string `json:"sha256,omitempty"`
}

func (s Source) String() string {
	if s.Ref == "" {
		return s.Kind
	}
	return s.Kind + ":" + s.Ref
}

// Grants 一个插件获得的授权。
//
// 能力与环境是分开的两件事：能力说"能不能用某类连接"，环境说"能用哪些环境"。
// 二者都要显式授予；只声明贡献永远不会带来能力（见 docs 的权限矩阵）。
type Grants struct {
	Capabilities []string `json:"capabilities,omitempty"`
	Envs         []string `json:"envs,omitempty"`
}

// Has 是否持有某能力。
func (g Grants) Has(capability string) bool {
	for _, c := range g.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

// AllowsEnv 是否被授权使用某环境（只对需要环境的连接能力有意义）。
func (g Grants) AllowsEnv(env string) bool {
	for _, e := range g.Envs {
		if e == env {
			return true
		}
	}
	return false
}

// InstalledVersion 一个已安装的版本（版本目录不可变；这里只记事实）。
type InstalledVersion struct {
	Version     string `json:"version"`
	InstalledAt string `json:"installedAt"`
	Source      Source `json:"source"`
}

// Record 一个插件的注册信息。
//
// 这里存的是**静态元数据 + 状态**：程序在版本目录里，数据在 data 目录里，
// 运行态不落这里（run 目录）。CLI 帮助与前端菜单只读这份记录，不启动插件。
type Record struct {
	ID            string `json:"id"`
	Name          string `json:"name,omitempty"`
	ActiveVersion string `json:"activeVersion"`
	// PreviousVersion 上一次切换前的活跃版本（`tt plugin rollback` 的落点）。
	//
	// 只有"成功切换过版本"的插件才有值；清理策略保证它的程序目录还在（保留当前 + 上一个）。
	// 回滚之后清空：再指着一个刚被证明不可用的版本只会误导人。
	PreviousVersion string             `json:"previousVersion,omitempty"`
	Versions        []InstalledVersion `json:"versions"`
	Enabled         bool               `json:"enabled"`
	Source          Source             `json:"source"`
	// Contributions 生效版本的静态贡献元数据（根帮助/菜单用）。
	Contributions Contributions `json:"contributions"`
	// Permissions 清单声明的权限（≠ 已授予）。
	Permissions []string `json:"permissions,omitempty"`
	Grants      Grants   `json:"grants"`
	InstalledAt string   `json:"installedAt,omitempty"`
	UpdatedAt   string   `json:"updatedAt,omitempty"`
	// LastError 最近一次启用/升级失败的原因（可诊断，不阻断禁用与回滚）。
	LastError string `json:"lastError,omitempty"`
}

// HasVersion 是否已安装某版本。
func (r *Record) HasVersion(v string) bool {
	for _, iv := range r.Versions {
		if iv.Version == v {
			return true
		}
	}
	return false
}

// Registry 注册表（registry.json）。
type Registry struct {
	SchemaVersion int       `json:"schemaVersion"`
	Plugins       []*Record `json:"plugins"`

	// RecoveredFrom 非持久化：本次是从备份恢复的（值为被移走的损坏文件名）。
	RecoveredFrom string `json:"-"`
}

// Find 按 ID 找记录。
func (r *Registry) Find(id string) *Record {
	if r == nil {
		return nil
	}
	for _, p := range r.Plugins {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// Enabled 返回已启用的记录（稳定顺序：按 ID）。
func (r *Registry) Enabled() []*Record {
	if r == nil {
		return nil
	}
	out := make([]*Record, 0, len(r.Plugins))
	for _, p := range r.Plugins {
		if p.Enabled {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Remove 删除记录；不存在返回 false。
func (r *Registry) Remove(id string) bool {
	for i, p := range r.Plugins {
		if p.ID == id {
			r.Plugins = append(r.Plugins[:i], r.Plugins[i+1:]...)
			return true
		}
	}
	return false
}

// Put 插入或替换记录。
func (r *Registry) Put(rec *Record) {
	for i, p := range r.Plugins {
		if p.ID == rec.ID {
			r.Plugins[i] = rec
			return
		}
	}
	r.Plugins = append(r.Plugins, rec)
}

// NamespacesInUse 已启用插件占用的 CLI 命名空间（excludeID 用于校验自身）。
func (r *Registry) NamespacesInUse(excludeID string) []string {
	var out []string
	for _, p := range r.Plugins {
		if !p.Enabled || p.ID == excludeID || p.Contributions.CLI == nil {
			continue
		}
		out = append(out, p.Contributions.CLI.Namespace)
	}
	return out
}

// RoutesInUse 已启用插件占用的前端路由。
func (r *Registry) RoutesInUse(excludeID string) []string {
	var out []string
	for _, p := range r.Plugins {
		if !p.Enabled || p.ID == excludeID || p.Contributions.Frontend == nil {
			continue
		}
		out = append(out, p.Contributions.Frontend.Route)
	}
	return out
}

// Store 注册表与插件目录的读写门面。
//
// 所有落点都由 config.Locations 派生（见 internal/config/plugins.go），
// 本包不自己拼第二套路径。
type Store struct {
	loc *config.Locations
	now func() time.Time
	// lockWait 拿注册表锁的最长等待（测试里调小）。
	lockWait time.Duration
}

// NewStore 用统一路径构造。
func NewStore(loc *config.Locations) *Store {
	return &Store{loc: loc, now: time.Now, lockWait: 2 * time.Second}
}

// Locations 返回底层落点（CLI/Web 拼展示路径用）。
func (s *Store) Locations() *config.Locations { return s.loc }

// RegistryPath 注册表文件路径。
func (s *Store) RegistryPath() string { return s.loc.PluginRegistryPath() }

// DataDir 某插件的持久数据目录。
func (s *Store) DataDir(id string) string { return s.loc.PluginDataDir(id) }

// RunDir 某插件的运行态目录。
func (s *Store) RunDir(id string) string { return s.loc.PluginRunDir(id) }

// LogsDir 某插件的日志目录。
func (s *Store) LogsDir(id string) string { return s.loc.PluginLogsDir(id) }

// VersionDir 某插件某版本的载荷目录。
func (s *Store) VersionDir(id, version string) string {
	return s.loc.PluginVersionDir(id, version)
}

// ActiveDir 某插件当前活跃版本的目录；未启用/无版本时为空。
func (s *Store) ActiveDir(rec *Record) string {
	if rec == nil || rec.ActiveVersion == "" {
		return ""
	}
	return s.VersionDir(rec.ID, rec.ActiveVersion)
}

// CacheDir 安装缓存目录（下载与解包暂存）。
func (s *Store) CacheDir() string { return s.loc.PluginCacheDir() }

// Load 读注册表。文件不存在 = 空注册表（没装过插件不是错误）。
//
// 损坏时按这个顺序自救：registry.json.bak 能读 → 用备份并把损坏文件留档；
// 备份也读不出来 → 返回错误（不静默当成"没装插件"：那会让所有插件突然消失，
// 而用户看到的是"插件都没了"，比一条明确的报错难查得多）。
func (s *Store) Load() (*Registry, error) {
	path := s.RegistryPath()
	if path == "" {
		return nil, errors.New("数据目录未确定，无法定位插件注册表")
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Registry{SchemaVersion: RegistrySchemaVersion}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("读取插件注册表失败: %w", err)
	}
	reg, derr := decodeRegistry(b)
	if derr == nil {
		return reg, nil
	}

	// 损坏：留档 + 尝试备份。
	aside := fmt.Sprintf("%s.corrupt-%s", path, s.now().Format("20060102-150405"))
	_ = os.Rename(path, aside)
	if bak, berr := os.ReadFile(path + ".bak"); berr == nil {
		if reg, derr2 := decodeRegistry(bak); derr2 == nil {
			reg.RecoveredFrom = filepath.Base(aside)
			return reg, nil
		}
	}
	return nil, fmt.Errorf("插件注册表损坏（%v）；已留档到 %s，备份也不可用。修好或删除该文件后重试",
		derr, filepath.Base(aside))
}

// decodeRegistry 严格解码并检查版本。
func decodeRegistry(b []byte) (*Registry, error) {
	var reg Registry
	if err := json.Unmarshal(b, &reg); err != nil {
		return nil, fmt.Errorf("不是合法 JSON: %w", err)
	}
	if reg.SchemaVersion > RegistrySchemaVersion {
		return nil, fmt.Errorf("注册表 schemaVersion = %d，本版 TT 只支持 %d（请升级 tt）",
			reg.SchemaVersion, RegistrySchemaVersion)
	}
	if reg.SchemaVersion == 0 {
		// 空文件/手写的最小注册表：补当前版本，其余按原样用。
		reg.SchemaVersion = RegistrySchemaVersion
	}
	for _, p := range reg.Plugins {
		if p == nil || p.ID == "" {
			return nil, errors.New("注册表含没有 id 的条目")
		}
	}
	return &reg, nil
}

// Save 原子写注册表：临时文件 + fsync + rename，并把上一份留成 .bak。
func (s *Store) Save(reg *Registry) error {
	path := s.RegistryPath()
	if path == "" {
		return errors.New("数据目录未确定，无法写插件注册表")
	}
	if reg.SchemaVersion == 0 {
		reg.SchemaVersion = RegistrySchemaVersion
	}
	// 稳定输出：按 ID 排序，人的 diff 才有意义。
	sort.Slice(reg.Plugins, func(i, j int) bool { return reg.Plugins[i].ID < reg.Plugins[j].ID })
	b, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化注册表失败: %w", err)
	}
	b = append(b, '\n')

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建插件目录失败: %w", err)
	}
	// 备份上一份（读不到就跳过：第一次写没有"上一份"）。
	if old, rerr := os.ReadFile(path); rerr == nil && len(old) > 0 {
		_ = os.WriteFile(path+".bak", old, 0o644)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "registry.json.tmp-*")
	if err != nil {
		return fmt.Errorf("创建注册表临时文件失败: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // rename 成功后这条是空操作
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return fmt.Errorf("写注册表失败: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("同步注册表失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("关闭注册表临时文件失败: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("替换注册表失败: %w", err)
	}
	return nil
}

// WithLock 在注册表锁内做一次"读-改-写"。
//
// 为什么需要锁：CLI 与 Web 是两个入口，可以同时对同一个注册表动手。
// 没有锁时典型后果是"刚装的插件被另一侧的旧快照覆盖掉"，而两边都报成功。
//
// 锁是**目录内的独占创建**（O_CREATE|O_EXCL），不依赖平台特有的文件锁；
// 陈旧锁（超过 staleLockAge）会被清掉并继续 —— 否则一次崩溃会把插件管理永久钉死。
func (s *Store) WithLock(fn func(*Registry) error) error {
	path := s.RegistryPath()
	if path == "" {
		return errors.New("数据目录未确定，无法锁定插件注册表")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建插件目录失败: %w", err)
	}
	lockPath := path + ".lock"
	deadline := s.now().Add(s.lockWait)
	for {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			fmt.Fprintf(f, "pid=%d at=%s\n", os.Getpid(), s.now().Format(time.RFC3339Nano))
			f.Close()
			break
		}
		if !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("创建注册表锁失败: %w", err)
		}
		// 陈旧锁：超过 staleLockAge 就清掉再试。
		if st, serr := os.Stat(lockPath); serr == nil && s.now().Sub(st.ModTime()) > staleLockAge {
			_ = os.Remove(lockPath)
			continue
		}
		if s.now().After(deadline) {
			return fmt.Errorf("插件注册表正被另一个 tt 进程占用（锁文件 %s）：稍后重试", lockPath)
		}
		time.Sleep(20 * time.Millisecond)
	}
	defer os.Remove(lockPath)

	reg, err := s.Load()
	if err != nil {
		return err
	}
	if err := fn(reg); err != nil {
		return err
	}
	return s.Save(reg)
}

// staleLockAge 多久算"上一把锁的主人已经不在了"。
const staleLockAge = 2 * time.Minute

// Stamp 返回 RFC3339 时间戳（注册表里的时间字段只有这一个格式）。
func (s *Store) Stamp() string { return s.now().Format(time.RFC3339) }

// EnsureDirs 建好插件根下的标准目录（安装/启用前调用）。
func (s *Store) EnsureDirs() error {
	for _, d := range []string{
		s.loc.PluginsDir(),
		s.loc.PluginInstalledRoot(),
		filepath.Join(s.loc.PluginsDir(), "data"),
		filepath.Join(s.loc.PluginsDir(), "run"),
		filepath.Join(s.loc.PluginsDir(), "logs"),
		s.loc.PluginCacheDir(),
	} {
		if d == "" {
			continue
		}
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("创建插件目录失败 (%s): %w", d, err)
		}
	}
	return nil
}

// Summary 一条用于展示/帮助的记录投影（CLI 帮助与 Web 菜单都读它）。
type Summary struct {
	ID              string   `json:"id"`
	Name            string   `json:"name,omitempty"`
	Version         string   `json:"version"`
	PreviousVersion string   `json:"previousVersion,omitempty"`
	Enabled         bool     `json:"enabled"`
	Namespace       string   `json:"namespace,omitempty"`
	CLISummary      string   `json:"cliSummary,omitempty"`
	Route           string   `json:"route,omitempty"`
	RouteTitle      string   `json:"routeTitle,omitempty"`
	HasService      bool     `json:"hasService"`
	SkillDirs       []string `json:"skillDirs,omitempty"`
	Permissions     []string `json:"permissions,omitempty"`
	Capabilities    []string `json:"capabilities,omitempty"`
	Envs            []string `json:"envs,omitempty"`
	LastError       string   `json:"lastError,omitempty"`

	// 运行态：只有 Web 管理面在 supervisor 存在时才填（见 internal/web/plugins.go）。
	// CLI 不填 —— `tt plugin list` 不该为了一行状态去碰 run 目录和进程表。
	State    string `json:"state,omitempty"`
	Addr     string `json:"addr,omitempty"`
	PID      int    `json:"pid,omitempty"`
	Restarts int    `json:"restarts,omitempty"`
}

// Summarize 把记录压成展示投影。
func Summarize(rec *Record) Summary {
	s := Summary{
		ID: rec.ID, Name: rec.Name, Version: rec.ActiveVersion, Enabled: rec.Enabled,
		PreviousVersion: rec.PreviousVersion,
		Permissions:     append([]string(nil), rec.Permissions...),
		Capabilities:    append([]string(nil), rec.Grants.Capabilities...),
		Envs:            append([]string(nil), rec.Grants.Envs...),
		LastError:       rec.LastError,
		HasService:      rec.Contributions.Service != nil,
	}
	if c := rec.Contributions.CLI; c != nil {
		s.Namespace, s.CLISummary = c.Namespace, c.Summary
	}
	if f := rec.Contributions.Frontend; f != nil {
		s.Route, s.RouteTitle = f.Route, f.Title
	}
	s.SkillDirs = append([]string(nil), rec.Contributions.Skills...)
	return s
}

// ContributionSummaryString 用一句话描述贡献面（CLI/服务/页面/技能）。
func ContributionSummaryString(c Contributions) string {
	var parts []string
	if c.CLI != nil {
		parts = append(parts, "cli:"+c.CLI.Namespace)
	}
	if c.Service != nil {
		parts = append(parts, "service")
	}
	if c.Frontend != nil {
		parts = append(parts, "ui:"+c.Frontend.Route)
	}
	if len(c.Skills) > 0 {
		parts = append(parts, fmt.Sprintf("skills×%d", len(c.Skills)))
	}
	if len(parts) == 0 {
		return "（无贡献）"
	}
	return strings.Join(parts, " ")
}
