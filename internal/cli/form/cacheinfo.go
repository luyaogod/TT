package form

// `tt form cache`：表单缓存的**只读概览**。
//
// 它回答三个问题：缓存里有多少张表单、共多少字节、最旧/最新是哪一份（什么时候）。
//
// **绝不删任何东西**：清理是统一框架的事（`tt cache clear` 走 `config.CleanCache`，
// 清单在 `config.CacheSubdirs` 里 —— `forms` 已登记）。在这里自己造一套删除，
// 就等于给"缓存"多出第二个清理口径，而两套口径迟早会有一套把不该删的删掉。
// 本文件里没有 `os.Remove` / `os.RemoveAll`，也不 `MkdirAll`（概览连目录都不建）。

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "表单缓存概览：多少张、多少字节、最旧/最新（只读，不删任何东西）",
	Long: `看一眼表单缓存（<数据目录>/forms/）：多少张表单、共多少字节、最旧/最新是哪一份。

  tt form cache [--json]

缓存里一个表单一个 .42f（**原始 XML**，契约 C2.4），由 tt form fetch / show 写入。
清理不在这里：走统一的 tt cache clear（清单在 config.CacheSubdirs，forms 已登记）。

退出码：0 成功（还没有缓存也算成功：0 张 / 0 字节）。`,
	Args: noArgs,
	RunE: runCacheInfo,
}

func init() { Group.AddCommand(cacheCmd) }

// cacheEntry 是缓存里的一份表单（相对路径 + 字节 + 修改时间）。
type cacheEntry struct {
	Env      string `json:"env,omitempty"`
	Module   string `json:"module,omitempty"`
	Form     string `json:"form"`
	Path     string `json:"path"`     // 相对缓存目录（env/module/form.42f）
	Bytes    int64  `json:"bytes"`    // 文件大小
	Modified string `json:"modified"` // RFC3339（缓存文件的 mtime）
}

// cacheReport 是 `--json` 的信封（新增键；C5.2 只冻结 show 的信封）。
type cacheReport struct {
	Dir     string       `json:"dir"`
	Exists  bool         `json:"exists"`
	Forms   int          `json:"forms"` // .42f 张数
	Files   int          `json:"files"` // 目录下所有文件数（正常情况下等于 Forms）
	Bytes   int64        `json:"bytes"` // 所有文件的字节合计
	Oldest  string       `json:"oldest,omitempty"`
	Newest  string       `json:"newest,omitempty"`
	Entries []cacheEntry `json:"entries"`
}

func runCacheInfo(cmd *cobra.Command, _ []string) error {
	cache, err := formCache()
	if err != nil {
		// 算不出落点 = 缓存这条路整条不可用（数据目录解析不出来，或清单漏登记）。
		// 按本组码表归 1（FORM_UNKNOWN），消息与提示把病因说清 —— 不静默退 0。
		return formUnknown("用 --config 指向配置文件；并确认 config.CacheSubdirs 里有 \""+
			CacheSubdirForm+"\"（清单外名字拿空串，见 C2.3）", "算不出表单缓存落点：%v", err)
	}
	rep := scanFormCache(cache.Dir)

	if wantJSON(cmd) {
		return common.PrintJSON(rep)
	}
	printCacheTable(rep)
	return nil
}

// scanFormCache 扫一遍缓存目录（只 stat + 读目录，不写不删）。
//
// 目录不存在 → `Exists=false` 的零值报告，**不是错误**：还没预热过是完全正常的状态，
// 把它当失败会让"第一次用"看起来像坏了。
func scanFormCache(dir string) cacheReport {
	rep := cacheReport{Dir: dir, Entries: []cacheEntry{}}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return rep
	}
	rep.Exists = true

	var oldest, newest time.Time
	// 读不到的子树跳过（与 config.dirUsage 同一条口径）：概览不该因为一个权限怪目录整趟失败。
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, werr error) error {
		if werr != nil || d.IsDir() {
			return nil
		}
		fi, ierr := d.Info()
		if ierr != nil {
			return nil
		}
		rep.Files++
		rep.Bytes += fi.Size()
		if oldest.IsZero() || fi.ModTime().Before(oldest) {
			oldest = fi.ModTime()
		}
		if newest.IsZero() || fi.ModTime().After(newest) {
			newest = fi.ModTime()
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".42f") {
			return nil
		}
		rel, rerr := filepath.Rel(dir, p)
		if rerr != nil {
			rel = d.Name()
		}
		rel = filepath.ToSlash(rel)
		e := cacheEntry{
			Form:     strings.TrimSuffix(d.Name(), ".42f"),
			Path:     rel,
			Bytes:    fi.Size(),
			Modified: fi.ModTime().Format(time.RFC3339),
		}
		if segs := strings.Split(rel, "/"); len(segs) >= 3 {
			e.Env, e.Module = segs[0], segs[1]
		}
		rep.Forms++
		rep.Entries = append(rep.Entries, e)
		return nil
	})

	if !oldest.IsZero() {
		rep.Oldest = oldest.Format(time.RFC3339)
	}
	if !newest.IsZero() {
		rep.Newest = newest.Format(time.RFC3339)
	}
	// 排序后输出：map/WalkDir 的顺序不可依赖，而这份清单会被人逐行比对（要可 diff）。
	sort.Slice(rep.Entries, func(i, j int) bool { return rep.Entries[i].Path < rep.Entries[j].Path })
	return rep
}

// printCacheTable 给人读：一行位置，一行三个真值，最旧/最新各一行，再逐份列出。
func printCacheTable(r cacheReport) {
	writeOut("表单缓存 %s\n", r.Dir)
	if !r.Exists {
		writeOut("  （还没有这份缓存：先跑 tt form fetch --env <环境> --module <模块> 预热）\n")
		return
	}
	writeOut("  表单 %d 张 · 文件 %d 个 · 共 %d 字节\n", r.Forms, r.Files, r.Bytes)
	if r.Oldest != "" {
		writeOut("  最旧 %s · 最新 %s\n", r.Oldest, r.Newest)
	}
	for _, e := range r.Entries {
		writeOut("  %-40s %8d B  %s\n", e.Path, e.Bytes, e.Modified)
	}
	writeOut("（只读概览：这里不删任何东西；清理走 tt cache clear）\n")
}
