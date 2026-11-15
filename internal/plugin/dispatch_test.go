package plugin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/internal/hostrpc"
	"github.com/luyaogod/TT/internal/testkit"
)

// ---- 参考插件夹具 ----

// refExeName 当前平台的可执行文件名。
func refExeName() string {
	if runtime.GOOS == "windows" {
		return "refplugin.exe"
	}
	return "refplugin"
}

// buildReferencePlugin 把 testdata/plugins/reference 拷到临时目录，并把
// tools/refplugin 编译成它的入口。
//
// 为什么要现编：夹具是要**真进程**的（参数转发、退出码、就绪行、健康检查），
// 而可执行文件不能入库。manifest 里的入口名按平台改写一次。
func buildReferencePlugin(t *testing.T) string {
	t.Helper()
	root := testkit.RepoRoot(t)
	pkg := filepath.Join(t.TempDir(), "refplugin")
	if err := copyTree(filepath.Join(root, "testdata", "plugins", "reference"), pkg, DefaultZipLimits); err != nil {
		t.Fatalf("复制夹具失败: %v", err)
	}
	exe := filepath.Join(pkg, "bin", refExeName())
	cmd := exec.Command("go", "build", "-o", exe, "./tools/refplugin")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("编译参考插件失败: %v\n%s", err, out)
	}

	// 把清单里的入口改成实际文件名（入库存的是 windows 名，其它平台现改）。
	manPath := filepath.Join(pkg, ManifestFileName)
	b, err := os.ReadFile(manPath)
	if err != nil {
		t.Fatal(err)
	}
	var man Manifest
	if err := json.Unmarshal(b, &man); err != nil {
		t.Fatal(err)
	}
	entry := "bin/" + refExeName()
	man.Contributions.CLI.Entry = entry
	man.Contributions.Service.Entry = entry
	nb, err := json.MarshalIndent(man, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manPath, append(nb, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return pkg
}

// installRef 安装参考插件并返回记录。
func installRef(t *testing.T, m *Manager, pkg string) *Record {
	t.Helper()
	rec, err := m.Install(InstallOptions{Source: pkg})
	if err != nil {
		t.Fatalf("安装参考插件失败: %v", err)
	}
	return rec
}

// allowAll 放行一切（测试 Host API 用；真实授权来自注册表）。
type allowAll struct{}

func (allowAll) Permitted(c connsvc.Caller, _ connsvc.Capability) bool { return true }
func (allowAll) AllowedEnvs(c connsvc.Caller, _ connsvc.Capability) ([]string, bool) {
	return nil, true
}

// refDispatcher 造一个跑参考插件的 Dispatcher。
func refDispatcher(t *testing.T, s *Store, tokens *hostrpc.TokenStore) *Dispatcher {
	t.Helper()
	cfgPath := s.Locations().ConfigPath
	if err := config.SaveHosts(cfgPath, config.Hosts{SSHs: []config.NamedSsh{
		{Name: "prod", SSHConfig: config.SSHConfig{Host: "10.0.0.1", Port: 22, User: "u"}},
	}}); err != nil {
		t.Fatal(err)
	}
	return &Dispatcher{
		Store:      s,
		Tokens:     tokens,
		ConnSvc:    connsvc.New(connsvc.Options{ConfigPath: cfgPath, Authorizer: allowAll{}}),
		Authorizer: allowAll{},
		TTVersion:  "test",
	}
}

// parseRefOutput 解析参考插件打出的 JSON 行。
func parseRefOutput(t *testing.T, out string) map[string]any {
	t.Helper()
	line := strings.TrimSpace(out)
	if i := strings.LastIndex(line, "\n"); i >= 0 {
		line = line[i+1:]
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("参考插件输出不是 JSON: %v\n%s", err, out)
	}
	return m
}

// ---- CLI 转发 ----

// TestDispatchForwardsArgsAndExitCode 参数原样、退出码原样。
func TestDispatchForwardsArgsAndExitCode(t *testing.T) {
	m, s := newManager(t)
	pkg := buildReferencePlugin(t)
	rec := installRef(t, m, pkg)
	if _, err := m.Enable(rec.ID, Grants{}); err != nil {
		t.Fatal(err)
	}

	tokens := hostrpc.NewTokenStore()
	d := refDispatcher(t, s, tokens)
	var out, errb bytes.Buffer
	d.Stdout, d.Stderr = &out, &errb

	args := []string{"你好 world", "--flag=va lue", "emoji-🚀", "--exit", "7"}
	res, err := d.Run(context.Background(), rec, "prod", "", args)
	if err != nil {
		t.Fatalf("转发失败: %v（stderr=%s）", err, errb.String())
	}
	if res.ExitCode != 7 {
		t.Fatalf("退出码该原样是 7，得 %d", res.ExitCode)
	}
	got := parseRefOutput(t, out.String())
	rawArgs, _ := got["args"].([]any)
	if len(rawArgs) != 3 {
		t.Fatalf("该收到 3 个业务参数，得 %v", rawArgs)
	}
	for i, want := range []string{"你好 world", "--flag=va lue", "emoji-🚀"} {
		if rawArgs[i] != want {
			t.Errorf("第 %d 个参数 = %q, want %q", i, rawArgs[i], want)
		}
	}
}

// TestDispatchEchoesStdinAndStdout 标准流直通（插件自己写 stdout，TT 不加工）。
func TestDispatchEchoesStdinAndStdout(t *testing.T) {
	m, s := newManager(t)
	pkg := buildReferencePlugin(t)
	rec := installRef(t, m, pkg)
	if _, err := m.Enable(rec.ID, Grants{}); err != nil {
		t.Fatal(err)
	}
	d := refDispatcher(t, s, hostrpc.NewTokenStore())
	var out bytes.Buffer
	d.Stdout = &out
	d.Stdin = strings.NewReader("逐字节原样\n第二行\n")

	res, err := d.Run(context.Background(), rec, "", "", []string{"--echo-stdin"})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("退出码 = %d", res.ExitCode)
	}
	// stdout 里先是 stdin 的内容，最后一行才是参数 JSON。
	if !strings.Contains(out.String(), "逐字节原样\n第二行\n") {
		t.Fatalf("stdin 没有原样回显: %q", out.String())
	}
}

