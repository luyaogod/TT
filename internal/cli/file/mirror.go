package file

// tt file mirror：离线镜像（source: 4gl/42f/42s/inc；interface: 42f/42s）。
//
//   mirror dir [<目录>]              查看/设置镜像根目录（config.json 顶层 mirror.dir）
//   mirror pull [<环境>]             全量拉取该环境镜像
//   mirror path [<环境>]             打印该环境镜像目录的绝对路径（AI 直接前往）
//
// 镜像根必须显式设置（不设默认），避免隐式落盘位置。目录结构与服务器同构：
// <根>/<环境>/erp/... com/...（source 含 4gl/**/*.4gl + 42f + 42s/zh_CN + *.inc，
// interface 只含 42f/42s）。
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
	var profile string
	var tmpDir string
	g := &cobra.Command{
		Use:   "mirror",
		Short: "管理 T100 本地镜像（source: 4gl/42f/42s/inc；interface: 只 42f/42s）",
		Long: `把某环境（T100 服务器）的源码/界面文件镜像到本地目录。
profile=source（默认）取四类：4gl 目录下的 .4gl 源码、42f（编译表单）、
42s/zh_CN（编译字符串）、*.inc（4GL include）。**这四类都是本地命令的饭票**：
tt outline 只认 4gl，tt interface 只认 42f/42s，而两者读的是同一个镜像目录。
所以默认 profile 故意收全，不要为了省体积把它改窄（改窄只会让某个命令“找不到文件”）。
profile=interface 只取 42f/42s，想只要画面时配一个独立目录用。
4fd 不镜像（设计期 fieldId 不稳定）。per、其它语言目录、.4gl.src 等不拉；
备份文件（*.bak/*.bck/*~）也排除。目录结构与服务器同构：<镜像目录>/erp/... com/...。

拉取方式只有一种：**全量拉取**。每次都在服务器上重压一遍白名单内的全部文件，
本地整目录替换（服务器上已删的文件，镜像里也会消失）。没有增量、没有服务器侧基线。

服务器上**不留文件**：归档先写到服务器临时目录（默认 ${TMPDIR:-/tmp}，可用
--tmp 或配置 mirror.tmpDir 改），下完就删；中途取消、失败也删。只有客户端被强杀
才可能留下一个归档，下次拉取前会自动扫掉 3 小时前的遗留。

子命令：
  dir [<目录>]      查看/设置全局根目录；加 --at <环境> 则设置该环境单独目录
  pull [<环境>]     全量拉取该环境镜像
  path [<环境>]     打印该环境镜像目录的绝对路径（AI 直接前往）

环境用 --at <环境> 点名（也可给位置参数，二选一）；全局根与该环境单独目录都未设置时，
该环境的 pull/path 会提示先 dir。`,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	dirCmd := newMirrorDirCmd()
	pullCmd := newMirrorPullCmd(&full, &profile, &tmpDir)
	pathCmd := newMirrorPathCmd()
	pullCmd.Flags().BoolVar(&full, "full", false, "已弃用：现在只有全量拉取，此开关无作用")
	pullCmd.Flags().StringVar(&tmpDir, "tmp", "", "服务器临时目录（覆盖配置 mirror.tmpDir；空=服务器 ${TMPDIR:-/tmp}）")
	pullCmd.Flags().StringVar(&profile, "profile", "source", "镜像白名单：source(4gl+42f+42s+inc) | interface(只 42f/42s)")
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

// readMirrorTmpDir 读 config.json 顶层 mirror.tmpDir（服务器临时目录）。
// 空 = 用服务器上的 ${TMPDIR:-/tmp}，这里不做任何默认值填充。
func readMirrorTmpDir(cfgPath string) (string, error) {
	root, err := config.Load(cfgPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(root.Mirror.TmpDir), nil
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

func newMirrorPullCmd(full *bool, profile *string, tmpDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "pull [<环境>]",
		Short: "全量拉取该环境镜像（每次重压全部白名单文件）",
		Long: `SSH 登录环境，按 profile 的白名单打包：source 取 4gl 目录下的 .4gl + 42f + 42s/zh_CN + *.inc；interface 只取 42f/42s，供 tt interface 合成画面。
下载到该环境的镜像目录（环境单独指定，未指定则用全局根目录/<环境名>）。

**只有全量拉取**：每次都在服务器上重压一遍白名单内的全部文件，本地整目录替换。
只对本 profile 的白名单负责：不属于它的既有文件（如给 interface 拉取时的 4gl/inc）
原样保留。服务器上已删的文件，镜像里也会消失。

归档先写到**服务器临时目录**（默认 ${TMPDIR:-/tmp}，可用 --tmp <目录> 或配置
mirror.tmpDir 指定），下载完立即删除；取消、失败也删。所以 $TOP 里不会多出任何文件，
也不要求登录账号对 $TOP 有写权限。

打包根（$TOP）按登录区域动态探测；探测失败直接报错。
--full 已弃用（现在恒为全量），保留只是为了不打断旧脚本。`,
		Example: `  tt file mirror pull --at 主机109
  tt file mirror pull --at 主机109 --profile interface
  tt file mirror pull --at 主机109 --tmp /u2/tmp`,
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
			prof, err := host.ParseMirrorProfile(*profile)
			if err != nil {
				return usageErr("%v", err)
			}
			tmp, err := readMirrorTmpDir(cfgPath)
			if err != nil {
				return err
			}
			if strings.TrimSpace(*tmpDir) != "" {
				tmp = strings.TrimSpace(*tmpDir)
			}
			jsonOut := common.IsJSON()
			if !jsonOut {
				fmt.Printf("正在拉取 %s 镜像（全量，%s）: %s -> %s\n", e.Name, prof.Label(), e.Host, envDir)
				if *full {
					fmt.Println("(--full 已弃用：现在恒为全量拉取，此开关无作用)")
				}
			}
			st, err := host.MirrorPullToDirProfile(cmd.Context(), e, envDir, prof, host.MirrorOptions{TmpDir: tmp})
			if err != nil {
				return outputErrConn(err)
			}
			if jsonOut {
				return common.PrintJSON(map[string]any{
					"ok": true, "env": e.Name, "dir": envDir,
					"files": st.Files, "bytes": st.Bytes, "pruned": st.Pruned,
					"elapsed": st.Elapsed, "topDir": st.TopDir, "tmpDir": st.TmpDir,
					"note": st.Note, "profile": string(st.Profile),
				})
			}
			if st.Note != "" {
				fmt.Printf("%s\n", st.Note)
			} else {
				fmt.Printf("镜像完成: %d 个文件, %s, 用时 %s\n", st.Files, fmtBytes(st.Bytes), st.Elapsed)
				if st.Pruned > 0 {
					fmt.Printf("已清理本地残留备份文件: %d 个（如 *.bak/*.bck；本 profile 只保留 %s）\n", st.Pruned, st.Profile.Label())
				}
			}
			fmt.Printf("镜像目录: %s\n", envDir)
			if st.TopDir != "" {
				fmt.Printf("服务器 TOP: %s\n", st.TopDir)
			}
			if st.TmpDir != "" {
				fmt.Printf("服务器临时目录: %s（归档已删除）\n", st.TmpDir)
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
