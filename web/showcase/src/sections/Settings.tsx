import { useState } from 'react'
import { useTheme } from 'next-themes'
import { Check, Download, Eye, EyeOff, Monitor, Moon, Plus, RefreshCw, RotateCcw, Star, Sun, Terminal, Trash2, Zap } from 'lucide-react'
import { Block, Demo } from '@/components/Block'
import { blocksOf } from '@/lib/ledger'
import { ACCOUNTS, APP_SETTINGS, CACHE_DIRS, DB_FORM, DEBUG_PARAMS, ENVS, MIRROR, RUNTIME_INFO, SSH_FORM, SYNC, UPDATE_INFO } from '@/lib/fixtures'

import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Separator } from '@/components/ui/separator'
import { Progress } from '@/components/ui/progress'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldDescription, FieldGroup, FieldLabel } from '@/components/ui/field'
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Item, ItemContent, ItemDescription, ItemGroup, ItemTitle } from '@/components/ui/item'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription,
  AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { cn } from 'cn'

const B = blocksOf('settings')

const SECTIONS = [
  { key: 'sites', label: '站点管理' },
  { key: 'data-dict', label: '数据字典' },
  { key: 'debug', label: 'DEBUG' },
  { key: 'app', label: '应用设置' },
]

export function SettingsSection() {
  return (
    <div className="space-y-6">
      <Shell />
      <CardAndRows />
      <RowsAndFields />
      <AccountsTable />
      <ConfirmDelete />
      <DirtyStates />
      <ProgressBlocks />
      <ThemeCards />
      <ReadonlyInfo />
      <StatusNotes />
    </div>
  )
}

function Shell() {
  const [section, setSection] = useState('sites')
  return (
    <Block block={B['s-shell']} pad={false}>
      <div className="flex min-h-80 text-xs">
        <div className="w-44 shrink-0 border-r p-2">
          {SECTIONS.map((s) => (
            <Button
              key={s.key}
              variant={section === s.key ? 'secondary' : 'ghost'}
              size="sm"
              className="w-full justify-start"
              onClick={() => setSection(s.key)}
            >
              {s.label}
            </Button>
          ))}
          <Separator className="my-2" />
          <div className="px-2 text-[11px] leading-5 text-muted-foreground">
            分区清单只有一份,左导航与深链接 <code className="font-mono">#settings/data-dict</code> 共用。
          </div>
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2 border-b px-4 py-3">
            <h2 className="text-sm font-medium">{SECTIONS.find((s) => s.key === section)?.label}</h2>
            <span className="text-[11px] text-amber-600 dark:text-amber-400">有未保存的修改</span>
            <span className="ml-auto flex items-center gap-2">
              <Button variant="ghost" size="sm" title="从 config.json 重新读取" disabled>
                <RefreshCw className="size-3.5" /> 回读
              </Button>
            </span>
          </div>
          <div className="mx-auto max-w-3xl space-y-3 p-4">
            <Card className="gap-3 py-4">
              <CardHeader className="px-4">
                <CardTitle className="text-xs">{section === 'sites' ? '服务器与数据库' : section === 'data-dict' ? '源码镜像' : section === 'debug' ? '调试参数' : '服务'}</CardTitle>
                <CardDescription className="text-[11px] leading-4">
                  {section === 'sites'
                    ? 'SSH 登录与数据库连接按环境一对一挂载。运行时用哪个数据库账号由 TOPENT 决定(服务器侧解析),客户端直连取账号列表首项。'
                    : section === 'data-dict'
                      ? '把 T100 服务器的源码(4gl / 4fd / 42s 中文 / *.inc)拉到本地,供 AI 直接检索。'
                      : section === 'debug'
                        ? '本机调试会话的默认参数。清空某一项 = 恢复该项的内置默认值。'
                        : '统一 Web 服务的监听地址 —— 这套 SPA(调试工作台)就挂在它上面。'}
                </CardDescription>
              </CardHeader>
            </Card>
          </div>
        </div>
      </div>
      <div className="border-t bg-muted/30 px-4 py-2 text-[11px] text-muted-foreground">
        官方 Sidebar 是"应用级导航"(带 cookie 持久化与移动端 Sheet),对设置页可能过重;更轻的选择就是上面这种 Button 列表或竖排 Tabs。
      </div>
    </Block>
  )
}

