package config

import "strings"

// KbPath 返回某环境生效的知识库目录：
//   - 环境自己配了 kbDir → 用它（绝对化）；
//   - 否则用顶层 kb.dir（绝对化）；
//   - 两处都没有 → 空串（调用方负责报"未设置"）。
//
// 与 MirrorPath 的关键差别：镜像目录**按环境拼**（root/<环境名>），知识库**不拼** ——
// 它是一份现成的目录，不是"根目录下按环境分出来的坑位"。要给某环境单独一套笔记，
// 就把那个目录写在它的 kbDir 上，别指望自动命名。
//
// 只做路径解析，不碰文件系统；存在性由调用方 os.Stat。
func (e *NamedSsh) KbPath(kbRoot string) string {
	if e != nil {
		if d := strings.TrimSpace(e.KbDir); d != "" {
			return AbsPath(d)
		}
	}
	if root := strings.TrimSpace(kbRoot); root != "" {
		return AbsPath(root)
	}
	return ""
}
