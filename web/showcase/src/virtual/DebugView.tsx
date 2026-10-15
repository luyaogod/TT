import { useEffect, useRef, useState } from 'react'
import {
  Bug, Eye, EyeOff, Loader2, Play, Ruler, SquareChevronDown, SquareChevronRight, X,
} from 'lucide-react'
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { AUTOVARS, BREAKPOINTS, RAW_LOG, SESSION, STACK_FRAMES, TIMELINE, WATCHES } from '@/lib/fixtures'
import { BP_LINE, CODE_LINES, CUR_LINE, type SessionState, type Tok } from './mock'
import { StatusBadge } from './status'
import { cn } from 'cn'

const TONE: Record<Tok, string> = {
  kw: 'text-info',
  str: 'text-warning',
  cmt: 'text-muted-foreground/80 italic',
  fn: 'text-foreground',
  num: 'text-foreground',
  plain: 'text-foreground',
}

export function DebugView({ state, collab, onToast }: {
  state: SessionState
  collab: boolean
  onToast: (msg: string) => void
}) {
  const stopped = state === 'stopped'
  const [bottomH, setBottomH] = useState(200)
  const [rightShown, setRightShown] = useState(true)
  const dragRef = useRef<{ startY: number; startH: number } | null>(null)

  // 分栏拖拽:保持手写(官方 Resizable 会丢掉比例持久化、夹取与"左栏 min-w-0"这条防反馈环的锁)
  useEffect(() => {
    const move = (e: MouseEvent) => {
      const d = dragRef.current
      if (!d) return
      setBottomH(Math.min(420, Math.max(120, d.startH - (e.clientY - d.startY))))
    }
    const up = () => {
      dragRef.current = null
      document.body.style.cursor = ''
    }
    window.addEventListener('mousemove', move)
    window.addEventListener('mouseup', up)
    return () => {
      window.removeEventListener('mousemove', move)
      window.removeEventListener('mouseup', up)
    }
  }, [])

  return (
    <div className="flex min-h-0 flex-1">
      <div className="flex min-h-0 min-w-0 flex-1 flex-col">
        {/* 页签栏:滚动只包页签,右端按钮 shrink-0 不跟着滚走 */}
        <div className="flex h-9 shrink-0 items-stretch border-b bg-background">
          <div className="flex min-w-0 flex-1 items-stretch overflow-x-auto">
            <Tab label={SESSION.prog} icon={<Bug className="size-3" />} active={false} />
            <Tab label="bsft001_wf.4gl" active closable />
            <Tab label="cl_err.4gl" closable />
            <Tab label="wssp01131.4gl (无源码)" closable muted />
          </div>
          <div className="flex shrink-0 items-center gap-0.5 border-l pl-1">
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon" className="size-7" aria-label="收起下方面板"><SquareChevronDown className="size-4" /></Button>
              </TooltipTrigger>
              <TooltipContent>收起下方面板</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon" className="size-7" aria-label="收起右方面板" onClick={() => setRightShown((v) => !v)}>
                  <SquareChevronRight className="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>收起右方面板</TooltipContent>
            </Tooltip>
          </div>
        </div>

        {/* 编辑器区:底色必须与 Monaco 的 editor.background 相等(= --card) */}
        <div className="min-h-0 flex-1 overflow-hidden bg-card">
          <EditorMock />
        </div>

        {/* 下方面板:自绘 5px 拖拽条(与产品现有的 .row-resizer 等价) */}
        <div
          className="group flex h-[5px] shrink-0 cursor-row-resize items-center justify-center"
          onMouseDown={(e) => { e.preventDefault(); dragRef.current = { startY: e.clientY, startH: bottomH }; document.body.style.cursor = 'row-resize' }}
          title="拖拽调整下方面板高度"
        >
          <span className="h-px w-full bg-border group-hover:h-0.5 group-hover:bg-ring" />
        </div>
        <div className="shrink-0 border-t" style={{ height: bottomH }}>
          <BottomPanel collab={collab} />
        </div>
      </div>

      {rightShown && <RightPanel state={state} collab={collab} onToast={onToast} stopped={stopped} />}
    </div>
  )
}

