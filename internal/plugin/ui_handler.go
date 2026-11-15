package plugin

// 插件前端资源与私有 RPC 代理：/plugins/{id}/ui/… 与 /plugins/{id}/rpc/…。
//
// 两条边界（见 .agents/docs/plugin-system/03-trust-model.md）：
//  1. 静态资源只从**活跃版本目录**读，每次请求现解析 —— 不能缓存旧版本路径，
//     否则升级后一半请求拿到旧 HTML、一半拿到新脚本（"串版"）；
//  2. RPC 代理只转发到**该插件自己的实例**，绝不做通用本机端口代理：
//     代理一旦能指到别的端口，插件就等于拿到了"访问本机任意服务"的能力。

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"tt/internal/connsvc"
	"tt/internal/hostrpc"
)

// PluginPrefix 插件子资源的固定前缀。
const PluginPrefix = "/plugins"

// WebHandler 提供插件静态资源与私有 RPC 代理。
type WebHandler struct {
	Store *Store
	// Addr 取插件服务的本机地址与实例令牌（由 supervisor 提供）；nil 或取不到时 RPC 返回 503。
	Addr func(pluginID string) (addr, token string, ok bool)
	// HostAPIURL tt 自己的 Host API 根地址（如 http://127.0.0.1:28670/hostrpc）。
	// 插件页面经消息桥调 Host API 时从这里转发 —— 页面本身**永远拿不到令牌**。
	HostAPIURL string
	// Tokens 实例令牌表：桥接时给"页面代表的那次调用"签发短期令牌。
	Tokens *hostrpc.TokenStore
}

// NewWebHandler 构造插件子资源 handler（自带 /plugins/ 前缀，挂载时不要再 StripPrefix）。
func NewWebHandler(store *Store, addr func(pluginID string) (string, string, bool)) *WebHandler {
	return &WebHandler{Store: store, Addr: addr}
}

// Handler 装配路由（在设置完 HostAPIURL/Tokens 之后调用）。
func (h *WebHandler) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+PluginPrefix+"/{id}/ui/", h.serveUI)
	mux.HandleFunc(PluginPrefix+"/{id}/ui", h.redirectUI)
	mux.HandleFunc(PluginPrefix+"/{id}/rpc/", h.serveRPC)
	// 页面到 Host API 的桥：父页面校验过 nonce 之后把请求打到 这里，
	// 由 TT 换成"该插件实例"的身份再转发给它自己的 Host API。
	mux.HandleFunc(PluginPrefix+"/{id}/hostrpc/", h.serveHostBridge)
	return mux
}

// redirectUI 把 /plugins/{id}/ui 补成带斜杠的形式（相对路径才拼得对）。
func (h *WebHandler) redirectUI(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.URL.Path+"/", http.StatusTemporaryRedirect)
}

// serveUI 从活跃版本目录提供插件页面与静态资源。
func (h *WebHandler) serveUI(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rec, uiRoot, err := h.uiRoot(id)
	if err != nil {
		writeUIError(w, http.StatusNotFound, err)
		return
	}
	_ = rec

	rel := strings.TrimPrefix(r.URL.Path, PluginPrefix+"/"+id+"/ui/")
	// 路径穿越：先按 URL 语义清一遍，再确认最终还是落在 uiRoot 之内。
	clean := path.Clean("/" + rel)
	if strings.Contains(clean, "..") {
		writeUIError(w, http.StatusBadRequest, errors.New("非法资源路径"))
		return
	}
	if clean == "/" || clean == "." {
		clean = "/index.html"
	}
	full := filepath.Join(uiRoot, filepath.FromSlash(strings.TrimPrefix(clean, "/")))
	if !withinDir(uiRoot, full) {
		writeUIError(w, http.StatusBadRequest, errors.New("资源路径越出插件目录"))
		return
	}

	st, serr := os.Stat(full)
	if serr != nil || st.IsDir() {
		// 插件页面是完整页面，深链接也回 index.html（与宿主 SPA 同一策略）。
		full = filepath.Join(uiRoot, "index.html")
		if _, ierr := os.Stat(full); ierr != nil {
			writeUIError(w, http.StatusNotFound, errors.New("插件页面资源缺失（重新安装该插件）"))
			return
		}
	}

	f, oerr := os.Open(full)
	if oerr != nil {
		writeUIError(w, http.StatusNotFound, errors.New("插件资源打不开"))
		return
	}
	defer f.Close()
	info, _ := f.Stat()

	h.setUIHeaders(w, full)
	// index.html 每次都要最新（升级后不能拿到旧页面）；带哈希的静态资源可长缓存。
	if filepath.Base(full) == "index.html" {
		w.Header().Set("Cache-Control", "no-cache")
	}
	http.ServeContent(w, r, filepath.Base(full), info.ModTime(), f)
}

