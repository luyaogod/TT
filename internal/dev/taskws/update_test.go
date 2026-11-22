package taskws

// Update 的测试。
//
// Update 的完整链路要连服务器（devdata.Check→SFTP、devdata.Fetch→SFTP），断网环境里
// 能钉的是**闸门顺序**与 `--check` 的"只读承诺"：
//
//   - 未落盘改动 / 包与基线不一致：必须在联网之前拒绝（这就是它们的价值）。
//   - `--check` 例外：它不写盘，而"服务器有没有新规格"恰恰是带改动时最想知道的。
//   - Check 能在本地成功（材料清单缺失时它按 missing-local 报 stale，不 probe），
//     所以 --check 的整条"核对→回报告"路径是能离线走完的；再往下的 Fetch 才需要网络，
//     用例断言它**在拨号之前**就失败（合成配置里环境没有 SSH 主机）。
//
// 拉材料之后的三条分支（重建视图/保留成果/只刷材料）需要 devdata.Fetch 成功，留给真机；
// 本文件把它们依赖的纯函数（dedupe/staleText/appendMdl）单独钉住。

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/devdata"
)

// TestUpdate_LoadError: 目录不是工作区时更新命令要报 ErrNoWorkspace（退出码语义
// 依赖它），而不是去联网拉一个不存在的工作区。
func TestUpdate_LoadError(t *testing.T) {
	if _, err := Update(context.Background(), UpdateRequest{Dir: t.TempDir()}); !errors.Is(err, ErrNoWorkspace) {
		t.Fatalf("err = %v，想要 ErrNoWorkspace", err)
	}
}

// TestUpdate_CheckOnlyOffline: --check 只向服务器核对、一个字节都不写。
// 材料清单缺失时 devdata.Check 按 missing-local 判 stale（不 probe），
// 所以这条"只读核对"能在断网环境里整条走完。
func TestUpdate_CheckOnlyOffline(t *testing.T) {
	f := newFixture(t)
	f.removeMatManifest()
	before := f.readTaskJSON()

	res, err := Update(context.Background(), UpdateRequest{
		CfgPath: f.Cfg, Dir: f.Root, CheckOnly: true,
	})
	if err != nil {
		t.Fatalf("Update --check: %v", err)
	}
	if !res.Checked {
		t.Error("Checked 未置位")
	}
	if !res.Stale || len(res.StaleReasons) == 0 {
		t.Errorf("清单缺失应判 stale: %+v", res)
	}
	if res.StaleReasons[0].Code != "missing-local" {
		t.Errorf("stale 原因 = %+v，想要 missing-local", res.StaleReasons)
	}
	if res.Fetch != nil || res.IndexRebuilt || res.Repacked || res.Reexported {
		t.Errorf("--check 不该动任何东西: %+v", res)
	}
	if !strings.Contains(strings.Join(res.Notes, "；"), "只核对") {
		t.Errorf("没有只核对提示: %v", res.Notes)
	}
	if res.Status == nil {
		t.Error("--check 也要回一份状态")
	}
	// task.json 的字节必须一个没动（UpdatedAt 也不许刷新）
	if string(f.readTaskJSON()) != string(before) {
		t.Error("--check 改了 task.json")
	}
	// 材料也没被重写
	if _, serr := os.Stat(filepath.Join(f.P.Materials(), "manifest.json")); !os.IsNotExist(serr) {
		t.Errorf("--check 不该补出材料清单: %v", serr)
	}
}

