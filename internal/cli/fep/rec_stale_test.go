package fep

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"tt/internal/cli/common"
	"tt/internal/fep"
	"tt/internal/testkit"
)

// ── A24 的命令面：`tt recording mark-stale <id> [--reason <文字>]` ─────────────
//
// 三条判据（用户决策 D-F2 + 契约 §7 A24）：
//
//	1. **只打标记**：身份文件多 stale/reason/staleAt，录制目录里一个文件不少、一字节不变；
//	2. **列表要显示**「不可回放（需重录）」（文本与 --json 两种输出都要能看出）；
//	3. **不加跳过开关**：`mark-stale` 只有 --reason；回放入口没有 --skip-gate / --force 之类。

// staleRecFixture 造一份"有身份、有录制文件"的录制，返回目录与 ID。
//
// 录制件是真的帧流（不是随便几个字节）：`rec list` 会当场从字节复算指纹，
// 而"读不了表"的行会走另一条渲染分支 —— 夹具得先像个真件。
func staleRecFixture(t *testing.T) (dir, id string) {
	t.Helper()
	dir = t.TempDir()
	id = "fep-20261017-101010"
	writeFrames(t, filepath.Join(dir, id+"-1-rt2gdc.log"),
		`om 1 {{an 0 Window 4097 {{name "w_axmt500"}} {{`+
			`{an 4097 FormField 6000 {{name "formonly.b_target"}{colName "b_target"}{fieldId "5"}{text "目标字段"}} {}} `+
			`}}}}`,
	)
	writeFrames(t, filepath.Join(dir, id+"-1-gdc2rt.log"),
		`event _om 1{}{{ConfigureEvent 0{{idRef "6000"}{ value "X"}}}}`,
	)
	started := time.Date(2026, 10, 17, 10, 10, 10, 0, time.UTC)
	if _, err := fep.WriteMeta(dir, fep.RecordingMeta{
		ID: id, Path: fep.PathRunnerRR, Prog: "azzp000", Env: "t35dev", Ent: "1",
		Started: started, Finished: started.Add(time.Minute),
		Origin: fep.OriginUser, Retention: fep.RetainLibrary,
	}); err != nil {
		t.Fatalf("写身份文件: %v", err)
	}
	return dir, id
}

// recFileSet 把录制文件（不含身份文件）逐个哈希 —— "标记不删不改"的判据。
func recFileSet(t *testing.T, dir string) map[string]string {
	t.Helper()
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("读目录: %v", err)
	}
	out := map[string]string{}
	for _, e := range ents {
		if e.IsDir() || strings.HasSuffix(e.Name(), fep.MetaSuffix) {
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


// withTableOutput 把输出切成"人读文本"（default 是 json），结束后还原。
func withTableOutput(t *testing.T) {
	t.Helper()
	oldFmt, oldJSON, oldFrom := common.Format, common.JSON, uiFrom
	common.Format, common.JSON = "table", false
	t.Cleanup(func() { common.Format, common.JSON, uiFrom = oldFmt, oldJSON, oldFrom })
}

// TestMarkStaleOnlyMarksAndListSaysIt 是 A24 的主判据。
func TestMarkStaleOnlyMarksAndListSaysIt(t *testing.T) {
	dir, id := staleRecFixture(t)
	withTableOutput(t)
	uiFrom = dir // `rec list` 与 `mark-stale` 都认 --from 的那一处

	before := recFileSet(t, dir)

	// ① 标记之前：列表里不该有那句话
	var err error
	_, out := testkit.CaptureStdout(t, func() int { err = recListCmd.RunE(recListCmd, nil); return 0 })
	if err != nil {
		t.Fatalf("rec list: %v", err)
	}
	if strings.Contains(out, fep.StaleNote) {
		t.Fatalf("还没标记就说了「%s」：\n%s", fep.StaleNote, out)
	}

	// ② 打标记
	recStaleReason = "该件早于身份上线（回放会被拒）"
	_, out = testkit.CaptureStdout(t, func() int { err = recStaleCmd.RunE(recStaleCmd, []string{id}); return 0 })
	if err != nil {
		t.Fatalf("mark-stale: %v", err)
	}
	if !strings.Contains(out, id) || !strings.Contains(out, "需重录") {
		t.Fatalf("打标记后要说清发生了什么：\n%s", out)
	}

	// ③ 元数据：stale + 原因 + 时间
	mt, merr := fep.ReadMeta(dir, id)
	if merr != nil {
		t.Fatalf("读回身份：%v", merr)
	}
	if !mt.Stale || mt.StaleReason != recStaleReason || mt.StaleAt.IsZero() {
		t.Fatalf("元数据里没落标记：%+v", mt)
	}
	if mt.Prog != "azzp000" || mt.Env != "t35dev" {
		t.Fatalf("标记把既有身份弄丢了：%+v", mt)
	}

	// ④ 录制目录一个文件不少、一字节不变
	after := recFileSet(t, dir)
	if len(before) != len(after) {
		t.Fatalf("录制文件个数变了：%d → %d（标记绝不许删文件）", len(before), len(after))
	}
	names := make([]string, 0, len(before))
	for n := range before {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if after[n] != before[n] {
			t.Fatalf("录制文件 %s 被改了（标记只许动身份文件）", n)
		}
	}

	// ⑤ 列表要显示「不可回放（需重录）」
	_, out = testkit.CaptureStdout(t, func() int { err = recListCmd.RunE(recListCmd, nil); return 0 })
	if err != nil {
		t.Fatalf("rec list: %v", err)
	}
	if !strings.Contains(out, fep.StaleNote) {
		t.Fatalf("列表里没有「%s」：\n%s", fep.StaleNote, out)
	}
	t.Logf("标记后的 rec list：\n%s", out)

	// ⑥ --json 也要能看出（默认输出形态就是 json）
	common.Format, common.JSON = "", true
	var jerr error
	_, jout := testkit.CaptureStdout(t, func() int { jerr = recListCmd.RunE(recListCmd, nil); return 0 })
	if jerr != nil {
		t.Fatalf("rec list --json: %v", jerr)
	}
	for _, key := range []string{`"stale": true`, `"staleNote"`} {
		if !strings.Contains(jout, key) {
			t.Fatalf("json 列表里没有 %s：\n%s", key, jout)
		}
	}
}

// TestNoSkipSwitchAnywhere：用户明令 —— 不加任何"跳过闸门"的开关（A24 第一句）。
func TestNoSkipSwitchAnywhere(t *testing.T) {
	for _, name := range []string{"skip-gate", "force", "no-identity", "allow-legacy"} {
		if f := recReplayCmd.Flags().Lookup(name); f != nil {
			t.Errorf("rec replay 上有 --%s：A24 明令不加跳过闸门", name)
		}
		if f := recStaleCmd.Flags().Lookup(name); f != nil {
			t.Errorf("rec mark-stale 上有 --%s：它只该有 --reason", name)
		}
	}
	if f := recStaleCmd.Flags().Lookup("reason"); f == nil {
		t.Error("rec mark-stale 缺 --reason")
	}
}
