package fep

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// 本文件是 M5：**从源码抽"作业骨架"**，让角色脱离运行期命名。
//
// 依据（T100 侧，不是 Genero 侧）：
//   - `.4fd`（fglform 的表单定义，XML）里每个控件一条元素，带 widget/colName/comment/
//     required/fieldType 等属性 —— 这是"有哪些控件、各是什么类型、绑哪一栏"的权威来源。
//     实测：name="pmdl_t.pmdl001" fieldType="TABLE_COLUMN" gridHeight="1" comment="cmt_pmdl001"
//   - `.4gl` 里 `ON ACTION controlp INFIELD <字段>` + 紧邻的 `CALL q_xxx()` —— 这是
//     "这个字段按开窗会调哪个开窗程序"的权威来源。
//   - Genero 官方文档 1514 说交互元素**按 name 绑定 Action** —— 所以源码里的字段名
//     就是稳定角色，比运行期中文名可靠（中文名会重名，实测 付款条件 的字段与列就撞了）。

// SrcField 是源码里的一个控件（.4fd 一条元素）。
type SrcField struct {
	Form       string `json:"form"`
	Name       string `json:"name"`              // 协议名（如 formonly.pmdl004 / pmdl_t.pmdl004）
	ColName    string `json:"colName,omitempty"` // 数据库栏位（如 pmdl004）
	Widget     string `json:"widget,omitempty"`
	FieldType  string `json:"fieldType,omitempty"` // TABLE_COLUMN / NON_DATABASE / COLUMN_LIKE …
	Comment    string `json:"comment,omitempty"`   // cmt_pmdl004（运行期对象上也带这个）
	Title      string `json:"title,omitempty"`     // 中文标题（lbl_* 或直接文字）
	Required   bool   `json:"required,omitempty"`
	NoEntry    bool   `json:"noEntry,omitempty"`
	Opener     string `json:"opener,omitempty"` // 开窗程序（来自 .4gl 的 controlp 段）
	PosX, PosY int    `json:"posX,omitempty"`
	TabIndex   int    `json:"tabIndex,omitempty"`
}

var (
	reElement = regexp.MustCompile(`<(Edit|ButtonEdit|ComboBox|CheckBox|DateEdit|FFLabel|Label|Grid|Table)\s+([^>]*)>`)
	reAttr    = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\s*=\s*"([^"]*)"`)
	reAction  = regexp.MustCompile(`(?m)^\s*ON ACTION\s+([A-Za-z_][A-Za-z0-9_]*)`)
	reInf     = regexp.MustCompile(`INFIELD\s+([A-Za-z_][A-Za-z0-9_.]*)`)
	reCall    = regexp.MustCompile(`CALL\s+(q_[A-Za-z0-9_]+)\s*\(`)
)

// ParseForm4fd 解析一个 .4fd，按文档顺序返回控件清单。
func ParseForm4fd(form, text string) []SrcField {
	var out []SrcField
	for _, m := range reElement.FindAllStringSubmatch(text, -1) {
		tag, raw := m[1], m[2]
		attrs := map[string]string{}
		for _, a := range reAttr.FindAllStringSubmatch(raw, -1) {
			attrs[a[1]] = a[2]
		}
		f := SrcField{
			Form:      form,
			Name:      attrs["name"],
			ColName:   attrs["colName"],
			Widget:    attrs["widget"],
			FieldType: attrs["fieldType"],
			Comment:   attrs["comment"],
			Title:     attrs["title"],
			Required:  attrs["required"] == "true" || attrs["notNull"] == "true",
			NoEntry:   attrs["noEntry"] == "true",
		}
		if f.Widget == "" {
			f.Widget = tag
		}
		f.PosX = atoiSafe(attrs["posX"])
		f.PosY = atoiSafe(attrs["posY"])
		f.TabIndex = atoiSafe(attrs["tabIndex"])
		out = append(out, f)
	}
	return out
}

// ParseOpeners4gl 从 .4gl 抽"字段 → 开窗程序"：每个 `ON ACTION controlp INFIELD <字段>`
// 段里出现的第一个 `CALL q_xxx()`。
func ParseOpeners4gl(text string) map[string]string {
	openers := map[string]string{}
	idx := reAction.FindAllStringSubmatchIndex(text, -1)
	for i, m := range idx {
		action := text[m[2]:m[3]]
		if action != "controlp" && action != "controlq" && !strings.HasPrefix(action, "zoom") {
			continue
		}
		end := len(text)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		seg := text[m[0]:end]
		inf := reInf.FindStringSubmatch(seg)
		if inf == nil {
			continue
		}
		call := reCall.FindStringSubmatch(seg)
		if call == nil {
			continue
		}
		field := inf[1]
		if _, seen := openers[field]; !seen {
			openers[field] = call[1]
		}
	}
	return openers
}

