// 本文件是「查询读到的表必须都在同步清单里」这条要求的闸门。
//
// 判据(dict 的**离线**可用性完全建立在它上面):查询层 SQL 里 FROM / JOIN 的每一张表,
// 都必须出现在 Families(展平后即 DictTables)里,或者在 nonDictTables 台账里写明为什么不进镜像。
// 清单里没有的表,镜像里就没有 —— 离线环境下那条命令必然失败。
//
// 抽取用**两层,缺一不可**:
//
//	第一层 go/ast + go/parser,只取字符串字面量(*ast.BasicLit / token.STRING)。
//	  注释不是 BasicLit,天然进不来 —— 本包自己的注释里就写着「JOIN dzeb_t 取字段名」,
//	  文本正则会把它当成一处查询。
//	第二层 在字面量内容里只认 FROM / JOIN 后面那个 *_t 标识符(见 reFromJoinTable)。
//	  仓库里有真实的反例:cli/dict/param.go 的 8 个 case "ooac_t"/"gzsa_t"/… 是**参数群的
//	  分组码**(单据别/系统级/企业级/据点级),不是被查询的表。判据若放宽成「扫所有 *_t
//	  字面量」,未改动的仓库第一天就会红 —— 然后这条测试就会被人删掉。
//
// **扫描范围 = internal/dict 与 internal/cli/dict 两棵子树下的全部非测试 .go**,不是手写
// 的文件清单:新加一个查询文件自动进范围。这两个根正好是 dict 这个功能的全部源码面
// (dict 的表只可能在这两处被读到)。明确**不在**范围里的是别的功能的 SQL:
// internal/debug/wslog.go 的 wsfa_t(工作日志表,与数据字典无关)、internal/cli/debug 帮助
// 文本里的示例 SQL —— 它们不该出现在这本账上。另有一处静态扫不到的:db.TableRowCounts
// 的表名来自参数,而调用方传进去的就是 DictTables 本身 —— 它不是新表的来源。
//
// 考虑过但放弃的方案:建一个只含 DictTables 的临时 SQLite、把 20 个 Source 方法都调一遍、
// 断言 IsMissingTable 为 false。放弃理由 —— ①RebuildTable(t, []string{"x"}, nil) 只建一列,
// 查询引用具体列名时报的是 no such column,而 IsMissingTable **不匹配**它,要让它绿就得手抄
// 第二份 schema;②只覆盖本地后端,「本地加了、远程忘了加」一条都看不见(那正是
// TestLocalAndRemoteReadTheSameTables 要管的);③要手工维护调用清单,拿同一类脆弱换另一类;
// ④静态扫描在表名写进代码的那一刻就红,运行时方案要等用户真的查表、而附表缺失多数根本不报错;
// ⑤假表建出来也是通的,它答不上「这些表该不该同步到本地」—— 而那才是要钉的属性。
package dbsync

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"tt/internal/testkit"
)

// reFromJoinTable 抽出 FROM / JOIN 后面的表名。只认 *_t 形状(T100 字典表的统一后缀),
// 所以 `FROM %s`、`FROM (SELECT …)`、英文散文里的 "from a table" 都不会误报。
//
// 允许 schema 限定名与双引号(`FROM "user"."dzeb_t"` 这种形态):今天 dict 的 SQL 里没有
// 一处这么写,但将来写了而这里不认,那张表就会**静默漏过**扫描 —— 正是本文件要防的事。
var reFromJoinTable = regexp.MustCompile(`(?i)\b(?:from|join)\s+(?:"?[a-z_][a-z0-9_]*"?\.)?"?([a-z][a-z0-9_]*_t)"?`)

// scanRoots 参与扫描的两棵子树(仓库相对、正斜杠)。
var scanRoots = []string{"internal/dict", "internal/cli/dict"}

// scanRootMinFiles 每个根**至少**要读到这么多非测试 .go,防空转:路径打错一个字符,
// 那个根会静默贡献 0 张表,而主闸门照样绿。实测 internal/dict 17、internal/cli/dict 21。
var scanRootMinFiles = map[string]int{"internal/dict": 15, "internal/cli/dict": 18}

// 整体下界(括号里是实测值),只抓「扫描整个坏了」这类失效。
const (
	minFilesScanned    = 30 // 实测 38
	minFilesWithTables = 8  // 实测 10
	minTablesFound     = 25 // 实测 33
	minOneBackendTabs  = 25 // 单看 db/ 或 live/,实测各 32
	minLedgerReason    = 20 // 台账理由的字数下限
)

