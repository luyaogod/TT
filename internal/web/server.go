// Package web 是统一本地服务的 HTTP 层。
//
// 一个进程、一个端口，同时承载三件事：
//
//	/debug/  调试工作台 SPA（原 TDebug 的前端）与其 REST/WS API
//	/dict/   字典/镜像/同步 SPA（原 TDictCli 的前端）与其 API
//	/api/*   统一 API —— 其中 /api/hosts 是两套页面共用的环境/数据库管理端点
//
// 合并前两个工具各有自己的 HTTP 服务、各自一个配置页、各自读写配置的不同节；
// 现在环境清单只有一份数据源（config.json 的 hosts 节），由本包的 /api/hosts 统一读写，
// 两套页面共用它。这就是"统一配置管理"落在 HTTP 层的样子。
package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"tt/internal/config"
	"tt/internal/dbconfig"
	"tt/internal/erpdb"
	"tt/internal/host"
)

// DefaultListen 是 serve 未显式配置监听地址时的默认地址。
// 合并前 TDebug 的 debug.listen 与 TDictCli 的 server.DefaultListen 就是同一个值，
// 合并后自然收敛成同一个端口。
const DefaultListen = config.DefaultListen

// maxPortTries 端口被占用时向后探测的次数（原 TDictCli 的行为：最多顺延 50 个端口）。
const maxPortTries = 50

// Options 装配一个统一服务所需的全部外部依赖。
type Options struct {
	// DebugFS / DictFS 两套前端各自的构建产物根目录（dist 下的 debug/ 与 dict/）。
	// 由调用方经 common.WebSub 取好；nil 或没构建时该页面回落成引导页。
	DebugFS fs.FS
	DictFS  fs.FS
	// ConfigPath 配置文件路径。空则按统一规则解析（写路径允许缺省落点）。
	ConfigPath string
	// Version 版本号，出现在 /api/health 与引导页上。
	Version string

	// Debug 调试工作台的 handler（由 internal/cli/debug 装配）。nil = 未接入。
	Debug http.Handler
	// DebugBase 调试工作台的挂载前缀，默认 "/debug"。
	DebugBase string
	// Dict 字典页的 handler（由 internal/dict/server 装配）。nil = 未接入。
	Dict http.Handler
	// DictBase 字典页的挂载前缀，默认 "/dict"。
	DictBase string

	// Reloaders 配置被本服务改写后需要重新加载内存态的子系统。
	//
	// 起作用的场景：两个配置页写的是同一份 hosts 节。调试服务在内存里持有一份
	// *Config 并做了热替换，若从字典页改了环境，它得跟着重读 —— 否则一次页面写
	// 下去，另一个服务还按旧环境连，表现是"改了没生效"。
	Reloaders []ConfigReloader

	// Shutdown 收到 POST /api/shutdown 时调用；nil 表示不注册该端点。
	// 桌面版（Electron 外壳）靠它让关窗时 Go 侧优雅收尾（会话/SSH 连接收口），
	// 而不是被直接 kill。
	Shutdown func()

	// SyncDefaultTarget 字典同步目标的缺省位置。
	//
	// 由挂载方注入，好让统一层与字典子系统**对同一个值**达成一致：设置页会显示
	// 「默认位置：X」，而字典页真正往那儿写。两边各算一次的话，显示的那个路径
	// 可能不是实际写入的那个 —— 用户按显示去核对会发现文件不在那里。
	//
	// 留空 = 用 config.DefaultSyncTarget()（当前目录的 erp_data.db）。
	SyncDefaultTarget string
}

// ConfigReloader 由持有配置内存态的子系统实现；本服务在配置写入成功后调用它。
type ConfigReloader interface {
	// ReloadConfig 重新从磁盘读取配置并热替换内存态。
	// 实现须保证：并发安全、幂等、失败时保留原有内存态不变。
	ReloadConfig() error
}

// Server 是统一服务。
type Server struct {
	opt  Options
	mux  *http.ServeMux
	once bool
}

