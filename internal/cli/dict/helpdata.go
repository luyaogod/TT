package dict

import (
	"fmt"

	"tt/internal/cli/common"
	"tt/internal/dict/dbsync"

	"github.com/spf13/cobra"
)

// baseHelpFunc 是 cobra 的默认 help 实现。用一个无父命令的临时命令取出来,
// 免得取到自己或别的命令 SetHelpFunc 装过的那份(那会重复输出)。
var baseHelpFunc = (&cobra.Command{}).HelpFunc()

// versionLine 一行版本描述。版本号由打包脚本经 -ldflags 注入到 common.Version;
// 本地 go build 未注入时显示 devel(带 commit 的完整描述见 `tt version`)。
func versionLine() string {
	if common.Version != "" {
		return common.Version
	}
	return "devel"
}

// localDataHint 返回一行「本地数据」状态;keys 为空 = 全部数据族。
// 只读检查本地库,任何失败都静默(不该因为环境问题让 --help 报错)。
func localDataHint(keys ...string) string {
	st, err := dbsync.InspectLocal(resolveSyncTarget())
	if err != nil {
		return ""
	}
	return fmt.Sprintf("本地数据(%s): %s", st.Path, st.Summary(keys...))
}

// attachDataHint 在命令组上装 help 钩子。子命令继承命令组的 help 函数,
// 所以这一处就能让所有命令的 --help 末尾都带上版本与本地数据状态。
func attachDataHint() {
	Group.SetHelpFunc(func(c *cobra.Command, args []string) {
		baseHelpFunc(c, args)
		out := c.OutOrStdout()
		fmt.Fprintln(out, "\n版本: tt dict "+versionLine())
		if c == Group {
			fmt.Fprintln(out, localDataHint())
			return
		}
		// 依赖哪些族由 Families 现场反推(见 dbsync.FamilyKeysForCommand),
		// 不在这里另存一份映射 —— 那份抄本一定会和 Families 漂开。
		if keys := dbsync.FamilyKeysForCommand(c.Name()); len(keys) > 0 {
			fmt.Fprintln(out, localDataHint(keys...))
		}
	})
}
