// 槽位注册表：插件模块 register() 时把组件挂到某个命名槽位，外壳按槽位取。
//
// 设计约束：
//   - 注册与注销都归插件所有：插件页卸载/切走时 clearPlugin 清掉它的全部注册；
//   - 同槽位同 id 后注册者覆盖（与后端 doctor 的告警口径一致）；
//   - 快照按版本号缓存，useSyncExternalStore 拿到的是稳定引用（否则无限重渲染）。

import { useSyncExternalStore } from 'react'
import type { RegisteredEntry, SlotRegistration } from './types'

const registry = new Map<string, RegisteredEntry[]>()
const listeners = new Set<() => void>()
let version = 0

// 快照缓存：同一版本、同一槽位返回同一个数组引用。
const snapshotCache = new Map<string, { version: number; list: RegisteredEntry[] }>()
const EMPTY: RegisteredEntry[] = []

function emit(): void {
  version++
  for (const l of listeners) l()
}

function sortEntries(list: RegisteredEntry[]): RegisteredEntry[] {
  return list.slice().sort((a, b) => (a.order ?? 0) - (b.order ?? 0) || a.pluginId.localeCompare(b.pluginId))
}

/** 注册一个槽位组件；返回注销函数。 */
export function registerSlot(pluginId: string, slot: string, reg: SlotRegistration): () => void {
  const list = registry.get(slot) ?? []
  const entry: RegisteredEntry = { ...reg, pluginId, slot }
  // 同 id 覆盖：保持"后注册者生效"，但旧条目要从数组里摘掉，避免渲染两份。
  const idx = reg.id ? list.findIndex((e) => e.pluginId === pluginId && e.id === reg.id) : -1
  if (idx >= 0) list[idx] = entry
  else list.push(entry)
  registry.set(slot, sortEntries(list))
  emit()
  return () => {
    const cur = registry.get(slot)
    if (!cur) return
    const next = cur.filter((e) => e !== entry)
    if (next.length) registry.set(slot, next)
    else registry.delete(slot)
    emit()
  }
}

/** 清掉一个插件的全部注册（页面卸载/插件禁用时调用）。 */
export function clearPlugin(pluginId: string): void {
  let changed = false
  for (const [slot, list] of registry) {
    const next = list.filter((e) => e.pluginId !== pluginId)
    if (next.length === list.length) continue
    changed = true
    if (next.length) registry.set(slot, next)
    else registry.delete(slot)
  }
  if (changed) emit()
}

/** 取一个槽位的当前条目（稳定引用，供 useSyncExternalStore 用）。 */
export function slotSnapshot(slot: string): RegisteredEntry[] {
  const cached = snapshotCache.get(slot)
  if (cached && cached.version === version) return cached.list
  const list = registry.get(slot)
  const out = list ? list.slice() : EMPTY
  snapshotCache.set(slot, { version, list: out })
  return out
}

function subscribe(listener: () => void): () => void {
  listeners.add(listener)
  return () => { listeners.delete(listener) }
}

/** React 侧读槽位。 */
export function useSlot(slot: string): RegisteredEntry[] {
  return useSyncExternalStore(subscribe, () => slotSnapshot(slot), () => EMPTY)
}

/** 仅供测试/调试：清空全部注册。 */
export function resetSlots(): void {
  registry.clear()
  snapshotCache.clear()
  emit()
}
