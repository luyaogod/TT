package debug

// 集成服务注册表查询（服务名称开窗 / 服务编号查询）。
//
// 口径来自 T100 q_gzja005_1（com/qry/4gl/q_gzja005_1.4gl:586-651）：
//   gzja_t LEFT JOIN gzjal_t（语言别） UNION ALL gzjc_t LEFT JOIN gzjcl_t
//   三列 = 服务规格编号 gzja001 / 服务名称 gzjal003 / API服务名称 gzja004
//   ORDER BY gzja001
// 这两张表都在系统账号 ds 下，与 wsfa_t 同源；不按企业分 schema。
//
// 与 T100 的差异（有意）：
//   - T100 的模糊开关 A-SYS-0093 只影响"是否自动补 *"；这里统一按"包含"匹配
//     （输入不带通配符时前后补 %），因为开窗的第一诉求是"记得半个名字也能找到"。
//   - T100 用 ROWNUM 包三层做分页；这里用 OFFSET/FETCH（Oracle）与 LIMIT/OFFSET（金仓），
//     与本仓 wsfa_t 列表同一写法。

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/luyaogod/TT/internal/host"
)

const (
	wsServiceLang     = "zh_CN"
	wsServicePageMin  = 10
	wsServicePageMax  = 500
	wsServicePageSize = 200
)

// WSServiceItem 服务注册表一行（开窗三列）。
type WSServiceItem struct {
	Spec string `json:"spec"`    // gzja001 服务规格编号（如 wssp000）
	Name string `json:"name"`    // gzjal003 服务名称（多语言）
	API  string `json:"apiName"` // gzja004 API服务名称（回传值，如 easy.query）
}

// WSServiceFilter 服务名称开窗的过滤条件。
type WSServiceFilter struct {
	Spec     string // 服务规格编号
	Name     string // 服务名称
	API      string // API服务名称
	Page     int
	PageSize int
}

// wsServiceLikePattern 把一个开窗过滤值翻成 LIKE 模式。
// ok=false 表示值里有不能安全内联的字符（单引号/反斜杠/控制字符），调用方应退化为空集。
//
// 规则：
//   - 用户输入 * / ? 按通配处理（与 T100 的 CONSTRUCT 习惯一致）；
//   - 不带通配符时前后补 %（T100 的模糊查询意图：包含匹配）；
//   - 字面 % / _ 先转义，避免用户输入被当成通配符；
//   - 统一走 ESCAPE '\'。
func wsServiceLikePattern(raw string) (string, bool) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", true
	}
	// 单引号是字符串逃逸；反斜杠是 ESCAPE 字符；控制字符只可能来自异常输入。
	if strings.ContainsAny(v, "'\\") || strings.ContainsFunc(v, func(r rune) bool { return r < 0x20 || r == 0x7f }) {
		return "", false
	}
	var b strings.Builder
	for _, r := range v {
		switch r {
		case '%':
			b.WriteString(`\%`)
		case '_':
			b.WriteString(`\_`)
		default:
			b.WriteRune(r)
		}
	}
	pat := b.String()
	if strings.ContainsAny(pat, "*?") {
		pat = strings.ReplaceAll(pat, "*", "%")
		pat = strings.ReplaceAll(pat, "?", "_")
		return pat, true
	}
	return "%" + pat + "%", true
}

// wsServiceWhere 拼开窗的 WHERE 子句（纯函数，便于单测）。
// 三个条件都是 AND；任一条非法输入 → AND 1=0（与 wslogWhere 的退化口径一致）。
func wsServiceWhere(f WSServiceFilter) string {
	wc := "1=1"
	for _, c := range []struct {
		val, col string
	}{
		{f.Spec, "svc_spec"},
		{f.Name, "svc_name"},
		{f.API, "svc_api"},
	} {
		if strings.TrimSpace(c.val) == "" {
			continue
		}
		pat, ok := wsServiceLikePattern(c.val)
		if !ok {
			wc += " AND 1=0"
			continue
		}
		// UPPER 是为了让编码/接口名大小写不敏感；中文不受影响。
		wc += fmt.Sprintf(" AND UPPER(%s) LIKE UPPER('%s') ESCAPE '\\'", c.col, strings.ReplaceAll(pat, "'", "''"))
	}
	return wc
}

// wsServiceFrom 服务注册表的统一子查询（两段 UNION ALL，列名统一成 svc_*）。
//
// 第二段的 API 列取 gzjc001 而不是 gzjcl001：LEFT JOIN 在没有对应语言行时
// gzjcl001 会是 NULL，而 gzjc001 一定在。
func wsServiceFrom() string {
	return `(SELECT gzja001 AS svc_spec, gzjal003 AS svc_name, gzja004 AS svc_api
FROM gzja_t LEFT JOIN gzjal_t ON gzja001 = gzjal001 AND gzjal002 = '` + wsServiceLang + `'
WHERE 1=1
UNION ALL
SELECT gzjc002 AS svc_spec, gzjcl003 AS svc_name, gzjc001 AS svc_api
FROM gzjc_t LEFT JOIN gzjcl_t ON gzjc001 = gzjcl001 AND gzjcl002 = '` + wsServiceLang + `'
WHERE 1=1) t`
}

