package file

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/host"
	"github.com/luyaogod/TT/internal/output"
)

func newGetCmd() *cobra.Command {
	var outPath string
	var maxMB int64
	var force bool
	c := &cobra.Command{
		Use:   "get <路径> [-o <本地|->]",
		Short: "下载一个文件（默认同名落当前目录）",
		Long: `把一个文件从服务器下载到本地（服务器侧仍是只读）。

  -o <本地>   写到指定路径；"-" 表示 stdout
  --max MB    大小上限（默认 200 MB；0 = 不限）
  --force     镜像里已有同一份时也照下（默认拒绝，见下）

先写 .part 再 rename：中断不会留下半截文件。

镜像护栏：没给 -o - 且文件已在本环境离线镜像里、size 与 mtime 都对得上时，
直接拒绝下载并给出镜像路径（镜像保留服务器 mtime，所以这就能判"同一份"，
不必为了算 hash 先下文件）。size 对不上或 mtime 不一致 = 镜像落后，照下并提示。`,
		Example: `  tt file get --at 主机109 erp/apm/4gl/apm_apmt520.4gl
  tt file get --at 主机109 erp/apm/4gl/apm_apmt520.4gl -o x.4gl
  tt file get --at 主机109 per/apm/apmt520.rpt -o -
  tt file get --at 主机109 com/sub/4gl/s_apmt510.4gl --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := cur.argPath(args[0])
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
			if self.IsDir() {
				return usageErr("get 不支持目录：%s（先用 tt file find <目录> 收窄，或 tt file ls <目录> 看文件）", display(cur.top, p))
			}
			limit := maxMB << 20
			if maxMB > 0 && self.Size > limit {
				return usageErr("文件 %s（%s）超过 --max %d MB；确要下载用 --max 0",
					display(cur.top, p), humanBytes(self.Size), maxMB)
			}
			// 镜像护栏：先看本地镜像里是不是已经有同一份，别为一个已有文件重下。
			// -o - 是把字节流交给调用方，不做拦截（否则管道语义会被打断）。
			var notes []string
			if outPath != "-" {
				rel := strings.TrimPrefix(strings.TrimPrefix(p, cur.top), "/")
				if mf, ok := cur.mirrorFileFor(rel); ok {
					st, _ := mirrorCompare(mf, self.Size, self.ModTime)
					switch st {
					case mirrorSame:
						if !force {
							return usageErr("镜像里已有同一份（size %s，mtime %s），未下载：%s",
								humanBytes(self.Size), self.ModTime.Format("2006-01-02 15:04:05"), mf).
								WithHint("确要重下加 --force；只想读内容就直接打开上面这个镜像文件（tt file cat/head 读的是远端）")
						}
						notes = append(notes, "镜像里已有同一份，"+mf+"；本次按 --force 重下")
					case mirrorStale:
						notes = append(notes, "镜像里的同名文件与远端不一致（远端 "+humanBytes(self.Size)+" / "+
							self.ModTime.Format("2006-01-02 15:04:05")+"），本次照下；镜像路径 "+mf)
					}
				}
			}
			f, err := cur.fs.Open(p)
			if err != nil {
				return ioErr("%v", host.WrapPathErr(display(cur.top, p), err))
			}
			defer f.Close()

			if outPath == "-" {
				_, truncated, err := host.CopyN(os.Stdout, f, limit)
				if err != nil {
					return ioErr("读取失败: %v", err)
				}
				if truncated {
					return usageErr("超过 --max %d MB：输出已截断", maxMB)
				}
				// -o - 是“把字节流交给调用方”（管道/重定向），不能再追加 JSON，
				// 否则文件内容后面会多一段信封，二进制直接损坏。
				return nil
			}

			dst := outPath
			if dst == "" {
				dst = path.Base(p)
			}
			abs, err := filepath.Abs(dst)
			if err != nil {
				return output.Errorf(output.CodeLocalIO, 5, "解析本地路径失败: %v", err)
			}
			if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
				return output.Errorf(output.CodeLocalIO, 5, "创建本地目录失败: %v", err)
			}
			tmp, err := os.CreateTemp(filepath.Dir(abs), ".tt-get-*.part")
			if err != nil {
				return output.Errorf(output.CodeLocalIO, 5, "创建本地临时文件失败: %v", err)
			}
			tmpName := tmp.Name()
			n, truncated, copyErr := host.CopyN(tmp, f, limit)
			closeErr := tmp.Close()
			if copyErr != nil || truncated || closeErr != nil {
				os.Remove(tmpName)
				if truncated {
					return usageErr("超过 --max %d MB：下载中止（未落盘）", maxMB)
				}
				if copyErr != nil {
					return ioErr("读取失败: %v", copyErr)
				}
				return output.Errorf(output.CodeLocalIO, 5, "写本地文件失败: %v", closeErr)
			}
			if err := os.Rename(tmpName, abs); err != nil {
				os.Remove(tmpName)
				return output.Errorf(output.CodeLocalIO, 5, "保存失败: %v", err)
			}
			if wantJSON(cmd) {
				out := map[string]any{"ok": true, "path": display(cur.top, p), "local": abs, "bytes": n}
				if len(notes) > 0 {
					out["notes"] = notes
				}
				return commonPrintJSON(out)
			}
			fmt.Fprintf(os.Stderr, "已保存 %s（%s）\n", abs, humanBytes(n))
			for _, nt := range notes {
				fmt.Fprintln(os.Stderr, nt)
			}
			return nil
		},
	}
	c.Flags().StringVarP(&outPath, "output", "o", "", `本地路径（"-" = stdout）`)
	c.Flags().Int64Var(&maxMB, "max", 200, "大小上限（MB；0 = 不限）")
	c.Flags().BoolVar(&force, "force", false, "镜像里已有同一份时也照下")
	return c
}
