package form

// 缓存路径三段名字（env/module/form）的白名单（台账 D37）。
//
// 现场实测：环境名 `主机109` ⇒ `tt form fetch` 90/90 全部写缓存失败，
// `%APPDATA%\tt\forms` 至今不存在。病因就在白名单只认 ASCII ——
// 于是"缓存永远不写、永远不读"，而且**一声不响**。
//
// 这一组用例钉住两件事：
//
//  1. CJK 名字要能写进缓存、并且能读回来（只放宽 `cacheNameOK` 那一行字符集不够，
//     得证明它真的落盘了 —— 临时目录，`t.TempDir()`，不碰真 `%APPDATA%`）。
//  2. 危险名字一律拒，**而且拒绝必须说清为什么**：`Path` 拿空串是既有契约，
//     但 `Put` 的报错里要带着病因，否则现场又回到"静默失败"。

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCacheAcceptsCJKEnvName 环境名 `主机109` 能写进缓存并能读回（台账 D37 的正面用例）。
func TestCacheAcceptsCJKEnvName(t *testing.T) {
	data := t.TempDir()
	c, err := NewCacheFromDataDir(data)
	if err != nil {
		t.Fatal(err)
	}

	const (
		env    = "主机109"
		module = "axm"
		form   = "axmt500_wf"
	)
	body := []byte(`<Form name="axmt500_wf"><Field id="f1"/></Form>`)

	p, perr := c.PathError(env, module, form)
	if perr != nil {
		t.Fatalf("中文环境名 %q 该被放行，却被拒：%v", env, perr)
	}
	if p == "" {
		t.Fatalf("中文环境名 %q 算出了空路径", env)
	}
	// 仍然必须**落在缓存目录之内**：放宽字符集不等于放宽落点。
	if !strings.HasPrefix(p, c.Dir+string(os.PathSeparator)) {
		t.Fatalf("路径 %q 不在缓存目录 %q 之内", p, c.Dir)
	}
	want := filepath.Join(c.Dir, env, module, form+".42f")
	if p != want {
		t.Fatalf("落点=%q，want %q", p, want)
	}

	if err := c.Put(env, module, form, body, time.Time{}); err != nil {
		t.Fatalf("写缓存失败（现场症状就是这一步 90/90 全挂）：%v", err)
	}

	// 真落盘了：文件在，字节就是原文（缓存只放原始 XML）。
	got, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("缓存文件没落盘：%v", err)
	}
	if string(got) != string(body) {
		t.Errorf("缓存内容=%q，want %q", got, body)
	}

	// 能读回来：src 为 nil ⇒ 只按 TTL 判新鲜（刚写进去，必然新鲜）。
	back, ok := c.Lookup(context.Background(), nil, env, module, form)
	if !ok {
		t.Fatalf("刚写进去的缓存读不回来：env=%q", env)
	}
	if string(back) != string(body) {
		t.Errorf("读回=%q，want %q", back, body)
	}
}

// TestCacheNameAllowsStillASCII 放宽字符集不能把 ASCII 那半边弄丢。
func TestCacheNameAllowsStillASCII(t *testing.T) {
	c := &Cache{Dir: filepath.Join(t.TempDir(), "forms")}
	for _, name := range []string{
		"demoenv", "t35prd", "a-b_c.d", "上机1", "主机_109-b.test", "日本語",
	} {
		if err := c.nameError(name, "环境名"); err != nil {
			t.Errorf("名字 %q 该放行，却被拒：%v", name, err)
		}
	}
}

// TestCacheRejectsDangerousNames 危险名字一律拒，且**拒绝要能说清为什么**。
//
// `want` 是错误文案里必须出现的病因关键词 —— 不静默返回空、也不给一句"非法名字"。
func TestCacheRejectsDangerousNames(t *testing.T) {
	data := t.TempDir()
	c, err := NewCacheFromDataDir(data)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		want string // 错误文案里必须出现的病因关键词
	}{
		{"../evil", ".."},
		{"..", ". 或 .."},
		{".", ". 或 .."},
		{"", "为空"},
		{"a/b", "路径分隔符"},
		{`a\b`, "路径分隔符"},
		{"a\x00b", "控制字符"},
		{"a\nb", "控制字符"},
		{"a:b", "保留字符"},
		{"a*b", "保留字符"},
		{"a?b", "保留字符"},
		{`a"b`, "保留字符"},
		{"a<b", "保留字符"},
		{"a>b", "保留字符"},
		{"a|b", "保留字符"},
		{"a b", "白名单外的字符"},
		{"a\xff b", "UTF-8"},
		{strings.Repeat("中", maxCacheNameRunes+1), "超过"},
	}

	for _, tc := range cases {
		// Path 的既有契约不变：不合法 ⇒ 空串（调用方靠这个不拼路径）。
		if p := c.Path(tc.name, "m", "f"); p != "" {
			t.Errorf("危险名字 %q 该拿空路径，得到 %q（拼进路径就是路径穿越）", tc.name, p)
		}
		if _, err := c.PathError(tc.name, "m", "f"); err == nil {
			t.Errorf("危险名字 %q 该报错（PathError）", tc.name)
		} else if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("危险名字 %q 的报错没说清为什么：%v\n    文案里该有 %q", tc.name, err, tc.want)
		}
		// Put 的报错要原样带出病因 —— 现场看到的就是这一行。
		err := c.Put(tc.name, "m", "f", []byte("<Form/>"), time.Time{})
		if err == nil {
			t.Errorf("写缓存 %q 该报错", tc.name)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("写缓存 %q 的报错没说清为什么：%v\n    文案里该有 %q", tc.name, err, tc.want)
		}
	}

	// 三段都查，不能只看 env。
	if _, err := c.PathError("e", "../m", "f"); err == nil {
		t.Error("module 段该被查")
	}
	if _, err := c.PathError("e", "m", "a/b"); err == nil {
		t.Error("form 段该被查")
	}
}
