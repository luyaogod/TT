// 虚拟应用的夹具:静态编辑器内容、状态机用的会话/日志/设置状态。
// 数据本身尽量复用 lib/fixtures.ts(与台账屏同一份真实文案),这里只补虚拟应用独有的部分。

export type Tok = 'kw' | 'str' | 'cmt' | 'fn' | 'num' | 'plain'

/** 编辑器静态示意:只有布局/密度/行高是决策对象,语法色仍归 Monaco 主题(此处仅示意) */
export const CODE_LINES: { n: number; parts: [Tok, string][] }[] = [
  { n: 402, parts: [['kw', 'FUNCTION'], ['fn', ' wf_confirm()']] },
  { n: 403, parts: [['kw', '  DEFINE'], ['plain', ' lp_str    LIKE wo_m.str,']] },
  { n: 404, parts: [['plain', '         li_cnt    '], ['kw', 'INTEGER']] },
  { n: 405, parts: [] },
  { n: 406, parts: [['cmt', '  # 校验工单是否存在(停站行在下面)']] },
  { n: 407, parts: [['kw', '  SELECT'], ['plain', ' '], ['fn', 'COUNT(*)'], ['kw', ' INTO'], ['plain', ' li_cnt']] },
  { n: 408, parts: [['kw', '    FROM'], ['plain', ' wo_m']] },
  { n: 409, parts: [['kw', '   WHERE'], ['plain', ' wo_m.ent = g_qryparam.parm1']] },
  { n: 410, parts: [['kw', '     AND'], ['plain', ' wo_m.str = lp_str']] },
  { n: 411, parts: [] },
  { n: 412, parts: [['kw', '  IF'], ['plain', ' li_cnt = '], ['num', '0'], ['kw', ' THEN']] },
  { n: 413, parts: [['kw', '    CALL'], ['fn', ' cl_err'], ['plain', ' ('], ['str', "'T100'"], ['plain', ', '], ['num', '1001'], ['plain', ', lp_str)']] },
  { n: 414, parts: [['kw', '    RETURN'], ['plain', ' '], ['kw', 'FALSE']] },
  { n: 415, parts: [['kw', '  END IF']] },
  { n: 416, parts: [] },
  { n: 417, parts: [['kw', '  LET'], ['plain', ' g_qryparam.parm3 = '], ['str', "'Y'"]] },
  { n: 418, parts: [['kw', '  RETURN'], ['plain', ' '], ['kw', 'TRUE']] },
  { n: 419, parts: [['kw', 'END FUNCTION']] },
]

export const BP_LINE = 412
export const CUR_LINE = 412

/** 会话状态机:虚拟应用可以在四种状态间切,看状态栏与控件可用性的变化 */
export const SESSION_STATES = [
  { key: 'stopped', label: '已停站', hint: '断点命中:变量/调用栈可读,可步进' },
  { key: 'running', label: '运行中', hint: '程序在 GDC 里跑:界面转观察' },
  { key: 'loading', label: '启动中', hint: '连接与登录:控件禁用' },
  { key: 'idle', label: '会话空闲', hint: '登录态保留,可直接启动下一轮' },
] as const

export type SessionState = (typeof SESSION_STATES)[number]['key']

/** 报文编辑器占位内容(与接口日志的 Request 页共用) */
export const PAYLOAD_LINES = [
  '{',
  '  "ent": "35",',
  '  "str": "WO-2026031-0007",',
  '  "plant": "P01",',
  '  "rows": [',
  '    { "item": "A-100231", "qty": 120, "unit": "PCS" },',
  '    { "item": "A-100232", "qty": 80,  "unit": "PCS" }',
  '  ],',
  '  "total": 2',
  '}',
]

/** 账号表的服务器侧验证状态(三种) */
export const ACCT_PROBE = [
  { schema: 'dsprd', state: 'ok' as const, note: '连接正常 ✓' },
  { schema: 'topprd', state: 'testing' as const, note: '验证中…' },
  { schema: 'report', state: 'err' as const, note: 'ORA-01017: invalid username/password' },
]

export const CACHE_TOTAL = { bytes: '56.6 MB', files: 1293 }
