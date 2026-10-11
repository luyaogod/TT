package fep

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tt/internal/testkit"
)

// W4：动作层的两道防线 —— **发送前**（动作必须属于当前交互指令）与**发送后**
//（运行期报错帧 = "作业停下了"，绝不能当成"这一步没生效"）。
//
// 为什么这两条必须有测试：
//   - 发送前那条：今天发到已关闭实例的动作只回 18 字节空 ACK，看起来像"没反应"，
//     我因此反复改参数、反复重试。要能在发之前就说清楚"你发错对象了"。
//   - 发送后那条：今天我把一个容器（Edit/ValueList）写值，把开窗程序**打停**了
//     （FORMS -6327），却完全没看出来 —— 之后十几条事件全是空 ACK。
//
// 夹具 `testdata/fep-recordings/program-stop.txt` 就是那次停摆的现场录制剪出来的。

// foldFixture 折叠一份语料：name 可以是 testdata 下的夹具名，也可以是一个已存在的路径。
func foldFixture(t *testing.T, name string) *Model {
	t.Helper()
	path := name
	if _, err := os.Stat(path); err != nil {
		path = filepath.Join(testkit.RepoRoot(t), "testdata", "fep-recordings", name)
	}
	bodies, err := ReadRecording(path)
	if err != nil {
		t.Fatalf("读语料 %s: %v", name, err)
	}
	m := NewModel()
	for _, b := range bodies {
		m.ApplyMessage(b)
	}
	if m.Len() == 0 {
		return nil // 空壳录制（会话刚起来就死）：交调用方跳过，不当作错误
	}
	return m
}

// TestProgramStopIsDetectedOnThePinnedFixture：停摆帧必须被认出来，并且能报出"哪儿停了"。
func TestProgramStopIsDetectedOnThePinnedFixture(t *testing.T) {
	m := foldFixture(t, "program-stop.txt")

	errs := RuntimeErrors(m)
	if len(errs) == 0 {
		t.Fatal("没认出停摆帧：这等于" +
			"作业已经停了还继续发事件（今天真实发生过，白试了十几步）")
	}
	var stop RuntimeError
	found := false
	for _, e := range errs {
		if e.ProgramStop() {
			stop, found = e, true
			break
		}
	}
	if !found {
		t.Fatalf("认出了 %d 条报错，但没有一条被判为 ProgramStop：%+v", len(errs), errs)
	}
	if got, want := stop.Where(), "qry_q_pmaa001_3.4gl:278 (-6327)"; got != want {
		t.Errorf("定位 = %q，想要 %q", got, want)
	}
	if !strings.Contains(stop.Comment, "ndialog.c") {
		t.Errorf("正文没带上运行库位置：%q", stop.Comment)
	}

	// 体检必须返回"运行失败"（退出码 1），而且不能是 nil。
	err := CheckRuntimeHealthy(m)
	if err == nil {
		t.Fatal("CheckRuntimeHealthy 放过了停摆")
	}
	ref, ok := err.(*Refusal)
	if !ok {
		t.Fatalf("错误类型 = %T，想要 *Refusal", err)
	}
	if ref.Exit != 1 {
		t.Errorf("退出码 = %d，想要 1（运行失败）", ref.Exit)
	}
	if !strings.Contains(ref.Message, "qry_q_pmaa001_3.4gl") {
		t.Errorf("报错里没点出停在哪个程序：%q", ref.Message)
	}
	if ref.Hint == "" {
		t.Error("没给修复建议（重开会话 / 去掉那一步）")
	}
}

// TestHealthyFixturesHaveNoProgramStop 是**反向对照**：正常的会话绝不能被判成停摆 ——
// 否则每跑一步都会"报运行失败"，那这条检查就成了噪声，人会开始忽略它。
func TestHealthyFixturesHaveNoProgramStop(t *testing.T) {
	for _, name := range []string{"openwin-with-dialogs.txt", "id-reuse.txt"} {
		m := foldFixture(t, name)
		for _, e := range RuntimeErrors(m) {
			if e.ProgramStop() {
				t.Errorf("%s：正常会话被判成停摆：%s", name, e.Comment)
			}
		}
		if err := CheckRuntimeHealthy(m); err != nil {
			t.Errorf("%s：CheckRuntimeHealthy 误报：%v", name, err)
		}
	}
}

// TestActionMustBelongToTheCurrentDialog：动作属于**当前**交互指令才放行。
//
// 官方 1514：动作由 Dialog 节点持有；实测运行期每次状态变化都重新分配动作 id，
// 所以"名字对了"不够，必须"这一份是对的"。
func TestActionMustBelongToTheCurrentDialog(t *testing.T) {
	m := NewModel()
	m.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	m.ApplyMessage(`om 1 {{an 178 Dialog 200 {{active "1"}} {{Action 201 {{name "accept"}}}}}}`)
	m.ApplyMessage(`om 2 {{an 178 Dialog 300 {{active "0"}} {{Action 301 {{name "cancel"}}}}}}`)

	if err := CheckActionInCurrentDialog(m, 201); err != nil {
		t.Errorf("当前指令里的动作被拒了：%v", err)
	}
	err := CheckActionInCurrentDialog(m, 301)
	if err == nil {
		t.Fatal("非当前指令的动作被放行了（发出去只会回空 ACK）")
	}
	ref, ok := err.(*Refusal)
	if !ok {
		t.Fatalf("错误类型 = %T，想要 *Refusal", err)
	}
	if ref.Exit != 4 {
		t.Errorf("退出码 = %d，想要 4（会话状态不符）", ref.Exit)
	}

	// 判据不足时**不拦**：模型里一个 active Dialog 都没有（这份录制没抓到 Dialog 帧），
	// 此时任何判断都是猜 —— 宁可不判，也不误拒。
	blank := NewModel()
	blank.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	blank.ApplyMessage(`om 1 {{an 178 Dialog 200 {{active "0"}} {{Action 201 {{name "accept"}}}}}}`)
	if err := CheckActionInCurrentDialog(blank, 201); err != nil {
		t.Errorf("没有判定依据时不该拦：%v", err)
	}
}

// TestFixtureFilesExist 守住"夹具是这条回归的牙"：目录里必须有那三份，
// 否则所有断言都会因为"读不到文件"而变成 t.Fatalf（那还算响），
// 但如果是被人删掉、改名，就会静默少覆盖一类现场 —— 这里钉死。
func TestFixtureFilesExist(t *testing.T) {
	dir := filepath.Join(testkit.RepoRoot(t), "testdata", "fep-recordings")
	for _, name := range []string{"openwin-with-dialogs.txt", "id-reuse.txt", "program-stop.txt"} {
		st, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Errorf("缺夹具 %s：%v", name, err)
			continue
		}
		if st.Size() == 0 {
			t.Errorf("夹具 %s 是空的", name)
		}
	}
}
