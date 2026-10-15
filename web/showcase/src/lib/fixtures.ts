// 演示数据:全部取自现有前端的真实文案与真实字段(见各 section 的 source 注记),
// 目的不是"像真的",而是让 UI 决策看到真实的内容长度、真实的中文标签与真实的状态分布。

// ---- 会话与站点 ----
export const ENVS = [
  { name: '36正式', zone: '36正式', port: 22, db: '金仓', active: true },
  { name: '35测试', zone: '35测试', port: 22, db: 'Oracle', active: false },
  { name: '31开发', zone: '31开发', port: 22, db: 'Oracle', active: false },
  { name: '(新环境)', zone: '', port: 0, db: '无库', active: false },
]

export const SSH_FORM = {
  name: '36正式',
  host: '10.20.36.11',
  port: '22',
  zone: '36正式',
  user: 't100prd',
  topent: '35',
  workspace: 'D:\\t100_wrok_dir\\客户\\prd',
}

export const DB_FORM = {
  type: 'kingbase',
  host: '10.20.36.12',
  port: '54321',
  dbname: 'topprd',
}

export const ACCOUNTS = [
  { schema: 'dsprd', password: '' },
  { schema: 'topprd', password: 'topprd_readonly' },
]

export const ENV_VARS = {
  topent: '35',
  note: '当前会话值(连接时已按环境配置下发)',
}

// ---- 调试会话 ----
export const SESSION = {
  env: '36正式',
  state: 'stopped' as const,
  stateLabel: '已停站',
  prog: 'bsft001_wf',
  file: 'bsft001_wf.4gl',
  line: 412,
  holdSeconds: 63,
}

export const STACK_FRAMES = [
  { idx: 0, func: 'bsft001_wf', file: 'bsft001_wf.4gl', line: 412 },
  { idx: 1, func: 'wf_confirm', file: 'bsft001_wf.4gl', line: 388 },
  { idx: 2, func: 'main', file: 'bsft001_wf.4gl', line: 120 },
]

export const AUTOVARS = [
  { expr: 'lp_str', value: 'WO-2026031-0007' },
  { expr: 'g_qryparam', value: 'g_qryparam.parm1 = "36"\ng_qryparam.parm2 = "\ng_qryparam.parm3 = "N"' },
  { expr: 'li_cnt', value: '12' },
]

export const WATCHES = [
  { expr: 'lp_str', value: 'WO-2026031-0007' },
  { expr: 'li_cnt', value: '12' },
  { expr: 'lp_bad', error: 'No symbol lp_bad in current context.' },
]

export const BREAKPOINTS = [
  { num: 1, func: 'bsft001_wf', file: 'bsft001_wf.4gl', line: 412, enabled: true, hit: true },
  { num: 2, func: 'wf_confirm', file: 'bsft001_wf.4gl', line: 388, enabled: true, hit: false },
  { num: 3, func: '', file: 'bsft001_wf.4gl', line: 120, enabled: false, hit: false },
]

export const VAR_TREE = [
  {
    name: 'g_qryparam',
    type: 'RECORD',
    open: true,
    children: [
      { name: 'parm1', value: '36', type: 'STRING' },
      { name: 'parm2', value: '', type: 'STRING' },
      { name: 'parm3', value: 'N', type: 'STRING' },
    ],
  },
  { name: 'lp_str', value: 'WO-2026031-0007', type: 'STRING' },
]

export const OUTLINE = [
  {
    label: 'MAIN',
    line: 12,
    depth: 0,
    children: [
      { label: 'DIALOG', line: 14, depth: 1, children: [
        { label: 'INPUT BY NAME', line: 18, depth: 2, children: [] },
        { label: 'BEFORE DIALOG', line: 44, depth: 2, children: [] },
        { label: 'AFTER INPUT', line: 96, depth: 2, children: [] },
      ] },
      { label: 'DISPLAY ARRAY', line: 118, depth: 1, children: [
        { label: 'BEFORE DISPLAY', line: 132, depth: 2, children: [] },
        { label: 'ON ACTION confirm', line: 168, depth: 2, children: [] },
      ] },
    ],
  },
  { label: 'FUNCTION wf_confirm()', line: 388, depth: 0, children: [] },
  { label: 'REPORT r_wo_list', line: 520, depth: 0, children: [
    { label: 'BEFORE GROUP OF', line: 536, depth: 1, children: [] },
    { label: 'ON EVERY ROW', line: 560, depth: 1, children: [] },
  ] },
]

