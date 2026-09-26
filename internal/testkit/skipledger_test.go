package testkit

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// 本文件是**跳过台账**：让"多一处跳过"变成一次必须改文件、因此会被 review 的动作。
//
// 为什么值得：跳过的测试与通过的测试在报告里长得一样无害。而这个仓库里
// `go test ./...` 默认档的覆盖面，正是靠"到底跳了多少、跳的是哪一类"来定义的 ——
// 一处新的 t.Skip 悄悄加进来，默认档就少验一块，而没有任何东西会响。
//
// 台账登记每一处 t.Skip / t.Skipf 的**位置**（文件 + 文案）与**它属于哪一层**。
// 多一处、少一处、文案改了 —— 都红。
//
// 两条设计取舍，都是刻意的：
//
//  1. **只记文件 + 文案，不记行号**。记行号的话，随便在文件前面加一行注释就会让
//     整份台账失效；而行号只在报错时现算。反过来，改**文案**会让它红 —— 那是对的：
//     文案是"这条跳过在说什么"，改它本就该被看见。
//  2. **扫描范围是 `internal/` 下所有 `.go`，不只是 `_test.go`**。这不是过度设计：
//     `internal/testkit/corpus.go` 是**普通 .go 文件**，里面有 2 处 t.Skipf
//     （语料发现的两条收场），而它是全仓共用的 —— 只扫 `_test.go` 会把这两处漏掉。
//     只看 `_test.go` 的数字也会因此偏小（36 而不是 38）。

// skipLayer 是这条跳过属于哪一层。取值见【二】的分层定义。
type skipLayer string

const (
	// layL3 真语料回归：语料 / 引擎产物 / 工作区不在这台机器上。
	layL3 skipLayer = "L3"
	// layL4 真机 E2E：真引擎 exe / 真 SSH。
	layL4 skipLayer = "L4"
	// layL3Missing 语料在，但里面没有**这一件**（没有 .tzs / 没有那个叶子 / 没有包……）。
	//
	// 与 layL3 分开是因为它**没有开关可点名** —— 补救是"换一份带这个的语料"，
	// 不是"设一个环境变量"。混在一起的话，"why 必须点名开关"那条纪律会把
	// 一堆本来就说得清清楚楚的条目判成不合格。
	layL3Missing skipLayer = "L3-缺件"
	// layEnv 本机环境：没有 git / 没有用户配置目录 / 没有旧配置。
	// 这些**不影响"全绿"这个结论** —— 它们说的是"这台机器上没那个东西"。
	layEnv skipLayer = "L0-环境"
	// layNoSubject 这条断言这一次没有对象（比如文档里那句话被改写了）。
	layNoSubject skipLayer = "无对象"
	// layPremise 测试自己构造的前提没满足。归这一类的要特别当心：
	// 前提长期不成立 = 这条测试**静默变成空转**，与删掉它没有区别。
	layPremise skipLayer = "前提"
)

// switchNames 是"跳过时该设哪个变量"的候选名。
//
// **只列会出现在跳过文案里的那些**，不是全仓开关总表：`TTZS_INSTALL`（设计器目录）
// 与 `TTZS_VALIDATE`（validate 的范围）都是真的开关，但没有一处跳过以它们为门槛，
// 列进来只会让"表里没有死名字"那条测试误报。
//
// 另注意：扫描器只抓 `t.Skip/Skipf` 的**第一个**字符串字面量，而有的跳过文案是
// 拼接出来的（如 e2e_test.go 那条，`TTZS_INSTALL` 在后半段未被捕获）。
// 所以这张表要按**捕获得到**的名字列。
var switchNames = []string{
	"TDEV_DEEP", "TTZS_DEEP", "TTZS_E2E",
	"TTZS_CORPUS", "TDEV_CORPUS",
	"TTZS_EXE", "TTZS_WS", "TTZS_PKG",
}

