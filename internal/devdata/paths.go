package devdata

// 目录布局与"先写临时目录、成功再原子改名"的落盘纪律（plan §5.4 / §10）。
//
// 两条不变量：
//
//  1. **路径只由 config.DevData* 拼**。调用方不许自己 filepath.Join 出这些路径 ——
//     否则 CLI 与 Web 迟早各算一份，表现为"设置页里看得到、命令行说没拉"。
//     所以本文件里的函数一律**接收 cfgPath**，内部走 config.DevData*ForEnv（含环境级覆盖）。
//  2. **失败/取消不留半套**。写到 `<目标>.tmp-<rand>`，全部成功后再 rename 成正式目录；
//     取消或出错就删掉临时目录。半套材料比没有材料更坏：model 会拿它算出"看起来对"的结论。

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/luyaogod/TT/internal/config"
)

// 领域层错误：Web 层据此映射 HTTP 码，CLI 据此映射退出码。用哨兵错误 + %w 包装，
// 调用方一律 errors.Is 判，不靠文案匹配。
var (
	// ErrEnvNotFound 环境不在 config.json 的 hosts.sshs 清单里 → 404。
	ErrEnvNotFound = errors.New("环境不存在")
	// ErrProgNotFound 没有这个程序（服务器上没有 / 本地没拉过）→ 404。
	//
	// 两种情况共用同一个哨兵：对调用方来说都是"你要的这个程序拿不到材料"，
	// 区别写在包装后的文案里（fetch 说"服务器上没有"，check 说"本地没有"）。
	ErrProgNotFound = errors.New("该程序不存在")
	// ErrMaterialIncomplete 程序存在，但本地这套材料缺件 —— 最典型的是从未客制过的
	// 标准程序：服务器上端到端就没有 .tap/.tsd（只有 .tab + .4gl）。
	// 对外同样是"拿不到材料"（HTTP 404 / 退出码 4），但文案必须说清"是缺件、
	// 不是没这个程序"，否则会被误读成程序编号打错。
	ErrMaterialIncomplete = errors.New("材料不成套")
	// ErrPointNotFound 点名不存在（程序与材料都正常，只是查不到这个点）→ 404。
	ErrPointNotFound = errors.New("没有匹配的点位")
	// ErrInvalidProg 程序名非法 → 400。
	ErrInvalidProg = errors.New("程序名非法")	// ErrInvalidEnv 环境名非法（空/含路径分隔符等）→ 400。
	ErrInvalidEnv = errors.New("环境名非法")
	// ErrExists 目标已存在且未给 force → 409 语义（由调用方决定 HTTP 码）。
	ErrExists = errors.New("目标已存在")
	// ErrBusy 同 (env, kind) 已有任务在跑 → 409。
	ErrBusy = errors.New("已有同类任务在执行")
	// ErrNoJob 没有可取消的任务 → 409。
	ErrNoJob = errors.New("没有正在执行的任务")
)

// reProg 程序名白名单（plan §5.3）：小写字母开头，后跟小写字母/数字/下划线，3–32 位。
//
// 为什么这么紧：程序名会被拼进**本地路径**与**服务器路径**。`../` 这种输入一旦进到
// join 里就是越界读写；大写与中文在 T100 实际也不存在（程序编号全小写）。
// 3 位下限是因为最短的合法程序编号就是 3 段式（如 aap）。
var reProg = regexp.MustCompile(`^[a-z][a-z0-9_]{2,31}$`)

// ValidProgName 校验程序名。返回的错误已包 ErrInvalidProg，可直接给用户看。
func ValidProgName(prog string) error {
	p := strings.TrimSpace(prog)
	if p == "" {
		return fmt.Errorf("%w: 程序名不能为空", ErrInvalidProg)
	}
	if !reProg.MatchString(p) {
		return fmt.Errorf("%w: %q（应为小写字母开头、仅含小写字母/数字/下划线、3–32 位，如 axmt500_wf）",
			ErrInvalidProg, prog)
	}
	return nil
}

// envInfo 取环境配置（不存在时报 ErrEnvNotFound）。空名字也走这条。
func envInfo(cfgPath, env string) (*config.NamedSsh, error) {
	name := strings.TrimSpace(env)
	if name == "" {
		return nil, fmt.Errorf("%w: 未指定环境（用 --at <环境> 或设置页选择）", ErrInvalidEnv)
	}
	if err := config.ValidateEnvName(name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidEnv, err)
	}
	root, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}
	e := root.Hosts.ByName(name)
	if e == nil {
		return nil, fmt.Errorf("%w: %q（可用环境：%s）", ErrEnvNotFound, name, strings.Join(root.Hosts.Names(), ", "))
	}
	return e, nil
}

