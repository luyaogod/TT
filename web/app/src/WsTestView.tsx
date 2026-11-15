// 集成服务检测（复刻 T100 awsq990 page_1）。
//
// 接口方式 → 集成方式 URL → 测试 → 结果小表 + Request/Response 本文。
// 请求从**该环境的 T100 服务器**发出（后端 SSH + curl），不是本机；
// 检测本身不写 wsfa_t，真实调用产生的日志由 T100 框架自己写。
import { useCallback, useEffect, useRef, useState } from 'react'
import { Play } from 'lucide-react'
import { Button, Input, Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../shared/ui'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../../shared/ui-radix'
import { useStore } from './store'
import { api, type WSTestMetaReply, type WSTestMethodMeta, type WSTestResult } from './api'
import { PayloadEditor } from './PayloadEditor'
import { useToast } from './toast'

// 元信息没到/读失败时的兜底：接口方式与端点映射是固定的，地址留空让用户填。
const FALLBACK_METHODS: WSTestMethodMeta[] = [
  { id: '1', label: 'Web Service', endpoint: 'awsp900', soap: true, defaultUrl: '' },
  { id: '3', label: 'Restful', endpoint: 'awsp920', soap: false, defaultUrl: '' },
  { id: '4', label: 'OpenApi Restful', endpoint: 'awsp940', soap: false, defaultUrl: '' },
  { id: '5', label: 'OpenApi Web Service', endpoint: 'awsp930', soap: true, defaultUrl: '' },
]

type Snapshot = {
  method: string
  url: string
  request: string
  response: string
  rawBody: string
  results: WSTestResult[]
  warn: string
}
const stateByEnv = new Map<string, Snapshot>()

const EMPTY: Snapshot = { method: '3', url: '', request: '', response: '', rawBody: '', results: [], warn: '' }

/** 结果区展示的文本：非 200 与传输失败按 T100 口径给一句话。 */
function displayResponse(r: WSTestResult): string {
  if (r.transportError) return r.transportError
  if (r.httpCode !== 200) return `HTTPS Error (${r.httpCode}) ${r.statusText}`.trim()
  return r.body
}

export function WsTestView() {
  const activeEnv = useStore((s) => s.activeEnv)
  const toast = useToast()
  const [meta, setMeta] = useState<WSTestMetaReply | null>(null)
  const [method, setMethod] = useState('3')
  const [url, setUrl] = useState('')
  const [request, setRequest] = useState('')
  const [response, setResponse] = useState('')
  const [rawBody, setRawBody] = useState('')
  const [results, setResults] = useState<WSTestResult[]>([])
  const [warn, setWarn] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')

  const methods = meta?.methods?.length ? meta.methods : FALLBACK_METHODS
  const current = methods.find((m) => m.id === method)

  // 按环境保存"这一页的输入与结果"（keep-alive + 切环境都原样）。
  const snapRef = useRef<Snapshot>(EMPTY)
  snapRef.current = { method, url, request, response, rawBody, results, warn }

  useEffect(() => {
    if (!activeEnv) return
    const saved = stateByEnv.get(activeEnv)
    if (saved) {
      setMethod(saved.method); setUrl(saved.url); setRequest(saved.request)
      setResponse(saved.response); setRawBody(saved.rawBody); setResults(saved.results); setWarn(saved.warn)
    } else {
      setMethod('3'); setUrl(''); setRequest(''); setResponse(''); setRawBody(''); setResults([]); setWarn('')
    }
    setMeta(null)
    setErr('')
    let cancelled = false
    api.wsTestMeta(activeEnv).then((m) => {
      if (cancelled) return
      setMeta(m)
      // 没有保存过地址（或保存的是空）时用默认地址；保存过的不覆盖。
      if (!saved || !saved.url) {
        const def = m.methods.find((x) => x.id === (saved?.method || '3'))?.defaultUrl
        if (def) setUrl(def)
      }
    }).catch((e: any) => {
      if (!cancelled) setErr(e?.message || String(e))
    })
    return () => {
      cancelled = true
      stateByEnv.set(activeEnv, snapRef.current)
    }
  }, [activeEnv])

  const changeMethod = (id: string) => {
    setMethod(id)
    const def = meta?.methods.find((x) => x.id === id)?.defaultUrl
    if (def) setUrl(def)
  }

  const run = useCallback(async () => {
    if (!activeEnv) { toast.error('没有生效环境'); return }
    setBusy(true)
    setErr('')
    try {
      const r = await api.wsTest({ env: activeEnv, method, url: url.trim() || undefined, request })
      setResults((prev) => [r.result, ...prev].slice(0, 30))
      setResponse(displayResponse(r.result))
      setRawBody(r.result.body || '')
      setWarn(r.warn || '')
      if (r.result.transportError) {
        toast.error('检测未完成', { description: r.result.transportError })
      }
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setBusy(false)
    }
  }, [activeEnv, method, url, request, toast])

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col">
      {/* 表单：接口方式 / 集成方式 / 测试 */}
      <div className="shrink-0 space-y-2 border-b border-border px-3 py-2">
        <div className="flex flex-wrap items-end gap-2">
          <label className="block">
            <span className="mb-1 block text-[11px] text-muted-foreground">接口方式</span>
            <Select value={method} onValueChange={changeMethod}>
              <SelectTrigger className="h-8 w-40 text-[13px]" aria-label="接口方式">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {methods.map((m) => (
                  <SelectItem key={m.id} value={m.id}>{m.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </label>
          <label className="block min-w-0 flex-1">
            <span className="mb-1 block text-[11px] text-muted-foreground">集成方式</span>
            <Input value={url} onChange={(e) => setUrl(e.target.value)}
              placeholder={current?.defaultUrl || 'http://…/ws/r/awsp920'}
              className="h-8 w-full font-mono text-xs" aria-label="集成方式" />
          </label>
          <Button size="sm" onClick={() => void run()} disabled={busy || !url.trim()}>
            <Play className="h-3.5 w-3.5" />
            {busy ? '测试中…' : '测试'}
          </Button>
        </div>
        <div className="flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
          <span>请求由 {meta?.host || '该环境'} 服务器发出{meta?.zoneName ? `（${meta.zoneName}）` : ''}</span>
          {current?.soap && <span>· SOAP 信封</span>}
          {warn && <span className="text-amber-600 dark:text-amber-500">· {warn}</span>}
        </div>
        {err && (
          <div className="rounded-ui border border-red-500/20 bg-red-500/10 px-2 py-1 text-[11px] text-red-600 dark:text-red-400">
            {err}
          </div>
        )}
      </div>

      {/* 结果小表：新结果在最前，只在本次页面会话里保留 */}
      <div className="max-h-40 shrink-0 overflow-auto border-b border-border">
        <Table container={false} className="table-fixed border-separate border-spacing-0 text-xs">
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead className="w-44 border-t-0 border-l-0 bg-background">起始时间</TableHead>
              <TableHead className="w-24 border-t-0 border-l-0 bg-background">HTTP code</TableHead>
              <TableHead className="border-t-0 border-l-0 bg-background">运行结果</TableHead>
              <TableHead className="w-32 border-t-0 border-l-0 bg-background">处理时间(秒)</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {results.map((r, i) => (
              <TableRow key={`${r.at}-${i}`} className="h-7 hover:bg-transparent">
                <TableCell className="border-t-0 border-l-0 px-2 py-0 font-mono text-[11px] text-muted-foreground">{r.at}</TableCell>
                <TableCell className="border-t-0 border-l-0 px-2 py-0 font-mono text-[11px]">
                  <span className={r.httpCode === 200 ? 'text-emerald-600 dark:text-emerald-500' : 'text-red-500'}>
                    {r.httpCode || '-'}
                  </span>
                </TableCell>
                <TableCell className="border-t-0 border-l-0 px-2 py-0">
                  <span className="block truncate" title={r.transportError || r.statusText}>
                    {r.transportError ? '未完成' : (r.statusText || '-')}
                  </span>
                </TableCell>
                <TableCell className="border-t-0 border-l-0 px-2 py-0 font-mono text-[11px] text-muted-foreground">
                  {r.durationSec.toFixed(3)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        {results.length === 0 && (
          <div className="p-4 text-center text-[11px] text-muted-foreground">还没有执行过检测</div>
        )}
      </div>

      {/* Request / Response 本文 */}
      <div className="grid min-h-0 flex-1 grid-cols-2 gap-2 p-2">
        <div className="flex min-h-0 flex-col border border-border bg-card">
          <div className="flex h-7 shrink-0 items-center justify-between border-b border-border px-2 text-[11px] text-muted-foreground">
            <span>Request 本文</span>
            <span className="opacity-70">可编辑</span>
          </div>
          <div className="min-h-0 flex-1">
            <PayloadEditor text={request} editable onChange={setRequest}
              path={`wstest:${activeEnv}:req`} emptyHint="在这里填请求报文" />
          </div>
        </div>
        <div className="flex min-h-0 flex-col border border-border bg-card">
          <div className="flex h-7 shrink-0 items-center justify-between border-b border-border px-2 text-[11px] text-muted-foreground">
            <span>Response 本文</span>
            {busy && <span>测试中…</span>}
          </div>
          <div className="min-h-0 flex-1">
            <PayloadEditor text={response} path={`wstest:${activeEnv}:rsp`} emptyHint="执行一次测试后显示响应" />
          </div>
          {!!rawBody && rawBody !== response && (
            <details className="shrink-0 border-t border-border px-2 py-1 text-[11px]">
              <summary className="cursor-pointer text-muted-foreground">网关原始响应</summary>
              <pre className="mt-1 max-h-32 overflow-auto whitespace-pre-wrap break-words text-[11px] text-foreground">{rawBody}</pre>
            </details>
          )}
        </div>
      </div>
    </div>
  )
}