// skipSite 是台账里的一处跳过。
type skipSite struct {
	file  string // 仓库相对路径（斜杠分隔）
	msg   string // t.Skip/Skipf 的第一个字符串字面量（**照抄**，标点都不能改）
	layer skipLayer
	why   string // 为什么它不算假绿 / 属于哪一层
}

// skipLedger 是**全部**跳过点的台账，逐条按 文件+文案 与代码比对。
//
// 条目是脚本从源码里抽出来的（避免手抄错），层与理由是人逐条标的。
var skipLedger = []skipSite{
	{file: "internal/config/paths_test.go", msg: "拿不到用户配置目录", layer: layEnv, why: "本机取不到用户配置目录 —— 这台机器上没有那个目录"},
	{file: "internal/config/realdata_test.go", msg: "定位不到统一用户目录", layer: layEnv, why: "本机取不到统一用户目录，那份真实旧配置无从谈起"},
	{file: "internal/config/realdata_test.go", msg: "本机没有合并前的旧配置，跳过", layer: layEnv, why: "本机没有合并前的旧配置 —— 只有本机真有旧配置时才验得到"},
	{file: "internal/dev/cli/corpus_test.go", msg: "深度语料回归未启用：设 TDEV_DEEP=1 并给足 -timeout 30m（见 README「测试与验收」）", layer: layL3, why: "需真实语料：开关 TDEV_DEEP=1"},
	{file: "internal/dev/cli/tzs_verb_e2e_test.go", msg: "需要真引擎：设 TTZS_E2E=1（可选 TTZS_EXE），见本文件顶部注释", layer: layL4, why: "需真引擎：开关 TTZS_E2E=1（可选 TTZS_EXE）"},
	{file: "internal/dev/cli/tzs_verb_e2e_test.go", msg: "找不到引擎 exe（TTZS_EXE=%s）：%v", layer: layL4, why: "需真引擎：开关 TTZS_E2E=1 + TTZS_EXE"},
	{file: "internal/dev/cli/tzs_verb_test.go", msg: "文档里没有「N 个动词」的声明 —— 这条断言没有对象了（改写了措辞就把它一起改）", layer: layNoSubject, why: "这条断言的对象是文档里那句话，措辞一改它就没有对象了"},
	{file: "internal/dev/cli/tzs_verb_test.go", msg: "没有引擎 exe（%v）：这条测的是期限机制本身", layer: layL4, why: "需真引擎：开关 TTZS_E2E=1（这条测的是期限机制本身）"},
	{file: "internal/dev/pkgfile/zipshape_test.go", msg: "语料里没找到设计器形态的 .tap 包", layer: layL3Missing, why: "需真实语料：这一类形态只在真语料里出现"},
	{file: "internal/dev/pkgfile/zipshape_test.go", msg: "语料里没有 Zip64 形态的 .tap 包", layer: layL3Missing, why: "需真实语料：这一类形态只在真语料里出现"},
	{file: "internal/dev/store/store_test.go", msg: "环境里没有 git，跳过 git 相关断言", layer: layEnv, why: "本机没装 git（装了才跑 git 相关断言）"},
	{file: "internal/dev/tapfile/tapfile_test.go", msg: "%s 下没有 .tap", layer: layL3Missing, why: "语料在，但里面没有 .tap"},
	{file: "internal/dev/tzs/chain_test.go", msg: "语料里没有 %s —— 这条链要一个真实的数据绑定叶子；换个语料就得同时改 chainQuery", layer: layL3Missing, why: "需真实语料里那个数据绑定叶子；换语料要同时改 chainQuery"},
	{file: "internal/dev/tzs/chain_test.go", msg: "%s 的祖先目录里没有 mta/，且没给 TTZS_WS —— 工作区绝不替你选", layer: layL3, why: "需真工作区：开关 TTZS_WS（工作区绝不替你选）"},
	{file: "internal/dev/tzs/chain_test.go", msg: "%s 里没有 %s（find_component 回 matchCount=%d）—— 这条链要一个已知存在的代号", layer: layL3Missing, why: "需真实语料里那个已知存在的代号"},
	{file: "internal/dev/tzs/chain_test.go", msg: "%s 的 %s 没有 posX 这个布局属性（可写属性：%s）", layer: layL3Missing, why: "需真实语料里带 posX 布局属性的叶子"},
	{file: "internal/dev/tzs/chain_test.go", msg: "%s 的 posX=%q 不是整数，无法「+1」出一个必然不同的值", layer: layL3Missing, why: "需真实语料里 posX 是整数的叶子（要 +1 出一个必然不同的值）"},
	{file: "internal/dev/tzs/corpus_test.go", msg: "语料深度回归未启用：设 TTZS_DEEP=1 并给足 -timeout 30m（见 corpus_test.go 顶部注释）", layer: layL3, why: "需真实语料：开关 TTZS_DEEP=1"},
	{file: "internal/dev/tzs/corpus_test.go", msg: "找不到引擎 exe（TTZS_EXE=%s）：%v；先在 engine/ 里跑 build.sh", layer: layL4, why: "需真引擎：开关 TTZS_E2E=1 + TTZS_EXE（先跑 engine/build.sh）"},
	{file: "internal/dev/tzs/corpus_test.go", msg: "%s 旁边没有 RoundTrip.exe：%v", layer: layL4, why: "需真引擎：TTZS_EXE 旁边的 RoundTrip.exe（engine/build.sh 产出）"},
	{file: "internal/dev/tzs/corpus_test.go", msg: "没有真实语料（%s / %s 都没指到一个目录）：设 TTZS_CORPUS，或准备 %s", layer: layL3, why: "需真实语料：开关 TTZS_CORPUS"},
	{file: "internal/dev/tzs/corpus_test.go", msg: "%s 下没有 .tzs（排除掉 %s 之后）", layer: layL3Missing, why: "语料在，但里面没有 .tzs"},
	{file: "internal/dev/tzs/corpus_test.go", msg: "没有真实语料：设 TTZS_CORPUS（或 TDEV_CORPUS），或准备 %s", layer: layL3, why: "需真实语料：开关 TTZS_CORPUS（或 TDEV_CORPUS）"},
	{file: "internal/dev/tzs/corpus_test.go", msg: "%s 下没有 .tzs", layer: layL3Missing, why: "语料在，但里面没有 .tzs"},
	{file: "internal/dev/tzs/corpus_test.go", msg: "pin 里 %d 条路径在本机一条都不存在（语料根=%s）：pin 描述的是另一份语料，", layer: layL3, why: "pin 描述的是另一份语料：用 TTZS_CORPUS 指到正确那份"},
	{file: "internal/dev/tzs/corpus_test.go", msg: "语料里没有包", layer: layL3Missing, why: "语料在，但里面没有包"},
	{file: "internal/dev/tzs/corpus_test.go", msg: "没有一个工作区能验（都缺 mta/mod-fd.spec，或包打不开）", layer: layL3, why: "需真工作区：开关 TTZS_WS"},
	{file: "internal/dev/tzs/corpus_test.go", msg: "没验到（语料里没有推得出工作区的包）", layer: layL3Missing, why: "需真实语料里能推出工作区的包"},
	{file: "internal/dev/tzs/dryrun_test.go", msg: "没验到（语料里挑不出可写的目标）", layer: layL3Missing, why: "需真实语料里可写的目标"},
	{file: "internal/dev/tzs/e2e_test.go", msg: "需要真引擎：设 TTZS_E2E=1 + 引擎 exe + 工作区（TTZS_EXE / TTZS_WS，", layer: layL4, why: "需真引擎：开关 TTZS_E2E=1 + TTZS_EXE + TTZS_WS"},
	{file: "internal/dev/tzs/e2e_test.go", msg: "找不到引擎 exe（%s）：%v", layer: layL4, why: "需真引擎：开关 TTZS_E2E=1 + TTZS_EXE"},
	{file: "internal/dev/tzs/e2e_test.go", msg: "没给工作区（TTZS_WS 或 config.local.json 的 workspace）。", layer: layL4, why: "需真工作区：开关 TTZS_WS（工作区绝不替你选）"},
	{file: "internal/dev/tzs/e2e_test.go", msg: "没给 TTZS_PKG（一个本工作区内的真 .tzs 路径）：逻辑键寻址要用真包验", layer: layL4, why: "需真包：开关 TTZS_PKG（逻辑键寻址要用真包验）"},
	{file: "internal/dev/tzs/e2e_test.go", msg: "没给 TTZS_PKG（本工作区内的真 .tzs 路径）：任务级动词要用真包验", layer: layL4, why: "需真包：开关 TTZS_PKG（任务级动词要用真包验）"},
	{file: "internal/dev/tzs/oplog_test.go", msg: "工作区 %s 里没有包，这条用例没法跑", layer: layL3, why: "需真工作区且里面有包：开关 TTZS_WS"},
	{file: "internal/dev/verify/verify_test.go", msg: "fence.Parse 自己就拦住了（%v）—— 这条测的是 gate1，不是解析层", layer: layPremise, why: "解析层先拦住了（那更好）—— 这条测试的对象就不存在了，如实说明而不是假装通过"},
	{file: "internal/testkit/corpus.go", msg: "没有真实语料：设 TDEV_CORPUS 或 TTZS_CORPUS 指到一份，或准备 %s", layer: layL3, why: "需真实语料：开关 TDEV_CORPUS / TTZS_CORPUS"},
	{file: "internal/testkit/corpus.go", msg: "%s 下没有 %s", layer: layL3Missing, why: "语料在，但里面没有该扩展名的文件"},
}

