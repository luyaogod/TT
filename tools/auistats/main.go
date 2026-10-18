// auistats 对一批 FEP 录制做「对象身份唯一性」统计，回答回放驱动最关心的四个问题：
//
//	Q1 (窗口, 类型, name) 唯一吗？不唯一时 tabName / tabIndex / 父 id / posX,posY 能消歧多少？
//	Q2 同一个 类型+name 跨窗口重复多少？
//	Q3 惰性对象（Completer 等）首次出现在哪一侧（运行期推来 vs 客户端请求出来）？
//	Q4 裸 id 复用：同一 (Scope, ID) 上类型/名字变过多少次、隔多少条消息？
//
// 用法：
//
//	go run ./tools/auistats -root <语料目录> [-json]
//
// 口径：
//   - 语料 = fep.ReadRecording 能读的一切（二进制帧流 .log 与 testdata/fep-recordings/*.txt 同路径）；
//   - **身份键是 (Scope, ID)**：Model 的 Object.Scope 是声明作用域，同一裸 id 在不同 Scope 下
//     是两个不同对象；混在一起会把"不同对象"误算成"复用"；
//   - Q1 分两桶：可寻址类型（控件/容器/动作）与其它（属性/集合项）；属性类同名同类型天然多发
//     （StyleAttribute.textFormat ×5），不该进"按名字定位"的分母。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tt/internal/fep"
)

type snapObj struct {
	ID            int
	Type          string
	Name          string
	Parent        int
	Win           int
	Scope         int
	Seq           int
	TabName       string
	ParentTN      string // 父的 (类型,name) —— **不含运行期 id**，回放可用
	Path          string // 整条祖先链的 (类型,name) 序列 —— 不含 id
	StructPath    string // 祖先链 + **每层同类型序号** —— 不含 id，且能区分"同名同类型的两个父"
	TabIndex      string
	PosX          string
	PosY          string
	Children      int
	ChildrenNames string
}

type tally struct {
	winGroupsAddr, winExtraAddr, winGroupsOther int
	addrObjects, addrObjectsCollided            int // 可寻址对象总数 / 落在撞名组里的对象数
	afterTabName, afterTabIndex, afterParent    int
	afterParentPos, afterParentCh               int
	afterParentTN, afterPath                    int
	afterStruct                                 int
	byTypeAddr, byTypeOther                     map[string]int
	samples                                     []string
	crossWin                                    map[string]map[int]bool
	lazyFirst                                   map[string]int
	idsSeen, reuseType, reuseName               int
	reuseSamples                                []string
	reuseDist                                   []int
	recordings, messages, snaps, objects        int
}

func newTally() *tally {
	return &tally{byTypeAddr: map[string]int{}, byTypeOther: map[string]int{},
		crossWin: map[string]map[int]bool{}, lazyFirst: map[string]int{}}
}

var addressingTypes = map[string]bool{	"Window": true, "Dialog": true, "Form": true, "Group": true, "Page": true, "Tab": true,
	"FormField": true, "TableColumn": true, "PhantomColumn": true, "PhantomFormField": true,
	"Label": true, "Button": true, "Edit": true, "ComboBox": true, "CheckBox": true,
	"RadioGroup": true, "SpinEdit": true, "DateEdit": true, "DateTimeEdit": true,
	"TextEdit": true, "Image": true, "Completer": true, "Aggregate": true,
	"ToolBar": true, "ToolBarItem": true, "Menu": true, "MenuItem": true, "Action": true,
	"Table": true, "Grid": true, "ScrollBar": true, "Message": true,
}

var lazyTypes = map[string]bool{
	"Completer": true, "Aggregate": true, "PhantomEdit": true, "PhantomFormField": true,
	"PhantomColumn": true, "DateTimeEdit": true, "Image": true,
}

