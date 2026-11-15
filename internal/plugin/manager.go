package plugin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tt/internal/config"
	"tt/internal/connsvc"
)

// Hooks 让后续阶段（Supervisor / 前端资源 / skills 部署）挂到启用/停用/卸载上。
//
// 为什么用钩子而不是让 Manager 直接依赖那些子系统：阶段 2 只装元数据，不该为了
// "以后要启动服务"就把 supervisor 拉进来。钩子把"状态变了之后还需要做什么"
// 留成注入点，缺省是空实现（阶段 2 的 CLI/Web 行为因此是确定的）。
type Hooks interface {
	// AfterEnable 在启用状态已落盘之后调用；返回错误会把插件回退成未启用并记 LastError。
	AfterEnable(rec *Record) error
	// BeforeDisable 在停用状态落盘之前调用（停服务、撤令牌、关会话）。
	BeforeDisable(rec *Record) error
	// AfterDisable 在停用状态落盘之后调用（重算 skills / 菜单）。
	AfterDisable(rec *Record) error
	// AfterUninstall 在注册项移除之后调用（清理只属于该插件的部署项）。
	AfterUninstall(id string)
}

// Manager 插件管理面：安装/启停/卸载/查询/自检。
//
// 所有状态变更都经 Store.WithLock（CLI 与 Web 同时操作不会互相覆盖）。
type Manager struct {
	Store *Store
	// ReservedNamespaces 内置命令 + 旧别名 + TT 保留名（由 CLI 层从命令树算出）。
	ReservedNamespaces []string
	// HostAPIVersion 默认 connsvc.HostAPIVersion。
	HostAPIVersion string
	// Envs 当前配置里的环境名清单（校验授权时用）。nil = 读 Store 的配置路径。
	Envs  func() ([]string, error)
	Hooks Hooks
	// SkillsSync 状态变更后重算「期望的已分发技能」的注入点。nil = 不做事。
	//
	// 为什么不在这里直接调 internal/skill：插件管理面只管注册表与载荷，而"技能怎么
	// 装出去"（来源合并、冲突、assets 指针、平台清单）只有 internal/skill 一份实现。
	// 命令层在两边之间接线（internal/cli/skills_sync.go），本包因此既不 import 技能
	// 管线，也不把技能规则复制一遍。
	//
	// 纪律（P4-02）：**能从平台目录删除什么，只由 .agents/.tt-manifest.json 决定**。
	// 注册表只说明"哪些技能应当出现"；撤销安装、刷新副本都只动平台清单记过的条目。
	SkillsSync func(reason string) error

	// ConfigPath 供缺省 Envs 实现读配置用。
	ConfigPath string

	// Runtime 升级/回滚期间对常驻服务与实例凭证的操作；nil = 本进程没有运行态。
	//
	// 为什么不放进 Hooks：Hooks 是"状态变了之后顺手做点什么"（可失败、可空实现），
	// 而升级里的"停服务再起服务"是**流程的一步**：它失败就该中止这次升级，不能当成
	// 事后清理。两者语义不同，所以分开注入（见 upgrade.go）。
	Runtime UpdateRuntime
}

// NewManager 构造管理面。
func NewManager(store *Store, reserved []string) *Manager {
	return &Manager{Store: store, ReservedNamespaces: append([]string(nil), reserved...)}
}

// hostAPIVersion 生效的 Host API 版本。
func (m *Manager) hostAPIVersion() string {
	if m.HostAPIVersion != "" {
		return m.HostAPIVersion
	}
	return connsvc.HostAPIVersion
}

// envNames 取环境名清单。
func (m *Manager) envNames() ([]string, error) {
	if m.Envs != nil {
		return m.Envs()
	}
	if m.ConfigPath == "" {
		return nil, nil
	}
	root, err := config.Load(m.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}
	return root.Hosts.Names(), nil
}

