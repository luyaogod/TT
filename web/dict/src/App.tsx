import { useState, type ReactNode } from 'react'
import { Server, Folder, Database, Settings, Bug } from 'lucide-react'
import { SettingsView } from './SettingsView'
import { MirrorView } from './MirrorView'
import { SyncView } from './SyncView'
import { AppSettingsView } from './AppSettingsView'
import { cn } from './ui'

// 单页应用:左侧活动栏切换「环境配置」(SSH/数据库)、「源码镜像」、「数据同步」、「设置」(命令行安装)。
// 栏底还有一个通往调试工作台的入口 —— 那是**另一套 SPA**(统一服务下挂在 /debug/),
// 所以只能整页跳转,不在本页的 view 状态里。
function ActivityIcon({ label, active, onClick, children }: {
  label: string; active: boolean; onClick: () => void; children: ReactNode
}) {
  return (
    <button type="button" title={label} onClick={onClick}
      className={cn('flex h-8 w-full shrink-0 items-center justify-center transition-colors',
        active ? 'bg-zinc-100 text-zinc-900 dark:bg-zinc-800 dark:text-zinc-50' : 'text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-100')}>
      {children}
    </button>
  )
}

// 跨页入口(普通 <a>,整页跳转):样式与 ActivityIcon 的非选中态逐字一致,
// 免得看起来像另一个控件
function ActivityLink({ label, href, children }: { label: string; href: string; children: ReactNode }) {
  return (
    <a href={href} title={label}
      className="flex h-8 w-full shrink-0 items-center justify-center transition-colors text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-100">
      {children}
    </a>
  )
}

// 视图常驻挂载,切换只改 display:输入框/滚动位置/轮询状态不丢。
function ViewPane({ show, children }: { show: boolean; children: ReactNode }) {
  return <div className={cn('h-full min-h-0 min-w-0', show ? 'block' : 'hidden')}>{children}</div>
}

export function App() {
  const [view, setView] = useState<'envs' | 'mirror' | 'sync' | 'settings'>('envs')
  return (
    <div className="flex h-full min-h-0 bg-zinc-50 text-zinc-800 dark:bg-zinc-950 dark:text-zinc-100">
      <nav className="flex w-10 shrink-0 flex-col items-center gap-1 border-r border-zinc-200 bg-white py-2 dark:border-zinc-800 dark:bg-zinc-900">
        <ActivityIcon label="环境配置(SSH 与数据库)" active={view === 'envs'} onClick={() => setView('envs')}>
          <Server className="h-5 w-5" />
        </ActivityIcon>
        <ActivityIcon label="源码镜像" active={view === 'mirror'} onClick={() => setView('mirror')}>
          <Folder className="h-5 w-5" />
        </ActivityIcon>
        <ActivityIcon label="数据同步(ERP → 本地 SQLite)" active={view === 'sync'} onClick={() => setView('sync')}>
          <Database className="h-5 w-5" />
        </ActivityIcon>
        <ActivityIcon label="设置(命令行安装)" active={view === 'settings'} onClick={() => setView('settings')}>
          <Settings className="h-5 w-5" />
        </ActivityIcon>
        {/* 跨页入口:调试工作台是另一套 SPA,整页跳过去(统一服务下挂在 /debug/) */}
        <ActivityLink label="调试工作台(切换到调试页)" href="/debug/">
          <Bug className="h-5 w-5" />
        </ActivityLink>
      </nav>
      <div className="min-h-0 min-w-0 flex-1">
        <ViewPane show={view === 'envs'}><SettingsView /></ViewPane>
        <ViewPane show={view === 'mirror'}><MirrorView /></ViewPane>
        <ViewPane show={view === 'sync'}><SyncView /></ViewPane>
        <ViewPane show={view === 'settings'}><AppSettingsView /></ViewPane>
      </div>
    </div>
  )
}