// uiRoot 解析某插件的 UI 根目录（页面入口所在目录）。
func (h *WebHandler) uiRoot(id string) (*Record, string, error) {
	reg, err := h.Store.Load()
	if err != nil {
		return nil, "", errors.New("读取插件注册表失败")
	}
	rec := reg.Find(id)
	if rec == nil {
		return nil, "", errors.New("没有这个插件")
	}
	if !rec.Enabled {
		return nil, "", errors.New("插件未启用")
	}
	if rec.Contributions.Frontend == nil {
		return nil, "", errors.New("这个插件没有页面")
	}
	dir := h.Store.ActiveDir(rec)
	if dir == "" {
		return nil, "", errors.New("插件没有活跃版本")
	}
	man, merr := LoadManifest(dir)
	if merr != nil {
		return nil, "", errors.New("插件清单读不出来")
	}
	rel, rerr := man.ResolveFrontendEntry()
	if rerr != nil {
		return nil, "", rerr
	}
	entry := filepath.Join(dir, filepath.FromSlash(rel))
	if !withinDir(dir, entry) {
		return nil, "", errors.New("页面入口越出插件目录")
	}
	return rec, filepath.Dir(entry), nil
}

// setUIHeaders 插件页面的安全头与缓存策略。
//
// CSP 为什么把 127.0.0.1/localhost 显式列出来：插件页面跑在 sandbox="allow-scripts"
// （没有 allow-same-origin）的 iframe 里，文档的 origin 是不透明的，`'self'` 对它的
// 含义不可依赖；显式列出宿主来源才是可工作的写法。
// `connect-src 'none'` 是刻意的：插件要数据走消息桥，别自己直连宿主 API。
func (h *WebHandler) setUIHeaders(w http.ResponseWriter, full string) {
	hdr := w.Header()
	hdr.Set("X-Content-Type-Options", "nosniff")
	hdr.Set("Referrer-Policy", "no-referrer")
	hdr.Set("Content-Security-Policy",
		"default-src 'self' http://127.0.0.1:* http://localhost:*; "+
			"script-src 'self' 'unsafe-inline' http://127.0.0.1:* http://localhost:*; "+
			"style-src 'self' 'unsafe-inline' http://127.0.0.1:* http://localhost:*; "+
			"img-src 'self' data: blob:; font-src 'self' data:; "+
			"connect-src 'none'; frame-ancestors 'self'; form-action 'none'; base-uri 'none'")
	if filepath.Base(full) != "index.html" && hashedName(filepath.Base(full)) {
		hdr.Set("Cache-Control", "public, max-age=31536000, immutable")
	}
}

