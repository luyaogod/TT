package web

// 录制解析结果的**读侧缓存**（契约 W5）。
//
// 为什么要有：明细页与实时帧流都在一秒一次地读同一份录制。每次请求从字节重解一遍
// （实测 448KB ≈ 10ms、更大的录制线性增长）在"看着它长"的时候纯属浪费。
//
// 判据只有一条：**每个方向文件的大小与 mtime 都没变**。变了就重算 —— 活录制追加字节时
// 大小必变，所以"正在录"照样能看到新帧；换成同名新文件（重建索引、重录一轮）mtime 也会变。
//
// 边界与取舍：
//   - 命中时返回**首层副本**：调用方可以随手改自己那份的字段，但改不坏缓存里的模型；
//   - 条目数封顶（见 recCacheMax）：这东西是加速器，不是存储 —— 超了直接清空重来，
//     不用 LRU 那套（录音明细的访问天然集中在一两份上）；
//   - 解析失败不入缓存（没有"负缓存"）：文件正在写、下一毫秒就好了的场景不该被记住。

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"tt/internal/fep"
)

const recCacheMax = 8

// recParseCount 只用来让测试断言"没有重复解析"（生产代码不读它）。
var recParseCount atomic.Int64

// recCache 是"某个目录下某个 id"的解析结果。
type recCache struct {
	mu sync.Mutex
	m  map[string]recCacheEntry
}

type recCacheEntry struct {
	stamp string
	read  *recRead
}

// recStamp 把一组 raw 文件的身份压成一句话：文件名 + 大小 + mtime。
//
// 用 stat 而不是读文件：这一步本身就是"要不要重解"的判据，读文件就白省了。
func recStamp(files []fep.RawInfo) string {
	var b strings.Builder
	for _, f := range files {
		var size, mt int64 = -1, -1
		if st, err := os.Stat(f.Path); err == nil {
			size, mt = st.Size(), st.ModTime().UnixNano()
		}
		fmt.Fprintf(&b, "%s|%d|%d;", f.Name, size, mt)
	}
	return b.String()
}

// recCacheGet 命中时返回**首层副本**（调用方改不到缓存里的那份）。
func (s *Server) recCacheGet(key, stamp string) (*recRead, bool) {
	s.rec.mu.Lock()
	defer s.rec.mu.Unlock()
	e, ok := s.rec.m[key]
	if !ok || e.stamp != stamp || e.read == nil {
		return nil, false
	}
	cp := *e.read
	return &cp, true
}

// recCachePut 记下这次解析。失败的结果不进（没有负缓存）。
func (s *Server) recCachePut(key, stamp string, read *recRead) {
	if read == nil {
		return
	}
	recParseCount.Add(1)
	s.rec.mu.Lock()
	defer s.rec.mu.Unlock()
	if s.rec.m == nil || len(s.rec.m) >= recCacheMax {
		// 满了就清空重来：录音明细的访问集中在一两份，LRU 不值得。
		s.rec.m = map[string]recCacheEntry{}
	}
	s.rec.m[key] = recCacheEntry{stamp: stamp, read: read}
}
