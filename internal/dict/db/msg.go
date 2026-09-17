package db

import "fmt"

// 系统消息档查询(gzze_t):T100 所有提示/报错消息由作业 azzi920 维护,
// 运行时 cl_err/cl_getmsg 按 (编号, 语言) 精确取用。消息编号(gzze001)形如
// std-00001 / azz-00041 / lib-xxxxx / -263(负整数 SQLCODE),语言(gzze002)为
// zh_CN/zh_TW/en_US 等。作业名称辅助表 gzzal_t 未同步时静默降级(名称留空)。

// MsgRow 一条消息档记录(编号+语言唯一)。
type MsgRow struct {
	Code     string `json:"编号"`
	Lang     string `json:"语言"`
	Text     string `json:"文本"`   // gzze003 消息文本
	Action   string `json:"建议处理"` // gzze004 建议处理方式
	ExecProg string `json:"建议作业"` // gzze005 建议执行作业编号(:EXEPROG=无)
	ProgName string `json:"作业名称"` // gzzal003(该语言名称;gzzal_t 未同步时为空)
	Detail   string `json:"技术细节"` // gzze006 程式人员详细讯息
	TypeCode string `json:"类型码"`  // gzze007 讯息类型: 0警告/1错误/2资讯 (SCC 106)
	ForceWin string `json:"强制开窗"` // gzze008 Y/N
	Status   string `json:"状态"`   // gzzestus Y=启用/N=停用
}

// QueryMsg 返回指定消息编号的全部语言行(按 gzze002 排序)。
// gzze_t 缺表时返回原错误(命令层用 IsMissingTable 提示先 sync);
// gzzal_t 缺表/查询失败仅导致作业名称为空,不阻断主结果。
func (d *DB) QueryMsg(code string) ([]MsgRow, error) {
	rows, err := d.conn.Query(`
		SELECT COALESCE(gzze001, ''), COALESCE(gzze002, ''), COALESCE(gzze003, ''),
		       COALESCE(gzze004, ''), COALESCE(gzze005, ''), COALESCE(gzze006, ''),
		       COALESCE(gzze007, ''), COALESCE(gzze008, ''), COALESCE(gzzestus, '')
		FROM gzze_t WHERE gzze001 = ? ORDER BY gzze002`, code)
	if err != nil {
		return nil, fmt.Errorf("query msg %s: %w", code, err)
	}
	defer rows.Close()

	var result []MsgRow
	progs := map[string]bool{}
	for rows.Next() {
		var r MsgRow
		if err := rows.Scan(&r.Code, &r.Lang, &r.Text, &r.Action, &r.ExecProg,
			&r.Detail, &r.TypeCode, &r.ForceWin, &r.Status); err != nil {
			return nil, fmt.Errorf("scan msg row: %w", err)
		}
		if r.ExecProg != "" && r.ExecProg != ":EXEPROG" {
			progs[r.ExecProg] = true
		}
		result = append(result, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return result, nil
	}

	// 作业名称:一次性查 gzzal_t(该语言的作业名称),缺表静默降级
	names, err := d.msgProgNames(progs)
	if err != nil {
		names = map[string]string{} // gzzal_t 未同步/缺失:名称留空
	}
	for i := range result {
		result[i].ProgName = names[result[i].ExecProg+"\x00"+result[i].Lang]
	}
	return result, nil
}

// msgProgNames 查作业多语言名称:key = "<作业号>\x00<语言>"。
func (d *DB) msgProgNames(progs map[string]bool) (map[string]string, error) {
	if len(progs) == 0 {
		return map[string]string{}, nil
	}
	ids := make([]any, 0, len(progs))
	ph := ""
	for p := range progs {
		ids = append(ids, p)
		ph += "?,"
	}
	ph = ph[:len(ph)-1]
	rows, err := d.conn.Query(`
		SELECT gzzal001, COALESCE(gzzal002, ''), COALESCE(gzzal003, '')
		FROM gzzal_t WHERE gzzal001 IN (`+ph+`)`, ids...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var prog, lang, name string
		if err := rows.Scan(&prog, &lang, &name); err != nil {
			return nil, err
		}
		out[prog+"\x00"+lang] = name
	}
	return out, rows.Err()
}