func main() {
	root := flag.String("root", "", "语料目录（递归找录制件）")
	asJSON := flag.Bool("json", false, "机器可读输出")
	flag.Parse()
	if *root == "" {
		fmt.Fprintln(os.Stderr, "需要 -root <语料目录>")
		os.Exit(2)
	}
	files := collect(*root)
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "%s 下没有可读的录制件\n", *root)
		os.Exit(1)
	}
	t := newTally()

	for _, f := range files {
		msgs, err := fep.ReadRecording(f)
		if err != nil || len(msgs) == 0 {
			continue
		}
		t.recordings++
		m := fep.NewModel()
		var seen map[int64]snapObj
		lastMsg := map[int64]int{}
		for i, msg := range msgs {
			m.ApplyMessage(msg)
			t.messages++
			snap := takeSnapshot(m)
			t.snaps++
			t.objects += len(snap)

			groups := map[string][]snapObj{}
			for _, o := range snap {
				if o.Name == "" {
					continue
				}
				if addressingTypes[o.Type] {
					t.addrObjects++
				}
				k := fmt.Sprintf("%d\x00%s\x00%s", o.Win, o.Type, o.Name)
				groups[k] = append(groups[k], o)
			}
			for _, g := range groups {
				if len(g) < 2 {
					continue
				}
				if addressingTypes[g[0].Type] {
					t.addrObjectsCollided += len(g)
				}
			}
			for _, g := range groups {
				if len(g) < 2 {
					continue
				}
				if !addressingTypes[g[0].Type] {
					t.winGroupsOther++
					t.byTypeOther[g[0].Type]++
					continue
				}
				t.winGroupsAddr++
				t.winExtraAddr += len(g) - 1
				t.byTypeAddr[g[0].Type]++
				if len(t.samples) < 10 {
					t.samples = append(t.samples, fmt.Sprintf("%s | %s.%q ×%d (win=%d, parent=%d)",
						short(f), g[0].Type, g[0].Name, len(g), g[0].Win, g[0].Parent))
				}
				t.afterTabName += stillDup(g, func(o snapObj) string { return o.TabName })
				t.afterTabIndex += stillDup(g, func(o snapObj) string { return o.TabIndex })
				t.afterParent += stillDup(g, func(o snapObj) string { return fmt.Sprint(o.Parent) })
				t.afterParentPos += stillDup(g, func(o snapObj) string { return o.PosX + "," + o.PosY })
				// **无 id 判据**（回放不能用运行期 id）：父的 (类型,name) 与整条祖先链。
				t.afterParentTN += stillDup(g, func(o snapObj) string { return o.ParentTN })
				t.afterPath += stillDup(g, func(o snapObj) string { return o.Path })
				t.afterStruct += stillDup(g, func(o snapObj) string { return o.StructPath })
				t.afterParentCh += stillDup(g, func(o snapObj) string {
					return fmt.Sprintf("%d/%d/%s", o.Parent, o.Children, o.ChildrenNames)
				})
			}

			for _, o := range snap {
				if o.Name == "" {
					continue
				}
				k := o.Type + "\x00" + o.Name
				if t.crossWin[k] == nil {
					t.crossWin[k] = map[int]bool{}
				}
				t.crossWin[k][o.Win] = true
			}

			// Q3：惰性对象首次出现（首个快照也算 —— 大模板树就是在第 1 条消息里长出来的）
			for key, o := range snap {
				if seen != nil {
					if _, ok := seen[key]; ok {
						continue
					}
				}
				if lazyTypes[o.Type] {
					t.lazyFirst[o.Type]++
				}
			}
			if seen != nil {
				for key, o := range snap {
					prev, ok := seen[key]
					if !ok {
						t.idsSeen++
						continue
					}
					if prev.Type != o.Type {
						t.reuseType++
						if at, ok2 := lastMsg[key]; ok2 {
							t.reuseDist = append(t.reuseDist, i-at)
						}
						if len(t.reuseSamples) < 10 {
							t.reuseSamples = append(t.reuseSamples,
								fmt.Sprintf("%s | (scope=%d,id=%d): %s.%q → %s.%q @msg#%d",
									short(f), o.Scope, o.ID, prev.Type, prev.Name, o.Type, o.Name, i))
						}
					} else if prev.Name != o.Name {
						t.reuseName++
					}
				}
			}
			for key := range snap {
				lastMsg[key] = i
			}
			seen = snap
		}
	}

	crossMulti, crossMax := 0, 0
	for _, wins := range t.crossWin {
		if len(wins) > 1 {
			crossMulti++
			if len(wins) > crossMax {
				crossMax = len(wins)
			}
		}
	}
	sort.Ints(t.reuseDist)

	if *asJSON {
		out := map[string]any{
			"recordings": t.recordings, "messages": t.messages, "snapshots": t.snaps, "objectSlots": t.objects,
			"q1": map[string]any{
				"addressingCollisionGroups": t.winGroupsAddr, "addressingExtraObjects": t.winExtraAddr,
				"otherCollisionGroups": t.winGroupsOther,
				"after": map[string]int{"tabName": t.afterTabName, "tabIndex": t.afterTabIndex,
					"parent": t.afterParent, "parentPos": t.afterParentPos, "parentChildrenNames": t.afterParentCh,
					"parentTypeName": t.afterParentTN, "ancestorPathNoID": t.afterPath},
				"byTypeAddressing": t.byTypeAddr, "byTypeOther": t.byTypeOther, "samples": t.samples,
			},
			"q2": map[string]any{"keys": len(t.crossWin), "multiWindow": crossMulti, "maxWindows": crossMax},
			"q3": t.lazyFirst,
			"q4": map[string]any{"keys": t.idsSeen, "typeChanged": t.reuseType, "nameChanged": t.reuseName,
				"samples": t.reuseSamples, "distMin": minInt(t.reuseDist), "distMax": maxInt(t.reuseDist)},
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return
	}

	fmt.Printf("语料：%d 份录制 / %d 条消息 / %d 个快照 / %d 个对象槽\n\n", t.recordings, t.messages, t.snaps, t.objects)
	fmt.Println("【Q1】同一窗口内 (类型, name) 撞名")
	fmt.Printf("  可寻址类型：撞名组 %d（多出对象 %d）\n", t.winGroupsAddr, t.winExtraAddr)
	if t.addrObjects > 0 {
		fmt.Printf("  可寻址对象 %d 个槽位，其中落在撞名组的 %d（%.2f%%）\n",
			t.addrObjects, t.addrObjectsCollided,
			100*float64(t.addrObjectsCollided)/float64(t.addrObjects))
	}
	printDist("    按类型", t.byTypeAddr)
	fmt.Printf("  其它类型（属性/集合项等，本就不该当身份）：撞名组 %d\n", t.winGroupsOther)
	printDist("    按类型", t.byTypeOther)
	fmt.Println("  加候选判据后仍撞名的组数（只在可寻址类型上算，越小越有用）：")
	fmt.Printf("    tabName            %d\n", t.afterTabName)
	fmt.Printf("    tabIndex           %d\n", t.afterTabIndex)
	fmt.Printf("    父 id              %d\n", t.afterParent)
	fmt.Printf("    posX,posY          %d\n", t.afterParentPos)
	fmt.Printf("    父 id + 子数/子名  %d\n", t.afterParentCh)
	fmt.Printf("    父的(类型,name)     %d   ← 无 id，回放可用\n", t.afterParentTN)
	fmt.Printf("    整条祖先链(无 id)   %d   ← 无 id，回放可用\n", t.afterPath)
	fmt.Printf("    结构路径(链+同层序号) %d   ← 无 id，回放可用（能区分同名同类型的父）\n", t.afterStruct)
	if len(t.samples) > 0 {
		fmt.Println("  可寻址类型的撞名样例：")
		for _, s := range t.samples {
			fmt.Println("    " + s)
		}
	}
	fmt.Println("\n【Q2】跨窗口重复")
	fmt.Printf("  类型+name 键 %d 个，其中跨窗口 %d 个，最多落在 %d 个窗口\n", len(t.crossWin), crossMulti, crossMax)
	fmt.Println("\n【Q3】惰性对象首次出现（-rt2gdc = 运行期→前端）")
	printDist("  ", t.lazyFirst)
	fmt.Println("\n【Q4】裸 id 复用（键 = (Scope, ID)）")
	fmt.Printf("  (scope,id) 键 %d 个；类型变过 %d 次、同类型换名 %d 次\n", t.idsSeen, t.reuseType, t.reuseName)
	if len(t.reuseDist) > 0 {
		fmt.Printf("  类型变化距上次出现的消息数：min=%d 中位=%d max=%d\n",
			t.reuseDist[0], t.reuseDist[len(t.reuseDist)/2], t.reuseDist[len(t.reuseDist)-1])
	}
	for _, s := range t.reuseSamples {
		fmt.Println("    " + s)
	}
}

