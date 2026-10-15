import { useState } from 'react'
import { Copy, RedoDot, Square, StepForward, ArrowDownToDot, ArrowUpFromDot } from 'lucide-react'
import { Block, Demo } from '@/components/Block'
import { blocksOf } from '@/lib/ledger'
import { SOURCE_SNIPPET, TIMELINE, VAR_TREE } from '@/lib/fixtures'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { cn } from 'cn'

const B = blocksOf('patterns')

export function PatternsSection() {
  return (
    <div className="space-y-6">
      <MonacoEditor />
      <HoverCard />
      <Gutter />
      <SplitSash />
      <CollabGate />
    </div>
  )
}

/** 静态示意:演示台不加载 Monaco(它有自己的主题与 worker,不属于"换 shadcn 组件"的范围) */
function MonacoEditor() {
  const lines = SOURCE_SNIPPET.split('\n')
  return (
    <Block block={B['p-monaco']} pad={false}>
      <div className="grid gap-4 p-4 lg:grid-cols-2">
        <Demo label="编辑器区(静态示意)">
          <div className="overflow-hidden rounded-lg border">
            <div className="flex font-mono text-[11px] leading-5">
              <div className="w-10 shrink-0 select-none border-r bg-muted/40 py-2 text-right text-muted-foreground">
                {lines.map((_, i) => <div key={i} className="pr-2">{412 + i}</div>)}
              </div>
              <div className="min-w-0 flex-1 overflow-x-auto py-2">
                {lines.map((l, i) => (
                  <div key={i} className={cn('whitespace-pre px-3', i === 6 && 'bg-amber-400/15')}>
                    {l.replace(/^ {2}/, '  ')}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </Demo>
        <div className="space-y-3 text-xs text-muted-foreground">
          <p><b>为什么不能用 shadcn 换</b>:Monaco 自带主题、worker 与布局,只有外层容器是 React。</p>
          <p><b>必须保住的四件事</b>:</p>
          <ul className="list-disc space-y-1 pl-5">
            <li>编辑器底色必须等于容器底色 —— 产品里 <code className="font-mono">--card</code> 就是 <code className="font-mono">editor.background</code>;换成 shadcn 默认 token 后要重新对齐(现在 card 暗色是 oklch(0.205 0 0),Monaco 主题里写的是 #1f1f1f)。</li>
            <li><code className="font-mono">editorOptions</code> 引用必须稳定,否则加断点这类重渲染会把滚动位置复位。</li>
            <li><code className="font-mono">automaticLayout</code> 要求父链高度确定(Card 的默认 gap/py 会顶出高度)。</li>
            <li><code className="font-mono">fixedOverflowWidgets: true</code> 是悬停卡片越界时的兜底,不能关。</li>
          </ul>
        </div>
      </div>
    </Block>
  )
}

function HoverCard() {
  const [hover, setHover] = useState(false)
  return (
    <Block block={B['p-hover']}>
      <div className="grid gap-4 lg:grid-cols-2">
        <Demo label="变量悬浮卡片(静态示意,鼠标移入显示复制钮)">
          <div
            className="relative w-fit min-w-36 max-w-[560px] rounded-lg border bg-card text-xs"
            onMouseEnter={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
          >
            <Button
              variant="ghost"
              size="icon"
              className={cn('absolute top-1 right-1 z-10 size-6 transition-opacity', hover ? 'opacity-100' : 'opacity-0')}
              aria-label="复制原始值"
            >
              <Copy className="size-3" />
            </Button>
            <div className="max-h-80 overflow-auto px-2.5 py-1.5">
              {VAR_TREE.map((n, i) => <VarNode key={i} node={n} depth={0} />)}
            </div>
          </div>
        </Demo>
        <div className="space-y-3 text-xs text-muted-foreground">
          <p><b>为什么不能用 Radix</b>:卡片是 Monaco 的 content widget(渲染进编辑器挂载节点,不在 React 树里),
            没有锚点 DOM、也无法与编辑器滚动同步;Portal 出来的浮层会被 minimap(z-index 5)盖住。</p>
          <p>能做的只有把 <code className="font-mono">.fgl-hover*</code> 里的写死值 token 化:
            底色 <code className="font-mono">--card</code>、边框 <code className="font-mono">--border</code>(这两条已经对了)、
            错误文案的 <code className="font-mono">text-red-400</code>、以及阴影 <code className="font-mono">rgba(0,0,0,0.35)</code>(暗色下偏重)。</p>
          <p>卡片还有两处手工逻辑不能丢:宽度夹紧(避开 minimap 与行号槽)与滚轮接管(长值只滚卡片、不滚代码)。</p>
        </div>
      </div>
    </Block>
  )
}

function VarNode({ node, depth }: { node: { name: string; value?: string; type?: string; open?: boolean; children?: unknown[] }; depth: number }) {
  const [open, setOpen] = useState(!!node.open)
  const kids = (node.children || []) as typeof node[]
  const hasKids = kids.length > 0
  return (
    <div>
      <div
        className={cn('flex items-baseline gap-1 rounded px-1 py-0.5', hasKids && 'cursor-pointer hover:bg-accent/50')}
        style={{ paddingLeft: depth * 14 + 4 }}
        onClick={() => hasKids && setOpen((v) => !v)}
      >
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
      {open && kids.map((c, i) => <VarNode key={i} node={c} depth={depth + 1} />)}
    </div>
  )
}

function Gutter() {
  return (
    <Block block={B['p-gutter']}>
      <div className="grid gap-4 lg:grid-cols-2">
        <Demo label="现状:写死色(两处来源)" hint="#ef4444 同时写在 CSS 与 overviewRuler 里;#eab308 同理">
          <GutterMock dotClass="bg-[#ef4444]" arrowClass="bg-[#eab308]" lineClass="bg-[#eab308]/12" />
        </Demo>
        <Demo label="建议:token 化(红=destructive,黄=warning 语义色)" hint="尺子上与行上取同一个 token,才不会只改一半">
          <GutterMock dotClass="bg-destructive" arrowClass="bg-amber-500" lineClass="bg-amber-500/12" />
        </Demo>
      </div>
      <div className="mt-3 text-xs text-muted-foreground">
        gutter 装饰是命令式 CSS 类(断点圆点 / 停站箭头 / 停站行高亮),挂在不在 Tailwind 扫描范围内的 index.css 上。
        另外 <code className="font-mono">.bp-dot.cur-arrow</code> 必须<b>合写</b>覆盖,否则箭头的 background 简写会把红点整个吃掉。
      </div>
    </Block>
  )
}

function GutterMock({ dotClass, arrowClass, lineClass }: { dotClass: string; arrowClass: string; lineClass: string }) {
  const lines = ['FUNCTION wf_confirm()', '  DEFINE lp_str LIKE wo_m.str,', '  SELECT COUNT(*) INTO li_cnt']
  return (
    <div className="overflow-hidden rounded-lg border font-mono text-[11px] leading-6">
      {lines.map((l, i) => (
        <div key={i} className={cn('flex items-center', i === 1 && lineClass)}>
          <span className="flex w-10 shrink-0 items-center justify-end gap-1 pr-1 text-muted-foreground">
            {i === 1 && <span className={cn('size-2.5 rounded-full', dotClass)} />}
            {i === 1 && <span className={cn('size-2', arrowClass)} />}
            412
          </span>
          <span className="whitespace-pre">{l}</span>
        </div>
      ))}
    </div>
  )
}

function SplitSash() {
  return (
    <Block block={B['p-split']}>
      <div className="grid gap-4 lg:grid-cols-2">
        <Demo label="比例式分栏 + 5px 拖拽条" hint="鼠标移上去看拖拽条变主色加粗">
          <div className="flex h-40 overflow-hidden rounded-lg border">
            <div className="min-w-0 flex-1 overflow-hidden bg-background p-2 text-[11px] text-muted-foreground">
              日志列表(min-w-0 flex-1 overflow-hidden)
            </div>
            <div
              className="group flex w-[5px] shrink-0 cursor-col-resize items-stretch bg-[linear-gradient(var(--border),var(--border))] bg-[length:1px_100%] bg-center bg-no-repeat hover:bg-[linear-gradient(var(--ring),var(--ring))] hover:bg-[length:2px_100%]"
              title="拖拽分配宽度"
            />
            <div className="w-2/5 shrink-0 overflow-hidden border-l bg-card p-2 text-[11px] text-muted-foreground">
              报文详情(shrink-0,比例 50%)
            </div>
          </div>
        </Demo>
        <div className="space-y-3 text-xs text-muted-foreground">
          <p><b>换官方 Resizable 会丢掉的四样</b>:比例持久化键、比例夹取区间(25%~75%)、5px 占位、以及
            <code className="font-mono">hover/dragging</code> 时主色加粗的提示。</p>
          <p>还有一条隐形锁:左栏必须 <code className="font-mono">min-w-0 flex-1 overflow-hidden</code>、详情栏必须 <code className="font-mono">shrink-0</code>,
            否则会形成宽度反馈环,整页横向溢出。</p>
          <p>结论:这类交互收益低、风险高,建议保持手写,只在视觉上对齐 token。</p>
        </div>
      </div>
    </Block>
  )
}

function CollabGate() {
  const [collab, setCollab] = useState(true)
  const tip = collab ? '协作模式下由 AI 主导' : ''
  return (
    <Block block={B['p-collab']}>
      <div className="space-y-4">
        <Demo label="协作模式:界面不装作能做(服务端同样 403)">
          <div className="flex flex-wrap items-center gap-2">
            <Button size="xs" variant="outline" onClick={() => setCollab((v) => !v)}>{collab ? '切回纯人工' : '交给 AI 主导'}</Button>
            {[StepForward, RedoDot, ArrowDownToDot, ArrowUpFromDot].map((Icon, i) => (
              <Tooltip key={i}>
                <TooltipTrigger asChild>
                  <span className="inline-flex">
                    <Button size="icon" variant="ghost" disabled={collab} aria-label="步进"><Icon className="size-4" /></Button>
                  </span>
                </TooltipTrigger>
                <TooltipContent>{tip || '继续 (F5) — 运行到下一个断点'}</TooltipContent>
              </Tooltip>
            ))}
            <Tooltip>
              <TooltipTrigger asChild>
                <span className="inline-flex"><Button size="icon" variant="ghost" className="text-red-600 dark:text-red-400" disabled={collab} aria-label="结束调试"><Square className="size-4" /></Button></span>
              </TooltipTrigger>
              <TooltipContent>{tip || '结束调试 — 只结束本轮运行(会话保留)'}</TooltipContent>
            </Tooltip>
            <Separator orientation="vertical" className="mx-1 h-5" />
            <span className="flex items-center gap-2 text-[11px]">
              <Badge variant="secondary" title="本次调试由 AI 主导:你只能观察与取值">协作模式 · AI 主导</Badge>
              <span className="text-muted-foreground">接管</span>
            </span>
          </div>
        </Demo>
        <div className="text-xs text-muted-foreground">
          协作模式下断点、选帧、命令、TOPENT、自动求值一律禁用 —— 这是<b>行为契约</b>而非样式:
          任何"顺手修好"的改造都会让界面撒谎。最近的一处真话在时间线里:
          {TIMELINE.filter((t) => t.origin === 'ai').map((t) => ` ${t.time} ${t.text}`)}
        </div>
      </div>
    </Block>
  )
}
