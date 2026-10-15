import { useState } from 'react'
import { useTheme } from 'next-themes'
import {
  Download, Eye, EyeOff, Monitor, Moon, Plus, RefreshCw, RotateCcw, Star, Sun, Terminal, Trash2, Zap,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Progress } from '@/components/ui/progress'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldDescription, FieldGroup, FieldLabel, FieldLegend, FieldSet } from '@/components/ui/field'
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { StatusNote } from './status'
import { useConfirm } from './useConfirm'
import {
  ACCOUNTS, APP_SETTINGS, CACHE_DIRS, DB_FORM, DEBUG_PARAMS, ENVS, MIRROR, RUNTIME_INFO, SSH_FORM, SYNC, UPDATE_INFO,
} from '@/lib/fixtures'
import { ACCT_PROBE, CACHE_TOTAL } from './mock'
import { cn } from 'cn'

const SECTIONS = [
  { key: 'sites', label: '站点管理' },
  { key: 'data-dict', label: '数据字典' },
  { key: 'debug', label: 'DEBUG' },
  { key: 'app', label: '应用设置' },
] as const

type SectionKey = (typeof SECTIONS)[number]['key']
/** 脏标记按配置节跟踪:一个按钮的可用性只依赖自己那一节 */
type DirtyKey = 'hosts' | 'query' | 'mirror' | 'sync' | 'debug' | 'app' | 'bdldoc'

export function SettingsView({ onToast, initialSection = 'sites' }: { onToast: (m: string) => void; initialSection?: SectionKey }) {
  const [section, setSection] = useState<SectionKey>(initialSection)
  const [dirty, setDirty] = useState<Record<DirtyKey, boolean>>({
    hosts: true, query: false, mirror: false, sync: false, debug: false, app: false, bdldoc: false,
  })
  const [saving, setSaving] = useState<DirtyKey | null>(null)

  const mark = (k: DirtyKey) => setDirty((d) => (d[k] ? d : { ...d, [k]: true }))
  const save = (k: DirtyKey, label: string) => {
    setSaving(k)
    window.setTimeout(() => {
      setSaving(null)
      setDirty((d) => ({ ...d, [k]: false }))
      onToast(`${label}已保存到 config.json`)
    }, 600)
  }

  const anyDirty = Object.values(dirty).some(Boolean)

  return (
    <div className="flex min-h-0 flex-1 text-xs">
      <div className="flex w-44 shrink-0 flex-col border-r p-2">
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
      </div>

      <div className="flex min-w-0 flex-1 flex-col">
        <div className="flex shrink-0 items-center gap-2 border-b px-4 py-2.5">
          <h2 className="text-sm font-medium">{SECTIONS.find((s) => s.key === section)?.label}</h2>
          {anyDirty
            ? <StatusNote tone="warning" className="py-0.5">有未保存的修改</StatusNote>
            : <span className="text-[11px] text-success">已保存 ✓</span>}
          <Button variant="ghost" size="sm" className="ml-auto" disabled={anyDirty} title="从 config.json 重新读取">
            <RefreshCw className="size-3.5" /> 回读
          </Button>
        </div>

        <div className="min-h-0 flex-1 overflow-auto">
          <div className={cn('mx-auto space-y-3 p-4', section === 'sites' ? 'max-w-5xl' : 'max-w-3xl')}>
            {section === 'sites' && <Sites dirty={dirty} saving={saving} mark={mark} save={save} onToast={onToast} />}
            {section === 'data-dict' && <DataDict dirty={dirty} saving={saving} mark={mark} save={save} onToast={onToast} />}
            {section === 'debug' && <DebugSection dirty={dirty} saving={saving} mark={mark} save={save} />}
            {section === 'app' && <AppSection dirty={dirty} saving={saving} mark={mark} save={save} onToast={onToast} />}
          </div>
        </div>
      </div>
    </div>
  )
}

interface SecProps {
  dirty: Record<DirtyKey, boolean>
  saving: DirtyKey | null
  mark: (k: DirtyKey) => void
  save: (k: DirtyKey, label: string) => void
  onToast: (m: string) => void
}

/** 卡片级保存按钮:不脏就不渲染(按钮存在性本身是状态表达) */
function SaveBtn({ k, label, dirty, saving, save }: { k: DirtyKey; label: string } & Pick<SecProps, 'dirty' | 'saving' | 'save'>) {
  if (!dirty[k]) return null
  return (
    <Button size="sm" disabled={saving === k} onClick={() => save(k, label)}>
      {saving === k ? '保存中…' : '保存'}
    </Button>
  )
}

