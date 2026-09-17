package debug

import (
	"testing"

	"tt/internal/host"
)

// 停站看门狗的默认值。配置里没写这个键时取 DefaultWatchdogSeconds ——
// 这里写死 1800 而**不引那个常量**,要钉住的正是"这个数是多少":
// docs/debug.md 一度写 1800 而代码里是 180,引常量的话两边一起漂就没人发现了。
func TestWatchdogDefaultSeconds(t *testing.T) {
	c := &Config{}
	c.applySsh(&host.NamedSsh{Name: "未配看门狗的环境"})
	if c.WatchdogSeconds != 1800 {
		t.Errorf("环境没覆盖时看门狗 = %d 秒,期望 1800", c.WatchdogSeconds)
	}
}

// 每环境的覆盖项优先于默认值;覆盖为 0(未配)不把默认值抹掉 ——
// applySsh 先合覆盖项、再 fillDefaults,顺序反了这里就会露馅。
func TestWatchdogEnvOverride(t *testing.T) {
	c := &Config{}
	c.applySsh(&host.NamedSsh{Name: "覆盖过的环境", WatchdogSeconds: 654})
	if c.WatchdogSeconds != 654 {
		t.Errorf("环境覆盖应生效:期望 654,得到 %d", c.WatchdogSeconds)
	}
}
