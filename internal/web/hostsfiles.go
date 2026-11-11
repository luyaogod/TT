package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"tt/internal/config"
)

// ── 站点的导出 / 导入 ──
//
// 这是"把一台环境搬到另一台机器"的落盘格式:导出写一个 JSON 文件到用户选的目录,
// 导入读一个目录里的这类文件。**两个端点都不改配置** —— 导入只把候选交回前端,
// 落盘仍走唯一的 PUT /api/hosts(那条路上有改名迁移、校验、保留未知字段等一整套)。
//
// 文件为什么要包一层 kind/schemaVersion,而不是直接吐一个 NamedSsh:
// 这种文件会被用户手工搬来搬去、落到任意目录里,导入端需要一句话判断"这是不是 TT 的
// 站点文件",而不是拿形状去猜;版本号则给以后的字段演进留门。
const (
	siteFileKind    = "tt-site"
	siteFileVersion = 1
	// siteFileMaxBytes 单个站点文件的上限。一份站点配置正常只有几 KB,
	// 设 1 MiB 是防"有人把整个 config.json 甚至别的东西丢进这个目录"。
	siteFileMaxBytes = 1 << 20
	// siteImportMaxFiles 一次最多扫多少个 *.json。目录里可能有成百上千个无关文件。
	siteImportMaxFiles = 200
)

type siteFile struct {
	Kind          string          `json:"kind"`
	SchemaVersion int             `json:"schemaVersion"`
	ExportedAt    string          `json:"exportedAt,omitempty"`
	Site          config.NamedSsh `json:"site"`
}

type siteExportReq struct {
	Name string `json:"name"`
	Dir  string `json:"dir"`
}

// hHostsExport 把一个站点导出成 <dir>/<站点名>.json。
//
// 数据源是**磁盘上的配置**:设置页的改动是自动保存的,若导出时还在保存窗口里,
// 前端会把按钮置灰(与镜像目录那几行同一套 isDirty 判断),所以这里不读表单。
//
// 注意:文件里含 SSH/数据库口令(与 config.json 同等级),因此落盘用 0600。
func (s *Server) hHostsExport(w http.ResponseWriter, r *http.Request) {
	var req siteExportReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("没有选中环境"))
		return
	}
	dir, err := requireDir(req.Dir, "导出目录")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}

	cfgPath, err := s.configPath(false)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	root, err := config.Load(cfgPath)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	var site *config.NamedSsh
	for i := range root.Hosts.SSHs {
		if root.Hosts.SSHs[i].Name == name {
			site = &root.Hosts.SSHs[i]
			break
		}
	}
	if site == nil {
		writeErr(w, http.StatusNotFound, fmt.Errorf("环境不存在: %s", name))
		return
	}

	body := siteFile{
		Kind:          siteFileKind,
		SchemaVersion: siteFileVersion,
		ExportedAt:    time.Now().Format(time.RFC3339),
		Site:          *site,
	}
	b, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	target := filepath.Join(dir, siteFileName(name))
	// 同名导出到同一目录 = 刷新那份文件,所以直接覆盖(前端会提示落地路径)。
	if err := os.WriteFile(target, append(b, '\n'), 0o600); err != nil {
		writeErr(w, http.StatusInternalServerError, fmt.Errorf("写文件失败: %w", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "path": target, "bytes": len(b)})
}

type siteImportReq struct {
	Dir string `json:"dir"`
}

// siteImportItem 一个候选站点。Error 非空 = 这个文件读不成一个站点(照样回给前端,
// 由它列出来告诉用户"哪个文件为什么没进来",而不是静默跳过)。
type siteImportItem struct {
	File  string           `json:"file"`
	Site  *config.NamedSsh `json:"site,omitempty"`
	Error string           `json:"error,omitempty"`
}

