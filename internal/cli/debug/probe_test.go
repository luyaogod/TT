package debug

// 这组测试盯的是"probe 连的是配置里那台机器"这条线。
//
// 从前 probe 只 LoadConfig、不 ApplyDefaultEnv,而 loadConfig 只把环境名写进 cfg.envName、
// 不同步到运行时字段(见 internal/debug/config.go 的 loadConfig/applySsh)。于是 cfg.SSH
// 是零值,probe 永远打印并去连一台空主机:
//
//	[probe] 服务器 :22 区域
//	SSH 连接失败: dial tcp :22: connectex: No connection could be made ...
//
// 这不是"那台服务器连不上",是**配置压根没选** —— 而 probe 恰恰是文档化的自检命令
// (internal/cli/debug/README.md 的「自检」一行),用来判断调试链路通不通。一条恒失败的自检
// 比没有自检更费时间:实测有人先被它带去查网络,真正的毛病却在这行代码里。
//
// 所以这条用例不看内部字段,只看命令自己打印出来的那一行 —— 谁去掉 ApplyDefaultEnv,它就红。

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tt/internal/cli/common"
	"tt/internal/testkit"
)

// probe 报出的服务器与区域必须是配置里的那一份(而不是零值)。
func TestDBGProbeReportsConfiguredHost(t *testing.T) {
	oldPath := common.ConfigPath
	common.ConfigPath = writeProbeConfig(t)
	t.Cleanup(func() { common.ConfigPath = oldPath })

	oldModule, oldProg := dbgModule, dbgProg
	dbgModule, dbgProg = "asf", "bsft001_wf"
	t.Cleanup(func() { dbgModule, dbgProg = oldModule, oldProg })

	_, out := testkit.CaptureStdout(t, func() int {
		// 连不上是这一条的**预期**:端口 1 上不会有 SSH。这里要的是"连哪台",不是"连得上"。
		_ = debugProbeCmd.RunE(debugProbeCmd, nil)
		return 0
	})

	if !strings.Contains(out, "[probe] 服务器 127.0.0.1:1 区域 36") {
		t.Fatalf("probe 没按配置里的环境连。\n漏掉 ApplyDefaultEnv 的表现是 `[probe] 服务器 :22 区域 `(主机空、区域空、端口回落成默认的 22):\n%s", out)
	}
}

// writeProbeConfig 落一份**只有 hosts.sshs 的统一配置**(schemaVersion 3)——
// 就是设置页写出来的那种形状,也是这条缺陷从前藏身的地方:顶层没有 ssh 字段,
// 不合并环境就什么都拿不到。
//
// 主机填 127.0.0.1、端口填 1:必定被立即拒绝,用例不必等超时,也绝不会真的连上做什么。
func writeProbeConfig(t *testing.T) string {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	cfg := `{"schemaVersion":3,"hosts":{"activeEnv":"探针机","sshs":[
	  {"name":"探针机","host":"127.0.0.1","port":1,"user":"tiptop","zone":"36","topent":"99",
	   "db":{"type":"oracle","host":"127.0.0.1","port":1521,"service":"t35prd"}}
	]}}`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	return cfgPath
}
