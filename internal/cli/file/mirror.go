package file

// tt file mirror：离线源码镜像（服务器 4gl/4fd/42s-zh_CN/*.inc → 本地目录）。
//
//   mirror dir [<目录>]              查看/设置镜像根目录（config.json 顶层 mirror.dir）
//   mirror pull [<环境>] [--full]    下载/更新该环境镜像（默认增量）
//   mirror path [<环境>]             打印该环境镜像目录的绝对路径（AI 直接前往）
//
// 镜像根必须显式设置（不设默认），避免隐式落盘位置。目录结构与服务器同构：
// <根>/<环境>/erp/... com/...（只含各模块 4gl/4fd、42s/zh_CN 与 *.inc）。
//
// 它是“离线口子”：联网时 AI 用 tt file cat/grep 现读；断网时读这份子集。

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/host"
)

// NewMirrorCmd 构造 mirror 子树（file 组与 dict 的隐藏转发各持一棵，
// 避免同一个 *cobra.Command 被挂到两个父命令下）。
func NewMirrorCmd() *cobra.Command {
	var full bool
	g := &cobra.Command{
		Use:   "mirror",
		Short: "管理 T100 本地源码镜像（服务器 4gl/4fd/42s/*.inc → 本地目录）",
		Long: `把某环境（T100 服务器）的源代码镜像到本地目录，供 AI 用本地文件工具读码：
只镜像各模块 4gl（源码）、4fd（前端字段描述）、42s（编译字符串，仅 zh_CN 语言目录），
以及 *.inc（4GL include，任意位置）；per/编译产物/其它语言等一律不拉；
备份文件（*.bak/*.bck/*~）也排除。目录结构与服务器同构：<镜像根>/<环境名>/erp/... com/...。

子命令：
  dir [<目录>]      查看或设置镜像根目录（config.json 顶层 mirror.dir；必须显式设置）
  pull [<环境>]     下载/更新该环境镜像（默认增量；--full 全量重建）
  path [<环境>]     打印该环境镜像目录的绝对路径（AI 直接前往）

环境用 --at <环境> 点名（也可给位置参数，二选一）；镜像根未设置时 pull/path 会提示先 dir。`,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	dirCmd := newMirrorDirCmd()
	pullCmd := newMirrorPullCmd(&full)
	pathCmd := newMirrorPathCmd()
	pullCmd.Flags().BoolVar(&full, "full", false, "全量重建（整目录替换，含删除服务器已不存在的文件）")
	g.AddCommand(dirCmd, pullCmd, pathCmd)
	return g
}

// ---- 配置读写 ----

func mirrorConfig() (*host.Hosts, string, error) {
	path, err := common.ResolveConfig(false)
	if err != nil {
		return nil, "", err
	}
	hosts, err := config.LoadHosts(path)
	if err != nil {
		return nil, "", err
	}
	return hosts, path, nil
}

// readMirrorDir 读 config.json 顶层 mirror.dir；返回绝对路径；ok=false=未配置。
func readMirrorDir(cfgPath string) (string, bool) {
	root, err := config.Load(cfgPath)
	if err != nil || root.Mirror.Dir == "" {
		return "", false
	}
	abs, err := filepath.Abs(root.Mirror.Dir)
	if err != nil {
		return root.Mirror.Dir, true
	}
	return abs, true
}

// writeMirrorDir 写回 config.json 顶层 mirror.dir（原子；其余顶层键原样保留）。
func writeMirrorDir(cfgPath, dir string) error {
	return config.EditSection(cfgPath, "mirror", nil, func(sec map[string]any) error {
		sec["dir"] = dir
		return nil
	})
}

// mirrorEnvName 取本次环境：--at 优先，位置参数也可（两者给得不一致就报错）。
func mirrorEnvName(args []string) (string, error) {
	flagEnv := strings.TrimSpace(common.Env)
	name := flagEnv
	if len(args) == 1 {
		arg := strings.TrimSpace(args[0])
		if flagEnv != "" && arg != "" && flagEnv != arg {
			return "", usageErr("--at 的环境（%s）与位置参数（%s）不一致：二选一", flagEnv, arg)
		}
		name = arg
	}
	if name == "" {
		return "", usageErr("需要点名环境：tt file mirror pull|path --at <环境>（tt env list 看全部）")
	}
	return name, nil
}

func mirrorEnvOf(hosts *host.Hosts, name string) (*host.NamedSsh, error) {
	ref, err := hosts.Resolve(name)
	if err != nil {
		return nil, usageErr("%v", err)
	}
	return ref.Env, nil
}

// ---- mirror dir ----

func newMirrorDirCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dir [<目录>]",
		Short: "查看或设置本地镜像根目录（config.json 顶层 mirror.dir）",
		Long: `不带参数显示当前镜像根目录；带参数把新目录写入 config.json（顶层 "mirror" 键）。
镜像根必须显式设置：未设置时 mirror pull/path 会提示先执行本命令。
目录可为相对路径，保存时转为绝对路径。`,
		Example: `  tt file mirror dir
  tt file mirror dir D:\dev\erp-src`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfgPath, err := common.ResolveConfig(false)
			if err != nil {
				return err
			}
			if len(args) == 1 {
				abs, err := filepath.Abs(args[0])
				if err != nil {
					return usageErr("解析目录路径失败: %v", err)
				}
				if err := writeMirrorDir(cfgPath, abs); err != nil {
					return err
				}
				fmt.Printf("镜像根目录已设置: %s\n(镜像将保存在 %s\\<环境名>\\ 下；执行 tt file mirror pull --at <环境> 拉取)\n", abs, abs)
				return nil
			}
			dir, ok := readMirrorDir(cfgPath)
			if !ok {
				fmt.Println("镜像根目录未设置。请执行: tt file mirror dir <目录>")
				return nil
			}
			fmt.Println(dir)
			return nil
		},
	}
}

