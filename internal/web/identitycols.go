package web

// 录制明细页的「身份 vs 句柄」三列（契约 C3 + A7/A9/A10/A11/A12/A15/A18 的消费面）。
//
// 一行要同时说清三件事，顺序固定（界面照 identityColumns 排，不自己猜顺序）：
//
//	colName    身份·稳        节点的 Attr("colName")   —— 编译期分配，表单内稳定
//	fieldId    表单内编号·稳   节点的 Attr("fieldId")   —— 与 colName 成对出现（`<colName>#<fieldId>` 双键）
//	sessionId  句柄·会复用     事件里的 idRef          —— 单步有效，LIFO 复用，rn 之后全量失效（H1/H2）
//
// 为什么要摆在一起：会话 id 是**临时句柄**，人看不出"这是哪个字段"（同一个 3541 前一生是工具栏项、
// 后一生是明细列 `formonly.check`）。而 colName/fieldId 是编译期分配的身份，跨会话稳定 ——
// 认字段只看前两列，号码只用来在**这一步**里唯一定位。
//
// 三条纪律（都是契约里现成的口径，改之前先读 internal/fep/identity.go 的头注释）：
//
//  1. **只读节点属性，不拼不猜**（A15/A18）：取不到就写"未解析"，绝不自己造一个
//     （例如 `w_`+剥后缀 那种启发式）。双键缺一半也算"未解析"，不拿另一半顶上。
//  2. **身份取的是"这一步发生时"的那个节点**，不是录制末尾那棵树：裸 id 会被复用
//     （实测 `rn` 之后同一 id 换成另一个对象），拿末尾回填等于把后来的节点安到早先的动作上。
//     排先后只用帧级时间（跨方向唯一的共同坐标）；没有帧级时间就退回末尾整棵树，
//     并**明说**这一点（identityNote），不装作时间能排。
//  3. **不改 internal/fep**：数据全从既有的 `Attr`/`Get`/`Model` 读，模型由读侧（recordingread.go）建。
//
// 旧键一个不动：三列挂在每行的 `identity` 上，`target`/`frame`/`events` 等照原样留着。

import (
	"fmt"
	"sort"
	"strings"

	"tt/internal/fep"
)

// identityUnresolved 是"取不到身份"时的**唯一**文案。
//
// 为什么不是空串：空串在界面上与"解出来了但是空的"分不开，而这两种情形对用户的**下一步**
// 完全不同 —— 前者要去查这条流里到底下发了什么，后者是数据本来就空。契约 A15/A18 也是这条：
// 取不到就是不认识，不许静默降级。
const identityUnresolved = "未解析"

// identityCols 是一行上的三列（JSON 键：colName / fieldId / sessionId）。
//
// Known 两个布尔是**机器可读的那一半**：界面拿它选样式（灰掉/加提示），
// **不许**去匹配 colName == "未解析" 这种文案（文案会改，判定不该跟着改）。
type identityCols struct {
	// ColName 是身份第一列：节点的 `colName`（或 "未解析"）。
	ColName string `json:"colName"`
	// FieldID 是身份第二列：节点的 `fieldId`（或 "未解析"）。
	//
	// 照**原文**给（字符串）：`fieldId` 可以合法为 "0"（实测 `{colName "check"} {fieldId "0"}`），
	// 所以不能拿 0 当"没给"，也不在这里解释成数字（语义归属 internal/fep）。
	FieldID string `json:"fieldId"`
	// SessionID 是第三列：事件里的 idRef —— **句柄**，会复用（同一行里它就是"号码"）。
	//
	// 不用 omitempty：`idRef=0`（纯键盘事件）与"没这个键"在界面上该长得不一样。
	SessionID int `json:"sessionId"`
	// ColNameKnown / FieldIDKnown：这两格到底解出来了没有。
	ColNameKnown bool `json:"colNameKnown"`
	FieldIDKnown bool `json:"fieldIdKnown"`
	// NodeType 是"这一刻这个 id 是什么"：复用发生时，类型最先说话（Item → TableColumn）。
	NodeType string `json:"nodeType,omitempty"`
	// Note 说清"为什么未解析"，或"身份是从哪儿取的"（末尾整棵树时要明说）。
	// **不许**留空装没事：空白格会让人以为是界面坏了。
	Note string `json:"note,omitempty"`
}

// identityColumn 是一列的自述（界面照这个顺序与文案渲染，别在两处各写一套）。
type identityColumn struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	// Stable：这一列稳不稳 —— true = 身份（跨会话稳定），false = 句柄（会复用）。
	Stable bool `json:"stable"`
	// Missing 是这一列取不到时的文案（只有身份两列有；句柄永远有值）。
	Missing string `json:"missing,omitempty"`
}

// identityColumnSpec 是**唯一**的列定义处（顺序即界面顺序：身份在前、句柄在后）。
var identityColumnSpec = []identityColumn{
	{Key: "colName", Label: "栏位名 colName（身份·稳）", Stable: true, Missing: identityUnresolved},
	{Key: "fieldId", Label: "表单内编号 fieldId（身份·稳）", Stable: true, Missing: identityUnresolved},
	{Key: "sessionId", Label: "会话 id（句柄·会复用）", Stable: false},
}

// identityHint 是页面上那一句人话（一句话，别长篇）。
const identityHint = "前两列是字段身份（编译期分配、表单内稳定），认字段只看它们；" +
	"第三列是运行期会话 id —— 单步句柄，LIFO 复用、rn 之后全量失效，别拿号码当名字。"

