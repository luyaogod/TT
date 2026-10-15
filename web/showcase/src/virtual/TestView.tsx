import { useMemo, useState } from 'react'
import { Play } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { StatusBadge } from './status'
import { WS_MODES, WS_TEST_REQUEST, WS_TEST_RESPONSE } from '@/lib/fixtures'
import { PAYLOAD_LINES } from './mock'
import { cn } from 'cn'

export function TestView({ onToast, wsLogService }: { onToast: (m: string) => void; wsLogService?: string }) {
  const [mode, setMode] = useState('3')
  const [soap, setSoap] = useState(false)
  const [body, setBody] = useState(WS_TEST_REQUEST)
  const [running, setRunning] = useState(false)
  const [result, setResult] = useState<{ httpCode: number; duration: number; response: string } | null>(null)

  const ep = useMemo(() => WS_MODES.find((m) => m.value === mode)?.ep || 'awsp920', [mode])
  const isSoap = mode === '1' || mode === '2' || mode === '5'
  const url = `http://127.0.0.1/wt35prd/ws/r/${ep}`

  const run = () => {
    setRunning(true)
    setResult(null)
    window.setTimeout(() => {
      setRunning(false)
      setResult({ httpCode: 200, duration: 0.412, response: WS_TEST_RESPONSE })
      onToast('接口调用完成:HTTP 200 · 0.412s')
    }, 900)
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="shrink-0 space-y-2 border-b px-3 py-2">
        <div className="flex flex-wrap items-center gap-2">
          <Select value={mode} onValueChange={(v) => { setMode(v); setBody(''); setResult(null) }}>
            <SelectTrigger className="w-72" title="接口方式(wsfc001 映射)"><SelectValue placeholder="接口方式" /></SelectTrigger>
            <SelectContent>
              {WS_MODES.map((m) => <SelectItem key={m.value} value={m.value} className="text-xs">{`${m.label} (${m.ep})`}</SelectItem>)}
            </SelectContent>
          </Select>
          <Input className="h-8 w-[420px] font-mono text-xs" defaultValue={url} key={url} title={url} />
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {isSoap && (
            <span className="flex items-center gap-2">
              <Checkbox id="soap" checked={soap} onCheckedChange={(v) => setSoap(v === true)} />
              <Label htmlFor="soap" className="font-normal">SOAP</Label>
            </span>
          )}
          <Button
            size="sm"
            className="bg-success text-success-foreground hover:bg-success/90"
            disabled={running || !body.trim()}
            onClick={run}
            title="执行接口调用(服务器侧 curl POST)"
          >
            {running ? <Spinner className="size-3.5" /> : <Play className="size-3.5" fill="currentColor" />}
            {running ? '执行中…' : '执行'}
          </Button>
          {wsLogService && (
            <Button
              variant="outline"
              size="sm"
              title={`带入日志报文:${wsLogService}`}
              onClick={() => { setBody(LOG_LIKE); setSoap(false); setResult(null); onToast(`已带入日志报文:${wsLogService}`) }}
            >
              从日志带入({wsLogService.length > 16 ? wsLogService.slice(0, 16) + '…' : wsLogService})
            </Button>
          )}
          <span className="text-[11px] text-muted-foreground">切换接口方式会清空已填 URL 与报文(现状行为,保留)</span>
        </div>
      </div>

      <div className="flex min-h-0 flex-1 flex-col lg:flex-row">
        <div className="flex min-h-40 flex-1 flex-col overflow-hidden border-b lg:border-b-0 lg:border-r">
          <div className="flex h-9 shrink-0 items-center border-b bg-card px-3 text-xs font-medium text-muted-foreground">
            请求报文(JSON / XML)
          </div>
          <div className="min-h-0 flex-1 overflow-hidden bg-card">
            <pre className="h-full overflow-auto p-3 font-mono text-[11px] leading-[18px]">{body || '(空)'}</pre>
          </div>
        </div>
        <div className="flex min-h-40 flex-1 flex-col overflow-hidden">
          <div className="flex h-9 shrink-0 items-center gap-3 border-b bg-card px-3 text-xs font-medium text-muted-foreground">
            响应
            {result && (
              <span className="ml-auto flex items-center gap-3 font-normal">
                <StatusBadge tone={result.httpCode === 200 ? 'success' : 'danger'} className="font-mono">HTTP {result.httpCode}</StatusBadge>
                <span className="font-mono">{result.duration.toFixed(3)}s</span>
              </span>
            )}
            {running && <Spinner className="ml-auto size-3.5" />}
          </div>
          <div className={cn('min-h-0 flex-1 overflow-hidden bg-card')}>
            {result
              ? <pre className="h-full overflow-auto p-3 font-mono text-[11px] leading-[18px]">{result.response}</pre>
              : <div className="flex h-full items-center justify-center text-xs text-muted-foreground">
                  {running ? '执行中…' : '执行后在此显示响应报文'}
                </div>}
          </div>
        </div>
      </div>
      <div className="flex shrink-0 items-center gap-2 border-t px-3 py-2 text-[11px] text-muted-foreground">
        <Badge variant="outline">静态示意</Badge>
        真实实现里这两栏是 Monaco(共用源码编辑器同一套明暗主题);底栏展示的报文行数:
        {` ${PAYLOAD_LINES.length} 行(JSON 解析示例)`}
      </div>
    </div>
  )
}

const LOG_LIKE = `{
  "ent": "35",
  "str": "WO-2026031-0007",
  "plant": "P01"
}`
