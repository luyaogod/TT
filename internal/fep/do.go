package fep

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tt/internal/output"
)

// 动作面：把"点一下/打个字"变成一条事件，先校验再发。
//
// 跨进程：常驻会话在另一个进程里，本层通过**注入口**（回环 TCP）与它说话，
// AUI 状态则从**录制件**里重建（运行期不会回答"当前界面是什么"，它只会推 om）。
// 所以这一层没有任何内存里的会话状态，进程重启后照样能干活。

// ConnStatus is one live connection as the resident process sees it.
type ConnStatus struct {
	ID      int    `json:"id"`
	Seq     int    `json:"seq"`
	Alive   bool   `json:"alive"`
	Remote  string `json:"remote,omitempty"`
	BytesRT int64  `json:"bytesRT,omitempty"`
	BytesFE int64  `json:"bytesFE,omitempty"`
}

// Client talks to a resident session.
type Client struct {
	InjectAddr string // 127.0.0.1:<port>
	RecordDir  string
	Tag        string
	// Window 可选：只在某个窗口里找目标（`--window`）。
	//
	// 取值 = 窗口名**原样**，口径就是 `m.WindowName`（契约 A12 精确化 / A18）：它可能是
	// 窗口标题形态（`单据别编号数据查询(q_ooba002_1)`），也可能为空串 —— **不是**干净的
	// AUI 名（那个是窗口对象的 `name` 属性，如 `w_qry_q_ooba002_1`）。
	// 判定与 `Scope.Window` 完全同口径（identity.go 的 windowNamed）：精确相等才算命中；
	// 不做前缀/后缀/大小写/子串模糊匹配，也不用 `w_` + 剥 `_wf/_tui/_gui` 之类的启发式
	// ——那些只许进诊断（A18）。
	Window string
}

// Status lists the live connections (and, per connection, the next `_om` number).
func (c *Client) Status() ([]ConnStatus, error) {
	reply, err := c.roundTrip("status")
	if err != nil {
		return nil, err
	}
	var out struct {
		Conns []ConnStatus `json:"conns"`
	}
	if err := json.Unmarshal([]byte(reply), &out); err != nil {
		return nil, fmt.Errorf("fep: 注入口回了看不懂的东西: %q", reply)
	}
	return out.Conns, nil
}

// Send injects one event on one connection. The resident process numbers it.
func (c *Client) Send(connID int, event string) (string, error) {
	reply, err := c.roundTrip(fmt.Sprintf("@%d %s", connID, strings.TrimSpace(event)))
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(reply, "ERR") {
		return reply, fmt.Errorf("fep: 注入口拒绝: %s", strings.TrimPrefix(reply, "ERR "))
	}
	return reply, nil
}

