package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tt/internal/dev/model"
	"tt/internal/dev/testutil"
)

// captureStdout 把命令的 stdout 收进字符串（命令里 `line(w, …)` 直接写 os.Stdout）。
//
// **必须边写边读**（2026-09-25 实测）。原来的写法是"先跑完 fn，关掉写端，再读"，注释里写着
// "输出量小，先关写端再读不会阻塞" —— 那个前提有边界，而边界就在 4 KB 左右：进程内管道的缓冲
// 是有限的，`tt dev tzs --help` 在**引擎可达**时会附上动词索引，整份输出 4,574 字节，恰好越过
// 缓冲。于是写端阻塞、而它在等 fn() 返回、fn() 在等写成功 —— 一条测试挂满 10 分钟被 go test
// 判超时，goroutine dump 停在 `Manifest.Index` 的 Fprintf 上。
//
// 这个坑一直存在，只是平时不显形：默认跑测试时引擎不可达（没有 `TT_CONFIG`），索引那段不打印，
// 输出就小。它是被"环境变量恰好把引擎指通了"这件事**偶然**翻出来的。
func captureStdout(t *testing.T, fn func() int) (int, string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("建管道失败: %v", err)
	}
	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	old := os.Stdout
	os.Stdout = w
	code := fn()
	os.Stdout = old
	w.Close()
	out := <-done
	r.Close()
	return code, out
}

// TestCaptureStdoutSurvivesBigOutput 钉住上面那个边界（约 4 KB）。
//
// 判据是"它回来了"—— 而这一条**必须见过它红**：把 captureStdout 退回"先跑完再读"的写法，
// 这条测试不会失败，它会**挂住**（写端阻塞在管道缓冲上，读端在等 fn 返回），最后靠 go test 的
// 超时兜底。所以输出量给到 64 KB，远超任何管道缓冲。
func TestCaptureStdoutSurvivesBigOutput(t *testing.T) {
	const n = 64 * 1024
	code, out := captureStdout(t, func() int {
		for i := 0; i < n; i++ {
			fmt.Fprint(os.Stdout, "x")
		}
		return 0
	})
	if code != 0 {
		t.Errorf("退出码该原样返回 0，得 %d", code)
	}
	if len(out) != n {
		t.Errorf("该收全 %d 字节，得 %d（收不全说明读端没跟上）", n, len(out))
	}
}

// exportFixture 导出一个合成包到临时工作区，返回 (工作区目录, 包路径, 渲染文档字节)。
func exportFixture(t *testing.T, dir string) (string, string, []byte) {
	t.Helper()
	p, err := testutil.NormalPkgPath(dir, "adzi999")
	if err != nil {
		t.Fatal(err)
	}
	ws := filepath.Join(dir, "ws")
	if code := silent(t, func() int { return cmdExport([]string{p, "-o", ws}) }); code != 0 {
		t.Fatalf("export 退出码 %d", code)
	}
	gl := filepath.Join(ws, "prog.full.4gl")
	b, err := os.ReadFile(gl)
	if err != nil {
		t.Fatal(err)
	}
	return ws, p, b
}

// insertAtLineStart 把 marker 插到 off 所在行的行首（off 必须落在行首）。
func insertAtLineStart(t *testing.T, doc []byte, off int, marker string) []byte {
	t.Helper()
	if off <= 0 || off > len(doc) || doc[off-1] != '\n' {
		t.Fatalf("off=%d 不在行首", off)
	}
	out := make([]byte, 0, len(doc)+len(marker))
	out = append(out, doc[:off]...)
	out = append(out, marker...)
	out = append(out, doc[off:]...)
	return out
}

// contentLineStart 返回「从 from 开始的下一行行首」。
func contentLineStart(doc []byte, from int) int {
	i := bytes.IndexByte(doc[from:], '\n')
	if i < 0 {
		return len(doc)
	}
	return from + i + 1
}

// TestApplyReportsReadonlyLine apply 拒绝只读区改动时，必须给出文件:行号 + 该行内容。
func TestApplyReportsReadonlyLine(t *testing.T) {
	dir := t.TempDir()
	ws, _, doc := exportFixture(t, dir)

	// 找到第一个只读区段的围栏行，往它正文的第一行行首插一行
	i := bytes.Index(doc, []byte("[READONLY"))
	if i < 0 {
		t.Fatal("合成包里没有 READONLY 区段？")
	}
	at := contentLineStart(doc, i)
	wantLine := model.LineOf(doc, at)
	marker := "  # tdev pos probe"
	if err := os.WriteFile(filepath.Join(ws, "prog.full.4gl"),
		insertAtLineStart(t, doc, at, marker+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, out := captureStdout(t, func() int { return cmdApply([]string{ws}) })
	if code != 4 {
		t.Fatalf("改动只读区应退出码 4，实际 %d\n输出：\n%s", code, out)
	}
	want := "prog.full.4gl:" + itoa(wantLine)
	if !strings.Contains(out, want) {
		t.Errorf("输出里没有 %q：\n%s", want, out)
	}
	if !strings.Contains(out, marker) {
		t.Errorf("输出里没有该行内容 %q：\n%s", marker, out)
	}
	if !strings.Contains(out, "位置：") {
		t.Errorf("输出里没有「位置：」：\n%s", out)
	}
}

// TestStatusReportsChangedLine status 的改动清单要带行号（改的是哪一行就指哪一行）。
func TestStatusReportsChangedLine(t *testing.T) {
	dir := t.TempDir()
	ws, _, doc := exportFixture(t, dir)

	// 往第一个可编辑**点**的正文首行行首插一行注释
	i := bytes.Index(doc, []byte("[EDITABLE "))
	if i < 0 {
		t.Fatal("合成包里没有可编辑点？")
	}
	at := contentLineStart(doc, i)
	wantLine := model.LineOf(doc, at)
	marker := "   # tdev status probe"
	if err := os.WriteFile(filepath.Join(ws, "prog.full.4gl"),
		insertAtLineStart(t, doc, at, marker+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, out := captureStdout(t, func() int { return cmdStatus([]string{ws}) })
	if code != 0 {
		t.Fatalf("status 退出码 %d（只读点正文应可改）\n%s", code, out)
	}
	if !strings.Contains(out, "（prog.full.4gl:"+itoa(wantLine)+"）") {
		t.Errorf("改动清单没有带行号（期望第 %d 行）：\n%s", wantLine, out)
	}
}

// TestVerifyReportsPosition verify 也要打印位置（全量视图）。
func TestVerifyReportsPosition(t *testing.T) {
	dir := t.TempDir()
	ws, _, doc := exportFixture(t, dir)
	i := bytes.Index(doc, []byte("[READONLY"))
	if i < 0 {
		t.Fatal("合成包里没有 READONLY 区段？")
	}
	at := contentLineStart(doc, i)
	wantLine := model.LineOf(doc, at)
	if err := os.WriteFile(filepath.Join(ws, "prog.full.4gl"),
		insertAtLineStart(t, doc, at, "   # tdev verify probe\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out := captureStdout(t, func() int { return cmdVerify([]string{ws}) })
	if code != 4 {
		t.Fatalf("verify 退出码 %d（只读区被改应 4）\n%s", code, out)
	}
	if !strings.Contains(out, "prog.full.4gl:"+itoa(wantLine)) {
		t.Errorf("verify 输出没有位置（期望第 %d 行）：\n%s", wantLine, out)
	}
}

// itoa 只为拼字符串，避免引入 strconv 与上面的 import 同时挤在一处。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