// reSkipCall 匹配 t.Skip / t.Skipf 且第一个参数是同行的字符串字面量。
//
// 全仓当前 38 处都长这样（实测：`t.Skip` 出现 38 次，其中 38 次同行带字面量）。
// 将来若有人写成跨行或拼接，这里会匹配不到 → 台账对不上 → 红。
// 那时该改的是**扫描规则**，不是把台账删掉。
var reSkipCall = regexp.MustCompile(`t\.Skipf?\(\s*"([^"]*)"`)

// scanSkipSites 走 internal/ 下所有 .go，按**文件名字典序 + 文件内出现顺序**产出跳过点。
//
// 顺序不重要（比对按 文件+文案 做键），但必须**确定**：同一份代码扫两次要给同一个结果，
// 否则报错信息会时有时无。
func scanSkipSites(t *testing.T) []skipSite {
	t.Helper()
	root := RepoRoot(t)
	base := filepath.Join(root, "internal")
	var out []skipSite
	err := filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			rel = p
		}
		rel = filepath.ToSlash(rel)
		for _, m := range reSkipCall.FindAllStringSubmatch(string(b), -1) {
			out = append(out, skipSite{file: rel, msg: m[1]})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("扫 %s 失败：%v", base, err)
	}
	return out
}

// ledgerKey 是台账比对的键。用 \x00 分隔两个字段，避免"文件结尾恰好连着文案"这种巧合。
func ledgerKey(file, msg string) string { return file + "\x00" + msg }

