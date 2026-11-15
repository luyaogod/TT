package debug

// 集成服务检测的 URL 规则与 `/api/status` 的 `wsPrefix` 共用这一份实现。
//
// 历史：`/api/wstest`（`WSTest`/`WSDefaultURLFor`/`soapEnvelope`/`stripWSDLSuffix`/
// `wsEndpoints`）曾在 f264ccc 整批删除；本批以 `/api/wslogs/test` 重新引入，
// 但补上了当时的缺口：URL/方法/报文全部先校验、请求体走 stdin 不落临时文件、
// 不写 `wsfa_t`、把"非 200 的正文"保留下来（原生会丢弃）。
// `wsURLPrefix` 仍是唯一实现：`/api/status` 的 `wsPrefix` 与检测默认地址都用它。

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
