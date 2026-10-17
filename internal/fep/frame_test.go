package fep

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	payload := []byte(`an 0 Window 5 {{name "w"}}`)
	wire := Encode(Frame{Type: FrameTypeAUI, Payload: payload})

	dec := NewDecoder()
	frames, err := dec.Feed(wire)
	if err != nil {
		t.Fatalf("Feed: %v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("got %d frames, want 1", len(frames))
	}
	if frames[0].Type != FrameTypeAUI || !bytes.Equal(frames[0].Payload, payload) {
		t.Fatalf("round trip mismatch: %+v", frames[0])
	}
	if dec.Buffered() != 0 {
		t.Fatalf("decoder still holds %d bytes", dec.Buffered())
	}
}

func TestDecoderHandlesSplitAndCoalescedFrames(t *testing.T) {
	a := Encode(Frame{Type: FrameTypeAUI, Payload: []byte("one")})
	b := Encode(Frame{Type: FrameTypeAUI, Payload: []byte("two")})
	stream := append(append([]byte{}, a...), b...)

	// Byte-at-a-time: every frame must still come out exactly once.
	dec := NewDecoder()
	var got []string
	for i := 0; i < len(stream); i++ {
		frames, err := dec.Feed(stream[i : i+1])
		if err != nil {
			t.Fatalf("Feed at %d: %v", i, err)
		}
		for _, f := range frames {
			got = append(got, string(f.Payload))
		}
	}
	if len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("split decode = %v", got)
	}

	// All at once.
	frames, err := NewDecoder().Feed(stream)
	if err != nil || len(frames) != 2 {
		t.Fatalf("coalesced decode: %v %v", frames, err)
	}
}

// The client interleaves binary (cached-object) bytes between messages; the
// decoder must resynchronise instead of trusting a fixed offset.
func TestDecoderResynchronises(t *testing.T) {
	noise := []byte{0x00, 0x00, 0x00, 0x02, 0xaa, 0xbb}
	wire := append(append([]byte{}, noise...), Encode(Frame{Type: FrameTypeAUI, Payload: []byte("ok")})...)

	frames, err := NewDecoder().Feed(wire)
	if err != nil {
		t.Fatalf("Feed: %v", err)
	}
	if len(frames) != 1 || string(frames[0].Payload) != "ok" {
		t.Fatalf("resync failed: %+v", frames)
	}
}

func TestDecoderSkipsOversizedHeader(t *testing.T) {
	// A 32-bit pair that declares a huge frame can appear inside a payload; the
	// decoder must resynchronise instead of failing or allocating.
	wire := make([]byte, 9)
	wire[0], wire[1], wire[2], wire[3] = 0x7f, 0xff, 0xff, 0xff
	wire[4], wire[5], wire[6], wire[7] = 0x7f, 0xff, 0xff, 0xff
	frames, err := NewDecoder().Feed(wire)
	if err != nil {
		t.Fatalf("Feed: %v", err)
	}
	if len(frames) != 0 {
		t.Fatalf("expected no frames, got %d", len(frames))
	}
}

// wireFrame 按线格式手工拼一帧：`[BE32 len][BE32 len][type] payload '\n'`
// （长度字段**包含**结尾换行 —— GDC 自己的约定）。
//
// 刻意不用 Encode：Encode 的契约是"渲染我认识的帧"，而这里要造的正是**未知类型**帧。
// 拿 Encode 造它等于让被测对象替测试准备输入。
func wireFrame(typ byte, payload []byte) []byte {
	n := len(payload) + 1
	out := make([]byte, 9+n)
	binary.BigEndian.PutUint32(out[0:4], uint32(n))
	binary.BigEndian.PutUint32(out[4:8], uint32(n))
	out[8] = typ
	copy(out[9:], payload)
	out[9+len(payload)] = '\n'
	return out
}

