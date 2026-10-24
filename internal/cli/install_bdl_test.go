package cli

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"tt/internal/cli/common"
	"tt/internal/config"
)

func makeBDLZip(t *testing.T, p string) {
	t.Helper()
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range map[string]string{
		"BDL-Markdown/a.md":     "a",
		"BDL-Markdown/sub/b.md": "b",
	} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestInstallBDLDocsExtractsAndSkips(t *testing.T) {
	src := filepath.Join(t.TempDir(), bdlZipName)
	makeBDLZip(t, src)
	dst := filepath.Join(t.TempDir(), "BDL-Markdown")

	files, skipped, err := installBDLDocs(src, dst, false)
	if err != nil {
		t.Fatalf("installBDLDocs: %v", err)
	}
	if skipped || files != 2 {
		t.Fatalf("files=%d skipped=%v, want 2/false", files, skipped)
	}
	for _, p := range []string{"a.md", "sub/b.md"} {
		if _, err := os.Stat(filepath.Join(dst, p)); err != nil {
			t.Fatalf("解压后缺少 %s: %v", p, err)
		}
	}
	// 去掉 zip 顶层 BDL-Markdown/：不应留下嵌套目录
	if _, err := os.Stat(filepath.Join(dst, "BDL-Markdown")); err == nil {
		t.Fatal("zip 顶层目录不该被保留")
	}

	if _, skipped, err = installBDLDocs(src, dst, false); err != nil || !skipped {
		t.Fatalf("同一 zip 第二次应跳过: skipped=%v err=%v", skipped, err)
	}
	if _, skipped, err = installBDLDocs(src, dst, true); err != nil || skipped {
		t.Fatalf("--force 应重新解压: skipped=%v err=%v", skipped, err)
	}
}

func TestInstallBdlCmdWritesBDLDocDir(t *testing.T) {
	t.Setenv("TT_CONFIG", "") // 隔离真实 TT_CONFIG,否则配置读写会落到生产 config.json
	src := filepath.Join(t.TempDir(), bdlZipName)
	makeBDLZip(t, src)
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	oldCfg, oldSrc := common.ConfigPath, bdlZipSource
	common.ConfigPath = cfgPath
	bdlZipSource = func() (string, error) { return src, nil }
	t.Cleanup(func() { common.ConfigPath, bdlZipSource = oldCfg, oldSrc })

	dst := filepath.Join(t.TempDir(), "docs")
	cmd := newInstallBdlCmd()
	cmd.SetArgs([]string{"--dir", dst})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("install bdl: %v", err)
	}
	root, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Clean(root.Bdldoc.Dir) != filepath.Clean(dst) {
		t.Fatalf("bdldoc.dir = %q, want %q", root.Bdldoc.Dir, dst)
	}
	if _, err := os.Stat(filepath.Join(dst, "a.md")); err != nil {
		t.Fatalf("文档没解压到 --dir: %v", err)
	}
}
