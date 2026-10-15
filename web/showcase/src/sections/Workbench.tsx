import { useMemo, useRef, useState } from 'react'
import {
  Bug, ChevronsDownUp, Eye, EyeOff, FlaskConical, Globe, GripVertical, ListTree, Loader2, Play,
  RedoDot, RefreshCw, Ruler, Server, Settings, Square, SquareChevronDown,
  SquareChevronRight, StepForward, ArrowDownToDot, ArrowUpFromDot, X,
} from 'lucide-react'
import { Block, Demo } from '@/components/Block'
import { blocksOf } from '@/lib/ledger'
import { AUTOVARS, BREAKPOINTS, OUTLINE, RAW_LOG, SESSION, SOURCE_SNIPPET, STACK_FRAMES, TIMELINE, VAR_TREE } from '@/lib/fixtures'

import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Toggle } from '@/components/ui/toggle'
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { cn } from 'cn'

const B = blocksOf('workbench')

export function WorkbenchSection() {
  return (
    <div className="space-y-6">
      <ActivityBar />
      <StatusBar />
      <SourceTabs />
      <FloatingToolbar />
      <AccordionPanels />
      <Trees />
      <Timeline />
      <EmptyStates />
    </div>
  )
}

function ActivityBar() {
  const [view, setView] = useState('debug')
  const items = [
    { key: 'debug', icon: Bug, label: '调试' },
    { key: 'wslogs', icon: Globe, label: '接口日志' },
    { key: 'wstest', icon: FlaskConical, label: '服务测试' },
    { key: 'settings', icon: Settings, label: '设置(站点 / 数据字典 / DEBUG / 应用)' },
  ]
  return (
    <Block block={B['w-activity']}>
      <div className="flex flex-wrap items-start gap-8">
        <Demo label="方案 A:ghost icon + 选中 accent(推荐)" hint="选中态用 secondary/accent,不加边框">
          <div className="flex w-10 flex-col items-center gap-2 rounded-xl border py-3">
            {items.map(({ key, icon: Icon, label }) => (
              <Tooltip key={key}>
                <TooltipTrigger asChild>
                  <Button
                    variant={view === key ? 'secondary' : 'ghost'}
                    size="icon"
                    aria-label={label}
                    onClick={() => setView(key)}
                  >
                    <Icon />
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="right">{label}</TooltipContent>
              </Tooltip>
            ))}
          </div>
        </Demo>
        <Demo label="方案 B:Toggle(outline) 表达" hint="选中态带边框,更像「栏」里的开关">
          <div className="flex w-10 flex-col items-center gap-2 rounded-xl border py-3">
            {items.map(({ key, icon: Icon, label }) => (
              <Toggle key={key} variant="outline" size="sm" pressed={view === key} onPressedChange={() => setView(key)} aria-label={label}>
                <Icon />
              </Toggle>
            ))}
          </div>
        </Demo>
        <Demo label="右侧 sheet 切换栏(会话 / 调试面板 / 大纲)">
          <div className="flex w-10 flex-col items-center gap-2 rounded-xl border py-3">
            {[[Server, '会话'], [Bug, '调试面板'], [ListTree, '大纲']].map(([Icon, label], i) => {
              const I = Icon as typeof Server
              return (
                <Button key={i} variant={i === 0 ? 'secondary' : 'ghost'} size="icon" aria-label={label as string}>
                  <I />
                </Button>
              )
            })}
          </div>
        </Demo>
      </div>
    </Block>
  )
}

