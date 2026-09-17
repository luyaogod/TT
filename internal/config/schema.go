package config

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"tt/internal/dbconfig"
)

// SchemaVersion 当前配置结构版本。0/缺失 = 合并前的旧结构，需要迁移。
const SchemaVersion = 2

// 缺省值。合并前三处各自硬编码，现在只有这一份。
const (
	DefaultListen          = "127.0.0.1:28670"
	DefaultLaunchArgs      = "BBDL512840855a 2 12345 'N' {prog}"
	DefaultWatchdogSeconds = 1800
	DefaultTermWidth       = 200
	DefaultTermHeight      = 50
	DefaultPrintElements   = 1000
	DefaultSSHPort         = 22
	DefaultWorkspaceSuffix = "-ws"
)

// ---------- 环境模型 ----------

// SSHConfig 远程服务器连接配置
type SSHConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// Addr 返回 SSH 地址 host:port
func (c *SSHConfig) Addr() string { return c.Host + ":" + strconv.Itoa(c.Port) }

// EntValue 企业编号(TOPENT)：数字或文本均可，兼容 JSON 数字。
// 需真实编号的场景用 Int()（非数字返回 false）。
type EntValue string

// UnmarshalJSON 同时接受 JSON 数字与字符串
func (e *EntValue) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" {
		*e = ""
		return nil
	}
	*e = EntValue(strings.Trim(s, `"`))
	return nil
}

// MarshalJSON 统一序列化为字符串
func (e EntValue) MarshalJSON() ([]byte, error) { return json.Marshal(string(e)) }

// Int 解析为数字（供数据库探测按企业编号匹配；非数字内容返回 false）
func (e EntValue) Int() (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(string(e)))
	return n, err == nil
}

// NamedSsh 服务器环境：SSH 连接 + 登录区域 + 默认企业(TOPENT)。
//
// 合并自 TDebug/host.NamedSsh 与 TDictCli/host.NamedSsh；TDebug 多出的
// LaunchArgs / WatchdogSeconds 是该环境的调试参数覆盖，omitempty 不影响 TDict 侧。
//
// 这是三个工具共用的唯一环境登记处：原来 tdebug 放在 debug.sshs、
// tdict 放在 hosts.sshs，schema 完全一致，合并后只留 hosts.sshs。
type NamedSsh struct {
	Name            string               `json:"name"`
	SSHConfig                            // 匿名嵌入：host/port/user/password 提升到 ssh 层
	Zone            string               `json:"zone,omitempty"`            // 登录后区域菜单代码：31开发 35测试 36正式 39PATCH t出货
	Topent          EntValue             `json:"topent,omitempty"`          // 默认企业编号(TOPENT)；连接会话即下发，会话内可覆盖
	LaunchArgs      string               `json:"launchArgs,omitempty"`      // 该环境的作业启动参数覆盖（空=用 debug.launchArgs）
	WatchdogSeconds int                  `json:"watchdogSeconds,omitempty"` // 该环境的停站超时覆盖（0=用 debug.watchdogSeconds）
	DB              *dbconfig.Connection `json:"db,omitempty"`              // 该环境的数据库连接（与 SSH 一对一）
}

// Hosts 服务器环境清单：config.json 顶层 hosts 节。
type Hosts struct {
	ActiveEnv string     `json:"activeEnv"`
	SSHs      []NamedSsh `json:"sshs"`
}

// ByName 按名取环境；name 为空取 activeEnv；未命中返回 nil
func (h *Hosts) ByName(name string) *NamedSsh {
	target := name
	if target == "" {
		target = h.ActiveEnv
	}
	if target == "" && len(h.SSHs) > 0 {
		target = h.SSHs[0].Name
	}
	for i := range h.SSHs {
		if h.SSHs[i].Name == target {
			return &h.SSHs[i]
		}
	}
	return nil
}

// Names 返回全部环境名（顺序与配置一致）。
func (h *Hosts) Names() []string {
	out := make([]string, 0, len(h.SSHs))
	for i := range h.SSHs {
		out = append(out, h.SSHs[i].Name)
	}
	return out
}

// Normalize 补全缺省值：activeEnv 缺失时取首条，端口缺失时取 22。
func (h *Hosts) Normalize() {
	for i := range h.SSHs {
		if h.SSHs[i].Port == 0 {
			h.SSHs[i].Port = DefaultSSHPort
		}
	}
	if h.ActiveEnv == "" && len(h.SSHs) > 0 {
		h.ActiveEnv = h.SSHs[0].Name
	}
}

// ---------- 工具设置节 ----------

