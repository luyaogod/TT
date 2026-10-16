package fep

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tt/internal/config"
)

// 长期库（R3）：录制件按 **环境 + ENT** 存放，与缓存目录分开。
//
// 为什么必须搬出缓存：缓存目录（`config.CacheSubdirs` 里的 `fep/`）的语义是
// "一轮一删也不心疼"，启动清理与 `tt cache clean` 都会删它 —— 而录制是"用户报一次 bug，
// 几小时后 AI 拿来重放"的东西，删了就再也造不出来。两者的寿命不一样，就不该住一处。
//
// 落点：`<数据目录>/records/<环境>/<ENT>/<录制 ID>/`。
// **目录树就是索引**：一份录制是一个目录，环境与 ENT 是路径本身。
// 刻意不另写一份 index.json —— 那会有两个真相，而它们迟早对不上（谁的 id 列表、谁的时间戳）。
//
// 录制 ID（= 代理落盘的 tag）不带环境与 ENT，所以按 ID 找人是**遍历三层**（`Find`）。
// 三层的遍历成本可以忽略，换来的是"ID 复制出去就能用"。
const (
	// LibraryDirName 长期库在数据目录下的名字。
	//
	// 刻意**不**加进 `config.CacheSubdirs`：那份清单是"可以删的"，而这里正是不能删的。
	LibraryDirName = "records"
	// EntUnset 是"这一轮没指定企业编号"的落点名字。
	//
	// 没给 --ent 表示沿用登录 shell 自己的 `$TOPENT`（见 OpenConfig.Ent），
	// 那个值我们不知道，所以不能假装知道 —— 如实叫"未指定"。
	EntUnset = "ENT-未指定"
)

// Library 是长期库的根。
type Library struct{ Root string }

// NewLibrary 指向数据目录下的长期库。
func NewLibrary(dataDir string) *Library {
	if dataDir == "" {
		return &Library{}
	}
	return &Library{Root: filepath.Join(dataDir, LibraryDirName)}
}

// EntDir 返回某个环境 + ENT 的落点。
func (l *Library) EntDir(env, ent string) string {
	if l == nil || l.Root == "" {
		return ""
	}
	return filepath.Join(l.Root, envSegment(env), EntSegment(ent))
}

// EntSegment 是 ENT 那一段目录名（空 → `ENT-未指定`）。
func EntSegment(ent string) string {
	if strings.TrimSpace(ent) == "" {
		return EntUnset
	}
	// 同 envSegment：人给的值只允许当"一段"，不许改落点层级。
	return filepath.Base(strings.ReplaceAll(strings.TrimSpace(ent), "\\", "_"))
}

// RecordingDir 返回一份录制的落点。
func (l *Library) RecordingDir(env, ent, id string) string {
	d := l.EntDir(env, ent)
	if d == "" || id == "" {
		return ""
	}
	return filepath.Join(d, filepath.Base(id))
}

// envSegment 把环境名变成一段目录名。
//
// 环境名是人给的（`主机109`、`尚阳通`），直接当目录名有两个坑：分隔符会被当成层级，
// 而 `.`/`..` 会把落点挪到别处。所以只取最后一段并剔掉路径分隔符 —— 名字本身照留（中文没问题）。
func envSegment(env string) string {
	if env == "" {
		return "环境-未指定"
	}
	s := strings.TrimSpace(env)
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.Trim(s, ".")
	if s == "" {
		return "环境-未指定"
	}
	return s
}

// LibraryItem 是库里的一份录制。
type LibraryItem struct {
	ID   string `json:"id"`
	Env  string `json:"env"`
	Ent  string `json:"ent"`
	Dir  string `json:"dir"`
	Meta *RecordingMeta
	// Files 是这份录制的 AUI 侧文件（按时间从早到晚），直接可以喂 SnapshotOfFiles。
	Files []string
}