func (c *Client) roundTrip(line string) (string, error) {
	if c.InjectAddr == "" {
		return "", fmt.Errorf("fep: 没有注入口地址（会话登记里没有 injectPort）")
	}
	conn, err := net.DialTimeout("tcp", c.InjectAddr, 3*time.Second)
	if err != nil {
		return "", fmt.Errorf("fep: 连不上注入口 %s: %w", c.InjectAddr, err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(10 * time.Second))
	if _, err := conn.Write([]byte(line + "\n")); err != nil {
		return "", err
	}
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil && reply == "" {
		return "", err
	}
	return strings.TrimSpace(reply), nil
}

// DoConfig describes one action.
type DoConfig struct {
	Kind    string // click | key | input
	Target  string // 名字或 #id（click/input 必填）
	Value   string // input
	KeyName string // key
	Conn    int    // 0 = 自动（给了 --window 就按窗口路由；只有一条连接时用它）
	DryRun  bool
	Wait    time.Duration // 发完等运行期回包多久（0 = 默认 10s；负数 = 不等）
	// RawEvent 直接指定要发的**事件体**（可含多个事件），用于像前端那样把
	// 多个事件放进同一条消息（实测：写格子 = 值 + 光标 同一条，拆开发不生效）
	RawEvent string
	// Commit 是"写完这个栏位后，主动离开它"要发的键（空 = 不离开）。
	//
	// 为什么必须离开（真机语料）：写值本身**只改显示值**，AFTER FIELD（校验、
	// 自动带值、下一栏位的联动）只在**光标离开那一刻**才跑。实测的真实序列：
	//
	//	_om 4  ConfigureEvent{idRef "271"}{ value "CNJ-A50-261000000099"}
	//	_om 5  ActionEvent{idRef "239"}          ← 立刻离开栏位
	//
	// 键盘离开栏位就是 `KeyEvent "Tab"`（语料里大量出现）。不离开的后果是：
	// 报告里只看得到"我写的那个值"，看不到系统自动带出来的值 —— 而那正是
	// "这个栏位填完之后系统自己做了什么"，是手册里最该写的东西。
	CommitKey string
}

// ValueChange is one field whose value changed while waiting.
type ValueChange struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
	From string `json:"from,omitempty"`
	To   string `json:"to"`
	// Kind 区分"这个值是谁改的"：
	//   sent —— 就是我这一步发过去的
	//   auto —— **运行期自己带出来的**（例如供应商编号 00000002 → 供应商名称 三洋开泰1）
	// 后者正是用户眼里"系统自己填好了"的那一步，操作手册必须能看见它。
	Kind string `json:"kind,omitempty"`
	// Label 是中文名（能从协议解析出来才填），Comment 是源码键（cmt_xxx）。
	Label   string `json:"label,omitempty"`
	Comment string `json:"comment,omitempty"`
	Role    string `json:"role,omitempty"`
}

// Changes summarises what the runtime did after an action.
// "事件发出去了"与"作业真的变了"是两件事，只有后者有用。
type Changes struct {
	RuntimeBytes int64         `json:"runtimeBytes"`
	NewMessages  int           `json:"newMessages"`
	Acked        bool          `json:"acked"`
	NewWindows   []string      `json:"newWindows,omitempty"`
	Values       []ValueChange `json:"values,omitempty"`
	TimedOut     bool          `json:"timedOut,omitempty"`
	// RuntimeError 是运行期推上来的报错（实测形态：Menu{text "ERROR"}{comment "Program stopped at …"}）。
	// 它一出现，"这一步没变化"就不是"没生效"而是"作业已经停了、后面全是空 ACK"。
	RuntimeError string `json:"runtimeError,omitempty"`
	// Tree 是**树本身的差分**（新增/删除/开关窗/动作增失）——"这一步之后树变了什么"。
	Tree *TreeDiff `json:"tree,omitempty"`
	// Focus 是**这一步之后的焦点**（谁拿着焦点、在哪个窗口/网格/第几行第几列）。
	// 它是"下一步能不能动手"的前置条件，也是手册里"光标现在在哪"的凭据。
	Focus *FocusReport `json:"focus,omitempty"`
}