function StatusBar() {
  return (
    <Block block={B['w-statusbar']}>
      <div className="space-y-5">
        <Demo label="官方 token 直接铺(问题所在)" hint="shadcn 默认里 primary 亮色近黑、暗色近白 —— 状态栏会变成黑条/白条,失去「停站橙/运行蓝」的信息">
          <div className="space-y-2">
            <Bar className="bg-primary text-primary-foreground" left={SESSION.file} chips={[SESSION.prog, `已停站 ${SESSION.file}:${SESSION.line} · 停留 ${SESSION.holdSeconds}s`]} />
            <Bar className="bg-secondary text-secondary-foreground" left={SESSION.file} chips={[SESSION.prog, '运行中 · 请到 GDC 操作']} />
          </div>
        </Demo>
        <Demo label="方案:为状态定义语义色(需要额外 token)" hint="现状用的就是这一套(#cc6633 / #0078d4),只是写死在组件里">
          <div className="space-y-2">
            <Bar className="bg-[#cc6633] text-white" left={SESSION.file} chips={[`已停站 ${SESSION.file}:${SESSION.line} · 停留 ${SESSION.holdSeconds}s`, SESSION.prog]} />
            <Bar className="bg-[#0078d4] text-white" left={SESSION.file} chips={['AI 已放行 · 在界面上观察即可', SESSION.prog]} />
            <Bar className="bg-accent text-foreground" left="" chips={['启动中…']} />
            <Bar className="bg-muted text-muted-foreground" left="" chips={['会话空闲(36正式) · 可直接启动调试']} />
          </div>
        </Demo>
        <div className="text-xs text-muted-foreground">
          决策点:状态栏底色是"这个界面的状态语言",不是可选项。建议在 shadcn 默认 token 之外显式定义
          <code className="mx-1 rounded bg-muted px-1.5 py-0.5 font-mono">--success / --warning / --info</code>
          三个语义色,状态栏与 Badge、Alert 四色语义共用同一套。
        </div>
      </div>
    </Block>
  )
}

function Bar({ className, left, chips }: { className: string; left: string; chips: string[] }) {
  return (
    <div className={cn('flex h-7 items-center gap-3 px-3 text-[11px]', className)}>
      <span className="min-w-0 flex-1 truncate font-mono">{left || ' '}</span>
      <span className="flex shrink-0 items-center gap-2">
        <span className="rounded-sm bg-black/20 px-1.5 py-px">协作模式 · AI 主导</span>
        <span className="rounded-sm bg-black/20 px-1.5 py-px font-mono">⟳ continue(已 12s)</span>
        {chips.map((c, i) => <span key={i}>{c}</span>)}
      </span>
    </div>
  )
}

function SourceTabs() {
  const tabs = [
    { key: 'debug', label: SESSION.prog, icon: Bug, closable: false, active: false },
    { key: 't1', label: 'bsft001_wf.4gl', closable: true, active: true },
    { key: 't2', label: 'cl_err.4gl', closable: true, active: false },
    { key: 't3', label: 'wssp01131.4gl (无源码)', closable: true, active: false },
  ]
  return (
    <Block block={B['w-tabs']} pad={false}>
      <div className="flex h-9 items-stretch border-b bg-background">
        {/* 关键结构:滚动只包页签区,右端按钮 shrink-0 不跟着滚走 */}
        <div className="flex min-w-0 flex-1 items-stretch overflow-x-auto">
          {tabs.map((t) => (
            <div key={t.key} className={cn('group flex shrink-0 items-stretch border-r', t.active && 'bg-card')}>
              <button
                className={cn(
                  'flex max-w-48 items-center gap-1.5 truncate px-3 text-xs',
                  t.active ? 'font-medium text-foreground shadow-[inset_0_1px_0_0_var(--primary)]' : 'text-muted-foreground hover:bg-accent/60',
                )}
              >
                {t.icon && <t.icon className="size-3" />}
                {t.label}
              </button>
              {t.closable && (
                <Button variant="ghost" size="icon" className="my-1 mr-1 size-6 opacity-40 hover:opacity-100" aria-label={`关闭 ${t.label}`}>
                  <X className="size-3" />
                </Button>
              )}
            </div>
          ))}
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
              <Button variant="ghost" size="icon" className="size-7" aria-label="收起右方面板"><SquareChevronRight className="size-4" /></Button>
            </TooltipTrigger>
            <TooltipContent>收起右方面板</TooltipContent>
          </Tooltip>
        </div>
      </div>
      <div className="border-b bg-muted/30 px-3 py-2 text-[11px] text-muted-foreground">
        要注意的两点:① 活动页签靠 <code className="font-mono">inset 0 1px 0 0 var(--primary)</code> 与编辑区融为一体(现在叫 .tab-active),
        官方 Tabs 表达不了"页签与内容同底 + 顶部主色细线";② 关闭钮是页签内的独立按钮,Radix TabsTrigger 里塞按钮会有嵌套交互问题。
      </div>
      <pre className="overflow-x-auto bg-card p-3 font-mono text-[11px] leading-5">{SOURCE_SNIPPET}</pre>
    </Block>
  )
}

