package plugin

import "github.com/luyaogod/TT/internal/connsvc"

// RegistryAuthorizer 用插件注册表做授权来源（connsvc.Authorizer 的实现）。
//
// v2 起采用"安装并启用即信任"：插件一旦被启用，就可以使用 Host API 的全部能力
// 与配置里的全部环境。理由见计划 —— TT 提供资源与工具，插件在自己的进程里本来
// 就能读配置、连数据库；逐项授权只会制造"能装不能跑"的假安全感，而权限边界由
// 操作系统与用户对插件的信任决定，不由清单字段决定。
//
// 仍然保留的判定是**生命周期**：未启用/未安装的插件不能调用 Host API。
type RegistryAuthorizer struct {
	Store *Store
}

// Permitted 已启用的插件放行全部能力；内置调用方始终放行。
func (a RegistryAuthorizer) Permitted(caller connsvc.Caller, _ connsvc.Capability) bool {
	if caller.IsBuiltin() {
		return true
	}
	return a.enabled(caller.PluginID) != nil
}

// AllowedEnvs 已启用的插件可用全部环境（nil, true）。
func (a RegistryAuthorizer) AllowedEnvs(caller connsvc.Caller, _ connsvc.Capability) ([]string, bool) {
	if caller.IsBuiltin() {
		return nil, true
	}
	if a.enabled(caller.PluginID) == nil {
		return nil, false
	}
	return nil, true
}

// enabled 取已启用的记录；读失败/未启用/不存在都返回 nil（一律拒绝）。
func (a RegistryAuthorizer) enabled(pluginID string) *Record {
	if a.Store == nil || pluginID == "" {
		return nil
	}
	reg, err := a.Store.Load()
	if err != nil {
		// 注册表读不出来时**拒绝**而不是放行：启停状态判定失败必须 fail-closed。
		return nil
	}
	rec := reg.Find(pluginID)
	if rec == nil || !rec.Enabled {
		return nil
	}
	return rec
}
