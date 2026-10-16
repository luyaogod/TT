package fep

import "strings"

// ActionKind classifies a user-intent event.
type ActionKind string

const (
	// ActionClick is an ActionEvent: a button, an open-window button, a grid
	// row, a combo item or a menu action.
	ActionClick ActionKind = "click"
	// ActionInput is a ConfigureEvent carrying a value: typing or ticking a box.
	ActionInput ActionKind = "input"
)

// Action is one semantic action extracted from an event message.
type Action struct {
	Kind  ActionKind
	ID    int    // idRef of the clicked object, or the id of the edited field
	Value string // input value (ActionInput only)
	Raw   string // the original event text, kept for the audit trail
}

// SplitEvents splits one `event _om N{}{{E1}{E2}…}` message body into the
// individual event texts (for example `ActionEvent 0{{idRef "5160"}}`).
func SplitEvents(body string) []string {
	// Drop the `event _om <n>{}` prefix; what remains is `{E1}{E2}…}`.
	s := body
	if i := strings.Index(s, "event _om "); i >= 0 {
		s = s[i+len("event _om "):]
		// skip the sequence number
		for len(s) > 0 && s[0] >= '0' && s[0] <= '9' {
			s = s[1:]
		}
		// skip the empty `{}` group
		s = strings.TrimPrefix(s, "{}")
	}
	var out []string
	i := 0
	for i < len(s) {
		if s[i] != '{' {
			i++
			continue
		}
		inner, next := groupAt(s, i)
		i = next
		inner = strings.TrimSpace(inner)
		if inner == "" {
			continue
		}
		if inner[0] == '{' { // still the outer wrapper
			out = append(out, SplitEvents(inner)...)
			continue
		}
		out = append(out, inner)
	}
	return out
}

// SemanticActions keeps only user intent and drops navigation. Validated
// ratio on a real create-order flow: 68 raw events -> 19 semantic actions, the
// 49 dropped ones being KeyEvent(Tab), cursor/cursor2, currentRow/currentColumn
// and tabIndex/bufferSize/pageSize reports.
func SemanticActions(events []string) []Action {
	var out []Action
	for _, e := range events {
		if a, ok := semanticAction(e); ok {
			out = append(out, a)
		}
	}
	return out
}

// semanticAction classifies a single event.
func semanticAction(e string) (Action, bool) {
	switch {
	case strings.HasPrefix(e, "ActionEvent"):
		if id, ok := idRefOf(e); ok {
			return Action{Kind: ActionClick, ID: id, Raw: e}, true
		}
		return Action{}, false
	case strings.HasPrefix(e, "ConfigureEvent"):
		if v, ok := attrOf(e, "value"); ok {
			a := Action{Kind: ActionInput, Value: v, Raw: e}
			if id, ok := idRefOf(e); ok {
				a.ID = id
			}
			return a, true
		}
		// cursor/currentRow/currentColumn/offset/bufferSize/pageSize: navigation
		// or status reports, never a step.
		return Action{}, false
	default:
		// KeyEvent (Tab …) and everything else is navigation.
		return Action{}, false
	}
}

// idRefOf extracts `idRef "<n>"`.
func idRefOf(e string) (int, bool) {
	v, ok := attrOf(e, "idRef")
	if !ok {
		return 0, false
	}
	n, _, ok := numAt(v, 0)
	if !ok {
		return 0, false
	}
	return n, true
}

// soleIDRef 返回载荷里**唯一**的那个 `idRef "<n>"`；出现两个不同的值就回 false。
//
// 为什么不能"取第一个"：一条 `event _om` 帧可以带一批事件（实测 `event _om 0` 里是
// 十来个 `{ConfigureEvent 0{{idRef "960"}{tabIndex "1005"}}}` 这类导航上报，id 各不相同）。
// 这时"这条消息指向哪个对象"没有答案，取第一个就是**编** —— 而它会被界面当"目标"显示。
func soleIDRef(e string) (int, bool) {
	found, n := 0, 0
	for i := 0; ; {
		j := strings.Index(e[i:], `idRef "`)
		if j < 0 {
			break
		}
		rest := e[i+j+len(`idRef "`):]
		v, _, ok := numAt(rest, 0)
		if !ok {
			i += j + len(`idRef "`)
			continue
		}
		if found == 0 {
			n = v
		} else if v != n {
			return 0, false
		}
		found++
		i += j + len(`idRef "`)
	}
	if found == 0 {
		return 0, false
	}
	return n, true
}

// attrOf extracts `key "value"` from an event text.
func attrOf(e, key string) (string, bool) {
	needle := key + " \""
	i := strings.Index(e, needle)
	if i < 0 {
		return "", false
	}
	rest := e[i+len(needle):]
	j := strings.IndexByte(rest, '"')
	if j < 0 {
		return "", false
	}
	return rest[:j], true
}
