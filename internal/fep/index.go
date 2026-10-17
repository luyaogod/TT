package fep

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// 帧索引（计划 C 的 D0）：录制时由代理顺手落一份 `<raw>.frames.jsonl`，
// 一行一帧，记下**帧级时刻**与**字节位置**。
//
// 为什么必须另存一份而不能从录制件现推：
//
//	raw 里只有字节，没有时间。`timeline.log` 有时间，但它是**块级**的、而且每块截断
//	1500 字节 —— 拿它当索引源会得到一个安静的错答案（"这条事件在这一块里"，但块里到底
//	有哪几帧、每帧在第几个字节，它答不出来）。时间只在录制的那一刻可得，所以只能当时记。
//
// 三条不可动摇的规矩：
//
//  1. **raw 是唯一权威**：索引只是它的派生视图。任何一帧都能拿 `off/len` 回 raw 取原字节，
//     索引缺了/坏了，重新解析 raw 就能重建（`tt recording index`）。
//  2. **索引不参与证据**：指纹、闸门、回放计划一律从 raw 算，绝不从索引算 ——
//     索引是缓存，缓存不该影响"这份录制还算不算数"。
//  3. **图片缓存不是消息**：type=5 那类帧（`!!__cached__!!`、图片清单）也落索引，
//     但 `kind=cache` —— 既不解析也不重放（H9），只是"这段时间在传图"这个事实。
//     真正没认出来的字节（握手、残块）留成 `off` 的跳跃，见 Noise。

// IndexSuffix 是帧索引的扩展名。
//
// 命名规则：**索引与它的 raw 同一个干**（`<ID>-<连接>-rt2gdc.log` →
// `<ID>-<连接>-rt2gdc.frames.jsonl`）。为什么不照计划里最初写的"一条连接一份"：
// 一条连接有**两个方向**，而 `off` 是对**各自的 raw 文件**算的 —— 合在一份里，
// 同一个 `off` 会同时指向两个文件，读侧每次都得先想起"这条记录属于哪一侧"。
// 分开存之后，"索引 ↔ raw"是纯粹的一对一。
const IndexSuffix = ".frames.jsonl"

// 方向取值，与 `timeline.log` 的 `>>`/`<<` 同义（**从运行期视角看**）。
const (
	// DirRuntime 是 `>>`：运行期 → 前端，raw 是 `*-rt2gdc.log`（界面状态在这一侧）。
	DirRuntime = 1
	// DirFront 是 `<<`：前端 → 运行期，raw 是 `*-gdc2rt.log`（用户/注入的事件在这一侧）。
	DirFront = 0
)

// indexVersion 是索引格式版本。字段或**收录范围**一变就要 +1：
// 读侧靠它决定"要不要重建"，而不是拿一份旧定义的索引去画时间轴。
//
// 版本史：
//
//	1 —— 只收 AUI 帧（type=1）。
//	2 —— 图片缓存（type=5）也收，`kind=cache`。真机实测一份录制两个方向各有 157 条，
//	     不收的话"缓存/其它"那条泳道就没有时间（只剩字节区间）。
const indexVersion = 2

// FrameRecord 是索引里的一行。
//
// 字段刻意短（json 名 t/conn/dir/off/len/seq/kind/idRef）：一份长录制的索引可以有几万行，
// 而这些名字就是它唯一的读者（我们自己的读侧）要看的东西。
type FrameRecord struct {
	T     int64  `json:"t,omitempty"`     // 帧**首块**到达时刻（Unix 毫秒）；0 = 不知道（老件重建）
	Conn  int    `json:"conn"`            // 连接号（就是文件名里那一段）
	Dir   int    `json:"dir"`             // DirRuntime / DirFront
	Off   int64  `json:"off"`             // 在对应 raw 文件里的起始偏移
	Len   int    `json:"len"`             // 占用字节数（含 9 字节帧头与结尾换行）
	Seq   int    `json:"seq,omitempty"`   // 消息序号：om/un/rn 之后的数字，event 的 _om 号
	Kind  string `json:"kind"`            // 消息类别：om/un/rn/event/meta/other
	IDRef int    `json:"idRef,omitempty"` // 这条消息**主要指向**的对象 id（见 classifyMessage）
}