function Tab({ label, active, closable, icon, muted }: {
  label: string; active?: boolean; closable?: boolean; icon?: React.ReactNode; muted?: boolean
}) {
  return (
    <div className={cn('group flex shrink-0 items-stretch border-r', active && 'bg-card')}>
      <button
        aria-selected={active}
        className={cn(
          'flex max-w-56 items-center gap-1.5 truncate px-3 text-xs',
          active
            ? 'font-medium text-foreground shadow-[inset_0_1px_0_0_var(--info)]'
            : 'text-muted-foreground hover:bg-accent/60',
          muted && !active && 'italic',
        )}
      >
        {icon}
        {label}
      </button>
      {closable && (
        <Button variant="ghost" size="icon" className="my-1 mr-1 size-6 opacity-0 group-hover:opacity-60 hover:opacity-100" aria-label={`关闭 ${label}`}>
          <X className="size-3" />
        </Button>
      )}
    </div>
  )
}

function EditorMock() {
  return (
    <div className="h-full overflow-auto font-mono text-[12px] leading-[19px]">
      {CODE_LINES.map((l) => {
        const cur = l.n === CUR_LINE
        return (
          <div key={l.n} className={cn('flex items-start', cur && 'bg-warning/15')}>
            <div className="flex w-16 shrink-0 items-center justify-end gap-1 pr-2 text-muted-foreground select-none">
              {l.n === BP_LINE && <span className="size-2.5 shrink-0 rounded-full bg-destructive" aria-label="断点" />}
              {cur && <span className="size-2 shrink-0 bg-warning" aria-label="当前停站行" />}
              {l.n}
            </div>
            <div className="min-w-0 flex-1 whitespace-pre pr-4">
              {l.parts.length === 0 ? ' ' : l.parts.map(([t, s], i) => <span key={i} className={TONE[t]}>{s}</span>)}
            </div>
          </div>
        )
      })}
    </div>
  )
}

function BottomPanel({ collab }: { collab: boolean }) {
  const [cmd, setCmd] = useState('')
  const hist = useRef<string[]>(['print lp_str', 'info breakpoints'])
  const [idx, setIdx] = useState(-1)
  const onKey = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      if (cmd.trim()) hist.current = [...hist.current.filter((x) => x !== cmd.trim()), cmd.trim()].slice(-50)
      setCmd(''); setIdx(-1); return
    }
    if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
      e.preventDefault()
      const h = hist.current
      if (!h.length) return
      const next = e.key === 'ArrowUp' ? (idx < 0 ? h.length - 1 : Math.max(0, idx - 1)) : (idx < 0 ? -1 : (idx + 1 >= h.length ? -1 : idx + 1))
      setIdx(next); setCmd(next >= 0 ? h[next] : '')
    }
  }
  const tone = (kind: string) =>
    kind === 'stop' ? 'text-warning' : kind === 'warn' ? 'text-destructive' : kind === 'command' ? 'text-info' : 'text-muted-foreground'
  return (
    <Tabs defaultValue="raw" className="h-full gap-0">
      <div className="flex h-9 items-center justify-between border-b px-2">
        <TabsList variant="line" className="h-9">
          <TabsTrigger value="raw">原始协议流 ({RAW_LOG.length})</TabsTrigger>
          <TabsTrigger value="timeline">操作时间线</TabsTrigger>
        </TabsList>
        <span className="text-[11px] text-muted-foreground">fgldb 直通</span>
      </div>
      <TabsContent value="raw" className="mt-0 flex min-h-0 flex-col">
        <div className="min-h-0 flex-1 overflow-auto p-2 font-mono text-[11px] leading-5 text-muted-foreground">
          {RAW_LOG.map((l, i) => <div key={i} className="break-all whitespace-pre-wrap">{l}</div>)}
        </div>
        <div className="flex h-9 shrink-0 items-center gap-2 border-t px-2">
          <span className="font-mono text-xs text-info">&gt;</span>
          <Input
            value={cmd}
            onChange={(e) => setCmd(e.target.value)}
            onKeyDown={onKey}
            disabled={collab}
            placeholder={collab ? '协作模式:命令由 AI 执行,请在对话里委托给 AI' : 'fgldb 命令,如 print lp_str / info breakpoints(↑↓ 历史),回车发送'}
            className="h-7 flex-1 border-0 bg-transparent font-mono text-xs shadow-none focus-visible:ring-0 dark:bg-transparent"
          />
        </div>
      </TabsContent>
      <TabsContent value="timeline" className="mt-0 min-h-0 flex-1 overflow-auto p-1 font-mono text-[11px] leading-6">
        {TIMELINE.map((t, i) => (
          <div key={i} className="flex items-start gap-2 px-1">
            <span className="shrink-0 text-muted-foreground">{t.time}</span>
            <Badge variant={t.origin === 'human' ? 'secondary' : 'outline'} className="mt-0.5 shrink-0">
              {t.origin === 'ai' ? 'AI' : t.origin === 'human' ? '用户' : '系统'}
            </Badge>
            <span className={cn('min-w-0', tone(t.kind))}>{t.text}</span>
          </div>
        ))}
      </TabsContent>
    </Tabs>
  )
}

