import { useEffect, useState, type ReactNode } from 'react'
import { Folder, Database, Settings, Bug } from 'lucide-react'
import { MirrorView } from './MirrorView'
import { SyncView } from './SyncView'
import { api } from './api'
import { cn } from '../../shared/utils'

// 字典页:左侧活动栏只有**两个操作视图**(源码镜像 / 数据同步)加两个跨页入口。
//
// 环境与数据库的配置、查询数据源、镜像目录、同步目标、BDL 文档目录 —— 这些"配置"已全部
// 收进调试工作台里的统一设置页(/debug/#settings/…)。本页保留的是**动作**:拉源码镜像、
// 跑字典同步(各自带长跑任务与进度)。合并前这里还有「环境配置」与「设置」两个页签,
// 与调试页的设置近乎重复,现在只留一份。
//
// 两个跨页入口都是普通 <a>:调试工作台是**另一套 SPA**(统一服务下挂在 /debug/),
// 不在本页的 view 状态里。
function ActivityIcon({ label, active, onClick, children }: {
  label: string; active: boolean; onClick: () => void; children: ReactNode
}) {
  return (
    <button type="button" title={label} onClick={onClick}
      className={cn('flex h-8 w-full shrink-0 items-center justify-center transition-colors',
        active ? 'bg-foreground/10 text-foreground' : 'text-muted-foreground hover:bg-foreground/5')}>
      {children}
    </button>
  )
}

// 跨页入口(普通 <a>,整页跳转):样式与 ActivityIcon 的非选中态逐字一致,
// 免得看起来像另一个控件
function ActivityLink({ label, href, children }: { label: string; href: string; children: ReactNode }) {
  return (
    <a href={href} title={label}
      className="flex h-8 w-full shrink-0 items-center justify-center text-muted-foreground transition-colors hover:bg-foreground/5">
      {children}
    </a>
  )
}

// 视图常驻挂载,切换只改 display:下拉选择/滚动位置/轮询状态不丢。
function ViewPane({ show, children }: { show: boolean; children: ReactNode }) {
  return <div className={cn('h-full min-h-0 min-w-0', show ? 'block' : 'hidden')}>{children}</div>
}

type View = 'mirror' | 'sync'

function viewFromHash(): View {
  return window.location.hash.replace(/^#/, '') === 'sync' ? 'sync' : 'mirror'
}

export function App() {
  const [view, setView] = useState<View>(viewFromHash)
  // 只有调试子系统在场时才显示「设置」入口:统一服务的路由允许只挂字典子系统,
  // 那种部署下 /debug/ 是空的,给一个死链不如不给。
  const [hasDebug, setHasDebug] = useState(true)

  useEffect(() => {
    void api.health().then((h) => setHasDebug(!!h.debug)).catch(() => { /* 探测不到就不显示 */ })
  }, [])

  // 片段 ↔ 视图:设置页里的「去运行」链接指向 /dict/#mirror 与 /dict/#sync
  useEffect(() => {
    const onHash = () => setView(viewFromHash())
    window.addEventListener('hashchange', onHash)
    return () => window.removeEventListener('hashchange', onHash)
  }, [])
  const go = (v: View) => {
    setView(v)
    if (window.location.hash !== '#' + v) window.history.replaceState(null, '', '#' + v)
  }

  return (
    <div className="flex h-full min-h-0 bg-background text-foreground">
      <nav className="flex w-10 shrink-0 flex-col items-center gap-1 border-r border-border bg-background py-2">
        <ActivityIcon label="源码镜像" active={view === 'mirror'} onClick={() => go('mirror')}>
          <Folder className="h-5 w-5" />
        </ActivityIcon>
        <ActivityIcon label="数据同步(ERP → 本地 SQLite)" active={view === 'sync'} onClick={() => go('sync')}>
          <Database className="h-5 w-5" />
        </ActivityIcon>
        {hasDebug && (
          <ActivityLink label="设置(环境 / 数据库 / 镜像目录 / 同步目标)" href="/debug/#settings/data-dict">
            <Settings className="h-5 w-5" />
          </ActivityLink>
        )}
        <ActivityLink label="调试工作台(切换到调试页)" href="/debug/">
          <Bug className="h-5 w-5" />
        </ActivityLink>
      </nav>
      <div className="min-h-0 min-w-0 flex-1">
        <ViewPane show={view === 'mirror'}><MirrorView /></ViewPane>
        <ViewPane show={view === 'sync'}><SyncView /></ViewPane>
      </div>
    </div>
  )
}