// DoResult is what happened (or would happen).
type DoResult struct {
	Event   string `json:"event"`
	Conn    int    `json:"connection"`
	Seq     int    `json:"seq"`
	Target  string `json:"target"`
	Name    string `json:"name,omitempty"`
	Sent    bool   `json:"sent"`
	Reply   string `json:"reply,omitempty"`
	Refused string `json:"refused,omitempty"`
	// Committed 是"这一步之后主动离开栏位"用的键（空 = 没离开）—— 有它才有 AFTER FIELD。
	Committed string   `json:"committed,omitempty"`
	NotSent   string   `json:"notSentReason,omitempty"`
	Suggest   string   `json:"suggestion,omitempty"`
	Changes   *Changes `json:"changes,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
}

// Do runs the pre-flight checks and, unless it is a dry run, sends the event.
//
// Everything that can be checked offline is checked before a single byte is sent:
// that is the whole point of the pre-flight layer (a wrong event resets the
// connection and can take the job down).
func Do(c *Client, cfg DoConfig) (*DoResult, error) {
	conns, err := c.Status()
	if err != nil {
		return nil, err
	}
	if len(conns) == 0 {
		return nil, fmt.Errorf("fep: 会话里没有活着的连接（作业还没连上来？）")
	}
	conn, err := pickConn(conns, cfg.Conn)
	if err != nil {
		return nil, err
	}
	m, r, _, err := LoadModelFrom(c.RecordDir, c.Tag)
	if err != nil {
		return nil, err
	}

	res := &DoResult{Conn: conn.ID, Seq: conn.Seq}
	// sentID/sentValue：这一步"我自己发的"是哪个对象、什么值（用于把变化分成
	// "我写的"与"运行期自动带出来的"两类）。key/raw 形态由各自分支填。
	sentID, sentValue := 0, ""
	switch cfg.Kind {
	case "click":
		id, name, err := resolveTarget(m, r, cfg.Target, c.Window, "click")
		if err != nil {
			return nil, err
		}
		res.Target, res.Name = fmt.Sprintf("#%d", id), name
		sentID = id
		if err := CheckClick(m, id); err != nil {
			var ref *Refusal
			if asRefusal(err, &ref) {
				res.Refused, res.Suggest = ref.Message, ref.Hint
				return res, err
			}
			return nil, err
		}
		res.Event = ActionEventText(id)
	case "key":
		if err := CheckKey(cfg.KeyName); err != nil {
			return nil, err
		}
		res.Event = KeyEventText(cfg.KeyName)
	case "pick":
		id, name, err := resolveTarget(m, r, cfg.Target, c.Window, "input")
		if err != nil {
			return nil, err
		}
		itemID, itemText, perr := PickOption(m, id, cfg.Value)
		if perr != nil {
			return nil, perr
		}
		res.Target, res.Name = fmt.Sprintf("#%d", itemID), name+" ← "+itemText
		sentID, sentValue = id, itemText
		res.Event = ActionEventText(itemID)
	case "input":
		id, name, err := resolveTarget(m, r, cfg.Target, c.Window, "input")
		if err != nil {
			return nil, err
		}
		target, via, err := ResolveInputTarget(m, id)
		if err != nil {
			return nil, err
		}
		res.Target, res.Name = fmt.Sprintf("#%d", target), name
		sentID, sentValue = target, cfg.Value
		if !via {
			res.NotSent = ""
			res.Suggest = "这个字段没有 FormField 子对象，直接发给它自己；真机上若回 -6327，请改发字段的 FormField"
		}
		res.Event = ConfigureValueText(target, cfg.Value, nil)
	case "raw":
		// 直接给事件体（可含多个事件）—— 像前端那样把"写值 + 光标"放进同一条消息。
		res.Event = strings.TrimSpace(cfg.RawEvent)
		res.Target = cfg.Target
		// raw 形态（如写格子：值 + 光标同一条消息）：抠出第一个 ConfigureEvent 的
		// idRef/value，好把"我写的那个值"从一堆自动带值里认出来。
		sentID, sentValue = firstConfigureTarget(cfg.RawEvent)
		if res.Event == "" {
			return nil, fmt.Errorf("fep: raw 事件体为空")
		}
	default:
		return nil, fmt.Errorf("fep: 不认识的动作 %q（click/key/input）", cfg.Kind)
	}

	// 正式区 + 写类动作：给一行警告（不硬禁，依决策②：安全靠审计/预演/警告，不靠禁令）
	if prod, ev := ProductionEvidence(m); prod && writesData(cfg, res) {
		res.Warnings = append(res.Warnings, "目标看起来是正式区（依据："+ev+"）：这一步会改数据，审计日志已记录")
	}

	if cfg.DryRun {
		res.Sent = false
		res.NotSent = "dry-run：只校验与打印，一个字节都没发"
		return res, nil
	}
	// 记下动作前的现场：用来回答“这一步到底让作业变了什么”
	before := snapshotOf(c, conn.ID)

	reply, err := c.Send(conn.ID, res.Event)
	res.Sent = err == nil
	res.Reply = reply
	if err != nil {
		return res, err
	}
	// 写完栏位后主动离开（否则 AFTER FIELD 不跑，自动带值看不见）
	if cfg.CommitKey != "" {
		time.Sleep(400 * time.Millisecond) // 让运行期先把这次写入吃掉，再谈离开
		if _, cerr := c.Send(conn.ID, KeyEventText(cfg.CommitKey)); cerr != nil {
			res.Warnings = append(res.Warnings, "离开栏位失败："+cerr.Error())
		} else {
			res.Committed = cfg.CommitKey
		}
	}
	if cfg.Wait >= 0 {
		res.Changes = waitChanges(c, conn.ID, before, cfg.Wait, sentID, sentValue)
		// 发送后验证：作业要是被这一步搞停了，就**不要**让调用方把它当成功。
		// 实测教训：停掉之后窗口对任何事件都只回空 ACK，看起来像"这一步没生效"，
		// 于是接着改参数接着试 —— 每一步都在浪费，而且看不出错在哪。
		if res.Changes != nil && res.Changes.RuntimeError != "" {
			return res, &Refusal{
				Code:    output.CodeQueryFailed,
				Exit:    1,
				Message: fmt.Sprintf("作业程序停下了：%s", res.Changes.RuntimeError),
				Hint: "重开会话再继续（tt recording stop <录制ID> && tt recording start --launch standard --prog <作业>）；" +
					"并把这一步从重放/计划里去掉：它把值发给了容器，或发给了不属于当前交互指令的动作",
			}
		}
	}
	return res, nil
}

// snapshot is the state we compare against after an action.
type snapshot struct {
	bytesRT int64
	msgs    int
	windows map[string]bool
	// values 按 **(id, 作用域)** 归键：同一个裸 id 在不同作用域各有一份（项目早已知），
	// 只按 id 归键会把别的窗口的值当成"这一步变了"（实测：每一步都误报同一批自动带值）。
	values map[ObjKey]string
	// model 是发送**之前**的整个模型：有了它才能回答"树变了什么"（差分），
	// 而不是只能报"运行期多了几个字节"。
	model *Model
}

func snapshotOf(c *Client, connID int) snapshot {
	s := snapshot{windows: map[string]bool{}, values: map[ObjKey]string{}}
	if conns, err := c.Status(); err == nil {
		for _, cc := range conns {
			if cc.ID == connID {
				s.bytesRT = cc.BytesRT
			}
		}
	}
	m, _, n, err := LoadModelFrom(c.RecordDir, c.Tag)
	if err != nil {
		return s
	}
	s.model = m
	s.msgs = n
	for _, w := range m.Windows() {
		s.windows[w.Attr("name")] = true
	}
	for _, o := range m.Objects() {
		if v := o.Attr("value"); v != "" {
			s.values[ObjKey{o.ID, o.Scope}] = v
		}
	}
	return s
}

// waitChanges blocks until the runtime pushes something (or the wait times out)
// and summarises the difference.
//
// sentID/sentValue：这一步**我自己发的**是哪个对象、什么值 —— 用来把变化分成
// "我写的"与"运行期自动带出来的"两类（后者是手册里"系统自己填好了"那一步）。
func waitChanges(c *Client, connID int, before snapshot, wait time.Duration, sentID int, sentValue string) *Changes {
	if wait == 0 {
		wait = 10 * time.Second
	}
	ch := &Changes{}
	deadline := time.Now().Add(wait)
	for {
		settled := false
		if conns, err := c.Status(); err == nil {
			for _, cc := range conns {
				if cc.ID == connID && cc.BytesRT != before.bytesRT {
					ch.RuntimeBytes = cc.BytesRT - before.bytesRT
					settled = true
				}
			}
		}
		if settled {
			break
		}
		if time.Now().After(deadline) {
			ch.TimedOut = true
			return ch
		}
		time.Sleep(200 * time.Millisecond)
	}
	// 让它把这一屏推完再读：半截的界面看起来跟完整的一样
	time.Sleep(500 * time.Millisecond)
	m, r, n, err := LoadModelFrom(c.RecordDir, c.Tag)
	if err != nil {
		return ch
	}
	ch.NewMessages = n - before.msgs
	ch.Tree = DiffModels(before.model, m)
	if f := m.FocusNow(); f.FocusID != 0 || f.CursorID != 0 {
		ch.Focus = &f
	}
	for _, w := range m.Windows() {
		name := w.Attr("name")
		if name != "" && !before.windows[name] {
			ch.NewWindows = append(ch.NewWindows, name)
		}
	}
	for _, o := range m.Objects() {
		v := o.Attr("value")
		if v == "" {
			continue
		}
		key := ObjKey{o.ID, o.Scope}
		if old, had := before.values[key]; had && old == v {
			continue
		}
		name := ""
		if fn, ok := r.Resolve(o.ID); ok && fn.Text != "" {
			name = fn.Text
		}
		// 角色：优先协议 name（驱动寻址用），其次中文名
		role := o.Attr("name")
		if role == "" {
			role = name
		}
		kind := "auto"
		if sentID != 0 && o.ID == sentID {
			kind = "sent"
		} else if sentID != 0 && sentValue != "" && v == sentValue && name != "" {
			// 没给到同一个 id（实测：输入往往落在 FormField，而运行期回收时
			// 报的是那个 Value）—— 值一样且不是自动带的，仍归"我写的"。
			kind = "sent"
		}
		ch.Values = append(ch.Values, ValueChange{
			ID: o.ID, Name: name, Label: name, Role: role, Comment: o.Attr("comment"),
			From: before.values[key], To: v, Kind: kind,
		})
	}
	// 最关键的一条：运行期有没有报错把程序停下。
	// 它出现在这里，所有的"无变化"都该读成"作业已停"，而不是"这一步没生效"。
	for _, e := range RuntimeErrors(m) {
		if !e.ProgramStop() {
			continue
		}
		where := e.Where()
		if where == "" {
			where = "（运行期没给位置）"
		}
		ch.RuntimeError = where + "：" + firstLine(e.Comment)
		break
	}
	ch.Acked = ch.NewMessages > 0 || ch.RuntimeBytes > 0
	return ch
}

// pickConn chooses the connection to act on.
func pickConn(conns []ConnStatus, want int) (ConnStatus, error) {
	if want != 0 {
		for _, c := range conns {
			if c.ID == want {
				return c, nil
			}
		}
		return ConnStatus{}, fmt.Errorf("fep: 没有连接 #%d（现有 %d 条，见 tt recording active）", want, len(conns))
	}
	live := make([]ConnStatus, 0, len(conns))
	for _, c := range conns {
		if c.Alive {
			live = append(live, c)
		}
	}
	switch len(live) {
	case 1:
		return live[0], nil
	case 0:
		return ConnStatus{}, fmt.Errorf("fep: 没有活着的连接")
	default:
		return ConnStatus{}, fmt.Errorf("fep: 有 %d 条连接，必须用 --conn 指定（多窗口作业每条窗口一条连接）", len(live))
	}
}

// resolveTarget turns "客户编号" or "#272" into an object id.
//
// 一个名字在一屏里往往命中多个对象（Label + ButtonEdit + FormField 都叫「客户编号」），
// 这是**真实情况**而不是异常。所以按动作类型收敛：点击只认可点击的、输入只认可写的；
// 收敛不到一个就照实报歧义让人用 #id 指定，绝不自己挑一个。
//
// window（`--window`）与 `Scope.Window` **同一个口径**：候选的窗口名是
// `m.WindowName(m.WindowOf(候选))`，**精确相等**才算命中（identity.go 的 windowNamed，
// 契约 A12 精确化 / A18）。以前这里比的是**候选自身的 text**（`m.WindowName(o.ID)`），
// 与窗口名不是一回事 —— 台账 D1：那样会选中不属于该窗口的对象，或者一个都选不出来。
func resolveTarget(m *Model, r *Resolver, target, window, kind string) (int, string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return 0, "", fmt.Errorf("fep: 没给目标")
	}
	if strings.HasPrefix(target, "#") {
		var id int
		if _, err := fmt.Sscanf(target, "#%d", &id); err != nil {
			return 0, "", fmt.Errorf("fep: 看不懂的目标 %q", target)
		}
		if _, ok := m.Get(id); !ok {
			return 0, "", &Refusal{Code: "TARGET_MISSING", Exit: 5, Message: fmt.Sprintf("AUI 里没有 %s", target)}
		}
		return id, "", nil
	}
	// 按名字找：只认高置信度的名字，低置信一律不当依据
	var hits, positioned []*Object
	for _, o := range m.Objects() {
		// A12/A18：窗口归属用 WindowOf 算、名字用 WindowName 取，与 Scope.Window 同一口径。
		// 绝不拿 `m.WindowName(o.ID)`（= 候选自己的 text）当窗口名，也不做任何模糊匹配。
		if window != "" && !windowNamed(m, o, window) {
			continue
		}
		fn, ok := r.Resolve(o.ID)
		if !ok || fn.Text != target || fn.Confidence != ConfHigh {
			continue
		}
		if !acceptableFor(kind, o.Type) {
			continue
		}
		hits = append(hits, o)
		// 同一屏会累积多个画面的同名对象：查询画面那份常常停在 (0,0)，是"没定位"的副本，
		// 点它等于点了空气（实测：点开窗按钮没反应）。优先挑真正有坐标的。
		if o.Attr("posX") != "" || o.Attr("posY") != "" {
			if o.Attr("posX") != "0" || o.Attr("posY") != "0" {
				positioned = append(positioned, o)
			}
		}
	}
	// 只有**点击**才优先有坐标的对象：按钮要点在它所在的位置。
	// 输入不能这么挑 —— 可写的 FormField 本来就没有坐标，照坐标挑会挑到容器（实测踩过，
	// 结果就是"输入目标变成 ButtonEdit 容器"并被规则拦下）。
	if kind == "click" && len(positioned) > 0 {
		hits = positioned
	}
	if len(hits) == 0 {
		msg := fmt.Sprintf("界面里没有叫 %q 且能接受该动作的目标（只认高置信度的名字）", target)
		hint := ""
		if window != "" {
			// 窗口不命中时**必须**把真实存在的窗口名原样列出来（A18 的硬规之一）：
			// 只报"没找到"会让人以为目标名写错了，实际多半是窗口名这一侧写错了
			// —— 而窗口名是"标题(名字)"形态，靠猜不可能猜对。
			msg += fmt.Sprintf("；--window 给的是 %q，而这一屏实际存在的窗口名是：%s", window, windowNamesDesc(m))
			hint = "窗口名照上面那份**原样**复制（可能是「标题(名字)」形态，也可能是空串）；" +
				"不做前缀/后缀/大小写/子串匹配（A18：启发式不参与判定）。" +
				"也可以不经 `--window`：用 `#id` 或 `--field <colName>` 点名"
		}
		return 0, "", &Refusal{Code: "TARGET_MISSING", Exit: 5, Message: msg, Hint: hint}
	}
	if len(hits) == 1 {
		return hits[0].ID, target, nil
	}
	// 输入优先字段值本身（FormField）：实测这是打字该发的地方
	if kind == "input" {
		var forms []*Object
		for _, o := range hits {
			if o.Type == "FormField" {
				forms = append(forms, o)
			}
		}
		if len(forms) == 1 {
			return forms[0].ID, target, nil
		}
	}
	// 点击优先 ToolBarItem：工具栏项与它的 Action 对象同名（如 5160 与 5365），
	// 而客户端点击时发的就是 ToolBarItem 的 id（录制件里 ActionEvent 写的就是 5160）。
	if kind == "click" {
		var bars []*Object
		for _, o := range hits {
			if o.Type == "ToolBarItem" {
				bars = append(bars, o)
			}
		}
		if len(bars) == 1 {
			return bars[0].ID, target, nil
		}
	}
	ids := make([]int, 0, len(hits))
	for _, o := range hits {
		ids = append(ids, o.ID)
	}
	return 0, "", &Refusal{Code: "TARGET_MISSING", Exit: 5,
		Message: fmt.Sprintf("%q 命中 %d 个可选目标（%v）：请用 #id 指定", target, len(hits), ids)}
}

