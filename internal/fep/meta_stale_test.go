package fep

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// ── A24（契约 §7，用户决策 D-F2）：旧件处置 = **只打标记，不加逃生门** ─────────
//
// 这三条测试守的是同一件事：`mark-stale` 只动身份文件，**录制本身一个字节都不碰**。
// 录制是不可再生的（retention.go 那套"住哪表达寿命"的反面：长期库不参与自动清理），
// 所以"标记"必须与"删除"在实现上分开 —— 分开的判据就在这里：文件集合与逐字节哈希不变。

// staleFixture 造一份"有身份文件"的录制（两个流量文件 + .meta.json）。
func staleFixture(t *testing.T) (dir, id string, started time.Time) {
	t.Helper()
	dir = t.TempDir()
	id = "fep-20261017-101010"
	started = time.Date(2026, 10, 17, 10, 10, 10, 0, time.UTC)
	for _, name := range []string{id + "-1-rt2gdc.log", id + "-1-gdc2rt.log", id + "-1-timeline.log"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("raw bytes of "+name), 0o644); err != nil {
			t.Fatalf("写录制文件 %s: %v", name, err)
		}
	}
	if _, err := WriteMeta(dir, RecordingMeta{
		ID: id, Path: PathRunnerRR, Prog: "azzp000", Env: "t35dev", Ent: "1",
		Started: started, Finished: started.Add(3 * time.Minute),
		Origin: OriginUser, Retention: RetainLibrary,
	}); err != nil {
		t.Fatalf("写身份文件: %v", err)
	}
	return dir, id, started
}

// staleFileSet 把目录里的**录制文件**逐个哈希（不含身份文件 —— 标记就是要改它）。
func staleFileSet(t *testing.T, dir string) map[string]string {
	t.Helper()
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("读目录: %v", err)
	}
	out := map[string]string{}
	for _, e := range ents {
		if e.IsDir() || strings.HasSuffix(e.Name(), MetaSuffix) {
			continue
		}
		b, rerr := os.ReadFile(filepath.Join(dir, e.Name()))
		if rerr != nil {
			t.Fatalf("读 %s: %v", e.Name(), rerr)
		}
		sum := sha256.Sum256(b)
		out[e.Name()] = hex.EncodeToString(sum[:])
	}
	return out
}

func sameFileSet(t *testing.T, before, after map[string]string) {
	t.Helper()
	var names []string
	for n := range before {
		names = append(names, n)
	}
	sort.Strings(names)
	if len(before) != len(after) {
		t.Fatalf("录制文件个数变了：%d → %d（标记绝不许删文件）", len(before), len(after))
	}
	for _, n := range names {
		if after[n] != before[n] {
			t.Fatalf("录制文件 %s 的字节被动了（标记只许改身份文件）", n)
		}
	}
}

