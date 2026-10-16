// 回放进度（计划 C 的 D7）：**一行**说清"回放到哪了"。
//
// 用户的原话是"播放过账中我可以观察走到哪了"，后来又删掉了逐条确认清单
// （2026-10-16："移除确认清单"）：73 行 `已确认` 是重复信息 —— 计数那一行已经说了
// 各是多少，每步的**明细在下面全文流表里**（对应行上有同一个徽标）。
// 所以这里只留：进度条 + 一句计数 + 正在等的那一步 + 四个命令。
//
// 三条纪律（都是实测或计划里钉过的）：
//
//  1. **进行中的步只画开始标记，不画时长**（D7.3）：`waiting` 那一步显示"已等 Nms"，
//     而不是一个 0 —— 没有回包 ≠ 0 毫秒。
//  2. **"归不了"与"无回包"分开**：`suspect=true`（点击类事件本该有回包）才值得去看运行期，
//     计数里也单独标出来。
//  3. **数字来自服务端**（D7.5）：`progress` 那一组计数是服务端按 jsonl 收敛出来的，
//     界面自己数一遍迟早两边对不上（重连补发、暂停、跳过都会让行数与步数不等）。
import { useEffect, useMemo, useState } from 'react'
import { Pause, Play, SkipForward, Square } from 'lucide-react'
import type { ReplayProgress, ReplayStep } from './api'
import { fmtDur } from './recordingTime'

interface Props {
  id: string
  /** 服务端给的计数（权威：界面不自己数） */
  progress: ReplayProgress
  /** 按 `n` 收敛后的每一步最新状态（**只用来找"正在等的那一步"**） */
  steps: ReplayStep[]
  running: boolean
  note?: string
  /** 人和回放下的命令（暂停/继续/单步/停止）。失败由调用方 toast。 */
  onCtl: (cmd: 'pause' | 'resume' | 'step' | 'stop') => void
}

export function ReplayProgressBar(p: Props) {
  const { progress: pr, steps } = p
  // 正在等回包的那一步（服务端按 n 收敛过，最多一步）
  const waiting = useMemo(() => steps.find((s) => s.result === 'waiting'), [steps])
  const now = useNow(waiting ? 800 : 0) // 只在等待时走时钟（否则白重渲染）
  const waited = waiting?.sentAt ? now - waiting.sentAt : 0

  const done = pr.total > 0 ? Math.min(100, Math.round((pr.done / pr.total) * 100)) : 0
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
      </div>
      {p.note && <div className="px-2 pb-1 text-[10px] text-muted-foreground">{p.note}</div>}
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