// TestDispatchForwardsLargeStdinAndStdout 1 MiB 级的 stdin/stdout 原样透传（P3-08）。
//
// 为什么单独一条：小数据走通只证明"接上了管子"，证明不了**没人截断、没人死锁**。
// 真实现场（插件吐一份大报表/收一份大报文）的失败方式恰好是这两种：
//   - 截断：中间某段用了固定大小的缓冲或 Scanner 默认上限；
//   - 死锁：一边写满管道另一边还没开始读（父进程先写 stdin 再读 stdout 就会这样）。
//
// 判据用**逐字节摘要**而不是长度：长度一样但内容错了同样不算"原样转发"。
// 数据用可复现的模式而不是全零：全零会让"只写了一半 + 补零"这类错误看不出来。
func TestDispatchForwardsLargeStdinAndStdout(t *testing.T) {
	const size = 1<<20 + 4093 // 1 MiB 出头，且不是整块大小（边界更容易暴露截断）
	big := make([]byte, size)
	for i := range big {
		big[i] = byte(i*131 + i/7)
	}
	want := sha256.Sum256(big)

	m, s := newManager(t)
	pkg := buildReferencePlugin(t)
	rec := installRef(t, m, pkg)
	if _, err := m.Enable(rec.ID, Grants{}); err != nil {
		t.Fatal(err)
	}
	d := refDispatcher(t, s, hostrpc.NewTokenStore())
	var out bytes.Buffer
	d.Stdout = &out
	d.Stdin = bytes.NewReader(big)

	// 死锁时的失败形态是"这条用例永远不返回"——给它一个上限，让失败变成可读的超时，
	// 而不是等 go test 的包级超时把整个包的其它用例一起拖死。
	type result struct {
		res DispatchResult
		err error
	}
	done := make(chan result, 1)
	go func() {
		res, err := d.Run(context.Background(), rec, "", "", []string{"--echo-stdin"})
		done <- result{res, err}
	}()
	var got result
	select {
	case got = <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("1 MiB 级 stdin/stdout 转发超时（管道读写在某处死锁了）")
	}
	if got.err != nil {
		t.Fatalf("转发失败: %v", got.err)
	}
	if got.res.ExitCode != 0 {
		t.Fatalf("退出码 = %d", got.res.ExitCode)
	}
	// stdout = stdin 原样回显 + 最后一行参数 JSON。
	if out.Len() <= size {
		t.Fatalf("stdout 只有 %d 字节，连回显的 %d 字节都不够（被截断）", out.Len(), size)
	}
	if sum := sha256.Sum256(out.Bytes()[:size]); sum != want {
		t.Fatalf("回显的 %d 字节与输入摘要不符（内容被改过或只转发了一部分）", size)
	}
	// 尾部 JSON 必须完整：它说明插件是正常跑完写出来的，而不是被中途掐掉。
	// 这里不能用 parseRefOutput（它按"最后一行"切，而回显内容是二进制、自己就带换行）；
	// 直接看"回显长度之后那一段"是不是完整的参数 JSON。
	tail := out.Bytes()[size:]
	if !bytes.HasPrefix(tail, []byte(`{"args":[]`)) {
		t.Fatalf("回显之后没有完整的参数 JSON（尾部 %q）", tail[:min(len(tail), 120)])
	}
}

