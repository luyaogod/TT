package host

// 出厂表单（`.42f`）的**按需单取**（契约 C2.1 / C2.2 的实现侧；A8 把接口定义在消费方）。
//
// 三条硬约束，照 H6「服务器只读」写死：
//
//  1. **只读**：只做 `find` 与 SFTP `get`。绝不写、不删、不起作业 —— 与
//     `internal/host/mirror.go` 的"往服务器上打包再拉下来"完全不同（那条路会写临时归档）。
//  2. **名字先过白名单**再拼进远端命令（`FormNameOK`）：模块名/表单名来自命令行，
//     直接拼进 `find` 就是一条命令注入；拼进 SFTP 路径就是路径穿越。
//  3. **候选不唯一就报错**（`PickFormPath`）：同一个表单名在 TOP 下有两份时，
//     "看起来对了其实取错"是最不容易被发现的一种错。
//
// 连服务器是**惰性**的（第一次 Form/List 才 Dial + 探测 TOP），因此构造它不会碰网络 ——
// 这也是本文件能被单元测试覆盖纯函数部分的前提。

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
)

// RemoteFormSource 从一个 T100 环境按需单取 `.42f`。
//
// 它实现消费方（`internal/cli/form`）的 `Form(ctx, env, module, form) []byte` /
// `List(ctx, env, module) []string` / 可选的 `Identify`（大小 + 时间，供缓存判新鲜）。
type RemoteFormSource struct {
	env *NamedSsh

	mu   sync.Mutex
	conn *SSHConn
	sf   *sftp.Client
	top  string // 探测到的 TOP（T100 无 $TOPDIR，只能登录后动态取，见 tenv.go）
}

// NewRemoteFormSource 绑定一个环境（不连网）。
func NewRemoteFormSource(e *NamedSsh) (*RemoteFormSource, error) {
	if e == nil {
		return nil, fmt.Errorf("host: 环境配置为空")
	}
	if e.Name == "" {
		return nil, fmt.Errorf("host: 环境名为空")
	}
	return &RemoteFormSource{env: e}, nil
}

// Close 关掉 SFTP 与 SSH 连接（重复调用安全）。
func (r *RemoteFormSource) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var err error
	if r.sf != nil {
		err = r.sf.Close()
		r.sf = nil
	}
	if r.conn != nil {
		if e := r.conn.Close(); err == nil {
			err = e
		}
		r.conn = nil
	}
	r.top = ""
	return err
}

// connect 惰性连接并探测 TOP（只探测一次，之后复用）。
func (r *RemoteFormSource) connect() (*SSHConn, *sftp.Client, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn != nil && r.sf != nil && r.top != "" {
		return r.conn, r.sf, r.top, nil
	}
	if r.env.Zone == "" {
		return nil, nil, "", fmt.Errorf("环境 %q 缺少登录区域(zone)，探测不了 T100 目录", r.env.Name)
	}
	conn, err := Dial(r.env.SSHConfig)
	if err != nil {
		return nil, nil, "", fmt.Errorf("连接 %s 失败: %w", r.env.SSHConfig.Addr(), err)
	}
	env, err := ProbeTEnv(conn, r.env.Zone)
	if err != nil {
		conn.Close()
		return nil, nil, "", fmt.Errorf("环境 %s(zone %s)登录探测 T100 目录失败: %w", r.env.Name, r.env.Zone, err)
	}
	sf, err := conn.SFTP()
	if err != nil {
		conn.Close()
		return nil, nil, "", fmt.Errorf("打开 SFTP 失败: %w", err)
	}
	r.conn, r.sf, r.top = conn, sf, strings.TrimSuffix(env.TOP, "/")
	return r.conn, r.sf, r.top, nil
}

// Form 取一张表单的**原始字节**（就是 `.42f` 文件本身，不解析 —— 解析在 internal/form42f）。
func (r *RemoteFormSource) Form(ctx context.Context, _, module, form string) ([]byte, error) {
	conn, sf, top, err := r.connect()
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	expr, err := FormFindExpr(top, module, form)
	if err != nil {
		return nil, err
	}
	out, err := conn.Output(expr, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("在服务器上找 %s.42f 失败: %w（输出：%s）", form, err, strings.TrimSpace(out))
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p, err := PickFormPath(ParseFormFindPaths(out))
	if err != nil {
		return nil, err
	}
	f, err := sf.Open(p)
	if err != nil {
		return nil, fmt.Errorf("打开 %s 失败: %w", p, err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("读 %s 失败: %w", p, err)
	}
	return b, nil
}

// List 列该模块下的表单名（去扩展名、去重、排序）。
func (r *RemoteFormSource) List(ctx context.Context, _, module string) ([]string, error) {
	conn, _, top, err := r.connect()
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	expr, err := FormListExpr(top, module)
	if err != nil {
		return nil, err
	}
	out, err := conn.Output(expr, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("在服务器上列 %s 模块的表单失败: %w（输出：%s）", module, err, strings.TrimSpace(out))
	}
	names := FormNamesFromPaths(ParseFormFindPaths(out))
	if len(names) == 0 {
		return nil, fmt.Errorf("服务器上 TOP=%s 下没有模块 %s 的 *.42f", top, module)
	}
	return names, nil
}

// Identify 报出这一份的大小与修改时间（缓存据此判新鲜；**只 Stat，不下载**）。
func (r *RemoteFormSource) Identify(ctx context.Context, _, module, form string) (int64, time.Time, error) {
	conn, sf, top, err := r.connect()
	if err != nil {
		return 0, time.Time{}, err
	}
	if err := ctx.Err(); err != nil {
		return 0, time.Time{}, err
	}
	expr, err := FormFindExpr(top, module, form)
	if err != nil {
		return 0, time.Time{}, err
	}
	out, err := conn.Output(expr, 30*time.Second)
	if err != nil {
		return 0, time.Time{}, err
	}
	p, err := PickFormPath(ParseFormFindPaths(out))
	if err != nil {
		return 0, time.Time{}, err
	}
	st, err := sf.Stat(p)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("stat %s 失败: %w", p, err)
	}
	return st.Size(), st.ModTime(), nil
}

// ── 纯函数（本文件可被单元测试覆盖的全部部分）────────────────────────────────

// maxFormNameLen 是表单/模块名的长度上限（远端路径名不会比这更长）。
const maxFormNameLen = 64

// FormNameOK 判一个名字能不能拼进远端命令与路径。
//
// 白名单（不是黑名单）：字母/数字/下划线/点/连字符，1–64 字节。
// 这条挡的是**命令注入**（`;` / `$(` / 反引号 / 引号 / 空白）与**路径穿越**（`/`、`\`、`..`）。
func FormNameOK(s string) bool {
	if s == "" || len(s) > maxFormNameLen || strings.Contains(s, "..") {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case c == '_' || c == '.' || c == '-':
		default:
			return false
		}
	}
	return true
}

