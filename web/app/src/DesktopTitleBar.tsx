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
import { useState, type CSSProperties } from 'react'
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
      className="flex w-full items-center gap-6 rounded-[5px] px-2.5 py-1.5 text-left text-[12px] text-foreground transition-colors hover:bg-accent"
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
            'rounded-[5px] px-2 py-1 text-[12px] transition-colors',
            open ? 'bg-accent text-foreground' : 'text-muted-foreground hover:bg-accent/70 hover:text-foreground',
          )}
        >
          {label}
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" sideOffset={6} className="min-w-[196px] rounded-[8px] border-border bg-popover p-1 shadow-lg">
        {entries.map((e, i) =>
          e.kind === 'sep'
            ? <div key={`sep-${i}`} className="my-1 h-px bg-border" />
            : <MenuItem key={e.label} entry={e} close={() => setOpen(false)} />,
        )}
      </PopoverContent>
    </Popover>
  )
}

export function DesktopTitleBar() {
  const shell = desktopShell()
  if (!shell) return null

  const fileMenu: Entry[] = [
    { kind: 'item', label: '在浏览器中打开', run: () => shell.openInBrowser() },
    { kind: 'item', label: '打开数据目录', run: () => shell.openDataDir() },
    { kind: 'sep' },
    { kind: 'item', label: '退出', run: () => shell.quit() },
  ]
  const viewMenu: Entry[] = [
    { kind: 'item', label: '重新加载', hint: 'F5', run: () => shell.reload() },
    { kind: 'item', label: '实际大小', hint: 'Ctrl+0', run: () => shell.resetZoom() },
    { kind: 'item', label: '全屏', hint: 'F11', run: () => shell.toggleFullscreen() },
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
      <Menu label="文件" entries={fileMenu} />
      <Menu label="视图" entries={viewMenu} />
      <Menu label="帮助" entries={helpMenu} />
    </div>
  )
}
