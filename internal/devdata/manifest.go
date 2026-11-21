package devdata

// manifest：一份材料的"身份证"（plan §5.2）。规则只有四条，都是为了让 model 能判断
// "这份材料还能不能用"：
//
//  1. **每个落盘文件都有 sha256**，外加源 mtime —— 前者判"内容被改过"，后者判"源变新了"。
//  2. 缺的东西进 `missing[]`，**不判失败**。有的程序本来就不产 .tap/_rep.4gl；
//     把可选缺失当失败，会让 fetch 在一半程序上不可用。
//  3. `versions` 记"规则是否仍适用"的锚点：模板版本（`code_a05.template` 的 `(Version:3)`）、
//     `.tgl` 头部的 SD/PR 版次、`DZAA_T` 的版次。
//  4. `warnings[]` 只承载**一致性提示**（蓝图比骨架新 → 产物可能滞后），不阻断。
//
// M 侧（共用资料）用同一个 schema；差别只有 `prog` 为空、`tables[]` 记逐表行数。
// `tables[]` 与 `files[]` 是**互补**而不是替换：`files[]` 只登记真正落盘的文件
// （规格库一个 `role=dict`、模板树每片一个 `role=mdl`），`tables[]` 登记库里的表。
// 不让每张表在 `files[]` 里各占一条（并各给一个 sha256）是刻意的：表没有独立文件，
// 硬编一个 sha 只会变成"看起来校验过、其实校验的是整库"的假证据。

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ManifestSchema 契约版本。改结构就升这个串（model 会按它分支）。
const ManifestSchema = "tt.devdata/v1"

// ManifestFileName 每份材料的清单文件名（env 根、prog 目录各一份）。
const ManifestFileName = "manifest.json"

// Manifest 一份材料的清单。
type Manifest struct {
	Schema    string       `json:"schema"`
	Kind      string       `json:"kind"` // mirror | fetch
	Env       string       `json:"env"`
	Module    string       `json:"module,omitempty"`
	Prog      string       `json:"prog,omitempty"`
	FetchedAt string       `json:"fetchedAt"`
	TTVersion string       `json:"ttVersion,omitempty"`
	Source    Source       `json:"source"`
	Versions  Versions     `json:"versions"`
	Files     []FileEntry  `json:"files"`
	Tables    []TableEntry `json:"tables,omitempty"`
	Missing   []string     `json:"missing"`
	Warnings  []string     `json:"warnings"`
}

// Source 材料来自哪里（复现一份材料时必须知道这两项）。
type Source struct {
	TopDir string `json:"topDir,omitempty"`
	Zone   string `json:"zone,omitempty"`
	Ent    int    `json:"ent,omitempty"`
	// DictDB 规格库文件（M 侧）；D 侧为空。
	DBPath string `json:"dbPath,omitempty"`
}

// Versions 版本锚点。
type Versions struct {
	// DZAA002 该程序当前的规格版次（D 侧，来自 DZAA_T 的 MAX(dzaa002)）；取不到为 0。
	DZAA002 int `json:"dzaa002,omitempty"`
	// TglHeader `.tgl` 头部的 SD/PR 版次（"0025(2025-11-13 17:33:55)" 这种原样串）。
	TglHeader *TglHeader `json:"tglHeader,omitempty"`
	// TemplateVersions 模板版本：键是模板名去掉 `code_` 前缀与 `.template` 后缀
	// （`slice/code_a05.template` → `a05`），值是 `(Version:3)` 里的数字。
	TemplateVersions map[string]int `json:"templateVersions,omitempty"`
}

// TglHeader .tgl 文件头部的版本信息（生成器写在注释里）。
type TglHeader struct {
	SDVer string `json:"sdVer,omitempty"`
	PRVer string `json:"prVer,omitempty"`
	Build string `json:"build,omitempty"`
	// Desc 程序说明（`#+ Description: 訂單維護作業`）。
	Desc string `json:"desc,omitempty"`
}

