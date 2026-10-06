package drawio

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// 图标一律内联成 data URI —— 库文件因此不依赖任何外部图片，单文件即可分发。
//
// 注意 README 里那条相反的结论：**不要用 `image=` 内联 SVG 画复杂图标**（drawio 里
// 位置会飘），图标一律用矢量图元拼。这里的 data URI 只服务于 part 的 `icon` 字段，
// 目前只有日历/放大镜这类小部件在用。

// iconCache 按名字缓存已内联的 svg（同一个图标被多个部件引用时只读一次）。
type iconCache struct {
	fsys fs.FS
	m    map[string]string
}

func newIconCache(fsys fs.FS) *iconCache {
	return &iconCache{fsys: fsys, m: map[string]string{}}
}

// uri 返回 `data:image/svg+xml,...`。
func (c *iconCache) uri(name string) (string, error) {
	if v, ok := c.m[name]; ok {
		return v, nil
	}
	b, err := fs.ReadFile(c.fsys, path.Join("icons", name+".svg"))
	if err != nil {
		return "", fmt.Errorf("找不到图标 icons/%s.svg", name)
	}
	uri := "data:image/svg+xml," + encodeURIComponent(strings.TrimSpace(string(b)))
	c.m[name] = uri
	return uri, nil
}

// encodeURIComponent 是 JavaScript 同名函数的等价实现。
//
// 为什么不能直接用 net/url：url.QueryEscape 把空格转成 `+`（不是 %20），
// url.PathEscape 又漏掉 `!*'()` 这几个 —— 两者都与 JS 不同。而这里的用途是把 SVG
// 塞进 drawio 的样式串，`;` 与 `#` 必须被编码（否则样式串会被 drawio 的解析器截断），
// 一个字符不对，图标就挂了。
//
// 不转义的字符集照抄 JS 规范：A-Z a-z 0-9 - _ . ! ~ * ' ( )
func encodeURIComponent(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isURIUnreserved(c) {
			b.WriteByte(c)
			continue
		}
		const hex = "0123456789ABCDEF"
		b.WriteByte('%')
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&0x0F])
	}
	return b.String()
}

func isURIUnreserved(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		return true
	}
	switch c {
	case '-', '_', '.', '!', '~', '*', '\'', '(', ')':
		return true
	}
	return false
}
