package cli

// tt plugin audit：看 Host API 的调用审计（JSONL）。只读，不执行插件代码。

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/audit"
	"github.com/luyaogod/TT/internal/cli/common"
)

func newPluginAuditCmd() *cobra.Command {
	var tail int
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "查看 Host API 调用审计（最近的在前为原始顺序）",
		Long: `查看 <配置目录>/logs/audit.jsonl：谁（内置/某个插件）在什么时候连了哪个环境、
做了什么、成功与否、耗时多少。

只记元数据，不含 SQL 文本与命令内容。`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := audit.DefaultPath(common.ConfigPath)
			b, err := os.ReadFile(path)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("还没有审计记录（", path, "）")
					return nil
				}
				return err
			}
			lines := strings.Split(strings.TrimSpace(string(b)), "\n")
			start := 0
			if tail > 0 && len(lines) > tail {
				start = len(lines) - tail
			}
			for _, line := range lines[start:] {
				if strings.TrimSpace(line) == "" {
					continue
				}
				if asJSON {
					fmt.Println(line)
				} else {
					fmt.Println(audit.FormatLine([]byte(line)))
				}
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&tail, "tail", 50, "只看最后 N 条（0 = 全部）")
	cmd.Flags().BoolVar(&asJSON, "json", false, "原样输出 JSONL")
	return cmd
}