// Actions4gl 返回程序里声明的动作名（ON ACTION <名字>），按出现顺序去重。
func Actions4gl(text string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range reAction.FindAllStringSubmatch(text, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

// Source 是一个作业的源码骨架（按环境镜像目录读）。
type Source struct {
	Root     string              `json:"root"`
	Program  string              `json:"program"`
	Forms    []string            `json:"forms"`
	Fields   []SrcField          `json:"fields"`
	Actions  []string            `json:"actions"`
	Openers  map[string]string   `json:"openers,omitempty"`
	ByCol    map[string]SrcField `json:"-"`
	ByComent map[string]SrcField `json:"-"`
	ByColNam map[string]SrcField `json:"-"`
}

// LoadSource 按作业程序名从源码镜像里找 .4fd/.4gl。
//
// 程序名形如 apmt500_wf：模块取前三码（apm），表单在 erp/<模块>/4fd/，
// 开窗等共用表单在 com/qry/4fd/。找不到就返回错误，调用方照实降级。
func LoadSource(root, program string) (*Source, error) {
	s := &Source{Root: root, Program: program,
		Openers: map[string]string{}, ByCol: map[string]SrcField{},
		ByComent: map[string]SrcField{}, ByColNam: map[string]SrcField{}}
	mod := ""
	if len(program) >= 3 {
		mod = strings.ToLower(program[:3])
	}
	// 主程序 + 同编号的兄弟程序（_01 子程序等）都读，动作/开窗可能在其中
	globs := []string{
		filepath.Join(root, "erp", mod, "4gl", program+".4gl"),
		filepath.Join(root, "erp", mod, "4gl", program+".4gl.src"),
	}
	if mod != "" {
		if extra, err := filepath.Glob(filepath.Join(root, "erp", mod, "4gl", program[:6]+"*.4gl")); err == nil {
			globs = append(globs, extra...)
		}
	}
	for _, g := range globs {
		b, err := os.ReadFile(g)
		if err != nil {
			continue
		}
		for f, o := range ParseOpeners4gl(string(b)) {
			if _, seen := s.Openers[f]; !seen {
				s.Openers[f] = o
			}
		}
		s.Actions = append(s.Actions, Actions4gl(string(b))...)
	}
	// 表单：主表单 + 主表单的兄弟（子画面）
	var forms []string
	for _, dir := range []string{
		filepath.Join(root, "erp", mod, "4fd"),
		filepath.Join(root, "com", "qry", "4fd"),
	} {
		matches, _ := filepath.Glob(filepath.Join(dir, program+"*.4fd"))
		forms = append(forms, matches...)
	}
	sort.Strings(forms)
	for _, f := range forms {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(filepath.Base(f), ".4fd")
		fields := ParseForm4fd(name, string(b))
		if len(fields) == 0 {
			continue
		}
		s.Forms = append(s.Forms, name)
		for i := range fields {
			if op, ok := s.Openers[fields[i].ColName]; ok {
				fields[i].Opener = op
			}
			fld := fields[i]
			if fld.ColName != "" {
				if _, dup := s.ByCol[fld.ColName]; !dup {
					s.ByCol[fld.ColName] = fld
				}
				if _, dup := s.ByComent["cmt_"+fld.ColName]; !dup {
					s.ByComent["cmt_"+fld.ColName] = fld
				}
			}
			if fld.Name != "" {
				if _, dup := s.ByColNam[fld.Name]; !dup {
					s.ByColNam[fld.Name] = fld
				}
			}
		}
		s.Fields = append(s.Fields, fields...)
	}
	if len(s.Fields) == 0 && len(s.Actions) == 0 {
		return nil, fmt.Errorf("在 %s 下没找到 %s 的 .4fd/.4gl（源码镜像可能没这个作业）", root, program)
	}
	return s, nil
}

// FindByComment 用运行期对象上的 comment（cmt_xxx）找源码字段。
func (s *Source) FindByComment(cmt string) (SrcField, bool) {
	if s == nil {
		return SrcField{}, false
	}
	f, ok := s.ByComent[cmt]
	return f, ok
}

// FindByColName 用数据库栏位名找源码字段。
func (s *Source) FindByColName(col string) (SrcField, bool) {
	if s == nil {
		return SrcField{}, false
	}
	f, ok := s.ByCol[col]
	return f, ok
}

// ApplySource 用源码骨架**重写角色**：角色优先用源码字段名（如 pmdl_t.pmdl004），
// 并把源码属性（widget/required/opener）挂到控件上。
//
// 为什么这一步是关键：运行期中文名会重名（实测 付款条件 的字段与列撞在一起），
// 而源码字段名是唯一的、跨会话稳定的 —— 这才是自动化脚本应该用的句柄。
func ApplySource(cat []Control, s *Source) int {
	if s == nil {
		return 0
	}
	n := 0
	for i := range cat {
		cmt := cat[i].Comment // 运行期对象带 cmt_xxx
		col := strings.TrimPrefix(cmt, "cmt_")
		f, ok := s.FindByComment(cmt)
		if !ok {
			f, ok = s.FindByColName(col)
		}
		if !ok {
			continue
		}
		if f.Name != "" {
			cat[i].Role = f.Name
		}
		cp := f
		cat[i].Source = &cp
		if _, hasOpen := OpsOf(cat[i].Kind); hasOpen && f.Opener != "" {
			cat[i].Ops = appendUnique(cat[i].Ops, "open")
		}
		n++
	}
	return n
}

func appendUnique(ops []string, want string) []string {
	for _, o := range ops {
		if o == want {
			return ops
		}
	}
	return append(ops, want)
}

func atoiSafe(v string) int {
	n := 0
	for i := 0; i < len(v); i++ {
		if v[i] < '0' || v[i] > '9' {
			return 0
		}
		n = n*10 + int(v[i]-'0')
	}
	return n
}
