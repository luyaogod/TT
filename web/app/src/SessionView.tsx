// 会话页:一条会话对应的工作面板。
//
// 两个 Tab 恒在(Debug / FEP录制),**顺序与默认落点由会话类型决定**:Debug 会话先看 Debug,
// FEP 会话先看 FEP录制。不要把"另一条 Tab"藏掉 —— 会话类型只决定"先看哪一个",
// 两条链路都要能从这里走到。
//
// Debug Tab 是原来整块调试工作台(见 DebugWorkbench.tsx);FEP录制 那一屏是**可复用控件**
// `FepRecording`(见 FepRecording.tsx):会话页给它 session 作用域(可起停),环境页给它
// env 作用域(只看当前在录什么)。同一份判定只写一处 —— 否则两块面板迟早各说一套。
import { useEffect } from 'react'
import { FolderOpen } from 'lucide-react'
import type { ReactNode } from 'react'
import { useStore, type SessionItem } from './store'
import type { SessionTab } from './routing'
import { Pane, TabStrip, ViewHeader, Meta, type TabItem } from './Shell'
import { DebugWorkbench } from './DebugWorkbench'
import { StatusBar } from './StatusBar'
import { SESSION_TYPE_LABEL } from './Sidebar'
import { FepRecording, useScreenInfo } from './FepRecording'


export function SessionView({ session, tab, onTab, onOpenEnv, onOpenRecording, leading }: {
  session: SessionItem
  tab: SessionTab
  onTab: (t: SessionTab) => void
  onOpenEnv: () => void
  /** 打开这一轮录制件的明细（环境页 → FEP录像 → 那一份）。 */
  onOpenRecording: (recId: string) => void
  leading?: ReactNode
}) {
  const screen = useScreenInfo({ kind: 'session', id: session.id, env: session.env })

  // 会话页一进来（以及状态每次变化）就问一句"要不要把母版源码打开"。
  //
  // 为什么放在这一层而不是绑定时一次性做完：会话刚建好是 idle、挂上 proxy 后变成
  // running、命中第一个断点又变 stopped —— 模块名要等**起跑那一刻**服务端才解析出来，
  // 绑定时拉不到。所以由状态变化驱动重试，`ensureSource` 自己带幂等判断（有源码就早退）。
  const state = useStore((s) => s.state)
  useEffect(() => {
    void useStore.getState().ensureSource()
  }, [session.id, state])

  const debugTab: TabItem<SessionTab> = { key: 'debug', label: 'Debug', hint: '断点、单步、变量与调用栈' }
  const fepTab: TabItem<SessionTab> = { key: 'fep', label: 'FEP录制', hint: '这条会话的画面通道与录制' }
  // 用途决定给什么（W7 的 PANE_POLICY）：
  //   - record：**不给编辑器**（已经放行跑着，防误操作），只剩 FEP 监控与收尾；
  //   - 其余：Debug + FEP 两条，顺序按用途（debug 先看 Debug，replay 先看 FEP）。
  const recordOnly = session.purpose === 'record'
  const tabs = recordOnly
    ? [fepTab]
    : (session.purpose === 'replay' ? [fepTab, debugTab] : [debugTab, fepTab])
  // 深链接可能写着 `/debug`，而 record 会话没有那一屏 —— 实际展示的那一屏以**用途**为准。
  const shownTab: SessionTab = recordOnly ? 'fep' : tab

  const meta = (
    <>
      <Meta title="会话所属环境">{session.env}</Meta>
      {session.ent && <><span className="opacity-30">·</span>
        <Meta title="企业编号（建会话时点名的那个）">ENT {session.ent}</Meta></>}
      {session.prog && <><span className="opacity-30">·</span>
        <Meta title="作业编号">作业 {session.prog}</Meta></>}
      {session.ref && <><span className="opacity-30">·</span>
        <Meta title="服务端规范 id（环境#序号）">{session.ref}</Meta></>}
      <span className="opacity-30">·</span>
      <Meta title={screen.attached ? '这条会话正挂着一块画面通道（录制代理）' : '这条会话当前没有画面通道'}>
        {screen.attached ? `录制 ${screen.screen?.tag || '已挂'}` : '无画面'}
      </Meta>
    </>
  )

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col bg-background">
      <ViewHeader title={session.name} meta={meta} leading={leading}
        actions={
          <button onClick={onOpenEnv} title={`打开环境「${session.env}」`}
            className="flex h-7 items-center gap-1 rounded-[7px] border border-border px-2 text-[12px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
            <FolderOpen className="h-3.5 w-3.5" />
            {session.env}
          </button>
        } />
      <TabStrip items={tabs} active={shownTab} onChange={onTab} />

      <div className="flex min-h-0 min-w-0 flex-1 flex-col">
        <Pane show={shownTab === 'debug'} label="Debug">
          <DebugWorkbench active={shownTab === 'debug'} />
        </Pane>
        <Pane show={shownTab === 'fep'} label="FEP录制">
          <FepRecording
            scope={{ kind: 'session', id: session.id, env: session.env, ref: session.ref, ent: session.ent, prog: session.prog }}
            mode={session.purpose === 'replay' ? 'replay' : 'both'}
            onOpenRecording={onOpenRecording}
          />
        </Pane>
      </div>
      {/* 状态栏（停站橙/运行蓝/空闲灰 + 接管按钮）属于**这条会话**，所以钉在会话页底部，
          不再横跨整个壳子 —— 左侧树和环境页没有"调试状态"可报。 */}
      <StatusBar />
    </div>
  )
}
