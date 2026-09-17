package web

import (
	"net/http"

	"tt/internal/config"
	"tt/internal/pathinstall"
)

// configStatus 是"配置里那些路径型取值"的派生状态:值本身在 /api/hosts 里,
// 这里补上**只在服务端才能算出来的部分** —— 目录/文件是否真的存在、生效值是哪个。
//
// 为什么单独一个端点而不是塞进 /api/hosts:这些判断要 os.Stat,而 /api/hosts 是
// 两套页面每次进入设置都会读的热路径;把它们掺进去等于每次读配置都摸一遍磁盘。
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
}

// hConfigStatus 返回上述派生状态。配置读不到时不报错,返回缺省值 —— 设置页在没有
// 配置文件的首次运行下也要能渲染。
func (s *Server) hConfigStatus(w http.ResponseWriter, r *http.Request) {
	path, _ := s.configPath(true)
	st := configStatus{OK: true, Config: path, Sync: config.FileStatusOf("", s.syncDefaultTarget())}

	root, err := config.Load(path)
	if err == nil {
		st.Mirror = config.DirStatusOf(root.Mirror.Dir)
		st.Bdldoc = config.DirStatusOf(root.Bdldoc.Dir)
		st.Sync = config.FileStatusOf(root.Sync.Target, s.syncDefaultTarget())
	}
	st.Install = pathinstall.Get()
	writeJSON(w, http.StatusOK, st)
}

// syncDefaultTarget 同步目标的缺省位置。由 Options 注入以便与字典子系统取同一个值
// (两边各算一次的话,设置页显示的"默认位置"可能和字典页实际写入的不是同一个)。
func (s *Server) syncDefaultTarget() string {
	if s.opt.SyncDefaultTarget != "" {
		return s.opt.SyncDefaultTarget
	}
	return config.DefaultSyncTarget()
}
