package fep

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 帧索引（计划 C 的 D0）。三条判据都在这里钉住：
//  1. 跨块的帧照样定位准确（`off/len` 能把原字节取回来）；
//  2. 索引与 raw 一对一，重建出来的帧数与现场解析一致；
//  3. 索引残缺（尾行没落盘）时**不报错**，只是照实标"不完整"。

func TestOffsetScannerLocatesFramesAcrossChunks(t *testing.T) {
	a := Encode(Frame{Type: FrameTypeAUI, Payload: []byte(`om 1 {{an 0 Window 5 {{}}}}`)})
	b := Encode(Frame{Type: FrameTypeAUI, Payload: []byte(`event _om 2{}{{ActionEvent 0{{idRef "42"}}}}`)})
	// 帧与帧之间夹"图片缓存"那类非帧字节（H9）：它是字节，占位置。
	noise := []byte{0x00, 0x00, 0x00, 0x03, 0xde, 0xad, 0xbe}
	stream := append(append(append(append([]byte{}, a...), noise...), b...), 0x01)

	// 一帧拆成三段喂进去：只该在收齐那一次报出来，而且位置是**帧首**。
	sc := NewOffsetScanner()
	cut := len(a) / 3
	first := sc.Feed(stream[:cut], time.UnixMilli(1000))
	if len(first) != 0 {
		t.Fatalf("半帧不该报出来：%+v", first)
	}
	second := sc.Feed(stream[cut:2*cut], time.UnixMilli(2000))
	if len(second) != 0 {
		t.Fatalf("还是半帧：%+v", second)
	}
	rest := sc.Feed(stream[2*cut:], time.UnixMilli(3000))

	got := append(append(first, second...), rest...)
	if len(got) != 2 {
		t.Fatalf("该解出 2 帧，得到 %d：%+v", len(got), got)
	}
	if got[0].Off != 0 || got[0].Len != len(a) {
		t.Fatalf("第一帧位置不对：off=%d len=%d（帧长 %d）", got[0].Off, got[0].Len, len(a))
	}
	wantOff := int64(len(a) + len(noise))
	if got[1].Off != wantOff || got[1].Len != len(b) {
		t.Fatalf("第二帧位置不对：off=%d len=%d（该是 off=%d len=%d）", got[1].Off, got[1].Len, wantOff, len(b))
	}
	if sc.Buffered() != 1 {
		t.Fatalf("stream 尾巴上那个孤字节该留着：Buffered=%d", sc.Buffered())
	}
	// 帧时间取**首块**时刻（H8）：第一帧在 1000 那次开始，第二帧在 3000 那次才见到。
	if got[0].T.UnixMilli() != 1000 {
		t.Fatalf("第一帧时间该是首块的 1000，得到 %d", got[0].T.UnixMilli())
	}
	if got[1].T.UnixMilli() != 3000 {
		t.Fatalf("第二帧时间该是 3000，得到 %d", got[1].T.UnixMilli())
	}
	// 取原字节必须能原样解回来（检查器就是靠这一条去读 raw 的）。
	for _, sp := range got {
		raw := stream[sp.Off : sp.Off+int64(sp.Len)]
		frames, err := NewDecoder().Feed(raw)
		if err != nil || len(frames) != 1 {
			t.Fatalf("off/len 切出来的字节解不出一帧：%v %v", frames, err)
		}
		if string(frames[0].Payload) != string(sp.Payload) {
			t.Fatalf("取回来的载荷不一样：%q vs %q", frames[0].Payload, sp.Payload)
		}
	}
}