// New 装配路由。返回的 Server 可直接交给 http.Server 使用。
func New(opt Options) *Server {
	if opt.DebugBase == "" {
		opt.DebugBase = "/debug"
	}
	if opt.DictBase == "" {
		opt.DictBase = "/dict"
	}
	s := &Server{opt: opt, mux: http.NewServeMux()}
	s.routes()
	return s
}

// Handler 返回顶层 http.Handler。
func (s *Server) Handler() http.Handler { return s.mux }

// routes 装配路由表。集中在一处，方便一眼看清整个服务的接口面。
//
// 两个子系统的挂载用带前缀的 API 路径 + 各自的 SPA，而不是把子系统 handler
// 整个挂在 /debug/ 下：后者的做法会让子系统的 "/" 兜底路由吃掉整个前缀下的
// 所有请求（Go 的 ServeMux 匹配到前缀后不再向外回落），SPA 就永远出不来。
// 只把 /*/api/ 转给子系统，页面由本包统一从 web/dist/<应用> 提供。
func (s *Server) routes() {
	m := s.mux

	// ---- 统一 API：两套页面共用的环境/数据库管理 ----
	m.HandleFunc("GET /api/health", s.hHealth)
	m.HandleFunc("GET /api/hosts", s.hHostsGet)
	m.HandleFunc("PUT /api/hosts", s.hHostsPut)
	m.HandleFunc("POST /api/dbprobe", s.hDBProbe)
	m.HandleFunc("POST /api/dbaccverify", s.hDBAccVerify)
	m.HandleFunc("POST /api/conntest", s.hConnTest)
	m.HandleFunc("GET /api/config/meta", s.hConfigMeta)
	m.HandleFunc("GET /api/config/status", s.hConfigStatus)
	// 命令行安装(PATH):应用级动作,归统一层 —— 合并前它在 /dict/api/install 下,
	// 于是共用的设置页要用它就得去调字典子系统的私有 API。
	m.HandleFunc("GET /api/install", s.hInstallGet)
	m.HandleFunc("POST /api/install", s.hInstallAdd)
	m.HandleFunc("DELETE /api/install", s.hInstallRemove)
	if s.opt.Shutdown != nil {
		m.HandleFunc("POST /api/shutdown", s.hShutdown)
	}

	// ---- 子系统：只接管各自的 API，前缀剥掉后交给它们 ----
	if s.opt.Debug != nil {
		m.Handle(s.opt.DebugBase+"/api/", http.StripPrefix(s.opt.DebugBase, s.opt.Debug))
	}
	if s.opt.Dict != nil {
		m.Handle(s.opt.DictBase+"/api/", http.StripPrefix(s.opt.DictBase, s.opt.Dict))
	}

	// ---- 两套页面（前端构建产物） ----
	m.Handle(s.opt.DebugBase+"/", SPAHandler(s.opt.DebugFS, s.opt.DebugBase+"/", "调试工作台未构建"))
	m.Handle(s.opt.DictBase+"/", SPAHandler(s.opt.DictFS, s.opt.DictBase+"/", "数据字典页未构建"))

	// ---- 根路径 ----
	m.HandleFunc("/", s.hRoot)
}

// ListenAndServe 在 addr 上启动服务。addr 为空时用配置里的 listen，再空用 DefaultListen。
// 端口被占用时向后顺延（最多 maxPortTries 个），返回实际使用的地址。
func (s *Server) ListenAndServe(addr string) (string, error) {
	if addr == "" {
		addr = s.listenFromConfig()
	}
	ln, used, err := listenWithFallback(addr)
	if err != nil {
		return "", err
	}
	srv := &http.Server{
		Handler:           s.mux,
		ReadHeaderTimeout: 10 * time.Second,
		// 调试工作台有长连接的 WebSocket 与长时间挂起的命令，
		// 所以不设 WriteTimeout / IdleTimeout。
	}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("[tt] HTTP 服务退出: %v", err)
		}
	}()
	return used, nil
}

