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
// **点一行 → 整屏换成明细页**(`#record/<id>`,见 RecordingDetail.tsx):条带 + 三层记录表
// + 检查器。身份、回放、复制 ID、删除这些"看着这一份做决定"的动作也搬到了明细页的头上 ——
// 一份录制只该有一个"我正看着它"的地方,列表只管挑。
import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react'
import { Play, RefreshCw, Square, Video } from 'lucide-react'
import { useStore } from './store'
import { api, type RecordingCurrent, type RecordingRow } from './api'
import { Input, Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../shared/ui'
import { useToast } from './toast'
import { RecordingDetail } from './RecordingDetail'

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

// 会话登记里那一轮**该怎么称呼**。
//
// 为什么要分四种:登记文件比进程活得久 —— 「回放」是另起一个进程 `rec reproduce --no-wait`,
// 它发完事件就退,留下的是"登记还在 + 一个还开着的窗口"(用户要看最后那一屏)。把这种残留
// 说成"正在录"是错的(用户 2026-10-15 撞上过:面板写着"正在录:fep-x",可那一轮其实是回放),
// 于是按 用途(录/回放) × 起它的进程还在不在 分四种说法,并且都指向同一个收工按钮。
function curLabel(cur: RecordingCurrent): string {
  const stale = !!cur.stale
  if (cur.kind === 'replay') return stale ? '回放留下的窗口' : '正在回放'
  return stale ? '上一轮录制没收干净' : '正在录'
}

// 收工按钮上的字:回放那一轮谈不上"停止录制"(它早就发完了),留住的是那个窗口。
function stopLabel(cur: RecordingCurrent | null): string {
  return cur?.kind === 'replay' ? '收工(关窗口)' : '停止录制'
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

// 指纹:hash(条数)。hash 空 = 这一轮根本没录到界面(作业没连上来 / 刚起来就断了),
// 那不是"少了几个字",而是"这份录像没有可回放的界面"。
function fmtFP(fp?: { total: number; hash: string }): string {
  if (!fp || !fp.hash) return '-'
  return `${fp.hash}(${fp.total} 条)`
}

export function RecordView() {
  const view = useStore((s) => s.view)
  // 环境与会话都取自**左侧那个会话面板**(用户 2026-10-15 要求):
  // 有了独立的会话栏,这里再摆一个环境下拉就是在两个地方管同一件事。
  // 于是本面板不再自己选环境 —— 列表按"当前会话的环境"过滤,起录也用同一个。
  const sessionEnv = useStore((s) => s.sessionEnv)
  const sessionId = useStore((s) => s.sessionId)
  const toast = useToast()

  const [prog, setProg] = useState('')
  // 「刷新」按钮的强制重拉信号(条件没变时也要重新拉一次)
  const [tick, setTick] = useState(0)

  const [rows, setRows] = useState<RecordingRow[]>([])
  const [cur, setCur] = useState<RecordingCurrent | null>(null)
  const [loading, setLoading] = useState(false)
  // 列表读取失败的原因（非空 = 界面要给"读不到"而不是"没有"）。
  const [loadErr, setLoadErr] = useState('')
  const [busy, setBusy] = useState<'' | 'start' | 'stop'>('')

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

  // 进面板就先拉一次。视图是 keep-alive(只切显隐、不卸载重挂),所以"进面板"这件事
  // 只能由 view 的变化来标识 —— 挂载那一刻 view 就已经是 record 了。
  // 会话换了环境(sessionEnv 变)也要跟着重拉:列表是按它过滤的。
  useEffect(() => {
    if (view !== 'record') return
    void load(sessionEnv)
  }, [view, sessionEnv, tick, load])

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
          if (!next) void load(sessionEnv)
        } catch { /* 轮询失败不动界面:下一拍再试 */ }
      })()
    }, 5000)
    return () => window.clearInterval(t)
  }, [view, cur, load, sessionEnv])

  // 点开一行就取它的身份。取不到(没有 .meta.json 的老件)照实说,不假装有。
  // 列表与明细页共用同一个「在录吗」轮询:明细页自己不再轮询 —— 那是同一件事。

  const onRefresh = () => setTick((t) => t + 1)

  // 起一轮录制。已经在录时服务端回 409 与一句中文原因,原样弹出来 ——
  // 所以这里只把"点不了"的两种情形(没会话环境 / 没作业名)挡在前面。
  //
  // 环境与 ENT 都跟着**当前会话**走(面板上不再单独选):环境用会话的环境名,
  // 企业编号读会话自己的 TOPENT(读取失败就不带 —— 那就等于沿用登录 shell 的 $TOPENT,
  // 与服务端 "空 = 沿用 $TOPENT" 的口径一致)。
  const onStart = async () => {
    // 三道闸门在这里**再走一遍**，而不是只靠按钮的 disabled：
    // 作业名那个输入框的回车也走这条路（实测：按钮已经禁用并写着"已经在录了"时，
    // 在框里按回车照样把请求发出去；同一个任务里连按两次回车会发出 2 个 POST）。
    // 按钮的 disabled 只是**界面**上的闸门，状态检查必须落在动作里。
    if (busy) return
    if (cur && !cur.stale) {
      toast.error('已经在录了', { description: `当前这一轮是 ${cur.id}（${cur.prog || '-'}）：先${stopLabel(cur)}再起新的（两个会话会抢同一组端口与调试会话）。` })
      return
    }
    if (!sessionEnv) {
      toast.error('还没有会话环境', { description: '先在左侧「会话」面板连一个环境 —— 录制落在哪个环境、用哪个企业,都跟着它走。' })
      return
    }
    const p = prog.trim()
    if (!p) {
      toast.error('请填作业名', { description: '要录哪个作业(如 bsft001_wf)。' })
      return
    }
    setBusy('start')
    try {
      let ent = ''
      if (sessionId) {
        try {
          const snap: any = await api.snapshot(sessionId)
          ent = snap?.topentSession || ''
        } catch { /* 读不到就不带 ENT:等同于沿用 $TOPENT */ }
      }
      const r = await api.recordingStart({ env: sessionEnv, prog: p, ent })
      const id = r.recording?.id || ''
      toast.success(`已开始录制:${id}`, {
        description: `作业 ${r.recording?.prog || p}${r.log ? `;日志 ${r.log}` : ''}`,
        duration: 8000, // 录制 ID 多留一会儿:默认 3.5 秒常常还没看清就没了
      })
      await load(sessionEnv)
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
    const replay = cur?.kind === 'replay'
    try {
      const r = await api.recordingStop()
      const id = r.stopped?.id || ''
      toast.success(replay ? `已收工:${id}` : `已停止录制:${id}`, {
        description: replay
          ? '回放留下的那个窗口已经关掉,会话登记也清了'
          : `${r.stopped?.prog || ''} 已收工,录制件落到长期库/缓存(稍等再点刷新)`,
      })
      await load(sessionEnv)
    } catch (e: any) {
      toast.error('停止录制失败', { description: e?.message || String(e) })
    } finally {
      setBusy('')
    }
  }

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col">
      {/* 工具条一行:**只有"录哪个作业"这一件事**。
          环境与企业编号本来在这里各有一个控件,现在都去掉了 —— 左侧独立的会话面板已经在管
          这两样(用户 2026-10-15:"因为我们有环境面板了"),同一个事实摆两处只会对不上。
          于是这一行只留:作业名 + 起录/停录 + 刷新,并且**靠左**(不再 ml-auto 顶到右边:
          左边的条件没了,再把人赶到右边纯属浪费)。
          字段名在左、控件在右,不写进 placeholder(与接口日志页的 QueryField 同一口径)。 */}
      <div className="flex shrink-0 flex-wrap items-center gap-2 border-b border-border px-2 py-2">
        <span className="shrink-0 text-xs text-muted-foreground"
          title="要录的作业编号。起录走的是服务端另起的 `tt fep rec start`,与调试会话互不干扰">作业名</span>
        <Input
          value={prog}
          onChange={(e) => setProg(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') void onStart() }}
          placeholder="如 bsft001_wf"
          title="要录的作业编号。起录走的是服务端另起的 `tt fep rec start`,与调试会话互不干扰;环境与企业编号都跟着左侧会话面板当前那个会话走"
          className="h-7 w-40 text-xs"
        />
        <button
          onClick={() => void onStart()}
          disabled={!!busy || !sessionEnv || !prog.trim() || (!!cur && !cur.stale)}
          title={cur && !cur.stale
            ? `已经在录了(${cur.id}):先停掉它再起新的(两个会话会抢同一组端口)`
            : cur && cur.stale
              ? `登记里还留着上一轮(${cur.id}),但起它的进程已经不在了 —— 可以直接起新的;那个窗口若还开着,点「${stopLabel(cur)}」收掉`
              : !sessionEnv
                ? '先在左侧「会话」面板连一个环境 —— 录制落在哪个环境跟着它走'
                : `起一轮录制(环境 ${sessionEnv}):服务端会另起一个脱离的进程,等它登记(秒级到几十秒)`}
          className="inline-flex shrink-0 items-center gap-1 border border-emerald-500/20 px-2 py-0.5 text-xs text-emerald-600 hover:bg-emerald-500/10 disabled:pointer-events-none disabled:opacity-40 dark:text-emerald-400">
          <Play className="h-3.5 w-3.5" />
          {busy === 'start' ? '正在起…' : '开始录制'}
        </button>
        <button
          onClick={() => void onStop()}
          disabled={!!busy || !cur}
          title={cur
            ? (cur.kind === 'replay'
              ? `收工:关掉回放留下的那个窗口(${cur.id},起它的回放进程早就退场了)`
              : `停止并收工:${cur.id}(它会落身份并搬进长期库)`)
            : '现在没有在录的会话'}
          className="inline-flex shrink-0 items-center gap-1 border border-border px-2 py-0.5 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:pointer-events-none disabled:opacity-40">
          <Square className="h-3.5 w-3.5" />
          {busy === 'stop' ? '正在停…' : stopLabel(cur)}
        </button>
        <button onClick={onRefresh} disabled={loading} title="重新拉列表与「在录吗」"
          className="p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:opacity-40">
          <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
        </button>
        {/* 列表按哪个环境过滤:说清它跟着会话走,不然用户会以为漏了一份 */}
        <span className="shrink-0 text-[11px] text-muted-foreground"
          title="列表按当前会话的环境过滤;还没连会话时列全部环境">
          {sessionEnv ? `环境 ${sessionEnv}` : '未连会话:列全部环境'}
        </span>
        {cur && (
          <span className={`flex min-w-0 items-center gap-1 text-xs ${cur.stale
            ? 'text-muted-foreground'
            : 'text-amber-600 dark:text-amber-400'}`}
            title={cur.kind === 'replay'
              ? (cur.stale
                ? `这是一轮回放留下的窗口(${cur.id}):回放进程发完事件就退场了,登记与那个窗口还在。「${stopLabel(cur)}」会收掉它`
                : `正在回放 ${cur.id}:进程还在(它会等停站)`)
              : (cur.stale
                ? `登记里还留着 ${cur.id},但起它的进程已经不在了(那一轮没收干净)。「${stopLabel(cur)}」可以收掉它留下的窗口`
                : `这一轮还活着(会话登记里有它)。${stopLabel(cur)}会让它收工并落身份:${PATH_LABEL[cur.path || ''] || cur.path || ''}`)}>
            <Video className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">{curLabel(cur)}:{cur.id}({cur.prog || '-'})</span>
          </span>
        )}
      </div>

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
                <TableRow className="hover:bg-transparent">
                  <TableHead className="sticky top-0 z-10 w-48 border-t-0 border-l-0 bg-background"
                    title="录制 ID:命令行 `tt fep rec show/rm/reproduce <ID>` 用的就是它">录制 ID</TableHead>
                  <TableHead className="sticky top-0 z-10 w-16 border-t-0 border-l-0 bg-background"
                    title="长期库:不自动清理,可在明细页里手动删;缓存:可再生的中间数据,随统一清理框架按年龄删">住哪</TableHead>
                  <TableHead className="sticky top-0 z-10 w-24 border-t-0 border-l-0 bg-background"
                    title="这一次作业是怎么起来的。回放闸门按它比对:路径不同的录像不能用">启动路径</TableHead>
                  <TableHead className="sticky top-0 z-10 w-32 border-t-0 border-l-0 bg-background">作业</TableHead>
                  <TableHead className="sticky top-0 z-10 w-36 border-t-0 border-l-0 bg-background"
                    title="环境 / 区域 / ENT(企业编号)">环境/区域/ENT</TableHead>
                  <TableHead className="sticky top-0 z-10 w-36 border-t-0 border-l-0 bg-background">时间</TableHead>
                  <TableHead className="sticky top-0 z-10 w-20 border-t-0 border-l-0 bg-background">体积</TableHead>
                  <TableHead className="sticky top-0 z-10 w-48 border-t-0 border-l-0 bg-background"
                    title="编号表摘要:sha256 前 16 位(条数)。整体比对用它,不看数字也不看声明顺序">指纹</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {rows.map((r) => {
                  return (
                    <TableRow key={`${r.where}:${r.id}`}
                      onClick={() => openDetail(r)}
                      title="点开这一份的明细(条带 + 三层记录表 + 检查器):看它录到了什么、回放前核对身份"
                      className="h-7 cursor-pointer hover:bg-accent/40">
                      <TableCell className="w-48 border-t-0 border-l-0 px-2 py-0 font-mono" title={r.id}>
                        <span className="block truncate font-mono">{r.id}</span>
                      </TableCell>
                      <TableCell className="w-16 border-t-0 border-l-0 px-2 py-0" title={r.where}>
                        <span className="block truncate">{r.where || '-'}</span>
                      </TableCell>
                      <TableCell className="w-24 border-t-0 border-l-0 px-2 py-0"
                        title={PATH_LABEL[r.path || ''] || (r.path || '身份里没写路径')}>
                        <span className="block truncate font-mono text-muted-foreground">{r.path || '-'}</span>
                      </TableCell>
                      <TableCell className="w-32 border-t-0 border-l-0 px-2 py-0" title={r.prog || ''}>
                        <span className="block truncate text-sky-600 dark:text-sky-400">{r.prog || '-'}</span>
                      </TableCell>
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
                  <div>{sessionEnv ? `环境 ${sessionEnv} 下还没有录制件。` : '还没有录制件。'}</div>
                  <div>填好上方的作业名,点「开始录制」录一轮;录完收工后它就会出现在这里。</div>
                  {!sessionEnv && (
                    <div>起录需要一个环境:先在左侧「会话」面板连一个(环境与企业编号都跟着它走)。</div>
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