// TestUpdate_CheckOnlyAllowedWhileDirty: 视图有未落盘改动时，普通 update 拒绝，
// 但 --check 必须放行 —— 它承诺不写盘（README 里记着实测：试用者被这条挡在门外）。
func TestUpdate_CheckOnlyAllowedWhileDirty(t *testing.T) {
	f := newFixture(t)
	f.dirtyView()
	f.removeMatManifest()

	// 普通 update 先在联网前被脏闸门拒绝
	if _, err := Update(context.Background(), UpdateRequest{CfgPath: f.Cfg, Dir: f.Root}); !errors.Is(err, ErrDirtyWorkspace) {
		t.Fatalf("普通 update: err = %v，想要 ErrDirtyWorkspace", err)
	}
	// --check：脏工作区里也放行，且不写盘
	before := f.readTaskJSON()
	res, err := Update(context.Background(), UpdateRequest{CfgPath: f.Cfg, Dir: f.Root, CheckOnly: true})
	if err != nil {
		t.Fatalf("--check 在脏工作区应放行: %v", err)
	}
	if !res.Checked || !strings.Contains(strings.Join(res.Notes, "；"), "只核对") {
		t.Errorf("结果 = %+v", res)
	}
	if string(f.readTaskJSON()) != string(before) {
		t.Error("--check 改了 task.json")
	}
}

// TestUpdate_CheckOnlyBypassesPackageMismatch: 包与视图基线不一致时普通 update 拒绝，
// 但 --check 只是核对服务器，不碰包 —— 也要放行（否则状态修好之前连"服务器有没有新规格"
// 都问不了）。
func TestUpdate_CheckOnlyBypassesPackageMismatch(t *testing.T) {
	f := newFixture(t)
	f.writePkg(map[string]string{
		f.Prog + ".tap": "<不是合法 XML",
		f.Prog + ".tgl": "x\n",
		f.Prog + ".4gl": "MAIN\nEND MAIN\n",
		"ver":           "1.0\n",
	})
	f.removeMatManifest()
	if _, err := Update(context.Background(), UpdateRequest{CfgPath: f.Cfg, Dir: f.Root}); !errors.Is(err, ErrPackageMismatch) {
		t.Fatalf("普通 update: err = %v，想要 ErrPackageMismatch", err)
	}
	if _, err := Update(context.Background(), UpdateRequest{CfgPath: f.Cfg, Dir: f.Root, CheckOnly: true}); err != nil {
		t.Fatalf("--check 不该被包不一致挡住: %v", err)
	}
}

// TestUpdate_StopsBeforeDialing: 真去拉材料时，合成配置里的环境没有 SSH 主机 ——
// 失败信息必须是"没有配置 SSH 主机"，也就是**没有真的拨号**。这条同时钉住
// "check 成功后 needFetch 为真才会走到 Fetch"。
func TestUpdate_StopsBeforeDialing(t *testing.T) {
	f := newFixture(t)
	f.removeMatManifest()
	before := f.readTaskJSON()

	_, err := Update(context.Background(), UpdateRequest{CfgPath: f.Cfg, Dir: f.Root})
	if err == nil {
		t.Fatal("没有 SSH 主机时 Fetch 应失败")
	}
	if errors.Is(err, ErrDirtyWorkspace) || errors.Is(err, ErrPackageMismatch) {
		t.Fatalf("应先过闸门再拉料，而不是在闸门被拒: %v", err)
	}
	if !strings.Contains(err.Error(), "没有配置 SSH 主机") {
		t.Errorf("失败点不在拨号之前（可能是真的联网了）: %v", err)
	}
	if string(f.readTaskJSON()) != string(before) {
		t.Error("拉料失败不该改 task.json")
	}
	// Fetch 的临时目录必须收干净（失败不留半套）
	entries, rerr := os.ReadDir(f.Root)
	if rerr != nil {
		t.Fatal(rerr)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("留下临时目录: %s", e.Name())
		}
	}
}

