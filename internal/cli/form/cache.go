package form

// 表单缓存（契约 C2.3 / C2.4 / C2.5 / Q3）。
//
// 三条口径，都是"照抄既有书面约定"而不是新发明：
//
//  1. **落点 `<数据目录>/forms/`**，且目录名必须走 `config.CacheDir` 校验 ——
//     清单外名字返回空串（B7），所以漏登记的表现不是报错而是"缓存永远不写、永远不读"。`forms` 已登记。
//  2. **只缓存原始 XML**：`Put` 写进去的就是取回来的那串字节，一个文件一份，
//     没有元数据旁车（sidecar）—— 解析每次现做（D5：一个作业的表单解析是毫秒级），
//     缓存里出现"解析结果"就会有一个第二真相。
//  3. **缓存键含 env/module/form + size + mtime**：前三个进路径；后两个用
//     "把缓存文件的 mtime 设成源那边的 mtime"来表达 —— 源实现了 `Identifiable` 时，
//     一次 Stat 就知道表有没有被重新编译（此时大小/时间任一不等即重取）；
//     源报不出身份时（本地目录）回落到 **TTL**（默认 10 分钟）。

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"tt/internal/config"
)

// CacheSubdirForm 是缓存子目录名（必须一字不差地出现在 `config.CacheSubdirs` 里）。
const CacheSubdirForm = "forms"

// DefaultCacheTTL 是源报不出身份时的缓存保鲜期。
//
// 10 分钟这个数与 `internal/config` 的企业目录快照同一量级：表单在一天里不会变，
// 但"改完表单想让 TT 立刻看见"是常见动作，所以不能按天长。
const DefaultCacheTTL = 10 * time.Minute

// Cache 是 `<数据目录>/forms/` 这一棵目录。
type Cache struct {
	Dir string
	TTL time.Duration
}

// NewCacheFromDataDir 由数据目录构造缓存。
//
// 两个错误都必须**响亮**（契约 C2.5 / B7）：数据目录为空、或 `forms` 没登记进清单
// （`config.CacheDir` 返回空串）时直接报错 —— 否则表现是"缓存静默失效"，
// 而没有人会因此收到任何提示。
func NewCacheFromDataDir(dataDir string) (*Cache, error) {
	if strings.TrimSpace(dataDir) == "" {
		return nil, fmt.Errorf("数据目录为空，算不出表单缓存落点（<数据目录>/%s/）", CacheSubdirForm)
	}
	dir := config.CacheDir(dataDir, CacheSubdirForm)
	if dir == "" {
		return nil, fmt.Errorf("缓存子目录 %q 没登记进 config.CacheSubdirs —— 落点会是空路径（契约 C2.3）",
			CacheSubdirForm)
	}
	return &Cache{Dir: dir, TTL: DefaultCacheTTL}, nil
}

// Path 给一份表单的缓存落点；名字不合法或缓存没配时返回空串。
//
// 名字要过白名单：env/module/form 最终会拼进路径，`..` 或分隔符就是路径穿越
// （缓存写到哪里去了会比"没缓存"严重得多）。
//
// 空串是既有契约（调用方靠它"不拼路径"）；**病因**在 PathError 里 ——
// 只返回空串而说不清为什么，现场就回到 D37 那种"缓存永远不写、一声不响"。
func (c *Cache) Path(env, module, form string) string {
	p, err := c.PathError(env, module, form)
	if err != nil {
		return ""
	}
	return p
}

// PathError 同 Path，但拒绝时把**为什么**说出来。
func (c *Cache) PathError(env, module, form string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("缓存没配（Cache 是 nil）—— 算不出落点")
	}
	if c.Dir == "" {
		return "", fmt.Errorf("缓存目录为空 —— 算不出表单缓存落点（<数据目录>/%s/）", CacheSubdirForm)
	}
	for _, seg := range [3]struct{ what, val string }{
		{"环境名", env}, {"模块名", module}, {"表单名", form},
	} {
		if err := c.nameError(seg.val, seg.what); err != nil {
			return "", err
		}
	}
	return filepath.Join(c.Dir, env, module, form+".42f"), nil
}

// nameError 给字段值套上白名单，报错时带上**是哪一段**（env/module/form）。
func (c *Cache) nameError(name, what string) error {
	if err := cacheNameError(name); err != nil {
		return fmt.Errorf("%s不可用（拼进缓存路径的就是它）：%w", what, err)
	}
	return nil
}

// Put 把**原始 XML** 写进缓存。mtime 非零时同时把缓存文件的修改时间设成它
// （源那边的身份，见文件头第 3 条）。
func (c *Cache) Put(env, module, form string, data []byte, mtime time.Time) error {
	p, perr := c.PathError(env, module, form)
	if perr != nil {
		return fmt.Errorf("写缓存失败（env=%q module=%q form=%q）：%w", env, module, form, perr)
	}
	if len(data) == 0 {
		return fmt.Errorf("不缓存空内容（空文件会被解析器判成「不是合法 .42f」，而真正的病因在取件那一步）")
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		return err
	}
	if !mtime.IsZero() {
		_ = os.Chtimes(p, time.Now(), mtime)
	}
	return nil
}

