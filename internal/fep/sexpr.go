package fep

import (
	"errors"
	"fmt"
	"strings"
)

// Node is one object definition or update as it appears in a message body.
//
// Type is the AUI object type ("Window", "Edit", "TableColumn", …) or the
// pseudo types "un"/"rn" for the two update forms. Parent is the declared
// parent id (0 for roots); updates carry Parent = 0 and are merged by the model
// instead of being inserted.
// Node 是 AUI 消息里的一个对象声明。
//
// Bare 表示它**没有 `an <父>` 前缀**。实测（本会话原始流）：
//
//	Value  110 {…}
//	Dialog 110 {…}
//
// 同一个 id 在顶层被声明成两个不同对象 —— 说明这类**无 an 的顶层声明**是
// “局部节点”（id 只在它自己的局部语境里有意义），不是全局对象身份。
// 而 `an 4097 ToolBarItem 5160 …` 这种带父的声明才是全局对象。
type Node struct {
	Parent int
	Type   string
	ID     int
	Bare   bool // 顶层无 an 前缀：不进全局模型
	Attrs  map[string]string
}

// maxParseDepth bounds recursion: AUI nesting is shallow, but the stream is
// untrusted.
const maxParseDepth = 64

// ParseMessage parses one message body and returns definitions and updates in
// stream order. It accepts the forms the client actually sends:
//
//	{an 0 Window 5195 {{attrs}} {{child}{child}}}   explicit parent
//	{TableColumn 5196 {{attrs}} {{child}}}          parent from the enclosing group
//	{{Value 6599 {{value ""}} {}}}                   children are each wrapped in braces
//	{un 401 {{value "1"}}}                           attribute-only update
//	{TableColumn 5196 {{attrs}} {{child}}}          parent from the enclosing group
//	om 3 {{an 0 Window 5195 {{attrs}}}}              verb prefix on the wire
//
// Children are written as a group of brace-wrapped objects — missing that extra
// layer is what silently flattens the tree and mis-parents every child after the
// first one.
// TrimNoise normalizes one recorded message line.
//
// Recordings written by the earlier prototype kept part of the frame header in
// front of the message (`…\x00\x00+om 0 {{…}}`), and that leading binary is enough
// to hide the verb: the parse then invents an object of type "om" and the whole
// screen disappears. The second return value reports whether a known verb was
// found at all, so a caller can drop lines that are not messages (the proxy's own
// `!!__cached__!!` markers, for example).
func TrimNoise(line string) (string, bool) {
	i := 0
	for i < len(line) && !isLetter(line[i]) {
		i++
	}
	s := line[i:]
	for _, v := range msgVerbs {
		if strings.HasPrefix(s, v) {
			rest := s[len(v):]
			if rest == "" || isSpace(rest[0]) || rest[0] == '{' {
				return strings.TrimSpace(s), true
			}
		}
	}
	return "", false
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// ParseMessage parses one message body. The body may carry a verb prefix —
// `om 3 {{…}}`, `meta Connection {{…}}` — so parsing skips it.
//
// Only known verbs are skipped. "Cut at the first `{`" looks equivalent and is
// not: a body may open with a bare object header (`Table 501 {…}`), and cutting
// there throws the type and id away — a whole screen collapses into objects with
// no name and no parent.
func ParseMessage(body string) []Node {
	if i := verbEnd(body); i > 0 {
		body = body[i:]
	}
	var out []Node
	parseSeq(body, 0, 0, &out)
	return out
}

// msgVerbs are the lowercase verbs the runtime puts in front of a message.
var msgVerbs = []string{"om", "meta", "un", "rn", "event"}

// ErrUnsupportedDialect 报"这条入站消息不是我们实现的那套方言"（契约 C6.2）。
//
// 上游证据（three-threads.md §①）：**同一条路径上有两套方言**，由运行期全局开关
// （`0x30c9b4`）切换：
//
//	{an …} {un …} {rn …} {om …}     原生 GDC —— 就是我们处理、也是本包唯一实现的这套
//	["a" …] ["u" …] ["r" …]         数组/JSON 方言（容器 `[`、引号、`:` 分隔），
//	                                UR/HTML5 客户端或启用 `gui.rendering` 时启用
//
// 第二套**不实现**（契约 C8：不做第二方言的收发），但**必须明确报错**：硬解它不会崩，
// 只会把对象名与父链解析成垃圾 —— 于是"方言不对"和"这窗口本来就这样"在调用方眼里
// 长得一模一样。这是计划里点名的两条**静默出错**路径之一，所以它是错误，不是降级。
var ErrUnsupportedDialect = errors.New("fep: 未支持的入站消息方言（数组/JSON 方言：结构从 [ 开始，UR/HTML5 客户端或 gui.rendering 启用）")

// ErrUnknownShape 报"这条入站文本既不是 S 表达式、也不是数组方言"（A16 增，2026-10-17）。
//
// 它**不是** ErrUnsupportedDialect，两者刻意不合并：方言错误说"对端换了前端"
// （该换解析器 / 该提示用户），形状错误说"这段字节根本不是一条消息"（帧头噪声、
// 半截载荷、别的什么东西 —— 该重取字节）。合成一个码，调用方就没得选。
var ErrUnknownShape = errors.New("fep: 未知的入站消息形状（既不是 { 结构的 S 表达式，也不是 [ 结构的数组/JSON 方言）")

// dialectEvidenceLen 是错误里附带证据片段的上限：诊断要能看懂，但不该把整条载荷
// （窗口声明可以几万字节）抄进日志。
const dialectEvidenceLen = 32

// CheckDialect 探测一条入站消息属于哪套方言/什么形状；是我们实现的 S 表达式时返回 nil。
//
// 判据（A16 放宽后）：去空白后**看第一个结构字符** ——
//
//	空 / 纯空白        → nil（交给 ParseMessage 的既有行为：0 个节点，不报错）
//	第一个结构字符 `{` → nil（S 表达式，放行）
//	第一个结构字符 `[` → ErrUnsupportedDialect（数组/JSON 方言，附 32 字节证据）
//	两者都没有         → ErrUnknownShape
//
// 为什么不只看首字节：S 表达式有两种**合法**写法并不以 `{` 开头 —— 带动词前缀
// （`om 3 {{…}}`、`meta Connection {{…}}`）与从裸对象头开始（`Table 501 {{…}}`，
// 见 ParseMessage 的文档与 TestBareObjectHeaderIsNotAVerb）。证据只支持一句话：
// **S 表达式的结构必然从 `{` 开始**（语料 7,805 帧无一例外）。所以"谁先出现就定谁"：
// 既堵住 A16 要堵的 `om 3 ["a",…]`（原判据只看首字节，会把它漏给 S 表达式解析器
// 静默解成垃圾），又不会误拒上面两种合法写法。
//
// 属性值里的方括号（`{text "a[b"}`）不影响判据：那时 `{` 已经先出现过了。
func CheckDialect(body string) error {
	i := skipWS(body, 0)
	if i >= len(body) {
		return nil
	}
	for j := i; j < len(body); j++ {
		switch body[j] {
		case '{':
			return nil
		case '[':
			return fmt.Errorf("%w：检测到数组/JSON 方言（`[\"a\",…]`/`[\"u\",…]`/`[\"r\",…]`），证据 %q",
				ErrUnsupportedDialect, clipEvidence(body[j:], dialectEvidenceLen))
		}
	}
	return fmt.Errorf("%w：证据 %q", ErrUnknownShape, clipEvidence(body[i:], dialectEvidenceLen))
}

// clipEvidence 截一段载荷当证据（超长加省略号；按字节切，诊断用不着顾字符边界）。
func clipEvidence(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// ParseMessageChecked 是**带方言探测的**解析入口：CheckDialect + ParseMessage。
//
// ParseMessage 本身一个字节都没改（老调用方不受影响，见 aui.go 的 ApplyMessage）；
// 新调用方走这里：命中未支持方言就返回 nil 节点 + 包着 `ErrUnsupportedDialect` 的错误，
// 形状不认就包着 `ErrUnknownShape`（两者都 `errors.Is` 可判），**绝不"尽力而为"地硬解**。
func ParseMessageChecked(body string) ([]Node, error) {
	if err := CheckDialect(body); err != nil {
		return nil, err
	}
	return ParseMessage(body), nil
}

// verbEnd returns the offset just past a leading verb (and its argument), or 0
// when the body starts with data.
func verbEnd(body string) int {
	for _, v := range msgVerbs {
		if !strings.HasPrefix(body, v) {
			continue
		}
		rest := body[len(v):]
		if rest != "" && !isSpace(rest[0]) {
			continue // `omx` is not the verb `om`
		}
		i := len(v)
		for i < len(body) && isSpace(body[i]) {
			i++
		}
		// skip the verb's argument: a number for om/un/rn, a type for meta/event
		for i < len(body) && !isSpace(body[i]) && body[i] != '{' {
			i++
		}
		for i < len(body) && isSpace(body[i]) {
			i++
		}
		if i < len(body) && body[i] == '{' {
			return i
		}
		return 0
	}
	return 0
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// skipWS returns i advanced past whitespace.
func skipWS(s string, i int) int {
	for i < len(s) {
		switch s[i] {
		case ' ', '\t', '\n', '\r':
			i++
		default:
			return i
		}
	}
	return i
}

// identAt reads an AUI identifier starting at i.
func identAt(s string, i int) (string, int) {
	start := i
	for i < len(s) {
		c := s[i]
		if c == '_' || c == '.' || c == '-' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			i++
			continue
		}
		break
	}
	return s[start:i], i
}

// numAt reads a non-negative integer starting at i.
func numAt(s string, i int) (int, int, bool) {
	start := i
	n := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		n = n*10 + int(s[i]-'0')
		i++
	}
	return n, i, i > start
}

// groupAt returns the text between the braces at i plus the index just past the
// matching '}'. It assumes s[i] == '{'.
func groupAt(s string, i int) (string, int) {
	depth := 0
	start := i + 1
	for ; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start:i], i + 1
			}
		}
	}
	return s[start:], len(s)
}