// TestUpdate_ForceStillStopsBeforeDialing: --force 只是允许"带着脏改动只刷材料"，
// 不是允许跳过闸门之后乱写；网络不可用时同样在拨号前失败。
func TestUpdate_ForceStillStopsBeforeDialing(t *testing.T) {
	f := newFixture(t)
	f.dirtyView()
	f.removeMatManifest()
	_, err := Update(context.Background(), UpdateRequest{CfgPath: f.Cfg, Dir: f.Root, Force: true})
	if err == nil {
		t.Fatal("没有 SSH 主机时应失败")
	}
	if errors.Is(err, ErrDirtyWorkspace) {
		t.Fatalf("--force 不该再被脏闸门拒绝: %v", err)
	}
	if !strings.Contains(err.Error(), "没有配置 SSH 主机") {
		t.Errorf("失败点不在拨号之前: %v", err)
	}
}

// TestDedupe: Check 与 Fetch 都会说"取的是客制模块"，两条一模一样的提示只留一条。
func TestDedupe(t *testing.T) {
	if got := dedupe(nil); len(got) != 0 {
		t.Errorf("nil = %v", got)
	}
	one := []string{"只一条"}
	if got := dedupe(one); len(got) != 1 || got[0] != "只一条" {
		t.Errorf("单条被改了: %v", got)
	}
	got := dedupe([]string{"a", "b", "a", "c", "b"})
	if strings.Join(got, ",") != "a,b,c" {
		t.Errorf("去重结果 = %v，想要 a,b,c（保序）", got)
	}
}

// TestStaleText: CLI 把 stale 原因拼成一句话；没有原因时是空串（不输出空括号）。
func TestStaleText(t *testing.T) {
	if got := staleText(nil); got != "" {
		t.Errorf("无原因 = %q，想要空", got)
	}
	reasons := []devdata.StaleReason{
		{Code: "remote-mtime", Detail: "服务器上的文件更新了"},
		{Code: "missing-local", Detail: "没有清单"},
	}
	want := "remote-mtime: 服务器上的文件更新了；missing-local: 没有清单"
	if got := staleText(reasons); got != want {
		t.Errorf("staleText = %q，想要 %q", got, want)
	}
}

// TestAppendMdl: update 刷新材料后顺带更新模板切片（可选步骤，失败只给提示）。
func TestAppendMdl(t *testing.T) {
	env := newFixEnv(t)
	env.writeSlice(t)
	dst := filepath.Join(t.TempDir(), DirMdl)
	if warns := appendMdl(env.Cfg, "e", dst, false); len(warns) != 0 {
		t.Errorf("切片可复制时不该有提示: %v", warns)
	}
	if _, err := os.Stat(filepath.Join(dst, "slice", "code_a05.template")); err != nil {
		t.Errorf("切片没复制到工作区: %v", err)
	}

	// --no-mdl：不更新，也不提示
	other := filepath.Join(t.TempDir(), DirMdl)
	if warns := appendMdl(env.Cfg, "e", other, true); len(warns) != 0 {
		t.Errorf("--no-mdl 不该有提示: %v", warns)
	}
	if _, err := os.Stat(other); !os.IsNotExist(err) {
		t.Errorf("--no-mdl 不该建目录")
	}

	// 环境里没镜像切片 → 一条提示（不阻断 update）
	env2 := newFixEnv(t)
	warns := appendMdl(env2.Cfg, "e", filepath.Join(t.TempDir(), DirMdl), false)
	if len(warns) != 1 || !strings.Contains(warns[0], "模板切片没镜像") {
		t.Errorf("没镜像时提示 = %v", warns)
	}
}

// TestUpdate_PropagatesCheckError: Check 自己失败时（这里是 task.json 里的程序名非法）
// Update 要原样把错误带出去，不能继续往下走。
func TestUpdate_PropagatesCheckError(t *testing.T) {
	root := t.TempDir()
	task := &Task{Schema: TaskSchema, Prog: "BAD_PROG", Env: "e", Layout: DefaultLayout()}
	if err := Save(task, Paths{Root: root}); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, DirMaterials), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := Update(context.Background(), UpdateRequest{Dir: root})
	if err == nil || !strings.Contains(err.Error(), "程序名非法") {
		t.Fatalf("err = %v，想要程序名校验错误", err)
	}
}
