// Package fep implements the TT side of the FEP (front-end protocol) used by
// the Genero Desktop Client: framing, S-expression parsing, the AUI object
// model, field-name resolution and semantic-action extraction.
//
// The package is deliberately offline: it never opens sockets, never touches
// windows. Everything here is driven by recorded byte streams so it can be
// unit-tested without a live job (see .agents/plan/tt-fep-view-plan.md, M0).
package fep

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"
)

// FrameTypeAUI is the only message type observed on the wire (payload is text).
const FrameTypeAUI byte = 0x01

// FrameTypeCache 是**图片缓存**那一类帧（type=5）。
//
// 实测（2026-10-16 一份真机录制，448 KB）：运行期→前端 35 条 AUI 帧 + **157 条 type=5**；
// 前端→运行期 34 条 AUI + 157 条 type=5。它们**不是消息**（既不解析也不重放，见 H9）：
// 载荷是 `!!__cached__!!` 与图片清单（`logo/dsc_logo.ico?s=1150&t=…`）。
//
// 但它们是**字节**，而且是有时刻的字节：帧索引把它们记成 `kind=cache`，明细页据此画
// "缓存/其它"那条泳道。把它们当"不存在"会让那条泳道没有时间（只剩字节区间），
// 而"什么时候在传图"恰恰是排查"界面卡住"时要看的东西。
const FrameTypeCache byte = 0x05

// MaxFrameLen caps a single payload. AUI screens are large but finite; the cap
// keeps a corrupt stream from allocating unbounded memory.
const MaxFrameLen = 8 << 20

// Frame is one FEP message. On the wire it is
//
//	[BE32 len][BE32 len][type] payload '\n'
//
// The length appears twice (the client uses the duplicate as a sanity check)
// and the trailing newline is part of the frame, not of the payload.
type Frame struct {
	Type    byte
	Payload []byte
}

// Encode renders f as wire bytes.
//
// 长度字段**包含结尾换行**：真机录制件里，GDC 自己发的事件帧声明 len = 文本长 + 1
// （末尾那个 0x0a 计入长度）。少算这一个字节，运行期按声明长度切完就再也对不上下一条
// 消息 —— 表现为**静默丢包**：连接是好的、字节也发出去了、作业一动不动，而且之后连
// 真人的事件也一起被错位吃掉。这是在真机上逐字节对比才捉到的。
func Encode(f Frame) []byte {
	n := len(f.Payload) + 1 // +1：结尾 \n 计入长度（GDC 的约定）
	out := make([]byte, 9+n)
	binary.BigEndian.PutUint32(out[0:4], uint32(n))
	binary.BigEndian.PutUint32(out[4:8], uint32(n))
	out[8] = f.Type
	copy(out[9:], f.Payload)
	out[9+len(f.Payload)] = '\n'
	return out
}

// ErrFrameTooLarge reports a payload longer than MaxFrameLen.
var ErrFrameTooLarge = errors.New("fep: frame too large")

// Decoder turns a byte stream into frames. It tolerates split frames, coalesced
// frames and the binary noise the client interleaves between messages
// (cached-object bytes), which is why it resynchronises by dropping one byte at
// a time instead of trusting a fixed offset.
type Decoder struct {
	buf []byte
}

// NewDecoder returns an empty decoder.
func NewDecoder() *Decoder { return &Decoder{} }

// Feed appends p and returns every frame that is now complete.
func (d *Decoder) Feed(p []byte) ([]Frame, error) {
	d.buf = append(d.buf, p...)
	var out []Frame
	for len(d.buf) >= 9 {
		n := int(binary.BigEndian.Uint32(d.buf[0:4]))
		dup := int(binary.BigEndian.Uint32(d.buf[4:8]))
		// An implausible header means we are looking at payload or noise: drop a
		// byte and try again. A single absurd 32-bit pair can occur inside a
		// payload, so this must not be fatal.
		if n <= 0 || n != dup || n > MaxFrameLen {
			d.buf = d.buf[1:]
			continue
		}
		if len(d.buf) < 9+n {
			break // wait for the rest
		}
		payload := make([]byte, n)
		copy(payload, d.buf[9:9+n])
		// 长度字段包含结尾换行（GDC 的约定），所以载荷末尾那一个 0x0a 是帧的一部分、
		// 不属于消息文本：读出来就切掉。旧录制（长度不含换行）切了也没副作用。
		payload = bytes.TrimRight(payload, "\r\n")
		out = append(out, Frame{Type: d.buf[8], Payload: payload})
		d.buf = d.buf[9+n:]
		// 兼容“长度不含换行”的写法：换行落在帧外时照样吃掉。
		if len(d.buf) > 0 && d.buf[0] == '\n' {
			d.buf = d.buf[1:]
		}
	}
	return out, nil
}

// Buffered reports how many undecoded bytes are held (diagnostics only).
func (d *Decoder) Buffered() int { return len(d.buf) }

// FrameSpan 是一帧**在原始字节流里的位置**：除了帧本身，还带它的起始偏移、占用字节数
// 与"首块到达时刻"。帧索引（index.go）靠它落盘，检查器靠 `Off/Len` 回 raw 里取原字节。
//
// 为什么时刻是"首块"而不是"整帧收齐"：代理是按 `Read` 回来的块打点的，一帧可能跨好几块，
// 而 `time.Now()` 只在块到手那一刻才真实可得。"收齐时刻"要拿最后一块的时间，
// 那会让一帧的时间取决于后面的字节什么时候到 —— 对"用户在什么时候点了这一下"更没用。
// 精度如实标：块级（见 index.go 的 `t`）。
type FrameSpan struct {
	Frame
	Off int64     // 帧首字节在流里的绝对偏移（含 9 字节帧头）
	Len int       // 这一帧占用的字节数（9 字节帧头 + 载荷 + 结尾换行，若在）
	T   time.Time // 帧首块到达的时刻（零值 = 不知道，例如从老录制件现场重建索引）
}

