package dict

// tt dict mirror:管理 T100 本地源码镜像(服务器 4gl/4fd/42s-zh_CN/*.inc → 本地目录,供 AI 读码)。
//   mirror dir [<目录>]   查看/设置镜像根目录(config.json 顶层 "mirror": {"dir": ...});
//   mirror pull [<环境名>] [--full]   下载/更新该环境镜像(默认增量);
//   mirror path [<环境名>]  打印该环境镜像目录的绝对路径(AI 直接前往)。
//
// 镜像根目录必须显式设置(不设默认),避免隐式落盘位置。镜像目录结构:
// <根>/<环境名>/erp/... com/...(与服务器同构;只含各模块 4gl/4fd、42s/zh_CN 与 *.inc)。

import (
	"fmt"
	"os"
	"path/filepath"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/host"

	"github.com/spf13/cobra"
)

var (
	mirrorCfg     *host.Hosts
	mirrorCfgPath string
	mirrorFull    bool
)

var mirrorCmd = &cobra.Command{
	Use:   "mirror",
	Short: "管理 T100 本地源码镜像(服务器 4gl/4fd/42s/*.inc → 本地目录)",
	Long: `把某环境(T100 服务器)的源代码镜像到本地目录,供 AI 用本地文件工具读码:
只镜像各模块 4gl(源码)、4fd(前端字段描述)、42s(编译字符串,仅 zh_CN 语言目录),
以及 *.inc(4GL include,任意位置);
per/编译产物/其它语言等一律不拉;备份文件(*.bak/*.bck/*~)也排除。
目录结构与服务器同构:<镜像根>/<环境名>/erp/... com/...。

子命令:
  dir [<目录>]     查看或设置镜像根目录(config.json 顶层 mirror.dir;必须显式设置)
  pull [<环境名>]  下载/更新该环境镜像(默认增量;--full 全量重建)
  path [<环境名>]  打印该环境镜像目录的绝对路径(AI 直接前往)

环境名缺省取 hosts.activeEnv,未设置取 sshs 首条。镜像根未设置时 pull/path 会提示先 dir。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 与 dbCmd 相同:覆盖 Group 的查询数据源钩子,镜像命令不依赖本地字典库
		if err := skipQuerySource(cmd, args); err != nil {
			return err
		}
		path, err := common.ResolveConfig(false)
		if err != nil {
			return err
		}
		cfg, err := config.LoadHosts(path)
		if err != nil {
			return err
		}
		mirrorCfgPath = path
		mirrorCfg = cfg
		return nil
	},
}

func init() {
	Group.AddCommand(mirrorCmd)
	mirrorCmd.AddCommand(mirrorDirCmd, mirrorPullCmd, mirrorPathCmd)
	mirrorPullCmd.Flags().BoolVar(&mirrorFull, "full", false, "全量重建(整目录替换,含删除服务器已不存在的文件)")
}

// ---- 顶层 mirror 配置读写 ----

// readMirrorDir 读 config.json 顶层 mirror.dir;返回绝对路径;ok=false=未配置。
func readMirrorDir() (string, bool) {
	root, err := config.Load(mirrorCfgPath)
	if err != nil || root.Mirror.Dir == "" {
		return "", false
	}
	abs, err := filepath.Abs(root.Mirror.Dir)
	if err != nil {
		return root.Mirror.Dir, true
	}
	return abs, true
}

// writeMirrorDir 写回 config.json 顶层 mirror.dir(原子;其余顶层键原样保留)。
func writeMirrorDir(dir string) error {
	return config.EditSection(mirrorCfgPath, "mirror", nil, func(sec map[string]any) error {
		sec["dir"] = dir
		return nil
	})
}

// mirrorEnv 按名(参数 > activeEnv > 首条)解析环境。
func mirrorEnv(name string) (*host.NamedSsh, error) {
	e := mirrorCfg.ByName(name)
	if e == nil {
		return nil, fmt.Errorf("未找到环境 %q(可用: tt env list 查看环境名)", name)
	}
	return e, nil
}

// ---- mirror dir ----

var mirrorDirCmd = &cobra.Command{
	Use:   "dir [<目录>]",
	Short: "查看或设置本地镜像根目录(config.json 顶层 mirror.dir)",
	Long: `不带参数显示当前镜像根目录;带参数把新的镜像根目录写入 config.json
(顶层 "mirror" 键,与 hosts/debug/query 平级)。镜像根必须显式设置:
未设置时 mirror pull/path 会提示先执行本命令。目录可为相对路径,保存时转为绝对路径。`,
	Example: `  tt dict mirror dir                      # 显示当前镜像根
  tt dict mirror dir D:\dev\erp-src       # 设置镜像根(写入 config.json)`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			abs, err := filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("解析目录路径失败: %w", err)
			}
			if err := writeMirrorDir(abs); err != nil {
				return err
			}
			fmt.Printf("镜像根目录已设置: %s\n(镜像将保存在 %s\\<环境名>\\ 下;执行 tt dict mirror pull <环境名> 拉取)\n", abs, abs)
			return nil
		}
		dir, ok := readMirrorDir()
		if !ok {
			fmt.Println("镜像根目录未设置。请执行: tt dict mirror dir <目录>")
			return nil
		}
		fmt.Println(dir)
		return nil
	},
}

