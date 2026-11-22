package task

// apply / verify / relink / unlock 这四个动词本身是 tzc 线的直通（`tzc.Run`），
// 工作区命令层在这里只做三件自己的事：
//
//  1. 把工作区定位成 `<工作区>/ws` 再交出去（地址口径统一）；
//  2. 在交出去**之前**做自己的闸门：apply 的 RebindPackage、verify 的"包被换过"拒绝、
//     relink 的脏工作区拒绝；
//  3. 把 tzc 的退出码原样透传（exitCode），并补上"不编译 BDL"这类只在工作区语境下
//     才说得清的人的提醒（stderr，不污染 stdout）。
//
// 所以这些用例断言的是"闸门 + 透传 + 提醒"，而不是重新验一遍 tzc 的管线。

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/dev/testutil"
)

// TestVerifyCommand 钉住 verify 的两种失败形态与成功形态。
func TestVerifyCommand(t *testing.T) {
	t.Run("干净工作区通过", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, true) // 人读：tzc 自己打结论
		code, out, errOut, err := runCmd(t, newVerifyCmd, "--dir", f.root)
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）\nstdout:\n%s\nstderr:\n%s", code, err, out, errOut)
		}
		if !strings.Contains(out, "结论") || !strings.Contains(out, "通过") {
			t.Errorf("verify 应打结论：\n%s", out)
		}
		// 这句是**命令层**补的：tzc 自己不知道"工作区语境下 BDL 没编译"。
		if !strings.Contains(errOut, "不编译 BDL") {
			t.Errorf("stderr 应提醒「只查围栏结构、不编译 BDL」：\n%s", errOut)
		}
	})

	t.Run("包被换过拒绝退4", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, true)
		replacePkg(t, f, testutil.NormalEntriesWith(fxProg, map[string]string{"env": "s"}))
		code, _, _, verr := runCmd(t, newVerifyCmd, "--dir", f.root)
		if code != exitRefused {
			t.Fatalf("退出码 = %d，期望 4（err=%v）", code, verr)
		}
		if verr == nil || !strings.Contains(verr.Error(), "verify 已拒绝") {
			t.Errorf("错误文案应点名 verify 拒绝并指路 relink：%v", verr)
		}
	})

	t.Run("strict把warn当失败退3", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, true) // 人读：tzc 的 verify 只在人读模式下打"结论"
		code, out, _, verr := runCmd(t, newVerifyCmd, "--dir", f.root, "--strict")
		if code != 3 {
			t.Fatalf("--strict 退出码 = %d，期望 3（err=%v）\\n%s", code, verr, out)
		}
		if !strings.Contains(out, "失败") {
			t.Errorf("--strict 下结论应是失败：\\n%s", out)
		}
	})

	t.Run("JSON输出", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, false) // JSON
		code, out, _, err := runCmd(t, newVerifyCmd, "--dir", f.root)
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		env := decodeEnvelope(t, out)
		if ok, _ := env["ok"].(bool); !ok {
			t.Errorf("verify --json 的 ok 应为 true：%v", env)
		}
		if env["prog"] != fxProg {
			t.Errorf("--json 应回显程序名：%v", env["prog"])
		}
	})
}

