// 桌面版的标题栏：原生标题栏与菜单在桌面版里是**收掉**的（见 desktop/src/main.js），
// 由这一条自己画。
//
// 为什么不用系统标题栏：Windows 的标题栏与菜单栏只跟**系统**的明暗走，而 tt 的主题是
// **应用内**的设置（默认暗色）。系统是浅色时，深色的工作面顶着一张白条 —— 那不是"简洁"，
// 是割裂。所以这里画一条跟着 html.dark 走的栏，右上角三颗按钮仍是**原生**的
// （titleBarOverlay），配色调给主进程（desktopShell.setTheme）——不自己画按钮，
// 才能保住 Windows 11 的最大化贴靠与高 DPI 下的正确绘制。
//
// 浏览器版没有 window.ttDesktopShell，这个组件直接不渲染。
//
// 这一条栏还负责「收起/展开左侧目录」这唯一一个入口：它钉在窗口最左上角，侧栏收起与否
// 都在同一个位置（不像从前那样"展开时按钮在侧栏里、收起后跑到主面板标题行上"）。
// 侧栏与主面板因此都不再管侧栏的收展，那个状态只归外壳（App.tsx 的 collapsed）。
import { useState, type CSSProperties } from 'react'
import { PanelLeft } from 'lucide-react'
import { Popover, PopoverContent, PopoverTrigger } from '../../shared/ui-radix'
import { desktopShell } from './desktopShell'
import { cn } from '../../shared/utils'

/** 与主进程 titleBarOverlay.height 必须一致；CSS 的 h-9 也是 36px。 */
const BAR_HEIGHT = 36

// -webkit-app-region 不在 React 的 CSSProperties 里，只能这样塞。
// 拖拽区只能盖在"没有交互"的地方：按钮自己标 no-drag，否则点不动。
const DRAG = { WebkitAppRegion: 'drag' } as CSSProperties
const NO_DRAG = { WebkitAppRegion: 'no-drag' } as CSSProperties

type Entry =
  | { kind: 'item'; label: string; hint?: string; run: () => void }
  | { kind: 'sep' }

function MenuItem({ entry, close }: { entry: Extract<Entry, { kind: 'item' }>; close: () => void }) {
  return (
    <button
      role="menuitem"
      onClick={() => { close(); entry.run() }}
      className="flex w-full items-center gap-6 rounded-ui px-2.5 py-1.5 text-left text-[12px] text-foreground transition-colors hover:bg-accent"
    >
      <span className="flex-1 whitespace-nowrap">{entry.label}</span>
      {entry.hint && <span className="text-[11px] text-muted-foreground">{entry.hint}</span>}
    </button>
  )
}

function Menu({ label, entries }: { label: string; entries: Entry[] }) {
  const [open, setOpen] = useState(false)
  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          style={NO_DRAG}
          className={cn(
            'rounded-ui px-2 py-1 text-[12px] transition-colors',
            open ? 'bg-accent text-foreground' : 'text-muted-foreground hover:bg-accent/70 hover:text-foreground',
          )}
        >
          {label}
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" sideOffset={6} className="min-w-[196px] border-border bg-popover p-1 shadow-lg">
        {entries.map((e, i) =>
          e.kind === 'sep'
            ? <div key={`sep-${i}`} className="my-1 h-px bg-border" />
            : <MenuItem key={e.label} entry={e} close={() => setOpen(false)} />,
        )}
      </PopoverContent>
    </Popover>
  )
}

export function DesktopTitleBar({ collapsed, onToggleSidebar }: {
  /** 左侧目录是否已收起（只用于按钮的提示文案，状态本身在外壳里）。 */
  collapsed: boolean
  onToggleSidebar: () => void
}) {
  const shell = desktopShell()
  if (!shell) return null

  const fileMenu: Entry[] = [
    { kind: 'item', label: '在浏览器中打开', run: () => shell.openInBrowser() },
    { kind: 'item', label: '打开数据目录', run: () => shell.openDataDir() },
    { kind: 'sep' },
    { kind: 'item', label: '退出', run: () => shell.quit() },
  ]
  // 菜单里**没有全屏**：F11 是调试台的「步入」（see App.tsx 的调试快捷键），
  // 留着这一项等于把同一个键解释成两件事，而且桌面版真的很少需要全屏。
  const viewMenu: Entry[] = [
    { kind: 'item', label: '重新加载', hint: 'F5', run: () => shell.reload() },
    { kind: 'item', label: '实际大小', hint: 'Ctrl+0', run: () => shell.resetZoom() },
  ]
  const helpMenu: Entry[] = [
    { kind: 'item', label: '检查更新', run: () => shell.checkUpdate() },
    { kind: 'item', label: '打开发布页', run: () => shell.openReleases() },
    { kind: 'sep' },
    { kind: 'item', label: '关于 TT Desktop', run: () => shell.about() },
  ]

  return (
    <div
      style={{ ...DRAG, height: BAR_HEIGHT }}
      className="flex shrink-0 select-none items-center gap-1 border-b border-border bg-card pl-1.5 pr-[148px]"
    >
      {/* 不放假品牌块：窗口按钮右边那三颗系统的已经说明"这是什么应用"，
          再挂一个 TT 字标只是把界面又缩小一寸。拖拽区靠右边的空白（整条都是 drag）。 */}
      {/* 收起/展开左侧目录：整窗唯一的入口，钉在最左上（DSH 同款位置）。
          用 panel-left 而不是 panel-left-close：它是"切换"，收起与展开是同一个动作的两种结果，
          按钮本身不该换图标 —— 换了就等于说"这是两件事"。 */}
      <button
        type="button"
        style={NO_DRAG}
        onClick={onToggleSidebar}
        title={collapsed ? '展开左侧目录' : '收起左侧目录'}
        aria-label={collapsed ? '展开左侧目录' : '收起左侧目录'}
        aria-pressed={!collapsed}
        className="flex h-7 w-7 shrink-0 items-center justify-center rounded-ui text-muted-foreground transition-colors hover:bg-accent/70 hover:text-foreground"
      >
        <PanelLeft className="h-4 w-4" />
      </button>
      <Menu label="文件" entries={fileMenu} />
      <Menu label="视图" entries={viewMenu} />
      <Menu label="帮助" entries={helpMenu} />
    </div>
  )
}
