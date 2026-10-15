// 「会话」sheet:管理全局单一常驻会话(每个环境一条,通常就是默认环境那一条)。
// 会话 = 与某 T100 环境的 SSH+登录态,debug 只是在其上跑一轮;结束调试/程序退出都只回空闲。
// 本面板负责 切换(换到另一环境并重连)/ 重启(断开重连同环境)/ 结束(彻底释放连接);
// 这些操作都会先结束当前 debug(由后端收口)。行只显示环境名 + 状态,保持精简。
// 两个可折叠区域(仿 debug 运行区手风琴):「环境变量」(ENT:开会话的前置参数,
// 随切换/重启请求带给后端,**不落任何配置**)、「会话」(环境行列表)。
// 顺序是刻意的:先填 ENT 再选环境 —— 选环境那一步就会用到它,且没填 ENT 不放行。
import { useCallback, useEffect, useState, type ReactNode } from 'react'
import { ArrowRightLeft, RefreshCw, RotateCcw, Power } from 'lucide-react'
import { useStore } from './store'
import { api } from './api'
import { Input } from '../../shared/ui'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from '../../shared/ui-radix'
import { useToast } from './toast'

const STATE_LABEL: Record<string, string> = {
  idle: '空闲', loading: '连接中…', stopped: '已停站', running: '运行中', exit: '已断开',
}
const STATE_DOT: Record<string, string> = {
  idle: 'bg-emerald-500',
  loading: 'bg-amber-400',
  stopped: 'bg-[#cc6633]',
  running: 'bg-[#0078d4]',
  exit: 'bg-muted-foreground/50',
}

// 折叠区头(仿 RightPanels 手风琴):标题(可点击)…附加控件…折叠箭头;
// topLine 为真时块顶加分割线(与会话内容分隔,避免与上一块底边线叠成双线)
function SectionHead({ title, open, onToggle, topLine, extra }: {
  title: string; open: boolean; onToggle: () => void; topLine?: boolean; extra?: ReactNode
}) {
  const tip = open ? `收起${title}` : `展开${title}`
  return (
    <div className={`flex h-8 shrink-0 items-center border-b border-border px-2.5 text-xs font-medium text-muted-foreground ${
      topLine ? 'border-t border-t-border' : ''
    }`}>
      <button onClick={onToggle} title={tip}
        className="min-w-0 flex-1 truncate text-left hover:text-foreground">
        {title}
      </button>
      {extra}
      <button onClick={onToggle} title={tip}
        className="p-0.5 text-muted-foreground transition-colors hover:text-foreground">
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
          stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
          className={`shrink-0 text-muted-foreground transition-transform duration-200 ${open ? 'rotate-180' : ''}`}>
          <path d="m6 9 6 6 6-6" />
        </svg>
      </button>
    </div>
  )
}

