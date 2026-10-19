// Package form 承载 `tt form` 命令组：出厂表单（`.42f`）的列出 / 查看 / 校验。
//
// 它在计划里的位置（唯一口径：`.agents/plan/tt-genero-formdef-contract.md`）：
//
//   - **C2 取件与寿命** —— 按需单取一张 `.42f`，只读，落 `<数据目录>/forms/`；
//   - **C5 CLI 面** —— `ls` / `show` / `verify`，`show --json` 的键名冻结在 C5.2；
//   - **§4 的 A8 接口归消费方** —— 取件接口（`Source`）定义在**这里**（消费方），
//     实现放两处：本地目录（`DirSource`，镜像/离线）与真机按需单取
//     （`internal/host/formsource.go`）。`internal/form42f` 只留纯解析，零 IO。
//
// 退出码（本组 —— 见 internal/cli/README.md 的码表）：
//
//	0  成功
//	1  FORM_UNKNOWN：表单取不到（网络/权限/路径）或不是合法 `.42f`
//	2  用法错（缺 `--env`/`--module`/`--form`、没有基线、语料里没有这个窗口）
//	6  体检不过：`verify` 与语料逐条比对出现不一致（`tt view` 的 4 是"产物自查"，这里是"字段表自查"）
//
// 输出：**本组默认给人读**（表格），`--json` 或显式 `--format json` 时走 C5.2 的冻结键名。
// 这条与本仓"默认 json"的通行取向不同，是契约 C5 明写的（`tt form show` 的 Agent 用法
// 是 `--json` 显式要），代价是脚本必须显式要 JSON —— 换来的是"不加开关就是给人看的"。
package form

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/output"
)

// 退出码（本组）。码表在 internal/cli/README.md；改这里等于改契约。
//
// 契约 §4 **A13** 把这一组写死为：
//
//	0  一致（成功）
//	1  取件失败（`FORM_UNKNOWN`：网络/权限/路径取不到，或不是合法 .42f）
//	2  用法错**或无基线**（缺 --env/--module/--form 或 --map；语料里没有点名/推到的那一节窗口）
//	6  verify **不一致**（复用既有 `output.CodePreflight` =「体检不过」，**不新造码名**）
const (
	exitOK          = 0
	exitFormUnknown = 1 // 取件失败 / 不是合法 .42f
	exitUsage       = 2 // 用法错，或无基线
	exitMismatch    = 6 // verify 与语料不一致（体检不过，CodePreflight）
)

// CodeFormUnknown 是契约 C4.2 冻结的码名之一（表单取不到，或不是合法 `.42f`）。
//
// 契约 §4 **A14**：M1 阶段它**暂定义在本包**（现在只有 `tt form` 一家用；过早集中会把
// 两个工作切片耦在同一条改动链上）；**M2 统一到 `internal/output`**，与 `CodeUsage` /
// `CodeTargetMissing` 同处。
//
// 名字逐字照抄契约，不新造码名（契约 §2：不许改错误码名）。另外三码
// （`FIELD_NOT_IN_SCOPE` / `FIELD_AMBIGUOUS`，A11 又加了 `FIELD_NOT_ACTIONABLE`）属
// **寻址面**，由 M2 的 `internal/form42f` 落；本组不寻址，所以只定义用得上的这一个。
const CodeFormUnknown = "FORM_UNKNOWN"