// canaryTables 每张代表一种 SQL 形态,抓「某一类形态失效」。
// 失败时的文案必须写明「不是清单少了它,是抽取器漏了这一类」—— 否则下一个人会去改 Families。
var canaryTables = []struct{ table, why string }{
	{"dzcd_t", "最常见的形态:`FROM dzcd_t c` + 别名(实测 FROM 4)"},
	{"dzcdl_t", "**只以 JOIN 出现**(实测 FROM 0 / JOIN 5)—— 判据里的 join 分支断了就会红"},
	{"gzzal_t", "msg 与 prog 两族共用、且是 JOIN 的目标(实测 FROM 2 / JOIN 10)"},
	{"gzdg_t", "只有 progtable 一族用、只出现在 4 个查询里 —— 最容易被整段漏掉"},
	{"dzep_t", "spec 族只有这一张,只有 desc 一条命令用它(实测 FROM 2)"},
	{"gzzk_t", "双键 ON 的 JOIN 目标(实测 FROM 0 / JOIN 2)"},
}

// nonDictTables —— 「查询层会读、但**故意**不进本地镜像」的表。今天只有一条。
//
// 三条机制让它不是随便开的口子:
//  1. 台账里的表必须**仍被读到**(TestTableLedgersAreAlive 查)—— 哪天那段查询删了,报「例外过期」;
//  2. 不许同时出现在 DictTables 里 —— 用「加进清单」把主闸门糊过去,对 gzou_t 来说正好是最坏结果
//     (把客户的企业→账号映射落到磁盘上);
//  3. 理由要写够字数 —— 没有理由的放过会变成下一次的放过。
var nonDictTables = map[string]string{
	"gzou_t": "企业编号→数据库账号的引导表:要先用它查出账号才连得上 ERP 库,所以读它发生在建主连接之前(仅 live 路径)。离线不连库、不需要账号解析,拉进镜像没有用处。它同时是客户的企业→账号映射,属凭据面 —— 落到本地 SQLite 等于把账号表写到磁盘上。见 cli/dict/entacct.go。",
}

// pendingQueries —— 已经进了 Families、但查询层还没接上的表。
//
// 「先加族、后接查询」是**合法顺序**(镜像可以先拉下来),所以不判红;但也不许无声增长:
// 加一张就要在这里留一行 + 一句为什么,那一行会出现在 diff 里被人看见。
// **不设条数上限** —— 上限是随手就能改大的数字,改大它比写理由省事。
var pendingQueries = map[string]string{
	"dzee_t": "在表字典族里,但全仓没有任何 FROM/JOIN 读它。入库自一次 squash 提交,没留下理由;保留是为了不改变现成的同步行为,待确认是删是接。",
	"dzef_t": "同 dzee_t:在表字典族里无人读,入库无理由记录。保留现状、登记在此,别无声地留在清单里。",
	"dzeg_t": "同 dzee_t:在表字典族里无人读,入库无理由记录。保留现状、登记在此,别无声地留在清单里。",
}

// tableScan 一次扫描的结果。
type tableScan struct {
	tables   map[string][]string // 表名(小写) → 读到它的文件(仓库相对路径、正斜杠、已排序)
	files    int                 // 读过的非测试 .go 数
	filesHit int                 // 其中至少抽到一张表的文件数
	perRoot  map[string]int      // 每个根读过的文件数
}