// wsServiceDataSQL / wsServiceCountSQL 纯函数，便于单测钉分页方言。
func wsServiceDataSQL(kingbase bool, where string, page, size int) string {
	n := (page - 1) * size
	if kingbase {
		return fmt.Sprintf(`SELECT svc_spec||'|'||coalesce(svc_name,'')||'|'||coalesce(svc_api,'') FROM %s WHERE %s ORDER BY svc_spec OFFSET %d LIMIT %d`, wsServiceFrom(), where, n, size)
	}
	return fmt.Sprintf(`SELECT svc_spec||'|'||coalesce(svc_name,'')||'|'||coalesce(svc_api,'') FROM %s WHERE %s ORDER BY svc_spec OFFSET %d ROWS FETCH FIRST %d ROWS ONLY`, wsServiceFrom(), where, n, size)
}

func wsServiceCountSQL(where string) string {
	return fmt.Sprintf(`SELECT COUNT(1) FROM %s WHERE %s`, wsServiceFrom(), where)
}

// parseWSServiceRows 解析 "spec|name|api" 行（纯函数，便于单测）。
func parseWSServiceRows(out string) []WSServiceItem {
	var items []WSServiceItem
	for _, ln := range strings.Split(out, "\n") {
		ln = strings.TrimRight(ln, " \r")
		if ln == "" {
			continue
		}
		parts := strings.SplitN(ln, "|", 3)
		if len(parts) < 2 {
			continue
		}
		it := WSServiceItem{Spec: strings.TrimSpace(parts[0])}
		it.Name = strings.TrimSpace(parts[1])
		if len(parts) == 3 {
			it.API = strings.TrimSpace(parts[2])
		}
		if it.Spec == "" && it.Name == "" && it.API == "" {
			continue
		}
		items = append(items, it)
	}
	return items
}

// parseWSServiceCount 取 COUNT(1) 的第一行数字。
func parseWSServiceCount(out string) int {
	for _, ln := range strings.Split(out, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		if n, err := strconv.Atoi(ln); err == nil {
			return n
		}
	}
	return 0
}

// wsServiceAcct 说明查服务注册表用的账号（与 wslogAcct 同形，前端可直接复用展示）。
func wsServiceAcct() WSLogAcct {
	return WSLogAcct{
		Account: wsSysAccount,
		Reason:  "gzja_t/gzjc_t 是系统账号 ds 下的集成服务注册表，不按企业分 schema",
	}
}

// listWSServices 查服务注册表（Oracle: OFFSET/FETCH；金仓: LIMIT/OFFSET）。
func listWSServices(conn *host.SSHConn, dbc *dbRun, f WSServiceFilter) (items []WSServiceItem, total int, hasMore bool, err error) {
	size := f.PageSize
	if size < wsServicePageMin || size > wsServicePageMax {
		size = wsServicePageSize
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	where := wsServiceWhere(f)
	connStr, err := dbc.dbConnStr(wsSysAccount)
	if err != nil {
		return nil, 0, false, err
	}
	kingbase := dbc.conn.Type == "kingbase"
	var countOut, dataOut string
	if kingbase {
		countOut, err = dbc.exec(conn, connStr, "", wsServiceCountSQL(where), 30*time.Second)
		if err != nil {
			return nil, 0, false, fmt.Errorf("查询服务注册表失败: %w (%s)", err, host.FirstLines(countOut, 3))
		}
		dataOut, err = dbc.exec(conn, connStr, "", wsServiceDataSQL(true, where, page, size), 40*time.Second)
	} else {
		pre := "set heading off\nset feedback off\nset pagesize 0\nset trimspool on\nset linesize 32767\n"
		countOut, err = dbc.exec(conn, connStr, pre+wsServiceCountSQL(where)+";", "", 30*time.Second)
		if err != nil {
			return nil, 0, false, fmt.Errorf("查询服务注册表失败: %w (%s)", err, host.FirstLines(countOut, 3))
		}
		dataOut, err = dbc.exec(conn, connStr, pre+wsServiceDataSQL(false, where, page, size)+";", "", 40*time.Second)
	}
	if err != nil {
		return nil, 0, false, fmt.Errorf("查询服务注册表失败: %w (%s)", err, host.FirstLines(dataOut, 3))
	}
	for _, ln := range strings.Split(countOut+"\n"+dataOut, "\n") {
		if isDBErrorLine(ln) {
			return nil, 0, false, fmt.Errorf("服务注册表查询出错: %s", strings.TrimSpace(ln))
		}
	}
	total = parseWSServiceCount(countOut)
	items = parseWSServiceRows(dataOut)
	hasMore = page*size < total
	if len(items) > size {
		items = items[:size]
	}
	return items, total, hasMore, nil
}
