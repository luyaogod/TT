package tzs

// argmap.go —— 命令行 `--k v` → args JSON。
//
// 定型照 manifest 走（`--offset 2` 是数字 2，`--paths a,b` 是数组），
// 本地只校验**引擎自己也会校验、而且现在就能判**的那几件事：
//
//	必填 / 未知参数 / 类型名能不能解析（int、bool）/ enum 的静态合法集
//
// **本地绝不拦的东西**（越界拦截是无声发生的，所以这条必须写下来）：
//
//   - `kind` 参数：合法集（field/hfield/pfield/rfield/mlfield/tree/act）**不在 manifest 里**，
//     Go 侧抄一份就是第二份真源，引擎一改就静默过期。
//   - `attr` 参数（from: "spec:<kind>" / "layout"）：合法集要从**活着的模型**里取，
//     静态根本判不了。典型反例：`set_spec_attr --attr bogus_attr --kind field` 必须
//     照发，让引擎用自己的白名单拒绝（它会回 E_ATTR_NOT_WHITELIST，kind=validation，退 2）。
//   - 任何 `t: "string"` 的参数：哪怕它的语义值域很窄。反例：`add_action --type db3`
//     —— 这个"型态"的合法集是**每张表单自己的**（来自该表单的 s_detail<n> 记录，
//     aapp320 有 12 个、别的表单 3 个），manifest 里它就是一个普通 string，没有 from。
//     在本地拦它会让「加一个 db* 型态」永远到不了引擎。
//
// 换句话说：本地只做**语法**层，语义层（值域）交给引擎。引擎拒绝的退出码是
// kind=validation → 2，和本地拦下来**一模一样**，所以这不会让报错变晚或变差。
//
// 一条与 tzs-cli 的有意差异：位置参数（不带 `--` 的裸词）在这里是**错误**，
// 而 tzs-cli 会「忽略非选项参数」并继续。无声忽略正是引擎 Manifest.Check 双向卡 arity
// 要治的病（`--attrr` 打错了和合法调用在结果上无法区分，直到有人去读文件），
// 所以这里宁可当场报错。

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Arg 是一个命令行选项：`--name value` / `--name=value` / 裸 `--name`（值 "true"）。
type Arg struct {
	Name  string
	Value string
}

// SplitArgs 把 `--k v` 拆成 Arg 列表 —— 只做语法，不碰 manifest。
//
// 规则（与引擎的 tzs-cli Flags() 同口径）：
//
//	--name value   下一个词不以 `--` 开头就吃掉它。`-1` 不以 `--` 开头，
//	               所以 `--value -1` 的值是字符串 "-1"（负号不是 flag ——
//	               把它当 flag 会把一个合法的负值变成「未知参数」）。
//	--name=v       `=` 后全是值，允许为空串（`--desc=` 是「清空」）。
//	--name         后面没有值 → "true"。这是为了让布尔参数读起来自然：
//	               `--force` / `--excluded` / `--cited`。
//
// 注意一个必然的歧义：`--path --force` 会让 path 取到 "true"（因为 --force 以 `--`
// 开头，谁也不吃谁）。不做特殊处理：真要用一个以 `--` 开头的值，写 `--path=--force`。
// 这种打错会被引擎当场拒（path "true" 找不到 → not_found → 退 2），不会静默发出去。
func SplitArgs(argv []string) ([]Arg, error) {
	var out []Arg
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		if !strings.HasPrefix(a, "--") {
			return nil, &UsageError{
				Msg:    "位置参数 " + strconv.Quote(a) + " 不是选项",
				Detail: []string{"本命令的参数一律写成 --<名> <值>（可用参数见该函数的 manifest）"},
			}
		}
		body := a[2:]
		if body == "" {
			return nil, &UsageError{Msg: "`--` 不是参数"}
		}
		name, val, hasVal := body, "", false
		if eq := strings.IndexByte(body, '='); eq >= 0 {
			name, val, hasVal = body[:eq], body[eq+1:], true
		}
		if !hasVal {
			if i+1 < len(argv) && !strings.HasPrefix(argv[i+1], "--") {
				val = argv[i+1]
				i++
			} else {
				val = "true"
			}
		}
		out = append(out, Arg{Name: name, Value: val})
	}
	return out, nil
}

// BuildArgs 按 manifest 把 Arg 列表定型成 args 对象（json.RawMessage）。
//
// 输出顺序 = manifest 里参数的声明顺序（不是命令行顺序）：同一份输入永远产出同一串字节，
// 于是 `--json` 里的帧、日志里的记录、测试里的断言都可以逐字节比对。
func BuildArgs(m *Manifest, fn string, args []Arg) (json.RawMessage, error) {
	spec := m.ByName(fn)
	if spec == nil {
		if fn == FnStop {
			// 这不是遗漏：stop 是传输级命令，引擎在查表之前就把它截走，所以它不在 manifest 里。
			return nil, &UsageError{Msg: FnStop + " 是传输级命令，不走参数定型"}
		}
		return nil, &UsageError{
			Msg:    "未知函数 " + fn,
			Detail: []string{"可用函数：" + clipNames(m.Names())},
		}
	}

	// 按名归并。列表参数的可重复是特性（`--paths a --paths b` 是两项），标量的重复是笔误。
	vals := map[string][]string{}
	for _, a := range args {
		if spec.Param(a.Name) == nil {
			return nil, &UsageError{
				Msg: "参数 " + a.Name + " 不在 manifest 里（函数 " + fn + "）",
				Detail: []string{"该函数的参数：" + clipNames(paramNames(spec)),
					"参数是**整个** --name value，不是 --name=value 之外还能缩写"},
			}
		}
		vals[a.Name] = append(vals[a.Name], a.Value)
	}

	var buf bytes.Buffer
	buf.WriteByte('{')
	first := true
	emit := func(name, raw string) {
		if !first {
			buf.WriteByte(',')
		}
		first = false
		buf.WriteString(jstr(name))
		buf.WriteByte(':')
		buf.WriteString(raw)
	}

	for _, p := range spec.Args {
		vs, ok := vals[p.Name]
		if !ok {
			if p.Required {
				return nil, &UsageError{
					Msg:    "缺必填参数 " + p.Name + "（函数 " + fn + "）",
					Detail: []string{"必填：" + clipNames(spec.RequiredParams())},
				}
			}
			continue
		}
		raw, err := paramJSON(fn, p, vs)
		if err != nil {
			return nil, err
		}
		emit(p.Name, raw)
	}
	buf.WriteByte('}')
	return json.RawMessage(buf.Bytes()), nil
}