// TestSkipLedger 台账与代码必须**严格对上**：多一处、少一处、文案改了都红。
//
// 它买到三件事：
//   - "静默跳过"变成"必须改台账"，而改台账是一个会被 review 的 diff
//   - 台账就是分层归属的唯一来源：L3 几条 = 开 TDEV_DEEP 该多跑多少条；
//     L4 几条 = 开 TTZS_E2E 该多跑多少条。"这次到底跳了多少"不再靠数
//   - L3/L4 的 why 必须点名开关名 → **人看到 skip 就知道怎么打开**
func TestSkipLedger(t *testing.T) {
	found := scanSkipSites(t)

	// 防空转：扫描坏了的话（路径变了、正则被改坏），下面所有比对都会"通过"。
	if len(found) < 20 {
		t.Fatalf("只扫到 %d 处 t.Skip —— 扫描本身坏了，此后的比对等于空转", len(found))
	}

	want := make(map[string]skipSite, len(skipLedger))
	for _, s := range skipLedger {
		k := ledgerKey(s.file, s.msg)
		if _, dup := want[k]; dup {
			t.Errorf("台账里 %s 的这条跳过登记了两次：%q", s.file, s.msg)
		}
		want[k] = s
	}

	seen := make(map[string]bool, len(found))
	var unregistered []skipSite
	for _, s := range found {
		k := ledgerKey(s.file, s.msg)
		seen[k] = true
		if _, ok := want[k]; !ok {
			unregistered = append(unregistered, s)
		}
	}
	if len(unregistered) > 0 {
		var b strings.Builder
		for _, s := range unregistered {
			b.WriteString("\n  " + s.file + "\n      " + s.msg)
		}
		t.Errorf("这些跳过没登记：%s\n\n"+
			"新增跳过必须登记进 skipLedger —— 这不是形式主义：\n"+
			"  · 属于**默认档**（L1/L2）的，它应当是断言而不是跳过，请改。\n"+
			"  · 属于 L3/L4 的，登记并注明**开关名**（TDEV_DEEP / TTZS_DEEP / TTZS_E2E /\n"+
			"    TTZS_CORPUS …）—— 人看到 skip 要能知道怎么打开。\n"+
			"  · 属于「L0-环境」的（无 git 之类），写清为什么它不影响「全绿」这个结论。\n"+
			"  · 属于「前提」的要特别当心：前提长期不成立 = 这条测试静默变成空转。\n\n"+
			"台账在 internal/testkit/skipledger_test.go。", b.String())
	}

	// 反向：台账里登记了、代码里却没有 —— 删了跳过却没同步台账。
	// 少了这条，台账会慢慢变成一份没人信的历史。
	var stale []skipSite
	for _, s := range skipLedger {
		if !seen[ledgerKey(s.file, s.msg)] {
			stale = append(stale, s)
		}
	}
	if len(stale) > 0 {
		var b strings.Builder
		for _, s := range stale {
			b.WriteString("\n  " + s.file + "\n      " + s.msg)
		}
		t.Errorf("台账里这些跳过，代码里已经没有了：%s\n\n"+
			"（删掉跳过、或改动了它的文案，都要同步台账。文案改动也算 ——\n"+
			"  文案是「这条跳过在说什么」，改它本就该被看见。）", b.String())
	}
}

