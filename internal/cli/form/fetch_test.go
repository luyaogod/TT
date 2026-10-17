package form

// `tt form fetch` 与 `tt form cache` 的**先红后绿**（契约 C2 / C5 / A13）。
//
// 纪律照本包既有测试：
//   - 取件通道一律走假 `Source`（`injectSource` 换掉 `sourceFor` 这一个裁决点），
//     或者 `--dir <临时镜像>`；**不连真机、不起会话**；
//   - 缓存落点是 `tempDataDir(t)` 造的临时数据目录，跑完随 t.TempDir 一起消失；
//   - **不删任何用户文件**：这些用例只在自己造出来的临时目录里读写。
//
// 这几条测试各自钉住哪一条要求，写在每条用例头上。

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// fetchForms 是那 5 张假表单的名字（都过得了名字白名单，能拼进缓存路径）。
var fetchForms = []string{"axmt500_wf", "axmt501_wf", "axmt502_wf", "axmt503_wf", "axmt504_wf"}

// rawFormXML 给每张表单一份**互不相同**的原始 XML（几百字节内的合成件）。
func rawFormXML(name string) []byte {
	return []byte(`<Form name="` + name + `" build="3.21.03"><FormField colName="` + name + `001"/></Form>`)
}

// multiSource 是"每张表单各有各的原始字节"的假件，并记下 Form 被调用了几次。
//
// 为什么要数调用次数：`--force` 与"二次命中"的唯一判据就是"底层到底被问了几次"——
// 只看输出里的计数，实现完全可以在不取件的情况下把数字写成 5。
type multiSource struct {
	fakeSource
	bodies map[string][]byte // 键：module + "/" + form
	failOn map[string]error  // 键同上：让指定表单取件失败
	calls  int               // Form 被调用的次数
	asked  []string          // 按顺序记下问过哪几张（便于断言"整批都问了"）
	err    error             // 非 nil 时 List 也失败
}

func (s *multiSource) Form(_ context.Context, _, module, form string) ([]byte, error) {
	s.calls++
	s.asked = append(s.asked, module+"/"+form)
	k := module + "/" + form
	if err, ok := s.failOn[k]; ok {
		return nil, err
	}
	b, ok := s.bodies[k]
	if !ok {
		return nil, fmt.Errorf("假件里没有 %s", k)
	}
	return b, nil
}

func (s *multiSource) List(ctx context.Context, env, module string) ([]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.fakeSource.List(ctx, env, module)
}

// fiveFormSource 造一份"5 张表单都取得到"的假件。
func fiveFormSource() *multiSource {
	bodies := make(map[string][]byte, len(fetchForms))
	for _, n := range fetchForms {
		bodies["demomod/"+n] = rawFormXML(n)
	}
	return &multiSource{
		fakeSource: fakeSource{names: append([]string{}, fetchForms...)},
		bodies:     bodies,
		failOn:     map[string]error{},
	}
}

// formCacheDir 是本次运行的表单缓存目录（与实现走同一条：数据目录 + config.CacheDir）。
func formCacheDir(t *testing.T, dataDir string) string {
	t.Helper()
	c, err := NewCacheFromDataDir(dataDir)
	if err != nil {
		t.Fatalf("算不出表单缓存落点: %v", err)
	}
	return c.Dir
}

// cachedFormFiles 列出缓存目录里的 `.42f`（相对路径，排序）。目录不存在返回 nil。
func cachedFormFiles(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	if _, err := os.Stat(dir); err != nil {
		return nil
	}
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, werr error) error {
		if werr != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".42f") {
			rel, rerr := filepath.Rel(dir, p)
			if rerr != nil {
				return rerr
			}
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("走缓存目录失败: %v", err)
	}
	sort.Strings(out)
	return out
}

// decodeDoc 把 `--json` 的信封解成 map。
func decodeDoc(t *testing.T, out string) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("--json 输出不是 JSON: %v\n%s", err, out)
	}
	return doc
}