// FormFindExpr 拼"找某一张某模块下的表单"的远端 find 命令。
//
// 只用 `find`（H6 允许的读操作之一），深度限死，`-type f` 只要文件；
// `2>/dev/null` 让"某个子树没权限"不污染结果 —— 权限不足时我们要的是"没找到"，
// 而不是一段混进路径清单里的错误文本。
func FormFindExpr(top, module, form string) (string, error) {
	if !FormNameOK(module) || !FormNameOK(form) {
		return "", fmt.Errorf("host: 模块名或表单名非法（只允许字母数字 _ . -，且不超过 %d 字节）：module=%q form=%q",
			maxFormNameLen, module, form)
	}
	top = strings.TrimSuffix(strings.TrimSpace(top), "/")
	if top == "" {
		return "", fmt.Errorf("host: TOP 为空，不知道去哪找表单")
	}
	leaf := form + ".42f"
	return "find " + shdq(top) + " -maxdepth 6 -type f " +
		formPathPredicate(module, leaf) + " 2>/dev/null", nil
}

// FormListExpr 拼"列该模块下所有表单"的远端 find 命令。
func FormListExpr(top, module string) (string, error) {
	if !FormNameOK(module) {
		return "", fmt.Errorf("host: 模块名非法（只允许字母数字 _ . -，且不超过 %d 字节）：%q",
			maxFormNameLen, module)
	}
	top = strings.TrimSuffix(strings.TrimSpace(top), "/")
	if top == "" {
		return "", fmt.Errorf("host: TOP 为空，不知道去哪找表单")
	}
	return "find " + shdq(top) + " -maxdepth 6 -type f " +
		formPathPredicate(module, "*.42f") + " 2>/dev/null", nil
}

// formPathPredicate 是"模块目录下的 42f 目录里的某个文件"的两种可能形态。
//
// 两种都要：实测布局里有 `erp/<模块>/<程序>/42f/` 这种，也可能出现 `erp/<模块>/42f/`
// （少一层）。只写一种的后果是"某些模块永远找不到表单"，而那种失败看起来像权限问题。
func formPathPredicate(module, leaf string) string {
	a := `-path "*/` + module + `/*/42f/` + leaf + `"`
	b := `-path "*/` + module + `/42f/` + leaf + `"`
	return `\( ` + a + ` -o ` + b + ` \)`
}

// ParseFormFindPaths 把 find 的输出切成路径清单（去 CR、去空白、去空行、去重、排序）。
func ParseFormFindPaths(out string) []string {
	seen := map[string]bool{}
	var paths []string
	for _, ln := range strings.Split(out, "\n") {
		p := strings.TrimSpace(strings.TrimSuffix(ln, "\r"))
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}

// PickFormPath 在候选里挑唯一的那份；0 个或多个都报错（绝不自己挑一个，D4）。
//
// 多候选的错误里**列出了全部候选**：那时人要做的判断是"哪一份才是这套环境的正式产物"，
// 而这个判断只有他能做。
func PickFormPath(paths []string) (string, error) {
	switch len(paths) {
	case 0:
		return "", fmt.Errorf("服务器上没找到这张表单（.42f）")
	case 1:
		return paths[0], nil
	default:
		return "", fmt.Errorf("找到 %d 份同名表单，无法自己挑一个：%s", len(paths), strings.Join(paths, "、"))
	}
}

// FormNamesFromPaths 从 find 的路径清单里取表单名（去扩展名、去重、排序）。
func FormNamesFromPaths(paths []string) []string {
	seen := map[string]bool{}
	var names []string
	for _, p := range paths {
		b := p
		if i := strings.LastIndexByte(p, '/'); i >= 0 {
			b = p[i+1:]
		}
		b = strings.TrimSuffix(b, ".42f")
		if b == "" || seen[b] {
			continue
		}
		seen[b] = true
		names = append(names, b)
	}
	sort.Strings(names)
	return names
}