// names 扫描到的表名,按字典序(失败信息要稳定可 diff)。
func (s tableScan) names() []string {
	out := make([]string, 0, len(s.tables))
	for t := range s.tables {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// scanQueriedTables 读 roots 下的全部非测试 .go,返回它们 SQL 里 FROM / JOIN 的表。
//
// 解析失败的文件**跳过并记日志**,不让一个语法变体把整条闸门判红;但 files 计数照记,
// 所以整体下界仍然盖得住「大面积解析失败」。
func scanQueriedTables(t *testing.T, roots []string) tableScan {
	t.Helper()
	repo := testkit.RepoRoot(t)
	out := tableScan{tables: map[string][]string{}, perRoot: map[string]int{}}
	fset := token.NewFileSet()
	for _, root := range roots {
		dir := filepath.Join(repo, filepath.FromSlash(root))
		seen := map[string]bool{}
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
			if werr != nil {
				return werr
			}
			name := d.Name()
			if d.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				return nil
			}
			src, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			out.files++
			out.perRoot[root]++
			f, perr := parser.ParseFile(fset, path, src, parser.SkipObjectResolution)
			if perr != nil {
				t.Logf("跳过无法解析的文件 %s: %v", path, perr)
				return nil
			}
			rel, relErr := filepath.Rel(repo, path)
			if relErr != nil {
				rel = path
			}
			rel = filepath.ToSlash(rel)
			hit := false
			ast.Inspect(f, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				s, uerr := strconv.Unquote(lit.Value)
				if uerr != nil {
					return true
				}
				for _, tb := range tablesInSQL(s) {
					hit = true
					if seen[tb] {
						continue
					}
					seen[tb] = true
					out.tables[tb] = append(out.tables[tb], rel)
				}
				return true
			})
			if hit {
				out.filesHit++
			}
			return nil
		})
		if err != nil {
			t.Fatalf("扫描 %s 失败: %v", dir, err)
		}
	}
	for tb := range out.tables {
		sort.Strings(out.tables[tb])
	}
	return out
}

// tablesInSQL 返回一段 SQL 文本里 FROM / JOIN 的表名(小写、已去重)。
// 统一小写:SQL 里大小写混着写是常事,而清单里是全小写。
func tablesInSQL(s string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range reFromJoinTable.FindAllStringSubmatch(s, -1) {
		tb := strings.ToLower(m[1])
		if seen[tb] {
			continue
		}
		seen[tb] = true
		out = append(out, tb)
	}
	return out
}

// syncedSet 把 DictTables 转成集合,顺手断言它确实是 Families 的展平(别处也有这条,
// 但本文件的所有断言都建立在它上面,不值得靠另一个包的测试来保证)。
func syncedSet() map[string]bool {
	out := make(map[string]bool, len(DictTables))
	for _, t := range DictTables {
		out[t] = true
	}
	return out
}

// TestQueriedTablesAreAllInSyncList 是**主闸门**:查询读到的表 ⊆ 同步清单 ∪ 例外台账。
func TestQueriedTablesAreAllInSyncList(t *testing.T) {
	sc := scanQueriedTables(t, scanRoots)
	synced := syncedSet()

	var bad []string
	for _, tb := range sc.names() {
		if synced[tb] || nonDictTables[tb] != "" {
			continue
		}
		bad = append(bad, fmt.Sprintf("  %s ← %s", tb, strings.Join(sc.tables[tb], "、")))
	}
	if len(bad) > 0 {
		t.Errorf("有 %d 张表会被 dict 的查询读到,却不在同步清单里 —— 离线(本地 SQLite 镜像)下这几条查询必然失败:\n%s\n"+
			"  改法(二选一):\n"+
			"   1. 把它加进 dbsync.go 的 Families 里**它服务的那一族**的 Tables(同一笔改动里),"+
			"`tt dict db sync` 才拉得下来;\n"+
			"   2. 它确实不该进镜像(比如是账号引导表、属凭据面)—— 在 nonDictTables 里登记一行并写明理由。\n"+
			"  只加表不挂族 = 加了张 db sync 永远拉不下来的表。", len(bad), strings.Join(bad, "\n"))
	}
}

// TestTableScanIsNotVacuous 盯住**扫描本身**还在工作。
// 主闸门是「集合包含」,扫描坏掉时它是**假绿**:读到 0 张表也满足包含关系。
func TestTableScanIsNotVacuous(t *testing.T) {
	sc := scanQueriedTables(t, scanRoots)

	if sc.files < minFilesScanned {
		t.Fatalf("只读到 %d 个非测试 .go(实测 %d 个)—— 扫描路径可能写错了", sc.files, minFilesScanned)
	}
	if sc.filesHit < minFilesWithTables {
		t.Fatalf("只有 %d 个文件抽出过表(实测 %d 个)—— 抽取器大面积失效", sc.filesHit, minFilesWithTables)
	}
	if len(sc.tables) < minTablesFound {
		t.Fatalf("只扫到 %d 张表(实测 %d 张)—— 抽取器或扫描范围出了问题", len(sc.tables), minTablesFound)
	}
	for _, root := range scanRoots {
		if got, want := sc.perRoot[root], scanRootMinFiles[root]; got < want {
			t.Fatalf("根 %s 只读到 %d 个非测试 .go(至少应 %d 个)—— 路径打错一个字符,"+
				"这个根就会静默贡献 0 张表,而主闸门照样绿", root, got, want)
		}
	}
	for _, c := range canaryTables {
		if len(sc.tables[c.table]) == 0 {
			t.Errorf("canary 表 %s 没被扫到 —— %s。\n"+
				"  注意:这**不是**同步清单少了它,是抽取器漏了这一类 SQL 形态。"+
				"别去改 Families,去改 reFromJoinTable 或扫描范围。", c.table, c.why)
		}
	}
}

