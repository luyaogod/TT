package fep

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// 录制件的读取。两种形态都是真实产物，只认一种会得到一个安静的错答案：
//
//	原始帧（Go 代理写的）：带长度头的字节流，重解码；
//	文本（早期原型写的）：一条消息一行，而且行首可能还留着帧头噪声。
//
// 会话里的 AUI 状态**全部**来自录制件 —— 运行期没有"查询当前界面"的请求，
// 只有它主动推来的 om 消息。所以谁想知道界面是什么样，就得读这里。

// TextRecording reports whether a recording is the text form: a frame header is a
// big-endian length, so its first byte is almost never a lowercase letter, while
// the verbs (meta/om/event) always are.
func TextRecording(b []byte) bool {
	return len(b) > 0 && b[0] >= 'a' && b[0] <= 'z'
}

// ReadRecording reads one recording into a list of message bodies.
func ReadRecording(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if TextRecording(b) {
		var out []string
		for _, line := range strings.Split(string(b), "\n") {
			if msg, ok := TrimNoise(line); ok {
				out = append(out, msg)
			}
		}
		return out, nil
	}
	d := NewDecoder()
	frames, err := d.Feed(b)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	var out []string
	for _, fr := range frames {
		if fr.Type == FrameTypeAUI {
			out = append(out, string(fr.Payload))
		}
	}
	return out, nil
}

// RecordingFiles lists recordings with the given suffix, newest first.
func RecordingFiles(dir, suffix string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	type cand struct {
		path string
		mod  time.Time
	}
	var cands []cand
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), suffix) {
			continue
		}
		// os.Stat 而不是 DirEntry.Info()：后者在 Windows 上可能是陈旧值（见 do_test.go 的实测）
		info, ierr := os.Stat(filepath.Join(dir, e.Name()))
		if ierr != nil {
			continue
		}
		cands = append(cands, cand{filepath.Join(dir, e.Name()), info.ModTime()})
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].mod.After(cands[j].mod) })
	out := make([]string, 0, len(cands))
	for _, c := range cands {
		out = append(out, c.path)
	}
	return out, nil
}

// LoadModelFrom rebuilds the AUI of a session from its recordings: every frame of
// every `<tag>*-rt2gdc.log`, oldest first (a screen is described by a stream of
// updates, so order matters).
func LoadModelFrom(recordDir, tag string) (*Model, *Resolver, int, error) {
	files, err := RecordingFiles(recordDir, "-rt2gdc.log")
	if err != nil {
		return nil, nil, 0, err
	}
	// 只取本次会话的（tag 前缀相同），并且按时间从早到晚喂
	var mine []string
	for _, f := range files {
		if tag == "" || strings.HasPrefix(filepath.Base(f), tag) {
			mine = append(mine, f)
		}
	}
	for i, j := 0, len(mine)-1; i < j; i, j = i+1, j-1 {
		mine[i], mine[j] = mine[j], mine[i]
	}
	m := NewModel()
	msgs := 0
	for _, f := range mine {
		bodies, err := ReadRecording(f)
		if err != nil {
			return nil, nil, 0, err
		}
		for _, body := range bodies {
			m.ApplyMessage(body)
			msgs++
		}
	}
	if msgs == 0 {
		return nil, nil, 0, fmt.Errorf("fep: %s 下没有本会话（%s*）的 AUI 录制件", recordDir, tag)
	}
	return m, NewResolver(m), msgs, nil
}
