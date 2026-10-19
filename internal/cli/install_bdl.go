package cli

// `tt install bdl` —— 把随包的 BDL 语言文档（BDL-Markdown.zip）解压到 bdldoc.dir。
//
// 文档本体很大（16 MB 的 markdown），不内嵌进二进制：与 skills/ 一样随包分发。
// 安装后（MSI 的自定义动作 / 便携包第一次 tt serve）解压一次；之后用 marker 判断
// 是否已是最新。目标目录优先用 config.json 里已有的 bdldoc.dir —— 用户可能已经
// 把文档放在别的盘上，不能覆盖他的选择。

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
)

const bdlZipName = "BDL-Markdown.zip"
const bdlMarkerName = ".tt-bdl.json"

// bdlZipSource 解析「exe 旁边的 BDL-Markdown.zip」；测试可替换。
var bdlZipSource = defaultBDLZipSource

func defaultBDLZipSource() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("取不到当前可执行文件路径: %w", err)
	}
	return filepath.Join(filepath.Dir(exe), bdlZipName), nil
}

type bdlMarker struct {
	Zip         string `json:"zip"`
	SHA256      string `json:"sha256"`
	ExtractedAt string `json:"extractedAt"`
	Files       int    `json:"files"`
}

func newInstallBdlCmd() *cobra.Command {
	var (
		dir   string
		force bool
	)
	cmd := &cobra.Command{
		Use:   "bdl [--dir <目录>] [--force]",
		Short: "解压随包的 BDL 语言文档到配置目录（bdldoc.dir）",
		Long: `把与 tt.exe 同目录的 BDL-Markdown.zip 解压到 BDL 文档目录。

目标目录按这个顺序：
  1. --dir <目录>（同时写回 config.json 的 bdldoc.dir）；
  2. config.json 里已有的 bdldoc.dir（不覆盖用户原来的选择）；
  3. 都没有 → <配置目录>/BDL-Markdown，并写回 bdldoc.dir。

同一个 zip 只解压一次（marker 记录 sha256）；升级后 zip 变了会自动刷新。
--force 强制重新解压。`,
		Example: `  tt install bdl
  tt install bdl --force
  tt install bdl --dir D:\T100\bdl-docs`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			src, err := bdlZipSource()
			if err != nil {
				return err
			}
			if _, err := os.Stat(src); err != nil {
				return fmt.Errorf("找不到 %s: %s\n（它应与 tt.exe 同目录；便携包/安装包自带）", bdlZipName, src)
			}
			cfgPath, err := common.ResolveConfig(true)
			if err != nil {
				return err
			}
			root, err := config.Load(cfgPath)
			if err != nil {
				return err
			}
			dst := strings.TrimSpace(dir)
			if dst == "" {
				dst = strings.TrimSpace(root.Bdldoc.Dir)
			}
			setCfg := dst == ""
			if dst == "" {
				dst = filepath.Join(filepath.Dir(cfgPath), "BDL-Markdown")
				setCfg = true
			}
			abs, err := filepath.Abs(dst)
			if err != nil {
				return fmt.Errorf("解析 BDL 文档目录失败: %w", err)
			}
			files, skipped, err := installBDLDocs(src, abs, force)
			if err != nil {
				return err
			}
			if setCfg || strings.TrimSpace(dir) != "" {
				if err := setBDLDocDir(cfgPath, abs); err != nil {
					return err
				}
			}
			if common.JSON {
				return common.PrintJSON(map[string]any{
					"ok": true, "source": src, "target": abs,
					"files": files, "skipped": skipped,
				})
			}
			if skipped {
				fmt.Printf("BDL 文档已是最新: %s\n", abs)
			} else {
				fmt.Printf("BDL 文档已解压: %d 个文件 -> %s\n", files, abs)
			}
			if setCfg || strings.TrimSpace(dir) != "" {
				fmt.Printf("已写入配置 bdldoc.dir: %s\n", abs)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "解压到指定目录（同时写回 config.json 的 bdldoc.dir）")
	cmd.Flags().BoolVar(&force, "force", false, "强制重新解压（默认同一 zip 只解一次）")
	return cmd
}

// installBDLDocs 解压 zip 到 dest（去掉 zip 里的顶层 BDL-Markdown/ 目录）。
// 返回解压文件数；skipped=true 表示 marker 显示同一 zip 已解压过。
func installBDLDocs(src, dest string, force bool) (files int, skipped bool, err error) {
	hash, err := fileSHA256(src)
	if err != nil {
		return 0, false, err
	}
	markerPath := filepath.Join(dest, bdlMarkerName)
	if !force {
		if m, ok := readBDLMarker(markerPath); ok && m.SHA256 == hash {
			return 0, true, nil
		}
	}
	zr, err := zip.OpenReader(src)
	if err != nil {
		return 0, false, fmt.Errorf("打开 %s 失败: %w", src, err)
	}
	defer zr.Close()

	if err := os.MkdirAll(dest, 0o755); err != nil {
		return 0, false, fmt.Errorf("建 BDL 文档目录失败 %s: %w", dest, err)
	}
	prefix := zipTopDir(zr.File)
	base := filepath.Clean(dest)
	for _, f := range zr.File {
		rel := strings.TrimPrefix(f.Name, prefix)
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			continue
		}
		target := filepath.Join(base, filepath.FromSlash(rel))
		if target != base && !strings.HasPrefix(target, base+string(os.PathSeparator)) {
			return 0, false, fmt.Errorf("zip 条目越出目标目录: %s", f.Name)
		}
		if f.FileInfo().IsDir() || strings.HasSuffix(f.Name, "/") {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return 0, false, err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return 0, false, err
		}
		mode := f.Mode().Perm()
		if mode == 0 {
			mode = 0o644
		}
		in, err := f.Open()
		if err != nil {
			return 0, false, err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
		if err != nil {
			in.Close()
			return 0, false, err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		in.Close()
		if copyErr != nil {
			return 0, false, copyErr
		}
		if closeErr != nil {
			return 0, false, closeErr
		}
		files++
	}
	m := bdlMarker{Zip: src, SHA256: hash, ExtractedAt: time.Now().Format(time.RFC3339), Files: files}
	if err := writeBDLMarker(markerPath, m); err != nil {
		return files, false, err
	}
	return files, false, nil
}

// zipTopDir 找 zip 的唯一顶层目录（本包是 BDL-Markdown/）；没有则返回空串。
// 兼容两种打包：显式目录条目，或只有文件（Go/PowerShell 都可能不打目录条目）。
func zipTopDir(files []*zip.File) string {
	prefix := ""
	for _, f := range files {
		name := strings.TrimSpace(strings.TrimSuffix(f.Name, "/"))
		if name == "" {
			continue
		}
		i := strings.IndexByte(name, '/')
		if i < 0 {
			if f.FileInfo().IsDir() {
				prefix = name // 显式顶层目录条目
				continue
			}
			return "" // 根级文件：不能剥掉任何一层
		}
		top := name[:i]
		if prefix == "" {
			prefix = top
		} else if prefix != top {
			return ""
		}
	}
	if prefix == "" {
		return ""
	}
	return prefix + "/"
}
func fileSHA256(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func readBDLMarker(p string) (bdlMarker, bool) {
	b, err := os.ReadFile(p)
	if err != nil {
		return bdlMarker{}, false
	}
	var m bdlMarker
	if err := json.Unmarshal(b, &m); err != nil {
		return bdlMarker{}, false
	}
	return m, m.SHA256 != ""
}

func writeBDLMarker(p string, m bdlMarker) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}

// ensureBDLDocs 在 tt serve 启动时调用：便携包第一次运行也能把文档解压出来。
// 任何失败都只提示，不阻止服务启动。
func ensureBDLDocs() {
	src, err := bdlZipSource()
	if err != nil {
		return
	}
	if _, err := os.Stat(src); err != nil {
		return
	}
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return
	}
	root, err := config.Load(cfgPath)
	if err != nil {
		return
	}
	dst := strings.TrimSpace(root.Bdldoc.Dir)
	setCfg := dst == ""
	if dst == "" {
		dst = filepath.Join(filepath.Dir(cfgPath), "BDL-Markdown")
	}
	abs, err := filepath.Abs(dst)
	if err != nil {
		return
	}
	files, skipped, err := installBDLDocs(src, abs, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "BDL 文档解压失败（不影响服务）: %v\n", err)
		return
	}
	if !skipped {
		fmt.Printf("BDL 文档已解压: %d 个文件 -> %s\n", files, abs)
	}
	if setCfg {
		_ = setBDLDocDir(cfgPath, abs)
	}
}

func setBDLDocDir(cfgPath, dir string) error {
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return err
	}
	return config.EditSection(cfgPath, "bdldoc", nil, func(sec map[string]any) error {
		sec["dir"] = dir
		return nil
	})
}