// 老录制件里"长度不含换行"的写法：换行落在帧外时，它也属于这一帧的字节 ——
// 不算进来的话 `off+len` 会停在换行前，读侧取原字节就少一个。
func TestOffsetScannerCountsTrailingNewlineOutsideFrame(t *testing.T) {
	payload := []byte("om 3 {{}}")
	wire := make([]byte, 0, 12)
	n := len(payload) // 故意不把换行算进长度（旧写法）
	wire = append(wire, byte(n>>24), byte(n>>16), byte(n>>8), byte(n), byte(n>>24), byte(n>>16), byte(n>>8), byte(n), FrameTypeAUI)
	wire = append(wire, payload...)
	wire = append(wire, '\n')
	stream := append(wire, Encode(Frame{Type: FrameTypeAUI, Payload: []byte("om 4 {{}}")})...)

	spans := DecodeOffsets(stream)
	if len(spans) != 2 {
		t.Fatalf("该解出 2 帧，得到 %d", len(spans))
	}
	if spans[0].Len != len(wire) {
		t.Fatalf("第一帧该把帧外的换行算进去：len=%d 该是 %d", spans[0].Len, len(wire))
	}
	if spans[1].Off != int64(len(wire)) {
		t.Fatalf("第二帧起点该紧接在第一帧之后：off=%d 该是 %d", spans[1].Off, len(wire))
	}
}

func TestFrameIndexRoundTripAndNoise(t *testing.T) {
	dir := t.TempDir()
	raw := filepath.Join(dir, "fep-1-1-rt2gdc.log")
	idx := IndexPathFor(raw)
	if filepath.Base(idx) != "fep-1-1-rt2gdc.frames.jsonl" {
		t.Fatalf("索引该与 raw 同一个干：%s", filepath.Base(idx))
	}

	w, err := NewFrameIndexWriter(idx, 1, DirRuntime)
	if err != nil {
		t.Fatal(err)
	}
	// 一段：噪声 + om 帧 + event 帧（event 帧里带 idRef）
	chunk1 := append([]byte{0xaa, 0xbb}, Encode(Frame{Type: FrameTypeAUI, Payload: []byte(`om 7 {{an 0 Window 5 {{}}}}`)})...)
	chunk2 := Encode(Frame{Type: FrameTypeAUI, Payload: []byte(`event _om 8{}{{ActionEvent 0{{idRef "5348"}}}}`)})
	chunk2 = append(chunk2, 0x00, 0x01, 0x02) // 尾巴上又是噪声
	w.Write(chunk1, time.UnixMilli(5000))
	w.Write(chunk2, time.UnixMilli(6500))
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	// raw 文件由代理写；这里模拟它，好让"覆盖 vs 文件大小"能对上。
	if err := os.WriteFile(raw, append(chunk1, chunk2...), 0o644); err != nil {
		t.Fatal(err)
	}

	ix, err := ReadFrameIndex(idx)
	if err != nil {
		t.Fatal(err)
	}
	if len(ix.Recs) != 2 || ix.Bad != 0 {
		t.Fatalf("该读回 2 帧 0 坏行：%d / %d", len(ix.Recs), ix.Bad)
	}
	if !ix.Complete() {
		t.Fatalf("尾行 + 大小都对得上，该判完整（End=%d RawSize=%d）", ix.End, ix.RawSize)
	}
	if !ix.Timed() {
		t.Fatal("录制时落的索引该带帧时间")
	}
	if ix.Recs[0].Kind != "om" || ix.Recs[0].Seq != 7 || ix.Recs[0].T != 5000 {
		t.Fatalf("第一条记录不对：%+v", ix.Recs[0])
	}
	if ix.Recs[1].Kind != "event" || ix.Recs[1].Seq != 8 || ix.Recs[1].IDRef != 5348 {
		t.Fatalf("event 那条该带上载荷里的 idRef：%+v", ix.Recs[1])
	}
	// om 那条**不填** idRef：一条 om 里可以有几十个对象，取哪个都是编。
	if ix.Recs[0].IDRef != 0 {
		t.Fatalf("om 帧不该假装有一个对象 id：%d", ix.Recs[0].IDRef)
	}
	// 噪声：头部 2 字节 + 尾部 3 字节
	noise := ix.Noise()
	var total int64
	for _, r := range noise {
		total += r.Len
	}
	if total != 5 || len(noise) != 2 {
		t.Fatalf("噪声该是两段共 5 字节：%v", noise)
	}
	// 索引名与 raw 名互相认得出来（搬库靠它认"这份的东西"）
	if got, ok := RecordingIDOfAny(filepath.Base(idx)); !ok || got != "fep-1" {
		t.Fatalf("索引文件名该能切出录制 ID：%q %v", got, ok)
	}
	if name, ok := RawOfIndex(filepath.Base(idx)); !ok || name != "fep-1-1-rt2gdc.log" {
		t.Fatalf("索引该能换回 raw 名：%q %v", name, ok)
	}
}

