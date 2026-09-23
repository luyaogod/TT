package config

import (
	"path/filepath"
	"testing"
)

// TestSkeletonHasNoDesignerDir 钉住这次改动：设计器目录不再是配置项。
//
// 设计器的程序集现在随包分发（见 engine/src/Designer/Bootstrap.cs 的 Install，默认是
// `<引擎自己的目录>\designer`），所以 config.json 里不该再出现 installDir。
//
// 这条断言是防止它被"顺手加回来"的：一旦它回来，同一份 tt 装在两台机器上就会跑两版设计器，
// 而现象是同一个 .tzs 行为不同、报错里却看不出来 —— 这正是把版本钉进包要消掉的那件事。
func TestSkeletonHasNoDesignerDir(t *testing.T) {
	sk := NewSkeleton()
	sec, ok := sk["tzs"].(map[string]any)
	if !ok {
		t.Fatalf("骨架里没有 tzs 节：%#v", sk["tzs"])
	}
	if _, bad := sec["installDir"]; bad {
		t.Error("tzs 节里不该有 installDir —— 设计器程序集随包分发，不是配置项")
	}
	for _, want := range []string{"workspace", "serverExe"} {
		if _, ok := sec[want]; !ok {
			t.Errorf("tzs 节少了 %s", want)
		}
	}
}

// TestTzsStatusDesignerFollowsEngine 钉住 Designer 的位置规则：跟着**生效的**引擎 exe 走。
//
// 引擎自己就是这么算的（`<自己所在目录>\designer`）。两边各算一次的话，设置页会报"有"
// 而引擎报"没有" —— 那是最难查的一类不一致。
func TestTzsStatusDesignerFollowsEngine(t *testing.T) {
	def := filepath.Join(t.TempDir(), "tzs", "tzs-server.exe")

	got := TzsStatusOf(TzsSettings{}, def).Designer.Dir
	if want := AbsPath(filepath.Join(filepath.Dir(def), "designer")); got != want {
		t.Errorf("Designer 缺省位置 = %q，想要 %q", got, want)
	}

	// serverExe 覆盖之后，Designer 要挪到覆盖后那份的旁边，而不是留在缺省旁边。
	other := filepath.Join(t.TempDir(), "elsewhere", "tzs-server.exe")
	got = TzsStatusOf(TzsSettings{ServerExe: other}, def).Designer.Dir
	if want := AbsPath(filepath.Join(filepath.Dir(other), "designer")); got != want {
		t.Errorf("覆盖 serverExe 后 Designer = %q，想要 %q", got, want)
	}

	// 引擎 exe 都算不出来时不硬凑一个路径（DirStatus 空 = 未配置）。
	if got := TzsStatusOf(TzsSettings{}, "").Designer.Dir; got != "" {
		t.Errorf("没有引擎 exe 时 Designer 该为空，得到 %q", got)
	}
}
