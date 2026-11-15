package web

// 插件管理 API（Core 管理面的一部分）。
//
// 为什么挂在 /api/ 而不是 /hostrpc/：插件令牌只在 /hostrpc/ 被接受，这两个前缀之间
// 没有转发关系 —— 普通插件令牌**够不到**这里（见 docs 的权限矩阵"管理面默认拒绝"）。
// 反过来，本机浏览器/脚本走的是与其它设置页端点同一扇门（loopback，无令牌），
// 与既有 /api/hosts 等的安全姿态一致，不另立一套。
//
// 这些端点只操作元数据与载荷（安装/启停/卸载/自检），**不启动**第三方代码；
// 常驻服务由 tt serve 的 supervisor 负责（阶段 3）。

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"

	"tt/internal/plugin"
)

// requireLocalManagement 给插件管理端点加一道"本机 + 本机来源"门。
//
// 为什么 /api/ 的整体安全姿态没变，这里却要单独加：既有 /api/* 是本机信任（浏览器与
// 脚本都在本机），但插件管理是**安装并启用可执行代码**的动作 —— 一个恶意网页向
// 127.0.0.1 发 POST 就能做到这件事，代价太高。所以这里额外要求：
//   - 远端地址必须是回环；
//   - Origin 若存在必须是本机（跨站表单/脚本一律拒）。
//
// 它挡不住"同用户的本机进程" —— 那是文档里写明的信任边界（插件本来就是本机可执行
// 代码，进程之间没有沙箱）。
func requireLocalManagement(w http.ResponseWriter, r *http.Request) bool {
	if !remoteIsLoopback(r.RemoteAddr) {
		writeErr(w, http.StatusForbidden, errors.New("插件管理只接受本机请求"))
		return false
	}
	if o := strings.TrimSpace(r.Header.Get("Origin")); o != "" {
		u, err := url.Parse(o)
		if err != nil || !hostIsLoopbackName(u.Host) {
			writeErr(w, http.StatusForbidden, errors.New("插件管理拒绝跨站来源（Origin 不是本机）"))
			return false
		}
	}
	return true
}

// remoteIsLoopback RemoteAddr 是否是回环地址。
func remoteIsLoopback(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = strings.Trim(remoteAddr, "[]")
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// hostIsLoopbackName host[:port] 是否本机名或回环地址。
func hostIsLoopbackName(hostport string) bool {
	if hostport == "" {
		return false
	}
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	} else {
		host = strings.Trim(hostport, "[]")
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// pluginManagerOf 返回管理面；未装配时 503（而不是 404：让人知道是"没接上"）。
func (s *Server) pluginManagerOf(w http.ResponseWriter) (*plugin.Manager, bool) {
	if s.opt.PluginMgr == nil {
		writeErr(w, http.StatusServiceUnavailable, errors.New("插件管理面未装配"))
		return nil, false
	}
	return s.opt.PluginMgr, true
}

// hPluginsList 列出插件（默认全部，含未启用）。
func (s *Server) hPluginsList(w http.ResponseWriter, r *http.Request) {
	if !requireLocalManagement(w, r) {
		return
	}
	m, ok := s.pluginManagerOf(w)
	if !ok {
		return
	}
	list, err := m.List()
	if err != nil {
		writePluginErr(w, err)
		return
	}
	onlyEnabled := r.URL.Query().Get("enabled") == "1"
	if onlyEnabled {
		filtered := make([]plugin.Summary, 0, len(list))
		for _, s := range list {
			if s.Enabled {
				filtered = append(filtered, s)
			}
		}
		list = filtered
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "count": len(list), "plugins": list})
}

// hPluginInfo 一个插件的详情（清单 + 授权 + 落点）。
func (s *Server) hPluginInfo(w http.ResponseWriter, r *http.Request) {
	if !requireLocalManagement(w, r) {
		return
	}
	m, ok := s.pluginManagerOf(w)
	if !ok {
		return
	}
	info, err := m.Info(r.PathValue("id"))
	if err != nil {
		writePluginErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "plugin": info})
}

// pluginInstallReq 安装请求体。
type pluginInstallReq struct {
	Source  string `json:"source"`
	Replace bool   `json:"replace"`
}