// List 列出全部插件（按 ID 稳定排序）。
func (m *Manager) List() ([]Summary, error) {
	reg, err := m.Store.Load()
	if err != nil {
		return nil, err
	}
	out := make([]Summary, 0, len(reg.Plugins))
	for _, rec := range reg.Plugins {
		out = append(out, Summarize(rec))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Paths 一个插件涉及的落点（展示与排障用）。
type Paths struct {
	VersionDir string `json:"versionDir"`
	DataDir    string `json:"dataDir"`
	RunDir     string `json:"runDir"`
	LogsDir    string `json:"logsDir"`
}

// Info 一个插件的完整信息。
type Info struct {
	Summary
	Manifest *Manifest          `json:"manifest,omitempty"`
	Versions []InstalledVersion `json:"versions"`
	Grants   Grants             `json:"grants"`
	Paths    Paths              `json:"paths"`
	// ManifestErr 清单读不出来时的原因（信息仍可展示：路径与版本来自注册表）。
	ManifestErr string `json:"manifestErr,omitempty"`
}

// Info 取一个插件的信息。
func (m *Manager) Info(id string) (*Info, error) {
	reg, err := m.Store.Load()
	if err != nil {
		return nil, err
	}
	rec := reg.Find(id)
	if rec == nil {
		return nil, fmt.Errorf("没有插件 %q（tt plugin list 看已装的）", id)
	}
	info := &Info{
		Summary:  Summarize(rec),
		Versions: append([]InstalledVersion(nil), rec.Versions...),
		Grants:   rec.Grants,
		Paths: Paths{
			VersionDir: m.Store.ActiveDir(rec),
			DataDir:    m.Store.DataDir(rec.ID),
			RunDir:     m.Store.RunDir(rec.ID),
			LogsDir:    m.Store.LogsDir(rec.ID),
		},
	}
	if dir := m.Store.ActiveDir(rec); dir != "" {
		man, merr := LoadManifest(dir)
		if merr != nil {
			info.ManifestErr = merr.Error()
		} else {
			info.Manifest = man
		}
	}
	return info, nil
}

// syncSkills 触发一次期望态重算（未注入回调时是空操作）。
//
// 失败只记 LastError、不许悄悄吞掉：调用方决定这次失败要不要让本次操作报错 ——
// 状态本身已经落盘，技能副本只是"还没跟上"，那是可重试的（tt skills deploy）。
func (m *Manager) syncSkills(id, reason string) error {
	if m.SkillsSync == nil {
		return nil
	}
	if err := m.SkillsSync(reason); err != nil {
		m.recordError(id, "技能同步失败: "+err.Error())
		return err
	}
	return nil
}

// Enable 启用插件并写入授权。
//
// grants 是"这次给它什么权限"，不是"清单里声明了什么"：清单只声明诉求，
// 授权必须由用户显式给出（见 docs 的权限矩阵）。
func (m *Manager) Enable(id string, grants Grants) (*Record, error) {
	reg, err := m.Store.Load()
	if err != nil {
		return nil, err
	}
	rec := reg.Find(id)
	if rec == nil {
		return nil, fmt.Errorf("没有插件 %q", id)
	}
	dir := m.Store.ActiveDir(rec)
	if dir == "" {
		return nil, fmt.Errorf("插件 %q 没有可用的版本目录", id)
	}
	man, err := LoadManifest(dir)
	if err != nil {
		m.recordError(id, err.Error())
		return nil, fmt.Errorf("启用失败：%w", err)
	}
	// 启用前重新校验一遍：版本目录里的文件可能被人删改过。
	if err := ValidateManifest(man, dir, ValidateOptions{
		HostAPIVersion:     m.hostAPIVersion(),
		ReservedNamespaces: m.ReservedNamespaces,
		OtherNamespaces:    reg.NamespacesInUse(id),
		OtherRoutes:        reg.RoutesInUse(id),
	}); err != nil {
		m.recordError(id, err.Error())
		return nil, fmt.Errorf("启用失败：%w", err)
	}
	if err := m.validateGrants(grants, man); err != nil {
		return nil, err
	}

	var out *Record
	err = m.Store.WithLock(func(reg *Registry) error {
		r := reg.Find(id)
		if r == nil {
			return fmt.Errorf("插件 %q 在启用过程中消失（注册表被并发修改）", id)
		}
		r.Enabled = true
		r.Grants = normalGrants(grants)
		r.LastError = ""
		r.UpdatedAt = m.Store.Stamp()
		out = r
		return nil
	})
	if err != nil {
		return nil, err
	}
	if m.Hooks != nil {
		if herr := m.Hooks.AfterEnable(out); herr != nil {
			// 启用后置动作失败 → 回退成未启用，并把原因留在记录上（可诊断）。
			m.recordError(id, herr.Error())
			_ = m.Store.WithLock(func(reg *Registry) error {
				if r := reg.Find(id); r != nil {
					r.Enabled = false
					r.LastError = herr.Error()
				}
				return nil
			})
			return nil, fmt.Errorf("启用失败（已回退为未启用）：%w", herr)
		}
	}
	// 启用成功 → 重算期望态：插件的技能应当出现在已经部署过技能的平台目录里。
	//
	// 这里**不回退启用状态**：插件本身（CLI/页面/服务）已经可用，少的是平台目录里的
	// 技能副本；回退会把一个能用的插件按掉，而用户真正要做的是把技能重装一次。
	if serr := m.syncSkills(id, "enable "+id); serr != nil {
		return out, fmt.Errorf("插件已启用，但技能没有同步：%w（修好后执行 tt skills deploy 重试）", serr)
	}
	// 返回最新状态。
	return m.reloadRecord(id)
}

// Disable 停用插件：先让钩子收口（停服务、撤令牌、关会话），再落状态。
func (m *Manager) Disable(id string) (*Record, error) {
	reg, err := m.Store.Load()
	if err != nil {
		return nil, err
	}
	rec := reg.Find(id)
	if rec == nil {
		return nil, fmt.Errorf("没有插件 %q", id)
	}
	if !rec.Enabled {
		return rec, nil
	}
	if m.Hooks != nil {
		if herr := m.Hooks.BeforeDisable(rec); herr != nil {
			return nil, fmt.Errorf("停止插件失败（未改动启用状态）：%w", herr)
		}
	}
	err = m.Store.WithLock(func(reg *Registry) error {
		r := reg.Find(id)
		if r == nil {
			return fmt.Errorf("插件 %q 在停用过程中消失（注册表被并发修改）", id)
		}
		r.Enabled = false
		r.UpdatedAt = m.Store.Stamp()
		return nil
	})
	if err != nil {
		return nil, err
	}
	out, _ := m.reloadRecord(id)
	if m.Hooks != nil && out != nil {
		if herr := m.Hooks.AfterDisable(out); herr != nil {
			m.recordError(id, herr.Error())
			return out, fmt.Errorf("插件已停用，但清理未完成：%w", herr)
		}
	}
	// 停用成功 → 重算期望态：它的技能要从平台目录撤走（只撤平台清单记过的条目，
	// 只撤"源确实已停用"的那些；见 skill.Withdraw）。
	if serr := m.syncSkills(id, "disable "+id); serr != nil {
		return out, fmt.Errorf("插件已停用，但技能没有同步：%w（修好后执行 tt skills deploy 重试）", serr)
	}
	return out, nil
}

// UninstallOptions 卸载参数。
type UninstallOptions struct {
	// PurgeData 连持久数据一起删。**默认不删** —— 数据是用户的，不是程序的一部分。
	PurgeData bool
}

// Uninstall 卸载插件：移除版本目录与注册项，默认保留数据（与日志）。
func (m *Manager) Uninstall(id string, opt UninstallOptions) error {
	reg, err := m.Store.Load()
	if err != nil {
		return err
	}
	rec := reg.Find(id)
	if rec == nil {
		return fmt.Errorf("没有插件 %q", id)
	}
	if rec.Enabled {
		return fmt.Errorf("插件 %q 正在启用中：先 tt plugin disable %q 再卸载", id, id)
	}

	// 期望态重算放在删载荷**之前**：技能名只能从版本目录里的清单读出来，删完就再也
	// 算不出"该从平台目录撤走哪些副本"。失败时直接中止 —— 此刻一个字节都还没删。
	if serr := m.syncSkills(id, "uninstall "+id); serr != nil {
		return fmt.Errorf("卸载前重算技能失败（没有删除任何内容）：%w", serr)
	}

	var removedVersions []string
	for _, iv := range rec.Versions {
		dir := m.Store.VersionDir(id, iv.Version)
		if dir == "" {
			continue
		}
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("删除版本 %s 失败: %w", iv.Version, err)
		}
		removedVersions = append(removedVersions, iv.Version)
	}
	// 版本父目录空了就一并删掉（只删这个插件自己的目录）。
	if len(rec.Versions) > 0 {
		if dir := m.Store.VersionDir(id, rec.Versions[0].Version); dir != "" {
			_ = os.Remove(filepath.Dir(dir))
		}
	}
	// 运行态目录清掉（里面的东西都是可重建的）；日志与数据保留。
	if dir := m.Store.RunDir(id); dir != "" {
		_ = os.RemoveAll(dir)
	}
	if opt.PurgeData {
		if dir := m.Store.DataDir(id); dir != "" {
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("删除插件数据失败: %w", err)
			}
		}
	}

	if err := m.Store.WithLock(func(reg *Registry) error {
		if !reg.Remove(id) {
			return fmt.Errorf("插件 %q 在卸载过程中消失（注册表被并发修改）", id)
		}
		return nil
	}); err != nil {
		return err
	}
	if m.Hooks != nil {
		m.Hooks.AfterUninstall(id)
	}
	return nil
}

