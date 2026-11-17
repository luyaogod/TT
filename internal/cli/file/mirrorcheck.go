package file

// 镜像护栏：`tt file get` 在下盘之前先问一句"镜像里是不是已经有同一份"。
//
// 为什么值得拦：镜像保留服务器的 mtime，所以 size + mtime 两项就能判"同一份"，
// 不必为了算 hash 先把文件下下来（那正好与"别白下"自相矛盾）。实测三个 4gl
// （966KB / 1.18MB / 2.26MB）本地镜像里都有、hash 与远端一致，却被白下了一遍。

import (
	"os"
	"path/filepath"
	"time"
)

// mirrorState 镜像里那一份与远端的关系。
type mirrorState int

const (
	// mirrorNone 镜像里没有这个文件（或该环境没配镜像目录）：按正常下载处理。
	mirrorNone mirrorState = iota
	// mirrorSame 同一份：size 相等且 mtime 一致 —— 下载只会得到已有的字节。
	mirrorSame
	// mirrorStale 有同名文件但内容对不上：照下，同时提示镜像落后。
	mirrorStale
)

// mirrorCompare 用远端的 size/mtime 比一次镜像文件（不发 SFTP）。
//
// mirrorFile 为空（未配镜像 / 路径不在 $TOP 内）时返回 mirrorNone。
func mirrorCompare(mirrorFile string, remoteSize int64, remoteMod time.Time) (mirrorState, string) {
	if mirrorFile == "" {
		return mirrorNone, ""
	}
	st, err := os.Stat(mirrorFile)
	if err != nil || !st.Mode().IsRegular() {
		return mirrorNone, mirrorFile
	}
	if st.Size() != remoteSize {
		return mirrorStale, mirrorFile
	}
	// SFTP 的 mtime 只到秒，镜像又保留服务器时间戳；2s 容差吸收取整与时钟差。
	d := st.ModTime().Sub(remoteMod)
	if d < 0 {
		d = -d
	}
	if d <= 2*time.Second {
		return mirrorSame, mirrorFile
	}
	return mirrorStale, mirrorFile
}

// mirrorFileFor 把 $TOP 内的相对路径映射到本环境镜像里的文件路径。
// 未配镜像目录时 ok=false —— 调用方据此静默跳过护栏（没镜像就没有可指的落点）。
func (t *target) mirrorFileFor(rel string) (string, bool) {
	if t == nil || t.cfgPath == "" || rel == "" {
		return "", false
	}
	dir, ok := readMirrorDir(t.cfgPath)
	if !ok || dir == "" {
		return "", false
	}
	envDir := t.ssh.MirrorPath(dir)
	if envDir == "" {
		return "", false
	}
	return filepath.Join(envDir, filepath.FromSlash(rel)), true
}
