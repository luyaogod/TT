// 录制明细页的**唯一**时间/位置刻度实现。
//
// 为什么必须只有这一份:条带上有一根游标、表格每一行有一个小方块,它们指的都是"这一帧在
// 时间轴上的位置"。两处各写一遍 `x = (t-t0)/(t1-t0)*w`,迟早出现"游标与行里的小方块指着
// 不同时刻" —— 而这正是排障时唯一能用来对齐两个方向的东西(两个方向是两条独立的字节流,
// 序号分属两侧,时间是对得上的唯一坐标)。
//
// 两种横轴(见 `StripDomain`):
//
//	time  —— 有帧级时间时用它。横轴 = 时刻。
//	bytes —— 没有帧级时间时(老件)用它并**明说**:横轴 = 各方向文件里的字节序。
//	         这时"位置"不是时间,但"图片缓存夹在哪两帧之间"照样看得见 —— 而那正是
//	         老件唯一能问的问题。绝不用字节序插值出假的时刻。

/** 横轴类型。`bytes` 是降级形态,界面必须把这件事说出来。 */
export type AxisMode = 'time' | 'bytes'

export interface Domain {
  mode: AxisMode
  min: number
  max: number
}

/** 一帧(或一个动作)在横轴上的取值:t>0 时给时刻,否则给字节偏移。 */
export function axisValue(f: { t: number; o: number }, mode: AxisMode): number {
  return mode === 'time' ? f.t : f.o
}

/** 把取值域铺到 [0,width] 上。宽 0 或域退化时返回常量函数(不产生 NaN 坐标)。 */
export function makeScale(d: Domain, width: number): (v: number) => number {
  const span = d.max - d.min
  if (!(width > 0) || !(span > 0)) return () => 0
  return (v: number) => ((v - d.min) / span) * width
}

/**
 * 域:数据的最小/最大取值,两端各留一点余量。
 *
 * 为什么留余量:块画在边界上时右半边会被裁掉,而"最后一帧"恰恰是最常看的那一帧。
 */
export function domainOf(values: number[], mode: AxisMode, pad = 0.01): Domain {
  const vals = values.filter((v) => Number.isFinite(v))
  if (vals.length === 0) return { mode, min: 0, max: 1 }
  let min = Math.min(...vals)
  let max = Math.max(...vals)
  if (max === min) {
    // 单点(只有一个时刻):给一个对称的小窗口,否则 span=0 所有块都堆在 0 上
    const d = mode === 'time' ? 50 : 1
    min -= d
    max += d
  } else {
    const d = (max - min) * pad
    min -= d
    max += d
  }
  return { mode, min, max }
}

/** 把域夹在边界内(放大/平移时不许跑到数据之外)。 */
export function clampDomain(d: Domain, bounds: Domain): Domain {
  const span = d.max - d.min
  const bSpan = bounds.max - bounds.min
  if (span >= bSpan) return { mode: d.mode, min: bounds.min, max: bounds.max }
  let min = d.min
  if (min < bounds.min) min = bounds.min
  if (min + span > bounds.max) min = bounds.max - span
  return { mode: d.mode, min, max: min + span }
}

/**
 * 缩放:`factor < 1` 放大(域变窄),`anchor` 是**保持不动**的那个取值(通常是鼠标下的时刻)。
 *
 * 为什么要锚点:滚轮缩放时如果锚在最左端,用户想看的那个位置会跑出视野 ——
 * 而滚轮的全部意义就是"把这里放大"。
 */
export function zoomDomain(d: Domain, bounds: Domain, factor: number, anchor: number): Domain {
  const span = d.max - d.min
  const bSpan = bounds.max - bounds.min
  const minSpan = Math.max(bSpan / 500, d.mode === 'time' ? 1 : 0.01) // 最多放到 500 倍
  let next = span * factor
  if (next < minSpan) next = minSpan
  if (next > bSpan) next = bSpan
  const ratio = span > 0 ? (anchor - d.min) / span : 0.5
  const min = anchor - ratio * next
  return clampDomain({ mode: d.mode, min, max: min + next }, bounds)
}

/** 平移:`frac` 是"移动了可视宽度的几分之几"(正 = 向未来/向后)。 */
export function panDomain(d: Domain, bounds: Domain, frac: number): Domain {
  const span = d.max - d.min
  const delta = span * frac
  return clampDomain({ mode: d.mode, min: d.min + delta, max: d.max + delta }, bounds)
}

/** 时钟(HH:MM:SS.mmm)。绝对时钟是默认口径:它能直接与服务器日志、`timeline.log` 对时。 */
export function fmtClock(t: number): string {
  if (!(t > 0)) return '—'
  const d = new Date(t)
  const p2 = (n: number) => String(n).padStart(2, '0')
  const p3 = (n: number) => String(n).padStart(3, '0')
  return `${p2(d.getHours())}:${p2(d.getMinutes())}:${p2(d.getSeconds())}.${p3(d.getMilliseconds())}`
}

/** 相对时刻(T+1.234s)。要给 `t0`(这一份的第一帧),没有就退回绝对时钟。 */
export function fmtRel(t: number, t0: number): string {
  if (!(t > 0)) return '—'
  if (!(t0 > 0)) return fmtClock(t)
  const s = (t - t0) / 1000
  return (s >= 0 ? 'T+' : 'T') + s.toFixed(3) + 's'
}

/** 耗时(毫秒 → 人读)。**0 与负数一律不画**:没有回包不是"0 毫秒"(见 D7 的纪律)。 */
export function fmtDur(ms: number): string {
  if (!(ms > 0)) return '—'
  if (ms < 1000) return `${Math.round(ms)} ms`
  if (ms < 60_000) return `${(ms / 1000).toFixed(2)} s`
  const m = Math.floor(ms / 60_000)
  return `${m} 分 ${((ms - m * 60_000) / 1000).toFixed(1)} 秒`
}

/** 大约多少字节(与列表里那份同一口径:1024 进位、先算档再取整)。 */
export function fmtBytes(n: number): string {
  if (!n) return '0 B'
  if (n >= 1 << 20) return (n / (1 << 20)).toFixed(1) + ' MB'
  if (n >= 1 << 10) return (n / (1 << 10)).toFixed(1) + ' KB'
  return n + ' B'
}

/** 一条时间尺刻度。 */
export interface Tick {
  v: number
  x: number
  label: string
}

/**
 * 刻度:挑一个"好看"的步长,再按域铺开。
 *
 * `t0` 只影响标签形态(相对/绝对),不影响位置 —— 位置一律由 `makeScale` 给,免得两处不一致。
 */
export function ticks(d: Domain, width: number, t0: number, px = 90): Tick[] {
  const span = d.max - d.min
  if (!(width > 0) || !(span > 0)) return []
  const want = Math.max(2, Math.floor(width / px))
  const raw = span / want
  const nice = niceStep(raw)
  const first = Math.ceil(d.min / nice) * nice
  const x = makeScale(d, width)
  const out: Tick[] = []
  for (let v = first; v <= d.max && out.length < 40; v += nice) {
    out.push({ v, x: x(v), label: d.mode === 'time' ? (t0 > 0 ? fmtRel(v, t0) : fmtClock(v)) : fmtBytes(v) })
  }
  return out
}

/** 把任意步长收成 1/2/5 × 10^k。 */
function niceStep(raw: number): number {
  const mag = Math.pow(10, Math.floor(Math.log10(raw)))
  const n = raw / mag
  const mult = n <= 1 ? 1 : n <= 2 ? 2 : n <= 5 ? 5 : 10
  return mult * mag
}
