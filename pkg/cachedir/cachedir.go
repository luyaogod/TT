// Package cachedir 定义数据目录下的**缓存**子目录：可再生的中间数据，
// 删了会自愈，不删会一直涨。
//
// 它与 config.json 同居一个数据目录，但性质完全不同 —— config.json 是用户数据
//（口令、环境、路径），删了就没了；这里列出的只是跑出来的副本与快照。
// "清缓存"必须只清后者，所以哪些算缓存在这里定义一次，CLI、Web、启动清理
// 都从这一份拿，免得某处漏了或某处手滑把配置一起删了。
//
// 不列入的：
//   - config.json      用户数据
//   - .tt-serve.json   **运行中**服务的 pid 文件，删了 `tt serve --stop` 就找不着它
//   - .tt-serve.log    日志（排障要看，且不影响正确性；想清可以用系统手段）
package cachedir

import "path/filepath"

// Subdirs 数据目录下的缓存子目录清单。
var Subdirs = []string{
	"ents",      // 企业目录快照(ENT→账号),10 分钟新鲜期,过期自动重查
	"srccache",  // 调试时的源码镜像,只活一轮调试
	"execlog",   // 调试执行的大输出落盘副本
	"debug-bps", // 断点存档
	"spill",     // 查询被截断时落盘的完整结果(见 tt dict 的返回上限)
	"update",    // tt 自身的更新：下载的安装包与最近一次检查结果(删了只是下次重下)
	"runtime",   // 登录后回读的 TOP/ERP/COM（tt file 用；30 分钟新鲜期，可 --refresh 重探）
}

// Dir 数据目录下的缓存子目录。name 必须在 Subdirs 清单里，
// 不在则返回空串 —— 缓存名只允许这份清单里的，调用方写错名字当场得到空串
//（而不是悄悄建出一个清单外的新目录）。凡是 <数据目录>/<名字> 这种拼接都必须走它。
func Dir(dataDir, name string) string {
	if dataDir == "" {
		return ""
	}
	for _, n := range Subdirs {
		if n == name {
			return filepath.Join(dataDir, name)
		}
	}
	return ""
}
