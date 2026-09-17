// Package server 提供字典页的 REST 接口与 config.json 读写:源码镜像、字典同步、
// BDL 文档目录、用户 PATH 安装。只覆盖配置管理,不含任何调试能力。
//
// SSH 连接与服务器侧数据库探测复用 internal/host(与 CLI 的 db discover 同源);
// 客户端直连测试走 internal/erpdb(与 db ping / 远程直查同链路)。
//
// 对外只暴露两个东西:构造函数 New 与 Handler()。挂载方(tt serve / internal/web)
// 用 http.StripPrefix("/dict", srv.Handler()) 把它挂在 /dict/ 下 —— 包内路径全部
// 相对挂载点,于是包内的 /api/mirror 对外是 /dict/api/mirror。
//
// 挂载前缀是必需的:同一个进程还挂着调试工作台(/debug/),两边各有自己的 /api/*,
// 不分开就会撞名。凡是两个工具都要的端点(主机/环境的读写、dbprobe/dbaccverify/
// conntest 探测、健康检查)由 internal/web 在顶层各提供一份,本包不再重复。
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"tt/internal/config"
)

// Server 字典页配置服务:config.json 读写 + REST 接口 + 后台任务状态。
type Server struct {
	cfgPath string // config.json 路径(读写;文件不存在时首次保存创建)
	dbPath  string // 数据同步目标 SQLite(由挂载方注入;空=当前目录 erp_data.db)

	// 源码镜像拉取任务(单实例:同一时间只允许一个)
	mirrorMu    sync.Mutex
	mirror      mirrorJob
	mirrorStart time.Time

	// 数据库同步任务(单实例:同一时间只允许一个)
	syncMu    sync.Mutex
	sync      dbSyncJob
	syncStart time.Time
}

// New 创建服务实例。cfgPath 是配置文件的读写位置。
//
// 监听地址不在这里:HTTP 服务由挂载方(tt serve / internal/web)统一建立,
// 本包只提供 Handler()。
func New(cfgPath string) *Server {
	return &Server{cfgPath: cfgPath}
}

// SetDBTarget 注入数据同步的默认目标(便携版:exe 同目录的 erp_data.db,由挂载方传入)。
func (s *Server) SetDBTarget(p string) { s.dbPath = p }

// defaultDBTarget 返回默认同步目标(未注入时用 config.DefaultSyncTarget)。
//
// 默认位置只有 config 里那一份实现:统一设置页会显示「默认位置:X」,
// 而这里真正往那儿写 —— 两边各算一次的话,显示的路径可能不是实际写入的那个。
func (s *Server) defaultDBTarget() string {
	if s.dbPath != "" {
		return s.dbPath
	}
	return config.DefaultSyncTarget()
}

// syncTarget 返回当前同步目标:优先 config.json 顶层 sync.target(页面可改),
// 否则用默认目标(exe 同目录)。文件不存在时由同步过程创建(含父目录)。
func (s *Server) syncTarget() string {
	if r, err := loadConfig(s.cfgPath); err == nil {
		if d := config.AbsPath(r.Sync.Target); d != "" {
			return d
		}
	}
	return s.defaultDBTarget()
}

