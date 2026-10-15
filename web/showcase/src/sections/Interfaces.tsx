import { useMemo, useState } from 'react'
import { CalendarDays, Check, ChevronLeft, ChevronRight, ChevronsUpDown, Copy, Fingerprint, Play, RefreshCw, Undo2, Bug, X } from 'lucide-react'
import { Block, Demo } from '@/components/Block'
import { blocksOf } from '@/lib/ledger'
import { LOG_DETAIL_FIELDS, LOG_REQUEST, LOG_RESPONSE, LOG_ROWS, WS_MODES, WS_TEST_REQUEST, WS_TEST_RESPONSE, type LogRow } from '@/lib/fixtures'

import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardHeader, CardTitle } from '@/components/ui/card'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Calendar } from '@/components/ui/calendar'
import {
  Pagination, PaginationContent, PaginationItem, PaginationLink, PaginationNext, PaginationPrevious,
} from '@/components/ui/pagination'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { InputGroup, InputGroupInput } from '@/components/ui/input-group'
import { Spinner } from '@/components/ui/spinner'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Item, ItemContent, ItemDescription, ItemGroup, ItemTitle } from '@/components/ui/item'
import { cn } from 'cn'

const B = blocksOf('interfaces')

const COLS_FULL: { key: keyof LogRow; head: string; className?: string; mono?: boolean }[] = [
  { key: 'code', head: '状态', className: 'w-16' },
  { key: 'service', head: '服务', className: 'w-40' },
  { key: 'job', head: '作业', className: 'w-28' },
  { key: 'pid', head: '服务程序序号', className: 'w-24', mono: true },
  { key: 'start', head: '开始时间', className: 'w-44', mono: true },
  { key: 'end', head: '结束时间', className: 'w-44', mono: true },
  { key: 'duration', head: '耗时(s)', className: 'w-20', mono: true },
  { key: 'sso', head: 'sso时间(秒)', className: 'w-24', mono: true },
  { key: 'origin', head: '发起端', className: 'w-24' },
  { key: 'server', head: '服务端', className: 'w-24' },
  { key: 'errMsg', head: '错误描述' },
]

const MANY_ROWS: LogRow[] = Array.from({ length: 6 }).flatMap((_, k) =>
  LOG_ROWS.map((r, i) => ({ ...r, pid: String(Number(r.pid) + k * 10 + i) })),
)

export function InterfacesSection() {
  return (
    <div className="space-y-6">
      <FilterBar />
      <LogTable />
      <Paging />
      <LogDetail />
      <CopyFeedback />
      <WsTest />
      <PayloadStates />
    </div>
  )
}

function FilterBar() {
  const [onlyFail, setOnlyFail] = useState(false)
  const [from, setFrom] = useState<Date | undefined>(new Date(2026, 2, 11))
  const [to, setTo] = useState<Date | undefined>(new Date(2026, 2, 11))
  const [loading, setLoading] = useState(false)
  return (
    <Block block={B['i-filter']}>
      <div className="space-y-3">
        <div className="flex flex-wrap items-center gap-2">
          <InputGroup className="h-7 w-64">
            <InputGroupInput className="text-xs" placeholder="服务名称(支持 * ? 通配,回车生效)" />
          </InputGroup>
          <Input className="h-7 w-40 text-xs" placeholder="作业编号(支持 * ? 通配)" />
          <Input className="h-7 w-28 text-xs" placeholder="服务端(如 T100)" />
          <Input className="h-7 w-28 text-xs" placeholder="发起端(如 OA)" />
          <Input className="h-7 w-32 text-xs" placeholder="处理结果(如 000)" />
          <Button variant="ghost" size="icon" className="size-7" disabled={loading} onClick={() => { setLoading(true); setTimeout(() => setLoading(false), 900) }} aria-label="按当前条件重新加载">
            {loading ? <Spinner className="size-4" /> : <RefreshCw className="size-4" />}
          </Button>
        </div>
        <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
          <span>起始时间</span>
          <DatePick value={from} onChange={setFrom} title="起始时间 wsfa003 的下界(含当天)" />
          <span>~</span>
          <span>结束时间</span>
          <DatePick value={to} onChange={setTo} title="结束时间 wsfa004 的上界(含当天,补到 23:59:59.99999)" />
          <span className="flex items-center gap-1.5 text-foreground">
            <Checkbox id="d-onlyfail" checked={onlyFail} onCheckedChange={(v) => setOnlyFail(v === true)} />
            <Label htmlFor="d-onlyfail" className="font-normal">仅失败</Label>
          </span>
          <span className="text-amber-600/80 dark:text-amber-500/80">调试会话忙碌中:重放将自动结束当前调试</span>
          <span className="ml-auto flex items-center gap-2">
            <Badge variant="secondary">{MANY_ROWS.length} 条</Badge>
            <Button variant="ghost" size="icon" className="size-7" aria-label="上一页"><ChevronLeft className="size-4" /></Button>
            <span className="font-mono">第 1 页</span>
            <Button variant="ghost" size="icon" className="size-7" aria-label="下一页"><ChevronRight className="size-4" /></Button>
          </span>
        </div>
        <div className="text-xs text-muted-foreground">
          交互不对称是有意的:文本条件<b>回车才生效</b>,日期与「仅失败」<b>即时生效</b>(注释里记着"靠 setTimeout 复用旧闭包"的历史 bug)。
          换 Input 只换外观,不要顺手把 onChange 改成即时提交。
        </div>
      </div>
    </Block>
  )
}

