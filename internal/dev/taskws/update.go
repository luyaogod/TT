package taskws

// 刷新工作区：材料/索引跟上服务器，视图与包按"有没有活"决定要不要重建。
//
// 三条判断（顺序即安全级别）：
//
//  1. 视图有未落盘的改动 → **拒绝**。材料换了视图没换，行号会静默错；先 apply 或 verify。
//     --force 只更新材料与索引（不动视图与包），用于"我知道自己在干什么"。
//  2. 包里有已 apply 的改动（包 sha ≠ 视图基线 sha）→ 不重建视图与包。
//     那些改动是任务成果，重建会把它们冲掉；材料更新后视图就"落后于服务器"，别再在上面改。
//  3. 视图干净且包没被 apply 过 → 重新打包 + 重新导出视图，工作区整体对齐到最新材料。
//
// --check 只向服务器核对（mtime/大小），一个字节都不写。

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/luyaogod/TT/internal/devdata"
)

// UpdateRequest 是一次刷新请求。
type UpdateRequest struct {
	CfgPath string
	// Dir 工作区根目录。
	Dir string
	// Full 不比对，直接整包重拉（源文件 mtime 没变也拉）。
	Full bool
	// CheckOnly 只核对，不写盘。
	CheckOnly bool
	// NoMdl 不更新模板切片。
	NoMdl bool
	// Force 有未落盘改动时仍然只刷新材料与索引（不动视图/包）。
	Force bool
	// Files 仅供测试注入（离线跑通 Check → Fetch → 重建索引/包/视图三条分支）；
	// nil 时按环境打开真 SFTP（与 devdata 的 StatusRequest/FetchRequest.Files 同一条缝）。
	Files devdata.FileSource
	// Progress 进度回调（可为 nil）。
	Progress devdata.ProgressFunc
}

// UpdateResult 是一次刷新的结果。
type UpdateResult struct {
	Checked      bool                  `json:"checked"`
	Stale        bool                  `json:"stale"`
	StaleReasons []devdata.StaleReason `json:"staleReasons,omitempty"`
	// Fetch 有值表示真拉了材料。
	Fetch        *devdata.FetchResult `json:"fetch,omitempty"`
	IndexRebuilt bool                 `json:"indexRebuilt"`
	Repacked     bool                 `json:"repacked"`
	Reexported   bool                 `json:"reexported"`
	Warnings     []string             `json:"warnings,omitempty"`
	Notes        []string             `json:"notes,omitempty"`
	Status       *Status              `json:"status"`
}

