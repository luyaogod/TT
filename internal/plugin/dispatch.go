package plugin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/internal/hostrpc"
	"github.com/luyaogod/TT/internal/winproc"
)

// Dispatching 说明第三方 CLI 的模型：
//
//	tt <namespace> …  →  一次性进程（插件自己的 exe）
//	                      └─ 期间有一个**临时 Host API**（本进程内的 loopback 服务）
//
// Host API 端点随插件进程一起生灭，不去找 tt serve 借令牌 —— CLI 与 serve 是两个进程，
// 内存里的令牌表不共享；让 CLI 自己起一个只服务这次调用的端点，既不需要共享秘密，
// 也不需要"必须先跑 tt serve"这个前置条件。

// Dispatcher 执行一次第三方 CLI 调用。
type Dispatcher struct {
	Store  *Store
	Tokens *hostrpc.TokenStore
	// ConnSvc 进程内连接服务（Host API 的实现）。nil = 按 Store 的配置路径懒建。
	ConnSvc *connsvc.Service
	// Authorizer 授权来源；nil = 只放行内置调用方（插件拿不到任何环境）。
	Authorizer connsvc.Authorizer
	// TTVersion 进 Capabilities。
	TTVersion string

	// 以下为测试缝（nil 走缺省）。
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	BaseEnv   []string
	NewServer func(http.Handler) (url string, stop func(), err error)
}

// DispatchResult 一次转发的结局。
type DispatchResult struct {
	// ExitCode 插件进程的退出码（**原样**返回，TT 不重映射）。
	ExitCode int
	// InstanceID 本次调用用的实例身份（日志/排障用）。
	InstanceID string
}

// ErrPluginSpawn 起不来（程序缺失/不可执行/清单坏了）—— 属于 TT 基础设施失败。
var ErrPluginSpawn = errors.New("插件进程启动失败")

// Run 在一次调用里跑完插件进程。
//
// 返回约定：
//   - error 只表示**基础设施**失败（清单坏了、入口不存在、起不来）；此时调用方用 TT 的
//     错误信封与退出码规则处理；
//   - (DispatchResult, nil) 表示插件真的跑过，ExitCode 是它自己的退出码（含非零）。
func (d *Dispatcher) Run(ctx context.Context, rec *Record, envName, ent string, args []string) (DispatchResult, error) {
	if rec == nil {
		return DispatchResult{}, fmt.Errorf("%w: 没有插件记录", ErrPluginSpawn)
	}
	dir := d.Store.ActiveDir(rec)
	if dir == "" {
		return DispatchResult{}, fmt.Errorf("%w: 插件 %s 没有活跃版本", ErrPluginSpawn, rec.ID)
	}
	man, err := LoadManifest(dir)
	if err != nil {
		return DispatchResult{}, fmt.Errorf("%w: %v", ErrPluginSpawn, err)
	}
	rel, err := man.ResolveCLIEntry()
	if err != nil {
		return DispatchResult{}, fmt.Errorf("%w: %v", ErrPluginSpawn, err)
	}
	exe := filepath.Join(dir, filepath.FromSlash(rel))
	if st, serr := os.Stat(exe); serr != nil || !st.Mode().IsRegular() {
		return DispatchResult{}, fmt.Errorf("%w: 入口不存在或不是普通文件: %s", ErrPluginSpawn, exe)
	}

	instanceID, ierr := newInstanceID("cli")
	if ierr != nil {
		return DispatchResult{}, ierr
	}
	caller := connsvc.Caller{Kind: connsvc.CallerCLIPlugin, PluginID: rec.ID, InstanceID: instanceID}

	tokens := d.Tokens
	if tokens == nil {
		tokens = hostrpc.NewTokenStore()
	}
	token := tokens.Issue(caller)
	defer tokens.Revoke(token)

	svc := d.connSvc()
	// 会话按实例收口：只关这次调用开的会话，不牵连同插件的其他实例。
	defer svc.CloseForInstance(rec.ID, instanceID)

	apiURL, stopAPI, err := d.startHostAPI(svc, tokens)
	if err != nil {
		return DispatchResult{}, err
	}
	defer stopAPI()

	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Dir = "" // 继承当前工作目录：插件的相对路径语义应当与用户在哪个目录无关地保持
	cmd.Stdin = d.stdin()
	cmd.Stdout = d.stdout()
	cmd.Stderr = d.stderr()
	cmd.Env = append(d.baseEnv(), pluginEnv(rec, d.Store.DataDir(rec.ID), instanceID, caller, token, apiURL, envName, ent)...)
	winproc.HideConsole(cmd)

	if err := cmd.Start(); err != nil {
		return DispatchResult{}, fmt.Errorf("%w: %v", ErrPluginSpawn, err)
	}

	// 取消时结束整棵进程树：只杀包装进程会留下孙进程继续占资源。
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-ctx.Done():
		_ = winproc.KillTree(cmd.Process.Pid)
		<-done
		return DispatchResult{ExitCode: 130, InstanceID: instanceID}, nil
	case werr := <-done:
		code := 0
		if cmd.ProcessState != nil {
			code = cmd.ProcessState.ExitCode()
		}
		if code < 0 {
			// 被信号杀掉：给一个可读的码，别让包装脚本拿到 -1。
			code = 1
		}
		if werr != nil && code == 0 {
			return DispatchResult{}, fmt.Errorf("%w: %v", ErrPluginSpawn, werr)
		}
		return DispatchResult{ExitCode: code, InstanceID: instanceID}, nil
	}
}

