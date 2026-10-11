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
)

// FrameTypeAUI is the only message type observed on the wire (payload is text).
const FrameTypeAUI byte = 0x01

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
