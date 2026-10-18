package debug

// 界面会话的挂接变化 → 事件（W2 的 screen.attached / screen.detached）。
//
// 为什么由**接线层**盯注册表，而不是让录制进程回报或让调试线去读：
//   - 界面会话的真相在 `.tt-fep.json`，写它的是**录制进程**（一个脱离终端的子进程）；
//   - 调试线不允许 import `internal/fep`（H3：它不认识 GDC / 代理 / 录制）。
// 所以这里做两件事：把注册表读成一个**只读投影**、把两次投影的差异转成事件。
// 差异计算是纯函数（`screenChanges`），好测；只有那个循环带 IO。

import (
	"context"
	"time"

	"tt/internal/debug"
	"tt/internal/fep"
)

// screenEntry 是注册表里我们关心的一行（只读投影）。
//
// 字段与 `.tt-fep.json` 一一对应；**不在这里做任何判断**（stale 与否由读侧决定，
// 这里只报告"变了"）。
type screenEntry struct {
	Ref        string
	Tag        string
	Kind       string
	Env        string
	Prog       string
	GDCPid     int
	ProxyPort  int
	InjectPort int
	OwnerPID   int
}

// screenSnapshot 把注册表列表收成 ref → entry 的表。
//
// 用 ref 做键：一条 SSH 会话最多挂一块画面（同 ref 的第二块会覆盖前一块 —— 那本来就是
// "这条会话现在挂着哪一块"的答案，旧的那块已经被收掉了）。
func screenSnapshot(list []*fep.Session) map[string]screenEntry {
	out := make(map[string]screenEntry, len(list))
	for _, s := range list {
		if s == nil || s.Ref == "" {
			continue
		}
		out[s.Ref] = screenEntry{
			Ref: s.Ref, Tag: s.Tag, Kind: string(s.Kind), Env: s.Env, Prog: s.Prog,
			GDCPid: s.GDCPID, ProxyPort: s.ProxyPort, InjectPort: s.InjectPort, OwnerPID: s.OwnerPID,
		}
	}
	return out
}

// screenChanges 比较两次快照，给出"新挂上的"与"收掉的"。
//
// 判据是 **ref + tag**：同 ref 换了 tag（收一块、起另一块）算"收掉旧的 + 挂上新的"——
// 只比 tag 不比 ref 就会把"同一 tag 换会话"漏掉，只比 ref 不比 tag 就会把"换了一份录制"
// 当成没变化（而前端正要看的是"现在录的是哪一份"）。
func screenChanges(prev, cur map[string]screenEntry) (attached, detached []screenEntry) {
	for ref, e := range cur {
		if old, ok := prev[ref]; ok && old.Tag == e.Tag {
			continue // 没变
		}
		attached = append(attached, e)
	}
	for ref, e := range prev {
		if now, ok := cur[ref]; ok && now.Tag == e.Tag {
			continue // 还在
		}
		detached = append(detached, e)
	}
	return attached, detached
}

// toScreenState 把投影翻成事件里那个形状（`debug.ScreenState`）。
func toScreenState(e screenEntry, reason string) debug.ScreenState {
	return debug.ScreenState{
		Ref: e.Ref, Tag: e.Tag, Kind: e.Kind, Env: e.Env, Prog: e.Prog,
		GDCPid: e.GDCPid, ProxyPort: e.ProxyPort, InjectPort: e.InjectPort,
		OwnerPID: e.OwnerPID, Reason: reason,
	}
}

// ScreenWatchInterval 是对比注册表的周期。
//
// 400ms 是"人点一下到页面变色"的感知阈值之内，同时把读文件的成本压到可以忽略
// （注册表通常只有一两行）。调大只会让界面变慢，调小没有收益。
const ScreenWatchInterval = 400 * time.Millisecond

// StartScreenWatch 起一个循环：盯 `.tt-fep.json`，变化时回调（接线层把它接到
// `Server.PublishScreenChange`）。
//
// 三条纪律：
//   - **读不到就跳过**（保持上一份快照）：宁可晚一点，也不许因为一次读失败就广播
//     "全收掉了"——那会让页面把正在看的面板关掉。
//   - ctx 取消即退出：服务收尾时不能留 goroutine（否则进程退不掉）。
//   - 第一次 tick 把所有**已挂着**的广播一遍：服务重启后页面靠它追上现状
//     （`GET /screen` 也能答，但事件到了就不用等下一次轮询）。
func StartScreenWatch(ctx context.Context, dataDir string, publish func(attached bool, st debug.ScreenState)) {
	if dataDir == "" || publish == nil {
		return
	}
	go func() {
		var prev map[string]screenEntry
		t := time.NewTicker(ScreenWatchInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
			list, err := fep.NewRegistry(dataDir).List()
			if err != nil {
				continue
			}
			cur := screenSnapshot(list)
			attached, detached := screenChanges(prev, cur)
			for _, e := range attached {
				publish(true, toScreenState(e, ""))
			}
			for _, e := range detached {
				publish(false, toScreenState(e, "stop"))
			}
			prev = cur
		}
	}()
}
