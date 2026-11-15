// Package fetch 是"取一份外来字节并校验"的共享原语：带重试的 GET、落盘下载、
// 大小与 SHA-256 校验、失败不留半成品。
//
// 为什么单独成包（P2-08）：TT 有两条**互不相干**的取字节路径 —— tt 自身更新
// （internal/update：固定 GitHub 仓库 + 整包 kind 规则）与插件更新（internal/plugin：
// 发布索引 + 按 {os}-{arch} 选资产）。两者的策略不该共用（仓库、资产命名、版本判据
// 都不同），但"下载、重试、校验、失败删半成品"是同一套字节级动作。抽到这里以后，
// 任何一侧修下载缺陷都不会漏掉另一侧。
//
// 这里**不含 URL 策略**：https-only 是调用方的判断（插件发布索引要求 https；tt 自更新
// 跟随自己的源；本机 file:// 只用于离线和测试）。把策略和原语混在一起，测试就没法在
// 不放开生产策略的前提下用 httptest 验下载。
package fetch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultTimeout 一次下载的时限。真正防挂死的是"每读一块就推进"的流式写法，
	// 这个数只是兜底。
	DefaultTimeout = 30 * time.Minute
	// DefaultRetries 网络失败的额外重试次数（不含首次尝试）。
	DefaultRetries = 2
	// DefaultMaxBytes 一个 JSON 响应体的读入上限（发布索引撑死几百 KB）。
	DefaultMaxBytes = 8 << 20
)

// Request 一次落盘下载。
//
// SHA256 为空 = 拒绝下载：没有摘要就没有"下到的字节 == 发布方发布的字节"这条判据，
// 而这是整条路径上唯一不靠信任的判断（与 tt 自更新同一条纪律）。
type Request struct {
	// URL 支持三种写法：https://…、file:///…、本机绝对路径（离线/测试用）。
	URL string
	// Name 展示名（错误信息里用；空则取 URL 末段）。
	Name string
	// Dest 目标文件路径。
	Dest string
	// SHA256 期望的十六进制摘要（大小写不敏感）。
	SHA256 string
	// Size 期望字节数；<=0 表示这次不校验大小（仍校验摘要）。
	Size int64
	// Agent User-Agent（有些源要求非空）。
	Agent string
	// Via 描述出口（代理/直连），只用于把失败信息说清楚。
	Via string
	// Timeout 单次尝试的时限；0 = DefaultTimeout。
	Timeout time.Duration
	// Retries 额外重试次数；0 = DefaultRetries，负数 = 一次也不重试（测试用）。
	Retries int
	// Progress 下载进度回调（可为 nil）。
	Progress func(got, total int64)
}

// Download 把 req.URL 下到 req.Dest 并校验摘要与大小。
//
// 两条纪律：
//   - Dest 已存在且摘要对得上 → 直接返回（下载是幂等的：`--dry-run` 之后重跑不必再下）。
//   - 先写 Dest.part 再改名；任何失败都删掉半成品，绝不把自己校验不过的字节留在目标位置。
func Download(ctx context.Context, client *http.Client, req Request) error {
	name := req.Name
	if name == "" {
		name = nameOf(req.URL)
	}
	want := normalizeHex(req.SHA256)
	if want == "" {
		return fmt.Errorf("资产 %s 没有 sha256 摘要：没有可校验的东西，拒绝下载", name)
	}
	if req.Dest == "" {
		return errors.New("下载目标路径为空")
	}
	// 幂等：目标已经是那份字节就不用再下。
	if got, err := SHA256File(req.Dest); err == nil && got == want {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(req.Dest), 0o755); err != nil {
		return fmt.Errorf("创建下载目录失败: %w", err)
	}

	kind, local, err := resolveSource(req.URL)
	if err != nil {
		return fmt.Errorf("资产 %s 的地址不可用: %w", name, err)
	}
	if kind == "file" {
		return copyLocal(name, local, req, want)
	}

	attempts := req.Retries
	if attempts < 0 {
		attempts = 0
	} else if attempts == 0 {
		attempts = DefaultRetries
	}
	var last error
	for attempt := 0; attempt <= attempts; attempt++ {
		if attempt > 0 {
			if err := SleepCtx(ctx, time.Duration(attempt)*2*time.Second); err != nil {
				return err
			}
		}
		err := downloadOnce(ctx, client, name, req, want)
		if err == nil {
			return nil
		}
		last = err
		if !retryable(err) {
			break
		}
	}
	return last
}

// httpError 一个可判定的 HTTP 状态错误（决定要不要重试）。
type httpError struct {
	url  string
	code int
	msg  string
}

func (e *httpError) Error() string { return e.msg }

func (e *httpError) retryable() bool { return e.code >= 500 || e.code == http.StatusTooManyRequests }

