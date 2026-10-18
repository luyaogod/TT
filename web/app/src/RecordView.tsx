// 「录制」面板:列出某个环境/ENT 下的录制件(长期库 + 缓存),点开一份看它的身份,并起一轮/停一轮。
//
// 为什么起录/停录只是两个按钮、列表要手动刷新:一轮录制 = 一个隔离 GDC + 一条代理 +
// 一个常驻 Runtime,服务端会**另起一个脱离的进程**(`tt fep rec start`,见
// internal/web/recordings.go 的文件头注释)。界面这边能做的只有"发请求 + 拉列表",
// 真正的活在子进程里 —— 所以起录要等它登记(秒级到几十秒),停录要等它收工落身份。
//
// 列表只摊开**六列**（用户 2026-10-16 定的）：`ID / 时间 / 位置 / 作业 / 环境 / 体积`。
// 挑录像时这几项够回答"这是哪一轮、在哪、多大"。原先还有「启动路径」与「指纹」两列，
// 那两项决定"这份能不能回放"（回放闸门按启动路径与指纹比对，见 internal/fep/path.go）——
// 它们没丢，只是搬到了**明细页的检查器**里：挑的时候不需要逐张比对指纹。
//
// **点一行 → 整屏换成明细页**(`#record/<id>`,见 RecordingDetail.tsx):条带 + 全文流记录表
// + 检查器。身份、回放、复制 ID、删除这些"看着这一份做决定"的动作也搬到了明细页的头上 ——
// 一份录制只该有一个"我正看着它"的地方,列表只管挑。
import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react'
import { RefreshCw } from 'lucide-react'
import { useStore } from './store'
import { api, type RecordingCurrent, type RecordingRow } from './api'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../shared/ui'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
  Checkbox,
} from '../../shared/ui-radix'
import { useToast } from './toast'
import { RecordingDetail } from './RecordingDetail'

const ORIGIN_LABEL: Record<string, string> = { user: '用户主动', auto: '系统自动' }
const RETENTION_LABEL: Record<string, string> = {
  library: '长期库(不自动清理)',
  cache: '缓存目录(随统一清理按年龄删)',
}

// 详情面板的宽度分配:与接口日志页同一套做法(拖分割条改占比、占比记在本地)。
// 三个数是那边同一份口径,量很小,先各留一份;哪天有第三处再抽出去。
const RATIO_MIN = 0.25
const RATIO_MAX = 0.75
const RESIZER_W = 1 // .col-resizer 的宽度(见 index.css:1px 线 + 透明溢出命中区)
// 两侧各要多少才"看得成"：低于这个总宽就不再左右分栏（见 narrowSplit）。
const MIN_LIST_W = 280
const MIN_DETAIL_W = 300
const MIN_DETAIL_ABS = 240

// 体积:与设置页那份同一口径(那边是视图内的本地函数,没有导出,这里照抄一份)。
// 1024 进位:1048575 字节要说 "1023.9 KB" 而不是 "1024.0 KB"(先算到哪一档、再取整)。
function fmtBytes(n: number): string {
  if (!n) return '0 B'
  if (n >= 1 << 30) return (n / (1 << 30)).toFixed(2) + ' GB'
  if (n >= 1 << 20) {
    const mb = n / (1 << 20)
    // 四舍五入到一位后正好封顶 1024 的（1048575 这种），按 MB 说
    return (mb >= 1023.95 ? '1.0 MB' : mb.toFixed(1) + ' MB')
  }
  if (n >= 1 << 10) {
    const kb = n / (1 << 10)
    return (kb >= 1023.95 ? '1.0 MB' : kb.toFixed(1) + ' KB')
  }
  return n + ' B'
}