// indexLine 是索引文件里一行的完整形状：帧记录 + 头行/尾行的元数据。
//
// 头行与尾行用**同一个** struct 的零值字段区分（V>0 是头、End>0 是尾、Kind 非空是帧），
// 这样读侧只有一个 json.Unmarshal 调用，也不会因为将来加字段就认不出老索引。
type indexLine struct {
	FrameRecord
	V       int    `json:"v,omitempty"`       // 头行：格式版本
	Raw     string `json:"raw,omitempty"`     // 头行：对应的 raw 文件名（搬库后仍然只是名字）
	Rebuilt bool   `json:"rebuilt,omitempty"` // 头行：这份索引是**事后重建**的（所以 t 全是空）
	End     int64  `json:"end,omitempty"`     // 尾行：raw 一共写了多少字节
	Frames  int    `json:"frames,omitempty"`  // 尾行：一共多少帧
}

// ErrNoIndex 表示这个方向没有帧索引（raw 还在，可以重建）。
var ErrNoIndex = errors.New("fep: 没有帧索引（raw 还在，可用 tt recording index 重建）")

// IndexPathFor 返回某个 raw 文件对应的索引落点（与 raw 同一个干）。
func IndexPathFor(raw string) string {
	base := raw
	if strings.HasSuffix(base, ".log") {
		base = strings.TrimSuffix(base, ".log")
	}
	return base + IndexSuffix
}

// RawOfIndex 是反过来的那一步：索引文件名 → 它记的是哪个 raw 文件。
//
// 第二个返回值报告"这个文件名是不是索引名"。读侧拿它把 `meta` 里记的 raw 名
// 换成索引名，不需要在 meta 里再存一份索引名（那是第二个真相，迟早对不上）。
func RawOfIndex(index string) (string, bool) {
	if !strings.HasSuffix(index, IndexSuffix) {
		return "", false
	}
	return strings.TrimSuffix(index, IndexSuffix) + ".log", true
}

// IsIndexFile 报告一个文件名是不是帧索引。
func IsIndexFile(name string) bool { _, ok := RawOfIndex(name); return ok }

// FrameIndexWriter 一边收原始块、一边把解出来的帧逐行落盘。
//
// 它**不碰 raw 的写句柄**：raw 由代理自己写，索引只记位置。两条路互不干扰 ——
// 索引写失败（磁盘满、权限）绝不该拖垮录制，raw 是权威。
type FrameIndexWriter struct {
	f     *os.File
	bw    *bufio.Writer
	sc    *OffsetScanner
	conn  int
	dir   int
	n     int   // 落了多少帧
	bytes int64 // 一共喂进来多少字节（= 这个 raw 文件应该有多大）
	cov   int64 // 被帧覆盖的字节（bytes - cov = 非帧字节）
	err   error
}

// NewFrameIndexWriter 创建索引文件并写下头行（录制进行时用这个）。
func NewFrameIndexWriter(path string, conn, dir int) (*FrameIndexWriter, error) {
	return newFrameIndexWriter(path, conn, dir, false)
}

// NewRebuiltFrameIndexWriter 同上，只是头行里标上"这是**事后重建**的索引"。
//
// 为什么要在文件里标出来：重建的索引 `t` 全是空。读侧必须能区分
// "这份录制本来就没有帧级时间"与"这个字段没人填"，否则界面会把 0 画成 1970 年。
func NewRebuiltFrameIndexWriter(path string, conn, dir int) (*FrameIndexWriter, error) {
	return newFrameIndexWriter(path, conn, dir, true)
}

func newFrameIndexWriter(path string, conn, dir int, rebuilt bool) (*FrameIndexWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := &FrameIndexWriter{
		f:    f,
		bw:   bufio.NewWriterSize(f, 64<<10),
		sc:   NewOffsetScanner(),
		conn: conn,
		dir:  dir,
	}
	head, _ := json.Marshal(indexLine{
		V: indexVersion, Raw: RawOfIndexName(path), Rebuilt: rebuilt,
		FrameRecord: FrameRecord{Conn: conn, Dir: dir},
	})
	w.bw.Write(head)
	w.bw.WriteByte('\n')
	return w, nil
}

