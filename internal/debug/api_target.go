package debug

// 服务端的**目标解析**（WP4b）：契约 C7.1 五处 `Manager.Current()` 兜底删掉之后，
// "这次要操作哪条会话"必须由**请求显式给出**，或者由会话池**唯一确定**。
//
// # 钉住的条款
//
//   - **Q4 / C8.2「服务端不保留隐式寻址」**：服务端不再"没给目标就替你挑一条"。
//     从前那五处（`sessOfWriteGlobal`:367 / `hSessionSwitch`:816 / `hWSLogDebug`:1679 /
//     `wslogTopent`:1724 / `hDBSQL`:1754）直呼 `Manager.Current()`，而它取的是 map 里
//     **任意一条**（`internal/debug/manager.go:434` 的实现 + `session_pool_behavior_test.go`
//     的实测）—— 多会话下那是"随机落一条，而输出看起来一切正常"，也就是计划开头那句
//     要根除的话：**不会出现"命令跑在另一条会话上却看起来成功"；落不上目标时，一定是报错。**
//   - **C8.1「只在多会话时收紧」**：池里 **0 条或 1 条**时，"省略目标"照旧确定性地落位
//     （0 条 = 没有东西可保护；1 条 = 唯一，不必问）；**>=2 条**才按 F1 报错并给候选。
//     这条与 CLI 侧 `internal/cli/debug/target.go` 的"唯一就落"同源：**唯一**是确定性的，
//     不是"替调用方挑一条"。
//   - **C5.15 / C5.16**：失败信封在现行 `{ok,error}` 形状上**增加** `code` 与 `candidates`，
//     `ok` / `error` 一个字都不改；缺目标/歧义用 **409 Conflict**。
//   - **C5.12**：候选行的键集是那**九键**（唯一出处是契约；本文件只投影，不另立形状）。
//
// # 它不做什么
//
//   - **不复刻 `internal/target` 的选择器文法**（那要 import 它，而 `internal/debug`
//     是那个包的"被引用方"，反向 import 会成环；Q25 的纪律是目标层零 `internal` 依赖）。
//     服务端只认它**自己**的原始句柄 `{id}` 与"池内唯一"这两种目标来源 ——
//     "`--session` 选择器怎么解析"这件事住在 CLI 的目标层里（Q12 / C8.2）。

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"tt/internal/output"
)

// candidateSession 是服务端失败信封里的一条候选行（契约 C5.12 的**九键**，
// 与 `session` 块同一个形状的子集）。
//
// 字段名与 `internal/target/view.go` 的 `Candidate` 逐字对齐（C5.12 是唯一出处，
// 本结构只是把同一套键在服务端再摆一次 —— 服务端**不能** import 目标层，见文件头）。
// 服务端今天还没有 `ref` / `display` / `name` / `n` / `ent` 这五个键（C5.3 的纯新增，
// 归别名与世代那一批），所以 `ref` / `display` 按现有字段现算、`n` 恒为 0 ——
// 不许编一个假序号出来：CLI 那一侧会用自己的目标层把 `ref` / `display` 重算一遍。
type candidateSession struct {
	ID      string `json:"id"`
	Ref     string `json:"ref"`
	Display string `json:"display"`
	Name    string `json:"name"`
	Env     string `json:"env"`
	Ent     string `json:"ent"`
	N       int    `json:"n"`
	Prog    string `json:"prog"`
	State   string `json:"state"`
}

