package fep

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// 本文件是**录制件的身份**。口径 = 契约 A19「指纹 v3」。
//
// 判据全部来自实测（出处见 .agents/plan/tt-agent-field-notes.md §9、
// .agents/Genero Original Factory Compatible/out/fglrun321/id-allocator-decoded.md）：
//
//   - AUI 的对象编号由**运行期**按声明顺序分配（单调计数器 + LIFO 空闲栈，会被复用），
//     所以 id 是**启动历史的函数**：同一套界面换一条启动路径，对象集几乎一样（1556 / 1558），
//     而共同条目的编号**一条都不相同**（偏移 −31 到 −4274，不是等量平移）。
//   - 同一表单内 `colName`+`fieldId` **稳定且唯一**（离线解析 vs 运行期语料 509/509 = 100%）。
//
// 于是 v3 的主判据是**身份**，不是号码：
//
//	身份多重集  = 全部节点的 `colName#fieldId`（排序后取哈希）
//	引用边集合  = 父子边，写成「父身份→子身份」（缺身份时退化为 type#占位），排序后取哈希
//	类型直方图  = 节点类型 → 个数
//	idDigest   = 旧算法的哈希（整张「坐标→编号」表），**只存不判**（诊断/对比用）
//
// 判决（冻结，取代 A19 的双向 Jaccard：**非对称覆盖**，见契约 A21）：
//
//	通过 ⟺ |live| ≥ max(10, 0.5×|rec|)
//	      且 live→rec 身份覆盖 ≥ 0.98 且 边覆盖 ≥ 0.98
//	      且 live 多余身份 ≤ 2% 且 类型直方图差异 ≤ 2%
//
// 方向单向（live→rec）是因为会话是渐进的：录制件是整会话、当前只开半屏是常态，
// 误拒正是本轮要消灭的失败类。阈值写死在本文件，读不到配置。
//
// 指纹回答的是"这份录制是在**哪个界面**下录的"，而**不是**"这个 id 现在是谁"。
// 用它寻址只在同一轮会话内成立（那是"安全闸"的事）；拿号码当签名是 v2 的错，
// 会把"同一界面、不同启动路径"判成不同 —— 跨路径回放的误拒就是这么来的。
//
// 指纹里**不含业务值**（text/value）：同一条路径连录两轮，单号会变，拿值当签名
// 会把能用的录制判成失效；因此指纹没有任何业务数据，可以随便复制流转。

// FPVersion 是指纹口径的版本号（A19 = 3，判决按 A21 改为非对称覆盖）。
const FPVersion = 3

// 判决阈值：A21 **冻结**，写死在代码里（不读配置、不做环境变量逃生门）。
const (
	// FPEvidenceMinNodes / FPEvidenceRecShare 是①"证据量足够"的两条线：
	// |live| ≥ max(10, 0.5×|rec|) —— 防"只有 3 个节点也说 100% 覆盖"。
	FPEvidenceMinNodes = 10
	FPEvidenceRecShare = 0.5

	// FPIdentJaccardMin 是②"live→rec 身份覆盖"的下限（A21 起是**单向**覆盖，
	// 不再是 A19 的双向 Jaccard：会话是渐进的，rec ⊃ live 是常态）。
	FPIdentJaccardMin = 0.98
	// FPEdgeIsoMin 是③"live→rec 边覆盖"的下限（同样单向）。
	FPEdgeIsoMin = 0.98
	// FPTypeHistMaxDiff 是④"live 多余身份占比"的上限，也是⑤"类型直方图差异"的上限
	// （⑤ = ½·Σ|Δ类| / 较多一侧的节点数 = 变了类型的节点占比，按 live 的身份集合收窄后比）。
	FPTypeHistMaxDiff = 0.02
)

// fpDiffLimit 是差异清单最多列多少条（清单是给人读的，不是日志倾倒）。
const fpDiffLimit = 12

