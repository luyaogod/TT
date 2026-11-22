package devdata

// 把已拉下来的材料**在本地**重新打成一个 .tzc —— 复刻设计器「下载程式包」的产物。
//
// 为什么不走服务器：真的"下载 tzc"（adzp050_download_code_pack）只是
// `sadzp050_zip_compress_to_code_zip` + `sadzp050_zip_download_code_zip`，
// 既不重产也不读数据库，包内成员就是 dzx 里的 tap/tgl 加 4gl 与一个恒为 "1.0" 的 ver。
// 而 tap/tgl/4gl 已经由 `tt gen fetch` 拉到本地了（且带 sha256），所以这一步纯本地：
// 服务器零写操作、零风险，任意程序任意时刻都能造出测试夹具。
//
// 包形态由 pkgfile.WriteZip 保证（flags 不置 bit 3、CRC 内联），与真包同构。

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/internal/dev/pkgfile"
)

// 真包里 ver 条目的内容恒为 "1.0\n"（5 个真包实测 crc 都是 6493f4df）。
const defaultPkgVer = "1.0"

// PackRequest 是打包请求。
type PackRequest struct {
	CfgPath  string
	Env      string
	Prog     string
	Identity string // c（客制）/ s（标准）；空 = 从 .tap 根属性推断，推断不出按 c
	Out      string // 目标文件或目录；空 = <环境>/pack/<程序>(<身份>).tzc
	Ver      string // ver 条目内容，空 = "1.0"
	Force    bool   // 目标已存在时覆盖
	// Dir 材料目录（空 = <环境>/prog/<程序>）。工作区把它指到 <工作区>/materials。
	Dir string
}

// PackEntry 是包内一个成员。
type PackEntry struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Size   int    `json:"size"`
	SHA256 string `json:"sha256"`
}

// PackResult 是打包结果。
type PackResult struct {
	Env      string      `json:"env"`
	Prog     string      `json:"prog"`
	Identity string      `json:"identity"`
	Path     string      `json:"path"`
	Bytes    int64       `json:"bytes"`
	SHA256   string      `json:"sha256"`
	Entries  []PackEntry `json:"entries"`
	Warnings []string    `json:"warnings"`
	Elapsed  string      `json:"elapsed"`
}

// 包内必须有的三个成员（缺一个就不成包）。
var packRoles = []struct {
	role string
	ext  string
}{
	{"4gl", ".4gl"},
	{"tap", ".tap"},
	{"tgl", ".tgl"},
}

var reIdentity = regexp.MustCompile(`identity="([a-z])"`)