// Grant 修改授权（不改变启用状态）。
func (m *Manager) Grant(id string, grants Grants) (*Record, error) {
	reg, err := m.Store.Load()
	if err != nil {
		return nil, err
	}
	rec := reg.Find(id)
	if rec == nil {
		return nil, fmt.Errorf("没有插件 %q", id)
	}
	man, merr := LoadManifest(m.Store.ActiveDir(rec))
	if merr != nil {
		return nil, fmt.Errorf("读清单失败，无法校验授权: %w", merr)
	}
	if err := m.validateGrants(grants, man); err != nil {
		return nil, err
	}
	if err := m.Store.WithLock(func(reg *Registry) error {
		r := reg.Find(id)
		if r == nil {
			return fmt.Errorf("插件 %q 不存在", id)
		}
		r.Grants = normalGrants(grants)
		r.UpdatedAt = m.Store.Stamp()
		return nil
	}); err != nil {
		return nil, err
	}
	return m.reloadRecord(id)
}

// validateGrants 校验授权：能力名合法 + 环境真实存在。
//
// 环境写成拼错的名字必须报错而不是"授权了个不存在的环境然后一切照旧被拒"：
// 后者在界面上看起来是"授权了但不生效"。
func (m *Manager) validateGrants(g Grants, man *Manifest) error {
	var ps Problems
	for _, c := range g.Capabilities {
		if err := ValidatePermission(c); err != nil {
			ps.addf("%v", err)
		}
	}
	if len(g.Envs) > 0 {
		known, err := m.envNames()
		if err != nil {
			return err
		}
		have := map[string]bool{}
		for _, e := range known {
			have[e] = true
		}
		for _, e := range g.Envs {
			if !have[e] {
				ps.addf("环境 %q 在配置里不存在（可用: %s）", e, strings.Join(known, "、"))
			}
		}
	}
	if len(g.Envs) > 0 && len(g.Capabilities) == 0 {
		ps.addf("授予了环境但没有任何能力（profiles.list / db.connect / ssh.connect）：环境授权不会单独生效")
	}
	// 声明了却没授予的能力只是提示，不当错误：用户可以先启用、后授权。
	_ = man
	return ps.Err()
}

