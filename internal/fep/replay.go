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
//
// ColName/FieldID 是这一步的**字段身份**（契约 C3 + A18），录制/计划阶段从当步目标
// 节点取（`IdentityOf`）。两处都**向后兼容**：旧计划没有这两个键 ⇒ 零值 = 未记录，
// 读侧照旧按 `idRef` 走，行为一字不变（见 ResolveStepTarget）。
type ReplayEvent struct {
	N     int    `json:"n"`
	Event string `json:"event"`
	// ColName 非空 ⇒ 回放时**优先**按它重新定位（`ByColName`，每步现解析、用完即弃）。
	ColName string `json:"colName,omitempty"`
	// FieldID 是录制时那一帧的编号，用来按 C3.4 收窄（0 = 未记录，见 ReplayIdentity）。
	FieldID int `json:"fieldId,omitempty"`
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
// 注意这里**不拿模型做取舍**：计划阶段不该做 AUI 判断 —— 那会把“后面才出现的按钮”
// （确定/审核/是）提前剔掉（实测踩过：保存与审核那 4 步根本没发）。
// AUI 安全闸只放在“发之前那一刻”，逐条判、逐条记原因。
//
// `Model` 是唯一的例外，而且它的用途被**收窄到一件事**：把字段身份（`colName`/
// `fieldId`）**记进步骤**，好让回放时能按身份重新定位。它**不**决定某条事件进不进
// 计划 —— 模型里没有的对象、没有 idRef 的按键事件，照样一条不少地留在计划里
// （反证测试：TestPlanStageModelNeverDropsEvents）。nil = 不记身份（旧行为）。
type ReplayOptions struct {
	StopAt int      // >0：只回放到第 N 条（含），用于“跑一半看效果”
	Skip   []string // 命中（子串）就跳过，便于绕开某一两步
	Model  *Model   // 只用于**记身份**，不参与取舍（见上）
}

// PlanReplay turns recorded events into a replay plan.
//
// 录制侧写入：`opt.Model` 非 nil 时，每条带 `idRef` 的事件都去模型里取它的
// `(colName, fieldId)` 记进步骤（`IdentityOf`；取不到就留空 = 未记录）。
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
		e := ReplayEvent{N: len(plan.Events) + 1, Event: ev}
		// 身份只**补记**：拿不到（模型里没这个对象 / 事件不带 idRef）就留零值。
		if id, ok := idRefOf(ev); ok {
			if idn := IdentityOf(opt.Model, id); idn.ColName != "" {
				e.ColName, e.FieldID = string(idn.ColName), int(idn.FieldID)
			}
		}
		plan.Events = append(plan.Events, e)
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
	// Refused 是"这一步带字段身份，但身份解析不过（四码）"的步数 —— 与 Failed 分开：
	// 它一个字节都没发，而且**这一轮到此为止**（后面的步骤依赖它，接着发只会连错）。
	Refused int `json:"refused,omitempty"`
	// Stopped 是"**人**按了停止"（不是失败、也不是跑完）—— 三者在界面上的说法完全不同：
	// 跑完是"完成"，失败要查原因，而"人停了"是正常操作。
	Stopped bool `json:"stopped,omitempty"`
	// Degraded 是"按录制时的临时编号发出去"的步数（没有正式身份 ⇒ 跨会话可能指到别的控件）。
	//
	// 它不是失败：这些步**照发**了（用户 2026-10-18 定），但必须能被看见 ——
	// 界面/CLI 把"发了几步、其中几步是降级"一起报出来。
	Degraded int `json:"degraded,omitempty"`
}

// RunReplay sends the plan through the resident session.
//
// 每条之间等一小会儿（默认 300ms）：T100 的运行期按顺序处理，压得太快会把上一屏
// 还没建好的对象当成“不存在”，于是 AUI 安全闸会误杀后面的步骤。
func (c *Client) RunReplay(plan ReplayPlan, conn int, settle time.Duration, dryRun bool,
	logf func(string, ...any)) (*ReplayResult, error) {
	return c.RunReplayWith(plan, ReplayRunOptions{Conn: conn, Settle: settle, DryRun: dryRun, Logf: logf})
}

