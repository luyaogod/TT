// Package server 提供字典侧的 REST 接口:源码镜像拉取、字典同步、BDL 文档目录。
// 只覆盖这几件事,不含任何调试能力。
//
// 它**不再对应一个独立页面** —— 合并前字典页有自己的 SPA(挂在 /dict/),本包的端点
// 也挂在 /dict/api/ 下;那个页面已并入调试工作台里的统一设置页,于是本包的端点直接挂到
// 共享层 /api/ 下(见 internal/web 的 routes)。包结构保留是因为它持有两个长跑任务的
// 状态与处理器,与"页面挂在哪"无关。
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

// Handler 返回字典侧的 HTTP 面(镜像/同步/BDL 文档的 REST 接口)。
//
// 路径是**绝对**的(/api/…),由 internal/web 直接挂在共享层 /api/ 下,不经 StripPrefix。
//
// 服务的路径(全部返回 JSON):
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
//
// 主机/环境的读写走共享的顶层 /api/hosts;dbprobe / dbaccverify / conntest、
// 健康检查、配置派生状态、PATH 安装也由 internal/web 各提供一份,故本包不注册这些端点
// (它们两边都有,挂在同一进程里会撞名)。
//
// 用户 PATH 安装(/api/install)也**已移到** internal/web:把 tt 加进 PATH 是应用级
// 动作,与字典查询无关。
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
	return mux
}

// hUnknownAPI 未注册的 /api/ 路径:明确 404。
//
// 这条兜底挂在共享层 /api/ 下,所以它也会接到与字典无关的路径 —— 文案因此写得通用些。
// 关键是**不能**落到 HTML 页:那样调用方拿到 200 + HTML,解析失败时报的错与真实原因
// (接口不存在)完全对不上。
func (s *Server) hUnknownAPI(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 404, map[string]any{
		"ok":    false,
		"error": "未知接口: " + r.URL.Path,
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
