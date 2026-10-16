package debug

// 这组测试盯的是"点名了环境就必须按它连"这条线。
//
// 从前 `tt debug start … --env X` 既不报错也不生效(--env 是根命令的开关,debug 组里
// 没人读它),方向永远取服务端当前环境;同机多环境(同库不同企业)时那就是"上错号",
// 所有按企业编号取数的查询返回 0 行,看起来与"配置缺失"一模一样。这些用例把三件事钉住:
// 点名怎么解析、生效身份怎么表述、不一致时怎么报。

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tt/internal/cli/common"
	"tt/internal/entdir"
	"tt/internal/target"
)

// 环境名解析:子命令上的 --env 优先,其次根命令的持久 --env,再没有就是空(用当前环境)。
func TestDBGTargetEnv(t *testing.T) {
	oldFlag, oldGlobal := dbgEnvFlag, common.Env
	t.Cleanup(func() { dbgEnvFlag, common.Env = oldFlag, oldGlobal })

	dbgEnvFlag, common.Env = "", ""
	if got := dbgTargetEnv(); got != "" {
		t.Fatalf("两条都没给时应为空(用服务端当前环境),got %q", got)
	}
	dbgEnvFlag, common.Env = "", "根命令给的"
	if got := dbgTargetEnv(); got != "根命令给的" {
		t.Fatalf("根命令 --env 应被认(别处能用,这里也得能用),got %q", got)
	}
	dbgEnvFlag, common.Env = "子命令给的", "根命令给的"
	if got := dbgTargetEnv(); got != "子命令给的" {
		t.Fatalf("子命令 --env 应优先,got %q", got)
	}
	dbgEnvFlag, common.Env = "  带空白  ", ""
	if got := dbgTargetEnv(); got != "带空白" {
		t.Fatalf("两侧空白应剔除,got %q", got)
	}
}

// 点名了却没生效:必须报警并判失败;一致或没点名则放行。
func TestDBGVerifyEnv(t *testing.T) {
	oldFlag, oldGlobal := dbgEnvFlag, common.Env
	t.Cleanup(func() { dbgEnvFlag, common.Env = oldFlag, oldGlobal })
	dbgEnvFlag, common.Env = "", ""

	snap := &dbgSnapView{Session: target.Session{Env: "测试1012", Prog: "capp115"}}
	if !dbgVerifyEnv("", snap) {
		t.Fatal("没点名环境时不该报失败")
	}
	if !dbgVerifyEnv("测试1012", snap) {
		t.Fatal("点名与实际一致时不该报失败")
	}
	if dbgVerifyEnv("恒烁正式区", snap) {
		t.Fatal("点名与实际不一致时必须判失败(否则调用方以为跑在了点名的环境上)")
	}
	if !dbgVerifyEnv("恒烁正式区", nil) {
		t.Fatal("拿不到快照时不误报(另一条路会给失败)")
	}
}

// 复用了宿主要有痕迹:只有"回包 id 与启动前那个空闲会话相同"才算复用。
func TestDBGReusedHost(t *testing.T) {
	cases := []struct {
		prev, id string
		want     bool
	}{
		{"s1791770107217", "s1791770107217", true},  // 同一个空闲会话被复用
		{"s1791770107217", "s1791957373680", false}, // 新建了一个
		{"", "s1791957373680", false},               // 启动前没有空闲会话
		{"s1791770107217", "", false},               // 没拿到 id:不猜
	}
	for _, c := range cases {
		if got := dbgReusedHost(c.prev, c.id); got != c.want {
			t.Errorf("dbgReusedHost(%q, %q) = %v, 期望 %v", c.prev, c.id, got, c.want)
		}
	}
}

