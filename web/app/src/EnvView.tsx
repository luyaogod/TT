// 环境页:一个环境对应的工作面板。
//
// 三个 Tab:创建会话 / 接口日志 / 接口调试。
// 其中两个是**已有的画面**,原样搬进来;「创建会话」是外壳自己的那一屏 ——
// 它已经从"只写本机"变成**真的向后端建会话**(见 store.createSession)。
import { useState, type ReactNode } from 'react'
import { Bug, Plus } from 'lucide-react'
import { Button, Input } from '../../shared/ui'
import type { SessionDraft, EnvBrief, SessionItem } from './store'
import type { EnvTab } from './routing'
import { Pane, TabStrip, ViewHeader, Meta, type TabItem } from './Shell'
import { WsLogView } from './WsLogView'
import { WsTestView } from './WsTestView'

const ENV_TABS: TabItem<EnvTab>[] = [
  { key: 'create', label: '创建会话', hint: '在这个环境下面新建一条会话' },
  { key: 'wslogs', label: '接口日志', hint: '看这个环境的接口调用报文' },
  { key: 'wstest', label: '接口调试', hint: '按报文参数调一次接口' },
]

export function EnvView({ env, tab, sessions, onTab, onCreate, leading }: {
  env: EnvBrief
  tab: EnvTab
  sessions: SessionItem[]
  onTab: (t: EnvTab) => void
  onCreate: (form: SessionDraft, onStep: (step: string) => void) => Promise<void>
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
          <CreateSessionPane env={env} count={sessions.length} onCreate={onCreate} />
        </Pane>
        <Pane show={tab === 'wslogs'} label="接口日志"><WsLogView /></Pane>
        <Pane show={tab === 'wstest'} label="接口调试"><WsTestView /></Pane>
      </div>
    </div>
  )
}

/**
 * 创建会话:企业编号 + 作业编号(+ 可选名字)。
 *
 * 为什么 ENT 在这里填:企业编号已经**不从环境配置里来**了(hosts.sshs[].topent 已移除),
 * 它决定用哪个账号连库,而"忘了点名"与"点错了"表现完全一样 —— 所以它必须显式给。
 * 作业编号则是"这条会话要调哪个作业"。
 */
function CreateSessionPane({ env, count, onCreate }: {
  env: EnvBrief
  count: number
  onCreate: (form: SessionDraft, onStep: (step: string) => void) => Promise<void>
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
    if (!prog.trim()) { setErr('作业编号必填：它就是这条会话要启动的作业'); return }
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
          会话会挂在左侧这棵树的「{env.name}」下面。建完自动跳到它的 Debug 页。
        </p>

        <div className="mt-5 space-y-4 rounded-[12px] border border-border bg-card p-4">
          <label className="block">
            <span className="mb-1.5 block text-xs text-muted-foreground">会话名称</span>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder={placeholder}
              disabled={busy}
              onKeyDown={(e) => { if (e.key === 'Enter') void submit() }}
              className="h-8 text-[13px]" />
            <span className="mt-1 block text-[11px] text-muted-foreground/80">
              留空就用「{placeholder}」。每条会话都会有个名字，便于在树里认。
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
            ENT 决定用哪个账号连库（不再沿用登录 shell 的 $TOPENT）；作业编号决定要启动哪个作业。
          </p>

          <div className="flex items-start gap-2 rounded-[8px] border border-border bg-muted/30 px-3 py-2 text-[11px] leading-relaxed text-muted-foreground">
            <Bug className="mt-0.5 h-3.5 w-3.5 shrink-0 opacity-70" />
            <span>建出来的是**调试会话**：停在作业入口，可以下断点、单步、看变量。</span>
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
          </div>
        </div>

      </div>
    </div>
  )
}