// Promote 把缓存目录里的一份录制搬进长期库（**移动**，不是拷贝）。
//
// 只搬这一份的东西：`<ID>-<连接>-{rt2gdc,gdc2rt,timeline}.log`、它们的帧索引
// `<同名>.frames.jsonl` 与 `<ID>.meta.json`。
// 先全部搬完再收尾 —— 中途失败时新的那份是不全的，所以要报错让调用方知道，
// 且**不删源**（宁可两边都有，也不要一份半的）。
func (l *Library) Promote(srcDir, id string) (LibraryItem, error) {
	var item LibraryItem
	if l == nil || l.Root == "" {
		return item, errors.New("fep: 没有数据目录，长期库落点算不出来")
	}
	if id == "" {
		return item, errors.New("fep: 搬哪一份？没给录制 ID")
	}
	mt, err := ReadMeta(srcDir, id)
	if err != nil {
		return item, fmt.Errorf("fep: 这份录制没有身份文件，不知道该进哪个环境/ENT 的库（%w）", err)
	}
	if strings.TrimSpace(mt.Env) == "" {
		return item, errors.New("fep: 录制身份里没有环境，长期库按环境+ENT 分层，不知道该放哪一层")
	}
	files, err := recordingLogFiles(srcDir, id)
	if err != nil {
		return item, err
	}
	if len(files) == 0 {
		return item, fmt.Errorf("fep: %s 里没有 %s 的录制件", srcDir, id)
	}
	dst := l.RecordingDir(mt.Env, mt.Ent, id)
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return item, err
	}
	var moved []string
	for _, src := range files {
		dstFile := filepath.Join(dst, filepath.Base(src))
		if err := moveFile(src, dstFile); err != nil {
			return item, fmt.Errorf("fep: 搬 %s 失败: %w（已搬 %d 份，源未删）", filepath.Base(src), err, len(moved))
		}
		moved = append(moved, dstFile)
	}
	// 身份文件最后搬：它是"这一份齐了"的标记，先搬它会让半份录制看起来是完整的。
	metaSrc, metaDst := MetaPath(srcDir, id), MetaPath(dst, id)
	if err := moveFile(metaSrc, metaDst); err != nil {
		return item, fmt.Errorf("fep: 搬身份文件失败: %w（录制件已搬入 %s）", err, dst)
	}
	item = LibraryItem{ID: id, Env: mt.Env, Ent: EntSegment(mt.Ent), Dir: dst, Meta: mt}
	item.Files, _ = RecordingFilesFor(dst, id)
	return item, nil
}

// recordingLogFiles 列出这一份录制的**流量文件**（不含身份文件：身份是"这一份齐了"
// 的标记，由 Promote 在最后单独搬，见那里的注释）。三个后缀都带：回放要事件侧，
// 排障要看 timeline；帧索引也带（它记着帧级时间，丢了明细页就只剩块级时间）。
func recordingLogFiles(dir, id string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		// 按切出来的 ID **精确**匹配：不然 "fep-1" 会把 "fep-10" 的文件一起搬走。
		if got, ok := RecordingIDOfAny(e.Name()); ok && got == id {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(out)
	return out, nil
}

// moveFile 先试改名（同一卷上是原子的），跨卷时退回"拷贝 + 删源"。
//
// Windows 上改名到一个**已存在**的目标会失败，所以先删目标：重复入库同一份时，
// 库里那份多半是上一次的同样内容，覆盖是对的（真出问题也是可重建的）。
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if _, err := os.Stat(dst); err == nil {
		if err := os.Remove(dst); err != nil {
			return err
		}
		if err := os.Rename(src, dst); err == nil {
			return nil
		}
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Remove(src)
}

// Remove 删掉库里的一份录制（整份：目录 + 里面所有文件）。返回删掉的文件数与字节数。
//
// 这是"前端手动清除"落地的那一下。刻意只认**库里的目录**（`<环境>/<ENT>/<ID>/`）：
// 传一个别的路径进来不该把缓存目录或别的什么东西删掉。
func (l *Library) Remove(item LibraryItem) (files int, bytes int64, err error) {
	if l == nil || l.Root == "" {
		return 0, 0, errors.New("fep: 没有数据目录，长期库落点算不出来")
	}
	if item.Dir == "" {
		return 0, 0, errors.New("fep: 这份录制没有落点，不猜着删")
	}
	rel, rerr := filepath.Rel(l.Root, item.Dir)
	if rerr != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return 0, 0, fmt.Errorf("fep: %s 不在长期库里（%s），不删", item.Dir, l.Root)
	}
	return config.RemoveTree(item.Dir)
}

