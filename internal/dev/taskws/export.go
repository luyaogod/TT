package taskws

// 包 → 工作区（渲染围栏视图）。这是 `tt dev tzc export` 的核心，被工作区与 tzc 两条
// 命令共用：同一份实现才不会出现"工作区导出的围栏"和"tzc export 的围栏"差一个字符。
//
// 之所以放在 dev/taskws 而不是 cli 层：这一步没有参数解析、没有输出格式，是纯领域动作
// （读包 → 合成 → 渲染 → 落盘 → git commit）。

import (
	"fmt"

	"github.com/luyaogod/TT/internal/dev/fence"
	"github.com/luyaogod/TT/internal/dev/model"
	"github.com/luyaogod/TT/internal/dev/pkgfile"
	"github.com/luyaogod/TT/internal/dev/store"
	"github.com/luyaogod/TT/internal/dev/synth"
)

// ExportRequest 是一次导出请求。
type ExportRequest struct {
	// Dir 工作区目录（不存在则创建；已存在且是工作区则拒绝）。
	Dir string
	// Pkg 源 .tzc 路径。
	Pkg string
	// Only 部分导出：只把这些点渲染成可编辑（空 = 全部可编辑点）。
	Only []string
	// AcceptVer 可接受的 ver 主次版本（空 = 1.0）。
	AcceptVer string
}

// ExportResult 是导出结果（CLI 的输出与 task.json 都从这里取数）。
type ExportResult struct {
	Dir     string
	Pkg     string
	Prog    string
	Kind    string
	Ver     string
	Regions int
	Stats   store.Stats
	Anchors map[string]bool
	Section store.SectionInfo
	Commit  string
	// Manifest 是落盘后的工作区清单（CLI 直接用，避免再 Open 一次）。
	Manifest *store.Manifest
}

// Export 把 .tzc 渲染成围栏工作区。
//
// 与 `tt dev tzc export` 的行为一致：拒绝覆盖已存在的工作区；git 提交失败只警告
// （工作区仍然可用）。唯一的差别是不打印任何东西 —— 输出由调用方决定。
func Export(req ExportRequest) (*ExportResult, error) {
	ver := req.AcceptVer
	if ver == "" {
		ver = "1.0"
	}
	pkg, err := pkgfile.Open(req.Pkg, pkgfile.OpenOptions{AcceptVer: ver})
	if err != nil {
		return nil, err
	}
	doc, err := synth.Synthesize(pkg, synth.Options{Only: req.Only})
	if err != nil {
		return nil, err
	}
	fenced, regions, spans, err := fence.Render(doc)
	if err != nil {
		return nil, err
	}
	doc.Text = fenced
	doc.Regions = regions
	doc.Spans = spans

	ws, err := store.Create(req.Dir, doc, pkg, fenced)
	if err != nil {
		return nil, err
	}
	commit, gerr := ws.GitCommit("tt dev open: " + doc.Prog)
	mf, merr := ws.Manifest()
	if merr != nil {
		return nil, fmt.Errorf("读工作区清单失败 (%s): %w", req.Dir, merr)
	}
	res := &ExportResult{
		Dir: req.Dir, Pkg: req.Pkg, Prog: doc.Prog, Kind: pkg.Kind.String(), Ver: pkg.Ver.Raw,
		Regions: len(doc.AllRegions()), Stats: mf.Stats, Anchors: mf.Anchors,
		Section: mf.Section, Commit: commit, Manifest: mf,
	}
	if gerr != nil {
		res.Commit = "" // 空 = 没有提交（工作区仍可用）
	}
	return res, nil
}

// ExportWarnings 收集导出后该让人看见的提示（git 没提交、框架未解锁）。
func ExportWarnings(res *ExportResult) []string {
	var out []string
	if res.Commit == "" {
		out = append(out, "git 提交失败（工作区仍可用，但改动不便于回溯）")
	}
	if res.Section.State != string(model.SectionUnlocked) {
		out = append(out, "框架未解锁（改加点不需要它；要改 section 正文才需 tt dev unlock，单向、有代价）")
	}
	return out
}