// hashedName 判断文件名是否带内容哈希（vite 产物形如 index-DCXJThG6.js）。
func hashedName(name string) bool {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	i := strings.LastIndexByte(base, '-')
	if i < 0 || len(base)-i-1 < 6 {
		return false
	}
	for _, c := range base[i+1:] {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

// serveRPC 把 /plugins/{id}/rpc/… 转发到该插件自己的服务实例。
func (h *WebHandler) serveRPC(w http.ResponseWriter, r *http.Request) {
	if !requireLocalManagement(w, r) {
		return
	}
	id := r.PathValue("id")
	reg, err := h.Store.Load()
	if err != nil {
		writeUIError(w, http.StatusServiceUnavailable, errors.New("读取插件注册表失败"))
		return
	}
	rec := reg.Find(id)
	if rec == nil {
		writeUIError(w, http.StatusNotFound, errors.New("没有这个插件"))
		return
	}
	if !rec.Enabled {
		writeUIError(w, http.StatusConflict, errors.New("插件未启用"))
		return
	}
	if rec.Contributions.Service == nil {
		writeUIError(w, http.StatusConflict, errors.New("这个插件没有常驻服务，无法代理 RPC"))
		return
	}
	if h.Addr == nil {
		writeUIError(w, http.StatusServiceUnavailable, errors.New("插件服务地址不可用（supervisor 未接入）"))
		return
	}
	addr, token, ok := h.Addr(id)
	if !ok || addr == "" {
		writeUIError(w, http.StatusServiceUnavailable, errors.New("插件服务没有在运行（看 tt plugin doctor / logs）"))
		return
	}
	target, perr := url.Parse("http://" + addr)
	if perr != nil || target.Host == "" {
		writeUIError(w, http.StatusInternalServerError, errors.New("插件服务地址不可用"))
		return
	}
	// 只允许转发到本机回环地址：supervisor 只会给出这种地址，这里再确认一次。
	host, _, herr := net.SplitHostPort(target.Host)
	if herr != nil {
		writeUIError(w, http.StatusInternalServerError, errors.New("插件服务地址不可用"))
		return
	}
	if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
		writeUIError(w, http.StatusForbidden, errors.New("拒绝转发到非回环地址"))
		return
	}

	suffix := strings.TrimPrefix(r.URL.Path, PluginPrefix+"/"+id+"/rpc")
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.Out.URL.Scheme = "http"
			pr.Out.URL.Host = target.Host
			pr.Out.URL.Path = suffix
			pr.Out.URL.RawQuery = r.URL.RawQuery
			pr.Out.Host = target.Host
			// 插件服务用它自己的实例令牌认证：TI 不把宿主的管理入口带过去。
			if token != "" {
				pr.Out.Header.Set("Authorization", "Bearer "+token)
			}
			pr.Out.Header.Del("Cookie")
		},
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeUIError(w, http.StatusBadGateway, errors.New("插件服务不可达: "+err.Error()))
		},
	}
	proxy.ServeHTTP(w, r)
}

// serveHostBridge 把插件页面的 Host API 调用转发给 TT 自己的 /hostrpc/。
//
// 为什么要有这一层：iframe 里**不能**放实例令牌（页面代码来自第三方，令牌等于把
// 该插件的连接权限交给任何能注入页面脚本的东西）。所以页面只知道"调桥"，桥这边：
//  1. 用**服务端**的注册表确认这个插件已启用；
//  2. 以该插件实例的身份签发一次性令牌（crypto/rand，调用结束即撤销）；
//  3. 转发到 /hostrpc/，由那边照常做能力/环境授权。
//
// 页面因此永远不持有凭证，而授权判定仍然只有一处（Host API 的拦截器）。
func (h *WebHandler) serveHostBridge(w http.ResponseWriter, r *http.Request) {
	if !requireLocalManagement(w, r) {
		return
	}
	id := r.PathValue("id")
	reg, err := h.Store.Load()
	if err != nil {
		writeUIError(w, http.StatusServiceUnavailable, errors.New("读取插件注册表失败"))
		return
	}
	rec := reg.Find(id)
	if rec == nil {
		writeUIError(w, http.StatusNotFound, errors.New("没有这个插件"))
		return
	}
	if !rec.Enabled {
		writeUIError(w, http.StatusConflict, errors.New("插件未启用"))
		return
	}
	if h.Tokens == nil {
		writeUIError(w, http.StatusServiceUnavailable, errors.New("Host API 未接入"))
		return
	}
	// Host API 地址：优先用显式注入的（测试/特殊部署）；否则用**本请求的 Host**
	// 拼一个回环地址 —— 这样不必等服务监听后回填（那会有"装配期写、请求期读"的竞态），
	// 也不会因为端口顺延而指向错的端口。前提是 Host 必须是本机名/回环地址。
	apiURL := h.HostAPIURL
	if apiURL == "" {
		if !hostIsLoopbackName(r.Host) {
			writeUIError(w, http.StatusForbidden, errors.New("Host 头不是本机地址"))
			return
		}
		apiURL = "http://" + r.Host + hostrpc.Prefix
	}
	target, perr := url.Parse(apiURL)
	if perr != nil || target.Host == "" {
		writeUIError(w, http.StatusInternalServerError, errors.New("Host API 地址不可用"))
		return
	}

	instanceID, ierr := newInstanceID("ui")
	if ierr != nil {
		writeUIError(w, http.StatusInternalServerError, ierr)
		return
	}
	caller := connsvc.Caller{Kind: connsvc.CallerServicePlugin, PluginID: id, InstanceID: instanceID}
	token := h.Tokens.Issue(caller)
	defer h.Tokens.Revoke(token)

	suffix := strings.TrimPrefix(r.URL.Path, PluginPrefix+"/"+id+"/hostrpc")
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.Out.URL.Scheme = target.Scheme
			pr.Out.URL.Host = target.Host
			pr.Out.URL.Path = strings.TrimSuffix(target.Path, "/") + "/" + strings.TrimPrefix(suffix, "/")
			pr.Out.URL.RawQuery = r.URL.RawQuery
			pr.Out.Host = target.Host
			pr.Out.Header.Set("Authorization", "Bearer "+token)
			// 浏览器带来的 Origin/Cookie 一律不带：桥的请求是 TT 代表插件发出的。
			pr.Out.Header.Del("Origin")
			pr.Out.Header.Del("Cookie")
		},
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeUIError(w, http.StatusBadGateway, errors.New("Host API 不可达: "+err.Error()))
		},
	}
	// 桥只承载一次性 unary 调用；请求体上限与消息桥的 1 MiB 对齐。
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	proxy.ServeHTTP(w, r)
}