// Handler 返回字典页的全部 HTTP 面(镜像/同步/PATH 安装/BDL 文档的 REST 接口 +
// 兜底页),所有路径相对于挂载点。
//
// tt serve 用 http.StripPrefix("/dict", srv.Handler()) 把它挂在 /dict/ 下,
// 于是包内注册的 /api/mirror 对外是 /dict/api/mirror。
//
// 服务的路径(全部返回 JSON;前端在 web/ 下):
//
//	GET    /api/mirror         镜像根 + 各环境镜像现状 + 拉取任务状态
//	PUT    /api/mirror         保存镜像根(mirror.dir)
//	POST   /api/mirror/pull    启动一次源码镜像拉取
//	GET    /api/dbsync         同步目标 + 可同步环境 + 同步任务状态
//	PUT    /api/dbsync         保存同步目标(sync.target;空=清除回默认)
//	POST   /api/dbsync         启动一次字典同步
//	GET    /api/bdldoc         BDL 文档目录
//	PUT    /api/bdldoc         保存 BDL 文档目录
//	(未知 /api/ 路径)        404 JSON(见 hUnknownAPI)
//	GET    /                  兜底页(见 hStatic)
//
// 主机/环境的读写走共享的顶层 /api/hosts;dbprobe / dbaccverify / conntest /
// 健康检查也由 internal/web 在顶层各提供一份,故本包不注册这些端点
// (它们两边都有,挂在同一进程里会撞名)。
//
// 用户 PATH 安装(/api/install)也**已移到** internal/web:把 tt 加进 PATH 是应用级
// 动作,与字典查询无关;先前挂在这里,导致共用的设置页要用它就得去调本子系统的私有 API。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/mirror", s.hMirrorGet)
	mux.HandleFunc("PUT /api/mirror", s.hMirrorPut)
	mux.HandleFunc("POST /api/mirror/pull", s.hMirrorPull)
	mux.HandleFunc("GET /api/dbsync", s.hDBSyncGet)
	mux.HandleFunc("POST /api/dbsync", s.hDBSyncPost)
	mux.HandleFunc("PUT /api/dbsync", s.hDBSyncPut)
	mux.HandleFunc("GET /api/bdldoc", s.hBdldocGet)
	mux.HandleFunc("PUT /api/bdldoc", s.hBdldocPut)
	// API 路径掉到这里说明接口不存在(写错了,或者是已经移到统一层的那些,如
	// /api/install)。必须回 404 JSON 而不是落到下面的 HTML 兜底页 —— 后者会让调用方
	// 拿到 200 + HTML,解析失败时报的错与真实原因(接口不存在)完全对不上。
	// 注意 Go 1.22 起具体模式优先于前缀模式,所以上面那些具名路由不受影响。
	mux.HandleFunc("/api/", s.hUnknownAPI)
	mux.HandleFunc("/", s.hStatic)
	return mux
}

// hUnknownAPI 未知的字典接口:明确 404,并提示它可能已经搬到统一层。
func (s *Server) hUnknownAPI(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 404, map[string]any{
		"ok":    false,
		"error": "未知的字典接口: " + r.URL.Path + "(共享类接口在 /api/ 下,如 /api/hosts、/api/install、/api/config/status)",
	})
}

// ---------- 工具 ----------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]any{"ok": false, "error": err.Error()})
}

func readBody[T any](w http.ResponseWriter, r *http.Request, out *T) bool {
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		fail(w, 400, fmt.Errorf("请求体解析失败: %v", err))
		return false
	}
	return true
}

// hStatic 兜底页:本包只提供字典页的 REST 接口,页面本身(内嵌 SPA)由 internal/web
// 统一提供(tt serve)。请求打到 /dict/ 下任何非 API 路径时给一页接口清单,
// 避免只看到光秃秃的 404。
//
// 注意:这个 "/" 兜底会遮蔽挂载点下的其它页面路由。若 internal/web 要让自己的
// SPA 接管 /dict/ 的页面路径,把 Handler 里最后那行 mux.HandleFunc("/", s.hStatic)
// 删掉即可 —— API 路由不受影响。
func (s *Server) hStatic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!doctype html><html lang="zh"><meta charset="utf-8">
<title>tt dict</title><body style="font-family:system-ui;padding:40px;line-height:1.8">
<h2>字典配置接口</h2>
<p>页面由 <code>tt serve</code> 提供;本页只说明接口。</p>
<p>API:<code>GET/PUT /api/mirror</code>、<code>POST /api/mirror/pull</code>、
<code>GET/PUT/POST /api/dbsync</code>、<code>GET/PUT /api/bdldoc</code>。</p>
<p>PATH 安装与配置状态由统一层提供:<code>/api/install</code>、<code>/api/config/status</code>。</p>
<p>主机与环境清单由共享端点 <code>/api/hosts</code> 读写。</p></body>`)
}