// TestTableQueryExtractor 是抽取器**自己**的形状判据。
// 主闸门只说「查到的都在清单里」,抽取器漏一类形态时它会假绿 —— 这一条把那些形态逐个钉住,
// 每个「必须找不到」的样本都对应仓库里真实存在的一处(英文散文、参数群分组码……)。
func TestTableQueryExtractor(t *testing.T) {
	mustFind := []struct{ name, sql string }{
		{"FROM + 别名", "SELECT a FROM dzeb_t b WHERE b.x = ?"},
		{"只以 JOIN 出现", "SELECT 1 FROM dzca_t c LEFT JOIN dzcdl_t l ON l.k = c.k"},
		{"schema 限定", "select count(*) from user.dzea_t"},
		{"双引号包住", `SELECT x FROM "gzcb_t"`},
		{"跨行", "SELECT 1\n  FROM gzzk_t k\n  JOIN gzzal_t g ON g.a = k.a"},
		{"大小写混写", "select x FrOm GzZe_T"},
	}
	for _, c := range mustFind {
		if got := tablesInSQL(c.sql); len(got) == 0 {
			t.Errorf("%s:这段 SQL 里的表没被抽出来:\n%s", c.name, c.sql)
		}
	}

	mustNotFind := []struct{ name, sql string }{
		{"英文散文 from a table", "TableInfo holds a single field's information from a table dictionary query."},
		{"英文散文 from the", "TableListItem holds a row from the table list query."},
		{"英文散文 from scratch", "This file is built from scratch; a crash may corrupt it."},
		{"参数群的分组码(不是被查的表)", `case "ooac_t":`},
		{"同上,另一个分组码", `case "gzsa_t":`},
		{"动态表名", "SELECT COUNT(*) FROM %s"},
		{"子查询不是表名", "SELECT * FROM (SELECT 1) x"},
		{"缺 _t 后缀", "SELECT 1 FROM gzca c JOIN gzcal l ON 1=1"},
	}
	for _, c := range mustNotFind {
		if got := tablesInSQL(c.sql); len(got) > 0 {
			t.Errorf("%s:这段文本不该被当成查询(抽出了 %v):\n%s"+
				"\n  放宽判据的后果:未改动的仓库第一天就红,然后这条测试会被人整段删掉。", c.name, got, c.sql)
		}
	}
}

// TestTableScanIgnoresComments 是**第一层**的判据:注释里的表名不算查询。
// 仓库里真实存在 6 处这种注释(本包 dbsync.go 自己就有两处「JOIN dzeb_t」)。
func TestTableScanIgnoresComments(t *testing.T) {
	const src = "package p\n" +
		"\n" +
		"// 参考 live/source.go 的 FROM gzzx_t 写法\n" +
		"/* 块注释 from gzzy_t 同样不算 */\n" +
		"func f() string { return `SELECT a FROM dzeb_t b JOIN dzebl_t l ON 1=1` }\n"
	f, err := parser.ParseFile(token.NewFileSet(), "synthetic.go", src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("合成源码解析失败(测试自己写错了): %v", err)
	}
	got := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		s, uerr := strconv.Unquote(lit.Value)
		if uerr != nil {
			return true
		}
		for _, tb := range tablesInSQL(s) {
			got[tb] = true
		}
		return true
	})
	for _, want := range []string{"dzeb_t", "dzebl_t"} {
		if !got[want] {
			t.Errorf("字面量里的 %s 没被抽出来(注释之外的正经查询必须能抽到)", want)
		}
	}
	for _, notWant := range []string{"gzzx_t", "gzzy_t"} {
		if got[notWant] {
			t.Errorf("注释里的 %s 被当成了查询 —— 注释不是 *ast.BasicLit,抽取不该看得见它", notWant)
		}
	}
}

