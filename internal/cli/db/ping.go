package db

// tt db ping —— 从**本机**连一次库（客户端直连，可经 viaSsh 隧道）。
// 原 tt dict db ping；区别是身份用 --at，且给 @企业 时用配置清单里的那个账号
// （不给企业就按客户端惯例用账号清单首项 —— 这一档只回答“库通不通”）。
//
// P6-09 的第一方迁移点：这条链路原来经 internal/dict/live 自己建隧道、自己开连接，
// 与设置页「测试连接」和第三方插件的 ConnTest 是**两套编排**。现在它直接调
// connsvc.Service.TestDBConn —— 第一方直连与 Host API 的 RPC 走同一份实现，
// "隧道怎么起、连一次、取版本、怎么收"只在这里之外的一处（internal/connsvc）。
// 输出形状与退出码保持原样：TT 的旧接口不因为内部换了实现而变。

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/cli/common"
	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/pkg/dbconfig"
	"github.com/luyaogod/TT/internal/output"
)

func newPingCmd() *cobra.Command {
	var entFlag int
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "从本机连一次库（客户端直连；支持 viaSsh 隧道）",
		Long: `从**本机**连一次库，验证网络/VPN/隧道与账号口令是否可用（客户端直连视角）。

  tt db ping --at 主机109          # 用环境配置里的账号清单（首项）连一次
  tt db ping --at 主机109@99       # 用配置 ENT 清单里企业 99 的账号连一次
  tt db ping --at 主机109@99 --json

对端能不能连用 tt db probe（服务器侧视角）；查数据用 tt db sql。`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, path, err := loadCfg()
			if err != nil {
				return output.Errorf("TARGET", 2, "%v", err)
			}
			ent := entOf(entFlag)
			cc, err := dbConnFor(cfg, ent)
			if err != nil {
				return output.Errorf("TARGET", 2, "%v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
			defer cancel()
			// 内置调用方：这是 TT 自己的链路，不经插件授权。
			svc := connsvc.New(connsvc.Options{ConfigPath: path})
			res := runDBPing(ctx, svc, *cc)
			if !res.OK {
				return output.Errorf("DB_UNREACHABLE", 4, "%v", pingFailure(res))
			}
			oneLine := strings.TrimSpace(strings.SplitN(res.ServerVersion, "\n", 2)[0])

			if common.IsJSON() {
				return common.PrintJSON(map[string]any{
					"ok": true, "env": cfg.EnvName(), "ent": ent,
					"account": cc.Account, "dialect": cc.Type, "target": cc.Address(),
					"version": oneLine,
				})
			}
			who := cfg.EnvName()
			if ent > 0 {
				who = fmt.Sprintf("%s@%d", cfg.EnvName(), ent)
			}
			acct := cc.Account
			if acct == "" {
				acct = "账号清单首项"
			}
			fmt.Printf("连接正常: %s (%s, %s) %s\n", who, acct, cc.Type, oneLine)
			return nil
		},
	}
	cmd.Flags().IntVar(&entFlag, "ent", 0, "企业编号（也可用 --at <环境>@<企业>；不给=账号清单首项）")
	return cmd
}

// runDBPing 走进程内连接服务测一次库。
//
// 单独成一个函数是为了可测：用例可以注入 Service 的 DialTunnel/OpenDB 替身，
// 再拿同一条 Service 去走 Host API 的 RPC 入口，证明两边是同一份实现。
func runDBPing(ctx context.Context, svc *connsvc.Service, conn dbconfig.Connection) *connsvc.TestDBResult {
	return svc.TestDBConn(ctx, conn)
}

// pingFailure 把连接层失败描述成一行"卡在哪一步"。
func pingFailure(res *connsvc.TestDBResult) string {
	stage := res.Stage
	if stage == "" {
		stage = "connect"
	}
	return fmt.Sprintf("连不上（阶段 %s，目标 %s）: %s", stage, res.Address, res.Message)
}
