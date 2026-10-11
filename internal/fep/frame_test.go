package fep

import (
	"bytes"
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
