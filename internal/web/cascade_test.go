package web

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"tt/internal/fep"
)

// 环境被删掉时，它的**录制库要跟着收**（用户 2026-10-15 定的策略：
// 用户主动录的那些长期留着，但环境没了它们就是没人认领的字节）。
// 环境可能是设置页删的（这条），也可能是手改 config.json 删的
// （那条走 `tt fep rec prune` 兜底，见 internal/cli/fep）。
func TestHostsPutCascadesRecordingLibrary(t *testing.T) {
	seed := `{"schemaVersion":3,"hosts":{"sshs":[
	  {"name":"要删的环境","host":"10.0.0.9","port":22,"user":"u","password":"p","zone":"36","db":{"type":"oracle","host":"10.0.0.9","port":1521,"service":"t35prd","accounts":[{"account":"ds","password":"x"}]}},
	  {"name":"留着的环境","host":"10.0.0.8","port":22,"user":"u","password":"p","zone":"36","db":{"type":"oracle","host":"10.0.0.8","port":1521,"service":"t35prd","accounts":[{"account":"ds","password":"x"}]}}]}}`
	s, cfgPath := newTestServer(t, seed)
	dataDir := filepath.Dir(cfgPath)

	// 两个环境各有一份录制（用户主动录的：住在 records/<环境>/<ENT>/<ID>/）
	seedRecordingDir := func(env, id string) string {
		dir := filepath.Join(dataDir, "records", env, "99", id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, id+"-1-rt2gdc.log"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	doomed := seedRecordingDir("要删的环境", "fep-doomed")
	kept := seedRecordingDir("留着的环境", "fep-kept")

	// 设置页保存：只留一个环境（等价于把另一个删掉）
	rec, _ := doJSON(t, s, http.MethodPut, "/api/hosts", map[string]any{
		"hosts": map[string]any{
			"sshs": []any{map[string]any{
				"name": "留着的环境", "host": "10.0.0.8", "port": 22, "user": "u", "password": "p", "zone": "36",
				"db": map[string]any{"type": "oracle", "host": "10.0.0.8", "port": 1521, "service": "t35prd",
					"accounts": []any{map[string]any{"account": "ds", "password": "x"}}},
			}},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("保存环境失败：%d %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(doomed); !os.IsNotExist(err) {
		t.Fatalf("被删环境的录制库该跟着删掉：%v", err)
	}
	if _, err := os.Stat(kept); err != nil {
		t.Fatalf("没被删的环境，它的录制不该动：%v", err)
	}
	// 库里那个环境的整段都没了（不是只删了一个录制目录）
	if _, err := os.Stat(filepath.Dir(doomed)); !os.IsNotExist(err) {
		t.Fatalf("该环境在库里整段都不该留：%v", err)
	}
}

// 改配置但**没删环境**（只改 host）时，一份录制都不许动。
func TestHostsPutKeepsRecordingsWhenNothingRemoved(t *testing.T) {
	seed := `{"schemaVersion":3,"hosts":{"sshs":[
	  {"name":"甲","host":"10.0.0.9","port":22,"user":"u","password":"p","zone":"36","db":{"type":"oracle","host":"10.0.0.9","port":1521,"service":"t35prd","accounts":[{"account":"ds","password":"x"}]}}]}}`
	s, cfgPath := newTestServer(t, seed)
	dir := filepath.Join(filepath.Dir(cfgPath), "records", "甲", "99", "fep-x")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	rec, _ := doJSON(t, s, http.MethodPut, "/api/hosts", map[string]any{
		"hosts": map[string]any{
			"sshs": []any{map[string]any{
				"name": "甲", "host": "10.0.0.10", "port": 22, "user": "u", "password": "p", "zone": "36",
				"db": map[string]any{"type": "oracle", "host": "10.0.0.10", "port": 1521, "service": "t35prd",
					"accounts": []any{map[string]any{"account": "ds", "password": "x"}}},
			}},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("保存失败：%d", rec.Code)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("只是改了 host，录制不该动：%v", err)
	}
}

// **改名不等于删除**：把环境改个名字保存，它名下的录像要一起搬到新名字下。
//
// 这条钉的是 2026-10-15 那次真实事故：设置页里把「环境名称」清空（前端静默补成 host-zone）
// 再保存，服务端只按名字集合判断"旧环境被删了"，于是级联清理把那个环境名下长期库里的
// 录像**全删了**（27 份文件 / 2.5 MB，`os.RemoveAll`，不可恢复）。名字是标签，录像才是
// 不可再生的东西 —— 改名必须走"搬家"。
func TestHostsPutRenamesEnvKeepsRecordings(t *testing.T) {
	seed := `{"schemaVersion":3,"hosts":{"activeEnv":"甲","sshs":[
	  {"name":"甲","host":"10.0.0.9","port":22,"user":"u","password":"p","zone":"36","db":{"type":"oracle","host":"10.0.0.9","port":1521,"service":"t35prd","accounts":[{"account":"ds","password":"x"}]}},
	  {"name":"乙","host":"10.0.0.8","port":22,"user":"u","password":"p","zone":"36","db":{"type":"oracle","host":"10.0.0.8","port":1521,"service":"t35prd","accounts":[{"account":"ds","password":"x"}]}}]}}`
	s, cfgPath := newTestServer(t, seed)
	dataDir := filepath.Dir(cfgPath)

	// 甲 名下有一份**真**录制（有身份文件 + 一个流量文件：ListIDs 靠后者认出这个 ID）
	id := "fep-rename-me"
	old := filepath.Join(dataDir, "records", "甲", "99", id)
	if err := os.MkdirAll(old, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, id+"-1-rt2gdc.log"), []byte("om 3 {{}}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := fep.WriteMeta(old, fep.RecordingMeta{
		ID: id, Path: fep.PathDebug, Prog: "bsft001_wf", Env: "甲", Ent: "99", Zone: "36",
		Fingerprint: fep.Fingerprint{Total: 3, Hash: "abcdef0123456789"},
	}); err != nil {
		t.Fatal(err)
	}

	// 设置页保存：甲 → 丙（**只有名字变了**，位置也不动）
	rec, _ := doJSON(t, s, http.MethodPut, "/api/hosts", map[string]any{
		"hosts": map[string]any{
			"sshs": []any{
				map[string]any{
					"name": "丙", "host": "10.0.0.9", "port": 22, "user": "u", "password": "p", "zone": "36",
					"db": map[string]any{"type": "oracle", "host": "10.0.0.9", "port": 1521, "service": "t35prd",
						"accounts": []any{map[string]any{"account": "ds", "password": "x"}}},
				},
				map[string]any{
					"name": "乙", "host": "10.0.0.8", "port": 22, "user": "u", "password": "p", "zone": "36",
					"db": map[string]any{"type": "oracle", "host": "10.0.0.8", "port": 1521, "service": "t35prd",
						"accounts": []any{map[string]any{"account": "ds", "password": "x"}}},
				},
			},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("保存环境失败：%d %s", rec.Code, rec.Body.String())
	}

	moved := filepath.Join(dataDir, "records", "丙", "99", id)
	if _, err := os.Stat(filepath.Join(moved, id+"-1-rt2gdc.log")); err != nil {
		t.Fatalf("改名后录像该在新名字下：%v", err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("旧名字下不该还留着：%v", err)
	}
	// 身份里的环境名也要跟着改：`rec reproduce` 是拿它去起会话的，
	// 留旧名字的话改名之后那些录像就再也回放不起来了（"环境不存在"）。
	mt, merr := fep.ReadMeta(moved, id)
	if merr != nil {
		t.Fatalf("身份文件该跟着搬过去：%v", merr)
	}
	if mt.Env != "丙" {
		t.Fatalf("身份里的环境名该改成新的（否则回放找不到环境）：%q", mt.Env)
	}
	// 另一个环境一点都不能动
	if _, err := os.Stat(filepath.Join(dataDir, "records", "乙")); !os.IsNotExist(err) {
		t.Fatalf("没被改名的环境不该凭空多出目录：%v", err)
	}
}

// 真删环境（少了一个、没补新的）仍然级联清 —— 这是用户定的策略，改名那条修复不该把它废掉。
func TestHostsPutStillCascadesWhenEnvReallyRemoved(t *testing.T) {
	seed := `{"schemaVersion":3,"hosts":{"sshs":[
	  {"name":"甲","host":"10.0.0.9","port":22,"user":"u","password":"p","zone":"36","db":{"type":"oracle","host":"10.0.0.9","port":1521,"service":"t35prd","accounts":[{"account":"ds","password":"x"}]}}]}}`
	s, cfgPath := newTestServer(t, seed)
	dataDir := filepath.Dir(cfgPath)
	dir := filepath.Join(dataDir, "records", "甲", "99", "fep-gone")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// 用一个别的环境替换掉它（删一个、加一个，但加的那个指向**另一台机器**：
	// 身份对不上，就不该被当成改名）
	rec, _ := doJSON(t, s, http.MethodPut, "/api/hosts", map[string]any{
		"hosts": map[string]any{
			"sshs": []any{map[string]any{
				"name": "全新的环境", "host": "10.9.9.9", "port": 22, "user": "u", "password": "p", "zone": "99",
				"db": map[string]any{"type": "oracle", "host": "10.9.9.9", "port": 1521, "service": "t99",
					"accounts": []any{map[string]any{"account": "ds", "password": "x"}}},
			}},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("保存失败：%d %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("真的删掉的环境，它的录像该级联清掉：%v", err)
	}
}

// envEntries 按顺序取条目：没有名字的照留（改名判定第一道判据就是"同一个下标"）。
func TestEnvEntriesKeepsOrderAndNameless(t *testing.T) {
	sec := map[string]any{"sshs": []any{
		map[string]any{"name": "甲"},
		"不是对象",
		map[string]any{"host": "无名"},
		map[string]any{"name": ""},
	}}
	got := envEntries(sec)
	if len(got) != 3 {
		t.Fatalf("该取到 3 条（跳过那个不是对象的）：%v", got)
	}
	if got[0].name != "甲" || got[0].idx != 0 {
		t.Errorf("第一条该是 甲（下标 0）：%+v", got[0])
	}
	if got[1].name != "" || got[1].idx != 2 {
		t.Errorf("没有名字的也要留，下标是它在列表里的位置：%+v", got[1])
	}
}

// diffEnvs 的三道判据。名字集合分不出"改名"与"删一个加一个"，而两者的后果天差地别
// （一个要搬家、一个要级联删），所以这道判据值得单独钉住。
func TestDiffEnvs(t *testing.T) {
	env := func(name, host string) map[string]any {
		return map[string]any{"name": name, "host": host, "port": 22, "zone": "36"}
	}
	entry := func(i int, m map[string]any) envEntry {
		n, _ := m["name"].(string)
		return envEntry{idx: i, name: n, raw: m}
	}

	// ① 同一个下标、只有名字变了 → 改名
	renames, removed := diffEnvs(
		[]envEntry{entry(0, env("甲", "10.0.0.1"))},
		[]envEntry{entry(0, env("丙", "10.0.0.1"))},
	)
	if len(renames) != 1 || renames[0].From != "甲" || renames[0].To != "丙" || len(removed) != 0 {
		t.Fatalf("同位置改名该算改名：%v %v", renames, removed)
	}

	// ② 换了位置、但身份相同 → 也算改名（重排 + 改名）
	renames, removed = diffEnvs(
		[]envEntry{entry(0, env("甲", "10.0.0.1")), entry(1, env("乙", "10.0.0.2"))},
		[]envEntry{entry(0, env("丙", "10.0.0.1")), entry(1, env("乙", "10.0.0.2"))},
	)
	if len(renames) != 1 || renames[0].To != "丙" || len(removed) != 0 {
		t.Fatalf("身份相同的一对一该算改名：%v %v", renames, removed)
	}

	// ③ 判不出来时偏向"搬"：数量相等就按下标配对（宁可留下字节，也不删不可再生的录像）
	renames, removed = diffEnvs(
		[]envEntry{entry(0, env("甲", "10.0.0.1"))},
		[]envEntry{entry(0, env("丁", "10.9.9.9"))},
	)
	if len(renames) != 1 || len(removed) != 0 {
		t.Fatalf("数量相等时该按改名处理（误判成删除的代价是不可再生的录像）：%v %v", renames, removed)
	}

	// ④ 真的删掉（删得比加得多）→ 级联清
	renames, removed = diffEnvs(
		[]envEntry{entry(0, env("甲", "10.0.0.1")), entry(1, env("乙", "10.0.0.2"))},
		[]envEntry{entry(0, env("乙", "10.0.0.2"))},
	)
	if len(removed) != 1 || removed[0] != "甲" || len(renames) != 0 {
		t.Fatalf("少了一个环境该级联清：%v %v", renames, removed)
	}

	// ⑤ 一个都没动 → 什么都不发生（这条最常走）
	renames, removed = diffEnvs(
		[]envEntry{entry(0, env("甲", "10.0.0.1"))},
		[]envEntry{entry(0, env("甲", "10.0.0.1"))},
	)
	if len(renames) != 0 || len(removed) != 0 {
		t.Fatalf("没改名没删除就不该有动作：%v %v", renames, removed)
	}
}