// TestDispatchInjectsMinimalEnv 只注入最小环境，不注入完整配置或别人的令牌。
func TestDispatchInjectsMinimalEnv(t *testing.T) {
	m, s := newManager(t)
	pkg := buildReferencePlugin(t)
	rec := installRef(t, m, pkg)
	if _, err := m.Enable(rec.ID, Grants{}); err != nil {
		t.Fatal(err)
	}
	d := refDispatcher(t, s, hostrpc.NewTokenStore())
	// 故意在基础环境里塞一份"上一层的配置与令牌"：必须被抹掉。
	d.BaseEnv = append(os.Environ(), "TT_CONFIG=C:/secret/config.json", "TT_PLUGIN_TOKEN=leaked")
	var out bytes.Buffer
	d.Stdout = &out

	if _, err := d.Run(context.Background(), rec, "prod", "123", []string{"noop"}); err != nil {
		t.Fatal(err)
	}
	got := parseRefOutput(t, out.String())
	env, _ := got["env"].(map[string]any)
	if env["TT_CONFIG"] != "" {
		t.Errorf("不该把 TT_CONFIG 透给插件，得 %v", env["TT_CONFIG"])
	}
	if env["TT_PLUGIN_ID"] != rec.ID {
		t.Errorf("TT_PLUGIN_ID = %v", env["TT_PLUGIN_ID"])
	}
	if env["TT_PLUGIN_VERSION"] != rec.ActiveVersion {
		t.Errorf("TT_PLUGIN_VERSION = %v", env["TT_PLUGIN_VERSION"])
	}
	if env["TT_ENV"] != "prod" || env["TT_ENT"] != "123" {
		t.Errorf("环境上下文不对: TT_ENV=%v TT_ENT=%v", env["TT_ENV"], env["TT_ENT"])
	}
	if dir, _ := env["TT_PLUGIN_DATA_DIR"].(string); dir == "" || !strings.Contains(dir, "plugins") {
		t.Errorf("TT_PLUGIN_DATA_DIR 不对: %v", env["TT_PLUGIN_DATA_DIR"])
	}
	// 令牌必须存在且**不是**基础环境里那个（后者被封掉了）。
	tok, _ := env["TT_PLUGIN_TOKEN"].(string)
	if tok == "" || tok == "leaked" {
		t.Errorf("令牌没有正确签发/覆盖: %q", tok)
	}
	if !strings.Contains(tok, "…(") {
		t.Errorf("夹具该只回报令牌形态，得 %q", tok)
	}
}

