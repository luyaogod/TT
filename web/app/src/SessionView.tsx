// 会话页:一条会话对应的工作面板。
//
// 两个 Tab 恒在(Debug / FEP录制),**顺序与默认落点由会话类型决定**:Debug 会话先看 Debug,
// FEP 会话先看 FEP录制。不要把"另一条 Tab"藏掉 —— 会话类型只决定"先看哪一个",
// 两条链路都要能从这里走到。
//
// Debug Tab 是原来整块调试工作台(见 DebugWorkbench.tsx);FEP录制 这一屏现在如实显示
// **这条会话挂着的画面通道**(建会话时默认就挂 —— 一个脱离进程的 `tt recording start`,
// 录制代理在中间)。没挂上就说没挂上,不留一块什么都不说的白板。
import { useEffect, useState } from 'react'
import { FolderOpen, Video } from 'lucide-react'
import type { ReactNode } from 'react'
import { api } from './api'
import type { SessionItem } from './store'
import type { SessionTab } from './routing'
import { Pane, TabStrip, ViewHeader, Meta, type TabItem } from './Shell'
import { DebugWorkbench } from './DebugWorkbench'
import { SESSION_TYPE_LABEL } from './Sidebar'

/** 画面通道(录制代理)的只读视图，与后端 `GET /api/sessions/{id}/screen` 的 `screen` 块同形。 */
interface ScreenInfo {
  tag?: string
  kind?: string
  prog?: string
  proxyPort?: number
  injectPort?: number
  recordDir?: string
  stale?: boolean
  started?: string
}

/**
 * 读这条会话挂着的画面通道。
 *
 * 它只回答"现在挂着没有"：`attached:false` 是正常状态(会话可以先没有画面)，不是错误，
 * 所以取不到就显示成"没有"，并把真实的报错留在 `err` 里 —— 两者不能混成一句话。
 */
function useScreenInfo(id: string) {
  const [state, setState] = useState<{ attached: boolean; screen: ScreenInfo | null; err: string }>(
    { attached: false, screen: null, err: '' },
  )
  const [nonce, setNonce] = useState(0)
  useEffect(() => {
    let cancelled = false
    if (!id) return
    setState((s) => ({ ...s, err: '' }))
    void api.screen(id).then((r) => {
      if (cancelled) return
      setState({ attached: !!r.attached, screen: (r.screen as ScreenInfo) || null, err: '' })
    }).catch((e: any) => {
      if (cancelled) return
      setState({ attached: false, screen: null, err: e?.message || String(e) })
    })
    return () => { cancelled = true }
  }, [id, nonce])
  return { ...state, refresh: () => setNonce((n) => n + 1) }
}

export function SessionView({ session, tab, onTab, onOpenEnv, leading }: {
  session: SessionItem
  tab: SessionTab
  onTab: (t: SessionTab) => void
  onOpenEnv: () => void
  leading?: ReactNode
}) {
  const screen = useScreenInfo(session.id)
  const debugTab: TabItem<SessionTab> = { key: 'debug', label: 'Debug', hint: '断点、单步、变量与调用栈' }
  const fepTab: TabItem<SessionTab> = { key: 'fep', label: 'FEP录制', hint: '这条会话的画面通道与录制' }
  const tabs = session.type === 'debug' ? [debugTab, fepTab] : [fepTab, debugTab]

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
      <TabStrip items={tabs} active={tab} onChange={onTab} />

      <div className="flex min-h-0 min-w-0 flex-1 flex-col">
        <Pane show={tab === 'debug'} label="Debug">
          <DebugWorkbench active={tab === 'debug'} />
        </Pane>
        <Pane show={tab === 'fep'} label="FEP录制">
          <FepPane session={session} screen={screen} />
        </Pane>
      </div>
    </div>
  )
}

/**
 * FEP录制:这条会话的画面通道。
 *
 * 写清楚四件事:这条会话是哪条、画面挂着没有、挂着的是哪一份录制、下一步能做什么。
 * 这一屏的功能(起停录制 / 回放 / 对拍)后续逐个接进来;现在能看的"挂着没有"本身
 * 就有用 —— 建会话默认挂代理，挂失败了必须能在这里看见，而不是只能翻日志。
 */
function FepPane({ session, screen }: {
  session: SessionItem
  screen: ReturnType<typeof useScreenInfo>
}) {
  const s = screen.screen
  return (
    <div className="min-h-0 flex-1 overflow-auto p-6">
      <div className="mx-auto w-full max-w-[560px]">
        <div className="flex items-center gap-2 text-[15px] font-semibold text-foreground">
          <Video className="h-4 w-4" />
          FEP 录制
        </div>
        <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
          画面通道 = 一块 GDC 画面 + 一个录制代理。建会话时默认就把它挂上了，所以这条会话
          {screen.attached ? '现在正录着' : '现在没有画面通道'}。
        </p>

        {screen.err && (
          <div className="mt-3 rounded-[8px] border border-red-500/30 bg-red-500/10 px-3 py-2 text-[11px] leading-relaxed text-red-600 dark:text-red-400">
            查画面通道失败：{screen.err}
          </div>
        )}

        <dl className="mt-4 space-y-2 rounded-[12px] border border-border bg-card p-4 text-xs">
          <Row k="环境" v={session.env} />
          <Row k="会话" v={session.ref || session.id} mono />
          <Row k="ENT" v={session.ent || '未记录'} />
          <Row k="作业" v={session.prog || '未记录'} />
          <Row k="画面" v={screen.attached ? (s?.stale ? '挂着，但进程已不在（stale）' : '已挂上') : '没有画面通道'} />
          {screen.attached && s?.tag && <Row k="录制 ID" v={s.tag} mono />}
          {screen.attached && s?.kind && <Row k="录制类型" v={s.kind} />}
          {screen.attached && (s?.proxyPort || s?.injectPort) && (
            <Row k="端口" v={`代理 ${s?.proxyPort || '?'} → ${s?.injectPort || '?'}`} mono />
          )}
          {screen.attached && s?.recordDir && <Row k="录制目录" v={s.recordDir} mono />}
        </dl>

        <div className="mt-3 flex items-center gap-2">
          <button onClick={screen.refresh}
            className="rounded-[7px] border border-border px-2 py-1 text-[12px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
            重新查一次
          </button>
          <span className="text-[11px] text-muted-foreground/80">
            这条会话的录制件也能在「环境页 → FEP录像」里按环境看到。
          </span>
        </div>
      </div>
    </div>
  )
}

function Row({ k, v, mono }: { k: string; v: string; mono?: boolean }) {
  return (
    <div className="flex items-start gap-3">
      <dt className="w-16 shrink-0 text-muted-foreground">{k}</dt>
      <dd className={mono ? 'min-w-0 break-all font-mono text-foreground' : 'min-w-0 break-all text-foreground'}>{v}</dd>
    </div>
  )
}
