package fep

import "strings"

// Node is one object definition or update as it appears in a message body.
//
// Type is the AUI object type ("Window", "Edit", "TableColumn", …) or the
// pseudo types "un"/"rn" for the two update forms. Parent is the declared
// parent id (0 for roots); updates carry Parent = 0 and are merged by the model
// instead of being inserted.
type Node struct {
	Parent int
	Type   string
	ID     int
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
	if strings.HasPrefix(s[i:], "an ") {
		j := skipWS(s, i+3)
		p, j2, ok := numAt(s, j)
		if !ok {
			return i, false
		}
		pid = p
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
	n := Node{Parent: pid, Type: typ, ID: id}
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
