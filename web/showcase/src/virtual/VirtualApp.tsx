// 目标态虚拟应用:把"我推荐的最佳实践"组装成一台能点、能切视图的整机。
//
// 它回答的不是"某个组件长什么样"(那是台账屏的事),而是:
//   编辑器式的工作台 × shadcn 默认明暗双色 × 官方尺寸变体,拼起来到底成不成立。
//
// 与产品现状的差别,只有列在左栏"已定的最佳实践"里的那几条(其余一律是 shadcn 默认)。
import { useState } from 'react'
import { useTheme } from 'next-themes'
import { toast } from 'sonner'
import {
  ArrowDownToDot, ArrowUpFromDot, Bug, FlaskConical, Globe, GripVertical, Monitor, Moon,
  RedoDot, Settings, Square, StepForward, Sun,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Toaster } from '@/components/ui/sonner'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { DebugView, SessionRibbon } from './DebugView'
import { LogsView } from './LogsView'
import { TestView } from './TestView'
import { SettingsView } from './SettingsView'
import { ConfirmProvider } from './useConfirm'
import { StatusBadge } from './status'
import { SESSION } from '@/lib/fixtures'
import { SESSION_STATES, type SessionState } from './mock'
import { cn } from 'cn'

export type ViewKey = 'debug' | 'logs' | 'test' | 'settings'

const VIEWS: { key: ViewKey; label: string; icon: typeof Bug }[] = [
  { key: 'debug', label: '调试', icon: Bug },
  { key: 'logs', label: '接口日志', icon: Globe },
  { key: 'test', label: '服务测试', icon: FlaskConical },
  { key: 'settings', label: '设置(站点 / 数据字典 / DEBUG / 应用)', icon: Settings },
]

