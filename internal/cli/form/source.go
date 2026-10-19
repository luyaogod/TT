package form

// 取件（契约 C2 / 变更记录 A8）。
//
// **接口归消费方**：`Source` 定义在这里（不是 `internal/form42f`）—— Go 惯例，且让
// `internal/form42f` 保住"零 IO 的纯解析"。两个实现：
//
//	DirSource      本地目录/镜像（离线与测试用，本文件）
//	host.RemoteFormSource  真机按需单取（只读 find + SFTP get，internal/host/formsource.go）
//
// 另外这里给出 **缓存包装**：只缓存原始 XML，落 `<数据目录>/forms/`（见 cache.go）。

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/host"
	"tt/internal/output"
)

// Source 是 `tt form` 需要的取件面（契约 A8：定义在消费方，只 1–2 个方法）。
//
// 约定：
//   - **只读**：实现方不做任何写/删/起作业（H6）；
//   - `form` 不含扩展名（如 `axmt500_wf`）；
//   - 取不到就返回错误，**不要**返回空字节 + nil —— 空内容会被解析器当成
//     "不是合法 .42f"，而真正的病因（权限/路径/网络）就丢了。
type Source interface {
	Form(ctx context.Context, env, module, form string) ([]byte, error)
	List(ctx context.Context, env, module string) ([]string, error)
}

// Identifiable 是**可选**能力：能报出"这一份在源那边是什么"（大小 + 修改时间）。
//
// 有了它，缓存的新鲜度按**源身份**判（契约 C2.4：缓存键含 env/module/form + size + mtime），
// 而不是按本地年龄猜 —— 服务器上重编译过一张表单时，TTL 还没到也不会拿到旧表。
type Identifiable interface {
	Identify(ctx context.Context, env, module, form string) (size int64, mtime time.Time, err error)
}

// ── 本地目录实现 ────────────────────────────────────────────────────────────

// DirSource 从本地目录（镜像/离线夹具/`--dir`）取件。
//
// 候选的**打分**（不是"第一个撞上的"）：路径里有 `42f` 段最重，其次是模块段、环境段。
// 理由：镜像里同一个表单名可能同时躺在 `42f/`（正式产物）与别处（备份/草稿），
// 而"取错了但解析成功"是最不容易被发现的一种错。
//
// 与 A15（"启发式不得作拒绝理由"）的关系：这里参与判定的量是**路径里的事实**
// （有没有 `42f` 目录、有没有该模块/环境的目录名），不是文件名启发式；而且**同分即拒**
// —— 拒绝只由候选数决定（与 A7/A9 同一条原则）。守它的是
// `TestDirSourcePrefersForm42fDir` 与 `TestDirSourceRefusesAmbiguous`。
type DirSource struct{ Root string }

func (d *DirSource) Form(_ context.Context, env, module, form string) ([]byte, error) {
	if strings.TrimSpace(d.Root) == "" {
		return nil, fmt.Errorf("本地目录没给（用 --dir <目录> 或 TT_FORM_DIR）")
	}
	cands, err := d.glob(form + ".42f")
	if err != nil {
		return nil, err
	}
	if len(cands) == 0 {
		return nil, fmt.Errorf("%s 下没有 %s.42f（模块 %s / 环境 %s）", d.Root, form, module, env)
	}
	best := bestByScore(cands, d.Root, env, module)
	if len(best) > 1 {
		return nil, fmt.Errorf("找到 %d 个同样合适的 %s.42f，无法自己挑一个：%s（把 --dir 指到更具体的目录）",
			len(best), form, strings.Join(best, "、"))
	}
	b, err := os.ReadFile(best[0])
	if err != nil {
		return nil, fmt.Errorf("读 %s 失败: %w", best[0], err)
	}
	return b, nil
}

