// 配置里那些"路径型"取值的公共解析。
//
// 合并前这套逻辑散在 internal/dict/server 的 config.go / bdldoc.go / dbsync.go 里
// (absPath + 各处自己 os.Stat),而设置页搬进调试工作台之后也需要同一份判断 ——
// 两边各算一遍的话,同一个 sync.target 会在两个页面上解析成不同结果,那是最难查的一类 bug。
// 所以收敛到这里,谁要用谁来取。
package config

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultSyncFileName 字典同步产出的本地 SQLite 文件名。
const DefaultSyncFileName = "erp_data.db"

// AbsPath 相对路径转绝对;空串原样返回空串(表示"未配置")。
func AbsPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}

// DirStatus 一个"目录型"配置值的状态:配置里的值 + 它在本机是否真的存在。
type DirStatus struct {
	// Dir 绝对路径;空 = 未配置。
	Dir string `json:"dir"`
	// Exists 该目录是否存在于本机。
	Exists bool `json:"exists"`
}

// DirStatusOf 解析一个目录型配置值并检查它是否存在。
func DirStatusOf(raw string) DirStatus {
	st := DirStatus{Dir: AbsPath(raw)}
	if st.Dir == "" {
		return st
	}
	if fi, err := os.Stat(st.Dir); err == nil && fi.IsDir() {
		st.Exists = true
	}
	return st
}

// FileStatus 一个"文件型"配置值的状态。
type FileStatus struct {
	// Target 当前生效值(绝对路径)。
	Target string `json:"target"`
	// Configured 配置里显式写下的值;空 = 用缺省位置。
	Configured string `json:"configured"`
	// DefaultTarget 缺省位置(未配置时生效的那个)。
	DefaultTarget string `json:"defaultTarget"`
	// Exists 生效值指向的文件当前是否存在。
	Exists bool `json:"exists"`
}

// FileStatusOf 汇总一个文件型配置值:生效值、显式配置值、缺省值、是否存在。
func FileStatusOf(configured, defaultTarget string) FileStatus {
	st := FileStatus{
		Configured:    AbsPath(configured),
		DefaultTarget: defaultTarget,
	}
	st.Target = st.Configured
	if st.Target == "" {
		st.Target = defaultTarget
	}
	if st.Target != "" {
		if _, err := os.Stat(st.Target); err == nil {
			st.Exists = true
		}
	}
	return st
}

// DefaultSyncTarget 同步目标的缺省位置:当前目录下的 erp_data.db(绝对化)。
//
// 与 `tt dict -d` 的解析顺序(cwd → exe 目录)第一条候选一致,所以
// "服务在这里同步、命令行在这里读"天然对得上。桌面外壳启动时 cwd 就是数据目录
// (见 desktop/main.js 的 spawn cwd),于是桌面场景下它自然落在配置旁边。
func DefaultSyncTarget() string {
	if abs, err := filepath.Abs(DefaultSyncFileName); err == nil {
		return abs
	}
	return DefaultSyncFileName
}
