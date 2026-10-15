// SSR 渲染入口:把每一屏用真实组件树渲染成字符串。
// 这不是产品的 SSR 能力(演示台是纯客户端应用),而是"渲染检查"的载体:
// 类型检查看不出运行时会炸的东西 —— Radix 需要 Provider、组件被误用、导出名不对等等。
import { renderToString } from 'react-dom/server'
import type { ReactNode } from 'react'
import { ThemeProvider } from 'next-themes'
import { TooltipProvider } from '@/components/ui/tooltip'
import { Overview } from '@/sections/Overview'
import { FoundationSection } from '@/sections/Foundation'
import { WorkbenchSection } from '@/sections/Workbench'
import { InterfacesSection } from '@/sections/Interfaces'
import { SettingsSection } from '@/sections/Settings'
import { PatternsSection } from '@/sections/Patterns'
import { VirtualApp } from '@/virtual/VirtualApp'

const wrap = (node: ReactNode) =>
  renderToString(
    <ThemeProvider attribute="class" defaultTheme="dark" enableSystem storageKey="tt.theme">
      <TooltipProvider>{node}</TooltipProvider>
    </ThemeProvider>,
  )

import { CATALOG } from '@/lib/catalog'

/** 台账里每个功能点的锚点 id,按屏分组 —— 渲染检查据此断言"每一块都在页面上有锚点" */
export const BLOCK_IDS: Record<string, string[]> = Object.fromEntries(
  CATALOG.map((s) => [s.id, s.blocks.map((b) => b.id)]),
)

export function renderAll(): Record<string, string> {
  return {
    overview: wrap(<Overview onJump={() => {}} />),
    foundation: wrap(<FoundationSection />),
    workbench: wrap(<WorkbenchSection />),
    interfaces: wrap(<InterfacesSection />),
    settings: wrap(<SettingsSection />),
    patterns: wrap(<PatternsSection />),
    // 目标态虚拟应用:四个视图各渲染一遍(虚拟应用内部按 activeView 只挂载一个)
    'virtual-debug': wrap(<VirtualApp initialView="debug" />),
    'virtual-logs': wrap(<VirtualApp initialView="logs" />),
    'virtual-test': wrap(<VirtualApp initialView="test" />),
    'virtual-settings': wrap(<VirtualApp initialView="settings" />),
    'virtual-settings-dict': wrap(<VirtualApp initialView="settings" initialSettingsSection="data-dict" />),
    'virtual-settings-app': wrap(<VirtualApp initialView="settings" initialSettingsSection="app" />),
  }
}
