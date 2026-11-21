package devdata

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestMirrorTablesMatchPlanAppendix 钉住"代码里的 53 张表"与计划文件附录 A 一致。
//
// 为什么要有这条测试：清单是**双份存在**的（代码 + 文档）。漂移的后果不是编译错误，
// 而是"文档说镜了 53 张、实际镜了 52 张"，而这种缺失要等到某次 model 结论错得离谱
// 才会被发现。
//
// 计划文件在全部完成后会被改名成 `✓tt-devdata-plan.md`（项目规范），所以按通配找，
// 不写死文件名。找不到就是**失败**而不是跳过 —— 计划文件是入库文件，它不在仓库里
// 说明测试环境不对，而"跳过"会让这条一致性断言静默空转。
func TestMirrorTablesMatchPlanAppendix(t *testing.T) {
	matches, err := filepath.Glob("../../.agents/plan/*tt-devdata-plan.md")
	if err != nil || len(matches) == 0 {
		t.Fatalf("找不到计划文件（.agents/plan/*tt-devdata-plan.md），无法做表清单一致性断言: %v", err)
	}
	b, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("读计划文件 %s 失败: %v", matches[0], err)
	}
	text := string(b)
	// 附录 A 是唯一一处连续列出全部表名的地方（DZAA DZAB … DZZC）
	idx := strings.Index(text, "## 附录 A")
	if idx < 0 {
		t.Fatalf("计划文件里找不到「## 附录 A」")
	}
	tail := text[idx:]
	codeBlock := regexp.MustCompile("(?s)```\n(.*?)```")
	mm := codeBlock.FindStringSubmatch(tail)
	if mm == nil {
		t.Fatalf("附录 A 里没有代码块")
	}
	fields := strings.Fields(mm[1])
	var want []string
	for _, f := range fields {
		if regexp.MustCompile(`^DZ[A-Z]{2}$`).MatchString(f) {
			want = append(want, f+"_T")
		}
	}
	if len(want) != 53 {
		t.Fatalf("附录 A 解析出 %d 张表，应为 53: %v", len(want), want)
	}
	sort.Strings(want)
	got := MirrorTableNames()
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("代码与文档的表清单不一致\n代码 %d 张: %v\n文档 %d 张: %v",
			len(got), got, len(want), want)
	}
}

// TestRolesComplete 每个角色都要有源路径、落点与必需性判定（结构完整）。
func TestRolesComplete(t *testing.T) {
	if len(FetchRoles) == 0 {
		t.Fatal("FetchRoles 为空")
	}
	required := map[string]bool{}
	for _, r := range FetchRoles {
		if r.Role == "" || r.Src == "" || r.Rel == "" {
			t.Errorf("角色定义不完整: %+v", r)
		}
		if !strings.Contains(r.Src, "%s") || !strings.Contains(r.Rel, "%s") {
			t.Errorf("角色的路径模板必须带 %%s: %+v", r)
		}
		if r.Required {
			required[r.Role] = true
		}
	}
	// 必需项 = 蓝图 / 骨架 / 产物；缺任何一个整套都不成立
	for _, role := range []string{"tab", "tgl", "4gl"} {
		if !required[role] {
			t.Errorf("%s 必须是必需项（缺它 model 无从判断）", role)
		}
	}
	// 同一个角色名可以有多条（inc 的 mask/rep），但落点不能撞
	rels := map[string]string{}
	for _, r := range FetchRoles {
		rel := r.RelPath("axmt500_wf")
		if prev, ok := rels[rel]; ok {
			t.Errorf("落点重复: %s 同时属于 %s 与 %s", rel, prev, r.Role)
		}
		rels[rel] = r.Role
	}
}

// TestTableSpecs 53 张表的名字、字段数、行数不能是零值（防止手滑漏填）。
func TestTableSpecs(t *testing.T) {
	if len(MirrorTables) != 53 {
		t.Fatalf("MirrorTables 有 %d 张，应为 53", len(MirrorTables))
	}
	for _, s := range MirrorTables {
		if !strings.HasSuffix(s.Name, "_T") {
			t.Errorf("表名应以 _T 结尾: %s", s.Name)
		}
		if s.Desc == "" {
			t.Errorf("%s 缺中文说明", s.Name)
		}
		if s.Fields <= 0 {
			t.Errorf("%s 字段数应为正: %d", s.Name, s.Fields)
		}
	}
	// 唯一 LatestOnly 的表是 DZAA_T（ADR-03 的补充口径）
	var latest []string
	for _, s := range MirrorTables {
		if s.LatestOnly {
			latest = append(latest, s.Name)
		}
	}
	if strings.Join(latest, ",") != "DZAA_T" {
		t.Fatalf("LatestOnly 只应给 DZAA_T，实际: %v", latest)
	}
	// 实测总行数要能对上计划里那个 1837 万（±1，防手滑改一个数字）
	var total int64
	for _, s := range MirrorTables {
		total += s.Rows
	}
	if total != 18371187 {
		t.Fatalf("实测行数合计 = %d，应为 18371187（2026-11-21 福联测试tst 实测）；改了数据请同步文档", total)
	}
}

