package fep

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScratchWinTables(t *testing.T) {
	dir := filepath.Join(os.Getenv("APPDATA"), "tt", "fep")
	files, _ := RecordingFiles(dir, "-rt2gdc.log")
	f := files[0]
	bodies, err := ReadRecording(f)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel()
	for _, b := range bodies {
		m.ApplyMessage(b)
	}
	t.Logf("文件 %s 消息 %d 当前窗口 #%d", filepath.Base(f), len(bodies), m.CurrentWindow())
	for w, gs := range m.winTables {
		t.Logf("  winTables[%d] = %v", w, gs)
	}
	// 看看最近一条消息里有没有 Window+Table
	for i := len(bodies) - 1; i >= 0 && i > len(bodies)-8; i-- {
		b := bodies[i]
		if len(b) > 300 {
			b = b[:300]
		}
		t.Logf("msg[%d] %s", i, b)
	}
}