// Update 刷新工作区。
func Update(ctx context.Context, req UpdateRequest) (*UpdateResult, error) {
	t, p, err := Load(req.Dir)
	if err != nil {
		return nil, err
	}
	st, err := GetStatus(StatusRequest{Dir: req.Dir})
	if err != nil {
		return nil, err
	}
	res := &UpdateResult{Status: st}

	// 0) 包与视图基线不一致：先把状态修好再谈刷新（否则重拉材料也救不了 apply）。
	// --check 例外：它只是核对服务器，不碰包。
	if st.Pkg.Present && st.Ws.PkgSha != "" && !st.Pkg.MatchesBase && !req.CheckOnly {
		return nil, fmt.Errorf("%w: 包与视图基线不一致（包被换过/工作区是复制来的），先 tt dev relink "+
			"（以当前包重建视图）或把基线包放回来", ErrPackageMismatch)
	}

	// 1) 未落盘改动：默认拒绝。--check 例外 —— 它承诺不写盘，而"服务器有没有新规格"
	// 恰恰是带着改动时最想知道的（实测试用者被这条挡在门外）。
	if st.Ws.Present && st.Ws.Dirty && !req.Force && !req.CheckOnly {
		return nil, fmt.Errorf("%w: %s\n"+
			"  先 tt dev apply（把改动落成包）或 tt dev verify（只审计）；\n"+
			"  确实只想更新材料（不要视图）：--force（视图与包保持原样，之后别在上面继续改）",
			ErrDirtyWorkspace, p.Ws())
	}

	// 2) 向服务器核对。
	chk, err := devdata.Check(ctx, devdata.StatusRequest{
		CfgPath: req.CfgPath, Env: t.Env, Prog: t.Prog, Dir: p.Materials(),
		Files: req.Files,
	})
	if err != nil {
		return nil, err
	}
	res.Checked = true
	res.Stale = chk.Stale
	res.StaleReasons = chk.StaleReasons
	res.Warnings = append(res.Warnings, chk.Warnings...)

	needFetch := chk.Stale || req.Full
	needIndex := !st.Index.Fresh
	if req.CheckOnly {
		res.Notes = append(res.Notes, "只核对（--check）：没有改任何文件")
		return res, nil
	}
	if !needFetch && !needIndex {
		res.Notes = append(res.Notes, "材料与索引都是最新的，什么都没改")
		return res, nil
	}

	// 3) 拉材料（整目录原子替换）。
	if needFetch {
		fetch, ferr := devdata.Fetch(ctx, devdata.FetchRequest{
			CfgPath: req.CfgPath, Env: t.Env, Prog: t.Prog, Force: true, Dir: p.Materials(),
			Files: req.Files,
		}, req.Progress)
		if ferr != nil {
			return nil, ferr
		}
		res.Fetch = fetch
		res.Warnings = append(res.Warnings, fetch.Warnings...)
		if mwarn := appendMdl(req.CfgPath, t.Env, p.Mdl(), req.NoMdl); len(mwarn) > 0 {
			res.Warnings = append(res.Warnings, mwarn...)
		}
	}

	// 4) 重建索引（材料一变就必须重建，否则行号指向旧文本）。
	doc, merr := devdata.Map(devdata.LinkRequest{
		CfgPath: req.CfgPath, Env: t.Env, Prog: t.Prog,
		Dir: p.Materials(), Out: p.Index(), Mdl: p.Mdl(),
	})
	if merr != nil {
		return nil, merr
	}
	res.IndexRebuilt = true

	// 5) 视图/包：只在"没有活"时重建。
	refreshView := !st.Ws.Dirty && !st.Pkg.Modified
	switch {
	case !refreshView:
		res.Warnings = append(res.Warnings,
			"视图或包里已有改动：本次只更新材料与索引，视图仍基于旧材料（别再在上面改；要接着干活请重新 open）")
	case !needFetch:
		// 材料没变，视图/包没理由动。
	default:
		// 视图干净 + 包未 apply：删掉旧视图，按新重新打包并导出。
		if st.Ws.Present {
			if rerr := os.RemoveAll(p.Ws()); rerr != nil {
				return nil, fmt.Errorf("删旧视图失败 (%s): %w", p.Ws(), rerr)
			}
		}
		// pkg/ 必须是**已存在的目录**，否则 Pack 会把目录路径当包文件名。
		if merr := os.MkdirAll(p.PkgDir(), 0o755); merr != nil {
			return nil, fmt.Errorf("建包目录失败 (%s): %w", p.PkgDir(), merr)
		}
		pack, perr := devdata.Pack(devdata.PackRequest{
			CfgPath: req.CfgPath, Env: t.Env, Prog: t.Prog, Identity: t.Identity,
			Out: p.PkgDir(), Force: true, Dir: p.Materials(),
		})
		if perr != nil {
			return nil, perr
		}
		res.Repacked = true
		res.Warnings = append(res.Warnings, pack.Warnings...)
		exp, eerr := Export(ExportRequest{Dir: p.Ws(), Pkg: pack.Path})
		if eerr != nil {
			return nil, eerr
		}
		res.Reexported = true
		res.Warnings = append(res.Warnings, ExportWarnings(exp)...)
	}

	// 6) 记一笔
	detail := "材料未变，仅重建索引"
	if res.Fetch != nil {
		detail = fmt.Sprintf("材料已更新（%d 个文件）", len(res.Fetch.Files))
	}
	if res.Repacked {
		detail += "；包与视图已重建"
	}
	t.AddEvent("update", detail)
	if serr := Save(t, p); serr != nil {
		res.Warnings = append(res.Warnings, "task.json 没写成功："+serr.Error())
	}
	res.Notes = append(res.Notes, fmt.Sprintf("索引重建：%d 个点 / 客制 %d 个",
		doc.Stats.Points, doc.Stats.PointsCustom))
	// 状态要**现算一遍**：进 Update 时那份是动手前的快照，直接拿它展示会显示旧的材料
	// 时间与"未 apply"（实测踩过）。
	if st2, serr := GetStatus(StatusRequest{Dir: req.Dir}); serr == nil {
		res.Status = st2
	}
	res.Warnings = dedupe(res.Warnings)
	return res, nil
}

// dedupe 去掉重复的提示（Check 与 Fetch 都会说"取的是客制模块"，两条一模一样）。
func dedupe(list []string) []string {
	if len(list) < 2 {
		return list
	}
	seen := make(map[string]bool, len(list))
	out := list[:0]
	for _, s := range list {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// appendMdl 更新模板切片；返回要展示的提示。
func appendMdl(cfgPath, env, dst string, noMdl bool) []string {
	_, warns := copyMdl(cfgPath, env, dst, noMdl)
	return warns
}

// staleText 把 stale 原因拼成一句（CLI 输出用）。
func staleText(reasons []devdata.StaleReason) string {
	if len(reasons) == 0 {
		return ""
	}
	parts := make([]string, 0, len(reasons))
	for _, r := range reasons {
		parts = append(parts, r.Code+": "+r.Detail)
	}
	return strings.Join(parts, "；")
}