// TestGuessModule 模块推断只看前三位。
func TestGuessModule(t *testing.T) {
	cases := map[string]string{
		"axmt500_wf": "axm",
		"apmt500":    "apm",
		"aap":        "aap",
		"ab":         "",
	}
	for prog, want := range cases {
		if got := GuessModule(prog); got != want {
			t.Errorf("GuessModule(%q) = %q, want %q", prog, got, want)
		}
	}
}

// TestTemplateKey 模板键的折算规则。
func TestTemplateKey(t *testing.T) {
	cases := []struct {
		rel string
		key string
		ok  bool
	}{
		{"code_a05.template", "a05", true},
		{"slice/code_qs01.template", "qs01", true},
		{"code_i02.template", "i02", true},
		{"4rp_01.template", "", false},
		{"F001_00.snippet", "", false},
		// 带日期的备份不是模板：它是旧版本的副本，记进版本表会多出几个不存在的模板
		{"code_g01.template.141119", "", false},
		{"code_s01.140128", "", false},
	}
	for _, c := range cases {
		key, ok := TemplateKey(c.rel)
		if ok != c.ok {
			t.Errorf("TemplateKey(%q) ok = %v, want %v", c.rel, ok, c.ok)
			continue
		}
		if ok && key != c.key {
			t.Errorf("TemplateKey(%q) = %q, want %q", c.rel, key, c.key)
		}
	}
}

// TestIsTemplateFile 状态里的「模板 N 个」口径：code_*（含备份）+ slice/ = 171。
func TestIsTemplateFile(t *testing.T) {
	yes := []string{"code_a05.template", "slice/code_a05.template", "code_g01.template.141119", "slice/code_d01.template"}
	no := []string{"4rp_01.template", "F001_00.snippet", "ComboxItemSetting.xml", "xrp/x.template"}
	for _, rel := range yes {
		if !IsTemplateFile(rel) {
			t.Errorf("IsTemplateFile(%q) 应为真", rel)
		}
	}
	for _, rel := range no {
		if IsTemplateFile(rel) {
			t.Errorf("IsTemplateFile(%q) 应为假", rel)
		}
	}
}

// TestPathValidation 程序名白名单：`../`、大写、超长、过短一律拒绝。
func TestPathValidation(t *testing.T) {
	bad := []string{"", "../../x", "../etc/passwd", "Axmt500", "a", "ab", "axmt500-wf", "axmt500_wf/../x",
		strings.Repeat("a", 40), "1axmt500"}
	for _, p := range bad {
		if err := ValidProgName(p); err == nil {
			t.Errorf("ValidProgName(%q) 应被拒绝", p)
		}
	}
	good := []string{"aap", "axmt500", "axmt500_wf", "cxmt500_wf_10"}
	for _, p := range good {
		if err := ValidProgName(p); err != nil {
			t.Errorf("ValidProgName(%q) 应通过: %v", p, err)
		}
	}
}

// TestEnvMustBeInConfig 环境名必须在配置清单里（否则 404），且**校验先于建目录**。
func TestEnvMustBeInConfig(t *testing.T) {
	cfg := writeConfig(t, `{"name":"福联测试tst","host":"10.0.0.1","zone":"2"}`)
	if _, err := EnvDir(cfg, "没有这个环境"); err == nil {
		t.Fatal("未配置的环境应报错")
	} else if !osIs(err, ErrEnvNotFound) {
		t.Fatalf("错误应包 ErrEnvNotFound: %v", err)
	}
	dir, err := EnvDir(cfg, "福联测试tst")
	if err != nil {
		t.Fatalf("已配置的环境应能解析: %v", err)
	}
	// 校验失败时不应留下任何目录
	if _, err := ProgDir(cfg, "没有这个环境", "axmt500_wf"); err == nil {
		t.Fatal("未配置环境 + 合法程序名应报错")
	}
	if _, statErr := os.Stat(filepath.Join(filepath.Dir(cfg), "devdata")); statErr == nil {
		t.Fatal("校验失败时不应创建 devdata 目录（先校验、后落盘）")
	}
	if !strings.Contains(dir, "devdata") {
		t.Fatalf("环境目录应在 devdata 下: %s", dir)
	}
}

// osIs 是 errors.Is 的本地小包装（本文件不 import errors，避免与其他测试重复导入冲突）。
func osIs(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
