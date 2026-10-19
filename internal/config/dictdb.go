package config

// 本地字典库的落点：**一个环境一个 SQLite 文件**。
//
// 以前只有一个 <数据目录>/erp_data.db，两个环境先后同步会写进同一个文件 ——
// 查「本地」时可能读到另一个客户的数据，而表现和「数据不存在」一模一样。
// 分文件之后，同步/查询/状态/过期判断都按环境各管各的，结构上不可能串。
//
// 旧文件 erp_data.db **不再默认使用**：它没有环境标记，拆不出来。显式 -d/--db
// 仍可打开它做对照。

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// DictDBDirName 是数据目录下存放各环境本地字典库的子目录。
const DictDBDirName = "dict"

var errLocalDBNotFound = errors.New("本地数据库文件未找到")

// DictDBDir 返回本地字典库目录（数据目录下 dict/）。
func DictDBDir(dataDir string) string {
	if strings.TrimSpace(dataDir) == "" {
		return ""
	}
	return filepath.Join(dataDir, DictDBDirName)
}

// DictDBFileName 把环境名变成安全的文件名：Windows 不允许的字符换成 _，
// 去掉首尾的点与空格；名字被改写过时追加短哈希，避免“A/B”与“A_B”撞车。
func DictDBFileName(env string) string {
	name := strings.TrimSpace(env)
	safe := strings.Map(func(r rune) rune {
		switch {
		case r == 92, r == '/', r == ':', r == '*', r == '?', r == '"', r == '<', r == '>', r == '|', r == 0:
			return '_'
		}
		if r < 0x20 {
			return '_'
		}
		return r
	}, name)
	safe = strings.Trim(safe, ". ")
	if safe == "" {
		safe = "env"
	}
	if safe != name && name != "" {
		sum := sha1.Sum([]byte(name))
		safe += "-" + hex.EncodeToString(sum[:])[:8]
	}
	return safe + ".db"
}

// DictDBPath 某环境的本地字典库文件：<数据目录>/dict/<环境>.db。
func DictDBPath(dataDir, env string) string {
	dir := DictDBDir(dataDir)
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, DictDBFileName(env))
}

// DictDBPathIn 在给定目录下拼某环境的库文件（设置页与同步任务共用同一份命名）。
func DictDBPathIn(dir, env string) string {
	if strings.TrimSpace(dir) == "" {
		return ""
	}
	return filepath.Join(dir, DictDBFileName(env))
}

// DictDBDir 数据目录下的本地字典库目录。
func (l *Locations) DictDBDir() string { return DictDBDir(l.DataDir()) }

// DictDBPath 某环境的本地字典库文件。
func (l *Locations) DictDBPath(env string) string { return DictDBPath(l.DataDir(), env) }

// DictDBCandidatesFor 某环境本地库的候选：-d/--db → <数据目录>/dict/<环境>.db。
//
// **不认 TDICT_DB**：那是旧版“一个共享库”的全局开关，留着它会让一个环境变量
// 悄悄覆盖所有环境（隔离失效）。点名了 local:<环境> 就只看这个环境的文件；
// 确实要打开旧共享库，用 -d/--db 或裸 local + TDICT_DB。
func (l *Locations) DictDBCandidatesFor(env, flagPath string) []string {
	var out []string
	if p := AbsPath(flagPath); p != "" {
		out = append(out, p)
	}
	if p := l.DictDBPath(env); p != "" {
		out = append(out, p)
	}
	return out
}

// ResolveDictDBFor 按环境定位本地库：返回第一个存在的候选；都缺时报错并列出尝试过的路径。
func (l *Locations) ResolveDictDBFor(env, flagPath string) (string, []string, error) {
	var tried []string
	for _, p := range l.DictDBCandidatesFor(env, flagPath) {
		abs, _ := filepath.Abs(p)
		if _, err := os.Stat(abs); err == nil {
			return abs, tried, nil
		}
		tried = append(tried, abs)
	}
	return "", tried, errLocalDBNotFound
}
