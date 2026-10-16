// 回放进度（计划 C 的 D7）：一条进度 + 每一步的结果，长在明细页的顶上。
//
// 用户的原话是"播放过账中我可以观察走到哪了"。所以这一块只回答三件事：
//
//	走到第几步 / 这一步是什么 / 运行期回没回
//
// 四条纪律（都是实测或计划里钉过的）：
//
//  1. **进行中的步只画开始标记，不画时长**（D7.3）：`waiting` 那一步显示"已等 Nms"，
//     而不是一个 0 —— 没有回包 ≠ 0 毫秒。
//  2. **"归不了"与"无回包"分开**：`suspect=true`（点击类事件本该有回包）才是黄灯的"8s 无回包"，
//     其余没回包的是灰灯的"本来无应答"。两者的下一步完全不同。
//  3. **跟随尾部，但用户向上滚动就停**（D7.4）：勾选框显示当前状态，可手动恢复。
//     不看这条的话，用户往回翻历史时会被每一拍的新行拽回底部 —— 那是没法读的。
//  4. **数字来自服务端**（D7.5）：`progress` 那一组计数是服务端按 jsonl 收敛出来的，
//     界面自己数一遍迟早两边对不上（重连补发、暂停、跳过都会让行数与步数不等）。
import { useEffect, useMemo, useRef, useState } from 'react'
import { Pause, Play, SkipForward, Square } from 'lucide-react'
import type { ReplayProgress, ReplayStep, ReplayStepResult } from './api'
import { fmtBytes, fmtDur } from './recordingTime'

/** 一步的徽标：文字 + 颜色 + 说明。五种终态 + 两个中间态（照计划 D7.2）。 */
function badge(s: ReplayStep): { text: string; cls: string; why: string } {
  switch (s.result) {
    case 'sent':
      return {
        text: '已确认', cls: 'border-emerald-500/30 text-emerald-600 dark:text-emerald-400',
        why: '运行期回了包' + (s.respAt && s.sentAt ? `（${fmtDur(s.respAt - s.sentAt)}）` : '') +
          (s.runtimeBytes ? `，推回 ${fmtBytes(s.runtimeBytes)}` : ''),
      }
    case 'no-response':
      return s.suspect
        ? {
          text: '8s 无回包（可疑）', cls: 'border-amber-500/40 text-amber-600 dark:text-amber-400',
          why: '点击类事件**通常一定有反应**，而这一步 8 秒内没有任何回包 —— 去看运行期/作业状态（' + (s.reason || '') + '）',
        }
        : {
          text: '本来无应答', cls: 'border-border text-muted-foreground',
          why: '这一步（光标/键盘/状态上报）本来就没有应答 —— 不是失败',
        }
    case 'skipped':
      return { text: '跳过', cls: 'border-border text-muted-foreground', why: s.reason || '按要求跳过' }
    case 'failed':
      return { text: '失败', cls: 'border-red-500/30 text-red-600 dark:text-red-400', why: s.reason || '发送出错' }
    case 'waiting':
      return { text: '进行中', cls: 'border-sky-500/30 text-sky-600 dark:text-sky-400', why: '已发出，正在等运行期回包（**进行中不画时长**）' }
    default:
      return { text: '已暂停', cls: 'border-border text-muted-foreground', why: s.reason || '等「继续」或「单步」' }
  }
}

/** 一步是不是"已经有终态"（waiting/paused 都还没定）。 */
function settled(r: ReplayStepResult): boolean {
  return r === 'sent' || r === 'no-response' || r === 'skipped' || r === 'failed'
}

interface Props {
  id: string
  /** 服务端给的计数（权威：界面不自己数） */
  progress: ReplayProgress
  /** 按 `n` 收敛后的每一步最新状态 */
  steps: ReplayStep[]
  running: boolean
  note?: string
  /** 人和回放下的命令（暂停/继续/单步/停止）。失败由调用方 toast。 */
  onCtl: (cmd: 'pause' | 'resume' | 'step' | 'stop') => void
  /** 重新读一次进度（按钮"刷新"与兜底轮询用）。 */
  onRefresh: () => void
}

