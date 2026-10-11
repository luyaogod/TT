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

// TextRecording reports whether a recording *looks* like the text form. Only a hint:
// the reliable decision is made by trying the frame decoder first (see ReadRecording),
// because a real recording opens with an unframed handshake that starts with a letter.
func TextRecording(b []byte) bool {
	return len(b) > 0 && b[0] >= 'a' && b[0] <= 'z'
}

// ReadRecording reads one recording into a list of message bodies.
func ReadRecording(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// 先按**帧**试：真实原始录制件的第一条消息往往是握手 `meta Connection …`，
	// 它是整个会话里唯一**不带帧头**的一条 —— 拿“首字节是不是字母”判定形态会被它骗过去，
	// 把二进制帧流当文本读，结果是一个安静的错答案（整屏解析成垃圾、字段值一个都不见）。
	if msgs := decodeFrames(b); len(msgs) > 0 {
		return msgs, nil
	}
	// 退回文本形态（早期原型的落盘：一条消息一行，行首可能带帧头噪声）
	var out []string
	for _, line := range strings.Split(string(b), "\n") {
		if msg, ok := TrimNoise(line); ok {
			out = append(out, msg)
		}
	}
	return out, nil
}

// decodeFrames 按帧解码，并只保留“看起来像一条消息”的载荷。
func decodeFrames(b []byte) []string {
	d := NewDecoder()
	frames, err := d.Feed(b)
	if err != nil {
		return nil
	}
	var out []string
	for _, fr := range frames {
		if fr.Type != FrameTypeAUI {
			continue
		}
		msg := strings.TrimSpace(string(fr.Payload))
		if msg == "" {
			continue
		}
		if _, ok := TrimNoise(msg); !ok && !strings.HasPrefix(msg, "{") {
			continue // 既不是动词开头也不是对象组：不是消息
		}
		out = append(out, msg)
	}
	return out
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

// ReadModelFrom rebuilds the AUI model from "a file or a directory".
//
//   - 单文件：自动配**同一连接**的另一侧（*-gdc2rt.log ↔ *-rt2gdc.log）——
//     拿事件文件去配别的连接的 AUI，对象编号对不上，动作会全变"名字未解析"；
//   - 目录：只取最新一份（一个目录里往往躺着好几轮录制，混着读会把不同会话的对象
//     堆进同一个模型，实测动作数将近翻倍、名字全对不上）。
func ReadModelFrom(from string) (*Model, *Resolver, int, error) {
	if from != "" {
		if st, err := os.Stat(from); err == nil && !st.IsDir() {
			if strings.HasSuffix(from, "-gdc2rt.log") {
				paired := strings.TrimSuffix(from, "-gdc2rt.log") + "-rt2gdc.log"
				if _, perr := os.Stat(paired); perr == nil {
					from = paired
				}
			}
			return loadModelFiles([]string{from})
		}
	}
	files, err := RecordingFiles(from, "-rt2gdc.log")
	if err != nil {
		return nil, nil, 0, err
	}
	if from == "" {
		return nil, nil, 0, fmt.Errorf("fep: 没给录制件")
	}
	if len(files) == 0 {
		return nil, nil, 0, fmt.Errorf("fep: %s 下没有 *-rt2gdc.log", from)
	}
	return loadModelFiles(files[:1])
}

func loadModelFiles(files []string) (*Model, *Resolver, int, error) {
	m := NewModel()
	msgs := 0
	for _, f := range files {
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
		return nil, nil, 0, fmt.Errorf("fep: 录制件里没有可用消息")
	}
	return m, NewResolver(m), msgs, nil
}
