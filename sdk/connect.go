// Connect 给**一次性**调用方（CLI 子命令）用：只建 Context，不接管进程生命周期。
//
// 服务插件用 sdk.Run；CLI 插件跑完就退，所以这里不等待 stdin/信号。

package sdk

import (
	"context"
	"os"
)

// Connect 读当前进程注入的 TT_* 环境变量并建立 Context。
//
// 第二个返回值是清理函数（当前无资源需要释放，预留给将来的连接池）。
func Connect() (Context, func(), error) {
	base := os.Getenv("TT_HOST_API_URL")
	token := os.Getenv("TT_PLUGIN_TOKEN")
	if base == "" || token == "" {
		return nil, nil, &Error{
			Code:    CodeInvalidArgument,
			Message: "缺少 TT_HOST_API_URL / TT_PLUGIN_TOKEN（插件必须由 tt 启动）",
		}
	}
	c := newContext(context.Background(), base, token,
		os.Getenv("TT_PLUGIN_ID"), os.Getenv("TT_PLUGIN_VERSION"), os.Getenv("TT_PLUGIN_DATA_DIR"))
	return c, func() {}, nil
}
