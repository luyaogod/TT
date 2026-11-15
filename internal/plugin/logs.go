package plugin

// 插件日志读取（P6-07）。
//
// supervisor 把服务 stdout/stderr 落到 logs/<id>/service-<时间戳>.log；一个插件长期运行
// 会攒很多个文件，所以"看日志"的默认语义是**最新的那个文件**的最后 N 行，而不是把整个
// 目录拼起来 —— 后者会把一次 10 分钟前的崩溃混进刚才那行错误里。
//
// 实现在 plugin 包而不是命令层：CLI 的 `tt plugin logs` 与 Web 的
// `GET /api/plugins/{id}/logs` 必须读同一个文件、用同一套"最新"判据。

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DefaultLogTail `tt plugin logs` 的默认行数。
const DefaultLogTail = 50

// maxLogTailBytes 一次最多为"取尾部"读多少字节（防止一行有几十 MB 时把内存吃光）。
const maxLogTailBytes = 8 << 20

// NewestLogFile 返回某插件日志目录里最新的日志文件；没有日志时返回空串。
//
// "最新"先看修改时间，再看文件名（同一秒内写的两份按名字排，结果稳定可复现）。
func (s *Store) NewestLogFile(id string) (string, error) {
	dir := s.LogsDir(id)
	if dir == "" {
		return "", errors.New("数据目录未确定，无法定位插件日志目录")
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("读取插件日志目录失败: %w", err)
	}
	type cand struct {
		path string
		mod  int64
		name string
	}
	var list []cand
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, ierr := e.Info()
		if ierr != nil {
			continue
		}
		list = append(list, cand{path: filepath.Join(dir, e.Name()), mod: info.ModTime().UnixNano(), name: e.Name()})
	}
	if len(list) == 0 {
		return "", nil
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].mod != list[j].mod {
			return list[i].mod > list[j].mod
		}
		return list[i].name > list[j].name
	})
	return list[0].path, nil
}

// TailLog 读某插件最新日志的最后 n 行。
//
// 返回的 file 是实际读的文件（没有日志时为空串）。没有日志不是错误：插件没起来过就
// 不会有日志，那是一个事实而不是一次失败。
func (s *Store) TailLog(id string, n int) (file string, lines []string, err error) {
	if n <= 0 {
		n = DefaultLogTail
	}
	file, err = s.NewestLogFile(id)
	if err != nil || file == "" {
		return file, nil, err
	}
	lines, err = ReadTailFile(file, n)
	return file, lines, err
}

// ReadTailFile 读一个文件最后 n 行（从文件尾往前按块读，不整文件载入）。
func ReadTailFile(path string, n int) ([]string, error) {
	if n <= 0 {
		n = DefaultLogTail
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := st.Size()
	if size == 0 {
		return nil, nil
	}

	const chunk = 64 << 10
	var tail []byte
	pos := size
	newlines := 0
	for pos > 0 && newlines <= n && len(tail) < maxLogTailBytes {
		start := pos - chunk
		if start < 0 {
			start = 0
		}
		b := make([]byte, pos-start)
		if _, rerr := f.ReadAt(b, start); rerr != nil && !errors.Is(rerr, io.EOF) {
			return nil, fmt.Errorf("读日志文件失败: %w", rerr)
		}
		tail = append(b, tail...)
		newlines = bytes.Count(tail, []byte{'\n'})
		pos = start
	}
	text := strings.TrimRight(string(tail), "\r\n")
	if text == "" {
		return nil, nil
	}
	all := strings.Split(text, "\n")
	if len(all) > n {
		all = all[len(all)-n:]
	}
	for i := range all {
		all[i] = strings.TrimSuffix(all[i], "\r")
	}
	return all, nil
}
