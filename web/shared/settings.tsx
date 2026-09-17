// 设置页的布局件(Card / SettingRow / Field / SectionTitle)。
//
// 两套页面共用:调试工作台用它搭统一设置页,字典页用它搭剩下的两个操作视图
// (只读的路径行 + 区块标题)。合并前两边各有一套自己的"像卡片的 <section> + 标题"
// 写法,样式靠各写一遍的 Tailwind 类维持一致。
import * as React from 'react'
import { cn } from './utils'

/**
 * 设置卡片:标题行(可带说明与右侧操作位) + 内容区。
 *
 * 刻意不用 Panel —— Panel 是 flex 列 + 自带 overflow-auto 的"面板",要撑满父容器;
 * 设置卡片是内容流里的静态盒子,滚动归外层的设置页内容列管。两者混用会套出嵌套滚动容器。
 */
export function Card({ id, title, description, right, children, className }: {
  /** 锚点 id:左树点击时按它滚动定位,深链接也用它(如 debug.launchArgs) */
  id?: string
  title: React.ReactNode
  description?: React.ReactNode
  /** 右侧操作位:通常是这张卡的「保存」按钮 */
  right?: React.ReactNode
  children: React.ReactNode
  className?: string
}) {
  return (
    <section id={id} className={cn('scroll-mt-3 border border-border bg-card', className)}>
      <div className="flex items-start justify-between gap-3 border-b border-border px-3 py-2">
        <div className="min-w-0">
          <h3 className="text-xs font-medium text-foreground">{title}</h3>
          {description && (
            <p className="mt-0.5 text-[11px] leading-4 text-muted-foreground">{description}</p>
          )}
        </div>
        {right && <div className="flex shrink-0 items-center gap-1.5">{right}</div>}
      </div>
      <div className="p-3">{children}</div>
    </section>
  )
}

/**
 * 一条设置:左侧标签+说明,右侧控件(VS Code 设置页的行式布局)。
 * 窄屏下自动改为上下堆叠。
 */
export function SettingRow({ id, label, description, control, className }: {
  /** 锚点 id,与 Card 的 id 一起构成深链接可定位的粒度 */
  id?: string
  label: React.ReactNode
  description?: React.ReactNode
  control: React.ReactNode
  className?: string
}) {
  return (
    <div
      id={id}
      className={cn(
        'flex scroll-mt-3 flex-col gap-1.5 py-2.5 first:pt-0 last:pb-0',
        'sm:flex-row sm:items-start sm:justify-between sm:gap-6',
        className,
      )}
    >
      <div className="min-w-0 sm:max-w-[58%]">
        <div className="text-xs text-foreground">{label}</div>
        {description && (
          <div className="mt-0.5 text-[11px] leading-4 text-muted-foreground">{description}</div>
        )}
      </div>
      <div className="w-full shrink-0 sm:w-72">{control}</div>
    </div>
  )
}

/** 表单字段:标签在上、控件在下(环境编辑器那种密集的两列栅格用) */
export function Field({ label, children, className }: {
  label: React.ReactNode; children: React.ReactNode; className?: string
}) {
  return (
    <label className={cn('block', className)}>
      <span className="mb-1 block text-[11px] text-muted-foreground">{label}</span>
      {children}
    </label>
  )
}

/** 区块小标题:大写短标签 + 一条细线。字典页的两个操作视图用它分区。 */
export function SectionTitle({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn(
      'mb-1.5 flex items-center gap-2 text-[11px] font-medium tracking-wide text-muted-foreground uppercase',
      className,
    )}>
      <span className="shrink-0">{children}</span>
      <span className="h-px min-w-0 flex-1 bg-border" />
    </div>
  )
}

/** 只读的信息行(运行信息、生效路径这类"看得见改不了"的值) */
export function InfoRow({ label, value, title }: {
  label: React.ReactNode; value: React.ReactNode; title?: string
}) {
  return (
    <div className="flex items-start gap-3 py-1.5 first:pt-0 last:pb-0">
      <span className="w-28 shrink-0 text-[11px] text-muted-foreground">{label}</span>
      <span className="min-w-0 flex-1 break-all font-mono text-[11px] text-foreground" title={title}>
        {value}
      </span>
    </div>
  )
}
