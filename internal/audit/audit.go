// Package audit 把 Host API 的调用事件写成 JSONL，供事后查"谁在什么时候连了哪个环境"。
//
// 只记元数据：动作、调用方、环境、连接类型、成功与否、耗时。**不记 SQL 文本、
// 不记命令内容、不记密码** —— 审计的价值是追责与排障，不是复制业务数据。
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/luyaogod/TT/internal/connsvc"
)

// maxSize 单个审计文件上限；超过就滚动到 .1（只保留一份历史，够排障）。
const maxSize = 8 << 20

// Writer 一个追加写的审计文件。
type Writer struct {
	mu   sync.Mutex
	path string
	f    *os.File
	size int64
}

// Open 打开（或创建）审计文件；父目录不存在会创建。
func Open(path string) (*Writer, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	st, _ := f.Stat()
	w := &Writer{path: path, f: f}
	if st != nil {
		w.size = st.Size()
	}
	return w, nil
}

// Close 关闭文件。
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return nil
	}
	err := w.f.Close()
	w.f = nil
	return err
}

type callerJSON struct {
	Kind       string `json:"kind"`
	PluginID   string `json:"pluginId,omitempty"`
	InstanceID string `json:"instanceId,omitempty"`
}

type eventJSON struct {
	Time       string     `json:"time"`
	Action     string     `json:"action"`
	Caller     callerJSON `json:"caller"`
	Env        string     `json:"env,omitempty"`
	Kind       string     `json:"kind,omitempty"`
	OK         bool       `json:"ok"`
	Code       string     `json:"code,omitempty"`
	DurationMS int64      `json:"durationMs,omitempty"`
}

// Log 追加一条事件（失败静默：审计不该把业务请求拖垮）。
func (w *Writer) Log(ev connsvc.Event) {
	if w == nil {
		return
	}
	line, err := json.Marshal(eventJSON{
		Time:   time.Now().Format(time.RFC3339Nano),
		Action: ev.Action,
		Caller: callerJSON{Kind: string(ev.Caller.Kind), PluginID: ev.Caller.PluginID, InstanceID: ev.Caller.InstanceID},
		Env:    ev.Env, Kind: string(ev.Kind), OK: ev.OK, Code: ev.Code, DurationMS: ev.DurationMS,
	})
	if err != nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return
	}
	n, err := w.f.Write(append(line, '\n'))
	if err != nil {
		return
	}
	w.size += int64(n)
	if w.size >= maxSize {
		w.rotateLocked()
	}
}

func (w *Writer) rotateLocked() {
	_ = w.f.Close()
	_ = os.Rename(w.path, w.path+".1")
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		w.f = nil
		return
	}
	w.f = f
	w.size = 0
}

// DefaultPath 审计文件位置：<配置目录>/logs/audit.jsonl。
func DefaultPath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "logs", "audit.jsonl")
}

// FormatLine 把一行 JSONL 转成人类可读的一行（给 CLI 用）；解析失败原样返回。
func FormatLine(line []byte) string {
	var ev eventJSON
	if err := json.Unmarshal(line, &ev); err != nil {
		return string(line)
	}
	status := "ok"
	if !ev.OK {
		status = "FAIL " + ev.Code
	}
	caller := ev.Caller.Kind
	if ev.Caller.PluginID != "" {
		caller += ":" + ev.Caller.PluginID
	}
	return fmt.Sprintf("%s  %-9s %-24s env=%-10s %-5s %s (%dms)",
		ev.Time, ev.Action, caller, ev.Env, ev.Kind, status, ev.DurationMS)
}
