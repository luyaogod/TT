package plugin

import "tt/internal/connsvc"

// RegistryAuthorizer 用插件注册表做授权来源（connsvc.Authorizer 的实现）。
//
// 每次判定都现读注册表，而不是缓存：授权是安全决定，而"改了授权要重启才生效"
// 正是这类系统最容易被忽略的坑（用户撤了权、插件还在用）。注册表很小（几十 KB），
// 一次 RPC 读一次文件的代价远小于"撤权没生效"。
type RegistryAuthorizer struct {
	Store *Store
}

// Permitted 调用方是否持有某能力。
func (a RegistryAuthorizer) Permitted(caller connsvc.Caller, capability connsvc.Capability) bool {
	if caller.IsBuiltin() {
		return true // 第一方模块不受插件授权限制
	}
	rec := a.enabled(caller.PluginID)
	if rec == nil {
		return false
	}
	return rec.Grants.Has(string(capability))
}

// AllowedEnvs 调用方在某能力下被授权的环境。
func (a RegistryAuthorizer) AllowedEnvs(caller connsvc.Caller, capability connsvc.Capability) ([]string, bool) {
	if caller.IsBuiltin() {
		return nil, true
	}
	rec := a.enabled(caller.PluginID)
	if rec == nil || !rec.Grants.Has(string(capability)) {
		return nil, false
	}
	return rec.Grants.Envs, false
}

// enabled 取已启用的记录；读失败/未启用/不存在都返回 nil（一律拒绝）。
func (a RegistryAuthorizer) enabled(pluginID string) *Record {
	if a.Store == nil || pluginID == "" {
		return nil
	}
	reg, err := a.Store.Load()
	if err != nil {
		// 注册表读不出来时**拒绝**而不是放行：授权判定失败必须 fail-closed。
		return nil
	}
	rec := reg.Find(pluginID)
	if rec == nil || !rec.Enabled {
		return nil
	}
	return rec
}