// 时间:RFC3339 → 本地时间。空 = 身份里没记时间(老件/没收到),不猜。
function fmtTime(s?: string): string {
  // Go 的零值时间(`0001-01-01T00:00:00Z`)是**照实**出现在 JSON 里的(结构体字段的
  // omitempty 管不到 time.Time):那不是"时间",是"没记"。渲染成 1/1/1 07:36:42 只会让人
  // 以为真有个一千年前的会话(实测单测抓到这个)。`fmtFinished` 早就有这道护栏,这里补上。
  if (!s || s.startsWith('0001-')) return '-'
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

export function RecordView() {
  const view = useStore((s) => s.view)
  // 环境与会话都取自**左侧那个会话面板**(用户 2026-10-15 要求):
  // 有了独立的会话栏,这里再摆一个环境下拉就是在两个地方管同一件事。
  // 于是本面板不再自己选环境 —— 列表按"当前会话的环境"过滤,起录也用同一个。
  const sessionEnv = useStore((s) => s.sessionEnv)
  // 列表按**当前环境**过滤:环境页的「FEP录像」Tab 是在看这个环境,不一定有会话。
  // 起一轮录制仍跟会话走(`sessionEnv`)—— 那是"用哪个登录态/企业",与"看哪个环境的件"是两件事。
  const activeEnv = useStore((s) => s.activeEnv)
  const listEnv = activeEnv || sessionEnv
  const toast = useToast()

  // 「刷新」按钮的强制重拉信号(条件没变时也要重新拉一次)
  const [tick, setTick] = useState(0)

  const [rows, setRows] = useState<RecordingRow[]>([])
  const [cur, setCur] = useState<RecordingCurrent | null>(null)
  const [loading, setLoading] = useState(false)
  // 列表读取失败的原因（非空 = 界面要给"读不到"而不是"没有"）。
  const [loadErr, setLoadErr] = useState('')

  // ── 管理模式（用户 2026-10-18 定）────────────────────────────────────────
  //
  // 列表默认只读（点一行=看明细）。点「管理」进管理模式：每行左侧出现勾选框（表头是全选），
  // 右上角变成「删除(N)」「退出」。删除是**彻底删除**（长期库 + 缓存一起清），
  // 所以一定要过一道确认对话框。
  const [manage, setManage] = useState(false)
  const [picked, setPicked] = useState<string[]>([])
  const [askDel, setAskDel] = useState<string[] | null>(null)
  const [deleting, setDeleting] = useState(false)
  const togglePick = (id: string) =>
    setPicked((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]))
  const exitManage = () => { setManage(false); setPicked([]) }

  // 明细页:`recordId` 非空就是"正在看这一份"(它是 store 里的状态,URL 片段由 App 同步)。
  const recordId = useStore((s) => s.recordId)
  const setRecordId = useStore((s) => s.setRecordId)

  // 列表的滚动位置:进明细再回来时要**回到原处**(keep-alive 语义)。
  // 两种做法都上:列表留在 DOM 里(只 display:none,浏览器一般会保住 scrollTop),
  // 再加一份显式存取 —— 后者不受浏览器行为影响,而"回来以后滚到别处去了"是实测会被抱怨的那类小事。
  const listScrollRef = useRef<HTMLDivElement | null>(null)
  const savedScroll = useRef(0)
  const openDetail = (r: RecordingRow) => {
    savedScroll.current = listScrollRef.current?.scrollTop || 0
    setRecordId(r.id)
  }
  useLayoutEffect(() => {
    if (recordId) return
    if (listScrollRef.current) listScrollRef.current.scrollTop = savedScroll.current
  }, [recordId])

  // 列表与「现在在录吗」一起拉:它们是同一屏的两个事实,分两次进入会看到互相矛盾的画面
  // (列表里已有刚收工的件、而"正在录"还挂着;或反过来)。
  // env 空 = 不按环境过滤(还没连会话时也能浏览全部),那时列表里的「环境/区域/ENT」列
  // 就是唯一的归属线索。
  const load = useCallback(async (env: string) => {
    setLoading(true)
    try {
      const [list, now] = await Promise.all([
        api.recordings(env, ''),
        api.recordingCurrent(),
      ])
      const next = list.recordings || []
      setRows(next)
      setCur(now.recording || null)
      setLoadErr('')
    } catch (e: any) {
      // 错误要留在界面上（loadErr），不能只发一条会消失的 toast：否则列表区会显示成
      // "还没有录制件"——那是一个**事实判断**，与"这次没读到"完全不是一回事（实测）。
      setLoadErr(e?.message || String(e))
      toast.error('读取录制失败', { description: e?.message || String(e) })
    } finally {
      setLoading(false)
    }
  }, [toast])

  // 勾选跟着列表收敛：删掉/消失的那些 id 不该继续占着"已选 N"。
  useEffect(() => {
    setPicked((prev) => (prev.length === 0 ? prev : prev.filter((id) => rows.some((r) => r.id === id))))
  }, [rows])

  // 批量彻底删除。**先确认再发**（确认框在下面）：录制不可再生，删了只能重录。
  const doDelete = async (ids: string[]) => {
    if (ids.length === 0) return
    setDeleting(true)
    try {
      const r = await api.recordingsBulkDelete(ids)
      const bad = (r.items || []).filter((x) => x.failed)
      if (bad.length > 0) {
        toast.error(`删掉 ${r.deleted} 份，${bad.length} 份没删成`, {
          description: bad.map((x) => `${x.id}：${x.failed}`).join("\n"),
        })
      } else {
        toast.success(`已彻底删除 ${r.deleted} 份录制`, { description: '长期库与缓存都清掉了' })
      }
      exitManage()
      void load(listEnv)
    } catch (e: any) {
      toast.error('批量删除失败', { description: e?.message || String(e) })
    } finally {
      setDeleting(false)
    }
  }

  // 进面板就先拉一次。视图是 keep-alive(只切显隐、不卸载重挂),所以"进面板"这件事
  // 只能由 view 的变化来标识 —— 挂载那一刻 view 就已经是 record 了。
  // 会话换了环境(sessionEnv 变)也要跟着重拉:列表是按它过滤的。
  useEffect(() => {
    if (view !== 'record') return
    void load(listEnv)
  }, [view, listEnv, tick, load])

  // 「正在录/正在回放」这条状态要**自己收敛**。
  //
  // 为什么需要轮询：会话可能被界面之外的东西收掉（`tt fep session close`、别的进程、CLI）。
  // 那时 `.tt-fep.json` 已经没了、`/api/recordings/current` 已经是 null，而面板什么都不点
  // 就永远挂着"正在录:fep-x"（实测：外部收掉后 25 秒纹丝不动，0 个新请求）。
  // 只在**有会话**时轮询，且只问那个便宜端点（读登记，不列文件）；状态真的变了才动界面，
  // 会话消失时顺带把列表重拉一次（那一轮已经收工落库了）。
  useEffect(() => {
    if (view !== 'record' || !cur) return
    const t = window.setInterval(() => {
      void (async () => {
        try {
          const now = await api.recordingCurrent()
          const next = now.recording || null
          const same = !!next && next.id === cur.id
            && !!next.stale === !!cur.stale && (next.kind || '') === (cur.kind || '')
          if (same) return
          setCur(next)
          if (!next) void load(listEnv)
        } catch { /* 轮询失败不动界面:下一拍再试 */ }
      })()
    }, 5000)
    return () => window.clearInterval(t)
  }, [view, cur, load, listEnv])

  // 点开一行就取它的身份。取不到(没有 .meta.json 的老件)照实说,不假装有。
  // 列表与明细页共用同一个「在录吗」轮询:明细页自己不再轮询 —— 那是同一件事。

  const onRefresh = () => setTick((t) => t + 1)


  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col">
      {/* 工具条一行:**只有"录哪个作业"这一件事**。
          环境与企业编号本来在这里各有一个控件,现在都去掉了 —— 左侧独立的会话面板已经在管
          这两样(用户 2026-10-15:"因为我们有环境面板了"),同一个事实摆两处只会对不上。
          于是这一行只留:作业名 + 起录/停录 + 刷新,并且**靠左**(不再 ml-auto 顶到右边:
          左边的条件没了,再把人赶到右边纯属浪费)。
          字段名在左、控件在右,不写进 placeholder(与接口日志页的 QueryField 同一口径)。 */}
      {/* 这里**不再有起停操作**（用户 2026-10-18 定：FEP 的唯一入口是「创建会话 → FEP 录制会话」）
          —— 这一屏只做归档浏览。刷新留着：它只读，不改变任何东西。 */}
      {!recordId && (
      <div className="flex shrink-0 flex-wrap items-center gap-2 border-b border-border px-2 py-2">
        <span className="text-[11px] text-muted-foreground">
          录制从「创建会话 → FEP 录制会话」开始；这里只浏览录制件。
        </span>
        <span className="shrink-0 text-[11px] text-muted-foreground/70"
          title="列表按左侧选中的环境过滤;没有选中环境时列全部">
          {listEnv ? `环境 ${listEnv}` : '未选环境:列全部环境'}
        </span>
        <div className="ml-auto flex items-center gap-1.5">
          {!manage ? (
            <button onClick={() => setManage(true)} disabled={rows.length === 0}
              title="管理模式：勾选多份录制，一次彻底删除（长期库 + 缓存）"
              className="rounded-ui border border-border px-2 py-0.5 text-[11px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:opacity-40">
              管理
            </button>
          ) : (
            <>
              <button onClick={() => setAskDel(picked)} disabled={picked.length === 0 || deleting}
                title="彻底删除勾中的录制（长期库 + 缓存，不可恢复）"
                className="rounded-ui border border-red-500/40 bg-red-500/10 px-2 py-0.5 text-[11px] text-red-600 transition-colors hover:bg-red-500/20 disabled:opacity-40 dark:text-red-400">
                {deleting ? '删除中…' : `删除${picked.length ? `(${picked.length})` : ''}`}
              </button>
              <button onClick={exitManage} disabled={deleting}
                className="rounded-ui border border-border px-2 py-0.5 text-[11px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:opacity-40">
                退出
              </button>
            </>
          )}
          <button onClick={onRefresh} disabled={loading} title="重新拉列表"
            className="p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:opacity-40">
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>
      )}

      {/* 查询/操作失败一律走 toast(与接口日志页同口径),列表区不内联错误行 —— 那会把表格挤动 */}

      {/* 列表与明细页:**同一时刻只看得见一个**,但两个都留在 DOM 里。
          列表用 hidden 而不是卸载,是为了回来时滚动位置与筛选都还在(keep-alive 语义;
          显式存取 scrollTop 见上面那条 useLayoutEffect —— 不依赖浏览器对 display:none 的处理)。 */}
      <div className={recordId ? 'hidden' : 'flex min-h-0 min-w-0 flex-1 overflow-hidden'}>
        <div className="flex min-w-0 flex-1 flex-col overflow-hidden bg-background">
          <div ref={listScrollRef} className="min-h-0 flex-1 overflow-auto">
            {/* table-fixed:列宽交给表头/单元格上的 w-* ,留空的那列吃剩余宽度;
                container=false 不套 shadcn Table 默认的 overflow-x-auto(它会让 th 的 sticky 失效);
                border-separate + 单元格只留"右+下"框线,理由与接口日志页列表相同。
                列表是整屏的(明细页在另一屏),所以每一列都留得下 —— 不再有"详情面板打开时列表
                只剩一半宽"这件事,列宽因此是固定的。
                没有「操作」列(用户 2026-10-15 要求移除):复制 ID、删除、回放都在明细页头上 ——
                行只负责"点开看",不摆按钮(与接口日志页的列表同一形态)。 */}
            <Table container={false} className="table-fixed border-separate border-spacing-0 text-xs">
              <TableHeader>
                {/* 表头**有底色**（用户 2026-10-16）：它是 sticky 的，滚动时下面要有东西把它
                    与行分开；顺带一眼看出"这一栏是表头不是数据"。列这时只剩六条（用户定的：
                    ID / 时间 / 位置 / 作业 / 环境 / 体积）—— 原先的「启动路径」与「指纹」两列去掉了，
                    它们都还在：指纹与启动路径在**明细页的检查器**里（回放闸门照样按它们比对，
                    只是不再占列表的一列）。 */}
                <TableRow className="hover:bg-transparent">
                  {manage && (
                    <TableHead className="sticky top-0 z-10 w-9 border-t-0 border-l-0 bg-muted px-0 text-center">
                      <Checkbox
                        checked={rows.length > 0 && picked.length === rows.length ? true : picked.length ? 'indeterminate' : false}
                        onCheckedChange={(v) => setPicked(v === true ? rows.map((r) => r.id) : [])}
                        className="mx-auto cursor-pointer"
                        title={picked.length === rows.length ? '取消全选' : '全选（当前列表里的全部）'} />
                    </TableHead>
                  )}
                  <TableHead className="sticky top-0 z-10 w-52 border-t-0 border-l-0 bg-muted"
                    title="录制 ID:命令行 `tt fep rec show/rm/reproduce <ID>` 用的就是它">ID</TableHead>
                  <TableHead className="sticky top-0 z-10 w-44 border-t-0 border-l-0 bg-muted">时间</TableHead>
                  <TableHead className="sticky top-0 z-10 w-20 border-t-0 border-l-0 bg-muted"
                    title="长期库:不自动清理,可在明细页里手动删;缓存:可再生的中间数据,随统一清理框架按年龄删">位置</TableHead>
                  <TableHead className="sticky top-0 z-10 w-40 border-t-0 border-l-0 bg-muted">作业</TableHead>
                  <TableHead className="sticky top-0 z-10 border-t-0 border-l-0 bg-muted"
                    title="环境 / 区域 / ENT(企业编号)">环境</TableHead>
                  <TableHead className="sticky top-0 z-10 w-20 border-t-0 border-l-0 bg-muted">体积</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {rows.map((r) => {
                  return (
                    <TableRow key={`${r.where}:${r.id}`}
                      onClick={() => (manage ? togglePick(r.id) : openDetail(r))}
                      title={manage ? '勾选这一份（管理模式）' : '点开这一份的明细(条带 + 三层记录表 + 检查器):看它录到了什么、回放前核对身份'}
                      className={`h-7 cursor-pointer hover:bg-accent/40 ${manage && picked.includes(r.id) ? 'bg-accent/30' : ''}`}>
                      {manage && (
                        <TableCell className="w-9 border-t-0 border-l-0 px-0 text-center">
                          <Checkbox checked={picked.includes(r.id)}
                            onCheckedChange={() => togglePick(r.id)}
                            onClick={(e) => e.stopPropagation()}
                            className="mx-auto cursor-pointer" />
                        </TableCell>
                      )}
                      <TableCell className="w-52 border-t-0 border-l-0 px-2 py-0 font-mono" title={r.id}>
                        <span className="block truncate font-mono">{r.id}</span>
                      </TableCell>
                      {/* 时间放第二列：挑录像时"这是哪一轮"往往比"住哪"更先要看 */}
                      <TableCell className="w-44 border-t-0 border-l-0 px-2 py-0" title={r.started || '身份里没记时间'}>
                        <span className="block truncate font-mono text-[11px] text-muted-foreground">{fmtTime(r.started)}</span>
                      </TableCell>
                      <TableCell className="w-20 border-t-0 border-l-0 px-2 py-0"
                        title={`${r.where || '-'}｜长期库:不自动清理,在明细页里手动删;缓存:可再生的中间数据,随统一清理框架按年龄删`}>
                        <span className="block truncate">{r.where || '-'}</span>
                      </TableCell>
                      <TableCell className="w-40 border-t-0 border-l-0 px-2 py-0" title={r.prog || ''}>
                        <span className="block truncate text-sky-600 dark:text-sky-400">{r.prog || '-'}</span>
                      </TableCell>
                      <TableCell className="border-t-0 border-l-0 px-2 py-0"
                        title={`环境 ${r.env || '-'} · 区域 ${r.zone || '-'} · ENT ${r.ent || '-'}`}>
                        <span className="block truncate font-mono text-[11px] text-muted-foreground">
                          {r.env || '-'}/{r.zone || '-'}/{r.ent || '-'}
                        </span>
                      </TableCell>
                      <TableCell className="w-20 border-t-0 border-l-0 px-2 py-0" title={`${r.bytes} 字节`}>
                        <span className="font-mono text-[11px] text-muted-foreground">{fmtBytes(r.bytes)}</span>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
            {rows.length === 0 && !loading && (
              // 失败 ≠ 没有录制件。原来这里不分青红皂白地说"还没有录制件"，把一次接口失败
              // 说成了一个事实判断（实测：CDP 拦掉 /api/recordings* 后，界面照样宣称"还没有录制件"）。
              loadErr ? (
                <div className="flex flex-col items-center gap-2 p-4 text-center text-xs">
                  <div className="text-red-600 dark:text-red-400">读取录制列表失败</div>
                  <div className="max-w-2xl whitespace-pre-wrap text-muted-foreground">{loadErr}</div>
                  <button onClick={onRefresh}
                    className="rounded-ui border border-border px-2 py-1 text-muted-foreground hover:bg-accent hover:text-foreground">
                    重试
                  </button>
                </div>
              ) : (
                <div className="space-y-1 p-4 text-center text-xs text-muted-foreground">
                  <div>{listEnv ? `环境 ${listEnv} 下还没有录制件。` : '还没有录制件。'}</div>
                  <div>填好上方的作业名,点「开始录制」录一轮;录完收工后它就会出现在这里。</div>
                  {!sessionEnv && (
                    <div>起录需要一个会话环境:先在左侧选一个环境、建一条会话,连上之后才能起录。</div>
                  )}
                </div>
              )
            )}
            {rows.length === 0 && loading && (
              <div className="p-4 text-center text-xs text-muted-foreground">加载中…</div>
            )}
          </div>
        </div>
      </div>

      {/* 批量删除的确认框（用户 2026-10-18 定：需要确认）。
          用 AlertDialog 而不是 Dialog：这是一次"要做/不做"的决定，不该点遮罩就放过。 */}
      <AlertDialog open={askDel !== null} onOpenChange={(o) => { if (!o) setAskDel(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>彻底删除 {askDel?.length ?? 0} 份录制？</AlertDialogTitle>
            <AlertDialogDescription>
              会把这些录制从**长期库与缓存**里都删掉，不可恢复（要再造一份只能重录一次）。
              {'\n'}只是不想看它们的话，删之前先想一下：里面可能留着还没回放过的现场。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => { const ids = askDel; setAskDel(null); if (ids) void doDelete(ids) }}>
              删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* 明细页:整屏换过去(列表留在 DOM 里、只 display:none,见上面的 wrapper)。
          它是"录制"视图内部的一级 —— 活动栏那五个图标不变、也不新增视图。 */}
      {recordId && (
        // `onChanged` = "明细页改了列表会关心的事实"（删除、起回放）→ 让列表重拉一次。
        // 不给它的话：删除后回到列表还看得见刚删的那一行（列表 keep-alive，view 没变就不重拉）。
        <RecordingDetail id={recordId} onBack={() => setRecordId('')} onChanged={onRefresh} />
      )}
    </div>
  )
}