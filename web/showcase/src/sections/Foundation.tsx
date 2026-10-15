import { useState } from 'react'
import { ArrowDownToDot, ArrowUpFromDot, CalendarDays, Check, Eye, Play, RedoDot, RefreshCw, RotateCcw, Square, StepForward } from 'lucide-react'
import { Block, Demo } from '@/components/Block'
import { blocksOf } from '@/lib/ledger'
import { ACCOUNTS, LOG_ROWS } from '@/lib/fixtures'

import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Switch } from '@/components/ui/switch'
import { Progress } from '@/components/ui/progress'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription,
  AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group'
import { cn } from 'cn'

const B = blocksOf('foundation')

export function FoundationSection() {
  return (
    <div className="space-y-6">
      <Buttons />
      <Badges />
      <InputsAndSelects />
      <Toggles />
      <Switches />
      <Tooltips />
      <ProgressBar />
      <Dialogs />
      <Alerts />
      <TabsDemo />
      <ScrollDemo />
    </div>
  )
}

function Buttons() {
  return (
    <Block block={B['f-button']}>
      <div className="space-y-4">
        <Demo label="六种变体" hint="官方默认高度 h-9、圆角跟 --radius(0.625rem)">
          <div className="flex flex-wrap items-center gap-2">
            <Button>保存</Button>
            <Button variant="secondary">取消</Button>
            <Button variant="outline">测试连接</Button>
            <Button variant="ghost">展开</Button>
            <Button variant="destructive">删除</Button>
            <Button variant="link">查看文档</Button>
          </div>
        </Demo>
        <Demo label="尺寸" hint="xs / sm / default / lg / icon —— 产品里最常用的是 h-7 这一档">
          <div className="flex flex-wrap items-center gap-2">
            <Button size="xs">xs 保存</Button>
            <Button size="sm">sm 保存</Button>
            <Button>default 保存</Button>
            <Button size="lg">lg 保存</Button>
            <Button size="icon" variant="outline" aria-label="刷新"><RefreshCw /></Button>
          </div>
        </Demo>
        <Demo label="状态" hint="禁用态在半透明之外还有 pointer-events-none,与产品现在的 disabled:opacity-30 口径不同">
          <div className="flex flex-wrap items-center gap-2">
            <Button disabled>启动中…</Button>
            <Button variant="outline" disabled>需停站后才能发送命令</Button>
            <Button variant="ghost" size="icon" disabled aria-label="复制"><Check /></Button>
          </div>
        </Demo>
      </div>
    </Block>
  )
}

function Badges() {
  return (
    <Block block={B['f-badge']}>
      <div className="space-y-4">
        <Demo label="方案 A:只用官方变体(默认口径,零额外 token)">
          <div className="flex flex-wrap items-center gap-2">
            <Badge>处理结果 000(成功)</Badge>
            <Badge variant="secondary">处理结果 空</Badge>
            <Badge variant="destructive">处理结果 100(失败)</Badge>
            <Badge variant="outline">已禁用</Badge>
          </div>
        </Demo>
        <Demo label="方案 B:保留四色语义(需要为 success/warning/info 额外定义 token)" hint="产品现在用 emerald/amber/red/sky 五种写死色表达状态">
          <div className="flex flex-wrap items-center gap-2">
            <Badge className="border-transparent bg-emerald-600/10 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-400">成功 000</Badge>
            <Badge className="border-transparent bg-amber-500/10 text-amber-700 dark:bg-amber-500/15 dark:text-amber-400">会话忙碌</Badge>
            <Badge className="border-transparent bg-red-600/10 text-red-700 dark:bg-red-500/15 dark:text-red-400">失败 100</Badge>
            <Badge className="border-transparent bg-sky-500/10 text-sky-700 dark:bg-sky-500/15 dark:text-sky-400">信息</Badge>
            <Badge variant="secondary">中性</Badge>
          </div>
        </Demo>
        <Demo label="真实落点" hint="状态码来自日志行;忙碌提示来自工具条">
          <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
            {LOG_ROWS.slice(0, 3).map((r, i) => (
              <span key={i} className="flex items-center gap-1.5">
                {r.code === '000'
                  ? <Badge className="border-transparent bg-emerald-600/10 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-400">000</Badge>
                  : r.code
                    ? <Badge variant="destructive">{r.code}</Badge>
                    : <Badge variant="secondary">-</Badge>}
                <span className="font-mono">{r.service}</span>
                {i < 2 && <Separator orientation="vertical" className="h-3" />}
              </span>
            ))}
          </div>
        </Demo>
      </div>
    </Block>
  )
}

