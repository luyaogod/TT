package dict

// 查询数据源(r.t/r.v/desc/scc/r.q 共用):本地 SQLite 镜像(*db.DB)与远程 ERP 库
// (*live.Live,金仓/Oracle)实现同一 db.Source 接口,数据源由配置/CLI 决定,
// 命令层不感知差异 —— 不再有 --online 开关。
//
// 选择优先级:
//  1. --env/--conn <环境名|local>(覆盖本次调用;**显式指定即强制**);
//  2. config.json 顶层 query.source(环境名 或 "local";可在 tt serve 的「设置」页切换);
//  3. 缺省 = **在线**:用默认环境(hosts.activeEnv)直查;一个环境都没配才用本地库。
//
// **明确不做自动降级**:在线就是在线、本地就是本地,连不上直接报错并提示怎么切
// (--env local,或前端「设置」页把「查询数据源」切成本地)。
//
// "local" = 现有 erp_data.db(--db / TDICT_DB);环境名 = hosts.sshs 中该环境的
// db(客户端直连,凭据取账号列表首项)。

import (
	"context"
	"fmt"
	"os"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/dbconfig"
	"tt/internal/dict/db"
	"tt/internal/dict/live"
	"tt/internal/host"
)

var (
	dataSrc  db.Source
	srcLocal bool // 当前源为本地 SQLite(错误提示文案分流)
)

// openQuerySource 在 Group PersistentPreRunE 打开数据源。
func openQuerySource() error {
	if common.Env != "" { // ① 命令行指定(--env/--conn):强制
		if common.Env == "local" {
			return openLocalSource()
		}
		return openRemoteSource(common.Env)
	}
	switch cfg := queryCfgSource(); cfg {
	case "local": // ② 配置指定本地
		return openLocalSource()
	case "", "auto": // ③ 未配置:在线(没配环境才本地)
		if env := defaultEnvName(); env != "" {
			if err := openRemoteSource(env); err != nil {
				return fmt.Errorf("%w\n提示: 想查本地库就加 `--env local`,或在 tt serve 的「设置 → 查询数据源」里切成「本地 SQLite」", err)
			}
			return nil
		}
		return openLocalSource()
	default: // ② 配置指定了环境名:强制
		return openRemoteSource(cfg)
	}
}

// closeQuerySource 在 Group PersistentPostRun 关闭数据源。
func closeQuerySource() {
	if dataSrc != nil {
		dataSrc.Close()
		dataSrc = nil
	}
}

// queryCfgSource 读 config.json 顶层 query.source("local" / 环境名 / "auto");
// 缺失、"auto" 或读不到配置时返回空串/auto(即走缺省策略:在线优先)。
func queryCfgSource() string {
	path, err := common.ResolveConfig(true)
	if err != nil {
		return ""
	}
	root, err := config.Load(path)
	if err != nil {
		return ""
	}
	return root.Query.Source
}

// openLocalSource 打开本地 SQLite 镜像(-d/TDICT_DB 定位,要求文件已存在)。
func openLocalSource() error {
	resolvedPath, err := resolveDBPath(dbPath)
	if err != nil {
		return err
	}
	if common.Verbose {
		fmt.Fprintf(os.Stderr, "[dict] 数据源: 本地 SQLite (%s)\n", resolvedPath)
	}
	database, err := db.Open(resolvedPath)
	if err != nil {
		return fmt.Errorf("打开本地数据库 %s 失败: %w", resolvedPath, err)
	}
	dataSrc = database
	srcLocal = true
	return nil
}

// openRemoteSource 打开指定 SSH 环境的远程库(客户端直连,首账号)。
func openRemoteSource(target string) error {
	path, err := common.ResolveConfig(false)
	if err != nil {
		return err
	}
	cfg, err := config.LoadHosts(path)
	if err != nil {
		return err
	}
	cc, err := envDBByName(cfg, target)
	if err != nil {
		return err
	}
	if common.Verbose {
		fmt.Fprintf(os.Stderr, "[dict] 数据源: 远程 %s (%s %s)\n", target, cc.Type, cc.Address())
	}
	l, err := live.Open(context.Background(), *cc)
	if err != nil {
		return fmt.Errorf("打开远程数据源 %q: %w", target, err)
	}
	dataSrc = l
	srcLocal = false
	return nil
}

// envDBByName 在 hosts.sshs 中按环境名取该环境 db 的深拷贝。
func envDBByName(cfg *host.Hosts, name string) (*dbconfig.Connection, error) {
	for i := range cfg.SSHs {
		e := &cfg.SSHs[i]
		if e.Name != name {
			continue
		}
		if e.DB == nil {
			return nil, fmt.Errorf("环境 %q 未配置数据库(可在 config.json hosts.sshs[].db 添加后可用 --env 直查)", name)
		}
		cc := *e.DB
		cc.Accounts = append([]dbconfig.DBAcct(nil), e.DB.Accounts...)
		return &cc, nil
	}
	return nil, fmt.Errorf("未找到环境 %q(可用: tt env list 查看环境名)", name)
}

// defaultEnvName 尽力取默认环境名(读配置失败或未配置时返回空串)。只用于错误提示,不报错。
func defaultEnvName() string {
	path, err := common.ResolveConfig(true)
	if err != nil {
		return ""
	}
	cfg, err := config.LoadHosts(path)
	if err != nil {
		return ""
	}
	if cfg.ActiveEnv != "" {
		return cfg.ActiveEnv
	}
	if len(cfg.SSHs) > 0 {
		return cfg.SSHs[0].Name
	}
	return ""
}

// missingHint 主字典表缺失(IsMissingTable)时的提示。本地源给**两条路**:全量同步,
// 以及"免写盘"的远程直查(有些环境不允许/不方便写本地库,只报 db sync 会把人引到死路);
// 远程源说明库的问题与回退方式。subject 如 "校验定义 (dzcd_t 等表)"。
func missingHint(subject string) string {
	if srcLocal {
		remote := "  tt dict <命令> --env <环境名>       # 免写盘,直接查远程"
		if env := defaultEnvName(); env != "" {
			remote = fmt.Sprintf("  tt dict <命令> --env %-11s # 免写盘,直接查远程(当前默认环境)", env)
		}
		return fmt.Sprintf("本地库尚未包含%s数据。二选一:\n  tt dict db sync                  # 从 ERP 全量刷新(原库自动备份为 .bak)\n%s\n看各数据族缺什么: tt dict db status", subject, remote)
	}
	return fmt.Sprintf("远程库缺少%s相关字典表(该环境数据库未同步或账号权限不足);可用 --env local 切回本地库", subject)
}

// emptyHint 查询零结果时的提示:本地源沿用"先 db sync"建议,远程源说明语义。
// subject 如 "校验定义"。
func emptyHint(subject string) string {
	if srcLocal {
		return fmt.Sprintf("暂无%s数据。请先执行: tt dict db sync", subject)
	}
	return fmt.Sprintf("未查询到%s数据(库中无记录或 --kw 无匹配;可换 --env <其他环境> 再试)", subject)
}
