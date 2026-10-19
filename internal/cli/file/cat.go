package file

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"tt/internal/host"
)

func newCatCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cat <文件...>",
		Short: "输出文件内容（多文件直接拼接）",
		Long: `输出一个或多个文件的内容。多个文件按参数顺序直接拼接（Unix cat 语义，无分隔头）。
参数支持最后一段通配（如 erp/apm/4gl/*.4gl），由本工具展开。

含 NUL 字节的文件默认拒绝输出（防止把终端/上下文打烂）；确要输出加 --binary。`,
		Example: `  tt file cat --at 主机109 erp/apm/4gl/apmi004.4gl
  tt file cat --at 主机109 'erp/apm/4gl/apmi004*.4gl'`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			for _, arg := range args {
				matches, err := cur.expandArg(arg)
				if err != nil {
					return err
				}
				for _, one := range matches {
					if err := catOne(one); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
}

func catOne(arg string) error {
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
	if self.IsDir() {
		return usageErr("cat 不支持目录：%s", display(cur.top, p))
	}
	f, _, err := openChecked(p)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(os.Stdout, f)
	f.Close()
	if copyErr != nil {
		return ioErr("%v", host.WrapPathErr(display(cur.top, p), copyErr))
	}
	return nil
}

func newHeadCmd() *cobra.Command {
	var n int
	c := &cobra.Command{
		Use:     "head [-n N] <文件...>",
		Short:   "前 N 行（默认 10）",
		Long:    `输出每个文件的前 N 行；多个文件时按 GNU 的 "==> 文件 <==" 分段。`,
		Example: `  tt file head -n 40 --at 主机109 erp/apm/4gl/apmi004.4gl`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runText(cmd, args, false, n)
		},
	}
	c.Flags().IntVarP(&n, "lines", "n", 10, "输出前 N 行")
	return c
}

func newTailCmd() *cobra.Command {
	var n int
	c := &cobra.Command{
		Use:     "tail [-n N] <文件...>",
		Short:   "后 N 行（默认 10）",
		Long:    `输出每个文件的后 N 行；只读文件尾部若干块，不整份下载。`,
		Example: `  tt file tail -n 40 --at 主机109 erp/apm/4gl/apmi004.4gl`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runText(cmd, args, true, n)
		},
	}
	c.Flags().IntVarP(&n, "lines", "n", 10, "输出后 N 行")
	return c
}