// 生效身份一行里必须同时有 环境/区域/企业/账号/库,以及"复用了宿主"的痕迹。
func TestDBGIdentity(t *testing.T) {
	oldPath := common.ConfigPath
	common.ConfigPath = writeIdentityConfig(t)
	t.Cleanup(func() { common.ConfigPath = oldPath })

	snap := &dbgSnapView{Session: target.Session{Env: "恒烁正式区"}, Zone: "1", TopentCf: "9999", Reused: true}
	got := dbgIdentity(snap)
	for _, want := range []string{"恒烁正式区", "区域 1", "企业 9999", "172.19.93.55:54321/topprd", "复用了已登录的宿主"} {
		if !strings.Contains(got, want) {
			t.Errorf("生效身份行缺少 %q:\n%s", want, got)
		}
	}
	// 没有企业快照时不许拨 SSH(慢时 40 秒),但要说明查不到并给出自己核对的命令
	if !strings.Contains(got, "没有该环境的企业快照") || !strings.Contains(got, "tt debug ents --env 恒烁正式区") {
		t.Errorf("没有快照时应说明为什么没有账号、并给出核对命令:\n%s", got)
	}
	// 三个裸字段全空时也不能打出空白:要么给值,要么写明没取到
	if g := dbgIdentity(&dbgSnapView{Session: target.Session{Env: "恒烁正式区"}}); !strings.Contains(g, "企业 (未取到)") {
		t.Errorf("企业三个字段都空时应写明未取到:\n%s", g)
	}
	if dbgIdentity(nil) != "" {
		t.Error("没有快照时应返回空串(不能打出半行)")
	}
}

// 有企业快照时,身份行要直接给出**账号** —— 企业编号错了就会连到另一个 schema,
// 而"上错号"与"数据不存在"在输出上一模一样,这一行就是用来分开它们的。
func TestDBGIdentityWithEntSnapshot(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeIdentityConfigIn(t, dir)

	// 快照文件名与内容都照 internal/entdir 的契约来(与 tt debug ents 落的是同一份)
	const (
		sshHost, sshUser, zone = "172.19.93.55", "tiptop", "1"
		dialect, dbHost, svc   = "kingbase", "172.19.93.55", "topprd"
	)
	fp := entdir.Fingerprint(sshHost, sshUser, zone, entdir.DBIdent(dialect, dbHost, 54321, svc))
	path := entdir.Path(dir, entdir.EnvSeg("恒烁正式区", sshHost, zone))
	if err := entdir.Write(path, &entdir.Snapshot{
		Version: entdir.Version, Env: "恒烁正式区", Fingerprint: fp, Zone: zone,
		Dialect: dialect, Target: svc, FetchedAt: time.Now(),
		Mappings: []entdir.Mapping{{Ent: 9999, Account: "dsdata"}, {Ent: 1012, Account: "dsdemo"}},
	}); err != nil {
		t.Fatalf("写快照失败: %v", err)
	}

	oldPath := common.ConfigPath
	common.ConfigPath = cfgPath
	t.Cleanup(func() { common.ConfigPath = oldPath })

	got := dbgIdentity(&dbgSnapView{Session: target.Session{Env: "恒烁正式区"}, Zone: "1", TopentCf: "9999"})
	if !strings.Contains(got, "账号 dsdata") {
		t.Errorf("有快照时应给出企业 9999 的账号 dsdata:\n%s", got)
	}
	if strings.Contains(got, "dsdemo") {
		t.Errorf("不该串到别的企业的账号上:\n%s", got)
	}
}

// writeIdentityConfig 落一份"单环境挂库"的配置,返回其路径。
func writeIdentityConfig(t *testing.T) string {
	t.Helper()
	return writeIdentityConfigIn(t, t.TempDir())
}

// writeIdentityConfigIn 同上,但落在指定目录(企业快照的 DataDir 就是配置所在目录)。
func writeIdentityConfigIn(t *testing.T, dir string) string {
	t.Helper()
	cfgPath := filepath.Join(dir, "config.json")
	cfg := `{"schemaVersion":3,"hosts":{"activeEnv":"恒烁正式区","sshs":[
	  {"name":"恒烁正式区","host":"172.19.93.55","port":22,"user":"tiptop","zone":"1","topent":"9999",
	   "db":{"type":"kingbase","host":"172.19.93.55","port":54321,"service":"topprd"}}
	]}}`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	return cfgPath
}
