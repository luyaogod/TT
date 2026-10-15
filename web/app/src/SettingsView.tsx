// 统一设置页:左侧四个分区的导航,右侧按分区渲染卡片。
//
// 「站点管理」环境清单 + SSH/数据库(原调试页「环境」节与原字典页「环境配置」的并集,只留一份);
// 「数据字典」查询数据源 / 源码镜像 / 数据同步 / BDL 文档;
// 「DEBUG」调试参数 / 默认环境;
// 「应用设置」外观 / 服务 / 运行信息。
//
// 环境清单走**共享**端点 /api/hosts(顶层,不带 API_BASE 前缀):一个进程里它是唯一的数据源
// —— config.json 的 hosts 节。合并前这份清单藏在 debug 节里
// (所以老代码收发的是整个 debug 节点),现在收发的是 hosts 节。
//
// 保存**按节提交**:设置页的各张卡片各改各的节,不拿陈旧快照覆盖对方刚改好的节
// (后端约定:PUT /api/hosts 里省略的节保持原样)。
//
// ⚠ 后端的 PUT 是**整节替换**,而这些结构体每个字段都带 omitempty —— 省略的键等于删除。
// 所以每张卡片的保存都必须提交**完整的节**(从读到的快照 + 编辑项构造),绝不能只发表单里
// 那几个字段:老代码在「高级」里只发了 5 个键,于是 fglserver / persistBreakpoints /
// activeEnv 在用户每次保存时被静默抹掉。debugPatchOf 就是为了堵这个。
import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react'
import {
  AlertCircle, CheckCircle2, Copy, Download, Eye, EyeOff,
  Monitor, Plus, RefreshCw, Star, Sun, Moon,
  Terminal, Trash2, Zap,
} from 'lucide-react'
import {
  api, type ConfigMeta, type ConfigStatus, type DBSyncJob, type DBSyncResp,
  type HostsDb, type HostsPatch, type HostsSsh, type HostsView, type MirrorJob, type MirrorResp,
  type UpdateState,
} from './api'
import {
  Button, Input, Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '../../shared/ui'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
  Progress,
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
  Tabs, TabsContent, TabsList, TabsTrigger,
} from '../../shared/ui-radix'
import { Card, Field, InfoRow, SectionTitle, SettingRow } from '../../shared/settings'
import { useStore, type ThemeMode } from './store'
import { cn } from '../../shared/utils'
import { SETTINGS_SECTIONS, settingsHash } from './routing'

// ---------- 表单模型 ----------

interface SshDb {
  type: string // oracle | kingbase
  host: string
  port: number
  service: string // oracle: SERVICE_NAME
  database: string // kingbase: 库名
  accounts: { account: string; password: string }[]
  // 是否允许 AI 执行只读 SQL(排查业务数据)。undefined = 未配置 = 开启;
  // 显式 false 才关。**载入/保存都必须带上它** —— 否则在设置页点一次保存,
  // 这个开关就被静默抹掉、悄悄变回开启(我实现时实测过)。
  readonlySql?: boolean
}

interface SshItem {
  name: string
  host: string
  port: number
  user: string
  password: string
  zone: string
  topent: string
  // workspace 该环境的 .tzs 工作区（含 mta/ 的目录，绝对路径）。
  // 与 launchArgs / watchdogSeconds 不同：它是**必填**的（引擎的默认工作区是真实客户目录），
  // 而且只有环境级这一份 —— 切环境就该跟着换工作区。
  workspace: string
  db: SshDb | null
}

// debug 节的完整表单。字段与 internal/config/schema.go 的 DebugSettings 一一对应 ——
// 少一个键就意味着保存时把它删了(整节替换),所以这里必须是全集。
interface DebugForm {
  launchArgs: string
  watchdogSeconds: number
  fglserver: string
  termWidth: number
  termHeight: number
  printElements: number
  // *bool 三态:nil = 跟随默认(开启)。复选框写不了三态,所以用下拉。
  persistBreakpoints: 'default' | 'on' | 'off'
}

interface DictForm {
  /** '' = 在线(默认环境) | 'local' = 本地 SQLite。刻意不留环境名那一档,见「查询数据源」卡片的注释。 */
  querySource: string
  mirrorDir: string
  bdldocDir: string
  // 刻意没有 syncTarget:字典库的落点不让用户在这一页改(它只是一个本地缓存文件,
  // 位置由服务端算)。位置只在卡片里只读显示,取值来自 /api/config/status。
  // **不要**为了"完整性"把它加回来再原样回传 —— 那等于把一个不展示的键交给这一页写。
}

/** 可保存的配置节。脏标记与保存按钮都按它划分。 */
type ConfigKey = 'hosts' | 'debug' | 'listen' | 'query' | 'mirror' | 'bdldoc'

// 主题切换卡片(shadcn 主题切换卡样式):三张卡各带一张迷你界面预览,选中卡描边+底色高亮。
const THEME_OPTIONS: { value: ThemeMode; label: string; icon: typeof Sun }[] = [
  { value: 'light', label: '亮色', icon: Sun },
  { value: 'dark', label: '暗色', icon: Moon },
  { value: 'system', label: '跟随系统', icon: Monitor },
]

// 外观选择:三张卡片,每张只有图标 + 名称。
//
// 原先每张卡里还画了一个"迷你界面"缩略图(左栏条 + 内容块,用被预览主题的配色)。
// 删掉的理由:那点 40px 高的示意既画不清主题差别(靠的是三块色块),又占掉了卡片
// 一半的高度;而"亮色/暗色/跟随系统"这三个词本身没有任何歧义,图标也一眼能认。
// 简化之后一行内三张卡更紧凑,选中的那档也只用一个边框 + 底色表达。
function ThemeCards({ value, onChange }: { value: ThemeMode; onChange: (t: ThemeMode) => void }) {
  return (
    <div className="grid max-w-md grid-cols-3 gap-2">
      {THEME_OPTIONS.map((o) => {
        const Icon = o.icon
        const active = value === o.value
        return (
          <button
            key={o.value}
            type="button"
            aria-pressed={active}
            title={`切换到「${o.label}」`}
            onClick={() => onChange(o.value)}
            className={cn('flex items-center justify-center gap-1.5 rounded-ui border px-3 py-2 text-xs transition-colors',
              active ? 'border-foreground/50 bg-accent/60 text-foreground' : 'border-border text-muted-foreground hover:bg-accent/40 hover:text-foreground')}
          >
            <Icon className="h-3.5 w-3.5 shrink-0" />
            {o.label}
          </button>
        )
      })}
    </div>
  )
}

/**
 * Go 的 time.Duration.String() → 秒数(后端把耗时按这个格式直接下发:
 * "2.1s" / "1m20.5s" / "1h2m3s")。解析不出来就返回 0,调用方据此不显示速度。
 */
function durSeconds(s: string): number {
  if (!s) return 0
  const re = /(\d+(?:\.\d+)?)(h|m|s)/g
  let total = 0, found = false, m: RegExpExecArray | null
  while ((m = re.exec(s))) {
    found = true
    const n = Number(m[1])
    total += m[2] === 'h' ? n * 3600 : m[2] === 'm' ? n * 60 : n
  }
  return found ? total : 0
}

// 镜像任务的数值行:**裸值**,从左到右按"先看什么"排 —— 大小(当前/总量)、速度、耗时、文件数。
// 速度是算出来的(bytes / elapsed):后端只下发耗时字符串,不为一个展示值再加 JSON 字段。
// 算不出来时(耗时为 0、刚开始)省掉那一格,不显示 "0 KB/s" 这种噪音。
function mirrorCellTexts(j: MirrorJob): ReactNode[] {
  const secs = durSeconds(j.elapsed)
  const rate = secs > 0 ? j.bytes / secs : 0
  return [
    fmtBytes(j.bytes) + (j.total > 0 ? ` / ${fmtBytes(j.total)}` : ''),
    rate > 0 ? `${(rate / 1024).toFixed(0)} KB/s` : '',
    j.elapsed ? (secs > 0 ? `${secs.toFixed(1)}s` : j.elapsed) : '',
    j.files > 0 ? `${j.files} 个文件` : '',
  ].filter(Boolean)
}

/**
 * 长跑任务的进度块(镜像拉取 / 字典同步共用)。
 *
 * 两类任务的"进度量"不是同一种东西:镜像按**字节**走,字典同步按**表**走。
 * 所以进度条只认一个 pct(由调用方按各自的量算好),下面那行数值由 pct 规则无关的
 * cells 描述符给出 —— 想显示什么就写什么,不必让这个组件认识两种任务的字段。
 *
 * 数值行刻意**不带描述性标签**("已传输:" / "文件数:" / "网络速度:" 这类):
 * 单位本身已经把含义说清(B/s 是速度、`a / b` 是总量、后面带 s 是秒数),
 * 再给每个数配个词只会把一行塞满,真正要看的数字被埋进文字里。
 */