function CardAndRows() {
  const [dirty, setDirty] = useState(true)
  const [pwd, setPwd] = useState('t100prd_pwd')
  const [showPwd, setShowPwd] = useState(false)
  return (
    <Block block={B['s-card']}>
      <div className="grid gap-5 lg:grid-cols-[16rem_1fr]">
        <Card className="gap-0 py-0">
          <CardHeader className="border-b px-4 py-3">
            <CardTitle className="text-xs">环境清单</CardTitle>
          </CardHeader>
          {/* 定高卡片:列表自己滚,底部按钮钉住 —— 官方 Card 要覆盖掉默认 padding 才能这么做 */}
          <CardContent className="flex h-56 flex-col p-0">
            <div className="min-h-0 flex-1 overflow-auto p-2">
              {ENVS.map((e) => (
                <div key={e.name} className={cn('flex items-center gap-1.5 rounded-md px-2 py-1.5 text-xs', e.active && 'bg-accent text-accent-foreground')}>
                  <span className="min-w-0 flex-1 truncate">
                    <span className="block truncate font-medium">{e.name}</span>
                    <span className="block truncate text-[11px] text-muted-foreground">
                      {e.zone || '-'} · {e.port || '-'} · {e.db}
                    </span>
                  </span>
                  <Star className={cn('size-3.5 shrink-0', e.active ? 'fill-amber-400 text-amber-400' : 'text-muted-foreground')} />
                </div>
              ))}
            </div>
            <div className="space-y-1.5 border-t p-2">
              <Button variant="outline" size="sm" className="w-full"><Plus className="size-3.5" /> 新增环境</Button>
              <div className="text-[11px] text-muted-foreground">星标 = 默认环境(36正式)</div>
            </div>
          </CardContent>
        </Card>

        <Card className="gap-0 py-0">
          <CardHeader className="flex flex-row items-start justify-between gap-3 px-4 py-3">
            <div className="min-w-0">
              <CardTitle className="text-xs">服务器与数据库</CardTitle>
              <CardDescription className="mt-1 text-[11px] leading-4">SSH 登录与数据库连接按环境一对一挂载。</CardDescription>
            </div>
            <CardAction className="flex items-center gap-1.5">
              {/* 关键:保存按钮"不脏时根本不渲染" —— 按钮存在性本身是状态表达 */}
              {dirty && <Button size="sm"><Check className="size-3.5" /> 保存</Button>}
              <Button variant="ghost" size="sm" onClick={() => setDirty((v) => !v)}>{dirty ? '放弃修改' : '改一下'}</Button>
            </CardAction>
          </CardHeader>
          <CardContent className="px-4 pb-4">
            <FieldGroup className="grid gap-x-6 gap-y-4 sm:grid-cols-2">
              <Field>
                <FieldLabel htmlFor="s-name" className="text-xs">环境名称(留空自动为主机-区域)</FieldLabel>
                <Input id="s-name" defaultValue={SSH_FORM.name} placeholder="如 正式区" onChange={() => setDirty(true)} />
              </Field>
              <Field>
                <FieldLabel htmlFor="s-host" className="text-xs">IP 主机</FieldLabel>
                <Input id="s-host" defaultValue={SSH_FORM.host} onChange={() => setDirty(true)} />
              </Field>
              <Field>
                <FieldLabel htmlFor="s-port" className="text-xs">端口</FieldLabel>
                <Input id="s-port" defaultValue={SSH_FORM.port} onChange={() => setDirty(true)} />
              </Field>
              <Field>
                <FieldLabel htmlFor="s-zone" className="text-xs">登录区域</FieldLabel>
                <Input id="s-zone" defaultValue={SSH_FORM.zone} placeholder="31开发/35测试/36正式/39PATCH/t出货" onChange={() => setDirty(true)} />
              </Field>
              <Field>
                <FieldLabel htmlFor="s-user" className="text-xs">账号</FieldLabel>
                <Input id="s-user" defaultValue={SSH_FORM.user} onChange={() => setDirty(true)} />
              </Field>
              <Field>
                <FieldLabel htmlFor="s-pwd" className="text-xs">密码</FieldLabel>
                <InputGroup>
                  <InputGroupInput
                    id="s-pwd"
                    type={showPwd ? 'text' : 'password'}
                    value={pwd}
                    onChange={(e) => { setPwd(e.target.value); setDirty(true) }}
                  />
                  <InputGroupAddon align="inline-end">
                    <Button variant="ghost" size="icon" className="size-7" onClick={() => setShowPwd((v) => !v)} title={showPwd ? '隐藏密码' : '显示密码'}>
                      {showPwd ? <EyeOff className="size-3.5" /> : <Eye className="size-3.5" />}
                    </Button>
                  </InputGroupAddon>
                </InputGroup>
              </Field>
              <Field className="sm:col-span-2">
                <FieldLabel htmlFor="s-topent" className="text-xs">TOPENT(默认企业;连接会话即下发,可数字或文本)</FieldLabel>
                <Input id="s-topent" defaultValue={SSH_FORM.topent} onChange={() => setDirty(true)} />
              </Field>
              <Field className="sm:col-span-2">
                <FieldLabel htmlFor="s-ws" className="text-xs">工作区目录(该环境的 .tzs 工作区;绝对路径)</FieldLabel>
                <Input id="s-ws" defaultValue={SSH_FORM.workspace} className="font-mono text-xs" onChange={() => setDirty(true)} />
                <FieldDescription className="text-[11px] leading-4">
                  调试会话按该服务器登录区域(区域 + TOPENT)。工作区目录按环境存（切环境就跟着换），是 tt dev tzs 的必填项。
                </FieldDescription>
              </Field>
            </FieldGroup>
          </CardContent>
        </Card>
      </div>
    </Block>
  )
}

