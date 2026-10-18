// 环境页:一个环境对应的工作面板。
//
// 四个 Tab(用户 2026-10-17 定的):创建会话 / 接口日志 / 接口调试 / FEP录像。
// 其中三个是**已有的画面**,原样搬进来(接口日志 = 报文日志、接口调试 = 服务测试、
// FEP录像 = 录制面板);「创建会话」是外壳自己的那一屏 —— 它已经从"只写本机"变成
// **真的向后端建会话**(见 store.createSession 的三步)。
import { useState, type ReactNode } from 'react'
import { Bug, Plus, Server, Video } from 'lucide-react'
import { Button, Input } from '../../shared/ui'
import { cn } from '../../shared/utils'
import type { SessionDraft, EnvBrief, SessionItem } from './store'
import type { EnvTab } from './routing'
import { Pane, TabStrip, ViewHeader, Meta, type TabItem } from './Shell'
import { WsLogView } from './WsLogView'
import { WsTestView } from './WsTestView'
import { RecordView } from './RecordView'
import { SESSION_PURPOSE_LABEL, SESSION_TYPE_LABEL } from './Sidebar'

const ENV_TABS: TabItem<EnvTab>[] = [
  { key: 'create', label: '创建会话', hint: '在这个环境下面新建一条会话' },
  { key: 'wslogs', label: '接口日志', hint: '看这个环境的接口调用报文' },
  { key: 'wstest', label: '接口调试', hint: '按报文参数调一次接口' },
  { key: 'record', label: 'FEP录像', hint: 'FEP 录制件与回放' },
]

export function EnvView({ env, tab, sessions, onTab, onCreate, onOpenSession, leading }: {
  env: EnvBrief
  tab: EnvTab
  sessions: SessionItem[]
  onTab: (t: EnvTab) => void
  onCreate: (form: SessionDraft, onStep: (step: string) => void) => Promise<void>
  onOpenSession: (s: SessionItem) => void
  leading?: ReactNode
}) {
  // 名字里的"环境"这一层信息:宿主与登录区域是设置里配的,壳子只照实显示,不改。
  const meta = (
    <>
      <Meta title="环境(设置 → 站点管理)">{env.host ? `${env.host}` : '未配置服务器地址'}</Meta>
      {env.zone && <><span className="opacity-30">·</span><Meta title="登录区域">区域 {env.zone}</Meta></>}
      <span className="opacity-30">·</span>
      <Meta title="该环境下的会话数">{sessions.length} 个会话</Meta>
    </>
  )

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col bg-background">
      <ViewHeader title={env.name} meta={meta} leading={leading} />
      <TabStrip items={ENV_TABS} active={tab} onChange={onTab} />

      <div className="flex min-h-0 min-w-0 flex-1 flex-col">
        <Pane show={tab === 'create'} label="创建会话">
          <CreateSessionPane env={env} count={sessions.length} sessions={sessions}
            onCreate={onCreate} onOpenSession={onOpenSession} />
        </Pane>
        <Pane show={tab === 'wslogs'} label="接口日志"><WsLogView /></Pane>
        <Pane show={tab === 'wstest'} label="接口调试"><WsTestView /></Pane>
        <Pane show={tab === 'record'} label="FEP录像"><RecordView /></Pane>
      </div>
    </div>
  )
}

/**
 * 创建会话:企业编号 + 作业编号(+ 可选名字)。
 *
 * 为什么 ENT 在这里填:企业编号已经**不从环境配置里来**了(hosts.sshs[].topent 已移除),
 * 它决定用哪个账号连库,而"忘了点名"与"点错了"表现完全一样 —— 所以它必须显式给。
 * 作业编号则是"这条会话要调哪个作业":录制代理录的就是它的界面。
 *
 * 默认带 proxy(用户 2026-10-18 定的):建完会话不是终点,建完还要把画面通道挂上 ——
 * 那一步在 store.createSession 里(等登录 → 起 `tt recording start` → 确认登记)。
 * 界面只负责把"现在走到哪一步"如实显示出来。
 */
