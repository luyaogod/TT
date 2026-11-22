package taskws

// 工作区状态：**全部现算**，一个字段都不从 task.json 读。
//
// 为什么要这么严：task.json 里存"是否脏""包 sha"这类量，第一次忘了更新就开始骗人，
// 而骗的方向总是"看起来还挺新"。现算的代价是读几份小文件 + 一次包 sha（10 MB 内存读），
// 几十毫秒，完全付得起。

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/luyaogod/TT/internal/dev/pkgfile"
	"github.com/luyaogod/TT/internal/dev/store"
	"github.com/luyaogod/TT/internal/devdata"
)

// StatusRequest 是一次状态查询。
type StatusRequest struct {
	// Dir 工作区根目录。
	Dir string
}

// MaterialStatus 是材料侧状态。
type MaterialStatus struct {
	Dir       string   `json:"dir"`
	FetchedAt string   `json:"fetchedAt,omitempty"`
	Module    string   `json:"module,omitempty"`
	Files     int      `json:"files"`
	Bytes     int64    `json:"bytes"`
	Missing   []string `json:"missing,omitempty"`
	DZAA002   int      `json:"dzaa002,omitempty"`
}

// IndexStatus 是索引侧状态。
type IndexStatus struct {
	Path        string `json:"path"`
	Present     bool   `json:"present"`
	Fresh       bool   `json:"fresh"`
	Schema      string `json:"schema,omitempty"`
	GeneratedAt string `json:"generatedAt,omitempty"`
	Points      int    `json:"points"`
	Custom      int    `json:"custom"`
	Sections    int    `json:"sections"`
	Slices      int    `json:"slices"`
}

// WsStatus 是可编辑视图状态。
type WsStatus struct {
	Dir          string `json:"dir"`
	Present      bool   `json:"present"`
	Dirty        bool   `json:"dirty"`
	SectionState string `json:"sectionState,omitempty"`
	Regions      int    `json:"regions"`
	Editable     int    `json:"editable"`
	// PkgSha 是视图导出时所基于的包 sha（apply 之后会变成新包 sha）。
	PkgSha string `json:"pkgSha,omitempty"`
}

// PkgStatus 是要交出去的包的状态。
type PkgStatus struct {
	Path    string `json:"path"`
	Present bool   `json:"present"`
	Bytes   int64  `json:"bytes"`
	Sha256  string `json:"sha256,omitempty"`
	// Modified true = 包里的 .tap/.tgl 与**材料**不同 —— 也就是 apply 过。
	// 这类包是任务成果（可以交出去）；但它上面的视图不再与包一致，别继续在视图上改。
	//
	// 判据刻意不是"包 sha 变了"：apply 之后 tzc 会把工作区基线一起更新，sha 比对会把
	// 刚做完的活判成"没动过"。
	Modified bool `json:"modified"`
	// MatchesBase true = 包的 sha256 与视图基线记的一致（可以安全 apply）。
	// false 且包存在 = 包被换过、或工作区是复制来的（此时 apply 会拒绝写回）。
	MatchesBase bool `json:"matchesBase"`
}

// MdlStatus 是模板切片状态。
type MdlStatus struct {
	Dir       string `json:"dir"`
	Present   bool   `json:"present"`
	Templates int    `json:"templates"`
}

// Status 是工作区的一页状态。
type Status struct {
	Dir       string         `json:"dir"`
	Prog      string         `json:"prog"`
	Env       string         `json:"env"`
	Module    string         `json:"module,omitempty"`
	Identity  string         `json:"identity,omitempty"`
	Task      string         `json:"task,omitempty"`
	Ent       int            `json:"ent,omitempty"`
	CreatedAt string         `json:"createdAt,omitempty"`
	UpdatedAt string         `json:"updatedAt,omitempty"`
	Materials MaterialStatus `json:"materials"`
	Index     IndexStatus    `json:"index"`
	Ws        WsStatus       `json:"ws"`
	Pkg       PkgStatus      `json:"pkg"`
	Mdl       MdlStatus      `json:"mdl"`
	Notes     []string       `json:"notes,omitempty"`
}

