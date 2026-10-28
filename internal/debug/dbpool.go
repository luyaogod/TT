package debug

// 常驻进程里的直连会话池（只服务 tt serve；CLI 一条命令一个进程，不适用）。
//
// 要解决的问题：`/api/dbsql` 每个请求都重新拨 SSH、并且走服务器 sqlplus（实测
// 1.3 s/条）。而 tt serve 是长期活着的进程 —— 把"建连 + 登录"留在进程里，之后的
// 查询只付执行时间（同一台机实测 10–30 ms）。
//
// 边界（刻意）：
//   - **不做健康检查**：多一次往返就把省下的时间还回去了。连接死了由查询报错暴露，
//     出错那条从池里摘掉，下一次请求重连。
//   - **不存口令**：池里只有连接本身，凭据仍在配置里，随配置热替换整池作废。
//   - 空闲超时回收（默认 5 分钟）：比探测缓存长，因为"登录"比"探测路径"贵。

import (
	"sync"
	"time"
)

// dbDirectIdleTTL 会话空闲多久回收。
const dbDirectIdleTTL = 5 * time.Minute

type dbPoolEntry struct {
	q        *DirectQuery
	lastUsed time.Time
}

// dbDirectPool 按 key（环境|企业|账号）复用直连会话。
type dbDirectPool struct {
	mu  sync.Mutex
	m   map[string]*dbPoolEntry
	ttl time.Duration
}

func newDBDirectPool() *dbDirectPool {
	return &dbDirectPool{m: make(map[string]*dbPoolEntry), ttl: dbDirectIdleTTL}
}

// get 取一条可复用的会话：命中且未空闲超时就直接给，否则用 open 新建。
func (p *dbDirectPool) get(key string, open func() (*DirectQuery, error)) (*DirectQuery, error) {
	p.mu.Lock()
	if e := p.m[key]; e != nil {
		if time.Since(e.lastUsed) < p.ttl {
			e.lastUsed = time.Now()
			q := e.q
			p.mu.Unlock()
			return q, nil
		}
		// 空闲超时：先摘掉再关，别让新请求拿到一条已经关掉的连接。
		delete(p.m, key)
		p.mu.Unlock()
		e.q.Close()
		return p.openAndStore(key, open)
	}
	p.mu.Unlock()
	return p.openAndStore(key, open)
}

// openAndStore 新建并登记；登记时若已有别处放进来的（并发），用先到的那条，
// 自己这条关掉 —— 否则同一 key 会攒下多条连接，越用越多。
func (p *dbDirectPool) openAndStore(key string, open func() (*DirectQuery, error)) (*DirectQuery, error) {
	q, err := open()
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	if e := p.m[key]; e != nil && time.Since(e.lastUsed) < p.ttl {
		e.lastUsed = time.Now()
		winner := e.q
		p.mu.Unlock()
		q.Close()
		return winner, nil
	}
	p.m[key] = &dbPoolEntry{q: q, lastUsed: time.Now()}
	p.mu.Unlock()
	return q, nil
}

// drop 把**出错的这一条**从池里摘掉（别的请求可能刚换成新的了，比对指针再删）。
func (p *dbDirectPool) drop(key string, q *DirectQuery) {
	p.mu.Lock()
	if e := p.m[key]; e != nil && e.q == q {
		delete(p.m, key)
	}
	p.mu.Unlock()
	if q != nil {
		q.Close()
	}
}

// closeAll 关掉全部：配置热替换（连接要素可能变了）与服务关闭时用。
func (p *dbDirectPool) closeAll() {
	if p == nil {
		return
	}
	p.mu.Lock()
	old := p.m
	p.m = make(map[string]*dbPoolEntry)
	p.mu.Unlock()
	for _, e := range old {
		e.q.Close()
	}
}
