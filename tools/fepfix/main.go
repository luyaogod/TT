// 把一份**现场录制**（二进制帧流）剪成仓库里的**夹具**（文本形态：一条消息一行）。
//
// 为什么需要它：`internal/fep` 的不变量回归（corpus_test.go）要有牙 ——
// 语料如果只存在于我这台机器的 %APPDATA%\tt\fep 下，那么换一台机器它就整条静默空转，
// 而"跳过"看起来和"通过"一模一样。所以现场录制必须能被**剪一份**进 testdata/：
// 小、可读、可 diff，而且在任何机器上都会真的跑。
//
// 剪的规则：
//   - 只保留 AUI 帧（type=1）：业务载荷（type=5 那些）与夹具无关；
//   - 丢掉单条超过 --max 的消息（默认 64KB）。丢掉的是**模板全量树**这类巨型消息，
//     留下的是 `an/un/rn` 的增量流 —— 本测试要的不变量（删了就不许活、裸 id 复用可判）
//     全在增量流里，不需要从根长出来的一棵完整树；
//   - 输出**文本形态**（一条一行）：ReadRecording 先试帧解码、失败才退回文本，
//     所以文本夹具走同一条读取路径，而且能 diff、能 review。
//
// 用法：
//
//	go run ./tools/fepfix <录制件.log> <输出.txt> [--max 65536]
//	go run ./tools/fepfix --list            # 看本机语料与各自会剪出多大
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tt/internal/fep"
)

func main() {
	max := flag.Int("max", 64<<10, "单条消息超过这个字节数就丢掉（保不住完整树，但保住增量流）")
	list := flag.Bool("list", false, "列出本机语料目录里各录制件剪出来会有多大")
	flag.Parse()

	if *list {
		dir := recordDir()
		files, err := fep.RecordingFiles(dir, "-rt2gdc.log")
		if err != nil {
			fatal("列语料 %s: %v", dir, err)
		}
		type row struct {
			name       string
			kept, drop int
			bytes      int
		}
		var rows []row
		for _, f := range files {
			msgs, err := fep.ReadRecording(f)
			if err != nil {
				continue
			}
			r := row{name: filepath.Base(f)}
			for _, m := range msgs {
				if len(m) > *max {
					r.drop++
					continue
				}
				r.kept++
				r.bytes += len(m) + 1
			}
			if r.kept > 0 {
				rows = append(rows, r)
			}
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].bytes > rows[j].bytes })
		fmt.Printf("%s（裁剪上限 %d 字节）\n", dir, *max)
		for _, r := range rows {
			fmt.Printf("  %8d 字节  留 %4d 条 / 丢 %3d 条  %s\n", r.bytes, r.kept, r.drop, r.name)
		}
		return
	}

	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "用法: fepfix <录制件.log> <输出.txt> [--max N]")
		os.Exit(2)
	}
	in, out := flag.Arg(0), flag.Arg(1)
	msgs, err := fep.ReadRecording(in)
	if err != nil {
		fatal("读 %s: %v", in, err)
	}
	var b strings.Builder
	kept, drop, bytes := 0, 0, 0
	for _, m := range msgs {
		if len(m) > *max {
			drop++
			continue
		}
		// 文本形态一条一行：消息体里不能有换行（AUI 消息实测是单行）。
		m = strings.ReplaceAll(m, "\r", "")
		m = strings.ReplaceAll(m, "\n", " ")
		b.WriteString(m)
		b.WriteString("\n")
		kept++
		bytes += len(m) + 1
	}
	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		fatal("写 %s: %v", out, err)
	}
	fmt.Printf("%s → %s：留 %d 条 / 丢 %d 条 / %d 字节\n", filepath.Base(in), out, kept, drop, bytes)
}

// recordDir 找本机语料：优先 TT_FEP_RECORD_DIR，其次 %APPDATA%\tt\fep。
func recordDir() string {
	if d := os.Getenv("TT_FEP_RECORD_DIR"); d != "" {
		return d
	}
	return filepath.Join(os.Getenv("APPDATA"), "tt", "fep")
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