// connSvc 取连接服务（懒建）。
func (d *Dispatcher) connSvc() *connsvc.Service {
	if d.ConnSvc != nil {
		return d.ConnSvc
	}
	cfgPath := ""
	if loc := d.Store.Locations(); loc != nil {
		cfgPath = loc.ConfigPath
	}
	d.ConnSvc = connsvc.New(connsvc.Options{
		ConfigPath: cfgPath,
		TTVersion:  d.TTVersion,
		Authorizer: d.Authorizer,
	})
	return d.ConnSvc
}

// startHostAPI 起一个只服务本次调用的 Host API 端点。
func (d *Dispatcher) startHostAPI(svc *connsvc.Service, tokens *hostrpc.TokenStore) (string, func(), error) {
	h := hostrpc.NewHandler(hostrpc.Deps{Service: svc, Auth: hostrpc.NewTokenAuthenticator(tokens)})
	if d.NewServer != nil {
		return d.NewServer(h)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, fmt.Errorf("%w: 起 Host API 端点失败: %v", ErrPluginSpawn, err)
	}
	srv := &http.Server{Handler: h, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = srv.Serve(ln) }()
	url := "http://" + ln.Addr().String() + hostrpc.Prefix
	stop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}
	return url, stop, nil
}

// pluginEnv 注入的最小环境变量。
//
// 刻意**不**注入：完整配置（含口令）、其他插件的令牌、TT 自己的内部开关。
// 插件该知道的只有"我是谁、我的数据在哪、Host API 在哪、这次是哪个环境"。
func pluginEnv(rec *Record, dataDir, instanceID string, caller connsvc.Caller, token, apiURL, envName, ent string) []string {
	env := []string{
		"TT_HOST_API_URL=" + apiURL,
		"TT_PLUGIN_TOKEN=" + token,
		"TT_PLUGIN_ID=" + rec.ID,
		"TT_PLUGIN_VERSION=" + rec.ActiveVersion,
		"TT_PLUGIN_INSTANCE_ID=" + instanceID,
		"TT_PLUGIN_KIND=" + string(caller.Kind),
	}
	if dataDir != "" {
		env = append(env, "TT_PLUGIN_DATA_DIR="+dataDir)
	}
	if envName != "" {
		env = append(env, "TT_ENV="+envName)
	}
	if ent != "" {
		env = append(env, "TT_ENT="+ent)
	}
	return env
}

// baseEnv 基础环境：继承当前进程的环境，但先抹掉所有 TT_PLUGIN_*/TT_HOST_API_URL/TT_ENV，
// 避免上一层的值漏进插件（嵌套调用时会串味）。
func (d *Dispatcher) baseEnv() []string {
	base := d.BaseEnv
	if base == nil {
		base = os.Environ()
	}
	out := make([]string, 0, len(base)+8)
	for _, kv := range base {
		key := kv
		if i := strings.IndexByte(kv, '='); i >= 0 {
			key = kv[:i]
		}
		switch key {
		case "TT_PLUGIN_TOKEN", "TT_HOST_API_URL", "TT_PLUGIN_ID", "TT_PLUGIN_VERSION",
			"TT_PLUGIN_INSTANCE_ID", "TT_PLUGIN_KIND", "TT_PLUGIN_DATA_DIR", "TT_ENV", "TT_ENT",
			"TT_CONFIG", "TT_PLUGIN_LOGS_DIR", "TT_PLUGIN_LISTEN":
			continue
		}
		out = append(out, kv)
	}
	return out
}

func (d *Dispatcher) stdin() io.Reader {
	if d.Stdin != nil {
		return d.Stdin
	}
	return os.Stdin
}

func (d *Dispatcher) stdout() io.Writer {
	if d.Stdout != nil {
		return d.Stdout
	}
	return os.Stdout
}

func (d *Dispatcher) stderr() io.Writer {
	if d.Stderr != nil {
		return d.Stderr
	}
	return os.Stderr
}

// newInstanceID 生成实例身份（前缀 + 随机，不依赖 PID：PID 会被复用）。
func newInstanceID(prefix string) (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成实例 ID 失败: %w", err)
	}
	return prefix + "-" + hex.EncodeToString(b), nil
}