// ReplayRunOptions 是 RunReplayWith 的全部参数（比一条长签名好读，也好往后加东西）。
type ReplayRunOptions struct {
	Conn   int
	Settle time.Duration
	DryRun bool
	Logf   func(string, ...any)
	// BreakAt 判"这一步发之前要不要停"（回放断点，用户 2026-10-18 定：**发之前停**）。
	//
	// nil = 没有断点能力（比如没有控制通道的调用方）。判据由调用方给：它知道
	// 断点文件在哪（`<回放日志>.breaks`），这里只管"到了就进暂停态"。
	BreakAt func(n int) bool
	// OnStep 逐步回调：**每一步的中间态与终态各叫一次**（waiting → sent/no-response）。
	//
	// 为什么连中间态也要叫：用户要看的是"现在走到哪了、等多久了"。只在结束时报告，
	// 得到的是一份"历史"，而屏幕上那 8 秒的等待仍然是黑的（那是实测里最难受的一段）。
	OnStep func(ReplayStep)
	// Ctl 是**人的命令**（暂停/单步/停止）在这一步之前读一次；返回空 = 继续。
	//
	// 为什么是"读一个函数"而不是"轮询一个文件"：文件那条路要处理"哪个文件、
	// 谁的日志"，而回放这一层只需要知道"现在该不该停"——把它留成注入点，
	// 命令行与面板各自给各自的实现（面板走缓存目录里的 .ctl 文件）。
	Ctl func() ReplayCtl
}

