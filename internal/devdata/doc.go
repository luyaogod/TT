// Package devdata 准备 model（生成器语义层）需要的**原材料**：规格库、模板树、按程序的
// 整套设计档。它不读代码、不做解析、不产出任何语义结论 —— 只保证材料齐、格式对、指纹可信。
//
// 与读侧镜像的分工（这是本包存在的第一理由）：
//
//	tt file mirror   给"读代码"用 → srccache/，求"查得快、看得全"，会被 tt cache 清理
//	tt dict db sync  给"查字典"用 → dict/<env>.db，整库替换
//	tt gen mirror    给"model 备料"用 → devdata/<env>/，按程序成套、带 manifest 指纹
//
// 三条硬纪律：
//
//  1. **原样保真**：XML/文本按字节落盘，不转码、不重排、不解析（plan ADR-05）。
//  2. **来源只读**：SFTP 只读、数据库只读事务，服务器上不写一个文件（plan ADR-06）。
//  3. **不进缓存**：不登记进 pkg/cachedir 的白名单，清理只走 `tt gen clean`（ADR-02）。
//
// 目录布局见 internal/config/devdata.go 的包注释；manifest 契约见 manifest.go。
package devdata