// RawOfIndexName 从一个索引路径换回它的 raw 文件名。
func RawOfIndexName(index string) string {
	if raw, ok := RawOfIndex(filepath.Base(index)); ok {
		return raw
	}
	return filepath.Base(index)
}

// Write 喂一段原始字节并落盘新收齐的帧。
//
// 时间语义（H8）：`at` 是**这一段字节到手**的时刻；跨块的帧取它**首块**所在那一段的时刻，
// 由 OffsetScanner 记着。
func (w *FrameIndexWriter) Write(chunk []byte, at time.Time) {
	if w == nil || w.err != nil || len(chunk) == 0 {
		return
	}
	w.bytes += int64(len(chunk))
	for _, sp := range w.sc.Feed(chunk, at) {
		kind, seq, idRef := FrameKind(sp)
		b, err := json.Marshal(FrameRecord{
			T:     msOf(sp.T),
			Conn:  w.conn,
			Dir:   w.dir,
			Off:   sp.Off,
			Len:   sp.Len,
			Seq:   seq,
			Kind:  kind,
			IDRef: idRef,
		})
		if err != nil {
			w.err = err
			return
		}
		w.bw.Write(b)
		w.bw.WriteByte('\n')
		w.n++
		w.cov += int64(sp.Len)
	}
}

// Count 返回已落盘的帧数。
func (w *FrameIndexWriter) Count() int { return w.n }

// Bytes 返回已喂进来的字节数（= 这个方向 raw 应有的字节数）。
func (w *FrameIndexWriter) Bytes() int64 { return w.bytes }

// Covered 返回被帧覆盖的字节数。
func (w *FrameIndexWriter) Covered() int64 { return w.cov }

// Close 写下尾行并落盘。
//
// 尾行是"索引完整"的唯一凭据：进程被硬杀时它不会出现，读侧据此说"索引不完整"
// （而不是把一份短索引当成完整索引用 —— 那会让界面在某个时刻突然断掉却不说为什么）。
func (w *FrameIndexWriter) Close() error {
	if w == nil {
		return nil
	}
	if w.err == nil {
		tail, _ := json.Marshal(indexLine{End: w.bytes, Frames: w.n})
		w.bw.Write(tail)
		w.bw.WriteByte('\n')
		if err := w.bw.Flush(); err != nil {
			w.err = err
		}
	}
	cerr := w.f.Close()
	if w.err != nil {
		return w.err
	}
	return cerr
}

// FrameIndex 是一个方向（一个 raw 文件）的帧索引。
type FrameIndex struct {
	Path string
	Raw  string // 它记的 raw 文件名（只是名字：索引与 raw 同住一个目录）
	Recs []FrameRecord
	Conn int
	Dir  int
	// Version 是这份索引的格式版本（头行里的 v）。与 indexVersion 不同就必须重建：
	// 索引的定义会随界面需求演进（例如 type=5 的图片缓存帧是后来才决定收进来的），
	// 拿旧定义的"完整索引"当完整，会得到一个**看起来对但缺一整类帧**的时间轴。
	Version int
	// Rebuilt 表示这份索引是事后重建的：帧位置是真的，**帧时间全是空的**。
	Rebuilt bool
	// End 是尾行报的 raw 字节数；HasEnd 表示尾行在不在（进程被杀时不在）。
	End    int64
	HasEnd bool
	// Bad 是读不动、被跳过的行数（>0 = 索引有损，界面要照实标）。
	Bad int
	// RawSize 是 raw 文件现在的实际大小（0 = 量不到）。
	RawSize int64
}

