package host

// 清理机制的单测。
//
// 这一层守的不是"功能对不对"，而是"服务器上会不会多出一个 265 MB 的文件"：
// 清理相关的那几行代码（脚本里的 rm、defer、3 小时清扫）删掉任意一处，
// 功能测试全都还是绿的，只有真机上才会表现为磁盘慢慢变满。

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 打包脚本必须是**纯全量**：没有 -newer、没有 touch mark、没有 set -e。
// 前三者任一残留，都意味着服务器上会重新长出长期存在的状态文件。
func TestMirrorPackScriptIsFullOnly(t *testing.T) {
	s := mirrorPackScript("/tmp/.tt-mirror-x-source-abc", "/u1/t35prd", MirrorSource)
	for _, bad := range []string{"-newer", "touch", ".mark", "set -e"} {
		if strings.Contains(s, bad) {
			t.Errorf("全量脚本里不该出现 %q:\n%s", bad, s)
		}
	}
	for _, want := range []string{
		"tar -czf",             // 打包
		"-T ",                  // 吃清单
		"TDICT_MIRROR_OK",      // 成功标记
		"TDICT_MIRROR_CD",      // 进不去 TOP
		"TDICT_MIRROR_TARFAIL", // tar 失败也要有标记
		"rm -f",                // 失败路径自清
	} {
		if !strings.Contains(s, want) {
			t.Errorf("脚本缺少 %q:\n%s", want, s)
		}
	}
	// 清单成功路径必须删掉（否则每次拉取在临时目录里留一个几 MB 的清单）
	if !strings.Contains(s, `rm -f `+"\""+"/tmp/.tt-mirror-x-source-abc.list\"") {
		t.Errorf("成功路径没有删清单:\n%s", s)
	}
	// tar 失败时半截归档也要删
	if !strings.Contains(s, `rm -f "`+"/tmp/.tt-mirror-x-source-abc.tar.gz"+`" "`+"/tmp/.tt-mirror-x-source-abc.list"+`"`) {
		t.Errorf("tar 失败路径没有删半截归档:\n%s", s)
	}
}

// 打包输出解析：取最后一个标记；各失败标记都要给人话，不能落进"格式不认识"。
func TestParseMirrorPackOutput(t *testing.T) {
	cases := []struct {
		name    string
		out     string
		n       int
		size    int64
		wantErr string
	}{
		{"成功", "tar: 警告\nTDICT_MIRROR_OK 34704 278409071\n", 34704, 278409071, ""},
		{"无文件", "TDICT_MIRROR_NONE\n", 0, 0, ""},
		{"取最后一个标记", "TDICT_MIRROR_NONE\nTDICT_MIRROR_OK 5 100\n", 5, 100, ""},
		{"进不去TOP", "TDICT_MIRROR_CD\n", 0, 0, "进不去服务器 TOP"},
		{"find失败", "TDICT_MIRROR_FINDFAIL\n", 0, 0, "find 失败"},
		{"tar失败", "TDICT_MIRROR_TARFAIL\n", 0, 0, "tar 打包失败"},
		{"没有标记", "bash: tar: command not found\n", 0, 0, "未输出预期标记"},
		{"标记残缺", "TDICT_MIRROR_OK abc def\n", 0, 0, "解析打包标记失败"},
	}
	for _, c := range cases {
		n, size, err := parseMirrorPackOutput(c.out, "/u1/t35prd")
		if c.wantErr == "" {
			if err != nil {
				t.Errorf("%s: 不该报错: %v", c.name, err)
				continue
			}
			if n != c.n || size != c.size {
				t.Errorf("%s: got (%d,%d), want (%d,%d)", c.name, n, size, c.n, c.size)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), c.wantErr) {
			t.Errorf("%s: err = %v, want 含 %q", c.name, err, c.wantErr)
		}
	}
}

// 临时目录探测脚本：默认落 /tmp、只清 3 小时前的自己的文件、不误删别人的东西。
func TestMirrorTmpProbeScript(t *testing.T) {
	s := mirrorTmpProbeScript("")
	if !strings.Contains(s, `${TMPDIR:-/tmp}`) {
		t.Errorf("默认应走 ${TMPDIR:-/tmp}:\n%s", s)
	}
	for _, want := range []string{
		`-mmin +180`,        // 只清 3 小时前的
		`-exec rm -f {} +`,  // POSIX 兼容的删除（不用 -delete）
		`df -Pk`,            // 空间预检
		`TDICT_MIRROR_TMP `, // 成功标记
		`TDICT_MIRROR_TMPMISSING`,
		`TDICT_MIRROR_TMPNOWRITE`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("探测脚本缺少 %q:\n%s", want, s)
		}
	}
	if strings.Contains(s, "-delete") {
		t.Errorf("不该用 -delete（非 POSIX）:\n%s", s)
	}
	// 自定义目录必须被引号包住（可能含空格）
	custom := mirrorTmpProbeScript("/u2/tmp dir")
	if !strings.Contains(custom, `T="/u2/tmp dir"`) {
		t.Errorf("自定义目录没正确引用:\n%s", custom)
	}
}