// RunReplayWith 与 RunReplay 同一条路，只是把每一步的遥测与人的控制点开出来。
func (c *Client) RunReplayWith(plan ReplayPlan, opt ReplayRunOptions) (*ReplayResult, error) {
	logf := opt.Logf
	if logf == nil {
		logf = func(string, ...any) {}
	}
	emit := opt.OnStep
	if emit == nil {
		emit = func(ReplayStep) {}
	}
	settle := opt.Settle
	if settle == 0 {
		settle = 300 * time.Millisecond
	}
	conn := opt.Conn
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
	total := len(plan.Events)
	// mode 是"人的命令"在这一步的状态：
	//
	//	""      正常跑
	//	"pause" 发之前停住（等继续/单步/停止）
	//	"step"  放行**这一步**，发完回到 "pause"
	//
	// 为什么不只用一个 bool：单步 = "放行一步再停"，它与"一直暂停"在**发完之后**的
	// 行为正好相反。用一个 bool 表达会写成 `if paused { paused = true }` 这种自欺的句子
	// （第一版就是这么写的，实测按单步会一直跑下去）。
	mode := ""
	// resumedAt 是"人刚按过继续"的那一步：断点不能在同一步上立刻又把它拦住
	// （否则「继续」等于没点 —— 真机上那就是死循环）。
	resumedAt := -1
	// pauseReason 决定停住时那行日志/面子上写什么（断点 vs 手动暂停）。
	pauseReason := "已暂停（等「继续」或「单步」）"
	for _, ev := range plan.Events {
		step := ReplayStep{N: ev.N, Total: total, Event: ev.Event, Kind: eventKind(ev.Event)}
		// skipped 是"这一步被人按了跳过"：**每一步都要重置** —— 声明在循环外会把
		// "跳过这一步"变成"从这里开始全跳过"。
		skipped := false
		if id, ok := idRefOf(ev.Event); ok {
			step.Target = id
		}

		// ① 人的命令在**这一步之前**生效
		if opt.Ctl != nil {
			switch opt.Ctl() {
			case CtlStop:
				logf("按用户要求停止（第 %d 步之前，已发 %d 条）", ev.N, res.Sent)
				res.Stopped = true
				return res, nil
			case CtlPause:
				mode = "pause"
			case CtlStep:
				mode = "step"
			case CtlResume:
				mode = ""
			case CtlSkip:
				// 没停在暂停态时按「跳过」= 跳过即将发的这一步；不改变暂停态。
				step.Result = ReplaySkipped
				step.Reason = "人要求跳过这一步"
				emit(step)
				res.Skipped++
				logf("跳过第 %d 步（人要求）", ev.N)
				skipped = true
			}
			if skipped {
				continue
			}
			// 断点：**发之前**停。人刚按过「继续」的那一步不拦（否则继续=没点）。
			if opt.BreakAt != nil && mode != "step" && !skipped && ev.N != resumedAt && opt.BreakAt(ev.N) {
				mode = "pause"
				pauseReason = fmt.Sprintf("断点：停在第 %d 步之前（点「发一步」或「继续」放行）", ev.N)
				logf("断点：停在第 %d 步之前", ev.N)
			}
			if mode == "pause" {
				// 只报**一次**"停在第几步"：这个等待循环每 200ms 转一圈，
				// 每圈都落一行会把日志冲成垃圾（而进度条只需要知道"停在哪"）。
				step.Result = ReplayPaused
				step.Reason = pauseReason
				emit(step)
				logf("已暂停在第 %d 步（等「继续」或「单步」）", ev.N)
				for mode == "pause" {
					time.Sleep(200 * time.Millisecond)
					switch opt.Ctl() {
					case CtlStop:
						res.Stopped = true
						logf("按用户要求停止（停在第 %d 步）", ev.N)
						return res, nil
					case CtlResume:
						mode = ""
						resumedAt = ev.N
					case CtlStep:
						mode = "step"
						resumedAt = ev.N
					case CtlSkip:
						// 停在第 N 步时按「跳过」：**不发它**，记一条 skipped，
						// 然后**继续停着**（下一步再等你按「发一步 / 继续」）。
						step.Result = ReplaySkipped
						step.Reason = "人要求跳过这一步"
						emit(step)
						res.Skipped++
						logf("跳过第 %d 步（人要求）", ev.N)
						skipped = true
					}
					if skipped {
						break
					}
				}
				if skipped {
					continue
				}
			}
		}

		if opt.DryRun {
			logf("（dry-run）#%d %s", ev.N, ev.Event)
			step.Result = ReplaySkipped
			step.Reason = "dry-run（只出计划不发）"
			emit(step)
			res.Skipped++
			continue
		}
		// 每条发之前重新读界面：前面几步会改屏，用旧模型做安全闸会误杀
		m, _, _, merr := LoadModelFrom(c.RecordDir, c.Tag)
		if merr != nil {
			m = nil
		}
		// ① 身份优先（契约 C3 + H1/H2）：步骤带 `colName` ⇒ **每步现解析、用完即弃**。
		//    解析不过就按四码拒绝，**一个字节都不发**（裸 id 跨会话会指到别的控件上）。
		if ColName(strings.TrimSpace(ev.ColName)) != "" {
			tgt, terr := ResolveStepTarget(m, ev, Scope{})
			if terr != nil {
				step.ColName, step.FieldID = ev.ColName, ev.FieldID
				if id, ok := idRefOf(ev.Event); ok {
					step.Target = id
				}
				step.Result = ReplayRefused
				step.Reason = refusalReason(terr)
				if ref, ok := terr.(*Refusal); ok {
					step.Code = ref.Code
				}
				emit(step)
				logf("拒绝 #%d：%s（这一步没发）", ev.N, step.Reason)
				res.Refused++
				return res, terr
			}
			ev.Event = tgt.Event // 只换目标，其余字节原样（回放求保真）
			step.Event, step.Target = tgt.Event, tgt.ID
			step.ColName, step.FieldID = string(tgt.ColName), int(tgt.FieldID)
			step.Resolved = "field"
		} else if step.Target != 0 {
			// 计划期没记身份 ⇒ **降级**：按录制时的临时编号发（用户 2026-10-18 定）。
			//
			// 为什么不再"整份拒绝"：真实操作必然含工具栏/表格点击，它们**天生没有字段身份**
			// （真机 37 份录制、87 个点击步骤实测：只有 1.1% 带 colName）。一票否决的实际效果
			// 是"真人录的录制永远不能回放"，而不是保护。
			//
			// 风险用**标注**兜住：这一步标 degraded、结果里计数、日志逐条写明；
			// 真正的硬拒仍留给"有身份但解析不过/有歧义"（那才是会点错的场合）。
			step.Resolved = "bare"
			if m != nil {
				if idn := IdentityOf(m, step.Target); idn.ColName != "" {
					step.ColName, step.FieldID = string(idn.ColName), int(idn.FieldID)
				}
			}
		}
		// ② 既有 AUI 安全闸：对象不在**当次界面**里就跳过（行为不变）
		if m != nil {
			if id, ok := idRefOf(ev.Event); ok {
				if _, exists := m.Get(id); !exists {
					logf("跳过 #%d：对象 #%d 不在当前界面", ev.N, id)
					step.Result = ReplaySkipped
					step.Reason = fmt.Sprintf("对象 #%d 不在当前界面（AUI 安全闸）", id)
					emit(step)
					res.Skipped++
					continue
				}
			}
		}
		// 真正发出去之前才计"降级"：被上面那道 AUI 闸跳过的步不算。
		if step.Resolved == "bare" {
			step.Degraded = true
			res.Degraded++
			logf("降级 #%d：没有正式身份，按录制时的临时编号 #%d 发送（现场若变了可能点错）", ev.N, step.Target)
		}
		if _, err := c.Send(conn, ev.Event); err != nil {
			logf("失败 #%d：%v", ev.N, err)
			step.Result = ReplayFailed
			step.Reason = err.Error()
			emit(step)
			res.Failed++
			continue
		}
		at := time.Now().UnixMilli()
		step.SentAt = &at
		step.Result = ReplayWaiting
		emit(step) // 中间态：界面从这一刻开始画"进行中 · 已等 Nms"，**不画时长**
		res.Sent++
		// 每一步都**等运行期回包**再发下一步：一屏查询要几秒，
		// 固定睡一小会儿根本不够（实测：不等回包就发，后面的步骤全都打在空气上）。
		got, delta := waitResponse(c, conn, 8*time.Second)
		step.RuntimeBytes = delta
		if got {
			rt := time.Now().UnixMilli()
			step.RespAt = &rt
			step.Result = ReplaySent
		} else {
			logf("第 %d 步运行期在 8s 内没有回包（继续）", ev.N)
			step.Result = ReplayNoResponse
			// 点击类事件通常一定有反应；光标/键盘那类本来就没应答（见 ReplayStep.Suspect）。
			step.Suspect = step.Kind == "ActionEvent"
			step.Reason = "8 秒内没有回包"
		}
		emit(step)
		// 单步：这一步发完了，回到"停住"（下一轮循环会重新报 paused 并等命令）。
		if mode == "step" {
			mode = "pause"
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

// eventKind 取一条事件文本的类别（`ActionEvent 0{{…}}` → `ActionEvent`）。
func eventKind(ev string) string {
	ev = strings.TrimSpace(ev)
	for i := 0; i < len(ev); i++ {
		if ev[i] == ' ' || ev[i] == '(' || ev[i] == '{' {
			return ev[:i]
		}
	}
	return ev
}

// waitResponse waits until the runtime pushes something on this connection.
// 返回（是否等到，期间运行期推了多少字节）。等不到不算失败（有些事件本来就没有应答，
// 比如光标移动），但要记下来 —— 后面的步骤是否可信，取决于这个信号。
func waitResponse(c *Client, conn int, timeout time.Duration) (bool, int64) {
	var before int64
	if conns, err := c.Status(); err == nil {
		for _, cc := range conns {
			if cc.ID == conn {
				before = cc.BytesRT
			}
		}
	}
	var got int64
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if conns, err := c.Status(); err == nil {
			for _, cc := range conns {
				if cc.ID == conn && cc.BytesRT != before {
					return true, cc.BytesRT - before
				}
			}
		}
		time.Sleep(150 * time.Millisecond)
	}
	return false, got
}