// RenameEnv 把某个环境的**整个库目录**改到新名字下（改名 = 搬家，不是删掉）。
//
// 为什么要有这一条：设置页保存一次 hosts，"环境改了名"与"环境被删了"在名字集合上
// 长得一模一样，而级联清理只认名字 —— 于是改个名字就把那个环境名下长期库里的录像全删了
// （2026-10-15 实测：清空环境名再保存，27 份录像、2.5 MB，`os.RemoveAll` 不可恢复）。
// 名字是**标签**，录像才是不可再生的东西：改名要把录像一起搬过去。
//
// 顺手把每一份的**身份文件里的 Env 也改成新名字**：`rec reproduce` 是拿身份里的环境去起
// 会话的，留旧名字的话改名之后那些录像就再也回放不起来了（"环境不存在"）。
// 返回搬过去的**录制份数**。
func (l *Library) RenameEnv(from, to string) (int, error) {
	if l == nil || l.Root == "" {
		return 0, errors.New("fep: 没有数据目录，长期库落点算不出来")
	}
	if strings.TrimSpace(from) == "" || strings.TrimSpace(to) == "" {
		return 0, errors.New("fep: 改名前后的环境名都要给全，不猜着搬")
	}
	src := filepath.Join(l.Root, EnvSegment(from))
	dst := filepath.Join(l.Root, EnvSegment(to))
	if src == dst {
		return 0, nil
	}
	if _, serr := os.Stat(src); serr != nil {
		return 0, nil // 这个环境本来就没录过：不是错
	}
	if _, derr := os.Stat(dst); derr != nil {
		// 目标还不存在：整棵搬过去（同一卷上是原子的）
		if err := os.Rename(src, dst); err != nil {
			return 0, fmt.Errorf("fep: 把 %s 的录制库搬到 %s 失败: %w", from, to, err)
		}
	} else {
		// 目标已经有了（新旧两个环境都在录过）：把每个 ENT 目录并进去。
		// 同名 ENT 目录里的录制 ID 不会撞（ID 带时间戳），所以并进去是安全的。
		ents, rerr := os.ReadDir(src)
		if rerr != nil {
			return 0, rerr
		}
		for _, e := range ents {
			s, d := filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())
			if err := os.MkdirAll(d, 0o755); err != nil {
				return 0, err
			}
			subs, serr := os.ReadDir(s)
			if serr != nil {
				return 0, serr
			}
			for _, sub := range subs {
				if err := os.Rename(filepath.Join(s, sub.Name()), filepath.Join(d, sub.Name())); err != nil {
					return 0, fmt.Errorf("fep: 并 %s 到 %s 失败: %w", sub.Name(), d, err)
				}
			}
			_ = os.Remove(s)
		}
		_ = os.Remove(src)
	}
	// 身份文件里的环境名要跟着改（见函数头：不改的话 `rec reproduce` 会找不到这个环境）。
	n, terr := l.rewriteEnvInMetas(dst, to)
	if terr != nil {
		return n, terr
	}
	return n, nil
}

// rewriteEnvInMetas 把一棵库目录下每份录制身份里的 Env 改成 to，返回改了几份。
//
// 目录形态照 List 的口径认两种：`<环境>/<ENT>/<录制 ID>/`（Promote 落的）与
// "直接摊在 ENT 目录下"（手工拷进来的）—— 两边都遍历一遍，别只认一种。
func (l *Library) rewriteEnvInMetas(envDir, to string) (int, error) {
	ents, err := os.ReadDir(envDir)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		entDir := filepath.Join(envDir, e.Name())
		dirs := []string{entDir}
		if subs, serr := os.ReadDir(entDir); serr == nil {
			for _, s := range subs {
				if s.IsDir() {
					dirs = append(dirs, filepath.Join(entDir, s.Name()))
				}
			}
		}
		for _, dir := range dirs {
			ids, lerr := ListIDs(dir)
			if lerr != nil {
				continue
			}
			for _, id := range ids {
				mt, merr := ReadMeta(dir, id)
				if merr != nil || mt.Env == to {
					continue
				}
				mt.Env = to
				if _, werr := WriteMeta(dir, *mt); werr != nil {
					return n, werr
				}
				n++
			}
		}
	}
	return n, nil
}

