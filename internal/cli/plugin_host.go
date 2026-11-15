package cli

// serve 侧的插件生命周期钩子：启用即起服务、停用即收服务/撤令牌/关会话。
//
// 为什么只有 serve 有钩子而 CLI 没有：常驻服务必须是**长命进程的子进程**。
// `tt plugin enable` 是一次性进程，它起的服务会随命令退出而死 —— 那看起来像
// "启用了但没在跑"。所以 CLI 只改状态，服务由 tt serve 在启动时/收到启用后拉起。

import (
	"context"
	"fmt"

	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/internal/hostrpc"
	"github.com/luyaogod/TT/internal/plugin"
)

// servePluginHooks 把插件状态变更接到 supervisor / 令牌表 / 连接服务上。
type servePluginHooks struct {
	sup    *plugin.Supervisor
	tokens *hostrpc.TokenStore
	conn   *connsvc.Service
	// logf 落服务日志。
	logf func(format string, a ...any)
}

// AfterEnable 启用后起服务（没有 service 贡献就什么也不做）。
func (h servePluginHooks) AfterEnable(rec *plugin.Record) error {
	if rec == nil || rec.Contributions.Service == nil {
		return nil
	}
	_, err := h.sup.Start(context.Background(), rec)
	if err != nil {
		return fmt.Errorf("服务启动失败: %w", err)
	}
	return nil
}

// BeforeDisable 停服务 → 撤令牌 → 关会话。顺序不能反：
// 先把"正在用连接的进程"停掉，再让它的凭证失效，最后收连接；
// 反过来会在停服务的过程中把它的连接抽走，日志里全是莫名其妙的断连。
func (h servePluginHooks) BeforeDisable(rec *plugin.Record) error {
	if rec == nil {
		return nil
	}
	if err := h.sup.Stop(rec.ID); err != nil {
		return fmt.Errorf("停止服务失败: %w", err)
	}
	h.tokens.RevokePlugin(rec.ID)
	n := h.conn.CloseForPlugin(rec.ID)
	if n > 0 && h.logf != nil {
		h.logf("插件 %s 停用：已关闭 %d 个会话", rec.ID, n)
	}
	return nil
}

// AfterDisable 停用后的收尾（skills/菜单在各自的阶段挂在这里）。
func (h servePluginHooks) AfterDisable(rec *plugin.Record) error {
	if rec != nil && h.logf != nil {
		h.logf("插件 %s 已停用（程序、授权与数据保留）", rec.ID)
	}
	return nil
}

// AfterUninstall 卸载后的收尾：注册项已经没了，凭证与会话也要清干净。
func (h servePluginHooks) AfterUninstall(id string) {
	h.tokens.RevokePlugin(id)
	h.conn.CloseForPlugin(id)
}

// startEnabledServices 启动时把已启用插件的服务拉起来。
//
// 单个插件失败不阻断其他插件与 tt serve 本身：记 LastError、继续下一个。
func startEnabledServices(m *plugin.Manager, sup *plugin.Supervisor, logf func(string, ...any)) {
	reg, err := m.Store.Load()
	if err != nil {
		if logf != nil {
			logf("读取插件注册表失败，跳过插件服务启动: %v", err)
		}
		return
	}
	for _, rec := range reg.Enabled() {
		if rec.Contributions.Service == nil {
			continue
		}
		if _, err := sup.Start(context.Background(), rec); err != nil {
			if logf != nil {
				logf("插件 %s 的服务启动失败: %v", rec.ID, err)
			}
		}
	}
}

// serveUpdateRuntime 升级/回滚期间对运行态的操作（只有 serve 进程有这一套）。
//
// 为什么 CLI 里没有：常驻服务必须是长命进程的子进程。`tt plugin update` 是一次性进程，
// 它既停不掉 serve 手里的子进程，也不该"起一个随命令退出就死的服务"。所以 CLI 走
// Manager.ensureServiceIdle 那条路（先确认没有别人的服务在跑），serve 走这一条。
type serveUpdateRuntime struct {
	sup    *plugin.Supervisor
	tokens *hostrpc.TokenStore
	conn   *connsvc.Service
}

// StopService 停服务（不存在时什么也不做）。
func (r serveUpdateRuntime) StopService(id string) error { return r.sup.Stop(id) }

// StartService 起服务并等健康检查通过。
func (r serveUpdateRuntime) StartService(rec *plugin.Record) error {
	if rec == nil || rec.Contributions.Service == nil {
		return nil
	}
	if _, err := r.sup.Start(context.Background(), rec); err != nil {
		return fmt.Errorf("服务启动失败: %w", err)
	}
	return nil
}

// InvalidateInstances 撤旧实例令牌 + 关它的会话（P6-05：新旧版本不能共用凭证）。
func (r serveUpdateRuntime) InvalidateInstances(id string) {
	r.tokens.RevokePlugin(id)
	r.conn.CloseForPlugin(id)
}