function RightPanel({ state, collab, stopped, onToast }: {
  state: SessionState; collab: boolean; stopped: boolean; onToast: (m: string) => void
}) {
  const [stackAuto, setStackAuto] = useState(true)
  const [autovarsAuto, setAutovarsAuto] = useState(false)
  const [open, setOpen] = useState<string[]>(['watches', 'bps'])
  const [target, setTarget] = useState('bsft001_wf')

  return (
    <div className="flex w-80 shrink-0 flex-col overflow-hidden border-l bg-background">
      <div className="flex h-9 shrink-0 items-center px-3 text-xs font-medium text-muted-foreground">运行</div>
      <div className="flex items-center gap-1 px-2 pb-2">
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="inline-flex">
              <Button
                variant="ghost"
                size="icon"
                className="size-7 text-success"
                disabled={collab || state === 'loading'}
                aria-label="启动调试会话"
                onClick={() => onToast(`已请求启动调试会话:${target}`)}
              >
                <Play className="size-4" fill="currentColor" />
              </Button>
            </span>
          </TooltipTrigger>
          <TooltipContent>{collab ? '协作模式下由 AI 主导' : '启动调试会话(Enter 同效)'}</TooltipContent>
        </Tooltip>
        <Input
          value={target}
          onChange={(e) => setTarget(e.target.value)}
          className="h-8 flex-1 text-xs"
          placeholder="作业编号,如 bsft001_wf 或 asf/bsft001_wf"
        />
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="inline-flex">
              <Button variant="ghost" size="icon" className="size-7 text-info" disabled={collab || !stopped} aria-label="行号校准">
                <Ruler className="size-4" />
              </Button>
            </span>
          </TooltipTrigger>
          <TooltipContent>{stopped ? '行号校准:协议行号与源码错位时点击对齐' : '行号校准(需停站后点击)'}</TooltipContent>
        </Tooltip>
      </div>

      <Accordion type="multiple" value={open} onValueChange={setOpen} className="flex min-h-0 flex-1 flex-col">
        <PanelItem value="stack" title={`调用栈 (${STACK_FRAMES.length})`} auto={stackAuto} onToggleAuto={() => setStackAuto((v) => !v)}>
          {STACK_FRAMES.map((f) => (
            <div key={f.idx} className={cn('flex cursor-pointer items-baseline gap-1.5 border-b px-3 py-1.5 text-xs last:border-b-0 hover:bg-accent/40', f.idx === 0 && 'bg-accent/60')}>
              <span className="text-muted-foreground">#{f.idx}</span>
              <span className="font-mono text-info">{f.func}</span>
              <span className="ml-auto font-mono text-muted-foreground">{f.file}:{f.line}</span>
            </div>
          ))}
        </PanelItem>

        <PanelItem value="autovars" title={`自动变量 (${AUTOVARS.length})`} auto={autovarsAuto || collab} onToggleAuto={() => !collab && setAutovarsAuto((v) => !v)} autoLocked={collab}>
          {AUTOVARS.map((v) => (
            <div key={v.expr} className="border-b px-3 py-1.5 text-xs last:border-b-0">
              <div className="font-mono text-muted-foreground">{v.expr}</div>
              <div className="break-all whitespace-pre-wrap text-success">{v.value}</div>
            </div>
          ))}
        </PanelItem>

        <PanelItem value="watches" title={`变量监视 (${WATCHES.length})`}>
          <div className="border-b p-2">
            <Input className="h-8 text-xs" placeholder="表达式,如 lp_str / g_qryparam.*(回车求值)" />
          </div>
          {WATCHES.map((w) => (
            <div key={w.expr} className="flex items-start gap-1.5 border-b px-3 py-1.5 text-xs last:border-b-0">
              <button className="shrink-0 text-muted-foreground hover:text-destructive" aria-label={`移除监视 ${w.expr}`}>×</button>
              <div className="min-w-0">
                <div className="font-mono text-muted-foreground">{w.expr}</div>
                {'error' in w && w.error
                  ? <div className="break-all text-destructive">{w.error}</div>
                  : <div className="break-all text-success">{w.value}</div>}
              </div>
            </div>
          ))}
        </PanelItem>

        <PanelItem value="bps" title={`断点 (${BREAKPOINTS.length})`}>
          {BREAKPOINTS.map((b) => (
            <div key={b.num} className={cn('relative flex items-center gap-2 border-b px-3 py-1.5 text-xs last:border-b-0', !b.enabled && 'opacity-50', b.hit && stopped && 'bg-warning/15')}>
              {b.hit && stopped && <span className="absolute inset-y-0 left-0 w-0.5 bg-warning" aria-hidden />}
              <Checkbox checked={b.enabled} disabled={collab} aria-label={`断点 ${b.num}`} />
              <span className="min-w-0 flex-1 truncate font-mono">
                {b.func && <span className="text-info">{b.func} </span>}
                {b.file}:{b.line}
                {!b.enabled && <span className="ml-1 font-sans text-[10px] text-muted-foreground">(已禁用)</span>}
              </span>
              <button className="text-muted-foreground hover:text-destructive" aria-label={`删除断点 ${b.num}`}>×</button>
            </div>
          ))}
        </PanelItem>
      </Accordion>
    </div>
  )
}

