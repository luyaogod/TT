// 「录制」面板:列出某个环境/ENT 下的录制件(长期库 + 缓存),点开一份看它的身份,并起一轮/停一轮。
//
// 为什么起录/停录只是两个按钮、列表要手动刷新:一轮录制 = 一个隔离 GDC + 一条代理 +
// 一个常驻 Runtime,服务端会**另起一个脱离的进程**(`tt fep rec start`,见
// internal/web/recordings.go 的文件头注释)。界面这边能做的只有"发请求 + 拉列表",
// 真正的活在子进程里 —— 所以起录要等它登记(秒级到几十秒),停录要等它收工落身份。
//
// 列表把「住哪 / 怎么起的 / 什么时候录的 / 多大 / 指纹 / 归属」一次摊开:
// 挑一份录像回放时,前四项决定"这份能不能用"(闸门按**启动路径**比对,见
// internal/fep/path.go),后三项决定"要不要留着"。
//
// 点一行 → 右侧详情面板(照接口日志页那一套:窄列 + 可拖的分割条 + 页签 + 右上角动作)。
// 面板说的是一件列表里放不下的事:**回放前自己核对一遍** —— 身份里的启动路径/作业/ENT 对不对、
// 编号表多少条、这份录制里到底有几条连接(闸门是**按连接**比对的,一条对得上就算过)。
import { useCallback, useEffect, useRef, useState } from 'react'
import { Check, Copy, Fingerprint, Play, RefreshCw, Square, Trash2, Video, X } from 'lucide-react'
import { useStore } from './store'
import { api, type RecordingCurrent, type RecordingMeta, type RecordingRow } from './api'
import { Input, Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../shared/ui'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '../../shared/ui-radix'
import { useToast } from './toast'

// 「这一次作业是怎么起来的」与「归谁 / 留哪」的中文名:与 Go 侧的 Label() 同一套措辞
// (internal/fep/path.go、retention.go),照抄一份 —— 前端无 i18n,中文硬编码。
// 值本身(r.r / debug / none…)照原样显示,中文只进 tooltip:它是身份的一部分,
// 不让用户在界面上看到"r.r 被翻译成了别的什么"而对不上命令行与日志。
const PATH_LABEL: Record<string, string> = {
  'r.r': 'r.r:T100 标准启动器(界面正常,不可调试)',
  debug: 'debug:我们自己的调试线(fglrun -d + fgldb)',
  'r.d': 'r.d:T100 自带调试启动器(gendbg + fgldeb,已不从界面起)',
  none: 'none:只起了隔离实例与代理,没起作业',
}
const ORIGIN_LABEL: Record<string, string> = { user: '用户主动', auto: '系统自动' }
const RETENTION_LABEL: Record<string, string> = {
  library: '长期库(不自动清理)',
  cache: '缓存目录(随统一清理按年龄删)',
}

// 详情面板的宽度分配:与接口日志页同一套做法(拖分割条改占比、占比记在本地)。
// 三个数是那边同一份口径,量很小,先各留一份;哪天有第三处再抽出去。
const RATIO_MIN = 0.25
const RATIO_MAX = 0.75
const RESIZER_W = 5

// 体积:与设置页那份同一口径(那边是视图内的本地函数,没有导出,这里照抄一份)。
function fmtBytes(n: number): string {
  if (!n) return '0 B'
  if (n >= 1 << 30) return (n / (1 << 30)).toFixed(2) + ' GB'
  if (n >= 1 << 20) return (n / (1 << 20)).toFixed(1) + ' MB'
  if (n >= 1 << 10) return (n / (1 << 10)).toFixed(1) + ' KB'
  return n + ' B'
}

// 时间:RFC3339 → 本地时间。空 = 身份里没记时间(老件/没收到),不猜。
function fmtTime(s?: string): string {
  if (!s) return '-'
  const d = new Date(s)
  if (Number.isNaN(d.getTime())) return s
  return d.toLocaleString('zh-CN', { hour12: false })
}

// time.Time 的零值会**照实**出现在 JSON 里(json 的 omitempty 对结构体不起作用):
// 没结束的那一份就是 `0001-01-01T00:00:00Z`,别把它当时间渲染出来。
function fmtFinished(s?: string): string {
  if (!s || s.startsWith('0001-')) return '-'
  return fmtTime(s)
}

// 录了多久(秒)。两端缺一个就报 0,不猜。
function secsBetween(a?: string, b?: string): number {
  if (!a || !b || b.startsWith('0001-')) return 0
  const d = (new Date(b).getTime() - new Date(a).getTime()) / 1000
  return Number.isFinite(d) && d > 0 ? Math.round(d) : 0
}

// 指纹:hash(条数)。hash 空 = 这一轮根本没录到界面(作业没连上来 / 刚起来就断了),
// 那不是"少了几个字",而是"这份录像没有可回放的界面"。
function fmtFP(fp?: { total: number; hash: string }): string {
  if (!fp || !fp.hash) return '-'
  return `${fp.hash}(${fp.total} 条)`
}

export function RecordView() {
  const view = useStore((s) => s.view)
  const toast = useToast()

  const [envNames, setEnvNames] = useState<string[]>([])
  // ENT 是**草稿**,改了要点「刷新」或回车才作用到列表(与接口日志页的"回车生效"同一口径);
  // 起录时用它当前的值 —— 用户刚打的字当然要用上。
  const [ent, setEnt] = useState('')
  const [prog, setProg] = useState('')
  // 已生效的查询条件。对象形式是为了让"环境变了/ENT 变了"都只走一条重拉路径。
  const [q, setQ] = useState<{ env: string; ent: string }>({ env: '', ent: '' })
  // 「刷新」按钮的强制重拉信号(条件没变时也要重新拉一次)
  const [tick, setTick] = useState(0)

  const [rows, setRows] = useState<RecordingRow[]>([])
  const [cur, setCur] = useState<RecordingCurrent | null>(null)
  const [loading, setLoading] = useState(false)
  const [busy, setBusy] = useState<'' | 'start' | 'stop'>('')
  // 复制反馈只记一行(局部 ok/fail,1.5s 复位),照接口日志页的 copyRowID
  const [copied, setCopied] = useState<{ id: string; state: 'ok' | 'fail' } | null>(null)
  const [pendingDel, setPendingDel] = useState<RecordingRow | null>(null)

  // 详情:选中的那一行 + 它的完整身份(列表里的行只有摘要,身份要单独取)。
  const [sel, setSel] = useState<RecordingRow | null>(null)
  const [selMeta, setSelMeta] = useState<RecordingMeta | null>(null)
  const [metaLoading, setMetaLoading] = useState(false)
  const [tab, setTab] = useState<'info' | 'conns' | 'raw'>('info')

  // 环境清单与环境配置同源(设置 → 站点管理);失败不打扰用户 ——
  // 列表那条路会把真正的错(服务没起来)报出来。
  const loadEnvs = useCallback(() => {
    void api.settings().then((c) => {
      const names = (c.sshs || []).map((e) => e.name).filter(Boolean)
      setEnvNames(names)
      // 首次进面板还没选过环境:落到第一个(列表按环境过滤,空着等于"看全部")
      setQ((s) => (s.env ? s : { ...s, env: names[0] || '' }))
    }).catch(() => { /* 服务未就绪:列表那边会报错 */ })
  }, [])

  // 列表与「现在在录吗」一起拉:它们是同一屏的两个事实,分两次进入会看到互相矛盾的画面
  // (列表里已有刚收工的件、而"正在录"还挂着;或反过来)。
  const load = useCallback(async (query: { env: string; ent: string }) => {
    setLoading(true)
    try {
      const [list, now] = await Promise.all([
        api.recordings(query.env, query.ent),
        api.recordingCurrent(),
      ])
      const next = list.recordings || []
      setRows(next)
      setCur(now.recording || null)
      // 选中项跟着列表走:还在就换成新的那行(体积/指纹可能刚更新),不在就关掉详情。
      // 用函数式写法读最新列表,免得把 sel 写进依赖、每次刷新都重拉一遍。
      setSel((s) => (s ? next.find((x) => x.id === s.id && x.where === s.where) || null : null))
    } catch (e: any) {
      toast.error('读取录制失败', { description: e?.message || String(e) })
    } finally {
      setLoading(false)
    }
  }, [toast])

  useEffect(() => { loadEnvs() }, [loadEnvs])
  // 进面板就先拉一次。视图是 keep-alive(只切显隐、不卸载重挂),所以"进面板"这件事
  // 只能由 view 的变化来标识 —— 挂载那一刻 view 就已经是 record 了。
  useEffect(() => {
    if (view !== 'record') return
    void load(q)
  }, [view, q, tick, load])

  // 点开一行就取它的身份。取不到(没有 .meta.json 的老件)照实说,不假装有。
  useEffect(() => {
    if (!sel) { setSelMeta(null); setMetaLoading(false); return }
    let alive = true
    setMetaLoading(true)
    api.recordingShow(sel.id)
      .then((r) => { if (alive) setSelMeta(r.meta || null) })
      .catch((e: any) => {
        if (!alive) return
        setSelMeta(null)
        toast.error('读取录制身份失败', { description: e?.message || String(e) })
      })
      .finally(() => { if (alive) setMetaLoading(false) })
    return () => { alive = false }
  }, [sel, toast])

  // Esc 关闭详情(仅在本视图可见时生效)。选中项变化时重置回第一个页签。
  useEffect(() => { setTab('info') }, [sel?.id, sel?.where])
  useEffect(() => {
    if (!sel || view !== 'record') return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') { e.preventDefault(); setSel(null) }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [sel, view])

  // ---- 宽度分配:列表吃剩余宽度,详情按占比取宽 + 5px 分隔条(同接口日志页) ----
  const splitRef = useRef<HTMLDivElement>(null)
  const [boxW, setBoxW] = useState(0)
  useEffect(() => {
    const el = splitRef.current
    if (!el) return
    const ro = new ResizeObserver(() => setBoxW(el.clientWidth))
    ro.observe(el)
    setBoxW(el.clientWidth)
    return () => ro.disconnect()
  }, [])
  const [ratio, setRatio] = useState(() => {
    const v = Number(localStorage.getItem('tt.recordRatio'))
    return v >= RATIO_MIN && v <= RATIO_MAX ? v : 0.5
  })
  const splittable = Math.max(1, boxW - RESIZER_W)
  const detailW = Math.round(splittable * ratio)

  const onDetailResizeDown = (e: React.MouseEvent<HTMLDivElement>) => {
    e.preventDefault()
    const el = e.currentTarget
    el.classList.add('dragging')
    const startX = e.clientX
    const startRatio = ratio
    let latest = ratio
    const move = (ev: MouseEvent) => {
      // 分隔线跟手:向左拖 = 详情面板变宽(与 VS Code 侧边栏一致);只改占比,总量不变
      const box = splitRef.current?.clientWidth || 0
      const sp = Math.max(1, box - RESIZER_W)
      latest = Math.min(RATIO_MAX, Math.max(RATIO_MIN, startRatio - (ev.clientX - startX) / sp))
      setRatio(latest)
    }
    const up = () => {
      el.classList.remove('dragging')
      localStorage.setItem('tt.recordRatio', String(latest))
      window.removeEventListener('mousemove', move)
      window.removeEventListener('mouseup', up)
    }
    window.addEventListener('mousemove', move)
    window.addEventListener('mouseup', up)
  }

  const onRefresh = () => {
    setQ((s) => (s.ent === ent.trim() ? s : { ...s, ent: ent.trim() }))
    setTick((t) => t + 1)
    loadEnvs()
  }

  const copyRow = async (id: string) => {
    try {
      await navigator.clipboard.writeText(id)
      setCopied({ id, state: 'ok' })
    } catch {
      setCopied({ id, state: 'fail' })
    }
    window.setTimeout(() => setCopied((c) => (c && c.id === id ? null : c)), 1500)
  }

  // 起一轮录制。已经在录时服务端回 409 与一句中文原因,原样弹出来 ——
  // 所以这里只把"点不了"的两种情形(没环境 / 没作业名)挡在前面。
  const onStart = async () => {
    if (!q.env) {
      toast.error('请先选环境', { description: '录制要落在某个环境上(列表也按它过滤)。' })
      return
    }
    const p = prog.trim()
    if (!p) {
      toast.error('请填作业名', { description: '要录哪个作业(如 bsft001_wf);它落在哪个企业只由 ENT 决定。' })
      return
    }
    setBusy('start')
    try {
      const r = await api.recordingStart({ env: q.env, prog: p, ent: ent.trim() })
      const id = r.recording?.id || ''
      toast.success(`已开始录制:${id}`, {
        description: `作业 ${r.recording?.prog || p}${r.log ? `;日志 ${r.log}` : ''}`,
        duration: 8000, // 录制 ID 多留一会儿:默认 3.5 秒常常还没看清就没了
      })
      await load(q)
    } catch (e: any) {
      toast.error('开始录制失败', { description: e?.message || String(e) })
    } finally {
      setBusy('')
    }
  }

  // 停一轮:收工 = 请常驻进程自己退(它会落身份 + 搬进长期库)+ 收掉调试会话。
  // 收工是异步的,所以刷新后可能还看不到那一件 —— 这是正常的,再点一次刷新即可。
  const onStop = async () => {
    setBusy('stop')
    try {
      const r = await api.recordingStop()
      const id = r.stopped?.id || ''
      toast.success(`已停止录制:${id}`, {
        description: `${r.stopped?.prog || ''} 已收工,录制件落到长期库/缓存(稍等再点刷新)`,
      })
      await load(q)
    } catch (e: any) {
      toast.error('停止录制失败', { description: e?.message || String(e) })
    } finally {
      setBusy('')
    }
  }

  // 删长期库里的那一份(入口在详情面板里:删除是看着这一份做的决定,不在列表里误点)。
  // 确认这一步由 AlertDialog 承担(与「会话」面板同一个基元),确认之后才带 ?confirm=1 发出去
  // —— 录制不可再生,不给误点留路。
  const doDelete = async () => {
    const row = pendingDel
    setPendingDel(null)
    if (!row) return
    try {
      const r = await api.recordingDelete(row.id)
      toast.success(`已删除录制:${r.id}`, { description: `清除 ${r.files} 个文件,释放 ${fmtBytes(r.bytes)}` })
      setSel(null)
      await load(q)
    } catch (e: any) {
      toast.error('删除录制失败', { description: e?.message || String(e) })
    }
  }

  const entDirty = ent.trim() !== q.ent
  const cp = copied && sel && copied.id === sel.id ? copied.state : null

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col">
      {/* 工具条一行(条件不多):左边 = 列表条件(环境/ENT/刷新),右边 = 起录/停录。
           每个条件都是「字段名 + 控件」两件东西,字段名在左、控件在右 —— 字段名不写进
           placeholder(那样看不出哪里是说明、哪里能打字;与接口日志页的 QueryField 同一口径)。
           下缘分割线与列表连成整体;窄了各自内部换行(flex-wrap),与接口日志页一致。 */}
      <div className="flex shrink-0 flex-wrap items-center gap-2 border-b border-border px-2 py-2">
        <span className="shrink-0 text-xs text-muted-foreground"
          title="环境:列表按它过滤,起录也用它(作业落在哪台机器)">环境</span>
        {envNames.length > 0 ? (
          <Select value={q.env} onValueChange={(v) => setQ((s) => ({ ...s, env: v }))}>
            <SelectTrigger className="h-7 w-40 text-xs"
              title="环境:列表按它过滤,起录也用它(作业落在哪台机器)">
              <SelectValue placeholder="选择" />
            </SelectTrigger>
            <SelectContent>
              {envNames.map((n) => <SelectItem key={n} value={n}>{n}</SelectItem>)}
            </SelectContent>
          </Select>
        ) : (
          <span className="text-xs text-muted-foreground">尚未配置环境(设置 → 站点管理)</span>
        )}
        <span className="shrink-0 text-xs text-muted-foreground"
          title="企业编号(ENT):起录时决定用哪个企业的账号连库;留空 = 沿用登录 shell 的 $TOPENT。作为列表条件时留空 = 不按 ENT 过滤">ENT</span>
        <Input
          value={ent}
          onChange={(e) => setEnt(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') onRefresh() }}
          placeholder="留空 = 沿用 $TOPENT"
          title="企业编号(ENT):起录时决定用哪个企业的账号连库;留空 = 沿用登录 shell 的 $TOPENT。作为列表条件时留空 = 不按 ENT 过滤。回车生效"
          className="h-7 w-40 text-xs"
        />
        <button onClick={onRefresh} disabled={loading} title="按当前环境/ENT 重新拉列表与「在录吗」"
          className="p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:opacity-40">
          <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
        </button>
        {entDirty && (
          <span className="text-[11px] text-amber-600 dark:text-amber-400">ENT 已改,点刷新或回车按新值过滤</span>
        )}
        {cur && (
          <span className="flex min-w-0 items-center gap-1 text-xs text-amber-600 dark:text-amber-400"
            title={`这一轮还活着(会话登记里有它)。停止录制会让它收工并落身份:${PATH_LABEL[cur.path || ''] || cur.path || ''}`}>
            <Video className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">正在录:{cur.id}({cur.prog || '-'})</span>
          </span>
        )}
        <div className="ml-auto flex items-center gap-2">
          <span className="shrink-0 text-xs text-muted-foreground"
            title="要录的作业编号。起录走的是服务端另起的 `tt fep rec start`,与调试会话互不干扰">作业名</span>
          <Input
            value={prog}
            onChange={(e) => setProg(e.target.value)}
            placeholder="如 bsft001_wf"
            title="要录的作业编号。起录走的是服务端另起的 `tt fep rec start`,与调试会话互不干扰"
            className="h-7 w-40 text-xs"
          />
          <button
            onClick={() => void onStart()}
            disabled={!!busy || !q.env || !prog.trim() || !!cur}
            title={cur
              ? `已经在录了(${cur.id}):先停掉它再起新的(两个会话会抢同一组端口)`
              : '起一轮录制:服务端会另起一个脱离的进程,等它登记(秒级到几十秒)'}
            className="inline-flex shrink-0 items-center gap-1 border border-emerald-500/20 px-2 py-0.5 text-xs text-emerald-600 hover:bg-emerald-500/10 disabled:pointer-events-none disabled:opacity-40 dark:text-emerald-400">
            <Play className="h-3.5 w-3.5" />
            {busy === 'start' ? '正在起…' : '开始录制'}
          </button>
          <button
            onClick={() => void onStop()}
            disabled={!!busy || !cur}
            title={cur ? `停止并收工:${cur.id}(它会落身份并搬进长期库)` : '现在没有在录的会话'}
            className="inline-flex shrink-0 items-center gap-1 border border-border px-2 py-0.5 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:pointer-events-none disabled:opacity-40">
            <Square className="h-3.5 w-3.5" />
            {busy === 'stop' ? '正在停…' : '停止录制'}
          </button>
        </div>
      </div>

      {/* 查询/操作失败一律走 toast(与接口日志页同口径),列表区不内联错误行 —— 那会把表格挤动 */}

      {/* 左右分配:左侧列表 flex-1 吃掉剩余宽度,右侧详情按占比取宽 + 5px 分隔条。
          min-w-0/overflow-hidden 锁住容器宽度:面板是 shrink-0,不加锁会把容器一路撑宽
          (宽度反馈环 → 整页横向溢出)。 */}
      <div ref={splitRef} className="flex min-h-0 min-w-0 flex-1 overflow-hidden">
        <div className="flex min-w-0 flex-1 flex-col overflow-hidden bg-background">
          <div className="min-h-0 flex-1 overflow-auto">
            {/* table-fixed:列宽交给表头/单元格上的 w-* ,留空的那列吃剩余宽度;
                container=false 不套 shadcn Table 默认的 overflow-x-auto(它会让 th 的 sticky 失效);
                border-separate + 单元格只留"右+下"框线,理由与接口日志页列表相同。
                详情面板打开时列表只剩一半宽,所以列要跟着收(与日志页的 COLS_COMPACT 同一个道理):
                面板里已经逐项列着的(环境·区域·ENT / 时间 / 体积 / 指纹)在这里让位,只留
                "这是哪一份"的四项 + 操作 —— 不然固定列宽加起来就吃掉整张表,操作列会被挤成 0 宽。
                宽度也随之压一档,让剩下的列还有地方放(每格都 truncate,不会撑破)。 */}
            <Table container={false} className="table-fixed border-separate border-spacing-0 text-xs">
              <TableHeader>
                <TableRow className="hover:bg-transparent">
                  <TableHead className={`sticky top-0 z-10 border-t-0 border-l-0 bg-background ${sel ? 'w-40' : 'w-48'}`}
                    title="录制 ID:命令行 `tt fep rec show/rm/reproduce <ID>` 用的就是它">录制 ID</TableHead>
                  <TableHead className={`sticky top-0 z-10 border-t-0 border-l-0 bg-background ${sel ? 'w-14' : 'w-16'}`}
                    title="长期库:不自动清理,可在详情里手动删;缓存:可再生的中间数据,随统一清理框架按年龄删">住哪</TableHead>
                  <TableHead className={`sticky top-0 z-10 border-t-0 border-l-0 bg-background ${sel ? 'w-20' : 'w-24'}`}
                    title="这一次作业是怎么起来的。回放闸门按它比对:路径不同的录像不能用">启动路径</TableHead>
                  <TableHead className={`sticky top-0 z-10 border-t-0 border-l-0 bg-background ${sel ? 'w-24' : 'w-32'}`}>作业</TableHead>
                  {!sel && (
                    <>
                      <TableHead className="sticky top-0 z-10 w-36 border-t-0 border-l-0 bg-background"
                        title="环境 / 区域 / ENT(企业编号)">环境/区域/ENT</TableHead>
                      <TableHead className="sticky top-0 z-10 w-36 border-t-0 border-l-0 bg-background">时间</TableHead>
                      <TableHead className="sticky top-0 z-10 w-20 border-t-0 border-l-0 bg-background">体积</TableHead>
                      <TableHead className="sticky top-0 z-10 w-48 border-t-0 border-l-0 bg-background"
                        title="编号表摘要:sha256 前 16 位(条数)。整体比对用它,不看数字也不看声明顺序">指纹</TableHead>
                    </>
                  )}
                  <TableHead className="sticky top-0 z-10 border-t-0 border-l-0 bg-background">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {rows.map((r) => {
                  const cpR = copied?.id === r.id ? copied.state : null
                  const selected = sel?.id === r.id && sel?.where === r.where
                  return (
                    <TableRow key={`${r.where}:${r.id}`} data-state={selected ? 'selected' : undefined}
                      onClick={() => setSel(r)}
                      title={selected ? undefined : '点击在右侧查看这份录制的身份(启动路径/作业/ENT/编号表/连接清单)'}
                      className="h-7 cursor-pointer hover:bg-accent/40 data-[state=selected]:bg-sky-500/10">
                      <TableCell className={`border-t-0 border-l-0 px-2 py-0 font-mono ${sel ? 'w-40' : 'w-48'}`} title={r.id}>
                        <span className="block truncate font-mono">{r.id}</span>
                      </TableCell>
                      <TableCell className={`border-t-0 border-l-0 px-2 py-0 ${sel ? 'w-14' : 'w-16'}`} title={r.where}>
                        <span className="block truncate">{r.where || '-'}</span>
                      </TableCell>
                      <TableCell className={`border-t-0 border-l-0 px-2 py-0 ${sel ? 'w-20' : 'w-24'}`}
                        title={PATH_LABEL[r.path || ''] || (r.path || '身份里没写路径')}>
                        <span className="block truncate font-mono text-muted-foreground">{r.path || '-'}</span>
                      </TableCell>
                      <TableCell className={`border-t-0 border-l-0 px-2 py-0 ${sel ? 'w-24' : 'w-32'}`} title={r.prog || ''}>
                        <span className="block truncate text-sky-600 dark:text-sky-400">{r.prog || '-'}</span>
                      </TableCell>
                      {!sel && (
                        <>
                          <TableCell className="border-t-0 border-l-0 px-2 py-0"
                            title={`环境 ${r.env || '-'} · 区域 ${r.zone || '-'} · ENT ${r.ent || '-'}`}>
                            <span className="block truncate font-mono text-[11px] text-muted-foreground">
                              {r.env || '-'}/{r.zone || '-'}/{r.ent || '-'}
                            </span>
                          </TableCell>
                          <TableCell className="border-t-0 border-l-0 px-2 py-0" title={r.started || '身份里没记时间'}>
                            <span className="block truncate font-mono text-[11px] text-muted-foreground">{fmtTime(r.started)}</span>
                          </TableCell>
                          <TableCell className="border-t-0 border-l-0 px-2 py-0" title={`${r.bytes} 字节`}>
                            <span className="font-mono text-[11px] text-muted-foreground">{fmtBytes(r.bytes)}</span>
                          </TableCell>
                          <TableCell className="border-t-0 border-l-0 px-2 py-0"
                            title={r.fingerprint?.hash
                              ? `编号表 ${r.fingerprint.total} 条;sha256 前 16 位 ${r.fingerprint.hash}`
                              : '这一轮没录到界面(作业没连上来,或刚起来就断了)'}>
                            <span className="block truncate font-mono text-[11px] text-muted-foreground">{fmtFP(r.fingerprint)}</span>
                          </TableCell>
                        </>
                      )}
                      <TableCell className="border-t-0 border-l-0 px-1 py-0">
                        {/* 行内只留最常用那一下:把 ID 复制走交给 AI / 命令行。
                            删除挪进了详情面板 —— 那是在看着这一份做决定,不该在列表里误点。
                            stopPropagation:别让点按钮顺带把行也选中。 */}
                        <button
                          onClick={(e) => { e.stopPropagation(); void copyRow(r.id) }}
                          title="复制这份录制的 ID。把它发给 AI / 贴到命令行(tt fep rec show/rm/reproduce <ID>)都用它"
                          className={`inline-flex shrink-0 items-center gap-1 border px-1.5 py-0 text-[11px] transition-colors ${
                            cpR === 'fail'
                              ? 'border-red-500/20 text-red-600 dark:text-red-400'
                              : 'border-border text-muted-foreground hover:bg-accent hover:text-foreground'
                          }`}>
                          {cpR === 'ok' ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                          {cpR === 'ok' ? '已复制' : cpR === 'fail' ? '复制失败' : '复制 ID'}
                        </button>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
            {rows.length === 0 && !loading && (
              <div className="space-y-1 p-4 text-center text-xs text-muted-foreground">
                {envNames.length === 0 ? (
                  <>
                    <div>尚未配置环境(设置 → 站点管理)。</div>
                    <div>配好环境后,这里才有可录、可看的东西。</div>
                  </>
                ) : (
                  <>
                    <div>{q.env || '当前条件'} 下还没有录制件。</div>
                    <div>填好上方的作业名,点「开始录制」录一轮;录完收工后它就会出现在这里。</div>
                  </>
                )}
              </div>
            )}
            {rows.length === 0 && loading && (
              <div className="p-4 text-center text-xs text-muted-foreground">加载中…</div>
            )}
          </div>
        </div>

        {/* 详情(右列,仅选中一行后出现):贴右缘整高嵌入,左侧 1px 分割线即拖拽把手 */}
        {sel && (
          <>
            <div className="col-resizer self-stretch" onMouseDown={onDetailResizeDown} title="拖拽分配宽度" />
            <div style={{ width: boxW ? detailW : '50%' }} className="flex shrink-0 flex-col overflow-hidden bg-card">
              <div className="flex h-8 shrink-0 items-center gap-1 border-b border-border px-2">
                <button onClick={() => setSel(null)} title="关闭详情(Esc)"
                  className="p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
                  <X className="h-3.5 w-3.5" />
                </button>
                {([['info', '基本信息'], ['conns', '连接'], ['raw', '身份文件']] as const).map(([k, label]) => (
                  <button key={k} onClick={() => setTab(k)}
                    className={`inline-flex items-center gap-1 px-2 py-0.5 text-xs ${tab === k ? 'bg-accent text-foreground' : 'text-muted-foreground hover:text-foreground'}`}>
                    {label}
                    {k === 'conns' && !metaLoading && !!selMeta?.conns?.length && (
                      <span className="text-[10px] text-muted-foreground">{selMeta.conns.length}</span>
                    )}
                  </button>
                ))}
                {/* 把这行的 ID 交给 AI:AI 用 tt fep rec reproduce <ID> 重放这一份 */}
                <button
                  onClick={() => void copyRow(sel.id)}
                  title="复制这份录制的 ID。把它发给 AI,AI 就能用 tt fep rec reproduce <ID> 重放这一份"
                  className={`ml-auto inline-flex shrink-0 items-center gap-1 border px-2 py-0.5 text-xs transition-colors ${
                    cp === 'fail'
                      ? 'border-red-500/20 text-red-600 dark:text-red-400'
                      : 'border-border text-muted-foreground hover:bg-accent hover:text-foreground'
                  }`}>
                  {cp === 'ok' ? <Check className="h-3.5 w-3.5" /> : <Fingerprint className="h-3.5 w-3.5" />}
                  {cp === 'ok' ? '已复制' : cp === 'fail' ? '复制失败' : '复制 ID'}
                </button>
                {/* 删除只对长期库里的那些开放:缓存里的那份归统一清理框架管(服务端会拒),
                    放在这里只会得到一个 409。 */}
                {sel.where === '长期库' && (
                  <button
                    onClick={() => setPendingDel(sel)}
                    title="从长期库里删掉这一份。录制不可再生,会先弹二次确认"
                    className="inline-flex shrink-0 items-center gap-1 border border-red-500/20 px-2 py-0.5 text-xs text-red-600 transition-colors hover:bg-red-500/10 dark:text-red-400">
                    <Trash2 className="h-3.5 w-3.5" />
                    删除
                  </button>
                )}
              </div>
              {/* 没录到界面的那一份要先说清:回放闸门会拒(没有编号表可比) */}
              {!sel.fingerprint?.hash && (
                <div className="shrink-0 border-b border-border bg-amber-500/10 px-2.5 py-1 text-[11px] text-amber-600 dark:text-amber-400">
                  这一轮没录到界面(作业没连上来,或刚起来就断了):没有编号表,回放时闸门会拒。
                </div>
              )}
              {/* 基本信息是普通内容流,需要外层滚动;身份文件是长文本,同样交给外层滚 */}
              <div className="min-h-0 flex-1 overflow-auto">
                {tab === 'info' && <InfoBody row={sel} meta={selMeta} loading={metaLoading} />}
                {tab === 'conns' && <ConnsBody meta={selMeta} loading={metaLoading} />}
                {tab === 'raw' && <RawBody meta={selMeta} loading={metaLoading} />}
              </div>
            </div>
          </>
        )}
      </div>

      {/* 删除的二次确认(与「会话」面板同一个基元:shadcn AlertDialog) */}
      <AlertDialog open={pendingDel !== null} onOpenChange={(o) => { if (!o) setPendingDel(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除录制「{pendingDel?.id || ''}」?</AlertDialogTitle>
            <AlertDialogDescription>
              将从长期库里删掉这一份({pendingDel?.prog || '未知作业'},{fmtBytes(pendingDel?.bytes ?? 0)})。
              录制不可再生 —— 要再拿到同一份,得重现一次问题再录一遍;此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void doDelete()}>确定删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

// 基本信息:两列键值(与接口日志页的 InfoBody 同款;字段名在左、值在右)。
// 值的口径与 Go 侧身份文件一一对应(internal/fep/meta.go),缺的那一项照实显示 '-',
// 不补默认值 —— "身份里没记"本身就是回放时要看的一条信息。
function InfoBody({ row, meta, loading }: { row: RecordingRow; meta: RecordingMeta | null; loading: boolean }) {
  const bin = meta?.binary
  const binText = bin && (bin.ref || bin.size || bin.modTime)
    ? [bin.ref, bin.size ? `${bin.size} 字节` : '', bin.modTime || ''].filter(Boolean).join(' / ')
    : '（没探到）'
  const dur = secsBetween(meta?.started || row.started, meta?.finished)
  const rows: [string, string, string?][] = [
    ['录制 ID', row.id, '命令行 tt fep rec show / rm / reproduce <ID> 用的就是它'],
    ['住哪', row.where, RETENTION_LABEL[row.retention || ''] || '身份里没说留哪'],
    ['目录', row.dir, '这份录制文件的落点(长期库按 环境/ENT 分层)'],
    ['启动路径', row.path || '-', PATH_LABEL[row.path || ''] || '身份里没写启动路径'],
    ['作业', row.prog || '-'],
    ['环境', row.env || '-'],
    ['区域', row.zone || '-'],
    ['企业编号(ENT)', row.ent || '-', '空 = 起录时沿用登录 shell 的 $TOPENT'],
    ['SSH 主机', meta?.host || '-', '不含账号与端口'],
    ['画面接线', meta?.fglserver || '-', 'FGLSERVER:这一轮的界面画到哪(本机 IP:代理端口-6400)'],
    ['开始时间', fmtTime(meta?.started || row.started)],
    ['结束时间', fmtFinished(meta?.finished), '收工那一刻(它同时也是身份落盘的时刻)'],
    ['时长(秒)', dur ? String(dur) : '-'],
    ['体积', `${fmtBytes(row.bytes)}（${row.bytes} 字节）`, '逐连接的 rt2gdc / gdc2rt / timeline 三个文件之和'],
    ['连接数', String(row.conns), '闸门按**连接**比对:一条对得上就算同一棵界面树'],
    ['指纹', fmtFP(row.fingerprint),
      row.fingerprint?.hash
        ? `编号表 ${row.fingerprint.total} 条;sha256 前 16 位 ${row.fingerprint.hash}`
        : '没有编号表(这一轮没录到界面)'],
    ['编译件', binText, '作业 .42r 的大小与修改时间;走 r.r 时无从探测,照实空着'],
    ['归属', ORIGIN_LABEL[row.origin || ''] || '-', '用户主动录的进长期库,系统自动录的进缓存'],
    ['备注', meta?.note || '-'],
  ]
  return (
    <div className="p-2 text-xs">
      {rows.map(([k, v, tip]) => (
        <div key={k} className="flex gap-2 border-b border-border/60 py-1" title={tip}>
          <span className="w-28 shrink-0 text-muted-foreground">{k}</span>
          <span className="min-w-0 flex-1 break-all text-foreground">
            {v || '-'}
            {k === '连接数' && loading && <span className="text-muted-foreground">（读身份中…）</span>}
          </span>
        </div>
      ))}
    </div>
  )
}

// 连接:一份录制里可能有多条连接,而且可以是**两棵不同的界面树**(实测 r.d 那一轮的
// #1 是调试器自己的界面 298 条、#2 才是作业 3976 条)。所以这里逐条摊开:
// 回放时对不上的那一条要能一眼看出是哪一条。
function ConnsBody({ meta, loading }: { meta: RecordingMeta | null; loading: boolean }) {
  if (loading && !meta) return <div className="p-3 text-xs text-muted-foreground">读身份中…</div>
  if (!meta?.conns?.length) {
    return (
      <div className="p-3 text-xs leading-5 text-muted-foreground">
        这一份没有连接清单 —— 老录制件没写身份,或者这一轮没录到任何连接。
      </div>
    )
  }
  return (
    <div className="p-2 text-xs">
      <div className="px-0.5 pb-2 text-[11px] leading-5 text-muted-foreground">
        回放闸门按连接比对:下面每一条连接各自一张编号表,只要有一条对得上就算同一棵界面树。
        同一个会话里完全可能有多条连接,而且是两棵不同的树。
      </div>
      {meta.conns.map((c) => (
        <div key={c.id} className="mb-2 border border-border/60">
          <div className="flex flex-wrap items-center gap-2 border-b border-border/60 bg-muted/40 px-2 py-1">
            <span className="font-mono text-foreground">#{c.id}</span>
            <span className="text-muted-foreground">{c.remote || '远端未知'}</span>
            <span className="ml-auto font-mono text-[11px] text-muted-foreground">
              运行→前端 {fmtBytes(c.bytesRT || 0)} ／ 前端→运行 {fmtBytes(c.bytesFE || 0)}
            </span>
          </div>
          <div className="space-y-0.5 px-2 py-1 font-mono text-[11px] leading-5 text-muted-foreground">
            <div className="break-all">AUI（运行期→前端）: {c.rt2gdc || '-'}</div>
            <div className="break-all">事件（前端→运行期）: {c.gdc2rt || '-'}</div>
            <div className="break-all">时间线: {c.timeline || '-'}</div>
          </div>
        </div>
      ))}
    </div>
  )
}

// 身份文件:原始 JSON。给"闸门为什么这么判"这类问题留一个能自己看的窗口 ——
// 上面两页是给人读的摘要,这里是原样。
function RawBody({ meta, loading }: { meta: RecordingMeta | null; loading: boolean }) {
  if (loading && !meta) return <div className="p-3 text-xs text-muted-foreground">读身份中…</div>
  if (!meta) {
    return (
      <div className="p-3 text-xs leading-5 text-muted-foreground">
        没有身份文件（.meta.json）—— 早期录的、或者进程被硬杀时录的件就是这样：
        指纹还能从录制件的字节算出来，但启动路径/作业/ENT 这些只有身份里才有。
      </div>
    )
  }
  return (
    <pre className="p-2 font-mono text-[11px] leading-5 break-all whitespace-pre-wrap text-muted-foreground">
      {JSON.stringify(meta, null, 2)}
    </pre>
  )
}