// attrs parses `{key "value"}` pairs; unknown shapes are ignored.
func attrs(s string) map[string]string {
	m := make(map[string]string)
	i := 0
	for i < len(s) {
		if s[i] != '{' {
			i++
			continue
		}
		inner, next := groupAt(s, i)
		i = next
		sp := strings.IndexByte(inner, ' ')
		if sp <= 0 {
			continue
		}
		key := strings.TrimSpace(inner[:sp])
		val := strings.TrimSpace(inner[sp+1:])
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
		if key != "" {
			m[key] = val
		}
	}
	return m
}

// parseSeq walks a sequence of items: brace-wrapped groups (recursed into, the
// wrapper does not change the parent) and bare objects.
func parseSeq(s string, parent, depth int, out *[]Node) {
	if depth > maxParseDepth {
		return
	}
	i := 0
	for {
		i = skipWS(s, i)
		if i >= len(s) {
			return
		}
		switch s[i] {
		case '}':
			i++ // stray closer (partial message): resynchronise
			continue
		case '{':
			inner, next := groupAt(s, i)
			parseSeq(inner, parent, depth+1, out)
			i = next
			continue
		}
		next, ok := parseObj(s, i, parent, depth, out)
		if !ok {
			i++
			continue
		}
		i = next
	}
}

