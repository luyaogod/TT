// 哈希路由:把「当前视图 / 设置页的哪个分区」映射到 URL 片段,让界面可深链接。
//
// 为什么需要它:设置页搬进调试工作台之后,要能直接落到具体的分区
// (/debug/#settings/data-dict),而不是把用户丢在调试页让他自己找。
//
// 约束(踩过):本文件**不得被 store.ts 导入**,且模块顶层不得访问 window/document/history。
// check:store 用裸 esbuild 打包 store.ts 并在 Node 里跑(桩 window/document),
// 顶层碰 DOM 会在测试启动前就抛错。把 DOM 访问全部留在函数体内、由 React effect 调用。

/** 设置页的四个分区 */
export type SettingsSectionKey = 'sites' | 'data-dict' | 'debug' | 'app'

/** 分区清单(键 + 显示名)。左树与深链接校验共用这一份,免得两处各列一遍。 */
export const SETTINGS_SECTIONS: { key: SettingsSectionKey; label: string }[] = [
  { key: 'sites', label: '站点管理' },
  { key: 'data-dict', label: '数据字典' },
  { key: 'debug', label: 'DEBUG' },
  { key: 'app', label: '应用设置' },
]

export const DEFAULT_SETTINGS_SECTION: SettingsSectionKey = 'sites'

/** 活动栏的五个视图(设置页是其中之一,录制面板是最后加的那个) */
export type ViewKey = 'debug' | 'wslogs' | 'wstest' | 'record' | 'settings'

export function isSettingsSection(v: string): v is SettingsSectionKey {
  return SETTINGS_SECTIONS.some((s) => s.key === v)
}

function isViewKey(v: string): v is ViewKey {
  return v === 'debug' || v === 'wslogs' || v === 'wstest' || v === 'record' || v === 'settings'
}

/**
 * 录制 ID 的合法形态:字母数字开头,后面允许字母数字与 `-` `_` `.`,最长 80。
 *
 * 为什么要白名单而不是"转义一下":这个值会进 URL 片段,而片段是**深链接**——用户会把它
 * 粘给别人。含 `/`、`..`、空格、中文的值只可能来自手改的地址栏或坏数据,那时**丢弃**
 * (回列表)比"猜他想看哪一份"安全:猜错了就会去打开另一个 ID 的文件。
 * 录制 ID 本身就是代理落盘的 tag(`fep-20261016-112410`),形态固定。
 */
export function isRecordID(v: string): boolean {
  return /^[A-Za-z0-9][A-Za-z0-9._-]{0,79}$/.test(v)
}

/** 录制视图的片段:列表是 `#record`,明细是 `#record/<录制 ID>`,定位到某一时刻再加 `?t=<Unix 毫秒>`。 */
export function recordHash(id: string, at?: number): string {
  if (!id) return '#record'
  return at && at > 0 ? `#record/${id}?t=${Math.round(at)}` : `#record/${id}`
}

/**
 * 解析 location.hash。认不出来的一律返回空对象(不猜,让调用方走默认)。
 *
 * `recordId` 只在**恰好两段**且第二段是合法 ID 时给出;多一段(`#record/a/b`)或形态不对
 * 一律当"就是列表" —— 不静默截断(旧写法 `split('/', 2)` 会把 `a` 当 ID 用)。
 * `?t=<ms>` 是"定位到某一时刻"(区间深链接):合法就给,不合法就当没有 —— 一个坏数字
 * 不该让整条深链接失效。
 */
export function parseHash(hash: string): { view?: ViewKey; section?: SettingsSectionKey; recordId?: string; at?: number } {
  const raw = (hash || '').replace(/^#/, '').trim()
  if (!raw) return {}
  const qi = raw.indexOf('?')
  const path = qi >= 0 ? raw.slice(0, qi) : raw
  const query = qi >= 0 ? raw.slice(qi + 1) : ''
  const parts = path.split('/')
  const head = parts[0]
  if (!isViewKey(head)) return {}
  if (head === 'record') {
    const at = parseAt(query)
    if (parts.length === 2 && isRecordID(parts[1])) {
      return at > 0 ? { view: 'record', recordId: parts[1], at } : { view: 'record', recordId: parts[1] }
    }
    if (parts.length > 2) return {}
    return { view: 'record' }
  }
  if (head !== 'settings') return { view: head }
  // #settings 后面可以没有分区(落到默认的第一个)
  const tail = parts[1]
  if (tail && isSettingsSection(tail)) return { view: 'settings', section: tail }
  return { view: 'settings' }
}

/** 从 `t=<Unix 毫秒>` 里取时刻(没有/非法/非正数一律 0 = 不定位)。 */
function parseAt(query: string): number {
  if (!query) return 0
  for (const kv of query.split('&')) {
    const [k, v] = kv.split('=')
    if (k !== 't') continue
    const n = Number(v)
    // 只认**合理的毫秒**：2001 年之后、2100 年之前。乱填的数字当没有 ——
    // 拿它去定位会得到一个空窗口，用户看到的是"页面打开了但什么都没有"。
    if (Number.isFinite(n) && n > 1e12 && n < 4e12) return Math.round(n)
  }
  return 0
}

/** 设置页某个分区的 hash(带前导 #,可直接交给 history.replaceState) */
export function settingsHash(section: SettingsSectionKey): string {
  return `#settings/${section}`
}

/**
 * 当前视图对应的 hash。设置页要带上分区(它自己分成四个分区),录制页的明细则带上录制 ID,
 * 别的视图就是视图名。
 *
 * 写片段这件事必须**双向**(读进来 + 写回去):只读不写的话,点图标切到别的视图时 URL
 * 还停在旧视图上,一刷新就把人弹回旧页(实测:在 `#debug` 里点「设置」,界面是设置页而
 * URL 仍是 `#debug`,刷新回调试页)。录制明细同理:打开了第 N 份却不在片段里,
 * 刷新就回列表、也没法把"我在看这一份"发给别人。
 */
export function viewHash(view: ViewKey, section: SettingsSectionKey, recordId?: string, at?: number): string {
  if (view === 'settings') return settingsHash(section)
  if (view === 'record') return recordHash(recordId || '', at)
  return `#${view}`
}