// ---- mirror pull ----

func newMirrorPullCmd(full *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "pull [<环境>]",
		Short: "下载/更新该环境源码镜像（默认增量；--full 全量重建）",
		Long: `SSH 登录环境，按白名单（各模块 4gl/4fd、42s 下 zh_CN 语言目录、*.inc）打包源码，
下载到 <镜像根>/<环境名>/。默认增量：服务器 marker 记录上次基线，只拉变更文件；
--full 全量重建。白名单升级后，下次拉取会自动转全量以补齐新增类型。

打包根（$TOP）按登录区域动态探测；探测失败直接报错。`,
		Example: `  tt file mirror pull --at 主机109
  tt file mirror pull --at 主机109 --full`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hosts, cfgPath, err := mirrorConfig()
			if err != nil {
				return err
			}
			dir, ok := readMirrorDir(cfgPath)
			if !ok {
				return usageErr("镜像根目录未设置。请先执行: tt file mirror dir <目录>")
			}
			name, err := mirrorEnvName(args)
			if err != nil {
				return err
			}
			e, err := mirrorEnvOf(hosts, name)
			if err != nil {
				return err
			}
			mode := "增量"
			if *full {
				mode = "全量"
			}
			fmt.Printf("正在拉取 %s 源码镜像（%s）: %s -> %s\n", e.Name, mode, e.Host, filepath.Join(dir, e.Name))
			st, err := host.MirrorPull(cmd.Context(), e, dir, *full)
			if err != nil {
				return outputErrConn(err)
			}
			if st.Note != "" {
				fmt.Printf("%s\n", st.Note)
			} else {
				fmt.Printf("镜像完成: %d 个文件, %s, 用时 %s\n", st.Files, fmtBytes(st.Bytes), st.Elapsed)
				if st.Pruned > 0 {
					fmt.Printf("已清理本地残留备份文件: %d 个（如 *.bak/*.bck；镜像只保留 4gl/4fd、42s_zh_CN 与 *.inc）\n", st.Pruned)
				}
			}
			if st.Full && !*full {
				fmt.Println("(本地无完整基线或镜像白名单已升级，本次自动按全量拉取)")
			}
			fmt.Printf("镜像目录: %s\n", filepath.Join(dir, e.Name))
			if st.TopDir != "" {
				fmt.Printf("服务器 TOP: %s\n", st.TopDir)
			}
			return nil
		},
	}
}

// ---- mirror path ----

func newMirrorPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path [<环境>]",
		Short: "打印该环境镜像目录的绝对路径（AI 直接前往）",
		Long: `打印 <镜像根>/<环境名> 的绝对路径，第一行即路径本身（便于 AI/脚本取用）。
目录尚未拉取时会在路径后附提示。`,
		Example: `  tt file mirror path --at 主机109`,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			hosts, cfgPath, err := mirrorConfig()
			if err != nil {
				return err
			}
			dir, ok := readMirrorDir(cfgPath)
			if !ok {
				return usageErr("镜像根目录未设置。请先执行: tt file mirror dir <目录>")
			}
			name, err := mirrorEnvName(args)
			if err != nil {
				return err
			}
			e, err := mirrorEnvOf(hosts, name)
			if err != nil {
				return err
			}
			p := filepath.Join(dir, e.Name)
			fmt.Println(p)
			if fi, err := os.Stat(p); err != nil || !fi.IsDir() {
				fmt.Printf("(该环境尚未拉取；先执行 tt file mirror pull --at %s)\n", e.Name)
			}
			return nil
		},
	}
}

// fmtBytes 字节数人性化显示。
func fmtBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
