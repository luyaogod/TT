// 设置页的布局件(Card / SettingRow / Field / SectionTitle)。
//
// 统一设置页用它搭各张卡片(含只读的路径行 + 区块标题)。合并前调试工作台与字典页
// 两边各有一套自己的"像卡片的 <section> + 标题"写法,样式靠各写一遍的 Tailwind 类维持一致。
import * as React from 'react'
import { cn } from './utils'

/**
 * 设置卡片:标题(可带说明与右侧操作位) + 一串字段,整张卡是一个竖直序列。
 *
 * 版式对齐 VS Code 的设置页:标题在上、字段在下,**卡片内部不打任何分割线**。
 * 卡片自己就是分组边界,里面再横线切割只会把一屏切成许多小格子,越看越碎。
 * 所以整张卡只有一层内边距,标题与字段之间靠间距分,不靠线。
 *
 * 刻意不用 Panel —— Panel 是 flex 列 + 自带 overflow-auto 的"面板",要撑满父容器;
 * 设置卡片是内容流里的静态盒子,滚动归外层的设置页内容列管。两者混用会套出嵌套滚动容器。
 */
export function Card({ id, title, description, right, children, className, panel }: {
  /** 锚点 id:锚点跳转落到这张卡时用它定位,scroll-mt-3 保证落点不贴着面板上沿 */
  id?: string
  title: React.ReactNode
  description?: React.ReactNode
  /** 右侧操作位:通常是这张卡的「保存」按钮 */
  right?: React.ReactNode
  children: React.ReactNode
  className?: string
  /**
   * 当"定高面板"用:自带高度的卡片(站点管理那两栏)必须打开它,否则内容一多就顶出边框。
   * 打开后卡片是个可收缩的 flex 列,内容区 `flex-1 min-h-0 overflow-auto` —— 长了就在卡片内滚。
   * 左栏那种"列表滚、底部按钮钉住"的还要在内容里再分一层(flex-1 min-h-0 overflow-auto 的列表 +
   * 固定高度的按钮行);里面分好之后内容区自己就不会溢出了,滚动条只出现在列表上。
   * 不传就是普通的内容流盒子,跟着内容长高(默认)。
   */
  panel?: boolean
}) {
  return (
    <section id={id} className={cn('scroll-mt-3 border border-border bg-card p-3', panel && 'flex min-h-0 flex-col', className)}>
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 className="text-xs font-medium text-foreground">{title}</h3>
          {description && (
            <p className="mt-1 text-[11px] leading-4 text-muted-foreground">{description}</p>
          )}
        </div>
        {right && <div className="flex shrink-0 items-center gap-1.5">{right}</div>}
      </div>
      <div className={cn('mt-3 space-y-3', panel && 'flex min-h-0 flex-1 flex-col overflow-auto')}>{children}</div>
    </section>
  )
}

/**
 * 一条设置:标签在上、说明其次、控件在下(VS Code 设置页的版式)。
 *
 * 控件宽度统一收到 max-w-md。内容列有 700px 出头,让输入框撑满会被拉成一条细长的线,
 * 同一张卡里的下拉、按钮也会各走各的宽度。要整行宽的控件(进度块、环境清单那种)
 * 不走这里,直接做卡片的子元素。
 *
 * 竖直排列的另一半原因是说明文字:原先标签在左、控件在右,说明只能挤在窄窄的左半栏里折行,
 * 折出来的三四行会把控件和它的说明拉开很远。
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
    <div id={id} className={cn('scroll-mt-3', className)}>
      <div className="text-xs text-foreground">{label}</div>
      {description && (
        <div className="mt-1 text-[11px] leading-4 text-muted-foreground">{description}</div>
      )}
      <div className="mt-1.5 w-full max-w-md">{control}</div>
    </div>
  )
}

/** 表单字段:标签在上、控件在下。与 SettingRow 同一套字级与间距,只是不带说明行。 */
export function Field({ label, children, className }: {
  label: React.ReactNode; children: React.ReactNode; className?: string
}) {
  return (
    <label className={cn('block', className)}>
      <span className="block text-xs text-foreground">{label}</span>
      <span className="mt-1.5 block">{children}</span>
    </label>
  )
}

/** 区块小标题:大写短标签 + 一条细线。统一设置页的卡片用它分区。 */
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

/**
 * 只读的信息行(运行信息、生效路径这类"看得见改不了"的值)。
 * 竖直间距交给 Card 的 space-y —— 原先自己带 py 会与卡片给的间距叠成两层。
 */
export function InfoRow({ label, value, title }: {
  label: React.ReactNode; value: React.ReactNode; title?: string
}) {
  return (
    <div className="flex items-start gap-3">
      <span className="w-28 shrink-0 text-[11px] text-muted-foreground">{label}</span>
      <span className="min-w-0 flex-1 break-all font-mono text-[11px] text-foreground" title={title}>
        {value}
      </span>
    </div>
  )
}
