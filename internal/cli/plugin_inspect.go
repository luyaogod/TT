package cli

// tt plugin inspect：只读地看清一个插件包或已安装插件 —— 清单、贡献、文件、
// 体积、问题清单。与 doctor 的区别：doctor 判"能不能跑"，inspect 回答"这是什么"。

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/internal/plugin"
)

func newPluginInspectCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "inspect <目录|zip|插件 id>",
		Short: "查看插件包或已安装插件的清单/文件/体积",
		Long: `查看插件包（目录或 zip）的清单与文件，或查看已安装插件的注册信息、路径与体积。

包形态只读：不安装、不执行、不改注册表。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := args[0]
			var report map[string]any
			if st, err := os.Stat(arg); err == nil {
				r, perr := inspectPackage(arg, st.IsDir())
				if perr != nil {
					return perr
				}
				report = r
			} else {
				m, merr := pluginManager()
				if merr != nil {
					return merr
				}
				info, ierr := m.Info(arg)
				if ierr != nil {
					return pluginErr(ierr)
				}
				report = map[string]any{
					"ok": true, "kind": "installed",
					"id": info.ID, "version": info.Version, "enabled": info.Enabled,
					"namespace": info.Namespace, "route": info.Route,
					"manifest": info.Manifest, "versions": info.Versions,
					"paths": info.Paths, "sizeBytes": dirBytes(info.Paths.VersionDir),
				}
			}
			if asJSON {
				b, _ := json.MarshalIndent(report, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			printInspect(report)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "输出 JSON")
	return cmd
}

// inspectPackage 读目录或 zip（不安装）。zip 会先安全解到临时目录再按目录检查。
func inspectPackage(p string, dir bool) (map[string]any, error) {
	orig := p
	if !dir {
		tmp, err := os.MkdirTemp("", "tt-inspect-*")
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(tmp)
		if err := plugin.ExtractZip(p, tmp, plugin.DefaultZipLimits); err != nil {
			return nil, fmt.Errorf("解包失败: %w", err)
		}
		p, dir = tmp, true
	}

	raw, err := os.ReadFile(filepath.Join(p, "plugin.json"))
	if err != nil {
		return nil, fmt.Errorf("包里没有 plugin.json: %w", err)
	}
	files := map[string]int64{}
	var size int64
	if err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(p, path)
		if strings.HasPrefix(filepath.ToSlash(rel), ".git/") {
			return nil
		}
		if i, err := d.Info(); err == nil {
			files[filepath.ToSlash(rel)] = i.Size()
			size += i.Size()
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var man plugin.Manifest
	var problems []string
	if err := json.Unmarshal(raw, &man); err != nil {
		return nil, fmt.Errorf("plugin.json 不合法: %w", err)
	}
	if err := plugin.ValidateManifest(&man, p, plugin.ValidateOptions{HostAPIVersion: connsvc.HostAPIVersion}); err != nil {
		problems = append(problems, strings.Split(err.Error(), "\n")...)
	}

	list := make([]string, 0, len(files))
	for k := range files {
		list = append(list, k)
	}
	sort.Strings(list)
	fileList := make([]map[string]any, 0, len(list))
	for _, k := range list {
		fileList = append(fileList, map[string]any{"path": k, "size": files[k]})
	}
	return map[string]any{
		"ok": len(problems) == 0, "kind": "package", "path": orig,
		"manifest": man, "files": fileList, "totalBytes": size,
		"problems": problems,
	}, nil
}
func dirBytes(dir string) int64 {
	var total int64
	_ = filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if i, err := d.Info(); err == nil {
			total += i.Size()
		}
		return nil
	})
	return total
}

func manifestOf(v any) (plugin.Manifest, bool) {
	switch m := v.(type) {
	case plugin.Manifest:
		return m, true
	case *plugin.Manifest:
		if m != nil {
			return *m, true
		}
	}
	return plugin.Manifest{}, false
}

func printInspect(r map[string]any) {
	kind, _ := r["kind"].(string)
	if kind == "package" {
		fmt.Printf("插件包: %v\n", r["path"])
	} else {
		fmt.Printf("已安装: %v  v%v  enabled=%v\n", r["id"], r["version"], r["enabled"])
	}
	if m, ok := manifestOf(r["manifest"]); ok {
		fmt.Printf("  id=%s name=%s version=%s tt=%s hostApi=%s\n", m.ID, m.Name, m.Version, m.TT, m.HostAPI)
		var kinds []string
		if m.Contributions.CLI != nil && m.Contributions.CLI.Namespace != "" {
			kinds = append(kinds, "cli:"+m.Contributions.CLI.Namespace)
		}
		if m.Contributions.Service != nil && m.Contributions.Service.Entry != "" {
			kinds = append(kinds, "service")
		}
		if len(m.Contributions.UI) > 0 {
			kinds = append(kinds, fmt.Sprintf("ui:%d", len(m.Contributions.UI)))
		}
		if len(kinds) > 0 {
			fmt.Println("  贡献:", strings.Join(kinds, " "))
		}
	}
	if v, ok := r["totalBytes"]; ok {
		fmt.Printf("  体积: %.1f KB\n", float64(v.(int64))/1024)
	}
	if v, ok := r["sizeBytes"]; ok {
		fmt.Printf("  体积: %.1f KB\n", float64(v.(int64))/1024)
	}
	if p, ok := r["problems"].([]string); ok && len(p) > 0 {
		fmt.Println("  问题:")
		for _, s := range p {
			fmt.Println("    -", s)
		}
	}
}