// ReadFrameIndex 读一个方向的帧索引。
//
// 坏行**不报错**：索引是缓存，一行读不动就把那一行记进 Bad 继续读 —— 为了一行坏数据
// 让整个明细页打不开，是把缓存的重要性放到了 raw 之上。
func ReadFrameIndex(path string) (*FrameIndex, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoIndex
		}
		return nil, err
	}
	defer f.Close()

	ix := &FrameIndex{Path: path, Raw: RawOfIndexName(path)}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var ln indexLine
		if err := json.Unmarshal(line, &ln); err != nil {
			// 坏行只计数：索引是缓存，为一行坏数据让整个明细页打不开，
			// 是把缓存的重要性放到了 raw 之上。
			ix.Bad++
			continue
		}
		switch {
		case ln.V > 0:
			ix.Conn, ix.Dir, ix.Rebuilt, ix.Version = ln.Conn, ln.Dir, ln.Rebuilt, ln.V
			if ln.Raw != "" {
				ix.Raw = ln.Raw
			}
		case ln.End > 0:
			ix.End, ix.HasEnd = ln.End, true
		default:
			ix.Recs = append(ix.Recs, ln.FrameRecord)
			if ix.Conn == 0 {
				ix.Conn, ix.Dir = ln.Conn, ln.Dir
			}
		}
	}
	// 尾行（`HasEnd`）才是"索引写到收尾"的凭据：进程被硬杀时它不会出现 ——
	// 这个区分决定了界面该说"索引不完整（可重建）"还是"索引有损"。
	if st, serr := os.Stat(filepath.Join(filepath.Dir(path), ix.Raw)); serr == nil {
		ix.RawSize = st.Size()
	}
	return ix, sc.Err()
}

// Covered 返回被帧覆盖的字节数。
func (ix *FrameIndex) Covered() int64 {
	if ix == nil {
		return 0
	}
	var n int64
	for _, r := range ix.Recs {
		n += int64(r.Len)
	}
	return n
}

// Complete 报告这份索引是否可信地覆盖了整个 raw 文件。
//
// 判据是**格式版本 + 尾行 + 大小对得上**，不是"覆盖了多少字节"：raw 开头那段不带帧头的握手
// （`meta Connection …`）本来就不在索引里，拿覆盖率当判据会把每一份完整索引都判成不完整。
func (ix *FrameIndex) Complete() bool {
	if ix == nil || !ix.HasEnd || ix.Version != indexVersion {
		return false
	}
	if ix.RawSize > 0 && ix.End != ix.RawSize {
		return false
	}
	return true
}

// StaleFormat 报告这份索引的格式版本与当前代码不同（要重建，别直接用）。
func (ix *FrameIndex) StaleFormat() bool {
	return ix != nil && ix.Version != indexVersion
}

// Timed 报告这份索引有没有帧级时间（重建出来的没有）。
func (ix *FrameIndex) Timed() bool {
	if ix == nil {
		return false
	}
	for _, r := range ix.Recs {
		if r.T > 0 {
			return true
		}
	}
	return false
}

// First / Last 返回时间范围（没有帧时返回 0）。
func (ix *FrameIndex) First() int64 {
	if ix == nil || len(ix.Recs) == 0 {
		return 0
	}
	return ix.Recs[0].T
}

// Last 返回最后一帧的时刻。
func (ix *FrameIndex) Last() int64 {
	if ix == nil || len(ix.Recs) == 0 {
		return 0
	}
	return ix.Recs[len(ix.Recs)-1].T
}

// ByteRange 是 raw 文件里的一段字节。
type ByteRange struct {
	Off int64 `json:"off"`
	Len int64 `json:"len"`
}

// Noise 返回 raw 里**没有落进任何一条索引记录**的区间：开头的握手（`meta Connection …`
// 是整条流里唯一不带帧头的一段）、以及没认出帧头来的残块。
//
// 为什么不单独记它们：`off` 的跳跃已经隐含了它们（H9），再记一份就是第二个真相。
// 这里只是把跳跃显式算出来，省得每个读侧各写一遍。
//
// 注意：图片缓存（type=5）**不在**这里面 —— 它们有记录（`kind=cache`）。
func (ix *FrameIndex) Noise() []ByteRange {
	if ix == nil || len(ix.Recs) == 0 {
		return nil
	}
	var out []ByteRange
	var prev int64
	for _, r := range ix.Recs {
		if r.Off > prev {
			out = append(out, ByteRange{Off: prev, Len: r.Off - prev})
		}
		end := r.Off + int64(r.Len)
		if end > prev {
			prev = end
		}
	}
	if ix.RawSize > prev {
		out = append(out, ByteRange{Off: prev, Len: ix.RawSize - prev})
	}
	return out
}

