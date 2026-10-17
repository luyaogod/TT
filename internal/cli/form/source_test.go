package form

// 取件通道的本地实现与缓存（契约 C2 / A8）。
//
// 这两条都用**本地临时目录**测，不碰真机：本地目录实现本身就是"镜像/离线"那一路，
// 真机那一路（`internal/host/formsource.go`）的纯函数在 `hostfinder_test.go` 里测。

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tt/internal/config"
)

// ── 本地目录实现 ────────────────────────────────────────────────────────────

// makeTree 造一棵"环境/模块/程序/42f/<表单>.42f"形状的镜像树。
func makeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestDirSourceFindsForm(t *testing.T) {
	root := makeTree(t, map[string]string{
		"t35prd/axm/axmt500/42f/demoform.42f": "<Form name=\"x\"/>",
		"t35prd/axm/axmt501/42f/other.42f":    "<Form name=\"y\"/>",
	})
	got, err := (&DirSource{Root: root}).Form(context.Background(), "t35prd", "axm", "demoform")
	if err != nil {
		t.Fatalf("取件失败: %v", err)
	}
	if string(got) != "<Form name=\"x\"/>" {
		t.Errorf("取到的是别的文件: %q", got)
	}
}

func TestDirSourcePrefersForm42fDir(t *testing.T) {
	// 同名文件既躺在 42f/ 下也躺在别处：必须取 42f/ 那份（另一个多半是备份/草稿）
	root := makeTree(t, map[string]string{
		"axm/a/42f/demoform.42f":  "<Form name=\"good\"/>",
		"axm/backup/demoform.42f": "<Form name=\"draft\"/>",
	})
	got, err := (&DirSource{Root: root}).Form(context.Background(), "", "axm", "demoform")
	if err != nil {
		t.Fatalf("取件失败: %v", err)
	}
	if !strings.Contains(string(got), "good") {
		t.Errorf("该优先 42f/ 下的那份: %q", got)
	}
}

func TestDirSourceRefusesAmbiguous(t *testing.T) {
	root := makeTree(t, map[string]string{
		"axm/a/42f/demoform.42f": "<Form name=\"a\"/>",
		"axm/b/42f/demoform.42f": "<Form name=\"b\"/>",
	})
	_, err := (&DirSource{Root: root}).Form(context.Background(), "", "axm", "demoform")
	if err == nil {
		t.Fatal("同分歧义必须报错，绝不自己挑一个（D4）")
	}
	if !strings.Contains(err.Error(), "demoform.42f") {
		t.Errorf("错误里该列出候选: %v", err)
	}
}

func TestDirSourceMissingForm(t *testing.T) {
	root := makeTree(t, map[string]string{"axm/a/42f/other.42f": "<Form/>"})
	_, err := (&DirSource{Root: root}).Form(context.Background(), "", "axm", "demoform")
	if err == nil {
		t.Fatal("取不到必须报错")
	}
	if !strings.Contains(err.Error(), "demoform") {
		t.Errorf("错误该点名是哪张表单: %v", err)
	}
}

func TestDirSourceList(t *testing.T) {
	root := makeTree(t, map[string]string{
		"axm/a/42f/zzz.42f":   "<Form/>",
		"axm/a/42f/aaa.42f":   "<Form/>",
		"oth/b/42f/other.42f": "<Form/>",
	})
	got, err := (&DirSource{Root: root}).List(context.Background(), "", "axm")
	if err != nil {
		t.Fatalf("列目录失败: %v", err)
	}
	if strings.Join(got, ",") != "aaa,zzz" {
		t.Errorf("List = %v, want [aaa zzz]（排序、去掉扩展名、只挑该模块的）", got)
	}
}

func TestDirSourceListUnknownModule(t *testing.T) {
	root := makeTree(t, map[string]string{"axm/a/42f/aaa.42f": "<Form/>"})
	if _, err := (&DirSource{Root: root}).List(context.Background(), "", "nope"); err == nil {
		t.Fatal("该模块下一个表单都没有时必须报错，不能返回空清单（空清单会被读成「这张表没表单」）")
	}
}

