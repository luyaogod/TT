package debug

// T2.1 契约测试：把 WP1 冻结契约（`.agents/plan/tt-refactor-contract.md`）里**可断言的形状**
// 变成断言。值全部取自同包的 `debugcontract_const_test.go`（**唯一出处**），
// 本文件里不出现散落的键名 / 码值 / id 文法字面量。
//
// 三条规矩：
//
//  1. **契约条款编号写进注释**（C5.1 / 4.1.1 / F7 …），让"测试 ↔ 契约"可追溯；
//  2. 今天还没实现的那部分用 `t.Skip`，文案写清**等什么**（等 WP3 的哪个东西），
//     并且**每一条都登记进 `internal/testkit` 的跳过台账**；
//  3. 失败信息要能照做：说清期望什么、实际什么、怎么办。
//
// 不在这里测的：`internal/debug/session.go` 主流程（要真 SSH + 真 fgldb）——
// 见本文件末尾「T2.4 不测什么」一节。

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"tt/internal/output"
	"tt/internal/target"
)

// ---------------------------------------------------------------------------
// T2.1-a 信封键集（C5.1 / C5.2 / C5.8 / C5.10 / C5.12）
// ---------------------------------------------------------------------------

// TestContractSessionEnvelopeKeysMatchSessionBrief 钉住**成功信封 `session` 块的键集**
// 与 `SessionBrief` 的字段集**完全一致**（C5.1 + C5.2 + C5.3）。
//
// 为什么值得：这是"契约里的键名到底有没有落进代码"的唯一机械判据。
// 它是**两向**的，因此能同时挡住两件事：
//
//   - 有人给 `SessionBrief` 改名或删键 → 契约 C5.3 明写「一个旧键都不改、改名、不删」；
//   - 有人往信封里塞了契约没写的第三份形状 → 前端与 CLI 会照契约解析，多的那个键没人认识。
//
// 它**不**断言"恒定出现"（那是 C5.1 的 omitempty 纪律，要搭真会话才验得了）：
// 这里断言的是**键的存在性与拼写**，恒定出现由 `TestContractFailureEnvelopeKeysAreStable`
// 的失败信封那一侧 + WP3 落地后的端到端用例承担。
func TestContractSessionEnvelopeKeysMatchSessionBrief(t *testing.T) {
	want := map[string]bool{}
	for _, k := range debugContractSessionEnvelopeKeys {
		want[k] = true
	}
	for _, k := range debugContractSessionOptionalKeys {
		if want[k] {
			t.Errorf("契约常量表自相矛盾：%q 同时出现在恒定键与 omitempty 键里 —— "+
				"一个键只能属一边，请改 debugcontract_const_test.go", k)
		}
		want[k] = true
	}

	rt := reflect.TypeOf(SessionBrief{})
	got := map[string]bool{}
	for i := 0; i < rt.NumField(); i++ {
		tag := rt.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			t.Errorf("SessionBrief.%s 没有 json 标签 —— 它是对外键集的一部分，"+
				"没有标签就会被 encoding/json 按字段名输出（=%q），契约里的键名当场失效",
				rt.Field(i).Name, rt.Field(i).Name)
			continue
		}
		name := strings.Split(tag, ",")[0]
		got[name] = true
	}

	var missing, extra []string
	for k := range want {
		if !got[k] {
			missing = append(missing, k)
		}
	}
	for k := range got {
		if !want[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(extra) > 0 {
		t.Errorf("SessionBrief 上有、契约键集里没有这些键：%v\n"+
			"    怎么办：要么它属于 C5.2 的 SessionBrief 键集（加进 debugContractSessionOptionalKeys），\n"+
			"    要么它不该出现在信封里。别让契约与代码各说一套。", extra)
	}
	if len(missing) > 0 {
		// 只允许缺 C5.3 定稿的那六个**纯新增**键（Q22）。缺了别的 = 有人改了/删了旧键，
		// 那是 C5.2 / C5.3 明写不许的（"一个旧键都不改、不改名、不删"）。
		newOnes := map[string]bool{}
		for _, k := range debugContractNewSessionKeys {
			newOnes[k] = true
		}
		var illegal []string
		for _, k := range missing {
			if !newOnes[k] {
				illegal = append(illegal, k)
			}
		}
		if len(illegal) > 0 {
			t.Errorf("契约里要求 `session` 块有、SessionBrief 上却没有这些键：%v\n"+
				"    它们**不在** C5.3 的新增名单（%v）里，所以这是旧键被改名或删掉了 —— "+
				"C5.2 / C5.3 明写「一个旧键都不改、不改名、不删」。",
				illegal, debugContractNewSessionKeys)
		}
		if len(illegal) == 0 {
			t.Skipf("等 WP3 把 C5.3 / Q22 的新增键加进 SessionBrief，今天还缺 %v", missing)
		}
	}
}

// TestContractFlagPlacementIsNotAssertedHere 记账：**标志落在哪些命令上这条断言不在这里**。
//
// WP2 交付时这里有一条跳过（`--session` / `--name` 没被任何命令注册，B19/B20），
// 它带一个"落地了就改成断言"的守卫。WP4a 把两个标志注册上去之后，那个守卫响了 ——
// 按它自己的交代，这里改成了断言，但**断言搬去了 `internal/cli/debug`**：
//
//   - `internal/cli/debug` 才有那棵命令树（`debug.Group`）；
//   - 而 `internal/debug` **不能** import `internal/cli/debug` —— 那是一个
//     import cycle（`internal/cli/debug` 本来就 import `internal/debug`），
//     编译期就报 `import cycle not allowed in test`。
//
// 所以这条留下的是一行**指路**，断言本体见 `internal/cli/debug/target_test.go` 的
// `TestCharTargetFlagPlacement`（逐条对着契约 1. 的命令角色表：哪些命令认 `--session`、
// 哪些命令不许认它、`--name` 只在起会话的命令上）。
func TestContractFlagPlacementIsNotAssertedHere(t *testing.T) {
	// 本文件里凡是"要摸 CLI 命令树"的断言都在别的包 —— 这条不该有跳过，
	// 它只是把"为什么不在 debug 包里测"记在案上（免得上层的契约测试清单里少一块没人发现）。
	t.Log("标志落点断言在 internal/cli/debug：TestCharTargetFlagPlacement（契约 1. 的命令角色表 / F16 / F17）")
}

// TestContractFailureEnvelopeKeysAreStable 钉住失败信封的**四键恒定 + 两键按条件**
// （C5.10 / C5.11），并用 `output.WriteError` 的**真实输出**核对，而不是只看常量表。
//
// 为什么用真实输出：契约说"恒定出现"，指的是**序列化出来的 JSON 里有这个键**。
// 只在常量表里写四个字符串，等于把"断言"写成了"抄写"。
//
// C5.15 给 HTTP 侧留了另一条：那边是 `{"ok":false,"error":…}` 加 `code`/`candidates`，
// 形状一字不改 —— 两侧的差别由 `TestContractHTTPFailureEnvelopeKeepsItsShape` 钉。
func TestContractFailureEnvelopeKeysAreStable(t *testing.T) {
	var b strings.Builder
	we := output.Errorf(output.CodeUsage, debugContractCodeExits[debugContractCodeUsage], "示例错")
	if err := output.WriteError(&b, we); err != nil {
		t.Fatalf("output.WriteError 失败：%v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(b.String()), &raw); err != nil {
		t.Fatalf("失败信封不是合法 JSON：%v\n%s", err, b.String())
	}
	for _, k := range debugContractFailureRequredKeys {
		if _, ok := raw[k]; !ok {
			t.Errorf("失败信封（C5.10）必须恒定有 %q 键，实际只有 %v\n"+
				"    怎么办：键集在 internal/output/errors.go 的 errorEnvelope。", k, keysOf(raw))
		}
	}
	// 反向：常量表里的每个键都必须真的在输出里（已经在上面的循环里逐条报过缺失）。
	// hint 没给时必须**不出现**（omitempty）：契约说"有修复建议才出现"。
	if _, ok := raw["hint"]; ok {
		t.Errorf("没给修复建议时失败信封不该有 hint 键（C5.10：有修复建议才出现），实际 %v", keysOf(raw))
	}
	// candidates 只在 C5.11 那五种码下出现；这条码（USAGE）不在其中。
	if _, ok := raw["candidates"]; ok {
		t.Errorf("code=%s（C5.11 不在必须带候选的五种码里）时不该有 candidates 键，实际 %v",
			debugContractCodeUsage, keysOf(raw))
	}
}

// TestContractCandidateKeysAreASubsetOfSessionKeys 钉住 C5.12 的**唯一出处**纪律：
// 候选九键必须是 `session` 键集的**真子集**（C5.4b：同一个 struct 的投影，不许手抄第二份 JSON 定义）。
func TestContractCandidateKeysAreASubsetOfSessionKeys(t *testing.T) {
	inSession := map[string]bool{}
	for _, k := range debugContractSessionEnvelopeKeys {
		inSession[k] = true
	}
	for _, k := range debugContractCandidateKeys {
		if !inSession[k] {
			t.Errorf("候选键 %q 不在 `session` 块的键集里 —— C5.12 要求九键是它的**真子集**，\n"+
				"    这样同一个渲染函数才能既打身份行又打候选行。要么改候选键集，要么改 session 键集，\n"+
				"    两处都在 debugcontract_const_test.go。", k)
		}
	}
	if len(debugContractCandidateKeys) != 9 {
		t.Errorf("C5.12 写死候选是**九键**，常量表里有 %d 个：%v",
			len(debugContractCandidateKeys), debugContractCandidateKeys)
	}
}

// TestContractCandidatesRequiredCodesMatchTheInvalidFormTable 双向核对
// C5.11 的"五种码必须带候选"与 2.4 那张表里标了"必须出现"的那些 F 条目。
//
// 这两处一个在 5.2、一个在 2.4，只是**引用关系**：表里少写一个码，就会出现
// "报歧义却不列候选"——调用方只能自己去列，那正是这次要根除的隐性行为。
func TestContractCandidatesRequiredCodesMatchTheInvalidFormTable(t *testing.T) {
	fromTable := map[string]bool{}
	for _, f := range debugContractInvalidForms {
		if f.Candidates {
			fromTable[f.Code] = true
		}
	}
	fromContract := map[string]bool{}
	for _, c := range debugContractCandidatesRequiredCodes {
		fromContract[c] = true
	}
	for c := range fromContract {
		if !fromTable[c] {
			t.Errorf("C5.11 说 %s 必须带候选清单，但 2.4 的表里没有任何一条标了 candidates=true —— "+
				"两处对不上，改哪边都要同步（都在 debugcontract_const_test.go）", c)
		}
	}
	for c := range fromTable {
		if !fromContract[c] {
			t.Errorf("2.4 的表里有 candidates=true 的条目用了 %s，但 C5.11 的五种码里没有它 —— "+
				"要么它真该带候选（加进 debugContractCandidatesRequiredCodes），要么那一条标错了", c)
		}
	}
}

// ---------------------------------------------------------------------------
// T2.1-b 七个机器码与退出码（4.1 / 4.1.1 / C5.13）
// ---------------------------------------------------------------------------

// TestContractCodeExitTable 逐条钉住 4.1.1 那张 `code` → 退出码的表。
//
// 用 `output.Error` 的真实字段核对（Code + Exit），而不是只比常量表：
// 契约说"机器码与退出码同源"（C4.1），这张测试就是那句话的判据。
func TestContractCodeExitTable(t *testing.T) {
	if len(debugContractCodeExits) == 0 {
		t.Fatal("退出码表是空的 —— 契约 4.1.1 里明明有一张表，常量表漏抄了")
	}
	codes := make([]string, 0, len(debugContractCodeExits))
	for c := range debugContractCodeExits {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	for _, code := range codes {
		wantExit := debugContractCodeExits[code]
		e := output.Errorf(code, wantExit, "示例")
		if e.Code != code {
			t.Errorf("output.Error.Code = %q，期望 %q", e.Code, code)
		}
		if e.Exit != wantExit {
			t.Errorf("code=%s 的退出码 = %d，契约 4.1.1 写的是 %d\n"+
				"    怎么办：改的是 debugContractCodeExits（契约那一列），不是这里。",
				code, e.Exit, wantExit)
		}
	}
	// USAGE 的两档口径（4.1.1 的两条）：新写法退 2，老路径（F16/F17）退 1。
	if got := debugContractCodeExits[debugContractCodeUsage]; got != debugContractExitUsage {
		t.Errorf("新写法下 USAGE 该退 %d，表里是 %d（4.1.1 第一条口径）", debugContractExitUsage, got)
	}
	if debugContractFlagOnWrongCommandExit != debugContractExitRunFailed {
		t.Errorf("F16/F17 走老路径该退 %d（Q18：不补 FlagErrorFunc），常量写的是 %d",
			debugContractExitRunFailed, debugContractFlagOnWrongCommandExit)
	}
}

// TestContractNewCodesAreRegisteredInOutput 钉住 C5.13：
// 五个新码要加进 `internal/output/errors.go` 的常量清单（加新码可以，改旧码名字不行）。
//
// 今天一个都没有 → 跳过，并登记台账（"等 WP3/WP4 把新码加进 output"）。
// 一旦加了，这条测试会逐字核对**取值**：常量名可以叫 `CodeSessionAmbiguous`，
// 但它必须等于契约里那个字符串。
func TestContractNewCodesAreRegisteredInOutput(t *testing.T) {
	// 新增码（C5.13 要求加进 output 的那五个）。
	// 键就是常量本身 —— 这里不重写字面量，否则"值只定义一次"当场破功。
	newCodes := []string{
		debugContractCodeAmbiguous,
		debugContractCodeNotFound,
		debugContractCodeStale,
		debugContractCodeTargetBusy,
		debugContractCodeNameTaken,
	}
	// 复用现行常量的四个（B5 / C5.13 末句）：今天就有，所以**现在就核对**。
	t.Run("复用的现行码取值不许变", func(t *testing.T) {
		cases := []struct{ code, want string }{
			{output.CodeUsage, debugContractCodeUsage},
			{output.CodeEnvUnknown, debugContractCodeEnvUnknown},
			{output.CodeNoSession, debugContractCodeNoSession},
			{output.CodeSessionState, debugContractCodeState},
		}
		for _, c := range cases {
			if c.code != c.want {
				t.Errorf("internal/output 的常量取值 = %q，契约 4.1.1 写的是 %q\n"+
					"    C5.13：加新码可以，**改旧码的名字/取值不行**（它是 agent 分支的依据）。",
					c.code, c.want)
			}
		}
	})
	t.Run("五个新增码", func(t *testing.T) {
		var absent []string
		for _, want := range newCodes {
			if !outputHasCode(want) {
				absent = append(absent, want)
			}
		}
		if len(absent) > 0 {
			sort.Strings(absent)
			t.Skipf("等 WP3/WP4 把 4.1.1 的新码加进 internal/output/errors.go 的常量清单（C5.13）：%s",
				strings.Join(absent, "、"))
		}
		// 加进来了就逐字核对取值（常量名随意，取值必须等于契约里的字符串）。
		e := output.Errorf(debugContractCodeAmbiguous, debugContractCodeExits[debugContractCodeAmbiguous], "x")
		if e.Code != debugContractCodeAmbiguous {
			t.Errorf("output 侧 SESSION_AMBIGUOUS 的取值 = %q，契约写的是 %q", e.Code, debugContractCodeAmbiguous)
		}
	})
}

// outputHasCode 判断 internal/output 里是否已经有这个取值的稳定错误码常量。
//
// 用一份**显式清单**：Go 里没有"枚举某个包的常量"的反射手段（常量不是反射对象），
// 所以这里只能把现行清单列出来。它的作用是"新码有没有进 output"这个机械判据 ——
// 而不是靠人去 grep 关键字面量（关键字面量在别处也会出现，那样判不准）。
//
// 清单随 output 增长时要同步（多加一个码忘了同步，只影响这条测试的**严格度**：
// 认不出的码会被当成"还没落地"而走跳过分支，不会假绿任何别的东西）。
func outputHasCode(want string) bool {
	for _, v := range []string{
		output.CodeUsage, output.CodeEnumInvalid, output.CodeEnvUnknown, output.CodeConfigBroken,
		output.CodeTableMissing, output.CodeConnFailed, output.CodeSQLRejected, output.CodeQueryFailed,
		output.CodeNotFound, output.CodeNoSession, output.CodeSessionState, output.CodeTargetMissing,
		output.CodePreflight, output.CodeUpdateUsage, output.CodeUpdateRefused,
		output.CodeUpdateChecksum, output.CodeUpdateNetwork, output.CodeUpdateApply,
	} {
		if v == want {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// T2.1-c 规范 id 的文法（2.1.2 / C2.0 / Q1-b）
// ---------------------------------------------------------------------------

// TestContractRefGrammar 钉住**规范 id 的文法**（C2.1.2 的定稿）：
//
//	规范 id := 环境名 "#" 序号        序号 := 十进制整数，>= 1
//
// 三条要点，都是定稿改过的：
//   - 地址**不含作业段、不含企业编号**（早先的 `环境/作业#序号` 因为 SetRun 会改写 prog 被收回，
//     `ent` 因为 SetTopent 可变被收回 —— C2.0 的"生命周期内不变"是根依据）；
//   - 环境名**精确匹配**，不做前缀、不做大小写折叠（C2.1）；
//   - 序号 >= 1（`#0` 非法，F9）。
func TestContractRefGrammar(t *testing.T) {
	re, err := regexp.Compile(debugContractRefPattern)
	if err != nil {
		t.Fatalf("规范 id 的正则编译不过（debugContractRefPattern）：%v", err)
	}
	if !strings.Contains(debugContractRefPattern, debugContractSeqSep) {
		t.Errorf("规范 id 的正则里没有序号分隔段 %q —— 它与 debugContractSeqSep 必须同源", debugContractSeqSep)
	}

	valid := []string{"恒烁正式区#1", "测试1012#12", "env#999999", "带 空格的环境#3", "a#1"}
	for _, s := range valid {
		if !re.MatchString(s) {
			t.Errorf("%q 该是合法的规范 id（环境名%q序号，2.1.2），正则 %q 却说不是",
				s, debugContractSeqSep, debugContractRefPattern)
		}
	}
	invalid := []struct{ s, why string }{
		{"恒烁正式区", "没有序号段 —— 光有环境名是「按环境找」（写法 ⑧），不是规范 id"},
		{"恒烁正式区#", "序号为空（F9 的 `#`）"},
		{"恒烁正式区#0", "序号必须 >= 1（F9）"},
		{"恒烁正式区#-1", "序号必须是十进制正整数（F9）"},
		{"恒烁正式区#1.5", "序号必须是十进制正整数（F9）"},
		{"恒烁正式区#x", "序号必须是十进制正整数（F9）"},
		{"恒烁正式区#01", "序号不写前导零 —— 与 #1 是同一个地址，写成两个会对不上账"},
		{"恒烁正式区/apmt520#1", "这是**人读形态**（display，写法 B），不是规范 id：作业段会随 SetRun 漂（C5.4）"},
		{"#1", "环境名不能为空（规范 id 是「环境名#序号」）"},
	}
	for _, c := range invalid {
		if re.MatchString(c.s) {
			t.Errorf("%q 不该被当成规范 id：%s", c.s, c.why)
		}
	}
}

// TestContractSelectorSecondSegmentKind 钉住 C2.3 的判据：
// 选择器第二段**全数字 → 企业编号；否则 → 作业名**，两种形态不会撞
// （作业名首字符是字母，B29），所以不需要额外消歧规则。
func TestContractSelectorSecondSegmentKind(t *testing.T) {
	for _, s := range []string{"99", "1012", "0", "1"} {
		if got := selectorSecondSegKind(s); got != debugContractSecondSegEnt {
			t.Errorf("第二段 %q 该判成企业编号（C2.3：全数字），得 %q", s, got)
		}
	}
	for _, s := range []string{"apmt520", "bsft001", "a", "A_1-x"} {
		if got := selectorSecondSegKind(s); got != debugContractSecondSegJob {
			t.Errorf("第二段 %q 该判成作业名（C2.3：非全数字），得 %q", s, got)
		}
	}
	// 判据的分界在「全数字」这四个字上：判据**不**先做 B29 的合法性校验，
	// 所以 `1a` 这种形态（B29 下不可能是合法作业名）按 C2.3 落到作业那一支，
	// 随后会在"环境/作业"那一步按 F7 报 SESSION_NOT_FOUND。
	// 它**不会**被静默当成企业编号 1 —— 那才是"看起来落上了、其实落在别处"。
	for _, s := range []string{"1a", "apmt520a1", "99x"} {
		if got := selectorSecondSegKind(s); got != debugContractSecondSegJob {
			t.Errorf("C2.3 的判据只看「全数字」：第二段 %q 含非数字字符，应当判成作业名（%q），得 %q。\n"+
				"    注：B29 下这类串不是合法作业名，所以它最终按 F7（SESSION_NOT_FOUND）被拒 —— "+
				"这一支绝不能静默退回企业编号，那会变成「看起来落上了、其实落在别处」。",
				s, debugContractSecondSegJob, got)
		}
	}
	// 作业名合法字符集本身也要钉（B29 / 2.1.4）。
	jre := regexp.MustCompile(debugContractJobPattern)
	for _, s := range []string{"apmt520", "a", "A_1-x", strings.Repeat("z", 64)} {
		if !jre.MatchString(s) {
			t.Errorf("作业名 %q 该合法（B29 的正则 %q）", s, debugContractJobPattern)
		}
	}
	for _, s := range []string{"1apmt", "apmt.520", "apmt 520", "", strings.Repeat("z", 65), "作业"} {
		if jre.MatchString(s) {
			t.Errorf("作业名 %q 该非法（B29 的正则 %q）", s, debugContractJobPattern)
		}
	}
}

// ---------------------------------------------------------------------------
// T2.1-d 非法写法一览 F1–F20（2.4 + Q23 的 F20）
// ---------------------------------------------------------------------------

// TestContractInvalidFormsAreComplete 钉住那张表**本身**：F1–F20 逐条都在、没有空壳。
//
// 有人删掉一条非法写法，这张测试当场红 —— 那张表就是将来 TDD 的输入集，
// 少一条等于少验一种"说不清的错"。
func TestContractInvalidFormsAreComplete(t *testing.T) {
	const total = 20
	if len(debugContractInvalidForms) != total {
		t.Fatalf("非法写法表该有 %d 条（F1–F20），实得 %d 条", total, len(debugContractInvalidForms))
	}
	seen := map[string]bool{}
	for i, f := range debugContractInvalidForms {
		wantKey := fmt.Sprintf("F%d", i+1)
		if f.Key != wantKey {
			t.Errorf("第 %d 条该是 %s，表里写的是 %q —— 键必须连号，否则「逐条覆盖」这句话就没人验得了",
				i+1, wantKey, f.Key)
		}
		if seen[f.Key] {
			t.Errorf("非法写法 %s 登记了两次", f.Key)
		}
		seen[f.Key] = true
		if strings.TrimSpace(f.Write) == "" {
			t.Errorf("%s 没有写下「怎么写出这个非法写法」—— 那是将来照着敲的输入，不能空着", f.Key)
		}
		if strings.TrimSpace(f.Code) == "" {
			t.Errorf("%s 没有机器码（2.4 表里那一列）", f.Key)
		}
		switch f.Exit {
		case debugContractExitUsage, debugContractExitRunFailed, debugContractExitNoSession,
			debugContractExitState, debugContractExitMissing, debugContractExitPreflight:
		default:
			t.Errorf("%s 的退出码 = %d，不在契约 4.1 的码表里（0/1/2/3/4/5/6）", f.Key, f.Exit)
		}
		// 2.4 的"机器码"一列必须落在 4.1.1 那张唯一清单里（C5.12 的同一条纪律：
		// 一个规矩只定义一处）。
		if _, ok := debugContractCodeExits[f.Code]; !ok {
			t.Errorf("%s 用的机器码 %q 不在 debugContractCodeExits（4.1.1 的唯一清单）里 —— "+
				"要么这条写错了码，要么 4.1.1 漏了这个码", f.Key, f.Code)
		}
	}
}

// TestContractInvalidFormsCodeSequence 把 2.4 那张表的机器码**逐条**钉下来。
//
// 这条测试的作用是"改一处必须改两处"：谁动了某条 F 的码，这里立刻红。
// 码的顺序就是契约 2.4 那张表从上到下的顺序（F1…F20）。
func TestContractInvalidFormsCodeSequence(t *testing.T) {
	want := []struct {
		key  string
		code string
	}{
		{"F1", debugContractCodeAmbiguous},
		{"F2", debugContractCodeNoSession},
		{"F3", debugContractCodeAmbiguous},
		{"F4", debugContractCodeAmbiguous},
		{"F5", debugContractCodeNotFound},
		{"F6", debugContractCodeEnvUnknown},
		{"F7", debugContractCodeNotFound},
		{"F8", debugContractCodeNotFound},
		{"F9", debugContractCodeUsage},
		{"F10", debugContractCodeUsage},
		{"F11", debugContractCodeUsage},
		{"F12", debugContractCodeUsage},
		{"F13", debugContractCodeUsage},
		{"F14", debugContractCodeUsage},
		{"F15", debugContractCodeStale},
		{"F16", debugContractCodeUsage},
		{"F17", debugContractCodeUsage},
		{"F18", debugContractCodeNameTaken},
		{"F19", debugContractCodeUsage},
		{"F20", debugContractCodeTargetBusy},
	}
	byKey := map[string]debugContractInvalidForm{}
	for _, f := range debugContractInvalidForms {
		byKey[f.Key] = f
	}
	for _, w := range want {
		f, ok := byKey[w.key]
		if !ok {
			t.Errorf("非法写法表里没有 %s", w.key)
			continue
		}
		if f.Code != w.code {
			t.Errorf("%s 的机器码 = %q，契约 2.4 写的是 %q\n"+
				"    怎么办：改哪边都要同步（表是 internal/target 的 InvalidForms，"+
				"debugcontract_const_test.go 只是引用它）",
				w.key, f.Code, w.code)
		}
	}
}

// TestContractInvalidFormsAreNotYetEnforced 是 F1–F20 里**服务端还没做的那部分**的记账。
//
// 为什么必须有一条这样的测试而不是不写：这二十条是本次重构的**验收输入集**。
// 没有它，WP4 全部完成后没人知道"是不是真的逐条跑了"。
//
// **WP4a 之后的现状（改这条的理由）**：
//
//   - 目标层（`internal/target`）已交付，且 WP4a 把它接进了 CLI ——
//     所以"选择器文法 / 落位 / 码 / 退出码 / 候选政策"这几层**已经有真实实现**，
//     逐条断言写在 `internal/target` 的测试里（`resolve_test.go` / `selector_test.go` /
//     `values_test.go` 的 `TestInvalidFormTableIsComplete`）；
//   - CLI 侧新写的那几条（F1 的"多会话不取第一项"+候选、退出码分档、会话路径只认 id）
//     在 `internal/cli/debug` 的 `target_test.go` 里；
//   - **还差的是"端到端"那一层**：F16 要求"标志挂错命令"的信封带 `code: USAGE`
//     （C2.8c），而 HTTP 侧的失败信封今天只有 `{ok,error}` 两键（B8）——
//     给 HTTP 侧补 `code` / `candidates` 是 **WP4c**（C5.15 / C5.16），本轮不许动。
//     所以"逐条跑 F1–F20 的**实测**（真 CLI 进程 + 真服务端）"这件事仍然只能等。
//
// 守卫换成**不再会误响**的那一个：它问"服务端的失败信封有没有 code"，
// 而不是"CLI 有没有注册 --session"（那个问题在 WP4a 已经翻页了）。
func TestContractInvalidFormsAreNotYetEnforced(t *testing.T) {
	if failureEnvelopeHasCode(t) {
		t.Fatalf("服务端失败信封已经有 code 了（C5.15 落地），请把这条跳过改成逐条跑 F1–F20 的实测。"+
			"当前表里有 %d 条：%v", len(debugContractInvalidForms), invalidFormKeys())
	}
	t.Skip("等 WP4c 给 HTTP 失败信封补 code/candidates（C5.15/C5.16）后，逐条端到端跑 F1–F20（契约 2.4 的输入集 B）")
}

// TestContractFlagConflictsIsNotYetEnforced 钉住四个标志的冲突矩阵（C3.3 的四种组合）
// 里**今天可测的两个半**：
//
//   - F12 / F13：`--session` 带的环境/企业与 `--env` / `--ent` 矛盾 → 退 2、`code: USAGE`；
//   - F14：`--session` 里的会话实际 ENT 与 `--ent` 不同 → 退 2、`code: USAGE`。
//
// **WP4a 之后这三条已经可断言了**（目标层是纯函数，`internal/target` 的
// `TestResolveF12EnvConflict` / `F13` / `F14` 逐条断言了码与退出码；CLI 那侧的包装
// 由 `internal/cli/debug/target_test.go` 钉住 `ExitFor(code, NewScenario, newSyntax)` 的分档 ——
// 分档判据是"**这个场景是不是新场景**"，不是"有没有打标志"，见 `target.Failure.NewScenario`）。
//
// 所以这里**不再跳过**：本包只做一件它能做的事 —— 核对 debugcontract 常量表里
// 这三条 F 的码与 `internal/target` 的常量一致（跨包字面量的唯一防线在那边，
// 这里只保证"本包引用的那份没漂"）。
func TestContractFlagConflictsIsNotYetEnforced(t *testing.T) {
	want := map[string]string{
		"F12": debugContractCodeUsage,
		"F13": debugContractCodeUsage,
		"F14": debugContractCodeUsage,
	}
	byKey := map[string]debugContractInvalidForm{}
	for _, f := range debugContractInvalidForms {
		byKey[f.Key] = f
	}
	for k, code := range want {
		f, ok := byKey[k]
		if !ok {
			t.Fatalf("非法写法表里没有 %s —— C3.3 的四种组合里这三条必须登记在案", k)
		}
		if f.Code != code {
			t.Errorf("%s 的机器码 = %q，C3.3 要求 %q", k, f.Code, code)
		}
		if f.Candidates {
			t.Errorf("%s 不该带候选清单（C5.11：标志冲突不是「该列候选」的场合）", k)
		}
	}
}

// TestContractAliasRulesAreNotYetEnforced 曾经钉住 `--name` 的取值纪律
// （C2.4 / C2.5 / F19 / Q17）：**非空、去空白后 1–64 字符、只排除 `/` 与 `#`**，
// 其余（中文、空格、标点）按原样允许。
//
// **WP4a 之后这条不再跳过**：`--name` 已在 `tt debug start` 上注册（B20 反面），
// 而取值纪律的实现与逐条判据在 `internal/target`（`ValidateAlias` 与它的
// `TestValidateAliasRules` / `TestAliasLengthCountsRunesNotBytes`）。
// 本包能做的、也只有意义的一件事是：核对本包引用的那份常量与契约逐字一致 ——
// 它们与 target 是一套值（跨包字面量由 `internal/target` 的 values 测试守着）。
func TestContractAliasRulesAreNotYetEnforced(t *testing.T) {
	if debugContractAliasForbid != "/#" {
		t.Fatalf("别名排除字符集常量 = %q，契约 C2.4/C2.5 写的是 %q", debugContractAliasForbid, "/#")
	}
	if debugContractAliasMinRunes != 1 || debugContractAliasMaxRunes != 64 {
		t.Fatalf("别名长度常量 = %d–%d，契约 C2.5 写的是 %d–%d",
			debugContractAliasMinRunes, debugContractAliasMaxRunes, 1, 64)
	}
	if target.AliasForbid != debugContractAliasForbid ||
		target.AliasMinRunes != debugContractAliasMinRunes ||
		target.AliasMaxRunes != debugContractAliasMaxRunes {
		t.Errorf("本包引用的一份（forbid=%q %d–%d）与 internal/target 的实现（forbid=%q %d–%d）不一致 ——\n"+
			"    两处必须同值：实现与判据同源，改一处就要改另一处。",
			debugContractAliasForbid, debugContractAliasMinRunes, debugContractAliasMaxRunes,
			target.AliasForbid, target.AliasMinRunes, target.AliasMaxRunes)
	}
}

// ---------------------------------------------------------------------------
// T2.1-e 规范 id 进出信封（C5.3 / C5.4 / F15 / Q22）
// ---------------------------------------------------------------------------

// TestContractRefKeyIsNotYetInSessionBrief 是 `ref` / `name` / `n` / `epoch` 四个新键的跳过。
//
// 它存在的意义：C5.3 说这四个键**纯新增、不改旧键**。落地之后，
// `TestContractSessionEnvelopeKeysMatchSessionBrief` 会自动开始要求它们；
// 在那之前，用这一条把"还差这四个键"记在案上，而不是让上面那条测试的失败信息孤零零地挂着。
func TestContractRefKeyIsNotYetInSessionBrief(t *testing.T) {
	if contractRefKeyLanded() {
		t.Fatal("ref 键已经加进 SessionBrief 了，请删掉这条跳过（C5.3 的四个新键已落地）")
	}
	t.Skip("等 WP3 实现 ref 字段（并同批补 name / n / epoch，契约 C5.3 / Q22）")
}

// TestContractEpochStalenessIsNotYetEnforced 是 F15（句柄过期）的跳过。
//
// F15 要求"环境与序号都对，但句柄是上一代"时报 `SESSION_STALE` + 候选块。
// 今天**没有"代"这个概念**（B31：SessionBrief 里无 epoch/generation），
// 所以这条无从判起 —— 等 WP3/WP4 把 epoch 落进 SessionBrief 并接上校验。
func TestContractEpochStalenessIsNotYetEnforced(t *testing.T) {
	if contractRefKeyLanded() {
		t.Fatal("epoch 已经落地了，请把这条跳过改成 F15 的世代校验断言")
	}
	t.Skip("等 WP3 实现 epoch 字段与世代校验后，断言 F15 的 SESSION_STALE 报错（C5.3 / C5.11）")
}

// TestContractNameIdempotencyIsNotYetEnforced 是 Q23 / F20 的记账：
// `--name X` 的别名当幂等键 —— 空闲就复用、正在跑/已停站报 `SESSION_TARGET_BUSY` 退 3。
//
// **WP4a 之后还差什么（这条守卫换过了）**：`--name` 已经在 `tt debug start` 上注册，
// 但"别名当幂等键"要的是**服务端记得住别名**（`SessionBrief.name`，C5.3 的纯新增键之一），
// 而服务端今天不吐 `name` / `ref` / `n` / `epoch`（那是 WP4c 的活）。
// 目标层那侧已经把三种结局实现好了（`ResolveName` → `NameFree` / `NameReuse` / `NameBusy`，
// 断言在 `internal/target` 的 `TestResolveNameIdempotency`）。
//
// 所以守卫从"CLI 有没有注册 --name"（这个已经翻页）换成"服务端记不记得别名"。
func TestContractNameIdempotencyIsNotYetEnforced(t *testing.T) {
	if contractRefKeyLanded() {
		t.Fatal("SessionBrief 已经有 ref/name/n/epoch 了，请把这条跳过改成 Q23 的端到端幂等断言" +
			"（空闲复用 / 忙碌报 SESSION_TARGET_BUSY）")
	}
	t.Skip("等 WP4c 把 name 键加进 SessionBrief（C5.3）后，断言 Q23：空闲复用、忙碌报 SESSION_TARGET_BUSY（F20 / C1.6）")
}

// ---------------------------------------------------------------------------
// T2.1-f 人话输出（C5.9 身份行 / C5.14 候选行）
// ---------------------------------------------------------------------------

// tryContractCandidateLine 按 C5.14 渲染一行人话候选块：
//
//	<ref>  <name>  <state>  ENT <ent>
//
// 两格缩进、字段间两个空格；`<name>` 未命名写 `(未命名)`。
// 第一个字段是 `ref`（恒定不变的规范 id），**不是** display（C5.14 的理由：候选行是给人照抄重发的）。
func tryContractCandidateLine(ref, name, state, ent string) string {
	if strings.TrimSpace(name) == "" {
		name = "(未命名)"
	}
	if strings.TrimSpace(ent) == "" {
		ent = "?"
	}
	return fmt.Sprintf("  %s  %s  %s  ENT %s", ref, name, state, ent)
}

// TestContractCandidateLineFormat 钉住 C5.14 的候选行格式。
//
// 为什么值得钉：候选行的用途是"照着它重发一枪"。少一个空格、把 ref 换成 display，
// 抄下去的就是一条会漂的写法 —— 那正是 C5.4 说的"要恒定用 ref"。
//
// 今天**没有渲染函数可调**（目标层还没落地），所以这里调的是上面那个按 C5.14 写死的
// 辅助函数，并对格式本身做断言。落地后应当删掉辅助函数，改成调真实现。
func TestContractCandidateLineFormat(t *testing.T) {
	got := tryContractCandidateLine("恒烁正式区#1", "汇率单", "已停站", "99")
	want := "  恒烁正式区#1  汇率单  已停站  ENT 99"
	if got != want {
		t.Errorf("候选行（C5.14）= %q\n    期望 %q\n"+
			"    格式：两格缩进 + <ref> + 两空格 + <name> + 两空格 + <state> + 两空格 + ENT <ent>",
			got, want)
	}
	if strings.Contains(got, "会话 ") {
		t.Errorf("候选行不该带 `会话 ` 前缀（C5.14 末段：那是 5.9 身份行的前缀，两行不是同一行）")
	}
	if got := tryContractCandidateLine("恒烁正式区#2", "", "空闲", ""); !strings.Contains(got, "(未命名)") || !strings.Contains(got, "ENT ?") {
		t.Errorf("未命名/取不到企业编号时要写字面量 (未命名) 与 ?（C5.14），实得 %q", got)
	}
	// 反向：第一个字段必须是 ref 的形态（环境#序号），不能是 display 的形态（含 /）。
	first := strings.Fields(strings.TrimSpace(want))[0]
	if strings.Contains(first, debugContractSelectorSep) {
		t.Errorf("候选行第一个字段 %q 含 %q —— 那是 display，会随会话当前作业漂（C5.14）",
			first, debugContractSelectorSep)
	}
}

// TestContractIdentityLineIsNotYetLanded 是 C5.9 身份行的跳过。
//
// C5.9 定的那一行 `会话 <display> · <name> · <state> · ENT <ent>` 要求同时印出
// `display` 与 `name` —— 而今天这两样都不存在（C5.3 的四个新键还没落地）。
// 落地前能测的只有 state 的中文标签，那一条已经在 T2.2 的表征测试里钉住。
func TestContractIdentityLineIsNotYetLanded(t *testing.T) {
	if contractRefKeyLanded() {
		t.Fatal("display/name 已经落地了，请把这条跳过改成 C5.9 身份行的逐字断言")
	}
	t.Skip("等 WP3/WP4 落地 display 与 name 后，断言 C5.9 身份行 `会话 <display> · <name> · <state> · ENT <ent>`")
}

// ---------------------------------------------------------------------------
// T2.1-g 端点与信封的两侧差别（C5.15 / C5.16 / 第 7 节的表）
// ---------------------------------------------------------------------------

// TestContractHTTPFailureEnvelopeKeepsItsShape 钉住 C5.15 的**现状半边**：
// HTTP 侧的失败信封是 `{"ok":false,"error":…}`（B8），**没有** `code`，也没有 `candidates`。
//
// 契约要求"在这形状上**增加** `code` 与 `candidates`，`ok`/`error` 一个字都不改"。
// 所以这一条今天必须绿（它钉住的是"不许动的部分"），而"新增的部分"由下面那条跳过盯着。
//
// 这与 CLI 侧（output.WriteError 的四键恒定）是**两套形状**，别混：
// HTTP 侧给前端，CLI 侧给 --json 的调用方。
func TestContractHTTPFailureEnvelopeKeepsItsShape(t *testing.T) {
	hs := actionTestServer(t)
	resp := doReq(t, "POST", hs.URL+"/api/wstest", "", "{}")
	if resp.StatusCode != 500 {
		t.Fatalf("前置不成立：期望拨号失败退 500，得 %d（%s）", resp.StatusCode, respError(t, resp))
	}
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("HTTP 失败信封不是合法 JSON：%v", err)
	}
	if raw["ok"] != false {
		t.Errorf("HTTP 失败信封的 ok 必须是 false（B8 / C5.15），实得 %v", raw["ok"])
	}
	if s, _ := raw["error"].(string); strings.TrimSpace(s) == "" {
		t.Errorf("HTTP 失败信封必须有非空的 error（B8），实得 %v", raw["error"])
	}
	// "一个字都不改"：今天**没有** code/candidates，所以它们不该出现。
	// 这条断言在 WP4 加进 code 之后必须一起改 —— 那时改的是这里，不是 C5.15。
	for _, k := range []string{"code", "candidates"} {
		if _, ok := raw[k]; ok {
			t.Errorf("HTTP 失败信封今天不该有 %q 键（B8 的形状是 {ok,error}）；"+
				"C5.15 要求加它时，请连同这条断言一起改", k)
		}
	}
}

// TestContractHTTPFailureCodeIsNotYetLanded 是 C5.15 / C5.16 的跳过：
// HTTP 侧要增加 `code`（必给）与 `candidates`（按 C5.11），缺目标/歧义用 **409**。
func TestContractHTTPFailureCodeIsNotYetLanded(t *testing.T) {
	hs := actionTestServer(t)
	resp := doReq(t, "POST", hs.URL+"/api/wstest", "", "{}")
	var raw map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&raw)
	if _, ok := raw["code"]; ok {
		t.Fatal("HTTP 失败信封已经有 code 了，请把这条跳过改成 C5.15 / C5.16 的断言")
	}
	t.Skip("等 WP4 给服务端失败信封加 code 与 candidates（C5.15 / C5.16：缺目标报 409 + SESSION_AMBIGUOUS）")
}

// TestContractStatusStaysTwoDocuments 钉住 C1.5b / Q21 里**今天就不许动**的那一半：
//
// `tt debug status` 默认输出 JSON（今天就是），而且是**两个** JSON 文档
// （`GET /api/status` 与 `GET /api/sessions` 各一行），**不合并成一个信封**。
// 契约只在每个 session 元素上按 C5.8 换成 5.1 的 `session` 对象（纯新增四个键）。
//
// 这里调的是服务端真实端点，不是拼字符串 —— 它同时是"端点没被顺手改掉"的判据。
func TestContractStatusStaysTwoDocuments(t *testing.T) {
	srv := NewServer(cfgFor("h1", "36"), nil, "")
	mux := http.NewServeMux()
	srv.routes(mux)
	hs := newTestHTTPServer(t, mux)

	// /api/status：服务级文档，与 C5.8 无关，一个键都不许动。
	resp := doReq(t, "GET", hs.URL+"/api/status", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/status 该退 200，得 %d（%s）", resp.StatusCode, respError(t, resp))
	}
	var st map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		t.Fatalf("/api/status 该回一个 JSON 文档：%v", err)
	}
	// 今天这个文档是**服务级状态**（server/ssh/zone/listen/watchdog…），**没有** ok 键。
	// C1.5b 管的是 `tt debug status` 命令默认打 JSON 这件事，两个文档的形状它一个字都没改
	// （契约原话：行数、两个文档的顺序、以及 ok 键的有无都不动）。
	// 所以这里断言的是"它在、且是个有内容的 JSON 对象"，不是"它有 ok 键"。
	if len(st) == 0 {
		t.Errorf("/api/status 的文档不该是空的（C1.5b：这两个文档今天就打，形状不许动）")
	}
	if _, ok := st["server"]; !ok {
		t.Errorf("/api/status 该有 server 键（服务级状态文档的既有形状），实得 %v", keysOf(st))
	}

	// /api/sessions：会话列表文档。今天空池就是空数组 —— 这是"默认 JSON、两行文档"里的第二行。
	resp2 := doReq(t, "GET", hs.URL+"/api/sessions", "", "")
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("/api/sessions 该退 200，得 %d（%s）", resp2.StatusCode, respError(t, resp2))
	}
	var ls map[string]any
	if err := json.NewDecoder(resp2.Body).Decode(&ls); err != nil {
		t.Fatalf("/api/sessions 该回一个 JSON 文档：%v", err)
	}
	if _, ok := ls["sessions"]; !ok {
		t.Errorf("/api/sessions 该有 sessions 键（C5.8 的宿主形状），实得 %v", keysOf(ls))
	}
}

// ---------------------------------------------------------------------------
// 小工具
// ---------------------------------------------------------------------------

// keysOf 把 map 的键排序后返回（失败信息里要用，顺序必须稳定）。
func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// selectorSecondSegKind 是 C2.3 的判据：第二段全数字 → 企业编号；否则 → 作业名。
//
// WP3 起它**委托**给 `internal/target` 的生产实现（`SecondSegmentKind`）：
// 判据只有一处，测试验的就是那一处（不再是"测试自己抄一份、再验自己"）。
func selectorSecondSegKind(seg string) string { return target.SecondSegmentKind(seg) }

// invalidFormKeys 返回表里现有的 F 键（失败信息用）。
func invalidFormKeys() []string {
	out := make([]string, 0, len(debugContractInvalidForms))
	for _, f := range debugContractInvalidForms {
		out = append(out, f.Key)
	}
	return out
}

// stateLabelOf 是 B30 / C5.9 的五个会话状态取值与它们的**中文标签**的表。
//
// 它是"值只定义一次"这条纪律在 T2.2 那一侧的落点：`internal/cli/debug` 的
// `dbgStateLabel`（debugenv.go:351）是这五个标签的**生产实现**，而 C5.9 的身份行
// `会话 <display> · <name> · <state> · ENT <ent>` 明确要求 `<state>` 用它。
//
// **一处已知的重复，如实写明**：`internal/cli/debug/currentsession_char_test.go` 的
// `TestCharDBGStateLabels` 里也有一份同样的五对。这不是偷懒 —— Go 的包**看不见**
// 另一个包 `_test.go` 里声明的符号（实测：`dbgcontract.StateLabelOf` 编译不过），
// 而 WP2 的纪律是只许新增 `_test.go`，不能把这五个标签挪进生产文件。
// 两边的成对断言（这条表 + 那条测试）就是防漂移的判据：任何一边改了标签，
// 另一边那条测试立刻红。
func stateLabelOf() map[string]string {
	return map[string]string{
		"idle":    "空闲",
		"loading": "连接中…",
		"stopped": "已停站",
		"running": "运行中",
		"exit":    "已断开",
	}
}

// newTestHTTPServer 起一个测试 HTTP 服务，挂上调试面自己的路由表。
// 复用 mode_test.go 的 doReq / respError。
func newTestHTTPServer(t *testing.T, h http.Handler) *httptest.Server {
	t.Helper()
	hs := httptest.NewServer(h)
	t.Cleanup(hs.Close)
	return hs
}

// ---------- "落地了没有"的三个判据（全是机械的，不是凭印象） ----------

// contractRefKeyLanded 问："SessionBrief 上已经有 `ref` 键了吗"（C5.3 的五个新键）。
func contractRefKeyLanded() bool {
	rt := reflect.TypeOf(SessionBrief{})
	for i := 0; i < rt.NumField(); i++ {
		if strings.Split(rt.Field(i).Tag.Get("json"), ",")[0] == "ref" {
			return true
		}
	}
	return false
}

// failureEnvelopeHasCode 问："**HTTP 侧**的失败信封已经有 `code` 键了吗"（契约 C5.15 的反面）。
//
// 判据是**实测一个真实的失败响应**（复用 actionTestServer 与 doReq），
// 不是"源码里出现过 code 这个词"：这条路由今天回的是 `{"ok":false,"error":…}` 两键（B8），
// 而 C5.15 要求它增加 `code`（必给）与 `candidates`（按 C5.11）。
//
// 它与 `TestContractHTTPFailureCodeIsNotYetLanded` 问的是同一件事，两处都保留是有意的：
// 那一条是"C5.15 本身的账"，这一条是"F1–F20 端到端还差什么"的守卫。
func failureEnvelopeHasCode(t *testing.T) bool {
	t.Helper()
	hs := actionTestServer(t)
	resp := doReq(t, "POST", hs.URL+"/api/wstest", "", "{}")
	var raw map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&raw)
	_, ok := raw["code"]
	return ok
}

// ---------------------------------------------------------------------------
// T2.4 不测什么（WP2 的第三类交付；计划 Q13 / WP2 一节）
// ---------------------------------------------------------------------------
//
// 这一节**不是注释装饰**：半年后会有人以为"这些测过了"。逐条写清为什么测不了，
// 以及替代它的是什么。
//
// 1. `internal/debug/session.go` 的主流程 —— **不测**。要真 SSH PTY + 真 fgldb 交互：
//    那里 19 处 `time.Now()`、多处硬编码 `time.Sleep`，断言只能落在"最终态"，
//    而最终态要真引擎。理由与现状见 `internal/debug/README.md` 的「故意不覆盖」一节
//    （原文：要测得先造一个 fgldb 协议模拟器，那是**一次依赖反转**；而买家是一个
//    已被 87 个用例与真机自检覆盖的包的内部行覆盖）。计划 Q13 的定稿就是
//    「不硬测，老实写下不测什么」。替代：WP6 的"双 Agent 并发实测"（人工跑一遍）。
//
// 2. `api.go` 的**快乐路径** —— 不测。要 `*host.SSHConn`，它的 `cli` 是私有字段
//    （`host/ssh.go:18`），从外部伪造得实现 8 个方法、其中三个返回不可构造的类型。
//    测试只注入**必然失败**的 dial，买到的是错误半边（见 `api_action_test.go` 的文件头）。
//    所以本文件里凡是要走到"拨号成功之后"的断言，一律**不能**在这里写。
//
// 3. 服务端五处 `mgr.Current()`（B11 / C7.1）**今天只能钉住它自己的语义**，
//    钉不住"每一处各自干了什么"：
//    - `sessOfWriteGlobal`（api.go:367）：能测到的只有"池空 → 返回 nil → 放行"
//      与"池里有空宿主 → 放行"这两个分支（后者要真会话，见第 2 条）；
//    - `hSessionSwitch`（api.go:816）：能测到"池里有**同环境**会话 → 幂等返回那一条"，
//      因为它在 `Current() != nil && sameEnvAs` 时**直接 return**，不碰 pty/conn；
//      剩下两条分支（不同环境 → Close 旧会话 + 新建；池空 → 新建）都会打到
//      `Close()`/`CreateSessionOn()`，而手工造的 Session 没有 pty/conn，`forceExit()`
//      会在 `s.pty.Close()` 上打到 nil（`session_reuse_test.go:73-75` 已记过这个坑）；
//    - `hWSLogDebug`（api.go:1679）：必须先拨 SSH 查 `wsfa_t` 拿到 item 才走到那一步，
//      见第 2 条；
//    - `wslogTopent`（api.go:1724）：同上（它由 `hWSLogDebug` 调用）；
//    - `hDBSQL`（api.go:1754）：拨号失败发生在使用 ent 之前，拿不到可分辨的行为。
//    替代判据：`internal/testkit` 的静态检查（每一个 `mgr.Current()` 都记在案上，
//    WP4 删干净才转绿）+ `Manager.Current()` 自身的表征测试（T2.2）。
//
// 4. `prepareSession`（manager.go:242）的**"换环境 → 收掉旧会话再新建"那一条分支** —— 不测。
//    它要 `live.Close()`，同上会打到 nil 的 pty/conn。可测的两条（同目标空闲复用、
//    同目标忙碌报错）已在 `session_reuse_test.go` 与本次新增的表征测试里钉住。
//
// 5. 真机相关的一切（真 SSH / 真库 / 真 fgldb）**不在契约测试范围**（契约第 10 节末段）。
//    技能在 `.agents/skills/tt-test`，它规定跑真机测试必须先取用户授权、且环境配置
//    不得进仓库 —— 所以它不可能是一条 `go test` 用例。
