package fep

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// timeline.log 的读取（`proxy.timelineLine` 是它的写入侧）。
//
// 它记的是**块级**事实：`HH:MM:SS.mmm >>` / `<<` + 这一块的可 grep 文本（每块截断 1500 字节）。
//
// 为什么还要读它：老录制件没有帧索引，唯一的时刻线索就在这里。明细页据此画"块级精度"的
// 刻度（H5：它**不能**当索引源 —— 没有字节偏移、还被截断，但它确实记得"那几秒有流量"）。
//
// 读者要记住的一件事：一条 timeline 行 ≠ 一帧。一行是 `Read` 回来的一块，里面可能有
// 半帧、几帧、或只是图片缓存的一部分。

// TimelineBlock 是 timeline.log 里的一行。
type TimelineBlock struct {
	// T 是这一块的时刻。文件里只有 `HH:MM:SS.mmm`（没有日期），所以日期从调用方
	// 给的 `day` 补上 —— 录制跨零点时按"时刻倒退就算第二天"处理（见 ReadTimeline）。
	T   time.Time
	Dir int    // DirRuntime（`>>`）/ DirFront（`<<`）
	Raw string // 原始的那段转义文本（照留：排障要能直接 grep）
	// Truncated 表示这一块在文件里就被截断了（>1500 字节的块）。**不许**把它当完整的。
	Truncated bool
	// Bad 表示这一行连时间都读不出来（格式不对）。调用方按"有损"处理，不静默跳过。
	Bad bool
}

// ReadTimeline 解析一个 timeline.log。`day` 提供日期（通常是录制身份的 Started）。
//
// 坏行不报错，而是标 `Bad`：它是诊断文件，一行读不动不该让整个明细页打不开。
func ReadTimeline(path string, day time.Time) ([]TimelineBlock, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	loc := day.Location()
	date := day.Format("2006-01-02")
	var out []TimelineBlock
	var prev time.Time
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64<<10), 4<<20) // 一块最多 1500 字节 + 转义，留够余量
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		sp := strings.IndexByte(line, ' ')
		if sp < 0 {
			out = append(out, TimelineBlock{Bad: true, Raw: line})
			continue
		}
		stamp, dir, rest := line[:sp], line[sp+1:], ""
		if j := strings.IndexByte(dir, ' '); j >= 0 {
			dir, rest = dir[:j], dir[j+1:]
		}
		d := DirRuntime
		switch dir {
		case ">>":
			d = DirRuntime
		case "<<":
			d = DirFront
		default:
			out = append(out, TimelineBlock{Bad: true, Raw: line})
			continue
		}
		t, perr := time.ParseInLocation("2006-01-02 15:04:05.000", date+" "+stamp, loc)
		if perr != nil {
			out = append(out, TimelineBlock{Bad: true, Raw: line})
			continue
		}
		// 跨零点：时刻比上一块早，就是第二天（录制跨零点时才发生）。
		if !prev.IsZero() && t.Before(prev) {
			t = t.Add(24 * time.Hour)
		}
		prev = t
		out = append(out, TimelineBlock{
			T: t, Dir: d, Raw: rest,
			Truncated: strings.HasSuffix(rest, "…（截断）"),
		})
	}
	return out, sc.Err()
}

// TimelineNote 给"块级精度"一句话说明（界面与 CLI 共用同一句，口径不能两处不一样）。
func TimelineNote(blocks int) string {
	return fmt.Sprintf("时间只能到**块级**（%d 块，来自 timeline.log）：每帧的确切时刻无从得知，"+
		"顺序按各方向文件内的字节序", blocks)
}