// windowNamesDesc 把这一屏**真实存在**的窗口名写成能直接复制回 `--window` 的一串。
//
// 口径 = `m.WindowName` 的**原样值**（与 windowNamed / Scope.Window 同一个口径，
// A12 精确化 / A18）：它可能是窗口标题形态（`单据别编号数据查询(q_ooba002_1)`），
// 也可能为空串 —— 空的**不列**，因为它粘不回 `--window`，也不解释"选错了窗口"。
// 同名窗口只列一次：这一串的用途是"照抄"，不是统计窗口个数（标题本来就会重名）。
//
// 只影响报错文本，**不参与任何判定**：这里一个名字都不进比较，比较只在 windowNamed 里。
func windowNamesDesc(m *Model) string {
	var names []string
	seen := map[string]bool{}
	for _, w := range m.Windows() {
		n := m.WindowName(w.ID)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		names = append(names, n)
	}
	if len(names) == 0 {
		return "（这一屏没有任何窗口名：运行期没报窗口标题，`--window` 无从匹配，去掉它或改用 #id/--field）"
	}
	sort.Strings(names)
	quoted := make([]string, 0, len(names))
	for _, n := range names {
		quoted = append(quoted, fmt.Sprintf("%q", n))
	}
	return strings.Join(quoted, "、")
}