export function ReplayProgressBar(p: Props) {
  const { progress: pr, steps } = p
  // 正在等回包的那一步（服务端按 n 收敛过，最多一步）
  const waiting = useMemo(() => steps.find((s) => s.result === 'waiting'), [steps])
  const now = useNow(waiting ? 800 : 0) // 只在等待时走时钟（否则白重渲染）
  const waited = waiting?.sentAt ? now - waiting.sentAt : 0

  const done = pr.total > 0 ? Math.min(100, Math.round((pr.done / pr.total) * 100)) : 0

  // 跟随尾部：默认开；用户往上滚就关（否则每一拍新行都会把人拽回底部）。
  const listRef = useRef<HTMLDivElement | null>(null)
  const [follow, setFollow] = useState(true)
  useEffect(() => {
    if (!follow || !listRef.current) return
    listRef.current.scrollTop = listRef.current.scrollHeight
  }, [follow, steps.length])
  const onScroll = () => {
    const el = listRef.current
    if (!el) return
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 8
    if (!atBottom && follow) setFollow(false)
    if (atBottom && !follow) setFollow(true)
  }

  const label = pr.running
    ? (pr.paused ? '已暂停' : waiting ? `进行中 · 第 ${waiting.n}/${pr.total} 步` : `正在回放 · 第 ${pr.done}/${pr.total} 步`)
    : `完成 ${pr.done}/${pr.total} · 失败 ${pr.failed}`

  return (
    <div className="border-b border-border bg-card/40">
      <div className="flex flex-wrap items-center gap-2 px-2 py-1 text-[11px]">
        <span className="font-mono">{label}</span>
        <div className="h-1.5 w-40 overflow-hidden rounded-ui bg-muted" title={`${pr.done}/${pr.total} 步已有结果`}>
          <div className={`h-full ${pr.failed ? 'bg-amber-500' : 'bg-emerald-500'}`} style={{ width: `${done}%` }} />
        </div>
        <span className="text-muted-foreground">
          已确认 {pr.sent} · 无应答 {pr.noResp}{pr.suspect ? `（其中可疑 ${pr.suspect}）` : ''} · 跳过 {pr.skipped} · 失败 {pr.failed}
        </span>
        {waiting && (
          <span className="font-mono text-sky-600 dark:text-sky-400"
            title="已发出，正在等运行期回包（最长 8 秒）。**进行中不画时长** —— 没有回包不等于 0 毫秒">
            已等 {fmtWait(waited)}
          </span>
        )}
        {waiting && (
          <span className="max-w-[420px] truncate font-mono text-muted-foreground" title={waiting.event}>
            {waiting.event}
          </span>
        )}
        {/* 命令按钮只在"还在跑"时给：已经结束的一轮点了也没有接收方（服务端会 409）。 */}
        {pr.running && (
          <span className="ml-auto flex items-center gap-1">
            {pr.paused ? (
              <button onClick={() => p.onCtl('resume')} className={btn}
                title="继续：从停住的那一步接着发">
                <Play size={11} /> 继续
              </button>
            ) : (
              <button onClick={() => p.onCtl('pause')} className={btn}
                title="暂停：在当前这一步的回包之后生效（它可能正在等运行期，最长 8 秒）">
                <Pause size={11} /> 暂停
              </button>
            )}
            <button onClick={() => p.onCtl('step')} className={btn} title="单步：放行当前这一步，然后重新停下">
              <SkipForward size={11} /> 单步
            </button>
            <button onClick={() => p.onCtl('stop')} className={`${btn} text-red-600 dark:text-red-400`}
              title="停止：不再发后面的步骤（已经发出去的收不回来）">
              <Square size={11} /> 停止
            </button>
          </span>
        )}
        <span className={`${pr.running ? '' : 'ml-auto'} flex items-center gap-2 text-muted-foreground`}>
          <label className="flex cursor-pointer items-center gap-1" title="跟随尾部：新的一步自动滚进来。你往上滚就自动关掉，滚回底部自动恢复">
            <input type="checkbox" checked={follow} onChange={(e) => setFollow(e.target.checked)} />
            跟随
          </label>
          {!pr.running && <button onClick={p.onRefresh} className={btn}>刷新</button>}
        </span>
      </div>
      {p.note && <div className="px-2 pb-1 text-[10px] text-muted-foreground">{p.note}</div>}

      {/* 每一步一行。最多列最近 200 步：几千步的录制全渲染会把这一块拖死，
          而"看走到哪了"只需要尾部（前面那些用下面的表与全文流看）。 */}
      <div ref={listRef} onScroll={onScroll} className="max-h-28 overflow-auto border-t border-border/60 px-2 py-0.5">
        {(steps.length > 200 ? steps.slice(steps.length - 200) : steps).map((s) => {
          const b = badge(s)
          return (
            <div key={s.n} className="flex items-center gap-2 py-px font-mono text-[10px]">
              <span className="w-10 shrink-0 text-right text-muted-foreground">#{s.n}</span>
              <span className={`shrink-0 rounded-ui border px-1 ${b.cls}`} title={b.why}>{b.text}</span>
              <span className="shrink-0 text-muted-foreground">{s.kind || ''}</span>
              <span className="min-w-0 flex-1 truncate text-muted-foreground" title={s.event}>{s.event}</span>
              <span className="shrink-0 text-muted-foreground">
                {s.sentAt ? (settled(s.result) ? (s.respAt ? fmtDur(s.respAt - s.sentAt) : '—') : '…') : ''}
              </span>
            </div>
          )
        })}
        {steps.length > 200 && (
          <div className="py-px text-[10px] text-muted-foreground">（只列最近 200 步，共 {steps.length} 步）</div>
        )}
      </div>
    </div>
  )
}

const btn = 'inline-flex items-center gap-1 rounded-ui border border-border px-1.5 py-0.5 hover:bg-accent'

/**
 * "已等"的显示。**与 `fmtDur` 分开**：`fmtDur(0)` 给的是"——"（它的意思是"没有这个数"，
 * 用在"耗时"那一栏是对的），而这里是真的"刚开始等，已等 0 ms"——一个正在等的步
 * 显示"已等 ——"会让人以为它没在等。
 */
function fmtWait(ms: number): string {
  if (!(ms > 0)) return '0 ms'
  return fmtDur(ms)
}

/** 只在 `ms > 0` 时走时钟（等回包时显示"已等 Nms"，其余时候不重渲染）。 */
function useNow(ms: number): number {
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    if (ms <= 0) return
    const t = window.setInterval(() => setNow(Date.now()), ms)
    return () => window.clearInterval(t)
  }, [ms])
  return now
}