// ── 缓存 ────────────────────────────────────────────────────────────────────

func TestNewCacheFromDataDir(t *testing.T) {
	dir := t.TempDir()
	c, err := NewCacheFromDataDir(dir)
	if err != nil {
		t.Fatalf("建缓存失败: %v", err)
	}
	if want := filepath.Join(dir, "forms"); c.Dir != want {
		t.Errorf("缓存落点 = %q, want %q（C2.3：<数据目录>/forms/）", c.Dir, want)
	}
	// 目录名必须走 config.CacheDir 校验：数据目录为空时拿不到落点，必须当场报错
	if _, err := NewCacheFromDataDir(""); err == nil {
		t.Fatal("数据目录解析不出来时必须报错（B7：清单外名字拿空串，静默失败最贵）")
	}
	if got := config.CacheDir(dir, "forms"); got != c.Dir {
		t.Errorf("缓存目录必须与 config.CacheDir 一致: %q vs %q", got, c.Dir)
	}
}

func TestCacheOnlyStoresRawXML(t *testing.T) {
	data := t.TempDir()
	c, err := NewCacheFromDataDir(data)
	if err != nil {
		t.Fatal(err)
	}
	raw := "<Form name=\"demo\"/>"
	if err := c.Put("t35prd", "axm", "demoform", []byte(raw), time.Time{}); err != nil {
		t.Fatalf("写缓存失败: %v", err)
	}

	// 只缓存原始 XML：那个目录下**只有**一个 .42f 文件，内容与取回来的字节一字不差
	sub := filepath.Join(c.Dir, "t35prd", "axm")
	ents, err := os.ReadDir(sub)
	if err != nil {
		t.Fatalf("缓存子目录没建出来: %v", err)
	}
	if len(ents) != 1 || ents[0].Name() != "demoform.42f" {
		t.Fatalf("缓存目录里该只有一个原始 XML: %v", ents)
	}
	b, err := os.ReadFile(filepath.Join(sub, "demoform.42f"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != raw {
		t.Errorf("缓存内容 = %q, want %q", b, raw)
	}

	got, ok := c.Lookup(context.Background(), nil, "t35prd", "axm", "demoform")
	if !ok || string(got) != raw {
		t.Errorf("缓存没命中: ok=%v got=%q", ok, got)
	}
}

func TestCacheTTLExpiry(t *testing.T) {
	data := t.TempDir()
	c, err := NewCacheFromDataDir(data)
	if err != nil {
		t.Fatal(err)
	}
	c.TTL = time.Hour
	if err := c.Put("e", "m", "f", []byte("<Form/>"), time.Time{}); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Lookup(context.Background(), nil, "e", "m", "f"); !ok {
		t.Fatal("刚写进去的该命中")
	}
	// 把它做旧（TTL 之外）→ 必须不命中
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(c.Path("e", "m", "f"), old, old); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Lookup(context.Background(), nil, "e", "m", "f"); ok {
		t.Error("超过 TTL 的缓存不该命中（否则服务器上换了表单我们永远看不到）")
	}
}

// identSource 是"能报出源那边大小/时间"的假件（真机实现走 SFTP Stat）。
type identSource struct {
	fakeSource
	size  int64
	mtime time.Time
}

func (s *identSource) Identify(_ context.Context, _, _, _ string) (int64, time.Time, error) {
	return s.size, s.mtime, nil
}

func TestCacheKeyIncludesSourceSizeAndTime(t *testing.T) {
	data := t.TempDir()
	c, err := NewCacheFromDataDir(data)
	if err != nil {
		t.Fatal(err)
	}
	inner := &identSource{size: int64(len("<Form/>")), mtime: time.Unix(1000, 0)}
	cs := &CachingSource{Inner: inner, Cache: c}

	if err := c.Put("e", "m", "f", []byte("<Form/>"), inner.mtime); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Lookup(context.Background(), inner, "e", "m", "f"); !ok {
		t.Fatal("源那边没变（size+mtime 相同）时该命中")
	}
	// 源那边变了（表格被重新编译）→ 必须不命中
	inner.size = 999
	if _, ok := c.Lookup(context.Background(), inner, "e", "m", "f"); ok {
		t.Error("源那边 size 变了必须重取（C2.4：缓存键含 size+mtime）")
	}
	inner.size = int64(len("<Form/>"))
	inner.mtime = time.Unix(2000, 0)
	if _, ok := c.Lookup(context.Background(), inner, "e", "m", "f"); ok {
		t.Error("源那边 mtime 变了必须重取")
	}
	_ = cs
}

func TestCacheRejectsUnsafeNames(t *testing.T) {
	data := t.TempDir()
	c, err := NewCacheFromDataDir(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range [][3]string{
		{"..", "m", "f"},
		{"e", "..", "f"},
		{"e", "m", "../escape"},
		{"e", "m", "a/b"},
		{"e", "m", "a\\b"},
	} {
		if p := c.Path(tc[0], tc[1], tc[2]); p != "" {
			t.Errorf("危险名字 %v 该拿空路径，得到 %q（拼进路径就是路径穿越）", tc, p)
		}
		if err := c.Put(tc[0], tc[1], tc[2], []byte("<Form/>"), time.Time{}); err == nil {
			t.Errorf("写缓存 %v 该报错", tc)
		}
	}
}

// ── 缓存包装 ────────────────────────────────────────────────────────────────

func TestCachingSourceFetchesOnceAndReuses(t *testing.T) {
	data := t.TempDir()
	c, err := NewCacheFromDataDir(data)
	if err != nil {
		t.Fatal(err)
	}
	inner := &countingSource{body: []byte(fixtureBasic(t))}
	cs := &CachingSource{Inner: inner, Cache: c}
	ctx := context.Background()

	first, err := cs.Form(ctx, "demoenv", "demomod", "demoform")
	if err != nil {
		t.Fatalf("第一次取件失败: %v", err)
	}
	second, err := cs.Form(ctx, "demoenv", "demomod", "demoform")
	if err != nil {
		t.Fatalf("第二次取件失败: %v", err)
	}
	if inner.calls != 1 {
		t.Errorf("取件被调用了 %d 次，want 1（第二次该走缓存）", inner.calls)
	}
	if string(first) != string(second) {
		t.Error("两次拿到的字节该一致")
	}

	// 假件也不改：缓存里的字节与刚取回来的一致（且是原始 XML，不是解析结果）
	b, err := os.ReadFile(c.Path("demoenv", "demomod", "demoform"))
	if err != nil {
		t.Fatal(err)
	}
	var probe struct {
		Build string `json:"build"`
	}
	if json.Unmarshal(b, &probe) == nil && probe.Build != "" {
		t.Errorf("缓存里是 JSON（解析结果）而不是原始 XML: %s", b)
	}
	if !strings.Contains(string(b), "<Form name=") {
		t.Errorf("缓存里该是原始 XML: %s", b)
	}
}

func TestCachingSourceFetchErrorIsNotCached(t *testing.T) {
	data := t.TempDir()
	c, err := NewCacheFromDataDir(data)
	if err != nil {
		t.Fatal(err)
	}
	inner := &countingSource{fakeSource: fakeSource{err: os.ErrNotExist}}
	cs := &CachingSource{Inner: inner, Cache: c}

	if _, err := cs.Form(context.Background(), "e", "m", "f"); err == nil {
		t.Fatal("取件失败要原样报出去")
	}
	if _, err := cs.Form(context.Background(), "e", "m", "f"); err == nil {
		t.Fatal("第二次仍要报错")
	}
	if inner.calls != 2 {
		t.Errorf("失败不该被缓存（否则网络恢复后还要等 TTL）: calls=%d", inner.calls)
	}
}

type countingSource struct {
	fakeSource
	body  []byte
	calls int
}

func (s *countingSource) Form(_ context.Context, _, _, _ string) ([]byte, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.body, nil
}