// FileEntry 一个落盘文件。
type FileEntry struct {
	Role   string `json:"role"` // tab/tsd/tap/tgl/tgx/tgx2/4gl/inc/form/design/mdl/dict
	Src    string `json:"src"`  // 服务器侧相对 $TOP 的路径（M 侧 dict 为 "dict"，design 为 "local"）
	Rel    string `json:"rel"`  // 本地相对路径（相对该材料目录）
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	// SrcMtime 源文件的 mtime（RFC3339，服务器时区原样）。本地归档（design）为空。
	SrcMtime string `json:"srcMtime,omitempty"`
	// Rows 只有 role=table 的历史形态会用到；现在的表行进 Tables。
	Rows int64 `json:"rows,omitempty"`
}

// TableEntry 规格库里的一张表。
type TableEntry struct {
	Table string `json:"table"`
	Rows  int64  `json:"rows"`
	// Skipped 因超行数上限被跳过（行数为实测值，未落库）。
	Skipped bool `json:"skipped,omitempty"`
	// Missing 本环境的库里没有这张表（已记进 missing[]，不判失败）。
	//
	// 单独留一条而不是干脆不记：`tables[]` 的长度要恒等于角色表的张数（53），
	// 否则"镜了几张、缺几张"要靠集合差集算，而差集算错没人会发现。
	Missing bool   `json:"missing,omitempty"`
	Desc    string `json:"desc,omitempty"`
}

// NewManifest 造一份空清单（字段都初始化成空数组而不是 nil —— JSON 里要出现 `[]`，
// 否则 model 侧 parse 出 null 会把"没有缺失"与"这块字段不存在"混成一种）。
func NewManifest(kind, env string) *Manifest {
	return &Manifest{
		Schema:   ManifestSchema,
		Kind:     kind,
		Env:      env,
		Files:    []FileEntry{},
		Missing:  []string{},
		Warnings: []string{},
		Versions: Versions{TemplateVersions: map[string]int{}},
	}
}

// AddFile 追一个文件条目并返回它（调用方常要再补 SrcMtime）。
func (m *Manifest) AddFile(e FileEntry) {
	m.Files = append(m.Files, e)
}

// AddMissing 记一个缺失项（去重）。
func (m *Manifest) AddMissing(rel string) {
	for _, x := range m.Missing {
		if x == rel {
			return
		}
	}
	m.Missing = append(m.Missing, rel)
}

// AddWarning 记一条警告（去重）。
func (m *Manifest) AddWarning(w string) {
	if strings.TrimSpace(w) == "" {
		return
	}
	for _, x := range m.Warnings {
		if x == w {
			return
		}
	}
	m.Warnings = append(m.Warnings, w)
}

// FileByRel 按本地相对路径取条目。
func (m *Manifest) FileByRel(rel string) *FileEntry {
	for i := range m.Files {
		if m.Files[i].Rel == rel {
			return &m.Files[i]
		}
	}
	return nil
}

// Bytes 本地文件总字节数。
func (m *Manifest) Bytes() int64 {
	var n int64
	for _, f := range m.Files {
		n += f.Size
	}
	return n
}

// ---------- 读写 ----------

// WriteManifest 原子写一份清单（先 .tmp 再 rename：半个 JSON 比没有清单更坏）。
func WriteManifest(dir string, m *Manifest) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败 (%s): %w", dir, err)
	}
	m.Schema = ManifestSchema
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 manifest 失败: %w", err)
	}
	b = append(b, '\n')
	final := filepath.Join(dir, ManifestFileName)
	tmp := final + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return fmt.Errorf("写 manifest 失败 (%s): %w", tmp, err)
	}
	if err := os.Rename(tmp, final); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("写 manifest 失败 (%s): %w", final, err)
	}
	return nil
}

// ReadManifest 读一份清单。文件不存在返回 (nil, nil) —— "没拉过"是正常状态，
// 不是错误（status 要能对没拉过的环境/程序安静地回答"未拉取"）。
func ReadManifest(dir string) (*Manifest, error) {
	p := filepath.Join(dir, ManifestFileName)
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("读 manifest 失败 (%s): %w", p, err)
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("%s 不是合法 JSON: %w", p, err)
	}
	if m.Schema != "" && m.Schema != ManifestSchema {
		// 不硬失败：旧契约的清单仍然能读出"拉取时间/文件"，只是 stale 判定可能不准。
		// 但必须让用户看见 —— 静默按新规则解释旧数据，结论会悄悄错。
		m.AddWarning(fmt.Sprintf("manifest 契约版本为 %q（当前 %q），建议重新拉取", m.Schema, ManifestSchema))
	}
	return &m, nil
}