// 进程被硬杀时尾行不会落盘：这时**不能**当完整索引用（那会让界面在某个时刻突然断掉
// 却不说为什么），但也不该报错 —— raw 还是好的，重建就行。
func TestFrameIndexWithoutTrailerIsIncompleteNotBroken(t *testing.T) {
	dir := t.TempDir()
	raw := filepath.Join(dir, "fep-2-1-rt2gdc.log")
	body := Encode(Frame{Type: FrameTypeAUI, Payload: []byte(`om 1 {{}}`)})
	if err := os.WriteFile(raw, body, 0o644); err != nil {
		t.Fatal(err)
	}
	// 手工造一份"只有头行 + 一帧、没有尾行"的索引
	head := `{"v":1,"conn":1,"dir":1,"raw":"fep-2-1-rt2gdc.log"}` + "\n"
	rec := `{"t":1000,"conn":1,"dir":1,"off":0,"len":` + itoa(len(body)) + `,"seq":1,"kind":"om"}` + "\n"
	if err := os.WriteFile(IndexPathFor(raw), []byte(head+rec), 0o644); err != nil {
		t.Fatal(err)
	}

	ix, err := ReadFrameIndex(IndexPathFor(raw))
	if err != nil {
		t.Fatalf("缺尾行不是错误：%v", err)
	}
	if len(ix.Recs) != 1 {
		t.Fatalf("已写下的帧照样要能读：%d", len(ix.Recs))
	}
	if ix.Complete() {
		t.Fatal("没有尾行就是不可信为完整")
	}
	// 一个字节都不认得的文件：ErrNoIndex，而不是空索引（两者对读侧意义不同）
	if _, err := ReadFrameIndex(filepath.Join(dir, "不存在.frames.jsonl")); err != ErrNoIndex {
		t.Fatalf("缺索引该回 ErrNoIndex，得到 %v", err)
	}
}

// 老录制件现场重建：帧数与"重新解析"一致，且**不编时间**。
func TestBuildFrameIndexFromLegacyRaw(t *testing.T) {
	dir := t.TempDir()
	raw := filepath.Join(dir, "fep-3-2-gdc2rt.log")
	var body []byte
	body = append(body, []byte("meta Connection {{compression \"none\"}}\n")...) // 不带帧头的握手
	body = append(body, Encode(Frame{Type: FrameTypeAUI, Payload: []byte(`un 0 {{currentWindow "119"}}`)})...)
	body = append(body, []byte{0x11, 0x22, 0x33}...) // 图片缓存
	body = append(body, Encode(Frame{Type: FrameTypeAUI, Payload: []byte(`event _om 12{}{{KeyEvent 0{{keyName "Tab"}}}}`)})...)
	if err := os.WriteFile(raw, body, 0o644); err != nil {
		t.Fatal(err)
	}
	build, err := BuildFrameIndex(raw, 2, DirFront, false)
	if err != nil {
		t.Fatal(err)
	}
	if build.Frames != 2 || !build.Rebuilt {
		t.Fatalf("重建该找到 2 帧并标 rebuilt：%+v", build)
	}
	ix, err := ReadFrameIndex(build.Path)
	if err != nil {
		t.Fatal(err)
	}
	if want := len(DecodeOffsets(body)); len(ix.Recs) != want {
		t.Fatalf("重建帧数 %d 与现场解析 %d 不一致", len(ix.Recs), want)
	}
	if ix.Timed() {
		t.Fatal("重建的索引不许有帧时间（老件没有帧级时间，编一个出来更坏）")
	}
	if !ix.Rebuilt {
		t.Fatal("重建出来的索引该标出来")
	}
	// 第二次调用不重建（已经有了）
	again, err := BuildFrameIndex(raw, 2, DirFront, false)
	if err != nil || again.Rebuilt {
		t.Fatalf("已有完整索引时不该重建：%+v %v", again, err)
	}
	// --force 才重建
	forced, err := BuildFrameIndex(raw, 2, DirFront, true)
	if err != nil || !forced.Rebuilt {
		t.Fatalf("force 该重建：%+v %v", forced, err)
	}
}

