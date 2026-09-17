// settings.go —— tdev 与 tt 统一配置（config.json 的 `tdev` 节）的**唯一接缝**。
//
// 原 TDev 完全没有配置文件：每个参数都由一次性的命令行 flag 传递，这条原则不变。
// 合并进 tt 之后新增了一个可选的 `tdev` 节，它只用来在**省略 flag 时**补一个默认值：
//
//	workspaceSuffix   tzc export 默认工作区后缀（缺省 -ws）
//	defaultOut        tzs export 默认解压目录
//
// 这是 tdev 唯一一处「行为可配置」的地方，所以三条纪律写死在这里：
//
//  1. **命令行 flag 永远优先**。本文件的函数只在 flag 取到空串时才被调用；
//     只要用户显式写了 -o，配置里的值一律不参与运算 —— 一次性的显式意图
//     不该被一个上次顺手写下的持久默认值盖掉（这与 A1「唯一编辑界面」同源：
//     改了什么是用户当场说的算）。
//  2. **绝不因为配置失败而失败**。配置文件缺失、定位不到、JSON 非法、没有
//     `tdev` 节 —— 一律静默退回原来的硬编码行为（-ws / <包所在目录>/<程序名>-unzip）。
//     没有配置文件的 tdev 用户必须照常可用：tdev 不要求、也不假设存在配置。
//  3. **不进热路径**。只有「确实要用默认值」的那一处（-o 缺省）才读一次配置，
//     解析失败也只是拿到零值，不会让任何命令中途失败。
package cli

import "tt/internal/config"

// loadTdevSettings 读取统一配置里的 `tdev` 节；任何一步不成立都返回零值。
//
// 零值即「没有配置」：调用方据此回退到硬编码默认值。这里的错误**不透出** ——
// 在 tdev 里配置只是可选的默认值，不该拥有让命令失败的能力（纪律 2）。
func loadTdevSettings() config.TdevSettings {
	// allowMissing=true：配置文件不存在时返回缺省落点而不是报错，
	// 正好对上「没有配置文件的 tdev 用户照常可用」。
	path, err := config.ResolvePath("", true)
	if err != nil || path == "" {
		return config.TdevSettings{}
	}
	root, err := config.Load(path)
	if err != nil {
		return config.TdevSettings{}
	}
	return root.Tdev
}
