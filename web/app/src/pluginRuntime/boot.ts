// 插件槽位启动器：应用启动时把所有已启用插件的浏览器模块激活一次，
// 让 nav/statusbar/settings.section/overlay 这些**不依赖进入页面**的槽位也能注册。
//
// 与 PluginPage 的关系：页面不再各自加载模块，而是 ensure 这里；页面切走也不清注册
// （否则侧边栏/状态栏的插件内容会跟着消失）。

import { useSyncExternalStore } from 'react'
import { activatePlugin, loadPluginClient } from './load'
import { clearPlugin } from './registry'
import { pluginUIAssetURL, type PluginDescriptor } from '../plugins'

interface Activation {
  dispose: () => void
  error: string
  loading: boolean
}

const activations = new Map<string, Activation>()
const listeners = new Set<() => void>()

function emit(): void {
  for (const l of listeners) l()
}

function subscribe(l: () => void): () => void {
  listeners.add(l)
  return () => { listeners.delete(l) }
}

function entryURL(p: PluginDescriptor): string | null {
  // 同一个模块可服务多个槽位；取第一个有 entry 的贡献。
  const hit = p.ui.find((u) => u.entry)
  if (!hit) return null
  return pluginUIAssetURL(p.id, hit.entry)
}

/** 激活一个插件模块（幂等）。返回错误文本（空 = 成功/加载中）。 */
export async function ensurePluginActivated(p: PluginDescriptor): Promise<string> {
  const cur = activations.get(p.id)
  if (cur && !cur.error && !cur.loading) return ''
  if (cur && cur.loading) return ''
  activations.set(p.id, { dispose: () => {}, error: '', loading: true })
  emit()
  try {
    const url = entryURL(p)
    if (!url) throw new Error('插件没有可加载的前端模块')
    const mod = await loadPluginClient(url)
    const dispose = activatePlugin(p.id, p.version, mod)
    activations.set(p.id, { dispose, error: '', loading: false })
    emit()
    return ''
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    activations.set(p.id, { dispose: () => {}, error: msg, loading: false })
    emit()
    return msg
  }
}

/** 启动时对全部已启用插件激活一遍（失败只记录，不影响其它插件）。 */
export async function ensureAllPlugins(plugins: PluginDescriptor[]): Promise<void> {
  await Promise.all(plugins.filter((p) => p.enabled).map((p) => ensurePluginActivated(p)))
}

/** 重新加载一个插件（页面上的"重试"用）。 */
export async function reloadPlugin(p: PluginDescriptor): Promise<string> {
  const cur = activations.get(p.id)
  cur?.dispose()
  clearPlugin(p.id)
  activations.delete(p.id)
  return ensurePluginActivated(p)
}

/** 某个插件的激活状态（页面用来显示加载/错误）。 */
export function pluginActivation(id: string): Activation | undefined {
  return activations.get(id)
}

export function usePluginActivation(id: string): Activation | undefined {
  return useSyncExternalStore(subscribe, () => activations.get(id), () => undefined)
}