// num 取 JSON 里的数字（encoding/json 解进 map 后是 float64）。
func num(t *testing.T, doc map[string]any, key string) int {
	t.Helper()
	v, ok := doc[key]
	if !ok {
		t.Fatalf("JSON 里没有 %q 键: %v", key, doc)
	}
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("JSON 的 %q 不是数字: %T", key, v)
	}
	return int(f)
}

// ── 1. 批量预热 ─────────────────────────────────────────────────────────────

// TestFormFetchPreheatsCache 钉住"列出该模块的表单 → 逐个取件 → 写进缓存"这条主链：
// 5 张表单 ⇒ 底层被问 5 次、缓存里 5 个文件、计数 5/5/0/0，且缓存内容是**原始 XML**。
func TestFormFetchPreheatsCache(t *testing.T) {
	data := tempDataDir(t)
	src := fiveFormSource()
	injectSource(t, src)

	out, err := runForm(t, "fetch", "--env", "demoenv", "--module", "demomod", "--json")
	if err != nil {
		t.Fatalf("fetch 失败: %v\n%s", err, out)
	}
	doc := decodeDoc(t, out)
	if got := num(t, doc, "total"); got != 5 {
		t.Errorf("total = %d, want 5", got)
	}
	if got := num(t, doc, "fetched"); got != 5 {
		t.Errorf("fetched（新取）= %d, want 5", got)
	}
	if got := num(t, doc, "cached"); got != 0 {
		t.Errorf("cached（命中缓存）= %d, want 0", got)
	}
	if got := num(t, doc, "failed"); got != 0 {
		t.Errorf("failed = %d, want 0", got)
	}
	if src.calls != 5 {
		t.Errorf("底层 Form 被调用 %d 次, want 5", src.calls)
	}

	dir := formCacheDir(t, data)
	files := cachedFormFiles(t, dir)
	if len(files) != 5 {
		t.Fatalf("缓存里 %d 个 .42f, want 5：%v", len(files), files)
	}
	for _, n := range fetchForms {
		want := "demoenv/demomod/" + n + ".42f"
		if !containsStr(files, want) {
			t.Errorf("缓存缺 %s（现有 %v）", want, files)
		}
		b, rerr := os.ReadFile(filepath.Join(dir, filepath.FromSlash(want)))
		if rerr != nil {
			t.Fatalf("读缓存 %s 失败: %v", want, rerr)
		}
		if string(b) != string(rawFormXML(n)) {
			t.Errorf("%s 的缓存字节与取回来的不一致：%q", want, b)
		}
	}
}

// TestFormFetchCachesRawXMLOnly 是**反证**（契约 C2.4）：缓存里必须是与取回来的
// 那串字节一字不差的**原始 XML**，不是解析结果。
//
// 把 Put 改成"缓存解析结果（JSON / 结构体）"→ 这条必红；show 的 golden 不会红，
// 因为它只看得见 show 的输出。
func TestFormFetchCachesRawXMLOnly(t *testing.T) {
	data := tempDataDir(t)
	injectSource(t, fiveFormSource())

	if _, err := runForm(t, "fetch", "--env", "demoenv", "--module", "demomod", "--format", "table"); err != nil {
		t.Fatalf("fetch 失败: %v", err)
	}
	dir := formCacheDir(t, data)
	for _, n := range fetchForms {
		p := filepath.Join(dir, "demoenv", "demomod", n+".42f")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("读缓存 %s 失败: %v", p, err)
		}
		if string(b) != string(rawFormXML(n)) {
			t.Errorf("缓存里不是原始 XML（C2.4）：%q", b)
		}
		if first := strings.TrimSpace(string(b)); strings.HasPrefix(first, "{") || strings.HasPrefix(first, "[") {
			t.Errorf("缓存里像解析结果（JSON）而不是原始 XML：%q", b)
		}
	}
	// 缓存目录里**只有** .42f 文件：没有元数据旁车（解析结果/索引），一个表单一份。
	var nonForm []string
	for _, f := range listAllFiles(t, dir) {
		if !strings.HasSuffix(strings.ToLower(f), ".42f") {
			nonForm = append(nonForm, f)
		}
	}
	if len(nonForm) != 0 {
		t.Errorf("缓存目录里多出了非 .42f 的文件（旁车/解析结果）：%v", nonForm)
	}
}

