package fep

import (
	"os"
	"path/filepath"
	"testing"
)

// 留存策略只有一条规矩：**归属决定留存**（用户 2026-10-15 定）——
// 用户主动录的进长期库、不参与自动清理；系统自动录的留缓存、走统一清理框架。
func TestRetentionFollowsOrigin(t *testing.T) {
	cases := []struct {
		name   string
		origin Origin
		want   Retention
		ask    Retention
	}{
		{"用户主动 → 长期库", OriginUser, RetainLibrary, ""},
		{"系统自动 → 缓存", OriginAuto, RetainCache, ""},
		{"显式要留（哪怕系统录的）", OriginAuto, RetainLibrary, RetainLibrary},
		{"显式要丢（哪怕用户录的）", OriginUser, RetainCache, RetainCache},
	}
	for _, c := range cases {
		if got := retentionFor(c.origin, c.ask); got != c.want {
			t.Errorf("%s：retentionFor(%q,%q) = %q，应为 %q", c.name, c.origin, c.ask, got, c.want)
		}
	}
}

// 默认值落在 fillDefaults 里（一个调用点写反了，症状是"用户主动录的那一轮悄悄被清掉"）。
func TestFillDefaultsDerivesRetention(t *testing.T) {
	user := OpenConfig{Prog: "x", NoJob: true}
	user.fillDefaults()
	if user.Origin != OriginUser || user.Retention != RetainLibrary {
		t.Fatalf("用户主动录的默认该进长期库：%+v", user)
	}
	auto := OpenConfig{Prog: "x", NoJob: true, Origin: OriginAuto}
	auto.fillDefaults()
	if auto.Retention != RetainCache {
		t.Fatalf("系统自动录的默认该留缓存（走统一清理）：%+v", auto)
	}
	noKeep := OpenConfig{Prog: "x", NoJob: true, NoKeep: true}
	noKeep.fillDefaults()
	if noKeep.Retention != RetainCache {
		t.Fatalf("--no-keep 是显式要留缓存：%+v", noKeep)
	}
}

// 删一个环境 = 删它整段（级联清理落地的那一下）。
func TestLibraryRemoveEnvOnlyTouchesThatEnv(t *testing.T) {
	dataDir := t.TempDir()
	cache := filepath.Join(dataDir, "fep")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	lib := NewLibrary(dataDir)
	for _, c := range []struct{ env, ent, id string }{
		{"主机109", "99", "fep-a"},
		{"主机109", "77", "fep-b"},
		{"尚阳通", "1", "fep-c"},
	} {
		seedRecording(t, cache, c.id, c.env, c.ent)
		if _, err := lib.Promote(cache, c.id); err != nil {
			t.Fatal(err)
		}
	}
	files, bytes, err := lib.RemoveEnv("主机109")
	if err != nil {
		t.Fatal(err)
	}
	if files == 0 || bytes == 0 {
		t.Fatalf("该报出删了多少：files=%d bytes=%d", files, bytes)
	}
	if _, err := os.Stat(filepath.Join(dataDir, "records", "主机109")); !os.IsNotExist(err) {
		t.Fatalf("被删环境的整段该没了：%v", err)
	}
	if _, err := os.Stat(filepath.Join(dataDir, "records", "尚阳通", "1", "fep-c")); err != nil {
		t.Fatalf("别的环境不该动：%v", err)
	}
	if envs, err := lib.Envs(); err != nil || len(envs) != 1 || envs[0] != "尚阳通" {
		t.Fatalf("库里的环境清单不对：%v（err=%v）", envs, err)
	}
	// 没录过的环境：不是错，删 0 个
	if f, _, err := lib.RemoveEnv("没录过的环境"); err != nil || f != 0 {
		t.Fatalf("没录过的环境该是空操作：files=%d err=%v", f, err)
	}
	// 空环境名：拒（不能变成"删库根"）
	if _, _, err := lib.RemoveEnv("  "); err == nil {
		t.Fatal("没给环境名必须拒：那会删掉整座库")
	}
}

// 删一份录制只认**库里**的那一份：传别的路径进来不猜着删。
func TestLibraryRemoveRefusesPathsOutsideTheRoot(t *testing.T) {
	dataDir := t.TempDir()
	cache := filepath.Join(dataDir, "fep")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	lib := NewLibrary(dataDir)
	seedRecording(t, cache, "fep-a", "主机109", "99")
	item, err := lib.Promote(cache, "fep-a")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := lib.Remove(LibraryItem{ID: "fep-a", Dir: cache}); err == nil {
		t.Fatal("缓存目录不在库里，必须拒")
	}
	if _, _, err := lib.Remove(LibraryItem{ID: "fep-a", Dir: lib.Root}); err == nil {
		t.Fatal("库根不是一份录制，必须拒")
	}
	if _, _, err := lib.Remove(LibraryItem{ID: "fep-a"}); err == nil {
		t.Fatal("没有落点时必须拒（不猜着删）")
	}
	files, bytes, err := lib.Remove(item)
	if err != nil || files == 0 || bytes == 0 {
		t.Fatalf("删库里那一份该成功并报数：files=%d bytes=%d err=%v", files, bytes, err)
	}
	if _, err := os.Stat(item.Dir); !os.IsNotExist(err) {
		t.Fatalf("删完该没了：%v", err)
	}
}

// 身份文件里要写下归属与留存：以后看一份录像为什么会在缓存里被清掉，答案就在那一栏。
func TestMetaCarriesOriginAndRetention(t *testing.T) {
	dir := t.TempDir()
	if _, err := WriteMeta(dir, RecordingMeta{
		ID: "fep-x", Origin: OriginAuto, Retention: RetainCache,
	}); err != nil {
		t.Fatal(err)
	}
	mt, err := ReadMeta(dir, "fep-x")
	if err != nil {
		t.Fatal(err)
	}
	if mt.Origin != OriginAuto || mt.Retention != RetainCache {
		t.Fatalf("归属/留存没落进身份：%+v", mt)
	}
	if mt.Origin.Label() == "" || mt.Retention.Label() == "" {
		t.Fatal("两栏都要能说给人听")
	}
}
