package file

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/luyaogod/TT/internal/config"
)

// mirrorCompare 是 `tt file get` 护栏的判定核心：不发 SFTP，只用远端的 size/mtime
// 比一次镜像文件。镜像保留服务器 mtime，所以这两项对得上就等于"同一份"。
func TestMirrorCompare(t *testing.T) {
	dir := t.TempDir()
	mod := time.Date(2026, 10, 30, 9, 4, 33, 0, time.UTC)

	same := filepath.Join(dir, "same.4gl")
	if err := os.WriteFile(same, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(same, mod, mod); err != nil {
		t.Fatal(err)
	}
	// mtime 只到秒：1s 差属于取整/时钟抖动，仍算同一份。
	near := filepath.Join(dir, "near.4gl")
	if err := os.WriteFile(near, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(near, mod.Add(time.Second), mod.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(dir, "old.4gl")
	if err := os.WriteFile(old, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(old, mod.Add(-time.Hour), mod.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	sizeDiff := filepath.Join(dir, "size.4gl")
	if err := os.WriteFile(sizeDiff, []byte("hello!!"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(sizeDiff, mod, mod); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		path  string
		size  int64
		mod   time.Time
		want  mirrorState
		wantP string
	}{
		{"未配镜像", "", 5, mod, mirrorNone, ""},
		{"镜像里没有", filepath.Join(dir, "nope.4gl"), 5, mod, mirrorNone, filepath.Join(dir, "nope.4gl")},
		{"同一份", same, 5, mod, mirrorSame, same},
		{"mtime 秒级抖动仍算同一份", near, 5, mod, mirrorSame, near},
		{"mtime 差得多=落后", old, 5, mod, mirrorStale, old},
		{"size 不同=落后", sizeDiff, 5, mod, mirrorStale, sizeDiff},
	}
	for _, c := range cases {
		got, gotPath := mirrorCompare(c.path, c.size, c.mod)
		if got != c.want || gotPath != c.wantP {
			t.Errorf("%s: mirrorCompare(%q)= (%v,%q)，期望 (%v,%q)",
				c.name, c.path, got, gotPath, c.want, c.wantP)
		}
	}
}

// 目录不是文件：镜像里同名的目录不该被当成"已有一份"。
func TestMirrorCompareDirIsNotHit(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub.4gl")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if got, _ := mirrorCompare(sub, 0, time.Now()); got != mirrorNone {
		t.Fatalf("目录应判为 mirrorNone，实得 %v", got)
	}
}

// mirrorFileFor 把 $TOP 内的相对路径映射到"本环境镜像目录/<相对路径>"。
func TestMirrorFileFor(t *testing.T) {
	root := t.TempDir()
	cfg := filepath.Join(root, "config.json")
	if err := os.WriteFile(cfg, []byte(`{"mirror":{"dir":`+quoteJSON(root)+`}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	tr := &target{cfgPath: cfg, ssh: &config.NamedSsh{Name: "正式"}}
	got, ok := tr.mirrorFileFor("com/sub/4gl/s_apmt510.4gl")
	if !ok {
		t.Fatal("配了 mirror.dir 就该给得出镜像路径")
	}
	want := filepath.Join(root, "正式", "com", "sub", "4gl", "s_apmt510.4gl")
	if got != want {
		t.Errorf("mirrorFileFor = %q，期望 %q", got, want)
	}

	// 没配镜像目录 → 静默跳过护栏。
	empty := filepath.Join(root, "empty.json")
	if err := os.WriteFile(empty, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := (&target{cfgPath: empty, ssh: tr.ssh}).mirrorFileFor("a.4gl"); ok {
		t.Error("未配 mirror.dir 时不该给出镜像路径")
	}
	// 没有 config 路径、或相对路径为空 → 也跳过。
	if _, ok := (&target{ssh: tr.ssh}).mirrorFileFor("a.4gl"); ok {
		t.Error("拿不到 config.json 时不该给出镜像路径")
	}
	if _, ok := tr.mirrorFileFor(""); ok {
		t.Error("空相对路径不该给出镜像路径")
	}
}

// quoteJSON 只为拼一个 Windows 路径进 JSON（反斜杠要转义）。
func quoteJSON(s string) string {
	out := make([]rune, 0, len(s)+2)
	out = append(out, '"')
	for _, r := range s {
		if r == '\\' || r == '"' {
			out = append(out, '\\')
		}
		out = append(out, r)
	}
	out = append(out, '"')
	return string(out)
}