function InputsAndSelects() {
  const [env, setEnv] = useState('36正式')
  const [dbType, setDbType] = useState('kingbase')
  return (
    <div className="space-y-6">
      <Block block={B['f-input']}>
        <div className="space-y-4">
          <Demo label="官方默认" hint="h-9 + focus ring;产品现在的输入框是 h-8 且查询框刻意压掉了 ring">
            <div className="grid max-w-2xl gap-3 sm:grid-cols-2">
              <div className="grid gap-1.5">
                <Label htmlFor="d-input">IP 主机</Label>
                <Input id="d-input" defaultValue="10.20.36.11" />
              </div>
              <div className="grid gap-1.5">
                <Label htmlFor="d-input2">端口</Label>
                <Input id="d-input2" defaultValue="22" disabled />
              </div>
            </div>
          </Demo>
          <Demo label="紧凑覆盖(查询工具条那一排)" hint="要保住 h-7 + 无聚焦环的搜索感,就得显式覆盖官方默认">
            <div className="flex flex-wrap items-center gap-2">
              <Input className="h-7 w-52 text-xs" placeholder="服务名称(支持 * ? 通配,回车生效)" />
              <Input className="h-7 w-36 text-xs" placeholder="作业编号(支持 * ? 通配)" />
              <Input className="h-7 w-28 text-xs" placeholder="服务端(如 T100)" />
              <Input className="h-7 w-28 text-xs" placeholder="发起端(如 OA)" />
              <Input className="h-7 w-28 text-xs" placeholder="处理结果(如 000)" />
            </div>
          </Demo>
          <Demo label="带图标/后缀的输入(官方 InputGroup)" hint="比手写 relative + 绝对定位图标更稳;产品现在没有这个基元">
            <div className="flex max-w-md flex-wrap gap-2">
              <InputGroup className="h-9">
                <InputGroupAddon><CalendarDays /></InputGroupAddon>
                <InputGroupInput defaultValue="2026-03-11" />
              </InputGroup>
            </div>
          </Demo>
        </div>
      </Block>

      <Block block={B['f-select']}>
        <div className="grid max-w-2xl gap-4 sm:grid-cols-2">
          <Demo label="环境选择(真实选项文案)">
            <Select value={env} onValueChange={setEnv}>
              <SelectTrigger className="w-full"><SelectValue placeholder="选择环境" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="36正式">36正式（默认）</SelectItem>
                <SelectItem value="35测试">35测试 · Oracle</SelectItem>
                <SelectItem value="31开发">31开发 · 未拉取</SelectItem>
              </SelectContent>
            </Select>
          </Demo>
          <Demo label="库类型(条件字段随之变化)">
            <Select value={dbType} onValueChange={setDbType}>
              <SelectTrigger className="w-full"><SelectValue placeholder="类型" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="oracle">Oracle</SelectItem>
                <SelectItem value="kingbase">金仓(KingbaseES)</SelectItem>
              </SelectContent>
            </Select>
          </Demo>
        </div>
      </Block>
    </div>
  )
}

function Toggles() {
  const [soap, setSoap] = useState(false)
  return (
    <Block block={B['f-checkbox']}>
      <div className="space-y-5">
        <Demo label="复选(官方 Checkbox)" hint="产品的设置页只有一处复选,用的是原生 input;日志页的两处已经走 Radix">
          <div className="flex items-center gap-2">
            <Checkbox id="d-onlyfail" />
            <Label htmlFor="d-onlyfail">仅失败</Label>
          </div>
        </Demo>
        <Demo label="真实落点(服务测试的 SOAP 勾选)">
          <div className="flex items-center gap-2">
            <Checkbox id="d-soap" checked={soap} onCheckedChange={(v) => setSoap(v === true)} />
            <Label htmlFor="d-soap">SOAP 报文(SOAPAction 空头)</Label>
          </div>
        </Demo>
      </div>
    </Block>
  )
}