// retryable 这次失败值不值得再试一次。
func retryable(err error) bool {
	var he *httpError
	if errors.As(err, &he) {
		return he.retryable()
	}
	// 本地文件/摘要/大小这类失败重试没有意义。
	var le *localError
	if errors.As(err, &le) {
		return false
	}
	if errors.Is(err, errDigest) || errors.Is(err, errSize) {
		return false
	}
	return true
}

var (
	errDigest = errors.New("摘要不符")
	errSize   = errors.New("大小不符")
)

// localError 本机来源失败（重试没有意义）。
type localError struct{ err error }

func (e *localError) Error() string { return e.err.Error() }
func (e *localError) Unwrap() error { return e.err }

// downloadOnce 一次尝试：HTTP 流式落盘 → 校验 → 改名。
func downloadOnce(ctx context.Context, client *http.Client, name string, req Request, want string) error {
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, req.URL, nil)
	if err != nil {
		return fmt.Errorf("下载 %s 失败: %w", name, err)
	}
	httpReq.Header.Set("User-Agent", agentOf(req.Agent))
	resp, err := httpClient(client).Do(httpReq)
	if err != nil {
		return fmt.Errorf("下载 %s 失败（%s）: %w", name, req.Via, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
		return statusErr(req.URL, resp.StatusCode)
	}

	part := req.Dest + ".part"
	f, err := os.Create(part)
	if err != nil {
		return fmt.Errorf("建临时文件 %s 失败: %w", part, err)
	}
	defer os.Remove(part) // 成功时目标已改名，这个删除是无操作

	h := sha256.New()
	total := req.Size
	if total <= 0 {
		total = resp.ContentLength
	}
	got := int64(0)
	buf := make([]byte, 256<<10)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				f.Close()
				return fmt.Errorf("写 %s 失败: %w", part, werr)
			}
			h.Write(buf[:n])
			got += int64(n)
			if req.Progress != nil {
				req.Progress(got, total)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			f.Close()
			return fmt.Errorf("下载 %s 中断于 %s: %w", name, HumanBytes(got), rerr)
		}
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return fmt.Errorf("刷 %s 失败: %w", part, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("关 %s 失败: %w", part, err)
	}

	if err := verify(name, part, want, req.Size, got, hex.EncodeToString(h.Sum(nil))); err != nil {
		return err
	}
	if err := os.Rename(part, req.Dest); err != nil {
		return fmt.Errorf("把 %s 改名为 %s 失败: %w", part, req.Dest, err)
	}
	return nil
}

// copyLocal 本机来源（file:// 或路径）：复制 + 同一套校验。
//
// 企业网络里"发布索引放在共享盘"是常态，离线安装也需要这条；它不做任何网络动作，
// 因此也不涉及 https 策略。
func copyLocal(name, src string, req Request, want string) error {
	in, err := os.Open(src)
	if err != nil {
		return &localError{fmt.Errorf("打开本机资产 %s 失败: %w", src, err)}
	}
	defer in.Close()
	part := req.Dest + ".part"
	_ = os.Remove(part)
	out, err := os.Create(part)
	if err != nil {
		return &localError{fmt.Errorf("建临时文件 %s 失败: %w", part, err)}
	}
	defer os.Remove(part)

	h := sha256.New()
	got, err := io.Copy(io.MultiWriter(out, h), in)
	if err != nil {
		out.Close()
		return &localError{fmt.Errorf("复制本机资产 %s 失败: %w", src, err)}
	}
	if err := out.Close(); err != nil {
		return &localError{fmt.Errorf("关 %s 失败: %w", part, err)}
	}
	if err := verify(name, part, want, req.Size, got, hex.EncodeToString(h.Sum(nil))); err != nil {
		return &localError{err}
	}
	if err := os.Rename(part, req.Dest); err != nil {
		return &localError{fmt.Errorf("把 %s 改名为 %s 失败: %w", part, req.Dest, err)}
	}
	return nil
}

// verify 摘要 + 大小两道校验（失败的文件已被调用方的 defer 删掉）。
func verify(name, part, want string, wantSize, got int64, gotHex string) error {
	if gotHex != want {
		return fmt.Errorf("%w: 下载的 %s 校验不通过：发布方声明 sha256 %s，实际 %s（文件已删）",
			errDigest, name, want, gotHex)
	}
	if wantSize > 0 && got != wantSize {
		return fmt.Errorf("%w: 下载的 %s 大小不符：声明 %d 字节，实际 %d 字节（文件已删）",
			errSize, name, wantSize, got)
	}
	return nil
}

// GetBytes 取一个文本响应（发布索引这类小文件），带同样的重试。
//
// maxBytes<=0 用 DefaultMaxBytes：坏响应不该把内存吃光。
func GetBytes(ctx context.Context, client *http.Client, rawurl, agent string, maxBytes int64) ([]byte, error) {
	kind, local, err := resolveSource(rawurl)
	if err != nil {
		return nil, err
	}
	if kind == "file" {
		b, rerr := os.ReadFile(local)
		if rerr != nil {
			return nil, fmt.Errorf("读取本机文件 %s 失败: %w", local, rerr)
		}
		return b, nil
	}
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	var last error
	for attempt := 0; attempt <= DefaultRetries; attempt++ {
		if attempt > 0 {
			if err := SleepCtx(ctx, time.Duration(attempt)*2*time.Second); err != nil {
				return nil, err
			}
		}
		body, retry, gerr := getOnce(ctx, client, rawurl, agent, maxBytes)
		if gerr == nil {
			return body, nil
		}
		last = gerr
		if !retry {
			break
		}
	}
	return nil, last
}