// TestUnknownFrameTypeIsDiagnosed 守 C6.3（契约 §1 C6.3 / B12）。
//
// 上游证据（GDC-逆向报告 §16① 与 §8.1）：帧头第 3 字节（type）的取值域**未穷尽** ——
// 实测只见过 `1`（AUI）和 `5`（图片缓存）。既然没穷尽，就不能假设"别的值不存在"：
//
//   - 遇到未知 type **不丢帧**（丢帧会让后面的事件全部错位，而且症状是"界面一动不动"，
//     没人能查到是这里吞的）；
//   - 但**也不能当 AUI 消息合并**（H9/B12：载荷不是消息，硬解等于往模型里灌垃圾）；
//   - 要留下**可查询的诊断**：计数 + 最近几个未知类型 + 它们在流里的字节偏移，
//     供调用方报"这条连接上有我没见过的帧类型"。
func TestUnknownFrameTypeIsDiagnosed(t *testing.T) {
	const unknown byte = 0x07
	auiBody := []byte(`om 1 {{an 0 Window 5 {{name "w"}}}}`)
	mysteryBody := []byte(`om 2 {{an 0 Window 6 {{name "w2"}}}}`)

	first := Encode(Frame{Type: FrameTypeAUI, Payload: auiBody})
	mystery := wireFrame(unknown, mysteryBody)
	stream := append(append([]byte{}, first...), mystery...)

	dec := NewDecoder()
	frames, err := dec.Feed(stream)
	if err != nil {
		t.Fatalf("Feed: %v", err)
	}
	// 未知类型**不得被丢帧**：两帧都得交出来，且类型原样。
	if len(frames) != 2 {
		t.Fatalf("未知类型被丢帧了：拿到 %d 帧，期望 2", len(frames))
	}
	if frames[0].Type != FrameTypeAUI || frames[1].Type != unknown {
		t.Fatalf("帧类型不对：%+v", []byte{frames[0].Type, frames[1].Type})
	}
	if !bytes.Equal(frames[1].Payload, mysteryBody) {
		t.Fatalf("未知帧载荷被改动了：%q", frames[1].Payload)
	}

	// 诊断：计数 1、最近列表带类型与偏移（这一帧紧跟在第一帧之后，偏移 = 第一帧的字节数）。
	diag := dec.Diagnostics()
	if diag.UnknownCount != 1 {
		t.Fatalf("未知类型计数 = %d，期望 1", diag.UnknownCount)
	}
	if len(diag.RecentUnknown) != 1 {
		t.Fatalf("最近未知帧 = %+v", diag.RecentUnknown)
	}
	u := diag.RecentUnknown[0]
	if u.Type != unknown {
		t.Fatalf("最近未知帧的类型 = %d，期望 %d", u.Type, unknown)
	}
	if u.Off != int64(len(first)) {
		t.Fatalf("最近未知帧的偏移 = %d，期望 %d", u.Off, len(first))
	}
	if u.PayloadLen != len(mysteryBody) {
		t.Fatalf("最近未知帧的载荷长度 = %d，期望 %d", u.PayloadLen, len(mysteryBody))
	}

	// 已知类型表：只认 AUI 与图片缓存这两类（实测穷尽值）。
	if !FrameTypeKnown(FrameTypeAUI) || !FrameTypeKnown(FrameTypeCache) {
		t.Fatal("已知类型被当成未知")
	}
	if FrameTypeKnown(unknown) || FrameTypeKnown(0) || FrameTypeKnown(0xff) {
		t.Fatal("未知类型被当成已知")
	}

	// 硬要求：未知类型的载荷里**恰好是一条合法 AUI 消息**，但它绝不能进模型 ——
	// 只有 type=1 的那一帧能被合并。
	m := NewModel()
	applied := 0
	for _, f := range frames {
		applied += m.ApplyFrame(f)
	}
	if applied == 0 {
		t.Fatal("type=1 的 AUI 帧没被合并")
	}
	if _, ok := m.Get(5); !ok {
		t.Fatal("type=1 的 Window 5 不在模型里")
	}
	if _, ok := m.Get(6); ok {
		t.Fatal("type=7 的载荷被当成 AUI 消息合并进模型了")
	}

	// 计数是累加的，最近列表有上限（不会因为一条长连接上的噪声无限长大）。
	for i := 0; i < 32; i++ {
		if _, err := dec.Feed(mystery); err != nil {
			t.Fatalf("Feed #%d: %v", i, err)
		}
	}
	diag = dec.Diagnostics()
	if diag.UnknownCount != 33 {
		t.Fatalf("未知类型计数 = %d，期望 33", diag.UnknownCount)
	}
	if len(diag.RecentUnknown) != UnknownFrameHistory {
		t.Fatalf("最近未知帧窗口 = %d，期望 %d", len(diag.RecentUnknown), UnknownFrameHistory)
	}
	if last := diag.RecentUnknown[len(diag.RecentUnknown)-1]; last.Off <= u.Off {
		t.Fatalf("最近未知帧没保持【最近】顺序：%+v", diag.RecentUnknown)
	}
}