// 三种"身份取不到/取得不精确"的原因，分开写清（它们的下一步不同）。
const (
	identityNoteNoRuntime = "这条连接（#%d）没有录到界面侧：colName/fieldId 无从得知，号码只是句柄"
	identityNoteFinalTree = "这一步没有帧级时间（索引缺失或由老件重建）：身份取自**录制末尾**那棵整树" +
		"—— 期间若有 id 复用，显示的可能是后来的那个节点"
)

// identityRows 给每个动作算出三列，下标与 d.Actions 对齐。
//
// 做法：按连接把界面侧消息**只走一遍**，走到这一步的时刻为止，然后读这个 id 现在是谁。
// 一遍而不是"每个动作重建一次模型"：真录制 383 帧、17 步，重建 17 次是白烧 CPU；
// 而两条规则（时刻不晚于这一步的算已发生、同刻按字节序）本来就允许单遍推进。
func identityRows(d *recRead) []identityCols {
	out := make([]identityCols, len(d.Actions))
	if d == nil {
		return out
	}
	byConn := map[int][]int{}
	for i, a := range d.Actions {
		out[i] = identityCols{ColName: identityUnresolved, FieldID: identityUnresolved, SessionID: a.ID}
		byConn[a.Conn] = append(byConn[a.Conn], i)
	}

	for conn, idxs := range byConn {
		rd := d.dirOf(conn, fep.DirRuntime)
		if rd == nil {
			for _, i := range idxs {
				out[i].Note = fmt.Sprintf(identityNoteNoRuntime, conn)
			}
			continue
		}
		// 可回溯的前提：这一侧的每一帧、以及每一步自己，都有帧级时刻。
		// 少一个就只能退回末尾整棵树 —— 那是有代价的，所以照实说（见 identityNoteFinalTree）。
		if !identityTimed(rd, d.Actions, idxs) {
			for _, i := range idxs {
				out[i] = identityFrom(rd.model, d.Actions[i].ID)
				out[i].Note = joinNotes(identityNoteFinalTree, out[i].Note)
			}
			continue
		}
		order := append([]int(nil), idxs...)
		sort.SliceStable(order, func(x, y int) bool {
			return d.Actions[order[x]].Frame.T < d.Actions[order[y]].Frame.T
		})
		msgs := append([]recFrame(nil), rd.Frames...)
		sort.SliceStable(msgs, func(x, y int) bool { return msgs[x].T < msgs[y].T })

		m := fep.NewModel()
		k := 0
		for _, i := range order {
			at := d.Actions[i].Frame.T
			for k < len(msgs) && msgs[k].T <= at {
				if msgs[k].Kind != "cache" && msgs[k].Text != "" {
					m.ApplyMessage(msgs[k].Text)
				}
				k++
			}
			out[i] = identityFrom(m, d.Actions[i].ID)
		}
	}
	return out
}

// identityTimed 判断"这一侧够不够排先后"：界面侧的每一帧与每一步都要有帧级时刻。
func identityTimed(rd *recDir, acts []*recAction, idxs []int) bool {
	if rd == nil || rd.model == nil {
		return false
	}
	for _, fr := range rd.Frames {
		if fr.T == 0 {
			return false
		}
	}
	for _, i := range idxs {
		if acts[i].Frame.T == 0 {
			return false
		}
	}
	return true
}

// identityFrom 从**此刻**的界面树读出三列。
//
// 只读节点属性：`Attr("colName")` / `Attr("fieldId")`。取不到就是"未解析"（A15/A18），
// 双键缺一半也只写那一半缺（不拿另一半顶上，也不从 `name` 后缀去拼 —— 那是启发式）。
func identityFrom(m *fep.Model, id int) identityCols {
	out := identityCols{ColName: identityUnresolved, FieldID: identityUnresolved, SessionID: id}
	if id == 0 {
		out.Note = "这条事件里没有 idRef：它不指向任何对象（例如纯键盘事件），没有身份可言"
		return out
	}
	if m == nil {
		out.Note = "没有可用的界面树：colName/fieldId 无从得知"
		return out
	}
	o, ok := m.Get(id)
	if !ok {
		out.Note = fmt.Sprintf("这一刻的界面树里没有 #%d：它还没下发，或已被 rn 删掉"+
			"（号码指向的是一个不在场的对象）", id)
		return out
	}
	out.NodeType = o.Type
	if v := strings.TrimSpace(o.Attr("colName")); v != "" {
		out.ColName, out.ColNameKnown = v, true
	}
	if v := strings.TrimSpace(o.Attr("fieldId")); v != "" {
		out.FieldID, out.FieldIDKnown = v, true
	}
	switch {
	case !out.ColNameKnown && !out.FieldIDKnown:
		out.Note = fmt.Sprintf("这个节点（%s）上没有 colName/fieldId：它不是字段"+
			"（容器/动作/树项就是这样下发的），身份未解析 —— 号码只能当句柄用", o.Type)
	case !out.ColNameKnown:
		out.Note = "这个节点上只有 fieldId、没有 colName：双键缺一半，不猜另一半（C3）"
	case !out.FieldIDKnown:
		out.Note = "这个节点上只有 colName、没有 fieldId：双键缺一半，不猜另一半（C3）"
	}
	return out
}

// joinNotes 把两条原因接成一句（都为空时回空串 —— 界面据此决定要不要显示提示行）。
func joinNotes(a, b string) string {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	switch {
	case a == "":
		return b
	case b == "":
		return a
	}
	return a + "；" + b
}