// TestMarkStaleMarksMetaOnly：A24 的核心 —— 只打标记，不删不改录制。
func TestMarkStaleMarksMetaOnly(t *testing.T) {
	dir, id, started := staleFixture(t)
	before := staleFileSet(t, dir)

	at := time.Date(2026, 10, 18, 9, 30, 0, 0, time.UTC)
	m, err := MarkStale(dir, id, "无身份（该件早于身份上线）", at)
	if err != nil {
		t.Fatalf("打标记失败：%v", err)
	}
	if !m.Stale {
		t.Fatalf("返回值里没有 stale=true：%+v", m)
	}
	if m.StaleReason != "无身份（该件早于身份上线）" {
		t.Fatalf("原因没记下来：%q", m.StaleReason)
	}
	if !m.StaleAt.Equal(at) {
		t.Fatalf("时间 = %v，想要 %v", m.StaleAt, at)
	}

	// ① 身份文件里真的有这三样（重新读一遍，不信内存里的那份）
	reread, err := ReadMeta(dir, id)
	if err != nil {
		t.Fatalf("读回身份：%v", err)
	}
	if !reread.IsStale() || reread.StaleReason == "" || reread.StaleAt.IsZero() {
		t.Fatalf("身份文件里没落标记：%+v", reread)
	}
	// 其它身份字段一个没丢（标记不是"重写身份"）
	if reread.Prog != "azzp000" || reread.Env != "t35dev" || !reread.Started.Equal(started) {
		t.Fatalf("标记把别的身份字段弄丢了：%+v", reread)
	}

	// ② 录制目录里**一个文件不少、一个字节不变**
	sameFileSet(t, before, staleFileSet(t, dir))

	// ③ 落盘的键名就是契约写的 stale / reason / staleAt（三个都 omitempty）
	raw, err := os.ReadFile(MetaPath(dir, id))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"stale": true`, `"reason"`, `"staleAt"`} {
		if !strings.Contains(string(raw), key) {
			t.Fatalf("身份文件里没有 %s：\n%s", key, raw)
		}
	}
	t.Logf("标记后的身份文件：\n%s", raw)
}

// TestOldMetaWithoutStaleFieldsReadsAsLive：老身份文件（没有这三个键）读到即 stale=false，
// 不报错、也不许因此崩（omitempty 的全部意义）。
func TestOldMetaWithoutStaleFieldsReadsAsLive(t *testing.T) {
	dir := t.TempDir()
	id := "fep-20261001-000000"
	const old = `{"id":"fep-20261001-000000","path":"r.r","prog":"azzp000",` +
		`"started":"2026-10-01T00:00:00Z","fingerprint":{"total":3}}`
	if err := os.WriteFile(MetaPath(dir, id), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := ReadMeta(dir, id)
	if err != nil {
		t.Fatalf("老身份文件读不动了：%v", err)
	}
	if m.Stale || m.StaleReason != "" || !m.StaleAt.IsZero() {
		t.Fatalf("老件里没有这三个键，该是零值：%+v", m)
	}
	if m.IsStale() {
		t.Fatal("IsStale 对老件要回 false")
	}
	if m.Prog != "azzp000" || m.Path != PathRunnerRR {
		t.Fatalf("老件的既有字段读坏了：%+v", m)
	}
	// 没标记的件写出去不该冒出 `stale` / `reason` 两个键（omitempty）。
	// `staleAt` 除外：encoding/json 不认识 time.Time 的零值（既有 `finished` 同样如此），
	// 所以它总会写一个零时刻 —— 判据只看 `stale`（字段注释里写明了）。
	b, err := json.Marshal(*m)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"stale"`, `"reason"`} {
		if strings.Contains(string(b), key) {
			t.Fatalf("没标记却写出了 %s 键：%s", key, b)
		}
	}
}

// TestMarkStaleWithoutMetaWritesOnlyMeta：连身份文件都没有的件也能标记 ——
// 补写一份**只有身份与标记**的 .meta.json，录制文件照旧一个字节不动。
// （老件正是最需要标记的那一类：它们既没有指纹也没有身份，只有"别拿它回放"这句话。）
func TestMarkStaleWithoutMetaWritesOnlyMeta(t *testing.T) {
	dir := t.TempDir()
	id := "fep-20260901-000000"
	raw := filepath.Join(dir, id+"-1-rt2gdc.log")
	if err := os.WriteFile(raw, []byte("raw bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadMeta(dir, id); err == nil {
		t.Fatal("夹具本来不该有身份文件")
	}
	at := time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC)
	m, err := MarkStale(dir, id, "", at)
	if err != nil {
		t.Fatalf("没有身份文件的件也该能标记：%v", err)
	}
	if m.ID != id || !m.Stale || !m.StaleAt.Equal(at) {
		t.Fatalf("补写的身份不对：%+v", m)
	}
	if m.StaleReason != "" {
		t.Fatalf("没给原因就该是空：%q", m.StaleReason)
	}
	if b, rerr := os.ReadFile(raw); rerr != nil || string(b) != "raw bytes" {
		t.Fatalf("录制文件被动了：%q / %v", b, rerr)
	}
	// 空原因不写空键（omitempty）
	mb, err := os.ReadFile(MetaPath(dir, id))
	if err != nil {
		t.Fatalf("补写的身份文件不在：%v", err)
	}
	if strings.Contains(string(mb), `"reason"`) {
		t.Fatalf("没给原因却写出了 reason 键：%s", mb)
	}
}
