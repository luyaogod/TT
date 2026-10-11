package fep

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// 每连接模型 + 等待条件。
//
// 连接与窗口的对应关系**就藏在文件名里**：`<tag>-<ID>-rt2gdc.log` 是那条连接收到的
// AUI。所以"这个窗口在哪条连接上"不需要额外记账 —— 这是本轮真实拓扑（多窗口 = 多连接）
// 唯一可靠的依据，也是把事件发到正确连接的前提（发错连接的后果实测过：窗口不动）。

// ConnectionModels rebuilds one AUI model per connection of a session.
// 返回 map[连接号]模型；键是文件名里的连接号（只有真机代理落盘的录制件才有）。
func ConnectionModels(dir, tag string) (map[int]*Model, map[int]*Resolver, error) {
	files, err := RecordingFiles(dir, "-rt2gdc.log")
	if err != nil {
		return nil, nil, err
	}
	type one struct {
		id   int
		path string
	}
	var mine []one
	for _, f := range files {
		base := filepath.Base(f)
		if tag != "" && !strings.HasPrefix(base, tag) {
			continue
		}
		id := connIDOf(base)
		if id <= 0 {
			continue
		}
		mine = append(mine, one{id: id, path: f})
	}
	if len(mine) == 0 {
		return nil, nil, fmt.Errorf("fep: %s 下没有本会话（%s*）的连接录制件", dir, tag)
	}
	// 同一连接可能有多份（重连、多轮），按时间从早到晚喂
	sort.Slice(mine, func(i, j int) bool { return mine[i].path < mine[j].path })
	models := map[int]*Model{}
	for _, o := range mine {
		bodies, rerr := ReadRecording(o.path)
		if rerr != nil {
			return nil, nil, rerr
		}
		m, ok := models[o.id]
		if !ok {
			m = NewModel()
			models[o.id] = m
		}
		for _, body := range bodies {
			m.ApplyMessage(body)
		}
	}
	res := map[int]*Resolver{}
	for id, m := range models {
		res[id] = NewResolver(m)
	}
	return models, res, nil
}

// connIDOf extracts the connection id from `<tag>-<ID>-rt2gdc.log`.
func connIDOf(base string) int {
	s := strings.TrimSuffix(base, "-rt2gdc.log")
	i := strings.LastIndexByte(s, '-')
	if i < 0 {
		return 0
	}
	id := 0
	for _, c := range s[i+1:] {
		if c < '0' || c > '9' {
			return 0
		}
		id = id*10 + int(c-'0')
	}
	return id
}

// ConnForWindow returns the connection that owns a window with this name.
// 多个连接都有同名窗口时**报错**（宁可让人指定，也不猜着发）。
func (c *Client) ConnForWindow(name string) (int, error) {
	if name == "" {
		return 0, fmt.Errorf("fep: 没给窗口名")
	}
	models, _, err := ConnectionModels(c.RecordDir, c.Tag)
	if err != nil {
		return 0, err
	}
	var hits []int
	for id, m := range models {
		for _, w := range m.Windows() {
			if w.Attr("name") == name {
				hits = append(hits, id)
				break
			}
		}
	}
	sort.Ints(hits)
	switch len(hits) {
	case 0:
		return 0, fmt.Errorf("fep: 没有叫 %q 的窗口（现有连接：%d 条）", name, len(models))
	case 1:
		return hits[0], nil
	default:
		return 0, fmt.Errorf("fep: %q 在 %d 条连接上都存在（%v）：请用 --conn 指定", name, len(hits), hits)
	}
}

// WaitCond is a condition to wait for. All set fields must hold.
type WaitCond struct {
	Window string // 出现这个窗口名
	Value  string // "字段名=值"，字段名按名字解析（只认高置信度）
	Text   string // 界面里出现这段文本（任意属性里）
}

func (c WaitCond) empty() bool { return c.Window == "" && c.Value == "" && c.Text == "" }

// WaitFor polls the session's recordings until the condition holds.
//
// 轮询的是**录制件**而不是屏幕：协议是唯一权威，而录制件是协议的忠实副本。
// 等到之前先看 status 的字节数，没有新数据时连解析都省了。
func (c *Client) WaitFor(cond WaitCond, timeout time.Duration) (*Model, error) {
	if cond.empty() {
		return nil, fmt.Errorf("fep: 没给等待条件")
	}
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	deadline := time.Now().Add(timeout)
	var lastBytes int64 = -1
	var m *Model
	for {
		conns, err := c.Status()
		if err != nil {
			return nil, err
		}
		var total int64
		for _, cc := range conns {
			total += cc.BytesRT
		}
		if total != lastBytes {
			lastBytes = total
			if mm, _, _, merr := LoadModelFrom(c.RecordDir, c.Tag); merr == nil {
				m = mm
				if cond.holds(m) {
					return m, nil
				}
			}
		}
		if time.Now().After(deadline) {
			return m, fmt.Errorf("fep: 等条件超时（%s）：%s", timeout, cond.String())
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func (c WaitCond) String() string {
	var parts []string
	if c.Window != "" {
		parts = append(parts, "窗口 "+c.Window)
	}
	if c.Value != "" {
		parts = append(parts, "值 "+c.Value)
	}
	if c.Text != "" {
		parts = append(parts, "文本 "+c.Text)
	}
	return strings.Join(parts, "、")
}

func (c WaitCond) holds(m *Model) bool {
	if c.Window != "" {
		found := false
		for _, w := range m.Windows() {
			if w.Attr("name") == c.Window {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if c.Text != "" {
		found := false
		for _, o := range m.Objects() {
			for _, v := range o.Attrs {
				if strings.Contains(v, c.Text) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}
	if c.Value != "" {
		name, want, ok := strings.Cut(c.Value, "=")
		if !ok {
			return false
		}
		r := NewResolver(m)
		found := false
		for _, o := range m.Objects() {
			if o.Attr("value") != want {
				continue
			}
			if fn, rok := r.Resolve(o.ID); rok && fn.Text == name && fn.Confidence == ConfHigh {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// ProductionEvidence reports whether the session looks like a production zone, and
// the evidence for it. The check is a **warning source**, not a gate (the tool is
// meant to be strong; safety lives in the audit log and --dry-run): the job window
// carries its zone in plain text, which is evidence we can quote.
func ProductionEvidence(m *Model) (bool, string) {
	for _, o := range m.Objects() {
		for _, key := range []string{"text", "name", "title"} {
			v := o.Attr(key)
			if v == "" {
				continue
			}
			if strings.Contains(v, "正式区") {
				return true, v
			}
			if i := strings.Index(v, ":"); i >= 0 {
				db := v[i+1:]
				if j := strings.IndexAny(db, ",]"); j >= 0 {
					db = db[:j]
				}
				if strings.HasSuffix(strings.ToLower(strings.TrimSpace(db)), "prd") {
					return true, v
				}
			}
		}
	}
	return false, ""
}
