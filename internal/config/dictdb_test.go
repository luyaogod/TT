package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDictDBPathIsPerEnv(t *testing.T) {
	dir := t.TempDir()
	a := DictDBPath(dir, "主机109")
	b := DictDBPath(dir, "尚阳通")
	if a == b {
		t.Fatalf("两个环境指向同一个库: %s", a)
	}
	if filepath.Dir(a) != filepath.Join(dir, DictDBDirName) {
		t.Fatalf("库文件应在 %s 下: %s", filepath.Join(dir, DictDBDirName), a)
	}
	if !strings.HasSuffix(a, ".db") || filepath.Base(a) == DefaultDictDBName {
		t.Fatalf("不该用旧的共享库名: %s", a)
	}
}

func TestDictDBFileNameSanitizes(t *testing.T) {
	name := DictDBFileName(`a/b:c*?"<>|`)
	if strings.ContainsAny(name, `/\:*?"<>|`) {
		t.Errorf("文件名仍含非法字符: %q", name)
	}
	if !strings.HasSuffix(name, ".db") {
		t.Errorf("应以 .db 结尾: %q", name)
	}
	if got := DictDBFileName("   "); got != "env.db" {
		t.Errorf("空环境名 = %q, want env.db", got)
	}
}

func TestDictDBCandidatesForEnv(t *testing.T) {
	t.Setenv("TDICT_DB", "D:/tmp/legacy.db")
	l := LocationsAt(filepath.Join(t.TempDir(), "config.json"))
	got := l.DictDBCandidatesFor("主机109", "")
	want := l.DictDBPath("主机109")
	if len(got) != 1 || got[0] != want {
		t.Fatalf("候选 = %v, want [%s]（默认不该回落到 erp_data.db，也不该被 TDICT_DB 覆盖）", got, want)
	}

	flag := filepath.Join(t.TempDir(), "flag.db")
	t.Setenv("TDICT_DB", "")
	got = l.DictDBCandidatesFor("主机109", flag)
	if got[0] != flag {
		t.Fatalf("-d 应优先: %v", got)
	}
}

func TestResolveDictDBForMissing(t *testing.T) {
	t.Setenv("TDICT_DB", "")
	l := LocationsAt(filepath.Join(t.TempDir(), "config.json"))
	if _, tried, err := l.ResolveDictDBFor("主机109", ""); err == nil || len(tried) == 0 {
		t.Fatalf("缺库应报错并列出尝试过的路径: err=%v tried=%v", err, tried)
	}

	p := l.DictDBPath("主机109")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, _, err := l.ResolveDictDBFor("主机109", "")
	if err != nil || got != p {
		t.Fatalf("ResolveDictDBFor = %q, %v; want %q", got, err, p)
	}
}
