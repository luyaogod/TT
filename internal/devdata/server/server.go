// Package server 是「开发资料」的 HTTP 面（/api/devdata/*）。
//
// 它**只做三件事**：校验入参、把长跑动作交给领域层的 Manager、把领域错误映射成
// 稳定的 HTTP 契约。真正的拉取/落盘/指纹全在 internal/devdata —— CLI 与 Web 共用
// 同一份实现，不存在"页面拉的和命令行拉的 sha 不一致"这种事。
//
// 错误信封与共享层一致：`{"ok":false,"code":...,"error":...}`。code 是稳定值，
// 调用方按 code 分支，不改文案。
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/luyaogod/TT/internal/devdata"
)

// 稳定错误码（对外契约的一部分，改了就是破坏性变更）。
const (
	CodeUsage    = "USAGE"     // 400 入参不合法（prog 非法 / only 非法 / 缺参数）
	CodeNotFound = "NOT_FOUND" // 404 环境不在清单 / 程序没拉过
	CodeConflict = "CONFLICT"  // 409 同类任务在跑 / 目标已存在 / 没任务可停
	CodeInternal = "INTERNAL"  // 500 其它
)

// Deps 领域层入口（可注入；nil 时用 internal/devdata 的真实现）。
//
// 注入的用途只有一个：**契约测试**要能断言"校验失败时根本没调领域层" ——
// 也就是"不得先建目录再报错"。为此不需要造一台假服务器，只需要换掉这几个函数。
type Deps struct {
	Mirror       func(ctx context.Context, req devdata.MirrorRequest, onProgress devdata.ProgressFunc) (*devdata.MirrorResult, error)
	Fetch        func(ctx context.Context, req devdata.FetchRequest, onProgress devdata.ProgressFunc) (*devdata.FetchResult, error)
	Status       func(ctx context.Context, req devdata.StatusRequest) (*devdata.EnvStatus, error)
	Check        func(ctx context.Context, req devdata.StatusRequest) (*devdata.ProgStatus, error)
	PlanClean    func(cfgPath, env, prog string) (*devdata.CleanPlan, error)
	ExecuteClean func(plan *devdata.CleanPlan) error
}

// Options 装配参数。
type Options struct {
	// ConfigPath config.json 路径（路径与身份解析的唯一入口）。
	ConfigPath string
	// Version tt 版本号（写进新 manifest 的 ttVersion）。
	Version string
	// Manager 进程内任务登记处；nil 时自建一个。
	Manager *devdata.Manager
	// Deps 领域层入口；零值 = 用真实现。
	Deps Deps
}

// Server 开发资料的 HTTP 面。
type Server struct {
	cfgPath string
	version string
	mgr     *devdata.Manager
	deps    Deps
}

// New 装配服务。
func New(opt Options) *Server {
	d := opt.Deps
	if d.Mirror == nil {
		d.Mirror = devdata.Mirror
	}
	if d.Fetch == nil {
		d.Fetch = devdata.Fetch
	}
	if d.Status == nil {
		d.Status = devdata.Status
	}
	if d.Check == nil {
		d.Check = devdata.Check
	}
	if d.PlanClean == nil {
		d.PlanClean = devdata.PlanClean
	}
	if d.ExecuteClean == nil {
		d.ExecuteClean = devdata.ExecuteClean
	}
	mgr := opt.Manager
	if mgr == nil {
		mgr = devdata.NewManager()
	}
	return &Server{cfgPath: opt.ConfigPath, version: opt.Version, mgr: mgr, deps: d}
}

// Manager 暴露任务登记处（测试与状态查询用）。
func (s *Server) Manager() *devdata.Manager { return s.mgr }

// Handler 返回挂在 /api/devdata/ 下的 HTTP 面（路径是绝对 /api/devdata/…）。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/devdata/status", s.hStatus)
	mux.HandleFunc("POST /api/devdata/mirror", s.hMirror)
	mux.HandleFunc("POST /api/devdata/fetch", s.hFetch)
	mux.HandleFunc("POST /api/devdata/stop", s.hStop)
	mux.HandleFunc("DELETE /api/devdata/prog", s.hDeleteProg)
	mux.HandleFunc("GET /api/devdata/check", s.hCheck)
	// 未注册的 /api/devdata/* 明确 404 JSON（不能落到 HTML 兜底页：
	// 那样调用方拿到 200 + HTML，解析失败的原因与真实原因完全对不上）
	mux.HandleFunc("/api/devdata/", func(w http.ResponseWriter, r *http.Request) {
		fail(w, http.StatusNotFound, CodeNotFound, fmt.Errorf("未知接口 %s", r.URL.Path))
	})
	return mux
}

