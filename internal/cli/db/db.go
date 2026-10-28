// Package db 是“数据库”命令组：**连通性测试** + **只读 SQL**。
//
// 它不碰字典内容（`tt dict r.t / desc / prog …` 是组合查询，留在 dict），
// 也不碰调试会话（那是 `tt debug`）。三条命令的边界：
//
//	sql    对库执行一条只读 SQL（业务数据）
//	ping   从**本机**连一次库（客户端直连，可经 viaSsh 隧道）
//	probe  从**服务器**连一次库（SSH + sqlplus/ksql，验证该企业的账号）
//
// 身份统一用 `--at <环境>[@<企业>]`（根命令的全局开关，见 internal/cli/at）：
// 企业与账号来自 config.json 的 ENT 清单（设置页「ENT」页签同步），
// 所以这几个命令**默认不联网解析企业**；连不上只会报“连不上”，不会伪装成“没数据”。
package db

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/at"
	"tt/internal/cli/common"
	"tt/internal/dbconfig"
	"tt/internal/debug"
)

// Register 把 db 组挂到根命令。
func Register(root *cobra.Command) {
	g := &cobra.Command{
		Use:   "db",
		Short: "数据库：连通性测试与只读 SQL",
		Long: `数据库：测试连不连得上，以及直接执行一条只读 SQL。

  tt db sql     执行一条只读 SQL（单条 SELECT/WITH，只读事务，最多 200 行；
                默认客户端直连，建连失败回退服务器 sqlplus，--route 可指定）
  tt db ping    从本机连一次库（客户端直连；支持 viaSsh 隧道）
  tt db probe   从服务器连一次库（SSH + sqlplus/ksql；验证该企业的账号）

身份用 --at 点名（取值来自 config.json 的 ENT 清单，不联网）：
  tt db sql   --at <环境>@<企业>      # 企业必填：它决定用哪个账号
  tt db ping  --at <环境>[@<企业>]    # 不给企业 = 账号清单首项（只验“库通不通”）
  tt db probe --at <环境>@<企业>      # 企业必填：验的是这个企业的账号

字典内容查询（r.t / desc / prog …）在 tt dict；本组不碰它们。

退出码：0 成功 / 2 用法或身份错（环境/ENT 不在清单、--at 冲突）/ 3 只读闸门拒绝 /
4 连接或查询失败。`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	g.AddCommand(newSQLCmd(), newPingCmd(), newProbeCmd())
	root.AddCommand(g)
}

// loadCfg 读配置并切到本次点名的环境（--at/--env 已落到 common.Env）。
//
// 与 tt debug 的读法一致：CloneEnv 把该环境的 SSH/Zone/DB 装进运行时字段，
// 但不 ApplyDefaultEnv —— 那次合并会把“没点名就落首条”的老行为带回来。
func loadCfg() (*debug.Config, string, error) {
	env := strings.TrimSpace(common.Env)
	if env == "" {
		return nil, "", fmt.Errorf("需要 --at <环境> 或 --env <环境>（可用：tt env ents）")
	}
	path, err := common.ResolveConfig(false)
	if err != nil {
		return nil, "", err
	}
	cfg, err := debug.LoadConfig(path)
	if err != nil {
		return nil, "", err
	}
	sub := cfg.CloneEnv(env)
	if sub == nil {
		return nil, "", fmt.Errorf("环境 %q 不存在（tt env list 看全部）", env)
	}
	sub.DataDir = filepath.Dir(path)
	return sub, path, nil
}

// entOf 取本次点名的企业：命令自己的 --ent 优先（--at 的 @企业 由 at.Apply 写进它）。
func entOf(flag int) int {
	if flag > 0 {
		return flag
	}
	return at.Ent
}

// entAccount 从配置的 ENT 清单取该企业的账号。
//
// 只读配置：这里不做“没有就现查 gzou_t”的回落 —— 那是同步（tt env ents --refresh /
// ENT 页签）的活。清单里没有这个企业就报错并列出可用的，绝不猜一个账号。
func entAccount(cfg *debug.Config, ent int) (string, error) {
	name := cfg.EnvName()
	for i := range cfg.SSHs {
		if cfg.SSHs[i].Name != name {
			continue
		}
		e := &cfg.SSHs[i]
		for _, x := range e.Ents {
			if x.Ent != ent {
				continue
			}
			if a := strings.TrimSpace(x.Account); a != "" && a != "-" {
				return a, nil
			}
			return "", fmt.Errorf("企业 %d 在配置的 ENT 清单里没有可用账号（tt env ents %s 可看，或到设置页「ENT」页签刷新）", ent, name)
		}
		if len(e.Ents) == 0 {
			return "", fmt.Errorf("环境 %q 还没同步过 ENT 清单：跑 tt env ents %s --refresh（或设置页「ENT」页签）", name, name)
		}
		return "", fmt.Errorf("企业 %d 不在 %s 的 ENT 清单里（tt env ents %s 看全部）", ent, name, name)
	}
	return "", fmt.Errorf("环境 %q 不在配置里", name)
}

// dbConnFor 把该环境的库连接取出来（可选覆盖成某个企业的账号）。
func dbConnFor(cfg *debug.Config, ent int) (*dbconfig.Connection, error) {
	if cfg.DB == nil {
		return nil, fmt.Errorf("环境 %q 未挂数据库连接（设置 → 站点管理 → 数据库）", cfg.EnvName())
	}
	cc := *cfg.DB // 浅拷贝：只改 Account 槽位，Accounts 切片只读
	if ent > 0 {
		acct, err := entAccount(cfg, ent)
		if err != nil {
			return nil, err
		}
		cc.Account = acct
	}
	return &cc, nil
}
