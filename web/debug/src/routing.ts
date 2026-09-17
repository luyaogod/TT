// 哈希路由:把「当前视图 / 设置页的哪个分区」映射到 URL 片段,让界面可深链接。
//
// 为什么需要它:设置页搬进调试工作台之后,字典页的「设置」入口要能直接落到
// 具体的分区(/debug/#settings/data-dict),而不是把用户丢在调试页让他自己找。
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

/** 活动栏的四个视图(设置页是其中之一) */
export type ViewKey = 'debug' | 'wslogs' | 'wstest' | 'settings'

export function isSettingsSection(v: string): v is SettingsSectionKey {
  return SETTINGS_SECTIONS.some((s) => s.key === v)
}

function isViewKey(v: string): v is ViewKey {
  return v === 'debug' || v === 'wslogs' || v === 'wstest' || v === 'settings'
}

/** 解析 location.hash。认不出来的一律返回空对象(不猜,让调用方走默认)。 */
export function parseHash(hash: string): { view?: ViewKey; section?: SettingsSectionKey } {
  const raw = (hash || '').replace(/^#/, '').trim()
  if (!raw) return {}
  const [head, tail] = raw.split('/', 2)
  if (!isViewKey(head)) return {}
  if (head !== 'settings') return { view: head }
  // #settings 后面可以没有分区(落到默认的第一个)
  if (tail && isSettingsSection(tail)) return { view: 'settings', section: tail }
  return { view: 'settings' }
}

/** 设置页某个分区的 hash(带前导 #,可直接交给 history.replaceState) */
export function settingsHash(section: SettingsSectionKey): string {
  return `#settings/${section}`
}
