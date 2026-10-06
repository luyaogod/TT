package drawio

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// 两份转义表照抄原 Node 实现（build.js / table.js 的 xmlEscape 与 xmlEscapeText），
// **字符集与顺序都不能改**。
//
// 为什么不用 encoding/xml 的 EscapeText：它会把 " 写成 &#34;、还会把制表符与换行
// 转成实体，输出与原来不一致。drawio 加载一个库走的是
//     mxUtils.parseXml(data) → JSON.parse(mxUtils.getTextContent(root))
// 也就是「先当 XML 解析、再取文本当 JSON」两层。哪一层少转了、或者多转了，
// 轻则形状显示错，重则整个库打不开（原项目为此踩过两次，报错是
// `Unexpected token 'T', "This page" ... is not valid JSON`）。
var (
	// attrEscaper 用在 XML 属性值里 —— 引号必须转，否则属性提前结束。
	attrEscaper = strings.NewReplacer(
		"&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;",
	)
	// textEscaper 用在 XML 文本节点里 —— 引号在文本里是合法的，转了反而改变内容。
	textEscaper = strings.NewReplacer(
		"&", "&amp;", "<", "&lt;", ">", "&gt;",
	)
)

// escapeAttr 转义 XML 属性值（5 个字符）。
func escapeAttr(v string) string { return attrEscaper.Replace(v) }

// escapeText 转义 XML 文本节点（3 个字符）。
func escapeText(v string) string { return textEscaper.Replace(v) }

// jsNum 按 JavaScript 模板字面量的方式打印一个数字（`${n}`）。
//
// 为什么需要它：原实现到处是 `${part.x}` 这种插值，Go 的默认格式化会给出不同的字面量
// （`%v` 对 100.0 打 `100`，对 4.8 打 `4.8`，但 `%g` 会打科学计数法）。坐标里有小数
// （ButtonEdit 的放大镜手柄是 `x=201.1 w=1.5`），所以必须逐值对齐，不能靠 %v 碰运气。
//
// 整数值一律不带小数点 —— 这与 JS 一致（`${100}` → `100`，不是 `100.0`）。
func jsNum(f float64) string {
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// mustNum 取一个部件坐标；缺值时由调用方保证已经过 validate。
func num(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// jsString 按 JavaScript 的 String(v) 打印一个 JSON 解出来的值 —— 表格行里的单元格
// 允许是数字或布尔（`[[1, 2], ...]`），原实现直接把它们插进模板字面量。
func jsString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		return jsNum(t)
	default:
		return fmt.Sprint(t)
	}
}