// 真实录制件上的对齐回归：索引里每一条的 `off/len` 必须能把**原字节**取回来，
// 而且取回来的那一帧与"现场解析"得到的是同一帧。
//
// 为什么必须钉在真件上：`off/len` 一旦算错，检查器会显示**别人的**字节 —— 而合成的
// 短流很难暴露这种错（真件里夹着图片缓存那类二进制、帧跨块、还有上万帧）。
//
// 现场语料是可选的（含真实业务数据，不随仓库走），所以这段**不跳过**：
// 没有语料就用合成流 + 仓库夹具（文本形态）跑，并在日志里说清这一轮到底跑了什么。
func TestFrameIndexAlignsWithRealRecordingBytes(t *testing.T) {
	// 1) 仓库夹具是**文本形态**（早期原型一条消息一行，没有帧头）：
	//    它必须老老实实报"0 帧"，而不是硬从文本里"找出"帧来。
	for _, f := range fixtureFiles(t) {
		tmp := filepath.Join(t.TempDir(), filepath.Base(f))
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(tmp, b, 0o644); err != nil {
			t.Fatal(err)
		}
		build, err := BuildFrameIndex(tmp, 1, DirRuntime, true)
		if err != nil {
			t.Fatal(err)
		}
		if build.Frames != 0 {
			t.Errorf("%s 是文本形态的录制件，不该「找出」%d 帧", filepath.Base(f), build.Frames)
		}
		// 但消息照样读得出来（降级路径：文本形态走 ReadRecording）。
		if bodies, err := ReadRecording(tmp); err != nil || len(bodies) == 0 {
			t.Errorf("%s 的文本形态读不出消息：%v", filepath.Base(f), err)
		}
	}

	// 2) 现场语料：每个方向的 raw 逐帧对齐。
	dir := recordDir(t)
	raws, err := RawFilesFor(dir, "")
	if err != nil || len(raws) == 0 {
		t.Logf("没有现场语料（%s）：只跑了合成流与文本夹具", dir)
		return
	}
	checked, frames, cache := 0, 0, 0
	for _, raw := range raws {
		n, c := verifyIndexAgainstRaw(t, raw)
		checked++
		frames += n
		cache += c
	}
	t.Logf("现场语料逐帧对齐：%d 个方向 / 共 %d 帧（其中图片缓存 type=5 共 %d 帧）", checked, frames, cache)
}

