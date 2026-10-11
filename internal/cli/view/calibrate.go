package view

import (
	"errors"
	"fmt"
	"image"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/fep"
	"tt/internal/view"
)

// tt view calibrate：量出"协议格 → 像素"的换算。
//
// 做法：往两个已知栏位各写一个值 → 各拍两张图 → 取**像素差**当锚点 → 解出映射。
// 不推断、不写死：真机上几何单位是混用的（见 internal/view/README），
// 而像素差是**量出来的**证据。
//
// 每个样本都要过一遍体检（差异框不能太大），但样本互相不一致（残差大）时**直接报错**：
// 勉强拟合出来的映射会把标注框画在栏位旁边，那比没有框更糟。

var (
	calFields  []string
	calOut     string
	calValue   string
	calMaxDiff int
	calSettle  time.Duration
	calDryRun  bool
	calCands   int
)

// candidate 是一个可用的校准样本来源：FormField（往它写值）+ 带几何的容器（取协议坐标）。
type candidate struct {
	Name       string
	FormField  int
	Box        view.Box
	Confidence string
}

var calibrateCmd = &cobra.Command{
	Use:   "calibrate",
	Short: "量出协议格到像素的换算（注入 + 像素差，实测不推断）",
	Long: `在活着的会话上量一次坐标换算，结果存成 JSON 供 marks 使用。

为什么必须实测：真实 AUI 的几何单位是混用的（同一屏里 ButtonEdit 的 posX 是"格"、
Window 的 width 是"像素"），静态推不出全局映射；而注一次输入之后，**变化的像素**
就是那个栏位在屏幕上的真实位置 —— 这是量出来的，不是推的。

做法：默认自动挑 3 个候选栏位（名字可解析、能写、坐标互不相同），往每个写一个值，
拍前后两张图取差异框；用两个以上互不矛盾的样本解出（格宽/格高/原点）。样本残差大、
或差异框大得不像"一个栏位的变化"（说明那一枪引起了别的重绘），就报错、绝不出勉强拟合的映射。

--dry-run 只列出候选与参考坐标，不注入任何东西。`,
	Args: noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newFEPClient()
		if err != nil {
			return err
		}
		m, r, _, merr := fep.ReadModelFrom("")
		if merr != nil {
			m, r, _, merr = loadModelFromCache()
			if merr != nil {
				return fail(exitFailed, "读 AUI 失败：%v", merr)
			}
		}
		cands := pickCandidates(m, r, calFields, calCands)
		if len(cands) < 2 {
			return fail(exitTarget, "能当校准样本的栏位不足（需要两个坐标不同的可写栏位）；用 --field 指定")
		}
		if calDryRun {
			return printJSONOrText(map[string]any{"candidates": cands}, func() {
				for _, cd := range cands {
					fmt.Printf("%-20s FormField#%-6d 协议 %d,%d %dx%d（%s）\n",
						cd.Name, cd.FormField, cd.Box.X, cd.Box.Y, cd.Box.W, cd.Box.H, cd.Confidence)
				}
				fmt.Println("（dry-run：以上是候选，一个字节都没发）")
			})
		}

		valBase := calValue
		if valBase == "" {
			valBase = "13579"
		}
		var samples []view.Sample
		var skipped []string
		for i, cd := range cands {
			value := strconv.Itoa(1000+i) + valBase
			before, berr := captureForCalibration()
			if berr != nil {
				return berr
			}
			_, derr := fep.Do(c, fep.DoConfig{Kind: "input", Target: "#" + strconv.Itoa(cd.FormField),
				Value: value, Wait: calSettle})
			after, aerr := captureForCalibration()
			if aerr != nil {
				return aerr
			}
			if derr != nil {
				skipped = append(skipped, fmt.Sprintf("%s：注入被拒（%v）", cd.Name, derr))
				continue
			}
			box, changed := view.DiffBox(before, after, 30)
			if changed == 0 {
				skipped = append(skipped, fmt.Sprintf("%s：写了值但屏幕没变化", cd.Name))
				continue
			}
			limit := calMaxDiff
			if limit <= 0 {
				limit = 600
			}
			if box.Dx() > limit || box.Dy() > 120 {
				skipped = append(skipped, fmt.Sprintf("%s：差异框 %v 太大（引起了别的重绘），这次样本不用",
					cd.Name, box))
				continue
			}
			samples = append(samples, view.Sample{Box: cd.Box, Pixels: box})
			fmt.Fprintf(cmd.ErrOrStderr(), "· 样本 %s：协议 (%d,%d) → 像素 %v（%d 像素变化）\n",
				cd.Name, cd.Box.X, cd.Box.Y, box, changed)
		}
		fm, problems := view.SolveMapping(samples)
		if len(problems) > 0 {
			return fail(exitQuality, "解不出可用的映射：%s（跳过：%s）",
				strings.Join(problems, "；"), strings.Join(skipped, "；"))
		}
		out := calOut
		if out == "" {
			dir, derr := cacheDir()
			if derr != nil {
				return derr
			}
			out = filepath.Join(dir, "calibration.json")
		}
		cal := view.Calibration{ScaleX: fm.ScaleX, ScaleY: fm.ScaleY, OriginX: fm.OriginX, OriginY: fm.OriginY,
			At: time.Now().Format(time.RFC3339), Samples: samples, Notes: skipped}
		if serr := view.SaveCalibration(out, cal); serr != nil {
			return fail(exitFailed, "写校准失败: %v", serr)
		}
		return printJSONOrText(map[string]any{"path": out, "scaleX": cal.ScaleX, "scaleY": cal.ScaleY,
			"originX": cal.OriginX, "originY": cal.OriginY, "samples": len(samples), "skipped": skipped}, func() {
			fmt.Printf("校准已存：%s\n", out)
			fmt.Printf("  格宽 %.2f px，格高 %.2f px，原点 (%d,%d)，用 %d 个样本\n",
				cal.ScaleX, cal.ScaleY, cal.OriginX, cal.OriginY, len(samples))
			for _, s := range skipped {
				fmt.Printf("  跳过：%s\n", s)
			}
		})
	},
}

