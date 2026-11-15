package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/luyaogod/TT/internal/connsvc"
)

// TestWriterLogsJSONL 事件落成 JSONL，字段齐全且不包含 SQL/命令内容。
func TestWriterLogsJSONL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs", "audit.jsonl")
	w, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	w.Log(connsvc.Event{
		Action: "db.query", Caller: connsvc.Caller{Kind: connsvc.CallerCLIPlugin, PluginID: "com.a"},
		Env: "prod", Kind: connsvc.KindDB, OK: true, DurationMS: 12,
	})
	w.Log(connsvc.Event{
		Action: "ssh.exec", Caller: connsvc.Caller{Kind: connsvc.CallerBuiltin},
		Env: "test", Kind: connsvc.KindSSH, OK: false, Code: "SSH_EXEC_FAILED",
	})
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := splitLines(string(b))
	if len(lines) != 2 {
		t.Fatalf("行数 = %d", len(lines))
	}
	var first eventJSON
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatal(err)
	}
	if first.Action != "db.query" || first.Caller.PluginID != "com.a" || !first.OK || first.Env != "prod" {
		t.Fatalf("事件 = %+v", first)
	}
	if first.Time == "" {
		t.Fatal("缺时间")
	}
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			if i > start {
				out = append(out, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