// DebugSettings 调试工具自有设置（原 TDebug 的 debug 节，减去 sshs）。
// 「当前环境」是运行时概念，不落盘：由会话选择/切换时确定，无会话时取 hosts.activeEnv。
type DebugSettings struct {
	// ActiveEnv 该工具的环境覆盖；留空即继承 hosts.activeEnv。
	ActiveEnv       string `json:"activeEnv,omitempty"`
	LaunchArgs      string `json:"launchArgs,omitempty"`      // 作业启动参数默认模板，{prog} 替换为作业名（环境可覆盖）
	WatchdogSeconds int    `json:"watchdogSeconds,omitempty"` // 停站停留超时默认（秒）；环境可覆盖
	FGLServer       string `json:"fglserver,omitempty"`       // 留空使用 T100 按 SSH 来源 IP 自动设置
	TermWidth       int    `json:"termWidth,omitempty"`
	TermHeight      int    `json:"termHeight,omitempty"`
	PrintElements   int    `json:"printElements,omitempty"`      // fgldb 单次 print 的数组元素上限；0=默认 1000
	PersistBPs      *bool  `json:"persistBreakpoints,omitempty"` // 断点持久化开关（nil 视为 true）
}

// FillDefaults 补全调试设置缺省值（不写盘，仅在内存里生效）。
func (d *DebugSettings) FillDefaults() {
	if d.LaunchArgs == "" {
		d.LaunchArgs = DefaultLaunchArgs
	}
	if d.WatchdogSeconds == 0 {
		d.WatchdogSeconds = DefaultWatchdogSeconds
	}
	if d.TermWidth == 0 {
		d.TermWidth = DefaultTermWidth
	}
	if d.TermHeight == 0 {
		d.TermHeight = DefaultTermHeight
	}
	if d.PrintElements == 0 {
		d.PrintElements = DefaultPrintElements
	}
}

// BPsPersisted 断点持久化是否启用
func (d *DebugSettings) BPsPersisted(dataDir string) bool {
	return dataDir != "" && (d.PersistBPs == nil || *d.PersistBPs)
}

// QuerySettings 字典查询的数据源选择。
//
// Source: "" 或 "auto"（缺省）= **在线优先**：用默认环境（hosts.activeEnv）的远程库直查，
// 一个环境都没配时才用本地 SQLite 镜像；"local" = 固定用本地镜像；<环境名> = 固定用该环境。
//
// 注意合并前 TDictCli 的 README 写着"默认是本地 SQLite 镜像"，与它自己的代码相反
// （cli/source.go 的注释写的是"缺省 = **在线**"）。这里以代码为准。
//
// 也**不做自动降级**：选了在线就是在线，连不上直接报错并提示怎么切，不静默回落本地 ——
// 静默回落会让用户以为查到的是实时数据，实际是几天前的镜像。
type QuerySettings struct {
	Source string `json:"source,omitempty"`
}

// MirrorSettings 源码镜像落点（绝对路径）。
type MirrorSettings struct {
	Dir string `json:"dir,omitempty"`
}

// BdldocSettings BDL 文档落点（绝对路径）。
type BdldocSettings struct {
	Dir string `json:"dir,omitempty"`
}

// SyncSettings 字典同步的本地 SQLite 落点（绝对路径）。
type SyncSettings struct {
	Target string `json:"target,omitempty"`
}

// TdevSettings 设计器包工具设置。原 TDev 完全没有配置文件，这是新增节；
// 只放跨调用稳定的默认值，命令行 flag 仍然优先。
type TdevSettings struct {
	WorkspaceSuffix string `json:"workspaceSuffix,omitempty"` // 导出工作区默认后缀（缺省 -ws）
	DefaultOut      string `json:"defaultOut,omitempty"`      // tzs 解压等命令的默认输出目录
}

// WorkspaceSuffixOrDefault 返回生效的工作区后缀。
func (t *TdevSettings) WorkspaceSuffixOrDefault() string {
	if t.WorkspaceSuffix != "" {
		return t.WorkspaceSuffix
	}
	return DefaultWorkspaceSuffix
}

// ---------- 配置根 ----------

// Root 是 config.json 的类型化视图。
//
// 持久化仍以 map（见 cfgfile.go）为准 —— Root 只是读侧的一层类型化收敛，
// 写侧一律走 EditSection，保证未知顶层键与未知节内键都不会被悄悄抹掉。
type Root struct {
	SchemaVersion int
	Hosts         Hosts
	Listen        string
	Debug         DebugSettings
	Query         QuerySettings
	Mirror        MirrorSettings
	Bdldoc        BdldocSettings
	Sync          SyncSettings
	Tdev          TdevSettings
}

