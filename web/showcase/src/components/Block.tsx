import type { ReactNode } from 'react'
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { cn } from 'cn'
import { RISK_LABEL, type Block, type Risk } from '@/lib/catalog'

const RISK_VARIANT: Record<Risk, 'destructive' | 'secondary' | 'outline'> = {
  high: 'destructive',
  mid: 'secondary',
  low: 'outline',
}

/** 风险徽标:全站只有这一个渲染口径 */
export function RiskBadge({ risk }: { risk: Risk }) {
  return (
    <Badge variant={RISK_VARIANT[risk]} title={`迁移风险:${RISK_LABEL[risk]}`}>
      风险 {RISK_LABEL[risk]}
    </Badge>
  )
}

/**
 * 一个功能点的展示单元:上面是"这是什么/现状在哪/换成什么/要小心什么",下面是实时预览。
 *
 * 刻意把预览放在带边框的内层容器里:预览里出现的圆角、间距、投影都是 shadcn 默认值,
 * 不加内层边框就分不清"哪些留白是组件自带的、哪些是页面给的"。
 */
export function Block({ block, children, pad = true }: { block: Block; children: ReactNode; pad?: boolean }) {
  return (
    <Card id={block.id} className="scroll-mt-20">
      <CardHeader>
        <CardTitle className="flex flex-wrap items-center gap-2 text-base">
          {block.title}
          <RiskBadge risk={block.risk} />
        </CardTitle>
        <CardDescription className="leading-relaxed">{block.note}</CardDescription>
        <CardAction className="hidden shrink-0 flex-col items-end gap-1.5 sm:flex">
          <code className="rounded-md bg-muted px-2 py-1 font-mono text-[11px] text-muted-foreground">
            {block.source}
          </code>
          <span className="flex flex-wrap justify-end gap-1">
            {block.target.length === 0 ? (
              <Badge variant="outline">保持手写</Badge>
            ) : (
              block.target.map((t) => (
                <Badge key={t} variant="outline" className="font-mono">
                  {t}
                </Badge>
              ))
            )}
          </span>
        </CardAction>
      </CardHeader>
      <CardContent>
        <Separator className="mb-4" />
        <div
          className={cn(
            'rounded-lg border border-dashed border-border bg-background',
            pad && 'p-4',
            // 极窄屏下预览允许横向滚,不让内容把整页撑破
            'overflow-x-auto',
          )}
        >
          {children}
        </div>
      </CardContent>
    </Card>
  )
}

/** 预览里的小标题:说明"这一段在看什么" */
export function Demo({ label, children, hint }: { label: string; children: ReactNode; hint?: string }) {
  return (
    <div className="space-y-2">
      <div className="flex flex-wrap items-baseline gap-2">
        <span className="text-xs font-medium text-muted-foreground">{label}</span>
        {hint && <span className="text-[11px] text-muted-foreground/80">{hint}</span>}
      </div>
      {children}
    </div>
  )
}