// hPluginInstall 安装一个本地插件包（目录或 zip）。
func (s *Server) hPluginInstall(w http.ResponseWriter, r *http.Request) {
	if !requireLocalManagement(w, r) {
		return
	}
	m, ok := s.pluginManagerOf(w)
	if !ok {
		return
	}
	var req pluginInstallReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	src := strings.TrimSpace(req.Source)
	if src == "" {
		writeErr(w, http.StatusBadRequest, errors.New("source 不能为空（本地插件目录或 zip 路径）"))
		return
	}
	rec, err := m.Install(plugin.InstallOptions{Source: src, Replace: req.Replace})
	if err != nil {
		writePluginErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "id": rec.ID, "version": rec.ActiveVersion, "enabled": rec.Enabled,
		"versionDir": m.Store.VersionDir(rec.ID, rec.ActiveVersion),
	})
}

// pluginGrantsReq 启用/授权请求体。
type pluginGrantsReq struct {
	Capabilities []string `json:"capabilities"`
	Envs         []string `json:"envs"`
}

// hPluginEnable 启用并授权。
func (s *Server) hPluginEnable(w http.ResponseWriter, r *http.Request) {
	if !requireLocalManagement(w, r) {
		return
	}
	m, ok := s.pluginManagerOf(w)
	if !ok {
		return
	}
	var req pluginGrantsReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	rec, err := m.Enable(r.PathValue("id"), plugin.Grants{Capabilities: req.Capabilities, Envs: req.Envs})
	if err != nil {
		writePluginErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "id": rec.ID, "enabled": rec.Enabled,
		"capabilities": rec.Grants.Capabilities, "envs": rec.Grants.Envs,
	})
}

// hPluginDisable 停用。
func (s *Server) hPluginDisable(w http.ResponseWriter, r *http.Request) {
	if !requireLocalManagement(w, r) {
		return
	}
	m, ok := s.pluginManagerOf(w)
	if !ok {
		return
	}
	rec, err := m.Disable(r.PathValue("id"))
	if err != nil {
		writePluginErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": rec.ID, "enabled": rec.Enabled})
}

// hPluginUninstall 卸载；purgeData 必须显式给，且必须是 POST 体里的字段。
func (s *Server) hPluginUninstall(w http.ResponseWriter, r *http.Request) {
	if !requireLocalManagement(w, r) {
		return
	}
	m, ok := s.pluginManagerOf(w)
	if !ok {
		return
	}
	var req struct {
		PurgeData bool `json:"purgeData"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	id := r.PathValue("id")
	if err := m.Uninstall(id, plugin.UninstallOptions{PurgeData: req.PurgeData}); err != nil {
		writePluginErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id, "purgedData": req.PurgeData})
}

// hPluginDoctor 自检（不启动插件、不连网）。
func (s *Server) hPluginDoctor(w http.ResponseWriter, r *http.Request) {
	if !requireLocalManagement(w, r) {
		return
	}
	m, ok := s.pluginManagerOf(w)
	if !ok {
		return
	}
	issues, err := m.Doctor(strings.TrimSpace(r.URL.Query().Get("id")))
	if err != nil {
		writePluginErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": len(issues) == 0, "count": len(issues), "issues": issues})
}

// writePluginErr 把插件错误映射成 HTTP 码。
//
// 映射只看错误种类：校验/包不安全 → 400；没有这个插件 → 404；忙/冲突 → 409。
func writePluginErr(w http.ResponseWriter, err error) {
	var ve *plugin.ValidationError
	switch {
	case errors.As(err, &ve):
		writeErr(w, http.StatusBadRequest, err)
	case errors.Is(err, plugin.ErrZipUnsafe):
		writeErr(w, http.StatusBadRequest, err)
	case strings.Contains(err.Error(), "没有插件"):
		writeErr(w, http.StatusNotFound, err)
	case strings.Contains(err.Error(), "启用中"),
		strings.Contains(err.Error(), "已安装"),
		strings.Contains(err.Error(), "冲突"),
		strings.Contains(err.Error(), "占用"):
		writeErr(w, http.StatusConflict, err)
	default:
		writeErr(w, http.StatusInternalServerError, err)
	}
}
