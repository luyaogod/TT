package config

// WP2 / 牙齿 d 的一半（契约 §3「一致性断言」、计划 WP2 的 `TestFormsDirInCacheSubdirs`）。
//
// 为什么漏登记缓存目录值得一条专门的测试：`CacheDir` 对**清单外**的名字返回空串
// （B7 的"住哪表达寿命"口径），于是漏登记的表现不是报错，而是"表单缓存永远不写、
// 永远不读"—— 取件照旧每次都走网络，而**没有任何一处会失败**。
//
// 落点口径（契约 C2.3）：`<数据目录>/forms/`。

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestFormsDirInCacheSubdirs(t *testing.T) {
	if !slices.Contains(CacheSubdirs, "forms") {
		t.Fatalf("CacheSubdirs 里没有 \"forms\"（现有：%v）—— .42f 表单缓存会静默落到空路径",
			CacheSubdirs)
	}

	dir := t.TempDir()
	if got, want := CacheDir(dir, "forms"), filepath.Join(dir, "forms"); got != want {
		t.Errorf("CacheDir(dir, \"forms\") = %q, want %q", got, want)
	}

	loc := LocationsAt(filepath.Join(dir, "config.json"))
	if got, want := loc.CacheDir("forms"), filepath.Join(dir, "forms"); got != want {
		t.Errorf("Locations.CacheDir(\"forms\") = %q, want %q（落点必须是 <数据目录>/forms）", got, want)
	}

	// 清单外名字仍然拿空串：这条同时是"上面那条不是靠枚举巧合过"的反证。
	if got := CacheDir(dir, "forms-nope"); got != "" {
		t.Errorf("清单外名字该拿空串，得到 %q —— 校验被绕过了", got)
	}

	st := CacheStatusOf(dir)
	found := false
	for _, d := range st.Dirs {
		if d.Name == "forms" {
			found = true
		}
	}
	if !found {
		t.Errorf("CacheStatusOf 没把 forms 算进来（`tt cache` 会看不到它）: %+v", st.Dirs)
	}
}
