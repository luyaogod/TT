package view

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/fep"
	"tt/internal/view"
)

// 这一层是**胶水**：把 AUI 的几何属性翻成 view 的 Box，再把几张东西串起来。
// internal/fep 与 internal/view 谁都不 import 谁 —— 两条线的交界就落在这里。

var (
	marksRegion []string
	marksCal    string
	marksFrom   string
	marksFields []string
	marksImage  string
	marksPNG    string
	marksOut    string
	marksRef    string
)

var marksCmd = &cobra.Command{
	Use:   "marks",
	Short: "按协议坐标算标注框（并可顺带过一遍“栏位在这一帧里看得见吗”）",
	Long: `把协议里的几何（posX/posY/gridWidth/gridHeight）换算成图上的框。

**坐标不是猜的**：比例与偏移从参照物（AUI 里的顶层窗口）与抓到的图派生出来；
推出来的东西离谱就报错，而不是接着画一个"看着像"的框。

给了 --png 就顺手过一遍可见性门（计划 §2.12）：说明里点名的每个栏位，必须在那一帧里
真的画出来了 —— 这一条直接防住"标题写在这里输入，而画面根本没走到那个栏位"那类错。

--out 会把框画到图上另存一张（合成用 Go 的 image/draw，逐像素、绝不缩放）。`,
	Args: noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(marksFields) == 0 && len(marksRegion) == 0 {
			return fail(exitUsage, "至少给一个 --field <栏位名> 或 --region <实测区域名>")
		}
		events, err := loadEventsFile(marksFrom)
		if err != nil {
			return err
		}
		_ = events // 目前只用 AUI；留着是为了将来按步骤筛框
		m, r, _, merr := loadModelFile(marksFrom)
		if merr != nil {
			return merr
		}
		var imgSize image.Rectangle
		var img *image.RGBA
		if marksPNG != "" {
			loaded, lerr := view.LoadPNG(marksPNG)
			if lerr != nil {
				return fail(exitFailed, "读图失败: %v", lerr)
			}
			img = loaded
			imgSize = loaded.Bounds()
		} else if marksImage != "" {
			w, h, perr := parseSize(marksImage)
			if perr != nil {
				return perr
			}
			imgSize = image.Rect(0, 0, w, h)
		} else {
			return fail(exitUsage, "要知道图有多大：给 --png <path> 或 --image WxH")
		}

		// 优先用**实测校准**：静态推不出全局映射（几何单位混用），实测的那一份才是可信的。
		var fm view.FrameMap
		var refNote string
		var problems []string
		calPath := marksCal
		if calPath == "" {
			if dir, derr := cacheDir(); derr == nil {
				calPath = filepath.Join(dir, "calibration.json")
			}
		}
		if cal, cerr := view.LoadCalibration(calPath); cerr == nil {
			fm = view.FrameMap{ScaleX: cal.ScaleX, ScaleY: cal.ScaleY, OriginX: cal.OriginX, OriginY: cal.OriginY}
			refNote = fmt.Sprintf("实测校准 %s（%s 量的，%d 个样本）", calPath, cal.At, len(cal.Samples))
		} else {
			ref, refObjs, note := contentExtent(m, marksRef)
			refNote = fmt.Sprintf("%s（%d 个对象，%dx%d 协议单位）", note, refObjs, ref.W, ref.H)
			fm, problems = view.Derive(ref, imgSize)
			if len(problems) > 0 {
				return fail(exitQuality, "坐标映射推不出来：%s（%s）。先跑一次 tt view calibrate 实测一份",
					strings.Join(problems, "；"), refNote)
			}
			reProblems := []string{"（提示：这是**推断**出来的映射，跑 tt view calibrate 可换成实测的）"}
			refNote = refNote + "；" + strings.Join(reProblems, "；")
		}

		// 实测区域优先：协议里没有坐标的东西（明细格子）只能靠量出来的矩形
		var regionMarks []view.Overlay
		var regionBoxes []view.Box
		if len(marksRegion) > 0 {
			rsPath := filepath.Join(filepath.Dir(markCalOrDefault(marksCal)), "regions.json")
			if marksCal != "" {
				rsPath = filepath.Join(filepath.Dir(marksCal), "regions.json")
			}
			rs, lerr := view.LoadRegions(rsPath)
			if lerr != nil {
				return fail(exitFailed, "读区域文件失败: %v", lerr)
			}
			for _, nm := range marksRegion {
				reg, ok := rs.Get(nm)
				if !ok {
					return fail(exitTarget, "区域文件里没有 %q（先 tt view measure --target <栏位>）", nm)
				}
				regionMarks = append(regionMarks, view.Overlay{Rect: reg.Rect, Number: len(regionMarks) + 1, Color: view.HighlightColor})
				regionBoxes = append(regionBoxes, view.Box{X: reg.Rect.Min.X, Y: reg.Rect.Min.Y, W: reg.Rect.Dx(), H: reg.Rect.Dy(), Label: nm + "（实测区域）"})
			}
		}

		var boxes []view.Box
		boxes = append(boxes, regionBoxes...)
		for _, name := range marksFields {
			b, berr := boxForField(m, r, name)
			if berr != nil {
				return berr
			}
			boxes = append(boxes, b)
		}
		frame := view.Frame{Image: img, Map: fm, Fields: boxes}
		type out struct {
			Reference string         `json:"reference"`
			Scale     float64        `json:"scale"`
			OriginY   int            `json:"originY"`
			Marks     []view.Box     `json:"marks"`
			Verdicts  []view.Verdict `json:"verdicts,omitempty"`
			Image     string         `json:"image,omitempty"`
		}
		res := out{Reference: refNote, Scale: fm.ScaleX, OriginY: fm.OriginY, Marks: boxes}
		if img != nil {
			res.Verdicts = view.GateFields(img, boxes, fm)
			res.Image = marksPNG
		}
		if marksOut != "" {
			if img == nil {
				return fail(exitUsage, "--out 需要同时给 --png（要往哪张图上画）")
			}
			overlays := append([]view.Overlay{}, regionMarks...)
			offset := len(overlays)
			for i, b := range boxes {
				if i < offset {
					continue // 实测区域已经加过
				}
				overlays = append(overlays, view.Overlay{Rect: fm.Map(b), Number: i + 1, Color: view.HighlightColor})
			}
			composed := view.Compose(img, overlays)
			if serr := view.SavePNG(marksOut, composed); serr != nil {
				return fail(exitFailed, "写图失败: %v", serr)
			}
			res.Image = marksOut
		}
		// 只给了图尺寸（--image）时没有像素可查：像素门只在真有图时跑。
		var bad []view.Verdict
		if img != nil {
			bad = frame.Gate()
		}
		if perr := printJSONOrText(res, func() {
			fmt.Printf("参照物 %s → 比例 %.3f，客户区上沿 %dpx\n", res.Reference, res.Scale, res.OriginY)
			for i, b := range boxes {
				r := fm.Map(b)
				fmt.Printf("%2d. %-22s 图内 %dx%d+%d+%d\n", i+1, b.Label, r.Dx(), r.Dy(), r.Min.X, r.Min.Y)
			}
			if len(res.Verdicts) > 0 {
				fmt.Printf("可见性门：%s\n", view.Describe(bad))
			}
			if marksOut != "" {
				fmt.Printf("已出图：%s\n", marksOut)
			}
		}); perr != nil {
			return perr
		}
		if len(bad) > 0 {
			return fail(exitQuality, "可见性门未通过：%s", view.Describe(bad))
		}
		return nil
	},
}