// IndexBuild 是一次索引重建的结果。
type IndexBuild struct {
	Path    string `json:"path"`
	Frames  int    `json:"frames"`
	Bytes   int64  `json:"bytes"`
	Covered int64  `json:"covered"`
	Rebuilt bool   `json:"rebuilt"`
}

// BuildFrameIndex 从一个 raw 文件现场解析并落索引（老录制件补建，`tt recording index`）。
//
// 时间：老录制**没有**帧级时间 —— 那时候还没记。所以 `t` 一律留空，读侧照实说
// "时间精度：块级"。刻意不去拿 `timeline.log` 的块级时间"补"出一个看起来精确的时刻：
// 那不是精度问题，是编。
func BuildFrameIndex(rawPath string, conn, dir int, force bool) (IndexBuild, error) {
	out := IndexBuild{Path: IndexPathFor(rawPath)}
	if !force {
		if ix, err := ReadFrameIndex(out.Path); err == nil && ix.Complete() {
			out.Frames, out.Bytes, out.Covered = len(ix.Recs), ix.End, ix.Covered()
			return out, nil
		}
	}
	f, err := os.Open(rawPath)
	if err != nil {
		return out, err
	}
	defer f.Close()
	w, err := NewRebuiltFrameIndexWriter(out.Path, conn, dir)
	if err != nil {
		return out, err
	}
	buf := make([]byte, 64<<10)
	for {
		n, rerr := f.Read(buf)
		if n > 0 {
			// 重建没有真时刻可用：零值进去，`t` 就留空。
			w.Write(buf[:n], time.Time{})
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			w.Close()
			return out, rerr
		}
	}
	if err := w.Close(); err != nil {
		return out, err
	}
	out.Frames, out.Bytes, out.Covered, out.Rebuilt = w.Count(), w.Bytes(), w.Covered(), true
	return out, nil
}

// FrameKind 给一帧定类别，并尽量取它的序号与对象 id。
//
// 帧类型决定类别：AUI（0x01）是消息，按动词分类；type=5 是**图片缓存**，如实叫 `cache`；
// 其它没见过的类型照编号说出来（`typeN`）—— 绝不把它混进消息里冒充一条 AUI 消息（H9）。
//
// 读侧（明细页）与写侧（索引）**必须**用同一个函数：两边各写一套分类，界面上就会出现
// "索引说是 om、明细页说是 cache"这种没人能查的偏差。
func FrameKind(sp FrameSpan) (kind string, seq, idRef int) {
	switch sp.Type {
	case FrameTypeAUI:
		return classifyMessage(sp.Payload)
	case FrameTypeCache:
		return "cache", 0, 0
	default:
		return fmt.Sprintf("type%d", sp.Type), 0, 0
	}
}

// classifyMessage 把一条载荷归到一个类别，并尽量取它的序号与对象 id。
//
// 重同步规则与 TrimNoise 一致（跳过帧头噪声、找第一个动词）。**取不到就留空，绝不猜**：
//
//	om <n> {{…}}         → seq = n；idRef 留空 —— 一条 om 里可以有几十个对象，
//	                        取哪一个都是编（读侧要具体对象时自己解析载荷）。
//	un <id> {{…}}        → seq = idRef = id（这条消息指的就是这个对象）。
//	rn <id>              → 同上（删除也是"指向它"）。
//	event _om <n>{}{{…}} → seq = n；idRef = 载荷里**唯一**的那个 `idRef "<n>"`
//	                        （一批导航上报里有好几个 id 时留空，见 soleIDRef）。
//	meta <Type> {{…}}    → 握手/元信息，无序号无对象。
func classifyMessage(payload []byte) (kind string, seq, idRef int) {
	s := string(payload)
	i := 0
	for i < len(s) && !isLetter(s[i]) {
		i++
	}
	s = s[i:]
	for _, v := range msgVerbs {
		if !strings.HasPrefix(s, v) {
			continue
		}
		rest := s[len(v):]
		if rest != "" && !isSpace(rest[0]) && rest[0] != '{' {
			continue // `omx` 不是动词 `om`
		}
		switch v {
		case "meta":
			return "meta", 0, 0
		case "event":
			r := strings.TrimSpace(rest)
			r = strings.TrimSpace(strings.TrimPrefix(r, "_om"))
			if n, _, ok := numAt(r, 0); ok {
				seq = n
			}
			// 一条事件帧可以带**一批**事件（实测 `event _om 0` 里十来个 tabIndex/hidden
			// 上报）。这时"这条消息指向哪个对象"没有唯一答案 —— 取第一个就是编，
			// 所以只在**唯一**时才填（见 soleIDRef）。
			if id, ok := soleIDRef(s); ok {
				idRef = id
			}
			return "event", seq, idRef
		case "om":
			if n, _, ok := numAt(strings.TrimSpace(rest), 0); ok {
				seq = n
			}
			return "om", seq, 0
		case "un", "rn":
			if n, _, ok := numAt(strings.TrimSpace(rest), 0); ok {
				seq, idRef = n, n
			}
			return v, seq, idRef
		}
	}
	return "other", 0, 0
}

