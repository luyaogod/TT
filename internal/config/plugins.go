// 插件落点：与其它自有存储一样，路径只在 config.Locations 派生一次。
//
// 为什么不让 internal/plugin 自己 filepath.Join：那就是第二份位置解析，
// 某天两边会悄悄分叉（历史上"服务写一处、命令行读另一处"正是这么来的）。
package config

import (
	"path/filepath"
	"strings"
)

// 插件子目录名（<数据目录>/plugins/ 之下）。
const (
	PluginRegistryFileName = "registry.json"
	pluginInstalledDirName = "installed"
	pluginDataDirName      = "data"
	pluginRunDirName       = "run"
	pluginLogsDirName      = "logs"
	pluginCacheDirName     = "cache"
)

// PluginsDir 插件根目录 <数据目录>/plugins。
func (l *Locations) PluginsDir() string {
	if dir := l.DataDir(); dir != "" {
		return filepath.Join(dir, "plugins")
	}
	return ""
}

// PluginRegistryPath 插件注册表 registry.json。
func (l *Locations) PluginRegistryPath() string {
	if dir := l.PluginsDir(); dir != "" {
		return filepath.Join(dir, PluginRegistryFileName)
	}
	return ""
}

// PluginInstalledRoot 全部插件载荷的父目录。
func (l *Locations) PluginInstalledRoot() string {
	return pluginSubdir(l, pluginInstalledDirName)
}

// PluginVersionDir 某插件某版本的**不可变**程序目录。
func (l *Locations) PluginVersionDir(id, version string) string {
	root := l.PluginInstalledRoot()
	if root == "" || !safePathSegment(id) || !safePathSegment(version) {
		return ""
	}
	return filepath.Join(root, id, version)
}

// PluginDataDir 插件持久数据目录（升级/卸载程序默认不动它）。
func (l *Locations) PluginDataDir(id string) string {
	root := pluginSubdir(l, pluginDataDirName)
	if root == "" || !safePathSegment(id) {
		return ""
	}
	return filepath.Join(root, id)
}

// PluginRunDir 插件运行态目录（实例状态、就绪信息、临时凭证引用）。
func (l *Locations) PluginRunDir(id string) string {
	root := pluginSubdir(l, pluginRunDirName)
	if root == "" || !safePathSegment(id) {
		return ""
	}
	return filepath.Join(root, id)
}

// PluginLogsDir 插件日志目录。
func (l *Locations) PluginLogsDir(id string) string {
	root := pluginSubdir(l, pluginLogsDirName)
	if root == "" || !safePathSegment(id) {
		return ""
	}
	return filepath.Join(root, id)
}

// PluginCacheDir 下载与解包缓存（安装失败留下的残余都在这里，可直接清）。
func (l *Locations) PluginCacheDir() string {
	return pluginSubdir(l, pluginCacheDirName)
}

// pluginSubdir 拼 <数据目录>/plugins/<name>。
func pluginSubdir(l *Locations, name string) string {
	dir := l.PluginsDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, name)
}

// safePathSegment 插件 ID/版本会直接当目录名用，先挡掉能逃出父目录的写法。
//
// 校验层（internal/plugin）另有更严的 ID/版本规则；这里只是路径拼装的最后一道门 ——
// 万一某天有个调用方绕过了校验，也不至于把目录拼到插件根之外。
func safePathSegment(s string) bool {
	if s == "" || s == "." || s == ".." {
		return false
	}
	if strings.ContainsAny(s, `/\:`) {
		return false
	}
	return !strings.Contains(s, "..")
}