function DatePick({ value, onChange, title }: { value?: Date; onChange: (d?: Date) => void; title: string }) {
  const [open, setOpen] = useState(false)
  const fmt = value ? `${value.getFullYear()}-${String(value.getMonth() + 1).padStart(2, '0')}-${String(value.getDate()).padStart(2, '0')}` : '选择日期'
  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm" title={title} className="h-7 justify-start gap-1.5 px-2 font-normal">
          <CalendarDays className="size-3.5" />
          {fmt}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar mode="single" selected={value} onSelect={(d) => { onChange(d); setOpen(false) }} />
      </PopoverContent>
    </Popover>
  )
}

function LogTable() {
  const [sel, setSel] = useState<number | null>(0)
  const cols = sel !== null ? COLS_FULL.filter((c) => ['code', 'service', 'job', 'start'].includes(String(c.key))) : COLS_FULL
  return (
    <Block block={B['i-table']} pad={false}>
      <div className="grid gap-4 p-4 lg:grid-cols-2">
        <Demo label="A. 官方 Table 默认(会踩坑)" hint="Table 自带 overflow-x-auto 容器 ⇒ 它成了滚动祖先,th 的 sticky 只相对它生效,纵向滚动时表头照样走">
          <div className="h-56 overflow-auto rounded-lg border">
            <Table className="table-fixed">
              <TableHeader>
                <TableRow>
                  {cols.map((c) => (
                    <TableHead key={String(c.key)} className={cn('sticky top-0 z-10 bg-muted text-xs', c.className)}>{c.head}</TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {MANY_ROWS.map((r, i) => (
                  <TableRow key={i} data-state={sel === i ? 'selected' : undefined} onClick={() => setSel(i)} className="h-7 cursor-pointer">
                    {cols.map((c) => (
                      <TableCell key={String(c.key)} className={cn('px-2 py-0 text-xs', c.className)}>
                        <Cell col={c.key} row={r} />
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </Demo>
        <Demo label="B. 实保结构(推荐:只借官方单元格样式,自己拿滚动祖先)" hint="用官方 TableHead/TableRow/TableCell + 自己的 <table> 与滚动容器 ⇒ sticky 生效、框线不叠">
          <div className="h-56 overflow-auto rounded-lg border">
            <table className="w-full table-fixed border-separate border-spacing-0 text-sm">
              <TableHeader>
                <TableRow className="hover:bg-transparent">
                  {cols.map((c) => (
                    <TableHead key={String(c.key)} className={cn('sticky top-0 z-10 border-b bg-background px-2 text-xs', c.className)}>
                      <button className="group inline-flex items-center gap-1" title="点击按此列升序排序">
                        {c.head}
                        <ChevronsUpDown className="size-3 opacity-0 group-hover:opacity-60" />
                      </button>
                    </TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {MANY_ROWS.map((r, i) => (
                  <TableRow key={i} data-state={sel === i ? 'selected' : undefined} onClick={() => setSel(i)} className={cn('h-7 cursor-pointer hover:bg-accent/40', sel === i && 'bg-accent/60')}>
                    {cols.map((c) => (
                      <TableCell key={String(c.key)} className={cn('border-b px-2 py-0 text-xs', c.className)}>
                        <Cell col={c.key} row={r} />
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </table>
          </div>
        </Demo>
      </div>
      <div className="px-4 pb-4 text-xs text-muted-foreground">
        四条硬约束(缺一条就出问题):<b>自己当滚动祖先</b>、<code className="font-mono">table-fixed + border-separate</code>、
        单元格只留<b>右+下</b>边框(否则相邻边叠成 2px)、表头背景<b>不透明且与列表同色</b>(默认 bg-muted/50 会透出行)。
        实测点上面两块的滚动:左侧表头会跟着走,右侧不会。
      </div>
    </Block>
  )
}

function Cell({ col, row }: { col: keyof LogRow; row: LogRow }) {
  const v = row[col]
  if (col === 'code') {
    if (!v) return <span className="text-muted-foreground">-</span>
    return v === '000'
      ? <Badge className="border-transparent bg-emerald-600/10 font-mono text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-400">000</Badge>
      : <Badge variant="destructive" className="font-mono">{v}</Badge>
  }
  if (col === 'job') return <span className="font-mono text-sky-600 dark:text-sky-400">{v}</span>
  if (col === 'errMsg') return v ? <span className="text-red-600 dark:text-red-400">{v}</span> : null
  return <span className={cn(col === 'service' && 'truncate', 'text-muted-foreground')}>{v || '-'}</span>
}

function Paging() {
  return (
    <Block block={B['i-paging']}>
      <div className="space-y-4">
        <Demo label="官方 Pagination">
          <Pagination className="justify-start">
            <PaginationContent>
              <PaginationItem><PaginationPrevious href="#" /></PaginationItem>
              <PaginationItem><PaginationLink href="#" isActive>1</PaginationLink></PaginationItem>
              <PaginationItem><PaginationLink href="#">2</PaginationLink></PaginationItem>
              <PaginationItem><PaginationNext href="#" /></PaginationItem>
            </PaginationContent>
          </Pagination>
        </Demo>
        <Demo label="紧凑形态(现状:左页 / 第 N 页 / 右页 + 条数)" hint="数据密集页更合适,官方 Pagination 的按钮偏大">
          <div className="flex items-center gap-2">
            <Badge variant="secondary">{MANY_ROWS.length} 条</Badge>
            <Button variant="ghost" size="icon" className="size-7" aria-label="上一页"><ChevronLeft className="size-4" /></Button>
            <span className="font-mono text-xs">第 1 页</span>
            <Button variant="ghost" size="icon" className="size-7" aria-label="下一页"><ChevronRight className="size-4" /></Button>
            <Separator orientation="vertical" className="mx-1 h-4" />
            <Select defaultValue="default">
              <SelectTrigger size="sm" className="w-40"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="default">默认(最新在前)</SelectItem>
                <SelectItem value="dur">按耗时降序</SelectItem>
                <SelectItem value="fail">仅失败</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </Demo>
        <div className="text-xs text-muted-foreground">
          排序只作用于<b>当前页</b>(服务端默认 wsfa003 desc),且空值与 NaN 恒排最后、不随升降序翻面。
          引入官方 DataTable(服务端排序)会变更这两条语义。
        </div>
      </div>
    </Block>
  )
}

function LogDetail() {
  const [tab, setTab] = useState('info')
  const [reqDirty, setReqDirty] = useState(true)
  const row = LOG_ROWS[0]
  return (
    <Block block={B['i-detail']} pad={false}>
      <div className="max-w-3xl border-b">
        <div className="flex h-10 items-center gap-1 px-2">
          <Button variant="ghost" size="icon" className="size-7" aria-label="关闭详情(Esc)"><X className="size-4" /></Button>
          <Tabs value={tab} onValueChange={setTab}>
            <TabsList variant="line">
              <TabsTrigger value="info">基本信息</TabsTrigger>
              <TabsTrigger value="request">
                Request
                {reqDirty && <span className="text-amber-600 dark:text-amber-400" title="入参已修改,重放会用它">•</span>}
              </TabsTrigger>
              <TabsTrigger value="response">Response</TabsTrigger>
            </TabsList>
          </Tabs>
          <div className="ml-auto flex items-center gap-1.5">
            {reqDirty && tab === 'request' && (
              <Button variant="ghost" size="sm" onClick={() => setReqDirty(false)} title="放弃修改,还原为日志里的原始入参">
                <Undo2 className="size-3.5" /> 还原入参
              </Button>
            )}
            <Button variant="outline" size="sm" className="border-emerald-600/20 text-emerald-700 hover:bg-emerald-600/10 dark:text-emerald-400" title="用该日志的原始报文重放此接口调用">
              <Bug className="size-3.5" /> 调试此调用
            </Button>
            <Button variant="outline" size="sm" title="复制这行日志的唯一标识(rowid / ctid)"><Fingerprint className="size-3.5" /> 复制标识</Button>
            {tab !== 'info' && <Button variant="outline" size="sm" title="复制报文到剪贴板"><Copy className="size-3.5" /> 复制</Button>}
          </div>
        </div>
      </div>
      <div className="max-w-3xl">
        {tab === 'info' && (
          <ItemGroup className="gap-0 p-2">
            {LOG_DETAIL_FIELDS.map((f) => (
              <Item key={f.label} className="flex-nowrap items-baseline gap-2 rounded-none border-b border-border/60 px-0 py-1.5 last:border-b-0">
                <ItemContent className="flex-row items-baseline gap-2">
                  <ItemTitle className="w-28 shrink-0 text-xs font-normal text-muted-foreground">{f.label}</ItemTitle>
                  <ItemDescription className="min-w-0 flex-1 break-all font-mono text-xs text-foreground">{f.value || '-'}</ItemDescription>
                </ItemContent>
              </Item>
            ))}
          </ItemGroup>
        )}
        {tab === 'request' && <pre className="max-h-72 overflow-auto bg-card p-3 font-mono text-[11px] leading-5">{LOG_REQUEST}</pre>}
        {tab === 'response' && <pre className="max-h-72 overflow-auto bg-card p-3 font-mono text-[11px] leading-5">{LOG_RESPONSE}</pre>}
      </div>
      <div className="border-t bg-muted/30 px-4 py-2 text-[11px] text-muted-foreground">
        request/response 页必须保持 <code className="font-mono">overflow-hidden</code>(滚动交给 Monaco 自己),
        给编辑器外面再套一层 overflow-auto 会出双滚动条。另外 Esc 现在是自己监听 window keydown 处理的,换 Radix Dialog/Sheet 后要明确谁负责关闭。
        真实行:{row.service} · {row.job}
      </div>
    </Block>
  )
}

function CopyFeedback() {
  const [state, setState] = useState<'idle' | 'ok' | 'fail'>('idle')
  return (
    <Block block={B['i-copy']}>
      <div className="flex flex-wrap items-center gap-3">
        <Button
          variant="outline"
          size="sm"
          className={cn(state === 'fail' && 'border-red-500/20 text-red-600 dark:text-red-400')}
          onClick={() => setState('ok')}
          onDoubleClick={() => setState('fail')}
        >
          {state === 'ok' ? <><Check className="size-3.5" /> 已复制</> : state === 'fail' ? <><X className="size-3.5" /> 复制失败</> : <><Copy className="size-3.5" /> 复制</>}
        </Button>
        <span className="text-xs text-muted-foreground" title="单击=成功,双击=失败">1.5s 后自动复位;失败态变红,不弹 toast</span>
      </div>
    </Block>
  )
}

function WsTest() {
  const [mode, setMode] = useState('3')
  const ep = useMemo(() => WS_MODES.find((m) => m.value === mode)?.ep || 'awsp920', [mode])
  const isSoap = mode === '1' || mode === '2' || mode === '5'
  const [soap, setSoap] = useState(false)
  const [running, setRunning] = useState(false)
  return (
    <Block block={B['i-wstest']} pad={false}>
      <div className="space-y-3 p-4">
        <div className="flex flex-wrap items-center gap-2">
          <Select value={mode} onValueChange={setMode}>
            <SelectTrigger title="接口方式(wsfc001 映射)" className="w-64"><SelectValue placeholder="接口方式" /></SelectTrigger>
            <SelectContent>
              {WS_MODES.map((m) => <SelectItem key={m.value} value={m.value} className="text-xs">{`${m.label} (${m.ep})`}</SelectItem>)}
            </SelectContent>
          </Select>
          <Input className="h-8 w-[420px] font-mono text-xs" defaultValue={`http://127.0.0.1/wt35prd/ws/r/${ep}`} />
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {isSoap && (
            <span className="flex items-center gap-1.5">
              <Checkbox id="d-soap" checked={soap} onCheckedChange={(v) => setSoap(v === true)} />
              <Label htmlFor="d-soap" className="font-normal">SOAP</Label>
            </span>
          )}
          <Button size="sm" className="bg-emerald-600 text-white hover:bg-emerald-600/90" disabled={running} onClick={() => { setRunning(true); setTimeout(() => setRunning(false), 900) }}>
            <Play className="size-3.5" fill="currentColor" /> {running ? '执行中…' : '执行'}
          </Button>
          <Button variant="outline" size="sm">从日志带入(icd.erp.wo.get)</Button>
          <span className="text-xs text-muted-foreground">切换接口方式会清空已填 URL(现状行为,重建时要保留)</span>
        </div>
      </div>
      <Card className="gap-0 rounded-none border-x-0 border-b-0 py-0">
        <div className="flex min-h-40 flex-col lg:flex-row">
          <div className="flex min-h-40 flex-1 flex-col overflow-hidden border-b lg:border-b-0 lg:border-r">
            <CardHeader className="px-2.5 py-0"><CardTitle className="py-2 text-xs font-medium text-muted-foreground">请求报文(JSON / XML)</CardTitle></CardHeader>
            <pre className="min-h-0 flex-1 overflow-auto bg-card p-3 font-mono text-[11px] leading-5">{WS_TEST_REQUEST}</pre>
          </div>
          <div className="flex min-h-40 flex-1 flex-col overflow-hidden">
            <CardHeader className="flex flex-row items-center justify-between px-2.5 py-0">
              <CardTitle className="py-2 text-xs font-medium text-muted-foreground">响应</CardTitle>
              <span className="flex items-center gap-3 py-2 text-xs">
                <Badge className="border-transparent bg-emerald-600/10 font-mono text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-400">HTTP 200</Badge>
                <span className="font-mono text-muted-foreground">0.412s</span>
              </span>
            </CardHeader>
            <pre className="min-h-0 flex-1 overflow-auto bg-card p-3 font-mono text-[11px] leading-5">{WS_TEST_RESPONSE}</pre>
          </div>
        </div>
      </Card>
      <div className="bg-muted/30 px-4 py-2 text-[11px] text-muted-foreground">
        官方 Card 自带 <code className="font-mono">gap-6 / py-6</code>,而这两个面板必须 <code className="font-mono">min-h-40 + flex-1</code> 且零多余内边距
        —— 上面的写法就是"用 Card 语义 + 覆盖默认间距"的样子;若嫌麻烦,直接用 Separator + 手写头。
      </div>
    </Block>
  )
}

function PayloadStates() {
  return (
    <Block block={B['i-payload']}>
      <div className="grid gap-4 sm:grid-cols-3">
        <Demo label="加载中">
          <div className="flex h-40 items-center justify-center rounded-lg border bg-card text-xs text-muted-foreground">加载中…</div>
        </Demo>
        <Demo label="空(超过入库上限且源文件已清理)">
          <div className="flex h-40 items-center justify-center rounded-lg border bg-card px-3 text-center text-xs text-muted-foreground">
            无报文(超过入库大小上限且源文件已清理)
          </div>
        </Demo>
        <Demo label="有内容(Monaco 区域,此处为静态示意)">
          <pre className="h-40 overflow-auto rounded-lg border bg-card p-3 font-mono text-[11px] leading-[18px]">{`{
  "code": "000",
  "rows": [ … ]
}`}</pre>
        </Demo>
      </div>
      <div className="mt-3 text-xs text-muted-foreground">
        三态底色必须等于 Monaco 的 <code className="font-mono">editor.background</code>(产品里等于 --card)。
        换成官方 Empty/Skeleton 时别引入 <code className="font-mono">bg-muted</code>,否则占位块与编辑器深一块浅一块。
      </div>
    </Block>
  )
}