// verifyIndexAgainstRaw 对一个真实 raw 文件逐帧验证索引的 `off/len`，
// 返回（帧数，其中图片缓存帧数）。
func verifyIndexAgainstRaw(t *testing.T, raw RawInfo) (int, int) {
	t.Helper()
	blob, err := os.ReadFile(raw.Path)
	if err != nil {
		t.Fatal(err)
	}
	want := DecodeOffsets(blob)
	// 索引落在 raw 旁边，而语料目录不一定让写（沙箱/权限）—— 拷到临时目录里建。
	tmp := filepath.Join(t.TempDir(), raw.Name)
	if err := os.WriteFile(tmp, blob, 0o644); err != nil {
		t.Fatal(err)
	}
	build, err := BuildFrameIndex(tmp, raw.Conn, raw.Dir, true)
	if err != nil {
		t.Fatalf("建索引失败 %s: %v", raw.Name, err)
	}
	ix, err := ReadFrameIndex(build.Path)
	if err != nil {
		t.Fatal(err)
	}
	if len(ix.Recs) != len(want) {
		t.Fatalf("%s：索引 %d 帧，现场解析 %d 帧 —— 两边必须一致", raw.Name, len(ix.Recs), len(want))
	}
	if !ix.Complete() {
		t.Errorf("%s：刚建完的索引该是完整的（HasEnd=%v End=%d RawSize=%d）", raw.Name, ix.HasEnd, ix.End, ix.RawSize)
	}
	cache := 0
	for i, rec := range ix.Recs {
		sp := want[i]
		if rec.Off != sp.Off || rec.Len != sp.Len {
			t.Fatalf("%s 第 %d 帧：索引 off/len = %d/%d，现场 = %d/%d",
				raw.Name, i, rec.Off, rec.Len, sp.Off, sp.Len)
		}
		// 图片缓存帧如实叫 cache；AUI 帧不许出现 cache 这个类别（那是两条泳道）。
		if sp.Type == FrameTypeCache {
			cache++
			if rec.Kind != "cache" {
				t.Fatalf("%s 第 %d 帧是 type=5（图片缓存），kind 该是 cache，得到 %q", raw.Name, i, rec.Kind)
			}
		} else if rec.Kind == "cache" {
			t.Fatalf("%s 第 %d 帧是 type=1，不该被叫成 cache", raw.Name, i)
		}
		end := rec.Off + int64(rec.Len)
		if end > int64(len(blob)) {
			t.Fatalf("%s 第 %d 帧：off+len 越界（%d > %d）", raw.Name, i, end, len(blob))
		}
		// 拿索引给的字节回去解一次：必须解出**同一帧**（检查器就是靠这一条读 raw 的）。
		got, derr := NewDecoder().Feed(blob[rec.Off:end])
		if derr != nil || len(got) != 1 {
			t.Fatalf("%s 第 %d 帧：off/len 切出来的字节解不出唯一一帧（%d 帧，err=%v）", raw.Name, i, len(got), derr)
		}
		if string(got[0].Payload) != string(sp.Payload) {
			t.Fatalf("%s 第 %d 帧：取回来的载荷不是同一帧\n  索引切出：%.60q\n  现场解析：%.60q",
				raw.Name, i, got[0].Payload, sp.Payload)
		}
	}
	// 字节账必须对得上：索引覆盖的字节 = 全部帧的字节和；差额就是没认出来的那些。
	if cov := ix.Covered(); cov > int64(len(blob)) {
		t.Fatalf("%s：索引覆盖 %d 字节，超过文件大小 %d", raw.Name, cov, len(blob))
	}
	return len(ix.Recs), cache
}

// 分类：取不到就留空，绝不猜（这条是索引侧最容易被"顺手取一个"的地方）。
func TestClassifyMessage(t *testing.T) {
	cases := []struct {
		body string
		kind string
		seq  int
		id   int
	}{
		{`om 34 {{an 4533 Dialog 3533 {{}}}}`, "om", 34, 0},
		{`un 0 {{currentWindow "119"}}`, "un", 0, 0},
		{`rn 4533 {}`, "rn", 4533, 4533},
		{`event _om 12{}{{ActionEvent 0{{idRef "5348"}}}}`, "event", 12, 5348},
		{`event _om 13{}{{KeyEvent 0{{keyName "Tab"}}}}`, "event", 13, 0},
		// 一批导航上报里有好几个不同的 id：没有唯一答案 → 留空，**绝不取第一个**
		// （实测 `event _om 0` 就是这样：十来个 tabIndex/hidden，id 各不相同）
		{`event _om 0{}{{ConfigureEvent 0{{idRef "960"}}{{tabIndex "1"}}}}{{ConfigureEvent 0{{idRef "4540"}}{{tabIndex "2"}}}}`,
			"event", 0, 0},
		// 同一个 id 出现多次：唯一，可以填
		{`event _om 1{}{{ConfigureEvent 0{{idRef "42"}}{{a "1"}}}}{{ConfigureEvent 0{{idRef "42"}}{{b "2"}}}}`,
			"event", 1, 42},
		{`meta Connection {{compression "none"}}`, "meta", 0, 0},
		{"\x00\x00\x00\x0b\x00\x00\x00\x0b\x01om 5 {{}}", "om", 5, 0}, // 帧头残渣在载荷里也要认出来
		{`omx 5 {{}}`, "other", 0, 0},                                 // `omx` 不是动词 om
		{`{an 0 Window 5 {{}}}`, "other", 0, 0},                       // 没有动词：归类为"其它"
	}
	for _, c := range cases {
		kind, seq, id := classifyMessage([]byte(c.body))
		if kind != c.kind || seq != c.seq || id != c.id {
			t.Errorf("%q → (%s,%d,%d)，该是 (%s,%d,%d)", c.body, kind, seq, id, c.kind, c.seq, c.id)
		}
	}
}