// listenFromConfig 读取配置里的监听地址；读不到就用缺省值。
func (s *Server) listenFromConfig() string {
	path, err := s.configPath(false)
	if err != nil {
		return DefaultListen
	}
	r, err := config.Load(path)
	if err != nil || r.Listen == "" {
		return DefaultListen
	}
	return r.Listen
}

// listenWithFallback 在 addr 上监听；被占用时按端口递增顺延。
// 返回监听器与实际使用的地址。
func listenWithFallback(addr string) (net.Listener, string, error) {
	ln, err := net.Listen("tcp", addr)
	if err == nil {
		return ln, ln.Addr().String(), nil
	}
	hostPart, portPart, splitErr := net.SplitHostPort(addr)
	if splitErr != nil {
		return nil, "", fmt.Errorf("监听地址格式不正确 %q: %w", addr, err)
	}
	port := 0
	if _, e := fmt.Sscanf(portPart, "%d", &port); e != nil || port <= 0 {
		return nil, "", fmt.Errorf("监听 %s 失败: %w", addr, err)
	}
	for i := 1; i <= maxPortTries; i++ {
		try := net.JoinHostPort(hostPart, fmt.Sprint(port+i))
		if ln, e := net.Listen("tcp", try); e == nil {
			log.Printf("[tt] %s 被占用，改用 %s", addr, try)
			return ln, ln.Addr().String(), nil
		}
	}
	return nil, "", fmt.Errorf("监听 %s 失败，且向后探测 %d 个端口都被占用: %w", addr, maxPortTries, err)
}

// configPath 解析配置文件路径。allowMissing=true 时返回缺省落点而不是报错。
func (s *Server) configPath(allowMissing bool) (string, error) {
	if s.opt.ConfigPath != "" {
		return s.opt.ConfigPath, nil
	}
	return config.ResolvePath("", allowMissing)
}

// ---------- 统一 API 处理器 ----------

func (s *Server) hHealth(w http.ResponseWriter, r *http.Request) {
	path, _ := s.configPath(true)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"version": s.opt.Version,
		"config":  path,
		"debug":   s.opt.Debug != nil,
		"dict":    s.opt.Dict != nil,
	})
}

// hHostsGet 返回共用环境清单。两套页面的「设置 → 环境」都读它。
func (s *Server) hHostsGet(w http.ResponseWriter, r *http.Request) {
	path, err := s.configPath(false)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	root, err := config.Load(path)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"config":    path,
		"activeEnv": root.Hosts.ActiveEnv,
		"sshs":      root.Hosts.SSHs,
		// 监听地址也在这一份里下发，设置页的「外观/高级」页要用
		"listen":  root.Listen,
		"debug":   root.Debug,
		"query":   root.Query,
		"mirror":  root.Mirror,
		"bdldoc":  root.Bdldoc,
		"sync":    root.Sync,
		"tdev":    root.Tdev,
		"version": s.opt.Version,
	})
}

// hostsPutReq 是 PUT /api/hosts 的请求体。
//
// 只带要改的节：nil 表示"这一节不动"。这样两套页面各改各的部分时不会互相覆盖 ——
// 原来 TDebug 的 hSettingsPut 就只替换自己那一节，这里沿用同样的约定。
type hostsPutReq struct {
	Hosts  *config.Hosts         `json:"hosts,omitempty"`
	Listen *string               `json:"listen,omitempty"`
	Debug  *config.DebugSettings `json:"debug,omitempty"`
	Query  *config.QuerySettings `json:"query,omitempty"`
	Mirror *config.MirrorSettings
	Bdldoc *config.BdldocSettings
	Sync   *config.SyncSettings
	Tdev   *config.TdevSettings
}

