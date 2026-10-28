package debug

// 常驻直连池的契约：命中复用、空闲回收、出错摘除、配置热替换整池作废。
// 全部离线：池里放的是不连库的令牌（&DirectQuery{}），只验证"复用/回收"的簿记。

import (
	"context"
	"sync"
	"testing"
	"time"
)

func token() *DirectQuery { return &DirectQuery{} }

func TestDBDirectPoolReusesSameConn(t *testing.T) {
	p := newDBDirectPool()
	opened := 0
	open := func() (*DirectQuery, error) { opened++; return token(), nil }

	a, err := p.get("k", open)
	if err != nil {
		t.Fatalf("不该报错: %v", err)
	}
	b, err := p.get("k", open)
	if err != nil {
		t.Fatalf("不该报错: %v", err)
	}
	if opened != 1 || a != b {
		t.Fatalf("同一 key 应复用同一条连接（open 次数 %d）", opened)
	}
	// 不同 key 各自一条
	if c, _ := p.get("k2", open); c == a || opened != 2 {
		t.Fatalf("不同 key 不能共用连接")
	}
}

func TestDBDirectPoolReopensAfterIdle(t *testing.T) {
	p := newDBDirectPool()
	p.ttl = time.Millisecond
	opened := 0
	open := func() (*DirectQuery, error) { opened++; return token(), nil }

	if _, err := p.get("k", open); err != nil {
		t.Fatalf("不该报错: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	if _, err := p.get("k", open); err != nil {
		t.Fatalf("不该报错: %v", err)
	}
	if opened != 2 {
		t.Errorf("空闲超时应重连（open 次数 %d，期望 2）", opened)
	}
}

func TestDBDirectPoolDropClosesAndForgets(t *testing.T) {
	p := newDBDirectPool()
	q, _ := p.get("k", func() (*DirectQuery, error) { return token(), nil })
	p.drop("k", q)
	if e := p.m["k"]; e != nil {
		t.Fatal("drop 之后池里不该还留着它")
	}
	// 再取一次必须是新连接
	q2, _ := p.get("k", func() (*DirectQuery, error) { return token(), nil })
	if q2 == q {
		t.Fatal("被摘掉的连接不能又被发出来")
	}
}

func TestDBDirectPoolDropOnlyRemovesMatchingConn(t *testing.T) {
	// 并发下别的请求可能刚把 key 换成新连接：旧连接出错的 drop 不能连新的也删掉。
	p := newDBDirectPool()
	old, _ := p.get("k", func() (*DirectQuery, error) { return token(), nil })
	fresh, _ := p.get("k2", func() (*DirectQuery, error) { return token(), nil })
	p.m["k"] = &dbPoolEntry{q: fresh, lastUsed: time.Now()} // 假装别处换过
	p.drop("k", old)
	if e := p.m["k"]; e == nil || e.q != fresh {
		t.Fatal("drop 只该摘掉出错的这一条，不该误删新连接")
	}
}

func TestDBDirectPoolCloseAll(t *testing.T) {
	p := newDBDirectPool()
	for _, k := range []string{"a", "b"} {
		if _, err := p.get(k, func() (*DirectQuery, error) { return token(), nil }); err != nil {
			t.Fatalf("不该报错: %v", err)
		}
	}
	p.closeAll()
	if len(p.m) != 0 {
		t.Errorf("closeAll 之后池应为空，剩 %d 条", len(p.m))
	}
}

func TestDBDirectPoolConcurrentOpenKeepsOne(t *testing.T) {
	p := newDBDirectPool()
	var opened int
	var mu sync.Mutex
	open := func() (*DirectQuery, error) {
		mu.Lock()
		opened++
		mu.Unlock()
		time.Sleep(5 * time.Millisecond) // 让并发请求都走到"自己开一条"
		return token(), nil
	}
	var wg sync.WaitGroup
	got := make([]*DirectQuery, 8)
	for i := range got {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got[i], _ = p.get("k", open)
		}(i)
	}
	wg.Wait()
	if len(p.m) != 1 {
		t.Fatalf("并发后池里应只有一条，实际 %d", len(p.m))
	}
	for i, q := range got {
		if q != got[0] {
			t.Errorf("第 %d 个调用拿到的不是同一条连接", i)
		}
	}
}

func TestDirectSQLFallsBackWhenNoDBConfigured(t *testing.T) {
	// 没挂库（测试里构造的 Server 就是这种）：直连这条路必须说"我不回答"，
	// 让调用方走原来的 SSH 路 —— 否则 /api/dbsql 的拨号失败契约会变。
	s := &Server{cfg: &Config{}, dbPool: newDBDirectPool()}
	if _, _, handled := s.directSQL(context.Background(), 99, 30, "select 1 from dual"); handled {
		t.Fatal("没挂库时不该由直连回答")
	}
}

func TestDirectSQLFallsBackWhenPoolMissing(t *testing.T) {
	// 直接字面量构造的 Server 可能没有池（老测试）：不能空指针崩，应当回退。
	s := &Server{cfg: &Config{}}
	if _, _, handled := s.directSQL(context.Background(), 99, 30, "select 1 from dual"); handled {
		t.Fatal("没有池时不该由直连回答")
	}
}