// acceptableFor 判断一个类型能不能接受该动作。
func acceptableFor(kind, typ string) bool {
	switch kind {
	case "input":
		return writableTypes[typ]
	case "click":
		return ClickableTypes[typ]
	}
	return true
}

func asRefusal(err error, out **Refusal) bool {
	if r, ok := err.(*Refusal); ok {
		*out = r
		return true
	}
	return false
}

// AuditEntry is one line of the audit trail. Every action is recorded, whether it
// was sent or refused: "what did the tool do to the job" must be answerable after
// the fact, and the refusal reasons are part of that answer.
type AuditEntry struct {
	Time    time.Time `json:"time"`
	Verb    string    `json:"verb"`
	Conn    int       `json:"connection"`
	Target  string    `json:"target,omitempty"`
	Name    string    `json:"name,omitempty"`
	Event   string    `json:"event,omitempty"`
	Sent    bool      `json:"sent"`
	DryRun  bool      `json:"dryRun,omitempty"`
	Result  string    `json:"result,omitempty"`
	Note    string    `json:"note,omitempty"`
	Session string    `json:"session,omitempty"`
}

// AppendAudit appends one entry as a JSON line.
func AppendAudit(path string, e AuditEntry) error {
	if path == "" {
		return nil
	}
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}

// Quit asks the resident process to stop (`tt recording stop <录制ID>` runs in another
// process, so it needs a way to say "you may go now" — the resident then runs its
// own full clean-up: remote job, proxy, instance, registry).
func (c *Client) Quit() error {
	reply, err := c.roundTrip("quit")
	if err != nil {
		return err
	}
	if !strings.HasPrefix(reply, "OK") {
		return fmt.Errorf("fep: 常驻进程拒绝退出: %s", reply)
	}
	return nil
}