// GetStatus 现算工作区状态。
func GetStatus(req StatusRequest) (*Status, error) {
	t, p, err := Load(req.Dir)
	if err != nil {
		return nil, err
	}
	out := &Status{
		Dir: p.Root, Prog: t.Prog, Env: t.Env, Module: t.Module, Identity: t.Identity,
		Task: t.Task, Ent: t.Ent, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
	// 材料
	out.Materials.Dir = p.Materials()
	if mf, merr := devdata.ReadManifest(p.Materials()); merr != nil {
		out.Notes = append(out.Notes, "材料清单读不出来："+merr.Error())
	} else if mf != nil {
		out.Materials.FetchedAt = mf.FetchedAt
		out.Materials.Module = mf.Module
		out.Materials.Files = len(mf.Files)
		out.Materials.Bytes = mf.Bytes()
		out.Materials.Missing = mf.Missing
		out.Materials.DZAA002 = mf.Versions.DZAA002
	} else {
		out.Notes = append(out.Notes, "没有材料清单（materials/manifest.json）—— 材料不完整")
	}

	// 索引
	out.Index.Path = p.Index()
	linkReq := devdata.LinkRequest{
		CfgPath: "", Env: t.Env, Prog: t.Prog, Dir: p.Materials(), Out: p.Index(), Mdl: p.Mdl(),
	}
	if doc, derr := devdata.ReadLinkDoc(p.Index()); derr == nil && doc != nil {
		out.Index.Present = true
		out.Index.Schema = doc.Schema
		out.Index.GeneratedAt = doc.GeneratedAt
		out.Index.Points = doc.Stats.Points
		out.Index.Custom = doc.Stats.PointsCustom
		out.Index.Sections = len(doc.Sections)
		out.Index.Slices = doc.Slices.Distinct
	}
	if fresh, ferr := devdata.IndexFresh(linkReq); ferr == nil {
		out.Index.Fresh = fresh
	} else {
		out.Notes = append(out.Notes, "索引新鲜度查不出来："+ferr.Error())
	}
	if out.Index.Present && !out.Index.Fresh {
		out.Notes = append(out.Notes, "索引与材料不一致（材料变过或 tt 版本变了）—— 用 tt dev map 重建")
	}

	// 视图
	out.Ws.Dir = p.Ws()
	if ws, werr := store.Open(p.Ws()); werr == nil {
		out.Ws.Present = true
		if edited, eerr := ws.ReadEdited(); eerr == nil {
			if base, berr := ws.ReadBase(); berr == nil {
				out.Ws.Dirty = !sameFile(edited, base)
			}
		}
		if mf, merr := ws.Manifest(); merr == nil && mf != nil {
			out.Ws.SectionState = mf.Section.State
			out.Ws.Regions = mf.Stats.Points + mf.Stats.Sections
			out.Ws.Editable = mf.Stats.Editable
			out.Ws.PkgSha = mf.Pkg.Sha256
		}
	} else if !os.IsNotExist(werr) {
		out.Notes = append(out.Notes, "视图读不出来："+werr.Error())
	}

	// 包
	out.Pkg.Path = p.Pkg(t.Prog, identityOf(t))
	if st, serr := os.Stat(out.Pkg.Path); serr == nil && !st.IsDir() {
		out.Pkg.Present = true
		out.Pkg.Bytes = st.Size()
		sha, _ := pkgSha(out.Pkg.Path)
		out.Pkg.Sha256 = sha
		modified, merr := pkgModified(out.Pkg.Path, p.Materials(), sha, out.Ws.PkgSha)
		if merr != nil {
			out.Notes = append(out.Notes, "包内容与材料对比失败："+merr.Error())
		}
		out.Pkg.Modified = modified
		out.Pkg.MatchesBase = out.Pkg.Sha256 != "" && out.Ws.PkgSha != "" && out.Pkg.Sha256 == out.Ws.PkgSha
		if out.Pkg.MatchesBase {
			// 正常情况（export 之后 / apply 之后）什么都不用说。
		} else if out.Ws.PkgSha != "" {
			out.Notes = append(out.Notes, "包与视图基线不一致（包被换过，或这个工作区是复制/移动来的）："+
				"apply/verify 会拒绝。两条真出路：把基线那份包放回来，或 tt dev relink"+
				"（以当前包重新导出视图，未落盘的改动会丢）")
		}
	}

	// 模板切片
	out.Mdl.Dir = p.Mdl()
	if n, ok := countSliceTemplates(p.Mdl()); ok {
		out.Mdl.Present = true
		out.Mdl.Templates = n
	}

	// 提示（只讲"下一步会踩的坑"）
	if out.Ws.Present && out.Ws.Dirty {
		out.Notes = append(out.Notes, "视图有未落盘的改动：先 tt dev apply（或 tt dev verify 看审计）")
	}
	return out, nil
}

// pkgModified 判断包里的 .tap/.tgl 是否与材料不同（= apply 过）。
//
// 材料里能取到 tap/tgl 时按字节比（**这是唯一诚实的判据**：apply 后工作区基线会被一起
// 更新，拿包 sha 与基线比会把刚做完的活判成没动过）。材料缺件时退回"包 sha ≠ 视图基线"，
// 并在调用方那边记一条提示。
func pkgModified(pkgPath, matDir, pkgSha, wsPkgSha string) (bool, error) {
	pkg, err := pkgfile.Open(pkgPath, pkgfile.OpenOptions{AcceptVer: "1.0"})
	if err != nil {
		return false, err
	}
	mf, merr := devdata.ReadManifest(matDir)
	if merr != nil || mf == nil || len(mf.Files) == 0 {
		return pkgSha != "" && wsPkgSha != "" && pkgSha != wsPkgSha, merr
	}
	compared := 0
	for _, role := range []string{"tap", "tgl"} {
		var rel string
		for _, f := range mf.Files {
			if f.Role == role {
				rel = f.Rel
				break
			}
		}
		if rel == "" {
			continue
		}
		want, rerr := os.ReadFile(filepath.Join(matDir, rel))
		if rerr != nil {
			continue
		}
		entry := ""
		for _, e := range pkg.Entries {
			if strings.HasSuffix(strings.ToLower(e.Name), "."+role) {
				entry = string(e.Data)
				break
			}
		}
		if entry == "" {
			continue // 包里没有这个成员（标准程序可能没有 tap）
		}
		compared++
		if !sameFile(want, []byte(entry)) {
			return true, nil
		}
	}
	if compared == 0 {
		return pkgSha != "" && wsPkgSha != "" && pkgSha != wsPkgSha, nil
	}
	return false, nil
}

// PackagePath 返回工作区里那个包的路径（不存在返回空串，不报错）。
func PackagePath(dir string) (string, error) {
	t, p, err := Load(dir)
	if err != nil {
		return "", err
	}
	path := p.Pkg(t.Prog, identityOf(t))
	if _, serr := os.Stat(path); serr != nil {
		// 与 PkgPoint 同一条回退：视图清单里记的源包也可能是真相
		// （老工作区/手工搬过包）。只在这里有、PkgPoint 没有，会让
		// `search --scope pkg` 静默搜不到而 `at --scope pkg` 又能看到。
		if ws, werr := store.Open(p.Ws()); werr == nil {
			if mf, merr := ws.Manifest(); merr == nil && mf.Pkg.Path != "" {
				// 目录不算包：清单可能被手改成目录，交给消费者打开只会得到
				// "是目录不是 .tzc" 的次生错误，不如按"没有包"处理。
				if st, e := os.Stat(mf.Pkg.Path); e == nil && !st.IsDir() {
					return mf.Pkg.Path, nil
				}
			}
		}
		return "", nil
	}
	return path, nil
}

// PkgPoint 从工作区的包里取某个点位的声明正文。
//
// 用途：apply 之后回答"我的改动进包没有"。此时 `.4gl` 还是服务器 build 的旧产物，
// 只有包里的 `.tap` 才是刚落盘的真相（实测：试用者拿 at 复核，看到的是空点）。
func PkgPoint(dir, name string) (*devdata.TapPointView, string, error) {
	t, p, err := Load(dir)
	if err != nil {
		return nil, "", err
	}
	pkgPath := p.Pkg(t.Prog, identityOf(t))
	if _, serr := os.Stat(pkgPath); serr != nil {
		// 退一步用视图清单里记的那份（老工作区可能把包放在别处）。
		if ws, werr := store.Open(p.Ws()); werr == nil {
			if mf, merr := ws.Manifest(); merr == nil && mf.Pkg.Path != "" {
				pkgPath = mf.Pkg.Path
			}
		}
	}
	pkg, err := pkgfile.Open(pkgPath, pkgfile.OpenOptions{AcceptVer: "1.0"})
	if err != nil {
		return nil, pkgPath, err
	}
	tap := pkg.Tap()
	if tap == nil {
		return nil, pkgPath, fmt.Errorf("%w: 包里没有 .tap", ErrPackageMismatch)
	}
	v, err := devdata.TapPointBytes(tap.Data, name)
	if err != nil {
		return nil, pkgPath, err
	}
	return v, pkgPath, nil
}

// identityOf 取身份（task.json 里记的；老工作区没记就从包名兜底）。
func identityOf(t *Task) string {
	if strings.TrimSpace(t.Identity) != "" {
		return t.Identity
	}
	return "c"
}

// countSliceTemplates 数 `<mdl>/slice/code_*.template`。
func countSliceTemplates(mdlDir string) (int, bool) {
	entries, err := os.ReadDir(filepath.Join(mdlDir, "slice"))
	if err != nil {
		return 0, false
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), "code_") && strings.HasSuffix(e.Name(), ".template") {
			n++
		}
	}
	return n, true
}

// Describe 是一句人读的定位（CLI 标题用）。
func (s *Status) Describe() string {
	name := s.Prog
	if s.Task != "" {
		name += " · " + s.Task
	}
	return fmt.Sprintf("%s（%s）", name, s.Env)
}