// TestApplyCommand 钉住 apply 的默认参数拼装（--dry-run/--yes/--json → tzc）、
// 真落盘通路，以及 RebindPackage 闸门。
func TestApplyCommand(t *testing.T) {
	t.Run("dry-run不落盘", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, true)
		before, err := os.ReadFile(f.pkg)
		if err != nil {
			t.Fatal(err)
		}
		code, _, errOut, aerr := runCmd(t, newApplyCmd, "--dir", f.root, "--dry-run")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）\n%s", code, aerr, errOut)
		}
		after, err := os.ReadFile(f.pkg)
		if err != nil {
			t.Fatal(err)
		}
		if string(before) != string(after) {
			t.Error("--dry-run 不该改包")
		}
	})

	t.Run("真落盘并提醒不编译BDL", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, true)
		code, _, errOut, aerr := runCmd(t, newApplyCmd, "--dir", f.root, "--yes")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）\n%s", code, aerr, errOut)
		}
		if !strings.Contains(errOut, "不编译 BDL") {
			t.Errorf("stderr 应提醒「apply 不编译 BDL、语法错要到设计器 F7 才暴露」：\n%s", errOut)
		}
		if _, err := os.Stat(f.pkg); err != nil {
			t.Errorf("apply 之后包应还在: %v", err)
		}
	})

	t.Run("JSON输出", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, false) // JSON：命令层要把 --json 透传给 tzc
		code, out, _, aerr := runCmd(t, newApplyCmd, "--dir", f.root, "--yes")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, aerr)
		}
		env := decodeEnvelope(t, out)
		if ok, _ := env["ok"].(bool); !ok {
			t.Errorf("apply --json 的 ok 应为 true：%v", env)
		}
	})
	t.Run("工作区被移动时把包路径绑回本地", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, true)
		// 只挪清单里的路径、不动包内容：这是"工作区被复制/移动"的形态。
		// apply 必须先把它绑回当前工作区（否则会静默写进原目录的包，实测事故）。
		repointWsManifest(t, f)
		code, _, errOut, aerr := runCmd(t, newApplyCmd, "--dir", f.root, "--yes")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）\\n%s", code, aerr, errOut)
		}
		if !strings.Contains(errOut, "已改到当前工作区") {
			t.Errorf("stderr 必须告诉人「这次写的是哪个包」：\\n%s", errOut)
		}
	})

	t.Run("包与基线不一致拒绝退4", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, false)
		// 任务层 apply 的闸门是 RebindPackage：清单里的源包路径指向别处 + 本地包 sha 对不上。
		breakPkgBinding(t, f)
		code, _, _, aerr := runCmd(t, newApplyCmd, "--dir", f.root)
		if code != exitRefused {
			t.Fatalf("退出码 = %d，期望 4（err=%v）", code, aerr)
		}
		if aerr == nil || !strings.Contains(aerr.Error(), "包与视图基线不一致") {
			t.Errorf("错误文案应说明包与基线不一致：%v", aerr)
		}
	})

	t.Run("包被原地替换拒绝退4", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, false)
		// 只换内容、不动路径与视图清单：RebindPackage 只比路径会放过，靠 tzc apply
		// 的"源包自 export 之后已被改动"拦住。这类替换同样该退 4（要 relink/放回基线包），
		// 不该退 5 让用户看到一个不指向出路的一般失败。
		replacePkg(t, f, testutil.NormalEntriesWith(fxProg, map[string]string{"env": "s"}))
		code, out, stderr, aerr := runCmd(t, newApplyCmd, "--dir", f.root)
		if code != exitRefused {
			t.Fatalf("退出码 = %d，期望 4（err=%v）", code, aerr)
		}
		// 直通动词的失败文案已经由 tzc 线打出去了（stdout/stderr），这里查文案里有没有出路。
		if !strings.Contains(out+stderr, "relink") {
			t.Errorf("输出应指出 relink 出路：stdout=%q stderr=%q", out, stderr)
		}
	})

	t.Run("非工作区退2", func(t *testing.T) {
		isolateConfig(t, false)
		setFormat(t, false)
		t.Chdir(t.TempDir())
		code, _, _, err := runCmd(t, newApplyCmd)
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2（err=%v）", code, err)
		}
	})
}

// TestRelinkCommand 钉住"包被换过"死局的两条出路：relink 承认当前包为新基线，
// 并且拒绝把未落盘的视图改动悄悄丢掉。
func TestRelinkCommand(t *testing.T) {
	t.Run("解死局", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, true)
		replacePkg(t, f, testutil.NormalEntriesWith(fxProg, map[string]string{"env": "s"}))
		// 换包后 verify 必须先拒绝（假绿灯是最坏的）
		if code, _, _, _ := runCmd(t, newVerifyCmd, "--dir", f.root); code != exitRefused {
			t.Fatalf("换包后 verify 退出码 = %d，期望 4", code)
		}
		code, out, _, rerr := runCmd(t, newRelinkCmd, "--dir", f.root)
		if code != 0 {
			t.Fatalf("relink 退出码 = %d（err=%v）", code, rerr)
		}
		if !strings.Contains(out, "已以当前包重建视图") || !strings.Contains(out, "下一步") {
			t.Errorf("relink 应说明做了什么、下一步是什么：\n%s", out)
		}
		// 死局解开：verify 恢复通过
		if code, _, _, verr := runCmd(t, newVerifyCmd, "--dir", f.root); code != 0 {
			t.Fatalf("relink 之后 verify 退出码 = %d（err=%v）", code, verr)
		}
	})

	t.Run("脏工作区拒绝", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, true)
		edited := filepath.Join(f.ws, "prog.full.4gl")
		b, err := os.ReadFile(edited)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(edited, append(b, []byte("\n# 未落盘\n")...), 0o644); err != nil {
			t.Fatal(err)
		}
		code, _, _, rerr := runCmd(t, newRelinkCmd, "--dir", f.root)
		if code != exitRefused {
			t.Fatalf("退出码 = %d，期望 4（err=%v）", code, rerr)
		}
		if rerr == nil || !strings.Contains(rerr.Error(), "未落盘的改动") {
			t.Errorf("错误文案应说明会丢改动：%v", rerr)
		}
		// --force 是用户显式确认"改动能丢"
		code, out, _, ferr := runCmd(t, newRelinkCmd, "--dir", f.root, "--force")
		if code != 0 {
			t.Fatalf("--force 后退出码 = %d（err=%v）", code, ferr)
		}
		if !strings.Contains(out, "已以当前包重建视图") {
			t.Errorf("--force 应重导视图：\n%s", out)
		}
	})

	t.Run("JSON输出", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, false) // JSON
		code, out, _, err := runCmd(t, newRelinkCmd, "--dir", f.root)
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		// relink 的 JSON 载荷是 RelinkResult（字段没有 json tag，所以键是首字母大写），
		// 它被 emitJSON 包在信封的 data 里。
		d := dataOf(t, decodeEnvelope(t, out))
		if d["PkgSha"] == nil || d["PkgSha"] == "" {
			t.Errorf("JSON 应给出新基线 sha：%v", d)
		}
		if num(t, d["Regions"]) == 0 {
			t.Errorf("Regions 应为正：%v", d["Regions"])
		}
	})
}