func (s *Server) hHostsPut(w http.ResponseWriter, r *http.Request) {
	path, err := s.configPath(true)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	var req hostsPutReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}

	err = config.Edit(path, nil, func(root map[string]any) error {
		if req.Hosts != nil {
			if err := validateHosts(req.Hosts); err != nil {
				return err
			}
			sec, err := structToSection(*req.Hosts)
			if err != nil {
				return err
			}
			// 保留前端表单不管理的 db 字段（目前是 viaSsh）。
			// 配置页只编辑 type/host/port/service/database/readonlySql/accounts，
			// 整节替换会把手工配的 SSH 隧道悄悄抹掉 —— 那是"保存一次就丢配置"，
			// 而用户完全看不出发生过什么。按环境名对齐，缺什么补什么。
			preserveUngovernedFields(root, sec)
			root["hosts"] = sec
		}
		if req.Listen != nil {
			root["listen"] = *req.Listen
		}
		// 逐节判空。注意**不能**写成 map 循环 + `if v == nil { continue }` ——
		// 那些字段是指针类型，把 nil 指针装进 any 得到的接口值不等于 nil，
		// 于是"本次没提交的节"会被序列化成 null 再当成空对象写回去，
		// 把用户原有的 query/mirror/debug 全清掉。这里逐个类型化判断。
		patches := []struct {
			key string
			set func() error
		}{
			{"debug", func() error {
				if req.Debug == nil {
					return nil
				}
				return putSection(root, "debug", req.Debug)
			}},
			{"query", func() error {
				if req.Query == nil {
					return nil
				}
				return putSection(root, "query", req.Query)
			}},
			{"mirror", func() error {
				if req.Mirror == nil {
					return nil
				}
				return putSection(root, "mirror", req.Mirror)
			}},
			{"bdldoc", func() error {
				if req.Bdldoc == nil {
					return nil
				}
				return putSection(root, "bdldoc", req.Bdldoc)
			}},
			{"sync", func() error {
				if req.Sync == nil {
					return nil
				}
				return putSection(root, "sync", req.Sync)
			}},
			{"tdev", func() error {
				if req.Tdev == nil {
					return nil
				}
				return putSection(root, "tdev", req.Tdev)
			}},
		}
		for _, p := range patches {
			if err := p.set(); err != nil {
				return err
			}
		}
		root["schemaVersion"] = config.SchemaVersion
		return nil
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	// 通知持有配置内存态的子系统（如调试服务）重新加载。
	// 失败只记日志：配置已经落盘成功，重新加载失败不该把一个成功的写报成失败，
	// 但必须让用户看见 —— 否则就成了"保存成功但没生效"。
	for _, r := range s.opt.Reloaders {
		if r == nil {
			continue
		}
		if err := r.ReloadConfig(); err != nil {
			log.Printf("[tt] 配置已保存，但子系统重新加载失败: %v", err)
		}
	}
	s.hHostsGet(w, r)
}

// preserveUngovernedFields 把旧配置里、配置页表单不管理的字段补进新节。
//
// 有两层：
//
//	环境级  hosts.sshs[].launchArgs / watchdogSeconds —— 该环境的调试参数覆盖
//	db 级   hosts.sshs[].db.viaSsh                  —— SSH 端口转发隧道
//
// 表单里都没有对应的输入控件，而 PUT 是整节替换，所以不做这一步就会"打开设置页点一次
// 保存，手工配的东西静默消失"。做法是按环境名对齐，新条目缺该键而旧条目有时原样搬过来。
//
// 权衡：这样也就无法通过配置页删掉它们 —— 想删就手改 config.json，或 `tt config` 改。
// 宁可"删不掉"也不要"保存一次就静默丢配置"。
func preserveUngovernedFields(oldRoot, newSec map[string]any) {
	oldHosts, _ := oldRoot["hosts"].(map[string]any)
	oldList, _ := oldHosts["sshs"].([]any)
	if len(oldList) == 0 {
		return
	}
	byName := make(map[string]map[string]any, len(oldList))
	for _, e := range oldList {
		m, _ := e.(map[string]any)
		if m == nil {
			continue
		}
		if name, _ := m["name"].(string); name != "" {
			byName[name] = m
		}
	}

	newList, _ := newSec["sshs"].([]any)
	for _, e := range newList {
		m, _ := e.(map[string]any)
		if m == nil {
			continue
		}
		name, _ := m["name"].(string)
		old, ok := byName[name]
		if !ok {
			continue
		}
		carryOver(old, m, ungovernedEnvKeys)
		oldDB, _ := old["db"].(map[string]any)
		newDB, _ := m["db"].(map[string]any)
		if oldDB == nil || newDB == nil {
			continue
		}
		carryOver(oldDB, newDB, ungovernedDBKeys)
	}
}

// carryOver 把 from 里、to 中缺席的键原样搬过去。
func carryOver(from, to map[string]any, keys []string) {
	for _, k := range keys {
		if _, has := to[k]; has {
			continue
		}
		if v, has := from[k]; has {
			to[k] = v
		}
	}
}

// ungovernedEnvKeys 配置页表单没有输入控件的**环境级**子键，保存时需从旧配置补回。
var ungovernedEnvKeys = []string{"launchArgs", "watchdogSeconds"}

// ungovernedDBKeys 配置页表单没有输入控件的 **db 级**子键，保存时需从旧配置补回。
var ungovernedDBKeys = []string{"viaSsh"}

// hDBProbe 「从服务器获取」：登录 SSH 后探测库连接要素，结果仅作回填参考。
func (s *Server) hDBProbe(w http.ResponseWriter, r *http.Request) {
	var req host.DBProbeReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	out, err := host.ProbeDBConfig(req)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// hDBAccVerify 逐账号在服务器侧验证。
func (s *Server) hDBAccVerify(w http.ResponseWriter, r *http.Request) {
	var req host.DBAccVerifyReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if err := host.VerifyDBAcct(req); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// connTestReq 客户端直连测试。
type connTestReq struct {
	Connection dbconfig.Connection `json:"connection"`
}

// hConnTest 从本机直连数据库，验证「环境 → 库」这条链路真的通。
// 服务器侧可达而客户端不可达时，连接里配的 viaSsh 隧道由调用方先建好再测。
func (s *Server) hConnTest(w http.ResponseWriter, r *http.Request) {
	var req connTestReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	c := req.Connection
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	conn, err := erpdb.Open(ctx, c)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": false, "stage": "connect", "error": err.Error(),
		})
		return
	}
	defer conn.Close()

	ver, err := conn.ServerVersion(ctx)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": false, "stage": "version", "error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "type": conn.Type(), "serverVersion": ver, "address": c.Address(),
	})
}