function FloatingToolbar() {
  const [collab, setCollab] = useState(false)
  return (
    <Block block={B['w-toolbar']}>
      <div className="space-y-4">
        <Demo label="浮动工具条(可拖拽手柄 + 六个动作)" hint="此预览不可拖拽;真实组件把位置记在 localStorage">
          <div className="inline-flex items-center gap-0.5 rounded-xl border bg-card px-1 py-0.5 shadow-xl">
            <span className="cursor-grab p-0.5 text-muted-foreground"><GripVertical className="size-3.5" /></span>
            <TIcon icon={StepForward} label={collab ? '协作模式下由 AI 主导' : '继续 (F5) — 运行到下一个断点'} disabled={collab} tone="sky" />
            <TIcon icon={RedoDot} label={collab ? '协作模式下由 AI 主导' : '步过 (F10)'} disabled={collab} tone="sky" />
            <TIcon icon={ArrowDownToDot} label={collab ? '协作模式下由 AI 主导' : '步入 (F11)'} disabled={collab} tone="sky" />
            <TIcon icon={ArrowUpFromDot} label={collab ? '协作模式下由 AI 主导' : '步出 (finish)'} disabled={collab} tone="sky" />
            <TIcon icon={RefreshCw} label={collab ? '协作模式下由 AI 主导' : '重新开始 — 复用会话重启同一作业'} disabled={collab} tone="green" />
            <TIcon icon={Square} label={collab ? '协作模式下由 AI 主导' : '结束调试 — 只结束本轮运行(会话保留)'} disabled={collab} tone="red" />
          </div>
        </Demo>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <Button size="xs" variant="outline" onClick={() => setCollab((v) => !v)}>{collab ? '退出协作模式' : '进入协作模式'}</Button>
          六个动作靠颜色表达语义(继续/步进=蓝、重启=绿、结束=红),官方 Button 只有 default/destructive 两档主色
          —— 要么收窄,要么与状态栏共用那套语义色 token。
        </div>
      </div>
    </Block>
  )
}

function TIcon({ icon: Icon, label, disabled, tone }: { icon: typeof StepForward; label: string; disabled?: boolean; tone: 'sky' | 'green' | 'red' }) {
  const color = tone === 'sky' ? 'text-sky-600 dark:text-sky-400' : tone === 'green' ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className="inline-flex">
          <Button variant="ghost" size="icon" className={cn('size-7', color)} disabled={disabled} aria-label={label}>
            <Icon className="size-4" />
          </Button>
        </span>
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  )
}