// Pack 把某程序已拉取的材料打成一个 .tzc。
func Pack(req PackRequest) (*PackResult, error) {
	start := time.Now()
	if err := ValidProgName(req.Prog); err != nil {
		return nil, err
	}
	if err := validIdentity(req.Identity); err != nil {
		return nil, err
	}
	progDir, err := PackMaterialDir(req)
	if err != nil {
		return nil, err
	}

	mf, err := readPackManifest(progDir, req.Prog)
	if err != nil {
		return nil, err
	}

	res := &PackResult{Env: req.Env, Prog: req.Prog}

	// 三个成员：按 role 找 manifest 里的相对路径，再核对 sha256（材料被本地改过要警告）。
	var members []packMember
	for _, want := range packRoles {
		rel, fe := relOfRole(mf, want.role)
		if rel == "" {
			return nil, fmt.Errorf("%w: %s 缺 %s —— 未被设计器接管过的标准程序在服务器上本来就没有 .tap/.tsd；"+
				"若该程序确实有，重新 tt gen fetch %s --force",
				ErrMaterialIncomplete, req.Prog, want.ext, req.Prog)
		}
		data, rerr := os.ReadFile(filepath.Join(progDir, rel))
		if rerr != nil {
			return nil, fmt.Errorf("读材料失败 (%s): %w", rel, rerr)
		}
		if fe != nil && fe.SHA256 != "" {
			if got := pkgfile.Sha256Hex(data); got != fe.SHA256 {
				res.Warnings = append(res.Warnings,
					fmt.Sprintf("%s 与 manifest 记的 sha256 不一致（本地被改过？）：%s → %s", rel, fe.SHA256[:12], got[:12]))
			}
		}
		members = append(members, packMember{role: want.role, rel: rel, data: data})
	}

	identity := strings.TrimSpace(req.Identity)
	if identity == "" {
		identity = detectIdentity(members)
		if identity == "" {
			identity = "c"
			res.Warnings = append(res.Warnings, "未能从 .tap 根属性读出 identity，按客制 (c) 命名；要改请用 --identity")
		}
	}
	res.Identity = identity

	ver := strings.TrimSpace(req.Ver)
	if ver == "" {
		ver = defaultPkgVer
	}
	if _, err := pkgfile.ParseVersion([]byte(ver)); err != nil {
		return nil, fmt.Errorf("%w: --ver %q 不是 Major.Minor 形式", ErrInvalidOpt, ver)
	}

	name := PackFileName(req.Prog, identity)
	outPath := strings.TrimSpace(req.Out)
	if outPath == "" {
		dir, derr := PackDir(req.CfgPath, req.Env)
		if derr != nil {
			return nil, derr
		}
		outPath = filepath.Join(dir, name)
	} else if isDirLike(outPath) {
		outPath = filepath.Join(outPath, name)
	}
	if fi, serr := os.Stat(outPath); serr == nil && fi.IsDir() {
		outPath = filepath.Join(outPath, name)
	}

	// 组装条目：真包顺序是 4gl / tap / tgl / ver（按扩展名字典序），ver 是 store。
	now := time.Now()
	entries := make([]pkgfile.WriteEntry, 0, len(members)+1)
	for _, m := range members {
		entries = append(entries, pkgfile.WriteEntry{
			Name:    req.Prog + fileExt(m.role),
			Data:    m.data,
			Mode:    0o664,
			ModTime: now,
		})
	}
	entries = append(entries, pkgfile.WriteEntry{
		Name:    "ver",
		Data:    []byte(ver + "\n"),
		Store:   true,
		Mode:    0o775,
		ModTime: now,
	})

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败 (%s): %w", filepath.Dir(outPath), err)
	}
	if fi, serr := os.Stat(outPath); serr == nil && !req.Force {
		return nil, fmt.Errorf("%w: %s（加 --force 覆盖）", ErrExists, outPath)
	} else if serr == nil && fi.IsDir() {
		return nil, fmt.Errorf("目标是个目录，不是包文件：%s", outPath)
	}

	tmp, err := os.CreateTemp(filepath.Dir(outPath), filepath.Base(outPath)+tmpSuffixPrefix+"*")
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败 (%s): %w", filepath.Dir(outPath), err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }
	if err := pkgfile.WriteZip(tmp, entries); err != nil {
		_ = tmp.Close()
		cleanup()
		return nil, fmt.Errorf("写包失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return nil, fmt.Errorf("关闭临时文件失败: %w", err)
	}
	if req.Force {
		_ = os.Remove(outPath) // Windows 上 rename 不覆盖已存在文件
	}
	if err := os.Rename(tmpName, outPath); err != nil {
		cleanup()
		return nil, fmt.Errorf("改名失败 (%s → %s): %w", tmpName, outPath, err)
	}

	// 自检：写出来的东西必须能被自己的读取器当成合法的 Code 包打开。
	pkg, err := pkgfile.Open(outPath, pkgfile.OpenOptions{AcceptVer: defaultPkgVer})
	if err != nil {
		return nil, fmt.Errorf("产出的包自检失败（%s）: %w", outPath, err)
	}
	if pkg.Kind != pkgfile.KindCode {
		return nil, fmt.Errorf("产出的包类型不是 Code（%s）", pkg.Kind)
	}

	raw, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("读回产出物失败: %w", err)
	}
	res.Path = outPath
	res.Bytes = int64(len(raw))
	res.SHA256 = pkgfile.Sha256Hex(raw)
	for i, e := range pkg.Entries {
		role := "ver"
		if i < len(members) {
			role = members[i].role
		}
		res.Entries = append(res.Entries, PackEntry{
			Name: e.Name, Role: role, Size: e.Size(), SHA256: pkgfile.Sha256Hex(e.Data),
		})
	}
	res.Elapsed = time.Since(start).Round(time.Millisecond).String()
	return res, nil
}

