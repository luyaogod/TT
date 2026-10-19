package target

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// 审计行（计划 WP6 的验收项之一：**每次会话操作有一条审计行**）（谁 / 哪条 / 干了什么）。
//
// 形状取本仓库既有审计的同一族：`internal/form42f` 的 `AppendAudit`（`internal/form42f/do.go:571-605`）
// 写的就是"一条 JSON 行"，字段名固定、时间为 RFC3339。这里不另立一种日记格式。
//
// 「哪条」写的是**规范 id**（`ref`）：契约 C5.4 明写"机器对账、错误文案、审计行用它"，
// 因为它在会话生命周期内恒定不变（C2.0）—— 拿会漂的 `display`（含作业段）记审计，
// 事后就没法把两条记录对上同一条会话。

// ActorEnv 是"谁"的缺省来源：环境变量 `TT_ACTOR`。
//
// 与 `internal/debug/api.go` 用的是**同一个**环境变量（那里把它当作 HTTP 调用者身份），
// 不另立第二个开关。
const ActorEnv = "TT_ACTOR"

// AuditEntry 是一条会话操作审计。
type AuditEntry struct {
	Time   time.Time `json:"time"`
	Actor  string    `json:"actor"`  // 谁
	Ref    string    `json:"ref"`    // 哪条：规范 id（C2.0 / C5.4）
	Action string    `json:"action"` // 干了什么
}

// NewAuditEntry 造一条审计。actor 取不到（空/空白）时给字面量占位 ——
// 审计行缺一段比写着一个 `?` 更难查。
func NewAuditEntry(actor, ref, action string, at time.Time) AuditEntry {
	if strings.TrimSpace(actor) == "" {
		actor = UnknownPlaceholder
	}
	return AuditEntry{Time: at, Actor: actor, Ref: ref, Action: action}
}

// JSONLine 把这条审计序列化成**一行**（不含换行），字段顺序固定。
//
// 没给时间就补当前时间（与 `internal/form42f.AppendAudit` 同一口径）：审计行不许出现零值时间。
func (e AuditEntry) JSONLine() (string, error) {
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// HumanLine 是人话形态：`审计 <time> · 谁 <actor> · 哪条 <ref> · 干了什么 <action>`。
func (e AuditEntry) HumanLine() string {
	at := e.Time
	if at.IsZero() {
		at = time.Now()
	}
	return fmt.Sprintf("审计 %s · 谁 %s · 哪条 %s · 干了什么 %s",
		at.Format(time.RFC3339), e.Actor, e.Ref, e.Action)
}

// DefaultActor 是"谁"的缺省来源：读 `TT_ACTOR`（见 ActorEnv）。
//
// getenv 由调用方注入（本包不直接读环境变量：那会让纯函数变成不可测的副作用）。
// 取不到就给字面量占位，绝不返回空串。
func DefaultActor(getenv func(string) string) string {
	if getenv != nil {
		if v := strings.TrimSpace(getenv(ActorEnv)); v != "" {
			return v
		}
	}
	return UnknownPlaceholder
}