function AccordionPanels() {
  return (
    <Block block={B['w-accordion']}>
      <div className="max-w-md overflow-hidden rounded-xl border">
        <div className="flex h-9 items-center border-b">
          <span className="flex-1 px-2.5 text-xs font-medium text-muted-foreground">运行</span>
        </div>
        <div className="flex items-center gap-1 border-b px-1.5 py-1.5">
          <Button variant="ghost" size="icon" className="size-7 text-green-600 dark:text-green-500" aria-label="启动调试会话">
            <Play className="size-4" fill="currentColor" />
          </Button>
          <Input defaultValue="bsft001_wf" className="h-8 flex-1 text-xs" placeholder="作业编号,如 bsft001_wf 或 asf/bsft001_wf" />
          <Button variant="ghost" size="icon" className="size-7 text-sky-600 dark:text-sky-400" aria-label="行号校准"><Ruler className="size-4" /></Button>
        </div>
        <Accordion type="multiple" defaultValue={['stack', 'watches', 'bps']}>
          <PanelItem value="stack" title={`调用栈 (${STACK_FRAMES.length})`} auto>
            <div>
              {STACK_FRAMES.map((f) => (
                <div key={f.idx} className={cn('flex items-baseline gap-1.5 border-b px-2 py-1.5 text-xs last:border-b-0', f.idx === 0 && 'bg-accent/60')}>
                  <span className="text-muted-foreground">#{f.idx}</span>
                  <span className="font-mono text-sky-600 dark:text-sky-400">{f.func}</span>
                  <span className="ml-auto text-muted-foreground">{f.file}:{f.line}</span>
                </div>
              ))}
            </div>
          </PanelItem>
          <PanelItem value="autovars" title={`自动变量 (${AUTOVARS.length})`}>
            <div>
              {AUTOVARS.map((v) => (
                <div key={v.expr} className="flex items-start gap-1.5 border-b px-2 py-1.5 text-xs last:border-b-0">
                  <span className="shrink-0 text-muted-foreground">+</span>
                  <div className="min-w-0">
                    <div className="text-muted-foreground">{v.expr}</div>
                    <div className="break-all whitespace-pre-wrap text-emerald-600 dark:text-emerald-400">{v.value}</div>
                  </div>
                </div>
              ))}
            </div>
          </PanelItem>
          <PanelItem value="watches" title={`变量监视 (${3})`}>
            <div>
              <div className="border-b p-1.5">
                <Input className="h-8 text-xs" placeholder="表达式,如 lp_str / g_qryparam.*(回车求值)" />
              </div>
              <div className="flex items-start gap-1.5 border-b px-2 py-1.5 text-xs last:border-b-0">
                <span className="shrink-0 text-muted-foreground">×</span>
                <div className="min-w-0">
                  <div className="text-muted-foreground">lp_bad</div>
                  <div className="break-all text-red-600 dark:text-red-400">No symbol lp_bad in current context.</div>
                </div>
              </div>
            </div>
          </PanelItem>
          <PanelItem value="bps" title={`断点 (${BREAKPOINTS.length})`}>
            <div>
              {BREAKPOINTS.map((b) => (
                <div key={b.num} className={cn('relative flex items-center gap-2 border-b px-2 py-1.5 text-xs last:border-b-0', !b.enabled && 'opacity-50', b.hit && 'bg-amber-500/10')}>
                  {b.hit && <span className="absolute inset-y-0 left-0 w-0.5 bg-amber-500" aria-hidden />}
                  <Check checked={b.enabled} />
                  <span className="min-w-0 flex-1 truncate">
                    {b.func && <span className="font-mono text-sky-600 dark:text-sky-400">{b.func} </span>}
                    <span className="font-mono">{b.file}:{b.line}</span>
                    {!b.enabled && <span className="ml-1 text-[10px] text-muted-foreground">(已禁用)</span>}
                  </span>
                  <span className="text-muted-foreground">×</span>
                </div>
              ))}
            </div>
          </PanelItem>
        </Accordion>
      </div>
    </Block>
  )
}

function Check({ checked }: { checked: boolean }) {
  return (
    <span className={cn('flex size-4 shrink-0 items-center justify-center rounded-[4px] border', checked ? 'border-primary bg-primary text-primary-foreground' : 'border-input')}>
      {checked && <span className="text-[10px] leading-none">✓</span>}
    </span>
  )
}

function PanelItem({ value, title, auto, children }: { value: string; title: string; auto?: boolean; children: React.ReactNode }) {
  const [on, setOn] = useState(!!auto)
  return (
    <AccordionItem value={value} className="border-b last:border-b-0">
      <div className="group flex items-stretch hover:bg-accent/50">
        <button
          className={cn('flex w-6 shrink-0 items-center justify-center', on ? 'text-emerald-600 dark:text-emerald-400' : 'text-muted-foreground')}
          onClick={() => setOn((v) => !v)}
          aria-label={on ? '自动刷新已开启' : '自动刷新已关闭(默认)'}
        >
          {on ? <Eye className="size-3.5" /> : <EyeOff className="size-3.5" />}
        </button>
        <AccordionTrigger className="flex-1 px-1 py-2 text-xs font-medium text-muted-foreground hover:no-underline group-hover:text-foreground">
          {title}
        </AccordionTrigger>
      </div>
      <AccordionContent className="pb-0">{children}</AccordionContent>
    </AccordionItem>
  )
}