// TestSkipLedgerL3L4NamesItsSwitch L3/L4 的台账条目，`why` 里必须点名一个开关。
//
// **这条测的是我写的 why，不是代码** —— 它拦不住"why 写得对但跳过本身是错的"。
// 它挡的是另一件常见的事：加一处跳过、why 随手写一句"环境不对"，于是没人知道
// 该设哪个变量才能跑起来。所以它是纪律断言，不是判据。
func TestSkipLedgerL3L4NamesItsSwitch(t *testing.T) {
	for _, s := range skipLedger {
		// layL3Missing 不在此列：那一类没有开关可点名（见它的注释）。
		if s.layer != layL3 && s.layer != layL4 {
			continue
		}
		named := false
		for _, n := range switchNames {
			if strings.Contains(s.why, n) {
				named = true
				break
			}
		}
		if !named {
			t.Errorf("%s（%s）的 why 没点名开关：%q\n"+
				"    改成「需…：开关 XXX=1」这样 —— 人看到 skip 要能知道怎么打开。\n"+
				"    跳过的是：%q", s.file, s.layer, s.why, s.msg)
		}
	}
}

// TestSkipLedgerHasNoStaleSwitchNames 反向：switchNames 里的每个名字，
// 至少要有一处跳过在用它 —— 否则它多半已经改名了，该从表里删掉。
func TestSkipLedgerHasNoStaleSwitchNames(t *testing.T) {
	joined := ""
	for _, s := range skipLedger {
		joined += s.why + "\n" + s.msg + "\n"
	}
	for _, n := range switchNames {
		if !strings.Contains(joined, n) {
			t.Errorf("开关名 %q 没有任何一处跳过在用它 —— 改名了？从 switchNames 里删掉", n)
		}
	}
}