// 空间预检：宁可在打包前拒绝，也不要压到一半 ENOSPC 留半截文件。
func TestMirrorCheckSpace(t *testing.T) {
	const mb = 1 << 20
	cases := []struct {
		name string
		tmp  serverTmp
		est  int64
		ok   bool
	}{
		{"空间充足", serverTmp{Dir: "/tmp", FreeKB: 2000 * 1024}, 200 * mb, true},
		{"刚好不够20%余量", serverTmp{Dir: "/tmp", FreeKB: 110 * 1024}, 100 * mb, false},
		{"足够20%余量", serverTmp{Dir: "/tmp", FreeKB: 130 * 1024}, 100 * mb, true},
		{"没有预估时按300MB兜底", serverTmp{Dir: "/tmp", FreeKB: 200 * 1024}, 0, false},
		{"没有预估但空间够", serverTmp{Dir: "/tmp", FreeKB: 400 * 1024}, 0, true},
		{"df没量到就不拦", serverTmp{Dir: "/tmp", FreeKB: 0}, 100 * mb, true},
	}
	for _, c := range cases {
		err := mirrorCheckSpace(c.tmp, c.est)
		if c.ok && err != nil {
			t.Errorf("%s: 不该报错: %v", c.name, err)
		}
		if !c.ok {
			if err == nil {
				t.Errorf("%s: 应该拒绝", c.name)
				continue
			}
			// 错误必须能指导用户：包含目录与改法
			if !strings.Contains(err.Error(), "/tmp") || !strings.Contains(err.Error(), "mirror.tmpDir") {
				t.Errorf("%s: 错误信息不可操作: %v", c.name, err)
			}
		}
	}
}

// 服务器清理命令：先杀掉还在写的 tar/ksh，再删两个文件；不能带通配。
func TestMirrorCleanScript(t *testing.T) {
	const base = "/tmp/.tt-mirror-abc-source-deadbeef"
	s := mirrorCleanScript(base)
	for _, want := range []string{"kill -TERM", "kill -KILL", "rm -f", `"$B.tar.gz"`, `"$B.list"`, `B="` + base + `"`} {
		if !strings.Contains(s, want) {
			t.Errorf("清理脚本缺少 %q: %s", want, s)
		}
	}
	if strings.ContainsAny(s, "*?") {
		t.Errorf("清理命令不该带通配（会误删）: %s", s)
	}
	// 必须排除自己：ps/awk 的命令行里也带着这个串，不排除就会杀掉清理脚本自己。
	if !strings.Contains(s, "$2!=me") {
		t.Errorf("清理脚本没有排除自身进程: %s", s)
	}
}

// 本地残留：只删 1 小时前的 .staging-* / .old-*，别碰正在写的和别人的东西。
func TestCleanLocalStaging(t *testing.T) {
	root := t.TempDir()
	envDir := filepath.Join(root, "109prd")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(root, ".staging-20200101-000000.000")
	fresh := filepath.Join(root, ".staging-fresh")
	oldOld := filepath.Join(root, ".old-20200101-000000.000")
	other := filepath.Join(root, "keepme")
	for _, d := range []string{old, fresh, oldOld, other} {
		if err := os.MkdirAll(filepath.Join(d, "erp"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// 把 old / oldOld 的 mtime 改到两小时前（目录本身，不含子项也能命中）
	two := time.Now().Add(-2 * time.Hour)
	for _, d := range []string{old, oldOld} {
		if err := os.Chtimes(d, two, two); err != nil {
			t.Fatal(err)
		}
	}

	cleanLocalStaging(envDir)

	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Errorf("两小时前的 .staging 应被删: %v", err)
	}
	if _, err := os.Stat(oldOld); !os.IsNotExist(err) {
		t.Errorf("两小时前的 .old 应被删: %v", err)
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Errorf("刚建的 .staging 不该动(可能正在写): %v", err)
	}
	if _, err := os.Stat(other); err != nil {
		t.Errorf("不认识的名字不该动: %v", err)
	}
	if _, err := os.Stat(envDir); err != nil {
		t.Errorf("镜像目录本身不该动: %v", err)
	}
}

// 没有 staging 目录时不该 panic（首次拉取、镜像根还没建）。
func TestCleanLocalStagingMissingRoot(t *testing.T) {
	cleanLocalStaging(filepath.Join(t.TempDir(), "nope", "env"))
}
