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
备份文件（*.bak/*.bck/*~）也排除。目录结构与服务器同构：<镜像目录>/erp/... com/...。
每个环境可以单独指定镜像目录；未单独指定时回退到全局根目录/<环境名>。

子命令：
  dir [<目录>]      查看/设置全局根目录；加 --at <环境> 则设置该环境单独目录
  pull [<环境>]     下载/更新该环境镜像（默认增量；--full 全量重建）
  path [<环境>]     打印该环境镜像目录的绝对路径（AI 直接前往）

环境用 --at <环境> 点名（也可给位置参数，二选一）；全局根与该环境单独目录都未设置时，
该环境的 pull/path 会提示先 dir。`,
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

// setEnvMirrorDir 写某环境自己的镜像目录（空串 = 删除覆盖，回退全局根目录）。
// 直接改 hosts.sshs[] 的 map，保留环境里本页不认识的键（launchArgs 等）。
func setEnvMirrorDir(cfgPath, env, dir string) error {
	return config.Edit(cfgPath, nil, func(root map[string]any) error {
		hosts, _ := root["hosts"].(map[string]any)
		sshs, _ := hosts["sshs"].([]any)
		for _, item := range sshs {
			m, _ := item.(map[string]any)
			if m == nil || m["name"] != env {
				continue
			}
			if strings.TrimSpace(dir) == "" {
				delete(m, "mirrorDir")
			} else {
				m["mirrorDir"] = dir
			}
			return nil
		}
		return fmt.Errorf("未找到环境 %q", env)
	})
}

// mirrorEnvPath 取某环境的最终镜像目录：环境自己的 mirrorDir 优先，否则顶层根目录/<环境名>。
func mirrorEnvPath(cfgPath string, e *host.NamedSsh) (string, error) {
	root, err := config.Load(cfgPath)
	if err != nil {
		return "", err
	}
	return e.MirrorPath(root.Mirror.Dir), nil
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
		return "", usageErr("需要点名环境：tt file mirror pull --at <环境>（或 tt file mirror path --at <环境>；可用环境见 tt env list）")
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
	var clear bool
	c := &cobra.Command{
		Use:   "dir [<目录>]",
		Short: "查看/设置镜像目录（全局根目录，或 --at 指定环境单独目录）",
		Long: `不带 --at：查看或设置全局镜像根目录（config.json 顶层 mirror.dir）。
带 --at <环境>：查看或设置该环境自己的镜像目录；不设置时回退到全局根目录/<环境名>。
--clear 只清除某环境的单独目录（必须同时给 --at <环境>），回退到全局根目录。

目录可为相对路径，保存时转为绝对路径。全局根目录和每个环境的单独目录都未设置时，
该环境的 mirror pull/path 会提示先执行本命令。`,
		Example: `  tt file mirror dir
  tt file mirror dir D:\dev\erp-src
  tt file mirror dir D:\dev\erp-src-109tst --at 109tst
  tt file mirror dir --at 109tst --clear`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := common.ResolveConfig(false)
			if err != nil {
				return err
			}
			env := strings.TrimSpace(common.Env)
			if clear && env == "" {
				return usageErr("--clear 需要同时给 --at <环境>（只清除该环境的单独镜像目录）")
			}
			if env != "" {
				hosts, err := config.LoadHosts(cfgPath)
				if err != nil {
					return err
				}
				e := hosts.ByName(env)
				if e == nil {
					return usageErr("未找到环境 %q（tt env list 看全部）", env)
				}
				if clear {
					if len(args) == 1 {
						return usageErr("--clear 与目录参数不能同时给")
					}
					if err := setEnvMirrorDir(cfgPath, env, ""); err != nil {
						return err
					}
					e.MirrorDir = "" // 本地副本也要清，下面算回退路径时才是真的回退
					dir, _ := mirrorEnvPath(cfgPath, e)
					if common.IsJSON() {
						return common.PrintJSON(map[string]any{"ok": true, "env": env, "dir": dir, "cleared": true})
					}
					if dir == "" {
						fmt.Printf("已清除环境 %s 的单独镜像目录；全局根目录也没有，mirror pull/path 仍会报未设置。\n", env)
					} else {
						fmt.Printf("已清除环境 %s 的单独镜像目录，现回退到: %s\n", env, dir)
					}
					return nil
				}
				if len(args) == 1 {
					abs, err := filepath.Abs(args[0])
					if err != nil {
						return usageErr("解析目录路径失败: %v", err)
					}
					if err := setEnvMirrorDir(cfgPath, env, abs); err != nil {
						return err
					}
					if common.IsJSON() {
						return common.PrintJSON(map[string]any{"ok": true, "env": env, "dir": abs, "updated": true})
					}
					fmt.Printf("环境 %s 的镜像目录已设置: %s\n", env, abs)
					return nil
				}
				dir, _ := mirrorEnvPath(cfgPath, e)
				custom := strings.TrimSpace(e.MirrorDir) != ""
				if common.IsJSON() {
					return common.PrintJSON(map[string]any{"ok": true, "env": env, "dir": dir, "configured": dir != "", "custom": custom})
				}
				if dir == "" {
					fmt.Printf("环境 %s 的镜像目录未设置。请执行: tt file mirror dir <目录> --at %s\n", env, env)
					return nil
				}
				fmt.Println(dir)
				return nil
			}
			if len(args) == 1 {
				abs, err := filepath.Abs(args[0])
				if err != nil {
					return usageErr("解析目录路径失败: %v", err)
				}
				if err := writeMirrorDir(cfgPath, abs); err != nil {
					return err
				}
				if common.IsJSON() {
					return common.PrintJSON(map[string]any{"ok": true, "dir": abs, "updated": true})
				}
				fmt.Printf("镜像根目录已设置: %s\n(镜像将保存在 %s\\<环境名>\\ 下；执行 tt file mirror pull --at <环境> 拉取)\n", abs, abs)
				return nil
			}
			dir, ok := readMirrorDir(cfgPath)
			if common.IsJSON() {
				return common.PrintJSON(map[string]any{"ok": true, "dir": dir, "configured": ok})
			}
			if !ok {
				fmt.Println("镜像根目录未设置。请执行: tt file mirror dir <目录>（或加 --at <环境> 只给该环境）")
				return nil
			}
			fmt.Println(dir)
			return nil
		},
	}
	c.Flags().BoolVar(&clear, "clear", false, "清除某环境自己的镜像目录，回退到全局根目录（需 --at）")
	return c
}

// ---- mirror pull ----

func newMirrorPullCmd(full *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "pull [<环境>]",
		Short: "下载/更新该环境源码镜像（默认增量；--full 全量重建）",
		Long: `SSH 登录环境，按白名单（各模块 4gl/4fd、42s 下 zh_CN 语言目录、*.inc）打包源码，
下载到该环境的镜像目录（环境单独指定，未指定则用全局根目录/<环境名>）。默认增量：
服务器 marker 记录上次基线，只拉变更文件；--full 全量重建。白名单升级后，下次拉取会自动转全量以补齐新增类型。

打包根（$TOP）按登录区域动态探测；探测失败直接报错。`,
		Example: `  tt file mirror pull --at 主机109
  tt file mirror pull --at 主机109 --full`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hosts, cfgPath, err := mirrorConfig()
			if err != nil {
				return err
			}
			name, err := mirrorEnvName(args)
			if err != nil {
				return err
			}
			e, err := mirrorEnvOf(hosts, name)
			if err != nil {
				return err
			}
			envDir, err := mirrorEnvPath(cfgPath, e)
			if err != nil {
				return err
			}
			if envDir == "" {
				return usageErr("环境 %s 的镜像目录未设置。请先执行: tt file mirror dir <目录>（可加 --at %s 只给该环境）", e.Name, e.Name)
			}
			mode := "增量"
			if *full {
				mode = "全量"
			}
			jsonOut := common.IsJSON()
			if !jsonOut {
				fmt.Printf("正在拉取 %s 源码镜像（%s）: %s -> %s\n", e.Name, mode, e.Host, envDir)
			}
			st, err := host.MirrorPullToDir(cmd.Context(), e, envDir, *full)
			if err != nil {
				return outputErrConn(err)
			}
			if jsonOut {
				return common.PrintJSON(map[string]any{
					"ok": true, "env": e.Name, "dir": envDir,
					"files": st.Files, "bytes": st.Bytes, "pruned": st.Pruned,
					"elapsed": st.Elapsed, "topDir": st.TopDir, "note": st.Note,
					"full": st.Full,
				})
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
			fmt.Printf("镜像目录: %s\n", envDir)
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
		Long: `打印该环境镜像目录的绝对路径（环境单独指定，未指定则用全局根目录/<环境名>），第一行即路径本身（便于 AI/脚本取用）。
目录尚未拉取时会在路径后附提示。`,
		Example: `  tt file mirror path --at 主机109`,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			hosts, cfgPath, err := mirrorConfig()
			if err != nil {
				return err
			}
			name, err := mirrorEnvName(args)
			if err != nil {
				return err
			}
			e, err := mirrorEnvOf(hosts, name)
			if err != nil {
				return err
			}
			p, err := mirrorEnvPath(cfgPath, e)
			if err != nil {
				return err
			}
			if p == "" {
				return usageErr("环境 %s 的镜像目录未设置。请先执行: tt file mirror dir <目录>（可加 --at %s 只给该环境）", e.Name, e.Name)
			}
			_, statErr := os.Stat(p)
			exists := statErr == nil
			if common.IsJSON() {
				return common.PrintJSON(map[string]any{"ok": true, "env": e.Name, "path": p, "exists": exists})
			}
			fmt.Println(p)
			if !exists {
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