export const SOURCE_SNIPPET = `FUNCTION wf_confirm()
  DEFINE lp_str    LIKE wo_m.str,
         li_cnt    INTEGER

  SELECT COUNT(*) INTO li_cnt
    FROM wo_m
   WHERE wo_m.ent = g_qryparam.parm1
     AND wo_m.str = lp_str

  IF li_cnt = 0 THEN
    CALL cl_err('T100', 1001, lp_str)
    RETURN FALSE
  END IF

  LET g_qryparam.parm3 = 'Y'
  RETURN TRUE
END FUNCTION`

// ---- 时间线 / 协议流 ----
export const TIMELINE = [
  { time: '14:02:11', origin: 'human', kind: 'command', text: '启动调试 bsft001_wf(36正式)' },
  { time: '14:02:19', origin: 'system', kind: 'info', text: '连接 10.20.36.11:22 · 登录区域 36正式 · TOPENT=35' },
  { time: '14:02:31', origin: 'ai', kind: 'command', text: 'continue → 运行到断点 #1' },
  { time: '14:02:44', origin: 'system', kind: 'stop', text: '停站 bsft001_wf.4gl:412 (断点 #1)' },
  { time: '14:02:47', origin: 'ai', kind: 'command', text: 'print lp_str' },
  { time: '14:02:52', origin: 'human', kind: 'warn', text: 'print 结果被截断:数组元素超过上限 1000' },
]

export const RAW_LOG = [
  'Connecting to 10.20.36.11:22 ...',
  '(fgldb) breakpoint 1 at bsft001_wf.4gl:412',
  '(fgldb) run bsft001_wf',
  'T100 runtime 4.00.02 · DVM line offset 168',
  'Breakpoint 1, wf_confirm () at bsft001_wf.4gl:412',
  '412   SELECT COUNT(*) INTO li_cnt',
]

// ---- 接口日志 ----
export type LogRow = {
  code: string
  service: string
  job: string
  pid: string
  start: string
  end: string
  duration: string
  sso: string
  origin: string
  server: string
  errMsg: string
}

export const LOG_ROWS: LogRow[] = [
  {
    code: '000', service: 'icd.erp.wo.get', job: 'wssp01131', pid: '10231',
    start: '2026-03-11 14:02:33.418', end: '2026-03-11 14:02:33.902', duration: '0.484',
    sso: '12', origin: 'OA', server: 'T100', errMsg: '',
  },
  {
    code: '100', service: 'icd.erp.wo.post', job: 'wssp01142', pid: '10240',
    start: '2026-03-11 14:03:02.117', end: '2026-03-11 14:03:04.688', duration: '2.571',
    sso: '12', origin: 'Athena', server: 'T100',
    errMsg: 'ORA-00001: unique constraint (TOPPRD.WO_M_PK) violated',
  },
  {
    code: '000', service: 'oa.schema.data.get', job: 'wssp01131', pid: '10231',
    start: '2026-03-11 14:03:10.002', end: '2026-03-11 14:03:10.140', duration: '0.138',
    sso: '12', origin: 'OA', server: 'T100', errMsg: '',
  },
  {
    code: '000', service: 'icd.erp.bom.query', job: 'wssp01088', pid: '10188',
    start: '2026-03-11 14:04:41.776', end: '2026-03-11 14:04:45.212', duration: '3.436',
    sso: '12', origin: 'OA', server: 'T100', errMsg: '',
  },
  {
    code: '', service: 'icd.erp.item.sync', job: 'wssp01190', pid: '10252',
    start: '2026-03-11 14:05:00.001', end: '',
    duration: '', sso: '', origin: 'OA', server: 'T100', errMsg: '',
  },
]

export const LOG_REQUEST = `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  <soapenv:Body>
    <icd:getWoList xmlns:icd="http://www.digiwin.com/t100/icd">
      <ent>35</ent>
      <str>WO-2026031-0007</str>
      <plant>P01</plant>
    </icd:getWoList>
  </soapenv:Body>
</soapenv:Envelope>`

export const LOG_RESPONSE = `{
  "code": "000",
  "message": "success",
  "rows": [
    { "wo_str": "WO-2026031-0007", "item": "A-100231", "qty": 120, "unit": "PCS" },
    { "wo_str": "WO-2026031-0007", "item": "A-100232", "qty": 80,  "unit": "PCS" }
  ],
  "total": 2
}`

