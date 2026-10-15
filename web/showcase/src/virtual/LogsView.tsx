import { useEffect, useMemo, useState } from 'react'
import {
  Bug, CalendarDays, Check, ChevronLeft, ChevronRight, ChevronsUpDown, Copy, Fingerprint, RefreshCw, Undo2, X,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { Calendar } from '@/components/ui/calendar'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { LOG_DETAIL_FIELDS, LOG_REQUEST, LOG_RESPONSE, LOG_ROWS, type LogRow } from '@/lib/fixtures'
import { StatusBadge, StatusNote } from './status'
import { cn } from 'cn'

type Col = { key: keyof LogRow; head: string; w?: string; mono?: boolean; sortable?: boolean }

const COLS: Col[] = [
  { key: 'code', head: '状态', w: 'w-16' },
  { key: 'service', head: '服务', w: 'w-44' },
  { key: 'job', head: '作业', w: 'w-28' },
  { key: 'pid', head: '服务程序序号', w: 'w-24', mono: true },
  { key: 'start', head: '开始时间', w: 'w-44', mono: true, sortable: true },
  { key: 'end', head: '结束时间', w: 'w-44', mono: true },
  { key: 'duration', head: '耗时(s)', w: 'w-20', mono: true, sortable: true },
  { key: 'sso', head: 'sso时间(秒)', w: 'w-24', mono: true },
  { key: 'origin', head: '发起端', w: 'w-24' },
  { key: 'server', head: '服务端', w: 'w-24' },
  { key: 'errMsg', head: '错误描述' },
]

const ROWS: LogRow[] = Array.from({ length: 7 }).flatMap((_, k) =>
  LOG_ROWS.map((r, i) => ({ ...r, pid: String(Number(r.pid) + k * 10 + i) })),
)

export function LogsView({ onToast, sessionReady }: { onToast: (m: string) => void; sessionReady: boolean }) {
  const [sel, setSel] = useState<number | null>(0)
  const [onlyFail, setOnlyFail] = useState(false)
  const [loading, setLoading] = useState(false)
  const [from, setFrom] = useState<Date | undefined>(new Date(2026, 2, 11))
  const [to, setTo] = useState<Date | undefined>(new Date(2026, 2, 11))
  const [sort, setSort] = useState<{ key: keyof LogRow; dir: 'asc' | 'desc' } | null>(null)

  // Esc 关详情:自己管,避免与 Radix 的 Dialog/Sheet 抢同一个按键
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && sel !== null) { e.preventDefault(); setSel(null) }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [sel])

  const rows = useMemo(() => {
    const base = onlyFail ? ROWS.filter((r) => r.code && r.code !== '000') : ROWS
    if (!sort) return base
    // 排序只作用于当前页;空值恒排最后(不随升降序翻面)
    const sorted = [...base].sort((a, b) => {
      const av = a[sort.key] || '', bv = b[sort.key] || ''
      if (!av && bv) return 1
      if (av && !bv) return -1
      const an = Number(av), bn = Number(bv)
      const na = !Number.isNaN(an) && av.trim() !== '' && sort.key === 'duration'
      const cmp = na ? an - bn : String(av).localeCompare(String(bv))
      return sort.dir === 'asc' ? cmp : -cmp
    })
    return sorted
  }, [onlyFail, sort])

  const cycleSort = (key: keyof LogRow) => {
    setSort((s) => (s?.key !== key ? { key, dir: 'asc' } : s.dir === 'asc' ? { key, dir: 'desc' } : null))
  }

  if (!sessionReady) {
    return (
      <Empty className="min-h-0 flex-1">
        <EmptyHeader>
          <EmptyTitle>未连接会话</EmptyTitle>
          <EmptyDescription>
            接口日志按当前会话所属环境的数据库查询。请先在左侧「会话」面板选择一个环境连接,再回到这里查看。
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {/* 筛选工具条:文本条件回车生效、日期与复选即时生效(这个不对称是有意的) */}
      <div className="shrink-0 space-y-2 border-b px-3 py-2">
        <div className="flex flex-wrap items-center gap-2">
          <Input className="h-8 w-64 text-xs" placeholder="服务名称(支持 * ? 通配,回车生效)" title="服务名称 wsfa001:支持 * ? 通配,如 icd.erp.wo*" />
          <Input className="h-8 w-40 text-xs" placeholder="作业编号(支持 * ? 通配)" />
          <Input className="h-8 w-28 text-xs" placeholder="服务端" />
          <Input className="h-8 w-28 text-xs" placeholder="发起端" />
          <Input className="h-8 w-32 text-xs" placeholder="处理结果" />
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="size-8"
                aria-label="按当前条件重新加载"
                disabled={loading}
                onClick={() => { setLoading(true); window.setTimeout(() => setLoading(false), 800) }}
              >
                {loading ? <Spinner className="size-4" /> : <RefreshCw className="size-4" />}
              </Button>
            </TooltipTrigger>
            <TooltipContent>按当前条件重新加载(回车等效)</TooltipContent>
          </Tooltip>
        </div>
        <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
          <span>起始时间</span>
          <DatePick value={from} onChange={setFrom} title="起始时间 wsfa003 的下界(含当天)" />
          <span>~</span>
          <span>结束时间</span>
          <DatePick value={to} onChange={setTo} title="结束时间 wsfa004 的上界(含当天,补到 23:59:59.99999)" />
          <span className="flex items-center gap-2 text-foreground">
            <Checkbox id="onlyfail" checked={onlyFail} onCheckedChange={(v) => setOnlyFail(v === true)} />
            <Label htmlFor="onlyfail" className="font-normal">仅失败</Label>
          </span>
          <StatusNote tone="warning" className="py-0.5">调试会话忙碌中:重放将自动结束当前调试</StatusNote>
          <span className="ml-auto flex items-center gap-2">
            <Badge variant="secondary">{rows.length} 条</Badge>
            <Button variant="ghost" size="icon" className="size-7" aria-label="上一页"><ChevronLeft className="size-4" /></Button>
            <span className="font-mono">第 1 页</span>
            <Button variant="ghost" size="icon" className="size-7" aria-label="下一页"><ChevronRight className="size-4" /></Button>
          </span>
        </div>
      </div>

      <div className="flex min-h-0 flex-1">
        {/* 表格:自供滚动祖先 + 表头不透明 + 单元格只留边框,三条一起 sticky 才成立 */}
        <div className="min-h-0 min-w-0 flex-1 overflow-auto">
          <table className="w-full table-fixed border-separate border-spacing-0 text-xs">
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                {COLS.map((c) => (
                  <TableHead key={String(c.key)} className={cn('sticky top-0 z-10 h-8 border-b bg-muted px-2 text-[11px] font-medium', c.w)}>
                    {c.sortable ? (
                      <button
                        className="group inline-flex items-center gap-1"
                        title={sort?.key === c.key ? (sort.dir === 'asc' ? '升序,点击改降序' : '降序,点击恢复默认(最新在前)') : '点击按此列升序排序'}
                        onClick={() => cycleSort(c.key)}
                      >
                        {c.head}
                        <ChevronsUpDown className={cn('size-3', sort?.key === c.key ? 'opacity-100' : 'opacity-0 group-hover:opacity-60')} />
                      </button>
                    ) : c.head}
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((r, i) => (
                <TableRow
                  key={i}
                  data-state={sel === i ? 'selected' : undefined}
                  onClick={() => setSel(i)}
                  className={cn('h-8 cursor-pointer hover:bg-accent/40', sel === i && 'bg-accent/60')}
                >
                  {COLS.map((c) => (
                    <TableCell key={String(c.key)} className={cn('border-b px-2 py-0', c.w, c.mono && 'font-mono text-[11px]')}>
                      <Cell col={c.key} row={r} />
                    </TableCell>
                  ))}
                </TableRow>
              ))}
            </TableBody>
          </table>
        </div>

        {sel !== null && (
          <>
            <div
              className="group flex w-[5px] shrink-0 cursor-col-resize items-stretch bg-border bg-[length:1px_100%] bg-center bg-no-repeat hover:bg-ring hover:bg-[length:2px_100%]"
              title="拖拽分配宽度"
            />
            <DetailPanel row={rows[sel] ?? ROWS[0]} onClose={() => setSel(null)} onToast={onToast} />
          </>
        )}
      </div>
    </div>
  )
}

function DatePick({ value, onChange, title }: { value?: Date; onChange: (d?: Date) => void; title: string }) {
  const [open, setOpen] = useState(false)
  const fmt = value ? `${value.getFullYear()}-${String(value.getMonth() + 1).padStart(2, '0')}-${String(value.getDate()).padStart(2, '0')}` : '选择日期'
  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm" title={title} className="h-8 justify-start gap-1.5 px-2 font-normal">
          <CalendarDays className="size-3.5" />{fmt}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar mode="single" selected={value} onSelect={(d) => { onChange(d); setOpen(false) }} />
      </PopoverContent>
    </Popover>
  )
}

function Cell({ col, row }: { col: keyof LogRow; row: LogRow }) {
  const v = row[col]
  if (col === 'code') {
    if (!v) return <span className="text-muted-foreground">-</span>
    return v === '000'
      ? <StatusBadge tone="success" className="font-mono">000</StatusBadge>
      : <StatusBadge tone="danger" className="font-mono">{v}</StatusBadge>
  }
  if (col === 'job') return <span className="font-mono text-info">{v}</span>
  if (col === 'errMsg') return v ? <span className="text-destructive">{v}</span> : null
  return <span className={cn('block truncate', col === 'service' ? 'text-foreground' : 'text-muted-foreground')}>{v || '-'}</span>
}

function DetailPanel({ row, onClose, onToast }: { row: LogRow; onClose: () => void; onToast: (m: string) => void }) {
  const [tab, setTab] = useState('info')
  const [dirty, setDirty] = useState(true)
  const [copied, setCopied] = useState<'id' | 'payload' | null>(null)

  const copy = (what: 'id' | 'payload') => {
    setCopied(what)
    onToast(what === 'id' ? `已复制标识 ${row.pid}` : '报文已复制到剪贴板')
    window.setTimeout(() => setCopied(null), 1500)
  }

  return (
    <div className="flex w-[620px] shrink-0 flex-col overflow-hidden border-l bg-card">
      <div className="flex h-11 shrink-0 items-center gap-1 border-b px-2">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="ghost" size="icon" className="size-7" aria-label="关闭详情(Esc)" onClick={onClose}><X className="size-4" /></Button>
          </TooltipTrigger>
          <TooltipContent>关闭详情(Esc)</TooltipContent>
        </Tooltip>
        <Tabs value={tab} onValueChange={setTab}>
          <TabsList variant="line" className="h-11">
            <TabsTrigger value="info">基本信息</TabsTrigger>
            <TabsTrigger value="request">
              Request
              {dirty && <span className="text-warning" title="入参已修改,重放会用它">•</span>}
            </TabsTrigger>
            <TabsTrigger value="response">Response</TabsTrigger>
          </TabsList>
        </Tabs>
        <div className="ml-auto flex items-center gap-1.5">
          {tab === 'request' && dirty && (
            <Button variant="ghost" size="sm" onClick={() => { setDirty(false); onToast('已还原为日志里的原始入参') }} title="放弃修改,还原为日志里的原始入参">
              <Undo2 className="size-3.5" /> 还原入参
            </Button>
          )}
          <Button
            size="sm"
            className="bg-success text-success-foreground hover:bg-success/90"
            title="用该日志的原始报文重放此接口调用"
            onClick={() => onToast(`已用原始报文重放 ${row.service}(现有会话会自动收口)`)}
          >
            <Bug className="size-3.5" /> 调试此调用
          </Button>
          <Button variant="outline" size="sm" title="复制这行日志的唯一标识,发给 AI 就能直接重放" onClick={() => copy('id')}>
            {copied === 'id' ? <Check className="size-3.5" /> : <Fingerprint className="size-3.5" />} 复制标识
          </Button>
          {tab !== 'info' && (
            <Button variant="outline" size="sm" title="复制报文到剪贴板" onClick={() => copy('payload')}>
              {copied === 'payload' ? <Check className="size-3.5" /> : <Copy className="size-3.5" />} 复制
            </Button>
          )}
        </div>
      </div>

      {tab === 'info' && (
        <div className="min-h-0 flex-1 overflow-auto p-3 text-xs">
          {LOG_DETAIL_FIELDS.map((f) => (
            <div key={f.label} className="flex items-baseline gap-3 border-b border-border/60 py-1.5 last:border-b-0">
              <span className="w-28 shrink-0 text-muted-foreground">{f.label}</span>
              <span className="min-w-0 flex-1 break-all font-mono">{f.value || '-'}</span>
            </div>
          ))}
        </div>
      )}

      {tab === 'request' && (
        <>
          <StatusNote tone="warning" className="rounded-none border-x-0 border-t-0">
            入参内容不完整(超过读取上限被截断,或源文件已清理只剩入库前 2000 字符):可以按原报文重放,但不能用修改后的入参重放。
          </StatusNote>
          <div className="min-h-0 flex-1 overflow-hidden bg-card">
            <pre className="h-full overflow-auto p-3 font-mono text-[11px] leading-[18px]">{LOG_REQUEST}</pre>
          </div>
        </>
      )}

      {tab === 'response' && (
        <div className="min-h-0 flex-1 overflow-hidden bg-card">
          <pre className="h-full overflow-auto p-3 font-mono text-[11px] leading-[18px]">{LOG_RESPONSE}</pre>
        </div>
      )}
    </div>
  )
}

/** 报文编辑器的空态(与真实组件同款文案,底色必须等于 Monaco 的 editor.background) */
export function PayloadEmpty({ hint }: { hint: string }) {
  return (
    <div className="flex h-full w-full items-center justify-center bg-card p-3 text-xs text-muted-foreground">
      {hint}
    </div>
  )
}
