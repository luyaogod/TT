package fep

import (
	"fmt"
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
func AllowedForReplay(ev string) bool {
	ev = strings.TrimSpace(ev)
	if ev == "" {
		return false
	}
	switch {
	case strings.HasPrefix(ev, "ActionEvent"):
		return true
	case strings.HasPrefix(ev, "KeyEvent"):
		return true
	case strings.HasPrefix(ev, "ConfigureEvent"):
		switch {
		case strings.Contains(ev, "value "):
			return true
		case strings.Contains(ev, "currentRow"), strings.Contains(ev, "currentColumn"):
			return true
		default:
			return false // tabIndex / bufferSize / pageSize 之类
		}
	}
	return false
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
type ReplayOptions struct {
	StopAt int      // >0：只回放到第 N 条（含），用于“跑一半看效果”
	Skip   []string // 命中（子串）就跳过，便于绕开某一两步
	Model  *Model   // AUI 安全闸：对象不在界面里就跳过（nil = 不校验）
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
		// AUI 安全闸：对象不在**当次界面**里，就绝不发（发了最轻是白费，最重是打断连接）
		if opt.Model != nil {
			if id, ok := idRefOf(ev); ok {
				if _, exists := opt.Model.Get(id); !exists {
					plan.Skips = append(plan.Skips, ReplaySkip{Raw: ev,
						Reason: fmt.Sprintf("对象 #%d 不在当前界面（AUI 安全闸）", id)})
					continue
				}
			}
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
