package file

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/host"
	"tt/internal/output"
)

// ls 一条记录（--json 用）。
//
// total / returned / truncated / filter 是本轮新增的**有界输出**元数据：
// 从前 `tt file ls com/lib/42m` 一次回 200 来条、30KB JSON，调用方要的往往只是
// "这个目录里有没有 wssp00120_wf 的编译产物" —— 筛选必须能在命令内部做完，
// 并且**截断不许静默**（要给 truncated 与可复现的下一步）。
// 旧键（ok/path/entries）一个没改名，既有脚本行为不变。
type lsJSON struct {
	OK        bool      `json:"ok"`
	Path      string    `json:"path"`
	Total     int       `json:"total"`
	Returned  int       `json:"returned"`
	Truncated int       `json:"truncated,omitempty"`
	Filter    string    `json:"filter,omitempty"`
	Entries   []lsEntry `json:"entries"`
	Hint      string    `json:"hint,omitempty"`
}

// filterEntries 按 glob 筛选；pat 为空 = 不筛。
//
// 取 glob（* / ?）与 `tt debug wslogs --job wssp*` 同一口径；**不区分大小写**，
// 因为服务器上真实存在 `WSSP00120_WF.42M` 这类混排名字，而调用方通常只想
// "有没有这个名字"。
func filterEntries(entries []lsEntry, pat string) []lsEntry {
	pat = strings.TrimSpace(pat)
	if pat == "" {
		return entries
	}
	lower := strings.ToLower(pat)
	out := make([]lsEntry, 0, len(entries))
	for _, e := range entries {
		if ok, _ := path.Match(lower, strings.ToLower(e.Name)); ok {
			out = append(out, e)
		}
	}
	return out
}

// limitEntries 截断到 limit 条；limit <= 0 = 不限。返回被截掉的条数（不许静默）。
func limitEntries(entries []lsEntry, limit int) ([]lsEntry, int) {
	if limit <= 0 || len(entries) <= limit {
		return entries, 0
	}
	return entries[:limit], len(entries) - limit
}

type lsEntry struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // file | dir | link | other
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"modTime"`
	Link    string `json:"link,omitempty"`
}

func entryType(e host.Entry) string {
	switch {
	case e.IsSymlink():
		return "link"
	case e.IsDir():
		return "dir"
	case e.Mode.IsRegular():
		return "file"
	default:
		return "other"
	}
}

func newLsCmd() *cobra.Command {
	var long, all bool
	var filter string
	var limit int
	c := &cobra.Command{
		Use:   "ls [-la] [--filter <glob>] [--limit n] [<路径>]",
		Short: "列出目录（默认 $TOP；支持按名字筛选与限制条数）",
		Long: `列出 $TOP 内的一个目录；不给路径就是 $TOP。

  -l  长格式（权限/大小/时间/名字）
  -a  含隐藏文件（名字以 . 开头）
  --human  大小人性化（配合 -l；-h 被 help 占用）

  --filter <glob>  只留名字匹配的条目（* 与 ? 通配，不区分大小写）。
  --limit <n>      最多返回 n 条（默认 0 = 不限）；截断时 truncated > 0 并给 hint。

JSON 形状（默认）：
  {"ok":true,"path":"com/cwss/42m","total":98,"returned":3,"truncated":0,
   "filter":"*wssp00120*","entries":[{"name":"cwss_wssp00120_wf.42m","type":"file",
   "size":129169,"mode":"-rw-rw-r--","modTime":"2026-07-02 17:35:41"}]}

退出码：0 成功（筛不到 = 空 entries）；1 路径不存在（NOT_FOUND）；4 读取失败。`,
		Example: `  tt file ls --at 主机109
  tt file ls -la --at 主机109 erp/apm/4gl
  tt file ls --at 主机109 com/cwss/42m --filter '*wssp00120*'
  tt file ls --at 主机109 com/lib/42m --filter '*.42m' --limit 20`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := ""
			if len(args) == 1 {
				arg = args[0]
			}
			p, err := cur.argPath(arg)
			if err != nil {
				return err
			}
			p, err = cur.resolveSafe(p)
			if err != nil {
				return err
			}
			self, err := cur.fs.Lstat(p)
			if err != nil {
				if host.NotFound(err) {
					return notFoundErr(cur.top, p)
				}
				return ioErr("%v", host.WrapPathErr(display(cur.top, p), err))
			}
			entries := []host.Entry{*self}
			if self.IsDir() {
				entries, err = cur.fs.List(p)
				if err != nil {
					return ioErr("%v", host.WrapPathErr(display(cur.top, p), err))
				}
			}
			shown := entries[:0]
			for _, e := range entries {
				if !all && strings.HasPrefix(e.Name, ".") {
					continue
				}
				shown = append(shown, e)
			}
			// 先转成对外形状，再在它上面筛选/截断：筛选口径与 JSON 里的名字永远是同一个。
			rows := make([]lsEntry, 0, len(shown))
			for _, e := range shown {
				rows = append(rows, lsEntry{
					Name: e.Name, Type: entryType(e), Size: e.Size,
					Mode: e.ModeText(), ModTime: e.ModTime.Format("2006-01-02 15:04:05"),
					Link: e.Link,
				})
			}
			total := len(rows)
			rows = filterEntries(rows, filter)
			matched := len(rows)
			rows, truncated := limitEntries(rows, limit)

			if wantJSON(cmd) {
				out := lsJSON{OK: true, Path: display(cur.top, p), Total: total,
					Returned: len(rows), Truncated: truncated, Filter: filter, Entries: rows}
				if truncated > 0 {
					out.Hint = fmt.Sprintf("还有 %d 条没返回；缩小办法：tt file ls %s --filter '<glob>' --limit %d",
						truncated, display(cur.top, p), total)
				}
				if total > 0 && matched == 0 {
					out.Hint = fmt.Sprintf("目录里有 %d 条，但没有名字匹配 %q 的；去掉 --filter 看全量", total, filter)
				}
				return commonPrintJSON(out)
			}
			for _, e := range rows {
				name := e.Name
				if e.Link != "" {
					name += " -> " + e.Link
				}
				if !long {
					fmt.Fprintln(os.Stdout, name)
					continue
				}
				size := fmt.Sprintf("%d", e.Size)
				if common.Human {
					size = humanBytes(e.Size)
				}
				fmt.Fprintf(os.Stdout, "%s %8s %s  %s\n",
					e.Mode, size, e.ModTime[:16], name)
			}
			if truncated > 0 {
				fmt.Fprintf(os.Stderr, "（还有 %d 条没返回；用 --filter 缩小或用 --limit 调大）\n", truncated)
			}
			if total > 0 && matched == 0 {
				fmt.Fprintf(os.Stderr, "（目录里有 %d 条，但没有名字匹配 %q 的）\n", total, filter)
			}
			return nil
		},
	}
	c.Flags().BoolVarP(&long, "long", "l", false, "长格式（权限/大小/时间/名字）")
	c.Flags().BoolVarP(&all, "all", "a", false, "含隐藏文件")
	c.Flags().StringVar(&filter, "filter", "", "只留名字匹配该 glob 的条目（* 与 ? 通配，不区分大小写）")
	c.Flags().IntVar(&limit, "limit", 0, "最多返回多少条（0 = 不限；截断时给 truncated 与提示）")
	return c
}

// commonPrintJSON 包一层，避免每个命令都 import output。
func commonPrintJSON(v any) error { return output.WriteJSONValue(os.Stdout, v) }