func (d *DirSource) List(_ context.Context, env, module string) ([]string, error) {
	if strings.TrimSpace(d.Root) == "" {
		return nil, fmt.Errorf("本地目录没给（用 --dir <目录> 或 TT_FORM_DIR）")
	}
	cands, err := d.glob("*.42f")
	if err != nil {
		return nil, err
	}
	if len(cands) == 0 {
		return nil, fmt.Errorf("%s 下一个 .42f 都没有", d.Root)
	}
	if module != "" {
		var scoped []string
		for _, p := range cands {
			if hasSegment(p, d.Root, module) {
				scoped = append(scoped, p)
			}
		}
		if len(scoped) == 0 {
			return nil, fmt.Errorf("%s 下没有模块 %q 的目录（找到 %d 个 .42f，都不在该模块下）",
				d.Root, module, len(cands))
		}
		cands = scoped
	}
	seen := map[string]bool{}
	names := make([]string, 0, len(cands))
	for _, p := range cands {
		n := strings.TrimSuffix(filepath.Base(p), ".42f")
		if !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names, nil
}

// glob 找出 Root 下所有 basename 匹配 pattern 的文件（全路径返回）。
// pattern 用 filepath.Match 的语义（`*.42f` 这类通配也认）。
func (d *DirSource) glob(pattern string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(d.Root, func(p string, e os.DirEntry, err error) error {
		if err != nil {
			return nil // 读不到的子树跳过：镜像里常有权限不一致的目录，不该因此整趟失败
		}
		if e.IsDir() {
			return nil
		}
		if ok, merr := filepath.Match(pattern, e.Name()); merr == nil && ok {
			out = append(out, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

// hasSegment 判 p（Root 下的绝对路径）的任一路径段是否等于 seg。
func hasSegment(p, root, seg string) bool {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		rel = p
	}
	for _, s := range strings.Split(filepath.ToSlash(rel), "/") {
		if strings.EqualFold(s, seg) {
			return true
		}
	}
	return false
}

// bestByScore 返回得分最高的那一批候选（同分即"无法自己挑"，调用方据此拒）。
func bestByScore(cands []string, root, env, module string) []string {
	best, bestScore := []string{}, -1
	for _, p := range cands {
		s := 0
		rel, err := filepath.Rel(root, p)
		if err != nil {
			rel = p
		}
		segs := strings.Split(filepath.ToSlash(rel), "/")
		for _, seg := range segs {
			switch {
			case strings.EqualFold(seg, "42f"):
				s += 100
			case module != "" && strings.EqualFold(seg, module):
				s += 50
			case env != "" && strings.EqualFold(seg, env):
				s += 25
			}
		}
		switch {
		case s > bestScore:
			best, bestScore = []string{p}, s
		case s == bestScore:
			best = append(best, p)
		}
	}
	return best
}

// ── 缓存包装 ────────────────────────────────────────────────────────────────

// CachingSource 在取件前面加一层本地缓存（只缓存原始 XML，见 cache.go）。
type CachingSource struct {
	Inner Source
	Cache *Cache
}

func (cs *CachingSource) Form(ctx context.Context, env, module, form string) ([]byte, error) {
	if b, ok := cs.Cache.Lookup(ctx, cs.Inner, env, module, form); ok {
		return b, nil
	}
	b, err := cs.Inner.Form(ctx, env, module, form)
	if err != nil {
		return nil, err
	}
	var mtime time.Time
	if id, ok := cs.Inner.(Identifiable); ok {
		if _, mt, ierr := id.Identify(ctx, env, module, form); ierr == nil {
			mtime = mt
		}
	}
	// 写不进缓存不该挡住取件（缓存是加速手段，不是正确性前提）
	_ = cs.Cache.Put(env, module, form, b, mtime)
	return b, nil
}

// List 不缓存：目录清单随开发变动，且一次 find 本就很便宜。
func (cs *CachingSource) List(ctx context.Context, env, module string) ([]string, error) {
	return cs.Inner.List(ctx, env, module)
}

// Identify 透传给内层（缓存据此判源身份）。
func (cs *CachingSource) Identify(ctx context.Context, env, module, form string) (int64, time.Time, error) {
	if id, ok := cs.Inner.(Identifiable); ok {
		return id.Identify(ctx, env, module, form)
	}
	return 0, time.Time{}, fmt.Errorf("内层实现报不出源身份")
}

// ── 本次调用用哪个取件实现（唯一裁决点）─────────────────────────────────────

// sourceOpts 是一次取件的选项。
type sourceOpts struct {
	Dir     string // 本地目录（非空则不连服务器）
	NoCache bool   // 关缓存
}

// sourceFor 是"本次取件走哪条路"的唯一裁决点。
//
// 测试替换它（假 Source / 本地目录）就等于换掉了取件通道 —— 命令函数不必为了测试
// 长出"依赖注入"参数，生产路径也只有一条。
var sourceFor = productionSource

// productionSource 是生产的裁决：
//   - 有 --dir（或 TT_FORM_DIR）→ 本地目录，不碰网络；
//   - 否则按环境名取 SSH 配置，走**按需单取**（只读）；
//   - 最后按需套上缓存（<数据目录>/forms/）。
func productionSource(env string, o sourceOpts) (Source, error) {
	var inner Source
	if o.Dir != "" {
		inner = &DirSource{Root: o.Dir}
	} else {
		cfgPath, err := common.ResolveConfig(false)
		if err != nil {
			return nil, err
		}
		hosts, err := config.LoadHosts(cfgPath)
		if err != nil {
			return nil, err
		}
		e := hosts.ByName(env)
		if e == nil {
			return nil, (&output.Error{
				Code:    output.CodeEnvUnknown,
				Exit:    exitUsage,
				Message: fmt.Sprintf("环境 %q 不存在", env),
				Hint:    "可用环境：" + strings.Join(hosts.Names(), " / "),
			})
		}
		remote, err := host.NewRemoteFormSource(e)
		if err != nil {
			return nil, err
		}
		inner = remote
	}
	if o.NoCache {
		return inner, nil
	}
	// 缓存落点：<数据目录>/forms/。数据目录解析不出来时不许静默降级 —— 见 cache.go。
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return nil, err
	}
	c, err := NewCacheFromDataDir(config.LocationsAt(cfgPath).DataDir())
	if err != nil {
		return nil, err
	}
	return &CachingSource{Inner: inner, Cache: c}, nil
}

// firstNonEmpty 取第一个非空（空格也算空）。多处用来做"开关优先、环境变量兜底"。
func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return strings.TrimSpace(a)
	}
	return strings.TrimSpace(b)
}
