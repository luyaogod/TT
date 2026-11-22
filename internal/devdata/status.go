package devdata

// 状态汇总：共用资料到什么程度、哪些程序拉过、有没有过期。
//
// **默认不出网**（plan §5.5）：设置页一打开就要出结果，所以 status 只读本地 manifest
// 与文件系统。`probe=1` / `tt gen check` 才连服务器比 mtime —— 那是"我现在就要确认
// 这份材料还新不新"的显式请求。两者的差集就是"本地被改过"（LocalStale）与
// "服务器那边变了"（RemoteStale）：前者不该触发联网，后者不联网根本看不见。

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/luyaogod/TT/internal/config"
)

// StatusRequest 一次状态查询的输入。
type StatusRequest struct {
	CfgPath string
	Env     string
	// Prog 只查某个程序（空 = 全部）。
	Prog string
	// Probe true = 向服务器核对 mtime/大小（慢，显式请求）。
	Probe bool
	// Dir 覆盖程序材料目录（工作区用；空 = <环境>/prog/<程序>）。
	Dir string
	// Files 仅供测试注入（probe 时才需要）；nil = 按环境打开真 SFTP。
	Files FileSource
	// Jobs 当前在跑的任务（由 Manager 提供；nil = 不出这一项）。
	Jobs []Job
}

// materialDir 本次要核对的材料目录（工作区显式给 Dir 时不再问配置）。
func (r StatusRequest) materialDir(prog string) (string, error) {
	if d := strings.TrimSpace(r.Dir); d != "" {
		return filepath.Abs(d)
	}
	return ProgDir(r.CfgPath, r.Env, prog)
}

// EnvStatus 一个环境的开发资料状态。
type EnvStatus struct {
	Env     string       `json:"env"`
	Dir     string       `json:"dir"`
	Present bool         `json:"present"` // 该环境目录是否存在
	Dict    *DictStatus  `json:"dict"`
	Mdl     *MdlStatus   `json:"mdl"`
	Progs   []ProgStatus `json:"progs"`
	Jobs    []Job        `json:"jobs,omitempty"`
	// Warnings 环境级警告（manifest 的 warnings + 本地不一致）。
	Warnings []string `json:"warnings,omitempty"`
}