// FPEntry 是编号表里的一条坐标：**编号是多少、叫什么、它的身份是什么、挂在谁下面**。
//
// Role 取协议 name（源码里的控件名）→ comment（cmt_<表><栏位>）→ 类型名。
// 它**不参与 v3 判决**（只进 idDigest 与诊断），留着是因为拒绝报文要能说人话。
type FPEntry struct {
	Kind  string `json:"kind"`
	Scope int    `json:"scope,omitempty"`
	ID    int    `json:"id"`
	Role  string `json:"role,omitempty"`

	// v3（A19）：**身份**。colName + fieldId 是编译期分配、表单内稳定且唯一的一对，
	// 与会话编号无关 —— 所以它们才是判据。
	ColName string `json:"colName,omitempty"`
	FieldID string `json:"fieldId,omitempty"`

	// Parent 是父对象的**编号**。它只用来在本张表里找回父的**身份**（边写成"身份→身份"），
	// 编号自身不进判决。
	Parent int `json:"parent,omitempty"`
}

// FPKey 是"同一张表里的同一条坐标"（不含编号）：编号不同的就是**不同**的表。
// v3 起它只服务 idDigest 的诊断（CompareFP / TranslateEntries），不做判决。
type FPKey struct {
	Kind  string
	Scope int
	Role  string
}

func (e FPEntry) key() FPKey { return FPKey{Kind: e.Kind, Scope: e.Scope, Role: e.Role} }

// Fingerprint 是一份录制件的界面指纹（v3）。
//
// **必须保持可比较**（`==`）：身份文件往返（meta.go）与既有测试都按值比。
// 所以三件套里落的是**哈希/规范串**，不是 map —— 类型直方图用"排序后拼接"的规范串表达，
// 既可比较、又能直接读（`Edit=3 TableColumn=2`）。
type Fingerprint struct {
	Total int    `json:"total"` // 表里有多少条
	Hash  string `json:"hash"`  // **idDigest**：旧算法（整张坐标→编号表）的 sha256 前 16 位，只存不判
	// Version = FPVersion(3) 表示这枚指纹带 v3 三件套；0 = 旧 v2 指纹（读到就忽略 + 记诊断）。
	Version int `json:"version,omitempty"`

	Ident     string `json:"ident,omitempty"`     // 身份多重集（排序后）的哈希
	IdentN    int    `json:"identN,omitempty"`    // 身份条目数（含重复）
	IdentUniq int    `json:"identUniq,omitempty"` // 身份种类数

	Edges string `json:"edges,omitempty"` // 引用边集合（排序后）的哈希
	EdgeN int    `json:"edgeN,omitempty"` // 边的条数

	Types string `json:"types,omitempty"` // 类型直方图的规范串（按类型名排序，`Edit=3 TableColumn=2`）
}

// IsV3 报告这枚指纹带不带 v3 三件套（旧 v2 指纹没有）。
func (f Fingerprint) IsV3() bool { return f.Version >= FPVersion }

// IDDigest 是旧算法的哈希（**只存不判**：判决里任何地方都不许拿它放行）。
func (f Fingerprint) IDDigest() string { return f.Hash }

// TypeHistogram 把类型直方图的规范串解析回 map（给诊断/展示用）。
func (f Fingerprint) TypeHistogram() map[string]int { return parseTypeHist(f.Types) }

// String 给人读的形态（跨路径比对时一眼能看出是不是同一张表）。
func (f Fingerprint) String() string {
	if f.Hash == "" && f.Ident == "" {
		return "（没有指纹）"
	}
	if !f.IsV3() {
		return fmt.Sprintf("%s（%d 条，v2 指纹）", f.Hash, f.Total)
	}
	ident := f.Ident
	if ident == "" {
		ident = "无身份集合"
	}
	return fmt.Sprintf("v3 %s（%d 条 / %d 类身份）", ident, f.Total, f.IdentUniq)
}