// ---- mirror pull ----

var mirrorPullCmd = &cobra.Command{
	Use:   "pull [<环境名>]",
	Short: "下载/更新该环境源码镜像(4gl/4fd/42s-zh_CN/*.inc;默认增量;--full 全量重建)",
	Long: `SSH 登录环境,按白名单(各模块 4gl/4fd、42s 下 zh_CN 语言目录、*.inc)打包源码
下载到本地镜像目录 <镜像根>/<环境名>/。默认增量:服务器 marker 记录上次基线,
只拉变更文件;--full 全量重建:本地整目录替换(删除服务器已不存在的残留文件)。
白名单升级(新增文件类型)后,下次拉取会自动转全量以补齐新增类型的历史文件。

打包根(TOP)按登录区域动态探测获取(T100 路径不允许静态配置);
探测失败的环境会直接报错(请检查该环境 SSH 登录与 zone 设置)。`,
	Example: `  tt dict mirror pull                     # 默认环境(activeEnv/首条)增量更新
  tt dict mirror pull 正式区           # 指定环境增量更新
  tt dict mirror pull 正式区 --full    # 全量重建(首次拉取也走全量)`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, ok := readMirrorDir()
		if !ok {
			return fmt.Errorf("镜像根目录未设置。请先执行: tt dict mirror dir <目录>")
		}
		envName := ""
		if len(args) == 1 {
			envName = args[0]
		}
		e, err := mirrorEnv(envName)
		if err != nil {
			return err
		}
		mode := "增量"
		if mirrorFull {
			mode = "全量"
		}
		fmt.Printf("正在拉取 %s 源码镜像(%s): %s -> %s\n", e.Name, mode, e.Host, filepath.Join(dir, e.Name))
		st, err := host.MirrorPull(e, dir, mirrorFull)
		if err != nil {
			return err
		}
		if st.Note != "" {
			fmt.Printf("%s\n", st.Note)
		} else {
			fmt.Printf("镜像完成: %d 个文件, %s, 用时 %s\n", st.Files, fmtBytes(st.Bytes), st.Elapsed)
			if st.Pruned > 0 {
				fmt.Printf("已清理本地残留备份文件: %d 个(如 *.bak/*.bck;镜像只保留 4gl/4fd、42s-zh_CN 与 *.inc)\n", st.Pruned)
			}
		}
		if st.Full && !mirrorFull {
			fmt.Println("(本地无完整基线或镜像白名单已升级,本次自动按全量拉取)")
		}
		fmt.Printf("镜像目录: %s\n", filepath.Join(dir, e.Name))
		if st.TopDir != "" {
			fmt.Printf("服务器 TOP: %s\n", st.TopDir)
		}
		return nil
	},
}

// ---- mirror path ----

var mirrorPathCmd = &cobra.Command{
	Use:   "path [<环境名>]",
	Short: "打印该环境镜像目录的绝对路径(AI 直接前往)",
	Long: `打印 <镜像根>/<环境名> 的绝对路径,第一行即路径本身(便于 AI/脚本取用)。
目录尚未拉取时会在路径后附提示。`,
	Example: `  tt dict mirror path
  tt dict mirror path 正式区`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, ok := readMirrorDir()
		if !ok {
			return fmt.Errorf("镜像根目录未设置。请先执行: tt dict mirror dir <目录>")
		}
		envName := ""
		if len(args) == 1 {
			envName = args[0]
		}
		e, err := mirrorEnv(envName)
		if err != nil {
			return err
		}
		p := filepath.Join(dir, e.Name)
		fmt.Println(p)
		if fi, err := os.Stat(p); err != nil || !fi.IsDir() {
			fmt.Printf("(该环境尚未拉取;先执行 tt dict mirror pull %s)\n", e.Name)
			return nil
		}
		return nil
	},
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
