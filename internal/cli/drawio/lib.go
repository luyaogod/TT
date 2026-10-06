package drawio

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/drawio"
)

func newLibCmd() *cobra.Command {
	var (
		out     string
		catalog bool
	)

	cmd := &cobra.Command{
		Use:   "lib",
		Short: "把形状库导出成 drawio 能加载的文件",
		Long: `把内置的 T100 形状库导出成一个 .xml，交给 drawio 加载。

  File → Open Library from → Device…   选导出的文件（桌面版与 app.diagrams.net 都一样）
  之后左侧图形面板里就多出这些控件，拖进画布即可 —— 尺寸与配色都是 T100 设计器的惯例。

导出的默认落点是**当前目录下的 dist/**（在仓库里跑就是仓库根的 dist/，与其它构建
产物同一处，同样不入库）。每次导出都是同一份字节，重跑即可覆盖。

  --catalog           不落盘，把形状清单打到 stdout（给 AI 看的紧凑表格）
  -o, --out <目录>     导出到这个目录（不存在会建）`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runLib(out, catalog)
		},
	}
	cmd.Flags().StringVarP(&out, "out", "o", "", "导出到哪个目录（默认：数据目录下的 drawio 缓存）")
	cmd.Flags().BoolVar(&catalog, "catalog", false, "只打形状清单到 stdout，不落盘")
	return cmd
}

func runLib(out string, catalog bool) error {
	src, err := drawio.Load()
	if err != nil {
		return inputErr("形状源坏了：%v", err)
	}
	for _, w := range src.Warnings {
		fmt.Fprintf(os.Stderr, "  ! %s\n", w)
	}

	// 形状清单：现打永远是最新的 —— 落一份到磁盘就是第二个家，一定会漂
	if catalog {
		fmt.Print(drawio.CatalogMarkdown(src.Catalogs))
		return nil
	}

	dir, err := libraryDir(out)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("建不了目录 %s: %w", dir, err)
	}

	written := make([]string, 0, len(src.Catalogs))
	for _, c := range src.Catalogs {
		payload, err := drawio.MxLibraryOf(c)
		if err != nil {
			return inputErr("%s 合成失败：%v", c.Library.Name, err)
		}
		p := filepath.Join(dir, fmt.Sprintf("%s-v%s.xml", c.Library.Output, c.Library.Version))
		if err := os.WriteFile(p, []byte(payload), 0o644); err != nil {
			return fmt.Errorf("写不了 %s: %w", p, err)
		}
		written = append(written, p)
	}

	if common.JSON {
		return common.PrintJSON(map[string]any{"ok": true, "dir": dir, "files": written})
	}

	fmt.Printf("形状库已导出：%d 个文件\n  目标: %s\n", len(written), dir)
	for _, p := range written {
		fmt.Printf("    %s\n", p)
	}
	fmt.Println("加载：drawio → File → Open Library from → Device… → 选上面任意一个文件")
	fmt.Println("      （库是内置形状源的产物，改形状源要重编 tt）")
	return nil
}
