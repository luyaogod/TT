package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"tt/internal/config"
	"tt/internal/dict/dbsync"
	"tt/internal/host"
)

// dbSyncJob 一次(或最近一次)数据同步任务的状态;前端据它画进度。
type dbSyncJob struct {
	Running    bool   `json:"running"`
	Env        string `json:"env"`
	Phase      string `json:"phase"` // open|table|index|replace|done|error
	Message    string `json:"message"`
	Table      string `json:"table"`
	TableIndex int    `json:"tableIndex"` // 当前第几张表(1-based)
	TableTotal int    `json:"tableTotal"`
	TableRows  int    `json:"tableRows"` // 当前表已写入行数
	TotalRows  int    `json:"totalRows"` // 累计行数
	Tables     int    `json:"tables"`    // 完成的表数
	Elapsed    string `json:"elapsed"`
	Target     string `json:"target"` // 目标 SQLite
	Backup     string `json:"backup,omitempty"`
	Warning    string `json:"warning,omitempty"`
	Error      string `json:"error,omitempty"`
	Done       bool   `json:"done"`
	StartedAt  string `json:"startedAt,omitempty"`
}

// dbSyncEnv 可用于同步的环境(挂了数据库的环境)。
type dbSyncEnv struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Address string `json:"address"`
}

type dbSyncResp struct {
	Target        string      `json:"target"`        // 当前生效目标(写入位置)
	Configured    string      `json:"configured"`    // config.json 顶层 sync.target(空=用默认)
	DefaultTarget string      `json:"defaultTarget"` // 默认目标(exe 同目录;便携版自带)
	Exists        bool        `json:"exists"`        // 目标文件当前是否已存在
	ActiveEnv     string      `json:"activeEnv"`
	Envs          []dbSyncEnv `json:"envs"`
	Job           dbSyncJob   `json:"job"`
}

// dbSyncView 汇总同步页面需要的状态(目标/环境/任务)。
func (s *Server) dbSyncView() dbSyncResp {
	target := s.syncTarget()
	resp := dbSyncResp{Target: target, DefaultTarget: s.defaultDBTarget(), Envs: []dbSyncEnv{}}
	if root, err := loadConfig(s.cfgPath); err == nil {
		resp.Configured = absPath(root.Sync.Target)
	}
	if _, err := os.Stat(target); err == nil {
		resp.Exists = true
	}
	if hosts, err := config.LoadHosts(s.cfgPath); err == nil {
		resp.ActiveEnv = hosts.ActiveEnv
		for i := range hosts.SSHs {
			e := &hosts.SSHs[i]
			if e.DB == nil {
				continue // 未挂库的环境无法同步
			}
			resp.Envs = append(resp.Envs, dbSyncEnv{Name: e.Name, Type: e.DB.Type, Address: e.DB.Address()})
		}
	}
	resp.Job = s.syncSnapshot()
	return resp
}

// hDBSyncGet 返回同步目标、可用环境与当前/最近一次任务(前端轮询用)。
func (s *Server) hDBSyncGet(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.dbSyncView())
}

// hDBSyncPut 设置同步目标(config.json 顶层 sync.target);target 为空 = 清除,回到默认(exe 同目录)。
func (s *Server) hDBSyncPut(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Target string `json:"target"`
	}
	if !readBody(w, r, &req) {
		return
	}
	t := strings.TrimSpace(req.Target)
	abs := absPath(t)
	if t != "" && abs == "" {
		fail(w, 400, fmt.Errorf("解析路径失败: %q", req.Target))
		return
	}
	if err := config.Edit(s.cfgPath, nil, func(root map[string]any) error {
		if t == "" {
			delete(root, "sync")
			return nil
		}
		config.SectionOrEmpty(root, "sync")["target"] = abs
		return nil
	}); err != nil {
		fail(w, 500, err)
		return
	}
	writeJSON(w, 200, s.dbSyncView())
}

// hDBSyncPost 启动一次同步(单实例:已有任务在跑返回 409);返回后前端轮询 /api/dbsync。
func (s *Server) hDBSyncPost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Env string `json:"env"`
	}
	if !readBody(w, r, &req) {
		return
	}
	envName := strings.TrimSpace(req.Env)
	if envName == "" {
		fail(w, 400, fmt.Errorf("请选择环境"))
		return
	}
	hosts, err := config.LoadHosts(s.cfgPath)
	if err != nil {
		fail(w, 500, err)
		return
	}
	e := hosts.ByName(envName)
	if e == nil {
		fail(w, 400, fmt.Errorf("未找到环境 %q", envName))
		return
	}
	if e.DB == nil {
		fail(w, 400, fmt.Errorf("环境 %q 未配置数据库", envName))
		return
	}

	target := s.syncTarget()
	s.syncMu.Lock()
	if s.sync.Running {
		s.syncMu.Unlock()
		fail(w, 409, fmt.Errorf("已有同步任务进行中"))
		return
	}
	s.sync = dbSyncJob{
		Running:   true,
		Env:       e.Name,
		Phase:     "open",
		Message:   "准备中…",
		Target:    target,
		StartedAt: time.Now().Format(time.RFC3339),
	}
	s.syncStart = time.Now()
	s.syncMu.Unlock()

	go s.runDBSync(e, target)
	writeJSON(w, 200, map[string]any{"ok": true})
}

// runDBSync 后台执行同步,把 dbsync 的进度回调同步到任务状态。
func (s *Server) runDBSync(e *host.NamedSsh, target string) {
	onProgress := func(p dbsync.Progress) {
		s.syncMu.Lock()
		if p.Phase != "" {
			s.sync.Phase = p.Phase
		}
		if p.Message != "" {
			s.sync.Message = p.Message
		}
		if p.Table != "" {
			s.sync.Table = p.Table
		}
		if p.TableIndex > 0 {
			s.sync.TableIndex = p.TableIndex
		}
		if p.TableTotal > 0 {
			s.sync.TableTotal = p.TableTotal
		}
		s.sync.TableRows = p.TableRows
		if p.TotalRows > 0 {
			s.sync.TotalRows = p.TotalRows
		}
		s.sync.Elapsed = time.Since(s.syncStart).Round(time.Second).String()
		s.syncMu.Unlock()
	}

	st, err := dbsync.Run(context.Background(), *e.DB, target, nil, onProgress)

	s.syncMu.Lock()
	defer s.syncMu.Unlock()
	s.sync.Running = false
	s.sync.Done = true
	s.sync.Elapsed = time.Since(s.syncStart).Round(100 * time.Millisecond).String()
	if err != nil {
		s.sync.Phase = "error"
		s.sync.Error = err.Error()
		return
	}
	s.sync.Phase = "done"
	s.sync.Tables = st.Tables
	s.sync.TotalRows = st.Rows
	s.sync.Backup = st.Backup
	if len(st.Warnings) > 0 {
		s.sync.Warning = strings.Join(st.Warnings, "\n")
	}
	s.sync.Message = fmt.Sprintf("同步完成:%d 张表,共 %d 行", st.Tables, st.Rows)
}

// syncSnapshot 取当前(或最近一次)同步状态副本。
func (s *Server) syncSnapshot() dbSyncJob {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()
	return s.sync
}