// hConfigMeta 给设置页展示"配置在哪、什么结构"用。
func (s *Server) hConfigMeta(w http.ResponseWriter, r *http.Request) {
	path, _ := s.configPath(true)
	writeJSON(w, http.StatusOK, map[string]any{
		"config":         path,
		"defaultConfig":  config.DefaultConfigPath(),
		"portable":       config.IsPortable(),
		"toolsHome":      config.ToolsHome(),
		"schemaVersion":  config.SchemaVersion,
		"defaultListen":  DefaultListen,
		"legacyTools":    config.LegacyToolConfigPaths(),
		"supportedTypes": []string{"oracle", "kingbase"},
	})
}

// hShutdown 桌面版关窗时调用：让 Go 侧优雅收尾（会话/SSH 连接收口），
// 而不是被外壳直接 kill 掉进程。
//
// 只在本机监听时可用 —— 服务默认绑 127.0.0.1，但用户可以把 listen 配成
// 0.0.0.0；那种情况下暴露一个能停服务的端点是不合适的，所以这里显式挡掉。
// 先回响应再触发停止：调用方要能读到结果，否则它只会看到一个连接被重置。
func (s *Server) hShutdown(w http.ResponseWriter, r *http.Request) {
	if !isLoopback(r.RemoteAddr) {
		writeErr(w, http.StatusForbidden, fmt.Errorf(
			"/api/shutdown 只接受本机请求（服务当前监听在非回环地址上）"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	go s.opt.Shutdown()
}

// isLoopback 判断请求来源是否本机。
func isLoopback(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// ---------- 静态资源 ----------

// hRoot 把根路径送到默认页面。两套页面都有自己的挂载前缀，这里只做跳转。
func (s *Server) hRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	// 默认进调试工作台（它是桌面版的主页）；字典页在 /dict/。
	if s.opt.Debug != nil {
		http.Redirect(w, r, s.opt.DebugBase+"/", http.StatusFound)
		return
	}
	if s.opt.Dict != nil {
		http.Redirect(w, r, s.opt.DictBase+"/", http.StatusFound)
		return
	}
	writeLanding(w, fmt.Sprintf("tt %s", s.opt.Version))
}

// ---------- 辅助 ----------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		// 头已经发出去了，只能记日志
		log.Printf("[tt] 写响应失败: %v", err)
	}
}

func writeErr(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]any{"ok": false, "error": err.Error()})
}