// Group 是 tt form 命令组。
var Group = &cobra.Command{
	Use:   "form",
	Short: "出厂表单：列出 / 查看字段表 / 与语料核对 / 预热缓存",
	Long: `按需取一张出厂表单（.42f，编译后的明文 XML）并把它读成**字段表**：工号（colName）、
表单内编号（fieldId）、控件、tabIndex、SQL 类型、必填、cmt_*/lbl_* 键与几何。

  tt form ls     --env <环境> --module <模块>
  tt form show   --env <环境> --module <模块> --form <表单名> [--json]
  tt form verify --env <环境> --module <模块> --form <表单名> --map <语料表.json>
  tt form fetch  --env <环境> --module <模块> [--force] [--json]
  tt form cache  [--json]

where 从哪来（三条路，优先级从上到下）：

  1. --dir <目录>（或 TT_FORM_DIR）：本地目录/镜像，离线可用；
  2. 否则按环境走 SSH 按需**单取**（只读：find + SFTP get，绝不写、不删、不起作业）；
  3. 取回来的原始 XML 进 <数据目录>/forms/ 缓存（--no-cache 可关）。

fetch 把该模块的表单**批量**取进同一份缓存（--force 全重取），cache 只看这份缓存的
概览（多少张/多少字节/最旧最新，**只读**，清理走 tt cache clear）。

退出码（契约 A13）：0 一致 / 1 表单取不到或不是合法 .42f（FORM_UNKNOWN）/
2 用法错或无基线 / 6 verify 与语料不一致（体检不过）。`,
	SilenceUsage: true,
	// 不认识的子命令必须非 0：cobra 默认会打一份帮助再退 0，而"我打错了命令"
	// 被读成成功是调用方最难自己发现的失败形态（internal/cli/common 里已有一处先例）。
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return common.UnknownSubcommand(cmd, args)
	},
}

// Register 把本命令组挂到根命令上（`internal/cli/root.go` 里一行）。
func Register(root *cobra.Command) { root.AddCommand(Group) }

func init() {
	// flag 打错也要按本组的码表（2 = 用法错），否则会落到 cobra 默认的 1。
	Group.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return failUsage("%v", err)
	})
	Group.AddCommand(lsCmd, showCmd, verifyCmd)
}

// noArgs 参数校验：本组三个动词都不收位置参数。
var noArgs = func(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return failUsage("%s 不接受位置参数（参数名一律用开关，如 --form）", cmd.CommandPath())
	}
	return nil
}

// ── 用法错与取件失败 ────────────────────────────────────────────────────────

// failUsage 造一个用法错（稳定码 USAGE，退出码 2）。
func failUsage(format string, a ...any) error {
	return &output.Error{Code: output.CodeUsage, Exit: exitUsage, Message: fmt.Sprintf(format, a...)}
}

// formUnknown 造一个 FORM_UNKNOWN（退出码 1）。hint 必须**可操作**（契约 C4.3）。
func formUnknown(hint, format string, a ...any) error {
	return (&output.Error{
		Code:    CodeFormUnknown,
		Exit:    exitFormUnknown,
		Message: fmt.Sprintf(format, a...),
	}).WithHint(hint)
}

// requireEnv 取环境名。环境一律**点名**（本仓没有"默认环境可回落"，见 config.Hosts）。
func requireEnv() (string, error) {
	env := strings.TrimSpace(common.Env)
	if env == "" {
		return "", failUsage("需要 --at <环境> 或 --env <环境名>（本仓没有默认环境可回落；离线可以用 --dir 指向本地镜像/目录）")
	}
	return env, nil
}

// requireModuleForm 取模块与表单名（两个都必填）。
func requireModuleForm(verb, module, form string) (string, string, error) {
	module, form = strings.TrimSpace(module), strings.TrimSpace(form)
	if module == "" || form == "" {
		return "", "", failUsage("%s 需要 --module <模块> 与 --form <表单名>（还要 --env <环境>）", verb)
	}
	return module, form, nil
}

// wantJSON 判本次是不是要走 JSON。
//
// 为什么要自己判一刀（而不是直接用 common.OutputFormat）：本仓 `--format` 的**默认值**
// 就是 json，于是"没给任何开关"与"显式 --format json"在 common 那个函数里长得一样。
// 本组的默认是**给人读**（契约 C5 的命令面），所以只有显式给了才走 JSON：
//
//	--json            → JSON（语法糖）
//	--format json     → JSON（显式）
//	--csv / --format table → 人读
//	什么都没给         → 人读
func wantJSON(cmd *cobra.Command) bool {
	if common.JSON {
		return true
	}
	if common.CSV {
		return false
	}
	if pf := cmd.Root().PersistentFlags(); pf.Lookup("format") != nil && pf.Changed("format") {
		return common.OutputFormat() == output.FormatJSON
	}
	return false
}

// writeOut 是本组唯一的"写给人看"出口（与既有命令组一致，直接写 os.Stdout，
// 于是 testkit.CaptureStdout 能把它收回来）。
func writeOut(s string, a ...any) { fmt.Fprintf(os.Stdout, s, a...) }
