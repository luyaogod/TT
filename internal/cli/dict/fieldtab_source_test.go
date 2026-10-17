package dict

// `tt dict fieldtab` 的**通道选择**钉子：默认在线、显式双模式、不许静默降级。
//
// 用户决策（本轮）：字典查询要有两种模式 —— 在线（直查服务器字典表）与本地（本地
// SQLite 镜像），**默认在线**。所以这一组钉的是：
//
//  1. 不给 `--source` ⇒ **在线通道被调用、本地通道调用 0 次**（缺省 = 在线）；
//  2. `--source local` / `--source online` 各自走对应通道（反过来也验）；
//  3. 在线不可用（没环境 / 连不上）⇒ 明确报错、指引里含 `--source local`，
//     且**本地通道调用 0 次**（契约 A15：拿不准不猜，不许拿本地库假装在线成功）；
//  4. 本地库不存在 ⇒ 报错、指引含 `tt dict sync`，并按 A31 归 **missing（确定缺失）**，
//     不是 unknown；
//  5. 输出里看得见模式：JSON 的新键 + 人读的一行；
//  6. 既有契约不破：`fieldtab` 原有的键一个不少；
//  7. 反证两条（跑过、见过红）：
//     ① 把缺省改成 local ⇒ 第 1 条红（DefValue 与"在线被调用"双断言）；
//     ② 把"在线不可用"改成静默回落本地 ⇒ 第 3 条红（计数 + e2e 都会红）。
//
// 在线通道**不连真机**：第三条用真实现 `openFieldtabOnline` 在"没配环境"这一层失败，
// 碰都碰不到网络；e2e 那条更狠 —— 本地库**明明能用**，缺省=在线仍必须失败。

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tt/internal/cli/common"
	"tt/internal/dict/formfields"
	"tt/internal/testkit"

	"github.com/spf13/cobra"

	_ "modernc.org/sqlite"
)

// ─────────────────────────────────────────────────────────────────────────────
// 1. 缺省 = 在线
// ─────────────────────────────────────────────────────────────────────────────