function Switches() {
  const [roSql, setRoSql] = useState(true)
  const [persist, setPersist] = useState('default')
  return (
    <Block block={B['f-switch']}>
      <div className="space-y-5">
        <Demo label="开关(语义匹配布尔权限)" hint="官方 Switch 没有第三态,所以它只用于「开/关」">
          <div className="flex items-center gap-2">
            <Switch id="d-rosql" checked={roSql} onCheckedChange={setRoSql} />
            <Label htmlFor="d-rosql">允许 AI 执行只读 SQL(默认开启;关掉后调试端点的 SQL 查询直接 403)</Label>
          </div>
        </Demo>
        <Demo label="三态不能用开关:断点持久化" hint="跟随默认(开启)/开启/关闭 —— 这处应保留 Select">
          <div className="max-w-xs">
            <Select value={persist} onValueChange={setPersist}>
              <SelectTrigger className="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="default">跟随默认(开启)</SelectItem>
                <SelectItem value="on">开启</SelectItem>
                <SelectItem value="off">关闭</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </Demo>
      </div>
    </Block>
  )
}

function Tooltips() {
  return (
    <Block block={B['f-tooltip']}>
      <div className="space-y-5">
        <Demo label="普通元素" hint="官方 Tooltip 是 bg-foreground/text-background 的反色块">
          <div className="flex flex-wrap gap-2">
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="outline" size="sm">悬停:默认环境星标</Button>
              </TooltipTrigger>
              <TooltipContent>设为默认环境(会话未指定环境时用它)</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="outline" size="sm">
                  <RotateCcw /> 重启会话
                </Button>
              </TooltipTrigger>
              <TooltipContent>重启会话(断开重连同环境)</TooltipContent>
            </Tooltip>
          </div>
        </Demo>
        <Demo label="禁用元素" hint="原生 title 在 disabled 按钮上仍会显示;Radix Tooltip 必须包一层可接收事件的容器">
          <div className="flex flex-wrap items-center gap-4">
            <Tooltip>
              <TooltipTrigger asChild>
                <span className="inline-flex">
                  <Button size="sm" disabled>步过 (F10)</Button>
                </span>
              </TooltipTrigger>
              <TooltipContent>协作模式下由 AI 主导</TooltipContent>
            </Tooltip>
            <Button size="sm" disabled title="原生 title:禁用时仍会显示">对照:原生 title</Button>
          </div>
        </Demo>
      </div>
    </Block>
  )
}

function ProgressBar() {
  return (
    <Block block={B['f-progress']}>
      <div className="max-w-lg space-y-5">
        <Demo label="确定进度(镜像拉取 72%)">
          <div className="space-y-1.5">
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>下载并解压…</span><span>72%</span>
            </div>
            <Progress value={72} />
          </div>
        </Demo>
        <Demo label="不确定进度" hint="产品现在用 bar-indeterminate,但那个类在仓库里没有任何 CSS 定义,动画是失效的">
          <div className="space-y-2">
            <div className="text-xs text-muted-foreground">现状等价物(静止的 1/3 宽色条)</div>
            <div className="h-1.5 w-full overflow-hidden rounded-full bg-primary/20">
              <div className="h-full w-1/3 rounded-full bg-primary" />
            </div>
            <div className="text-xs text-muted-foreground">官方 Progress 没有不确定态:要么给 val(哪怕估算),要么用 Spinner 表达"在跑但不知多久"</div>
          </div>
        </Demo>
      </div>
    </Block>
  )
}

