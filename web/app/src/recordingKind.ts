// 帧类别的**唯一**颜色与名字：条带上的块、表里的类型 tag 都从这里取。
//
// 为什么单独一个文件：这两处的颜色必须**逐字一致**（用户 2026-10-16 的要求是
// "表里的类型列与条带同样颜色"），各写一份迟早漂移 —— 那时同一帧在条带上是绿、
// 在表里是蓝，而它俩本来指的是同一件事。放这里之后，改颜色只有一处。
//
// 颜色语义（与协议里的类别一一对应）：
//
//	om     运行期推来的界面状态（绿）
//	un     运行期的属性更新（深绿）
//	rn     运行期的删节点（红）—— 三种 AUI 操作各一色，因为它们在排障时意义完全不同
//	meta   握手/元信息（紫）—— 整条流里唯一不带帧头的那几条
//	event  前端发出去的用户事件（蓝）
//	cache  图片缓存（灰）—— 不是消息，不解析也不重放，但占时间
const KIND_CLASS: Record<string, string> = {
  om: 'bg-emerald-500',
  un: 'bg-emerald-600',
  rn: 'bg-rose-500',
  meta: 'bg-violet-500',
  event: 'bg-sky-500',
  cache: 'bg-muted-foreground/40',
}
const OTHER_CLASS = 'bg-muted-foreground/60'

/** 帧类别 → Tailwind 背景色类（条带的块与表里的 tag 共用）。 */
export function kindClass(k: string): string {
  if (k === 'cache') return KIND_CLASS.cache
  if (k.startsWith('type')) return OTHER_CLASS // 没见过的类型：照编号说出来，别装成已知的
  return KIND_CLASS[k] || OTHER_CLASS
}

/** 帧类别 → 给人看的短名（中文只给"握手/缓存"这两个，其余保持协议里的原词）。 */
export function kindLabel(k: string): string {
  switch (k) {
    case 'meta': return '握手'
    case 'cache': return '缓存'
    default: return k
  }
}

/**
 * 帧类别 → 表里那个类型 tag 的完整类（底色与条带的块**同一个** `kindClass`）。
 *
 * 字色按底色分两档：有色的五种（emerald/rose/violet/sky）白字；灰的两种（缓存与没见过的
 * 类型）白字在浅灰底上根本看不清，所以用前景色 —— 这是唯一的例外，理由写在脸上。
 */
export function kindTagClass(k: string): string {
  const grey = k === 'cache' || k.startsWith('type')
  return `${kindClass(k)} ${grey ? 'text-foreground' : 'text-white'}`
}