// normalGrants 归一化授权清单（去重 + 排序，写盘稳定）。
func normalGrants(g Grants) Grants {
	return Grants{
		Capabilities: SortedUnique(g.Capabilities),
		Envs:         SortedUnique(g.Envs),
	}
}

// recordError 把最近一次失败记住（只影响展示与诊断，不改启用状态）。
func (m *Manager) recordError(id, msg string) {
	_ = m.Store.WithLock(func(reg *Registry) error {
		if r := reg.Find(id); r != nil {
			r.LastError = msg
		}
		return nil
	})
}

// reloadRecord 重新读一条记录的最新状态。
func (m *Manager) reloadRecord(id string) (*Record, error) {
	reg, err := m.Store.Load()
	if err != nil {
		return nil, err
	}
	rec := reg.Find(id)
	if rec == nil {
		return nil, fmt.Errorf("插件 %q 不存在", id)
	}
	return rec, nil
}

// Issue 自检发现的一条问题。
type Issue struct {
	Severity string `json:"severity"` // error | warning
	Code     string `json:"code"`
	Message  string `json:"message"`
	Hint     string `json:"hint,omitempty"`
}

// Doctor 自检一个插件（不给 id 时检查全部）。
//
// 覆盖：注册表记录与磁盘是否一致、清单是否可读可校验、授权是否有效、
// 启用但有过错误的状态。**不启动插件、不连网**。
func (m *Manager) Doctor(id string) ([]Issue, error) {
	reg, err := m.Store.Load()
	if err != nil {
		return nil, err
	}
	var issues []Issue
	targets := reg.Plugins
	if id != "" {
		rec := reg.Find(id)
		if rec == nil {
			return nil, fmt.Errorf("没有插件 %q", id)
		}
		targets = []*Record{rec}
	}
	for _, rec := range targets {
		issues = append(issues, m.doctorOne(reg, rec)...)
	}
	return issues, nil
}

