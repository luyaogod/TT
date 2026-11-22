package taskws

// 假 devdata.FileSource：在内存里模拟 $TOP 下的服务器文件树，用来离线跑通
// Create / Update 的"联网半段"。
//
// 为什么必须有它：Create 的 fetch→Map→Pack→Export 与 Update 的三条刷新分支，
// 之前唯一不可测的原因就是"材料只能从 SFTP 来"。生产侧在 CreateRequest/UpdateRequest
// 上开了 Files 注入缝之后，整条链路可以在断网环境里端到端跑 —— 这份假来源就是那个
// "$TOP"：路径一律相对 $TOP（如 erp/axm/dzx/tap/<程序>.tap），Stat/ReadFile 的语义
// 与真实现对齐（不存在统一成 fs.ErrNotExist、超上限报错），并记录被访问过的路径，
// 用来断言"请求里让拉什么，真的去拉了什么"（Form / 模块 / 身份）。
//
// 形态照 internal/devdata/fake_test.go 的 fakeFiles，但它是 devdata 包内私有类型，
// 这里只实现接口需要的五个方法（Top/Stat/ReadFile/Walk/Close）。

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/luyaogod/TT/internal/devdata"
)

// fakeBaseTime 是假服务器上文件的默认 mtime（固定值 → 测试可重复）。
var fakeBaseTime = time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)

// fakeSource 是内存版 FileSource。
type fakeSource struct {
	top    string
	files  map[string][]byte
	mtimes map[string]time.Time
	dirs   map[string]bool
	failOn map[string]error // rel → ReadFile 时返回该错误（一次）
	stats  []string         // 被 Stat 过的相对路径（按调用顺序）
	reads  []string         // 被 ReadFile 过的相对路径（按调用顺序）
	walked []string         // 被 Walk 过的相对路径
	closed bool
}

// newFakeSource 造一个空 $TOP。
func newFakeSource(top string) *fakeSource {
	return &fakeSource{
		top: top, files: map[string][]byte{}, mtimes: map[string]time.Time{},
		dirs: map[string]bool{}, failOn: map[string]error{},
	}
}

// Top 返回 $TOP（写进材料 manifest.source.topDir，可事后复核材料来自哪台机）。
func (f *fakeSource) Top() string { return f.top }

// norm 把相对路径规整成 fake 的键（与 devdata 的 sftpSource 同样处理前导 /）。
func norm(rel string) string { return strings.TrimPrefix(path.Clean("/"+rel), "/") }

// put 在假服务器上放一个文本文件。
func (f *fakeSource) put(rel, body string, mt time.Time) {
	rel = norm(rel)
	f.files[rel] = []byte(body)
	f.mtimes[rel] = mt
}

// putBytes 放一段二进制（.42f 之类）。
func (f *fakeSource) putBytes(rel string, body []byte, mt time.Time) {
	rel = norm(rel)
	f.files[rel] = append([]byte(nil), body...)
	f.mtimes[rel] = mt
}

// mkdir 在假服务器上放一个目录（模块探测会 Stat 目录）。
func (f *fakeSource) mkdir(rel string) { f.dirs[norm(rel)] = true }

// remove 把文件从假服务器上删掉（造"服务器上已没有"）。
func (f *fakeSource) remove(rel string) { delete(f.files, norm(rel)) }

// failRead 让下一次读该路径失败（造 Fetch 中途失败）。
func (f *fakeSource) failRead(rel string, err error) { f.failOn[norm(rel)] = err }

// statLog / readLog / walkLog 返回被访问过的路径副本。
func (f *fakeSource) statLog() []string { return append([]string(nil), f.stats...) }
func (f *fakeSource) readLog() []string { return append([]string(nil), f.reads...) }
func (f *fakeSource) walkLog() []string { return append([]string(nil), f.walked...) }

// resetLog 清空访问记录（多次请求之间断言"这次拉了哪些"）。
func (f *fakeSource) resetLog() { f.stats, f.reads, f.walked = nil, nil, nil }

