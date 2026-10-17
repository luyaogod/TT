package form

// `tt form ls`：列出某个模块下的表单名（契约 C5.1）。
//
// 它是 `show` 的前一步：`--form` 打不出来时先 ls，别靠猜。清单**不缓存**（见
// CachingSource.List），且"该模块下一个都没有"要报错而不是给空清单 ——
// 空清单会被读成"这个模块没有表单"，而真相多半是路径/模块名不对。

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
)

var (
	lsModule string
	lsDir    string
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "列出该模块下的表单名",
	Long: `列出该模块下的 .42f 表单名（不含扩展名），可直接喂给 tt form show --form。

取件通道与 show 相同：--dir <本地目录> / TT_FORM_DIR 走离线，否则按环境走 SSH 只读单取。`,
	Args: noArgs,
	RunE: runLs,
}

func init() {
	lsCmd.Flags().StringVar(&lsModule, "module", "", "模块名（如 axm）")
	lsCmd.Flags().StringVar(&lsDir, "dir", "", "从本地目录/镜像列（离线；也可用 TT_FORM_DIR）")
}

func runLs(cmd *cobra.Command, _ []string) error {
	env, err := requireEnv()
	if err != nil {
		return err
	}
	module := strings.TrimSpace(lsModule)
	if module == "" {
		return failUsage("ls 需要 --module <模块>（还要 --env <环境>）")
	}
	dir := firstNonEmpty(lsDir, os.Getenv("TT_FORM_DIR"))
	src, err := sourceFor(env, sourceOpts{Dir: dir, NoCache: true})
	if err != nil {
		return formUnknown("检查 --env 对应的环境配置，或用 --dir 指向本地镜像/目录", "%v", err)
	}
	names, err := src.List(cmd.Context(), env, module)
	if err != nil {
		return formUnknown("检查模块名与取件通道（--dir / 环境配置）；"+
			"确认取件通道本身没问题可以试：tt form ls --env "+env+" --module "+module,
			"列不出 %s 模块下的表单（环境 %s）：%v", module, env, err)
	}
	if names == nil {
		names = []string{}
	}

	if wantJSON(cmd) {
		return common.PrintJSON(map[string]any{
			"env":    env,
			"module": module,
			"count":  len(names),
			"forms":  names,
			"source": sourceLabel(dir),
		})
	}
	for _, n := range names {
		writeOut("%s\n", n)
	}
	writeOut("（%s / %s：%d 张表单；取件通道 = %s）\n", env, module, len(names), sourceLabel(dir))
	return nil
}

// sourceLabel 把"这次从哪取的"写进输出 —— 排查"为什么和刚才不一样"时，
// 第一个要回答的就是这个。
func sourceLabel(dir string) string {
	if dir != "" {
		return "本地目录 " + dir
	}
	return "环境 SSH（按需单取，只读）"
}
