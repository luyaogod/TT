import { useEffect, useRef, useState } from 'react'
import { useTheme } from 'next-themes'
import { BookOpen, LayoutList, Monitor, Moon, Sun } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { CATALOG } from '@/lib/catalog'
import { riskCounts } from '@/lib/ledger'
import { Overview } from '@/sections/Overview'
import { FoundationSection } from '@/sections/Foundation'
import { WorkbenchSection } from '@/sections/Workbench'
import { InterfacesSection } from '@/sections/Interfaces'
import { SettingsSection } from '@/sections/Settings'
import { PatternsSection } from '@/sections/Patterns'
import { VirtualApp, type ViewKey } from '@/virtual/VirtualApp'
import { PRACTICES } from '@/virtual/BestPractices'
import { cn } from 'cn'

type Mode = 'virtual' | 'ledger'

const SECTIONS: { id: string; title: string; node: () => React.ReactNode }[] = [
  { id: 'overview', title: '决策总览', node: () => null },
  { id: 'foundation', title: '基元与语义色', node: () => <FoundationSection /> },
  { id: 'workbench', title: '调试工作台外壳', node: () => <WorkbenchSection /> },
  { id: 'interfaces', title: '接口日志与服务测试', node: () => <InterfacesSection /> },
  { id: 'settings', title: '统一设置页', node: () => <SettingsSection /> },
  { id: 'patterns', title: '必须手写的部分', node: () => <PatternsSection /> },
]

