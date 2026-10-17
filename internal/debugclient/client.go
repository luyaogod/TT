// Package debugclient 是**调试面 REST 的唯一客户端**。
//
// # 为什么要有它
//
// 从前有两份"调调试面"的实现：
//
//   - `internal/cli/debug` 里的 `dbgAPI`（私有的，只给 debug 命令用）；
//   - `internal/cli/fep` 里的 `runSelf(cmd, "debug", …)` —— 完全绕开 HTTP，改为**起子进程**。
//
// 第二条是这套代码里最后一块跨域耦合：fep 想"收口会话 / 点名环境 / 设 ENT / 起作业"，
// 只能把 `tt debug` 当命令再跑一遍。代价是每次一个进程、参数只能用 argv 表达、错误码
// 得从子进程的退出码里猜 —— 于是"编排"只能活在 CLI 层，永远沉不到内核里。
//
// 本包把这件事收成一份：**谁要调调试面，都用它**。cli/debug 与 cli/fep 都依赖它，
// 彼此仍然互不 import（那是解耦的底线）。
//
// # 边界
//
// 只做三件事：拼 URL、发请求、把失败信封翻成 `*Failure`。不含任何业务判断 ——
// "该落哪条会话""名字合不合法"一律由服务端的信封回答（`code` / `candidates` / `hint`）。
package debugclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tt/internal/config"
	"tt/internal/debug"
	"tt/internal/target"
)

// State 是后台服务写的状态文件（`<数据目录>/.tt-serve.json`）里本包关心的部分。
//
// 字段名必须与 `internal/cli/debug` 写出去的那份一致 —— 读取侧只认文件，不认谁的代码写的。
type State struct {
	PID     int    `json:"pid"`
	URL     string `json:"url"`
	APIBase string `json:"apiBase,omitempty"`
}

// ReadState 读数据目录里的服务状态文件（读不到就是"没有实例在跑"，返回 nil, nil）。
func ReadState(dir string) (*State, error) {
	b, err := os.ReadFile(filepath.Join(dir, config.ServeStateFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var st State
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, fmt.Errorf("debugclient: 状态文件 %s 解析失败: %w", dir, err)
	}
	return &st, nil
}

// Discover 算出调试面的基址（已含挂载前缀）。
//
// 顺序与 CLI 侧一致：状态文件里的真实地址 > 配置里的 listen > 内置默认。
// 前缀（`/debug/api` 或空）只有状态文件知道 —— 统一服务挂在 `/debug/api`，独立调试服务在根。
func Discover(dir, listenFallback string) string {
	prefix := ""
	url := ""
	if st, err := ReadState(dir); err == nil && st != nil {
		url, prefix = st.URL, st.APIBase
	}
	if url == "" {
		if l := strings.TrimSpace(listenFallback); l != "" {
			url = "http://" + l
		} else {
			url = "http://" + debug.DefaultListen
		}
	}
	return joinBase(url, prefix)
}

// joinBase 补挂载前缀；URL 自己已经指到 API 上的不重复拼。
func joinBase(url, prefix string) string {
	u := strings.TrimRight(url, "/")
	if strings.HasSuffix(u, "/api") {
		return u
	}
	return u + prefix
}

// Failure 是服务端失败信封的客户端形态。
//
// 它实现 `ExitCode()`，所以调用方（`tt debug` / `tt fep`）的退出码自动与码表一致 ——
// 这正是"两条线一套规矩"的落点：从前 fep 只能拿子进程的退出码反推发生了什么。
type Failure struct {
	Status     int                // HTTP 状态（只用于诊断，分支一律读 Code）
	Code       string             // 稳定机器码（契约 C5.15）
	Message    string             // 给人和 AI 看的一句话
	Hint       string             // 可执行的下一步（可空）
	Candidates []target.Candidate // C5.12 九键候选（只在码表要求时出现）
}

func (f *Failure) Error() string {
	if f == nil {
		return ""
	}
	msg := f.Message
	if msg == "" {
		msg = f.Code
	}
	if f.Hint = strings.TrimSpace(f.Hint); f.Hint != "" {
		msg += "（" + f.Hint + "）"
	}
	return msg
}

// ExitCode 让调用方把 `*Failure` 当作带码的错误返回（根命令按它取退出码）。
//
// 分档：这里一律按"新场景"取表 —— 走 HTTP 的都是新写法（`--session`/`--name` 这一族），
// 老脚本不会出现在这条路径上。
func (f *Failure) ExitCode() int {
	if f == nil || f.Code == "" {
		return 1
	}
	return target.ExitFor(f.Code, true, false)
}

// Client 是调试面的客户端。零值不可用，请用 New。
type Client struct {
	Base  string // 已含挂载前缀
	Actor string // X-Actor：ai / human
	HTTP  *http.Client
}

// New 按基址建客户端。base 应已含挂载前缀（用 Discover 算）。
func New(base string) *Client {
	return &Client{
		Base:  strings.TrimRight(base, "/"),
		Actor: actorDefault(),
		HTTP:  &http.Client{Timeout: 5 * time.Minute},
	}
}

// actorDefault 与 `internal/cli/debug` 的 actorSelf 同一套取值：
// TT_ACTOR 由"人点的按钮"那条链路注入（serve 起子进程时带上），默认按 ai 算。
func actorDefault() string {
	if v := strings.TrimSpace(os.Getenv(debug.ActorEnv)); v != "" {
		return v
	}
	return "ai"
}

// Do 发一次请求并返回响应体；非 2xx 翻成 `*Failure`。
func (c *Client) Do(method, path string, body any) ([]byte, error) {
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.Base+path, rd)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Actor", c.Actor)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("无法连接调试服务(%s): %w —— 请先运行 tt serve(或 tt debug serve)", c.Base, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		if f := failureOf(resp.StatusCode, data); f != nil {
			return nil, f
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return data, nil
}

