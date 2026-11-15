// 槽位渲染：把注册到某个槽位的组件画出来，并给每个组件套一层错误边界。
//
// 错误边界是"信任模型"之外唯一的技术兜底：插件组件抛错只黑掉自己的槽位，
// 不把整个外壳带下去。

import { Component, type ReactNode } from 'react'
import { useSlot } from './registry'
import type { RegisteredEntry } from './types'

interface BoundaryProps {
  label: string
  children: ReactNode
}

interface BoundaryState {
  error: Error | null
}

/** 单个槽位条目的错误边界。 */
class SlotBoundary extends Component<BoundaryProps, BoundaryState> {
  state: BoundaryState = { error: null }

  static getDerivedStateFromError(error: Error): BoundaryState {
    return { error }
  }

  render(): ReactNode {
    if (this.state.error) {
      return (
        <div role="alert" className="m-2 border border-red-500/30 bg-red-500/10 p-3 text-xs text-red-600 dark:text-red-400">
          <div className="flex items-center gap-2 font-medium">
            <span>插件组件渲染失败（{this.props.label}）</span>
          </div>
          <div className="mt-1 break-all opacity-80">{this.state.error.message}</div>
        </div>
      )
    }
    return this.props.children
  }
}

/** 渲染一个槽位的全部条目。filter 用于页面级过滤（只看本插件本路由）。 */
export function SlotHost({
  slot, filter,
}: {
  slot: string
  filter?: (entry: RegisteredEntry) => boolean
}) {
  const entries = useSlot(slot)
  const shown = filter ? entries.filter(filter) : entries
  return (
    <>
      {shown.map((entry) => {
        const Comp = entry.component
        return (
          <SlotBoundary key={`${entry.pluginId}:${entry.id ?? entry.route ?? slot}`} label={`${slot}:${entry.pluginId}`}>
            <Comp />
          </SlotBoundary>
        )
      })}
    </>
  )
}
/** 有内容才渲染（用于状态栏/导航这类"有就占位、没有就消失"的槽位）。 */
export function SlotHostIfAny({ slot }: { slot: string }) {
  const entries = useSlot(slot)
  if (entries.length === 0) return null
  return <SlotHost slot={slot} />
}

/** 设置页的分区槽位：每条注册渲染成一张带标题的卡片。 */
export function PluginSettingsSections() {
  const entries = useSlot('settings.section')
  if (entries.length === 0) return null
  return (
    <>
      {entries.map((entry) => {
        const Comp = entry.component
        return (
          <section key={`${entry.pluginId}:${entry.id ?? entry.title ?? ''}`}
            className="rounded-ui border border-border bg-card p-3">
            <div className="mb-2 text-[13px] font-medium text-foreground">{entry.title || entry.pluginId}</div>
            <SlotBoundary label={`settings.section:${entry.pluginId}`}>
              <Comp />
            </SlotBoundary>
          </section>
        )
      })}
    </>
  )
}