// ── 2. 二次命中 ─────────────────────────────────────────────────────────────

// TestFormFetchSecondRunHitsCache 钉住"再跑一次不该再取件"：
// 底层调用次数不变、计数变成 新取 0 / 命中 5。
func TestFormFetchSecondRunHitsCache(t *testing.T) {
	tempDataDir(t)
	src := fiveFormSource()
	injectSource(t, src)

	if _, err := runForm(t, "fetch", "--env", "demoenv", "--module", "demomod", "--format", "table"); err != nil {
		t.Fatalf("第一次 fetch 失败: %v", err)
	}
	if src.calls != 5 {
		t.Fatalf("第一次该取 5 张，实际 %d", src.calls)
	}

	out, err := runForm(t, "fetch", "--env", "demoenv", "--module", "demomod", "--json")
	if err != nil {
		t.Fatalf("第二次 fetch 失败: %v", err)
	}
	if src.calls != 5 {
		t.Errorf("第二次又去取件了（底层共 %d 次，want 5）：缓存没被用上", src.calls)
	}
	doc := decodeDoc(t, out)
	if got := num(t, doc, "fetched"); got != 0 {
		t.Errorf("新取 = %d, want 0", got)
	}
	if got := num(t, doc, "cached"); got != 5 {
		t.Errorf("命中缓存 = %d, want 5", got)
	}
	if got := num(t, doc, "total"); got != 5 {
		t.Errorf("total = %d, want 5（总数要算上命中的）", got)
	}
}

// ── 3. --force ─────────────────────────────────────────────────────────────

// TestFormFetchForceRefetchesAll 钉住 `--force`：忽略缓存全部重取。
//
// 反证：把 `--force` 改成"给了也照旧查缓存"→ 底层会停在 5 次、本用例红。
func TestFormFetchForceRefetchesAll(t *testing.T) {
	tempDataDir(t)
	src := fiveFormSource()
	injectSource(t, src)

	if _, err := runForm(t, "fetch", "--env", "demoenv", "--module", "demomod", "--format", "table"); err != nil {
		t.Fatalf("第一次 fetch 失败: %v", err)
	}
	if src.calls != 5 {
		t.Fatalf("第一次该取 5 张，实际 %d", src.calls)
	}

	out, err := runForm(t, "fetch", "--env", "demoenv", "--module", "demomod", "--force", "--json")
	if err != nil {
		t.Fatalf("--force fetch 失败: %v", err)
	}
	if src.calls != 10 {
		t.Errorf("--force 之后底层共被调用 %d 次, want 10（5 + 5：忽略缓存全部重取）", src.calls)
	}
	doc := decodeDoc(t, out)
	if got := num(t, doc, "fetched"); got != 5 {
		t.Errorf("--force 的新取 = %d, want 5", got)
	}
	if got := num(t, doc, "cached"); got != 0 {
		t.Errorf("--force 的命中缓存 = %d, want 0（不是查缓存，是全重取）", got)
	}
	if forced, _ := doc["forced"].(bool); !forced {
		t.Errorf("--force 该在输出里自报 forced=true: %v", doc)
	}
}

// ── 4. 失败不静默 ───────────────────────────────────────────────────────────