// captureForCalibration 抓当前作业窗口（用会话登记里的隔离实例 pid）。
func captureForCalibration() (*image.RGBA, error) {
	pid, ok := sessionGDCPID()
	if !ok {
		return nil, fail(exitUsage, "没有 fep 会话可以推断 pid")
	}
	wins, err := view.ListTimeout(pid, 5*time.Second)
	if err != nil {
		return nil, fail(exitFailed, "%v", err)
	}
	picked := view.PickWindows(wins, pid)
	if len(picked) == 0 {
		return nil, fail(exitTarget, "没有可见窗口")
	}
	img, cerr := view.Capture(picked[0].HWND)
	if cerr != nil {
		return nil, fail(exitFailed, "%v", cerr)
	}
	return view.TrimShadow(img), nil
}

// pickCandidates 挑可用的校准样本：名字可解析（高置信）、能写（FormField）、坐标互不相同。
func pickCandidates(m *fep.Model, r *fep.Resolver, want []string, limit int) []candidate {
	if len(want) > 0 {
		var out []candidate
		for _, name := range want {
			if cd, ok := candidateFor(m, r, name); ok {
				out = append(out, cd)
			}
		}
		return out
	}
	if limit <= 0 {
		limit = 3
	}
	seenName := map[string]bool{}
	var out []candidate
	for _, o := range m.Objects() {
		if len(out) >= limit {
			break
		}
		fn, ok := r.Resolve(o.ID)
		if !ok || fn.Confidence != fep.ConfHigh || fn.Text == "" || seenName[fn.Text] {
			continue
		}
		cd, ok := candidateFor(m, r, fn.Text)
		if !ok {
			continue
		}
		// 坐标必须与已有的不同（同格解不出比例）
		dup := false
		for _, e := range out {
			if e.Box.X == cd.Box.X && e.Box.Y == cd.Box.Y {
				dup = true
			}
		}
		if dup {
			continue
		}
		seenName[fn.Text] = true
		out = append(out, cd)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Box.Y != out[j].Box.Y {
			return out[i].Box.Y < out[j].Box.Y
		}
		return out[i].Box.X < out[j].Box.X
	})
	return out
}