// decodeBody 读请求体并解析 JSON。体积上限 4 MiB（配置里可能有较长的账号列表）。
func decodeBody(r *http.Request, v any) error {
	if r.Body == nil {
		return fmt.Errorf("请求体为空")
	}
	dec := json.NewDecoder(io.LimitReader(r.Body, 4<<20))
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("请求体不是合法 JSON: %w", err)
	}
	return nil
}

// putSection 把类型化结构写回某个顶层节。
func putSection(root map[string]any, key string, v any) error {
	sec, err := structToSection(v)
	if err != nil {
		return fmt.Errorf("配置节 %s 无法序列化: %w", key, err)
	}
	root[key] = sec
	return nil
}

// structToSection 把类型化结构转成可持久化的 map（经 JSON 标签，保留 omitempty 语义）。
func structToSection(v any) (map[string]any, error) {
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

// validateHosts 校验环境清单。环境页保存前调用 —— 拦在这里，比让命令在几秒后
// 拿着一个错端口去连要好得多。
func validateHosts(h *config.Hosts) error {
	if len(h.SSHs) == 0 {
		return fmt.Errorf("至少要配置一个 SSH 环境")
	}
	seen := map[string]bool{}
	for i := range h.SSHs {
		e := &h.SSHs[i]
		where := fmt.Sprintf("第 %d 个环境", i+1)
		if strings.TrimSpace(e.Name) == "" {
			return fmt.Errorf("%s 缺少名称", where)
		}
		if e.Name != strings.TrimSpace(e.Name) {
			return fmt.Errorf("环境名 %q 首尾不能有空白", e.Name)
		}
		if seen[e.Name] {
			return fmt.Errorf("环境名 %q 重复", e.Name)
		}
		seen[e.Name] = true
		if e.Host == "" {
			return fmt.Errorf("环境 %q 缺少服务器地址", e.Name)
		}
		if e.Port < 0 || e.Port > 65535 {
			return fmt.Errorf("环境 %q 的 SSH 端口 %d 超出范围", e.Name, e.Port)
		}
		if e.User == "" {
			return fmt.Errorf("环境 %q 缺少登录账号", e.Name)
		}
		if e.DB == nil {
			continue
		}
		if e.DB.Port < 0 || e.DB.Port > 65535 {
			return fmt.Errorf("环境 %q 的数据库端口 %d 超出范围", e.Name, e.DB.Port)
		}
		switch e.DB.Type {
		case "oracle", "kingbase":
		case "":
			return fmt.Errorf("环境 %q 的数据库未选择类型", e.Name)
		default:
			return fmt.Errorf("环境 %q 的数据库类型 %q 不支持（应为 oracle 或 kingbase）", e.Name, e.DB.Type)
		}
	}
	if h.ActiveEnv != "" && !seen[h.ActiveEnv] {
		return fmt.Errorf("默认环境 %q 不在环境列表里", h.ActiveEnv)
	}
	return nil
}
