package dict

// fieldtab 的**通道选择**：在线（直查服务器字典表）与本地（本地 SQLite 镜像）。
//
// 用户决策：**默认在线**。四条纪律，每条都在下面的代码里（不是注释里的承诺）：
//
//  1. **缺省 online 写在开关的缺省值里**（`--source` 的 DefValue），不靠"猜用户想要哪条"；
//  2. **不静默降级**（契约 A15：拿不准不猜）：`--source online` 而连不上/没环境 ⇒
//     明确报错 + 可执行的下一步，**绝不**开本地库假装在线成功；
//  3. **本地库不存在是确定缺失**（契约 A31 的 missing），不是"查不了"（unknown）——
//     两者后续动作不同：前者去 `tt dict sync`，后者去查权限/网络；
//  4. **只读**：在线只发 SELECT（live 包，见 internal/dict/live），本地只开读；
//     本文件不写任何库、不下任何同步指令（`tt dict sync` 是另一条命令）。
//
// 两条通道都**复用既有实现**（`openRemoteSource` / `openLocalSource`，见 source.go），
// 不另起一套连库/解析代码；本文件只做"选哪条 + 选不到怎么说话"。

import (
	"fmt"
	"strings"

	"tt/internal/cli/common"
	"tt/internal/dict/formfields"
)

// SourceMode 是字典查询的通道（也就是输出里 `dictSource` 的取值）。
type SourceMode string

const (
	// SourceOnline：在线 —— 直查服务器的字典表（`dzeb_t`/`dzebl_t` 等），只发 SELECT。
	SourceOnline SourceMode = "online"
	// SourceLocal：本地 —— 查本地 SQLite 镜像（`-d`/`TDICT_DB`/数据目录下的 `erp_data.db`）。
	SourceLocal SourceMode = "local"
	// SourceNone：显式不接字典（`--no-dict`）。中文名整列 `unresolved/no-dict`，命令照样成功。
	SourceNone SourceMode = "none"
)

// DefaultSourceMode 缺省通道 = **在线**（用户决策）。
const DefaultSourceMode = SourceOnline

// fieldtabSourceName `--source` 的取值：`online`（缺省）| `local`。
var fieldtabSourceName = string(DefaultSourceMode)

// parseSourceMode 把开关值解析成通道。空串（没传）= 缺省 = 在线。
// 认不出来的值**报错并列出可选项** —— 不猜（A15）。
func parseSourceMode(s string) (SourceMode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", string(SourceOnline):
		return SourceOnline, nil
	case string(SourceLocal):
		return SourceLocal, nil
	}
	return "", fmt.Errorf("--source %q 不认识：只能是 %s（在线直查，缺省）或 %s（本地 SQLite 镜像）",
		s, SourceOnline, SourceLocal)
}

// fieldtabSourceMode 本次该用哪条通道。`--no-dict` 优先于 `--source`（它是"干脆不接字典"）。
func fieldtabSourceMode() (SourceMode, error) {
	if fieldtabNoDict {
		return SourceNone, nil
	}
	return parseSourceMode(fieldtabSourceName)
}

// fieldtabDict 一条**已打开**的字典通道：中文名解析器 + 它的落脚点 + 关闭动作。
type fieldtabDict struct {
	Mode  SourceMode            // online | local | none
	Namer formfields.FieldNamer // nil = 不接字典（整列 unresolved/no-dict）
	Where string                // 落脚点：在线=环境名；本地=库文件路径；未接=空
	close func()                // 关闭底层连接（可空）
}

// Close 释放底层连接（本地文件句柄 / 远程连接与隧道）。重复调用安全。
func (d *fieldtabDict) Close() {
	if d != nil && d.close != nil {
		d.close()
	}
}

// ModeLine 人读的一行：这一轮用的到底是哪种通道（输出里必须看得见）。
func (d *fieldtabDict) ModeLine() string {
	switch d.Mode {
	case SourceOnline:
		line := "在线(直查服务器字典表)"
		if d.Where != "" {
			line += " · 环境 " + d.Where
		}
		return line
	case SourceLocal:
		return "本地(本地 SQLite 镜像)"
	}
	return "未接(--no-dict)"
}

// dictOpener 打开一条通道。做成参数而不是包级变量：测试把假实现传进来就行，
// 不需要在包级藏一个可变的接缝（reset_test.go 的清单也就不会再多一项）。
type dictOpener func() (*fieldtabDict, error)

// openFieldtabDict 按 mode 打开通道。**不降级**：online 只会调 online，
// local 只会调 local；谁失败就把错误原样交上去（由调用方补指引）。
func openFieldtabDict(mode SourceMode, online, local dictOpener) (*fieldtabDict, error) {
	switch mode {
	case SourceOnline:
		return online()
	case SourceLocal:
		return local()
	case SourceNone:
		return &fieldtabDict{Mode: SourceNone}, nil
	}
	return nil, fmt.Errorf("未知的字典通道 %q", mode)
}

