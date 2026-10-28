package tzc

// 标记门禁的单测与命令级验证（见 tag.go 的说明）。
//
// 两条都要有：
//   - checkModTag 的纯函数用例钉住"什么算合法"（格式、uuid、memo、块配对）；
//   - 一条走 cmdApply 的真路径，钉住"没标记就不写包"——纯函数绿了但没接进命令，
//     等于没有门禁。

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tt/internal/dev/model"
	"tt/internal/testkit"
)

func TestCheckModTag(t *testing.T) {
	const memo = "#A12345#01 20261028 By 12345 新增客户信用额度校验\n"
	cases := []struct {
		name   string
		source string
		tag    string
		op     string
		wantOK bool
		want   string // 不通过时，问题里应出现的片段
	}{
		{"正常标记", memo, "A12345#01", "12345", true, ""},
		{"缺 tag", memo, "", "12345", false, "缺少 --tag"},
		{"格式错", memo, "A12345", "12345", false, "格式不对"},
		{"缺工号", memo, "A12345#01", "", false, "缺少 --operator"},
		{"memo 没写", "# 别的注释\n", "A12345#01", "12345", false, "摘要行"},
		{"memo 少了说明", "#A12345#01 20261028 By 12345\n", "A12345#01", "12345", false, "摘要行"},
		{"memo 工号不符", "#A12345#01 20261028 By 99999 说明\n", "A12345#01", "12345", false, "摘要行"},
		{"uuid + agent", "#3f9a1c07b2d84e56 20261028 By agent 说明\n", "3f9a1c07b2d84e56", "agent", true, ""},
		{"uuid 配真工号", "#3f9a1c07b2d84e56 20261028 By 12345 说明\n", "3f9a1c07b2d84e56", "12345", false, "agent"},
		{"块未闭合", memo + "#A12345#01 -s\n...\n", "A12345#01", "12345", false, "没有闭合"},
		{"多余的 -e", memo + "#A12345#01 -e\n", "A12345#01", "12345", false, "没有对应的 `-s`"},
		{"块配对", memo + "#A12345#01 -s\n...\n#A12345#01 -e\n", "A12345#01", "12345", true, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := checkModTag([]byte(c.source), c.tag, c.op)
			if c.wantOK {
				if len(got) != 0 {
					t.Fatalf("应通过，却报：%v", got)
				}
				return
			}
			if len(got) == 0 {
				t.Fatal("应被拒绝，却通过了")
			}
			joined := strings.Join(got, "\n")
			if !strings.Contains(joined, c.want) {
				t.Errorf("问题里应含 %q，实际：%s", c.want, joined)
			}
		})
	}
}

// TestApplyRequiresModTag 走命令真路径：没标记必须拒（退出码 3）且包一个字节都不动。
func TestApplyRequiresModTag(t *testing.T) {
	dir := t.TempDir()
	ws, pkgPath, _ := exportFixture(t, dir)
	sha0, err := model.Sha256File(pkgPath)
	if err != nil {
		t.Fatal(err)
	}

	code, out := testkit.CaptureStdout(t, func() int { return cmdApply([]string{ws}) })
	if code != 3 {
		t.Fatalf("没标记的 apply 应退出 3，实际 %d\n%s", code, out)
	}
	if !strings.Contains(out, "缺少 --tag") || !strings.Contains(out, "缺少 --operator") {
		t.Errorf("拒绝文案应同时点出 tag 与工号：\n%s", out)
	}
	sha1, _ := model.Sha256File(pkgPath)
	if sha0 != sha1 {
		t.Error("被标记门禁拒绝的 apply 不应改包")
	}

	// 有标记但没写进 memo：仍然拒绝，并指出摘要行
	code, out = testkit.CaptureStdout(t, func() int {
		return cmdApply([]string{"--tag", "A12345#01", "--operator", "12345", ws})
	})
	if code != 3 {
		t.Fatalf("memo 缺失应退出 3，实际 %d\n%s", code, out)
	}
	if !strings.Contains(out, "摘要行") {
		t.Errorf("应提示补 global.memo 摘要行：\n%s", out)
	}
	sha2, _ := model.Sha256File(pkgPath)
	if sha0 != sha2 {
		t.Error("memo 缺失被拒的 apply 不应改包")
	}

	// --no-tag 是给合成/测试包的逃生门：没有标记也能走完管线。
	code, out = testkit.CaptureStdout(t, func() int { return cmdApply([]string{"--no-tag", ws}) })
	if code != 0 {
		t.Fatalf("--no-tag 的 apply 应成功，实际 %d：%s", code, out)
	}
}

// TestApplyAcceptsMemoTag 正向路径：memo 里写了标记，apply 放行并真的写回。
func TestApplyAcceptsMemoTag(t *testing.T) {
	dir := t.TempDir()
	ws, pkgPath, _ := exportFixture(t, dir)
	// global.memo 落在框架区段里：按技能「不能硬改只读区段」的要求，先显式解锁。
	if code := testkit.Silent(t, func() int { return cmdUnlock([]string{"--yes", ws}) }); code != 0 {
		t.Fatalf("unlock 退出码 %d（测试前置失败）", code)
	}
	gl := filepath.Join(ws, "prog.full.4gl")
	doc, err := os.ReadFile(gl)
	if err != nil {
		t.Fatal(err)
	}

	// 在 global.memo 的 add-point 起始行之后插入摘要行。
	lines := strings.Split(string(doc), "\n")
	at := -1
	for i, ln := range lines {
		if strings.Contains(ln, "global.memo") {
			at = i
			break
		}
	}
	if at < 0 {
		t.Skip("合成包里没有 global.memo add-point，无法验证 memo 正向路径")
	}
	memo := "#A12345#01 20261028 By 12345 新增客户信用额度校验"
	lines = append(lines[:at+1], append([]string{memo}, lines[at+1:]...)...)
	if err := os.WriteFile(gl, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}

	sha0, _ := model.Sha256File(pkgPath)
	code, out := testkit.CaptureStdout(t, func() int {
		return cmdApply([]string{"--tag", "A12345#01", "--operator", "12345", ws})
	})
	if code != 0 {
		t.Fatalf("带合法标记的 apply 应成功，实际 %d：%s", code, out)
	}
	sha1, _ := model.Sha256File(pkgPath)
	if sha0 == sha1 {
		t.Error("apply 成功却没改包：这次修改可能没落盘")
	}
}
