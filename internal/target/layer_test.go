package target

// 层级校验（Q7 / Q8 / 契约第 7 节、以及 `tt-target-levels.md` 的组默认 + 例外）。
//
// 两条纪律在这条测试里变成判据：
//   - Q7：**L4 不作寻址层** —— 认路只做 L0–L3（`tt debug` 的组默认 L4 夹到 L3）；
//   - Q8：登记表只有"组默认 + 具名例外"，**清单从命令树派生**，不许写成 180 条全清单。

import (
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestLayerRegistryIsGroupDefaultsPlusNamedExceptions 钉住 Q8 的形状：
// 组默认 8 条（`tt-target-levels.md` 的总览那一列）+ 具名例外 40 条。
//
// **40 不是矩阵「例外条数」一列的和（50）**：矩阵那份是逐条判据（`debug` 那组的"基准 L4"
// 是状态语义），登记表这份只留**改变寻址要求**的条目。`debug` 的组默认按 Q7 定成 L3
// （寻址层），于是矩阵里 12 条 `debug xxx: L3` 与组默认同值、改变不了任何行为，
// 按 Q8「与组默认相同的命令一条都不登」全部摘掉：50 − 12 = 38（`debug` 19 → 7）。
// `fep` 组本轮改判（WP4c①）：组默认 L1 → **L3**（Q9），于是"要界面活着"的 10 条
// 与组默认同值、一条都不登，而 `rec replay`（要活界面 ⇒ L3）也从例外表里退出，
// `rec reproduce` / `rec start` 收口为 L1、`fe status` / `fe send` 判为 L0 ——
// 净变化 38 + 2 = 40（`fep` 21 → 23）。
func TestLayerRegistryIsGroupDefaultsPlusNamedExceptions(t *testing.T) {
	r := DefaultLayerRegistry()

	// 组默认：debug / dict / dev tzc / dev tzs / fep / view / drawio / session / recording（+ 顶层）
	// （顶层 / debug / dict / dev tzc / dev tzs / fep / view / drawio）。
	const wantGroups = 10
	if len(r.Groups) != wantGroups {
		t.Errorf("组默认该有 %d 条（矩阵总览那一列），实得 %d：%v", wantGroups, len(r.Groups), groupKeys(r))
	}

	// 具名例外：只登"真的改变寻址要求"的那些（Q8）；与组默认同值的一条不登。
	// 新增 session / recording 两组后，两组的 L0/L1 例外已按 Q8 归位：
	// recording 组默认 L0（起录/回放抬到 L1），session 组默认 L3（session ls 降到 L0）。
	const wantExceptions = 32
	if len(r.Exceptions) != wantExceptions {
		var pending int
		for _, l := range r.Exceptions {
			if l == LayerUndetermined {
				pending++
			}
		}
		t.Errorf("具名例外该有 %d 条（debug 7 / dict 5 / dev tzs 3 / view 2 / fep 23；"+
			"debug 的 12 条 L3 与组默认同值、按 Q8 不登），实得 %d（其中待定 %d）",
			wantExceptions, len(r.Exceptions), pending)
	}

	// 登记表**不是**命令清单：8 + 40 = 48 条，远小于矩阵的 186 行明细。
	if n := len(r.Groups) + len(r.Exceptions); n >= 186 {
		t.Errorf("登记表有 %d 条 —— Q8 要的是『组默认 + 具名例外』，不是全清单（矩阵明细 186 行）", n)
	}

	// 每一条例外都必须**真的是例外**：与所在组的组默认不同（Q8）。
	// 判据抽成 exceptionEqualsGroupDefault，好让它自己也被测一条（见
	// TestExceptionDisciplineCatchesEqualToGroupDefault）—— 本轮摘掉 12 条 debug L3
	// 就是这条判据"审"出来的，它得是**真的会红**的那种绿。
	for _, path := range sortedKeys(r.Exceptions) {
		if why := exceptionEqualsGroupDefault(r, path); why != "" {
			t.Errorf("%s", why)
		}
	}

	// 组默认自己不能被登记成例外（那会让"组默认"这件事有两个说法）。
	for _, g := range groupKeys(r) {
		if g == "" {
			continue
		}
		if _, ok := r.Exceptions[g]; ok {
			t.Errorf("命令组 %q 被登记成了具名例外 —— 组默认与例外是两件事", g)
		}
	}

	// 交出来的 map 是副本：调用方改了不会污染包级那份（测试要试别的登记形状就得靠这条）。
	before := len(layerExceptions)
	r.Exceptions["debug exec"] = Layer3
	r.Groups["debug"] = Layer1
	if len(layerExceptions) != before {
		t.Errorf("改了 DefaultLayerRegistry() 交出的例外表，包级 layerExceptions 跟着变了（%d → %d）—— 该交副本",
			before, len(layerExceptions))
	}
	r2 := DefaultLayerRegistry()
	if r2.Groups["debug"] != Layer3 || len(r2.Exceptions) != wantExceptions {
		t.Errorf("后一次 DefaultLayerRegistry() 被上一次的改动污染了：debug 组默认 = %s，例外 %d 条",
			r2.Groups["debug"], len(r2.Exceptions))
	}
}

// TestLayerRegistryDoesNotEnumerateTheCommandTree 钉住 Q8 的另一半：
// 层级**等于组默认**的那些叶子命令一条都不许登记（清单从命令树派生）。
func TestLayerRegistryDoesNotEnumerateTheCommandTree(t *testing.T) {
	r := DefaultLayerRegistry()
	for _, path := range []string{
		"dict r.t", "dict prog", "dict spill list", // dict 组默认 L0
		"dev tzc export", "dev tzc apply", "dev tzc verify", // tzc 全 L0
		"dev tzs open", "dev tzs field_add", "dev tzs save", // tzs 组默认 L0
		"view ls", "view shot", "view marks", // view 组默认 L0
		"drawio lib", "drawio compose", // drawio 无例外
		"config show", "version", "cache clear", // 顶层 L0
		"fep do type", "fep do key", "fep wait", "fep job run", // fep 组默认 L3（Q9 ⇒ 要界面活着）
		// 本轮 fep 改判后新落进"组默认"的那几条：`rec replay` 要一条**活着的界面**
		// （非 --dry-run 时 `newClient()` + 指纹闸比当前界面 + 注入），因此与组默认同值，
		// 一条都不许再登记（Q8）。do * 的另一半也一并钉住。
		"fep do act", "fep do cell", "fep do lookup", "fep do pick", "fep do role",
		"debug locate", "debug why", "debug wait", // debug 组默认 L3（Q7 的寻址层）
		// 本轮摘掉的 12 条：组默认已按 Q7 定成 L3，它们**一条都不许**再出现在例外表里。
		"debug start", "debug wsdebug", "debug stop", "debug exec",
		"debug interrupt", "debug quit", "debug mode", "debug status",
		"debug env", "debug topent", "debug serve", "debug logs",
	} {
		if _, ok := r.Exceptions[path]; ok {
			t.Errorf("%q 被登记进了例外表，但它的层级就是所在组的组默认 ——\n"+
				"    Q8：登记表只有『组默认 + 具名例外』，清单从命令树派生。", path)
		}
	}
}

// TestAddressingLayer 钉住 Q7：认路只做 L0–L3，待定的照旧待定。
//
// 登记表的 `debug` 组默认**本身就是 L3**（Q7：L4 不作寻址层），所以 L4 夹到 L3 这条
// 现在只是"万一有人把 L4 填回来"的兜底；`debug` 那 12 条 L3 例外摘掉之后认路结果一条不变。
func TestAddressingLayer(t *testing.T) {
	r := DefaultLayerRegistry()
	cases := []struct {
		path string
		want Layer
		why  string
	}{
		// ---- L3：要界面活着（靠 `fep` 组默认，不登记）----
		{"fep do act", Layer3, "点一个动作对象：`registry().Load()` 无会话即退 3，事件发给活界面"},
		{"fep do cell", Layer3, "往网格当前行的格子写值（先发定位事件再写值，都要活界面）"},
		{"fep do key", Layer3, "往活着的界面按一个键"},
		{"fep do lookup", Layer3, "开窗取数：`newClient()` + `Status()` 要活连接，且每步现读运行期"},
		{"fep do pick", Layer3, "在下拉里选一项（对运行期发 ActionEvent）"},
		{"fep do role", Layer3, "按角色操作控件：`registry().Load()` 无会话即退 3"},
		{"fep do type", Layer3, "往栏位输入值（校验在本地，事件仍要发给活界面）"},
		{"fep wait", Layer3, "轮询**运行中**的作业是否变了样子（`newClient()` 无会话即退 3）"},
		{"fep job run", Layer3, "每步现读运行期状态 → 判可用 → 发 → 验效果（`newClient()`）"},
		// ---- 会话级动作：要一条活着的会话（`session` 组默认 L3）----
		{"session ent", Layer3, "改会话的企业编号：`SetTopent` 要往那条会话的 shell 下发"},
		{"session quit", Layer3, "结束本轮调试：`EndRun` 打在那条会话上"},
		{"session rm", Layer3, "收掉会话：断连接、清资源"},
		{"session exec", Layer3, "透传 fgldb 命令：没有会话就没有可发的上下文"},
		{"session show", Layer3, "看会话详情：身份与现场都在那条会话上"},
		// ---- L0：纯本地（录制件 / 本地登记 / 本机进程），`recording` 组默认 ----
		{"recording ls", Layer0, "列缓存 + 长期库两处**本地**录制件（当场复算指纹）"},
		{"recording show", Layer0, "读本地录制件；有会话只是顺手比一次闸门，没有也照答"},
		{"recording index", Layer0, "给本地录制件建派生索引（raw 不动）"},
		{"recording keep", Layer0, "把本地录制件搬进本地长期库"},
		{"recording rm", Layer0, "删本地长期库里的一份录制"},
		{"recording prune", Layer0, "拿库里的环境目录名与本地配置对一遍，清多出来的"},
		{"recording active", Layer0, "读本地登记 + 判本机 GDC pid 是否还活着"},
		{"recording stop", Layer0, "只杀登记里的本机 pid（对调试会话转调 `tt debug quit` 也仍是本地动作）"},
		{"session ls", Layer0, "列会话：只要服务在，不要求已有会话"},
		{"fep ui snapshot", Layer0, "读录制件的字节汇总界面结构（离线）"},
		{"fep ui find", Layer0, "读录制件的字节折 AUI 再筛（离线）"},
		{"fep ui tree", Layer0, "读录制件里的字节打窗口子树（离线）"},
		{"fep ui roles", Layer0, "读录制件的字节折角色目录（--source 才顺带读一次本地登记，读不到也照答）"},
		{"fep ui state", Layer0, "读录制件里的运行期状态推送（离线）"},
		{"fep ui focus", Layer0, "读录制件里的 focus/cursor 推送（离线）"},
		{"fep src openwin", Layer0, "读本地源码镜像里开窗程序的交互语义"},
		{"fep job plan", Layer0, "从源码镜像推导操作计划，不碰运行期"},
		{"fep fe listen", Layer0, "起本机无头前端、在本机端口上**等**运行期连入"},
		{"fep fe status", Layer0, "拨本机控制口（`fep fe listen` 起的那个本机进程）"},
		{"fep fe send", Layer0, "经本机控制口发一条事件（同上）"},
		// ---- L1：只要一个环境，会话由它自己起（Q14 的"起会话"）----
		{"recording start", Layer1, "经 SSH 登录点名环境（`--env` 必给），起隔离实例 + 代理 + 作业"},
		{"recording replay", Layer1, "默认从零起自己的会话（环境/作业取自录制身份）；--attached 才接别人的"},
		{"session new", Layer1, "建会话：点名的环境就是要连的那台；--prog 还会起作业"},
	}
	for _, c := range cases {
		if got := r.AddressingLayer(c.path); got != c.want {
			t.Errorf("AddressingLayer(%q) = %s，期望 %s（%s）", c.path, got, c.want, c.why)
		}
	}
	// 这条表自己也要"不漏"：条数与下面的 Q8 检查都按它走。
	if len(cases) != 37 {
		t.Errorf("这张表该覆盖 37 条命令，实得 %d 条", len(cases))
	}

	// Q8：与**所在组的组默认**同值的一条都不许登记；不同值的必须在例外表里。
	for _, c := range cases {
		groupDefault := r.Groups[r.GroupOf(c.path)]
		_, registered := r.Exceptions[c.path]
		if c.want == groupDefault {
			if registered {
				t.Errorf("%q 的层级（%s）就是它所在组 %q 的组默认，却登记进了例外表 —— Q8：与组默认相同的一条都不登",
					c.path, c.want, r.GroupOf(c.path))
			}
			continue
		}
		if !registered {
			t.Errorf("%q（%s）与组默认 %s 不同，该登记进例外表却没有", c.path, c.want, groupDefault)
		}
	}
}
// TestProvidedLayer 钉住"提供了多少信息 → 提供到第几层"。
func TestProvidedLayer(t *testing.T) {
	cases := []struct {
		p    Provided
		want Layer
		why  string
	}{
		{Provided{}, Layer0, "什么都没给"},
		{Provided{Env: "恒烁正式区"}, Layer1, "只要环境"},
		{Provided{Env: "恒烁正式区", Ent: "99"}, Layer2, "环境 + 企业编号"},
		{Provided{Env: "恒烁正式区", Ent: "99", HasSession: true}, Layer3, "落到了一条会话"},
		{Provided{HasSession: true}, Layer3, "会话本身就隐含了环境"},
		{Provided{Ent: "99"}, Layer0, "B23：环境一律点名 —— 没有环境，企业编号解决不了任何事"},
	}
	for _, c := range cases {
		if got := c.p.Layer(); got != c.want {
			t.Errorf("Provided%+v.Layer() = %s，期望 %s（%s）", c.p, got, c.want, c.why)
		}
	}
}

// TestCheckLayerMismatch 钉住层级校验本身：
// "这条命令要求 L3 而目标只解析出 L1"要报出来（契约第 7 节 / 矩阵 (b) 的报错文案），
// 而不是让命令带着不够的目标继续跑。
//
// 码用 `USAGE`（**不新增码**：4.1.1 是唯一清单，里面没有"层级不足"这一条）。
func TestCheckLayerMismatch(t *testing.T) {
	r := DefaultLayerRegistry()

	f := CheckLayer(r, "debug exec", Provided{Env: "恒烁正式区"})
	if f == nil {
		t.Fatal("`tt debug exec` 要一条 SSH 会话（L3），只给到环境（L1）该报错")
	}
	if f.Code != CodeUsage || f.Exit != ExitUsage {
		t.Errorf("层级不足的码该是 %s / 退 %d（4.1.1 里没有第二个可用的码），实得 %s / 退 %d",
			CodeUsage, ExitUsage, f.Code, f.Exit)
	}
	for _, frag := range []string{"debug exec", "L3", "L1"} {
		if !strings.Contains(f.Message, frag) {
			t.Errorf("层级不足的文案 = %q，该说清『哪条命令要哪一层、只给到哪一层』（缺 %q）", f.Message, frag)
		}
	}

	// 够用就不报。
	for _, c := range []struct {
		path string
		p    Provided
	}{
		{"debug exec", Provided{Env: "E", HasSession: true}},
		{"dict r.t", Provided{}},                              // L0 什么都不需要
		{"debug sql", Provided{Env: "E", Ent: "99"}},          // L2 够了
		{"fep do type", Provided{Env: "E", HasSession: true}}, // 组默认 L3（Q9）：落到一条会话才够
		{"recording start", Provided{Env: "E"}},                 // fep 的 L1 例外：环境就够，会话它自己起
		{"recording ls", Provided{}},                          // 例外 L0
		{"fep fe send", Provided{}},                           // 例外 L0（本机控制口）
		{"dev tzs list_tables", Provided{}},                   // 待定不猜：不拦
		{"view measure", Provided{}},                          // 待定不猜
	} {
		if f := CheckLayer(r, c.path, c.p); f != nil {
			t.Errorf("CheckLayer(%q, %+v) 该放行，却报 %s：%s", c.path, c.p, f.Code, f.Message)
		}
	}

	// `fep` 组默认本轮从 L1 正名为 L3（WP4c① / Q9）：只给到环境（L1）必须报出来 ——
	// 只给环境就往下发事件，正是"带着不够的目标继续跑"那种静默落错。
	fFep := CheckLayer(r, "fep do type", Provided{Env: "E"})
	if fFep == nil {
		t.Error("`tt fep do type` 要一条活着的会话（L3），只给到环境（L1）该报错（Q9）")
	} else {
		for _, frag := range []string{"fep do type", "L3", "L1"} {
			if !strings.Contains(fFep.Message, frag) {
				t.Errorf("`tt fep do type` 层级不足的文案 = %q，缺 %q", fFep.Message, frag)
			}
		}
	}
	// 离线读录制件那批（L0）与"会话自己起"那批（L1）**不该**被这条改判带上——
	// 它们不是"要界面活着"，报错就是误拦。
	if f := CheckLayer(r, "recording ls", Provided{}); f != nil {
		t.Errorf("`tt recording ls`（L0）什么都不给就该放行，却报 %s：%s", f.Code, f.Message)
	}
	if f := CheckLayer(r, "recording start", Provided{Env: "E"}); f != nil {
		t.Errorf("`tt recording start`（L1，Q14 的起会话）给到环境就该放行，却报 %s：%s", f.Code, f.Message)
	}

	// L2 的命令只给到 L1 也要报（缺的是企业编号）。
	f2 := CheckLayer(r, "debug sql", Provided{Env: "E"})
	if f2 == nil || !strings.Contains(f2.Message, "L2") {
		t.Errorf("`tt debug sql`（L2）只给到环境该报缺 L2，实得 %v", f2)
	} else if !strings.Contains(f2.Hint, "--ent") {
		t.Errorf("缺 L2 时的修复建议该点名 --ent，实得 hint = %q", f2.Hint)
	}
	// L3 的命令缺会话时，修复建议要点名 --session。
	f3 := CheckLayer(r, "debug exec", Provided{})
	if f3 == nil || !strings.Contains(f3.Hint, "--session") {
		t.Errorf("缺 L3 时的修复建议该点名 --session，实得 %v", f3)
	}
}

// TestLayerNames 钉住层级的显示名（报错文案与文档都读它）。
func TestLayerNames(t *testing.T) {
	want := map[Layer]string{
		Layer0: "L0", Layer1: "L1", Layer2: "L2", Layer3: "L3", Layer4: "L4",
		LayerUndetermined: "待定",
	}
	for l, w := range want {
		if got := l.String(); got != w {
			t.Errorf("Layer(%d).String() = %q，期望 %q", int(l), got, w)
		}
	}
	for l, w := range map[Layer]string{
		Layer0: "无远端目标", Layer1: "要一个环境", Layer2: "要环境 + 企业编号",
		Layer3: "要一条 SSH 会话", Layer4: "要一条会话 + 一轮运行",
	} {
		if got := LayerNeed(l); got != w {
			t.Errorf("LayerNeed(%s) = %q，期望 %q", l, got, w)
		}
	}
}

// groupKeys 返回登记表里的组名（排序，失败信息用）。
func groupKeys(r LayerRegistry) []string {
	out := make([]string, 0, len(r.Groups))
	for g := range r.Groups {
		out = append(out, g)
	}
	sort.Strings(out)
	return out
}

// sortedKeys 返回例外表里的命令路径（排序，失败信息用）。
func sortedKeys(m map[string]Layer) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// exceptionEqualsGroupDefault 是 Q8 那条判据："这条例外真的与它所属组的组默认不同吗"。
//
// **两种"同值"都要拦**：
//
//  1. 原始声明值相同（`DeclaredLayer` vs `Groups[g]`）—— 登了就是重复登记；
//  2. **认路值**相同（`AddressingLayer` vs 组默认的认路值）—— 一条都不改变行为的登记。
//     本轮清掉的正是这一种：组默认 L4 夹成 L3，那 12 条 `debug xxx: L3` 在认路时
//     与组默认同值，于是 Q7 把组默认定到 L3 之后它们就是纯冗余（Q8 不许登）。
//
// 返回空串 = 合规（真的不同，或"待定"—— 待定不参与这条判据）；
// 返回非空 = 一条现成的失败文案（调用方直接报出去）。
func exceptionEqualsGroupDefault(r LayerRegistry, path string) string {
	if r.Exceptions[path] == LayerUndetermined {
		return "" // "待定"不参与这条判据（它本来就还没定）
	}
	g := r.GroupOf(path)
	declared := r.DeclaredLayer(path)
	groupDeclared := r.Groups[g]
	// 组默认自己的认路值：L4 的组默认在认路上就是 L3（Q7）。
	groupAddressing := groupDeclared
	if groupAddressing == Layer4 {
		groupAddressing = Layer3
	}
	if declared != groupDeclared && r.AddressingLayer(path) != groupAddressing {
		return ""
	}
	return "例外 " + strconv.Quote(path) + " 的层级（" + declared.String() + "）与它所属组 " +
		strconv.Quote(g) + " 的组默认（" + groupDeclared.String() + "）同值 —— 那不是例外；\n" +
		"    Q8：清单从命令树派生，与组默认相同的命令**不登记**（认路上改变不了任何行为的也不算例外）。"
}

// TestExceptionDisciplineCatchesEqualToGroupDefault 证明上面那条判据**真的会红** —— 三种形状都试：
//
//  1. **改前的形状**（组默认 L4 + 12 条 `debug xxx: L3`）：判据必须点出全部 12 条。
//     这正是上一轮那个问题的形状 —— 它们与**夹过**的组默认（L4 → L3）同值；
//     只比原始声明值的判据会放过它们（L3 ≠ L4），只比认路值的判据又漏掉别的，两种都得看。
//  2. **改后的形状**（组默认 L3）：真实登记表上一条都不报（全绿）；
//     但把 `debug exec` 这类 L3 条目塞回例外表时立刻点名。
//  3. **本轮 fep 换了组默认**（L1 → L3，WP4c①）：那 10 条"要界面活着"的若被谁登记回去，
//     判据同样必须当场红 —— 换默认值**不许**把这条纪律弄松。
//
// 没有这条元测试，"判据还绿着"只说明没人喂它坏数据，说明不了它拦得住。
func TestExceptionDisciplineCatchesEqualToGroupDefault(t *testing.T) {
	// ① 改前的形状：组默认 L4，12 条 `debug xxx: L3` 在例外表里。
	//    （`DefaultLayerRegistry()` 交出来的两份 map 都是副本，随便改。）
	old := DefaultLayerRegistry()
	old.Groups["debug"] = Layer4
	redundant := []string{
		"debug start", "debug wsdebug", "debug stop", "debug exec",
		"debug interrupt", "debug quit", "debug mode", "debug status",
		"debug env", "debug topent", "debug serve", "debug logs",
	}
	for _, p := range redundant {
		old.Exceptions[p] = Layer3
	}
	for _, p := range redundant {
		why := exceptionEqualsGroupDefault(old, p)
		if why == "" {
			t.Errorf("改前的形状下，例外 %q（L3）与组默认 L4 夹出来的认路值（L3）同值 —— 判据却放行；\n"+
				"    这正是本轮要清掉的冗余条目，判据必须点名它", p)
		} else if !strings.Contains(why, p) {
			t.Errorf("判据的失败文案该点名 %q，实得 %q", p, why)
		}
	}

	// ② 改后的形状：真实登记表必须一条都不被点名（那 12 条摘掉之后）。
	r := DefaultLayerRegistry()
	if _, ok := r.Exceptions["debug exec"]; ok {
		t.Fatal("前置条件不成立：`debug exec` 还在例外表里（本轮该已摘掉）")
	}
	for _, p := range sortedKeys(r.Exceptions) {
		if why := exceptionEqualsGroupDefault(r, p); why != "" {
			t.Errorf("默认登记表里出现了等于组默认的条目：%s", why)
		}
	}
	// 塞一条 L3 的 `debug` 条目回去 —— 判据必须当场红。
	r.Exceptions["debug exec"] = Layer3
	if why := exceptionEqualsGroupDefault(r, "debug exec"); why == "" {
		t.Error("把 `debug exec`（L3）塞回例外表，判据该红 —— 否则它拦不住『登了等于默认的条目』")
	}
	// 与组默认真不同的、以及"待定"的，判据不该报。
	r.Exceptions["dict r.t"] = Layer0 // = dict 组默认 L0 → 该报
	if why := exceptionEqualsGroupDefault(r, "dict r.t"); why == "" {
		t.Error("`dict r.t`（L0）等于 dict 组默认（L0），判据该报它")
	}
	if why := exceptionEqualsGroupDefault(r, "debug sql"); why != "" {
		t.Errorf("`debug sql`（L2）与 debug 组默认（L3）不同，判据不该报它，实得 %q", why)
	}
	if why := exceptionEqualsGroupDefault(r, "dev tzs list_tables"); why != "" {
		t.Errorf("『待定』不参与 Q8 这条判据，判据不该报它，实得 %q", why)
	}
	// 组默认变了之后，这条判据对"真的与组默认不同"的例外仍然放行：
	// `session` 组默认是 L3，而 `session ls`（L0）与 `session new`（L1）都真的不同。
	if why := exceptionEqualsGroupDefault(r, "session ls"); why != "" {
		t.Errorf("`session ls`（L0）与 session 组默认（L3）不同，判据不该报它，实得 %q", why)
	}
	if why := exceptionEqualsGroupDefault(r, "session new"); why != "" {
		t.Errorf("`session new`（L1）与 session 组默认（L3）不同，判据不该报它，实得 %q", why)
	}

	// ③ **组默认之上，判据仍然拦得住"等于默认"的条目**（WP4c① 对元测试的要求）：
	//    `session` 组默认已是 L3，所以把一条 L3 的 session 条目塞回例外表必须**当场红**。
	if r.Groups["session"] != Layer3 {
		t.Fatalf("前置条件不成立：session 组默认 = %s，期望 L3（WP4c① / Q9）", r.Groups["session"])
	}
	for _, p := range []string{"session show", "session ent", "session exec"} {
		if _, ok := layerExceptions[p]; ok {
			t.Errorf("`%s` 与 session 的组默认（L3）同值，不该出现在包级例外表里（Q8）", p)
		}
		r.Exceptions[p] = Layer3 // = 组默认 → 判据必须点名
		if why := exceptionEqualsGroupDefault(r, p); why == "" {
			t.Errorf("把 `%s`（L3）塞回例外表，判据该红 —— 否则它拦不住『登记了等于组默认的条目』", p)
		}
	}
}
