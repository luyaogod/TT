package file

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/cli/common"
	"github.com/luyaogod/TT/internal/host"
	"github.com/luyaogod/TT/internal/output"
)

// ls 一条记录（--json 用）。
type lsJSON struct {
	OK        bool     `json:"ok"`
	Path      string   `json:"path"`
	Entries   any      `json:"entries"` // []lsEntry；--name-only 时是 []string
	TotalRows int      `json:"totalRows"`
	Returned  int      `json:"returned"`
	Truncated bool     `json:"truncated"`
	Notes     []string `json:"notes,omitempty"`
}

type lsEntry struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // file | dir | link | other
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"modTime"`
	Link    string `json:"link,omitempty"`
}

// lsOptions 展示选项：排序 + 截断。默认有界（--limit 200），避免整目录 dump
// 把调用方的上下文一次打满；截断了必须说清楚（totalRows/returned/truncated）。
type lsOptions struct {
	limit    int
	sort     string // name | time | size
	desc     bool
	nameOnly bool
}

const lsDefaultLimit = 200

// applyLsOptions 排序后截断。返回展示项、总数与是否截断；入参不被改动。
func applyLsOptions(entries []host.Entry, opt lsOptions) (shown []host.Entry, total int, truncated bool) {
	sorted := make([]host.Entry, len(entries))
	copy(sorted, entries)
	switch opt.sort {
	case "time":
		sort.SliceStable(sorted, func(i, j int) bool {
			if sorted[i].ModTime.Equal(sorted[j].ModTime) {
				return sorted[i].Name < sorted[j].Name
			}
			return sorted[i].ModTime.Before(sorted[j].ModTime)
		})
	case "size":
		sort.SliceStable(sorted, func(i, j int) bool {
			if sorted[i].Size == sorted[j].Size {
				return sorted[i].Name < sorted[j].Name
			}
			return sorted[i].Size < sorted[j].Size
		})
	default: // name（host.FS.List 已按名字升序，这里只保证 desc 也成立）
		sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	}
	if opt.desc {
		for i, j := 0, len(sorted)-1; i < j; i, j = i+1, j-1 {
			sorted[i], sorted[j] = sorted[j], sorted[i]
		}
	}
	total = len(sorted)
	if opt.limit > 0 && total > opt.limit {
		return sorted[:opt.limit], total, true
	}
	return sorted, total, false
}

// lsTruncNote 截断提示：一句说清"少了多少"和"怎么要看全"。
func lsTruncNote(total, returned int) string {
	return fmt.Sprintf("还有 %d 项未显示（共 %d，已显示 %d）：要全看用 --limit 0；更常见的做法是收窄路径，或用 --sort time 只看最近的",
		total-returned, total, returned)
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
	var opt lsOptions
	c := &cobra.Command{
		Use:   "ls [-la] [--limit N] [--sort name|time|size] [--name-only] [<路径>]",
		Short: "列出目录（默认 $TOP）",
		Long: `列出 $TOP 内的一个目录；不给路径就是 $TOP。

  -l  长格式（权限/大小/时间/名字）
  -a  含隐藏文件（名字以 . 开头）
  --human  大小人性化（配合 -l；-h 被 help 占用）
  --limit N  最多列 N 项（默认 200；0 = 不限）
  --sort K   排序：name（默认）/ time / size
  --desc     反向排序（配合 --sort time 就是"最近的在前"）
  --name-only  只列名字（不带类型/大小/时间）

默认有界：超过 --limit 会截断，并在 totalRows/returned/truncated 与 notes 里
如实报出还有多少项没显示 —— 整目录 dump 会把调用方的上下文一次打满。`,
		Example: `  tt file ls --at 主机109
  tt file ls -la --at 主机109 erp/apm/4gl
  tt file ls --at 主机109 erp/apm/4gl --sort time --desc --limit 5
  tt file ls --at 主机109 erp/apm/4gl --name-only --limit 0`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch opt.sort {
			case "name", "time", "size":
			default:
				return usageErr("--sort 只认 name / time / size，收到 %q", opt.sort)
			}
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
			visible := entries[:0]
			for _, e := range entries {
				if !all && strings.HasPrefix(e.Name, ".") {
					continue
				}
				visible = append(visible, e)
			}
			shown, total, truncated := applyLsOptions(visible, opt)
			var notes []string
			if truncated {
				notes = append(notes, lsTruncNote(total, len(shown)))
			}
			if wantJSON(cmd) {
				out := lsJSON{
					OK: true, Path: display(cur.top, p),
					TotalRows: total, Returned: len(shown), Truncated: truncated, Notes: notes,
				}
				if opt.nameOnly {
					names := make([]string, 0, len(shown))
					for _, e := range shown {
						names = append(names, e.Name)
					}
					out.Entries = names
				} else {
					list := make([]lsEntry, 0, len(shown))
					for _, e := range shown {
						list = append(list, lsEntry{
							Name: e.Name, Type: entryType(e), Size: e.Size,
							Mode: e.ModeText(), ModTime: e.ModTime.Format("2006-01-02 15:04:05"),
							Link: e.Link,
						})
					}
					out.Entries = list
				}
				if err := commonPrintJSON(out); err != nil {
					return err
				}
			} else {
				for _, e := range shown {
					if opt.nameOnly {
						fmt.Fprintln(os.Stdout, e.Name)
						continue
					}
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
						e.ModeText(), size, e.ModTime.Format("2006-01-02 15:04"), name)
				}
			}
			for _, n := range notes {
				fmt.Fprintln(os.Stderr, n)
			}
			return nil
		},
	}
	c.Flags().BoolVarP(&long, "long", "l", false, "长格式（权限/大小/时间/名字）")
	c.Flags().BoolVarP(&all, "all", "a", false, "含隐藏文件")
	c.Flags().IntVar(&opt.limit, "limit", lsDefaultLimit, "最多列 N 项（0 = 不限）")
	c.Flags().StringVar(&opt.sort, "sort", "name", "排序：name / time / size")
	c.Flags().BoolVar(&opt.desc, "desc", false, "反向排序")
	c.Flags().BoolVar(&opt.nameOnly, "name-only", false, "只列名字")
	return c
}

// commonPrintJSON 包一层，避免每个命令都 import output。
func commonPrintJSON(v any) error { return output.WriteJSONValue(os.Stdout, v) }