// candidateFor 找一个栏位：能写的 FormField + 带几何的容器。
func candidateFor(m *fep.Model, r *fep.Resolver, name string) (candidate, bool) {
	ff := 0
	var boxObj *fep.Object
	for _, o := range m.Objects() {
		fn, ok := r.Resolve(o.ID)
		if !ok || fn.Text != name || fn.Confidence != fep.ConfHigh {
			continue
		}
		if o.Type == "FormField" && ff == 0 {
			ff = o.ID
		}
		if b := boxOf(o); b.W > 0 && b.H >= 0 && (boxObj == nil || area(b) > area(boxOf(boxObj))) {
			boxObj = o
		}
	}
	if ff == 0 || boxObj == nil {
		return candidate{}, false
	}
	b := boxOf(boxObj)
	if b.W <= 0 {
		return candidate{}, false
	}
	if b.H <= 0 {
		b.H = 1 // 行高按 1 格算（真实 AUI 里 height 常常缺省）
	}
	return candidate{Name: name, FormField: ff, Box: b, Confidence: "high"}, true
}

func boxOf(o *fep.Object) view.Box {
	return view.Box{X: attrInt(o, "posX"), Y: attrInt(o, "posY"), W: attrInt(o, "gridWidth", "width"),
		H: attrInt(o, "gridHeight", "height")}
}

func attrInt(o *fep.Object, keys ...string) int {
	for _, k := range keys {
		if v := strings.TrimSpace(o.Attr(k)); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				return n
			}
		}
	}
	return 0
}

func area(b view.Box) int { return b.W * b.H }

// newFEPClient 造一个与常驻会话说话的 Client（与 tt fep 那边同一套登记与落点）。
func newFEPClient() (*fep.Client, error) {
	cfgPath, err := configResolve()
	if err != nil {
		return nil, err
	}
	dir := dataDirOf(cfgPath)
	s, lerr := fep.NewRegistry(dir).Load()
	if errors.Is(lerr, fep.ErrNoSession) {
		return nil, fail(exitNoSession2, "没有登记中的会话（先跑 tt fep session open）")
	}
	if lerr != nil {
		return nil, fail(exitFailed, "%v", lerr)
	}
	recDir := s.RecordDir
	if recDir == "" {
		if recDir, err = cacheDir(); err != nil {
			return nil, err
		}
	}
	return &fep.Client{InjectAddr: netJoin(s.InjectPort), RecordDir: recDir, Tag: s.Tag}, nil
}

func loadModelFromCache() (*fep.Model, *fep.Resolver, int, error) {
	c, err := newFEPClient()
	if err != nil {
		return nil, nil, 0, err
	}
	return fep.LoadModelFrom(c.RecordDir, c.Tag)
}

func init() {
	calibrateCmd.Flags().StringArrayVar(&calFields, "field", nil, "指定用于校准的栏位名（可重复；默认自动挑 3 个）")
	calibrateCmd.Flags().StringVar(&calOut, "out", "", "校准文件落点（默认 <缓存>/calibration.json）")
	calibrateCmd.Flags().StringVar(&calValue, "value", "", "注入的值（默认 13579，前面补序号）")
	calibrateCmd.Flags().IntVar(&calMaxDiff, "max-diff", 600, "差异框宽度上限")
	calibrateCmd.Flags().DurationVar(&calSettle, "wait", 5*time.Second, "每次注入后等运行期多久")
	calibrateCmd.Flags().BoolVar(&calDryRun, "dry-run", false, "只列候选，不注入")
	calibrateCmd.Flags().IntVar(&calCands, "candidates", 3, "自动挑几个候选栏位")
	Group.AddCommand(calibrateCmd)
}
