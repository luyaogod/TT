package debug

// 服务测试（`/api/wstest`）本批已**从产品里取消**（仓库主人定性：它不是标准接口，
// 是临时测试工具 —— 所以是去掉端点，不是给它补闸门）。整个 `wstest.go` 连同
// `WSDefaultURLFor` / `WSTest` / `soapEnvelope` / `stripWSDLSuffix` / `wsEndpoints` 一起删了。
//
// **`wsURLPrefix` 必须留下**：它不是那条端点的私有件 —— `/api/status` 的 `wsPrefix`
// 字段（前端拼占位符地址用）就是它，与 `/api/wstest` 在不在无关。
// 单独抽到本文件，是为了让"这条规则只剩这一份实现"这件事在删掉 wstest.go 之后
// 仍然看得见（它的注释是从原文件逐字搬过来的，出处仍然准确）。
// 于是现在只有 `/api/status` 一处调用方。

// wsURLPrefix 是服务地址的前缀：出貨區(topprd)走 `/ws`，其余区走 `/w`。
//
// 单独抽出来是给 `/api/status` 用的（前端要拿它拼占位符地址）—— 这条规则只有这一份实现，
// 两边各写一遍的后果是"界面上的默认地址与真正发出去的地址不一致"（实测 zone=t 时
// 前端给 /wtopprd、后端给 /wstopprd，而这种偏差在本机 zone=36 时看不出来）。
//
// ⚠ 入参是**站点配置里的区域码**（菜单选项号，如 35/36/t），不是 ZONE 字符串。
// 判据与 `host.ZoneTNSName` **不是同一件事，别再合回去**：那张表是「猜 TNS 别名」的
// 兜底表，只覆盖旧式菜单码(1/2/3/4)、对不上返回空串；出货区判定要认**站点自己的
// 菜单码** —— 本机出貨區的菜单码是 `t`（ZONE=topprd），老站点是 `1`。
// 混用会让 zone=t 的站点拿到 `/w`（前端拼出 /wtopprd，而后端要 /wstopprd）。
func wsURLPrefix(zone string) string {
	switch zone {
	case "t", "1", "topprd": // 本机菜单码 / 旧式菜单码 / 已经是 ZONE 字符串
		return "/ws"
	}
	return "/w"
}
