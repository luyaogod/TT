import { useEffect, useRef, useState, type ReactNode } from 'react'
import { RotateCcw, WifiOff, Bug, BugPlay, Globe, FlaskConical, Video, Settings, ListTree, Server, type LucideIcon } from 'lucide-react'
import { connectWS, useStore } from './store'
import { Toolbar } from './Toolbar'
import { SourceView, editorRef } from './SourceView'
import { RightPanels, TimelinePanel } from './Panels'
import { OutlinePanel } from './OutlinePanel'
import { SessionPanel } from './SessionPanel'
import { WsLogView } from './WsLogView'
import { WsTestView } from './WsTestView'
import { RecordView } from './RecordView'
import { SettingsView } from './SettingsView'
import { StatusBar, STATUS_BAR_H } from './StatusBar'
import { api } from './api'
import { parseHash, viewHash, type ViewKey } from './routing'
import { ErrorBoundary } from './ErrorBoundary'

// 视图键 → 中文名(只给 tooltip 用:界面上的说法要与右侧图标栏、README 一致)。
const VIEW_LABEL: Partial<Record<ViewKey, string>> = {
  debug: '调试', wslogs: '接口日志', wstest: '服务测试', record: '录制', settings: '设置',
}

// 活动栏图标按钮(左右两条栏共用):悬停/选中都是浮在栏底上的小方块,不贴边 ——
// 直角色块贴到栏壁会看起来像"栏被切掉一块",而留白 + 圆角让选中成为一个明确的"按钮"。
//
// 底色走 accent 语义 token,不再是半透明的 bg-foreground/10:
// 半透明叠加的灰会随栏底色变化(亮暗两套因此是两个不同的灰),而 accent 是调色板里
// 已经定好的"选中/悬停"色,与下拉项、设置页导航的选中态同色。
// 悬停给半档(bg-accent/60),选中给整档,层级关系保住。
//
// 圆角是对全站零圆角(tokens.css 把 radius 全部清零)的一处刻意例外:参照的是
// "Ghost 按钮选中"的观感,直角做不出那个形状。
function ActivityIcon({ icon: Icon, label, active, onClick }: {
  icon: LucideIcon; label: string; active: boolean; onClick: () => void
}) {
  return (
    <button title={label} onClick={onClick}
      className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-ui transition-colors ${
        active ? 'bg-accent text-foreground' : 'text-muted-foreground hover:bg-accent/60 hover:text-foreground'
      }`}>
      <Icon className="h-5 w-5" />
    </button>
  )
}

// 单页面应用 keep-alive:视图首次访问后常驻挂载,切换只改 display,不卸载重挂。
// 编辑器(Monaco)实例、源码页签、表单输入、滚动位置等状态全部保留,切换零闪烁、
// 零重载——这正是"切换页面不像网页那样重新加载"的实现。
//
// 每个视图各包一层错误边界(见 ErrorBoundary.tsx):某个视图崩了,会话栏、工具条、
// 其余视图都还在,切走就能继续干活 —— 而不是整页白屏(实测过,2026-10-15)。
function ViewPane({ show, label, children }: { show: boolean; label: string; children: ReactNode }) {
  const [mounted, setMounted] = useState(show)
  useEffect(() => { if (show) setMounted(true) }, [show])
  if (!mounted) return null
  // min-w-[320px]:主内容有个下限,窄窗口下宁可让工作区出横向滚动条,也不把自己压成 0 宽
  // (实测 360 宽的窗口里编辑器列宽恰好 0:面板还在、内容没了,而且没有滚动条能到达它)。
  return (
    <div className={`min-h-0 min-w-[320px] flex-1 ${show ? 'flex' : 'hidden'}`}>
      <ErrorBoundary what={`视图「${label}」`}>{children}</ErrorBoundary>
    </div>
  )
}

// 浏览器标签页标题反馈:停站时标题闪烁(仅后台),运行/启动改前缀。
// 停站标题在「⚠ 停站 file:line」与原名之间每秒交替,后台标签会被浏览器高亮提醒。
function useDocTitleBlink() {
  const sessionId = useStore((s) => s.sessionId)
  const state = useStore((s) => s.state)
  const file = useStore((s) => s.stop?.file)
  const line = useStore((s) => s.stop?.line)
  useEffect(() => {
    const base = 'TT'
    const setT = (t: string) => { document.title = t }
    if (!sessionId) { setT(base); return }
    let timer: number | undefined
    if (state === 'stopped') {
      const loc = file && line ? `${file}:${line}` : '已停站'
      const flashA = `⚠ 停站 ${loc}`
      setT(flashA)
      let toggle = true
      const onVis = () => {
        if (document.hidden) {
          toggle = true
          timer = window.setInterval(() => {
            document.title = toggle ? base : flashA
            toggle = !toggle
          }, 1000)
        } else {
          if (timer) { clearInterval(timer); timer = undefined }
          setT(flashA)
        }
      }
      document.addEventListener('visibilitychange', onVis)
      if (document.hidden) onVis()
      return () => {
        document.removeEventListener('visibilitychange', onVis)
        if (timer) clearInterval(timer)
        setT(base)
      }
    }
    if (state === 'running') setT(`▶ 运行中 · ${base}`)
    else if (state === 'loading') setT(`⏳ 启动中 · ${base}`)
    else setT(base)
  }, [sessionId, state, file, line])
}

export function App() {
  useDocTitleBlink()
  const jumpedRef = useRef(false)
  useEffect(() => {
    if (jumpedRef.current) return
    jumpedRef.current = true
    // 先认 URL 片段:深链接必须优先于下面那条兜底 —— 形如 #settings/data-dict 的链接
    // 要落到指定分区,若先跑"没有环境就跳设置"会被覆盖成站点管理。
    const r = parseHash(location.hash)
    if (r.view) {
      useStore.getState().setView(r.view)
      if (r.view === 'settings' && r.section) useStore.getState().setSettingsSection(r.section)
      return
    }
    // 兜底:还没有任何服务器环境时(首次运行,或本地还没配过)落到「设置 → 站点管理」。
    // 正常情况下 store 的初始 view 已经是 settings(见 store.ts),所以这条通常**不产生动作**;
    // 它留着的意义是:哪怕哪天默认视图被改回调试页,空环境这个"什么都干不了"的状态
    // 也不会把人丢在一个没用的页面上。只在本页首次加载时判一次。
    void api.settings().then((c) => {
      const sshs = (c as { sshs?: unknown[] })?.sshs
      if (!Array.isArray(sshs) || sshs.length === 0) useStore.getState().setView('settings')
    }).catch(() => { /* 服务未就绪时不动:等用户自己操作 */ })
  }, [])
  // 浏览器前进/后退,以及从别处再点一次同样的深链接:跟着片段走
  useEffect(() => {
    const onHash = () => {
      const r = parseHash(location.hash)
      if (!r.view) return
      const st = useStore.getState()
      st.setView(r.view)
      if (r.view === 'settings' && r.section) st.setSettingsSection(r.section)
    }
    window.addEventListener('hashchange', onHash)
    return () => window.removeEventListener('hashchange', onHash)
  }, [])
  useEffect(() => {
    const close = connectWS()
    void api.status().catch(() => {})
    void useStore.getState().adoptExisting()
    // 调试快捷键:F5 继续 / F10 步过 / F11 步入 / Shift+F11 步出
    const onKey = (e: KeyboardEvent) => {
      const st = useStore.getState()
      if (!st.sessionId) return
      const stopped = st.state === 'stopped'
      if (e.key === 'F5') { e.preventDefault(); if (stopped) void st.control('continue') }
      else if (e.key === 'F10') { e.preventDefault(); if (stopped) void st.control('next') }
      else if (e.key === 'F11') {
        e.preventDefault()
        if (stopped) void st.control(e.shiftKey ? 'finish' : 'step')
      }
    }
    window.addEventListener('keydown', onKey)
    return () => { window.removeEventListener('keydown', onKey); close() }
  }, [])

  const showRight = useStore((s) => s.showRight)
  const showBottom = useStore((s) => s.showBottom)
  const bottomMax = useStore((s) => s.bottomMax)
  const rightView = useStore((s) => s.rightView)
  const setRightView = useStore((s) => s.setRightView)
  const view = useStore((s) => s.view)
  const setView = useStore((s) => s.setView)
  const settingsSection = useStore((s) => s.settingsSection)
  // 视图/分区变化 → **把片段写回去**。这是"刷新后还在这一页"的唯一保证（见 routing.ts 的 viewHash）：
  // 只读片段、不写回去的话，点图标切到别的视图时 URL 还停在旧视图上，一刷新就把人弹回旧页
  // （实测：在 `#debug` 里点「设置」，界面是设置页而 URL 仍是 `#debug`）。
  //
  // 首跑让一次：URL 已经指着某个视图时，那是深链接，而上面那条 mount effect 正在把它解析成
  // store 状态 —— 这一刻的 view/section 还是**初始渲染快照**，照写会把深链接覆盖成默认页
  // （SettingsView 里那条分区同步 effect 让的是同一件事）。
  // 往后每换一次才写，且用 replace：push 会让浏览器后退要一格一格走遍所有视图。
  const hashWroteRef = useRef(false)
  useEffect(() => {
    if (!hashWroteRef.current) {
      hashWroteRef.current = true
      // 首跑：片段已经指着某个视图时，**URL 是权威**（那是深链接），而这一刻的 store 还是
      // 初始渲染快照（mount 那条 effect 正把它解析出来）—— 所以目标按**片段自己**算：
      // 合法片段一个字节都不改；写法不规范（`#settings/site` 单数、`#settings` 没带分区）
      // 就顺手改成规范的，免得地址栏长期留着一个界面并不使用的值。
      const r = parseHash(window.location.hash)
      if (r.view) {
        const want0 = viewHash(r.view, r.section || useStore.getState().settingsSection)
        if (window.location.hash !== want0) window.history.replaceState(null, '', want0)
        return
      }
    }
    const want = viewHash(view, settingsSection)
    if (window.location.hash !== want) window.history.replaceState(null, '', want)
  }, [view, settingsSection])
  // launchError 不再在这里渲染(改走统一 toast);store 那个字段仍由启动失败写入,
  // 供别处/调试读取,应用外壳不再订阅它 —— 订阅了也只会白白多一次重渲染。
  const backendDead = useStore((s) => s.backendDead)
  const wsConnected = useStore((s) => s.wsConnected)
  const launching = useStore((s) => s.launching)
  const sessionId = useStore((s) => s.sessionId)
  // 注意:sessionId 必须是独立 hook 调用——嵌在 || 条件里会在短路时少调用一次,
  // 违反 React Hooks 规则导致 hook 队列错乱(Should have a queue / 无限重渲染)
  const banner = backendDead || (sessionId && !wsConnected ? '服务连接断开,自动重连中…' : '') 
  const restart = useStore((s) => s.restart)

  // 侧边栏/代码区宽度拖拽(记忆到 localStorage)。
  // 宽度放 **store**：浮动调试工具条要按它算默认落点，而 localStorage 只在 mouseup 时才写 ——
  // 组件本地 state 的话工具条读不到拖拽中的值，会一直停在旧位置压住右栏（实测重叠 184px）。
  const panelW = useStore((s) => s.panelW)
  const setPanelW = useStore((s) => s.setPanelW)
  // 左侧会话栏的宽度:与右侧栏同一套口径(同一档 clamp),只是记忆键与缺省值不同 ——
  // 会话面板要列出环境名与状态,300 比 320 让代码区多留一点。
  const [sessionW, setSessionW] = useState(() => {
    const v = Number(localStorage.getItem('tt.sessionW'))
    return v >= 240 && v <= 640 ? v : 300
  })
  // 会话栏的展开/收起:只由左侧那个图标按钮驱动,直接读 localStorage 而不是进 store ——
  // 因为它没有第二个消费者,进 store 只会多一层。默认展开(只有明确关过才是 '0')。
  const [showSession, setShowSession] = useState(() => localStorage.getItem('tt.sessionOpen') !== '0')
  // 最后一个"带会话栏"的视图(设置页不算)。设置页里点会话图标要能回到这里 ——
  // 否则在设置页点它毫无动静(设置页压根不渲染会话栏),用户看到的就是一个死按钮(用户实测反馈)。
  const lastViewRef = useRef<Exclude<ViewKey, 'settings'>>('debug')
  useEffect(() => { if (view !== 'settings') lastViewRef.current = view as Exclude<ViewKey, 'settings'> }, [view])
  // 下方面板(操作时间线/协议流)高度拖拽(记忆到 localStorage)
  const [bottomH, setBottomH] = useState(() => {
    const v = Number(localStorage.getItem('tt.bottomH'))
    return v >= 120 && v <= 600 ? v : 192
  })
  // 视口变窄时按运行时上限收窄两侧面板 —— **不写回持久化**：窗口拉大就恢复原宽度。
  //
  // 为什么需要：两侧面板都是 shrink-0 + 固定像素宽，而工作区是 overflow-hidden。
  // 窗口一窄（实测 360×640，或 800 宽配两个 640 的面板），主内容会被压成 0～61px、
  // 右侧栏被裁在视口外，且没有任何滚动条能到达它们 —— 静默裁剪，用户连"编辑器去哪了"
  // 都看不出来。这里按"视口 − 图标栏 − 间隙 − 主内容下限(320)"给两侧一个总预算，
  // 超了按比例收（各自不低于 240）；窗口再窄到连下限都放不下时，由工作区的横向滚动兜底。
  const [vw, setVw] = useState(() => window.innerWidth)
  useEffect(() => {
    const onResize = () => setVw(window.innerWidth)
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])
  const sideRoom = Math.max(480, vw - 40 - 8 - 320)
  const sideWanted = (showSession ? sessionW : 0) + (showRight ? panelW : 0)
  const sideK = sideWanted > sideRoom ? sideRoom / sideWanted : 1
  const sessionWEff = showSession ? Math.max(240, Math.round(sessionW * sideK)) : sessionW
  const panelWEff = showRight ? Math.max(240, Math.round(panelW * sideK)) : panelW

  // keep-alive 视图切换下,Monaco 容器被 display:none 期间 ResizeObserver 会拿到 0 尺寸;
  // 切回调试视图后补一次 layout(),确保编辑器渲染与视口不残留错位
  useEffect(() => {
    if (view !== 'debug') return
    const t = setTimeout(() => editorRef.current?.layout(), 60)
    return () => clearTimeout(t)
  }, [view])

  const onResizeDown = (e: React.MouseEvent<HTMLDivElement>) => {
    e.preventDefault()
    const el = e.currentTarget
    el.classList.add('dragging')
    const startX = e.clientX
    const startW = panelW
    let latest = startW
    const move = (ev: MouseEvent) => {
      latest = Math.min(640, Math.max(240, startW - (ev.clientX - startX)))
      setPanelW(latest)
    }
    const up = () => {
      el.classList.remove('dragging')
      localStorage.setItem('tt.panelW', String(latest))
      window.removeEventListener('mousemove', move)
      window.removeEventListener('mouseup', up)
    }
    window.addEventListener('mousemove', move)
    window.addEventListener('mouseup', up)
  }

  // 左侧会话栏的宽度拖拽:与上面那条同源(mousemove 跟手 + mouseup 落盘),
  // 唯一的差别是方向 —— 会话栏在分隔线**左侧**,所以往右拖 = 变宽(取 clientX - startX)。
  const onSessionResizeDown = (e: React.MouseEvent<HTMLDivElement>) => {
    e.preventDefault()
    const el = e.currentTarget
    el.classList.add('dragging')
    const startX = e.clientX
    const startW = sessionW
    let latest = startW
    const move = (ev: MouseEvent) => {
      latest = Math.min(640, Math.max(240, startW + (ev.clientX - startX)))
      setSessionW(latest)
    }
    const up = () => {
      el.classList.remove('dragging')
      localStorage.setItem('tt.sessionW', String(latest))
      window.removeEventListener('mousemove', move)
      window.removeEventListener('mouseup', up)
    }
    window.addEventListener('mousemove', move)
    window.addEventListener('mouseup', up)
  }

  // 会话栏的显示开关:active 态即"面板此刻可见"。写进 localStorage 让折叠意图跨刷新保持
  // (与宽度、面板显隐同一档持久化口径)。
  const toggleSession = () => {
    setShowSession((v) => {
      localStorage.setItem('tt.sessionOpen', v ? '0' : '1')
      return !v
    })
  }
  // 会话图标真正干的事:在四个容器里就是收/展会话栏;在**设置页**则先切回上一个容器
  // 再把它展开 —— 设置页不渲染会话栏,不这么办这个按钮在设置页按下去毫无反应。
  const onSessionClick = () => {
    if (view === 'settings') {
      setView(lastViewRef.current)
      setShowSession(true)
      localStorage.setItem('tt.sessionOpen', '1')
      return
    }
    toggleSession()
  }
  const sessionIconLabel = view === 'settings'
    ? `会话(离开设置,回到「${VIEW_LABEL[lastViewRef.current] || lastViewRef.current}」并展开会话栏)`
    : showSession ? '收起会话面板' : '展开会话面板'
  // 下方面板(操作时间线/协议流)高度拖拽(记忆到 localStorage)
  const onBottomResizeDown = (e: React.MouseEvent<HTMLDivElement>) => {
    e.preventDefault()
    const el = e.currentTarget
    el.classList.add('dragging')
    const startY = e.clientY
    const startH = bottomH
    let latest = startH
    const move = (ev: MouseEvent) => {
      // 分隔线跟手:向上拖 = 面板变高(与 VS Code 底部面板一致)
      latest = Math.min(600, Math.max(120, startH - (ev.clientY - startY)))
      setBottomH(latest)
    }
    const up = () => {
      el.classList.remove('dragging')
      localStorage.setItem('tt.bottomH', String(latest))
      window.removeEventListener('mousemove', move)
      window.removeEventListener('mouseup', up)
    }
    window.addEventListener('mousemove', move)
    window.addEventListener('mouseup', up)
  }

  // 布局:上一条与状态栏同高的留白 / 下一条状态栏;中部从左往右 = 活动栏 → 会话栏 →
  // 主内容面板 → 视图图标栏,各自成块、彼此留间隙,面板带边线与圆角。
  //
  // 会话栏为什么**单独一列**(而不是继续挂在调试视图的右侧栏里):
  //   · 会话是"干活前的那个动作"(选环境 / 重启 / 结束),四个视图都要用到它 ——
  //     挂在调试页的 sheet 里,别的视图就够不着;
  //   · 它必须是**全站同一份**:一旦跟着右侧栏切换卸载重挂,手风琴展开态与 TOPENT
  //     输入会当场丢,所以它挂在 app 外壳这一层、不挂进任何一个 ViewPane;
  //   · 放左边还有个好处:它与右侧栏的宽度拖拽互不抢位置,左右各自独立收展。
  // 而**设置页不显示它**:设置页自己是"先配环境才能干活"的地方,一个列环境与会话的
  // 面板在那里只是重复(环境就在设置页里管),徒占一列宽度;右侧那四个图标栏同理。
  // 左侧活动栏本身(设置 / 会话)是常驻的 —— 它是这一屏唯一的收展开关,不能被自己藏掉。
  //
  // (顶层那条横栏已撤掉:它只剩两个面板收展按钮,而那两个只管这一屏,
  //  现在钉在调试页签栏右端;浮动调试工具条归调试视图,所以跟着挪进 ViewPane)
  return (
    <div className="flex h-screen flex-col bg-background text-foreground">
      {/* 顶部留白,高度取状态栏那一份:内容不贴着窗口上沿,上下留白也就对称了 */}
      <div className={`${STATUS_BAR_H} shrink-0`} />
      {/* 「启动失败」原先是这里一条整行红色横幅(还把整个工作台往下压)。
          已改走统一 toast,并且**不自动关**:启动失败必须看见(见 store 的 launch catch)。
          下面那条「后端断开」保留在这里 —— 它不是一次性消息,而是**持续状态** + 一个动作
          (重新启动按钮),那类东西不该做成会自己消失的 toast。 */}
      {banner && (
        <div className="flex shrink-0 items-center gap-2 border-b border-red-500/20 bg-red-500/10 px-3 py-1.5 text-xs text-red-600 dark:text-red-600 dark:text-red-400">
          <WifiOff className="h-3.5 w-3.5" />
          <span>{banner}</span>
          {backendDead && (
            <button
              className="ml-auto inline-flex items-center gap-1 border border-red-500/30 px-2 py-0.5 hover:bg-red-500/10 disabled:opacity-40"
              disabled={launching}
              onClick={() => void restart()}
            >
              <RotateCcw className="h-3 w-3" />
              {launching ? '重启中…' : '重新启动'}
            </button>
          )}
        </div>
      )}
      {/* 四周留白对称:行内边距与栏-面板间隙都取 4px,于是栏里的小按钮**居中**在
          「窗口边」与「面板边」之间 —— 钮宽 32、栏宽 40,栏内本就留了 4px,
          两边各再加 4px,两侧都是 8px。之前间隙给 8px,内侧就成了 12px、外侧 8px,
          看上去按钮贴着窗口而离面板远。底部 8px 与状态栏隔开。 */}
      <div className="flex min-h-0 flex-1 gap-1 px-1 pb-2">
        {/* VS Code 经典活动栏:栏与内容各自成块,靠间隙分开,不再用分割线。
            只有两个图标:入口(设置)与干活前的那一个动作(会话)。四个视图容器
            已经挪到右边的图标栏(顺序与"先配后干"无关,所以不再需要靠位置表达主次);
            顺序:**设置排第一** —— 首次进来落在设置页(见 store 的初始 view 与
            下面那条"没有环境就跳设置"),入口与落点排在同一侧才不用回头找。 */}
        <div className="flex w-10 shrink-0 flex-col items-center gap-1 py-2">
          <ActivityIcon icon={Settings} label="设置(站点 / 数据字典 / DEBUG / 应用)" active={view === 'settings'} onClick={() => setView('settings')} />
          {/* active 即"会话栏此刻可见":它是个收展开关,不是视图切换 —— 所以不跟 view 走。
              在设置页它不是"没反应":点一下会离开设置、回到上一个容器并把会话栏展开
              (设置页不渲染会话栏,不这么办这个按钮在设置页就是个死键)。 */}
          <ActivityIcon icon={Server} label={sessionIconLabel} active={view !== 'settings' && showSession} onClick={onSessionClick} />
        </div>
        {/* 工作区 = 会话栏 + 主内容,**同一个面板**(共用一圈边线、中间只隔那条可拖的分割线)。
            为什么不各自成块:两块各自带边线时,中间会同时出现"4px 间隙 + 两条平行边线",
            看着就是一道莫名其妙的空隙(用户反馈)。这里与调试视图内部(代码区 | 侧边栏)同一套做法。
            会话栏不渲染时(设置页、或用户收起),这层壳里只剩主内容,边线与圆角照旧。
            会话栏自身:min-w-0 + overflow-hidden —— 里面是 auto 滚动区,不给这两条的话
            长环境名会把整列顶宽、再挤主内容(就是"宽度反馈环"那条老账)。
            它不挂在任何一个 ViewPane 里,所以四个容器之间切换时一次都不重挂:
            手风琴展开态与 TOPENT 输入跟着一直在(那正是"为什么要单独一列")。 */}
        <div className="flex min-h-0 min-w-0 flex-1 overflow-x-auto overflow-y-hidden rounded-ui border border-border">
          {/* 会话栏:收起时 / 在设置页上**只隐藏、不卸载**。
              原来写的是 `view !== 'settings' && showSession && <SessionPanel/>` —— 一进设置页
              整块被卸载，于是 TOPENT 输入与手风琴折叠态当场丢（实测：填好 XYZ9 + 折叠「环境变量」
              → 点设置 → 从设置页回来，两样都没了），而这与上面那句"挂一次就一直在"的承诺相反。
             现在只有 display 在变，组件与它的本地状态一直在。 */}
          <div style={{ width: sessionWEff }}
            className={view !== 'settings' && showSession
              ? 'flex min-h-0 min-w-0 shrink-0 flex-col overflow-hidden'
              : 'hidden'}>
            <SessionPanel />
          </div>
          {/* 分割条放在会话栏**右边**:它是这一列真正的边(和右侧栏那条一样,分割线始终
              贴着被拖的那条边)。放在左边的话,往右拖时面板变宽、而把手留在原地,
              光标与那条 1px 线当场分家。 */}
          {view !== 'settings' && showSession && (
            <div className="col-resizer" onMouseDown={onSessionResizeDown} title="拖拽调整会话面板宽度" />
          )}
          {/* 五个视图全部 keep-alive:首次访问后常驻挂载,切换仅改 display。
              它们与上面的会话栏是同一层的兄弟,只是同时只有一个不是 display:none。 */}
          <ViewPane show={view === 'debug'} label="调试">
            <Toolbar />
            {/* 面板紧贴(VS Code 经典密度):无外边距,仅靠 sash 分割线分区
               min-w-0 + overflow-hidden:Monaco 会给编辑器写内联像素宽度,
               否则 flex 最小宽度被钉死,收起再展开时编辑区不回缩、右面板被挤出屏幕 */}
            <div className="flex min-h-0 min-w-0 flex-1 overflow-hidden">
              <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
                {/* 编辑器:下方面板最大化时整块让位(那个"最大化"就是"我要专看这段协议流") */}
                <div className={`min-h-0 flex-1 overflow-hidden bg-card ${bottomMax ? 'hidden' : ''}`}>
                  <SourceView />
                </div>
                {showBottom && (
                  /* 最大化时这一列改吃整列高度(flex-1),**分割条不渲染** —— 那时高度由
                     "最大化"接管,拖它什么也不会发生,留着只会让人以为还能拖。 */
                  <div className={bottomMax ? 'flex min-h-0 flex-1 flex-col' : 'flex shrink-0 flex-col'}>
                    {!bottomMax && (
                      <div className="row-resizer" onMouseDown={onBottomResizeDown} title="拖拽调整下方面板高度" />
                    )}
                    <div style={bottomMax ? undefined : { height: bottomH }}
                      className={bottomMax ? 'min-h-0 flex-1' : 'shrink-0'}>
                      {/* 下方面板单独一层边界:时间线崩了不该把代码区一起带走 */}
                      <ErrorBoundary what="下方面板（操作时间线）"><TimelinePanel /></ErrorBoundary>
                    </div>
                  </div>
                )}
              </div>
              {showRight && (
                <>
                  <div className="col-resizer" onMouseDown={onResizeDown} title="拖拽调整代码区与侧边栏宽度" />
                  <div style={{ width: panelWEff }} className="flex min-h-0 shrink-0 flex-col overflow-hidden">
                    {/* 侧边栏内部的 sheet 切换行:横排、钉在栏顶,与代码区仍在同一个面板里
                        (它切的是这一列的内容,不是整屏的视图 —— 所以不做成第二条活动栏)。
                        只有两个 sheet:会话已经搬到左边独立一列(见上面的布局注释),
                        留在这里等于同一份会话有两处入口,状态还得互相对齐。 */}
                    <div className="flex h-8 shrink-0 items-center gap-1 border-b border-border px-1">
                      <ActivityIcon icon={BugPlay} label="调试面板" active={rightView === 'debug'} onClick={() => setRightView('debug')} />
                      <ActivityIcon icon={ListTree} label="大纲" active={rightView === 'outline'} onClick={() => setRightView('outline')} />
                    </div>
                    {/* 多 sheet keep-alive:切换仅改 display,保留手风琴展开/监视输入等本地状态。
                        两个 sheet 必须包在一层 flex 项里:panel 自己的类里有 flex-1
                        (占满剩余宽高),直接做列容器的子项会被上面那行图标挤着算高度 ——
                        包一层后这一层才是"图标行下面的剩余高度",panel 在它里面照旧占满。 */}
                    <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
                      {/* 右侧栏单独一层边界:变量树/大纲崩了,代码区与工具条都还在
                          （实测过一次:一个畸形变量值把整页打成白屏）。 */}
                      <div className={rightView === 'debug' ? 'h-full min-h-0 w-full' : 'hidden'}>
                        <ErrorBoundary what="调试面板（变量/断点/调用栈）"><RightPanels /></ErrorBoundary>
                      </div>
                      <div className={rightView === 'outline' ? 'h-full min-h-0 w-full' : 'hidden'}>
                        <ErrorBoundary what="大纲面板"><OutlinePanel /></ErrorBoundary>
                      </div>
                    </div>
                  </div>
                </>
              )}
            </div>
          </ViewPane>
          <ViewPane show={view === 'wslogs'} label="接口日志">
            <WsLogView />
          </ViewPane>
          <ViewPane show={view === 'wstest'} label="服务测试">
            <WsTestView />
          </ViewPane>
          <ViewPane show={view === 'record'} label="录制">
            <RecordView />
          </ViewPane>
          <ViewPane show={view === 'settings'} label="设置">
            <SettingsView />
          </ViewPane>
        </div>
        {/* 右侧视图图标栏(仿左侧活动栏,竖排):调试 / 接口日志 / 服务测试 / 录制 ——
            就是"主区域里放哪一个容器",与左侧的入口(设置)和前置动作(会话)分开两边。
            顺序即主次:调试是主场,其余三个按"有环境之后"的常用度排下去。
            它就是"这一屏放哪个视图",不是视图内容,所以留在主面板外面做它的右邻;
            设置页整条不渲染 —— 设置不是这四个容器之一,它自己占满主区域。 */}
        {view !== 'settings' && (
          <div className="flex w-10 shrink-0 flex-col items-center gap-1 py-2">
            <ActivityIcon icon={Bug} label="调试" active={view === 'debug'} onClick={() => setView('debug')} />
            <ActivityIcon icon={Globe} label="接口日志" active={view === 'wslogs'} onClick={() => setView('wslogs')} />
            <ActivityIcon icon={FlaskConical} label="服务测试" active={view === 'wstest'} onClick={() => setView('wstest')} />
            <ActivityIcon icon={Video} label="录制" active={view === 'record'} onClick={() => setView('record')} />
          </div>
        )}
      </div>
      <StatusBar />
    </div>
  )
}
