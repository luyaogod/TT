// 同页插件 UI 的运行时类型。宿主与插件模块共享这一份形状。

import type { ComponentType } from 'react'

/** 一次槽位注册。component 由插件模块提供，宿主只负责渲染与隔离。 */
export interface SlotRegistration {
  /** 槽位内的稳定标识（菜单项/设置分区用它区分；页面可省）。 */
  id?: string
  /** 页面路由（slot=page 时用于跟 #plugin/<route> 对齐）。 */
  route?: string
  title?: string
  /** 同槽位内排序，小的在前。 */
  order?: number
  component: ComponentType<any>
}

/** 注册表里的一条：带上是哪个插件注册的。 */
export interface RegisteredEntry extends SlotRegistration {
  pluginId: string
  slot: string
}
