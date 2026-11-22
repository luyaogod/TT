// 动词的接线信息：cobra 叶命令的展示面（Use/Short/Long）+ 核心函数。
//
// 拆成一份清单是刻意的：父包 register.go 只负责把清单挂上 cobra 树（路由、组级帮助、
// 退出码透传），动词的"它叫什么、说明书写什么、核心是哪个函数"都住在这层，
// 改动词不用碰父包。
package tzc

import "strings"

// Verb 是一条 tzc 动词的接线信息。Run 收**原始参数**（含 flag）——
// 参数解析权在动词自己的解析器（cli.go 的 parseArgs），cobra 不碰。
type Verb struct {
	Use   string
	Short string
	Long  string
	Run   func(args []string) int
}

// Run 按动词名跑一条 tzc 动词（工作区命令包一层默认参数时用，比如 `tt dev apply`
// 就是把 `<工作区>/ws` 填成第一个位置参数再交给 apply）。
// 未知动词返回 2，与组命令的"未知动词退 2"一致。
func Run(verb string, args []string) int {
	for _, v := range Verbs() {
		if fields := strings.Fields(v.Use); len(fields) > 0 && fields[0] == verb {
			return v.Run(args)
		}
	}
	return 2
}

// Verbs 按帮助顺序列出八个动词，外加一个生成修改标记的 uuid 小工具。
func Verbs() []Verb {
	return []Verb{
		{"export <pkg.tzc>", "渲染带围栏的 4GL 工作区（只读包）", exportLong, cmdExport},
		{"status [<dir>]", "报告工作区相对上次 export/apply 的改动（不写盘）", statusLong, cmdStatus},
		{"verify [<dir>]", "不写盘跑完整验证管线（gate1 + gate2）", verifyLong, cmdVerify},
		{"apply [<dir>]", "验证管线全阶段 + 原子写包 + git commit（唯一会改 .tzc 的命令）", applyLong, cmdApply},
		{"unlock [<dir>]", "框架解锁：单向状态迁移，只改 workspace", unlockLong, cmdUnlock},
		{"rename [<dir>] <旧函数名> <新函数名>", "结构事务：改自订定义点的函数名（只改 workspace）", renameLong, cmdRename},
		{"newfn [<dir>]", "新增自订定义点（只改 workspace）", newfnLong, cmdNewfn},
		{"points [<dir>]", "列区域（点名 / 行数 / 能否编辑）", pointsLong, cmdPoints},
		{"show [<dir>] <点名>", "只看一个区域的正文（带行号）", showLong, cmdShowRegion},
		{"find [<dir>] <点名> <子串>", "在区域里定位（带上下文）", findLong, cmdFindInRegion},
		{"patch [<dir>] <点名> --old … --new …", "区域内锚点替换（不要求整段重写）", patchLong, cmdPatchRegion},
		{"set [<dir>] <点名> --file/--stdin", "整段替换（仅小区域；超限请用 patch）", setLong, cmdSetRegion},
		{"uuid", "生成 16 位修改标记（用户拒绝提供单号时用）", uuidLong, cmdUUID},
		{"selftest", "内置 31 项对抗用例（全部合成包）", selftestLong, cmdSelftest},
	}
}

// TzcLong 是 tt dev tzc 的组级说明（与 Usage 的生命周期段同源措辞）。
const TzcLong = `.tzc 代码包管线：Package → Document → Workspace（export，只读包）→ 编辑 prog.full.4gl →
verify（gate1 + gate2）/ apply（三道闸门 + 原子写包）。

四个生命周期动词 + unlock 状态迁移 + 五个区域级动词 + uuid 小工具 + selftest。

区域级动词（points/show/find/patch/set）只改 workspace，不写包：
  points  列区域（点名/行数/能否编辑）
  show    只看一个区域的正文（十几行，不必读 7.7 万行的整份文件）
  find    在区域里定位（大区段专用）
  patch   区域内锚点替换（原文必须唯一命中）
  set     整段替换（仅小区域；超过 --max-lines 请用 patch）

改框架区段必须先 unlock：解锁改变的是门，不是所有房间（锚点区段永远只读）。

报错定位：验证失败会打印「prog.full.4gl:<行号>」+ 该行内容
（基线侧问题给「.tdev/base.full.4gl:<行号>」）；--json 里有 file/line/snippet 字段。

退出码：0 成功 / 2 包格式错 / 3 验证失败 / 4 写入被拒 / 5 IO·环境失败

工作区动词的 <dir> 可以省略：先 cd 进工作区，命令就不用再写目录。
  cd D:\pkg\capt110-ws
  tt dev tzc status          # = tt dev tzc status D:\pkg\capt110-ws
  tt dev tzc rename adzi999_calc adzi999_count   # 省略 <dir> 时 rename 收 2 个位置参数`

