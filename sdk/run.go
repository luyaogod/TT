// 插件运行入口：连接 TT Core、调用 Start，等到 TT 要求退出后调用 Stop。
//
// 退出通道有两条：TT 的 supervisor 关掉 stdin（约定路径），或系统信号。
// 插件自己的 HTTP 服务（如果有）由插件在 Start 里起、在 Stop 里收。

package sdk

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

// Run 启动插件并阻塞到 TT 要求退出。
func Run(p Plugin) {
	if p == nil {
		fmt.Fprintln(os.Stderr, "sdk: 插件实现为空")
		os.Exit(2)
	}
	base := os.Getenv("TT_HOST_API_URL")
	token := os.Getenv("TT_PLUGIN_TOKEN")
	if base == "" || token == "" {
		fmt.Fprintln(os.Stderr, "sdk: 缺少 TT_HOST_API_URL / TT_PLUGIN_TOKEN（插件必须由 TT 启动）")
		os.Exit(2)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := newContext(ctx, base, token,
		os.Getenv("TT_PLUGIN_ID"), os.Getenv("TT_PLUGIN_VERSION"), os.Getenv("TT_PLUGIN_DATA_DIR"))

	if err := p.Start(c); err != nil {
		fmt.Fprintf(os.Stderr, "sdk: 插件启动失败: %v\n", err)
		os.Exit(1)
	}

	done := make(chan struct{})
	go func() {
		// supervisor 停止时关 stdin；读到 EOF 即收工。
		_, _ = io.Copy(io.Discard, os.Stdin)
		close(done)
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case <-done:
	case <-sig:
	case <-ctx.Done():
	}
	cancel()
	if err := p.Stop(c); err != nil {
		fmt.Fprintf(os.Stderr, "sdk: 插件停止失败: %v\n", err)
	}
}