// TestDispatchHostAPICallWorks 插件经注入的地址与令牌真的能调 Host API。
func TestDispatchHostAPICallWorks(t *testing.T) {
	m, s := newManager(t)
	pkg := buildReferencePlugin(t)
	rec := installRef(t, m, pkg)
	if _, err := m.Enable(rec.ID, Grants{Capabilities: []string{"profiles.list"}}); err != nil {
		t.Fatal(err)
	}
	tokens := hostrpc.NewTokenStore()
	d := refDispatcher(t, s, tokens)
	var out bytes.Buffer
	d.Stdout = &out

	if _, err := d.Run(context.Background(), rec, "prod", "", []string{"--hostrpc"}); err != nil {
		t.Fatal(err)
	}
	got := parseRefOutput(t, out.String())
	if e, _ := got["hostrpcError"].(string); e != "" {
		t.Fatalf("插件调 Host API 失败: %s", e)
	}
	hr, _ := got["hostrpc"].(map[string]any)
	if hr["hostApiVersion"] != connsvc.HostAPIVersion {
		t.Errorf("HostAPI 版本 = %v", hr["hostApiVersion"])
	}
	profs, _ := hr["profiles"].([]any)
	if len(profs) != 1 || profs[0] != "prod" {
		t.Errorf("插件该看到 prod 一个环境，得 %v", profs)
	}
	// 调用结束后令牌必须被撤销（插件进程已经退出）。
	if tokens.Count() != 0 {
		t.Errorf("退出后令牌该被撤销，剩 %d 个", tokens.Count())
	}
}

// TestDispatchInfraFailureIsNotPluginExit 入口缺失是**基础设施**失败，不是插件退出码。
func TestDispatchInfraFailureIsNotPluginExit(t *testing.T) {
	m, s := newManager(t)
	pkg := buildReferencePlugin(t)
	rec := installRef(t, m, pkg)
	if _, err := m.Enable(rec.ID, Grants{}); err != nil {
		t.Fatal(err)
	}
	// 删掉入口：模拟"版本目录被改坏"。
	dir := s.VersionDir(rec.ID, rec.ActiveVersion)
	if err := os.Remove(filepath.Join(dir, "bin", refExeName())); err != nil {
		t.Fatal(err)
	}
	d := refDispatcher(t, s, hostrpc.NewTokenStore())
	_, err := d.Run(context.Background(), rec, "", "", []string{"noop"})
	if err == nil {
		t.Fatal("入口缺失该报错")
	}
	if !errors.Is(err, ErrPluginSpawn) {
		t.Fatalf("该归类为 ErrPluginSpawn，得 %v", err)
	}
}

// ---- supervisor ----

// TestSupervisorStartsHealthChecksAndStops 启动 → 就绪 → 健康检查 → 停止收口。
func TestSupervisorStartsHealthChecksAndStops(t *testing.T) {
	m, s := newManager(t)
	pkg := buildReferencePlugin(t)
	rec := installRef(t, m, pkg)
	if _, err := m.Enable(rec.ID, Grants{}); err != nil {
		t.Fatal(err)
	}

	tokens := hostrpc.NewTokenStore()
	sup := NewSupervisor(s, tokens, nil, "", "test")
	sup.StartTimeout = 10 * time.Second
	sup.HealthTimeout = 10 * time.Second
	sup.StopGrace = 2 * time.Second
	sup.Logf = t.Logf

	inst, err := sup.Start(context.Background(), rec)
	if err != nil {
		t.Fatalf("启动服务失败: %v", err)
	}
	if inst.State != "running" || inst.Addr == "" || inst.PID <= 0 {
		t.Fatalf("实例状态不对: %+v", inst)
	}
	if _, ok := sup.Addr(rec.ID); !ok {
		t.Error("Addr 该能取到")
	}
	// instance.json 落盘（重启后能看出上一轮是谁）。
	statePath := filepath.Join(s.RunDir(rec.ID), "instance.json")
	if _, err := os.Stat(statePath); err != nil {
		t.Errorf("实例状态文件没写: %v", err)
	}
	// 日志文件里有 ready 行。
	logs, _ := os.ReadDir(s.LogsDir(rec.ID))
	if len(logs) == 0 {
		t.Error("该有服务日志")
	}

	if err := sup.Stop(rec.ID); err != nil {
		t.Fatalf("停止失败: %v", err)
	}
	if _, ok := sup.Addr(rec.ID); ok {
		t.Error("停止后不该还有地址")
	}
	if tokens.Count() != 0 {
		t.Errorf("停止后令牌该撤销，剩 %d", tokens.Count())
	}
}