// RemoveEnv 删掉某个环境的**整个库目录**（级联清理）。
//
// 什么时候用它：环境配置被**删掉**（设置页移除那个环境）时，它名下的录制一起收掉 ——
// 那些录制件的画面与业务数据都属于那个环境，环境没了它们就是一堆没人认领的字节。
//
// 只认"真的删了"：改名走 RenameEnv（见那里的说明），调用方要用 web.envDiff 把两者分开。
func (l *Library) RemoveEnv(env string) (files int, bytes int64, err error) {
	if l == nil || l.Root == "" {
		return 0, 0, errors.New("fep: 没有数据目录，长期库落点算不出来")
	}
	if strings.TrimSpace(env) == "" {
		return 0, 0, errors.New("fep: 没给环境名，不猜着删")
	}
	dir := filepath.Join(l.Root, EnvSegment(env))
	if _, serr := os.Stat(dir); serr != nil {
		return 0, 0, nil // 这个环境本来就没录过：不是错
	}
	return config.RemoveTree(dir)
}

// Envs 列出库里有哪些环境（目录名，已按 EnvSegment 归一）。
//
// 用途是"环境配置没了、库里还留着"的兜底清理（`tt fep rec prune`）：
// 拿它与配置里的环境列表对一遍，多出来的就是没人认领的那些。
func (l *Library) Envs() ([]string, error) {
	if l == nil || l.Root == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(l.Root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

// EnvSegment 是环境名在库里的那一段目录名（`tt fep rec prune` 要用它把配置里的
// 环境名与库里的目录名对齐 —— 两边必须用同一个函数，否则永远对不上）。
func EnvSegment(env string) string { return envSegment(env) }

// List 按 环境 → ENT → 录制 ID 列出库里的全部录制。
func (l *Library) List() ([]LibraryItem, error) {
	if l == nil || l.Root == "" {
		return nil, nil
	}
	envs, err := os.ReadDir(l.Root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []LibraryItem
	for _, env := range envs {
		if !env.IsDir() {
			continue
		}
		entRoot := filepath.Join(l.Root, env.Name())
		ents, err := os.ReadDir(entRoot)
		if err != nil {
			continue
		}
		for _, ent := range ents {
			if !ent.IsDir() {
				continue
			}
			entDir := filepath.Join(entRoot, ent.Name())
			// 一份录制一个目录（Promote 就是这么落的）。同时也认"直接摊在 ENT 目录下"
			// 的那种形态 —— 手工拷进来的一份不该看不见。
			dirs := []string{entDir}
			if subs, serr := os.ReadDir(entDir); serr == nil {
				for _, s := range subs {
					if s.IsDir() {
						dirs = append(dirs, filepath.Join(entDir, s.Name()))
					}
				}
			}
			for _, dir := range dirs {
				ids, err := ListIDs(dir)
				if err != nil {
					continue
				}
				for _, id := range ids {
					item := LibraryItem{ID: id, Env: env.Name(), Ent: ent.Name(), Dir: dir}
					if mt, merr := ReadMeta(dir, id); merr == nil {
						item.Meta = mt
					}
					item.Files, _ = RecordingFilesFor(dir, id)
					out = append(out, item)
				}
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Env != out[j].Env {
			return out[i].Env < out[j].Env
		}
		if out[i].Ent != out[j].Ent {
			return out[i].Ent < out[j].Ent
		}
		return out[i].ID > out[j].ID // 同一层里新的在前（ID 里带时间戳）
	})
	return out, nil
}

// Find 按录制 ID 在库里找一份（ID 不带环境/ENT，所以是遍历）。
func (l *Library) Find(id string) (LibraryItem, bool, error) {
	items, err := l.List()
	if err != nil {
		return LibraryItem{}, false, err
	}
	for _, it := range items {
		if it.ID == id {
			return it, true, nil
		}
	}
	return LibraryItem{}, false, nil
}