// paramJSON 把一个参数的（可能多个）命令行取值定型成一段 JSON。
func paramJSON(fn string, p *Param, vs []string) (string, error) {
	switch p.Type {
	case TypePathList, TypeStrList:
		// 列表：合并**所有**出现（`--paths a,b --paths c` → a,b,c），每次出现再按逗号切。
		// 空串等于空数组（引擎收 [] 表示「什么也不做」，比 [] 更明确的只有不传）。
		var items []string
		for _, v := range vs {
			if v == "" {
				continue
			}
			for _, part := range strings.Split(v, ",") {
				items = append(items, strings.TrimSpace(part))
			}
		}
		if items == nil {
			items = []string{}
		}
		var b bytes.Buffer
		b.WriteByte('[')
		for i, it := range items {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(jstr(it))
		}
		b.WriteByte(']')
		return b.String(), nil

	case TypeInt:
		if len(vs) > 1 {
			return "", repeated(fn, p)
		}
		n, err := strconv.ParseInt(strings.TrimSpace(vs[0]), 10, 64)
		if err != nil {
			return "", &UsageError{
				Msg:    "参数 " + p.Name + " 需要整数，收到 " + strconv.Quote(vs[0]),
				Detail: []string{"函数 " + fn},
			}
		}
		return strconv.FormatInt(n, 10), nil

	case TypeBool:
		if len(vs) > 1 {
			return "", repeated(fn, p)
		}
		b, ok := parseBool(vs[0])
		if !ok {
			return "", &UsageError{
				Msg:    "参数 " + p.Name + " 需要 true/false，收到 " + strconv.Quote(vs[0]),
				Detail: []string{"函数 " + fn, "裸 --" + p.Name + " 等价于 --" + p.Name + " true"},
			}
		}
		if b {
			return "true", nil
		}
		return "false", nil

	case TypeEnum:
		if len(vs) > 1 {
			return "", repeated(fn, p)
		}
		v := strings.TrimSpace(vs[0])
		// Values 为空时跳过校验（防御：一个没带 values 的 enum 在引擎那边会接受任意字符串，
		// 我们跟着放开，免得把引擎能处理的值在本地判死）。
		if len(p.Values) > 0 && !contains(p.Values, v) {
			return "", &UsageError{
				Msg:    "参数 " + p.Name + " 的取值不在允许集合内：" + strconv.Quote(vs[0]),
				Detail: []string{"函数 " + fn, "可用：" + strings.Join(p.Values, "|")},
			}
		}
		return jstr(vs[0]), nil
	}

	// string / path / handle / kind / attr / 未知类型：一律按字符串发。
	//
	// 未知类型不报错是刻意的（见 KnownType）：「引擎加了一个新类型」时，
	// 按字符串发过去至少能让引擎给出明确拒绝（E_BAD_PARAM），
	// 而在本地判成「未知类型」会让一个合法调用永远到不了引擎。
	if len(vs) > 1 {
		return "", repeated(fn, p)
	}
	return jstr(vs[0]), nil
}

func repeated(fn string, p *Param) error {
	return &UsageError{
		Msg:    "参数 " + p.Name + " 给了一次以上（它不是列表）",
		Detail: []string{"函数 " + fn, "可重复的只有 path[] / string[] 类型（--paths a --paths b）"},
	}
}

// parseBool 认这几种写法（大小写不敏感）：true/1/y/yes/t 与 false/0/n/no/f。
//
// 比引擎侧 tzs-cli 的 Flags() 宽：那边只认小写的 true/1/yes。这里是给人打字的入口，
// `--force TRUE` 被拒只会浪费一轮；而 Bool 的值最终只有 true/false 两种，
// 宽一点不会有歧义（与 TypeProblem 的布尔校验一致）。
func parseBool(s string) (val, ok bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "y", "yes", "t":
		return true, true
	case "false", "0", "n", "no", "f":
		return false, true
	}
	return false, false
}

func contains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

// jstr 把一个字符串编成 JSON 字符串（关掉 HTML 转义，见 MarshalRequest 的注释）。
func jstr(s string) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		// encoding/json 对 string 不会失败；真失败也只能给出一个合法的 JSON 串。
		return `""`
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func paramNames(f *SpecFn) []string {
	out := make([]string, 0, len(f.Args))
	for _, p := range f.Args {
		out = append(out, p.Name)
	}
	return out
}

// clipNames 把参数/函数名清单裁短（报错文案里不该出现一屏 49 个名字）。
func clipNames(names []string) string {
	if len(names) == 0 {
		return "(无)"
	}
	const max = 12
	if len(names) <= max {
		return strings.Join(names, " ")
	}
	return fmt.Sprintf("%s … 共 %d 个", strings.Join(names[:max], " "), len(names))
}
