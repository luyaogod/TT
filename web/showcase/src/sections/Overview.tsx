import { AlertTriangle, Boxes, Palette, Ruler, Sparkles } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { RiskBadge } from '@/components/Block'
import { CATALOG, ROADMAP, type Risk } from '@/lib/catalog'
import { riskCounts } from '@/lib/ledger'

const DEFAULTS = [
  { icon: Palette, label: '风格与基色', value: 'new-york · baseColor=neutral', hint: '官方 registry 的默认组合;token 逐字取自 /r/colors/neutral.json' },
  { icon: Ruler, label: '圆角', value: '--radius: 0.625rem', hint: '产品现在是"token 清零 + 按钮硬写 6px"两套口径并存' },
  { icon: Boxes, label: '组件来源', value: '36 个组件来自官方 registry', hint: 'npx shadcn@latest add …(new-york-v4 通道,统一 radix-ui + 官方 cn 包)' },
  { icon: Sparkles, label: '明暗主题', value: 'next-themes · class 策略', hint: "storageKey='tt.theme'、值域 light/dark/system —— 与产品同一份偏好,可无痛切换" },
]

const DECISIONS = [
  {
    title: '决策点 1:是否一起换依赖底座',
    body: '官方 v4 registry 的组件从统一包 radix-ui 引入、并从官方 cn 包引入 cn(它自带 clsx + tailwind-merge 的能力)。产品现在用的是逐个 @radix-ui/react-* 包 + 自写 web/shared/utils.ts。两者都能跑,但混用会出现"同一个 Radix 装两份"。建议整站切到统一包,一次到位。',
  },
  {
    title: '决策点 2:默认 token 里没有"成功/警告/信息"',
    body: 'shadcn 默认只给 default/secondary/destructive/outline 四档,而产品现在用了 emerald(成功)、amber(警告/强调)、sky(信息/链接)、red(危险)五种语义。要么收窄语义(状态栏会变成黑条/白条),要么显式补 --success / --warning / --info 三个 token 并与状态栏、Badge、Alert 共用。这是本次改造最需要拍板的一条。',
  },
  {
    title: '决策点 3:编辑器式密度是否保留',
    body: '产品是"面板紧贴 + 1px 分割线 + 零圆角"的编辑器密度(输入框 h-7/h-8、表格行 h-7)。官方默认是 h-9/h-8、行高更松、圆角 0.625rem。演示台里每个块都能直接看出差别 —— 逐块选"保留紧凑覆盖"还是"接受官方默认"。',
  },
]

export function Overview({ onJump }: { onJump: (sectionId: string, blockId?: string) => void }) {
  const counts = riskCounts()
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">这一屏在决策什么</CardTitle>
          <CardDescription className="leading-relaxed">
            下面是现有前端(调试工作台 / 接口日志 / 服务测试 / 统一设置页)的<b>全部功能点</b>,每个都用
            <b>纯官方 shadcn 组件 + shadcn 默认明暗双色主题</b>重建了一遍,并标注了现状落点、目标组件、风险级别与迁移要点。
            看完每一屏就能回答:"这个功能点换过去长什么样、会不会踩坑、要不要保留紧凑密度"。
          </CardDescription>
          <CardAction>
            <Badge variant="secondary">{counts.total} 个功能点</Badge>
          </CardAction>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {DEFAULTS.map((d) => (
              <div key={d.label} className="rounded-lg border p-3">
                <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                  <d.icon className="size-3.5" />
                  {d.label}
                </div>
                <div className="mt-1 font-mono text-xs">{d.value}</div>
                <div className="mt-1.5 text-[11px] leading-4 text-muted-foreground">{d.hint}</div>
              </div>
            ))}
          </div>
          <div className="mt-4 flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
            风险分布:
            <Badge variant="outline">低 {(counts as Record<Risk | 'total', number>).low}</Badge>
            <Badge variant="secondary">中 {(counts as Record<Risk | 'total', number>).mid}</Badge>
            <Badge variant="destructive">高 {(counts as Record<Risk | 'total', number>).high}</Badge>
            <Separator orientation="vertical" className="h-4" />
            风险口径:低=纯展示/纯受控表单;中=受控状态或语义色映射;高=滚动祖先、键盘流、Monaco、异步确认、性能上限。
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-4 lg:grid-cols-3">
        {DECISIONS.map((d) => (
          <Card key={d.title}>
            <CardHeader>
              <CardTitle className="flex items-center gap-1.5 text-sm">
                <AlertTriangle className="size-4 text-amber-600 dark:text-amber-400" />
                {d.title}
              </CardTitle>
              <CardDescription className="text-[11px] leading-4">{d.body}</CardDescription>
            </CardHeader>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">建议推进顺序</CardTitle>
          <CardDescription>先做"所有后续工作都依赖"的那一层;把硬约束留在最后,逐块验证。</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {ROADMAP.map((r, i) => (
              <div key={r.phase} className="flex gap-3">
                <div className="flex size-6 shrink-0 items-center justify-center rounded-full bg-secondary text-[11px] font-medium">{i + 1}</div>
                <div className="min-w-0">
                  <div className="text-xs font-medium">{r.phase} · {r.scope}</div>
                  <div className="mt-0.5 text-[11px] leading-4 text-muted-foreground">{r.why}</div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">迁移台账(全部功能点)</CardTitle>
          <CardDescription>点任意一行跳到那一屏的对应块。风险级别决定它排在第几步做。</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {CATALOG.map((section) => (
            <div key={section.id} className="space-y-2">
              <div className="flex flex-wrap items-baseline gap-2">
                <span className="text-sm font-medium">{section.title}</span>
                <span className="text-[11px] text-muted-foreground">{section.question}</span>
              </div>
              <div className="rounded-lg border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-56 text-xs">功能点</TableHead>
                      <TableHead className="text-xs">现状落点</TableHead>
                      <TableHead className="w-56 text-xs">目标组件</TableHead>
                      <TableHead className="w-24 text-xs">风险</TableHead>
                      <TableHead className="w-20 text-xs" />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {section.blocks.map((b) => (
                      <TableRow key={b.id} className="align-top">
                        <TableCell className="whitespace-normal text-xs">{b.title}</TableCell>
                        <TableCell className="whitespace-normal font-mono text-[11px] text-muted-foreground">{b.source}</TableCell>
                        <TableCell className="whitespace-normal">
                          <span className="flex flex-wrap gap-1">
                            {b.target.length === 0
                              ? <Badge variant="outline">保持手写</Badge>
                              : b.target.map((t) => <Badge key={t} variant="outline" className="font-mono">{t}</Badge>)}
                          </span>
                        </TableCell>
                        <TableCell><RiskBadge risk={b.risk} /></TableCell>
                        <TableCell>
                          <Button variant="ghost" size="xs" onClick={() => onJump(section.id, b.id)}>去看</Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