// chunkMark 记"从哪个偏移开始的一段字节是何时到的"。
//
// 存在的理由：帧可以跨块，一条 `Read` 回来的块里可能有前一帧的尾巴。想知道某一帧的
// "首块时刻"，就得记住每一段的起点与时刻 —— 只记"当前块的时间"会把跨块帧的时间
// 记成**后一块**的时间（看起来是"这一帧来晚了"，其实是它早就开始了）。
type chunkMark struct {
	off int64
	t   time.Time
}

// OffsetScanner 与 Decoder 是同一套重同步规则（逐字节丢弃、不信任固定偏移），
// 区别只有一条：它额外报出每一帧在流里的偏移与占用长度。
//
// 它是**增量**的：`Feed` 可以一段一段喂，半帧留在缓冲里等下一次。这很重要 ——
// 代理是按 TCP 块拿到字节的，一帧跨块是常态而不是异常。
type OffsetScanner struct {
	buf   []byte
	base  int64       // buf[0] 在流里的绝对偏移
	marks []chunkMark // 各段的起点与时刻（帧时间取"帧首所在那一段"）
}

// NewOffsetScanner returns an empty scanner.
func NewOffsetScanner() *OffsetScanner { return &OffsetScanner{} }

// abs 返回"下一个将要追加的字节"在流里的绝对偏移。
func (s *OffsetScanner) abs() int64 { return s.base + int64(len(s.buf)) }

// Feed 追加一段字节（`at` = 这段字节到手的时刻）并返回这段字节里**新收齐**的帧。
func (s *OffsetScanner) Feed(p []byte, at time.Time) []FrameSpan {
	if len(p) == 0 {
		return nil
	}
	s.marks = append(s.marks, chunkMark{off: s.abs(), t: at})
	s.buf = append(s.buf, p...)

	var out []FrameSpan
	consumed := 0
	for len(s.buf)-consumed >= 9 {
		rest := s.buf[consumed:]
		n := int(binary.BigEndian.Uint32(rest[0:4]))
		dup := int(binary.BigEndian.Uint32(rest[4:8]))
		// 与 Decoder 同一条判据：不像帧头就丢一个字节再对齐（载荷里能出现看起来
		// 像帧头的 32 位对，所以这里绝不能致命）。
		if n <= 0 || n != dup || n > MaxFrameLen {
			consumed++
			continue
		}
		if len(rest) < 9+n {
			break // 半帧：等下一段
		}
		payload := make([]byte, n)
		copy(payload, rest[9:9+n])
		payload = bytes.TrimRight(payload, "\r\n")
		length := 9 + n
		// 兼容"长度不含换行"的写法：换行落在帧外时也属于这一帧占的字节 ——
		// 不把它算进来的话，`off+len` 会落在换行上，读侧取原字节就少一个。
		if len(rest) > length && rest[length] == '\n' {
			length++
		}
		off := s.base + int64(consumed)
		out = append(out, FrameSpan{
			Frame: Frame{Type: rest[8], Payload: payload},
			Off:   off,
			Len:   length,
			T:     s.timeAt(off),
		})
		consumed += length
	}
	s.base += int64(consumed)
	// 剩下的半帧搬到前面。刻意不用 `s.buf = s.buf[consumed:]`：那只是移动切片起点，
	// 底层数组会随着一条长连接一路长大（一帧 8 MB 上限），而这里只需要剩下的那点。
	if consumed > 0 {
		left := copy(s.buf, s.buf[consumed:])
		s.buf = s.buf[:left]
		s.pruneMarks()
	}
	return out
}

// timeAt 返回偏移 off 所在那一段的到达时刻（没有记录就返回零值）。
func (s *OffsetScanner) timeAt(off int64) time.Time {
	var t time.Time
	for _, m := range s.marks {
		if m.off <= off {
			t = m.t
			continue
		}
		break
	}
	return t
}

// pruneMarks 丢掉已经不可能再被问到的段落时刻。
//
// 判据：所有起点 < base 的段都已经**整段**被消费掉了，只有最后那一段例外 ——
// 剩下那半帧正是从它里面开始的，它的时刻还要留给下一次 Feed。
func (s *OffsetScanner) pruneMarks() {
	keep := 0
	for keep < len(s.marks) && s.marks[keep].off < s.base {
		keep++
	}
	if keep > 1 {
		s.marks = append(s.marks[:0], s.marks[keep-1:]...)
	}
}

// Buffered reports how many undecoded bytes the scanner holds (diagnostics only).
func (s *OffsetScanner) Buffered() int { return len(s.buf) }

// DecodeOffsets 一次性解码一整段字节流，报出每一帧的位置（老录制件现场重建索引用它）。
//
// 与 `dec.Feed(b)` 的关系：同一套判据、同一批帧，只是多了 `Off/Len`。
func DecodeOffsets(b []byte) []FrameSpan {
	return NewOffsetScanner().Feed(b, time.Time{})
}