// doctorOne 单个插件的检查项。
//
// 覆盖面（P6-07）：清单与 Host API 兼容、UI 入口、技能来源与同名冲突、服务运行态与
// run/<id>/instance.json 是否一致、多出来的旧版本。**只读本地事实**：不启动插件、不连网。
func (m *Manager) doctorOne(reg *Registry, rec *Record) []Issue {
	var out []Issue
	add := func(sev, code, msg, hint string) {
		out = append(out, Issue{Severity: sev, Code: code, Message: msg, Hint: hint})
	}

	if rec.ActiveVersion == "" {
		add("error", "NO_ACTIVE_VERSION", fmt.Sprintf("%s 没有活跃版本", rec.ID), "重新安装该插件")
		return out
	}
	dir := m.Store.ActiveDir(rec)
	if dir == "" {
		add("error", "NO_ACTIVE_VERSION", fmt.Sprintf("%s 的活跃版本不能用作目录名: %q", rec.ID, rec.ActiveVersion),
			"卸载后重新安装")
		return out
	}
	if _, err := os.Stat(dir); err != nil {
		add("error", "VERSION_DIR_MISSING", fmt.Sprintf("%s 的活跃版本目录不存在: %s", rec.ID, dir),
			"重新安装该版本；若只是数据目录被移动过，请把 plugins/ 整个目录搬回来")
		return out
	}
	man, merr := LoadManifest(dir)
	if merr != nil {
		add("error", "MANIFEST_BROKEN", fmt.Sprintf("%s 的清单读不出来: %v", rec.ID, merr), "重新安装该插件")
		return out
	}

	// Host API 兼容性单列一条：它是最常见、也最容易被笼统的"清单不合法"盖住的原因
	// （TT 升了大版本之后老插件就是这条）。
	if merr := ValidateHostAPIRange(man.HostAPI); merr != nil {
		add("error", "HOSTAPI_INCOMPATIBLE", fmt.Sprintf("%s 的 hostApi %q 不是合法范围: %v", rec.ID, man.HostAPI, merr),
			"重新安装该插件")
	} else if ok, herr := HostAPISatisfies(man.HostAPI, m.hostAPIVersion()); herr != nil || !ok {
		add("error", "HOSTAPI_INCOMPATIBLE", fmt.Sprintf("%s 要求 hostApi %q，当前 TT 的 Host API 是 %s",
			rec.ID, man.HostAPI, m.hostAPIVersion()),
			"升级该插件（tt plugin update），或等插件作者发布兼容版本")
	}

	if verr := ValidateManifest(man, dir, ValidateOptions{
		HostAPIVersion:     m.hostAPIVersion(),
		ReservedNamespaces: m.ReservedNamespaces,
		OtherNamespaces:    reg.NamespacesInUse(rec.ID),
		OtherRoutes:        reg.RoutesInUse(rec.ID),
	}); verr != nil {
		// hostApi 已经单列一条：从这条里剔掉，别把同一件事报两次。但它不该顺带把
		// 其它问题（命名空间、入口路径、技能目录）一起藏起来。
		if rest := nonHostAPIProblems(verr); len(rest) > 0 {
			add("error", "MANIFEST_INVALID", fmt.Sprintf("%s 的清单不再合法: %s", rec.ID, strings.Join(rest, "；")),
				"重新安装该插件")
		}
	}

	// UI 入口：路径在清单里合法，不代表文件还在（版本目录是可读可改的普通目录）。
	if ferr := checkUIEntry(man, dir); ferr != nil {
		add("error", "UI_ENTRY_MISSING", fmt.Sprintf("%s 的页面入口不可用: %v", rec.ID, ferr),
			"重新安装该插件（版本目录里的文件被删改过）")
	}

	if gerr := m.validateGrants(rec.Grants, man); gerr != nil {
		add("warning", "GRANTS_INVALID", fmt.Sprintf("%s 的授权不再有效: %v", rec.ID, gerr),
			"用 tt plugin grant 重新授权")
	}
	if rec.Enabled && rec.LastError != "" {
		add("warning", "LAST_ERROR", fmt.Sprintf("%s 启用期间有过错误: %s", rec.ID, rec.LastError),
			"tt plugin disable 后重试，或 tt plugin logs 看细节")
	}

	out = append(out, m.doctorService(rec, man, add)...)
	out = append(out, m.doctorSkills(reg, rec, man, dir, add)...)
	// 旧版本：保留"当前 + 上一个"是策略，多出来的就是该清的。
	keep := map[string]bool{rec.ActiveVersion: true}
	if rec.PreviousVersion != "" {
		keep[rec.PreviousVersion] = true
	}
	var extras []string
	for _, iv := range rec.Versions {
		if !keep[iv.Version] {
			extras = append(extras, iv.Version)
		}
	}
	if len(extras) > 0 {
		add("warning", "STALE_VERSION", fmt.Sprintf("%s 还有 %d 个不再需要的旧版本: %s",
			rec.ID, len(extras), strings.Join(extras, "、")),
			"下次 tt plugin update 会顺手清理（升级只保留当前与上一个版本）")
	}
	if rec.PreviousVersion != "" && !rec.HasVersion(rec.PreviousVersion) {
		add("warning", "PREVIOUS_VERSION_MISSING", fmt.Sprintf("%s 记录的上一个版本 %s 已不在本机",
			rec.ID, rec.PreviousVersion), "tt plugin rollback 会因此失败；升级时会重新记上一个版本")
	}

	// 未被任何贡献使用的权限只是提示（清单可以预留）。
	for _, p := range rec.Permissions {
		if !rec.Grants.Has(p) && rec.Enabled {
			add("warning", "PERMISSION_NOT_GRANTED", fmt.Sprintf("%s 声明了 %s 但未授予", rec.ID, p),
				"用 tt plugin grant --allow "+p)
		}
	}
	return out
}

