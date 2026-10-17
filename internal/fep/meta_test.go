package fep

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 录制件的身份文件：它必须能跨进程、跨时间地说明"这份录制是谁录的"。

func TestRecordingIDFromFileName(t *testing.T) {
	cases := []struct {
		name string
		want string
		ok   bool
	}{
		{"fep-20261015-170031-1-rt2gdc.log", "fep-20261015-170031", true}, // 连接号是数字 → 去掉
		{"fep-20261015-170031-12-gdc2rt.log", "", false},                  // 不是 AUI 侧
		{"r1-rt2gdc.log", "r1", true},                                     // 没有连接号：整段当 ID
		{"2026-rt2gdc.log", "2026", true},
		{"随便.txt", "", false},
	}
	for _, c := range cases {
		got, ok := RecordingIDFromFile(c.name)
		if got != c.want || ok != c.ok {
			t.Errorf("RecordingIDFromFile(%q) = (%q,%v)，应为 (%q,%v)", c.name, got, ok, c.want, c.ok)
		}
	}
}

// 拿一份录制去比对时，"前缀撞上另一个会话"是不能容忍的错答案 ——
// 这条钉的就是"按切出来的 ID 精确匹配"，而不是 LoadModelFrom 那种前缀匹配。
func TestRecordingFilesForMatchesExactID(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{
		"fep-20261015-170031-1-rt2gdc.log",
		"fep-20261015-170031-2-rt2gdc.log",
		"fep-20261015-170031-1-gdc2rt.log",
		"fep-20261015-17003-1-rt2gdc.log", // 少一位：前缀会撞上，精确匹配不会
	} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	files, err := RecordingFilesFor(dir, "fep-20261015-170031")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("只该取这一份的两条连接，got %v", files)
	}
	for _, f := range files {
		if got, _ := RecordingIDFromFile(filepath.Base(f)); got != "fep-20261015-170031" {
			t.Fatalf("撞上了别的会话：%s", f)
		}
	}
	ids, err := ListIDs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("目录里是两份录制（170031 与 17003），got %v", ids)
	}
}

func TestMetaRoundTripAndErrNoMeta(t *testing.T) {
	dir := t.TempDir()
	mt := RecordingMeta{
		ID: "fep-1", Path: PathDebug, Prog: "bsft001_wf",
		Env: "主机109", Zone: "36", Ent: "99", Host: "10.1.25.33",
		FGLServer:   "192.168.14.193:1200",
		Started:     time.Date(2026, 10, 15, 17, 0, 31, 0, time.UTC),
		Finished:    time.Date(2026, 10, 15, 17, 4, 0, 0, time.UTC),
		Binary:      BinaryID{Ref: "/u1/t35prd/erp/asf/42r/bsft001_wf.42r", Size: 310855, ModTime: "2026-09-24T10:06:28"},
		Conns:       []MetaConn{{ID: 1, RT2GDC: "x-1-rt2gdc.log", BytesRT: 272515, BytesFE: 33167}},
		Fingerprint: Fingerprint{Total: 3976, Hash: "493b95ef0b0df2af"},
	}
	path, err := WriteMeta(dir, mt)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "fep-1"+MetaSuffix {
		t.Fatalf("身份文件落点不对：%s", path)
	}
	got, err := ReadMeta(dir, "fep-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != PathDebug || got.Binary != mt.Binary || got.Fingerprint != mt.Fingerprint {
		t.Fatalf("读回来的身份不是写进去的：%+v", got)
	}
	if got.Duration() != 209*time.Second {
		t.Fatalf("时长应该算得出来：%s", got.Duration())
	}

	if _, err := ReadMeta(dir, "没有这一份"); !errors.Is(err, ErrNoMeta) {
		t.Fatalf("没有身份文件要报 ErrNoMeta（不是读失败）：%v", err)
	}
	if _, err := WriteMeta(dir, RecordingMeta{}); err == nil {
		t.Fatal("没有 ID 的身份文件必须被拒：它落下去之后没人找得到")
	}
}

// 收工落身份：指纹要从**录制下来的字节**算，不是从内存里的模型。
// 这条直接调 Runtime.writeMeta（不起实例、不连 SSH），所以谁都能跑。
func TestWriteMetaComputesFingerprintFromRecordings(t *testing.T) {
	dir := t.TempDir()
	const tag = "fep-20261015-170031"
	// 文本形态的 AUI 录制件（一条消息一行，行首是 `om <n> {{…}}`，与真夹具同形）——
	// ReadRecording 两种形态都吃
	body := "om 3 {{" + demoAUI() + "}}"
	if err := os.WriteFile(filepath.Join(dir, tag+"-1-rt2gdc.log"), []byte(body+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := &Runtime{
		cfg:       OpenConfig{Tag: tag, RecordDir: dir, Prog: "bsft001_wf", Path: PathRunnerRR, Env: "主机109", Zone: "36", Ent: "99"},
		Session:   &Session{Started: time.Now().Add(-time.Minute)},
		fglServer: "192.168.14.193:1200",
		// 编译件标识归 Job 管（P4）：只填它需要的那两个字段即可。
		job: &Job{binary: BinaryID{Ref: "/u1/t35prd/erp/asf/42r/bsft001_wf.42r", Size: 310855}},
	}
	if _, err := r.writeMeta([]ConnInfo{{ID: 1, BytesRT: 100, BytesFE: 50}}); err != nil {
		t.Fatal(err)
	}
	mt, err := ReadMeta(dir, tag)
	if err != nil {
		t.Fatal(err)
	}
	if mt.Fingerprint.Total == 0 || mt.Fingerprint.Hash == "" {
		t.Fatalf("指纹没从录制件算出来：%+v", mt.Fingerprint)
	}
	m := NewModel()
	m.ApplyMessage(body)
	if want := FingerprintOf(EntriesOf(m)); mt.Fingerprint != want {
		t.Fatalf("指纹该等于从同一份字节建出来的表：%+v vs %+v", mt.Fingerprint, want)
	}
	if mt.Path != PathRunnerRR || mt.FGLServer != "192.168.14.193:1200" || mt.Binary.Size != 310855 {
		t.Fatalf("身份里的路径/接线/编译件都该写进去：%+v", mt)
	}
	if len(mt.Conns) != 1 || mt.Conns[0].RT2GDC == "" || mt.Conns[0].GDC2RT == "" {
		t.Fatalf("连接清单要带两侧的文件落点：%+v", mt.Conns)
	}
}

// 一条 AUI 都没录到时，身份要**这么说**（而不是留一个空指纹假装录到了）。
func TestWriteMetaSaysWhenNothingWasRecorded(t *testing.T) {
	dir := t.TempDir()
	r := &Runtime{
		cfg:     OpenConfig{Tag: "fep-empty", RecordDir: dir, Path: PathNone},
		Session: &Session{Started: time.Now()},
	}
	if _, err := r.writeMeta(nil); err != nil {
		t.Fatal(err)
	}
	mt, err := ReadMeta(dir, "fep-empty")
	if err != nil {
		t.Fatal(err)
	}
	if mt.Fingerprint.Total != 0 || mt.Note == "" {
		t.Fatalf("没录到界面就必须说出来：%+v", mt)
	}
}