function CreateSessionPane({ env, count, sessions, onCreate, onOpenSession }: {
  env: EnvBrief
  count: number
  sessions: SessionItem[]
  onCreate: (form: SessionDraft, onStep: (step: string) => void) => Promise<void>
  onOpenSession: (s: SessionItem) => void
}) {
  const [name, setName] = useState('')
  const [ent, setEnt] = useState('')
  const [prog, setProg] = useState('')
  const [busy, setBusy] = useState(false)
  const [step, setStep] = useState('')
  const [err, setErr] = useState('')
  const placeholder = `新会话 ${count + 1}`

  const submit = async () => {
    if (busy) return
    // 两道硬前置先在界面拦一下:后端也会拒,但不必为一句"没填"白花一次 SSH 握手。
    if (!ent.trim()) { setErr('企业编号（ENT）必填：它决定用哪个账号连库，环境配置里已经没有它了'); return }
    if (!prog.trim()) { setErr('作业编号必填：要录的就是这个作业的界面'); return }
    setErr('')
    setBusy(true)
    setStep('')
    try {
      await onCreate({ ent: ent.trim(), prog: prog.trim(), name: name.trim() }, setStep)
      // 成功后这条会话已经跳走了,但本面板是 keep-alive 的 —— 不清空的话下次回来
      // 还留着上一轮的 ENT/作业,再点一次就会建出一条一模一样的会话。
      setName(''); setEnt(''); setProg('')
    } catch (e: any) {
      // 建会话的失败原因必须原样显示(服务端信封里的那句话):是环境不存在、ENT 没给、
      // 还是连不上,处置完全不同。
      setErr(e?.message || String(e))
    } finally {
      setBusy(false)
      setStep('')
    }
  }

  return (
    <div className="min-h-0 flex-1 overflow-auto p-6">
      <div className="mx-auto w-full max-w-[580px]">
        <h2 className="text-[17px] font-semibold tracking-tight text-foreground">在「{env.name}」下创建会话</h2>
        <p className="mt-1 text-xs text-muted-foreground">
          会话会挂在左侧这棵树的「{env.name}」下面。建完会自动跳到它的 Debug 页。
        </p>

        <div className="mt-5 space-y-4 rounded-[12px] border border-border bg-card p-4">
          <label className="block">
            <span className="mb-1.5 block text-xs text-muted-foreground">会话名称</span>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder={placeholder}
              disabled={busy}
              onKeyDown={(e) => { if (e.key === 'Enter') void submit() }}
              className="h-8 text-[13px]" />
            <span className="mt-1 block text-[11px] text-muted-foreground/80">
              留空就用「{placeholder}」。录制靠名字落回同一条会话，所以每条都会有个名字。
            </span>
          </label>

          <div className="grid grid-cols-2 gap-3">
            <label className="block">
              <span className="mb-1.5 block text-xs text-muted-foreground">
                企业编号 <span className="text-muted-foreground/70">ENT</span>
              </span>
              <Input value={ent} onChange={(e) => setEnt(e.target.value)} placeholder="如 99"
                disabled={busy}
                onKeyDown={(e) => { if (e.key === 'Enter') void submit() }}
                className="h-8 text-[13px]" />
            </label>
            <label className="block">
              <span className="mb-1.5 block text-xs text-muted-foreground">作业编号</span>
              <Input value={prog} onChange={(e) => setProg(e.target.value)} placeholder="如 axmt500_wf"
                disabled={busy}
                onKeyDown={(e) => { if (e.key === 'Enter') void submit() }}
                className="h-8 text-[13px]" />
            </label>
          </div>
          <p className="-mt-2 text-[11px] text-muted-foreground/80">
            ENT 决定用哪个账号连库（不再沿用登录 shell 的 $TOPENT）；作业编号决定录哪个界面。
          </p>

          <div>
            <span className="mb-1.5 block text-xs text-muted-foreground">会话类型</span>
            <div className="grid grid-cols-2 gap-2">
              <TypeCard on icon={Bug} title={`${SESSION_TYPE_LABEL.debug} 会话`}
                desc="断点、单步、变量与调用栈 —— 调试那条线" onClick={() => {}} />
              <TypeCard on={false} disabled icon={Video} title={SESSION_TYPE_LABEL.fep + ' 录制会话'}
                desc="录一轮 FEP 报文、回放与对拍 —— 这一段还没接进来" onClick={() => {}} />
            </div>
          </div>

          {err && (
            <div className="rounded-[8px] border border-red-500/30 bg-red-500/10 px-3 py-2 text-[11px] leading-relaxed text-red-600 dark:text-red-400">
              {err}
            </div>
          )}

          <div className="flex flex-wrap items-center gap-2 pt-1">
            <Button size="sm" onClick={() => void submit()} disabled={busy}>
              <Plus className="h-3.5 w-3.5" />
              {busy ? '创建中…' : '创建会话'}
            </Button>
            <span className="text-[11px] text-muted-foreground">
              {busy
                ? (step || '正在创建…')
                : '默认同时挂上画面通道（proxy），并跳到它的 Debug 页。'}
            </span>
          </div>

          <p className="border-t border-border/60 pt-3 text-[11px] leading-relaxed text-muted-foreground/80">
            点「创建会话」会依次做三件事:在这个环境上建一条**空闲**会话(登录 SSH)→ 等它就绪 →
            给它挂上画面通道(起一个脱离本进程的 <code className="font-mono">tt recording start</code>)。
            作业就在这个代理后面跑,所以这条会话默认就带着录制;连不上时会在上面原样报出原因。
          </p>
        </div>

        {/* 已有会话:环境页也给一份。左侧树是导航,这里是"这个环境里有什么"的清单 ——
            建完会话后不必先回树里找它;两处点进去是同一个会话页。 */}
        <div className="mt-6">
          <div className="mb-2 flex items-center gap-2 text-xs text-muted-foreground">
            <Server className="h-3.5 w-3.5" />
            <span>这个环境下的会话</span>
            <span className="text-[11px] opacity-70">{count}</span>
          </div>
          {sessions.length === 0 ? (
            <div className="rounded-[10px] border border-dashed border-border px-3 py-3 text-[11px] text-muted-foreground">
              还没有会话。上面建一条,它会同时出现在这里和左侧「{env.name}」下面。
            </div>
          ) : (
            <div className="divide-y divide-border/60 overflow-hidden rounded-[10px] border border-border bg-card">
              {sessions.map((s) => {
                const Icon = s.type === 'fep' ? Video : Bug
                return (
                  <button key={s.id} onClick={() => onOpenSession(s)}
                    className="flex w-full items-center gap-2 px-3 py-2 text-left text-xs transition-colors hover:bg-accent/50">
                    <Icon className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                    <span className="min-w-0 flex-1 truncate text-foreground">{s.name}</span>
                    {s.prog && (
                      <span className="shrink-0 font-mono text-[11px] text-muted-foreground">{s.prog}</span>
                    )}
                    {s.ent && (
                      <span className="shrink-0 text-[11px] text-muted-foreground">ENT {s.ent}</span>
                    )}
                    <span className="shrink-0 text-[11px] text-muted-foreground">{SESSION_PURPOSE_LABEL[s.purpose]}</span>
                  </button>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function TypeCard({ on, icon: Icon, title, desc, onClick, disabled }: {
  on: boolean; icon: typeof Bug; title: string; desc: string; onClick: () => void; disabled?: boolean
}) {
  return (
    <button onClick={onClick} disabled={disabled}
      title={disabled ? '下一阶段接入：这一段还没有后端' : undefined}
      className={cn('rounded-[10px] border p-3 text-left transition-colors',
        disabled ? 'cursor-not-allowed border-border opacity-50'
          : on ? 'border-[var(--dsh-accent)] bg-accent/40' : 'border-border hover:bg-accent/30')}>
      <div className="flex items-center gap-1.5 text-[13px] font-medium text-foreground">
        <Icon className="h-4 w-4" />
        {title}
      </div>
      <p className="mt-1 text-[11px] leading-relaxed text-muted-foreground">{desc}</p>
    </button>
  )
}