// TestSupervisorStartTimeout 不就绪要有明确超时错误，而不是无限等。
func TestSupervisorStartTimeout(t *testing.T) {
	m, s := newManager(t)
	pkg := buildReferencePlugin(t)
	rec := installRef(t, m, pkg)
	if _, err := m.Enable(rec.ID, Grants{}); err != nil {
		t.Fatal(err)
	}

	sup := NewSupervisor(s, hostrpc.NewTokenStore(), nil, "", "test")
	sup.StartTimeout = 400 * time.Millisecond
	sup.HealthTimeout = 400 * time.Millisecond
	sup.StopGrace = time.Second
	sup.ExtraEnv = []string{"REF_NO_READY=1"}

	_, err := sup.Start(context.Background(), rec)
	if err == nil {
		t.Fatal("不打 ready 行该超时失败")
	}
	if !strings.Contains(err.Error(), "没有就绪") {
		t.Fatalf("错误该说明没就绪，得 %v", err)
	}
	// 失败要留日志路径，方便排障。
	if !strings.Contains(err.Error(), "service-") {
		t.Errorf("错误该带日志路径，得 %v", err)
	}
	sup.StopAll()
}

// ---- 授权 ----

// TestRegistryAuthorizerTrustsEnabledPlugins v2：启用即信任（不再逐项能力/环境授权），
// 但未启用/未知插件仍然拒绝。
func TestRegistryAuthorizerTrustsEnabledPlugins(t *testing.T) {
	m, s := newManager(t)
	pkg := buildReferencePlugin(t)
	rec := installRef(t, m, pkg)
	if _, err := m.Enable(rec.ID, Grants{}); err != nil {
		t.Fatal(err)
	}
	a := RegistryAuthorizer{Store: s}
	caller := connsvc.Caller{Kind: connsvc.CallerCLIPlugin, PluginID: rec.ID, InstanceID: "i"}

	if !a.Permitted(caller, connsvc.CapProfilesList) || !a.Permitted(caller, connsvc.CapDBConnect) {
		t.Error("启用后所有能力都该放行")
	}
	envs, all := a.AllowedEnvs(caller, connsvc.CapProfilesList)
	if !all || len(envs) != 0 {
		t.Errorf("启用后该返回全部环境, 得 %v all=%v", envs, all)
	}
	// 内置调用方不受插件授权限制。
	if !a.Permitted(connsvc.BuiltinCaller(), connsvc.CapDBConnect) {
		t.Error("内置调用方该放行")
	}
	// 未知插件一律拒绝（fail-closed）。
	ghost := connsvc.Caller{Kind: connsvc.CallerCLIPlugin, PluginID: "com.example.ghost", InstanceID: "i"}
	if a.Permitted(ghost, connsvc.CapProfilesList) {
		t.Error("未知插件该拒绝")
	}

	// 停用后立即失效（不需要重启）。
	if _, err := m.Disable(rec.ID); err != nil {
		t.Fatal(err)
	}
	if a.Permitted(caller, connsvc.CapProfilesList) {
		t.Error("停用后该立即拒绝")
	}
}

// ---- 安装夹具变体 ----

// TestInstallRejectsFixtureVariants 三个"坏包"变体都要在安装阶段被拒。
func TestInstallRejectsFixtureVariants(t *testing.T) {
	root := testkit.RepoRoot(t)
	cases := []struct {
		name string
		dir  string
		want string
	}{
		{"无贡献", "no-contribution", "没有任何贡献"},
		{"清单坏了", "invalid-manifest", "JSON"},
		{"命名空间冲突", "namespace-collision", "冲突"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, _ := newManager(t)
			src := filepath.Join(root, "testdata", "plugins", "variants", c.dir)
			_, err := m.Install(InstallOptions{Source: src})
			if err == nil {
				t.Fatalf("该拒绝 %s", c.dir)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Fatalf("错误该含 %q，得 %v", c.want, err)
			}
			reg, lerr := m.Store.Load()
			if lerr != nil {
				t.Fatal(lerr)
			}
			if len(reg.Plugins) != 0 {
				t.Errorf("失败安装不该留下注册项: %+v", reg.Plugins)
			}
		})
	}
}