function RowsAndFields() {
  return (
    <Block block={B['s-row']}>
      <div className="grid gap-6 lg:grid-cols-2">
        <Card className="gap-3 py-4">
          <CardHeader className="px-4"><CardTitle className="text-xs">调试参数(六行,控件宽度收到 max-w-md)</CardTitle></CardHeader>
          <CardContent className="space-y-4 px-4">
            <Field>
              <FieldLabel htmlFor="r-launch" className="text-xs">启动参数模板</FieldLabel>
              <Input id="r-launch" defaultValue={DEBUG_PARAMS.launch} className="max-w-md font-mono text-xs" />
              <FieldDescription className="text-[11px]">{'{prog} 会替换为作业编号。清空 = 用内置默认。'}</FieldDescription>
            </Field>
            <Field>
              <FieldLabel htmlFor="r-wd" className="text-xs">停站看门狗(秒)</FieldLabel>
              <Input id="r-wd" defaultValue={DEBUG_PARAMS.watchdog} className="max-w-md" />
              <FieldDescription className="text-[11px]">停站后原地停留超过该时长就判定会话失联。</FieldDescription>
            </Field>
            <Field>
              <FieldLabel className="text-xs">终端尺寸(宽 × 高)</FieldLabel>
              <div className="flex max-w-md items-center gap-2">
                <Input defaultValue={DEBUG_PARAMS.termW} className="flex-1" />
                <span className="text-muted-foreground">×</span>
                <Input defaultValue={DEBUG_PARAMS.termH} className="flex-1" />
              </div>
            </Field>
          </CardContent>
        </Card>

        <div className="space-y-4">
          <Card className="gap-3 py-4">
            <CardHeader className="px-4"><CardTitle className="text-xs">数据库(条件字段随类型变)</CardTitle></CardHeader>
            <CardContent className="space-y-4 px-4">
              <Field>
                <FieldLabel className="text-xs">类型</FieldLabel>
                <Select defaultValue={DB_FORM.type}>
                  <SelectTrigger className="max-w-md"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="oracle">Oracle</SelectItem>
                    <SelectItem value="kingbase">金仓(KingbaseES)</SelectItem>
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel htmlFor="r-host" className="text-xs">主机地址</FieldLabel>
                <Input id="r-host" defaultValue={DB_FORM.host} className="max-w-md" placeholder="客户端与服务器侧均可达" />
              </Field>
              <Field>
                <FieldLabel htmlFor="r-db" className="text-xs">库名</FieldLabel>
                <Input id="r-db" defaultValue={DB_FORM.dbname} className="max-w-md" placeholder="如 topprd" />
              </Field>
            </CardContent>
          </Card>

          <Card className="gap-3 py-4">
            <CardHeader className="px-4"><CardTitle className="text-xs">服务 / 文档目录(单字段行)</CardTitle></CardHeader>
            <CardContent className="space-y-4 px-4">
              <Field>
                <FieldLabel htmlFor="r-listen" className="text-xs">监听地址</FieldLabel>
                <Input id="r-listen" defaultValue={APP_SETTINGS.listen} className="max-w-md font-mono text-xs" placeholder="127.0.0.1:28670" />
                <FieldDescription className="text-[11px]">改完需要重启 tt serve 才生效(监听只在启动时绑定一次)。端口被占用时会自动向后顺延。</FieldDescription>
              </Field>
              <Field>
                <FieldLabel htmlFor="r-bdl" className="text-xs">文档目录</FieldLabel>
                <Input id="r-bdl" defaultValue={APP_SETTINGS.bdlDoc} className="max-w-md font-mono text-xs" />
                <FieldDescription className="text-[11px]">该目录存在于本机。</FieldDescription>
              </Field>
            </CardContent>
          </Card>
        </div>
      </div>
      <div className="mt-3 text-xs text-muted-foreground">
        官方 Field 族的 <code className="font-mono">FieldLabel / FieldDescription</code> 与现有 SettingRow 几乎一比一对应,是这次改造收益最高的一处。
        唯一要保住的是"控件宽度统一收到 max-w-md":否则同一张卡里的下拉、按钮会各走各的宽度。
      </div>
    </Block>
  )
}