// TestFormFetchReportsFailures 钉住"取件失败不许静默吞"：
// 失败的表单名 + 原因要出现在输出里，其它表单照常写入，退出码非 0 且**不是 6**。
func TestFormFetchReportsFailures(t *testing.T) {
	data := tempDataDir(t)
	src := fiveFormSource()
	src.failOn["demomod/axmt503_wf"] = errors.New("SFTP: no such file")
	injectSource(t, src)

	out, err := runForm(t, "fetch", "--env", "demoenv", "--module", "demomod", "--json")
	if err == nil {
		t.Fatalf("有失败必须非 0 退出：\n%s", out)
	}
	if got := exitOf(t, err); got != exitFormUnknown {
		t.Errorf("退出码 = %d, want %d（取件失败，契约 A13）", got, exitFormUnknown)
	}
	if got := exitOf(t, err); got == exitMismatch {
		t.Errorf("有失败不能退 %d（那是 verify 的「体检不过」，A13 明写）", exitMismatch)
	}

	doc := decodeDoc(t, out)
	if got := num(t, doc, "total"); got != 5 {
		t.Errorf("total = %d, want 5", got)
	}
	if got := num(t, doc, "fetched"); got != 4 {
		t.Errorf("fetched = %d, want 4（失败那张不算取到）", got)
	}
	if got := num(t, doc, "failed"); got != 1 {
		t.Errorf("failed = %d, want 1", got)
	}
	if ok, _ := doc["ok"].(bool); ok {
		t.Errorf("有失败时 ok 该是 false: %v", doc)
	}
	fails, _ := doc["failures"].([]any)
	if len(fails) != 1 {
		t.Fatalf("failures 该恰好一条: %v", doc["failures"])
	}
	fl, _ := fails[0].(map[string]any)
	if fl["form"] != "axmt503_wf" {
		t.Errorf("failures[0].form = %v, want axmt503_wf（必须点名是哪张）", fl["form"])
	}
	if reason, _ := fl["reason"].(string); !strings.Contains(reason, "no such file") {
		t.Errorf("failures[0].reason = %q, want 含底层原因（不许只说「失败了」）", reason)
	}
	if !strings.Contains(err.Error(), "axmt503_wf") {
		t.Errorf("错误消息里也要点名失败的表单: %v", err)
	}

	// 其它照常写入：缓存里 4 个文件，失败那张没有
	dir := formCacheDir(t, data)
	files := cachedFormFiles(t, dir)
	if len(files) != 4 {
		t.Errorf("缓存里 %d 个 .42f, want 4（失败那张不写）: %v", len(files), files)
	}
	if containsStr(files, "demoenv/demomod/axmt503_wf.42f") {
		t.Errorf("取件失败的表单不该出现在缓存里: %v", files)
	}

	// 人读那一路也要看得见：再跑一次（这次 4 张命中、1 张仍失败）
	hout, herr := runForm(t, "fetch", "--env", "demoenv", "--module", "demomod", "--format", "table")
	if herr == nil {
		t.Fatalf("人读那一路有失败也必须非 0：\n%s", hout)
	}
	for _, want := range []string{"axmt503_wf", "no such file"} {
		if !strings.Contains(hout, want) {
			t.Errorf("人读输出里缺 %q：\n%s", want, hout)
		}
	}
}

// TestFormFetchListFailureIsNonZero 钉住"列不出表单"这条路：不取件、不写缓存、非 0。
func TestFormFetchListFailureIsNonZero(t *testing.T) {
	data := tempDataDir(t)
	src := fiveFormSource()
	src.err = errors.New("find: permission denied")
	injectSource(t, src)

	out, err := runForm(t, "fetch", "--env", "demoenv", "--module", "demomod", "--json")
	if err == nil {
		t.Fatalf("列不出表单必须非 0：\n%s", out)
	}
	if got := exitOf(t, err); got != exitFormUnknown {
		t.Errorf("退出码 = %d, want %d", got, exitFormUnknown)
	}
	if src.calls != 0 {
		t.Errorf("清单都没拿到就不该盲目取件（Form 被调用 %d 次）", src.calls)
	}
	if files := cachedFormFiles(t, formCacheDir(t, data)); len(files) != 0 {
		t.Errorf("清单失败时不该写缓存: %v", files)
	}
}

