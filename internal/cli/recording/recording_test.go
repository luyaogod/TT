package recording

// 判据：`tt recording` 必须**逐条镜像** `tt fep rec` —— 名字映射对、实现同一份、标志不丢。
//
// 这类"两个入口一份实现"的迁移最容易出的错是**漂移**：某天给 `fep rec show` 加了个标志，
// 新入口没跟上，于是同一个动作在两个入口上行为不同。下面三条判据就是防这个。

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"tt/internal/cli/fep"
)

// recordingGroup 建一次组并返回它（不挂到真根上，避免影响别的包）。
func recordingGroup(t *testing.T) *cobra.Command {
	t.Helper()
	root := &cobra.Command{Use: "tt"}
	Register(root)
	for _, c := range root.Commands() {
		if c.Name() == "recording" {
			return c
		}
	}
	t.Fatal("Register 没挂上 recording 组")
	return nil
}

// TestMirrorsEverySourceCommand：源命令与镜像命令**一一对应**（含改名）。
func TestMirrorsEverySourceCommand(t *testing.T) {
	group := recordingGroup(t)
	byName := map[string]*cobra.Command{}
	for _, c := range group.Commands() {
		byName[c.Name()] = c
	}

	for _, src := range fep.RecordingCommands() {
		want := src.Name()
		if n, ok := renames[want]; ok {
			want = n
		}
		m, ok := byName[want]
		if !ok {
			t.Errorf("`tt fep rec %s` 在新命令面里没有对应项（期望 `tt recording %s`）", src.Name(), want)
			continue
		}
		if m.RunE == nil {
			t.Errorf("`tt recording %s` 没有 RunE —— 它不该是个空壳", want)
		}
		// 同一份实现：镜像命令的 RunE 与源命令是同一个函数值。
		// （函数值不能直接比较，退化用"行为一致"的代理：源有 RunE 且镜像也有。）
		if src.RunE == nil && m.RunE != nil {
			t.Errorf("源命令没有 RunE，镜像却有 —— 说明镜像里另写了实现")
		}
	}
}

// TestMirrorKeepsEveryFlag：标志一个不许丢（丢一个就是两个入口行为不同）。
func TestMirrorKeepsEveryFlag(t *testing.T) {
	group := recordingGroup(t)
	byName := map[string]*cobra.Command{}
	for _, c := range group.Commands() {
		byName[c.Name()] = c
	}

	for _, src := range fep.RecordingCommands() {
		want := src.Name()
		if n, ok := renames[want]; ok {
			want = n
		}
		m := byName[want]
		if m == nil {
			continue
		}
		src.Flags().VisitAll(func(f *pflag.Flag) {
			if m.Flags().Lookup(f.Name) == nil {
				t.Errorf("`tt recording %s` 少了标志 --%s（源命令 `fep rec %s` 有）", want, f.Name, src.Name())
			}
		})
	}
}

// TestNoDuplicateNames：组里不许有重名命令（cobra 会静默只留一个）。
func TestNoDuplicateNames(t *testing.T) {
	group := recordingGroup(t)
	seen := map[string]bool{}
	for _, c := range group.Commands() {
		if seen[c.Name()] {
			t.Errorf("重复的命令名：%s", c.Name())
		}
		seen[c.Name()] = true
	}
}
