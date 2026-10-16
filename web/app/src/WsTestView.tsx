// 服务测试视图:复刻 awsq990「集成服务测试」——接口方式/网址/请求报文,直接执行看响应。
// 后端经 SSH 在服务器上以 curl 调用(与 awsq990 同网络位置)。
//
// 这里**没有**执行历史表:早先的注释说"用一张常驻表格记录每次执行(与原生 g_wsfa2_d 同列)",
// 而界面上从来没有那张表(实测被测试者点出来过)。要留历史得先想清楚它归谁管(会话?
// 数据目录?)—— 在那之前不要用注释把不存在的东西说成存在。
import { useEffect, useState } from 'react'
import { Play } from 'lucide-react'
import { isSoapMode, useStore } from './store'
import { api } from './api'
import { PayloadEditor } from './PayloadEditor'
import { useToast } from './toast'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue, Checkbox } from '../../shared/ui-radix'

// 接口方式选项(wsfc001 映射,与 awsq990 一致)
const MODES: [string, string, string][] = [
  ['1', 'awsp900', 'Web service (SOAP)'],
  ['2', 'awsp900', 'Web service (SOAP) 备选'],
  ['3', 'awsp920', 'RESTful'],
  ['4', 'awsp940', 'OpenApi restful'],
  ['5', 'awsp930', 'OpenApi Web service'],
]