// hHostsImport 读目录里的站点 JSON,返回候选清单(**不写配置**)。
//
// 接受两种形状:TT 导出的 {kind,schemaVersion,site:{...}},以及用户手写/从别处
// 拷来的裸站点对象(只要有 name/host/user)。后者是刻意留的口子 —— 用户想让 AI
// 照着格式造一个站点文件时,不必先造外壳。
func (s *Server) hHostsImport(w http.ResponseWriter, r *http.Request) {
	var req siteImportReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	dir, err := requireDir(req.Dir, "导入目录")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("无法读取目录: %v", err))
		return
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	truncated := false
	if len(names) > siteImportMaxFiles {
		names = names[:siteImportMaxFiles]
		truncated = true
	}

	items := make([]siteImportItem, 0, len(names))
	for _, n := range names {
		p := filepath.Join(dir, n)
		item := siteImportItem{File: p}
		site, err := readSiteFile(p)
		if err != nil {
			item.Error = err.Error()
		} else {
			item.Site = site
		}
		items = append(items, item)
	}
	// 空的候选也要给 [] 而不是 null:前端会直接 .map / .length。
	if items == nil {
		items = []siteImportItem{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "dir": dir, "sites": items, "truncated": truncated,
	})
}

// readSiteFile 读一个文件并归一成一个可直接放进配置的站点。
//
// 归一规则与设置页的表单同一口径:端口缺省 22、名字/地址/账号去空白后必须非空。
// 不合格时返回**给人看的原因**,由调用方落到 item.Error。
func readSiteFile(path string) (*config.NamedSsh, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("读不到文件: %v", err)
	}
	if st.Size() > siteFileMaxBytes {
		return nil, fmt.Errorf("文件太大(%d 字节),站点配置不该这么大", st.Size())
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读不到文件: %v", err)
	}

	var wrap siteFile
	if err := json.Unmarshal(b, &wrap); err != nil {
		return nil, fmt.Errorf("不是合法 JSON: %v", err)
	}
	site := wrap.Site
	if wrap.Kind == "" && site.Name == "" {
		// 不是外壳 → 试裸站点对象(用户手写的那种)。
		var bare config.NamedSsh
		if err := json.Unmarshal(b, &bare); err != nil {
			return nil, fmt.Errorf("不是合法 JSON: %v", err)
		}
		site = bare
	}
	if wrap.Kind != "" && wrap.Kind != siteFileKind {
		return nil, fmt.Errorf("不是 TT 站点文件(kind=%q)", wrap.Kind)
	}

	site.Name = strings.TrimSpace(site.Name)
	site.Host = strings.TrimSpace(site.Host)
	site.User = strings.TrimSpace(site.User)
	site.Zone = strings.TrimSpace(site.Zone)
	site.Workspace = strings.TrimSpace(site.Workspace)
	site.MirrorDir = strings.TrimSpace(site.MirrorDir)
	site.DictDir = strings.TrimSpace(site.DictDir)
	if site.Name == "" {
		return nil, fmt.Errorf("缺少环境名(name)")
	}
	if site.Host == "" {
		return nil, fmt.Errorf("缺少服务器地址(host)")
	}
	if site.User == "" {
		return nil, fmt.Errorf("缺少登录账号(user)")
	}
	if site.Port <= 0 || site.Port > 65535 {
		site.Port = 22
	}
	return &site, nil
}

// requireDir 校验一个"用户选出来的目录":必须非空、存在、且真的是目录。
func requireDir(raw, what string) (string, error) {
	dir := strings.TrimSpace(raw)
	if dir == "" {
		return "", fmt.Errorf("先选一个%s", what)
	}
	st, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("%s不存在或无法访问: %s", what, dir)
	}
	if !st.IsDir() {
		return "", fmt.Errorf("%s不是目录: %s", what, dir)
	}
	return dir, nil
}

// siteFileName 生成导出文件名:<站点名>.json,把路径分隔符与 Windows 非法字符换成 _。
//
// 站点名是用户输入的自由文本(可以叫 "109tst / 测试"),直接拼进路径不但会写错位置,
// 还可能越出所选目录 —— 所以这里只保留安全字符。
func siteFileName(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' ||
			r == '<' || r == '>' || r == '|' || unicode.IsControl(r):
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	out = strings.Trim(out, ".") // 别造出 "name." / 隐藏文件
	if out == "" {
		out = "site"
	}
	// 过长名字在 Windows 上会撞 MAX_PATH(目录本身也可能很深)。
	if r := []rune(out); len(r) > 64 {
		out = string(r[:64])
	}
	return out + ".json"
}
