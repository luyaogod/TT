package form

// 真机取件实现（`internal/host/formsource.go`）的**纯函数**部分。
//
// 为什么不在这里连服务器：本切片所有用例都不碰真机（纪律 5）。于是把"能测的"从
// "不能测的"里切出来 —— 名字白名单（防命令注入/路径穿越）、远端 `find` 命令的拼法、
// `find` 输出的解析、候选唯一性 —— 这些都纯函数化，剩下的只有 `sftp get` 那一步。

import (
	"strings"
	"testing"

	"tt/internal/host"
)

func TestFormNameOK(t *testing.T) {
	yes := []string{"axmt500_wf", "fglfind_tui", "q_imaf001_15", "a.b-c_1", "A1"}
	no := []string{"", "../x", "a/b", `a\b`, "$(id)", ";rm -rf /", "a b", "a;b", "a'b", "x`id`",
		strings.Repeat("a", 65)}
	for _, s := range yes {
		if !host.FormNameOK(s) {
			t.Errorf("FormNameOK(%q) = false, want true（合法表单名）", s)
		}
	}
	for _, s := range no {
		if host.FormNameOK(s) {
			t.Errorf("FormNameOK(%q) = true —— 这个名字会拼进远端命令与路径，必须挡掉", s)
		}
	}
}

func TestFormFindExprRejectsBadNames(t *testing.T) {
	if _, err := host.FormFindExpr("/u1/t35prd", "axm;rm -rf /", "ok"); err == nil {
		t.Error("模块名非法必须报错，不能拼进命令")
	}
	if _, err := host.FormFindExpr("/u1/t35prd", "axm", "$(id)"); err == nil {
		t.Error("表单名非法必须报错")
	}
	expr, err := host.FormFindExpr("/u1/t35prd", "axm", "axmt500_wf")
	if err != nil {
		t.Fatalf("合法输入不该报错: %v", err)
	}
	for _, want := range []string{"/u1/t35prd", "axm", "axmt500_wf.42f", "42f"} {
		if !strings.Contains(expr, want) {
			t.Errorf("find 命令里缺 %q: %s", want, expr)
		}
	}
}

func TestParseFormFindPaths(t *testing.T) {
	out := "/u1/t35prd/erp/axm/axmt500/42f/demo.42f\r\n" +
		"\n" +
		"/u1/t35prd/erp/axm/axmt500/42f/demo.42f\n" +
		"  /u1/t35prd/com/other/42f/demo.42f  \n"
	got := host.ParseFormFindPaths(out)
	if len(got) != 2 {
		t.Fatalf("候选 = %v, want 2 条（去空行/去重）", got)
	}
	if got[0] != "/u1/t35prd/com/other/42f/demo.42f" || got[1] != "/u1/t35prd/erp/axm/axmt500/42f/demo.42f" {
		t.Errorf("解析结果不对（该排序 + 去 CR + 去空白）: %v", got)
	}
}

func TestFormListExpr(t *testing.T) {
	if _, err := host.FormListExpr("/u1/t35prd", "../x"); err == nil {
		t.Error("模块名非法必须报错")
	}
	expr, err := host.FormListExpr("/u1/t35prd/", "axm")
	if err != nil {
		t.Fatalf("合法输入不该报错: %v", err)
	}
	for _, want := range []string{"/u1/t35prd", "axm", "42f", "*.42f"} {
		if !strings.Contains(expr, want) {
			t.Errorf("list 命令里缺 %q: %s", want, expr)
		}
	}
	if strings.Contains(expr, "/u1/t35prd/ ") || strings.Contains(expr, "//") {
		t.Errorf("顶层路径该规整（去掉尾斜杠）: %s", expr)
	}
}

func TestFormNamesFromPaths(t *testing.T) {
	got := host.FormNamesFromPaths([]string{
		"/u1/x/42f/b.42f",
		"/u1/y/42f/a.42f",
		"/u1/z/42f/a.42f",
		"",
	})
	if strings.Join(got, ",") != "a,b" {
		t.Errorf("表单名 = %v, want [a b]（去扩展名、去重、排序）", got)
	}
}

func TestPickFormPath(t *testing.T) {
	if _, err := host.PickFormPath(nil); err == nil {
		t.Error("一个候选都没有必须报错")
	}
	p, err := host.PickFormPath([]string{"/u1/x/42f/demo.42f"})
	if err != nil || p != "/u1/x/42f/demo.42f" {
		t.Errorf("唯一候选该直接返回: %q %v", p, err)
	}
	if _, err := host.PickFormPath([]string{"/u1/a/42f/demo.42f", "/u1/b/42f/demo.42f"}); err == nil {
		t.Error("多候选必须报错并列出（绝不自己挑一个）")
	} else if !strings.Contains(err.Error(), "/u1/a") || !strings.Contains(err.Error(), "/u1/b") {
		t.Errorf("错误里该列出全部候选: %v", err)
	}
}
