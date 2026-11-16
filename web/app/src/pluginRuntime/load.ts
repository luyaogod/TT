// 加载插件的浏览器模块并调用它的 register(ctx)。
//
// 模块地址是 /plugins/<id>/ui/<entry>（TT 从活跃版本目录提供）。这里用
// /* @vite-ignore */ 的动态 import：插件代码**不进宿主 bundle**，运行时才取。

import * as React from 'react'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  Checkbox,
  Progress,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../../../shared/ui-radix'
import { registerSlot } from './registry'
import type { SlotRegistration } from './types'

/** 插件模块拿到的上下文。React 由宿主注入 —— 保证单实例且不用 import map。 */
export interface PluginUIContext {
  pluginId: string
  version: string
  React: typeof React
  /** 注册一个槽位组件；页面卸载时宿主会清掉本插件的全部注册。 */
  slot: (slot: string, reg: SlotRegistration) => void
  /** 宿主提供的 shadcn/Radix 组件（与宿主共用同一个 React 实例）。 */
  ui: {
    Checkbox: typeof Checkbox
    Progress: typeof Progress
    /** shadcn Select 的五个部件：Root/Trigger/Value/Content/Item。 */
    Select: {
      Root: typeof Select
      Trigger: typeof SelectTrigger
      Value: typeof SelectValue
      Content: typeof SelectContent
      Item: typeof SelectItem
    }
    /** shadcn AlertDialog：确认弹窗用，别再让插件弹原生 window.confirm。 */
    AlertDialog: {
      Root: typeof AlertDialog
      Content: typeof AlertDialogContent
      Header: typeof AlertDialogHeader
      Footer: typeof AlertDialogFooter
      Title: typeof AlertDialogTitle
      Description: typeof AlertDialogDescription
      Action: typeof AlertDialogAction
      Cancel: typeof AlertDialogCancel
    }
  }
  log: (...args: unknown[]) => void
}

/** 插件模块必须导出的形状。 */
export interface PluginClientModule {
  register?: (ctx: PluginUIContext) => void | (() => void)
}

/** 加载并校验插件模块（不调用 register）。 */
export async function loadPluginClient(url: string): Promise<PluginClientModule> {
  const mod = (await import(/* @vite-ignore */ url)) as PluginClientModule
  if (!mod || typeof mod.register !== 'function') {
    throw new Error('插件模块缺少 register(ctx) 导出')
  }
  return mod
}

/** 调用 register，并返回它给出的清理函数。 */
export function activatePlugin(pluginId: string, version: string, mod: PluginClientModule): () => void {
  const context: PluginUIContext = {
    pluginId,
    version,
    React,
    slot: (slot, reg) => { registerSlot(pluginId, slot, reg) },
    ui: {
      Checkbox,
      Progress,
      Select: {
        Root: Select,
        Trigger: SelectTrigger,
        Value: SelectValue,
        Content: SelectContent,
        Item: SelectItem,
      },
      AlertDialog: {
        Root: AlertDialog,
        Content: AlertDialogContent,
        Header: AlertDialogHeader,
        Footer: AlertDialogFooter,
        Title: AlertDialogTitle,
        Description: AlertDialogDescription,
        Action: AlertDialogAction,
        Cancel: AlertDialogCancel,
      },
    },
    log: (...args: unknown[]) => { console.log(`[plugin:${pluginId}]`, ...args) },
  }
  const ret = mod.register?.(context)
  return typeof ret === 'function' ? ret : () => {}
}