// resetLandingPoint 清掉上一条命令留下的"落脚点"。真机上一命令一进程，这一步是空操作；
// 同一进程里连着跑两条（测试、tt serve 内的复用）时，它保证信封里不会留着上一轮的
// 环境/路径 —— 那正是"查错了地方却看着像查过了"的成因。
func resetLandingPoint() {
	srcTarget = nil
	srcNotes = nil
	srcLocal = false
}

// ensureQueryPolicy 补上命令组钩子本来会做的那一步：解析"这次用哪份配置"（srcConfigPath）。
//
// 为什么本命令要自己做：fieldtab 覆盖了 Group.PersistentPreRunE（见 skipQuerySource），
// 而 srcConfigPath 是**本地库定位**（数据目录下的 erp_data.db）与落盘的唯一依据 ——
// 少了它，`--source local` 会连"该去哪找库"都不知道（候选列表全是空的）。
// 读一次配置只为拿这个路径；读不到不算错（缺省落点照常能定位）。
func ensureQueryPolicy() { loadQueryPolicy() }

// ── 在线 ────────────────────────────────────────────────────────────────────

// openFieldtabOnline 打开在线通道：直查某环境的服务器字典表（只发 SELECT）。
//
// 环境从 `--env/--conn` 来（显式即强制），其次是配置里的 `query.source`。**都没有就报错** ——
// 从前那条"没配环境就用本地副本"的回落被这里刻意掐断（A15）：在线就是在线。
func openFieldtabOnline() (*fieldtabDict, error) {
	resetLandingPoint()
	ensureQueryPolicy()

	env := onlineEnvName()
	if env == "" {
		return nil, fmt.Errorf(
			"在线字典不可用：没有可用的环境 -- 既没给 --env/--conn，配置里 query.source 也不是环境名。\n\n"+
				"（--source 缺省就是 online；这一轮**不会**改走本地库来假装成功。）\n\n%s",
			fieldtabOnlineGuidance())
	}
	if err := openRemoteSource(env, dictEnt); err != nil {
		return nil, fmt.Errorf("在线字典不可用（环境 %q）：%w\n\n%s",
			env, err, fieldtabOnlineGuidance())
	}
	return &fieldtabDict{
		Mode:  SourceOnline,
		Namer: dataSrc,
		Where: env,
		close: closeQuerySource,
	}, nil
}

// onlineEnvName 本次可用的在线环境名；空 = 没有（那就该报错，不是回落）。
//
// `local` 与 `auto` 都不是环境名：它们是"用本地"，而本轮要的是在线 —— 如实返回空串。
func onlineEnvName() string {
	if e := strings.TrimSpace(common.Env); e != "" && !strings.EqualFold(e, string(SourceLocal)) {
		return e
	}
	if c := strings.TrimSpace(queryCfgSource()); c != "" &&
		!strings.EqualFold(c, "auto") && !strings.EqualFold(c, string(SourceLocal)) {
		return c
	}
	return ""
}

// fieldtabOnlineGuidance 在线不可用时的下一步：两条**能直接照抄**的命令。
func fieldtabOnlineGuidance() string {
	return "怎么办（二选一）：\n" +
		"  tt dict fieldtab <模块>/<表单> --source online --env <环境名>   # 指一个配了数据库的环境\n" +
		"  tt dict fieldtab <模块>/<表单> --source local                 # 改用本地镜像（需先 tt dict sync）"
}

// ── 本地 ────────────────────────────────────────────────────────────────────

// openFieldtabLocal 打开本地通道：本地 SQLite 镜像（只读）。
//
// 库文件不在是**确定缺失**（A31 的 missing）——报错、说清"不是查不了"，并给出同步命令。
// 从前这里是"打不开就算了、整列未解析、命令退 0"：那会把"本地库没同步"读成
// "字典里没有这个字段"，而两者的下一步完全不同。
func openFieldtabLocal() (*fieldtabDict, error) {
	resetLandingPoint()
	ensureQueryPolicy()

	// 先定位一次，好把"文件不在"与"文件在但打不开"分开说（两种原因，两种下一步）。
	path, err := resolveDBPath(dbPath)
	if err != nil {
		return nil, fmt.Errorf("%w\n\n"+
			"本地字典库不存在（missing：确定缺失，不是「查不了」unknown）。\n\n"+
			"怎么办：\n"+
			"  tt dict sync                                                 # 从 ERP 全量同步到本地镜像\n"+
			"  tt dict db status                                            # 先看各数据族缺什么\n"+
			"  tt dict fieldtab <模块>/<表单> --source online --env <环境名>   # 不落盘，直接查服务器", err)
	}
	if err := openLocalSource(); err != nil {
		return nil, err
	}
	return &fieldtabDict{
		Mode:  SourceLocal,
		Namer: dataSrc,
		Where: path,
		close: closeQuerySource,
	}, nil
}