function Trees() {
  return (
    <Block block={B['w-tree']}>
      <div className="grid gap-6 sm:grid-cols-2">
        <Demo label="变量树(监视/自动变量共用)" hint="展开收起是纯前端折叠;产品现在手写 ▸/▾ 文本箭头">
          <div className="rounded-lg border p-2 text-xs">
            {VAR_TREE.map((n, i) => <VarNode key={i} node={n} depth={0} />)}
          </div>
        </Demo>
        <Demo label="大纲(按深度着色)" hint="顶层=foreground、一层=sky、更深=muted;点击跳行落到名称列">
          <div className="rounded-lg border">
            <div className="flex items-center justify-between border-b px-2.5 py-1.5">
              <span className="text-xs font-medium text-muted-foreground">大纲</span>
              <span className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                {OUTLINE.length} 个顶层节点
                <Button variant="ghost" size="icon" className="size-6" aria-label="全部折叠"><ChevronsDownUp className="size-3.5" /></Button>
              </span>
            </div>
            <div className="py-1 text-xs">
              {OUTLINE.map((n, i) => <OutlineRow key={i} node={n} />)}
            </div>
          </div>
        </Demo>
      </div>
    </Block>
  )
}

function VarNode({ node, depth }: { node: { name: string; value?: string; type?: string; open?: boolean; children?: unknown[] }; depth: number }) {
  const [open, setOpen] = useState(!!node.open)
  const kids = (node.children || []) as typeof node[]
  const hasKids = kids.length > 0
  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <CollapsibleTrigger asChild disabled={!hasKids}>
        <div className={cn('flex items-baseline gap-1 rounded px-1 py-0.5', hasKids && 'cursor-pointer hover:bg-accent/50')} style={{ paddingLeft: depth * 14 + 4 }}>
          <span className={cn('w-3 shrink-0 text-muted-foreground', !hasKids && 'opacity-0')}>{open ? '▾' : '▸'}</span>
          <span className="font-mono whitespace-pre text-sky-700 dark:text-sky-300">{node.name}</span>
          {node.value !== undefined && (
            <>
              <span className="whitespace-pre text-muted-foreground"> = </span>
              <span className="break-all">{node.value || '(空)'}</span>
            </>
          )}
          {node.type && <span className="ml-1 shrink-0 text-muted-foreground">{node.type}</span>}
        </div>
      </CollapsibleTrigger>
      <CollapsibleContent>
        {kids.map((c, i) => <VarNode key={i} node={c} depth={depth + 1} />)}
      </CollapsibleContent>
    </Collapsible>
  )
}

function OutlineRow({ node }: { node: { label: string; line: number; depth: number; children?: unknown[] } }) {
  const [open, setOpen] = useState(true)
  const kids = (node.children || []) as typeof node[]
  const tone = node.depth === 0 ? 'font-medium text-foreground' : node.depth === 1 ? 'text-sky-600 dark:text-sky-400' : 'text-muted-foreground'
  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <div className={cn('flex h-6 items-center gap-1 pr-2 hover:bg-accent/40', tone)} style={{ paddingLeft: node.depth * 14 + 4 }}>
        {kids.length > 0 ? (
          <CollapsibleTrigger asChild>
            <button className="w-3 shrink-0 text-muted-foreground" aria-label={open ? '收起' : '展开'}>{open ? '▾' : '▸'}</button>
          </CollapsibleTrigger>
        ) : <span className="w-3 shrink-0" />}
        <span className="min-w-0 truncate">{node.label}</span>
      </div>
      <CollapsibleContent>
        {kids.map((k, i) => <OutlineRow key={i} node={k} />)}
      </CollapsibleContent>
    </Collapsible>
  )
}