function JobProgress({ env, phaseLabel, tag, pct, cells, error, message, stopping, onStop }: {
  env: string; phaseLabel: string
  /** 右上角的小标记(全量/增量、进行中/已结束) */
  tag?: string
  /** 0-100。算不出来(如刚开始、总量未知)时传 undefined,进度条显示为 0 */
  pct?: number
  /** 数值行:**裸值,不写前缀标签**。按"先看什么"从左到右排 */
  cells: ReactNode[]
  error?: string; message?: string
  /** 已经在停但还没停完(后端取消是异步的:SFTP 读到下一个分片才返回) */
  stopping?: boolean
  /** 有它才显示「停止」。取消后不该只是把按钮变灰 —— 要点得动,并由 stopping 反馈 */
  onStop?: () => void
}) {
  const failed = !!error
  const v = Math.max(0, Math.min(100, pct ?? 0))
  return (
    <div className="border border-border bg-muted/20 p-2">
      <div className="mb-1.5 flex items-center gap-2 text-xs">
        <span className="font-medium">{env}</span>
        <span className="text-muted-foreground">{phaseLabel}</span>
        {onStop && (
          // 停止放在这一行的最右:它作用的对象就是这条进度(误点「全量重建」后想收手)
          <Button variant="outline" size="sm" className="ml-auto h-6 px-2 text-[11px]"
            disabled={stopping} onClick={onStop}>
            {stopping ? '正在停止…' : '停止'}
          </Button>
        )}
        {tag && <span className={cn('text-[11px] text-muted-foreground', !onStop && 'ml-auto')}>{tag}</span>}
      </div>
      {/* 失败时把填充改成 destructive:进度条是这一块里唯一的视觉主体,
          出错还留着主色填充会让人以为还在正常跑 */}
      <Progress value={v} barClassName={failed ? 'bg-destructive' : undefined} />
      <div className="mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-muted-foreground">
        <span className="font-mono text-foreground">{v}%</span>
        {cells.map((c, i) => <span key={i} className="font-mono">{c}</span>)}
      </div>
      {message && (
        <p className={cn('mt-1.5 flex items-center gap-1 text-xs',
          failed ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400')}>
          {failed ? <AlertCircle className="h-3.5 w-3.5" /> : <CheckCircle2 className="h-3.5 w-3.5" />}
          {error || message}
        </p>
      )}
    </div>
  )
}

// 阶段标签与字节格式化:原先在字典页的两个视图里,随功能一起搬进来。
const MIRROR_PHASE: Record<string, string> = {
  connect: '连接服务器…', probe: '探测 T100 目录…', pack: '服务器打包…',
  download: '下载并解压…', done: '完成', error: '失败',
}
const SYNC_PHASE: Record<string, string> = {
  open: '连接远程数据库…', table: '拉取数据表…', index: '创建主键索引…',
  replace: '替换本地数据库…', done: '完成', error: '失败',
}
/**
 * 表格里的「状态」格:同一时刻只可能有一个环境在跑(后端一次只跑一个动作),
 * 所以正在跑的那个显示阶段与进度,其余显示空闲词(已有镜像 / 未拉取…) ——
 * 于是"哪个环境在忙、忙到哪一步"一眼能落在具体那一行上,不用去别处找进度条。
 */
function EnvStatus({ busy, running, idle, error }: {
  busy: boolean; running: boolean; idle: string; error?: string
}) {
  if (!busy) return <span className="text-[11px] text-muted-foreground">{idle}</span>
  return (
    <span className={cn('flex items-center gap-1 text-[11px]',
      error ? 'text-red-600 dark:text-red-400' : 'text-foreground')}>
      {running
        ? <RefreshCw className="h-3 w-3 shrink-0 animate-spin" />
        : error
          ? <AlertCircle className="h-3 w-3 shrink-0" />
          : <CheckCircle2 className="h-3 w-3 shrink-0 text-emerald-600 dark:text-emerald-400" />}
      {error || idle}
    </span>
  )
}
function fmtBytes(n: number): string {
  if (!n) return '0 B'
  if (n >= 1 << 30) return (n / (1 << 30)).toFixed(2) + ' GB'
  if (n >= 1 << 20) return (n / (1 << 20)).toFixed(1) + ' MB'
  if (n >= 1 << 10) return (n / (1 << 10)).toFixed(1) + ' KB'
  return n + ' B'
}

// 设置控件的统一尺寸。w-full 是必须的:控件都住在「说明 + 控件」这一列里,
// 不写宽度时它们靠 flex 默认的 align-items:stretch 撑满,而 SelectTrigger 自带 w-full、
// Input 却只有 hidden 的块级宽度 —— 两种来源不一样,换一次外层布局就会一部分宽一部分窄。
// 显式写死 w-full,控件宽度就只由那一列决定。
const input = 'h-7 w-full min-w-0 text-xs'
// 单元格「正在编辑」提示环:用 ring 语义 token,不写死颜色(共享层里也只有这一份)
const CELL_EDITING = 'focus-within:ring-1 focus-within:ring-inset focus-within:ring-ring/60'

const blankDb = (): SshDb => ({ type: 'oracle', host: '', port: 1521, service: 'topprd', database: '', accounts: [] })
// 新增环境默认带三个常用账号(账号=密码),省去手工录入
const defaultAccounts = () => ['ds', 'dsdata', 'dsdemo'].map((a) => ({ account: a, password: a }))
// 新增环境用的空档:登录区域默认 1、服务名默认 topprd —— 都是现场最常见的取值,
// 先填上能省掉每次手敲(不对再改)。注意 service 是**真值**不是占位符:
// placeholder 在框里是灰字、不会被提交,而这里要的是"保存下去就是它"。
const blankSsh = (): SshItem => ({
  name: '', host: '', port: 22, user: '', password: '', zone: '1', topent: '', workspace: '',
  db: { ...blankDb(), accounts: defaultAccounts() },
})

// ---------- 表单 ↔ 配置节的互转 ----------

/**
 * 环境清单 → hosts 节。**整节提交**,所以这里必须给出完整的 sshs 列表。
 *
 * 仍未覆盖的字段:`hosts.sshs[].launchArgs` / `watchdogSeconds`(每环境的调试覆盖项,
 * 本页不暴露)。它们由服务端的保留机制补回(见 internal/web 的 preserveUngovernedDBFields)。
 * 而 `workspace` 是这一页管得着的(表单里有输入框) —— 不要把它当成“未覆盖”加进那份保留名单,
 * 否则用户清空它时会被静默补回。
 */
function hostsPatchOf(list: SshItem[], prevActiveEnv: string) {
  const sshsOut: HostsSsh[] = list.filter((x) => x.host).map((x) => {
    const o: HostsSsh = {
      name: x.name || `${x.host}-${x.zone}`.replace(/-$/, ''),
      host: x.host, port: x.port || 22, user: x.user, password: x.password,
    }
    if (x.zone.trim()) o.zone = x.zone.trim()
    if (x.topent.trim()) o.topent = x.topent.trim()
    // 工作区：留空就是没配（不写这个键），与服务端“环境级没有兜底”一致。
    if (x.workspace.trim()) o.workspace = x.workspace.trim()
    if (x.db) {
      const db: HostsDb = { type: x.db.type || 'oracle', host: x.db.host.trim(), port: x.db.port || 0 }
      if (x.db.type === 'oracle') { if (x.db.service.trim()) db.service = x.db.service.trim() }
      else { if (x.db.database.trim()) db.database = x.db.database.trim() }
      const accounts = x.db.accounts.map((a) => ({ account: a.account.trim(), password: a.password })).filter((a) => a.account)
      if (accounts.length) db.accounts = accounts
      // 只有"显式关掉"才写:没配 = 默认开启,不必往 JSON 里塞一堆 true
      if (x.db.readonlySql === false) db.readonlySql = false
      o.db = db
    }
    return o
  })
  // 默认环境本页只在「环境清单」里改它,但 PUT 提交的是**整节**,必须原样回传:
  // 若它刚好被这次删除带走了,就顺延到首条 —— 否则后端会以「默认环境不在环境列表里」整次 400 拒掉。
  return {
    activeEnv: sshsOut.some((e) => e.name === prevActiveEnv) ? prevActiveEnv : (sshsOut[0]?.name || ''),
    sshs: sshsOut,
  }
}

/**
 * debug 节 → 完整节对象。
 *
 * **这里是那个数据丢失 bug 的修法**:不按表单里"用户改过哪些"挑选,而是把 debug 节的
 * 每个键都给出来。老代码只发 5 个键,于是 fglserver / persistBreakpoints / activeEnv
 * 每次保存都被删掉。
 *
 * ⚠ 后端的 PUT 是**整节替换**(internal/web 的 putSection:`root["debug"] = sec`),
 * 而 DebugSettings 每个字段都是 omitempty —— **省略一个键 = 从 config.json 里删掉它**。
 * 而 debug 节**没有** hosts 那样的"保留未管理键"机制(那条只罩 hosts.sshs[].db)。
 *
 * 所以下面几个"界面已经不显示"的键仍然要**原样回传**:它们只是不再提供编辑入口,
 * 不是不再存在。不回传的话,用户在设置页点一次保存,自己手工写进 config.json 的
 * 自定义启动参数 / 终端尺寸就被静默抹掉了 —— 正是这个文件开头那段注释里描述的同一类 bug。
 */
function debugPatchOf(d: DebugForm): NonNullable<HostsPatch['debug']> {
  const out: NonNullable<HostsPatch['debug']> = {}
  // 只有看门狗还在界面上(见「调试参数」卡片)
  if (d.watchdogSeconds) out.watchdogSeconds = d.watchdogSeconds
  // 以下五项**界面已移除**(一律跟内置默认走,用户不需要改),但按上面的理由原样回传。
  // 空值仍然省略:omitempty 语义下"省略 = 用默认",与"写一个空字符串"不是一回事。
  if (d.launchArgs) out.launchArgs = d.launchArgs
  if (d.fglserver.trim()) out.fglserver = d.fglserver.trim()
  if (d.termWidth) out.termWidth = d.termWidth
  if (d.termHeight) out.termHeight = d.termHeight
  if (d.printElements) out.printElements = d.printElements
  if (d.persistBreakpoints === 'on') out.persistBreakpoints = true
  else if (d.persistBreakpoints === 'off') out.persistBreakpoints = false
  // activeEnv **不再回传**:「默认环境」那张卡已整张移除,它属于"用户不需要改"的那一类,
  // 省略即回落到 hosts.activeEnv(站点管理里那个绿点),这正是想要的行为。
  return out
}

export function SettingsView() {
  const section = useStore((s) => s.settingsSection)
  const setSection = useStore((s) => s.setSettingsSection)
  const theme = useStore((s) => s.theme)
  const setTheme = useStore((s) => s.setTheme)

  const [envTab, setEnvTab] = useState<'ssh' | 'db'>('ssh')
  const [cfg, setCfg] = useState<HostsView | null>(null)
  const [sshs, setSshs] = useState<SshItem[]>([])
  const [dbg, setDbg] = useState<DebugForm>({
    launchArgs: '', watchdogSeconds: 0, fglserver: '',
    termWidth: 200, termHeight: 50, printElements: 1000, persistBreakpoints: 'default',
  })
  const [dict, setDict] = useState<DictForm>({ querySource: '', mirrorDir: '', bdldocDir: '' })
  const [listen, setListen] = useState('')
  // 派生状态(路径存不存在只有服务端算得出来)与配置元信息,以及 PATH 安装状态
  const [status, setStatus] = useState<ConfigStatus | null>(null)
  const [meta, setMeta] = useState<ConfigMeta | null>(null)
  const [installBusy, setInstallBusy] = useState(false)
  const [installNote, setInstallNote] = useState('')
  const [cacheBusy, setCacheBusy] = useState(false)
  const [cacheNote, setCacheNote] = useState('')
  // tt 自身的更新:状态靠 /api/update 读(不联网),查/装都是人去点才动作
  const [upd, setUpd] = useState<UpdateState | null>(null)
  const [updBusy, setUpdBusy] = useState('')
  const [updNote, setUpdNote] = useState<{ kind: 'ok' | 'err'; text: string } | null>(null)
  // 源码镜像 / 字典同步的运行态:它们是**动作**(长跑任务 + 进度),不再是独立页面,
  // 就住在「数据字典」分区的对应卡片里。目标环境不再是这里的一份 state ——
  // 它由表格里被点的那一行给出(原先那份 state 只是"下拉里选中的那个",改成表格后没有了)。
  const [mirror, setMirror] = useState<MirrorResp | null>(null)
  const [sync, setSync] = useState<DBSyncResp | null>(null)
  const [opBusy, setOpBusy] = useState('')
  // 源码镜像 / 字典同步的失败提示。**没有成功态** —— 成功由进度块自己表达(见 doMirror 的注释)。
  const [opNote, setOpNote] = useState<{ kind: 'err'; text: string } | null>(null)

  // 需要二次确认的动作(全量重建 / 开始同步 / 升级 tt)。
  // 这几个动作原先用 window.confirm:它是浏览器原生弹窗,样式与本界面无关、
  // 在 iframe/无头环境里还会被直接拦掉(拦掉时返回值按"取消"算,点了没反应还没提示)。
  // 统一改成与「删除环境」同一套 AlertDialog:文案里带上会被影响的环境与路径,
  // 确认按钮的文字按动作给(不是笼统的"确定")。
  const [ask, setAsk] = useState<{
    title: string; desc: string; confirmText: string; run: () => void
  } | null>(null)

  const [selSsh, setSelSsh] = useState(0)
  const [err, setErr] = useState('')
  const [saveState, setSaveState] = useState<'idle' | 'saving' | 'saved'>('idle')
  // 脏标记按**配置节**记。保存时只提交改过的节 —— 设置页的各张卡片共用 /api/hosts,
  // 把没改的节一起发过去就是用陈旧快照覆盖别张卡片刚改好的节。
  const dirtyRef = useRef<Record<ConfigKey, boolean>>({
    hosts: false, debug: false, listen: false, query: false, mirror: false, bdldoc: false,
  })
  const [dirty, setDirty] = useState(false)
  const markDirty = (...keys: ConfigKey[]) => {
    keys.forEach((k) => { dirtyRef.current[k] = true })
    setDirty(true)
  }
  const cleanDirty = (keys?: ConfigKey[]) => {
    if (keys) keys.forEach((k) => { dirtyRef.current[k] = false })
    else (Object.keys(dirtyRef.current) as ConfigKey[]).forEach((k) => { dirtyRef.current[k] = false })
    setDirty((Object.keys(dirtyRef.current) as ConfigKey[]).some((k) => dirtyRef.current[k]))
  }
  const isDirty = (...keys: ConfigKey[]) => keys.some((k) => dirtyRef.current[k])

  const [showPwd, setShowPwd] = useState(false)
  // 数据库页操作状态
  const [busy, setBusy] = useState<string>('')
  const [note, setNote] = useState<{ kind: 'ok' | 'err' | 'info'; text: string } | null>(null)
  const [accProbe, setAccProbe] = useState<{ i: number; state: 'testing' | 'ok' | 'err'; msg: string } | null>(null)
  const [addAcct, setAddAcct] = useState({ account: '', password: '' })
  const [delTarget, setDelTarget] = useState<number | null>(null) // 待删除环境的下标(弹窗确认)

  const sshName = (e: SshItem) => e.name || `${e.host}-${e.zone}`.replace(/-$/, '')

  const loadSettings = useCallback(() => {
    api.settings().then((c) => {
      setCfg(c)
      setSshs((c.sshs || []).map((e) => {
        const d = e.db
        return {
          name: e.name || '', host: e.host || '', port: e.port || 22, user: e.user || '', password: e.password || '',
          zone: e.zone || '', topent: e.topent != null ? String(e.topent) : '',
          workspace: e.workspace || '',
          db: d ? {
            type: d.type || 'oracle', host: d.host || '', port: d.port || 0,
            service: d.service || '', database: d.database || '',
            accounts: (d.accounts || []).map((a) => ({ account: a.account || '', password: a.password || '' })),
            readonlySql: d.readonlySql === false ? false : undefined,
          } : null,
        }
      }))
      setDbg({
        // 这几项界面已不显示,但仍然要读进来:保存时原样回传,避免整节替换把它们删掉
        launchArgs: c.debug?.launchArgs || '',
        watchdogSeconds: c.debug?.watchdogSeconds || 0,
        fglserver: c.debug?.fglserver || '',
        termWidth: c.debug?.termWidth || 200,
        termHeight: c.debug?.termHeight || 50,
        printElements: c.debug?.printElements || 1000,
        persistBreakpoints: c.debug?.persistBreakpoints === true ? 'on'
          : c.debug?.persistBreakpoints === false ? 'off' : 'default',
      })
      setDict({
        querySource: c.query?.source || '',
        mirrorDir: c.mirror?.dir || '',
        bdldocDir: c.bdldoc?.dir || '',
      })
      setListen(c.listen || '')
      cleanDirty(); setSaveState('idle'); setErr('')
    }).catch((e) => setErr(e.message))
    // 派生状态与元信息:取不到不影响主流程(它们在卡片里只是提示)
    void api.configStatus().then(setStatus).catch(() => { /* 静默:提示性信息 */ })
    void api.configMeta().then(setMeta).catch(() => { /* 静默:提示性信息 */ })
    // 更新状态同样是不联网的读端点(检查结论来自缓存)
    void api.updateState().then(setUpd).catch(() => { /* 静默:提示性信息 */ })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => { loadSettings() }, [loadSettings])
  // 本页被 keep-alive 常驻挂载(切走只是 display:none),所以**不能轮询** ——
  // 改为"每次切回本页、且没有未保存修改时"回读一次,拿别处(命令行)改过的值。
  const dirtyRefAny = () => (Object.keys(dirtyRef.current) as ConfigKey[]).some((k) => dirtyRef.current[k])
  const visibleRef = useRef(false)
  const view = useStore((s) => s.view)
  useEffect(() => {
    const visible = view === 'settings'
    if (visible && !visibleRef.current && !dirtyRefAny()) void loadSettings()
    visibleRef.current = visible
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [view])

  // 分区变化 → 同步到 URL 片段(用 replace 而不是 push:否则浏览器后退会一格一格走遍所有分区)
  useEffect(() => {
    const want = settingsHash(section)
    if (window.location.hash !== want) window.history.replaceState(null, '', want)
  }, [section])

  // ---- 统一保存:按节提交;keys 指明这次要提交哪些节 ----
  const save = async (keys: ConfigKey[], list?: SshItem[]) => {
    if (saveState === 'saving' || !cfg) return
    const body: HostsPatch = {}
    if (keys.includes('hosts')) body.hosts = hostsPatchOf(list || sshs, cfg.activeEnv)
    if (keys.includes('debug')) body.debug = debugPatchOf(dbg)
    if (keys.includes('listen')) body.listen = listen.trim()
    if (keys.includes('query')) body.query = { source: dict.querySource }
    if (keys.includes('mirror')) body.mirror = { dir: dict.mirrorDir.trim() }
    // 刻意不提交 sync 节:字典库的落点这一页不可编辑,交出去只会把服务端算好的值覆盖掉。
    // 后端约定"省略的节保持原样",所以不写它才是保留原值。
    if (keys.includes('bdldoc')) body.bdldoc = { dir: dict.bdldocDir.trim() }
    setErr(''); setSaveState('saving')
    try {
      const next = await api.saveSettings(body)
      setCfg(next)
      setSshs((prev) => prev.map((x) => ({ ...x, name: sshName(x) })))
      cleanDirty(keys); setSaveState('saved')
      // 路径型取值的 exists 是服务端算的,保存后要回读一次才对得上新值
      void api.configStatus().then(setStatus).catch(() => { /* 提示性信息,取不到不影响 */ })
    } catch (ex: any) {
      // 服务端是校验的权威(环境名重复 / 端口越界 / 至少要有一个环境…),
      // 它给的中文说明原样显示,不要用本地判断把它盖掉
      setErr(ex.message || String(ex)); setSaveState('idle')
    }
  }

  // ---- 环境编辑 ----
  const patchSsh = (i: number, patch: Partial<SshItem>) => {
    const next = sshs.map((x, j) => (j === i ? { ...x, ...patch } : x))
    setSshs(next); markDirty('hosts')
  }
  const patchDb = (i: number, patch: Partial<SshDb>) => {
    const cur = sshs[i]
    if (!cur) return
    patchSsh(i, { db: cur.db ? { ...cur.db, ...patch } : { ...blankDb(), ...patch } })
  }
  const patchAcct = (i: number, j: number, patch: Partial<{ account: string; password: string }>) => {
    const cur = sshs[i]
    if (!cur?.db) return
    patchDb(i, { accounts: cur.db.accounts.map((a, k) => (k === j ? { ...a, ...patch } : a)) })
  }
  const delAcct = (i: number, j: number) => {
    const cur = sshs[i]
    if (!cur?.db) return
    patchDb(i, { accounts: cur.db.accounts.filter((_, k) => k !== j) })
  }
  const pushAcct = () => {
    const cur = sshs[selSsh]
    if (!cur?.db) return
    const account = addAcct.account.trim()
    if (!account) return
    patchDb(selSsh, { accounts: [...cur.db.accounts, { account, password: addAcct.password }] })
    setAddAcct({ account: '', password: '' })
  }
  const addSsh = () => { setSelSsh(sshs.length); setSshs([...sshs, blankSsh()]); markDirty('hosts') }
  /**
   * 复制当前环境。
   *
   * 用途是"同一套库、换一台服务器":SSH 四元组与库连接整份抄过来,改个地址就能用,
   * 比从空白重敲一遍省事。所以这里抄的是**当前正在编辑的那份 state**(含还没保存的改动),
   * 不是 config.json 里的旧值 —— 否则"改完→复制"会把刚改的丢掉。
   *
   * 名称必须当场定下来:后端按名字认环境、重名直接 400。而截断显示名是按
   * `主机-区域` 兜底算的,空名字的两条会撞成同一个键 —— 所以这里始终写一个显式名字
   * (原名/显示名 + -copy,撞了就再往后取 -copy2、-copy3)。
   *
   * 刻意**不**顺手把副本设为默认:复制多半是为了试一套新参数,把线上默认环境
   * 悄悄挪到未验证的副本上,代价比多点一下「设为默认」大得多。两个动作分开。
   */
  const dupSsh = (i: number) => {
    const src = sshs[i]
    if (!src) return
    const base = (src.name || sshName(src) || '环境') + '-copy'
    const taken = new Set(sshs.map((x) => sshName(x)))
    let name = base
    for (let n = 2; taken.has(name); n++) name = `${base}${n}`
    // 深拷贝:db 与 accounts 都是嵌套对象,浅拷贝会让两个环境共用同一个账号数组,
    // 改一处两处都变(而且 dirty 标记只认 hosts 节,这种串改极难排查)
    const item: SshItem = {
      ...src, name,
      db: src.db ? { ...src.db, accounts: src.db.accounts.map((a) => ({ ...a })) } : null,
    }
    setSshs([...sshs, item])
    setSelSsh(sshs.length)
    markDirty('hosts')
  }
  const setDefaultEnv = (name: string) => {
    if (!cfg || name === cfg.activeEnv) return
    setCfg({ ...cfg, activeEnv: name })
    void save(['hosts'])
  }
  // 删除环境:先弹窗二次确认(误删代价高且不可撤销),确定后立即落盘保存
  const askDelSsh = () => { if (sshs[selSsh]) setDelTarget(selSsh) }
  const confirmDelSsh = async () => {
    const i = delTarget
    setDelTarget(null)
    if (i === null || !sshs[i]) return
    const next = sshs.filter((_, j) => j !== i)
    setSelSsh(Math.max(0, Math.min(i, next.length - 1)))
    setSshs(next)
    await save(['hosts'], next) // 点击「确定」= 删除并直接保存,不再需要手动点保存
  }

  // ---- 数据库页操作(针对当前环境的 db) ----
  // 从服务器获取:登录当前环境的 SSH 自动探测连接要素回填(辅助)
  const fetchFromServer = async () => {
    const s = sshs[selSsh]
    const d = s?.db
    if (!s?.host || !s?.user) { setNote({ kind: 'err', text: '请先填写 SSH 主机与账号' }); return }
    if (!d) return
    setBusy('fetch'); setNote(null)
    try {
      const r = await api.probeDB({ host: s.host, port: s.port || 22, user: s.user, password: s.password, zone: s.zone, type: d.type || 'oracle' })
      const patch: Partial<SshDb> = {}
      // 主机地址:服务器解析出的是 IP 就直接用;是服务器内部主机名(客户端多半不可达)则回填 SSH 主机
      // —— T100 的库通常与应用服务器同机,SSH 能连上就意味着该地址客户端可达。
      const probedHost = (r.host || '').trim()
      const isIP = /^\d{1,3}(\.\d{1,3}){3}$/.test(probedHost)
      const hostVal = (isIP ? probedHost : '') || (s.host || '').trim()
      if (hostVal) patch.host = hostVal
      if (r.type === 'kingbase') {
        if (r.database) patch.database = r.database
        if (r.port) patch.port = r.port
        setNote(r.note
          ? { kind: 'info', text: r.note }
          : { kind: 'ok', text: `获取成功:金仓 ${r.database || '?'} @ ${hostVal || '?'}:${r.port || '?'}(主机地址按 SSH 主机回填,如需从别处连库请手工调整)` })
      } else {
        if (r.service || r.tns) patch.service = r.service || r.tns || ''
        if (r.port) patch.port = r.port
        const inner = probedHost && !isIP ? `(服务器内部名 ${probedHost} 客户端多半不可达,已改用 SSH 主机)` : ''
        setNote({
          kind: r.note ? 'err' : 'info',
          text: (r.note ? r.note + ';' : '') +
            `服务器解析地址 ${probedHost || '未知'}:${r.port || '?'}(service ${r.service || r.tns || '?'});主机地址已回填 ${hostVal || '?'}${inner}`,
        })
      }
      if (Object.keys(patch).length) patchDb(selSsh, patch)
    } catch (ex: any) {
      setNote({ kind: 'err', text: '获取失败: ' + (ex.message || String(ex)) })
    } finally { setBusy('') }
  }
  // 附加账号行验证:服务器侧以 账号+密码 连当前环境的显式目标库
  const verifyAcct = async (j: number) => {
    const i = selSsh
    const s = sshs[i]
    const d = s?.db
    const row = d?.accounts?.[j]
    if (!s?.host || !s?.user) { setAccProbe({ i: j, state: 'err', msg: '请先填写 SSH 主机与账号' }); return }
    if (!d?.host.trim()) { setAccProbe({ i: j, state: 'err', msg: '请先填写数据库主机地址' }); return }
    if (!row?.account.trim()) { setAccProbe({ i: j, state: 'err', msg: '账号不能为空' }); return }
    setAccProbe({ i: j, state: 'testing', msg: '' })
    try {
      await api.dbAccVerify({
        host: s.host, port: s.port || 22, user: s.user, password: s.password, zone: s.zone,
        type: d.type || 'oracle', account: row.account.trim(), acctPassword: row.password,
        dbHost: d.host.trim(), dbPort: d.port || 0,
        ...(d.type === 'kingbase' ? { dbDatabase: d.database.trim() } : { dbSvc: d.service.trim() }),
      })
      setAccProbe({ i: j, state: 'ok', msg: '连接正常 ✓' })
    } catch (ex: any) {
      setAccProbe({ i: j, state: 'err', msg: ex.message || String(ex) })
    }
  }
  // 客户端直连测试:验证「本机 → 库」这条链路真的通(与服务器侧验证互补)
  const testConn = async () => {
    const d = sshs[selSsh]?.db
    if (!d?.host.trim()) { setNote({ kind: 'err', text: '请先填写数据库主机地址' }); return }
    setBusy('conn'); setNote(null)
    try {
      const r = await api.connTest({
        type: d.type || 'oracle', host: d.host.trim(), port: d.port || 0,
        service: d.service.trim(), database: d.database.trim(),
        accounts: d.accounts.map((a) => ({ account: a.account.trim(), password: a.password })).filter((a) => a.account),
      })
      const ver = (r.serverVersion || '').split('\n')[0]
      setNote({ kind: r.ok ? 'ok' : 'err', text: r.ok ? `连接正常:${ver}` : `${r.stage === 'version' ? '取版本失败' : '连接失败'}:${r.error}` })
    } catch (ex: any) {
      setNote({ kind: 'err', text: '连接失败: ' + (ex.message || String(ex)) })
    } finally { setBusy('') }
  }

  // ---- 命令行集成(用户 PATH) ----
  const doInstall = async (add: boolean) => {
    setInstallBusy(true); setInstallNote('')
    try {
      const st = add ? await api.installAdd() : await api.installRemove()
      setStatus((prev) => (prev ? { ...prev, install: st } : prev))
      setInstallNote(add ? `已把 ${st.exeDir} 加入用户 PATH,新开的终端即可直接敲 tt` : `已从用户 PATH 移除 ${st.exeDir}`)
    } catch (ex: any) {
      setInstallNote('操作失败: ' + (ex.message || String(ex)))
    } finally { setInstallBusy(false) }
  }

  // ---- tt 自身的更新 ----
  // 查:人去点才联网(服务不轮询——见 internal/web/update.go 的三条纪律)。
  const doUpdateCheck = async () => {
    setUpdBusy('check'); setUpdNote(null)
    try {
      const r = await api.updateCheck()
      if (!r.ok) {
        setUpdNote({ kind: 'err', text: (r.error || '检查未成') + (r.where ? `（${r.where}）` : '') })
        return
      }
      setUpd((prev) => (prev ? { ...prev, check: r.check, hint: r.hint } : prev))
      setUpdNote({ kind: 'ok', text: r.check?.newer ? (r.hint || `有新版本 ${r.check?.latest}`) : `已是最新（${r.check?.latest}）` })
    } catch (ex: any) {
      setUpdNote({ kind: 'err', text: '检查失败: ' + (ex.message || String(ex)) })
    } finally { setUpdBusy('') }
  }

  // 装:服务端拉一个 `tt update --yes` 子进程去做 —— 本服务几分钟后会被它停掉替换,
  // 所以这里只能告诉用户“已接手”,不能等结果。
  // 这条确认尤其不能省:它会把当前服务停掉,页面随后断开。
  const doUpdateApply = () => {
    setAsk({
      title: '要现在升级 tt 吗?',
      desc: '下载并校验完成后会停掉本服务并替换程序文件,\n本页随后会断开,请稍后重新打开。',
      confirmText: '开始升级',
      run: () => void applyUpdate(),
    })
  }
  const applyUpdate = async () => {
    setUpdBusy('apply'); setUpdNote(null)
    try {
      const r = await api.updateApply()
      setUpdNote({ kind: 'ok', text: r.note || `更新器已接手（pid ${r.pid}）` })
    } catch (ex: any) {
      setUpdNote({ kind: 'err', text: '交棒失败: ' + (ex.message || String(ex)) })
    } finally { setUpdBusy('') }
  }

  // ---- 缓存清理 ----
  // 只删可再生的中间数据(后端只认 internal/config 那份清单);config.json 不碰。
  // 后端把清完的状态一起回,所以这里不必再打一次 configStatus。
  const doClearCache = async () => {
    setCacheBusy(true); setCacheNote('')
    try {
      const r = await api.cacheClear()
      setStatus((prev) => (prev ? { ...prev, cache: r.cache } : prev))
      setCacheNote(r.removed > 0
        ? `已清理 ${r.removed} 个文件,释放 ${fmtBytes(r.freed)}`
        : '没有可清理的文件')
    } catch (ex: any) {
      setCacheNote('清理失败: ' + (ex.message || String(ex)))
    } finally { setCacheBusy(false) }
  }

  // 「已保存」提示短暂停留后回到空闲
  useEffect(() => {
    if (saveState !== 'saved') return
    const t = window.setTimeout(() => setSaveState('idle'), 2000)
    return () => window.clearTimeout(t)
  }, [saveState])

  // ---- 源码镜像 / 字典同步(动作) ----
  // 只在「数据字典」分区且正在跑的时候轮询:本页是常驻挂载的,无脑轮询会一直打后端。
  // 环境不再需要在这里挑一个"当前项"了 —— 动作由表格里那一行给出(见 pullMirror/runSync 的参数)。
  const refreshMirror = useCallback(async () => {
    try { setMirror(await api.mirror()) } catch { /* 取不到就不显示操作区,不打扰 */ }
  }, [])
  const refreshSync = useCallback(async () => {
    try { setSync(await api.dbsync()) } catch { /* 同上 */ }
  }, [])
  useEffect(() => {
    if (section !== 'data-dict') return
    void refreshMirror(); void refreshSync()
  }, [section, refreshMirror, refreshSync])
  const mirrorRunning = !!mirror?.job?.running
  const syncRunning = !!sync?.job?.running
  useEffect(() => {
    if (!mirrorRunning) return
    const t = window.setInterval(() => { void refreshMirror() }, 800)
    return () => window.clearInterval(t)
  }, [mirrorRunning, refreshMirror])
  useEffect(() => {
    if (!syncRunning) return
    const t = window.setInterval(() => { void refreshSync() }, 800)
    return () => window.clearInterval(t)
  }, [syncRunning, refreshSync])

  // 环境名由调用的那一行给出 —— 表格里每行一个按钮,不再有"先选环境"这一步。
  // 全量重建会整目录替换,先弹确认;确认后走 doMirror(动作与确认分开,
  // 弹窗那一侧只负责 setAsk 与描述文案)。
  const pullMirror = (full: boolean, env: string) => {
    if (!env) return
    if (!full) { void doMirror(false, env); return }
    setAsk({
      title: `全量重建环境「${env}」?`,
      desc: '将整目录替换本地镜像(删除服务器上已不存在的残留)。\n首次或按需执行,耗时较长。',
      confirmText: '全量重建',
      run: () => void doMirror(true, env),
    })
  }
  const doMirror = async (full: boolean, env: string) => {
    setOpBusy('mirror'); setOpNote(null)
    try {
      await api.mirrorPull(env, full)
      await refreshMirror()
      // 刻意不给"已开始…"这种成功提示:进度块马上就会出现,并写着环境、
      // 全量/增量与进行中状态 —— 同一件事在一屏里说两遍就是噪音。
      // (失败才留提示:失败原因只有它能说清,见下面的 catch。)
    } catch (ex: any) {
      setOpNote({ kind: 'err', text: '拉取失败: ' + (ex.message || String(ex)) })
    } finally { setOpBusy('') }
  }
  /**
   * 停止正在跑的拉取。不弹二次确认:这个动作本身就是"撤销误点",再让它过一道确认
   * 等于把最需要即时性的操作拖慢一步。后端的取消是**异步**的(SFTP 读到下一个分片才返回),
   * 所以这里不 await 到"真的停了"—— 立刻回读一次让轮询接管,按钮随即变成「正在停止…」。
   */
  const stopMirror = async () => {
    try {
      await api.mirrorStop()
      await refreshMirror()
    } catch (ex: any) {
      setOpNote({ kind: 'err', text: '停止失败: ' + (ex.message || String(ex)) })
    }
  }
  const runSync = (env: string) => {
    if (!env) return
    setAsk({
      title: `同步环境「${env}」的字典数据?`,
      desc: `目标:${sync?.target || ''}\n将覆盖本地 SQLite(原库自动备份为 .bak),约 85 万行,需数分钟。`,
      confirmText: '开始同步',
      run: () => void doSync(env),
    })
  }
  const doSync = async (env: string) => {
    setOpBusy('sync'); setOpNote(null)
    try {
      await api.dbsyncRun(env)
      await refreshSync()
      // 成功不给提示,理由同 doMirror(进度块随后就把这件事说清了)
    } catch (ex: any) {
      setOpNote({ kind: 'err', text: '同步失败: ' + (ex.message || String(ex)) })
    } finally { setOpBusy('') }
  }
  /** 停止正在跑的同步(理由同 stopMirror:不二次确认,异步取消不等待)。 */
  const stopSync = async () => {
    try {
      await api.dbsyncStop()
      await refreshSync()
    } catch (ex: any) {
      setOpNote({ kind: 'err', text: '停止失败: ' + (ex.message || String(ex)) })
    }
  }

  if (!cfg) return <div className="p-6 text-sm text-muted-foreground">{err || '加载配置中…'}</div>
  const cur = sshs[selSsh]
  const curDb = cur?.db
  // 当前选中的这个环境是不是默认环境(「设为默认」按钮的禁用依据,与左侧绿点同一判据)
  const isCurDefault = !!cur && cfg.activeEnv === sshName(cur)
  // 表格里那些行级动作的共用禁用条件:别的操作在跑、或这个动作自身在跑
  const mirrorBusy = mirrorRunning || opBusy !== ''
  const syncBusy = syncRunning || opBusy !== ''
  const mirrorDirOk = !!dict.mirrorDir.trim() && !isDirty('mirror')

  // 卡片的保存按钮:只在真的有改动时才出现。
  // 常驻一个灰着的「保存」挂在每张卡右上角,既不传达信息(它一直是灰的),又让人以为
  // 有什么东西没保存。有没有未保存的改动看内容区右上角那行状态文字就够了,这条按钮
  // 是「动手保存」的入口,没得保存时它就不该在。
  const SaveBtn = ({ keys }: { keys: ConfigKey[] }) => {
    if (!isDirty(...keys)) return null
    return (
      <Button size="sm" disabled={saveState === 'saving'} onClick={() => void save(keys)}>
        {saveState === 'saving' ? '保存中…' : '保存'}
      </Button>
    )
  }

  return (
    // flex-1 + min-w-0:填满活动栏右侧的整块区域。少了它这个根会按内容宽度撑开,
    // 于是外层的滚动条落在内容列的右缘(页面中间),右边留一大片空白 —— 滚动条看着像没贴着边。
    // 填满之后里面的 mx-auto max-w-3xl 才真正居中,标题与卡片始终对齐在同一条竖线上。
    <div className="flex h-full min-h-0 min-w-0 flex-1 text-xs">
      {/* 左侧导航:四个分区各一行。
          分区不带图标,也不给「设置」这类标题 —— 这一栏本身就在设置页里,
          重复一遍栏名是废话,而图标只是把每个分区名往右推了一格。
          原先选中项下面还会展开该分区的卡片清单,已经去掉:右边一屏就是那个分区的全部卡片,
          清单只是把它们再念一遍。
          p-1 是给选中项的 rounded-ui 留出边距:栏内不留白的话,圆角会被这一栏的
          overflow-auto 直接裁掉,选中项看着还是直角。 */}
      <div className="flex w-44 shrink-0 flex-col gap-0.5 overflow-auto border-r border-border p-1">
        {SETTINGS_SECTIONS.map((s) => {
          const active = section === s.key
          return (
            <button
              key={s.key}
              onClick={() => setSection(s.key)}
              className={cn('flex w-full items-center rounded-ui px-2 py-1.5 text-left transition-colors',
                active ? 'bg-accent font-medium text-accent-foreground' : 'text-muted-foreground hover:bg-accent/60 hover:text-foreground')}
            >
              <span className="min-w-0 flex-1 truncate">{s.label}</span>
            </button>
          )
        })}
      </div>

      {/* 右侧内容区:当前分区的全部卡片,整体滚动。
          站点管理是左右两栏(左边环境清单、右边连接表单),列宽放宽一档给表单留出原来的舒适宽度;
          其余分区都是单列设置列表,仍然收在 max-w-3xl。
          px-1 不是给排版留白,是给**输入框的聚焦环**留位置:内层 mx-auto 的宽度就等于卡片宽度,
          卡片里贴边的输入框 focus 时 ring 要往外扩约 2px,没有这点余量就会被本层的
          overflow-auto 裁掉 —— 表现是"选中时高亮缺一条边/一角"。 */}
      <div className="min-h-0 flex-1 overflow-auto px-1">
        <div className={cn('mx-auto space-y-3 p-4', section === 'sites' ? 'max-w-5xl' : 'max-w-3xl')}>
          <div className="flex items-center justify-between gap-2">
            <h2 className="text-sm font-medium text-foreground">
              {SETTINGS_SECTIONS.find((s) => s.key === section)?.label}
            </h2>
            {/* 这里原先挂了一个「从 config.json 重新读取」的刷新按钮,已去掉:
                本页切走再切回来就会自动回读一次(见上面 view 的那个 effect),而按钮只在
                没有未保存修改时才可点 —— 也就是说它唯一能覆盖的场景是"人一直停在本页、
                同时在命令行改了配置",代价却是每个分区标题右侧都常驻一个用不上的图标。 */}
            <span className="flex min-w-0 items-center gap-2">
              {err && <span className="truncate text-[11px] text-red-600 dark:text-red-400" title={err}>{err}</span>}
              {!err && (
                <span className={cn('text-[11px]',
                  saveState === 'saved' ? 'text-emerald-600 dark:text-emerald-400'
                    : dirty ? 'text-amber-600 dark:text-amber-400' : '')}>
                  {saveState === 'saving' ? '保存中…' : saveState === 'saved' ? '已保存 ✓' : dirty ? '有未保存的修改' : ''}
                </span>
              )}
            </span>
          </div>

          {/* ============ 站点管理(左右两栏:环境清单 | 连接表单) ============ */}
          {section === 'sites' && (
            <div className="grid h-[40rem] grid-cols-[16rem_1fr] gap-3">
              {/* 两栏定高:跟环境数量、表单长短都无关,两栏永远一样高,页面也不会跟着内容忽长忽短。
                  高度取 40rem —— 常见窗口高度下装得下,也不需要跟着视口算。
                  这张卡**没有**自己的「保存」:左右两栏改的是同一个 hosts 节
                  (清单里的增删与默认环境、右边那个环境的连接信息),各挂一个按钮时
                  同一屏会出现两个"保存",点哪个、会不会互相覆盖都说不清。保存只留右侧那一个。 */}
              <Card
                id="card-sites-list"
                panel
                title="环境清单"
              >
                {/* 列表撑满卡片余下的高度、自己滚:环境再多也不顶高卡片,
                    「新增环境」钉在卡片底部,不用翻到列表末尾去找 */}
                <div className="min-h-0 flex-1 space-y-0.5 overflow-auto">
                  {sshs.length === 0 && <div className="py-2 text-muted-foreground">(空)</div>}
                  {sshs.map((e, i) => {
                    const isDefault = cfg.activeEnv === sshName(e)
                    return (
                      <div key={i}
                        className={cn('flex items-center gap-1 rounded-ui px-1.5 py-1 transition-colors',
                          i === selSsh ? 'bg-accent text-accent-foreground' : 'hover:bg-accent/50')}>
                        <button className="min-w-0 flex-1 text-left" onClick={() => setSelSsh(i)}>
                          <span className="block truncate">{sshName(e) || '(新环境)'}</span>
                          <span className="block truncate text-[11px] text-muted-foreground">
                            {e.zone || '-'} · {e.port || '-'} · {e.db ? (e.db.type === 'kingbase' ? '金仓' : 'Oracle') : '无库'}
                          </span>
                        </button>
                        {/* 默认环境标记:**只表示状态,不可点**。改默认走右侧卡片的「设为默认」按钮
                            —— 一个绿点既是标记又当开关的话,点它到底发生什么全靠猜(原来是星标,
                            点亮/取消两种含义混在一个图标里)。非默认环境这里留同宽空位,
                            免得名字列的右边界跟着"有没有绿点"忽宽忽窄。 */}
                        <span className="flex w-3.5 shrink-0 justify-center" title={isDefault ? '默认环境' : undefined}>
                          {isDefault && <span className="h-2 w-2 rounded-full bg-emerald-500" />}
                        </span>
                      </div>
                    )
                  })}
                </div>
                <div className="flex items-center gap-2">
                  <Button variant="outline" size="sm" onClick={addSsh}><Plus className="h-3.5 w-3.5" />新增环境</Button>
                </div>
              </Card>

              <Card id="card-sites-conn" panel>
                {!cur ? (
                  <div className="py-2 text-muted-foreground">先在左边的「环境清单」里选一个环境,或新增一个。</div>
                ) : (
                  <>
                    {/* 一行放完:页签在左、操作在右,右边跟着这张卡的「保存」。
                        卡片头部（Card 的 right 位）是右对齐的,单独放页签会偏到最右边,
                        而它下面还要再起一行放按钮 —— 上下两块都占宽度,中间空出一大片。
                        这里合成一行,页面少一行高度,视线也不用上下折返。
                        「设为默认」不用星形图标:文字已经把动作说全,图标只是多一个要解读的形状。 */}
                    <Tabs value={envTab} onValueChange={(v) => setEnvTab(v as 'ssh' | 'db')}>
                      <div className="flex items-center gap-2">
                        <TabsList>
                          <TabsTrigger value="ssh">SSH 服务器</TabsTrigger>
                          <TabsTrigger value="db">数据库</TabsTrigger>
                        </TabsList>
                        <span className="ml-auto flex items-center gap-2">
                          <Button variant="ghost" size="sm" title="复制该环境(SSH 与数据库连接整份抄一份,名称自动加 -copy)"
                            onClick={() => dupSsh(selSsh)}>
                            <Copy className="h-3.5 w-3.5" />复制
                          </Button>
                          <Button variant="ghost" size="sm" title="删除该环境" onClick={askDelSsh}>
                            <Trash2 className="h-3.5 w-3.5" />删除
                          </Button>
                          <Button variant="outline" size="sm"
                            title={isCurDefault ? '当前已是默认环境' : '设为默认环境(会话未指定环境时用它)'}
                            disabled={isCurDefault}
                            onClick={() => setDefaultEnv(sshName(cur))}>
                            设为默认
                          </Button>
                          <SaveBtn keys={['hosts']} />
                        </span>
                      </div>

                      <TabsContent value="ssh" forceMount className="mt-3 space-y-3 data-[state=inactive]:hidden">
                        <div className="grid grid-cols-2 gap-x-4 gap-y-3">
                          <Field label="环境名称">
                            <Input className={input} value={cur.name} placeholder="如 正式区"
                              onChange={(e) => patchSsh(selSsh, { name: e.target.value })} />
                          </Field>
                          <Field label="IP 主机">
                            <Input className={input} value={cur.host} onChange={(e) => patchSsh(selSsh, { host: e.target.value })} />
                          </Field>
                          <Field label="端口">
                            <Input className={input} type="number" value={cur.port || ''}
                              onChange={(e) => patchSsh(selSsh, { port: Number(e.target.value) || 22 })} />
                          </Field>
                          <Field label="登录区域">
                            <Input className={input} value={cur.zone} placeholder="31开发/35测试/36正式/39PATCH/t出货"
                              onChange={(e) => patchSsh(selSsh, { zone: e.target.value })} />
                          </Field>
                          <Field label="账号">
                            <Input className={input} value={cur.user} onChange={(e) => patchSsh(selSsh, { user: e.target.value })} />
                          </Field>
                          <Field label="密码">
                            <div className="relative">
                              <Input className={input} type={showPwd ? 'text' : 'password'} value={cur.password}
                                onChange={(e) => patchSsh(selSsh, { password: e.target.value })} />
                              <button type="button" title={showPwd ? '隐藏密码' : '显示密码'}
                                onClick={() => setShowPwd((v) => !v)}
                                className="absolute top-1/2 right-1 -translate-y-1/2 p-0.5 text-muted-foreground hover:text-foreground">
                                {showPwd ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
                              </button>
                            </div>
                          </Field>
                          {/* 工作区目录排在 TOPENT 之前:它决定"这套源码在哪、包往哪出",
                              是跑 tt dev tzs 的**前置条件**(留空运行类动词直接拒启),
                              比 TOPENT(只有一个默认企业值)更靠前需要确认。 */}
                          <Field label="工作区目录" className="col-span-2">
                            <Input className={input} value={cur.workspace} placeholder="D:\\t100_wrok_dir\\客户\\prd"
                              onChange={(e) => patchSsh(selSsh, { workspace: e.target.value })} />
                          </Field>
                          <Field label="TOPENT" className="col-span-2">
                            <Input className={input} value={cur.topent} onChange={(e) => patchSsh(selSsh, { topent: e.target.value })} />
                          </Field>
                        </div>
                      </TabsContent>

                      <TabsContent value="db" forceMount className="mt-3 space-y-3 data-[state=inactive]:hidden">
                        {curDb == null ? (
                          <div className="border border-dashed border-border p-4 text-center text-muted-foreground">
                            <p className="mb-2">该环境尚未挂载数据库连接</p>
                          <Button variant="outline" size="sm" onClick={() => patchDb(selSsh, {})}>添加数据库</Button>
                        </div>
                      ) : (
                        <>
                          <div className="grid grid-cols-2 gap-x-4 gap-y-3">
                            <Field label="类型">
                              <Select value={curDb.type || 'oracle'} onValueChange={(v) => patchDb(selSsh, { type: v })}>
                                <SelectTrigger className="h-7 w-full text-xs"><SelectValue /></SelectTrigger>
                                <SelectContent>
                                  <SelectItem value="oracle">Oracle</SelectItem>
                                  <SelectItem value="kingbase">金仓(KingbaseES)</SelectItem>
                                </SelectContent>
                              </Select>
                            </Field>
                            <Field label="主机地址">
                              <Input className={input} value={curDb.host} placeholder="客户端与服务器侧均可达"
                                onChange={(e) => patchDb(selSsh, { host: e.target.value })} />
                            </Field>
                            <Field label="端口">
                              <Input className={input} type="number" value={curDb.port || ''}
                                placeholder={curDb.type === 'kingbase' ? '54321' : '1521'}
                                onChange={(e) => patchDb(selSsh, { port: Number(e.target.value) || 0 })} />
                            </Field>
                            {curDb.type === 'kingbase' ? (
                              <Field label="库名">
                                <Input className={input} value={curDb.database} placeholder="如 topprd"
                                  onChange={(e) => patchDb(selSsh, { database: e.target.value })} />
                              </Field>
                            ) : (
                              <Field label="服务名(SERVICE_NAME)">
                                <Input className={input} value={curDb.service} placeholder="如 t35prd"
                                  onChange={(e) => patchDb(selSsh, { service: e.target.value })} />
                              </Field>
                            )}
                          </div>
                          <label className="mt-3 flex items-center gap-2 text-[11px] text-muted-foreground">
                            <input type="checkbox" checked={curDb.readonlySql !== false}
                              onChange={(e) => patchDb(selSsh, { readonlySql: e.target.checked ? undefined : false })} />
                            允许 AI 执行只读 SQL(默认开启;关掉后调试端点的 SQL 查询直接 403)
                          </label>
                          <div className="mt-3 flex items-center gap-2">
                            <Button variant="outline" size="sm" disabled={busy !== '' || !cur.host || !cur.user}
                              title={!cur.host || !cur.user ? '先填 SSH 主机与账号' : '按区域登录服务器,解析库地址/service 并回填'}
                              onClick={() => void fetchFromServer()}>
                              <RefreshCw className={cn('h-3.5 w-3.5', busy === 'fetch' && 'animate-spin')} />从服务器获取
                            </Button>
                            <Button variant="outline" size="sm" disabled={busy !== '' || !curDb.host.trim()}
                              title="从本机直连该库,验证网络与账号"
                              onClick={() => void testConn()}>
                              <Zap className={cn('h-3.5 w-3.5', busy === 'conn' && 'animate-spin')} />测试连接
                            </Button>
                          </div>
                          {note && (
                            <p className={cn('mt-2 border-l-2 px-2 py-1 text-[11px]',
                              note.kind === 'ok' ? 'border-emerald-500/50 bg-emerald-500/5 text-emerald-700 dark:text-emerald-300'
                                : note.kind === 'err' ? 'border-red-500/50 bg-red-500/5 text-red-700 dark:text-red-300'
                                  : 'border-sky-500/50 bg-sky-500/5 text-muted-foreground')}>
                              {note.text}
                            </p>
                          )}

                          <div className="mt-4">
                            <div className="mb-1 text-[11px] text-muted-foreground">
                              账号列表(账号即 schema 名;客户端直连取首项,服务器侧调试按 TOPENT 解析出账号后在此查密码)
                            </div>
                            {/* 账号可能很多:表格封顶 16rem 后就地滚,表头 sticky 常驻。
                                滚动容器由我们自己提供(container={false}) —— 用 Table 自带的
                                那个 overflow-x-auto 包一层的话,它就成了滚动祖先,th 的 sticky 会失效。
                                表头底色必须不透明(默认 bg-muted/50 是半透明的),否则行会从表头透出来。 */}
                            <div className="max-h-64 overflow-auto">
                              <Table container={false}>
                                <TableHeader>
                                  <TableRow className="hover:bg-transparent">
                                    <TableHead className="sticky top-0 z-10 w-1/2 bg-muted">账号(schema)</TableHead>
                                    <TableHead className="sticky top-0 z-10 bg-muted">密码(缺省 = 账号)</TableHead>
                                    <TableHead className="sticky top-0 z-10 w-16 bg-muted">操作</TableHead>
                                  </TableRow>
                                </TableHeader>
                                <TableBody>
                                  <TableRow>
                                    <TableCell className={CELL_EDITING}>
                                      <input className={cn('cell-input font-mono')} placeholder="新增账号…" value={addAcct.account}
                                        onChange={(e) => setAddAcct((s) => ({ ...s, account: e.target.value }))}
                                        onKeyDown={(e) => { if (e.key === 'Enter') pushAcct() }} />
                                    </TableCell>
                                    <TableCell className={CELL_EDITING}>
                                      <input className="cell-input font-mono" placeholder="留空 = 与账号相同" value={addAcct.password}
                                        onChange={(e) => setAddAcct((s) => ({ ...s, password: e.target.value }))}
                                        onKeyDown={(e) => { if (e.key === 'Enter') pushAcct() }} />
                                    </TableCell>
                                    <TableCell className="text-center">
                                      <Button variant="ghost" size="sm" title="添加账号" disabled={!addAcct.account.trim()} onClick={pushAcct}>
                                        <Plus className="h-3.5 w-3.5" />
                                      </Button>
                                    </TableCell>
                                  </TableRow>
                                  {curDb.accounts.map((a, j) => (
                                    <TableRow key={j}>
                                      <TableCell className={CELL_EDITING}>
                                        <input className="cell-input font-mono" value={a.account}
                                          onChange={(e) => patchAcct(selSsh, j, { account: e.target.value })} />
                                      </TableCell>
                                      <TableCell className={cn(CELL_EDITING, 'relative')}>
                                        <input className="cell-input font-mono" value={a.password}
                                          onChange={(e) => patchAcct(selSsh, j, { password: e.target.value })} />
                                      </TableCell>
                                      <TableCell className="text-center whitespace-nowrap">
                                        <Button variant="ghost" size="sm" title="在服务器侧验证该账号"
                                          disabled={busy !== '' || accProbe?.state === 'testing'}
                                          onClick={() => void verifyAcct(j)}>
                                          <Zap className={cn('h-3.5 w-3.5', accProbe?.i === j && accProbe.state === 'testing' && 'animate-spin')} />
                                        </Button>
                                        <Button variant="ghost" size="sm" title="删除该账号" onClick={() => delAcct(selSsh, j)}>
                                          <Trash2 className="h-3.5 w-3.5" />
                                        </Button>
                                      </TableCell>
                                    </TableRow>
                                  ))}
                                </TableBody>
                              </Table>
                            </div>
                            {accProbe && (
                              <p className={cn('mt-1 text-[11px]',
                                accProbe.state === 'ok' ? 'text-emerald-600 dark:text-emerald-400'
                                  : accProbe.state === 'err' ? 'text-red-600 dark:text-red-400' : 'text-muted-foreground')}>
                                {accProbe.msg || '验证中…'}
                              </p>
                            )}
                          </div>
                        </>
                      )}
                      </TabsContent>
                    </Tabs>
                  </>
                )}
              </Card>
            </div>
          )}

          {/* ============ 数据字典 ============ */}
          {section === 'data-dict' && (
            <>
              {/* 「查询数据源」把数据源选择与「数据同步」合在一张卡里:两者讲的是同一件事 ——
                  本地那份 SQLite 从哪来、查的时候走哪条路。原先分成两张卡,第一张只放一个下拉,
                  第二张才是"这份数据怎么来的",读的人得自己把两处接起来。 */}
              <Card id="card-dict-query" title="查询数据源"
                description="r.t / r.v / desc / scc / r.q 等查询命令走哪条路取数。"
                right={<SaveBtn keys={['query']} />}>
                <SettingRow
                  id="query.source"
                  label="数据源"
                  description="选「本地」须先同步下面任一环境;连不上远程也不会静默回落本地。"
                  control={
                    <Select value={dict.querySource === 'local' ? 'local' : '__auto__'}
                      onValueChange={(v) => { setDict((s) => ({ ...s, querySource: v === '__auto__' ? '' : v })); markDirty('query') }}>
                      <SelectTrigger className="h-7 w-full text-xs"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {/* 只有两档:**在线(默认环境)** 与 **本地 SQLite**。
                            原先还按环境列了一遍「在线(主机109)」「在线(尚阳通)」——
                            但"用哪个环境"在「站点管理」里已经有一个默认环境(左侧那个绿点),
                            在这里再选一次等于让同一件事有两个出处:两处不一致时该信谁?
                            所以这里只留"在线/本地"这个**路线**选择,环境一律走默认那个。 */}
                        <SelectItem value="__auto__">在线(默认环境)</SelectItem>
                        <SelectItem value="local">本地(SQLite)</SelectItem>
                      </SelectContent>
                    </Select>
                  }
                />

                {/* 数据同步(原「数据同步」卡片):一个环境一行,每行一个「开始同步」。
                    它必须留在这张卡里 —— 上面选了「本地」就等于声明"我读本地那份",
                    而本地那份正是这张表同步出来的。 */}
                <SectionTitle>数据同步(写入本地 SQLite)</SectionTitle>
                {/* 只显示**失败**:成功不需要提示 —— 进度块随后就会写明环境、全量/增量
                    与进行中状态,再挂一句"已开始…"是把同一件事说两遍。
                    失败必须留:原因只有这里能说清(进度块只显示状态,不含错误详情)。 */}
                {opNote?.kind === 'err' && (
                  <p className="text-[11px] text-red-600 dark:text-red-400">{opNote.text}</p>
                )}
                {sync && sync.envs.length === 0 && (
                  <p className="text-[11px] text-muted-foreground">
                    没有可同步的环境 —— 需要先在「站点管理」为环境挂上数据库连接。
                  </p>
                )}
                {sync && sync.envs.length > 0 && (
                  <Table>
                    <TableHeader>
                      <TableRow className="hover:bg-transparent">
                        <TableHead>环境</TableHead>
                        <TableHead className="w-28">类型</TableHead>
                        <TableHead>地址</TableHead>
                        <TableHead className="w-32">状态</TableHead>
                        <TableHead className="w-28">操作</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {sync.envs.map((e) => {
                        const busy = !!sync.job && sync.job.env === e.name && (sync.job.running || sync.job.done)
                        const failed = busy && sync.job.phase === 'error'
                        return (
                          <TableRow key={e.name} className="hover:bg-transparent">
                            <TableCell className="px-2">
                              {e.name}
                              {e.name === sync.activeEnv && <span className="ml-1 text-[11px] text-muted-foreground">(默认)</span>}
                            </TableCell>
                            <TableCell className="px-2">{e.type}</TableCell>
                            <TableCell className="px-2">
                              <span className="block truncate font-mono text-[11px] text-muted-foreground" title={e.address}>{e.address}</span>
                            </TableCell>
                            <TableCell className="px-2">
                              <EnvStatus busy={busy} running={!!sync.job?.running} idle="空闲"
                                error={failed ? (sync.job.error || sync.job.message || '失败') : undefined} />
                            </TableCell>
                            <TableCell className="px-2">
                              {/* 表格里的行内动作一律用 ghost 文本按钮(与账号表、环境动作行一致):
                                  一整列实心主色块会把"表格"读成"一堆按钮",行与行的主次也没了。
                                  ghost 只是没有底色,仍是完整按钮 —— 点击区、禁用态、悬停反馈都在。 */}
                              <Button variant="ghost" size="sm" disabled={syncBusy}
                                title={`从「${e.name}」拉取字典数据到本地 SQLite`}
                                onClick={() => void runSync(e.name)}>
                                开始同步
                              </Button>
                            </TableCell>
                          </TableRow>
                        )
                      })}
                    </TableBody>
                  </Table>
                )}
                {sync?.job && (sync.job.running || sync.job.done) && (
                  <JobProgress
                    env={sync.job.env}
                    phaseLabel={SYNC_PHASE[sync.job.phase] || sync.job.phase}
                    tag={sync.job.running ? '进行中' : '已结束'}
                    pct={sync.job.phase === 'done' ? 100
                      : sync.job.tableTotal > 0 ? Math.round((sync.job.tableIndex * 100) / sync.job.tableTotal) : 0}
                    error={sync.job.phase === 'error' ? (sync.job.error || sync.job.message) : undefined}
                    message={sync.job.running ? undefined
                      : (sync.job.phase === 'cancelled' ? sync.job.message
                        : (sync.job.phase === 'done' ? sync.job.message : undefined))}
                    cells={[
                      sync.job.tableTotal > 0 ? `${sync.job.tableIndex}/${sync.job.tableTotal} 表` : '—',
                      sync.job.table || '',
                      sync.job.tableRows > 0 ? `${sync.job.tableRows.toLocaleString()} 行` : '',
                      sync.job.totalRows > 0 ? `共 ${sync.job.totalRows.toLocaleString()} 行` : '',
                      sync.job.elapsed || '',
                    ].filter(Boolean)}
                    stopping={sync.job.phase !== 'error' && sync.job.message === '正在停止…'}
                    onStop={sync.job.running ? () => void stopSync() : undefined}
                  />
                )}
                {sync?.job?.backup && <p className="text-[11px] text-muted-foreground">原库已备份:{sync.job.backup}</p>}
                {sync?.job?.warning && <pre className="whitespace-pre-wrap text-[11px] text-amber-600 dark:text-amber-400">{sync.job.warning}</pre>}

                {/* 本地库的落点**只读**:它是个本地缓存文件,位置由服务端算(缺省在数据目录下),
                    让用户在设置页里挑一个路径,除了能把 CLI/Web/查询三边指向不同的库之外没有好处。
                    要改仍在 config.json 里写 sync.target(后端照旧认它),这里只如实显示生效值。
                    原先这里还跟了一句状态说明(文件在不在、同步会先备份 .bak、位置是否来自配置):
                    那些都是"用户不需要为它做决定"的信息 —— 在不在不影响他点「开始同步」,
                    备份与失败保护是服务端自动行为,位置来源看这个只读值本身也够。 */}
                <InfoRow label="保存位置" value={status?.sync.target || '—'} />
              </Card>

              <Card id="card-dict-mirror" title="源码镜像"
                description="把 T100 源码(4gl / 4fd / 42s / *.inc)拉到本地,供 AI 检索。"
                right={<SaveBtn keys={['mirror']} />}>
                <SettingRow
                  id="mirror.dir"
                  label="镜像根目录"
                  description={
                    !dict.mirrorDir.trim()
                      ? '留空 = 未设置,镜像与 AI 检索不可用。'
                      : status?.mirror.exists
                        ? '目录已存在。'
                        : '目录不存在,拉取时自动创建。'
                  }
                  control={<Input className={input} value={dict.mirrorDir} placeholder="如 D:\t100\mirror"
                    onChange={(e) => { setDict((s) => ({ ...s, mirrorDir: e.target.value })); markDirty('mirror') }} />}
                />
                {/* 一个环境一行:镜像本来就是**按环境**分开的(每个环境一个镜像目录、一份基线),
                    原先"先在下拉里挑一个环境、再点按钮"把这件事藏进了两步操作里 ——
                    看不出总共有几个环境、各自拉没拉过。改成表之后这些直接摊开。 */}
                {mirror && mirror.envs.length === 0 && (
                  <p className="text-[11px] text-muted-foreground">还没有 SSH 环境,先去「站点管理」添加。</p>
                )}
                {mirror && mirror.envs.length > 0 && (
                  <Table>
                    <TableHeader>
                      <TableRow className="hover:bg-transparent">
                        <TableHead>环境</TableHead>
                        <TableHead className="w-28">状态</TableHead>
                        <TableHead>镜像目录</TableHead>
                        <TableHead className="w-44">操作</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {mirror.envs.map((e) => {
                        const busy = !!mirror.job && mirror.job.env === e.name && (mirror.job.running || mirror.job.done)
                        const failed = busy && mirror.job.phase === 'error'
                        return (
                          <TableRow key={e.name} className="hover:bg-transparent">
                            <TableCell className="px-2">
                              {e.name}
                              {e.name === mirror.activeEnv && <span className="ml-1 text-[11px] text-muted-foreground">(默认)</span>}
                            </TableCell>
                            <TableCell className="px-2">
                              <EnvStatus busy={busy} running={!!mirror.job?.running}
                                idle={e.ready ? '已有镜像' : '未拉取'}
                                error={failed ? (mirror.job.error || mirror.job.message || '失败') : undefined} />
                            </TableCell>
                            <TableCell className="px-2">
                              <span className="block truncate font-mono text-[11px] text-muted-foreground" title={e.path}>{e.path}</span>
                            </TableCell>
                            <TableCell className="px-2">
                              <div className="flex items-center gap-1 whitespace-nowrap">
                                {/* 同上:两个都走 ghost 文本按钮。原来"增量更新"是实心主色,
                                    一行里最重的视觉落在按钮上而不是"哪个环境、拉没拉过"。
                                    「全量」也去掉 Download 图标:同一格两个按钮,一个带图标一个不带
                                    会让它们看着不是一类东西;动作名已经把事说清了。 */}
                                <Button variant="ghost" size="sm" disabled={mirrorBusy || !mirrorDirOk}
                                  title={!dict.mirrorDir.trim() ? '先填镜像根目录' : isDirty('mirror') ? '镜像根目录有未保存的修改,先保存' : '只拉服务器上变过的文件'}
                                  onClick={() => void pullMirror(false, e.name)}>
                                  增量更新
                                </Button>
                                <Button variant="ghost" size="sm" disabled={mirrorBusy || !mirrorDirOk}
                                  title={!dict.mirrorDir.trim() ? '先填镜像根目录' : isDirty('mirror') ? '镜像根目录有未保存的修改,先保存' : '整目录替换(含服务器上已删的残留)'}
                                  onClick={() => void pullMirror(true, e.name)}>
                                  全量
                                </Button>
                              </div>
                            </TableCell>
                          </TableRow>
                        )
                      })}
                    </TableBody>
                  </Table>
                )}
                {mirror?.job && (mirror.job.running || mirror.job.done) && (
                  <JobProgress
                    env={mirror.job.env}
                    phaseLabel={MIRROR_PHASE[mirror.job.phase] || mirror.job.phase}
                    tag={`${mirror.job.full ? '全量' : '增量'} · ${mirror.job.running ? '进行中' : '已结束'}`}
                    pct={mirror.job.total > 0
                      ? Math.round((mirror.job.bytes * 100) / mirror.job.total)
                      : mirror.job.phase === 'done' ? 100 : 0}
                    error={mirror.job.phase === 'error' ? (mirror.job.error || mirror.job.message) : undefined}
                    message={mirror.job.running ? undefined
                      : (mirror.job.phase === 'cancelled' ? mirror.job.message
                        : (mirror.job.phase === 'done' ? mirror.job.message : undefined))}
                    cells={[
                      ...mirrorCellTexts(mirror.job),
                      !mirror.job.running && mirror.job.elapsed ? `${mirror.job.elapsed}` : '',
                    ].filter(Boolean)}
                    stopping={mirror.job.phase !== 'error' && mirror.job.message === '正在停止…'}
                    onStop={mirror.job.running ? () => void stopMirror() : undefined}
                  />
                )}
              </Card>

              <Card id="card-dict-bdldoc" title="BDL 文档"
                description="Genero BDL 文档的本地目录。"
                right={<SaveBtn keys={['bdldoc']} />}>
                <SettingRow
                  id="bdldoc.dir"
                  label="文档目录"
                  description={
                    !dict.bdldocDir.trim()
                      ? '留空 = 未设置。'
                      : status?.bdldoc.exists
                        ? '目录已存在。'
                        : '目录不存在。'
                  }
                  control={<Input className={input} value={dict.bdldocDir} placeholder="如 D:\t100\bdldoc"
                    onChange={(e) => { setDict((s) => ({ ...s, bdldocDir: e.target.value })); markDirty('bdldoc') }} />}
                />
              </Card>
            </>
          )}

          {/* ============ DEBUG ============ */}
          {section === 'debug' && (
            <>
              {/* 只剩看门狗一项。
                  其余五项(启动参数模板 / FGLServer / print 上限 / 终端尺寸 / 断点持久化)
                  已从界面移除:它们都有内置默认值(见 internal/config 的 FillDefaults:
                  BBDL…{prog} / 按 SSH 来源 IP / 1000 / 200×50 / 开启),现场几乎不需要改,
                  摆出来只会让人以为"必须调"。
                  ⚠ 但它们的值仍然**原样回传**(见 debugPatchOf 的注释):后端的 PUT 是
                  整节替换 + 字段 omitempty,不回传等于把这些键从 config.json 里删掉。
                  「默认环境」那张卡整张去掉 —— 调试会话不指定环境时继承站点默认
                  (站点管理里那个绿点),这一层覆盖没必要暴露。 */}
              <Card id="card-debug-params" title="调试参数"
                description="本机调试会话的默认参数。"
                right={<SaveBtn keys={['debug']} />}>
                <SettingRow id="debug.watchdogSeconds" label="停站看门狗(秒)"
                  description="停站后超过该时长无响应即判定失联。"
                  control={<Input className={input} type="number" value={dbg.watchdogSeconds || ''} placeholder="1800"
                    onChange={(e) => { setDbg((s) => ({ ...s, watchdogSeconds: Number(e.target.value) || 0 })); markDirty('debug') }} />} />
              </Card>
            </>
          )}

          {/* ============ 应用设置 ============ */}
          {section === 'app' && (
            <>
              <Card id="card-app-theme" title="外观"
                description="全站共用一份;切换后各视图都跟着变。">
                <ThemeCards value={theme} onChange={setTheme} />
              </Card>

              <Card id="card-app-service" title="服务"
                description="统一 Web 服务的监听地址。"
                right={<SaveBtn keys={['listen']} />}>
                <SettingRow id="listen" label="监听地址"
                  description="改完需重启 tt serve;端口占用时自动向后顺延。"
                  control={<Input className={input} value={listen} placeholder="127.0.0.1:28670"
                    onChange={(e) => { setListen(e.target.value); markDirty('listen') }} />} />
              </Card>

              <Card id="card-app-install" title="命令行集成"
                description="把 tt.exe 目录加入用户 PATH,之后任意终端可直接敲 tt。只改当前用户,不需管理员。">
                <SettingRow
                  id="app.install"
                  label="用户 PATH"
                  description={
                    !status?.install.supported
                      ? (status?.install.note || '本平台不支持自动写入')
                      : status.install.inUserPath
                        ? '已在用户 PATH 中'
                        : '尚未加入 —— 加完之后新开的终端才能直接敲 tt'
                  }
                  control={
                    <div className="flex items-center gap-1.5">
                      {status?.install.inUserPath ? (
                        <Button variant="outline" size="sm" disabled={installBusy}
                          onClick={() => void doInstall(false)}>
                          <Trash2 className="h-3.5 w-3.5" />{installBusy ? '处理中…' : '从 PATH 移除'}
                        </Button>
                      ) : (
                        <Button size="sm" disabled={installBusy || !status?.install.supported}
                          onClick={() => void doInstall(true)}>
                          <Terminal className="h-3.5 w-3.5" />{installBusy ? '处理中…' : '加入用户 PATH'}
                        </Button>
                      )}
                    </div>
                  }
                />
                {status?.install.exeDir && <InfoRow label="程序目录" value={status.install.exeDir} />}
                {!status?.install.supported && status?.install.manual && (
                  <pre className="overflow-x-auto border border-border bg-muted/30 p-2 text-[11px]">{status.install.manual}</pre>
                )}
                {installNote && (
                  <p className="text-[11px] text-muted-foreground">{installNote}</p>
                )}
              </Card>

              <Card id="card-app-info" title="运行信息">
                <InfoRow label="版本" value={cfg.version || '(未知)'} />
                <InfoRow label="配置文件" value={cfg.config} />
                <InfoRow label="缺省配置位置" value={meta?.defaultConfig || '—'} />
                <InfoRow label="便携模式" value={meta ? (meta.portable ? '是(配置留在程序目录)' : '否') : '—'} />
                <InfoRow label="统一工具目录" value={meta?.toolsHome || '—'} />
                <InfoRow label="配置结构版本" value={meta ? String(meta.schemaVersion) : '—'} />
                <InfoRow label="支持的库类型" value={meta?.supportedTypes?.join(' / ') || '—'} />
                <InfoRow label="默认环境" value={cfg.activeEnv || '(未设置)'} />
                <InfoRow label="环境数" value={String(cfg.sshs?.length ?? 0)} />
              </Card>

              <Card id="card-app-update" title="更新"
                description="检查/升级 tt 自身。查由你手动触发;升级会校验 sha256 后交由独立更新器完成。">
                <InfoRow label="当前版本" value={upd?.version || cfg.version || '(未知)'} />
                <InfoRow label="安装形态" value={upd ? `${upd.kind}${upd.canInstall ? '' : '（不让自装）'}` : '—'} />
                <InfoRow label="更新源" value={upd?.source || '—'} />
                {upd?.check && (
                  <InfoRow label="上次检查"
                    value={`${upd.check.latest}${upd.check.newer ? '（有新版本）' : '（已是最新）'} · ${new Date(upd.check.checkedAt).toLocaleString()}`} />
                )}
                {upd?.last && (
                  <InfoRow label="上次升级"
                    value={`${upd.last.phase}${upd.last.target ? ` → ${upd.last.target}` : ''}${upd.last.error ? ` · ${upd.last.error}` : ''}`} />
                )}
                {upd?.skillsDrift && <InfoRow label="技能" value={upd.skillsDrift} />}
                {upd && !upd.canInstall && upd.refusal && (
                  <p className="text-[11px] text-muted-foreground">{upd.refusal}</p>
                )}
                <div className="flex items-center gap-1.5">
                  <Button variant="outline" size="sm" disabled={updBusy !== ''}
                    onClick={() => void doUpdateCheck()}>
                    <RefreshCw className="h-3.5 w-3.5" />{updBusy === 'check' ? '检查中…' : '检查更新'}
                  </Button>
                  {upd?.check?.newer && upd.canInstall && (
                    <Button size="sm" disabled={updBusy !== ''} onClick={() => void doUpdateApply()}>
                      <Download className="h-3.5 w-3.5" />{updBusy === 'apply' ? '交棒中…' : `升级到 ${upd.check.latest}`}
                    </Button>
                  )}
                  {(updBusy !== '' || upd?.check?.newer) && (
                    <span className="text-[11px] text-muted-foreground">
                      {updBusy === 'apply' ? '升级期间本页会断开，之后重新打开即可' : '也可以直接敲 tt update'}
                    </span>
                  )}
                </div>
                {updNote && (
                  <p className={`text-[11px] ${updNote.kind === 'ok' ? 'text-muted-foreground' : 'text-destructive'}`}>{updNote.text}</p>
                )}
              </Card>

              <Card id="card-app-cache" title="缓存">
                <p className="text-[11px] text-muted-foreground">
                  跑出来的可再生数据(企业快照、源码镜像、查询落盘、断点存档)。它们与 config.json
                  同目录,但性质相反 —— 清除只动这些,配置文件一律不碰。
                </p>
                <InfoRow label="缓存目录" value={status?.cache?.dir || '—'} />
                {(status?.cache?.dirs ?? []).map((d) => (
                  <div key={d.name} className="flex items-center justify-between gap-2 text-[11px]">
                    <span className="font-mono">{d.name}</span>
                    <span className="text-muted-foreground">
                      {d.exists ? `${fmtBytes(d.bytes)} · ${d.files} 个文件` : '无'}
                    </span>
                  </div>
                ))}
                <InfoRow
                  label="共"
                  value={`${fmtBytes(status?.cache?.bytes ?? 0)} · ${status?.cache?.files ?? 0} 个文件`}
                />
                <div className="flex flex-wrap items-center gap-2 pt-1">
                  <Button size="sm" variant="outline" disabled={cacheBusy}
                    onClick={() => void doClearCache()}>
                    <Trash2 className="mr-1 h-3.5 w-3.5" />
                    {cacheBusy ? '清理中…' : '清除缓存'}
                  </Button>
                  {cacheNote && <span className="text-[11px] text-muted-foreground">{cacheNote}</span>}
                </div>
                <p className="text-[11px] text-muted-foreground">
                  tt serve 启动时会自动清理超过{' '}
                  {Math.max(1, Math.round((status?.cache?.maxAgeHours ?? 168) / 24))} 天的缓存
                  —— 刻意不"启动即全清":查询落盘的那些文件正是"刚才那条查询的完整结果",
                  启动就删会把上一条命令刚告诉你的东西删掉。
                </p>
              </Card>
            </>
          )}
        </div>
      </div>

      {/* 全量重建 / 开始同步 / 升级 tt 的二次确认:同一套 AlertDialog。
          文案里的换行(desc 里的 \n)由 AlertDialogDescription 的 whitespace-pre-line 呈现。 */}
      <AlertDialog open={ask !== null} onOpenChange={(o) => { if (!o) setAsk(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{ask?.title}</AlertDialogTitle>
            <AlertDialogDescription>{ask?.desc}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => { const a = ask; setAsk(null); a?.run() }}>
              {ask?.confirmText}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* 删除环境二次确认 */}
      <AlertDialog open={delTarget !== null} onOpenChange={(o) => { if (!o) setDelTarget(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除环境「{delTarget !== null ? sshName(sshs[delTarget]) : ''}」?</AlertDialogTitle>
            <AlertDialogDescription>
              该环境的 SSH 连接、登录区域、TOPENT 与数据库账号配置将一并移除。
              点击「删除」后立即写入 config.json,不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void confirmDelSsh()}>删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
