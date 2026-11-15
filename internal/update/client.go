package update

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"tt/internal/fetch"
)

// TimeoutAPI 一次 API 调用（取发布信息）的时限。
const timeoutAPI = 60 * time.Second

// timeoutDownload 一次资产下载的时限。40 MB 的安装包在慢线路上要几分钟，
// 所以给得宽；真正防挂死的是"每读一块就推进"的流式写法，而不是这个数。
const timeoutDownload = 30 * time.Minute

// ExplicitProxy 返回要显式使用的代理地址，空串表示"交给 Go 的环境变量规则"。
//
// 优先级：命令行 --proxy → 配置 net.proxy → 环境变量（由 http.ProxyFromEnvironment 处理）。
// 为什么最后那一档不自己读 HTTP(S)_PROXY：环境变量有一整套既有语义（HTTPS_PROXY 优先、
// NO_PROXY 例外、大小写两种写法），自己实现一份等于把 NO_PROXY 丢掉。
func ExplicitProxy(flagVal, cfgVal string) string {
	for _, v := range []string{flagVal, cfgVal} {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// Client 是对外 HTTP 的最小封装：一个发布源的读、一个资产的下载。
//
// 不带 cookie、不带 token、不带重定向到别的站点的宽容：更新器是"下载可执行文件"的
// 路径，能少一个变量就少一个。
type Client struct {
	HTTP  *http.Client
	Proxy string // 显式代理；""=按环境变量
	Agent string // User-Agent（GitHub API 要求带）

	LatestAPI string
	ListAPI   string
}

// NewClient 造一个客户端。proxy 为空时按环境变量走代理（含 NO_PROXY 例外）。
func NewClient(proxy, agent string) (*Client, error) {
	tr, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, errors.New("默认 HTTP 传输不可克隆")
	}
	rt := tr.Clone()
	if proxy != "" {
		u, err := url.Parse(proxy)
		if err != nil || u.Host == "" {
			return nil, fmt.Errorf("代理地址 %q 不合法（要形如 http://127.0.0.1:10808）", proxy)
		}
		rt.Proxy = http.ProxyURL(u)
	} else {
		rt.Proxy = http.ProxyFromEnvironment
	}
	if agent == "" {
		agent = "tt-update"
	}
	return &Client{
		HTTP:      &http.Client{Transport: rt},
		Proxy:     proxy,
		Agent:     agent,
		LatestAPI: LatestAPI,
		ListAPI:   ListAPI,
	}, nil
}

// Where 描述"这次会怎么出网"，给失败信息与 --dry-run 用：出网失败时，
// 用户第一个要问的就是"到底走没走代理"。
func (c *Client) Where() string {
	if c.Proxy != "" {
		return "经代理 " + c.Proxy
	}
	return "按环境变量里的代理设置（无则直连）"
}

// Latest 取最新发布。pre=true 时连预发布一起看（`/releases/latest` 有意跳过预发布）。
func (c *Client) Latest(ctx context.Context, pre bool) (*Release, error) {
	if !pre {
		b, err := c.get(ctx, c.LatestAPI)
		if err != nil {
			return nil, err
		}
		return ParseRelease(b)
	}
	b, err := c.get(ctx, c.ListAPI)
	if err != nil {
		return nil, err
	}
	list, err := ParseReleases(b)
	if err != nil {
		return nil, err
	}
	// 取最新的**能认出**的那个。认不出的 tag 只能跳过：装不了自己认不出名字的东西，
	// 但一个都认不出时必须报错 —— 静默退回一个更旧的正式版会骗人。
	var firstBad string
	for _, r := range list {
		if _, err := r.Version(); err == nil {
			return r, nil
		}
		if firstBad == "" {
			firstBad = r.Tag
		}
	}
	if firstBad != "" {
		return nil, fmt.Errorf("发布列表里没有能识别的版本号（第一个认不出的是 tag %q）", firstBad)
	}
	return nil, errors.New("发布列表是空的")
}

// get 取一个 JSON 响应体，带重试。
func (c *Client) get(ctx context.Context, u string) ([]byte, error) {
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			if err := sleepCtx(ctx, time.Duration(attempt)*2*time.Second); err != nil {
				return nil, err
			}
		}
		body, retry, err := c.getOnce(ctx, u)
		if err == nil {
			return body, nil
		}
		last = err
		if !retry {
			break
		}
	}
	return nil, last
}

func (c *Client) getOnce(ctx context.Context, u string) (body []byte, retry bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutAPI)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.Agent)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, true, fmt.Errorf("请求 %s 失败（%s）: %w", u, c.Where(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, retryableStatus(resp.StatusCode), statusError(u, resp)
	}
	// 发布 JSON 撑死几百 KB；给 8 MB 上限只为不让坏响应把内存吃光。
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, fmt.Errorf("读 %s 的响应失败: %w", u, err)
	}
	return b, false, nil
}

// statusError 把 HTTP 状态翻成"下一步该做什么"，而不是只报一个数字。
func statusError(u string, resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusForbidden, http.StatusTooManyRequests:
		return fmt.Errorf("访问 %s 被拒（HTTP %d）：GitHub 对未认证请求限流（每小时 60 次）。"+
			"稍后再试，或改用已下载的安装包", u, resp.StatusCode)
	case http.StatusNotFound:
		return fmt.Errorf("访问 %s 返回 404：这个仓库还没有发布（或发布被删了）", u)
	default:
		return fmt.Errorf("访问 %s 返回 HTTP %d", u, resp.StatusCode)
	}
}

func retryableStatus(code int) bool { return code >= 500 || code == http.StatusTooManyRequests }

// Download 把资产下到 dest，并校验 sha256 与发布方给的摘要一致。
//
// 实现已抽到 internal/fetch（P2-08）：下载/重试/校验/失败不留半成品是**字节级**原语，
// 插件更新用的是同一份。这里保留本包的方法，只把"TT 自更新的策略"留在本包：
// 摘要必须来自发布方（Asset.HexDigest）、超时、UA、以及"出口是哪条代理"这句提示。
//
// wantHex 为空表示这次没有可比对的摘要 —— 调用方不该走到这里（Asset.HexDigest 就是门槛）。
func (c *Client) Download(ctx context.Context, a Asset, dest, wantHex string, progress func(got, total int64)) error {
	return fetch.Download(ctx, c.HTTP, fetch.Request{
		URL: a.URL, Name: a.Name, Dest: dest,
		SHA256: wantHex, Size: a.Size,
		Agent: c.Agent, Via: c.Where(),
		Timeout:  timeoutDownload,
		Progress: progress,
	})
}

// fileSHA256 算一个已存在文件的 sha256（同包内复用共用原语）。
func fileSHA256(path string) (string, error) { return fetch.SHA256File(path) }

// sleepCtx 可被取消的等待（重试之间的退避）。
func sleepCtx(ctx context.Context, d time.Duration) error { return fetch.SleepCtx(ctx, d) }

// humanBytes 人读的大小，用在进度与失败信息里。
func humanBytes(n int64) string { return fetch.HumanBytes(n) }

// HumanBytes 供命令层复用（同一个数字只声明一次）。
func HumanBytes(n int64) string { return fetch.HumanBytes(n) }