// parseObj parses one object: `[an <parent>] <Type> <id> {{attrs}} [{{children}}…]`.
// It returns the index just past the object and whether a header was found.
func parseObj(s string, i, parent, depth int, out *[]Node) (int, bool) {
	pid := parent
	hadAn := false
	if strings.HasPrefix(s[i:], "an ") {
		j := skipWS(s, i+3)
		p, j2, ok := numAt(s, j)
		if !ok {
			return i, false
		}
		pid = p
		hadAn = true
		i = skipWS(s, j2)
	}
	typ, j := identAt(s, i)
	if typ == "" || typ == "un" || typ == "rn" {
		// `un <id> {{attrs}}` / `rn <id> {{attrs}}`: attribute-only update.
		if typ == "" {
			return i, false
		}
		j = skipWS(s, j)
		id, k, ok := numAt(s, j)
		if !ok {
			return i, false
		}
		n := Node{Parent: 0, Type: typ, ID: id}
		k = skipWS(s, k)
		if k < len(s) && s[k] == '{' {
			body, next := groupAt(s, k)
			n.Attrs = attrs(body)
			k = skipWS(s, next)
		}
		*out = append(*out, n)
		return k, true
	}
	j = skipWS(s, j)
	id, k, ok := numAt(s, j)
	if !ok {
		return i, false
	}
	n := Node{Parent: pid, Type: typ, ID: id, Bare: !hadAn}
	k = skipWS(s, k)
	if k < len(s) && s[k] == '{' { // attribute group (may be empty)
		body, next := groupAt(s, k)
		n.Attrs = attrs(body)
		k = skipWS(s, next)
	}
	*out = append(*out, n)
	// Any following group belongs to this object's children.
	for k < len(s) && s[k] == '{' {
		body, next := groupAt(s, k)
		parseSeq(body, id, depth+1, out)
		k = skipWS(s, next)
	}
	return k, true
}
