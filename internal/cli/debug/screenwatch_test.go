package debug

// 界面会话挂接变化 → 事件（W2）的判据。
//
// 一半是纯函数（快照 / 差异 / 投影），一半是那个循环（真的读写一个临时注册表）。
// 分开测的理由：纯函数定"变了什么"，循环只负责"什么时候看"。

import (
	"context"
	"testing"
	"time"

	"tt/internal/debug"
	"tt/internal/fep"
)

func entry(ref, tag string) screenEntry {
	return screenEntry{Ref: ref, Tag: tag, Kind: "rec", Env: "主机109", Prog: "axmt500_wf",
		GDCPid: 100, ProxyPort: 7600, InjectPort: 7899, OwnerPID: 200}
}

// 快照：按 ref 收表，跳过空 ref / nil（没有目标可指向的登记不该变成事件）。
func TestScreenSnapshotSkipsUntargeted(t *testing.T) {
	list := []*fep.Session{
		{Ref: "主机109#1", Tag: "fep-a", GDCPID: 1},
		nil,
		{Ref: "", Tag: "fep-无ref"},
		{Ref: "主机109#2", Tag: "fep-b", GDCPID: 2},
	}
	got := screenSnapshot(list)
	if len(got) != 2 {
		t.Fatalf("该只收两条带 ref 的，实得 %d 条：%+v", len(got), got)
	}
	if got["主机109#1"].Tag != "fep-a" || got["主机109#1"].GDCPid != 1 {
		t.Errorf("投影的字段不对：%+v", got["主机109#1"])
	}
}

// 差异：首次看到 = 挂上；消失 = 收掉；tag 变了 = 收掉旧的 + 挂上新的；没变 = 什么都不发。
func TestScreenChanges(t *testing.T) {
	base := map[string]screenEntry{"主机109#1": entry("主机109#1", "fep-a")}

	t.Run("首次看到全部算挂上", func(t *testing.T) {
		att, det := screenChanges(nil, base)
		if len(att) != 1 || len(det) != 0 {
			t.Fatalf("att=%v det=%v", att, det)
		}
	})
	t.Run("消失算收掉", func(t *testing.T) {
		att, det := screenChanges(base, nil)
		if len(att) != 0 || len(det) != 1 || det[0].Tag != "fep-a" {
			t.Fatalf("att=%v det=%v", att, det)
		}
	})
	t.Run("没变就不发", func(t *testing.T) {
		att, det := screenChanges(base, base)
		if len(att)+len(det) != 0 {
			t.Fatalf("同 tag 不该产生事件：att=%v det=%v", att, det)
		}
	})
	t.Run("同 ref 换 tag 要先收后挂", func(t *testing.T) {
		next := map[string]screenEntry{"主机109#1": entry("主机109#1", "fep-b")}
		att, det := screenChanges(base, next)
		if len(att) != 1 || att[0].Tag != "fep-b" {
			t.Fatalf("该挂上 fep-b：%v", att)
		}
		if len(det) != 1 || det[0].Tag != "fep-a" {
			t.Fatalf("该收掉 fep-a：%v", det)
		}
	})
	t.Run("新增与移除各自一条", func(t *testing.T) {
		next := map[string]screenEntry{"主机109#2": entry("主机109#2", "fep-c")}
		att, det := screenChanges(base, next)
		if len(att) != 1 || att[0].Ref != "主机109#2" || len(det) != 1 || det[0].Ref != "主机109#1" {
			t.Fatalf("att=%v det=%v", att, det)
		}
	})
}

// 投影：字段一个不少、reason 只透传（事件里 attached 不带 reason，detached 带）。
func TestToScreenState(t *testing.T) {
	st := toScreenState(entry("主机109#3", "fep-x"), "")
	if st.Ref != "主机109#3" || st.Tag != "fep-x" || st.ProxyPort != 7600 || st.OwnerPID != 200 {
		t.Fatalf("投影丢字段：%+v", st)
	}
	if st.Reason != "" {
		t.Errorf("attached 不该带 reason，实得 %q", st.Reason)
	}
	if got := toScreenState(entry("r", "t"), "stop").Reason; got != "stop" {
		t.Errorf("detached 该带 reason=stop，实得 %q", got)
	}
}

// 循环：真的写一份注册表，等它把 attached 报出来；删掉后报 detached。
//
// 这条是"接线层真的接上了"的判据 —— 上面那些纯函数通过、循环没接，等于没做。
func TestScreenWatchReportsRealRegistry(t *testing.T) {
	dir := t.TempDir()
	reg := fep.NewRegistry(dir)

	type ev struct {
		attached bool
		ref      string
		tag      string
	}
	got := make(chan ev, 8)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	StartScreenWatch(ctx, dir, func(attached bool, st debug.ScreenState) {
		got <- ev{attached: attached, ref: st.Ref, tag: st.Tag}
	})

	waitFor := func(wantAttached bool) ev {
		t.Helper()
		deadline := time.After(3 * time.Second)
		for {
			select {
			case e := <-got:
				if e.attached == wantAttached {
					return e
				}
			case <-deadline:
				t.Fatalf("3 秒内没等到 attached=%v 的事件", wantAttached)
			}
		}
	}

	if err := reg.Save(&fep.Session{Tag: "fep-watch-1", Ref: "主机109#9", Env: "主机109",
		Prog: "axmt500_wf", Kind: "rec", ProxyPort: 7600, InjectPort: 7899}); err != nil {
		t.Fatalf("写注册表失败：%v", err)
	}
	e := waitFor(true)
	if e.ref != "主机109#9" || e.tag != "fep-watch-1" {
		t.Fatalf("attached 事件不对：%+v", e)
	}

	if err := reg.Remove("fep-watch-1"); err != nil {
		t.Fatalf("摘注册失败：%v", err)
	}
	d := waitFor(false)
	if d.tag != "fep-watch-1" {
		t.Fatalf("detached 事件不对：%+v", d)
	}
}
