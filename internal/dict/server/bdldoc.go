package server

import (
	"net/http"
	"os"

	"tt/internal/config"
)

// bdldocStatus BDL(4GL)语言文档目录的当前配置(GET/PUT /api/bdldoc)。
type bdldocStatus struct {
	OK         bool   `json:"ok"`
	Dir        string `json:"dir"`        // 当前配置的目录(绝对路径;空=未设置)
	Exists     bool   `json:"exists"`     // 该目录是否存在于本机
	ConfigPath string `json:"configPath"` // 写入的配置文件
}

func (s *Server) bdldocStatus() bdldocStatus {
	st := bdldocStatus{OK: true, ConfigPath: s.cfgPath}
	if root, err := loadConfig(s.cfgPath); err == nil {
		st.Dir = config.AbsPath(root.Bdldoc.Dir)
	}
	if st.Dir != "" {
		if fi, err := os.Stat(st.Dir); err == nil && fi.IsDir() {
			st.Exists = true
		}
	}
	return st
}

// hBdldocGet 返回当前 BDL 文档目录与是否存在于本机(等价 tt dict bdldoc dir)。
func (s *Server) hBdldocGet(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.bdldocStatus())
}

// hBdldocPut 写入 config.json 顶层 bdldoc.dir(其余节与本节未知子键原样保留;
// 只改设置,不移动文档文件)。
func (s *Server) hBdldocPut(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Dir string `json:"dir"`
	}
	if !readBody(w, r, &req) {
		return
	}
	abs := config.AbsPath(req.Dir)
	if abs == "" {
		writeJSON(w, 200, map[string]any{"ok": false, "error": "请填写 BDL 文档目录"})
		return
	}
	if err := config.EditSection(s.cfgPath, "bdldoc", nil, func(sec map[string]any) error {
		sec["dir"] = abs
		return nil
	}); err != nil {
		fail(w, 500, err)
		return
	}
	writeJSON(w, 200, s.bdldocStatus())
}