// withinDir 判断 target 是否在 dir 之内（含 dir 本身）。
func withinDir(dir, target string) bool {
	rel, err := filepath.Rel(dir, target)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."
}

// writeUIError 插件子资源的错误页：**可见**的说明，不是空白 404。
//
// 为什么不复用宿主的 writeErr：这是给浏览器（iframe）看的，一段可读的 HTML
// 比一个 JSON 信封有用得多；而且插件页面加载失败时用户看到的应当是一句人话。
func writeUIError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	msg := "插件页面不可用"
	if err != nil {
		msg = err.Error()
	}
	_, _ = io.WriteString(w, "<!doctype html><meta charset=\"utf-8\"><title>插件页面不可用</title>"+
		"<div style=\"font:14px/1.6 system-ui;padding:24px\"><h3>插件页面不可用</h3><p>"+
		htmlEscape(msg)+"</p><p style=\"color:#666\">可以用 <code>tt plugin list --all</code> 看状态，"+
		"<code>tt plugin doctor</code> 做自检。</p></div>")
}

// htmlEscape 最小 HTML 转义（错误文案来自我们自己的固定句式，这里只防意外）。
func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")
	return r.Replace(s)
}

// requireLocalManagement 给"会改变宿主状态"的子资源端点加本机来源门。
//
// 与 web 的插件管理端点同一套判据（见 internal/web/plugins.go）：远端地址必须是回环，
// Origin 若存在必须是本机。这里不 import internal/web（那会造成 web → plugin → web 的环），
// 所以这份判据在本包内独立实现，语义保持一致。
func requireLocalManagement(w http.ResponseWriter, r *http.Request) bool {
	if !remoteIsLoopback(r.RemoteAddr) {
		writeUIError(w, http.StatusForbidden, errors.New("只接受本机请求"))
		return false
	}
	if o := strings.TrimSpace(r.Header.Get("Origin")); o != "" {
		u, uerr := url.Parse(o)
		if uerr != nil || !hostIsLoopbackName(u.Host) {
			writeUIError(w, http.StatusForbidden, errors.New("拒绝跨站来源（Origin 不是本机）"))
			return false
		}
	}
	return true
}

// remoteIsLoopback RemoteAddr 是否回环地址。
func remoteIsLoopback(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = strings.Trim(remoteAddr, "[]")
	}
	return hostIsLoopbackName(host)
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