// ---------- sha256 ----------

// FileSHA256 算本地文件的 sha256（十六进制小写）。
func FileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// BytesSHA256 算一段字节的 sha256（下载时边写边算，省一次读盘）。
func BytesSHA256(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ---------- 版本解析 ----------

// reTemplateVersion 匹配模板头部的 `(Version:3)`。
//
// 形态来自实际语料（`code_a05.template` 第一行）：
//
//	#應用 a05 樣板自動產生(Version:3)
var reTemplateVersion = regexp.MustCompile(`\(Version:(\d+)\)`)

// ParseTemplateVersion 从模板内容里取版本号；取不到返回 (0,false)。
//
// 只看前 4 KB：版本标记在文件头，扫全文件对大模板（`code_t01.template` 279 KB）没必要。
func ParseTemplateVersion(head []byte) (int, bool) {
	if len(head) > 4096 {
		head = head[:4096]
	}
	m := reTemplateVersion.FindSubmatch(head)
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(string(m[1]))
	if err != nil {
		return 0, false
	}
	return n, true
}

// TemplateKey 把模板文件名折成 `versions.templateVersions` 的键：
// `code_a05.template` → `a05`；`slice/code_qs01.template` → `qs01`。
//
// 判据只认 `code_` 前缀 + `.template` 后缀：
//   - `4rp_*.template`/`*.snippet` 没有这个版本标记，硬套一个键会让"没有版本"与
//     "版本 0"混起来；
//   - `code_g01.template.141119` 这类**带日期的备份**也不是模板（它是旧版本的副本），
//     把它的版本号也记进 templateVersions 只会让"模板版本表"里多出几个不存在的模板。
func TemplateKey(rel string) (string, bool) {
	base := filepath.Base(rel)
	if !strings.HasPrefix(base, "code_") || !strings.HasSuffix(base, ".template") {
		return "", false
	}
	base = strings.TrimPrefix(base, "code_")
	base = strings.TrimSuffix(base, ".template")
	if base == "" {
		return "", false
	}
	return base, true
}

// IsTemplateFile 判"这个文件算不算一个程序样板"（状态里的「模板 N 个」按它计数）。
//
// 口径与 plan §4 冻结的一致：`code_*`（45，含 7 个带日期的备份副本）+ `slice/`（126）
// = 171。带日期的备份**算文件、不算版本锚点**（见 TemplateKey）。
//
// 为什么把备份也算进来：它们是生成器会去读的真实文件（按名字 `code_i01.template` 之类
// 取模板时不会取备份，但它们在同一个目录里、也在镜像范围内）。状态的"模板数"回答的是
// "这套材料里有多少个样板相关文件"，而"版本锚点"回答的是"哪个模板是第几版"——
// 两个问题的答案本来就不一样，混成一个数只会让人对不上文档。
func IsTemplateFile(rel string) bool {
	base := filepath.Base(rel)
	if strings.HasPrefix(base, "code_") {
		return true
	}
	// slice/ 下的切片是 code_*.template 的另一批（同样以 code_ 开头），
	// 这里兜住路径形态，避免将来切片改了命名就统计不到。
	dir := filepath.ToSlash(filepath.Dir(rel))
	return strings.HasSuffix(dir, "slice")
}

// reSDVer / rePRVer / reBuild / reDesc 解析 `.tgl` 头部。
//
// 形态来自实际语料（axmt500_wf.tgl 第 5–9 行）：
//
//	#+ Standard Version.....: SD版次:0025(2025-11-13 17:33:55), PR版次:
//	#+ Customerized Version.: SD版次:0000(1900-01-01 00:00:00), PR版次:
//	#+ Build......: 000037
//	#+ Description: 訂單維護作業
var (
	reSDVer = regexp.MustCompile(`SD版次:\s*([^,)\s]*\([^)]*\)|[^,)\s]*)`)
	rePRVer = regexp.MustCompile(`PR版次:\s*([^,\s]*)`)
	reBuild = regexp.MustCompile(`#\+\s*Build\.*\s*:\s*(\S+)`)
	reDesc  = regexp.MustCompile(`(?m)#\+\s*Description\s*:\s*(.+)$`)
)

// ParseTglHeader 从 `.tgl` 头部取版次；一个都没取到返回 nil。
//
// 只看前 8 KB：头部在最前面，而 `.tgl` 本体可以到 1.3 MB（实测 axmt500_wf.tgl）。
func ParseTglHeader(head []byte) *TglHeader {
	if len(head) > 8192 {
		head = head[:8192]
	}
	text := string(head)
	h := &TglHeader{}
	if m := reSDVer.FindStringSubmatch(text); m != nil {
		h.SDVer = strings.TrimSpace(m[1])
	}
	if m := rePRVer.FindStringSubmatch(text); m != nil {
		h.PRVer = strings.TrimSpace(m[1])
	}
	if m := reBuild.FindStringSubmatch(text); m != nil {
		h.Build = strings.TrimSpace(m[1])
	}
	if m := reDesc.FindStringSubmatch(text); m != nil {
		h.Desc = strings.TrimSpace(m[1])
	}
	if h.SDVer == "" && h.PRVer == "" && h.Build == "" && h.Desc == "" {
		return nil
	}
	return h
}

// ---------- stale 判定 ----------

// StaleReason 一条 stale 原因（给人看的中文短句）。
type StaleReason struct {
	Code   string `json:"code"` // missing-local | sha-mismatch | remote-newer | remote-gone | version-changed
	Detail string `json:"detail"`
}

// RemoteStamp 服务器侧一个文件的当前状态（probe 时得到）。
type RemoteStamp struct {
	Src     string
	Size    int64
	ModTime time.Time
	Exists  bool
}

// LocalStale 只查本地：清单里登记的文件是否还在、内容是否被改过。
//
// 不出网，所以 status 打开即快（plan §5.5：status 不触发网络，除非 probe=1）。
func LocalStale(dir string, m *Manifest) []StaleReason {
	if m == nil {
		return []StaleReason{{Code: "missing-local", Detail: "没有 manifest（未完成或未拉取）"}}
	}
	var out []StaleReason
	for _, f := range m.Files {
		p := filepath.Join(dir, filepath.FromSlash(f.Rel))
		st, err := os.Stat(p)
		if err != nil {
			out = append(out, StaleReason{Code: "missing-local", Detail: "本地缺文件: " + f.Rel})
			continue
		}
		if f.Size > 0 && st.Size() != f.Size {
			out = append(out, StaleReason{Code: "sha-mismatch",
				Detail: fmt.Sprintf("本地文件大小变了: %s（清单 %d B，实际 %d B）", f.Rel, f.Size, st.Size())})
			continue
		}
		if f.SHA256 == "" {
			continue
		}
		got, err := FileSHA256(p)
		if err != nil {
			out = append(out, StaleReason{Code: "missing-local", Detail: "本地文件读不了: " + f.Rel + "（" + err.Error() + "）"})
			continue
		}
		if got != f.SHA256 {
			out = append(out, StaleReason{Code: "sha-mismatch",
				Detail: "本地文件内容变了: " + f.Rel})
		}
	}
	return out
}

// RemoteStale 拿服务器现状与清单比对（`tt gen check` / `status --probe`）。
//
// 四种触发（plan §8.1）：源文件变新（remote-newer）、源文件消失（remote-gone）、
// 大小变化（remote-newer）、版本变化（version-changed，模板版本与 .tgl 头部）。
// 只比 mtime 与大小，不重算服务器侧 sha：把 1.3 MB 的 .tgl 全下来算哈希，
// 只为回答"变了没"，代价远大于收益。
func RemoteStale(m *Manifest, remote map[string]RemoteStamp, localVersions map[string]int, localTgl *TglHeader) []StaleReason {
	if m == nil {
		return nil
	}
	var out []StaleReason
	// 按 src 路径排序，输出稳定（map 遍历无序，测试与人工比对都需要稳定顺序）
	srcs := make([]string, 0, len(remote))
	for s := range remote {
		srcs = append(srcs, s)
	}
	sort.Strings(srcs)
	for _, s := range srcs {
		r := remote[s]
		f := fileBySrc(m, s)
		if f == nil {
			continue // 清单里没有的源文件不参与比对（新加的素材不算 stale）
		}
		if !r.Exists {
			out = append(out, StaleReason{Code: "remote-gone", Detail: "服务器上已没有: " + s + "（本地 " + f.Rel + "）"})
			continue
		}
		if r.Size != f.Size {
			out = append(out, StaleReason{Code: "remote-newer",
				Detail: fmt.Sprintf("服务器文件大小变了: %s（清单 %d B，现在 %d B）", s, f.Size, r.Size)})
			continue
		}
		if f.SrcMtime != "" {
			if mt, err := time.Parse(time.RFC3339, f.SrcMtime); err == nil {
				// 留 2 秒容差：SFTP 的 mtime 只到秒，而不同服务器的时区/精度处理有差异。
				// 不设容差时，一次"刚拉完就 check"会自己报 stale。
				if r.ModTime.After(mt.Add(2 * time.Second)) {
					out = append(out, StaleReason{Code: "remote-newer",
						Detail: fmt.Sprintf("服务器文件更新了: %s（清单 %s，现在 %s）",
							s, mt.Format("2006-01-02 15:04:05"), r.ModTime.Format("2006-01-02 15:04:05"))})
				}
			}
		}
	}
	// 版本变化：模板版本（M 侧）与 .tgl 头部（D 侧）
	for k, v := range localVersions {
		if old, ok := m.Versions.TemplateVersions[k]; ok && old != v {
			out = append(out, StaleReason{Code: "version-changed",
				Detail: fmt.Sprintf("模板 %s 版本变了（清单 %d，现在 %d）", k, old, v)})
		}
	}
	if localTgl != nil && m.Versions.TglHeader != nil {
		if localTgl.SDVer != "" && m.Versions.TglHeader.SDVer != "" && localTgl.SDVer != m.Versions.TglHeader.SDVer {
			out = append(out, StaleReason{Code: "version-changed",
				Detail: fmt.Sprintf(".tgl 的 SD 版次变了（清单 %s，现在 %s）", m.Versions.TglHeader.SDVer, localTgl.SDVer)})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

// fileBySrc 按源路径找清单条目。
func fileBySrc(m *Manifest, src string) *FileEntry {
	for i := range m.Files {
		if m.Files[i].Src == src {
			return &m.Files[i]
		}
	}
	return nil
}

// ConsistencyWarnings 生成一致性提示（不阻断）。
//
// 核心判据：**tab 比 tgl 新 → 产物可能滞后**。设计器改了蓝图但没重产时，.tgl/.4gl
// 还是旧骨架，而 model 拿到的"为什么长这样"会指向一个已经不存在的设计。
// 只警示这个方向：反过来（tgl 比 tab 新）是正常的重产流程，天天报只会让人忽略警告。
func ConsistencyWarnings(m *Manifest) []string {
	if m == nil {
		return nil
	}
	tab := fileByRole(m, "tab")
	tgl := fileByRole(m, "tgl")
	if tab != nil && tgl != nil && tab.SrcMtime != "" && tgl.SrcMtime != "" {
		tm, err1 := time.Parse(time.RFC3339, tab.SrcMtime)
		gm, err2 := time.Parse(time.RFC3339, tgl.SrcMtime)
		if err1 == nil && err2 == nil && tm.After(gm.Add(time.Minute)) {
			mins := int(tm.Sub(gm).Minutes())
			m.AddWarning(fmt.Sprintf("tab(%s) 比 tgl(%s) 新 %d 分钟：蓝图改过但骨架未重产，产物可能滞后",
				tm.Format("01-02 15:04"), gm.Format("01-02 15:04"), mins))
		}
	}
	// 反向的"骨架比蓝图新"不报；但"只有 tgl 没有 tab"必须报 —— 那说明 fetch 拉了一缺一套。
	if tab == nil {
		m.AddWarning("缺少 tab（蓝图）：这套材料无法支撑\"为什么长这样\"的结论")
	}
	if tgl == nil {
		m.AddWarning("缺少 tgl（排序骨架）：无法判断产物顺序的来由")
	}
	return m.Warnings
}

// fileByRole 取第一个某角色的条目。
func fileByRole(m *Manifest, role string) *FileEntry {
	for i := range m.Files {
		if m.Files[i].Role == role {
			return &m.Files[i]
		}
	}
	return nil
}