// TestTableLedgersAreAlive 给两本台账装上反面判据,免得它们烂成"谁也不看的名单"。
func TestTableLedgersAreAlive(t *testing.T) {
	sc := scanQueriedTables(t, scanRoots)
	synced := syncedSet()

	for _, tb := range sortedKeys(nonDictTables) {
		why := nonDictTables[tb]
		if len([]rune(why)) < minLedgerReason {
			t.Errorf("nonDictTables[%q] 的理由只有 %d 字(至少 %d 字)—— "+
				"没有理由的放过会变成下一次的放过", tb, len([]rune(why)), minLedgerReason)
		}
		if synced[tb] {
			t.Errorf("%s 同时出现在 nonDictTables 与 DictTables 里 —— "+
				"这本台账是给「不进镜像」的表用的;一旦进了清单就必须从台账里删掉,否则就是"+
				"拿「加进清单」把主闸门糊过去(对它来说正好是最坏结果)", tb)
		}
		if len(sc.tables[tb]) == 0 {
			t.Errorf("**例外过期了**:台账登记 %s 说它是「查询层会读、但故意不进镜像」,"+
				"可现在没有任何 FROM/JOIN 读它。要么把查询找回来,要么删掉这一行。", tb)
		}
	}

	for _, tb := range sortedKeys(pendingQueries) {
		why := pendingQueries[tb]
		if len([]rune(why)) < minLedgerReason {
			t.Errorf("pendingQueries[%q] 的理由只有 %d 字(至少 %d 字)", tb, len([]rune(why)), minLedgerReason)
		}
		if !synced[tb] {
			t.Errorf("pendingQueries[%q] 不在 DictTables 里 —— "+
				"这本台账记的是「已进清单、查询还没接上」的表;不在清单里的表属于另一本台账", tb)
		}
		if files := sc.tables[tb]; len(files) > 0 {
			t.Errorf("pendingQueries[%q] 的查询已经接上了(%s 读了它)—— "+
				"把它从这本台账里删掉,它现在是正常成员", tb, strings.Join(files, "、"))
		}
	}

	// 反面:进了清单却没人读的表,必须逐条登记 —— 否则「先加族后接查询」会变成无声的常驻。
	var unregistered []string
	for _, tb := range DictTables {
		if len(sc.tables[tb]) == 0 && pendingQueries[tb] == "" {
			unregistered = append(unregistered, tb)
		}
	}
	if len(unregistered) > 0 {
		t.Errorf("这些表在同步清单里,但没有任何查询读它们,也没在 pendingQueries 登记:%v\n"+
			"  后果:镜像永远多拉这几张,db status 的族完整性也说不清为什么要拉它们。\n"+
			"  改法:要么把查询接上(那这一行自己会消失),要么在 pendingQueries 里加一行写明为什么先加族。",
			unregistered)
	}
}

// TestLocalAndRemoteReadTheSameTables 把 db/README.md 那句「表名或 JOIN:本地镜像与远程直查
// 必须同步改」变成**集合相等**断言。
//
// 为什么值得一条单独测试:主闸门看的是「两侧的并集 ⊆ 清单」,一侧读了一张清单里的表、另一侧
// 没读,它是绿的 —— 而那条命令切到另一个数据源就会失败(或少了 JOIN 出来的那一列)。
func TestLocalAndRemoteReadTheSameTables(t *testing.T) {
	local := scanQueriedTables(t, []string{"internal/dict/db"})
	remote := scanQueriedTables(t, []string{"internal/dict/live"})

	if len(local.tables) < minOneBackendTabs || len(remote.tables) < minOneBackendTabs {
		t.Fatalf("本地侧扫到 %d 张、远程侧扫到 %d 张(两侧实测各 %d 张)—— "+
			"任一侧扫描失效都会让下面的相等性变成假绿", len(local.tables), len(remote.tables), minOneBackendTabs)
	}
	only := func(a tableScan, b tableScan) []string {
		var out []string
		for _, tb := range a.names() {
			if len(b.tables[tb]) == 0 {
				out = append(out, tb)
			}
		}
		return out
	}
	if onlyLocal, onlyRemote := only(local, remote), only(remote, local); len(onlyLocal) > 0 || len(onlyRemote) > 0 {
		t.Errorf("本地镜像与远程直查读的表不一样 —— 两侧必须同步改:\n"+
			"  只有本地镜像在读:%v\n  只有远程直查在读:%v\n"+
			"  后果:把数据源切到另一边时,这条查询要么报缺表,要么少一列 —— 而 --help 的提示还说「齐了」。",
			onlyLocal, onlyRemote)
	}
}

// sortedKeys 让 map 的遍历顺序确定,失败信息才可 diff。
func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
