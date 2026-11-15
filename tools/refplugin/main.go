// refplugin 是测试用的**参考插件**程序：同一个二进制既能当一次性 CLI，也能当常驻服务。
//
// 它属于测试夹具，不进发行包。存在的理由：
//   - CLI 转发与 supervisor 的契约（参数原样、标准流、退出码、就绪行、健康检查、
//     /shutdown、Host API 调用）必须能用**真进程**验，而不是靠桩函数猜；
//   - 可控故障（崩溃/拒绝退出/慢启动）只能由真进程演。
//
// 用法（由测试驱动，不是给人敲的）：
//
//	refplugin args …            打印收到的参数与关键环境变量（JSON 一行）
//	refplugin --exit N …        以 N 退出
//	refplugin --crash           立即 panic（模拟插件崩溃）
//	refplugin --sleep-ms N      睡 N 毫秒再继续
//	refplugin --echo-stdin      把 stdin 原样拷到 stdout
//	refplugin --hostrpc         调 Host API：GetCapabilities + List，打印结果
//	refplugin service           常驻模式：监听 TT_PLUGIN_LISTEN，打 ready 行
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/luyaogod/TT/internal/gen/hostv1"
	"github.com/luyaogod/TT/internal/gen/hostv1/hostv1connect"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "service" {
		runService()
		return
	}
	runCLI(args)
}

// runCLI 一次性模式。
func runCLI(args []string) {
	exitCode := 0
	crash := false
	sleepMS := 0
	echoStdin := false
	hostrpcCheck := false
	rest := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--exit":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil {
					exitCode = n
				}
				i++
			}
		case "--crash":
			crash = true
		case "--sleep-ms":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil {
					sleepMS = n
				}
				i++
			}
		case "--echo-stdin":
			echoStdin = true
		case "--hostrpc":
			hostrpcCheck = true
		default:
			rest = append(rest, a)
		}
	}

	if sleepMS > 0 {
		time.Sleep(time.Duration(sleepMS) * time.Millisecond)
	}
	if echoStdin {
		if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
			fmt.Fprintln(os.Stderr, "读 stdin 失败:", err)
			os.Exit(1)
		}
	}
	if crash {
		panic("refplugin: 模拟插件崩溃")
	}

	out := map[string]any{
		"args": rest,
		"env": map[string]string{
			"TT_HOST_API_URL":       os.Getenv("TT_HOST_API_URL"),
			"TT_PLUGIN_ID":          os.Getenv("TT_PLUGIN_ID"),
			"TT_PLUGIN_VERSION":     os.Getenv("TT_PLUGIN_VERSION"),
			"TT_PLUGIN_INSTANCE_ID": os.Getenv("TT_PLUGIN_INSTANCE_ID"),
			"TT_PLUGIN_KIND":        os.Getenv("TT_PLUGIN_KIND"),
			"TT_PLUGIN_DATA_DIR":    os.Getenv("TT_PLUGIN_DATA_DIR"),
			"TT_ENV":                os.Getenv("TT_ENV"),
			"TT_ENT":                os.Getenv("TT_ENT"),
			// 这些**不该**出现在插件环境里（断言用）。
			"TT_CONFIG":       os.Getenv("TT_CONFIG"),
			"TT_PLUGIN_TOKEN": maskToken(os.Getenv("TT_PLUGIN_TOKEN")),
		},
	}
	if hostrpcCheck {
		out["hostrpc"], out["hostrpcError"] = callHostAPI()
	}
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
	os.Exit(exitCode)
}

// maskToken 只报"有没有令牌"，不把令牌本身打到 stdout（它会被测试日志留存）。
func maskToken(t string) string {
	if t == "" {
		return ""
	}
	if len(t) <= 8 {
		return "***"
	}
	return t[:4] + "…(" + strconv.Itoa(len(t)) + ")"
}

// callHostAPI 用 TT 注入的地址与令牌调两个方法。
func callHostAPI() (any, string) {
	base := os.Getenv("TT_HOST_API_URL")
	token := os.Getenv("TT_PLUGIN_TOKEN")
	if base == "" || token == "" {
		return nil, "缺少 TT_HOST_API_URL 或 TT_PLUGIN_TOKEN"
	}
	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: &authTransport{base: http.DefaultTransport, token: token},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	hosts := hostv1connect.NewHostServiceClient(httpClient, base)
	caps, err := hosts.GetCapabilities(ctx, connect.NewRequest(&hostv1.GetCapabilitiesRequest{}))
	if err != nil {
		return nil, "GetCapabilities: " + err.Error()
	}
	profs := hostv1connect.NewProfilesServiceClient(httpClient, base)
	list, err := profs.List(ctx, connect.NewRequest(&hostv1.ListRequest{}))
	result := map[string]any{
		"hostApiVersion": caps.Msg.GetCapabilities().GetHostApiVersion(),
		"features":       len(caps.Msg.GetCapabilities().GetFeatures()),
	}
	if err != nil {
		result["profilesError"] = err.Error()
		return result, ""
	}
	names := make([]string, 0, len(list.Msg.GetProfiles()))
	for _, p := range list.Msg.GetProfiles() {
		names = append(names, p.GetEnv())
	}
	result["profiles"] = names
	return result, ""
}

// authTransport 给每个请求加 Bearer 令牌。
type authTransport struct {
	base  http.RoundTripper
	token string
}

func (t *authTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(r)
}

// runService 常驻模式：按契约打 ready 行、响应健康检查与 /shutdown。
func runService() {
	addr := strings.TrimSpace(os.Getenv("TT_PLUGIN_LISTEN"))
	if addr == "" {
		addr = "127.0.0.1:0"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "监听失败:", err)
		os.Exit(1)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		go func() { time.Sleep(50 * time.Millisecond); os.Exit(0) }()
	})
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"plugin":"refplugin"}`))
	})
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	// 测试用的两个可控开关：
	//   REF_SLOW_READY_MS 延迟打 ready 行（测启动超时与就绪判定）；
	//   REF_NO_READY      干脆不打（模拟"进程活着但永远不就绪"）。
	if v := strings.TrimSpace(os.Getenv("REF_SLOW_READY_MS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			time.Sleep(time.Duration(n) * time.Millisecond)
		}
	}
	// ready 行：**必须在开始 Serve 之前**打，否则测试可能先看到端口再等不到行。
	if os.Getenv("REF_NO_READY") == "" {
		fmt.Printf("{\"ready\":true,\"addr\":%q}\n", ln.Addr().String())
		_ = os.Stdout.Sync()
	}
	go func() { _ = srv.Serve(ln) }()

	// stdin 关闭即优雅退出（与 tt serve 的约定一致）。
	go func() {
		_, _ = io.Copy(io.Discard, os.Stdin)
		os.Exit(0)
	}()
	select {}
}
