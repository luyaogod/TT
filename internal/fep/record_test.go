package fep

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 真实录制件的形状：第一条握手 `meta …` **不带帧头**，其后的消息都带。
// 拿“首字节是不是字母”判形态会在这种文件上翻车（整屏解析成垃圾），所以必须先按帧试。
func TestReadRecordingPrefersFramesOverTextHeuristic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x-1-rt2gdc.log")
	handshake := []byte(`meta Connection {{encoding "UTF-8"}}` + "\n")
	framed := Encode(Frame{Type: FrameTypeAUI, Payload: []byte(`om 4 {{an 0 Window 4097 {{name "w"}} {{an 4097 FormField 271 {{value ""}} {}}}}}`)})
	framed2 := Encode(Frame{Type: FrameTypeAUI, Payload: []byte(`om 7 {{un 271 {{value "CNJ-A50"}}}}`)})
	blob := append(append(append([]byte{}, handshake...), framed...), framed2...)
	if err := os.WriteFile(path, blob, 0o600); err != nil {
		t.Fatal(err)
	}
	msgs, err := ReadRecording(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) < 2 {
		t.Fatalf("解析出的消息太少（%d）：形态判定错了", len(msgs))
	}
	m := NewModel()
	for _, msg := range msgs {
		m.ApplyMessage(msg)
	}
	o, ok := m.Get(271)
	if !ok {
		t.Fatal("对象 271 没建出来")
	}
	if o.Attr("value") != "CNJ-A50" {
		t.Fatalf("value = %q（attrs=%v）", o.Attr("value"), o.Attrs)
	}
	// 握手（无帧头）不该被当成消息正文误读成对象
	for _, msg := range msgs {
		if strings.HasPrefix(msg, "meta ") && strings.Contains(msg, "{{") && !strings.Contains(msg, "an ") {
			t.Fatalf("握手被当成 AUI 消息了: %q", msg)
		}
	}
}

// 早期原型的文本落盘（行首带帧头噪声）也必须继续能读。
func TestReadRecordingStillHandlesTextForm(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "t-2-gdc2rt.log")
	blob := "\x00\x00\x00+   +om 0 {{an 0 UserInterface 0 {{name \"x\"}}}}\n" +
		"\x00\x00\x18\x05\x06\xff!!__cached__!!\n"
	if err := os.WriteFile(path, []byte(blob), 0o600); err != nil {
		t.Fatal(err)
	}
	msgs, err := ReadRecording(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 || !strings.HasPrefix(msgs[0], "om 0") {
		t.Fatalf("msgs = %q", msgs)
	}
}
