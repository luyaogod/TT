package taskws

// relink：以**当前工作区里的包**为基线，重新导出视图。
//
// 为什么需要：apply 会校验"源包自 export 之后没被改过"。包一旦被替换（拷错、手工覆盖、
// 别人给了一份），这个校验会一直失败，而它给的出路（"把基线包放回来"）要求你手里还有
// 那份基线 —— 通常没有。于是工作区成了死局。
//
// relink 把这个死局打开：承认"当前这个包就是新的基线"，按它重导视图。代价是视图里
// 未落盘的改动会丢（所以默认拒绝脏工作区，要 --force）。

import (
	"fmt"
	"os"
)

// RelinkRequest 是一次 relink 请求。
type RelinkRequest struct {
	Dir string
	// Force 视图有未落盘改动时仍然重导（改动会丢）。
	Force bool
}

// RelinkResult 是重导结果。
type RelinkResult struct {
	Dir      string
	Pkg      string
	PkgSha   string
	Regions  int
	Editable int
	Warnings []string
}

// Relink 以当前包为基线重建视图。
func Relink(req RelinkRequest) (*RelinkResult, error) {
	t, p, err := Load(req.Dir)
	if err != nil {
		return nil, err
	}
	local := p.Pkg(t.Prog, identityOf(t))
	sha, _ := pkgSha(local)
	if sha == "" {
		return nil, fmt.Errorf("%w: 当前工作区没有包 %s（把包放回去再 relink）", ErrPackageMismatch, local)
	}
	if st, serr := GetStatus(StatusRequest{Dir: req.Dir}); serr == nil && st.Ws.Dirty && !req.Force {
		return nil, fmt.Errorf("%w: 视图有未落盘改动，relink 会按包重导视图、丢掉它们；先 tt dev apply，或 --force",
			ErrDirtyWorkspace)
	}
	// 旧视图直接删：它有 git 历史，而且与"当前包"已经对不上（否则不会走到 relink）。
	if err := os.RemoveAll(p.Ws()); err != nil {
		return nil, fmt.Errorf("删旧视图失败 (%s): %w", p.Ws(), err)
	}
	exp, err := Export(ExportRequest{Dir: p.Ws(), Pkg: local})
	if err != nil {
		return nil, err
	}
	t.AddEvent("relink", fmt.Sprintf("以当前包重建视图（%s）", short(sha)))
	if serr := Save(t, p); serr != nil {
		return nil, serr
	}
	return &RelinkResult{
		Dir: p.Ws(), Pkg: local, PkgSha: sha,
		Regions: exp.Regions, Editable: exp.Stats.Editable,
		Warnings: ExportWarnings(exp),
	}, nil
}
