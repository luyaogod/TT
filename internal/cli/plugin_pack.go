package cli

// tt plugin pack：把一个插件目录打成 zip，并把清单里的 tt 下限写成**打包时**
// 的 SDK 版本 —— 插件用到了哪个版本的 Host API，就要求哪个版本的 TT，避免
// "装到旧 TT 上跑一半才发现 RPC 不存在"。

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/internal/plugin"
)

func newPluginPackCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "pack <目录>",
		Short: "打包插件（自动写回 tt 下限）",
		Long: `把一个插件目录打成 zip，并把 plugin.json 里的 tt 下限改写成打包时的 SDK 版本。

SDK 版本取自插件目录的 go.mod：优先读 replace 指向的本地 TT 的 VERSION，
否则读 require 的 github.com/luyaogod/TT 版本号。上限沿用清单里已有的，
缺省 <1.0.0。

包里会排除 .git / dist / .tmp / node_modules / *.zip。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, rng, id, ver, err := packPlugin(args[0], out)
			if err != nil {
				return err
			}
			b, _ := json.Marshal(map[string]any{
				"ok": true, "id": id, "version": ver, "tt": rng, "path": path,
			})
			fmt.Println(string(b))
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "输出 zip 路径（缺省 <目录>/dist/<id>-<version>.zip）")
	return cmd
}

// packPlugin 打包并返回 (zip 路径, tt 范围, id, version)。
func packPlugin(dir, out string) (string, string, string, string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", "", "", "", err
	}
	raw, err := os.ReadFile(filepath.Join(abs, "plugin.json"))
	if err != nil {
		return "", "", "", "", fmt.Errorf("读 plugin.json 失败: %w", err)
	}
	var man plugin.Manifest
	if err := json.Unmarshal(raw, &man); err != nil {
		return "", "", "", "", fmt.Errorf("plugin.json 不合法: %w", err)
	}
	if err := plugin.ValidateManifest(&man, abs, plugin.ValidateOptions{HostAPIVersion: connsvc.HostAPIVersion}); err != nil {
		return "", "", "", "", err
	}

	sdkVer := detectSDKVersion(abs)
	upper := "<1.0.0"
	if m := regexp.MustCompile(`<\s*([0-9]+\.[0-9]+\.[0-9]+)`).FindStringSubmatch(man.TT); len(m) == 2 {
		upper = "<" + m[1]
	}
	rng := man.TT
	if sdkVer != "" {
		rng = ">=" + sdkVer + " " + upper
	}

	// 改写清单（保留未知字段）。
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", "", "", "", err
	}
	obj["tt"] = rng
	newRaw, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", "", "", "", err
	}

	stage, err := os.MkdirTemp("", "tt-pack-*")
	if err != nil {
		return "", "", "", "", err
	}
	defer os.RemoveAll(stage)
	if err := copyTreeSkipping(abs, stage); err != nil {
		return "", "", "", "", err
	}
	if err := os.WriteFile(filepath.Join(stage, "plugin.json"), append(newRaw, '\n'), 0o644); err != nil {
		return "", "", "", "", err
	}

	if out == "" {
		out = filepath.Join(abs, "dist", man.ID+"-"+man.Version+".zip")
	}
	if err := zipDir(stage, out); err != nil {
		return "", "", "", "", err
	}
	return out, rng, man.ID, man.Version, nil
}

// detectSDKVersion 从 go.mod 推断插件编译时用的 TT SDK 版本。
func detectSDKVersion(dir string) string {
	b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return ""
	}
	s := string(b)
	// 本地 replace 优先：读被指向仓库的 VERSION。
	if m := regexp.MustCompile(`(?m)^replace\s+github\.com/luyaogod/TT\s+=>\s+(\S+)`).FindStringSubmatch(s); len(m) == 2 {
		p := m[1]
		if !filepath.IsAbs(p) {
			p = filepath.Join(dir, p)
		}
		if v, err := os.ReadFile(filepath.Join(p, "VERSION")); err == nil {
			if ver := strings.TrimSpace(string(v)); plugin.ValidateSemver(ver) == nil {
				return ver
			}
		}
	}
	if m := regexp.MustCompile(`github\.com/luyaogod/TT\s+v([0-9]+\.[0-9]+\.[0-9]+)`).FindStringSubmatch(s); len(m) == 2 {
		return m[1]
	}
	return ""
}

func copyTreeSkipping(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil || rel == "." {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			switch name {
			case ".git", "dist", ".tmp", "node_modules":
				return fs.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		if strings.HasSuffix(strings.ToLower(name), ".zip") || strings.HasSuffix(name, ".log") {
			return nil
		}
		return copyFile(p, filepath.Join(dst, rel))
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func zipDir(src, out string) error {
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)
	walkErr := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		w, err := zw.Create(filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()
		_, err = io.Copy(w, in)
		return err
	})
	if walkErr != nil {
		zw.Close()
		f.Close()
		return walkErr
	}
	if err := zw.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