// TestFormFetchEmptyBodyIsAFailure 钉住"取回来是 0 字节"不算取到：
// 空内容不是一张表单（C1.7 的来源之一），要按失败报出来、写清病因、别写进缓存。
func TestFormFetchEmptyBodyIsAFailure(t *testing.T) {
	data := tempDataDir(t)
	src := fiveFormSource()
	src.bodies["demomod/axmt502_wf"] = []byte{} // 真机镜像里确实会撞上 0 字节的 .42f
	injectSource(t, src)

	out, err := runForm(t, "fetch", "--env", "demoenv", "--module", "demomod", "--json")
	if err == nil {
		t.Fatalf("有表单取回空内容必须非 0：\n%s", out)
	}
	if got := exitOf(t, err); got != exitFormUnknown {
		t.Errorf("退出码 = %d, want %d", got, exitFormUnknown)
	}
	doc := decodeDoc(t, out)
	if got := num(t, doc, "fetched"); got != 4 {
		t.Errorf("fetched = %d, want 4（空内容那张不算取到）", got)
	}
	fails, _ := doc["failures"].([]any)
	if len(fails) != 1 {
		t.Fatalf("failures 该恰好一条: %v", doc["failures"])
	}
	fl, _ := fails[0].(map[string]any)
	if fl["form"] != "axmt502_wf" {
		t.Errorf("failures[0].form = %v, want axmt502_wf", fl["form"])
	}
	if reason, _ := fl["reason"].(string); !strings.Contains(reason, "空") {
		t.Errorf("原因该说清「取回的内容是空的」: %q", reason)
	}
	files := cachedFormFiles(t, formCacheDir(t, data))
	if len(files) != 4 || containsStr(files, "demoenv/demomod/axmt502_wf.42f") {
		t.Errorf("空内容不该进缓存: %v", files)
	}
}

// ── 5. cache 概览 ───────────────────────────────────────────────────────────

// TestFormCacheInfoIsTruthful 钉住概览的三个数字都是**真值**（数量/字节/时间），
// 并且**没有删除任何文件**（跑前后文件集合逐项相同）。
func TestFormCacheInfoIsTruthful(t *testing.T) {
	data := tempDataDir(t)
	c, err := NewCacheFromDataDir(data)
	if err != nil {
		t.Fatalf("算不出缓存落点: %v", err)
	}

	// 三张表 + 各自的 mtime（相差整分钟，避开文件系统时间精度的边界）
	bodies := map[string][]byte{
		"e1/m1/f1": []byte("<Form name=\"f1\"/>"),
		"e1/m1/f2": []byte("<Form name=\"f2\" build=\"3.21.03\"/>"),
		"e2/m2/f3": []byte("<Form name=\"f3\"/>"),
	}
	base := time.Now().Add(-3 * time.Hour).Truncate(time.Minute)
	mtimes := map[string]time.Time{
		"e1/m1/f1": base.Add(0 * time.Minute),
		"e1/m1/f2": base.Add(1 * time.Minute),
		"e2/m2/f3": base.Add(2 * time.Minute),
	}
	var wantBytes int64
	for k, b := range bodies {
		parts := strings.Split(k, "/")
		if err := c.Put(parts[0], parts[1], parts[2], b, mtimes[k]); err != nil {
			t.Fatalf("写缓存 %s 失败: %v", k, err)
		}
		wantBytes += int64(len(b))
	}
	before := snapshotTree(t, c.Dir)

	out, err := runForm(t, "cache", "--json")
	if err != nil {
		t.Fatalf("cache 失败: %v", err)
	}
	doc := decodeDoc(t, out)
	if got := num(t, doc, "forms"); got != 3 {
		t.Errorf("forms = %d, want 3", got)
	}
	if n, ok := doc["bytes"].(float64); !ok || int64(n) != wantBytes {
		t.Errorf("bytes = %v, want %d（必须是真值）", doc["bytes"], wantBytes)
	}
	if doc["dir"] != c.Dir {
		t.Errorf("dir = %v, want %v", doc["dir"], c.Dir)
	}

	oldest, _ := doc["oldest"].(string)
	newest, _ := doc["newest"].(string)
	oldestT, perr := time.Parse(time.RFC3339, oldest)
	if perr != nil {
		t.Fatalf("oldest 不是 RFC3339: %q（%v）", oldest, perr)
	}
	newestT, perr := time.Parse(time.RFC3339, newest)
	if perr != nil {
		t.Fatalf("newest 不是 RFC3339: %q（%v）", newest, perr)
	}
	if !oldestT.Truncate(time.Second).Equal(mtimes["e1/m1/f1"].Truncate(time.Second)) {
		t.Errorf("oldest = %v, want %v（最旧那一份的 mtime）", oldestT, mtimes["e1/m1/f1"])
	}
	if !newestT.Truncate(time.Second).Equal(mtimes["e2/m2/f3"].Truncate(time.Second)) {
		t.Errorf("newest = %v, want %v（最新那一份的 mtime）", newestT, mtimes["e2/m2/f3"])
	}

	// 人读那一路同样打得出来
	hout, herr := runForm(t, "cache", "--format", "table")
	if herr != nil {
		t.Fatalf("cache --format table 失败: %v", herr)
	}
	for _, want := range []string{"f1", "f2", "f3"} {
		if !strings.Contains(hout, want) {
			t.Errorf("人读概览里缺 %q：\n%s", want, hout)
		}
	}

	// **不删任何东西**（也不改）：前后文件集合逐项相同
	after := snapshotTree(t, c.Dir)
	if strings.Join(before, "\n") != strings.Join(after, "\n") {
		t.Errorf("cache 概览动过缓存目录（它只该看，不该删/改）\n--- before ---\n%s\n--- after ---\n%s",
			strings.Join(before, "\n"), strings.Join(after, "\n"))
	}
}