// failureEnvelope 是 **HTTP 侧**的失败信封（契约 C5.15）。
//
// 它与 `fail()` 那个两键形状（B8：`{"ok":false,"error":…}`）的关系是"**增加**"：
// `ok` 与 `error` 一字不改，多出来的只有 `code`（必给）与 `candidates`（按 C5.11）。
// 为什么不在 `fail()` 上直接改：那是**所有**端点的失败出口，而 C5.15 只要求
// "要一条会话却拿不到目标"这一族带码 —— 顺手给全站加上 `code` 会把
// `TestContractHTTPFailureEnvelopeKeepsItsShape`（它钉的正是"今天不许有 code"）打红，
// 而那条测试的改造属于别名与世代那一批。
type failureEnvelope struct {
	OK         bool                `json:"ok"`
	Code       string              `json:"code"`
	Error      string              `json:"error"`
	Candidates *[]candidateSession `json:"candidates,omitempty"`
}

// resolveReasons 是"解析目标"的四种结局。
type resolveReasons int

const (
	// targetOK：拿到了一条确定的会话。
	targetOK resolveReasons = iota
	// targetEmptyPool：会话池里 0 条 —— **不是错误**（C7.1-1：没有东西可保护，放行）。
	//
	// 为什么它是"放行"而不是"报 NO_SESSION"：`hLaunch`（起会话）走的就是这条 ——
	// 池子空的时候恰恰是最需要能起会话的时候。若把它判失败，"第一条会话怎么来的"
	// 就没有答案了（C8.1 的单会话便利会当场死掉）。
	targetEmptyPool
	// targetMissing：调用方点名了一条会话，而它不在池里。→ 404 + `SESSION_NOT_FOUND`。
	targetMissing
	// targetAmbiguous：池里 >=2 条且调用方没点名。→ **409 + `SESSION_AMBIGUOUS` + 候选清单**（F1 / C5.16）。
	targetAmbiguous
)

// resolveTarget 在会话池上落一条**确定**的会话（C8.1：0 条或 1 条确定，>=2 条报错）。
//
// 两条来源，按顺序判：
//
//	① 点名的原始句柄（`targetID`，来自 URL 的 `{id}` 或请求体的 `session`）——
//	   它是**权威**：给了就只认它，找不到就报 `SESSION_NOT_FOUND`（不回落池子，
//	   否则"我点名的那条没了"会被静默换成另一条，正是要根除的那件事）；
//	② 省略目标 —— 池里**恰好一条**时落它（唯一即确定，与 CLI 的"唯一就落"同源）；
//	   0 条 → `targetEmptyPool`（放行）；>=2 条 → `targetAmbiguous` 并列出全部候选。
//
// `cands` 只在 `targetAmbiguous` 时非空（C5.11：候选只在那几种码下出现）。
func (s *Server) resolveTarget(targetID string) (sess *Session, reason resolveReasons, cands []candidateSession) {
	if id := strings.TrimSpace(targetID); id != "" {
		if sess := s.mgr.Get(id); sess != nil {
			return sess, targetOK, nil
		}
		return nil, targetMissing, nil
	}
	briefs := s.mgr.Snapshot()
	switch len(briefs) {
	case 0:
		return nil, targetEmptyPool, nil
	case 1:
		return s.mgr.Get(briefs[0].ID), targetOK, nil
	default:
		return nil, targetAmbiguous, sessionsToCandidates(briefs)
	}
}

// mustTarget 要一条**确定的**会话,不对就当场写好失败信封并返回 false。
//
// 它把"落位 → 判模式闸门"这组动作收成一处:两个端点（切环境 / 按日志重放）都要
// "先落位、再按那一条判谁能写",两处各写一遍必然漂。
//
// 与 `resolveTarget` 的差别只有一处:池子空**也报错**（`targetEmptyPool` → 409 +
// `NO_SESSION`）。理由:这两个端点干的事都建立在"有一条会话"之上 —— 切环境要收掉它、
// 按日志重放也要先收掉它,而池空时服务端在"没有目标"这件事上什么都判不了。
// 报一句"先起一条会话"比替调用方在**零目标**上静默做一半要诚实。
// （`hLaunch` 不走这里:它是"池空也要能跑"的那条路,它自己那种 `nil` 目标见
// `sessOfWriteGlobal` 的注释。）
//
// `what` 见 `failTarget`；`targetID` 是调用方点名的原始句柄（空 = 省略目标）。
func (s *Server) mustTarget(w http.ResponseWriter, r *http.Request, targetID, what string) (*Session, bool) {
	sess, reason, cands := s.resolveTarget(targetID)
	if reason == targetOK {
		if !s.sessOfWriteGlobal(w, r, sess) {
			return nil, false
		}
		return sess, true
	}
	failTarget(w, reason, cands, what)
	return nil, false
}

