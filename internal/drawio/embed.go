// Package drawio 实现 T100 的 drawio 图纸能力：形状库的合成，以及
// 「排版规格 spec → .drawio」与「.tzs 表单包 → 原型图」两条出图路径。
//
// 本包是从 T100Drawio（原独立 Node 项目）整套搬进来的 —— 形状源、合成规则、
// 排版规则、动态表格规则都在这一层，运行期不依赖 Node。
//
// 三条边界：
//   - 只依赖标准库。形状源与合成规则是纯数据变换，不需要任何第三方包。
//   - 形状源只此一份（shapes/ 整棵 embed）。catalog、mxlibrary、形状清单全是
//     从它派生的，**一律不落库、不入库** —— 落盘的第二个家一定会漂。
//   - 不碰网络、不碰配置；落盘只有调用方显式点名的那一个产物。
package drawio

import (
	"embed"
	"io/fs"
)

// shapes 是形状源（原 T100Drawio 的 src/ 整棵搬进来）。
//
// 为什么 embed 而不是像 skills/ 那样随包发目录：形状源是**代码级数据**，必须与
// 二进制同版本；它不是给用户编辑的文档（skills/ 不 embed 的理由在它身上不成立）。
// embed 还让"源文件少一个"变成编译期失败，而不是运行期读盘找不到。
//
//go:embed all:shapes
var shapes embed.FS

// ShapesFS 返回形状源的根（已去掉 `shapes/` 前缀）。
//
// 目录布局：`library.json`、`presets.json`、`controls/*.json`、`business/*.json`、
// `icons/*.svg`、`raw/*.xml` —— 子目录名就是库名，与 library.json 的键一一对应。
func ShapesFS() fs.FS {
	sub, err := fs.Sub(shapes, "shapes")
	if err != nil {
		// 只可能在 embed 指令被改动时发生；那时的正确结果是当场炸，不是静默降级。
		panic("drawio: 形状源目录结构异常: " + err.Error())
	}
	return sub
}
