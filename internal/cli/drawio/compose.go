package drawio

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/drawio"
)

func newComposeCmd() *cobra.Command {
	var out string

	cmd := &cobra.Command{
		Use:   "compose <spec.json> [-o <文件>]",
		Short: "把排版规格 spec 展开成 .drawio",
		Long: `把一份「排版规格」展开成可以直接打开的 .drawio。

规格里只写「用哪个控件、放在第几列第几行、写什么字」——查库、算栅格、平移坐标、
重新编号、拼 XML 全由这条命令做。**不要自己写 XML**：库里的控件是「group + 多个部件」
的组合结构（DateEdit 有 10 个单元格），手写必然出错。

  {
    "library": "controls",
    "title": "采购订单",
    "grid": { "originX": 60, "originY": 60, "colGap": 24, "rowGap": 12 },
    "items": [
      { "shape": "ButtonEdit", "col": 0, "row": 0, "text": { "label": "供应商" } },
      { "shape": "DateEdit",   "col": 1, "row": 0, "text": { "label": "下单日期" } }
    ]
  }

可用控件与它们的文字槽位：tt drawio lib --catalog

不带 -o 时把 .drawio 打到 stdout（摘要与列宽报告走 stderr）。`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runCompose(args[0], out)
		},
	}
	cmd.Flags().StringVarP(&out, "out", "o", "", "写到哪个文件（缺省：打到 stdout）")
	return cmd
}

func runCompose(specPath, out string) error {
	if common.JSON && out == "" {
		return fmt.Errorf("--json 要配合 -o：不带 -o 时输出就是 .drawio 本身，没有别的可报")
	}

	b, err := os.ReadFile(specPath)
	if err != nil {
		return inputErr("读不到规格文件 %q：%v", specPath, err)
	}
	spec, err := drawio.ParseSpec(b)
	if err != nil {
		return inputErr("%v", err)
	}

	src, err := drawio.Load()
	if err != nil {
		return inputErr("形状源坏了：%v", err)
	}
	for _, w := range src.Warnings {
		fmt.Fprintf(os.Stderr, "  ! %s\n", w)
	}

	res, err := src.Compose(spec)
	if err != nil {
		var sc *drawio.SelfCheckError
		if errors.As(err, &sc) {
			return selfCheckErr("%s", sc.Msg)
		}
		return inputErr("%v", err)
	}

	summary := fmt.Sprintf("%d 个控件 / %d 个单元格，来自「%s」", res.Count, res.Cells, res.Library)

	if out == "" {
		fmt.Print(res.XML)
		fmt.Fprintf(os.Stderr, "# %s\n# %s\n", summary, res.Report)
		return nil
	}

	target, err := filepath.Abs(out)
	if err != nil {
		return fmt.Errorf("解析目标路径失败 %s: %w", out, err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("建不了目录 %s: %w", filepath.Dir(target), err)
	}
	if err := os.WriteFile(target, []byte(res.XML), 0o644); err != nil {
		return fmt.Errorf("写不了 %s: %w", target, err)
	}

	if common.JSON {
		return common.PrintJSON(map[string]any{
			"ok": true, "out": target, "shapes": res.Count, "cells": res.Cells, "report": res.Report,
		})
	}
	fmt.Printf("已生成 %s —— %s\n%s\n", out, summary, res.Report)
	return nil
}