// readOnlyActions 是"不改数据"的动作名（依据：真机工具栏里实际存在的项 ——
// 查询/流程状态/打印/汇出/切换视图这类）。判断只用来出**警告行**，不硬禁：
// 名单外的动作（录入/更改/删除/复制/确定…）都当成写类。
var readOnlyActions = map[string]bool{
	"查询": true, "二次筛选": true, "流程状态": true, "退出": true,
	"转换至简易画面": true, "打印": true, "快速打印": true,
	"列印语言别": true, "汇出Excel": true, "汇出EXCEL": true,
}

// writesData reports whether this action plausibly changes data.
func writesData(cfg DoConfig, res *DoResult) bool {
	switch cfg.Kind {
	case "input":
		return true
	case "click":
		if res.Name != "" && readOnlyActions[res.Name] {
			return false
		}
		return true
	}
	return false
}

// PickOption finds the item of a field's combo box whose text matches.
//
// 真机上用 GDC 自带的 **AUI Debug Tree**（Ctrl+右键）看到的结构是：
//
//	FormField 352 → ComboBox 353 → Item 47 / 6297 / 6298 / 6299（每个 Item 带 text）
//
// 所以"下拉的候选项"是**协议对象**，可以枚举、也可以点（点中的就是"选择"这一语义动作）。
// 弹出来的那一下仍然是前端本地渲染（协议看不见），但**选择结果在协议里**。
func PickOption(m *Model, fieldID int, text string) (itemID int, itemText string, err error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, "", fmt.Errorf("fep: 没给选项文本")
	}
	// 收集该栏位（或它的后代）下面的 ComboBox
	var combos []int
	for _, o := range m.Objects() {
		if o.Type != "ComboBox" {
			continue
		}
		if o.ID == fieldID || o.Parent == fieldID {
			combos = append(combos, o.ID)
			continue
		}
		if _, _, win := m.Chain(o.ID); win != 0 {
			// 沿父链找 fieldID
			cur := o.Parent
			for i := 0; i < 64 && cur != 0; i++ {
				if cur == fieldID {
					combos = append(combos, o.ID)
					break
				}
				p, ok := m.Get(cur)
				if !ok {
					break
				}
				cur = p.Parent
			}
		}
	}
	if len(combos) == 0 {
		return 0, "", &Refusal{Code: "TARGET_MISSING", Exit: 5,
			Message: fmt.Sprintf("#%d 下面没有 ComboBox（这个栏位不是下拉）", fieldID)}
	}
	type hit struct {
		id    int
		text  string
		exact bool
	}
	var hits []hit
	for _, cid := range combos {
		for _, o := range m.Objects() {
			if o.Type != "Item" || o.Parent != cid {
				continue
			}
			t := o.Attr("text")
			if t == "" {
				t = o.Attr("name")
			}
			if strings.Contains(t, text) {
				hits = append(hits, hit{id: o.ID, text: t, exact: t == text})
			}
		}
	}
	if len(hits) == 0 {
		return 0, "", &Refusal{Code: "TARGET_MISSING", Exit: 5,
			Message: fmt.Sprintf("这个下拉里没有含 %q 的选项", text)}
	}
	// 精确匹配优先；否则若只剩一个就它
	for _, h := range hits {
		if h.exact {
			return h.id, h.text, nil
		}
	}
	if len(hits) == 1 {
		return hits[0].id, hits[0].text, nil
	}
	var names []string
	for _, h := range hits {
		names = append(names, h.text)
	}
	return 0, "", &Refusal{Code: "TARGET_MISSING", Exit: 5,
		Message: fmt.Sprintf("%q 命中 %d 个选项（%s）：写出更完整的选项文本", text, len(hits), strings.Join(names, " / "))}
}
