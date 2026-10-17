package fep

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

// tt recording mark-stale —— 旧件的两条路之一（契约 §7 A24 / 用户决策 D-F2）。
//
// 背景：回放**强制身份**（A23）之后，早于身份上线的旧件、以及 v2 指纹的旧件，
// 真跑会被拒。用户定的处置只有两条：**重录**（推荐）、或**打标记**。
//
// 这条命令只做后者的"那一半"：给身份文件写上 `stale: true` + 原因 + 时间。
// 它**绝不删文件、绝不改录制内容** —— 录制是不可再生的（用户再报一次 bug、再操作一遍
// 才有这一份），所以"别拿它回放"这句话必须能说出口，又不必把证据毁掉。
//
// 为什么不给 `--skip-gate` / `--force` 之类的跳过开关：用户明令不加（A24 第一句）。
// 打标记不改变"回放会被拒"这个事实，它改变的是**看得见**：列表与明细都会写
// 「不可回放（需重录）」，于是没人会以为"能展示就能跑"。

var recStaleReason string

var recStaleCmd = &cobra.Command{
	Use:   "mark-stale <录制 ID>",
	Short: "把一份录制标成「" + fep.StaleNote + "」（只改身份文件，一个录制文件都不删）",
	Long: `把一份录制标成「` + fep.StaleNote + `」。

它只往身份文件（` + "`<ID>.meta.json`" + `）里写三样：stale=true、原因、时间。
录制件本身（-rt2gdc / -gdc2rt / -timeline / 帧索引）**一个字节都不动** ——
录制不可再生，标记是为了让人别拿它回放，不是为了把证据毁掉。

没有身份文件的老件也能标记：会补写一份只有 ID 与标记的身份文件（备注里写明是补写的）。

标记之后：` + "`tt recording ls`" + ` 与 ` + "`tt recording show`" + ` 都会显示「` + fep.StaleNote + `」。
回放仍然会被拒（A23）—— 这里**不加**任何跳过闸门的开关。要真能跑，只有重录：
  tt recording start <身份>`,
	Args: exactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := resolveRecording(uiFrom, args[0])
		if err != nil {
			return err
		}
		at := time.Now()
		m, merr := fep.MarkStale(t.Dir, t.ID, recStaleReason, at)
		if merr != nil {
			return fail(exitFailed, "标记 %s 失败：%v", t.ID, merr)
		}
		return printJSONOrText(map[string]any{
			"id": m.ID, "dir": t.Dir, "stale": m.Stale, "reason": m.StaleReason,
			"staleAt": m.StaleAt.Format(time.RFC3339), "staleNote": fep.StaleNote,
			"metaPath": fep.MetaPath(t.Dir, t.ID), "files": len(t.Files),
		}, func() {
			fmt.Printf("已标记 %s → %s\n", m.ID, fep.StaleNote)
			if m.StaleReason != "" {
				fmt.Printf("  原因：%s\n", m.StaleReason)
			}
			fmt.Printf("  时间：%s\n", m.StaleAt.Format(time.RFC3339))
			fmt.Printf("  只改了身份文件：%s\n", fep.MetaPath(t.Dir, t.ID))
			fmt.Printf("  录制文件 %d 份一个都没动（标记只是把后果写在脸上，回放仍会被拒）\n", len(t.Files))
			fmt.Println("  要真能回放：重录一份（tt recording start）；这份留着对照/取证")
		})
	},
}

func init() {
	recCmd.AddCommand(recStaleCmd)
	recStaleCmd.Flags().StringVar(&recStaleReason, "reason", "", "为什么标记它（一句话，写进身份文件）")
}