function AccountsTable() {
  const [rows, setRows] = useState(ACCOUNTS.map((a) => ({ ...a })))
  const [draft, setDraft] = useState({ schema: '', password: '' })
  const add = () => {
    if (!draft.schema.trim()) return
    setRows((r) => [...r, { schema: draft.schema.trim(), password: draft.password.trim() }])
    setDraft({ schema: '', password: '' })
  }
  return (
    <Block block={B['s-accounts']} pad={false}>
      <div className="p-4">
        <div className="flex flex-wrap items-center gap-2">
          <Label className="text-xs">数据库账号(schema)</Label>
          <Badge variant="secondary" className="ml-auto">允许 AI 执行只读 SQL</Badge>
          <Switch defaultChecked aria-label="允许 AI 执行只读 SQL" />
        </div>
        <div className="mt-3 max-h-56 overflow-auto rounded-lg border">
          {/* 自供滚动容器 + 表头不透明 + 单元格无边框填充:三条一起才成立 */}
          <table className="w-full border-separate border-spacing-0 text-xs">
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="sticky top-0 z-10 border-b bg-muted px-2 text-xs">账号(schema)</TableHead>
                <TableHead className="sticky top-0 z-10 border-b bg-muted px-2 text-xs">密码(缺省 = 账号)</TableHead>
                <TableHead className="sticky top-0 z-10 w-24 border-b bg-muted px-2 text-xs">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow className="hover:bg-transparent">
                <TableCell className="border-b p-0 focus-within:ring-1 focus-within:ring-ring/60 focus-within:ring-inset">
                  <Input
                    value={draft.schema}
                    onChange={(e) => setDraft((d) => ({ ...d, schema: e.target.value }))}
                    onKeyDown={(e) => { if (e.key === 'Enter') add() }}
                    placeholder="新增账号…"
                    className="h-8 rounded-none border-0 bg-transparent font-mono text-xs shadow-none focus-visible:ring-0 dark:bg-transparent"
                  />
                </TableCell>
                <TableCell className="border-b p-0 focus-within:ring-1 focus-within:ring-ring/60 focus-within:ring-inset">
                  <Input
                    value={draft.password}
                    onChange={(e) => setDraft((d) => ({ ...d, password: e.target.value }))}
                    onKeyDown={(e) => { if (e.key === 'Enter') add() }}
                    placeholder="留空 = 与账号相同"
                    className="h-8 rounded-none border-0 bg-transparent font-mono text-xs shadow-none focus-visible:ring-0 dark:bg-transparent"
                  />
                </TableCell>
                <TableCell className="border-b px-2">
                  <Button variant="ghost" size="icon" className="size-7" onClick={add} title="添加账号"><Plus className="size-3.5" /></Button>
                </TableCell>
              </TableRow>
              {rows.map((a, i) => (
                <TableRow key={`${a.schema}-${i}`} className="hover:bg-transparent">
                  <TableCell className="border-b p-0">
                    <Input
                      value={a.schema}
                      onChange={(e) => setRows((r) => r.map((x, j) => (j === i ? { ...x, schema: e.target.value } : x)))}
                      className="h-8 rounded-none border-0 bg-transparent font-mono text-xs shadow-none focus-visible:ring-0 dark:bg-transparent"
                    />
                  </TableCell>
                  <TableCell className="border-b p-0">
                    <Input
                      value={a.password}
                      onChange={(e) => setRows((r) => r.map((x, j) => (j === i ? { ...x, password: e.target.value } : x)))}
                      className="h-8 rounded-none border-0 bg-transparent font-mono text-xs shadow-none focus-visible:ring-0 dark:bg-transparent"
                    />
                  </TableCell>
                  <TableCell className="border-b px-2">
                    <span className="flex items-center gap-0.5">
                      <Button variant="ghost" size="icon" className="size-7" title="在服务器侧验证该账号"><Zap className="size-3.5" /></Button>
                      <Button variant="ghost" size="icon" className="size-7 text-red-600 dark:text-red-400" title="删除该账号" onClick={() => setRows((r) => r.filter((_, j) => j !== i))}>
                        <Trash2 className="size-3.5" />
                      </Button>
                    </span>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </table>
        </div>
        <div className="mt-2 flex items-center gap-3 text-[11px] text-muted-foreground">
          <span>回车提交新增行(不点加号也行)</span>
          <span className="text-emerald-600 dark:text-emerald-400">连接正常 ✓</span>
          <span className="text-red-600 dark:text-red-400">请先填写 SSH 主机与账号</span>
          <span>验证中…</span>
        </div>
      </div>
      <div className="border-t bg-muted/30 px-4 py-2 text-[11px] text-muted-foreground">
        三重耦合:自供滚动容器(官方 Table 的容器会让 sticky 失效)、表头必须<b>不透明</b>(默认 bg-muted/50 会透行)、
        格内输入必须<b>无边框填满格</b>。且 Enter 即提交 —— 任何包装组件吞掉 keydown 都会静默失效(无测试覆盖)。
      </div>
    </Block>
  )
}

function ConfirmDelete() {
  return (
    <Block block={B['s-confirm']}>
      <div className="flex flex-wrap items-center gap-3">
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button variant="destructive" size="sm"><Trash2 className="size-3.5" /> 删除环境</Button>
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
        <span className="text-xs text-muted-foreground">
          这处已经是官方形态;真正要保住的是:确认后<b>删除 + 立即保存</b>,且被删的是默认环境时默认环境要<b>顺延到首条</b>(否则后端 400)。
        </span>
      </div>
    </Block>
  )
}

function DirtyStates() {
  const states = [
    { label: '干净(无按钮)', dirty: false, saving: false },
    { label: '脏(保存按钮出现)', dirty: true, saving: false },
    { label: '保存中(按钮禁用 + 文案变)', dirty: true, saving: true },
  ]
  return (
    <Block block={B['s-dirty']}>
      <div className="grid gap-3 sm:grid-cols-3">
        {states.map((s) => (
          <Card key={s.label} className="gap-2 py-4">
            <CardHeader className="px-4">
              <CardTitle className="text-xs">{s.label}</CardTitle>
              <CardAction>
                {s.dirty && <Button size="sm" disabled={s.saving}>{s.saving ? '保存中…' : '保存'}</Button>}
              </CardAction>
            </CardHeader>
          </Card>
        ))}
      </div>
      <div className="mt-3 text-xs text-muted-foreground">
        脏标记按 <b>7 个配置节</b>分别跟踪,一个按钮的 disabled 只依赖自己那一节。
        换成全局脏标记会回归"用陈旧快照覆盖别张卡刚改好的节"这个已修复的 bug。
      </div>
    </Block>
  )
}

function ProgressBlocks() {
  return (
    <Block block={B['s-progress']}>
      <div className="grid gap-4 lg:grid-cols-2">
        <Card className="gap-3 py-4">
          <CardHeader className="px-4">
            <CardTitle className="text-xs">源码镜像</CardTitle>
            <CardDescription className="text-[11px]">增量按服务器 marker 记基线;本地无完整镜像时自动转全量。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 px-4">
            <Field>
              <FieldLabel className="text-xs">镜像根目录</FieldLabel>
              <Input defaultValue={MIRROR.dir} className="font-mono text-xs" />
              <FieldDescription className="text-[11px]">该目录存在于本机。</FieldDescription>
            </Field>
            <div className="space-y-1.5">
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>{MIRROR.phase}</span><span>{MIRROR.percent}%</span>
              </div>
              <Progress value={MIRROR.percent} />
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Button size="sm" variant="outline"><RefreshCw className="size-3.5" /> 增量更新</Button>
              <Button size="sm" variant="outline">全量重建</Button>
            </div>
          </CardContent>
        </Card>

        <Card className="gap-3 py-4">
          <CardHeader className="px-4">
            <CardTitle className="text-xs">数据同步</CardTitle>
            <CardDescription className="text-[11px]">把远程 ERP 的字典表同步成本地 SQLite,查询命令再读它。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 px-4">
            <Field>
              <FieldLabel className="text-xs">目标数据库文件</FieldLabel>
              <InputGroup>
                <InputGroupInput defaultValue={SYNC.target} className="font-mono text-xs" />
                <InputGroupAddon align="inline-end">
                  <Button variant="ghost" size="icon" className="size-7" title="清空 = 恢复默认位置"><RotateCcw className="size-3.5" /></Button>
                </InputGroupAddon>
              </InputGroup>
              <FieldDescription className="text-[11px]">该文件已存在(同步时会自动备份为 .bak)。</FieldDescription>
            </Field>
            <div className="space-y-1.5">
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>{SYNC.phase}</span><span>{SYNC.percent}%</span>
              </div>
              <Progress value={SYNC.percent} />
            </div>
            <Button size="sm" disabled>同步中…</Button>
            <Alert className="border-amber-600/30 text-amber-700 dark:text-amber-400">
              <AlertDescription className="whitespace-pre-wrap font-mono text-[11px] text-amber-700/90 dark:text-amber-400/90">{SYNC.warning}</AlertDescription>
            </Alert>
          </CardContent>
        </Card>
      </div>
      <div className="mt-3 text-xs text-muted-foreground">
        两个任务的阶段语义与百分比算法不同,轮询(800ms)只在 running 时挂载 —— 换 Progress 时不要把轮询挪进常驻 effect。
      </div>
    </Block>
  )
}

function ThemeCards() {
  const { theme, setTheme } = useTheme()
  const items = [
    { key: 'light', label: '亮色', icon: Sun },
    { key: 'dark', label: '暗色', icon: Moon },
    { key: 'system', label: '跟随系统', icon: Monitor },
  ]
  return (
    <Block block={B['s-theme']}>
      <div className="grid max-w-md grid-cols-3 gap-2">
        {items.map(({ key, label, icon: Icon }) => (
          <button
            key={key}
            aria-pressed={theme === key}
            title={`切换到「${label}」`}
            onClick={() => setTheme(key)}
            className={cn(
              'flex flex-col gap-2 rounded-lg border p-2 text-left transition-colors',
              theme === key ? 'border-foreground/50 bg-accent' : 'border-border hover:bg-accent/50',
            )}
          >
            <Preview kind={key} />
            <span className="flex items-center gap-1.5 text-xs"><Icon className="size-3.5" />{label}</span>
          </button>
        ))}
      </div>
      <div className="mt-3 text-xs text-muted-foreground">
        这三张卡里的色块是<b>预览内容</b>而非主题取值 —— 不要被"禁止写死色板"的规矩扫掉。
        点击可真实切换整个演示台的主题(走 next-themes,键名 tt.theme,与产品同一份偏好)。
      </div>
    </Block>
  )
}

function Preview({ kind }: { kind: string }) {
  const light = (
    <span className="flex h-12 flex-1 flex-col gap-1 rounded-md border bg-white p-1.5">
      <span className="h-1.5 w-2/3 rounded-sm bg-[#d4d4d4]" />
      <span className="h-6 rounded-sm bg-[#f4f4f4]" />
    </span>
  )
  const dark = (
    <span className="flex h-12 flex-1 flex-col gap-1 rounded-md border border-[#4a4a4a] bg-[#1f1f1f] p-1.5">
      <span className="h-1.5 w-2/3 rounded-sm bg-[#4a4a4a]" />
      <span className="h-6 rounded-sm bg-[#2b2b2b]" />
    </span>
  )
  return (
    <span className="flex gap-1">
      {kind === 'light' && light}
      {kind === 'dark' && dark}
      {kind === 'system' && <>{light}{dark}</>}
    </span>
  )
}

function ReadonlyInfo() {
  return (
    <Block block={B['s-readonly']}>
      <div className="grid gap-4 lg:grid-cols-2">
        <Card className="gap-2 py-4">
          <CardHeader className="px-4"><CardTitle className="text-xs">运行信息</CardTitle></CardHeader>
          <CardContent className="px-4">
            <ItemGroup className="gap-0">
              {RUNTIME_INFO.map((r) => (
                <Item key={r.label} className="flex-nowrap items-baseline gap-3 rounded-none border-b px-0 py-1.5 last:border-b-0">
                  <ItemContent className="flex-row items-baseline gap-3">
                    <ItemTitle className="w-28 shrink-0 text-[11px] font-normal text-muted-foreground">{r.label}</ItemTitle>
                    <ItemDescription className="min-w-0 flex-1 break-all font-mono text-[11px] text-foreground">{r.value}</ItemDescription>
                  </ItemContent>
                </Item>
              ))}
            </ItemGroup>
          </CardContent>
        </Card>

        <div className="space-y-4">
          <Card className="gap-2 py-4">
            <CardHeader className="px-4"><CardTitle className="text-xs">缓存</CardTitle></CardHeader>
            <CardContent className="space-y-3 px-4">
              <div className="space-y-0.5">
                {CACHE_DIRS.map((c) => (
                  <div key={c.name} className="flex items-baseline justify-between gap-3 text-[11px]">
                    <span className="font-mono text-muted-foreground">{c.name}</span>
                    <span className="font-mono">{c.files === 0 ? '无' : `${c.bytes} · ${c.files} 个文件`}</span>
                  </div>
                ))}
              </div>
              <Separator />
              <div className="flex items-center justify-between">
                <span className="text-[11px] text-muted-foreground">共 56.6 MB · 1293 个文件</span>
                <Button size="sm" variant="outline"><Trash2 className="size-3.5" /> 清除缓存</Button>
              </div>
            </CardContent>
          </Card>

          <Card className="gap-2 py-4">
            <CardHeader className="px-4">
              <CardTitle className="text-xs">更新</CardTitle>
              <CardDescription className="text-[11px] leading-4">
                查 tt 自己有没有新版。查是你点才联网(服务不会替你轮询);装会先下载并校验 sha256,再交给一个独立的更新器进程去做。
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 px-4">
              <ItemGroup className="gap-0">
                {UPDATE_INFO.map((r) => (
                  <Item key={r.label} className="flex-nowrap items-baseline gap-3 rounded-none border-b px-0 py-1.5 last:border-b-0">
                    <ItemContent className="flex-row items-baseline gap-3">
                      <ItemTitle className="w-28 shrink-0 text-[11px] font-normal text-muted-foreground">{r.label}</ItemTitle>
                      <ItemDescription className="min-w-0 flex-1 text-[11px] text-foreground">{r.value}</ItemDescription>
                    </ItemContent>
                  </Item>
                ))}
              </ItemGroup>
              <div className="flex flex-wrap gap-2">
                <Button size="sm" variant="outline"><RefreshCw className="size-3.5" /> 检查更新</Button>
                <Button size="sm"><Download className="size-3.5" /> 升级到 0.2.4</Button>
              </div>
              <div className="text-[11px] text-muted-foreground">升级期间本页会断开,之后重新打开即可</div>
            </CardContent>
          </Card>
        </div>
      </div>
    </Block>
  )
}

function StatusNotes() {
  return (
    <Block block={B['s-info']}>
      <div className="space-y-3">
        <Demo label="顶部状态簇(全页级)">
          <div className="flex flex-wrap items-center gap-3 text-xs">
            <span className="text-red-600 dark:text-red-400">保存失败:config.json 只读</span>
            <span className="text-emerald-600 dark:text-emerald-400">已保存 ✓</span>
            <span className="text-amber-600 dark:text-amber-400">有未保存的修改</span>
            <span className="text-muted-foreground">保存中…</span>
          </div>
        </Demo>
        <Demo label="卡片级 note(就近显示,不要收编到顶部或 toast)">
          <div className="grid gap-3 sm:grid-cols-2">
            <Alert className="border-emerald-600/30 text-emerald-700 dark:text-emerald-400">
              <AlertTitle className="text-xs">连接正常 ✓</AlertTitle>
              <AlertDescription className="text-[11px] text-emerald-700/80 dark:text-emerald-400/80">KingbaseES V008R006C008B0014</AlertDescription>
            </Alert>
            <Alert variant="destructive">
              <AlertTitle className="text-xs">获取失败</AlertTitle>
              <AlertDescription className="text-[11px]">请先填写 SSH 主机与账号</AlertDescription>
            </Alert>
          </div>
        </Demo>
        <Demo label="只读信息行(命令 / 路径)">
          <div className="flex items-center gap-2">
            <Button size="sm" variant="outline"><Terminal className="size-3.5" /> 加入用户 PATH</Button>
            <Button size="sm" variant="destructive"><Trash2 className="size-3.5" /> 从 PATH 移除</Button>
            <Badge variant="outline" className="font-mono">D:\Programs\TT</Badge>
          </div>
        </Demo>
      </div>
    </Block>
  )
}
