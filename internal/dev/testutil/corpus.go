package testutil

// corpus.go —— 真实语料的发现逻辑（`.tzc` 与 `.tzs` 两条管线共用）。
//
// 为什么抽到这里：`.tzc` 的语料回归（internal/dev/cli/corpus_test.go）与 `.tzs` 的
// （internal/dev/tzs/corpus_test.go）要找的是**同一批目录**，只是扩展名不同。抄两份的
// 后果不是「多几行」，而是两边会各自漂移：一边认了 TDEV_CORPUS、另一边只认 TTZS_CORPUS，
// 于是同一条命令在一台机器上跑全量、在另一台上静默跑零个包（"0 个包全部通过"是最坏的一种
// 假绿）。所以这里把「根怎么定」与「怎么走」各写一次，两个环境变量都认。
//
// 本文件只做**发现**：不读包、不解压、不判好坏。语料的判定（sha pin、RoundTrip 定点、
// validate 零新增）住在各自的管线里。

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// corpusEnvVars 是语料根的覆盖变量，按顺序取第一个设置了的。
//
// 两个名字都认是刻意的：TDEV_CORPUS 是 `.tzc` 那条管线先用的名字，TTZS_CORPUS 是
// `.tzs` 这条的。认两个不等于有两份真源 —— 值只有一个，谁先设谁说话。
var corpusEnvVars = []string{"TDEV_CORPUS", "TTZS_CORPUS"}

// DefaultCorpusRoot 是语料的内置缺省（引擎那侧 TZSCLI_WS 的缺省也在它下面）。
const DefaultCorpusRoot = `D:\t100_wrok_dir`

// CorpusRoot 解析真实语料根，解析不出来时返回 ""。
//
// 覆盖变量**设了但指不到一个目录**时返回 ""（不回落到缺省）：那说明调用方想指出一份
// 语料而指错了，静默换一份别的语料去跑，会让「我以为跑的是这份」这个错误结论留下来。
func CorpusRoot() string {
	for _, k := range corpusEnvVars {
		v := strings.TrimSpace(os.Getenv(k))
		if v == "" {
			continue
		}
		if st, err := os.Stat(v); err == nil && st.IsDir() {
			return v
		}
		return ""
	}
	if st, err := os.Stat(DefaultCorpusRoot); err == nil && st.IsDir() {
		return DefaultCorpusRoot
	}
	return ""
}

// CorpusFiles 递归收集 root 下扩展名匹配（大小写不敏感）的语料文件，路径升序。
//
// excludePrefix 里的名字前缀会被跳过 —— 那是驱动脚本自己留下的草稿产物（_ai/_dw/_ed/
// _del/_ac，四个 batch 脚本各自排除集的并集；见 engine/make-manifest.sh）。排除是必须的：
// 那些文件会在两次运行之间出现或消失，把它们算进基线，基线就不可复现 —— 这正是
// engine/corpus.manifest 存在的理由。
func CorpusFiles(root, ext string, excludePrefix ...string) []string {
	var out []string
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(p), ext) {
			return nil
		}
		base := filepath.Base(p)
		for _, pre := range excludePrefix {
			if strings.HasPrefix(base, pre) {
				return nil
			}
		}
		out = append(out, p)
		return nil
	})
	sort.Strings(out)
	return out
}