// TestFormCacheInfoOnEmptyCacheIsZero 钉住"还没有缓存"不是错误：0 张、0 字节、退 0，
// 且**不建目录**（概览是只读的）。
func TestFormCacheInfoOnEmptyCacheIsZero(t *testing.T) {
	data := tempDataDir(t)
	dir := formCacheDir(t, data)

	out, err := runForm(t, "cache", "--json")
	if err != nil {
		t.Fatalf("没有缓存时 cache 不该失败: %v\n%s", err, out)
	}
	doc := decodeDoc(t, out)
	if got := num(t, doc, "forms"); got != 0 {
		t.Errorf("forms = %d, want 0", got)
	}
	if got, _ := doc["exists"].(bool); got {
		t.Errorf("exists = %v, want false", doc["exists"])
	}
	if _, serr := os.Stat(dir); serr == nil {
		t.Errorf("概览不该顺手把缓存目录建出来：%s", dir)
	}
}

// ── 6. 命令面与顺序 ─────────────────────────────────────────────────────────

// TestGroupRegistersFetchAndCache 钉住两个新动词确实挂在 `tt form` 下。
func TestGroupRegistersFetchAndCache(t *testing.T) {
	got := map[string]bool{}
	for _, c := range Group.Commands() {
		got[c.Name()] = true
	}
	for _, want := range []string{"ls", "show", "verify", "fetch", "cache"} {
		if !got[want] {
			t.Errorf("tt form 缺子命令 %q（现有：%v）", want, got)
		}
	}
}

// TestFormFetchRequiresEnvModule 钉住用法错：两个都必填，退 2。
func TestFormFetchRequiresEnvModule(t *testing.T) {
	tempDataDir(t)
	injectSource(t, fiveFormSource())
	for _, args := range [][]string{
		{"fetch", "--module", "demomod"},
		{"fetch", "--env", "demoenv"},
	} {
		_, err := runForm(t, args...)
		if err == nil {
			t.Fatalf("%v 缺必填参数必须报错", args)
		}
		if got := exitOf(t, err); got != exitUsage {
			t.Errorf("%v 的退出码 = %d, want %d（用法错）", args, got, exitUsage)
		}
	}
}

// ── 小工具 ─────────────────────────────────────────────────────────────────

func containsStr(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// listAllFiles 列出 dir 下所有文件的相对路径（排序）。
func listAllFiles(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	if _, err := os.Stat(dir); err != nil {
		return nil
	}
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, werr error) error {
		if werr != nil || d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(dir, p)
		if rerr != nil {
			return rerr
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("走目录 %s 失败: %v", dir, err)
	}
	sort.Strings(out)
	return out
}

// snapshotTree 是"文件集合 + 每份的大小与 mtime"的快照（比对前后是否被动过）。
func snapshotTree(t *testing.T, dir string) []string {
	t.Helper()
	files := listAllFiles(t, dir)
	out := make([]string, 0, len(files))
	for _, rel := range files {
		fi, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("stat %s 失败: %v", rel, err)
		}
		out = append(out, fmt.Sprintf("%s|%d|%d", rel, fi.Size(), fi.ModTime().UnixNano()))
	}
	return out
}