function Timeline() {
  const [cmd, setCmd] = useState('')
  const hist = useRef<string[]>(['print lp_str', 'info breakpoints'])
  const [idx, setIdx] = useState(-1)
  const onKey = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      if (cmd.trim()) hist.current = [...hist.current.filter((x) => x !== cmd.trim()), cmd.trim()].slice(-50)
      setCmd('')
      setIdx(-1)
      return
    }
    if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
      e.preventDefault()
      const h = hist.current
      if (!h.length) return
      const next = e.key === 'ArrowUp' ? (idx < 0 ? h.length - 1 : Math.max(0, idx - 1)) : (idx < 0 ? -1 : (idx + 1 >= h.length ? -1 : idx + 1))
      setIdx(next)
      setCmd(next >= 0 ? h[next] : '')
    }
  }
  const kindTone = useMemo(() => ({
    stop: 'text-amber-600 dark:text-amber-400',
    warn: 'text-red-600 dark:text-red-400',
    command: 'text-sky-600 dark:text-sky-400',
    info: 'text-muted-foreground',
  } as Record<string, string>), [])
  return (
    <Block block={B['w-timeline']} pad={false}>
      <Tabs defaultValue="raw">
        <div className="border-b px-1.5">
          <TabsList variant="line">
            <TabsTrigger value="raw">原始协议流 ({RAW_LOG.length})</TabsTrigger>
            <TabsTrigger value="timeline">操作时间线</TabsTrigger>
          </TabsList>
        </div>
        <TabsContent value="raw" className="mt-0">
          <div className="max-h-44 overflow-auto p-1 font-mono text-[11px] leading-5 text-muted-foreground">
            {RAW_LOG.map((l, i) => <div key={i} className="break-all whitespace-pre-wrap">{l}</div>)}
          </div>
          <div className="flex h-9 items-center gap-1.5 border-t px-2.5">
            <span className="shrink-0 font-mono text-xs leading-none text-sky-600 dark:text-sky-400" title="fgldb 命令直通">&gt;</span>
            {/* 控制台语义:无边框 + ↑↓ 历史 —— 换 Input 需要覆盖外观,历史键必须自己保留 */}
            <Input
              value={cmd}
              onChange={(e) => setCmd(e.target.value)}
              onKeyDown={onKey}
              placeholder="fgldb 命令,如 print lp_str / info breakpoints(↑↓ 历史),回车发送"
              className="h-7 flex-1 border-0 bg-transparent font-mono text-xs shadow-none focus-visible:ring-0 dark:bg-transparent"
            />
          </div>
        </TabsContent>
        <TabsContent value="timeline" className="mt-0">
          <div className="max-h-56 overflow-auto p-1 font-mono text-[11px] leading-6">
            {TIMELINE.map((t, i) => (
              <div key={i} className="flex items-start gap-2">
                <span className="shrink-0 text-muted-foreground">{t.time}</span>
                <Badge variant={t.origin === 'ai' ? 'default' : t.origin === 'human' ? 'secondary' : 'outline'} className="mt-0.5 shrink-0">
                  {t.origin === 'ai' ? 'AI' : t.origin === 'human' ? '用户' : '系统'}
                </Badge>
                <span className={cn('min-w-0', kindTone[t.kind])}>{t.text}</span>
              </div>
            ))}
          </div>
        </TabsContent>
      </Tabs>
    </Block>
  )
}

function EmptyStates() {
  return (
    <Block block={B['w-empty']}>
      <div className="grid gap-6 sm:grid-cols-3">
        <Demo label="中央空状态(未连接会话)">
          <Empty className="border-0 p-0">
            <EmptyHeader>
              <EmptyTitle>未连接会话</EmptyTitle>
              <EmptyDescription>
                接口日志按当前会话所属环境的数据库查询。请先在「会话」面板选择一个环境连接,再回到这里查看。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        </Demo>
        <Demo label="加载骨架(加载配置中…)">
          <div className="space-y-2">
            <Skeleton className="h-4 w-2/3" />
            <Skeleton className="h-4 w-1/2" />
            <Skeleton className="h-20 w-full" />
          </div>
        </Demo>
        <Demo label="忙碌遮罩(启动调试会话…)">
          <div className="relative flex h-32 items-center justify-center rounded-lg border bg-card">
            <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center gap-2 bg-background/60">
              <Loader2 className="size-6 animate-spin text-muted-foreground" />
              <span className="text-xs text-muted-foreground">正在启动调试会话…</span>
            </div>
          </div>
        </Demo>
      </div>
      <Separator className="my-4" />
      <div className="text-xs text-muted-foreground">
        三处空状态文案与形态各异(居中块 / 居中卡片 / 左对齐一行)。可统一到官方 Empty,但「该模块无源码(仅 42m),当前停站: xxx.4gl:412」这类信息必须逐字保留。
      </div>
    </Block>
  )
}