// ---------- 端点 ----------

// hStatus 返回某环境的共用资料状态 + 程序清单 + 在跑的任务。
//
// 默认**不出网**（`?probe=1` 才向服务器核对 mtime）：设置页一打开就要出结果。
func (s *Server) hStatus(w http.ResponseWriter, r *http.Request) {
	env, ok := s.requireEnv(w, r)
	if !ok {
		return
	}
	probe := r.URL.Query().Get("probe") == "1"
	prog := strings.TrimSpace(r.URL.Query().Get("prog"))
	if prog != "" {
		if !s.validProg(w, prog) {
			return
		}
	}
	st, err := s.deps.Status(r.Context(), devdata.StatusRequest{
		CfgPath: s.cfgPath, Env: env, Probe: probe, Prog: prog,
		Jobs: s.mgr.JobsFor(env),
	})
	if err != nil {
		s.failDomain(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "env": st.Env, "dir": st.Dir, "present": st.Present,
		"dict": st.Dict, "mdl": st.Mdl, "progs": st.Progs, "jobs": st.Jobs,
		"warnings": st.Warnings,
	})
}

// hMirror 启动共用资料同步（单飞：已在跑 → 409）。
func (s *Server) hMirror(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Env   string `json:"env"`
		Only  string `json:"only"`
		Force bool   `json:"force"`
		Ent   int    `json:"ent"`
	}
	if !readBody(w, r, &req) {
		return
	}
	env, ok := s.requireEnvName(w, req.Env)
	if !ok {
		return
	}
	only := strings.ToLower(strings.TrimSpace(req.Only))
	if only != "" && only != devdata.OnlyDict && only != devdata.OnlyMdl {
		fail(w, http.StatusBadRequest, CodeUsage, fmt.Errorf("only 只支持 dict / mdl，得到 %q", req.Only))
		return
	}
	job, err := s.mgr.Start(devdata.KindMirror, env, "", func(ctx context.Context, report devdata.ProgressFunc) (any, error) {
		return s.deps.Mirror(ctx, devdata.MirrorRequest{
			CfgPath: s.cfgPath, Env: env, Ent: req.Ent, Only: only, Force: req.Force, TTVersion: s.version,
		}, report)
	})
	if err != nil {
		s.failDomain(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "job": job})
}

// hFetch 启动按程序拉取（单飞按程序：同一程序不许并发，不同程序可以）。
func (s *Server) hFetch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Env        string `json:"env"`
		Prog       string `json:"prog"`
		Module     string `json:"module"`
		WithDesign bool   `json:"withDesign"`
		Form       bool   `json:"form"`
		Force      bool   `json:"force"`
		Ent        int    `json:"ent"`
	}
	if !readBody(w, r, &req) {
		return
	}
	env, ok := s.requireEnvName(w, req.Env)
	if !ok {
		return
	}
	prog := strings.TrimSpace(req.Prog)
	if !s.validProg(w, prog) {
		return
	}
	job, err := s.mgr.Start(devdata.KindFetch, env, prog, func(ctx context.Context, report devdata.ProgressFunc) (any, error) {
		return s.deps.Fetch(ctx, devdata.FetchRequest{
			CfgPath: s.cfgPath, Env: env, Prog: prog, Module: req.Module, Ent: req.Ent,
			WithDesign: req.WithDesign, Form: req.Form, Force: req.Force, TTVersion: s.version,
		}, report)
	})
	if err != nil {
		s.failDomain(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "job": job})
}

