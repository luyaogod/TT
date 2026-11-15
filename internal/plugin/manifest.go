package plugin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ManifestFileName 插件包根部的清单文件名。首版只用严格 JSON —— 不引入 YAML 依赖。
const ManifestFileName = "plugin.json"

// ManifestSchemaVersion 当前清单结构版本。更高主版本一律拒绝（不是"尽力解析"）。
const ManifestSchemaVersion = 1

// Manifest 插件清单。字段与 .agents/docs/plugin-system/ 的契约文档一致。
type Manifest struct {
	SchemaVersion int           `json:"schemaVersion"`
	ID            string        `json:"id"`
	Version       string        `json:"version"`
	HostAPI       string        `json:"hostApi"`
	Name          string        `json:"name,omitempty"`
	Description   string        `json:"description,omitempty"`
	Contributions Contributions `json:"contributions"`
	Permissions   []string      `json:"permissions,omitempty"`
}

// Contributions 一个插件包可以贡献的能力。四类互相独立：可以只声明其中一类。
type Contributions struct {
	CLI      *CLIContribution      `json:"cli,omitempty"`
	Service  *ServiceContribution  `json:"service,omitempty"`
	Frontend *FrontendContribution `json:"frontend,omitempty"`
	Skills   []string              `json:"skills,omitempty"`
}

// CLIContribution 第三方 CLI：tt <namespace> … 转发到 entry。
type CLIContribution struct {
	Namespace string `json:"namespace"`
	// Entry 包内相对路径，允许 {os}/{arch} 占位（如 bin/{os}-{arch}/tool.exe）。
	Entry   string `json:"entry"`
	Summary string `json:"summary,omitempty"`
}

// ServiceContribution 可选常驻服务。
type ServiceContribution struct {
	Entry string `json:"entry"`
	// Args 启动服务时传给入口的参数（缺省为空）。
	//
	// 为什么需要：插件常把 CLI 与常驻服务装在同一个可执行文件里，用子命令区分
	// （`myplugin service`）。没有这个字段就只能靠环境变量约定，而那是隐式契约。
	Args []string `json:"args,omitempty"`
	// HealthPath 健康检查路径，缺省 /healthz。
	HealthPath string `json:"healthPath,omitempty"`
}

// FrontendContribution 第三方前端页面（独立 iframe + 消息桥）。
type FrontendContribution struct {
	Route string `json:"route"`
	Title string `json:"title"`
	Entry string `json:"entry"`
}

// HealthPathOrDefault 生效的健康检查路径。
func (s *ServiceContribution) HealthPathOrDefault() string {
	if s == nil {
		return ""
	}
	if p := strings.TrimSpace(s.HealthPath); p != "" {
		return ensureLeadingSlash(p)
	}
	return "/healthz"
}

// ensureLeadingSlash 补前导斜杠。
func ensureLeadingSlash(p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	return "/" + p
}

// LoadManifest 从目录读清单（严格解码：未知字段即错误）。
func LoadManifest(dir string) (*Manifest, error) {
	path := filepath.Join(dir, ManifestFileName)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 %s 失败: %w", ManifestFileName, err)
	}
	m, err := DecodeManifest(b)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// DecodeManifest 严格解码清单 JSON。
//
// 为什么 DisallowUnknownFields：清单里写错的键（比如把 entry 写成 entryPath）如果被静默
// 忽略，插件会以"缺字段"的姿态安装成功、运行时才炸；宁可安装时就拒绝并指出多余字段。
func DecodeManifest(b []byte) (*Manifest, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var m Manifest
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("清单不是合法 JSON 或含未知字段: %w", err)
	}
	// 只允许一个 JSON 值：多余内容说明文件被拼接了。
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, errors.New("清单文件在第一个 JSON 对象之后还有多余内容")
	}
	return &m, nil
}

// expandEntry 把 {os}/{arch} 占位换成当前平台，并校验是包内相对路径。
//
// 占位只支持这两个：插件的入口文件按平台分包（bin/windows-amd64/…），
// 而这里能确定的平台事实只有 GOOS/GOARCH。没有占位时就原样当相对路径。
func expandEntry(entry string) (string, error) {
	e := strings.TrimSpace(entry)
	if e == "" {
		return "", errors.New("入口路径为空")
	}
	e = strings.ReplaceAll(e, "{os}", runtime.GOOS)
	e = strings.ReplaceAll(e, "{arch}", runtime.GOARCH)
	return normalizeRelPath(e)
}

// normalizeRelPath 校验并规范化包内相对路径。
//
// 规则：不接受绝对路径、盘符、UNC、`..` 段、NUL；分隔符统一成 `/`。
// 包内路径必须是"相对且被包根关住"的，否则一次解包就能写到插件目录之外。
func normalizeRelPath(p string) (string, error) {
	s := strings.TrimSpace(p)
	if s == "" {
		return "", errors.New("路径为空")
	}
	if strings.ContainsRune(s, 0) {
		return "", errors.New("路径含 NUL")
	}
	// 统一分隔符：zip 里两种都见过。
	s = strings.ReplaceAll(s, `\`, "/")
	if strings.HasPrefix(s, "/") {
		return "", fmt.Errorf("不允许绝对路径: %q", p)
	}
	// 盘符（C:）与 UNC（//host）在上面两条之后已经不可能通过。
	if len(s) >= 2 && s[1] == ':' {
		return "", fmt.Errorf("不允许盘符路径: %q", p)
	}
	parts := strings.Split(s, "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		switch part {
		case "", ".":
			continue
		case "..":
			return "", fmt.Errorf("不允许路径穿越: %q", p)
		default:
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return "", fmt.Errorf("路径规范化后为空: %q", p)
	}
	return strings.Join(out, "/"), nil
}

// ResolveEntry 返回一个贡献入口在版本目录内的**规范化相对路径**。
func (m *Manifest) ResolveCLIEntry() (string, error) {
	if m.Contributions.CLI == nil {
		return "", errors.New("没有 cli 贡献")
	}
	return expandEntry(m.Contributions.CLI.Entry)
}

// ResolveServiceEntry 服务入口的相对路径。
func (m *Manifest) ResolveServiceEntry() (string, error) {
	if m.Contributions.Service == nil {
		return "", errors.New("没有 service 贡献")
	}
	return expandEntry(m.Contributions.Service.Entry)
}

// ResolveFrontendEntry 前端入口的相对路径。
func (m *Manifest) ResolveFrontendEntry() (string, error) {
	if m.Contributions.Frontend == nil {
		return "", errors.New("没有 frontend 贡献")
	}
	return expandEntry(m.Contributions.Frontend.Entry)
}

// SkillDirs 技能目录（相对路径，已校验）。顺序即声明顺序。
func (m *Manifest) SkillDirs() ([]string, error) {
	out := make([]string, 0, len(m.Contributions.Skills))
	for _, s := range m.Contributions.Skills {
		p, err := normalizeRelPath(s)
		if err != nil {
			return nil, fmt.Errorf("技能目录 %q: %w", s, err)
		}
		out = append(out, p)
	}
	return out, nil
}

// HasContribution 是否有任何一类贡献（四类全无的包对 TT 没有意义）。
func (m *Manifest) HasContribution() bool {
	c := m.Contributions
	return c.CLI != nil || c.Service != nil || c.Frontend != nil || len(c.Skills) > 0
}
