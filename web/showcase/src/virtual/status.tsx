// 语义状态件:官方 Badge 只有 default/secondary/destructive/outline 四档,而 TT 的信息设计有四类状态。
//
// 最佳实践是**不动 registry 组件**(升级时才能 --overwrite 重拉),在应用层加一层薄的语义包装:
// 颜色来自 index.css 里新补的 --success / --warning / --info 三个 token,与状态栏共用同一份来源。
import type { ReactNode } from 'react'
import { Badge } from '@/components/ui/badge'
import { cn } from 'cn'

export type Tone = 'success' | 'warning' | 'info' | 'danger' | 'neutral' | 'outline'

const TONE_CLASS: Record<Tone, string> = {
  success: 'border-transparent bg-success text-success-foreground',
  warning: 'border-transparent bg-warning text-warning-foreground',
  info: 'border-transparent bg-info text-info-foreground',
  danger: 'border-transparent bg-destructive text-white',
  neutral: 'border-transparent bg-secondary text-secondary-foreground',
  outline: '',
}

export function StatusBadge({ tone, children, className, title }: {
  tone: Tone; children: ReactNode; className?: string; title?: string
}) {
  return (
    <Badge
      variant={tone === 'outline' ? 'outline' : 'default'}
      className={cn(TONE_CLASS[tone], className)}
      title={title}
    >
      {children}
    </Badge>
  )
}

/** 状态点:会话是否是当前连接、断点是否启用这类"有/无"的表达 */
export function StatusDot({ tone, className, title }: { tone: Tone; className?: string; title?: string }) {
  const bg: Record<Tone, string> = {
    success: 'bg-success',
    warning: 'bg-warning',
    info: 'bg-info',
    danger: 'bg-destructive',
    neutral: 'bg-muted-foreground/40',
    outline: 'bg-border',
  }
  return <span title={title} className={cn('inline-block size-2 shrink-0 rounded-full', bg[tone], className)} />
}

/** 语义化的内联状态条:比官方 Alert 轻,用于"每张卡就近显示的结果" */
export function StatusNote({ tone, children, className }: { tone: Tone; children: ReactNode; className?: string }) {
  const cls: Record<Tone, string> = {
    success: 'border-success/40 bg-success/5 text-success',
    warning: 'border-warning/50 bg-warning/10 text-warning',
    info: 'border-info/40 bg-info/5 text-info',
    danger: 'border-destructive/40 bg-destructive/5 text-destructive',
    neutral: 'border-border bg-muted/40 text-muted-foreground',
    outline: 'border-border text-muted-foreground',
  }
  return (
    <div className={cn('rounded-md border border-l-2 px-2.5 py-1.5 text-[11px] leading-4', cls[tone], className)}>
      {children}
    </div>
  )
}