// getOnce 一次 HTTP GET。
func getOnce(ctx context.Context, client *http.Client, rawurl, agent string, maxBytes int64) (body []byte, retry bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", agentOf(agent))
	resp, err := httpClient(client).Do(req)
	if err != nil {
		return nil, true, fmt.Errorf("请求 %s 失败: %w", rawurl, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
		e := statusErr(rawurl, resp.StatusCode)
		return nil, e.retryable(), e
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, true, fmt.Errorf("读 %s 的响应失败: %w", rawurl, err)
	}
	if int64(len(b)) > maxBytes {
		return nil, false, fmt.Errorf("%s 的响应超过 %s 上限（这不是一份合法的索引/清单）", rawurl, HumanBytes(maxBytes))
	}
	return b, false, nil
}

// statusErr 把 HTTP 状态翻成"下一步该做什么"，而不是只报一个数字。
func statusErr(u string, code int) *httpError {
	return &httpError{url: u, code: code, msg: fmt.Sprintf("访问 %s 返回 HTTP %d", u, code)}
}

// RequireHTTPS 校验一个地址必须是 https。
//
// allowInsecure 只给测试用（httptest 的普通 HTTP 服务端）；生产命令一律不放开。
// 本机绝对路径与 file:// 不看这条：它们不经过网络，https 没有意义。
func RequireHTTPS(rawurl string, allowInsecure bool) error {
	kind, _, err := resolveSource(rawurl)
	if err != nil {
		return err
	}
	if kind == "file" {
		return nil
	}
	u, uerr := url.Parse(rawurl)
	if uerr != nil {
		return fmt.Errorf("地址 %q 不合法: %w", rawurl, uerr)
	}
	if u.Host == "" {
		return fmt.Errorf("地址 %q 没有主机名", rawurl)
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return nil
	case "http":
		if allowInsecure {
			return nil
		}
		return fmt.Errorf("地址 %q 是 http：插件安装包必须走 https（明文传输等于把可执行文件交给中间人）", rawurl)
	default:
		if allowInsecure {
			return nil
		}
		return fmt.Errorf("地址 %q 的协议 %q 不受支持（只接受 https 或本机文件）", rawurl, u.Scheme)
	}
}

// resolveSource 把三种写法分开：http(s) URL / file URL / 本机路径。
//
// Windows 盘符（C:\dir\a.json）会被 url.Parse 当成 scheme "c"，所以先按 "://" 判断
// 是不是一个真 URL，再决定当成路径。
func resolveSource(raw string) (kind, path string, err error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", "", errors.New("地址为空")
	}
	if !strings.Contains(s, "://") {
		return "file", s, nil
	}
	u, uerr := url.Parse(s)
	if uerr != nil {
		return "", "", fmt.Errorf("地址 %q 不合法: %w", s, uerr)
	}
	if strings.EqualFold(u.Scheme, "file") {
		p := u.Path
		// Windows 的 file:///C:/x 解析出来是 /C:/x。
		if len(p) >= 3 && p[0] == '/' && p[2] == ':' {
			p = p[1:]
		}
		if p == "" {
			return "", "", fmt.Errorf("file 地址 %q 没有路径", s)
		}
		return "file", filepath.FromSlash(p), nil
	}
	return "http", s, nil
}

// SHA256File 算一个已存在文件的 sha256。
func SHA256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// SleepCtx 可被取消的等待（重试之间的退避）。
func SleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// HumanBytes 人读的大小（进度与失败信息共用一份措辞）。
func HumanBytes(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// normalizeHex 归一化摘要写法（小写；非 64 位十六进制一律视为"没有摘要"）。
func normalizeHex(s string) string {
	s = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s), "sha256:")))
	if len(s) != 64 {
		return ""
	}
	for _, c := range s {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return ""
		}
	}
	return s
}

// httpClient 缺省客户端（有超时，不无限挂）。
func httpClient(c *http.Client) *http.Client {
	if c != nil {
		return c
	}
	return &http.Client{Timeout: 5 * time.Minute}
}

// agentOf 缺省 UA。
func agentOf(s string) string {
	if strings.TrimSpace(s) != "" {
		return s
	}
	return "tt"
}

// nameOf 从 URL 取一个展示名。
func nameOf(rawurl string) string {
	if u, err := url.Parse(rawurl); err == nil && u.Path != "" {
		return filepath.Base(u.Path)
	}
	return filepath.Base(rawurl)
}
