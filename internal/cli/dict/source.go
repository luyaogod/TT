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
	"path/filepath"
	"time"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/dbconfig"
	"tt/internal/dict/db"
	"tt/internal/dict/live"
	"tt/internal/output"
)

var (
	dataSrc  db.Source
	srcLocal bool // 当前源为本地 SQLite(错误提示文案分流)

	// srcTarget 当前数据源的"落脚点"(环境/账号/库/路径)—— openQuerySource 时填好,
	// 由查询命令打进输出信封。它是"我这次到底查的是哪儿"的唯一答案。
	srcTarget *dbconfig.Target
	// srcNotes 打开数据源时产生的降级提示(如企业账号解析失败改用了列表首项)。
	srcNotes []string
	// srcStarted 数据源打开的时刻,用来算信封里的"用时" —— 它含连接与环境解析,
	// 不只是查询本身,那正是"这条命令为什么慢"要看的数。
	srcStarted time.Time
)

// openQuerySource 在 Group PersistentPreRunE 打开数据源。
func openQuerySource() error {
	srcStarted = time.Now()
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
	srcTarget = &dbconfig.Target{Env: "local", Route: dbconfig.RouteLocalSQLite, Path: resolvedPath}
	srcNotes = nil
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
	ref, err := cfg.Resolve(target)
	if err != nil {
		return err
	}
	cc, err := dbConnOf(ref)
	if err != nil {
		return err
	}
	srcTarget = dbconfig.NewTarget(cc, ref.Name, ref.Env.Host, ref.Env.User, ref.Env.Zone,
		string(ref.Env.Topent), dbconfig.RouteClientDirect)
	// 用哪个账号由企业编号(topent)决定:与 tt debug 同一条规则、同一份快照 —— 见 entacct.go
	srcNotes = hookEntAccount(cc, ref, dataDirOf(path))
	if common.Verbose {
		fmt.Fprintf(os.Stderr, "[dict] 数据源: 远程 %s (账号 %s/%s, %s %s)\n",
			ref.Name, srcTarget.Account, srcTarget.AccountSource, cc.Type, cc.Address())
	}
	l, err := live.Open(context.Background(), *cc)
	if err != nil {
		return fmt.Errorf("打开远程数据源 %q: %w", ref.Name, err)
	}
	dataSrc = l
	srcLocal = false
	return nil
}

// dataDirOf 数据目录 = 配置文件所在目录(与调试侧同一条约定:ents/、srccache/ 都跟着它)。
// 企业目录快照两处共用一个文件,靠的就是这条约定。
func dataDirOf(configPath string) string {
	if configPath == "" {
		return ""
	}
	return filepath.Dir(configPath)
}

// dbConnOf 取某环境一对一挂载的库连接(深拷贝:调用方要改 User/Password 等运行期槽位,
// 不能写回配置里的那份)。
func dbConnOf(ref *config.EnvRef) (*dbconfig.Connection, error) {
	if ref.Env.DB == nil {
		return nil, fmt.Errorf("环境 %q 未配置数据库(设置-环境-数据库页添加,或编辑 config.json hosts.sshs[].db)", ref.Name)
	}
	cc := *ref.Env.DB
	cc.Accounts = append([]dbconfig.DBAcct(nil), ref.Env.DB.Accounts...)
	return &cc, nil
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
	ref, err := cfg.Resolve("")
	if err != nil {
		return ""
	}
	return ref.Name
}

// missingTableErr 主字典表缺失。**这是错误(退出码 3),不是一句提示。**
//
// 从前它在十来处都是"打一句提示、然后退出 0",脚本与 agent 会把"没查过"
// 读成"查过了,没有" —— 而这两件事的后续动作完全不同(前者去 sync,后者改条件)。
// 缺表意味着数据源不完整,那是失败。
//
// 注意只用于**命令的主数据**;程序族那种"缺了还能靠另一族兜底"的降级
// (printProgTables、程序详情的用表索引)仍然打提示、不报错。
func missingTableErr(subject string) error {
	return output.Errorf(output.CodeTableMissing, output.ExitMissing, "%s", missingHint(subject))
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
