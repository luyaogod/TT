package server

// 「清理」：把某个环境在本机留下的那份**可再生数据**删掉。
//
// 两个动作的边界不一样，所以分成两个端点：
//
//	POST /api/dbsync/clear  {env}  删该环境的本地字典库（<env>.db，连带 -wal/-shm/.bak）
//	POST /api/mirror/clear  {env}  删该环境的源码镜像目录（整棵）
//
// 它们都是破坏性动作，所以每条都带拒绝条件：任务在跑不倒、目录被别的环境共用不倒、
// 指向根目录/盘符不倒。删掉的是可再生数据（重新同步/重新拉取即可恢复），但**不可撤销**，
// 所以界面上都要过一道确认（前端 setAsk）。
//
// 为什么拒绝"共用目录"：镜像目录允许每个环境单独指定，两个环境填成同一个目录是能配出来的
// （见 settings 里那两行），而删一棵树会把别人的镜像一起删掉 —— 那种错误一次点击就发生，
// 没有任何提示能补救。宁可不删，让用户先把目录分开。

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"tt/internal/config"
)

// hDBSyncClear 删掉某环境的本地字典库。
func (s *Server) hDBSyncClear(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Env string `json:"env"`
	}
	if !readBody(w, r, &req) {
		return
	}
	env := strings.TrimSpace(req.Env)
	if env == "" {
		fail(w, 400, fmt.Errorf("请选择环境"))
		return
	}

	s.syncMu.Lock()
	running := s.sync.Running
	s.syncMu.Unlock()
	if running {
		fail(w, 409, fmt.Errorf("同步任务正在跑，先停止再清理"))
		return
	}

	hosts, err := config.LoadHosts(s.cfgPath)
	if err != nil {
		fail(w, 500, err)
		return
	}
	if hosts.ByName(env) == nil {
		fail(w, 400, fmt.Errorf("未找到环境 %q", env))
		return
	}

	p := s.dbPathFor(env)
	if strings.TrimSpace(p) == "" {
		fail(w, 400, fmt.Errorf("算不出环境 %q 的本地库路径", env))
		return
	}
	if !filepath.IsAbs(p) {
		fail(w, 400, fmt.Errorf("本地库路径不是绝对路径: %s", p))
		return
	}

	freed, removed, err := removeLocalDB(p)
	if err != nil {
		fail(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "path": p, "freed": freed, "removed": removed})
}

// removeLocalDB 删 <库>.db 与它的伴生文件。
//
// 伴生文件都要删：SQLite 的 -wal/-shm 里可能还压着没落盘的数据，.bak 是上一次同步
// 留下的旧库 —— 只删主文件会留一堆"看着像库"的残骸，下次排查时没人分得清哪个是真的。
func removeLocalDB(p string) (freed int64, removed int, err error) {
	for _, f := range []string{p, p + "-wal", p + "-shm", p + "-journal", p + ".bak"} {
		st, statErr := os.Stat(f)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				continue
			}
			return freed, removed, statErr
		}
		if st.IsDir() {
			continue
		}
		if rmErr := os.Remove(f); rmErr != nil {
			return freed, removed, fmt.Errorf("删除 %s 失败（可能有查询正打开着它）: %w", f, rmErr)
		}
		freed += st.Size()
		removed++
	}
	return freed, removed, nil
}

// hMirrorClear 删掉某环境的源码镜像目录（整棵）。
func (s *Server) hMirrorClear(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Env string `json:"env"`
	}
	if !readBody(w, r, &req) {
		return
	}
	env := strings.TrimSpace(req.Env)
	if env == "" {
		fail(w, 400, fmt.Errorf("请选择环境"))
		return
	}

	s.mirrorMu.Lock()
	running := s.mirror.Running
	s.mirrorMu.Unlock()
	if running {
		fail(w, 409, fmt.Errorf("拉取任务正在跑，先停止再清理"))
		return
	}

	root := ""
	if c, err := loadConfig(s.cfgPath); err == nil {
		root = c.Mirror.Dir
	}
	hosts, err := config.LoadHosts(s.cfgPath)
	if err != nil {
		fail(w, 500, err)
		return
	}
	e := hosts.ByName(env)
	if e == nil {
		fail(w, 400, fmt.Errorf("未找到环境 %q", env))
		return
	}
	dir := e.MirrorPath(root)
	if dir == "" {
		fail(w, 400, fmt.Errorf("环境 %q 没有镜像目录（先在站点管理里给它指定，或填上面的根目录）", env))
		return
	}
	if !filepath.IsAbs(dir) {
		fail(w, 400, fmt.Errorf("镜像目录不是绝对路径: %s", dir))
		return
	}

	// 别的环境也落在同一个目录 → 一次点击会把别人的镜像一起删掉，拒绝。
	for i := range hosts.SSHs {
		o := &hosts.SSHs[i]
		if o.Name == e.Name {
			continue
		}
		if p := o.MirrorPath(root); p != "" && config.SamePath(p, dir) {
			fail(w, 409, fmt.Errorf("该目录同时是环境 %q 的镜像目录；先把两行的目录分开，再清理", o.Name))
			return
		}
	}
	// 根目录本身 / 盘符根：这不是"某个环境的镜像"，整棵删掉会把别人的东西一起带走。
	if rootAbs := config.AbsPath(root); rootAbs != "" && config.SamePath(rootAbs, dir) {
		fail(w, 400, fmt.Errorf("该环境的镜像目录就是根目录本身，拒绝整棵删除；请给它单独指定一个目录"))
		return
	}
	if isDriveRoot(dir) {
		fail(w, 400, fmt.Errorf("镜像目录是盘符根目录，拒绝删除: %s", dir))
		return
	}

	st, statErr := os.Stat(dir)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			writeJSON(w, 200, map[string]any{"ok": true, "path": dir, "freed": 0, "note": "目录本来就不存在"})
			return
		}
		fail(w, 500, statErr)
		return
	}
	if !st.IsDir() {
		fail(w, 400, fmt.Errorf("%s 不是目录", dir))
		return
	}

	freed := dirSize(dir)
	if rmErr := os.RemoveAll(dir); rmErr != nil {
		fail(w, 500, fmt.Errorf("删除镜像目录失败（可能有文件正被占用）: %w", rmErr))
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "path": dir, "freed": freed})
}

// isDriveRoot 判"盘符根 / 卷根"（Windows 的 D:\、UNC 的 \\server\share）。非 Windows 上
// 卷名是空的，返回 false —— 那边由根目录那条检查兜底。
func isDriveRoot(p string) bool {
	c := filepath.Clean(p)
	vol := filepath.VolumeName(c)
	if vol == "" {
		return false
	}
	return c == vol || c == vol+string(os.PathSeparator)
}

// dirSize 递归统计目录大小（只为把"删掉了多少"回给界面；失败就按已统计到的算）。
func dirSize(root string) int64 {
	var total int64
	_ = filepath.WalkDir(root, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, ierr := d.Info(); ierr == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}
