package devdata

// tt gen pack：把已拉到的材料打成 .tzc。
//
// 断言的重点不是"生成了一个文件"，而是：
//   · 包内成员就是 4gl / tap / tgl / ver，名字与真包（<prog>.<ext> + ver）一致；
//   · 身份从 .tap 根属性读得出，包名跟着变（(c) / (s)）；
//   · 写出来的东西能被 pkgfile 当 Code 包打开（mandatory .tap/.tgl + ver 可解析）；
//   · 缺材料时给的是"先 fetch"的引导，而不是一个空包。

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/dev/pkgfile"
)

const testProg = "axmt500_wf"

const testTap = `<?xml version="1.0" encoding="UTF-8"?>
<add_points prog="axmt500_wf" std_prog="axmt500_wf" erpver="3.0" module="CXM" ver="3" identity="c" section_flag="N">
<section id="axmt500_wf.input" ver="56" src="s">
</section>
</add_points>
`

// setupPackEnv 造一份"已经 fetch 过"的程序材料，返回 config 路径与程序目录。
func setupPackEnv(t *testing.T, with4gl bool) (cfg, progDir string) {
	t.Helper()
	cfg = writeConfig(t, `{"name":"e","host":"h","port":22,"user":"u"}`)
	envDir := filepath.Join(filepath.Dir(cfg), "devdata", "e")
	progDir = filepath.Join(envDir, "prog", testProg)

	files := map[string]string{
		"dzx/" + testProg + ".tap": testTap,
		"dzx/" + testProg + ".tgl": "TGL-HEADER\n",
	}
	if with4gl {
		files["src/"+testProg+".4gl"] = "PROGRAM " + testProg + "\n"
	}
	var entries []FileEntry
	for rel, body := range files {
		p := filepath.Join(progDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("建目录: %v", err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("写文件: %v", err)
		}
		role := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
		switch {
		case strings.HasSuffix(rel, ".4gl"):
			role = "4gl"
		case strings.HasSuffix(rel, ".tap"):
			role = "tap"
		case strings.HasSuffix(rel, ".tgl"):
			role = "tgl"
		}
		entries = append(entries, FileEntry{
			Role: role, Src: "erp/cxm/" + rel, Rel: rel,
			SHA256: pkgfile.Sha256Hex([]byte(body)), Size: int64(len(body)),
		})
	}
	mf := Manifest{Schema: "tt.devdata/v1", Kind: "fetch", Env: "e", Prog: testProg, Module: "cxm", Files: entries}
	b, _ := json.MarshalIndent(mf, "", "  ")
	if err := os.WriteFile(filepath.Join(progDir, "manifest.json"), b, 0o644); err != nil {
		t.Fatalf("写 manifest: %v", err)
	}
	return cfg, progDir
}

func TestPackCreatesCodePackage(t *testing.T) {
	cfg, progDir := setupPackEnv(t, true)

	res, err := Pack(PackRequest{CfgPath: cfg, Env: "e", Prog: testProg})
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}

	// 命名与落点：<环境>/pack/<程序>(<身份>).tzc
	wantPath := filepath.Join(filepath.Dir(cfg), "devdata", "e", "pack", testProg+"(c).tzc")
	if res.Path != wantPath {
		t.Errorf("路径 = %s，想要 %s", res.Path, wantPath)
	}
	if res.Identity != "c" {
		t.Errorf("identity = %q，想要从 .tap 读出的 c", res.Identity)
	}

	names := []string{}
	for _, e := range res.Entries {
		names = append(names, e.Name)
	}
	if got, want := strings.Join(names, ","), testProg+".4gl,"+testProg+".tap,"+testProg+".tgl,ver"; got != want {
		t.Fatalf("成员 = %s，想要 %s", got, want)
	}

	// 包必须能被自己的读取器当 Code 包打开，且 ver 可解析。
	pkg, err := pkgfile.Open(res.Path, pkgfile.OpenOptions{AcceptVer: "1.0"})
	if err != nil {
		t.Fatalf("产出的包打不开: %v", err)
	}
	if pkg.Kind != pkgfile.KindCode {
		t.Fatalf("Kind = %s，想要 Code", pkg.Kind)
	}
	if pkg.Ver.Raw != "1.0" {
		t.Errorf("ver = %q，想要 1.0", pkg.Ver.Raw)
	}

	// 成员内容必须与材料逐字节一致（不能是"再生成"的东西）。
	byName := map[string][]byte{}
	for _, e := range pkg.Entries {
		byName[e.Name] = e.Data
	}
	src, err := os.ReadFile(filepath.Join(progDir, "dzx", testProg+".tap"))
	if err != nil {
		t.Fatalf("读源 tap: %v", err)
	}
	if string(byName[testProg+".tap"]) != string(src) {
		t.Error("包内 .tap 与材料不一致")
	}
	if strings.TrimSpace(string(byName["ver"])) != "1.0" {
		t.Errorf("ver 内容 = %q，想要 1.0", byName["ver"])
	}
	if res.Bytes <= 0 || res.SHA256 == "" {
		t.Errorf("结果缺体积/摘要: %d / %q", res.Bytes, res.SHA256)
	}
}

