package split

// renameBlock 的坐标防御回归：Region 可能来自老 regions.json 或手工构造，
// 签名行/正文区间越界时必须返回退出码 3 的 VerifyError，不能在切片处 panic。
// 这条不依赖 Split 主链（主链里坐标由 fence.Parse 重算，越界不可达），
// 钉的是"写回路径永远不崩"这条底线。

import (
	"errors"
	"testing"

	"github.com/luyaogod/TT/internal/dev/fence"
	"github.com/luyaogod/TT/internal/dev/model"
)

func TestRenameBlockOutOfBoundsRegion(t *testing.T) {
	doc := &model.Document{Text: []byte("short")} // len=5

	cases := []struct {
		name string
		sig  *model.ByteRange
		body model.ByteRange
	}{
		{"签名行越界", &model.ByteRange{Start: 0, End: 10}, model.ByteRange{Start: 0, End: 5}},
		{"签名行起点越界", &model.ByteRange{Start: 9, End: 9}, model.ByteRange{Start: 0, End: 5}},
		{"签名行反向", &model.ByteRange{Start: 3, End: 1}, model.ByteRange{Start: 0, End: 5}},
		{"正文越界", &model.ByteRange{Start: 0, End: 1}, model.ByteRange{Start: 0, End: 99}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			edited := &model.Region{
				Name:          "function.f",
				SignatureSpan: c.sig,
				ContentSpan:   c.body,
			}
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("renameBlock panic: %v", r)
					}
				}()
				_, err = renameBlock(doc, edited, edited, "g", fence.Pair{})
			}()
			var ve *VerifyError
			if !errors.As(err, &ve) {
				t.Errorf("越界应返回 VerifyError(3)，实际 %v", err)
			}
		})
	}
}