export function VirtualApp({ initialView = 'debug', initialSettingsSection = 'sites' }: {
  initialView?: ViewKey
  initialSettingsSection?: 'sites' | 'data-dict' | 'debug' | 'app'
} = {}) {
  const [view, setView] = useState<ViewKey>(initialView)
  const [state, setState] = useState<SessionState>('stopped')
  const [collab, setCollab] = useState(false)
  const onToast = (m: string) => toast.success(m, { duration: 2200 })
  const stopped = state === 'stopped'
  const inRun = state !== 'idle'

  return (
    <ConfirmProvider>
      <div className="relative flex h-[820px] flex-col overflow-hidden rounded-xl border bg-background shadow-2xl">
        {/* 窗口标题栏:虚拟应用的"外壳",不属于产品 UI(真实产品就是浏览器里的页面) */}
        <div className="flex shrink-0 flex-wrap items-center gap-2 border-b bg-card px-3 py-2">
          <Badge variant="outline" className="gap-1.5"><span className="size-2 rounded-full bg-success" /> TT · 目标态预览</Badge>
          <span className="text-[11px] text-muted-foreground">纯官方 shadcn 组件 · 默认明暗双色 · 数据为虚拟夹具</span>
          <div className="ml-auto flex flex-wrap items-center gap-2">
            <span className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
              会话
              <Select value={state} onValueChange={(v) => setState(v as SessionState)}>
                <SelectTrigger size="sm" className="w-32"><SelectValue /></SelectTrigger>
                <SelectContent>
                  {SESSION_STATES.map((s) => (
                    <SelectItem key={s.key} value={s.key} className="text-xs">
                      {s.label} · {s.hint}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </span>
            <span className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
              <Switch id="va-collab" checked={collab} onCheckedChange={setCollab} />
              <Label htmlFor="va-collab" className="font-normal">协作模式</Label>
            </span>
            <ThemePicker />
          </div>
        </div>

        {/* 主体:左侧图标栏 + 视图 + (调试视图的)右侧面板 */}
        <div className="flex min-h-0 flex-1">
          <div className="flex w-12 shrink-0 flex-col items-center gap-1 border-r bg-sidebar py-3">
            {VIEWS.map(({ key, label, icon: Icon }) => (
              <Tooltip key={key}>
                <TooltipTrigger asChild>
                  <Button
                    variant={view === key ? 'secondary' : 'ghost'}
                    size="icon"
                    aria-label={label}
                    aria-pressed={view === key}
                    className="size-9"
                    onClick={() => setView(key)}
                  >
                    <Icon className="size-4.5" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="right">{label}</TooltipContent>
              </Tooltip>
            ))}
            <Separator className="my-1 w-6" />
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon" className="size-9" aria-label="会话" onClick={() => setView('debug')}>
                  <Badge variant="secondary" className="size-5 justify-center rounded-full p-0 text-[10px]">3</Badge>
                </Button>
              </TooltipTrigger>
              <TooltipContent side="right">会话(已配置 3 个环境)</TooltipContent>
            </Tooltip>
          </div>

          <div className="flex min-h-0 min-w-0 flex-1 flex-col">
            {view === 'debug' && <SessionRibbon state={state} collab={collab} />}
            {view === 'debug' && <DebugView state={state} collab={collab} onToast={onToast} />}
            {view === 'logs' && <LogsView onToast={onToast} sessionReady={state !== 'loading'} />}
            {view === 'test' && <TestView onToast={onToast} wsLogService={view === 'test' ? 'icd.erp.wo.get' : undefined} />}
            {view === 'settings' && <SettingsView onToast={onToast} initialSection={initialSettingsSection} />}
          </div>
        </div>

        {/* 底部状态栏:整条底色表达会话状态 —— 这正是"必须补三个语义色"的理由 */}
        <StatusBar state={state} collab={collab} onToggleCollab={() => setCollab((v) => !v)} stopped={stopped} inRun={inRun} />

        {/* 浮动调试工具条:只在调试视图出现,可拖拽(此处为静态定位示意) */}
        {view === 'debug' && (
          <div className="absolute top-24 right-[21.5rem] z-20 inline-flex items-center gap-0.5 rounded-lg border bg-card/95 px-1 py-0.5 shadow-xl">
            <span className="cursor-grab p-0.5 text-muted-foreground" title="拖拽移动工具条"><GripVertical className="size-3.5" /></span>
            <TIcon icon={StepForward} label={collab ? '协作模式下由 AI 主导' : '继续 (F5) — 运行到下一个断点'} disabled={collab || !stopped} tone="info" onClick={() => onToast('continue:运行到下一个断点')} />
            <TIcon icon={RedoDot} label={collab ? '协作模式下由 AI 主导' : '步过 (F10)'} disabled={collab || !stopped} tone="info" onClick={() => onToast('next:步过')} />
            <TIcon icon={ArrowDownToDot} label={collab ? '协作模式下由 AI 主导' : '步入 (F11)'} disabled={collab || !stopped} tone="info" onClick={() => onToast('step:步入')} />
            <TIcon icon={ArrowUpFromDot} label={collab ? '协作模式下由 AI 主导' : '步出 (finish)'} disabled={collab || !stopped} tone="info" onClick={() => onToast('finish:步出')} />
            <TIcon icon={RedoDot} label={collab ? '协作模式下由 AI 主导' : '重新开始 — 复用会话重启同一作业'} disabled={collab || !inRun} tone="success" onClick={() => onToast('restart:复用会话重启同一作业')} />
            <TIcon icon={Square} label={collab ? '协作模式下由 AI 主导' : '结束调试 — 只结束本轮运行(会话保留)'} disabled={collab || !inRun} tone="danger" onClick={() => onToast('quit:只结束本轮运行')} />
          </div>
        )}
      </div>
      <Toaster position="bottom-right" richColors closeButton />
    </ConfirmProvider>
  )
}

function ThemePicker() {
  const { theme, setTheme } = useTheme()
  const items = [
    { key: 'light', label: '亮色', icon: Sun },
    { key: 'dark', label: '暗色', icon: Moon },
    { key: 'system', label: '跟随系统', icon: Monitor },
  ]
  return (
    <div className="flex items-center gap-0.5 rounded-lg border p-0.5">
      {items.map(({ key, label, icon: Icon }) => (
        <Button
          key={key}
          variant={theme === key ? 'secondary' : 'ghost'}
          size="icon"
          className="size-6"
          aria-label={label}
          aria-pressed={theme === key}
          title={label}
          onClick={() => setTheme(key)}
        >
          <Icon className="size-3.5" />
        </Button>
      ))}
    </div>
  )
}

function TIcon({ icon: Icon, label, disabled, tone, onClick }: {
  icon: typeof StepForward; label: string; disabled?: boolean
  tone: 'info' | 'success' | 'danger'; onClick: () => void
}) {
  const color = tone === 'info' ? 'text-info' : tone === 'success' ? 'text-success' : 'text-destructive'
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className="inline-flex">
          <Button variant="ghost" size="icon" className={cn('size-7', color)} disabled={disabled} aria-label={label} onClick={onClick}>
            <Icon className="size-4" />
          </Button>
        </span>
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  )
}

function StatusBar({ state, collab, stopped, inRun, onToggleCollab }: {
  state: SessionState; collab: boolean; stopped: boolean; inRun: boolean; onToggleCollab: () => void
}) {
  // 语义色:停站=warning、运行=info、启动中=secondary、空闲=muted(现状用的橙/蓝由同一套 token 表达)
  const bar =
    state === 'stopped' ? 'bg-warning text-warning-foreground'
      : state === 'running' ? 'bg-info text-info-foreground'
        : state === 'loading' ? 'bg-secondary text-secondary-foreground'
          : 'bg-muted text-muted-foreground'
  const chip = 'shrink-0 rounded-sm bg-black/15 px-1.5 py-px'
  const msg =
    state === 'stopped' ? `已停站 ${SESSION.file}:${SESSION.line} · 停留 ${SESSION.holdSeconds}s`
      : state === 'running' ? (collab ? 'AI 已放行 · 在界面上观察即可' : '运行中 · 请到 GDC 操作')
        : state === 'loading' ? '启动中…'
          : '会话空闲(36正式) · 可直接启动调试'
  return (
    <div className={cn('flex h-7 shrink-0 items-center gap-3 px-3 text-[11px] transition-colors', bar)}>
      <span className="min-w-0 flex-1 truncate font-mono">{SESSION.env} · {SESSION.file}{stopped ? `:${SESSION.line}` : ''}</span>
      <span className="flex shrink-0 items-center gap-2">
        {collab && <span className={chip} title="本次调试由 AI 主导:你只能观察与取值">协作模式 · AI 主导</span>}
        {inRun && <span className={cn(chip, 'font-mono')} title="正在执行的调试命令">⟳ continue(已 12s)</span>}
        <button
          className="shrink-0 rounded-sm border border-current/40 px-1.5 py-px opacity-80 hover:bg-black/10 hover:opacity-100"
          onClick={onToggleCollab}
          title={collab ? '切回纯人工:界面恢复全部写操作' : '把本次调试交给 AI 主导:你转为只读观察'}
        >
          {collab ? '接管' : '交给 AI'}
        </button>
        <span>{msg}</span>
        <StatusBadge tone="neutral" className="bg-black/15 text-current">{SESSION.prog}</StatusBadge>
      </span>
    </div>
  )
}
