package drawio

import (
	"bytes"
	"encoding/json"
	"strings"
)

// MarshalJSON 按 JavaScript 的 `JSON.stringify(v, null, 2)` 输出。
//
// 为什么不能用 json.MarshalIndent：Go 默认把尖括号与和号转成十六进制实体（防 HTML
// 注入），而 JS 原样输出。形状源里到处是 XML 片段，这一条会让产物面目全非。
// SetEscapeHTML(false) 关掉它，SetIndent 补上两空格缩进，最后去掉 Encoder 多打的换行
// （stringify 不产生尾随换行）。
//
// 已知的一处微小差异：Go 无条件转义行分隔符与段分隔符（U+2028 / U+2029），
// JS 原样输出。这只会影响**字节**，不影响解码后的值 —— 判据一律比解码后的语义，
// 不比字节。
func MarshalJSON(v any) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}
