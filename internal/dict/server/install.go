package server

import (
	"net/http"

	"tt/internal/pathinstall"
)

// ---------- 命令行安装(把可执行文件目录加入用户 PATH) ----------
//
// 实现在 pathinstall 包(与 CLI 的 `tt dict install path` 共用同一份)。

// hInstallGet 返回安装状态:可执行文件/目录、是否已在用户 PATH。
func (s *Server) hInstallGet(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, pathinstall.Get())
}

// hInstallAdd 把当前可执行文件所在目录加入用户 PATH(幂等,用户级无需管理员)。
func (s *Server) hInstallAdd(w http.ResponseWriter, r *http.Request) {
	st, err := pathinstall.Add()
	if err != nil {
		fail(w, 500, err)
		return
	}
	writeJSON(w, 200, st)
}

// hInstallRemove 从用户 PATH 移除该目录(幂等)。
func (s *Server) hInstallRemove(w http.ResponseWriter, r *http.Request) {
	st, err := pathinstall.Remove()
	if err != nil {
		fail(w, 500, err)
		return
	}
	writeJSON(w, 200, st)
}