// TestUnlockCommand 钉住框架解锁的三条闸门：
// 客制环境要 --yes（单向、有代价）、已解锁是 no-op、标准环境要看 adzi052 授权。
func TestUnlockCommand(t *testing.T) {
	t.Run("未确认退4", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, true)
		code, out, _, uerr := runCmd(t, newUnlockCmd, "--dir", f.root)
		if code != exitRefused {
			t.Fatalf("退出码 = %d，期望 4（err=%v）", code, uerr)
		}
		if !strings.Contains(out, "拒绝解锁") || !strings.Contains(out, "需要二次确认") {
			t.Errorf("未加 --yes 应打拒绝与原因：\n%s", out)
		}
	})

	t.Run("确认后解锁并说明pending", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, true)
		code, out, _, uerr := runCmd(t, newUnlockCmd, "--dir", f.root, "--yes")
		if code != 0 {
			t.Fatalf("--yes 退出码 = %d（err=%v）\n%s", code, uerr, out)
		}
		if !strings.Contains(out, "已解开框架") || !strings.Contains(out, "不可逆") {
			t.Errorf("解锁成功应说明已做什么并提醒不可逆：\n%s", out)
		}
		// 刚解锁、还没 apply 时是 pending_unlock：再跑一次要明说"待落盘"，
		// 而不是静默地又重渲染一遍（用户得知道 apply 之前 section_flag 还是 N）。
		code, out, _, uerr = runCmd(t, newUnlockCmd, "--dir", f.root, "--yes")
		if code != 0 {
			t.Fatalf("二次 unlock 退出码 = %d（err=%v）", code, uerr)
		}
		if !strings.Contains(out, "待落盘的解锁态") {
			t.Errorf("pending_unlock 阶段应明说待落盘：\n%s", out)
		}
	})

	t.Run("包本来就是解开态是no-op", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, true)
		// 包根 section_flag="Y"：设计器分支①直接放行，工作区状态已是 Unlocked 且无 pending。
		replacePkg(t, f, testutil.NormalEntriesWith(fxProg, map[string]string{"section_flag": "Y"}))
		reexportWsFromPkg(t, f)
		code, out, _, uerr := runCmd(t, newUnlockCmd, "--dir", f.root, "--yes")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）\n%s", code, uerr, out)
		}
		if !strings.Contains(out, "本来就是解开态") {
			t.Errorf("已是解开态应是 no-op 并说清楚：\n%s", out)
		}
	})

	t.Run("JSON闸门", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, false) // JSON
		code, out, _, uerr := runCmd(t, newUnlockCmd, "--dir", f.root)
		if code != exitRefused {
			t.Fatalf("退出码 = %d，期望 4（err=%v）", code, uerr)
		}
		env := decodeEnvelope(t, out)
		if ok, _ := env["ok"].(bool); ok {
			t.Error("拒绝解锁的 JSON 里 ok 必须是 false")
		}
		if env["stage"] != "unlock-gate" {
			t.Errorf("stage = %v，期望 unlock-gate（机器可读的闸门名）", env["stage"])
		}
		if need, _ := env["need_yes"].(bool); !need {
			t.Errorf("need_yes 应为 true：%v", env)
		}
	})

	t.Run("标准环境无授权退4", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, true)
		// env=s + 非 topstd + std_section_verify=N → 设计器分支②：无 adzi052 授权，--yes 也拒。
		replacePkg(t, f, testutil.NormalEntriesWith(fxProg, map[string]string{
			"env": "s", "login_user": "tiptop", "std_section_verify": "N",
		}))
		reexportWsFromPkg(t, f)
		code, out, _, uerr := runCmd(t, newUnlockCmd, "--dir", f.root, "--yes")
		if code != exitRefused {
			t.Fatalf("退出码 = %d，期望 4（err=%v）\n%s", code, uerr, out)
		}
		if !strings.Contains(out, "无解开框架授权") {
			t.Errorf("标准环境应点名「没有授权」：\n%s", out)
		}
	})

	t.Run("非工作区退2", func(t *testing.T) {
		isolateConfig(t, false)
		setFormat(t, false)
		t.Chdir(t.TempDir())
		code, _, _, err := runCmd(t, newUnlockCmd)
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2（err=%v）", code, err)
		}
	})
}
