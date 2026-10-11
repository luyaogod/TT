package fep

import (
	"strings"
	"time"
)

// 回放：把一轮真人动作重新发一遍。
//
// 与“语义动作”（`rec steps`，给人看的步骤清单）**不是一回事**：步骤清单要的是干净
// （丢导航），回放要的是**保真**，所以白名单比步骤清单宽。
//
// 宽到什么程度是实测出来的：当年把 `ConfigureEvent{currentRow/currentColumn}`
// （前端回抱“当前行/列”）当噪声过滤掉，回放到开窗那一步网格全是空的，作业还弹
// “传入的料件编号为空!” —— 那对属性**就是“选中哪一行”的语义**。
// 仍然排除的：tabIndex / bufferSize / pageSize 这类纯状态噪声。

// AllowedForReplay reports whether one event should be re-sent.
//
// 回放求的是**保真**：前端当时发出去的东西，就按当时的顺序再发一遍。
// 早前我把 ConfigureEvent{cursor…}（焦点）当"导航噪声"滤掉，结果回放跑不动 ——
// 今天查清楚了：**active/focus 是运行期的前置条件**（运行期自己都在推 un 0 {{focus "…"}}），
// 前端恰是先用 cursor 事件告诉它"现在光标在这个栏位"，运行期才认后面的动作。
// 所以回放白名单 = 全都放行（给"看得懂的步骤清单"用的过滤在 SemanticActions，不在这里）。
func AllowedForReplay(ev string) bool {
	ev = strings.TrimSpace(ev)
	if ev == "" {
		return false
	}
	return strings.HasPrefix(ev, "ActionEvent") ||
		strings.HasPrefix(ev, "KeyEvent") ||
		strings.HasPrefix(ev, "ConfigureEvent")
}

// ReplayEvent is one event scheduled for replay.
type ReplayEvent struct {
	N     int    `json:"n"`
	Event string `json:"event"`
}

// ReplaySkip records one event that will not be replayed, and why.
type ReplaySkip struct {
	Raw    string `json:"raw"`
	Reason string `json:"reason"`
}

// ReplayPlan is what a replay would send, plus what it would skip and why.
// 先出计划再执行：让人能先看一遍（`--dry-run` 打的就是它）。
type ReplayPlan struct {
	Events  []ReplayEvent `json:"events"`
	Skips   []ReplaySkip  `json:"skips"`
	Scanned int           `json:"scanned"`
}

// ReplayOptions narrows a replay.
//
// 注意这里**没有** Model：计划阶段不该做 AUI 判断 —— 那会把“后面才出现的按钮”
// （确定/审核/是）提前剔掉（实测踩过：保存与审核那 4 步根本没发）。
// AUI 安全闸只放在“发之前那一刻”，逐条判、逐条记原因。
type ReplayOptions struct {
	StopAt int      // >0：只回放到第 N 条（含），用于“跑一半看效果”
	Skip   []string // 命中（子串）就跳过，便于绕开某一两步
}

// PlanReplay turns recorded events into a replay plan.
func PlanReplay(events []string, opt ReplayOptions) ReplayPlan {
	plan := ReplayPlan{}
	for _, ev := range events {
		plan.Scanned++
		ev = strings.TrimSpace(ev)
		if !AllowedForReplay(ev) {
			plan.Skips = append(plan.Skips, ReplaySkip{Raw: ev, Reason: "不在回放白名单（导航/状态噪声）"})
			continue
		}
		if hit := skippedBy(ev, opt.Skip); hit != "" {
			plan.Skips = append(plan.Skips, ReplaySkip{Raw: ev, Reason: "按要求跳过：" + hit})
			continue
		}
		plan.Events = append(plan.Events, ReplayEvent{N: len(plan.Events) + 1, Event: ev})
		if opt.StopAt > 0 && len(plan.Events) >= opt.StopAt {
			break
		}
	}
	return plan
}

func skippedBy(ev string, skip []string) string {
	for _, s := range skip {
		if s != "" && strings.Contains(ev, s) {
			return s
		}
	}
	return ""
}

// ReplayResult summarises a replay run.
type ReplayResult struct {
	Sent         int   `json:"sent"`
	Skipped      int   `json:"skipped"`
	Failed       int   `json:"failed"`
	RuntimeBytes int64 `json:"runtimeBytes"`
}

// RunReplay sends the plan through the resident session.
//
// 每条之间等一小会儿（默认 300ms）：T100 的运行期按顺序处理，压得太快会把上一屏
// 还没建好的对象当成“不存在”，于是 AUI 安全闸会误杀后面的步骤。
func (c *Client) RunReplay(plan ReplayPlan, conn int, settle time.Duration, dryRun bool,
	logf func(string, ...any)) (*ReplayResult, error) {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	if settle == 0 {
		settle = 300 * time.Millisecond
	}
	if conn == 0 {
		conns, err := c.Status()
		if err != nil {
			return nil, err
		}
		picked, perr := pickConn(conns, 0)
		if perr != nil {
			return nil, perr
		}
		conn = picked.ID
	}
	res := &ReplayResult{}
	var bytesBefore int64
	if conns, err := c.Status(); err == nil {
		for _, cc := range conns {
			if cc.ID == conn {
				bytesBefore = cc.BytesRT
			}
		}
	}
	for _, ev := range plan.Events {
		if dryRun {
			logf("（dry-run）#%d %s", ev.N, ev.Event)
			res.Skipped++
			continue
		}
		// 每条发之前重新读界面：前面几步会改屏，用旧模型做安全闸会误杀
		if m, _, _, err := LoadModelFrom(c.RecordDir, c.Tag); err == nil {
			if id, ok := idRefOf(ev.Event); ok {
				if _, exists := m.Get(id); !exists {
					logf("跳过 #%d：对象 #%d 不在当前界面", ev.N, id)
					res.Skipped++
					continue
				}
			}
		}
		if _, err := c.Send(conn, ev.Event); err != nil {
			logf("失败 #%d：%v", ev.N, err)
			res.Failed++
			continue
		}
		res.Sent++
		// 每一步都**等运行期回包**再发下一步：一屏查询要几秒，
		// 固定睡一小会儿根本不够（实测：不等回包就发，后面的步骤全都打在空气上）。
		if !waitResponse(c, conn, 8*time.Second) {
			logf("第 %d 步运行期在 8s 内没有回包（继续）", ev.N)
		}
		time.Sleep(settle)
	}
	if conns, err := c.Status(); err == nil {
		for _, cc := range conns {
			if cc.ID == conn {
				res.RuntimeBytes = cc.BytesRT - bytesBefore
			}
		}
	}
	return res, nil
}

// waitResponse waits until the runtime pushes something on this connection.
// 返回是否等到了。等不到不算失败（有些事件本来就没有应答，比如光标移动），
// 但要记下来 —— 后面的步骤是否可信，取决于这个信号。
func waitResponse(c *Client, conn int, timeout time.Duration) bool {
	var before int64
	if conns, err := c.Status(); err == nil {
		for _, cc := range conns {
			if cc.ID == conn {
				before = cc.BytesRT
			}
		}
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if conns, err := c.Status(); err == nil {
			for _, cc := range conns {
				if cc.ID == conn && cc.BytesRT != before {
					return true
				}
			}
		}
		time.Sleep(150 * time.Millisecond)
	}
	return false
}