// Stat 看一个相对路径；不存在时返回 fs.ErrNotExist（与 sftpSource 同契约）。
func (f *fakeSource) Stat(_ context.Context, rel string) (devdata.RemoteFile, error) {
	rel = norm(rel)
	f.stats = append(f.stats, rel)
	if f.dirs[rel] {
		return devdata.RemoteFile{Rel: rel, IsDir: true}, nil
	}
	b, ok := f.files[rel]
	if !ok {
		return devdata.RemoteFile{Rel: rel}, fmt.Errorf("%w: %s", fs.ErrNotExist, rel)
	}
	mt := f.mtimes[rel]
	if mt.IsZero() {
		mt = fakeBaseTime
	}
	return devdata.RemoteFile{Rel: rel, Size: int64(len(b)), ModTime: mt}, nil
}

// ReadFile 读一个文件；max>0 且超限时返回错误（与 host.ErrFileTooLarge 语义一致）。
func (f *fakeSource) ReadFile(_ context.Context, rel string, max int64) ([]byte, error) {
	rel = norm(rel)
	f.reads = append(f.reads, rel)
	if err := f.failOn[rel]; err != nil {
		delete(f.failOn, rel) // 只失败一次，避免后续步骤也被它绊住
		return nil, err
	}
	b, ok := f.files[rel]
	if !ok {
		return nil, fmt.Errorf("%w: %s", fs.ErrNotExist, rel)
	}
	if max > 0 && int64(len(b)) > max {
		return nil, fmt.Errorf("文件 %s 超过上限 %d", rel, max)
	}
	return append([]byte(nil), b...), nil
}

// Walk 递归列一棵子树里的文件（本包的 Create/Update 不带 WithDesign，基本用不到；
// 实现出来是为了接口完整 + 未来 design 归档的用例）。
func (f *fakeSource) Walk(ctx context.Context, rel string, fn func(devdata.RemoteFile) error) error {
	rel = norm(rel)
	f.walked = append(f.walked, rel)
	prefix := rel + "/"
	keys := make([]string, 0, len(f.files))
	for k := range f.files {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return fmt.Errorf("%w: %s", fs.ErrNotExist, rel)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := fn(devdata.RemoteFile{Rel: k, Size: int64(len(f.files[k])), ModTime: f.mtimes[k]}); err != nil {
			return err
		}
	}
	return nil
}

// Close 记录关闭（注入来源由调用方持有，devdata 不会替我们关）。
func (f *fakeSource) Close() { f.closed = true }

// seedProgram 往假 $TOP 放一套"设计器接管过的程序"材料（标准模块 axm）。
//
// entries 是 testutil 的包条目（<prog>.tap/.tgl/.4gl）；缺哪个就真的不在服务器上
// —— 用来造"标准程序没有 .tap"这种 Create 必须拒绝的输入。tab 单独合成
// （testutil.NormalEntries 不含 .tab）。
func seedProgram(t *testing.T, src *fakeSource, prog string, entries map[string]string) {
	t.Helper()
	mt := fakeBaseTime
	src.put("erp/axm/dzx/tab/"+prog+".tab", fixTab(prog), mt)
	if body, ok := entries[prog+".tap"]; ok && body != "" {
		src.put("erp/axm/dzx/tap/"+prog+".tap", body, mt)
	}
	if body, ok := entries[prog+".tgl"]; ok && body != "" {
		src.put("erp/axm/dzx/tgl/"+prog+".tgl", body, mt)
	}
	if body, ok := entries[prog+".4gl"]; ok && body != "" {
		src.put("erp/axm/4gl/"+prog+".4gl", body, mt)
	}
}

// seedForm 放一个 .42f（Form 选项的透传断言用）。
func seedForm(t *testing.T, src *fakeSource, prog string) {
	t.Helper()
	src.putBytes("erp/axm/42f/"+prog+".42f", []byte("\x00\x01fake-42f"), fakeBaseTime)
}