// contentExtent 取“参照物”：**所有对象几何的并集**。
//
// 为什么不是拿某个对象当参照物：真实 AUI 里几何属性的单位并不统一 —— 栏位的位置是
// **字符格**（实测 posY "2"、width "8"），而窗口对象自己的尺寸是**像素**（1912）。
// 拿窗口当参照物会推出一个荒唐的比例（纵向下移 1038px）。用并集则不依赖“哪个对象是
// 客户区”这个说不准的判断：整屏内容的包围盒，就该铺满客户区。
func contentExtent(m *fep.Model, windowName string) (view.Box, int, string) {
	winID := 0
	note := "全部对象"
	if windowName != "" {
		for _, w := range m.Windows() {
			if w.Attr("name") == windowName {
				winID = w.ID
				note = "窗口 " + windowName
			}
		}
		if winID == 0 {
			return view.Box{}, 0, "没找到窗口 " + windowName
		}
	}
	minX, minY := 1<<30, 1<<30
	maxX, maxY := -(1 << 30), -(1 << 30)
	n := 0
	for _, o := range m.Objects() {
		if o.Attr("posX") == "" && o.Attr("posY") == "" {
			continue
		}
		if winID != 0 {
			if w, ok := m.Window(o.ID); !ok || w != winID {
				continue
			}
		}
		b := geomOf(o)
		if b.W <= 0 || b.H <= 0 {
			continue
		}
		if b.X < minX {
			minX = b.X
		}
		if b.Y < minY {
			minY = b.Y
		}
		if b.X+b.W > maxX {
			maxX = b.X + b.W
		}
		if b.Y+b.H > maxY {
			maxY = b.Y + b.H
		}
		n++
	}
	if n == 0 {
		return view.Box{}, 0, "没有任何对象带几何属性"
	}
	return view.Box{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}, n, note
}