func TestFieldtabSourceDefaultIsOnline(t *testing.T) {
	resetDictPackageState(t)

	// ① 开关存在，且**缺省值就是 online**（"默认在线"要写在缺省值里，不靠注释承诺）。
	f := fieldtabCmd.Flags().Lookup("source")
	if f == nil {
		t.Fatal("tt dict fieldtab 没有 --source 开关：双模式选择必须显式可点")
	}
	if f.DefValue != string(SourceOnline) {
		t.Errorf("--source 的缺省值 = %q，want %q（用户决策：默认在线）", f.DefValue, SourceOnline)
	}

	// ② 解析层：空串（"没传"）与 "online" 都落在线。
	for _, in := range []string{"", "online", "ONLINE", " online "} {
		if got, err := parseSourceMode(in); err != nil || got != SourceOnline {
			t.Errorf("parseSourceMode(%q) = %q, %v；want %q, nil", in, got, err, SourceOnline)
		}
	}

	// ③ 分发层：缺省这一轮**在线被调用、本地没被调用**。
	mode, err := fieldtabSourceMode()
	if err != nil {
		t.Fatalf("取缺省通道失败：%v", err)
	}
	if mode != SourceOnline {
		t.Fatalf("不传 --source 时通道 = %q，want %q", mode, SourceOnline)
	}
	var n openerCalls
	d, err := openFieldtabDict(mode, n.opener(SourceOnline), n.opener(SourceLocal))
	if err != nil {
		t.Fatalf("缺省通道该能打开：%v", err)
	}
	defer d.Close()
	if n.online != 1 {
		t.Errorf("在线通道被调用 %d 次，want 1（缺省=在线）", n.online)
	}
	if n.local != 0 {
		t.Errorf("本地通道被调用 %d 次，want 0（缺省=在线，不许顺带开本地库）", n.local)
	}
	if d.Mode != SourceOnline {
		t.Errorf("通道标记 = %q，want %q（输出里要照它写 dictSource）", d.Mode, SourceOnline)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 2. 显式切模式：两条通道各走各的
// ─────────────────────────────────────────────────────────────────────────────

func TestFieldtabSourceExplicitSwitch(t *testing.T) {
	cases := []struct {
		flag       string
		want       SourceMode
		wantOnline int
		wantLocal  int
	}{
		{"local", SourceLocal, 0, 1},
		{"LOCAL", SourceLocal, 0, 1},
		{"online", SourceOnline, 1, 0},
		{"", SourceOnline, 1, 0},
	}
	for _, c := range cases {
		resetDictPackageState(t)
		fieldtabSourceName = c.flag // 等价于命令行 - --source <c.flag>（空串 = 没传）
		mode, err := fieldtabSourceMode()
		if err != nil {
			t.Errorf("--source %q 解析失败：%v", c.flag, err)
			continue
		}
		if mode != c.want {
			t.Errorf("--source %q → %q，want %q", c.flag, mode, c.want)
			continue
		}
		var n openerCalls
		d, err := openFieldtabDict(mode, n.opener(SourceOnline), n.opener(SourceLocal))
		if err != nil {
			t.Errorf("--source %q 打开失败：%v", c.flag, err)
			continue
		}
		d.Close()
		if n.online != c.wantOnline || n.local != c.wantLocal {
			t.Errorf("--source %q：在线 %d 次 / 本地 %d 次，want 在线 %d / 本地 %d",
				c.flag, n.online, n.local, c.wantOnline, c.wantLocal)
		}
	}
}

// 认不出来的值必须报错并列出可选项 —— **不猜**（A15）。
func TestFieldtabSourceRejectsUnknownValue(t *testing.T) {
	resetDictPackageState(t)
	fieldtabSourceName = "server"
	if _, err := fieldtabSourceMode(); err == nil {
		t.Fatal("--source server 该报错（只认 online/local）")
	} else {
		msg := err.Error()
		for _, want := range []string{"online", "local"} {
			if !strings.Contains(msg, want) {
				t.Errorf("错误文案该列出可选项 %q：%q", want, msg)
			}
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 3. 在线不可用：明确报错 + 指引，且不许偷偷走本地（A15）
// ─────────────────────────────────────────────────────────────────────────────

func TestFieldtabOnlineUnavailableIsErrorNotFallback(t *testing.T) {
	resetDictPackageState(t)
	// 一份"什么环境都没配"的配置：解析环境这一步就失败，连不到网络。
	writeConfig(t, `{"schemaVersion": 2}`)
	common.Env = ""
	fieldtabSourceName = string(SourceOnline)

	var n openerCalls
	_, err := openFieldtabDict(SourceOnline, openFieldtabOnline, n.opener(SourceLocal))
	if err == nil {
		t.Fatal("在线拿不到环境时必须报错 —— 静默回落本地库会把「没查到」读成「查过了没有」")
	}
	msg := err.Error()
	if !strings.Contains(msg, "--source local") {
		t.Errorf("指引里必须给出可执行的那一步 --source local：\n%s", msg)
	}
	if !strings.Contains(msg, "--env") {
		t.Errorf("指引里该同时给出「配置 --env <环境名>」这条路：\n%s", msg)
	}
	if n.local != 0 {
		t.Errorf("本地通道被调用 %d 次，want 0（A15：不许拿本地库假装在线成功）", n.local)
	}
	if dataSrc != nil {
		t.Error("打开失败不该留下半个数据源")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 4. 本地库不存在：报错 + 指引 tt dict sync（A31 的 missing，不是 unknown）
// ─────────────────────────────────────────────────────────────────────────────

func TestFieldtabLocalMissingSaysHowToSync(t *testing.T) {
	resetDictPackageState(t)
	writeConfig(t, `{}`) // 数据目录 = 临时目录 ⇒ 里面没有 erp_data.db
	t.Setenv("TDICT_DB", "")
	dbPath = filepath.Join(t.TempDir(), "肯定不存在.db")
	fieldtabSourceName = string(SourceLocal)

	var n openerCalls
	_, err := openFieldtabDict(SourceLocal, n.opener(SourceOnline), openFieldtabLocal)
	if err == nil {
		t.Fatal("本地库不存在时必须报错（A31：这是确定缺失，不是「查不了」）")
	}
	msg := err.Error()
	for _, want := range []string{"tt dict sync", "missing", "本地数据库文件未找到"} {
		if !strings.Contains(msg, want) {
			t.Errorf("错误文案该含 %q：\n%s", want, msg)
		}
	}
	if n.online != 0 {
		t.Errorf("在线通道被调用 %d 次，want 0（--source local 就是本地，不许换通道）", n.online)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 5–6. 输出里看得见模式，且既有键一个不少
// ─────────────────────────────────────────────────────────────────────────────

// JSON 载荷：既有键原样保留，只新增 dictSource。
func TestFieldtabOutKeepsExistingKeysAndAddsDictSource(t *testing.T) {
	tab := sampleTable()

	plain := marshalMap(t, tab)
	wrapped := marshalMap(t, fieldtabOut{Table: tab, DictSource: string(SourceLocal)})

	for k := range plain {
		if _, ok := wrapped[k]; !ok {
			t.Errorf("既有顶层键 %q 在加了 dictSource 之后不见了", k)
		}
	}
	// 契约里点名的那几个键（顶层 + 字段级）逐个在。
	for _, k := range []string{"module", "form", "fields", "duplicates", "skipped",
		"labelResolved", "labelUnresolved", "labelSource"} {
		if _, ok := wrapped[k]; !ok {
			t.Errorf("既有契约键 %q 不见了", k)
		}
	}
	if wrapped["dictSource"] != string(SourceLocal) {
		t.Errorf("dictSource = %v，want %q", wrapped["dictSource"], SourceLocal)
	}
	fields, _ := wrapped["fields"].([]any)
	if len(fields) == 0 {
		t.Fatal("fields 空了")
	}
	first, _ := fields[0].(map[string]any)
	for _, k := range []string{"colName", "fieldId", "tag", "label", "labelState", "labelFrom"} {
		if _, ok := first[k]; !ok {
			t.Errorf("字段级既有键 %q 不见了", k)
		}
	}
}

// e2e（本地通道，真 SQLite 库 + 真 .42f）：--json 里有 dictSource，中文名真解析到。
func TestFieldtabE2ELocalSourceInJSON(t *testing.T) {
	root := setupLocalMirror(t)

	out := runDict(t, "fieldtab", "t100/demofm", "--root", root, "--source", "local")
	env := decodeEnvelope(t, out)
	data, ok := env["data"].(map[string]any)
	if !ok {
		t.Fatalf("信封里没有 data 对象：\n%s", out)
	}
	if data["dictSource"] != string(SourceLocal) {
		t.Errorf("data.dictSource = %v，want %q（模式必须在 JSON 里可见）", data["dictSource"], SourceLocal)
	}
	if data["labelSource"] != formfields.LabelSourceDict {
		t.Errorf("labelSource = %v，want %q（本地库在，就该真接上）", data["labelSource"], formfields.LabelSourceDict)
	}
	if data["labelResolved"] != float64(1) {
		t.Errorf("labelResolved = %v，want 1（真 SQLite 里的 dzeb_t 给出中文名）", data["labelResolved"])
	}
	fld := firstField(t, data)
	if fld["label"] != "演示栏位" || fld["labelState"] != string(formfields.LabelResolved) {
		t.Errorf("中文名没解析到：label=%v labelState=%v", fld["label"], fld["labelState"])
	}
}

// e2e（人读）：多出一行"这一轮用的是哪种"，原有表头与"未解析"规则不变。
func TestFieldtabE2ELocalSourceInText(t *testing.T) {
	root := setupLocalMirror(t)
	common.Format = "table" // 人读形态

	out := runDict(t, "fieldtab", "t100/demofm", "--root", root, "--source", "local")
	if !strings.Contains(out, "字典源: 本地") {
		t.Errorf("人读输出里该有一行写清这一轮用的是本地通道：\n%s", out)
	}
	for _, want := range []string{"colName", "中文名", "已解析"} {
		if !strings.Contains(out, want) {
			t.Errorf("既有的人读列不见了：%q\n%s", want, out)
		}
	}
}

// 在线通道的人读一行（不连真机：直接构造一个已打开的通道对象）。
func TestFieldtabTextShowsOnlineSource(t *testing.T) {
	resetDictPackageState(t)
	d := &fieldtabDict{Mode: SourceOnline, Where: "生产环境"}
	tab := sampleTable()
	_, out := captureFieldtab(t, func() { printFieldtab(tab, d) })
	if !strings.Contains(out, "字典源: 在线") {
		t.Errorf("人读输出该写清用的是在线通道：\n%s", out)
	}
	if !strings.Contains(out, "生产环境") {
		t.Errorf("在线通道该把环境名写出来（否则「在线」等于没说）：\n%s", out)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 7. 反证的核心：缺省=在线时，**哪怕本地库能用**，也必须失败（不许静默降级）
// ─────────────────────────────────────────────────────────────────────────────

func TestFieldtabE2EDefaultOnlineNeverUsesLocalMirror(t *testing.T) {
	resetDictPackageState(t)
	root := setupLocalMirror(t) // 本地库**在**，而且能解析出中文名

	// ① 缺省（不传 --source）、没有环境 ⇒ 必须报错。
	out, err := runDictErr(t, "fieldtab", "t100/demofm", "--root", root)
	if err == nil {
		t.Fatalf("缺省=在线且没有任何环境时必须失败；它却成功了，说明偷偷用了本地库（A15）：\n%s", out)
	}
	if !strings.Contains(err.Error(), "--source local") {
		t.Errorf("错误里该给下一步 --source local：%v", err)
	}
	if strings.Contains(out, `"dictSource"`) {
		t.Errorf("失败时不该产出结果信封：\n%s", out)
	}

	// ② 同一套夹具换成 --source local ⇒ 成功。证明 ① 不是因为"本地也不成"。
	out = runDict(t, "fieldtab", "t100/demofm", "--root", root, "--source", "local")
	data, ok := decodeEnvelope(t, out)["data"].(map[string]any)
	if !ok {
		t.Fatalf("本地通道该成功：\n%s", out)
	}
	if data["dictSource"] != string(SourceLocal) || data["labelResolved"] != float64(1) {
		t.Errorf("本地通道该真接上字典：dictSource=%v labelResolved=%v",
			data["dictSource"], data["labelResolved"])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 夹具与工具
// ─────────────────────────────────────────────────────────────────────────────

// openerCalls 记两条通道各被调用几次 —— "没偷偷改走另一条"全靠这两个计数说话。
type openerCalls struct{ online, local int }

func (c *openerCalls) opener(mode SourceMode) dictOpener {
	return func() (*fieldtabDict, error) {
		if mode == SourceOnline {
			c.online++
		} else {
			c.local++
		}
		return &fieldtabDict{Mode: mode, Where: "fake/" + string(mode)}, nil
	}
}

// synth42f 合成 .42f：一条明细列（cmt_/lbl_/colName 齐全）+ 一条 RecordView 声明它属于 demo_t。
// 与 internal/dict/formfields 的合成夹具同一形状（那里不能 import，只能同形）。
const synth42f = `<?xml version="1.0" encoding="UTF-8"?>
<Form name="demofm.tmp" build="3.21.03">
  <TableColumn name="formonly.b_demo001" colName="b_demo001" sqlType="SMALLINT" fieldId="7" sqlTabName="formonly" tabIndex="1" text="lbl_demo001">
    <Edit width="8" comment="cmt_demo001"><LStr comment="cmt_demo001"/></Edit>
  </TableColumn>
  <RecordView tabName="demo_t">
    <Link colName="b_demo001" fieldIdRef="7"/>
  </RecordView>
</Form>`

// setupLocalMirror 造一份"本地镜像确实可用"的现场：临时配置目录（= 数据目录）+ 真 SQLite
// 字典库 + 一张 .42f。返回 42f 根目录。
func setupLocalMirror(t *testing.T) string {
	t.Helper()
	resetDictPackageState(t)
	dir := tempConfigDir(t) // common.ConfigPath → <dir>/config.json；数据目录 = <dir>
	t.Setenv("TDICT_DB", "")
	makeLocalDict(t, filepath.Join(dir, "erp_data.db"))

	root := t.TempDir()
	mod := filepath.Join(root, "t100")
	if err := os.MkdirAll(mod, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mod, "demofm.42f"), []byte(synth42f), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// makeLocalDict 建一张最小可用的字典库：dzeb_t + dzebl_t，一条 demo_t.demo001 = 演示栏位。
// 列名与 db.QueryTable 的 SELECT 对齐（少一列就报 "no such column"，不会假绿）。
func makeLocalDict(t *testing.T, path string) {
	t.Helper()
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("建本地字典库失败：%v", err)
	}
	defer conn.Close()
	for _, stmt := range []string{
		`CREATE TABLE dzeb_t (dzeb001 TEXT, dzeb002 TEXT, dzeb003 TEXT, dzeb004 TEXT,
			dzeb005 TEXT, dzeb007 TEXT, dzeb008 TEXT, dzeb021 TEXT, dzeb024 TEXT)`,
		`CREATE TABLE dzebl_t (dzebl001 TEXT, dzebl002 TEXT, dzebl003 TEXT)`,
		`INSERT INTO dzeb_t (dzeb001, dzeb002, dzeb003, dzeb021)
			VALUES ('demo_t', 'demo001', '演示栏位', '1')`,
	} {
		if _, err := conn.Exec(stmt); err != nil {
			t.Fatalf("建本地字典库失败（%s）：%v", stmt, err)
		}
	}
}

// runDictErr 与 runDict 同一个命令树，但把错误交回给调用方（拒绝路径要用）。
func runDictErr(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var runErr error
	_, out := captureFieldtab(t, func() {
		root := newDictTestRoot(args...)
		runErr = root.Execute()
		testkit.ResetFlags(Group) // 下一次调用要像全新进程
	})
	return out, runErr
}

// newDictTestRoot 只挂 tt dict 组（全局开关不在树上：测试直接改 common 的包级 var）。
func newDictTestRoot(args ...string) *cobra.Command {
	root := &cobra.Command{Use: "tt", SilenceUsage: true, SilenceErrors: true}
	root.AddCommand(Group)
	root.SetArgs(append([]string{"dict"}, args...))
	return root
}

// captureFieldtab 抓一次命令的 stdout。捕获实现只有一处（testkit.CaptureStdout），
// 这里只是把"返回退出码"的签名适配成"无返回值"的那半边。
func captureFieldtab(t *testing.T, fn func()) (int, string) {
	t.Helper()
	return testkit.CaptureStdout(t, func() int { fn(); return 0 })
}

func marshalMap(t *testing.T, v any) map[string]any {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("序列化失败：%v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("反序列化失败：%v", err)
	}
	return m
}

func firstField(t *testing.T, data map[string]any) map[string]any {
	t.Helper()
	fields, _ := data["fields"].([]any)
	if len(fields) == 0 {
		t.Fatalf("data.fields 为空：%v", data)
	}
	f, ok := fields[0].(map[string]any)
	if !ok {
		t.Fatalf("fields[0] 不是对象：%v", fields[0])
	}
	return f
}
