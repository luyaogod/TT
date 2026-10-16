package debug

// WP4a 的两个新标志：`--session`（会话级目标）与 `--name`（起会话时的别名）。
//
// 为什么落点在这一个包：契约 1. 的**命令角色表**说的是"哪条命令认哪个标志"，
// 而命令树只在这里（`debug.Group`）。`internal/debug` 的契约测试想测它也不行 ——
// 那会是一个 import cycle（`internal/cli/debug` 本来就 import `internal/debug`），
// 所以 `internal/debug/contract_debug_test.go` 里留的是一行指路，断言本体在这儿。
//
// 三条断言逐条对着契约：
//
//  1. **会话级操作命令认 `--session`**（契约 1. 的命令角色表 / C3.1 / Q9）；
//  2. **会话外命令不认它**（C1.2 / C1.3 / C1.5）：status / sql / wslogs / wsdebug /
//     source / resolve / logs / ents / db / probe / serve —— 挂上去会让 F16 那条
//     "标志挂错命令"的自纠提示（`该命令不接受 --session`）落空；
//  3. **`--name` 只在起会话的命令上**（C3.5 / F17）：今天 = `start`。
//
// 另加两条"注册得上"的地基断言：标志名能 Lookup 到、且 help 文案非空 ——
// 后者是 `--help` 里能看见这条规矩的判据（C5.8/C2.5 那些取值纪律要写在文案里）。

import "testing"

// TestCharTargetFlagPlacement 见文件头。
func TestCharTargetFlagPlacement(t *testing.T) {
	acceptSession := []string{
		"start", "exec", "quit", "why", "wait", "mode", "stop", "locate", "interrupt", "env", "topent",
	}
	rejectSession := []string{
		"status", "sql", "wslogs", "wsdebug", "source", "resolve", "logs", "ents", "db", "probe", "serve",
	}
	acceptName := []string{"start"}
	rejectName := []string{
		"exec", "quit", "why", "wait", "mode", "stop", "locate", "interrupt", "env", "topent",
		"status", "sql", "wslogs", "wsdebug", "source", "resolve", "logs", "ents", "db", "probe", "serve",
	}

	// byName 把命令树摊成 名字 → 命令；顺带确认断言表里的名字都还在（改了命令名要一起改这里）。
	byName := map[string]bool{}
	for _, c := range Group.Commands() {
		byName[c.Name()] = true
	}
	hasFlag := func(cmdName, flag string) bool {
		for _, c := range Group.Commands() {
			if c.Name() == cmdName {
				return c.Flags().Lookup(flag) != nil
			}
		}
		return false
	}

	for _, c := range acceptSession {
		if !byName[c] {
			t.Errorf("断言表里的 `tt debug %s` 不在命令树上 —— 命令改名了就把这张表一起改", c)
			continue
		}
		if !hasFlag(c, "session") {
			t.Errorf("`tt debug %s` 该认 --session（契约 1. 的命令角色表：它是会话级操作）", c)
		}
	}
	for _, c := range rejectSession {
		if !byName[c] {
			t.Errorf("断言表里的 `tt debug %s` 不在命令树上 —— 命令改名了就把这张表一起改", c)
			continue
		}
		if hasFlag(c, "session") {
			t.Errorf("`tt debug %s` **不该**认 --session（C1.2/C1.3/C1.5：它的目标不是某一条会话）——\n"+
				"    挂上去会让 F16 那条「标志挂错命令」的自纠提示落空。", c)
		}
	}
	for _, c := range acceptName {
		if !hasFlag(c, "name") {
			t.Errorf("`tt debug %s` 该认 --name（C3.5：只在起会话的命令上有意义）", c)
		}
	}
	for _, c := range rejectName {
		if hasFlag(c, "name") {
			t.Errorf("`tt debug %s` 不该认 --name（F17：挂到非建会话命令上必须是「unknown flag」老路径）", c)
		}
	}

	// 地基：两个标志都真的注册上了，而且 help 文案非空 ——
	// 取值纪律（C2.4/C2.5 的"非空、1–64、只排除 / 与 #"）要能在 --help 里读到。
	for _, c := range Group.Commands() {
		for _, flag := range []string{"session", "name"} {
			f := c.Flags().Lookup(flag)
			if f == nil {
				continue
			}
			if f.Usage == "" {
				t.Errorf("`tt debug %s --%s` 的 help 文案是空的 —— 取值纪律要在 --help 里看得见", c.Name(), flag)
			}
		}
	}
}
