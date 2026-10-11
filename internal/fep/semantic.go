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