// failureOf 把失败信封解成 `*Failure`；不是信封就返回 nil（由调用方退化成 HTTP 错）。
func failureOf(status int, data []byte) *Failure {
	var e struct {
		Code       string             `json:"code"`
		Error      string             `json:"error"`
		Hint       string             `json:"hint"`
		Candidates []target.Candidate `json:"candidates"`
	}
	if json.Unmarshal(data, &e) != nil || strings.TrimSpace(e.Error) == "" {
		return nil
	}
	return &Failure{Status: status, Code: e.Code, Message: e.Error, Hint: e.Hint, Candidates: e.Candidates}
}

// Brief 是会话列表里的一项（身份七键 + 作业状态）。
//
// 只声明本包**真的会读**的字段：多出来的键照旧能被调用方从原始 JSON 里取，
// 但客户端不该替服务端再抄一份完整事实（C5.17–C5.19）。
type Brief struct {
	ID      string `json:"id"`
	Ref     string `json:"ref"`
	Display string `json:"display"`
	Name    string `json:"name"`
	Env     string `json:"env"`
	Ent     string `json:"ent"`
	State   string `json:"state"`
	N       int    `json:"n"`
	Epoch   string `json:"epoch"`
	Prog    string `json:"prog"`
}

// Sessions 列出全部会话。
func (c *Client) Sessions() ([]Brief, error) {
	data, err := c.Do(http.MethodGet, "/sessions", nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Sessions []Brief `json:"sessions"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("debugclient: 会话列表解析失败: %w", err)
	}
	return out.Sessions, nil
}

// Resolve 把"别名 / ref / 句柄"解析成一条会话。
//
// 规则与 CLI 侧的目标层一致：**精确命中一条才算**；0 条 `SESSION_NOT_FOUND`、
// 多条 `SESSION_AMBIGUOUS`（都带候选）—— 客户端不做"唯一就落"以外的推断。
func (c *Client) Resolve(want string) (Brief, error) {
	want = strings.TrimSpace(want)
	if want == "" {
		return Brief{}, &Failure{Code: target.CodeUsage, Message: "没有点名会话（别名 / ref / 句柄）"}
	}
	list, err := c.Sessions()
	if err != nil {
		return Brief{}, err
	}
	var hits []Brief
	for _, b := range list {
		if b.ID == want || b.Ref == want || (b.Name != "" && strings.EqualFold(b.Name, want)) {
			hits = append(hits, b)
		}
	}
	switch len(hits) {
	case 0:
		return Brief{}, &Failure{Code: target.CodeNotFound,
			Message: "没有这条会话: " + want,
			Hint:    "tt debug status 可看全部会话，照抄 <ref> 那列重发",
		}
	case 1:
		return hits[0], nil
	default:
		cands := make([]target.Candidate, 0, len(hits))
		for _, h := range hits {
			cands = append(cands, target.Candidate{
				ID: h.ID, Ref: h.Ref, Display: h.Display, Name: h.Name,
				Env: h.Env, Ent: h.Ent, N: h.N, Prog: h.Prog, State: h.State,
			})
		}
		return Brief{}, &Failure{Code: target.CodeAmbiguous,
			Message: fmt.Sprintf("%q 命中了 %d 条会话，无法确定是哪一条", want, len(hits)),
			Hint:    "照抄 <ref> 那一列重发 --session",
			Candidates: cands,
		}
	}
}

// Snapshot 取一条会话的完整快照（原样返回 JSON，调用方按需取字段）。
func (c *Client) Snapshot(id string) (map[string]any, error) {
	data, err := c.Do(http.MethodGet, "/sessions/"+id, nil)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("debugclient: 快照解析失败: %w", err)
	}
	return out, nil
}

// WaitStopped 轮询到会话进入停止态（或超时）。
//
// 返回 (快照, 是否停站, 错误)。超时**不是**错误 —— 调用方通常要照常往下走并报告超时。
// 轮询而不是长请求：与既有 `waitStopped` 的语义保持一致，服务端不必为它挂住一个 handler。
func (c *Client) WaitStopped(id string, timeout time.Duration) (map[string]any, bool, error) {
	deadline := time.Now().Add(timeout)
	for {
		snap, err := c.Snapshot(id)
		if err != nil {
			return nil, false, err
		}
		switch stateOf(snap) {
		case "stopped":
			return snap, true, nil
		case "exit":
			return snap, false, fmt.Errorf("会话 %s 已断开", id)
		}
		if time.Now().After(deadline) {
			return snap, false, nil
		}
		time.Sleep(time.Second)
	}
}

// stateOf 读快照里的会话状态。
func stateOf(snap map[string]any) string {
	if v, ok := snap["state"].(string); ok {
		return v
	}
	return ""
}