function Sites({ dirty, saving, mark, save, onToast }: SecProps) {
  const confirm = useConfirm()
  const [env, setEnv] = useState('36正式')
  const [tab, setTab] = useState('ssh')
  const [showPwd, setShowPwd] = useState(false)
  const [roSql, setRoSql] = useState(true)
  const [accts, setAccts] = useState(ACCOUNTS.map((a) => ({ ...a })))
  const [draft, setDraft] = useState({ schema: '', password: '' })
  const [probe, setProbe] = useState<'idle' | 'ok' | 'err'>('ok')

  const addAcct = () => {
    if (!draft.schema.trim()) return
    setAccts((r) => [...r, { schema: draft.schema.trim(), password: draft.password.trim() }])
    setDraft({ schema: '', password: '' })
    mark('hosts')
  }

  return (
    <div className="grid gap-3 lg:grid-cols-[15rem_1fr]">
      <Card className="gap-0 py-0">
        <CardHeader className="border-b px-4 py-3">
          <CardTitle className="text-xs">环境清单</CardTitle>
        </CardHeader>
        <CardContent className="flex h-72 flex-col p-0">
          <div className="min-h-0 flex-1 space-y-0.5 overflow-auto p-2">
            {ENVS.map((e) => (
              <button
                key={e.name}
                onClick={() => setEnv(e.name)}
                className={cn('flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left transition-colors', env === e.name ? 'bg-accent text-accent-foreground' : 'hover:bg-accent/50')}
              >
                <span className="min-w-0 flex-1">
                  <span className="block truncate font-medium">{e.name}</span>
                  <span className="block truncate text-[11px] text-muted-foreground">
                    {e.zone || '-'} · {e.port || '-'} · {e.db}
                  </span>
                </span>
                {e.active && <Star className="size-3.5 shrink-0 fill-warning text-warning" aria-label="当前默认环境" />}
              </button>
            ))}
          </div>
          <div className="space-y-1.5 border-t p-2">
            <Button variant="outline" size="sm" className="w-full" onClick={() => onToast('已新增环境(虚拟预览不会真的写入)')}>
              <Plus className="size-3.5" /> 新增环境
            </Button>
            <div className="text-[11px] text-muted-foreground">星标 = 默认环境(36正式)</div>
          </div>
        </CardContent>
      </Card>

      <Card className="gap-0 py-0">
        <CardHeader className="px-4 py-3">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div className="min-w-0">
              <CardTitle className="text-xs">服务器与数据库</CardTitle>
              <CardDescription className="mt-1 text-[11px] leading-4">
                SSH 登录与数据库连接按环境一对一挂载。运行时用哪个数据库账号由 TOPENT 决定(服务器侧解析)。
              </CardDescription>
            </div>
            <CardAction className="flex items-center gap-1.5">
              <SaveBtn k="hosts" label="服务器与数据库" {...{ dirty, saving, save }} />
              <Button
                variant="ghost"
                size="sm"
                className="text-destructive"
                onClick={async () => {
                  const ok = await confirm({
                    title: `删除环境「${env}」?`,
                    description: '该环境的 SSH 连接、登录区域、TOPENT 与数据库账号配置将一并移除。点击「删除」后立即写入 config.json,不可撤销。',
                    confirmText: '删除',
                  })
                  if (ok) onToast(`已删除环境「${env}」并写入 config.json`)
                }}
              >
                <Trash2 className="size-3.5" /> 删除该环境
              </Button>
            </CardAction>
          </div>
        </CardHeader>

        <CardContent className="px-4 pb-4">
          <Tabs value={tab} onValueChange={setTab}>
            <TabsList>
              <TabsTrigger value="ssh">SSH 服务器</TabsTrigger>
              <TabsTrigger value="db">数据库</TabsTrigger>
            </TabsList>

            {/* 最佳实践:切页签不丢输入 —— Radix 默认卸载非活动页,非受控输入会被重置回 defaultValue。
                所以这里 forceMount + 非活动时 hidden(官方 TabsContent 只给 flex-1 outline-none,
                不管隐藏;隐藏得自己写)。产品现状"切 tab 不卸载状态"靠的就是同一件事。 */}
            <TabsContent value="ssh" forceMount className="mt-3 data-[state=inactive]:hidden">
              <FieldSet>
                <FieldLegend variant="label" className="sr-only">SSH 登录</FieldLegend>
                <FieldGroup className="grid gap-x-6 gap-y-4 sm:grid-cols-2">
                  <Field>
                    <FieldLabel htmlFor="v-name">环境名称(留空自动为主机-区域)</FieldLabel>
                    <Input id="v-name" defaultValue={SSH_FORM.name} placeholder="如 正式区" onChange={() => mark('hosts')} />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="v-host">IP 主机</FieldLabel>
                    <Input id="v-host" defaultValue={SSH_FORM.host} onChange={() => mark('hosts')} />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="v-port">端口</FieldLabel>
                    <Input id="v-port" defaultValue={SSH_FORM.port} onChange={() => mark('hosts')} />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="v-zone">登录区域</FieldLabel>
                    <Input id="v-zone" defaultValue={SSH_FORM.zone} placeholder="31开发/35测试/36正式/39PATCH/t出货" onChange={() => mark('hosts')} />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="v-user">账号</FieldLabel>
                    <Input id="v-user" defaultValue={SSH_FORM.user} onChange={() => mark('hosts')} />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="v-pwd">密码</FieldLabel>
                    <InputGroup>
                      <InputGroupInput id="v-pwd" type={showPwd ? 'text' : 'password'} defaultValue="t100prd_pwd" onChange={() => mark('hosts')} />
                      <InputGroupAddon align="inline-end">
                        <Button variant="ghost" size="icon" className="size-7" title={showPwd ? '隐藏密码' : '显示密码'} onClick={() => setShowPwd((v) => !v)}>
                          {showPwd ? <EyeOff className="size-3.5" /> : <Eye className="size-3.5" />}
                        </Button>
                      </InputGroupAddon>
                    </InputGroup>
                  </Field>
                  <Field className="sm:col-span-2">
                    <FieldLabel htmlFor="v-topent">TOPENT(默认企业;连接会话即下发,可数字或文本)</FieldLabel>
                    <Input id="v-topent" defaultValue={SSH_FORM.topent} onChange={() => mark('hosts')} />
                  </Field>
                  <Field className="sm:col-span-2">
                    <FieldLabel htmlFor="v-ws">工作区目录(该环境的 .tzs 工作区;绝对路径)</FieldLabel>
                    <Input id="v-ws" defaultValue={SSH_FORM.workspace} className="font-mono" onChange={() => mark('hosts')} />
                    <FieldDescription>
                      按环境存(切环境就跟着换),是 tt dev tzs 的必填项:它定位包与 out,留空时运行类动词会拒绝启动。
                    </FieldDescription>
                  </Field>
                </FieldGroup>
              </FieldSet>
            </TabsContent>

            <TabsContent value="db" forceMount className="mt-3 space-y-4 data-[state=inactive]:hidden">
              <FieldGroup className="grid gap-x-6 gap-y-4 sm:grid-cols-2">
                <Field>
                  <FieldLabel>类型</FieldLabel>
                  <Select defaultValue={DB_FORM.type}>
                    <SelectTrigger className="w-full"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="oracle">Oracle</SelectItem>
                      <SelectItem value="kingbase">金仓(KingbaseES)</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
                <Field>
                  <FieldLabel htmlFor="v-dbhost">主机地址</FieldLabel>
                  <Input id="v-dbhost" defaultValue={DB_FORM.host} placeholder="客户端与服务器侧均可达" onChange={() => mark('hosts')} />
                </Field>
                <Field>
                  <FieldLabel htmlFor="v-dbport">端口</FieldLabel>
                  <Input id="v-dbport" defaultValue={DB_FORM.port} onChange={() => mark('hosts')} />
                </Field>
                <Field>
                  <FieldLabel htmlFor="v-dbname">库名</FieldLabel>
                  <Input id="v-dbname" defaultValue={DB_FORM.dbname} placeholder="如 topprd" onChange={() => mark('hosts')} />
                </Field>
              </FieldGroup>

              <div className="flex flex-wrap items-center gap-3">
                <span className="flex items-center gap-2">
                  <Switch id="v-rosql" checked={roSql} onCheckedChange={(v) => { setRoSql(v); mark('hosts') }} />
                  <Label htmlFor="v-rosql" className="font-normal">允许 AI 执行只读 SQL(默认开启;关掉后调试端点的 SQL 查询直接 403)</Label>
                </span>
                <span className="ml-auto flex gap-2">
                  <Button variant="outline" size="sm" onClick={() => onToast('已按区域登录服务器,解析库地址并回填')}>
                    <RefreshCw className="size-3.5" /> 从服务器获取
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => setProbe('ok')}>
                    <Zap className="size-3.5" /> 测试连接
                  </Button>
                </span>
              </div>

              {probe === 'ok'
                ? <StatusNote tone="success">连接正常:KingbaseES V008R006C008B0014</StatusNote>
                : <StatusNote tone="danger">获取失败: 请先填写 SSH 主机与账号</StatusNote>}

              <div>
                <div className="mb-2 flex items-center gap-2">
                  <Label className="text-xs">数据库账号(schema)</Label>
                  <Badge variant="outline" className="ml-auto">库连接显式给地址与 service|库名,不依赖服务器侧 TNS 配置</Badge>
                </div>
                {/* 可编辑表格:自供滚动容器 + 表头不透明 + 格内无边框输入 + Enter 提交 */}
                <div className="max-h-56 overflow-auto rounded-lg border">
                  <table className="w-full border-separate border-spacing-0 text-xs">
                    <TableHeader>
                      <TableRow className="hover:bg-transparent">
                        <TableHead className="sticky top-0 z-10 h-8 border-b bg-muted px-2 text-[11px]">账号(schema)</TableHead>
                        <TableHead className="sticky top-0 z-10 h-8 border-b bg-muted px-2 text-[11px]">密码(缺省 = 账号)</TableHead>
                        <TableHead className="sticky top-0 z-10 h-8 w-40 border-b bg-muted px-2 text-[11px]">验证</TableHead>
                        <TableHead className="sticky top-0 z-10 h-8 w-20 border-b bg-muted px-2 text-[11px]">操作</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      <TableRow className="hover:bg-transparent">
                        <TableCell className="border-b p-0 focus-within:ring-1 focus-within:ring-ring/60 focus-within:ring-inset">
                          <Input value={draft.schema} onChange={(e) => setDraft((d) => ({ ...d, schema: e.target.value }))} onKeyDown={(e) => { if (e.key === 'Enter') addAcct() }} placeholder="新增账号…" className="h-8 rounded-none border-0 bg-transparent font-mono shadow-none focus-visible:ring-0 dark:bg-transparent" />
                        </TableCell>
                        <TableCell className="border-b p-0 focus-within:ring-1 focus-within:ring-ring/60 focus-within:ring-inset">
                          <Input value={draft.password} onChange={(e) => setDraft((d) => ({ ...d, password: e.target.value }))} onKeyDown={(e) => { if (e.key === 'Enter') addAcct() }} placeholder="留空 = 与账号相同" className="h-8 rounded-none border-0 bg-transparent font-mono shadow-none focus-visible:ring-0 dark:bg-transparent" />
                        </TableCell>
                        <TableCell className="border-b px-2 text-[11px] text-muted-foreground">回车提交</TableCell>
                        <TableCell className="border-b px-2">
                          <Button variant="ghost" size="icon" className="size-7" title="添加账号" onClick={addAcct}><Plus className="size-3.5" /></Button>
                        </TableCell>
                      </TableRow>
                      {accts.map((a, i) => (
                        <TableRow key={`${a.schema}-${i}`} className="hover:bg-transparent">
                          <TableCell className="border-b p-0">
                            <Input value={a.schema} onChange={(e) => { setAccts((r) => r.map((x, j) => (j === i ? { ...x, schema: e.target.value } : x))); mark('hosts') }} className="h-8 rounded-none border-0 bg-transparent font-mono shadow-none focus-visible:ring-0 dark:bg-transparent" />
                          </TableCell>
                          <TableCell className="border-b p-0">
                            <Input value={a.password} onChange={(e) => { setAccts((r) => r.map((x, j) => (j === i ? { ...x, password: e.target.value } : x))); mark('hosts') }} className="h-8 rounded-none border-0 bg-transparent font-mono shadow-none focus-visible:ring-0 dark:bg-transparent" />
                          </TableCell>
                          <TableCell className="border-b px-2 text-[11px]">
                            {ACCT_PROBE[i]
                              ? <span className={cn(ACCT_PROBE[i].state === 'ok' ? 'text-success' : ACCT_PROBE[i].state === 'err' ? 'text-destructive' : 'text-muted-foreground')}>{ACCT_PROBE[i].note}</span>
                              : null}
                          </TableCell>
                          <TableCell className="border-b px-2">
                            <span className="flex items-center gap-0.5">
                              <Button variant="ghost" size="icon" className="size-7" title="在服务器侧验证该账号" onClick={() => onToast(`已在服务器侧验证 ${a.schema}`)}><Zap className="size-3.5" /></Button>
                              <Button variant="ghost" size="icon" className="size-7 text-destructive" title="删除该账号" onClick={() => { setAccts((r) => r.filter((_, j) => j !== i)); mark('hosts') }}><Trash2 className="size-3.5" /></Button>
                            </span>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </table>
                </div>
              </div>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}

function DataDict({ dirty, saving, mark, save, onToast }: SecProps) {
  const confirm = useConfirm()
  const [running, setRunning] = useState<'mirror' | 'sync' | null>(null)
  const mirrorPct = 72
  const syncPct = 91

  return (
    <>
      <Card className="gap-3 py-4">
        <CardHeader className="px-4">
          <CardTitle className="text-xs">查询数据源</CardTitle>
          <CardDescription className="text-[11px] leading-4">
            r.t / r.v / desc / scc / r.q 等查询命令用哪个数据源。缺省是在线优先:用默认环境的远程库直查;一个环境都没配时才用本地镜像。
          </CardDescription>
        </CardHeader>
        <CardContent className="px-4">
          <Field className="max-w-md">
            <FieldLabel>数据源</FieldLabel>
            <Select defaultValue="__auto__" onValueChange={() => mark('query')}>
              <SelectTrigger className="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="__auto__">在线(默认环境)</SelectItem>
                <SelectItem value="local">本地 SQLite 镜像</SelectItem>
                <SelectItem value="36正式">在线(36正式)</SelectItem>
              </SelectContent>
            </Select>
            <FieldDescription>
              本地镜像由 tt dict db sync 生成;连不上不会静默回落本地 —— 那会让你以为查到的是实时数据。
            </FieldDescription>
          </Field>
        </CardContent>
      </Card>

      <Card className="gap-3 py-4">
        <CardHeader className="px-4">
          <CardTitle className="text-xs">源码镜像</CardTitle>
          <CardDescription className="text-[11px]">把 T100 服务器的源码(4gl / 4fd / 42s 中文 / *.inc)拉到本地,供 AI 直接检索。</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 px-4">
          <Field className="max-w-md">
            <FieldLabel htmlFor="v-mirror">镜像根目录</FieldLabel>
            <Input id="v-mirror" defaultValue={MIRROR.dir} className="font-mono" onChange={() => mark('mirror')} />
            <FieldDescription>该目录存在于本机。</FieldDescription>
          </Field>
          <Field className="max-w-md">
            <FieldLabel>拉取环境</FieldLabel>
            <Select defaultValue="36正式" onValueChange={() => mark('mirror')}>
              <SelectTrigger className="w-full"><SelectValue placeholder="选择环境" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="36正式">36正式（默认） · 已有镜像</SelectItem>
                <SelectItem value="35测试">35测试 · 未拉取</SelectItem>
              </SelectContent>
            </Select>
          </Field>
          {running === 'mirror' && (
            <div className="max-w-md space-y-1.5">
              <div className="flex justify-between text-[11px] text-muted-foreground"><span>{MIRROR.phase}</span><span>{mirrorPct}%</span></div>
              <Progress value={mirrorPct} />
            </div>
          )}
          <div className="flex flex-wrap items-center gap-2">
            <Button size="sm" variant="outline" disabled={running !== null} onClick={() => { setRunning('mirror'); window.setTimeout(() => { setRunning(null); onToast('镜像增量更新完成') }, 1500) }}>
              <RefreshCw className="size-3.5" /> 增量更新
            </Button>
            <Button
              size="sm"
              variant="outline"
              disabled={running !== null}
              onClick={async () => {
                const ok = await confirm({
                  title: '全量重建环境「36正式」?',
                  description: '将整目录替换本地镜像(删除服务器上已不存在的残留),首次或按需执行,耗时较长。',
                  confirmText: '继续',
                })
                if (!ok) return
                setRunning('mirror')
                window.setTimeout(() => { setRunning(null); onToast('镜像全量重建完成') }, 1500)
              }}
            >
              全量重建
            </Button>
            <span className="text-[11px] text-muted-foreground">增量按服务器 marker 记基线;本地无完整镜像时自动转全量。</span>
          </div>
        </CardContent>
      </Card>

      <Card className="gap-3 py-4">
        <CardHeader className="px-4">
          <CardTitle className="text-xs">数据同步</CardTitle>
          <CardDescription className="text-[11px]">把远程 ERP 的字典表同步成本地 SQLite,查询命令再读它。</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 px-4">
          <Field className="max-w-md">
            <FieldLabel htmlFor="v-sync">目标数据库文件</FieldLabel>
            <InputGroup>
              <InputGroupInput id="v-sync" defaultValue={SYNC.target} className="font-mono" onChange={() => mark('sync')} />
              <InputGroupAddon align="inline-end">
                <Button variant="ghost" size="icon" className="size-7" title="清空 = 恢复默认位置" onClick={() => mark('sync')}><RotateCcw className="size-3.5" /></Button>
              </InputGroupAddon>
            </InputGroup>
            <FieldDescription>该文件已存在(同步时会自动备份为 .bak)。</FieldDescription>
          </Field>
          <Field className="max-w-md">
            <FieldLabel>同步环境</FieldLabel>
            <Select defaultValue="35测试" onValueChange={() => mark('sync')}>
              <SelectTrigger className="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="35测试">35测试 · Oracle</SelectItem>
                <SelectItem value="36正式">36正式 · 金仓</SelectItem>
              </SelectContent>
            </Select>
          </Field>
          {running === 'sync' && (
            <div className="max-w-md space-y-1.5">
              <div className="flex justify-between text-[11px] text-muted-foreground"><span>{SYNC.phase}</span><span>{syncPct}%</span></div>
              <Progress value={syncPct} />
            </div>
          )}
          <div className="flex flex-wrap items-center gap-2">
            <Button
              size="sm"
              disabled={running !== null}
              onClick={async () => {
                const ok = await confirm({
                  title: '从环境「35测试」拉取字典数据到本地?',
                  description: `目标文件:\n${SYNC.target}\n\n将覆盖本地 SQLite(原库自动备份为 .bak),约 85 万行,需数分钟。`,
                  confirmText: '开始同步',
                })
                if (!ok) return
                setRunning('sync')
                window.setTimeout(() => { setRunning(null); onToast('字典同步完成:856,204 行') }, 1500)
              }}
            >
              {running === 'sync' ? '同步中…' : '开始同步'}
            </Button>
            <span className="text-[11px] text-muted-foreground">逐表拉取 → 建索引 → 原子替换;原库自动备份为 .bak,失败不影响原库。</span>
          </div>
          <StatusNote tone="warning" className="max-w-md font-mono">warning: table RD_M 的行数(1,204,882)超过预期上限,请确认服务器字典版本</StatusNote>
        </CardContent>
      </Card>

      <Card className="gap-3 py-4">
        <CardHeader className="px-4">
          <CardTitle className="text-xs">BDL 文档</CardTitle>
          <CardDescription className="text-[11px]">Genero BDL 语言文档的本地落点(随发行包不附带,指向你自己的文档目录)。</CardDescription>
        </CardHeader>
        <CardContent className="px-4">
          <Field className="max-w-md">
            <FieldLabel htmlFor="v-bdl">文档目录</FieldLabel>
            <Input id="v-bdl" defaultValue={APP_SETTINGS.bdlDoc} className="font-mono" onChange={() => mark('bdldoc')} />
            <div className="flex items-center gap-2">
              <FieldDescription>该目录存在于本机。</FieldDescription>
              <SaveBtn k="bdldoc" label="BDL 文档" {...{ dirty, saving, save }} />
            </div>
          </Field>
        </CardContent>
      </Card>
    </>
  )
}

function DebugSection({ dirty, saving, mark, save }: Omit<SecProps, 'onToast'>) {
  return (
    <>
      <Card className="gap-3 py-4">
        <CardHeader className="px-4">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <CardTitle className="text-xs">调试参数</CardTitle>
              <CardDescription className="mt-1 text-[11px]">本机调试会话的默认参数。清空某一项 = 恢复该项的内置默认值。</CardDescription>
            </div>
            <CardAction><SaveBtn k="debug" label="调试参数" {...{ dirty, saving, save }} /></CardAction>
          </div>
        </CardHeader>
        <CardContent className="space-y-4 px-4">
          <FieldGroup className="max-w-xl gap-4">
            <Field>
              <FieldLabel htmlFor="d-launch">启动参数模板</FieldLabel>
              <Input id="d-launch" defaultValue={DEBUG_PARAMS.launch} className="font-mono" onChange={() => mark('debug')} />
              <FieldDescription>{'{prog} 会替换为作业编号。清空 = 用内置默认。'}</FieldDescription>
            </Field>
            <Field>
              <FieldLabel htmlFor="d-watchdog">停站看门狗(秒)</FieldLabel>
              <Input id="d-watchdog" defaultValue={DEBUG_PARAMS.watchdog} onChange={() => mark('debug')} />
              <FieldDescription>停站后原地停留超过该时长就判定会话失联。</FieldDescription>
            </Field>
            <Field>
              <FieldLabel htmlFor="d-fgl">FGLServer</FieldLabel>
              <Input id="d-fgl" defaultValue={DEBUG_PARAMS.fglserver} placeholder="留空 = 自动" onChange={() => mark('debug')} />
              <FieldDescription>留空 = 由 T100 按 SSH 来源 IP 自动设置;自定义时才填。</FieldDescription>
            </Field>
            <Field>
              <FieldLabel htmlFor="d-print">print 数组元素上限</FieldLabel>
              <Input id="d-print" defaultValue={DEBUG_PARAMS.printLimit} onChange={() => mark('debug')} />
              <FieldDescription>fgldb 单次 print 的数组元素上限。清空 = 内置默认 1000。</FieldDescription>
            </Field>
            <Field>
              <FieldLabel>终端尺寸(宽 × 高)</FieldLabel>
              <div className="flex items-center gap-2">
                <Input defaultValue={DEBUG_PARAMS.termW} onChange={() => mark('debug')} />
                <span className="text-muted-foreground">×</span>
                <Input defaultValue={DEBUG_PARAMS.termH} onChange={() => mark('debug')} />
              </div>
              <FieldDescription>驱动 PTY 时用的终端大小,影响远端输出换行。</FieldDescription>
            </Field>
            <Field>
              <FieldLabel>断点持久化</FieldLabel>
              <Select defaultValue={DEBUG_PARAMS.bpPersist} onValueChange={() => mark('debug')}>
                <SelectTrigger className="w-full"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="default">跟随默认(开启)</SelectItem>
                  <SelectItem value="on">开启</SelectItem>
                  <SelectItem value="off">关闭</SelectItem>
                </SelectContent>
              </Select>
              <FieldDescription>三态不能用开关表达 —— 这是保留 Select 的原因。</FieldDescription>
            </Field>
          </FieldGroup>
        </CardContent>
      </Card>

      <Card className="gap-3 py-4">
        <CardHeader className="px-4">
          <CardTitle className="text-xs">默认环境</CardTitle>
          <CardDescription className="text-[11px] leading-4">
            调试会话未显式指定环境时用哪个。与「站点管理」的默认环境是两层:这里是调试工具的覆盖,留空即继承站点默认。
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3 px-4">
          <Field className="max-w-md">
            <FieldLabel>调试默认环境</FieldLabel>
            <Select defaultValue="__inherit__" onValueChange={() => mark('debug')}>
              <SelectTrigger className="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="__inherit__">继承站点默认(36正式)</SelectItem>
                <SelectItem value="35测试">35测试</SelectItem>
              </SelectContent>
            </Select>
          </Field>
          <StatusNote tone="warning" className="max-w-xl">
            注意:「调试参数」与「默认环境」两张卡改的是**同一个** debug 配置节,任一张点保存都会提交整节(后端的 PUT 是整节替换)。
            所以两处会一起落盘,不会互相覆盖。
          </StatusNote>
        </CardContent>
      </Card>
    </>
  )
}

function AppSection({ dirty, saving, mark, save, onToast }: SecProps) {
  const confirm = useConfirm()
  const { theme, setTheme } = useTheme()
  const [pathState, setPathState] = useState<'in' | 'out'>('out')
  const [busy, setBusy] = useState(false)

  return (
    <>
      <Card className="gap-3 py-4">
        <CardHeader className="px-4">
          <CardTitle className="text-xs">外观</CardTitle>
          <CardDescription className="text-[11px]">各视图共用同一份选择:这里切换后,整个界面都跟着变。</CardDescription>
        </CardHeader>
        <CardContent className="px-4">
          <ToggleGroup
            type="single"
            value={theme || 'dark'}
            onValueChange={(v) => { if (v) setTheme(v) }}
            variant="outline"
            className="max-w-md"
          >
            <ToggleGroupItem value="light" className="flex-1 gap-1.5 text-xs"><Sun className="size-3.5" /> 亮色</ToggleGroupItem>
            <ToggleGroupItem value="dark" className="flex-1 gap-1.5 text-xs"><Moon className="size-3.5" /> 暗色</ToggleGroupItem>
            <ToggleGroupItem value="system" className="flex-1 gap-1.5 text-xs"><Monitor className="size-3.5" /> 跟随系统</ToggleGroupItem>
          </ToggleGroup>
        </CardContent>
      </Card>

      <Card className="gap-3 py-4">
        <CardHeader className="px-4">
          <CardTitle className="text-xs">服务</CardTitle>
          <CardDescription className="text-[11px]">统一 Web 服务的监听地址 —— 这套 SPA(调试工作台)就挂在它上面。</CardDescription>
        </CardHeader>
        <CardContent className="px-4">
          <Field className="max-w-md">
            <FieldLabel htmlFor="v-listen">监听地址</FieldLabel>
            <Input id="v-listen" defaultValue={APP_SETTINGS.listen} className="font-mono" onChange={() => mark('app')} />
            <div className="flex items-center gap-2">
              <FieldDescription>改完需要重启 tt serve 才生效(监听只在启动时绑定一次)。端口被占用时会自动向后顺延。</FieldDescription>
              <SaveBtn k="app" label="服务" {...{ dirty, saving, save }} />
            </div>
          </Field>
        </CardContent>
      </Card>

      <Card className="gap-3 py-4">
        <CardHeader className="px-4">
          <CardTitle className="text-xs">命令行集成</CardTitle>
          <CardDescription className="text-[11px]">把 tt.exe 所在目录加入用户 PATH,之后在任意终端直接敲 tt。只改当前用户的环境变量,不需要管理员。</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3 px-4">
          <div className="flex flex-wrap items-center gap-2">
            {pathState === 'in' ? (
              <Button size="sm" variant="destructive" disabled={busy} onClick={() => { setBusy(true); window.setTimeout(() => { setBusy(false); setPathState('out'); onToast('已从用户 PATH 移除 D:\\Programs\\TT') }, 600) }}>
                <Trash2 className="size-3.5" /> {busy ? '处理中…' : '从 PATH 移除'}
              </Button>
            ) : (
              <Button size="sm" disabled={busy} onClick={() => { setBusy(true); window.setTimeout(() => { setBusy(false); setPathState('in'); onToast('已把 D:\\Programs\\TT 加入用户 PATH,新开的终端即可直接敲 tt') }, 600) }}>
                <Terminal className="size-3.5" /> {busy ? '处理中…' : '加入用户 PATH'}
              </Button>
            )}
            <span className="text-[11px] text-muted-foreground">
              {pathState === 'in' ? '已在用户 PATH 中' : '尚未加入 —— 加完之后新开的终端才能直接敲 tt'}
            </span>
          </div>
          <div className="flex items-baseline gap-3">
            <span className="w-28 shrink-0 text-[11px] text-muted-foreground">程序目录</span>
            <span className="min-w-0 flex-1 break-all font-mono text-[11px]">C:\Program Files\TT</span>
          </div>
        </CardContent>
      </Card>

      <Card className="gap-3 py-4">
        <CardHeader className="px-4"><CardTitle className="text-xs">运行信息</CardTitle></CardHeader>
        <CardContent className="px-4">
          {RUNTIME_INFO.map((r) => (
            <div key={r.label} className="flex items-baseline gap-3 border-b border-border/60 py-1.5 last:border-b-0">
              <span className="w-28 shrink-0 text-[11px] text-muted-foreground">{r.label}</span>
              <span className="min-w-0 flex-1 break-all font-mono text-[11px]">{r.value}</span>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card className="gap-3 py-4">
        <CardHeader className="px-4">
          <CardTitle className="text-xs">更新</CardTitle>
          <CardDescription className="text-[11px] leading-4">
            查 tt 自己有没有新版。查是你点才联网(服务不会替你轮询);装会先下载并校验 sha256,再交给一个独立的更新器进程去做。
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3 px-4">
          {UPDATE_INFO.map((r) => (
            <div key={r.label} className="flex items-baseline gap-3 border-b border-border/60 py-1.5 last:border-b-0">
              <span className="w-28 shrink-0 text-[11px] text-muted-foreground">{r.label}</span>
              <span className="min-w-0 flex-1 break-all text-[11px]">{r.value}</span>
            </div>
          ))}
          <div className="flex flex-wrap items-center gap-2">
            <Button size="sm" variant="outline" onClick={() => onToast('已是最新(0.2.3)')}><RefreshCw className="size-3.5" /> 检查更新</Button>
            <Button
              size="sm"
              onClick={async () => {
                const ok = await confirm({
                  title: '要现在升级 tt 吗?',
                  description: '下载校验完成后会停掉本服务并替换程序文件,页面随后会断开,请稍后重新打开。',
                  confirmText: '升级',
                })
                if (ok) onToast('更新器已接手(pid 10231)')
              }}
            >
              <Download className="size-3.5" /> 升级到 0.2.4
            </Button>
            <span className="text-[11px] text-muted-foreground">升级期间本页会断开,之后重新打开即可</span>
          </div>
        </CardContent>
      </Card>

      <Card className="gap-3 py-4">
        <CardHeader className="px-4">
          <CardTitle className="text-xs">缓存</CardTitle>
          <CardDescription className="text-[11px] leading-4">
            跑出来的可再生数据(企业快照、源码镜像、查询落盘、断点存档)。它们与 config.json 同目录,但性质相反 —— 清除只动这些,配置文件一律不碰。
          </CardDescription>
        </CardHeader>
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
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-[11px] text-muted-foreground">共 {CACHE_TOTAL.bytes} · {CACHE_TOTAL.files} 个文件</span>
            <Button
              size="sm"
              variant="outline"
              className="ml-auto"
              onClick={async () => {
                const ok = await confirm({
                  title: '清除缓存文件?',
                  description: `将删除 ${CACHE_TOTAL.files} 个可再生文件(共 ${CACHE_TOTAL.bytes})。config.json 与 .tt-serve.json 不在其列。`,
                  confirmText: '清除',
                })
                if (ok) onToast(`已清理 ${CACHE_TOTAL.files} 个文件,释放 ${CACHE_TOTAL.bytes}`)
              }}
            >
              <Trash2 className="size-3.5" /> 清除缓存
            </Button>
          </div>
          <FieldDescription>
            tt serve 启动时会自动清理超过 7 天的缓存 —— 刻意不"启动即全清":查询落盘的那些文件正是"刚才那条查询的完整结果"。
          </FieldDescription>
        </CardContent>
      </Card>
    </>
  )
}
