package config

// 「开发资料」（devdata）的落点：**一个环境一整棵目录**，与读侧镜像/字典库完全分开。
//
// 为什么另起一棵树，而不是复用 mirror.dir 或 dict/：这两处是"给读代码用的"，
// 有各自的清理口径（`tt cache` 会按白名单统计，`tt dict db sync` 整库替换）。
// 开发资料是**给 model 的原材料**：按程序成套、带 manifest 指纹、过期判断靠版本，
// 被删了要重新拉且代价高。放进缓存清单里，7 天清理规则会在启动时把它删掉。
// 所以：不进 pkg/cachedir 的白名单，清理只走 `tt gen clean`（见 plan ADR-02）。
//
// 目录形态（plan §5.1）：
//
//	<data>/devdata/<env>/manifest.json         M 侧清单
//	<data>/devdata/<env>/dict/<env>.db         M 侧规格库（53 张 DZ*_T）
//	<data>/devdata/<env>/mdl/…                 M 侧模板树（保持 $TOP/erp/mdl 相对路径）
//	<data>/devdata/<env>/prog/<prog>/…         D 侧按程序成套
//
// 与字典库的差别只有根：`<data>/dict/<env>.db` 是读侧，这里是 `devdata/<env>/dict/<env>.db`。
// 两处同名是**故意**的（同一个"环境→一个库"的直觉），所以 README 互相指路。

import (
	"os"
	"path/filepath"
	"strings"
)

// DevDataDirName 数据目录下存放各环境开发资料的子目录名。
const DevDataDirName = "devdata"

// DevDataSubdirDict / DevDataSubdirMdl / DevDataSubdirProg 是环境目录下的固定子目录名。
const (
	DevDataSubdirDict = "dict"
	DevDataSubdirMdl  = "mdl"
	DevDataSubdirProg = "prog"
)

// DevDataDir 开发资料根目录（数据目录下 devdata/）。
func DevDataDir(dataDir string) string {
	if strings.TrimSpace(dataDir) == "" {
		return ""
	}
	return filepath.Join(dataDir, DevDataDirName)
}

// DevDataEnvDir 某环境的开发资料目录：<data>/devdata/<env>。
//
// 环境名直接当目录名（plan §5.1：含中文按原样建目录）。安全由**写入前的白名单**
// 保证（环境名必须来自配置里的 hosts.sshs，见 ValidateEnvName），不在这里做转义 ——
// 转义会让"配置里的名字"与"磁盘上的目录"对不上，而那正是排查时最费劲的一类问题。
func DevDataEnvDir(dataDir, env string) string {
	root := DevDataDir(dataDir)
	if root == "" || strings.TrimSpace(env) == "" {
		return ""
	}
	return filepath.Join(root, env)
}

// DevDataProgDir 某程序在开发资料里的目录：<data>/devdata/<env>/prog/<prog>。
func DevDataProgDir(dataDir, env, prog string) string {
	envDir := DevDataEnvDir(dataDir, env)
	if envDir == "" || strings.TrimSpace(prog) == "" {
		return ""
	}
	return filepath.Join(envDir, DevDataSubdirProg, prog)
}

// DevDataDictPath 某环境规格库文件：<data>/devdata/<env>/dict/<env>.db。
//
// 与 DictDBPath 一样，文件名固定为 `<环境>.db`（而不是让调用方随便指一个 .db）：
// 否则同一个环境可能被 CLI 指向 A、被界面指向 B，而"读到另一个库"与"数据不存在"
// 在输出上分不出来。
func DevDataDictPath(dataDir, env string) string {
	envDir := DevDataEnvDir(dataDir, env)
	if envDir == "" {
		return ""
	}
	return filepath.Join(envDir, DevDataSubdirDict, DictDBFileName(env))
}

