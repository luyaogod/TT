package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"tt/internal/config"
	"tt/internal/host"
)

// mirrorJob 一次(或最近一次)源码镜像拉取的状态;前端据它画进度条。
type mirrorJob struct {
	Running   bool   `json:"running"`
	Env       string `json:"env"`
	Full      bool   `json:"full"`
	Phase     string `json:"phase"` // connect|probe|pack|download|done|error
	Message   string `json:"message"`
	Bytes     int64  `json:"bytes"`
	Total     int64  `json:"total"` // 下载阶段=归档总字节;0=未知(pack 阶段)
	Files     int    `json:"files"`
	Pruned    int    `json:"pruned"` // 本地清理掉的残留备份文件数
	Elapsed   string `json:"elapsed"`
	Error     string `json:"error,omitempty"`
	Note      string `json:"note,omitempty"`
	Done      bool   `json:"done"`
	StartedAt string `json:"startedAt,omitempty"`
}

// mirrorEnv 单个环境的镜像现状。
type mirrorEnv struct {
	Name  string `json:"name"`
	Zone  string `json:"zone"`
	Path  string `json:"path"`
	Ready bool   `json:"ready"` // 本地已有完整基线(增量前提)
}

type mirrorResp struct {
	MirrorDir string      `json:"mirrorDir"`
	ActiveEnv string      `json:"activeEnv"` // 默认环境(前端下拉默认选中)
	Envs      []mirrorEnv `json:"envs"`
	Job       mirrorJob   `json:"job"`
}

// hMirrorGet 返回镜像根、各环境镜像现状与当前/最近一次拉取任务(前端轮询用)。
func (s *Server) hMirrorGet(w http.ResponseWriter, r *http.Request) {
	dir := ""
	if root, err := loadConfig(s.cfgPath); err == nil {
		dir = absPath(root.Mirror.Dir)
	}
	resp := mirrorResp{MirrorDir: dir, Envs: []mirrorEnv{}}
	if hosts, err := config.LoadHosts(s.cfgPath); err == nil {
		resp.ActiveEnv = hosts.ActiveEnv
		for i := range hosts.SSHs {
			e := &hosts.SSHs[i]
			resp.Envs = append(resp.Envs, mirrorEnv{
				Name:  e.Name,
				Zone:  e.Zone,
				Path:  host.MirrorEnvDir(dir, e.Name),
				Ready: host.MirrorReady(dir, e.Name),
			})
		}
	}
	resp.Job = s.mirrorSnapshot()
	writeJSON(w, 200, resp)
}

// hMirrorPut 保存镜像根目录(顶层 mirror.dir;其余节与本节未知子键原样保留;
// 目录由拉取时创建)。
func (s *Server) hMirrorPut(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Dir string `json:"dir"`
	}
	if !readBody(w, r, &req) {
		return
	}
	abs := absPath(req.Dir)
	if abs == "" {
		writeJSON(w, 200, map[string]any{"ok": false, "error": "请填写镜像根目录"})
		return
	}
	if err := config.EditSection(s.cfgPath, "mirror", nil, func(sec map[string]any) error {
		sec["dir"] = abs
		return nil
	}); err != nil {
		fail(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "mirrorDir": abs})
}

// hMirrorPull 启动一次拉取(单实例:已有任务在跑返回 409);返回后前端轮询 /api/mirror。
func (s *Server) hMirrorPull(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Env  string `json:"env"`
		Full bool   `json:"full"`
	}
	if !readBody(w, r, &req) {
		return
	}
	envName := strings.TrimSpace(req.Env)
	if envName == "" {
		fail(w, 400, fmt.Errorf("请选择环境"))
		return
	}
	dir := ""
	if root, err := loadConfig(s.cfgPath); err == nil {
		dir = absPath(root.Mirror.Dir)
	}
	if dir == "" {
		writeJSON(w, 200, map[string]any{"ok": false, "error": "请先设置镜像根目录"})
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

	s.mirrorMu.Lock()
	if s.mirror.Running {
		s.mirrorMu.Unlock()
		fail(w, 409, fmt.Errorf("已有拉取任务进行中"))
		return
	}
	s.mirror = mirrorJob{
		Running:   true,
		Env:       e.Name,
		Full:      req.Full,
		Phase:     "connect",
		Message:   "准备中…",
		StartedAt: time.Now().Format(time.RFC3339),
	}
	s.mirrorStart = time.Now()
	s.mirrorMu.Unlock()

	go s.runMirror(e, dir, req.Full)
	writeJSON(w, 200, map[string]any{"ok": true})
}

// runMirror 后台执行拉取,把 host 的进度回调同步到任务状态。
func (s *Server) runMirror(e *host.NamedSsh, dir string, full bool) {
	onProgress := func(p host.MirrorProgress) {
		s.mirrorMu.Lock()
		if p.Phase != "" {
			s.mirror.Phase = p.Phase
		}
		if p.Message != "" {
			s.mirror.Message = p.Message
		}
		if p.Bytes > 0 {
			s.mirror.Bytes = p.Bytes
		}
		if p.Total > 0 {
			s.mirror.Total = p.Total
		}
		if p.Files > 0 {
			s.mirror.Files = p.Files
		}
		s.mirror.Elapsed = time.Since(s.mirrorStart).Round(time.Second).String()
		s.mirrorMu.Unlock()
	}

	st, err := host.MirrorPullProgress(e, dir, full, onProgress)

	s.mirrorMu.Lock()
	defer s.mirrorMu.Unlock()
	s.mirror.Running = false
	s.mirror.Done = true
	s.mirror.Elapsed = time.Since(s.mirrorStart).Round(100 * time.Millisecond).String()
	if err != nil {
		s.mirror.Phase = "error"
		s.mirror.Error = err.Error()
		return
	}
	s.mirror.Phase = "done"
	s.mirror.Full = st.Full // 可能因缺基线或白名单升级而自动转全量,以实际为准
	s.mirror.Bytes = st.Bytes
	s.mirror.Total = st.Bytes
	s.mirror.Files = st.Files
	s.mirror.Pruned = st.Pruned
	if st.Note != "" {
		s.mirror.Note = st.Note
		s.mirror.Message = st.Note
	} else {
		s.mirror.Message = fmt.Sprintf("完成:%d 个文件,共 %s", st.Files, humanBytes(st.Bytes))
		if st.Pruned > 0 {
			s.mirror.Message += fmt.Sprintf(";清理本地备份 %d 个", st.Pruned)
		}
	}
}

// mirrorSnapshot 取当前(或最近一次)拉取状态副本。
func (s *Server) mirrorSnapshot() mirrorJob {
	s.mirrorMu.Lock()
	defer s.mirrorMu.Unlock()
	return s.mirror
}

// humanBytes 字节数人性化(服务器侧消息用)。
func humanBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