// Lookup 查缓存。src 可为 nil（那就只按 TTL 判新鲜）。
//
// 顺序写死：**先问源身份，再按 TTL 回落** —— 顺序反了不行：源报得出身份的场合，
// 缓存文件的 mtime 就是**源那边的** mtime，用"这个文件多老"判新鲜毫无意义
// （刚缓存下来的表也可能带着去年编译的时间戳）。
func (c *Cache) Lookup(ctx context.Context, src Source, env, module, form string) ([]byte, bool) {
	p := c.Path(env, module, form)
	if p == "" {
		return nil, false
	}
	fi, err := os.Stat(p)
	if err != nil || fi.IsDir() {
		return nil, false
	}
	identified := false
	if id, ok := src.(Identifiable); ok && id != nil {
		size, mtime, ierr := id.Identify(ctx, env, module, form)
		if ierr == nil {
			identified = true
			same := fi.Size() == size && !mtime.IsZero() &&
				fi.ModTime().Truncate(time.Second).Equal(mtime.Truncate(time.Second))
			if !same {
				return nil, false // 源那边变了（表单被重新编译）→ 重取
			}
		}
		// 问不到源那边（断网/权限）：落回 TTL 判
	}
	if !identified && c.TTL > 0 && time.Since(fi.ModTime()) > c.TTL {
		return nil, false
	}
	b, rerr := os.ReadFile(p)
	if rerr != nil {
		return nil, false
	}
	return b, true
}

// 缓存路径一段名字的长度上限。
//
// 为什么是"字"和"字节"两个上限：中文一个字 3 字节，只卡 64 字节的话 21 个汉字就满了
// （原来那条只认 ASCII 的白名单正是 1–64 字节）。但 Windows 单个路径分量是 255 个
// UTF-16 单元，而这里还要再拼一个 ".42f" 后缀，所以字节上限留到 192 —— 两个上限都不
// 允许"靠放宽字符集把路径撑爆"。
const (
	maxCacheNameRunes = 64
	maxCacheNameBytes = 192
)

// winReservedChars 是 Windows 文件名里出现就建不出来的字符。
const winReservedChars = `:*?"<>|`

// cacheNameOK 是拼进缓存路径的三段名字的白名单（台账 D37）。
//
// 规则（**白名单**，不是黑名单）：
//
//   - **放行**：中英文（`unicode.IsLetter`，含中日韩）、数字（`unicode.IsDigit`）、`-` `_` `.`；
//   - **拒绝**：空串、`.` 与 `..` 单独出现、任何位置出现 `..`、路径分隔符 `/` `\`、
//     控制字符（含 `\x00`）、Windows 保留字符 `: * ? " < > |`、白名单外的一切字符
//     （含空格、emoji），以及长度上限。
//
// 放行 CJK 是**必须**的：现场环境名就是 `主机109`，只认 ASCII 会让缓存 90/90 全挂
// 而且一声不响（D37 的原始病因）。
//
// 为什么放宽之后仍然安全：卡住的不是"字符集"，而是**结构**——
//   - 分隔符 `\` `/` 被拒 ⇒ 名字只能是一段，拼不出第二层目录；
//   - `.` `..` 与任何含 `..` 的名字被拒 ⇒ 拼不出"上一层"；
//   - 控制字符与 Windows 保留字符被拒 ⇒ 不给出"名字被系统改写/截断"的机会
//     （`a\0b` 这类在 C 层会变成 `a`，Windows 去掉尾点尾空格同理）；
//   - 剩下被放行的字符集里，没有一个能表达"往哪儿走"。
//
// 所以 `主机109` 通过，而 `../evil` / `a/b` / `a\b` / `..` / `.` / 空 / 含 NUL 全拒。
func cacheNameOK(s string) bool { return cacheNameError(s) == nil }

// cacheNameError 是 cacheNameOK 的"说得出为什么"版本。
//
// 现场排障只看到一行日志时，`名字非法` 帮不上忙；要能看出是"含 .."还是"含控制字符"。
func cacheNameError(s string) error {
	if s == "" {
		return fmt.Errorf("名字为空 —— 拼出来的路径会少一段，落点就跑偏了")
	}
	if s == "." || s == ".." {
		return fmt.Errorf("%q 就是 . 或 .. —— 拼进路径会指向当前目录或上一层", s)
	}
	if !utf8.ValidString(s) {
		return fmt.Errorf("%q 不是合法 UTF-8 —— 落盘时会被系统改名/截断", s)
	}
	if strings.Contains(s, "..") {
		return fmt.Errorf("%q 里含 .. —— 这是路径穿越", s)
	}
	if n := utf8.RuneCountInString(s); n > maxCacheNameRunes {
		return fmt.Errorf("名字有 %d 个字，超过上限 %d", n, maxCacheNameRunes)
	}
	if n := len(s); n > maxCacheNameBytes {
		return fmt.Errorf("名字有 %d 字节，超过上限 %d（还要给 .42f 后缀留位置）", n, maxCacheNameBytes)
	}
	for _, r := range s {
		switch {
		case r == '/' || r == '\\':
			return fmt.Errorf("%q 含路径分隔符 %q —— 名字会被拆成两层目录", s, r)
		case unicode.IsControl(r):
			return fmt.Errorf("%q 含控制字符 %U —— 文件名里不能出现", s, r)
		case strings.ContainsRune(winReservedChars, r):
			return fmt.Errorf("%q 含 Windows 保留字符 %q —— 这个文件名在 Windows 上建不出来", s, r)
		case r == '-' || r == '_' || r == '.', unicode.IsLetter(r), unicode.IsDigit(r):
			// 白名单内
		default:
			return fmt.Errorf("%q 含白名单外的字符 %q —— 只收中英文、数字与 - _ .", s, r)
		}
	}
	return nil
}