export function WsTestView() {
  const mode = useStore((s) => s.wsTestMode)
  const url = useStore((s) => s.wsTestUrl)
  const body = useStore((s) => s.wsTestBody)
  const soap = useStore((s) => s.wsTestSoap)
  const result = useStore((s) => s.wsTestResult)
  const running = useStore((s) => s.wsTestRunning)
  // 失败提示走统一 toast(在 store 的 runWsTest 里发),这里不再读 wsTestErr。
  const setWsTest = useStore((s) => s.setWsTest)
  const runWsTest = useStore((s) => s.runWsTest)
  const toast = useToast()
  const wsLogSel = useStore((s) => s.wsLogSel)
  const wsLogContent = useStore((s) => s.wsLogContent)
  // 默认地址里的区域别名(36→t35prd)与**前缀**(出貨區走 /ws、其余区走 /w)都取自
  // /api/status:这条拼法只有后端一份实现(WSDefaultURLFor),前端自己再推一遍容易漏掉
  // 「topprd 走 /ws」那一档 —— 实测 zone=t 时前端给 /wtopprd、后端给 /wstopprd,
  // 而本机 zone=36 时这个偏差看不出来。
  const [zoneName, setZoneName] = useState('')
  const [wsPrefix, setWsPrefix] = useState('/w')
  useEffect(() => {
    void api.status().then((s: any) => {
      setZoneName(s.zoneName || '')
      if (s.wsPrefix) setWsPrefix(s.wsPrefix)
    }).catch(() => {})
  }, [])

  const isSoap = isSoapMode(mode)
  const ep = MODES.find((m) => m[0] === mode)?.[1] || 'awsp920'
  const defaultUrl = `http://127.0.0.1${wsPrefix}${zoneName || 't35prd'}/ws/r/${ep}`

  const doRun = () => void runWsTest()

  // 「从日志带入」：只带**真的取到的报文**。
  //
  // 之前写的是 `wsLogContent?.request || wsLogSel.reqPath` —— 报文还没取到（或取得失败，
  // store 里 wsLogContent 恒为 null）时，它会退到 `reqPath`，那是**服务器上的文件路径**：
  // 一点「执行」就把 `/u1/toptest/ws/req/…/req_0001.xml` 当载荷 POST 出去（实测）。
  // 取不到就说取不到，别拿一个看起来像报文的东西顶上。
  //
  // 带进来的报文是**已经封好的**那一份，所以同时把 SOAP 勾选放掉（勾着会让后端再包一层信封）。
  const importFromLog = () => {
    const text = wsLogContent?.request
    if (!wsLogSel) return
    if (!text || !text.trim()) {
      toast.error('这条日志的请求报文还没取到', {
        description: '先在「接口日志」页点开这一行的明细，等报文出来再带入 —— ' +
          '服务器上的文件路径不是载荷，发出去只会得到一个看不懂的报错。',
        duration: 0,
      })
      return
    }
    setWsTest({ body: text, soap: false })
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {/* 工具条分两行(与日志页同样式):第一行 接口方式 + 集成方式(地址);
          第二行 SOAP 勾选 + 执行 + 从日志带入。字段不再拉满整行,留白更清楚。 */}
      <div className="shrink-0 border-b border-border">
        <div className="flex flex-wrap items-center gap-2 px-2 pt-2 pb-1.5">
          {/* 字段名在左、控件在右(与日志页的 QueryField 同一口径):
              早先把字段名写进控件的 placeholder(「接口方式」),等于标签与输入混在一件东西里。 */}
          <span className="shrink-0 text-xs text-muted-foreground" title="接口方式(wsfc001 映射)">接口方式</span>
          <Select value={mode} onValueChange={(v) => setWsTest({ mode: v, url: '' })}>
            <SelectTrigger className="h-7 w-[200px] text-xs" title="接口方式(wsfc001 映射)"
              aria-label="接口方式">
              <SelectValue placeholder="选择" />
            </SelectTrigger>
            <SelectContent>
              {MODES.map(([v, ep2, label]) => (
                <SelectItem key={v} value={v} className="text-xs">{label} ({ep2})</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <span className="shrink-0 text-xs text-muted-foreground" title={`默认地址(区域 ${zoneName || 't35prd'})`}>地址</span>
          <input
            value={url}
            onChange={(e) => setWsTest({ url: e.target.value })}
            placeholder={defaultUrl}
            aria-label="地址"
            className="h-7 w-[440px] min-w-0 rounded-ui border border-border bg-background px-2 font-mono text-xs text-foreground placeholder:text-muted-foreground focus:border-border focus:outline-none"
            title={defaultUrl}
          />
        </div>
        <div className="flex flex-wrap items-center gap-2 px-2 pt-1.5 pb-2">
          {isSoap && (
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground" title="SOAP 报文(SOAPAction 空头)">
              <Checkbox id="wstest-soap" checked={soap} onCheckedChange={(v) => setWsTest({ soap: v === true })} />
              <label htmlFor="wstest-soap" className="cursor-pointer select-none">SOAP</label>
            </div>
          )}
          <button onClick={doRun} disabled={running || !body.trim()}
            title={!body.trim()
              ? '请求报文是空的：至少要有载荷才能执行'
              : running ? '上一次还在跑' : '执行接口调用(服务器侧 curl POST)'}
            className="inline-flex items-center gap-1 border border-emerald-500/20 px-2.5 py-1 text-xs text-emerald-600 dark:text-emerald-400 hover:bg-emerald-500/10 disabled:opacity-40">
            <Play className="h-3.5 w-3.5" fill="currentColor" />
            {running ? '执行中…' : '执行'}
          </button>
          {wsLogSel && (
            <button
              onClick={importFromLog}
              title={`带入日志报文:${wsLogSel.service}(报文还没取到时不会带 —— 那会把服务器上的文件路径当载荷发出去)`}
              className="border border-border px-2 py-1 text-xs text-muted-foreground hover:bg-accent"
            >
              从日志带入({wsLogSel.service.length > 16 ? wsLogSel.service.slice(0, 16) + '…' : wsLogSel.service})
            </button>
          )}
        </div>
      </div>

      {/* 失败提示已改走统一 toast(见 store 的 runWsTest):那条横幅原先顶在这里,
          一出现就把请求/响应两个编辑区整体下推。 */}
      <div className="flex min-h-0 flex-1 flex-col lg:flex-row">
        {/* 请求报文:竖排时下缘分割线,横排时右缘分割线(与响应区紧贴相连)。
            背景交给 Monaco 主题(editor.background,与源码编辑器同色),故这里不再铺 bg。 */}
        <div className="flex min-h-40 flex-1 flex-col overflow-hidden border-b border-border lg:border-b-0 lg:border-r">
          <div className="flex h-8 shrink-0 items-center border-b border-border bg-card px-2.5 text-xs font-medium text-muted-foreground">
            请求报文(JSON / XML)
          </div>
          <div className="min-h-0 flex-1">
            <PayloadEditor text={body} editable onChange={(v) => setWsTest({ body: v })} path="wstest:request" />
          </div>
        </div>

        {/* 响应 */}
        <div className="flex min-h-40 flex-1 flex-col overflow-hidden">
          <div className="flex h-8 shrink-0 items-center gap-3 border-b border-border bg-card px-2.5 text-xs font-medium text-muted-foreground">
            响应
            {result && result.httpCode > 0 && (
              <>
                <span className={`font-mono ${result.httpCode === 200 ? 'text-emerald-600 dark:text-emerald-500' : 'text-red-500'}`}>
                  HTTP {result.httpCode}
                </span>
                <span className="font-mono text-muted-foreground">{result.durationSec.toFixed(3)}s</span>
              </>
            )}
            {result?.error && <span className="text-red-600 dark:text-red-400">{result.error}</span>}
          </div>
          <div className="min-h-0 flex-1">
            <PayloadEditor text={result?.response} loading={running} path="wstest:response"
              emptyHint={result
                ? '这次没有响应报文（原因见上方状态）'
                : '执行后在此显示响应报文'} />
          </div>
        </div>
      </div>
    </div>
  )
}
