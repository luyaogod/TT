package fep

// 录制的**归属与留存**（用户 2026-10-15 定的策略）。
//
// 两类录制的寿命不一样，所以住的地方也不一样：
//
//	用户主动录的（面板点"开始录制"、或人在命令行上起一轮）
//	  → 长期库 <数据目录>/records/<环境>/<ENT>/<录制 ID>/
//	  → **不参与自动清理**；要在前端手动删（`tt recording rm`），或在环境被删掉时**级联删**
//	    （`tt recording prune` 是它的兜底：环境配置没了、库里还留着的那种）
//
//	系统自动录的（调试会话顺手录下来的上下文）
//	  → 留在缓存目录 <数据目录>/fep/
//	  → 走**系统统一的清理框架**：`tt serve` 启动清超过 7 天的、`tt cache clear` 全清
//
// 为什么用"住哪"来表达寿命，而不是加一个"过期时间"字段：清理框架只认
// `config.CacheSubdirs` 里列出的目录（`config.CleanCache`），所以"在不在那份清单里"
// 就是"会不会被自动清"的唯一判据。多一个时间字段就会有两个判据，而它们迟早对不上。
type Origin string

const (
	// OriginUser 用户主动录的：长期留着。
	OriginUser Origin = "user"
	// OriginAuto 系统自动录的：调试会话的上下文，定期清掉。
	OriginAuto Origin = "auto"
)

// Retention 是这份录制**留在哪**。
type Retention string

const (
	// RetainLibrary 进长期库，不参与自动清理。
	RetainLibrary Retention = "library"
	// RetainCache 留在缓存目录，随统一清理框架按年龄删。
	RetainCache Retention = "cache"
)

// retentionFor 是策略本身：**归属决定留存**，除非调用方显式指定（`--no-keep` 那种
// "我知道这轮只是试验，别留在库里"）。
//
// 把默认值写在这里而不是各个调用点：一个调用点写反了，症状是"用户主动录的那一轮
// 悄悄被清掉了"——而那种错在下次要用它的时候才发现。
func retentionFor(origin Origin, want Retention) Retention {
	if want != "" {
		return want
	}
	if origin == OriginAuto {
		return RetainCache
	}
	return RetainLibrary
}

// Known 报告这个归属是不是认识的（写进身份文件的值不允许是别的）。
func (o Origin) Known() bool { return o == OriginUser || o == OriginAuto }

// Label 给人读的名字。
func (o Origin) Label() string {
	switch o {
	case OriginUser:
		return "用户主动录的"
	case OriginAuto:
		return "系统自动录的"
	case "":
		return "（没说谁录的）"
	}
	return string(o)
}

// Label 给人读的名字。
func (r Retention) Label() string {
	switch r {
	case RetainLibrary:
		return "长期库（不自动清理）"
	case RetainCache:
		return "缓存目录（随统一清理按年龄删）"
	case "":
		return "（没说留哪）"
	}
	return string(r)
}
