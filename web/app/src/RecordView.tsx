// 「录制」面板:列出某个环境/ENT 下的录制件(长期库 + 缓存),并起一轮/停一轮。
//
// 为什么起录/停录只是两个按钮、列表要手动刷新:一轮录制 = 一个隔离 GDC + 一条代理 +
// 一个常驻 Runtime,服务端会**另起一个脱离的进程**(`tt fep rec start`,见
// internal/web/recordings.go 的文件头注释)。界面这边能做的只有"发请求 + 拉列表",
// 真正的活在子进程里 —— 所以起录要等它登记(秒级到几十秒),停录要等它收工落身份。
//
// 列表把「住哪 / 怎么起的 / 什么时候录的 / 多大 / 指纹 / 归属」一次摊开:
// 挑一份录像回放时,前四项决定"这份能不能用"(闸门按**启动路径**比对,见
// internal/fep/path.go),后三项决定"要不要留着"。
import { useCallback, useEffect, useState } from 'react'
import { Check, Copy, Play, RefreshCw, Square, Trash2, Video } from 'lucide-react'
import { useStore } from './store'
import { api, type RecordingCurrent, type RecordingRow } from './api'
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
      setRows(list.recordings || [])
      setCur(now.recording || null)
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

  // 删长期库里的那一份。确认这一步由 AlertDialog 承担(与「会话」面板同一个基元),
  // 确认之后才带 ?confirm=1 发出去 —— 录制不可再生,不给误点留路。
  const doDelete = async () => {
    const row = pendingDel
    setPendingDel(null)
    if (!row) return
    try {
      const r = await api.recordingDelete(row.id)
      toast.success(`已删除录制:${r.id}`, { description: `清除 ${r.files} 个文件,释放 ${fmtBytes(r.bytes)}` })
      await load(q)
    } catch (e: any) {
      toast.error('删除录制失败', { description: e?.message || String(e) })
    }
  }

  const entDirty = ent.trim() !== q.ent

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
      <div className="min-h-0 flex-1 overflow-auto">
        {/* table-fixed:列宽交给表头/单元格上的 w-* ,留空的那列吃剩余宽度;
            container=false 不套 shadcn Table 默认的 overflow-x-auto(它会让 th 的 sticky 失效);
            border-separate + 单元格只留"右+下"框线,理由与接口日志页列表相同。 */}
        <Table container={false} className="table-fixed border-separate border-spacing-0 text-xs">
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead className="sticky top-0 z-10 w-52 border-t-0 border-l-0 bg-background"
                title="录制 ID:命令行 `tt fep rec show/rm <ID>` 用的就是它">录制 ID</TableHead>
              <TableHead className="sticky top-0 z-10 w-16 border-t-0 border-l-0 bg-background"
                title="长期库:不自动清理,可在这里手动删;缓存:可再生的中间数据,随统一清理框架按年龄删(不在这里删)">住哪</TableHead>
              <TableHead className="sticky top-0 z-10 w-24 border-t-0 border-l-0 bg-background"
                title="这一次作业是怎么起来的。回放闸门按它比对:路径不同的录像不能用">启动路径</TableHead>
              <TableHead className="sticky top-0 z-10 w-32 border-t-0 border-l-0 bg-background">作业</TableHead>
              <TableHead className="sticky top-0 z-10 w-44 border-t-0 border-l-0 bg-background"
                title="环境 / 区域 / ENT(企业编号)">环境/区域/ENT</TableHead>
              <TableHead className="sticky top-0 z-10 w-40 border-t-0 border-l-0 bg-background">时间</TableHead>
              <TableHead className="sticky top-0 z-10 w-20 border-t-0 border-l-0 bg-background">体积</TableHead>
              <TableHead className="sticky top-0 z-10 w-44 border-t-0 border-l-0 bg-background"
                title="编号表摘要:sha256 前 16 位(条数)。整体比对用它,不看数字也不看声明顺序">指纹</TableHead>
              <TableHead className="sticky top-0 z-10 w-24 border-t-0 border-l-0 bg-background"
                title="这份录制是谁录的(用户主动 / 系统自动)。留住哪由归属推默认:用户主动的进长期库,系统自动的进缓存">归属</TableHead>
              <TableHead className="sticky top-0 z-10 border-t-0 border-l-0 bg-background">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((r) => {
              const cp = copied?.id === r.id ? copied.state : null
              return (
                <TableRow key={`${r.where}:${r.id}`} className="h-7">
                  <TableCell className="border-t-0 border-l-0 px-2 py-0 font-mono" title={r.id}>
                    <span className="block truncate font-mono">{r.id}</span>
                  </TableCell>
                  <TableCell className="border-t-0 border-l-0 px-2 py-0" title={r.where}>
                    <span className="block truncate">{r.where || '-'}</span>
                  </TableCell>
                  <TableCell className="border-t-0 border-l-0 px-2 py-0"
                    title={PATH_LABEL[r.path || ''] || (r.path || '身份里没写路径')}>
                    <span className="block truncate font-mono text-muted-foreground">{r.path || '-'}</span>
                  </TableCell>
                  <TableCell className="border-t-0 border-l-0 px-2 py-0" title={r.prog || ''}>
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
                  <TableCell className="border-t-0 border-l-0 px-2 py-0"
                    title={`${ORIGIN_LABEL[r.origin || ''] || '身份里没说谁录的'} · ${RETENTION_LABEL[r.retention || ''] || '身份里没说留哪'}${r.note ? ` · ${r.note}` : ''}`}>
                    <span className="block truncate text-muted-foreground">{ORIGIN_LABEL[r.origin || ''] || '-'}</span>
                  </TableCell>
                  <TableCell className="border-t-0 border-l-0 px-1 py-0">
                    <span className="flex items-center gap-0.5">
                      <button
                        onClick={() => void copyRow(r.id)}
                        title="复制这份录制的 ID。把它发给 AI / 贴到命令行(tt fep rec show/rm <ID>)都用它"
                        className={`inline-flex shrink-0 items-center gap-1 border px-1.5 py-0 text-[11px] transition-colors ${
                          cp === 'fail'
                            ? 'border-red-500/20 text-red-600 dark:text-red-400'
                            : 'border-border text-muted-foreground hover:bg-accent hover:text-foreground'
                        }`}>
                        {cp === 'ok' ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                        {cp === 'ok' ? '已复制' : cp === 'fail' ? '复制失败' : '复制 ID'}
                      </button>
                      {/* 删除只对长期库里的那些开放:缓存里的那份归统一清理框架管(服务端会拒),
                          放在这里只会得到一个 409。 */}
                      {r.where === '长期库' && (
                        <button
                          onClick={() => setPendingDel(r)}
                          title="从长期库里删掉这一份。录制不可再生,会先弹二次确认"
                          className="inline-flex shrink-0 items-center gap-1 border border-red-500/20 px-1.5 py-0 text-[11px] text-red-600 transition-colors hover:bg-red-500/10 dark:text-red-400">
                          <Trash2 className="h-3 w-3" />
                          删除
                        </button>
                      )}
                    </span>
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