export const LOG_DETAIL_FIELDS: { label: string; value: string }[] = [
  { label: '服务名称', value: 'icd.erp.wo.get' },
  { label: '作业', value: 'wssp01131' },
  { label: '服务程序序号', value: '10231' },
  { label: '处理结果', value: '000' },
  { label: '开始时间', value: '2026-03-11 14:02:33.418' },
  { label: '结束时间', value: '2026-03-11 14:02:33.902' },
  { label: '耗时(s)', value: '0.484' },
  { label: 'sso时间(秒)', value: '12' },
  { label: '发起端', value: 'OA' },
  { label: '服务端', value: 'T100' },
  { label: '错误消息', value: '' },
  { label: '请求报文文件', value: 'D:\\t100_data\\wslog\\2026-03-11\\10231_req.xml' },
  { label: '响应报文文件', value: 'D:\\t100_data\\wslog\\2026-03-11\\10231_rsp.json' },
  { label: '请求大小(字节)', value: '482' },
  { label: '响应大小(字节)', value: '1963' },
]

// ---- 服务测试 ----
export const WS_MODES = [
  { value: '1', label: 'Web service (SOAP)', ep: 'awsp900' },
  { value: '2', label: 'Web service (SOAP) 备选', ep: 'awsp900' },
  { value: '3', label: 'RESTful', ep: 'awsp920' },
  { value: '4', label: 'OpenApi restful', ep: 'awsp940' },
  { value: '5', label: 'OpenApi Web service', ep: 'awsp930' },
]

export const WS_TEST_REQUEST = `{
  "ent": "35",
  "str": "WO-2026031-0007"
}`

export const WS_TEST_RESPONSE = `{
  "code": "000",
  "message": "success",
  "duration": "0.412s"
}`

// ---- 设置页 ----
export const MIRROR = {
  dir: 'D:\\t100\\mirror',
  env: '36正式',
  phase: '下载并解压…',
  percent: 72,
  note: '该目录存在于本机。',
}

export const SYNC = {
  target: 'D:\\tt\\erp_data.db',
  env: '35测试',
  phase: '创建主键索引…',
  percent: 91,
  warning: 'warning: table RD_M 的行数(1,204,882)超过预期上限,请确认服务器字典版本',
}

export const DEBUG_PARAMS = {
  launch: "BBDL512840855a 2 12345 'N' {prog}",
  watchdog: '1800',
  fglserver: '',
  printLimit: '1000',
  termW: '120',
  termH: '40',
  bpPersist: 'default',
}

export const APP_SETTINGS = {
  listen: '127.0.0.1:28670',
  bdlDoc: 'D:\\t100\\bdldoc',
}

export const RUNTIME_INFO: { label: string; value: string }[] = [
  { label: '版本', value: '0.2.3' },
  { label: '配置文件', value: 'C:\\Users\\18526\\AppData\\Roaming\\tt\\config.json' },
  { label: '缺省配置位置', value: 'C:\\Users\\18526\\AppData\\Roaming\\tt\\config.json' },
  { label: '便携模式', value: '否' },
  { label: '统一工具目录', value: 'C:\\Users\\18526\\AppData\\Roaming\\tt' },
  { label: '配置结构版本', value: '3' },
  { label: '支持的库类型', value: 'Oracle / 金仓(KingbaseES)' },
  { label: '默认环境', value: '36正式' },
  { label: '环境数', value: '3' },
]

export const CACHE_DIRS = [
  { name: 'ents', bytes: '1.2 MB', files: 18 },
  { name: 'srccache', bytes: '48.7 MB', files: 1264 },
  { name: 'execlog', bytes: '312 KB', files: 6 },
  { name: 'debug-bps', bytes: '8 KB', files: 2 },
  { name: 'spill', bytes: '6.4 MB', files: 3 },
  { name: 'update', bytes: '无', files: 0 },
]

export const UPDATE_INFO: { label: string; value: string }[] = [
  { label: '当前版本', value: '0.2.3' },
  { label: '安装形态', value: 'portable' },
  { label: '更新源', value: 'github.com/luyaogod/TT' },
  { label: '上次检查', value: '0.2.4（有新版本） · 2026/3/11 13:58:20' },
  { label: '上次升级', value: 'idle → 0.2.3' },
  { label: '技能', value: '已装到 2 个目录' },
]