func TestPackIdentityFlagAndOverwrite(t *testing.T) {
	cfg, _ := setupPackEnv(t, true)

	res, err := Pack(PackRequest{CfgPath: cfg, Env: "e", Prog: testProg, Identity: "s"})
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}
	if filepath.Base(res.Path) != testProg+"(s).tzc" {
		t.Errorf("包名 = %s，想要 %s(s).tzc", filepath.Base(res.Path), testProg)
	}

	// 同名再打一次：不给 --force 必须拒绝，不能悄悄覆盖。
	if _, err := Pack(PackRequest{CfgPath: cfg, Env: "e", Prog: testProg, Identity: "s"}); !errors.Is(err, ErrExists) {
		t.Errorf("重复打包的错 = %v，想要 ErrExists", err)
	}
	if _, err := Pack(PackRequest{CfgPath: cfg, Env: "e", Prog: testProg, Identity: "s", Force: true}); err != nil {
		t.Errorf("--force 覆盖失败: %v", err)
	}
}

func TestPackMissingMaterial(t *testing.T) {
	cfg, _ := setupPackEnv(t, false) // 只有 tap/tgl，没有 4gl

	_, err := Pack(PackRequest{CfgPath: cfg, Env: "e", Prog: testProg})
	if err == nil {
		t.Fatal("缺 4gl 应该失败")
	}
	if !errors.Is(err, ErrMaterialIncomplete) {
		t.Errorf("错误 = %v，想 ErrMaterialIncomplete 包一层", err)
	}
	if !strings.Contains(err.Error(), "fetch") {
		t.Errorf("错误里应该给出下一步（tt gen fetch）：%v", err)
	}
}

func TestPackNoManifest(t *testing.T) {
	cfg := writeConfig(t, `{"name":"e","host":"h","port":22,"user":"u"}`)
	_, err := Pack(PackRequest{CfgPath: cfg, Env: "e", Prog: testProg})
	if !errors.Is(err, ErrProgNotFound) {
		t.Fatalf("没拉取过的程序：err = %v，想要 ErrProgNotFound", err)
	}
}

func TestPackBadIdentity(t *testing.T) {
	cfg, _ := setupPackEnv(t, true)
	if _, err := Pack(PackRequest{CfgPath: cfg, Env: "e", Prog: testProg, Identity: "x"}); !errors.Is(err, ErrInvalidOpt) {
		t.Errorf("非法 identity：err = %v，想要 ErrInvalidOpt", err)
	}
}

func TestPackBadVer(t *testing.T) {
	cfg, _ := setupPackEnv(t, true)
	if _, err := Pack(PackRequest{CfgPath: cfg, Env: "e", Prog: testProg, Ver: "abc"}); !errors.Is(err, ErrInvalidOpt) {
		t.Errorf("非法 ver：err = %v，想要 ErrInvalidOpt", err)
	}
}
