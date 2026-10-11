package fep

import "strings"

// NameConfidence grades a resolved field name.
//
// High means the name can be published in a manual: the source is explicit
// (label text, dictionary key or the column's own text) and the object's parent
// chain is cycle-free and reaches a job window. Low means the chain is broken
// or cyclic (partial messages merged across windows), so the name is kept for
// diagnostics but must not be quoted. None means nothing was found: never guess.
type NameConfidence string

const (
	ConfHigh NameConfidence = "high"
	ConfLow  NameConfidence = "low"
	ConfNone NameConfidence = "unresolved"
)

// FieldName is a resolved field/column label with its provenance.
type FieldName struct {
	Text       string
	Source     string // "comment-key" | "name-suffix" | "own-text" | "inherit#<id>"
	Confidence NameConfidence
}

// Resolver turns AUI objects into human labels using the rules validated on a
// live job (plan §2.10):
//
//	R1 comment-key    field.comment == a Label's comment that carries text
//	R2 name-suffix    field.col  -> object whose "name" ends with .<col> / .b_<col> and carries text
//	R3 own-text       the object carries text itself
//	R4 inherit        parent chain, field-like parents only
type Resolver struct {
	m      *Model
	dict   map[string]string // cmt_x -> label text
	byName map[string]string // name suffix -> label text
	stats  stats
}

type stats struct {
	resolved int
	unknown  int
}

// NewResolver indexes the model's labels.
func NewResolver(m *Model) *Resolver {
	r := &Resolver{m: m, dict: make(map[string]string), byName: make(map[string]string)}
	for _, o := range m.Objects() {
		cmt := o.Attr("comment")
		if o.Type == "Label" && cmt != "" && o.Attr("text") != "" {
			if _, seen := r.dict[cmt]; !seen {
				r.dict[cmt] = o.Attr("text")
			}
		}
		name := o.Attr("name")
		text := o.Attr("text")
		if text == "" || name == "" {
			continue
		}
		seg := name
		if i := strings.LastIndexByte(name, '.'); i >= 0 {
			seg = name[i+1:]
		}
		if _, seen := r.byName[seg]; !seen {
			r.byName[seg] = text
		}
		if b := strings.TrimPrefix(seg, "b_"); b != seg {
			if _, seen := r.byName[b]; !seen {
				r.byName[b] = text
			}
		}
	}
	return r
}

// fieldLikeParents are the only parents R4 may inherit through. Containers
// (VBox/HBox/Page/Folder/Form/Window/Grid…) stop the walk: crossing one is what
// produced confidently-wrong names in the prototype.
var fieldLikeParents = map[string]bool{
	"FormField": true, "TableColumn": true, "Edit": true, "ButtonEdit": true,
	"ComboBox": true, "CheckBox": true, "DateEdit": true, "TextEdit": true,
}

// Resolve returns the label for an object.
func (r *Resolver) Resolve(id int) (FieldName, bool) {
	o, ok := r.m.Get(id)
	if !ok {
		r.stats.unknown++
		return FieldName{Confidence: ConfNone}, false
	}
	if fn, ok := r.direct(o); ok {
		return r.grade(o, fn), true
	}
	// R4: inherit, field-like parents only, and only on a sane chain.
	if _, cycle, _ := r.m.Chain(id); !cycle {
		cur := o
		for hops := 0; hops < 8; hops++ {
			if cur.Parent == 0 {
				break
			}
			p, ok := r.m.Get(cur.Parent)
			if !ok || !fieldLikeParents[p.Type] {
				break
			}
			if fn, ok := r.direct(p); ok {
				fn.Source = "inherit#" + itoa(p.ID)
				return r.grade(o, fn), true
			}
			cur = p
		}
	}
	r.stats.unknown++
	return FieldName{Confidence: ConfNone}, false
}

// direct applies R1–R3 to one object.
func (r *Resolver) direct(o *Object) (FieldName, bool) {
	cmt := o.Attr("comment")
	col := o.Attr("col")
	if col == "" {
		col = o.Attr("colName")
	}
	key := strings.TrimPrefix(col, "b_")
	if cmt != "" {
		if t, ok := r.dict[cmt]; ok {
			return FieldName{Text: t, Source: "comment-key", Confidence: ConfHigh}, true
		}
	}
	if key != "" {
		if t, ok := r.byName[key]; ok {
			return FieldName{Text: t, Source: "name-suffix", Confidence: ConfHigh}, true
		}
	}
	if cmt != "" && strings.HasPrefix(cmt, "cmt_") {
		if t, ok := r.byName[strings.TrimPrefix(cmt, "cmt_")]; ok {
			return FieldName{Text: t, Source: "name-suffix", Confidence: ConfHigh}, true
		}
	}
	if t := o.Attr("text"); t != "" {
		return FieldName{Text: t, Source: "own-text", Confidence: ConfHigh}, true
	}
	return FieldName{}, false
}

// grade downgrades names whose chain is cyclic (window-scoping violation).
func (r *Resolver) grade(o *Object, fn FieldName) FieldName {
	_, cycle, win := r.m.Chain(o.ID)
	switch {
	case cycle:
		fn.Confidence = ConfLow
	case win == 0:
		fn.Confidence = ConfLow
	default:
		fn.Confidence = ConfHigh
	}
	r.stats.resolved++
	return fn
}

// Stats reports resolution counts for the whole model.
func (r *Resolver) Stats() (resolved, unresolved int) {
	return r.stats.resolved, r.stats.unknown
}

// dataFieldTypes are the object kinds a manual can name.
var dataFieldTypes = map[string]bool{
	"FormField": true, "TableColumn": true, "Edit": true, "ButtonEdit": true,
	"ComboBox": true, "CheckBox": true, "DateEdit": true, "TextEdit": true,
}

// Coverage walks every data field and reports how many resolve. It is the
// number quoted in the plan (94.4% on a live axmt500_wf session).
func (r *Resolver) Coverage() (resolved, total int) {
	for _, o := range r.m.Objects() {
		if !dataFieldTypes[o.Type] {
			continue
		}
		if o.Attr("col") == "" && o.Attr("comment") == "" {
			continue
		}
		total++
		if _, ok := r.Resolve(o.ID); ok {
			resolved++
		}
	}
	return resolved, total
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