// nonHostAPIProblems 从清单校验错误里剔除 hostApi 相关的问题（doctor 单独报过）。
func nonHostAPIProblems(err error) []string {
	var ve *ValidationError
	if errors.As(err, &ve) {
		out := make([]string, 0, len(ve.Problems))
		for _, p := range ve.Problems {
			if strings.Contains(p, "hostApi") || strings.Contains(p, "Host API") {
				continue
			}
			out = append(out, p)
		}
		return out
	}
	return []string{err.Error()}
}

// doctorService 服务运行态 vs run/<id>/instance.json 是否一致。//
// 三种不一致都要能说出来：声明了服务却没在跑、状态文件说在跑但进程没了（残留）、
// 在跑的实例是**上一个版本**（升级没收拾干净 —— 这正是"页面新、服务旧"的现场）。
func (m *Manager) doctorService(rec *Record, man *Manifest, add func(sev, code, msg, hint string)) []Issue {
	if man.Contributions.Service == nil {
		return nil
	}
	st, err := m.Store.ReadInstanceState(rec.ID)
	if err != nil {
		add("warning", "SERVICE_STATE_BROKEN",
			fmt.Sprintf("%s 的运行状态文件读不出来: %v", rec.ID, err),
			"删掉 "+filepath.Join(m.Store.RunDir(rec.ID), "instance.json")+" 后重启 tt serve")
		return nil
	}
	if st == nil || st.State != "running" {
		if rec.Enabled {
			add("warning", "SERVICE_NOT_RUNNING",
				fmt.Sprintf("%s 声明了常驻服务，但没有运行中的实例状态", rec.ID),
				"重启 tt serve 或 tt plugin disable/enable 重试")
		}
		return nil
	}
	if !st.Alive() {
		add("warning", "SERVICE_STALE_STATE",
			fmt.Sprintf("%s 的状态文件说服务在跑（pid=%d），但那个进程已经不在了", rec.ID, st.PID),
			"重启 tt serve（它会重写运行状态）")
		return nil
	}
	if st.Version != "" && st.Version != rec.ActiveVersion {
		add("error", "SERVICE_VERSION_STALE",
			fmt.Sprintf("%s 运行中的服务是版本 %s，活跃版本是 %s（页面/CLI 与服务会串版）",
				rec.ID, st.Version, rec.ActiveVersion),
			"先 tt serve --stop，再 tt plugin update "+rec.ID+"（或 disable/enable）")
	}
	if st.PluginID != "" && st.PluginID != rec.ID {
		add("warning", "SERVICE_STATE_MISMATCH",
			fmt.Sprintf("%s 的运行状态文件记的是插件 %s（目录被移动/覆盖过）", rec.ID, st.PluginID),
			"删掉该 instance.json 后重启 tt serve")
	}
	return nil
}