// FPDiff 是两张表的差异。**方向是有意的**：live 是"现在这次会话"，
// rec 是"录制件"。闸门要回答的是"现在这个界面，能不能用那份录制来驱动"。
//
// 前六个字段是 v2 的逐条口径（编号对不对得上），v3 起**只作诊断**：
// 判决只看 IdentJaccard / EdgeIso / TypeDiff 三个数 + 后面那份差异清单。
type FPDiff struct {
	Same      int `json:"same"`      // 坐标与编号都对得上
	Differ    int `json:"differ"`    // 坐标对得上、编号不同（跨路径时的主症状）
	OnlyNew   int `json:"onlyNew"`   // 只在当前会话里（录制件里没有这条坐标）
	OnlyRec   int `json:"onlyRec"`   // 只在录制件里（多半是"录制时后来才出现的对象"）
	RecTotal  int `json:"recTotal"`  // 录制件表里有多少条
	LiveTotal int `json:"liveTotal"` // 当前会话表里有多少条

	// A21 判据的实测值（GateVerdict 的报文与 Report 都要报这些数，不报百分比黑箱）。
	//
	// 方向是**单向**的：live→rec（"现在这个界面，能不能用这份录制驱动"）。
	LiveCoverage float64 `json:"liveCoverage"` // ② live 的每个身份在 rec 里都找得到吗
	EdgeCoverage float64 `json:"edgeCoverage"` // ③ live 的每条边在 rec 里都找得到吗
	ExtraShare   float64 `json:"extraShare"`   // ④ live 多余身份占比（=1-②，但契约分列两条，照样算）
	TypeDiff     float64 `json:"typeDiff"`     // ⑤ 类型直方图差异（按 live 的身份集合收窄 rec 后比）
	EvidenceOK   bool    `json:"evidenceOK"`   // ① 证据量足够（|live| ≥ max(10, 0.5×|rec|)）
	IdentRec     int     `json:"identRec,omitempty"`
	IdentLive    int     `json:"identLive,omitempty"`
	EdgeRec      int     `json:"edgeRec,omitempty"`
	EdgeLive     int     `json:"edgeLive,omitempty"`

	// 差异清单**分两栏**（A21：这是"半屏"能被读懂的关键）：
	//
	//	意外（live 有、rec 无）      = 界面对不上的**拒绝理由**
	//	尚未加载（rec 有、live 无）  = 会话渐进，正常，**只作信息，不拒**
	IdentUnexpected []string `json:"identUnexpected,omitempty"`
	IdentPending    []string `json:"identPending,omitempty"`
	EdgeUnexpected  []string `json:"edgeUnexpected,omitempty"`
	EdgePending     []string `json:"edgePending,omitempty"`
	TypeChanged     []string `json:"typeChanged,omitempty"`
}

// MatchPct 是"当前界面有多少条能在录制件里原样找到"，0–100。
// v2 的旧口径，v3 起**只作诊断**（判决不用它）。
func (d FPDiff) MatchPct() int {
	if d.LiveTotal == 0 {
		return 0
	}
	return d.Same * 100 / d.LiveTotal
}

// Fit 是"这两面有多像"的一个可比数（挑最接近的那一对连接时用）：
// 判据里最弱的那条。判决本身不靠它，只有"best 是谁"靠它。
func (d FPDiff) Fit() float64 {
	fit := d.LiveCoverage
	if d.EdgeCoverage < fit {
		fit = d.EdgeCoverage
	}
	// 类型直方图差异是"越小越好"，换算成"越大越好"再比。
	fitType := 1 - d.TypeDiff
	if fitType < fit {
		fit = fitType
	}
	return fit
}

// EntriesOf 把一个模型摊成编号表。**每一份对象都进表**（不只可操作控件）：
// 跨路径实测的 1556 条就是这个口径。
func EntriesOf(m *Model) []FPEntry {
	if m == nil {
		return nil
	}
	objs := m.Objects()
	out := make([]FPEntry, 0, len(objs))
	for _, o := range objs {
		out = append(out, FPEntry{
			Kind: o.Type, Scope: o.Scope, ID: o.ID, Role: fpRole(o),
			ColName: strings.TrimSpace(o.Attr("colName")),
			FieldID: strings.TrimSpace(o.Attr("fieldId")),
			Parent:  o.Parent,
		})
	}
	sortEntries(out)
	return out
}