// hStop 取消正在跑的任务。
func (s *Server) hStop(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Env  string `json:"env"`
		Kind string `json:"kind"`
		Prog string `json:"prog"`
	}
	if !readBody(w, r, &req) {
		return
	}
	if _, ok := s.requireEnvName(w, req.Env); !ok {
		return
	}
	kind := strings.TrimSpace(req.Kind)
	if kind == "" {
		kind = devdata.KindMirror
	}
	if kind != devdata.KindMirror && kind != devdata.KindFetch {
		fail(w, http.StatusBadRequest, CodeUsage, fmt.Errorf("kind 只支持 mirror / fetch，得到 %q", req.Kind))
		return
	}
	if kind == devdata.KindFetch && !s.validProg(w, strings.TrimSpace(req.Prog)) {
		return
	}
	if err := s.mgr.Stop(kind, strings.TrimSpace(req.Env), strings.TrimSpace(req.Prog)); err != nil {
		s.failDomain(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// hDeleteProg 删除某程序的材料目录（必须 ?yes=1，默认拒绝 —— 删材料不可撤销）。
func (s *Server) hDeleteProg(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("yes") != "1" {
		fail(w, http.StatusBadRequest, CodeUsage, errors.New("删除不可撤销；确认请加 ?yes=1"))
		return
	}
	env, ok := s.requireEnv(w, r)
	if !ok {
		return
	}
	prog := strings.TrimSpace(r.URL.Query().Get("prog"))
	if !s.validProg(w, prog) {
		return
	}
	plan, err := s.deps.PlanClean(s.cfgPath, env, prog)
	if err != nil {
		s.failDomain(w, err)
		return
	}
	if err := s.deps.ExecuteClean(plan); err != nil {
		s.failDomain(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "removed": plan})
}

// hCheck 只比对服务器（`?env=&prog=`），输出 stale 明细。
func (s *Server) hCheck(w http.ResponseWriter, r *http.Request) {
	env, ok := s.requireEnv(w, r)
	if !ok {
		return
	}
	prog := strings.TrimSpace(r.URL.Query().Get("prog"))
	if !s.validProg(w, prog) {
		return
	}
	st, err := s.deps.Check(r.Context(), devdata.StatusRequest{CfgPath: s.cfgPath, Env: env, Prog: prog})
	if err != nil {
		s.failDomain(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "prog": st})
}

// ---------- 校验 ----------

// requireEnv 取并校验 `?env=`（空 → 400，不在清单 → 404）。
func (s *Server) requireEnv(w http.ResponseWriter, r *http.Request) (string, bool) {
	return s.requireEnvName(w, r.URL.Query().Get("env"))
}

// requireEnvName 校验环境名：空 → 400，不在配置清单 → 404。
//
// 走领域层的 EnvDir 而不是自己读配置：环境名合法的判据只有一处（hosts.sshs 清单），
// 两处各判一次迟早会出现"Web 认为合法、CLI 认为非法"。而且这一步**不落盘、不拨号**。
func (s *Server) requireEnvName(w http.ResponseWriter, env string) (string, bool) {
	name := strings.TrimSpace(env)
	if name == "" {
		fail(w, http.StatusBadRequest, CodeUsage, errors.New("缺少 env 参数（环境一律点名）"))
		return "", false
	}
	if _, err := devdata.EnvDir(s.cfgPath, name); err != nil {
		s.failDomain(w, err)
		return "", false
	}
	return name, true
}

// validProg 校验程序名（必须在任何落盘/拨号之前）。
func (s *Server) validProg(w http.ResponseWriter, prog string) bool {
	if err := devdata.ValidProgName(prog); err != nil {
		s.failDomain(w, err)
		return false
	}
	return true
}

// failDomain 把领域错误映射成 HTTP 契约。
//
// 映射只按哨兵错误（errors.Is），不按文案 —— 改文案不该改状态码。
func (s *Server) failDomain(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, devdata.ErrEnvNotFound), errors.Is(err, devdata.ErrProgNotFound),
		errors.Is(err, devdata.ErrMaterialIncomplete):
		fail(w, http.StatusNotFound, CodeNotFound, err)
	case errors.Is(err, devdata.ErrInvalidProg), errors.Is(err, devdata.ErrInvalidEnv),
		errors.Is(err, devdata.ErrInvalidOpt), errors.Is(err, devdata.ErrModuleUnknown):
		fail(w, http.StatusBadRequest, CodeUsage, err)
	case errors.Is(err, devdata.ErrBusy), errors.Is(err, devdata.ErrNoJob), errors.Is(err, devdata.ErrExists):
		fail(w, http.StatusConflict, CodeConflict, err)
	case errors.Is(err, context.Canceled):
		// 取消是正常结果，不是服务的错；但走到 HTTP 这一层说明请求本身在取消
		fail(w, http.StatusRequestTimeout, CodeConflict, errors.New("请求已取消"))
	default:
		fail(w, http.StatusInternalServerError, CodeInternal, err)
	}
}

// ---------- 工具 ----------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

func fail(w http.ResponseWriter, code int, errCode string, err error) {
	writeJSON(w, code, map[string]any{"ok": false, "code": errCode, "error": err.Error()})
}

func readBody[T any](w http.ResponseWriter, r *http.Request, out *T) bool {
	if r.Body == nil {
		fail(w, http.StatusBadRequest, CodeUsage, errors.New("请求体为空"))
		return false
	}
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		fail(w, http.StatusBadRequest, CodeUsage, fmt.Errorf("请求体解析失败: %v", err))
		return false
	}
	return true
}