// DictStatus 规格库状态。
type DictStatus struct {
	Present  bool   `json:"present"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Tables   int    `json:"tables"`
	Skipped  int    `json:"skipped"`
	Rows     int64  `json:"rows"`
	SyncedAt string `json:"syncedAt,omitempty"`
	SHA256   string `json:"sha256,omitempty"`
}

// MdlStatus 模板树状态。
type MdlStatus struct {
	Present          bool           `json:"present"`
	Dir              string         `json:"dir"`
	Files            int            `json:"files"`
	Bytes            int64          `json:"bytes"`
	SyncedAt         string         `json:"syncedAt,omitempty"`
	TemplateVersions map[string]int `json:"templateVersions,omitempty"`
	// TemplateCount plan 里的"模板 171 个"口径：code_* 模板 + slice 切片。
	TemplateCount int `json:"templateCount"`
}

// ProgStatus 一个程序的材料状态。
type ProgStatus struct {
	Prog         string        `json:"prog"`
	Module       string        `json:"module,omitempty"`
	Dir          string        `json:"dir"`
	Exists       bool          `json:"exists"`
	Files        int           `json:"files"`
	Bytes        int64         `json:"bytes"`
	FetchedAt    string        `json:"fetchedAt,omitempty"`
	Stale        bool          `json:"stale"`
	StaleReasons []StaleReason `json:"staleReasons,omitempty"`
	Missing      []string      `json:"missing,omitempty"`
	Warnings     []string      `json:"warnings,omitempty"`
	// Probed true = 本次比对过服务器。
	Probed bool `json:"probed,omitempty"`
}

// Status 汇总某环境的状态（默认不出网）。
func Status(ctx context.Context, req StatusRequest) (*EnvStatus, error) {
	envDir, err := EnvDir(req.CfgPath, req.Env)
	if err != nil {
		return nil, err
	}
	e, err := envInfo(req.CfgPath, req.Env)
	if err != nil {
		return nil, err
	}
	out := &EnvStatus{Env: e.Name, Dir: envDir, Progs: []ProgStatus{}, Jobs: req.Jobs}
	if st, serr := os.Stat(envDir); serr != nil || !st.IsDir() {
		// 没拉过：仍然回一个完整的空状态（前端据此画"未同步"），不报错
		out.Dict = &DictStatus{Path: config.DevDataDictPathForEnv(req.CfgPath, e.Name)}
		out.Mdl = &MdlStatus{Dir: config.DevDataMdlDirForEnv(req.CfgPath, e.Name)}
		return out, nil
	}
	out.Present = true

	m, merr := ReadManifest(envDir)
	if merr != nil {
		out.Warnings = append(out.Warnings, "manifest 读不出来："+merr.Error())
	}
	if m != nil {
		out.Warnings = append(out.Warnings, m.Warnings...)
	}

	dictPath := config.DevDataDictPathForEnv(req.CfgPath, e.Name)
	out.Dict = &DictStatus{Path: dictPath}
	if st, serr := os.Stat(dictPath); serr == nil {
		out.Dict.Present = true
		out.Dict.Size = st.Size()
	}
	if m != nil {
		out.Dict.Tables = countTables(m)
		out.Dict.Rows = sumTableRows(m)
		for _, t := range m.Tables {
			if t.Skipped {
				out.Dict.Skipped++
			}
		}
		out.Dict.SyncedAt = m.FetchedAt
		if f := m.FileByRel(filepath.ToSlash(filepath.Join(config.DevDataSubdirDict, config.DictDBFileName(e.Name)))); f != nil {
			out.Dict.SHA256 = f.SHA256
		}
	}
	out.Mdl = &MdlStatus{Dir: config.DevDataMdlDirForEnv(req.CfgPath, e.Name)}
	if m != nil {
		out.Mdl.Files = countRole(m, "mdl")
		out.Mdl.Bytes = sumRoleBytes(m, "mdl")
		out.Mdl.SyncedAt = m.FetchedAt
		out.Mdl.TemplateVersions = m.Versions.TemplateVersions
		for _, f := range m.Files {
			if f.Role != "mdl" {
				continue
			}
			if IsTemplateFile(trimMdlPrefix(f.Rel)) {
				out.Mdl.TemplateCount++
			}
		}
	}
	if st, serr := os.Stat(out.Mdl.Dir); serr == nil && st.IsDir() {
		out.Mdl.Present = true
	}

	// 程序列表：以磁盘上的目录为准（不是以 manifest 为准）—— 目录在就说明拉过，
	// 即使 manifest 被删了也要能看见（那种状态下 stale=true、原因写明没有 manifest）。
	progRoot := filepath.Join(envDir, config.DevDataSubdirProg)
	entries, _ := os.ReadDir(progRoot)
	names := make([]string, 0, len(entries))
	for _, en := range entries {
		if !en.IsDir() || isTempName(en.Name()) {
			continue
		}
		names = append(names, en.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		if req.Prog != "" && name != req.Prog {
			continue
		}
		ps, perr := progStatus(ctx, req, filepath.Join(progRoot, name), name)
		if perr != nil {
			return nil, perr
		}
		out.Progs = append(out.Progs, *ps)
	}
	return out, nil
}

// Check 只查一个程序的新鲜度（plan 的 `tt gen check`：强制 probe）。
func Check(ctx context.Context, req StatusRequest) (*ProgStatus, error) {
	if err := ValidProgName(req.Prog); err != nil {
		return nil, err
	}
	progDir, err := req.materialDir(strings.TrimSpace(req.Prog))
	if err != nil {
		return nil, err
	}
	if _, serr := os.Stat(progDir); serr != nil {
		return nil, fmt.Errorf("%w（本地没有该程序的材料，先 tt gen fetch %s）", ErrProgNotFound, strings.TrimSpace(req.Prog))
	}
	req.Probe = true
	return progStatus(ctx, req, progDir, strings.TrimSpace(req.Prog))
}

// progStatus 汇总一个程序目录。
func progStatus(ctx context.Context, req StatusRequest, dir, prog string) (*ProgStatus, error) {
	out := &ProgStatus{Prog: prog, Dir: dir, Exists: true}
	m, err := ReadManifest(dir)
	if err != nil {
		out.Stale = true
		out.StaleReasons = append(out.StaleReasons, StaleReason{Code: "missing-local", Detail: err.Error()})
		return out, nil
	}
	if m == nil {
		out.Stale = true
		out.StaleReasons = append(out.StaleReasons, StaleReason{Code: "missing-local", Detail: "没有 manifest（未完成或未拉取）"})
		return out, nil
	}
	out.Module = m.Module
	out.FetchedAt = m.FetchedAt
	out.Files = len(m.Files)
	out.Bytes = m.Bytes()
	out.Missing = m.Missing
	out.Warnings = m.Warnings

	reasons := LocalStale(dir, m)
	if req.Probe {
		src, closeAll, ferr := probeSource(req)
		if ferr != nil {
			return nil, ferr
		}
		if closeAll != nil {
			defer closeAll()
		}
		remote := map[string]RemoteStamp{}
		for _, f := range m.Files {
			// 只核对**服务器侧**的素材：design 是本地归档（Src=local:*），dict 是库文件
			if f.Src == "" || len(f.Src) > 6 && f.Src[:6] == "local:" || f.Role == "dict" {
				continue
			}
			st, serr := src.Stat(ctx, f.Src)
			if serr != nil {
				if errors.Is(serr, fs.ErrNotExist) || errors.Is(serr, os.ErrNotExist) {
					remote[f.Src] = RemoteStamp{Src: f.Src, Exists: false}
					continue
				}
				return nil, serr
			}
			remote[f.Src] = RemoteStamp{Src: f.Src, Size: st.Size, ModTime: st.ModTime, Exists: true}
		}
		var tv map[string]int
		var tgl *TglHeader
		if f := fileByRole(m, "tgl"); f != nil {
			if data, rerr := os.ReadFile(filepath.Join(dir, filepath.FromSlash(f.Rel))); rerr == nil {
				tgl = ParseTglHeader(data)
			}
		}
		reasons = append(reasons, RemoteStale(m, remote, tv, tgl)...)
		out.Probed = true
	}
	if len(reasons) > 0 {
		out.Stale = true
		out.StaleReasons = reasons
	}
	return out, nil
}

// probeSource 拿一个（必要时现开的）文件来源；返回的 closeAll 由调用方 defer。
func probeSource(req StatusRequest) (FileSource, func(), error) {
	if req.Files != nil {
		return req.Files, nil, nil
	}
	e, err := envInfo(req.CfgPath, req.Env)
	if err != nil {
		return nil, nil, err
	}
	return OpenFileSource(e)
}

// isTempName 判临时目录名（`xxx.tmp-` 由 NewTempDir 生成，这里挡住构建残留）。
func isTempName(name string) bool {
	return strings.Contains(name, ".tmp-")
}

// trimMdlPrefix 去掉 manifest.rel 里的 `mdl/` 前缀。
func trimMdlPrefix(rel string) string {
	const p = "mdl/"
	if len(rel) > len(p) && rel[:len(p)] == p {
		return rel[len(p):]
	}
	return rel
}

// AgeText 把"上次同步时间"渲染成"3 分钟前"这种相对说法（CLI/Web 共用）。
func AgeText(stamp string) string {
	if stamp == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, stamp)
	if err != nil {
		return stamp
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "刚刚"
	case d < time.Hour:
		return fmt.Sprintf("%d 分钟前", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d 小时前", int(d.Hours()))
	default:
		return fmt.Sprintf("%d 天前", int(d.Hours()/24))
	}
}