// DevDataMdlDir 某环境模板树根：<data>/devdata/<env>/mdl。
func DevDataMdlDir(dataDir, env string) string {
	envDir := DevDataEnvDir(dataDir, env)
	if envDir == "" {
		return ""
	}
	return filepath.Join(envDir, DevDataSubdirMdl)
}

// DevDataRootDirFor 开发资料根目录（读配置）：顶层 devdata.dir 优先，否则 <data>/devdata。
//
// 库/模板位置一旦可配，CLI 与 Web 必须读**同一处**配置，才不会出现"命令行读 A、
// 界面读 B"。所以这里统一传配置路径、内部 Load，而不是让各调用点自己算。
// 配置读不到 / 写坏时安静回退到缺省落点 —— 资料位置不该因为配置有问题就变成空串。
func DevDataRootDirFor(cfgPath string) string {
	if strings.TrimSpace(cfgPath) != "" {
		if r, err := Load(cfgPath); err == nil {
			if d := strings.TrimSpace(r.DevData.Dir); d != "" {
				return AbsPath(d)
			}
		}
	}
	return DevDataDir(LocationsAt(cfgPath).DataDir())
}

// DevDataEnvDirForEnv 某环境实际使用的开发资料目录：
// 环境 hosts.sshs[].devdataDir > 顶层 devdata.dir > <数据目录>/devdata，再拼环境名。
//
// 只有**根目录**能按环境覆盖，环境目录名始终是配置里的环境名 ——「每环境一个任意路径」
// 会让同一份材料在 CLI/Web 两边指向不同位置。
func DevDataEnvDirForEnv(cfgPath, env string) string {
	if strings.TrimSpace(env) == "" {
		return ""
	}
	if strings.TrimSpace(cfgPath) != "" {
		if r, err := Load(cfgPath); err == nil {
			for i := range r.Hosts.SSHs {
				if r.Hosts.SSHs[i].Name != env {
					continue
				}
				if d := strings.TrimSpace(r.Hosts.SSHs[i].DevDataDir); d != "" {
					return filepath.Join(AbsPath(d), env)
				}
				break
			}
		}
	}
	root := DevDataRootDirFor(cfgPath)
	if root == "" {
		return ""
	}
	return filepath.Join(root, env)
}

// DevDataProgDirForEnv 某程序在开发资料里的目录（读配置，含环境级覆盖）。
func DevDataProgDirForEnv(cfgPath, env, prog string) string {
	envDir := DevDataEnvDirForEnv(cfgPath, env)
	if envDir == "" || strings.TrimSpace(prog) == "" {
		return ""
	}
	return filepath.Join(envDir, DevDataSubdirProg, prog)
}

// DevDataDictPathForEnv 某环境规格库文件（读配置，含环境级覆盖）。
func DevDataDictPathForEnv(cfgPath, env string) string {
	envDir := DevDataEnvDirForEnv(cfgPath, env)
	if envDir == "" {
		return ""
	}
	return filepath.Join(envDir, DevDataSubdirDict, DictDBFileName(env))
}

// DevDataMdlDirForEnv 某环境模板树根（读配置，含环境级覆盖）。
func DevDataMdlDirForEnv(cfgPath, env string) string {
	envDir := DevDataEnvDirForEnv(cfgPath, env)
	if envDir == "" {
		return ""
	}
	return filepath.Join(envDir, DevDataSubdirMdl)
}

// DevDataEnvName 把环境名还原成"目录名"（目前就是原样，保留这个函数是为了让
// 「环境 → 目录名」的换算只有一处；将来若要改（例如大小写不敏感的文件系统），
// 改这里一处就够）。
func DevDataEnvName(env string) string { return strings.TrimSpace(env) }

// DevDataEnvExists 某环境的开发资料目录是否存在（status 判定"未拉取"用）。
func DevDataEnvExists(dir string) bool {
	if dir == "" {
		return false
	}
	st, err := os.Stat(dir)
	return err == nil && st.IsDir()
}