// geomOf 读一个对象的几何。
//
// 真实 AUI 里几何属性是混着用的（实测）：posX/posY 必有，宽高可能写作
// gridWidth/gridHeight，也可能只写作 width/height（gridHeight 尤其罕见）。
// 所以哪个有用哪个，两个都没有才当“没有几何”。
func geomOf(o *fep.Object) view.Box {
	w, h := num(o, "gridWidth"), num(o, "gridHeight")
	if w <= 0 {
		w = num(o, "width")
	}
	if h <= 0 {
		h = num(o, "height")
	}
	return view.Box{X: num(o, "posX"), Y: num(o, "posY"), W: w, H: h}
}

func areaOf(o *fep.Object) int { b := geomOf(o); return b.W * b.H }

// boxForField 找一个栏位的几何：先按名字解析（只认高置信度），再取它的几何。
func boxForField(m *fep.Model, r *fep.Resolver, name string) (view.Box, error) {
	var hits, positioned []*fep.Object
	for _, o := range m.Objects() {
		fn, ok := r.Resolve(o.ID)
		if !ok || fn.Text != name || fn.Confidence != fep.ConfHigh {
			continue
		}
		// 宽度必须有；高度在真实 AUI 里常常缺省（gridHeight 罕见），缺省就按 1 格算
		b := geomOf(o)
		if b.W <= 0 {
			continue
		}
		hits = append(hits, o)
		// 真实 AUI 里同一屏会累积多个画面的同名对象：查询画面那份常常停在 (0,0)，
		// 那是"没定位"的副本，不能拿来当落点。优先挑真正有位置的。
		if b.X > 0 || b.Y > 0 {
			positioned = append(positioned, o)
		}
	}
	if len(positioned) > 0 {
		hits = positioned
	}
	// 标注画在**栏位**上：优先可写对象（FormField/ButtonEdit/Value…），而不是它旁边的 Label
	var writable []*fep.Object
	for _, o := range hits {
		if fep.IsWritableType(o.Type) {
			writable = append(writable, o)
		}
	}
	if len(writable) > 0 {
		hits = writable
	}
	if len(hits) == 0 {
		return view.Box{}, fail(exitTarget, "界面里没有叫 %q 且有几何的栏位（只认高置信度的名字）", name)
	}
	// 同名的多个对象里取面积最大的那个（可点的容器通常比它内部的 Label 大）
	best := hits[0]
	for _, o := range hits[1:] {
		if areaOf(o) > areaOf(best) {
			best = o
		}
	}
	b := geomOf(best)
	if b.H <= 0 {
		b.H = 1
	}
	b.Label = fmt.Sprintf("%s %s#%d", name, best.Type, best.ID)
	return b, nil
}

func num(o *fep.Object, key string) int {
	n, err := strconv.Atoi(strings.TrimSpace(o.Attr(key)))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func parseSize(s string) (int, int, error) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(s)), "x")
	if len(parts) != 2 {
		return 0, 0, fail(exitUsage, "--image 要写成 WxH：%q", s)
	}
	w, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || w <= 0 || h <= 0 {
		return 0, 0, fail(exitUsage, "--image 里的数字读不出来：%q", s)
	}
	return w, h, nil
}

func init() {
	marksCmd.Flags().StringArrayVar(&marksRegion, "region", nil, "实测区域名（可重复；先 tt view measure 量）")
	marksCmd.Flags().StringVar(&marksCal, "calibration", "", "校准文件（默认 <缓存>/calibration.json，没有就用推导）")
	marksCmd.Flags().StringVar(&marksFrom, "from", "", "录制件：单个 .log 或目录（默认 fep 缓存目录最新一份）")
	marksCmd.Flags().StringArrayVar(&marksFields, "field", nil, "要标注的栏位名（可重复）")
	marksCmd.Flags().StringVar(&marksRef, "ref", "", "参照窗口名（默认 AUI 里最大的 Window）")
	marksCmd.Flags().StringVar(&marksImage, "image", "", "图尺寸 WxH（没有图时也要知道多大）")
	marksCmd.Flags().StringVar(&marksPNG, "png", "", "抓好的图：给了就过可见性门")
	marksCmd.Flags().StringVar(&marksOut, "out", "", "把框画到图上另存（需要 --png）")
	Group.AddCommand(marksCmd)
	_ = os.MkdirAll
	_ = time.Now
}

// markCalOrDefault 返回校准文件的实际落点（默认 <缓存>/calibration.json）。
func markCalOrDefault(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if dir, err := cacheDir(); err == nil {
		return filepath.Join(dir, "calibration.json")
	}
	return "calibration.json"
}