// PackMaterialDir 打包读材料的目录：显式 Dir 优先，否则环境的程序目录。
func PackMaterialDir(req PackRequest) (string, error) {
	if d := strings.TrimSpace(req.Dir); d != "" {
		return filepath.Abs(d)
	}
	return ProgDir(req.CfgPath, req.Env, req.Prog)
}

// PackFileName 包文件名：`<程序>(<身份>).tzc`（设计器下载的命名形态）。
func PackFileName(prog, identity string) string {
	return fmt.Sprintf("%s(%s).tzc", prog, identity)
}

// PackDir 某环境的包目录：<环境>/pack。
func PackDir(cfgPath, env string) (string, error) {
	dir, err := EnvDir(cfgPath, env)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, config.DevDataSubdirPack), nil
}

// validIdentity 身份只能是 c（客制）或 s（标准）—— 设计器的包名只有这两种。
func validIdentity(s string) error {
	switch strings.TrimSpace(s) {
	case "", "c", "s":
		return nil
	}
	return fmt.Errorf("%w: --identity 只能是 c（客制）或 s（标准），给的是 %q", ErrInvalidOpt, s)
}

// readPackManifest 读程序材料的 manifest；没有材料时给一条能直接照做的提示。
func readPackManifest(progDir, prog string) (*Manifest, error) {
	b, err := os.ReadFile(filepath.Join(progDir, "manifest.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: 本地没有 %s 的材料（先 tt gen fetch %s）", ErrProgNotFound, prog, prog)
		}
		return nil, fmt.Errorf("读 manifest 失败 (%s): %w", progDir, err)
	}
	var mf Manifest
	if err := json.Unmarshal(b, &mf); err != nil {
		return nil, fmt.Errorf("manifest 不是合法 JSON (%s): %w", progDir, err)
	}
	if mf.Kind != "fetch" {
		return nil, fmt.Errorf("%w: %s 的材料不是按程序拉取的（manifest.kind=%q）",
			ErrProgNotFound, prog, mf.Kind)
	}
	return &mf, nil
}

// relOfRole 按 role 找本地相对路径与登记信息。
func relOfRole(mf *Manifest, role string) (string, *FileEntry) {
	for i := range mf.Files {
		if mf.Files[i].Role == role && mf.Files[i].Rel != "" {
			return mf.Files[i].Rel, &mf.Files[i]
		}
	}
	return "", nil
}

// packMember 是一个待入包的材料成员（role 决定包内扩展名）。
type packMember struct {
	role, rel string
	data      []byte
}

// detectIdentity 从 .tap 根部读 identity（`<add_points … identity="c" …>`）。
func detectIdentity(members []packMember) string {
	for _, m := range members {
		if m.role != "tap" {
			continue
		}
		head := m.data
		if len(head) > 4096 {
			head = head[:4096]
		}
		if mm := reIdentity.FindSubmatch(head); mm != nil {
			return string(mm[1])
		}
	}
	return ""
}

// fileExt 由 role 得到包内扩展名。
func fileExt(role string) string {
	switch role {
	case "4gl":
		return ".4gl"
	case "tap":
		return ".tap"
	case "tgl":
		return ".tgl"
	}
	return "." + role
}

// isDirLike 判断 --out 是不是"当目录用"（末尾带分隔符）。
func isDirLike(p string) bool {
	return strings.HasSuffix(p, string(os.PathSeparator)) || strings.HasSuffix(p, "/")
}
