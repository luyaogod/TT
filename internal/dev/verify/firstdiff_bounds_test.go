package verify

// firstDiffDesc 的越界回归：首个差异落在较短串末尾 20 字节内时，老实现会在
// `b[lo:i+20]` 处 panic（slice bounds out of range），把本该返回退出码 3 的
// Gate1 预检变成崩溃。这里直接钉住"任何长度组合都不 panic 且给出偏移"。

import (
	"strings"
	"testing"
)

func TestFirstDiffDescShortAndUnevenLengths(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		off  int
	}{
		{"短串尾部差异", "abcdef", "abXdef", 2},
		{"a比b长且差异在b尾部", "0123456789abcdefghij", "0123456789abcdefgX", 18},
		{"b比a长且差异在a尾部", "0123456789abcdefgX", "0123456789abcdefghij", 18},
		{"首字节差异", "x", "y", 0},
		{"两侧都短于20", "ab", "ac", 1},
		{"完全相同长度不同", "abc", "abcd", 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got string
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("firstDiffDesc panic: %v", r)
					}
				}()
				got = firstDiffDesc([]byte(c.a), []byte(c.b))
			}()
			if !strings.Contains(got, "偏移") && !strings.Contains(got, "长度不同") {
				t.Errorf("描述既没偏移也没长度: %q", got)
			}
			if !strings.Contains(got, "偏移") && c.a[:c.off] != c.b[:c.off] {
				t.Errorf("前 %d 字节应相同，描述该给偏移: %q", c.off, got)
			}
		})
	}
}

// 差异落在尾部时的偏移必须准确（不是被截断成"长度不同"）。
func TestFirstDiffDescTailOffsetExact(t *testing.T) {
	got := firstDiffDesc([]byte("0123456789"), []byte("012345678X"))
	if !strings.Contains(got, "偏移 9") {
		t.Errorf("尾部差异应报偏移 9，实际 %q", got)
	}
}