// EnvDir 某环境的开发资料目录（已校验环境存在）。
func EnvDir(cfgPath, env string) (string, error) {
	e, err := envInfo(cfgPath, env)
	if err != nil {
		return "", err
	}
	dir := config.DevDataEnvDirForEnv(cfgPath, e.Name)
	if dir == "" {
		return "", fmt.Errorf("无法确定 %q 的开发资料目录（配置里没有数据目录？）", e.Name)
	}
	return dir, nil
}

// ProgDir 某程序的材料目录（已校验环境与程序名）。
func ProgDir(cfgPath, env, prog string) (string, error) {
	if err := ValidProgName(prog); err != nil {
		return "", err
	}
	dir, err := EnvDir(cfgPath, env)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, config.DevDataSubdirProg, strings.TrimSpace(prog)), nil
}

// DictPath 某环境规格库文件（已校验环境存在）。
func DictPath(cfgPath, env string) (string, error) {
	if _, err := envInfo(cfgPath, env); err != nil {
		return "", err
	}
	return config.DevDataDictPathForEnv(cfgPath, env), nil
}

// MdlDir 某环境模板树根（已校验环境存在）。
func MdlDir(cfgPath, env string) (string, error) {
	if _, err := envInfo(cfgPath, env); err != nil {
		return "", err
	}
	return config.DevDataMdlDirForEnv(cfgPath, env), nil
}

// ProgRoot 某环境的程序目录根（…/prog）。
func ProgRoot(cfgPath, env string) (string, error) {
	dir, err := EnvDir(cfgPath, env)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, config.DevDataSubdirProg), nil
}

// 临时目录命名：`<目标>.tmp-<随机>`。放在**同一个父目录**下 —— 只有同一文件系统内
// 的 rename 才是原子的；落到 %TEMP% 再跨盘搬就不是了。
const tmpSuffixPrefix = ".tmp-"

// NewTempDir 为 final 造一个临时目录（不创建；调用方 os.MkdirAll）。
//
// 用 os.MkdirTemp 而不是自拼随机数：并发两个任务的随机后缀撞车时，MkdirTemp 会重试，
// 自拼的不会。
func NewTempDir(final string) (string, error) {
	parent := filepath.Dir(final)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", fmt.Errorf("创建父目录失败 (%s): %w", parent, err)
	}
	base := filepath.Base(final)
	dir, err := os.MkdirTemp(parent, base+tmpSuffixPrefix)
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败 (%s): %w", parent, err)
	}
	return dir, nil
}

// CommitDir 把临时目录换成正式目录：先删旧的（force 时），再 rename。
//
// 顺序是刻意的：**先确认临时目录里的东西齐了再动旧的**。反过来（先删旧目录、
// 再从临时目录拷）中途失败就是"旧材料没了、新的也没到位"。
//
// force=false 且目标已存在 → ErrExists（由调用方决定是 409 还是提示加 --force）。
func CommitDir(tmp, final string, force bool) error {
	if _, err := os.Stat(final); err == nil {
		if !force {
			return fmt.Errorf("%w: %s（加 --force 覆盖）", ErrExists, final)
		}
		// 旧目录删不掉时**不继续**：rename 到已存在的非空目录在 Windows 上会失败，
		// 在 Unix 上会失败或把新目录塞进去 —— 两种都不是我们想要的，不如当场报清楚。
		if err := os.RemoveAll(final); err != nil {
			return fmt.Errorf("删除旧目录失败 (%s): %w", final, err)
		}
	}
	if err := os.Rename(tmp, final); err != nil {
		return fmt.Errorf("改名失败 (%s → %s): %w", tmp, final, err)
	}
	return nil
}

// CleanupTemp 失败/取消时清掉临时目录；顺带扫掉同前缀的历史残留（进程被 kill 时留下的）。
//
// 残留扫描是"取消验证"的一部分（plan §5.5 验收：`无 .tmp-* 残留`）——
// 只删自己那一个的话，上一次崩溃留下的空壳会一直在 status 里出现，让人以为还有任务在跑。
func CleanupTemp(tmp, final string) {
	if tmp != "" {
		_ = os.RemoveAll(tmp)
	}
	if final == "" {
		return
	}
	parent := filepath.Dir(final)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return
	}
	prefix := filepath.Base(final) + tmpSuffixPrefix
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			_ = os.RemoveAll(filepath.Join(parent, e.Name()))
		}
	}
}

// TempDirs 列出某目录下所有 `<前缀>.tmp-*` 残留（status 里报告"未完成"用）。
func TempDirs(dir, base string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	prefix := base + tmpSuffixPrefix
	var out []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out
}

// nowStamp 统一的 manifest 时间戳形态（RFC3339，带本地时区）。
//
// 带时区是必须的：服务器在本机之外的时区时，只写"2026-11-21 14:30"会让 stale 判定
// 与 srcMtime 比较出错，而错的方向是"看起来还新"。
func nowStamp() string { return time.Now().Format(time.RFC3339) }