// 各动词的 Long：用法行 + 关键开关，与 Usage 里的对应行同源措辞；
// 动词缺参数时核心自己打的"用法："也是同一行，两处不会漂。
const (
	exportLong = `把 .tzc 代码包渲染成带围栏的 4GL 工作区（只读包）。

用法：tt dev tzc export <pkg.tzc> [-o <dir>] [--only <点名>...] [--json]
  -o 省略时默认导出到 <包所在目录>/<程序名>-ws（身份后缀 (c)/(s) 会去掉）
  --only 部分导出：只渲染这些点为可编辑（可重复/逗号分隔）
  --allow-sec 已废弃：请改用 tt dev tzc unlock <dir>`

	statusLong = `报告工作区相对上次 export/apply 的改动（含行号，不写盘）。

用法：tt dev tzc status [<dir>] [--json]   # <dir> 省略时用当前目录`

	verifyLong = `不写盘跑完整验证管线（gate1 + gate2）。

用法：tt dev tzc verify [<dir>] [--json] [--strict]   # <dir> 省略时用当前目录
  --strict 把 warn 级发现也算作失败（供 CI 使用）`

	applyLong = `验证管线全阶段 + 原子写包 + git commit（唯一会改 .tzc 的命令）。

用法：tt dev tzc apply [<dir>] [-o <pkg.tzc>] [--dry-run] [--yes] [--json]   # <dir> 省略时用当前目录
  --dry-run 跑完管线并输出逐条目 diff，不落盘、不 commit
  --yes 已废弃（形参只为兼容旧脚本，apply 不弹二次确认）：
        要写区段必须先 tt dev tzc unlock --yes 把 workspace 迁到 Unlocked，否则 gate1 直接退 4`

	unlockLong = `框架解锁：单向状态迁移，只改 workspace（落到 .tzc 仍走 apply）。

用法：tt dev tzc unlock [<dir>] [--yes] [--json]   # <dir> 省略时用当前目录
  解开不可逆：解开后规格调整不再自动生成程序代码；需要显式 --yes（AI 不得自主加）`

	renameLong = `结构事务：改自订定义点的函数名（只改 workspace，落盘仍走 apply）。

用法：tt dev tzc rename [<dir>] <旧函数名> <新函数名> [--scope PUBLIC|PRIVATE] [--desc <描述>] [--json]
  <dir> 省略时用当前目录；围栏 fn + 签名行 + 描述块三处原子同步`

	newfnLong = `新增自订定义点（只改 workspace，落盘仍走 apply）。

用法：tt dev tzc newfn [<dir>] --type FUNCTION|DIALOG|REPORT [--name <名>] [--scope …] [--json]
  <dir> 省略时用当前目录；--name 省略则默认 <prog>_newfunc 并自动去重`

	selftestLong = `内置 31 项对抗用例（全部合成包，不需要真实语料）。

用法：tt dev tzc selftest [--json]`

	uuidLong = `生成一个 16 位十六进制修改标记（8 字节随机数）。

用法：tt dev tzc uuid [--json]
  用户在开工前拒绝提供 <单号>#<项次> 时，用它生成标记（工号写 agent）。
  命令行只负责生成 id；摘要行与 -s/-e 块的写法属于项目规范，见 tt-dev 技能。`
)

// 区域级动词（points/show/find/patch/set）：把"区域"变成一等对象。
// 共同点：只改 workspace，不写包；落包仍走 verify / apply（三道闸门不变）。
const (
	pointsLong = `列区域：点名 / 类型 / 行数 / 能不能改（不能改的给拒绝码与原因）。

用法：tt dev tzc points [<dir>] [--kind point|section] [--match <子串>] [--editable] [--nonempty] [--json]
  --match 只列点名包含该子串的区域；--editable 只看能改的；--nonempty 只看正文非空的`

	showLong = `只看一个区域的正文（带行号），不读整份 prog.full.4gl。

用法：tt dev tzc show [<dir>] <点名> [--max N] [--json]`

	findLong = `在区域内按子串定位（带上下文），用于大区域（解开框架后的 section 正文）。

用法：tt dev tzc find [<dir>] <点名> <子串> [--context N] [--case] [--limit N] [--json]`

	patchLong = `区域内锚点替换：原文必须在该区域里**唯一命中**，不要求整段重写。

用法：tt dev tzc patch [<dir>] <点名> (--line N --text S | --old S --new S | --old-file F --new-file F) [--dry-run] [--json]
  --line/--text 按 prog.full.4gl 行号改整行（保留原缩进）；show/find 打印的就是这个行号
  --old/--new 按子串替换，须在区域内唯一命中（0 处或多处都被拒）
  --dry-run 只预览替换结果，不写盘`

	setLong = `整段替换区域正文。**只适合小区域**：超过 --max-lines 直接拒，改用 patch。

用法：tt dev tzc set [<dir>] <点名> (--file F | --stdin) [--max-lines N] [--dry-run] [--json]`
)