func printDist(label string, m map[string]int) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return m[keys[i]] > m[keys[j]] })
	if len(keys) == 0 {
		fmt.Println(label + "：（无）")
		return
	}
	for _, k := range keys {
		fmt.Printf("%s  %-20s %d\n", label, k, m[k])
	}
}

func takeSnapshot(m *fep.Model) map[int64]snapObj {
	objs := m.Objects()
	out := make(map[int64]snapObj, len(objs))
	childCount := map[int]int{}
	childNames := map[int][]string{}
	byID := make(map[int]*fep.Object, len(objs))
	for _, o := range objs {
		byID[o.ID] = o
	}
	// 同层序号：同一父下同类型兄弟里的次序（用模型 Objects() 的顺序，声明序稳定）
	sibOrd := map[int]int{}
	perKey := map[string]int{}
	for _, o := range objs {
		k := fmt.Sprintf("%d\x00%s", o.Parent, o.Type)
		sibOrd[o.ID] = perKey[k]
		perKey[k]++
	}
	for _, o := range objs {
		childCount[o.Parent]++
		if n := o.Attr("name"); n != "" {
			childNames[o.Parent] = append(childNames[o.Parent], n)
		}
	}
	for _, o := range objs {
		names := childNames[o.ID]
		sort.Strings(names)
		// 无 id 的路径：一路向上的 (类型:name) 序列。用它消歧，回放时不需要运行期 id。
		var path []string
		var structPath []string
		cur := o
		for depth := 0; depth < 16 && cur != nil; depth++ {
			path = append(path, cur.Type+":"+cur.Attr("name"))
			// 同层序号：在"同一个父下、同类型"的兄弟里排第几（按模型里的稳定顺序）
			structPath = append(structPath, fmt.Sprintf("%s:%s#%d", cur.Type, cur.Attr("name"), sibOrd[cur.ID]))
			if cur.Parent == 0 {
				break
			}
			cur = byID[cur.Parent]
		}
		parentTN := ""
		if p, ok := byID[o.Parent]; ok {
			parentTN = p.Type + ":" + p.Attr("name")
		}
		out[scopeKey(o.Scope, o.ID)] = snapObj{
			ID: o.ID, Type: o.Type, Name: o.Attr("name"), Parent: o.Parent,
			Win: o.Win, Scope: o.Scope, Seq: o.Seq,
			TabName: o.Attr("tabName"), TabIndex: o.Attr("tabIndex"),
			PosX: o.Attr("posX"), PosY: o.Attr("posY"),
			ParentTN: parentTN, Path: strings.Join(path, "/"),
			StructPath: strings.Join(structPath, "/"),
			Children: childCount[o.ID], ChildrenNames: strings.Join(names, ","),
		}
	}
	return out
}

func scopeKey(scope, id int) int64 { return int64(scope)<<32 | int64(uint32(id)) }

func stillDup(g []snapObj, key func(snapObj) string) int {
	seen := map[string]bool{}
	for _, o := range g {
		seen[key(o)] = true
	}
	if len(seen) < len(g) {
		return 1
	}
	return 0
}

func minInt(a []int) int {
	if len(a) == 0 {
		return 0
	}
	return a[0]
}

func maxInt(a []int) int {
	if len(a) == 0 {
		return 0
	}
	return a[len(a)-1]
}

func short(p string) string { return strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)) }

func collect(root string) []string {
	var out []string
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		n := info.Name()
		switch {
		case strings.HasSuffix(n, "-rt2gdc.log"), strings.HasSuffix(n, "-gdc2rt.log"):
			out = append(out, p)
		case strings.HasSuffix(n, ".txt") && strings.Contains(p, "fep-recordings"):
			out = append(out, p)
		}
		return nil
	})
	sort.Strings(out)
	return out
}