// fpRole 给一条编号取一个**跨路径稳定**的名字。
//
// 顺序是有意的：name 是源码里的控件名（最稳），comment 是 cmt_<表><栏位>（次稳），
// 两样都没有就留空。它不进 v3 判决（判决只用身份/边/类型），只进 idDigest 与报文。
func fpRole(o *Object) string {
	for _, k := range []string{"name", "comment"} {
		if v := strings.TrimSpace(o.Attr(k)); v != "" {
			return v
		}
	}
	return ""
}

func sortEntries(es []FPEntry) {
	sort.Slice(es, func(i, j int) bool {
		a, b := es[i], es[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Role != b.Role {
			return a.Role < b.Role
		}
		if a.Scope != b.Scope {
			return a.Scope < b.Scope
		}
		return a.ID < b.ID
	})
}

// fpIdentOf 返回一条编号的**身份串**（A19）：`colName#fieldId`。
//
// 缺一半（或两样都缺）⇒ 无身份：这样的节点**不进**身份多重集（A19 的身份就是
// colName#fieldId，没有别的东西可以冒充），在引用边里退化为占位（见 fpPlaceholders）。
func fpIdentOf(e FPEntry) (string, bool) {
	col, fid := strings.TrimSpace(e.ColName), strings.TrimSpace(e.FieldID)
	if col == "" || fid == "" {
		return "", false
	}
	return col + "#" + fid, true
}

// fpPlaceholders 给每条编号算"在引用边里怎么称呼它"：有身份就用身份，
// 没有身份就按 A19 退化为 `type#占位`。
//
// 占位优先用 `type#Role`：Role 取协议 name/comment，是源码里的名字，跨启动路径稳定。
// 连 Role 都没有（裸控件）才用 `type#~序号` —— 序号是同一 type 里按表的顺序数出来的，
// 只在这条退无可退的退化路上出现，且它不含编号（编号进边会立刻把 v3 拉回 v2）。
func fpPlaceholders(es []FPEntry) []string {
	out := make([]string, len(es))
	counter := make(map[string]int, 4)
	for i, e := range es {
		if id, ok := fpIdentOf(e); ok {
			out[i] = id
			continue
		}
		if role := strings.TrimSpace(e.Role); role != "" {
			out[i] = e.Kind + "#" + role
			continue
		}
		out[i] = e.Kind + "#~" + strconv.Itoa(counter[e.Kind])
		counter[e.Kind]++
	}
	return out
}

// identityMultiset 是**身份多重集**（A19 第一件套）：全部带身份的节点的
// `colName#fieldId`，排序后返回（保留重复 —— 张数也是信息）。
func identityMultiset(es []FPEntry) []string {
	out := make([]string, 0, len(es))
	for _, e := range es {
		if id, ok := fpIdentOf(e); ok {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// edgeMultiset 是**引用边集合**（A19 第二件套）：每条父子边写成「父身份→子身份」。
//
// 父身份在本张表里按编号找（同一个编号有多份时优先同作用域的那一份 —— 模型会复用裸 id）；
// 父根本不在表里（读不全的录制件）就写 `parent#missing`，**不写编号**。
func edgeMultiset(es []FPEntry) []string {
	names := fpPlaceholders(es)
	byID := make(map[int][]int, len(es))
	for i, e := range es {
		byID[e.ID] = append(byID[e.ID], i)
	}
	out := make([]string, 0, len(es))
	for i, e := range es {
		if e.Parent == 0 {
			continue // 没有父：不是一条边
		}
		parent := "parent#missing"
		if cand := byID[e.Parent]; len(cand) > 0 {
			best := cand[0]
			for _, j := range cand {
				if es[j].Scope == e.Scope {
					best = j
					break
				}
			}
			parent = names[best]
		}
		out = append(out, parent+"→"+names[i])
	}
	sort.Strings(out)
	return out
}

// typeHistogram 是**类型直方图**（A19 第三件套）：节点类型 → 个数。
func typeHistogram(es []FPEntry) map[string]int {
	h := make(map[string]int, 8)
	for _, e := range es {
		h[e.Kind]++
	}
	return h
}

// typeHistString 把直方图写成**规范串**（按类型名排序）：可比较、可读、可进 JSON。
func typeHistString(h map[string]int) string {
	if len(h) == 0 {
		return ""
	}
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+strconv.Itoa(h[k]))
	}
	return strings.Join(parts, " ")
}

// parseTypeHist 是 typeHistString 的逆运算（诊断/展示用）。
func parseTypeHist(s string) map[string]int {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	out := map[string]int{}
	for _, part := range strings.Fields(s) {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		out[k] = n
	}
	return out
}

// fpHashOf 是"一串规范化的行"的哈希（空 ⇒ 空串：**没有**≠"空串的 sha256"）。
func fpHashOf(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	h := sha256.New()
	for _, l := range lines {
		fmt.Fprintf(h, "%s\n", l)
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// FingerprintOf 算一份编号表的指纹（v3）。
//
// 先排序再哈希：指纹要的是"这张表长什么样"，不是"按什么顺序念出来"，
// 这样它对读数顺序不敏感（同一条路径两次录制的消息切分可以不同）。
// `Hash` 仍然是旧算法（idDigest），**只存不判**。
func FingerprintOf(entries []FPEntry) Fingerprint {
	if len(entries) == 0 {
		// 空表就是**没有**指纹，不是"空串的 sha256" —— 后者看起来像个真指纹。
		// 实测会撞上：只录到握手、没录到界面的那条连接就能算出 `e3b0c44298fc1c14`。
		return Fingerprint{}
	}
	sorted := append([]FPEntry(nil), entries...)
	sortEntries(sorted)
	h := sha256.New()
	for _, e := range sorted {
		fmt.Fprintf(h, "%s|%s|%d|%d\n", e.Kind, e.Role, e.Scope, e.ID)
	}
	sum := hex.EncodeToString(h.Sum(nil))

	idents := identityMultiset(sorted)
	edges := edgeMultiset(sorted)
	hist := typeHistogram(sorted)
	return Fingerprint{
		Total:     len(sorted),
		Hash:      sum[:16], // idDigest：诊断用，判决不看它
		Version:   FPVersion,
		Ident:     fpHashOf(idents),
		IdentN:    len(idents),
		IdentUniq: len(uniqStrings(idents)),
		Edges:     fpHashOf(edges),
		EdgeN:     len(edges),
		Types:     typeHistString(hist),
	}
}

// SameInterface 只用**三件套**判两枚指纹是不是同一个界面（全等口径）。
//
// **绝不看 Hash（idDigest）**：那是诊断。旧 v2 指纹没有三件套 ⇒ 判不了（false），
// 不许"两边都空所以相等"地放行。
func SameInterface(a, b Fingerprint) bool {
	if !a.IsV3() || !b.IsV3() {
		return false
	}
	return a.Ident == b.Ident && a.Edges == b.Edges && a.Types == b.Types
}

// JudgeSameInterface 是**唯一判决入口**（A21 非对称覆盖，取代 A19 的双向 Jaccard）。
//
//	通过 ⟺ ① 证据量足够：|live| ≥ max(10, 0.5×|rec|)
//	      且 ② live→rec 身份覆盖 ≥ 0.98   （live 的每个身份都能在 rec 里找到）
//	      且 ③ live→rec 边覆盖   ≥ 0.98
//	      且 ④ live 的多余身份 ≤ 2%
//	      且 ⑤ 类型直方图差异 ≤ 2%（按 **live 的身份集合** 在 rec 里取对应节点后比较）
//
// 方向是**单向**的（live→rec）：会话是渐进的（`an` 加 / `un` 改 / `rn` 删），
// `rec ⊃ live`（录制件是整会话、当前只开半屏）是**常态**，不能误拒 ——
// 这正是本轮要消灭的失败类。反过来 "live 有 rec 没有" 才是界面真的变了。
//
// 返回的 FPDiff 里既有五个判据的数，也有**分两栏**的差异清单（意外 / 尚未加载）。
func JudgeSameInterface(rec, live []FPEntry) (FPDiff, bool) {
	recIdent, liveIdent := identityMultiset(rec), identityMultiset(live)
	recEdges, liveEdges := edgeMultiset(rec), edgeMultiset(live)

	// 逐条口径（v2）保留下来只作诊断：它能说"编号不同"这种号码层面的事。
	d := CompareFP(rec, live)
	d.IdentRec, d.IdentLive = len(recIdent), len(liveIdent)
	d.EdgeRec, d.EdgeLive = len(recEdges), len(liveEdges)

	// ① 证据量：|live| ≥ max(10, 0.5×|rec|)（节点数）。
	// 为什么必须有这条：否则"只有 3 个节点"的残缺界面也能拿 100% 覆盖把闸门骗开。
	floor := FPEvidenceRecShare * float64(len(rec))
	if floor < FPEvidenceMinNodes {
		floor = FPEvidenceMinNodes
	}
	d.EvidenceOK = float64(len(live)) >= floor

	// ② live→rec 身份覆盖；④ live 多余身份占比（④ 是 ② 的补，契约分列两条，各自算）。
	d.LiveCoverage = coverage(liveIdent, recIdent)
	d.ExtraShare = 1 - d.LiveCoverage

	// ③ live→rec 边覆盖。
	d.EdgeCoverage = coverage(liveEdges, recEdges)

	// ⑤ 类型直方图：把 rec **收窄到 live 的身份集合**再比（尚未加载的节点不参与）。
	liveSet := make(map[string]bool, len(liveIdent))
	for _, id := range liveIdent {
		liveSet[id] = true
	}
	recNarrow := restrictToIdentities(rec, liveSet)
	recHist, liveHist := typeHistogram(recNarrow), typeHistogram(live)
	d.TypeDiff = typeHistDiff(recHist, liveHist)
	d.TypeChanged = typeHistDiffList(recHist, liveHist, fpDiffLimit)

	// 差异清单**分两栏**：rec-only = 尚未加载（信息），live-only = 意外（拒的理由）。
	d.IdentPending, d.IdentUnexpected = multisetDiff(recIdent, liveIdent, fpDiffLimit)
	d.EdgePending, d.EdgeUnexpected = multisetDiff(recEdges, liveEdges, fpDiffLimit)

	ok := d.EvidenceOK &&
		d.LiveCoverage >= FPIdentJaccardMin &&
		d.EdgeCoverage >= FPEdgeIsoMin &&
		d.ExtraShare <= FPTypeHistMaxDiff &&
		d.TypeDiff <= FPTypeHistMaxDiff
	return d, ok
}

// coverage 是**单向覆盖**（多重集）：a 里有几条能在 b 里找到 / |a|。
// a 为空 ⇒ 1（没有东西要覆盖；"证据够不够"是 ① 的事）。
func coverage(a, b []string) float64 {
	if len(a) == 0 {
		return 1
	}
	ca, cb := countOf(a), countOf(b)
	inter := 0
	for k, n := range ca {
		if m := cb[k]; m > 0 {
			if n < m {
				inter += n
			} else {
				inter += m
			}
		}
	}
	return float64(inter) / float64(len(a))
}

// restrictToIdentities 把表收窄到"身份在 keep 里"的节点；**无身份的节点（容器/裸控件）
// 照原样留着** —— 它们两边都有，剔掉会让类型直方图失真。
func restrictToIdentities(es []FPEntry, keep map[string]bool) []FPEntry {
	out := make([]FPEntry, 0, len(es))
	for _, e := range es {
		id, ok := fpIdentOf(e)
		if !ok || keep[id] {
			out = append(out, e)
		}
	}
	return out
}

// LegacyFingerprintNotice 说清"这枚指纹是旧 v2 的，已忽略"（A19/A21：读到就忽略并记诊断，
// 不许因此崩、也不许因此拒或放）。
func LegacyFingerprintNotice(f Fingerprint) string {
	if f.IsV3() || f.Hash == "" {
		return ""
	}
	return fmt.Sprintf("旧 v2 指纹已忽略（total=%d hash=%s）：v3 的判决只用身份集合/引用边/类型直方图，"+
		"idDigest 只存不判", f.Total, f.Hash)
}

func uniqStrings(ss []string) []string {
	if len(ss) == 0 {
		return nil
	}
	out := make([]string, 0, len(ss))
	var last string
	for i, s := range ss {
		if i == 0 || s != last {
			out = append(out, s)
			last = s
		}
	}
	return out
}

func countOf(ss []string) map[string]int {
	out := make(map[string]int, len(ss))
	for _, s := range ss {
		out[s]++
	}
	return out
}

// typeHistDiff 是"类型直方图差异"：½·Σ|Δ类| / max(两边节点数)
// ——也就是"**变了类型的节点占多少比例**"（两个分布的总变差距离）。
//
// rec 一侧是**按 live 的身份集合收窄后**的直方图（A21 ⑤）：尚未加载的节点不参与比较。
// 分母取较多的一侧；一个节点换类型只算一次（不因为"这类少 1、那类多 1"被算成两处）。
func typeHistDiff(rec, live map[string]int) float64 {
	recN, liveN := 0, 0
	for _, n := range rec {
		recN += n
	}
	for _, n := range live {
		liveN += n
	}
	den := recN
	if liveN > den {
		den = liveN
	}
	if den == 0 {
		return 0
	}
	sum := 0
	for t, n := range rec {
		diff := n - live[t]
		if diff < 0 {
			diff = -diff
		}
		sum += diff
	}
	for t, n := range live {
		if _, ok := rec[t]; !ok {
			sum += n
		}
	}
	return float64(sum) / 2 / float64(den)
}

// multisetDiff 列出"录制件有而当前没有"（missing）与"当前多出来"（extra）的条目，
// 按名字排序、最多 limit 条（超出的折成一句"… 等 N 条"）。
func multisetDiff(rec, live []string, limit int) (missing, extra []string) {
	cr, cl := countOf(rec), countOf(live)
	missing = diffSide(cr, cl, limit)
	extra = diffSide(cl, cr, limit)
	return missing, extra
}

// listTrueCount 返回"清单**本来**有多少条"：截断标记要把被折掉的 N 加回去。
//
// 为什么需要它：`diffSide` / `typeHistDiffList` 在超过 limit 时会把余下的折成一句
// `… 等 N 条`，而**那一句自己也算一个元素**。判决行若直接写 `len(list)`，就会
// "把标记算成一条、又把折掉的 N 漏掉" —— 真机出现过「明细 13 条、尾部写着 118 条」
// 的错位（台账 D31：人读会以为漏了 100 多条）。
func listTrueCount(list []string) int {
	n := len(list)
	if n == 0 {
		return 0
	}
	const p = "… 等 "
	last := list[n-1]
	if !strings.HasPrefix(last, p) {
		return n
	}
	rest := strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(last, p), " 条"), " 类")
	if k, err := strconv.Atoi(rest); err == nil {
		return n - 1 + k
	}
	return n
}

func diffSide(a, b map[string]int, limit int) []string {	keys := make([]string, 0, len(a))
	for k, n := range a {
		if n > b[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if len(out) >= limit {
			out = append(out, fmt.Sprintf("… 等 %d 条", len(keys)-limit))
			break
		}
		deficit := a[k] - b[k]
		if deficit > 1 {
			out = append(out, fmt.Sprintf("%s ×%d", k, deficit))
			continue
		}
		out = append(out, k)
	}
	return out
}

// typeHistDiffList 列出"哪些类节点数变了"，形如 `Label 50 → 47`。
func typeHistDiffList(rec, live map[string]int, limit int) []string {
	kinds := make([]string, 0, len(rec)+len(live))
	for k := range rec {
		kinds = append(kinds, k)
	}
	for k := range live {
		if _, ok := rec[k]; !ok {
			kinds = append(kinds, k)
		}
	}
	sort.Strings(kinds)
	out := make([]string, 0, len(kinds))
	for _, k := range kinds {
		if rec[k] == live[k] {
			continue
		}
		if len(out) >= limit {
			out = append(out, fmt.Sprintf("… 等 %d 类", len(kinds)-limit))
			break
		}
		out = append(out, fmt.Sprintf("%s %d → %d", k, rec[k], live[k]))
	}
	return out
}

// CompareFP 逐条比对两张表（rec = 录制件，live = 当前会话）。
//
// v3 起它**只作诊断**（判决走 JudgeSameInterface）：哈希只回答"一样不一样"，
// 而闸门拒绝时必须说清"差在哪" —— 逐条比才是那句话的来源。
func CompareFP(rec, live []FPEntry) FPDiff {
	d := FPDiff{RecTotal: len(rec), LiveTotal: len(live)}
	recExact := make(map[FPEntry]int, len(rec))
	recKeys := make(map[FPKey]int, len(rec))
	for _, e := range rec {
		recExact[e]++
		recKeys[e.key()]++
	}
	for _, e := range live {
		switch {
		case recExact[e] > 0:
			recExact[e]--
			d.Same++
		case recKeys[e.key()] > 0:
			d.Differ++
		default:
			d.OnlyNew++
		}
	}
	for _, n := range recExact {
		d.OnlyRec += n
	}
	return d
}

// TranslateEntries 把两张表里"坐标对得上但编号不同"的前几条翻成人话。
//
// 这是拒绝报文里最有用的一段：跨路径时它会直说"录制件里 #327 是那个东西，
// 现在 #327 是这个东西" —— 它同时是"v3 为什么不看编号"的现场证据。
func TranslateEntries(rec, live []FPEntry, limit int) []string {
	byKeyNew := map[FPKey][]FPEntry{}
	for _, e := range live {
		byKeyNew[e.key()] = append(byKeyNew[e.key()], e)
	}
	recByKey := map[FPKey][]FPEntry{}
	for _, e := range rec {
		recByKey[e.key()] = append(recByKey[e.key()], e)
	}
	var keys []FPKey
	for k, es := range recByKey {
		live := byKeyNew[k]
		if len(live) == 0 {
			continue
		}
		if !sameIDs(es, live) {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Kind != keys[j].Kind {
			return keys[i].Kind < keys[j].Kind
		}
		return keys[i].Role < keys[j].Role
	})
	var out []string
	for _, k := range keys {
		if len(out) >= limit {
			break
		}
		out = append(out, fmt.Sprintf("%s %s（作用域 %d）：录制件编号 %s → 现在 %s",
			k.Kind, dashIfEmpty(k.Role), k.Scope, idList(recByKey[k]), idList(byKeyNew[k])))
	}
	return out
}

func sameIDs(a, b []FPEntry) bool {
	if len(a) != len(b) {
		return false
	}
	x := make([]int, 0, len(a))
	y := make([]int, 0, len(b))
	for i := range a {
		x = append(x, a[i].ID)
	}
	for i := range b {
		y = append(y, b[i].ID)
	}
	sort.Ints(x)
	sort.Ints(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func idList(es []FPEntry) string {
	ids := make([]int, 0, len(es))
	for _, e := range es {
		ids = append(ids, e.ID)
	}
	sort.Ints(ids)
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, fmt.Sprintf("#%d", id))
	}
	return strings.Join(parts, ",")
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "（无名）"
	}
	return s
}
