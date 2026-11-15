package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/luyaogod/TT/internal/connsvc"
)

// 形状规则。故意都收紧：这些值会直接变成目录名、命令行命名空间与前端路由。
var (
	// reID 反向 DNS：至少两段、小写、点分隔。
	reID = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?)+$`)
	// reNamespace CLI 命名空间：小写字母开头，字母数字与连字符。
	reNamespace = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)
	// reRoute 前端路由：小写字母数字与连字符，不能以连字符开头。
	reRoute = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)
)

// Problems 一组校验问题。校验收集**全部**问题再一次性报 —— 一次装包只被拒绝一次，
// 不该让人反复"改一个错、再撞下一个"。
type Problems []string

func (p *Problems) addf(format string, a ...any) {
	*p = append(*p, fmt.Sprintf(format, a...))
}

// Err 有问题时返回 *ValidationError，否则 nil。
func (p Problems) Err() error {
	if len(p) == 0 {
		return nil
	}
	return &ValidationError{Problems: p}
}

// ValidationError 校验失败（可能含多条）。
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	if len(e.Problems) == 1 {
		return e.Problems[0]
	}
	return "插件包不合法：" + strings.Join(e.Problems, "；")
}

// ValidateID 校验插件 ID（反向 DNS）。
func ValidateID(id string) error {
	if id == "" {
		return fmt.Errorf("id 不能为空")
	}
	if len(id) > 190 {
		return fmt.Errorf("id 过长（%d 字节，上限 190）", len(id))
	}
	if !reID.MatchString(id) {
		return fmt.Errorf("id %q 不是反向 DNS 形态（小写字母/数字/连字符，至少两段，如 com.example.myplugin）", id)
	}
	return nil
}

// ValidateNamespace 校验 CLI 命名空间。
func ValidateNamespace(ns string) error {
	if ns == "" {
		return fmt.Errorf("cli.namespace 不能为空")
	}
	if !reNamespace.MatchString(ns) {
		return fmt.Errorf("cli.namespace %q 不合法（小写字母开头，只允许小写字母/数字/连字符，最长 32）", ns)
	}
	return nil
}

// ValidateRoute 校验前端路由。
func ValidateRoute(route string) error {
	if route == "" {
		return fmt.Errorf("frontend.route 不能为空")
	}
	if !reRoute.MatchString(route) {
		return fmt.Errorf("frontend.route %q 不合法（小写字母/数字/连字符，最长 32）", route)
	}
	return nil
}

// ValidateSemver 校验插件版本号必须是**严格** SemVer。
//
// 为什么要 stricter 而不是宽松：版本会参与升级比较与回滚判定，`1.0` 这种"看起来能用"
// 的写法在比较时行为不确定，宁可安装时拒绝。
func ValidateSemver(v string) error {
	if v == "" {
		return fmt.Errorf("version 不能为空")
	}
	if _, err := semver.StrictNewVersion(v); err != nil {
		return fmt.Errorf("version %q 不是严格的 SemVer（如 0.1.0；不接受 v 前缀与 1.0 这种省略）", v)
	}
	return nil
}

// ValidateHostAPIRange 校验 hostApi 兼容范围写法。
func ValidateHostAPIRange(rng string) error {
	if strings.TrimSpace(rng) == "" {
		return fmt.Errorf("hostApi 不能为空（例: \">=1.0.0 <2.0.0\"）")
	}
	if _, err := semver.NewConstraint(rng); err != nil {
		return fmt.Errorf("hostApi %q 不是合法的版本范围: %v", rng, err)
	}
	return nil
}

// HostAPISatisfies 判断当前 Host API 版本是否落在插件声明的范围内。
func HostAPISatisfies(rng, current string) (bool, error) {
	c, err := semver.NewConstraint(rng)
	if err != nil {
		return false, fmt.Errorf("hostApi %q 不是合法的版本范围: %v", rng, err)
	}
	v, err := semver.StrictNewVersion(current)
	if err != nil {
		return false, fmt.Errorf("Host API 版本 %q 不是严格 SemVer: %v", current, err)
	}
	if c.Check(v) {
		return true, nil
	}
	return false, fmt.Errorf("插件要求 hostApi %q，当前 TT 的 Host API 是 %s", rng, current)
}

// ValidatePermission 校验能力名。
func ValidatePermission(p string) error {
	if connsvc.ValidCapability(p) {
		return nil
	}
	names := make([]string, 0, len(connsvc.AllCapabilities))
	for _, c := range connsvc.AllCapabilities {
		names = append(names, string(c))
	}
	return fmt.Errorf("未知权限 %q（可用: %s）", p, strings.Join(names, " / "))
}

// ValidateOptions 校验清单所需的上下文。
type ValidateOptions struct {
	// HostAPIVersion 当前 TT 的 Host API 版本（来自 connsvc.HostAPIVersion）。
	HostAPIVersion string
	// ReservedNamespaces 已被占用的 CLI 顶层命令名（含内置命令、旧别名与 TT 保留名）。
	ReservedNamespaces []string
	// OtherNamespaces 其他**已启用**插件占用的命名空间（冲突检查）。
	OtherNamespaces []string
	// OtherRoutes 其他已启用插件占用的前端路由。
	OtherRoutes []string
}

// ValidateManifest 校验清单本身 + 声明的入口/技能目录在 dir 下真实存在。
//
// 只做"包内自洽"的检查：与其他插件的冲突由注册表侧在安装/启用时判（那里有全量视图）。
func ValidateManifest(m *Manifest, dir string, opt ValidateOptions) error {
	if m == nil {
		return &ValidationError{Problems: []string{"清单为空"}}
	}
	var ps Problems

	if m.SchemaVersion != ManifestSchemaVersion {
		ps.addf("schemaVersion = %d，本版 TT 只支持 %d", m.SchemaVersion, ManifestSchemaVersion)
	}
	if err := ValidateID(m.ID); err != nil {
		ps.addf("%v", err)
	}
	if err := ValidateSemver(m.Version); err != nil {
		ps.addf("%v", err)
	}
	if err := ValidateHostAPIRange(m.HostAPI); err != nil {
		ps.addf("%v", err)
	} else if ok, err := HostAPISatisfies(m.HostAPI, opt.HostAPIVersion); err != nil || !ok {
		if err != nil {
			ps.addf("%v", err)
		} else {
			ps.addf("插件要求 hostApi %q，当前 TT 的 Host API 是 %s", m.HostAPI, opt.HostAPIVersion)
		}
	}
	if !m.HasContribution() {
		ps.addf("没有任何贡献（cli/service/frontend/skills 至少要有一类）")
	}

	// 权限名 + 去重。
	seenPerm := map[string]bool{}
	for _, p := range m.Permissions {
		if err := ValidatePermission(p); err != nil {
			ps.addf("%v", err)
			continue
		}
		if seenPerm[p] {
			ps.addf("权限 %q 重复声明", p)
		}
		seenPerm[p] = true
	}

	// CLI 贡献。
	if cli := m.Contributions.CLI; cli != nil {
		if err := ValidateNamespace(cli.Namespace); err != nil {
			ps.addf("%v", err)
		} else {
			if isReserved(cli.Namespace, opt.ReservedNamespaces) {
				ps.addf("cli.namespace %q 与 TT 内置命令/保留名冲突", cli.Namespace)
			}
			if isReserved(cli.Namespace, opt.OtherNamespaces) {
				ps.addf("cli.namespace %q 已被其他已启用插件占用", cli.Namespace)
			}
		}
		checkEntry(&ps, dir, "cli.entry", cli.Entry, m)
	}

	// 常驻服务贡献。
	if svc := m.Contributions.Service; svc != nil {
		checkEntry(&ps, dir, "service.entry", svc.Entry, m)
		if hp := svc.HealthPath; hp != "" && !strings.HasPrefix(ensureLeadingSlash(hp), "/") {
			ps.addf("service.healthPath %q 必须以 / 开头", hp)
		}
	}

	// 前端贡献。
	if fe := m.Contributions.Frontend; fe != nil {
		if err := ValidateRoute(fe.Route); err != nil {
			ps.addf("%v", err)
		} else if isReserved(fe.Route, opt.OtherRoutes) {
			ps.addf("frontend.route %q 已被其他已启用插件占用", fe.Route)
		}
		if strings.TrimSpace(fe.Title) == "" {
			ps.addf("frontend.title 不能为空")
		}
		checkEntry(&ps, dir, "frontend.entry", fe.Entry, m)
	}

	// 技能目录：必须存在且是一份合规技能（有 SKILL.md）。
	if dirs, err := m.SkillDirs(); err != nil {
		ps.addf("%v", err)
	} else {
		for _, rel := range dirs {
			full := filepath.Join(dir, filepath.FromSlash(rel))
			st, serr := os.Stat(full)
			if serr != nil || !st.IsDir() {
				ps.addf("技能目录 %q 在包内不存在", rel)
				continue
			}
			if _, serr := os.Stat(filepath.Join(full, "SKILL.md")); serr != nil {
				ps.addf("技能目录 %q 缺少 SKILL.md", rel)
			}
		}
	}

	return ps.Err()
}

// checkEntry 校验一个入口：可展开、包内相对、真实存在且是普通文件。
func checkEntry(ps *Problems, dir, field, entry string, m *Manifest) {
	rel, err := expandEntry(entry)
	if err != nil {
		ps.addf("%s: %v", field, err)
		return
	}
	full := filepath.Join(dir, filepath.FromSlash(rel))
	st, err := os.Stat(full)
	if err != nil {
		ps.addf("%s %q 在包内不存在（按当前平台 %s/%s 展开后）", field, rel, runtime.GOOS, runtime.GOARCH)
		return
	}
	if !st.Mode().IsRegular() {
		ps.addf("%s %q 不是普通文件", field, rel)
		return
	}
	if st.Size() == 0 {
		ps.addf("%s %q 是空文件", field, rel)
	}
}

// isReserved 判断名字是否在保留清单里（大小写不敏感：clap 的命令名本身是敏感的，
// 但"看起来一样"的名字会在帮助与文档里制造歧义，干脆一律拒绝）。
func isReserved(name string, list []string) bool {
	for _, r := range list {
		if strings.EqualFold(r, name) {
			return true
		}
	}
	return false
}

// SortedUnique 对字符串清单去重并排序（展示与稳定输出用）。
func SortedUnique(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