function PanelItem({ value, title, auto, onToggleAuto, autoLocked, children }: {
  value: string; title: string; auto?: boolean; onToggleAuto?: () => void; autoLocked?: boolean; children: React.ReactNode
}) {
  return (
    <AccordionItem value={value} className="border-b last:border-b-0">
      <div className="group flex items-stretch transition-colors hover:bg-accent/50">
        {onToggleAuto ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                aria-pressed={auto}
                aria-label={auto ? '关闭自动刷新' : '开启自动刷新'}
                onClick={onToggleAuto}
                className={cn('flex w-8 shrink-0 items-center justify-center', auto ? 'text-success' : 'text-muted-foreground')}
              >
                {auto ? <Eye className="size-3.5" /> : <EyeOff className="size-3.5" />}
              </button>
            </TooltipTrigger>
            <TooltipContent>
              {autoLocked ? '协作模式:自动求值开关由 AI 负责' : auto ? '自动刷新已开启:每次停站抓取(点击关闭)' : '自动刷新已关闭(默认):点击开启'}
            </TooltipContent>
          </Tooltip>
        ) : (
          <span className="block w-8 shrink-0" aria-hidden />
        )}
        <AccordionTrigger className="flex-1 py-2 pr-3 pl-0 text-xs font-medium text-muted-foreground hover:no-underline group-hover:text-foreground">
          {title}
        </AccordionTrigger>
      </div>
      <AccordionContent className="p-0">{children}</AccordionContent>
    </AccordionItem>
  )
}

export function SessionRibbon({ state, collab }: { state: SessionState; collab: boolean }) {
  return (
    <div className="flex items-center gap-2 border-b bg-card px-3 py-2 text-xs">
      <StatusBadge tone={state === 'stopped' ? 'warning' : state === 'running' ? 'info' : 'neutral'}>
        {state === 'stopped' ? '已停站' : state === 'running' ? '运行中' : state === 'loading' ? '启动中' : '会话空闲'}
      </StatusBadge>
      <span className="font-mono text-muted-foreground">{SESSION.file}:{SESSION.line}</span>
      {collab && <Badge variant="outline">协作模式 · AI 主导</Badge>}
      <Separator orientation="vertical" className="h-4" />
      <span className="text-muted-foreground">停留 {SESSION.holdSeconds}s · 作业 {SESSION.prog}</span>
      {state === 'loading' && <Loader2 className="ml-auto size-3.5 animate-spin text-muted-foreground" />}
    </div>
  )
}