// 同一份池子两次跑出的文案必须一样（否则报错信息没法当判据用）。
//
// "按 id 排序"是刻意的：`Manager.Snapshot()` 走的是 map 遍历，顺序本身就是随机的；
// 不排一下，同一次失败在两次请求里会列出两种顺序。
func sessionsToCandidates(briefs []SessionBrief) []candidateSession {
	out := make([]candidateSession, 0, len(briefs))
	for _, b := range briefs {
		if strings.TrimSpace(b.ID) == "" {
			continue
		}
		env := strings.TrimSpace(b.Env)
		// ref（规范 id `环境#序号`）**留空**：序号 `n` 是 C5.3 的纯新增键，服务端今天
		// 还没有它（`Manager.Current()` 那五处与它无关，本轮不许为它改服务端）。
		// 没有序号就拼不出规范 id，而**编一个假序号比留空更坏** —— ref 是恒定不变的
		// 地址（C2.0 / C5.4），一个错的地址会让人照抄着重发一枪。
		// display（人读形态）只用环境 + 作业，不需要序号，能给就给。
		display := env
		if env != "" && b.Prog != "" {
			display = env + "/" + b.Prog
		}
		out = append(out, candidateSession{
			ID: b.ID, Ref: "", Display: display,
			Env: env, Prog: b.Prog, State: b.State,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// failTarget 把"要一条会话却拿不到目标"这一族写成 C5.15 / C5.16 的失败信封。
//
//   - `targetAmbiguous` → **409** + `SESSION_AMBIGUOUS` + 候选清单（F1 的形状）；
//   - `targetMissing`   → **404** + `SESSION_NOT_FOUND`（指名的那条不在池里）；
//   - `targetEmptyPool` → **409** + `NO_SESSION`（调用方这一步**必须**有一条会话才做得下去）；
//   - `targetOK`        → 调用方判错了状态，如实说一句（不当成静默成功）。
//
// `what` 是"哪个端点在哪一步要目标"和（点名时的）那个句柄，两样都要进文案：
// 读错误的人得能一眼看出是切换环境、按日志重放、还是查库那一步缺目标。
func failTarget(w http.ResponseWriter, reason resolveReasons, cands []candidateSession, what string) {
	what = strings.TrimSpace(what)
	switch reason {
	case targetAmbiguous:
		list := cands
		if list == nil {
			list = []candidateSession{}
		}
		msg := what + "：有 " + strconv.Itoa(len(list)) + " 条 SSH 会话，无法确定你要哪一条（请带上目标会话的 id）"
		writeJSON(w, http.StatusConflict, failureEnvelope{
			OK: false, Code: output.CodeSessionAmbiguous, Error: msg, Candidates: &list,
		})
	case targetMissing:
		writeJSON(w, http.StatusNotFound, failureEnvelope{
			OK: false, Code: output.CodeSessionNotFound,
			Error: "会话不存在: " + what,
		})
	case targetEmptyPool:
		writeJSON(w, http.StatusConflict, failureEnvelope{
			OK: false, Code: output.CodeNoSession,
			Error: what + "：没有 SSH 会话 —— 先运行 tt debug start <作业> --env <环境> 起一条",
		})
	default:
		writeJSON(w, http.StatusConflict, failureEnvelope{
			OK: false, Code: output.CodeSessionState,
			Error: "这次请求要的会话状态不对: " + what,
		})
	}
}
