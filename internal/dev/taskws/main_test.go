package taskws

// TestMain：本包的 Export 会在工作区里跑 `git init/commit` 做**版本留痕** —— 这是
// 可选能力（store 里 git 缺失时降级为"不建版本库，工作区仍可用"）。单测不依赖留痕：
// 统一把 git 从 PATH 里拿掉，换来两件事：
//
//  1. 去掉"测试机有没有 git / git 多快"这个隐性环境依赖（CI 上更可重复）；
//  2. 省掉几十次进程启动 —— 本包测试原本 80s+，去掉后约 30s（验收上限是 120s）。
//
// "git 缺失时怎么降级"这条生产行为并没有因此漏测：TestExportWarnings_GitMissing
// 显式构造了它（而且此时 PATH 本来就是空的，断言同样成立）。

import (
	"os"
	"testing"
)

// TestMain 清掉 PATH 后跑全部用例。
func TestMain(m *testing.M) {
	_ = os.Setenv("PATH", "")
	os.Exit(m.Run())
}