// doctorSkills 技能来源是否存在、是否与别的插件同名冲突。
//
// 冲突只看插件之间：与官方 skills 的先后顺序由 internal/skill 的"官方优先"规则决定，
// 本包不复制那套判断（见 docs 05/11）。
func (m *Manager) doctorSkills(reg *Registry, rec *Record, man *Manifest, dir string, add func(sev, code, msg, hint string)) []Issue {
	dirs, err := man.SkillDirs()
	if err != nil {
		add("error", "SKILLS_SOURCE_MISSING", fmt.Sprintf("%s 的技能目录不可用: %v", rec.ID, err),
			"重新安装该插件")
		return nil
	}
	for _, rel := range dirs {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if _, serr := os.Stat(filepath.Join(full, "SKILL.md")); serr != nil {
			add("error", "SKILLS_SOURCE_MISSING",
				fmt.Sprintf("%s 声明的技能 %q 缺少 SKILL.md（分发出去会是一份坏技能）", rec.ID, rel),
				"重新安装该插件")
		}
		// 同名 = 同一个技能目录名。插件之间撞名时谁生效取决于登记顺序，必须让人看见。
		name := filepath.Base(filepath.FromSlash(rel))
		for _, other := range reg.Plugins {
			if other.ID == rec.ID {
				continue
			}
			for _, od := range other.Contributions.Skills {
				if strings.EqualFold(filepath.Base(filepath.FromSlash(od)), name) {
					add("warning", "SKILLS_CONFLICT",
						fmt.Sprintf("%s 的技能 %q 与插件 %s 同名", rec.ID, name, other.ID),
						"两者只会有一份生效（先登记优先）：改名或只留一个")
				}
			}
		}
	}
	return nil
}

// FindByNamespace 按 CLI 命名空间找已启用的记录（CLI 转发用）。
func (m *Manager) FindByNamespace(ns string) (*Record, error) {
	reg, err := m.Store.Load()
	if err != nil {
		return nil, err
	}
	for _, rec := range reg.Enabled() {
		if c := rec.Contributions.CLI; c != nil && c.Namespace == ns {
			return rec, nil
		}
	}
	return nil, nil
}

// FindByRoute 按前端路由找已启用的记录（页面资源服务用）。
func (m *Manager) FindByRoute(route string) (*Record, error) {
	reg, err := m.Store.Load()
	if err != nil {
		return nil, err
	}
	for _, rec := range reg.Enabled() {
		if f := rec.Contributions.Frontend; f != nil && f.Route == route {
			return rec, nil
		}
	}
	return nil, nil
}

// errNoManager 缺省管理面缺失时的错误（装配错误）。
var errNoManager = errors.New("插件管理面未装配")
