package fep

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tt/internal/config"
)

// 长期库（R3）：录制件按 环境+ENT 存放，且**清缓存不碰它**。
// 这条测的正是那个差别：缓存目录是"一轮一删也不心疼"，而录制是"以后还要拿来重放"的东西。

// seedRecording 在 dir 里落一份最小可用的录制（AUI 侧 + 事件侧 + 身份）。
func seedRecording(t *testing.T, dir, id, env, ent string) {
	t.Helper()
	body := "om 3 {{" + demoAUI() + "}}\n"
	for _, n := range []string{id + "-1-rt2gdc.log", id + "-1-gdc2rt.log", id + "-1-timeline.log"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := WriteMeta(dir, RecordingMeta{
		ID: id, Path: PathRunnerRR, Prog: "bsft001_wf", Env: env, Ent: ent,
		Started: time.Now(), Finished: time.Now().Add(time.Minute),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestPromoteMovesOneRecordingIntoEnvAndEnt(t *testing.T) {
	dataDir := t.TempDir()
	cache := filepath.Join(dataDir, "fep")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	seedRecording(t, cache, "fep-20261015-170031", "主机109", "99")
	// 前缀陷阱：另一份的 ID 是这份的前缀延伸，绝不能被一起搬走
	seedRecording(t, cache, "fep-20261015-1700312", "主机109", "99")

	lib := NewLibrary(dataDir)
	item, err := lib.Promote(cache, "fep-20261015-170031")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dataDir, "records", "主机109", "99", "fep-20261015-170031")
	if item.Dir != want {
		t.Fatalf("落点该是 <数据目录>/records/<环境>/<ENT>/<ID>：%s", item.Dir)
	}
	// 三份录制件 + 身份都到了；源里那一份不留
	for _, n := range []string{"-1-rt2gdc.log", "-1-gdc2rt.log", "-1-timeline.log", MetaSuffix} {
		if _, err := os.Stat(filepath.Join(want, "fep-20261015-170031"+n)); err != nil {
			t.Errorf("库里少了 %s: %v", n, err)
		}
		if _, err := os.Stat(filepath.Join(cache, "fep-20261015-170031"+n)); !os.IsNotExist(err) {
			t.Errorf("源里该没有 %s 了（是移动不是拷贝）", n)
		}
	}
	if _, err := os.Stat(filepath.Join(cache, "fep-20261015-1700312-1-rt2gdc.log")); err != nil {
		t.Fatalf("另一份录制被牵连了：%v", err)
	}
	// 库里能按 ID 找回，且带的文件正好是这一份的
	if got, found, err := lib.Find("fep-20261015-170031"); err != nil || !found {
		t.Fatalf("按 ID 找不回：found=%v err=%v", found, err)
	} else if len(got.Files) != 1 || filepath.Base(got.Files[0]) != "fep-20261015-170031-1-rt2gdc.log" {
		t.Fatalf("找回的文件不对：%v", got.Files)
	}
	if items, err := lib.List(); err != nil || len(items) != 1 {
		t.Fatalf("库里该只有搬进去的那一份：%v（err=%v）", items, err)
	}
}

// 没给 ENT 时落 "ENT-未指定"：不知道就不能假装知道（那一轮沿用登录 shell 的 $TOPENT）。
func TestPromoteWithoutEntLandsInUnsetSegment(t *testing.T) {
	dataDir := t.TempDir()
	cache := filepath.Join(dataDir, "fep")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	seedRecording(t, cache, "fep-x", "主机109", "")
	lib := NewLibrary(dataDir)
	item, err := lib.Promote(cache, "fep-x")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dataDir, "records", "主机109", EntUnset, "fep-x"); item.Dir != want {
		t.Fatalf("没给 ENT 该落 %s：%s", want, item.Dir)
	}
}

// 身份是"进哪一层"的唯一依据：没有它、或它没记环境，就必须拒（不猜）。
func TestPromoteRefusesWithoutIdentity(t *testing.T) {
	dataDir := t.TempDir()
	cache := filepath.Join(dataDir, "fep")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cache, "old-1-rt2gdc.log"), []byte("om 0 {{}}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lib := NewLibrary(dataDir)
	if _, err := lib.Promote(cache, "old"); !errors.Is(err, ErrNoMeta) {
		t.Fatalf("没有身份文件要拒（并说是缺身份）：%v", err)
	}
	// 有身份但没环境：同样拒
	seedRecording(t, cache, "noenv", "", "99")
	if _, err := lib.Promote(cache, "noenv"); err == nil {
		t.Fatal("身份里没有环境时必须拒：库是按环境分层的")
	}
}

// R3.3：清缓存**不碰**长期库。缓存目录与长期库都在数据目录下，靠"records 不在
// CacheSubdirs 清单里"这一条分开 —— 这条测试就是那句话的牙。
func TestCleanCacheNeverTouchesTheLibrary(t *testing.T) {
	dataDir := t.TempDir()
	cache := filepath.Join(dataDir, "fep")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	seedRecording(t, cache, "keep-me", "主机109", "99")
	lib := NewLibrary(dataDir)
	if _, err := lib.Promote(cache, "keep-me"); err != nil {
		t.Fatal(err)
	}
	// 缓存里再留一份没搬的（模拟"下一轮正在录"）
	seedRecording(t, cache, "cache-only", "主机109", "99")

	kept := filepath.Join(dataDir, "records", "主机109", "99", "keep-me")
	before, err := os.ReadDir(kept)
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := config.CleanCache(dataDir, 0); err != nil { // 0 = 整个缓存子目录删掉（用户手动清缓存）
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cache, "cache-only-1-rt2gdc.log")); !os.IsNotExist(err) {
		t.Fatal("手动清缓存该把缓存里的录制删掉（这就是为什么长期库要另落一处）")
	}
	after, err := os.ReadDir(kept)
	if err != nil {
		t.Fatalf("清缓存把长期库删了：%v", err)
	}
	if len(after) != len(before) {
		t.Fatalf("长期库少了文件：%d → %d", len(before), len(after))
	}
	if _, err := os.Stat(MetaPath(kept, "keep-me")); err != nil {
		t.Fatalf("长期库里的身份也不该被动：%v", err)
	}
	// 按年龄清理（启动清理走这条）同样只碰缓存
	if _, _, err := config.CleanCache(dataDir, time.Nanosecond); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(MetaPath(kept, "keep-me")); err != nil {
		t.Fatalf("按年龄清理也不该动长期库：%v", err)
	}
}

// 环境名是人给的字符串，只允许当"一段"目录名 —— 不然一个带分隔符的名字
// 就能把落点挪到别处（`../../..`）。
func TestLibrarySegmentsCannotEscapeTheRoot(t *testing.T) {
	lib := NewLibrary(`C:\data`)
	if got := lib.RecordingDir("../../etc", "..", "id"); got != "" {
		if rel, err := filepath.Rel(lib.Root, got); err != nil || len(rel) >= 2 && rel[:2] == ".." {
			t.Fatalf("落点跑到库外面去了：%s", got)
		}
	}
	if got := lib.EntDir("主机109", ""); filepath.Base(got) != EntUnset {
		t.Fatalf("没给 ENT 该用 %s：%s", EntUnset, got)
	}
}