export function SessionPanel() {
  const sessionId = useStore((s) => s.sessionId)
  const sessionEnv = useStore((s) => s.sessionEnv)
  const state = useStore((s) => s.state)
  const sessionOp = useStore((s) => s.sessionOp)

  const [envNames, setEnvNames] = useState<string[]>([])
  const [busy, setBusy] = useState<string | null>(null)
  // 失败提示统一走 toast(见 runOp / entMissing),这里不再留"面板级错误"的 state。
  const toast = useToast()

  // 两个可折叠区域(默认展开,仿 debug 运行区)
  const [open, setOpen] = useState({ session: true, env: true })
  const toggle = (k: 'session' | 'env') => setOpen((o) => ({ ...o, [k]: !o[k] }))

  // 环境清单来自设置(名称即会话目标);当前环境由会话本身决定(选环境 = 连会话)
  const load = useCallback(() => {
    void api.settings().then((c: any) => {
      const names = (c.sshs || [])
        .map((e: any) => e.name || `${e.host || ''}-${e.zone || ''}`.replace(/-$/, ''))
        .filter(Boolean)
      setEnvNames(names)
    }).catch(() => {})
  }, [])
  useEffect(() => { load() }, [load])
  // 连接中 loading→idle 等状态跟随;每 3s 轻量刷新一次
  useEffect(() => {
    const t = window.setInterval(load, 3000)
    return () => window.clearInterval(t)
  }, [load])

  // 会话所属环境:直接取会话快照里的环境名(不再有"默认环境"这种配置项)
  const curEnv = sessionEnv || ''
  // 行 = 已配置环境;当前会话不在列表(顶层直连等)时补一行。
  // 圆点只表示"是否连着会话":未连接一律灰点,连上的那行才按状态上色(空闲绿/运行蓝/停站橙)。
  const connectedEnv = sessionId && curEnv ? [curEnv] : []
  const rows = Array.from(new Set([...envNames, ...(curEnv ? [curEnv] : [])]))

  // 会话操作:有副作用的先弹 shadcn 二次确认(替代浏览器 confirm),确认后再执行
  const [pending, setPending] = useState<{ op: 'close' | 'restart' | 'switch'; env?: string; title: string; desc: string } | null>(null)

  const runOp = async (op: 'close' | 'restart' | 'switch', env?: string) => {
    setBusy(`${op}:${env || ''}`)
    try {
      // ENT 随操作带走:环境配置里已不存企业编号,不给就是"沿用登录 shell 的 $TOPENT"。
      await sessionOp(op, env, ent.trim())
    } catch (e: any) {
      // 会话操作的失败原先内联在这个窄面板里(一行小红字),会把下面的环境列表挤动。
      const label = op === 'close' ? '结束会话' : op === 'restart' ? '重启会话' : '切换会话'
      toast.error(label + '失败', { description: e?.message || String(e) })
    } finally {
      setBusy(null)
    }
  }

  // **必须填 ENT 才能开会话。**
  //
  // 为什么卡在这里:企业编号决定用哪个账号(schema)连库。空着去连,企业就由这台机器
  // 登录脚本里的 $TOPENT 决定 —— 而"用错了企业"的表现是查不到数据,跟"数据不存在"
  // 一模一样。所以宁可不放行,也不替用户接受一个来路不明的默认值。
  // (仅对"要建立/重建连接"的操作生效;close 随时可做。)
  //
  // 拦下来时说清原因:走统一消息出口(原先是在输入框下面挤一行红字,
  // 会把下面的环境列表顶下去)。按钮本身也已经显灰禁用,这里是第二道。
  const entMissing = () => {
    if (ent.trim()) return false
    toast.error('请先填 ENT(企业编号)', {
      description: '它决定这次连哪个企业/账号。留空会沿用机器登录默认,容易连错企业。',
    })
    return true
  }

  const doOp = (op: 'close' | 'restart' | 'switch', env?: string) => {
    // close 只是断开,不需要 ENT;switch/restart 都会重新登录,必须点名企业。
    if (op !== 'close' && entMissing()) return
    const hasRun = !!sessionId && state !== 'idle' && state !== '' && state !== 'exit'
    const envLabel = env || curEnv || ''
    const entLabel = ent.trim() ? `(ENT ${ent.trim()})` : ''
    if (op === 'close') {
      setPending({ op, env, title: `结束会话「${envLabel}」?`, desc: '将断开该环境的 SSH 连接并结束当前调试。' })
    } else if (op === 'restart') {
      setPending({ op, env, title: `重启会话「${envLabel}」?`, desc: `将断开并重新连接该环境,会先结束当前调试。${entLabel}` })
    } else if (op === 'switch' && hasRun) {
      setPending({ op, env, title: `切换会话到「${envLabel}」?`, desc: `会先结束当前调试并断开当前连接。${entLabel}` })
    } else {
      void runOp(op, env) // 空闲态切换无需确认
    }
  }

  const confirmPending = () => {
    const p = pending
    setPending(null)
    if (p) void runOp(p.op, p.env)
  }

  // ---- 环境变量:ENT(企业编号) ----
  //
  // **这里只是一个前端参数**,不再"保存"到任何地方:
  //   · 环境配置里已经没有企业编号了(hosts.sshs[].topent 已移除);
  //   · 它的作用范围是"下一次开会话/重启会话" —— 值随 switch/restart 请求带给后端,
  //     后端在登录后把它 export 进宿主 shell。
  // 所以没有保存按钮,也没有"配置值 vs 覆盖值"那套三方对比。
  //
  // ent       = 输入框里的值(要用哪个企业);
  // entShell  = 当前会话登录后回读到的真实 $TOPENT(没给 ENT 时它就是实际生效值)。
  const [ent, setEnt] = useState('')
  const [entShell, setEntShell] = useState('')
  const mode = useStore((s) => s.mode)

  // **输入框永远可编辑。**
  //
  // 它曾经被 `disabled={!!sessionId && mode!=='collab'}` 锁住 —— 那是把它当"会话内属性"
  // 看待时的遗留:没有会话(恰恰是最需要填 ENT 的时候)它反而是灰的,填不进去。
  // 现在它只是一个本地参数(下次开会话用哪个企业),与会话存不存在无关。
  // 唯一保留的例外是协作模式:那时 ENT 归 AI(它会改企业编号,直接影响程序连哪个库),
  // 人工改会被覆盖,所以照样锁住并说明原因。
  const entLocked = mode === 'collab'

  // 依赖里必须带 state:会话先建好(sessionId 先到、state=loading),登录与 ENT 回读
  // 完成后才有值;只在 sessionId 变化时取会拿到登录前的空值。
  const loadEnt = useCallback(() => {
    if (!sessionId) { setEntShell(''); return }
    void api.snapshot(sessionId).then((snap: any) => {
      setEntShell(snap.topentShell || '')
      // 会话已有的值回填到输入框(用户没改过时),便于确认"这次连的是哪个企业"。
      setEnt((cur) => cur || snap.topentSession || '')
    }).catch(() => {})
  }, [sessionId, state])
  useEffect(() => { loadEnt() }, [loadEnt])

  return (
    <div className="flex h-full min-h-0 w-full flex-col overflow-hidden bg-background">
      <div className="min-h-0 flex-1 overflow-auto">
        {/* 区域一:环境变量(ENT)。
            **排在「会话」之前**:ENT 是开会话的前置条件 —— 选环境那一步就会用到它,
            所以顺序上先填参数、再选环境。 */}
        <SectionHead title="环境变量" open={open.env} onToggle={() => toggle('env')} />
        {open.env && (
          <div className="py-1">
            <div className="flex items-center gap-1.5 px-2 py-0.5">
              <span className="w-[52px] shrink-0 text-xs text-muted-foreground"
                title="TOPENT(企业编号):T100 运行时的企业环境,作业运行与数据库连接以此为准">TOPENT</span>
              <Input value={ent} disabled={entLocked}
                onChange={(e) => setEnt(e.target.value)}
                placeholder="必填,如 99"
                title={entLocked
                  ? '协作模式下企业编号归 AI(它会直接改会话的企业编号);要自己填请先切回纯人工模式'
                  : '企业编号(ENT):决定这次连哪个企业/账号。它随「切换/重启会话」带给后端,在登录后 export 到宿主 shell;留空不放行 —— 空着连会沿用机器登录默认,容易连错企业'}
                className="h-6 min-w-0 flex-1 px-1.5 font-mono text-xs" />
            </div>
            {/* 只在**有内容可说**时才起这一行(填了值 / 已连上且读到了实际企业)。
                从前留空时也硬写一句"未指定;将沿用机器登录的企业(不推荐)" —— 那时输入框
                本身就是灰的、填不进去,这句话既无效又自相矛盾,所以整句删掉。 */}
            {(ent.trim() || entShell) && (
              <div className="px-2.5 pb-1 text-[11px] text-muted-foreground">
                {ent.trim()
                  ? (entShell && entShell === ent.trim()
                    ? '当前会话正在用这个企业'
                    : `${sessionId ? '下次「重启/切换会话」生效' : '建立会话时生效'}${entShell ? `;当前 shell 实际为 ${entShell}` : ''}`)
                  : `当前 shell 实际为 ${entShell}`}
              </div>
            )}
          </div>
        )}

        {/* 区域二:会话(环境行列表) */}
        <SectionHead title="会话" open={open.session} onToggle={() => toggle('session')} topLine={open.env}
          extra={
            <button title="刷新环境/会话状态" onClick={load}
              className="ml-1 p-0.5 text-muted-foreground transition-colors hover:text-foreground">
              <RefreshCw className="h-3.5 w-3.5" />
            </button>
          } />
        {open.session && (
          <div className="py-1">
            {rows.length === 0 && (
              <div className="space-y-1 p-2 text-xs text-muted-foreground">
                <div>尚未配置环境(设置 → 环境)。</div>
                <div>加一个环境后在这里点名连接(需先填上方 TOPENT)。</div>
              </div>
            )}
            {rows.map((name) => {
              const connected = connectedEnv.length > 0 && name === curEnv
              const st = connected ? (state || '') : ''
              // 没填 ENT 就不让"开会话":按钮显灰禁用(而不是等点了才弹错)。
              // 「结束会话」不受此限 —— 断开连接不需要知道企业编号。
              const needEnt = !ent.trim()
              return (
                <div key={name}
                  className={`group flex h-7 items-center gap-1.5 px-2.5 text-xs ${
                    connected ? 'bg-accent/40' : 'hover:bg-accent/30'
                  }`}
                  title={connected
                    ? `当前会话 · ${STATE_LABEL[st] || st}`
                    : (needEnt ? `环境 ${name}(请先在上方填 TOPENT 企业编号)` : `环境 ${name}`)}>
                  <span className={`h-2 w-2 shrink-0 rounded-full ${connected ? (STATE_DOT[st] || 'bg-accent') : 'bg-muted-foreground/25'}`} />
                  <span className="min-w-0 flex-1 truncate font-medium">{name}</span>
                  {connected && (
                    <span className="shrink-0 text-[11px] text-muted-foreground">{STATE_LABEL[st] || st}</span>
                  )}
                  {connected ? (
                    <span className={`flex shrink-0 items-center gap-0.5 ${needEnt ? 'opacity-60' : 'opacity-70 group-hover:opacity-100'}`}>
                      <button
                        title={needEnt ? '请先填 TOPENT(企业编号)才能重启会话' : '重启会话(断开重连同环境)'}
                        disabled={!!busy || needEnt}
                        onClick={() => void doOp('restart', name)}
                        className="p-1 text-muted-foreground transition-colors hover:text-foreground disabled:pointer-events-none disabled:opacity-30 disabled:hover:text-muted-foreground">
                        <RotateCcw className="h-3.5 w-3.5" />
                      </button>
                      <button title="结束会话(彻底断开连接)" disabled={!!busy}
                        onClick={() => void doOp('close')}
                        className="p-1 text-muted-foreground transition-colors hover:text-red-600 dark:hover:text-red-400 disabled:pointer-events-none disabled:opacity-30">
                        <Power className="h-3.5 w-3.5" />
                      </button>
                    </span>
                  ) : (
                    <span className={`flex shrink-0 items-center gap-0.5 ${needEnt ? 'opacity-60' : 'opacity-70 group-hover:opacity-100'}`}>
                      <button
                        title={needEnt
                          ? '请先在上方填 TOPENT(企业编号)才能连接该环境'
                          : `切换会话到 ${name}(结束当前 debug 并连接该环境)`}
                        disabled={!!busy || needEnt}
                        onClick={() => void doOp('switch', name)}
                        className="p-1 text-muted-foreground transition-colors hover:text-foreground disabled:pointer-events-none disabled:opacity-30">
                        <ArrowRightLeft className="h-3.5 w-3.5" />
                      </button>
                    </span>
                  )}
                </div>
              )
            })}
          </div>
        )}
      </div>

      {/* 会话操作二次确认(shadcn AlertDialog;确认后执行切换/重启/结束) */}
      <AlertDialog open={pending !== null} onOpenChange={(o) => { if (!o) setPending(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{pending?.title || ''}</AlertDialogTitle>
            <AlertDialogDescription>{pending?.desc || ''}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={confirmPending}>确定</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
