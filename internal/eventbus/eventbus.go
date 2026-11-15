// Package eventbus 进程内的事件环：宿主把配置变更、插件热重载等事件放进环，
// 插件的 SDK 轮询增量读取（after_id），避免长连接。
package eventbus

import (
	"encoding/json"
	"sync"
	"time"
)

// Event 一条宿主事件。
type Event struct {
	ID      int64
	Name    string
	Time    time.Time
	Payload json.RawMessage
}

// Bus 固定容量的环形事件表（满了覆盖最旧的）。
type Bus struct {
	mu   sync.Mutex
	next int64
	max  int
	ring []Event
}

// New 造一个最多保留 max 条事件的环（max<=0 时用 200）。
func New(max int) *Bus {
	if max <= 0 {
		max = 200
	}
	return &Bus{max: max}
}

// Publish 发布一条事件；payload 会被 JSON 编码（nil = 空对象）。
func (b *Bus) Publish(name string, payload any) Event {
	var raw json.RawMessage
	if payload != nil {
		if data, err := json.Marshal(payload); err == nil {
			raw = data
		}
	}
	if raw == nil {
		raw = json.RawMessage("{}")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.next++
	ev := Event{ID: b.next, Name: name, Time: time.Now(), Payload: raw}
	b.ring = append(b.ring, ev)
	if len(b.ring) > b.max {
		b.ring = b.ring[len(b.ring)-b.max:]
	}
	return ev
}

// Since 返回 id 大于 after 的事件（按 id 升序）。
func (b *Bus) Since(after int64) []Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]Event, 0, len(b.ring))
	for _, ev := range b.ring {
		if ev.ID > after {
			out = append(out, ev)
		}
	}
	return out
}

// Latest 当前最新事件 id（SDK 首次订阅时从"现在"开始，不重放历史）。
func (b *Bus) Latest() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.next
}