function Dialogs() {
  const [open, setOpen] = useState(false)
  return (
    <Block block={B['f-dialog']}>
      <div className="space-y-5">
        <Demo label="危险确认(官方 AlertDialog)" hint="已与产品现有的删除环境确认同构">
          <div className="flex flex-wrap gap-2">
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="destructive" size="sm">删除该环境</Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>删除环境「36正式」?</AlertDialogTitle>
                  <AlertDialogDescription>
                    该环境的 SSH 连接、登录区域、TOPENT 与数据库账号配置将一并移除。点击「删除」后立即写入 config.json,不可撤销。
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>取消</AlertDialogCancel>
                  <AlertDialogAction>删除</AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
            <Button variant="outline" size="sm" onClick={() => setOpen(true)}>打开普通 Dialog(全屏替换 / 字典同步那种要表单的场合)</Button>
          </div>
        </Demo>
        <Demo label="为什么不能顺手把 window.confirm 换成 AlertDialog" hint="三处 confirm 都要同步布尔返回值">
          <pre className="overflow-x-auto rounded-md bg-muted p-3 font-mono text-[11px] leading-5 text-muted-foreground">{`// 现状(同步,立刻拿到答案)
if (!window.confirm('要现在升级 tt 吗？…')) return

// 换 Radix 之后必须变成异步:二选一
await confirmDialog({ title: '要现在升级 tt 吗？', … })   // 自己维护一个 Promise 化的确认器
// 或者
setPending({ kind: 'update' }); return                       // 状态机:确认回调里再继续`}</pre>
        </Demo>
      </div>
      <AlertDialog open={open} onOpenChange={setOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>全量重建环境「36正式」?</AlertDialogTitle>
            <AlertDialogDescription>
              将整目录替换本地镜像(删除服务器上已不存在的残留),首次或按需执行,耗时较长。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction>继续</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Block>
  )
}

function Alerts() {
  return (
    <Block block={B['f-alert']}>
      <div className="space-y-4">
        <Demo label="官方两态">
          <div className="space-y-2">
            <Alert>
              <AlertTitle>已加入用户 PATH</AlertTitle>
              <AlertDescription>已把 D:\Programs\TT 加入用户 PATH,新开的终端即可直接敲 tt</AlertDescription>
            </Alert>
            <Alert variant="destructive">
              <AlertTitle>连接失败</AlertTitle>
              <AlertDescription>连接失败:ORA-12541: TNS:no listener</AlertDescription>
            </Alert>
          </div>
        </Demo>
        <Demo label="四色语义(现状的四种结果条收敛成一个组件)" hint="成功/失败/信息/警告是产品里出现频率最高的四种反馈">
          <div className="space-y-2">
            <Alert className="border-emerald-600/30 text-emerald-700 dark:text-emerald-400">
              <AlertTitle>连接正常 ✓</AlertTitle>
              <AlertDescription className="text-emerald-700/80 dark:text-emerald-400/80">KingbaseES V008R006C008B0014</AlertDescription>
            </Alert>
            <Alert className="border-amber-600/30 text-amber-700 dark:text-amber-400">
              <AlertTitle>有未保存的修改</AlertTitle>
              <AlertDescription className="text-amber-700/80 dark:text-amber-400/80">「调试参数」与「默认环境」改的是同一个配置节,任一张点保存都会提交整节。</AlertDescription>
            </Alert>
            <Alert className="border-sky-600/30 text-sky-700 dark:text-sky-400">
              <AlertTitle>获取失败: 先填 SSH 主机与账号</AlertTitle>
              <AlertDescription className="text-sky-700/80 dark:text-sky-400/80">按区域登录服务器,解析库地址/service 并回填。</AlertDescription>
            </Alert>
          </div>
        </Demo>
      </div>
    </Block>
  )
}

function TabsDemo() {
  return (
    <Block block={B['f-tabs']}>
      <div className="space-y-5">
        <Demo label="默认变体(pill) —— 设置页的 SSH 服务器 / 数据库">
          <Tabs defaultValue="ssh" className="max-w-lg">
            <TabsList>
              <TabsTrigger value="ssh">SSH 服务器</TabsTrigger>
              <TabsTrigger value="db">数据库</TabsTrigger>
            </TabsList>
            <TabsContent value="ssh" className="rounded-lg border p-3 text-xs text-muted-foreground">
              环境名称 / IP 主机 / 端口 / 登录区域 / 账号 / 密码 / TOPENT / 工作区目录
            </TabsContent>
            <TabsContent value="db" className="rounded-lg border p-3 text-xs text-muted-foreground">
              类型 / 主机地址 / 端口 / 库名 —— 未挂载时显示「该环境尚未挂载数据库连接」
            </TabsContent>
          </Tabs>
        </Demo>
        <Demo label="line 变体 —— 面板页签(原始协议流 / 操作时间线)与日志详情">
          <Tabs defaultValue="raw" className="max-w-lg">
            <TabsList variant="line">
              <TabsTrigger value="raw">原始协议流 (6)</TabsTrigger>
              <TabsTrigger value="timeline">操作时间线</TabsTrigger>
              <TabsTrigger value="detail">基本信息</TabsTrigger>
            </TabsList>
            <TabsContent value="raw" className="font-mono text-[11px] text-muted-foreground">(fgldb) breakpoint 1 at bsft001_wf.4gl:412</TabsContent>
            <TabsContent value="timeline" className="text-xs text-muted-foreground">14:02:44 停站 bsft001_wf.4gl:412 (断点 #1)</TabsContent>
            <TabsContent value="detail" className="text-xs text-muted-foreground">服务名称 icd.erp.wo.get</TabsContent>
          </Tabs>
        </Demo>
        <Demo label="页签里塞状态标记(日志详情的「入参已修改」)" hint="官方 TabsTrigger 是按钮,里面可以放任意内容">
          <Tabs defaultValue="req" className="max-w-lg">
            <TabsList variant="line">
              <TabsTrigger value="info">基本信息</TabsTrigger>
              <TabsTrigger value="req">
                Request <span className="text-amber-600 dark:text-amber-400" title="入参已修改,重放会用它">•</span>
              </TabsTrigger>
              <TabsTrigger value="rsp">Response</TabsTrigger>
            </TabsList>
            <TabsContent value="req" className="text-xs text-muted-foreground">报文编辑器(可编辑)</TabsContent>
          </Tabs>
        </Demo>
        <Demo label="注意事项" hint="换 Tabs 必须确认的三件事">
          <ul className="list-disc space-y-1 pl-5 text-xs text-muted-foreground">
            <li>设置页切页签<b>不卸载状态</b>(showPwd / addAcct / accProbe 靠这一点存活)—— Radix Tabs 默认卸载非活动内容,需要 <code className="font-mono">forceMount</code>。</li>
            <li>源码编辑器页签栏<b>不能</b>用 Tabs:右端两个面板收展按钮必须固定在滚动区之外,且活动页签靠 inset 主色细线表达。</li>
            <li>时间线页签现在的下划线是"压住分割线"的画法,line 变体的下划线在列表内部,视觉位置会差 1px。</li>
          </ul>
        </Demo>
      </div>
    </Block>
  )
}

function ScrollDemo() {
  const items = ACCOUNTS.map((a) => a.schema).concat(['topprd_read', 'dsprd_read', 'topprd_batch', 'dsprd_batch', 'iface', 'report'])
  return (
    <Block block={B['f-scroll']}>
      <div className="grid max-w-3xl gap-5 sm:grid-cols-2">
        <Demo label="ScrollArea" hint="官方滚动条样式,与产品现在全局改 webkit 滚动条的做派不同">
          <ScrollArea className="h-32 rounded-lg border">
            <div className="space-y-1 p-3 text-xs">
              {items.map((s) => <div key={s} className="font-mono text-muted-foreground">{s}</div>)}
            </div>
          </ScrollArea>
        </Demo>
        <Demo label="原生 overflow-auto + sticky 表头" hint="ScrollArea 里 sticky 会失效:它多了一层 overflow:hidden 的视口">
          <div className="h-32 overflow-auto rounded-lg border">
            <table className="w-full border-separate border-spacing-0 text-xs">
              <thead>
                <tr>
                  <th className="sticky top-0 z-10 border-b bg-background px-2 py-1.5 text-left font-medium">账号(schema)</th>
                  <th className="sticky top-0 z-10 border-b bg-background px-2 py-1.5 text-left font-medium">类型</th>
                </tr>
              </thead>
              <tbody>
                {items.map((s, i) => (
                  <tr key={s} className="odd:bg-muted/30">
                    <td className="border-b px-2 py-1 font-mono">{s}</td>
                    <td className="border-b px-2 py-1 text-muted-foreground">{i % 2 === 0 ? 'Oracle' : '金仓'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Demo>
      </div>
      <div className={cn('mt-3 text-xs text-muted-foreground')}>
        结论:<b>列表类滚动保持原生 overflow-auto</b>,ScrollArea 只用在无 sticky 的普通面板上。
      </div>
    </Block>
  )
}

/** 基础件那一节里,图标按钮的对照(产品里有大量"无边框图标按钮"落点) */
export function IconButtonRow() {
  return (
    <div className="flex items-center gap-1">
      {[Eye, Play, StepForward, RedoDot, ArrowDownToDot, ArrowUpFromDot].map((Icon, i) => (
        <Button key={i} variant="ghost" size="icon" aria-label="debug">
          <Icon />
        </Button>
      ))}
      <Separator orientation="vertical" className="mx-1 h-5" />
      <Button variant="ghost" size="icon" aria-label="刷新"><RefreshCw /></Button>
      <Button variant="ghost" size="icon" className="text-red-600 dark:text-red-400" aria-label="结束"><Square /></Button>
    </div>
  )
}
