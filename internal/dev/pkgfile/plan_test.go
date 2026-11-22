package pkgfile

// Plan 的"改没改"判定：apply 每次都会把 tap/tgl 传进来（哪怕视图没动），
// 所以不能按"传了就改写"标记，否则 --dry-run 永远显示「改写 3.6 MB → 3.6 MB」，
// 让人以为工具在乱写（实测被两个试用者抓到）。

import (
	"os"
	"path/filepath"
	"testing"
)

func writePlanPkg(t *testing.T, dir string, tap, tgl, four []byte) string {
	t.Helper()
	path := filepath.Join(dir, "p(c).tzc")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteZip(f, []WriteEntry{
		{Name: "p.4gl", Data: four},
		{Name: "p.tap", Data: tap},
		{Name: "p.tgl", Data: tgl},
		{Name: "ver", Data: []byte("1.0\n"), Store: true},
	}); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPlan_IdenticalBytesAreNotChanged(t *testing.T) {
	dir := t.TempDir()
	tap := []byte(`<add_points prog="p" identity="c"><section id="p.a" type="s"></section></add_points>`)
	tgl := []byte("{<section id=\"p.a\" type=\"s\" >}\n{</section>}\n")
	path := writePlanPkg(t, dir, tap, tgl, []byte("MAIN\n"))

	pkg, err := Open(path, OpenOptions{AcceptVer: "1.0"})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	actions, err := pkg.Plan(Rebuild{Tap: tap, Tgl: tgl})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	for _, a := range actions {
		if a.Changed {
			t.Errorf("%s 内容没变却标成改写：%d → %d", a.Name, a.OldSize, a.NewSize)
		}
	}

	// 真改一个字节：必须报改写，且 sha 跟着变
	tap2 := append([]byte(nil), tap...)
	tap2[1] = 'X'
	actions, err = pkg.Plan(Rebuild{Tap: tap2})
	if err != nil {
		t.Fatalf("Plan(改): %v", err)
	}
	found := false
	for _, a := range actions {
		if a.Name != "p.tap" {
			continue
		}
		found = true
		if !a.Changed {
			t.Error("改了内容却没标改写")
		}
		if a.NewSha256 == a.OldSha256 {
			t.Error("新 sha 与旧 sha 相同")
		}
	}
	if !found {
		t.Error("结果里没有 p.tap")
	}
}
