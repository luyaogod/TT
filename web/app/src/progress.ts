// 进度条与"那条数值行"的纯函数。
//
// 为什么单独成文件：这些规则（速度怎么算、打包阶段显示什么、百分比的分母是谁）
// 是最容易被"顺手改一下"改坏的地方，放在组件里只能靠肉眼验；放这里就能被
// scripts/progress.test.mjs 直接测。
//
// 速度为什么由前端自己采样：后端只给"累计字节 + 自任务开始的总耗时"。
// 两者相除等于把连接、探测、服务器打包的时间全摊进下载速度里，而且是从 0 开始的
// 平均值 —— 一开始偏小、越跑越接近真实值，看着永远对不上。这里按最近窗口的差分算。

export type Sample = { t: number; bytes: number }

/** 追加一个采样点，并丢掉窗口外的老点（至少保留两个点才算得出速度）。 */
export function pushSample(list: Sample[], bytes: number, now: number, windowMs = 5000): Sample[] {
  const next = list.concat({ t: now, bytes })
  const kept = next.filter((s) => s.t >= now - windowMs)
  return kept.length >= 2 ? kept : next.slice(-2)
}

/** 采样点 → 字节/秒。不足两点、时间没走、字节没涨都返回 0（调用方据此不显示）。 */
export function rateOf(list: Sample[]): number {
  if (list.length < 2) return 0
  const first = list[0]
  const last = list[list.length - 1]
  const dt = (last.t - first.t) / 1000
  const db = last.bytes - first.bytes
  if (dt <= 0 || db <= 0) return 0
  return db / dt
}

export function fmtRate(bps: number): string {
  if (!(bps > 0)) return ''
  if (bps >= 1024 * 1024) return (bps / (1024 * 1024)).toFixed(1) + ' MB/s'
  return (bps / 1024).toFixed(0) + ' KB/s'
}

export function fmtBytes(n: number): string {
  if (!n || n <= 0) return '0 B'
  if (n >= 1 << 30) return (n / (1 << 30)).toFixed(2) + ' GB'
  if (n >= 1 << 20) return (n / (1 << 20)).toFixed(1) + ' MB'
  if (n >= 1 << 10) return (n / (1 << 10)).toFixed(0) + ' KB'
  return n + ' B'
}

/** 镜像任务里前端真正用得上的那几个字段（不 import SettingsView 的类型，便于测试）。 */
export interface MirrorJobLike {
  phase: string
  bytes: number
  total: number
  elapsed: string
  files: number
  running: boolean
}

/**
 * 进度百分比。
 *
 * 分母是谁由阶段决定：
 *   · download —— 归档的真实字节数（后端在下载开始时给出）；
 *   · pack    —— **估算**的归档大小（后端量完源文件后按经验压缩比折算）；
 *   · done    —— 100。
 * 总量拿不到时返回 undefined：进度条交给调用方画成"不确定"，而不是假装 0%。
 */
export function mirrorPct(j: MirrorJobLike): number | undefined {
  if (j.phase === 'done') return 100
  if (j.total > 0) return Math.max(0, Math.min(100, Math.round((j.bytes * 100) / j.total)))
  return undefined
}

/**
 * 数值行（裸值，不写前缀标签）。
 *
 * 打包阶段**不显示速度**：那会儿后端报的字节是"服务器上归档文件已经写到多大"，
 * 拿它除以耗时得到的是打包速率，不是这条任务要看的传输速率 —— 显示出来只会
 * 让人以为下载已经开始了。这一阶段该看的是"已打包多少 / 预估多大"。
 */
export function mirrorCells(j: MirrorJobLike, rateBps: number): string[] {
  const pack = j.phase === 'pack'
  const size = pack
    ? (j.bytes > 0 ? `已打包 ${fmtBytes(j.bytes)}` : '')
    : fmtBytes(j.bytes) + (j.total > 0 ? ` / ${fmtBytes(j.total)}` : '')
  const cells = [
    size,
    pack ? (j.total > 0 ? `预估 ${fmtBytes(j.total)}` : '') : fmtRate(rateBps),
    j.elapsed || '',
    !pack && j.files > 0 ? `${j.files} 个文件` : '',
  ]
  return cells.filter(Boolean)
}