export function App() {
  // 分享链接:?mode=ledger 看台账屏;?view=debug|logs|test|settings 直接落到虚拟应用的某个视图
  const q = new URLSearchParams(window.location.search)
  const [mode, setMode] = useState<Mode>(q.get('mode') === 'ledger' ? 'ledger' : 'virtual')
  const [virtualView] = useState<ViewKey>(() => {
    const v = q.get('view')
    return v === 'logs' || v === 'test' || v === 'settings' ? v : 'debug'
  })
  const [current, setCurrent] = useState('overview')
  const pendingAnchor = useRef<string | null>(null)
  const scrollRef = useRef<HTMLDivElement | null>(null)

  const jump = (sectionId: string, blockId?: string) => {
    pendingAnchor.current = blockId ?? null
    setCurrent(sectionId)
    try {
      history.replaceState(null, '', blockId ? `#${blockId}` : `#${sectionId}`)
    } catch { /* 忽略 */ }
  }

  // 跳转后把目标块滚进视口(等这一屏渲染完)
  useEffect(() => {
    const id = pendingAnchor.current
    if (!id) {
      scrollRef.current?.scrollTo({ top: 0 })
      return
    }
    pendingAnchor.current = null
    const t = window.setTimeout(() => {
      document.getElementById(id)?.scrollIntoView({ block: 'start' })
    }, 40)
    return () => window.clearTimeout(t)
  }, [current])

  // 首次进入按 URL 片段落到对应屏
  useEffect(() => {
    const h = location.hash.replace(/^#/, '')
    if (!h) return
    const owner = CATALOG.find((s) => s.blocks.some((b) => b.id === h))
    if (owner) {
      pendingAnchor.current = h
      setCurrent(owner.id)
    } else if (SECTIONS.some((s) => s.id === h)) {
      setCurrent(h)
    }
  }, [])

  const counts = riskCounts()
  const section = SECTIONS.find((s) => s.id === current)!
  const catalog = CATALOG.find((s) => s.id === current)

  return (
    <div className="flex h-screen bg-background text-foreground">
      <aside className="flex w-72 shrink-0 flex-col border-r bg-sidebar text-sidebar-foreground">
        <div className="space-y-1 px-4 py-4">
          <div className="flex items-center gap-2">
            <BookOpen className="size-4" />
            <span className="text-sm font-semibold">TT 前端 UI 决策演示台</span>
          </div>
          <p className="text-[11px] leading-4 text-muted-foreground">
            纯官方 shadcn 组件 + 默认明暗双色主题。这是"看与决策"的页面,不是产品的一部分。
          </p>
        </div>
        <Separator />
        {/* 两种看法:整机效果 / 逐块台账 */}
        <div className="grid grid-cols-2 gap-1 p-2">
          <Button
            variant={mode === 'virtual' ? 'secondary' : 'ghost'}
            size="sm"
            className="justify-center text-xs"
            onClick={() => setMode('virtual')}
          >
            目标态整机
          </Button>
          <Button
            variant={mode === 'ledger' ? 'secondary' : 'ghost'}
            size="sm"
            className="justify-center text-xs"
            onClick={() => setMode('ledger')}
          >
            逐块台账
          </Button>
        </div>
        <Separator />
        <nav className="min-h-0 flex-1 overflow-auto p-2">
          {mode === 'virtual' ? (
            <div className="space-y-3 px-1 py-1">
              <div className="text-[11px] font-medium text-muted-foreground">已定的最佳实践(共 {PRACTICES.length} 条)</div>
              {PRACTICES.map((p, i) => (
                <div key={p.title} className="space-y-0.5">
                  <div className="flex gap-1.5 text-[11px] font-medium">
                    <span className="text-muted-foreground">{i + 1}.</span>
                    <span>{p.title}</span>
                  </div>
                  <p className="pl-4 text-[11px] leading-4 text-muted-foreground">{p.why}</p>
                </div>
              ))}
            </div>
          ) : (
            <>
              <Button
                variant={current === 'overview' ? 'secondary' : 'ghost'}
                size="sm"
                className="w-full justify-start"
                onClick={() => jump('overview')}
              >
                <LayoutList className="size-3.5" /> 决策总览
              </Button>
              <Separator className="my-2" />
              {CATALOG.map((s) => (
                <div key={s.id} className="mb-1">
                  <Button
                    variant={current === s.id ? 'secondary' : 'ghost'}
                    size="sm"
                    className="w-full justify-start text-left"
                    onClick={() => jump(s.id)}
                  >
                    <span className="truncate">{s.title}</span>
                    <Badge variant="outline" className="ml-auto shrink-0">{s.blocks.length}</Badge>
                  </Button>
                  {current === s.id && (
                    <div className="mt-0.5 mb-1 ml-2 space-y-0.5 border-l pl-2">
                      {s.blocks.map((b) => (
                        <button
                          key={b.id}
                          className="block w-full truncate rounded px-1.5 py-1 text-left text-[11px] text-muted-foreground hover:bg-accent/60 hover:text-foreground"
                          onClick={() => jump(s.id, b.id)}
                          title={b.title}
                        >
                          {b.title}
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </>
          )}
        </nav>
        <Separator />
        <div className="space-y-2 p-3">
          <ThemeToggle />
          <div className="flex flex-wrap items-center gap-1.5 text-[11px] text-muted-foreground">
            <Badge variant="outline">低 {counts.low}</Badge>
            <Badge variant="secondary">中 {counts.mid}</Badge>
            <Badge variant="destructive">高 {counts.high}</Badge>
            <span>共 {counts.total} 个功能点</span>
          </div>
        </div>
      </aside>

      <main className="flex min-w-0 flex-1 flex-col">
        <header className="flex flex-wrap items-center gap-3 border-b px-6 py-3">
          <div className="min-w-0">
            <h1 className="text-sm font-semibold">
              {mode === 'virtual' ? '目标态虚拟应用' : section.title}
            </h1>
            {mode === 'virtual' ? (
              <p className="text-[11px] text-muted-foreground">
                按左栏那几条最佳实践组装的整机预览:官方组件、默认 token、官方尺寸变体。数据是虚拟夹具,交互可点。
              </p>
            ) : catalog ? (
              <p className="text-[11px] text-muted-foreground">{catalog.question}</p>
            ) : (
              <p className="text-[11px] text-muted-foreground">这一屏回答:要不要换、换成什么、按什么顺序换。</p>
            )}
          </div>
          <div className="ml-auto flex items-center gap-2">
            <span className="text-[11px] text-muted-foreground">主题</span>
            <ThemeToggle compact />
          </div>
        </header>
        <div ref={scrollRef} className={cn('min-h-0 flex-1 overflow-auto', mode === 'virtual' ? 'p-4' : 'px-6 py-5')}>
          {mode === 'virtual' ? (
            <VirtualApp initialView={virtualView} />
          ) : (
            <div className="mx-auto max-w-6xl pb-16">
              {current === 'overview' ? <Overview onJump={jump} /> : section.node()}
            </div>
          )}
        </div>
      </main>
    </div>
  )
}

function ThemeToggle({ compact = false }: { compact?: boolean }) {
  const { theme, setTheme } = useTheme()
  const items = [
    { key: 'light', label: '亮色', icon: Sun },
    { key: 'dark', label: '暗色', icon: Moon },
    { key: 'system', label: '跟随系统', icon: Monitor },
  ]
  return (
    <div className={cn('flex items-center gap-1', !compact && 'w-full')}>
      {items.map(({ key, label, icon: Icon }) => (
        <Tooltip key={key}>
          <TooltipTrigger asChild>
            <Button
              variant={theme === key ? 'secondary' : 'ghost'}
              size="sm"
              className={cn('gap-1.5', !compact && 'flex-1')}
              onClick={() => setTheme(key)}
              aria-pressed={theme === key}
            >
              <Icon className="size-3.5" />
              {!compact && <span className="text-[11px]">{label}</span>}
            </Button>
          </TooltipTrigger>
          <TooltipContent>{label}</TooltipContent>
        </Tooltip>
      ))}
    </div>
  )
}