// msOf 把时刻转成 Unix 毫秒（零值 → 0，表示"不知道"）。
func msOf(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixMilli()
}

// IndexFilesFor 列出某份录制在各方向上的索引文件（存在才列），按文件名排序。
//
// raw 文件在、索引不在，是完全正常的状态（老件、或索引写失败）—— 调用方据此决定
// "现场重建"还是"降级到块级时间"，这里不替它决定。
func IndexFilesFor(dir, id string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() || !IsIndexFile(e.Name()) {
			continue
		}
		if got, ok := RecordingIDOfAny(e.Name()); !ok || (id != "" && got != id) {
			continue
		}
		out = append(out, filepath.Join(dir, e.Name()))
	}
	sort.Strings(out)
	return out, nil
}

// NoiseSummary 把噪声区间归并成"人看的一行"：段数与总字节。
//
// 为什么归并：图片缓存是**碎**的（帧与帧之间夹几十字节），一段段列出来没有信息量。
func NoiseSummary(rs []ByteRange) string {
	if len(rs) == 0 {
		return "无"
	}
	var total int64
	for _, r := range rs {
		total += r.Len
	}
	return fmt.Sprintf("%d 段 / %d 字节", len(rs), total)
}

// RawInfo 描述一份流量文件属于哪条连接的哪个方向。
type RawInfo struct {
	Path string
	Name string // 文件名
	Conn int
	Dir  int
}

// DirLabel 是方向的人读名字（与 `timeline.log` 的箭头同义）。
func DirLabel(dir int) string {
	if dir == DirRuntime {
		return "运行期→前端"
	}
	return "前端→运行期"
}

// RawFileInfo 从一个文件名切出连接号与方向。
//
// 为什么从文件名切、不另存一份账：连接号与方向**本来就编码在文件名里**（代理落盘时定的，
// 见 proxy.openLogs）。再记一份就是第二个真相，搬库/改名之后必然对不上。
func RawFileInfo(path string) (RawInfo, bool) {
	base := filepath.Base(path)
	var dir int
	var stem string
	switch {
	case strings.HasSuffix(base, "-rt2gdc.log"):
		dir, stem = DirRuntime, strings.TrimSuffix(base, "-rt2gdc.log")
	case strings.HasSuffix(base, "-gdc2rt.log"):
		dir, stem = DirFront, strings.TrimSuffix(base, "-gdc2rt.log")
	default:
		return RawInfo{}, false
	}
	i := strings.LastIndexByte(stem, '-')
	if i < 0 {
		return RawInfo{}, false
	}
	conn := 0
	for _, c := range stem[i+1:] {
		if c < '0' || c > '9' {
			return RawInfo{}, false
		}
		conn = conn*10 + int(c-'0')
	}
	if conn <= 0 {
		return RawInfo{}, false
	}
	return RawInfo{Path: path, Name: base, Conn: conn, Dir: dir}, true
}

// RawFilesFor 列出某份录制两个方向的流量文件（按文件名排序）。
func RawFilesFor(dir, id string) ([]RawInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []RawInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, ok := RawFileInfo(e.Name())
		if !ok {
			continue
		}
		if got, ok := RecordingIDOfAny(e.Name()); !ok || (id != "" && got != id) {
			continue
		}
		info.Path = filepath.Join(dir, e.Name())
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
