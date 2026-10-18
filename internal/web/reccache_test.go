package web

// 录制解析结果的读侧缓存（契约 W5）。
//
// 三条要钉的事：
//  1. **文件没变就不重解** —— 明细页/实时帧流一秒一读，重复解析是纯浪费；
//  2. **文件一变就重解** —— 活录制追加了字节必须看到新帧（否则实时监控就是假的）；
//  3. 命中时给的是**副本** —— 调用方改自己那份，不许把缓存里的模型改坏。

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecordingParseIsCachedUntilFilesGrow(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	id := seedDetailRecording(t, s, cfgPath)

	base := recParseCount.Load()
	if code, _ := getJSON(t, s, "/api/recordings/"+id+"/timeline?limit=5"); code != 200 {
		t.Fatalf("第一次读该 200，得 %d", code)
	}
	if got := recParseCount.Load() - base; got != 1 {
		t.Fatalf("第一次读该解析一次，实得 %d 次", got)
	}

	// 同一个端点再读两次：文件没变 → 一次都不该再解析。
	for i := 0; i < 2; i++ {
		if code, _ := getJSON(t, s, "/api/recordings/"+id+"/timeline?limit=5"); code != 200 {
			t.Fatalf("第二次读该 200，得 %d", code)
		}
	}
	if got := recParseCount.Load() - base; got != 1 {
		t.Fatalf("文件没变时不许重复解析，实得累计 %d 次", got)
	}

	// 换一个端点（overview）也不该重解析：缓存是"这份录制"，不是"这个端点"。
	if code, _ := getJSON(t, s, "/api/recordings/"+id+"/overview"); code != 200 {
		t.Fatalf("overview 该 200，得 %d", code)
	}
	if got := recParseCount.Load() - base; got != 1 {
		t.Fatalf("换个端点也该命中缓存，实得累计 %d 次", got)
	}

	// 追加一帧（活录制）→ 大小变了 → 必须重解，并且新帧要能被读到。
	raw := filepath.Join(filepath.Dir(cfgPath), "fep", id+"-1-rt2gdc.log")
	f, err := os.OpenFile(raw, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(auiFrame(`om 9 {{un 42 {{value "ZZ"}}}}`)); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	code, out := getJSON(t, s, "/api/recordings/"+id+"/timeline?limit=200")
	if code != 200 {
		t.Fatalf("追加后该 200，得 %d", code)
	}
	if got := recParseCount.Load() - base; got != 2 {
		t.Fatalf("文件变了必须重解，实得累计 %d 次", got)
	}
	total, _ := out["total"].(float64)
	if total <= 0 {
		t.Fatalf("追加后该有帧，实得 total=%v", out["total"])
	}
	// 副本：把调用方拿到的那份改掉，下一次读不该看到改动。
	out["frames"] = []any{}
	if code, again := getJSON(t, s, "/api/recordings/"+id+"/timeline?limit=200"); code != 200 || len(again["frames"].([]any)) == 0 {
		t.Fatalf("缓存被调用方改坏了：第二次读 frames=%v", again["frames"])
	}
}
