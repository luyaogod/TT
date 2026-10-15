package fep

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// 本文件是**录制件的身份**：不拿 id 寻址，拿整张"坐标 → 编号"表当签名。
//
// 判据全部来自实测（出处见 .agents/plan/tt-agent-field-notes.md §9）：
//
//   - AUI 的对象编号由**运行期**按声明顺序分配，所以 id 是**启动历史的函数**。
//     同一套界面换一条启动路径：对象集几乎一样（1556 / 1558），而共同条目的编号
//     **一条都不相同**（偏移 −31 到 −4274，不是等量平移）。
//   - 同一条路径逐条可复现：独立会话各跑三次，关键字段恒为 281/290/327/495；
//     调试线路径的编号表 2627 条逐条相同、0 差异。
//
// 于是 id 有两种用法，必须分开：
//
//	用它**寻址**（回放时发事件）—— 只在同一轮会话内成立，那是"安全闸"管的事；
//	用它**当签名**（这里）—— 整张表恰好就是"哪条路径、哪次编译"的签名。
//
// 所以指纹回答的是"这份录制是在什么界面下录的"，而**不是**"这个 id 现在是谁"。
// 跨路径回放之所以要拒，正是因为实测里安全闸拦不住它：6/6 全部落到别的对象上
// （#327 从「加工类型编号」变成「重测/返工否」的勾选框）**且不报错**。

// FPEntry 是编号表里的一条坐标：**这是什么、在谁下面、编号是多少、叫什么**。
//
// Role 取协议 name（源码里的控件名）→ comment（cmt_<表><栏位>）→ 类型名。
// **不含 text / value**：那是业务数据，同一条路径两次录制里必然不同
// （实测同一作业连录两轮，单号从 …003 变成 …004），拿它当签名会把"能用的录制"
// 判成失效。指纹里因此也不带任何业务值 —— 它可以随便复制流转。
type FPEntry struct {
	Kind  string `json:"kind"`
	Scope int    `json:"scope,omitempty"`
	ID    int    `json:"id"`
	Role  string `json:"role,omitempty"`
}

// FPKey 是"同一张表里的同一条坐标"（不含编号）：编号不同的就是**不同**的表。
type FPKey struct {
	Kind  string
	Scope int
	Role  string
}

func (e FPEntry) key() FPKey { return FPKey{Kind: e.Kind, Scope: e.Scope, Role: e.Role} }

// Fingerprint 是一份录制件的界面指纹：整张编号表的摘要。
type Fingerprint struct {
	Total int    `json:"total"` // 表里有多少条
	Hash  string `json:"hash"`  // 整张表的 sha256 前 16 位
}

// String 给人读的形态（跨路径比对时一眼能看出不是同一张表）。
func (f Fingerprint) String() string {
	if f.Hash == "" {
		return "（没有指纹）"
	}
	return fmt.Sprintf("%s（%d 条）", f.Hash, f.Total)
}

// FPDiff 是两张编号表的差异。**方向是有意的**：live 是"现在这次会话"，
// rec 是"录制件"。闸门要回答的是"现在这个界面，能不能用那份录制来驱动"。
type FPDiff struct {
	Same      int `json:"same"`      // 坐标与编号都对得上
	Differ    int `json:"differ"`    // 坐标对得上、编号不同（跨路径时的主症状）
	OnlyNew   int `json:"onlyNew"`   // 只在当前会话里（录制件里没有这条坐标）
	OnlyRec   int `json:"onlyRec"`   // 只在录制件里（多半是"录制时后来才出现的对象"）
	RecTotal  int `json:"recTotal"`  // 录制件表里有多少条
	LiveTotal int `json:"liveTotal"` // 当前会话表里有多少条
}

// MatchPct 是"当前界面有多少条能在录制件里原样找到"，0–100。
func (d FPDiff) MatchPct() int {
	if d.LiveTotal == 0 {
		return 0
	}
	return d.Same * 100 / d.LiveTotal
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
		out = append(out, FPEntry{Kind: o.Type, Scope: o.Scope, ID: o.ID, Role: fpRole(o)})
	}
	sortEntries(out)
	return out
}

// fpRole 给一条编号取一个**跨路径稳定**的名字。
//
// 顺序是有意的：name 是源码里的控件名（最稳），comment 是 cmt_<表><栏位>（次稳），
// 两样都没有就退回类型名 —— 那会让这一条退化成"只有类型+作用域+编号"，
// 但这仍然够用：跨路径时编号本身就全都不一样。
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

// FingerprintOf 算一张表的指纹。
//
// 先排序再哈希：指纹要的是"这张表长什么样"，不是"按什么顺序念出来"，
// 这样它对读数顺序不敏感（同一条路径两次录制的消息切分可以不同）。
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
	return Fingerprint{Total: len(sorted), Hash: sum[:16]}
}

// CompareFP 逐条比对两张表（rec = 录制件，live = 当前会话）。
//
// 为什么逐条而不是只比哈希：哈希只回答"一样不一样"，而闸门拒绝时必须说清
// **差在哪、差多少** —— 只报一句"指纹不符"的话，人只能靠猜。
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
// 现在 #327 是这个东西" —— 实测的 6/6 误点就是这样被看出来的。
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