// Load 读取并解析配置，返回类型化视图。
// 文件不存在时返回一份填好缺省值的空配置（首次运行没有文件不是错误）。
func Load(path string) (*Root, error) {
	root, err := Open(path)
	if err != nil {
		return nil, err
	}
	r := &Root{}
	r.SchemaVersion = intOf(root["schemaVersion"])
	if err := decodeSection(root, "hosts", &r.Hosts); err != nil {
		return nil, err
	}
	r.Hosts.Normalize()
	r.Listen = stringOf(root["listen"])
	if r.Listen == "" {
		// 旧结构把监听地址放在 debug.listen 里
		var legacy struct {
			Listen string `json:"listen"`
		}
		if err := decodeSection(root, "debug", &legacy); err == nil {
			r.Listen = legacy.Listen
		}
	}
	if r.Listen == "" {
		r.Listen = DefaultListen
	}
	if err := decodeSection(root, "debug", &r.Debug); err != nil {
		return nil, err
	}
	if err := decodeSection(root, "query", &r.Query); err != nil {
		return nil, err
	}
	if err := decodeSection(root, "mirror", &r.Mirror); err != nil {
		return nil, err
	}
	if err := decodeSection(root, "bdldoc", &r.Bdldoc); err != nil {
		return nil, err
	}
	if err := decodeSection(root, "sync", &r.Sync); err != nil {
		return nil, err
	}
	if err := decodeSection(root, "tdev", &r.Tdev); err != nil {
		return nil, err
	}
	return r, nil
}

// LoadHosts 读取配置文件并返回环境清单，要求至少配置了一个环境
// （CLI 命令都按"有环境可连"的前提工作）。activeEnv 缺省取首条。
//
// 合并前 TDictCli/host.LoadHosts 的替代：那时它还要兼容顶层 debug 键，
// 现在兼容性由 migrate.go 一次性做掉，读路径只认 hosts。
func LoadHosts(path string) (*Hosts, error) {
	r, err := Load(path)
	if err != nil {
		return nil, err
	}
	h := r.Hosts
	if len(h.SSHs) == 0 {
		return nil, fmt.Errorf("尚未配置 SSH 环境:请运行 tt serve 打开配置页添加，或编辑 config.json 的 hosts.sshs")
	}
	return &h, nil
}

// ActiveHost 返回生效环境：工具的覆盖优先，其次 hosts.activeEnv，最后首条。
func (r *Root) ActiveHost(tool string) *NamedSsh {
	switch tool {
	case "debug":
		if r.Debug.ActiveEnv != "" {
			if h := r.Hosts.ByName(r.Debug.ActiveEnv); h != nil {
				return h
			}
		}
	}
	return r.Hosts.ByName("")
}

// ---------- 读写辅助 ----------

// EditSection 只改一个顶层节，其余键（含本节的未知子键）原样保留。
// mutate 收到的 sec 一定是非 nil 的 map，可直接改。
func EditSection(path, key string, validate func(root map[string]any) error, mutate func(sec map[string]any) error) error {
	return Edit(path, validate, func(root map[string]any) error {
		if mutate == nil {
			return nil
		}
		return mutate(SectionOrEmpty(root, key))
	})
}

// SaveHosts 用类型化的环境清单整体替换 hosts 节（其余顶层节不动）。
func SaveHosts(path string, h Hosts) error {
	return Edit(path, nil, func(root map[string]any) error {
		sec, err := toMap(h)
		if err != nil {
			return err
		}
		root["hosts"] = sec
		ensureSchemaVersion(root)
		return nil
	})
}

// EditHosts 读-改-写环境清单：f 拿到当前清单（已补缺省），改完整体写回。
// 环境页的保存走它。
func EditHosts(path string, f func(h *Hosts) error) error {
	return Edit(path, nil, func(root map[string]any) error {
		var h Hosts
		if err := decodeSection(root, "hosts", &h); err != nil {
			return err
		}
		h.Normalize()
		if err := f(&h); err != nil {
			return err
		}
		sec, err := toMap(h)
		if err != nil {
			return err
		}
		root["hosts"] = sec
		ensureSchemaVersion(root)
		return nil
	})
}

// SetListen 写统一监听地址。
func SetListen(path, listen string) error {
	return Edit(path, nil, func(root map[string]any) error {
		root["listen"] = listen
		ensureSchemaVersion(root)
		return nil
	})
}

// ensureSchemaVersion 打上当前结构版本号。
func ensureSchemaVersion(root map[string]any) {
	root["schemaVersion"] = SchemaVersion
}

// ---------- map ↔ 类型转换 ----------

// decodeSection 把 root[key] 反序列化到 v；节缺失时把 v 留作零值（不是错误），
// 节存在但不是对象时报错。
func decodeSection(root map[string]any, key string, v any) error {
	raw, ok := root[key]
	if !ok || raw == nil {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("配置节 %q 无法序列化: %w", key, err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("配置节 %q 格式不正确: %v", key, err)
	}
	return nil
}

// toMap 把类型化结构转成可持久化的 map（经由 JSON 标签，保留 omitempty 语义）。
func toMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

// intOf 从 JSON 解出来的值取整数（数字/字符串皆可），取不到返回 0。
func intOf(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		i, _ := strconv.Atoi(n)
		return i
	}
	return 0
}

// stringOf 从 JSON 解出来的值取字符串，取不到返回空串。
func stringOf(v any) string {
	s, _ := v.(string)
	return s
}
