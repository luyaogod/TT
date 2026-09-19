package web

import (
	"net/http"
	"os"
	"path/filepath"

	"tt/internal/config"
	"tt/internal/pathinstall"
)

// configStatus 是"配置里那些路径型取值"的派生状态:值本身在 /api/hosts 里,
// 这里补上**只在服务端才能算出来的部分** —— 目录/文件是否真的存在、生效值是哪个。
//
// 为什么单独一个端点而不是塞进 /api/hosts:这些判断要 os.Stat,而 /api/hosts 是
// 设置页每次进入设置都会读的热路径;把它们掺进去等于每次读配置都摸一遍磁盘。
//
// 为什么放在共享层而不是字典子系统的私有 API:统一设置页的「数据字典」分区要显示这些
// 状态,而设置页是**共用**的 —— 让它去调 /dict/api/* 会把它绑死在字典子系统上
// (路由本身支持只有调试子系统的部署,那时数据字典卡片就整个打不开了)。
type configStatus struct {
	OK bool `json:"ok"`
	// Config 配置文件绝对路径。
	Config string `json:"config"`
	// Mirror 源码镜像根目录。
	Mirror config.DirStatus `json:"mirror"`
	// Bdldoc BDL 文档目录。
	Bdldoc config.DirStatus `json:"bdldoc"`
	// Sync 字典同步的本地 SQLite。
	Sync config.FileStatus `json:"sync"`
	// Install 命令行安装状态(用户 PATH),不是 config.json 里的东西。
	Install pathinstall.Status `json:"install"`
	// Tzs .tzs 表单引擎的运行时依赖:引擎 exe(随包分发)、设计器目录与工作区(用户配的)。
	Tzs config.TzsStatus `json:"tzs"`
}

// defaultEngineExe 引擎 exe 的缺省位置:<tt.exe 所在目录>\tzs\tzs-server.exe。
// 与 internal/dev/cli 那侧是同一条规则 —— 那边算出来是为了真的启动它,这里只是为了
// 让设置页能显示它在不在。放同一处的理由:两边算出不同的路径,设置页就会说"有"而命令说"没有"。
func defaultEngineExe() string {
	self, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(self), "tzs", "tzs-server.exe")
}

// hConfigStatus 返回上述派生状态。配置读不到时不报错,返回缺省值 —— 设置页在没有
// 配置文件的首次运行下也要能渲染。
func (s *Server) hConfigStatus(w http.ResponseWriter, r *http.Request) {
	path, _ := s.configPath(true)
	st := configStatus{OK: true, Config: path, Sync: config.FileStatusOf("", s.syncDefaultTarget())}

	var tzsCfg config.TzsSettings
	root, err := config.Load(path)
	if err == nil {
		st.Mirror = config.DirStatusOf(root.Mirror.Dir)
		st.Bdldoc = config.DirStatusOf(root.Bdldoc.Dir)
		st.Sync = config.FileStatusOf(root.Sync.Target, s.syncDefaultTarget())
		tzsCfg = root.Tzs
	}
	// 引擎 exe 的状态与配置读没读到无关(它随包分发),所以放在分支外。
	st.Tzs = config.TzsStatusOf(tzsCfg, defaultEngineExe())
	st.Install = pathinstall.Get()
	writeJSON(w, http.StatusOK, st)
}

// syncDefaultTarget 同步目标的缺省位置。由 Options 注入以便与字典子系统取同一个值
// (两边各算一次的话,设置页显示的"默认位置"可能和字典子系统实际写入的不是同一个)。
func (s *Server) syncDefaultTarget() string {
	if s.opt.SyncDefaultTarget != "" {
		return s.opt.SyncDefaultTarget
	}
	return config.DefaultSyncTarget()
}
