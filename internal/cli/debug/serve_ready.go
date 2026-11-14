package debug

// 桌面版（TT Desktop）的启动契约：**一行 JSON**。
//
// 为什么不能用"服务已启动 …"那段人读文字：壳需要三件它自己算不出来的事实 ——
// 真实地址（端口被占用会 +1 顺延）、pid（退出时按它收尾）、配置文件路径（日志与
// "打开数据目录"要用）。那段文字里三件都有，但它是给人看的界面文案，改一次字距
// 壳就解析不出来；而且它一打就是三行。所以这里另给一行**机器契约**：
//
//	{"ok":true,"attached":false,"url":"http://127.0.0.1:28670","pid":123,"config":"C:\\…\\config.json"}
//
// 与 `tt dev` 那条线同一个约定：`--json` 下 stdout 恰好一个对象。所以给出
// --ready-json 时**不再**打那段人读块（否则 stdout 上有两个对象，消费者要自己挑）。
//
// 失败不走这里：沿用全仓的错误信封（也是 stdout 一行 {"ok":false,…}），退出码非 0。
// 壳的判据因此只有一条：读到 ok 且 attached 就开窗，否则展示 error。

import (
	"encoding/json"
	"io"
)

// readyLine 是 --ready-json 的输出形状。字段只增不改：加了键旧壳照样能读，
// 改键却会让已经发出去的壳读到一个空地址。
type readyLine struct {
	OK       bool   `json:"ok"`
	Attached bool   `json:"attached"`
	URL      string `json:"url"`
	PID      int    `json:"pid"`
	Config   string `json:"config"`
}

// WriteServeReady 向 w 打一行就绪 JSON（含行尾换行，恰好一行）。
//
// attached=true 表示"同目录已有实例，我没有启动新的"：此时 pid 是那个实例的 pid，
// 壳**不能**在退出时收掉它 —— 它不是壳拉起来的。
func WriteServeReady(w io.Writer, st *ServeInfo, cfgPath string, attached bool) error {
	return json.NewEncoder(w).Encode(readyLine{
		OK:       true,
		Attached: attached,
		URL:      st.URL,
		PID:      st.PID,
		Config:   cfgPath,
	})
}
