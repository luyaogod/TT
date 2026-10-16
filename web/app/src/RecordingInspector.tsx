// 检查器(计划 C 的 D6):选中一行/一块之后,这里回答"它到底是什么"。
//
// 四个页签照 DSH 的 Summary/Preview/Raw/Source:
//
//	概览     类型 / 目标 / 时序 / **流量**(这一步之后运行期推了多少字节 —— 被确认的量度)
//	原始帧   那一段字节的 hex 与可打印形态(来自 /raw,越界与超限由服务端拒)
//	引起的回应 这一步之后、下一步之前运行期推回来的帧(归属规则见 internal/web/recordingread.go)
//	名字来源  这个名字是怎么来的(R1–R4 命中哪条、父链、置信度、没解出来时的原因)
//
// 一条硬纪律贯穿四个页签:**解不出来就说解不出来**。名字那一栏在 `unresolved` 时显示
// `Type#id` 加一句为什么,而不是留空让人以为是界面坏了(实测 17 个动作里只有 6 个能解出名字)。
import { useEffect, useState } from 'react'
import { api, type RecordingAction, type RecordingFrame, type RecordingMeta } from './api'
import { fmtBytes, fmtClock, fmtDur, fmtRel } from './recordingTime'

/** /raw 单次最多取这么多(服务端上限 8 KB,这里留一半,免得一次把正文挤爆)。 */
const RAW_CHUNK = 4096

type Tab = 'summary' | 'raw' | 'responses' | 'name'

interface Props {
  id: string
  /** 选中的帧(全文流/回应行/条带上的块都会设它) */
  sel: RecordingFrame | null
  /** 选中的语义动作(分层记录的行会设它;它自带那一帧) */
  action: RecordingAction | null
  precision: 'frame' | 'block'
  /** 取消选中(条带上右键、或点"清空")时回到"整份录制"的概览 */
  counts: { frames: number; messages: number; cache: number; eventFrames: number; events: number; actions: number }
  notes?: string[]
  directions: { conn: number; dir: number; dirLabel: string; bytes: number; frames: number; raw: string; index: { state: string; frames: number; rebuilt: boolean } }[]
  firstT: number
  /** 身份(来自 /api/recordings/{id}):没选中任何东西时,这一栏就是"这份录制是什么" */
  meta: RecordingMeta | null
  /** 身份**问过了**(false = 还在读)。它与"这份没有身份文件"是两件事,不许合并。 */
  metaLoaded: boolean
}

export function RecordingInspector(p: Props) {
  const frame = p.action ? p.action.frame : p.sel
  const [tab, setTab] = useState<Tab>('summary')
  // 换选中目标时回到"概览":停在上一个目标的"原始帧"页签上会显示上一个目标的字节(错得很难发现)。
  useEffect(() => { setTab('summary') }, [frame?.c, frame?.d, frame?.o, p.action?.n])

  const tabs: { key: Tab; label: string }[] = [
    { key: 'summary', label: '概览' },
    { key: 'raw', label: '原始帧' },
    { key: 'responses', label: p.action ? `引起的回应 (${p.action.responsesTotal || 0})` : '引起的回应' },
    { key: 'name', label: '名字来源' },
  ]

  return (
    <div className="flex h-full min-h-0 min-w-0 flex-col bg-card">
      <div className="flex h-7 shrink-0 items-center gap-1 overflow-x-auto whitespace-nowrap border-b border-border px-1 text-[11px]">
        {tabs.map((t) => (
          <button key={t.key} onClick={() => setTab(t.key)}
            className={`rounded-ui px-2 py-0.5 ${tab === t.key ? 'bg-accent text-foreground' : 'text-muted-foreground hover:bg-accent/50'}`}>
            {t.label}
          </button>
        ))}
        <span className="ml-auto pr-1 font-mono text-[10px] text-muted-foreground">
          {frame ? `${frame.c}:${frame.d}:${frame.o}` : '未选中'}
        </span>
      </div>
      <div className="min-h-0 flex-1 overflow-auto p-2 text-[11px]">
        {!frame && tab === 'summary' && <WholeRecording p={p} />}
        {!frame && tab !== 'summary' && (
          <Empty text="这个页签说的是**选中的那一帧**：在下面的表里点一行，或在条带上点一块。" />
        )}
        {frame && tab === 'summary' && <Summary p={p} frame={frame} />}
        {frame && tab === 'raw' && <RawTab id={p.id} frame={frame} />}
        {frame && tab === 'responses' && <Responses p={p} />}
        {frame && tab === 'name' && <NameSource p={p} />}
      </div>
    </div>
  )
}

function Empty({ text }: { text: string }) {
  return <div className="space-y-1 text-muted-foreground">{text.split('**').map((s, i) => (i % 2 ? <b key={i}>{s}</b> : <span key={i}>{s}</span>))}</div>
}

/** 没选中任何东西时的概览:整份录制的数字、各方向的字节账,以及它的**身份**。
 *
 * 身份这一块原来在列表页的侧栏里(点一行看它)。明细页取代侧栏之后,它必须跟过来 ——
 * 「回放前自己核对一遍」正是这一块存在的理由(启动路径/作业/ENT/编号表/编译件)。 */
function WholeRecording({ p }: { p: Props }) {
  const c = p.counts
  const m = p.meta
  const bin = m?.binary
  const binText = bin && (bin.ref || bin.size || bin.modTime)
    ? [bin.ref, bin.size ? `${bin.size} 字节` : '', bin.modTime || ''].filter(Boolean).join(' / ')
    : '（没探到）'
  return (
    <div className="space-y-2">
      <div className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-0.5">
        <K>帧</K><V>{c.frames} 帧（AUI 消息 {c.messages} · 图片缓存 {c.cache}）</V>
        <K>事件</K><V>事件侧 {c.eventFrames} 帧 → {c.events} 个原始事件</V>
        <K>语义动作</K><V>{c.actions} 步（= 用户眼里的"点了几下 / 输了几个值"）</V>
        <K>时间精度</K>
        <V>{p.precision === 'frame' ? '帧级（帧索引给了每帧首块时刻）' : '块级（没有帧索引；顺序按各方向字节序）'}</V>
        {p.metaLoaded && !m && (
          <><K>身份</K><V className="text-amber-600 dark:text-amber-400">这一份没有身份文件（.meta.json）：启动路径、作业、环境/ENT、编译件都无从得知 —— 帧数据照样能看（它是从录制件的字节现场解出来的）</V></>
        )}
        <K>启动路径</K><V><span className="font-mono">{m?.path || '—'}</span>（回放闸门按它比对：路径不同的录像不能用）</V>
        <K>作业</K><V><span className="font-mono">{m?.prog || '—'}</span></V>
        <K>环境/ENT</K><V><span className="font-mono">{m?.env || '—'}/{m?.zone || '—'}/{m?.ent || '（未指定）'}</span></V>
        <K>时间</K><V className="font-mono">{fmtClock(m?.started ? new Date(m.started).getTime() : 0)} → {m?.finished ? fmtClock(new Date(m.finished).getTime()) : '（没记结束）'}</V>
        <K>指纹</K><V className="font-mono">{m?.fingerprint?.hash ? `${m.fingerprint.hash}(${m.fingerprint.total} 条)` : '—'}（编号表摘要；整体比对用它）</V>
        <K>编译件</K><V>{binText}</V>
        {m?.host ? <><K>寄存主机</K><V className="font-mono">{m.host}</V></> : null}
        {m?.fglserver ? <><K>画面接线</K><V className="font-mono">{m.fglserver}</V></> : null}
        {m?.note ? <><K>备注</K><V>{m.note}</V></> : null}
      </div>
      <div>
        <div className="mb-0.5 text-muted-foreground">各方向</div>
        <table className="w-full table-fixed border-separate border-spacing-0 font-mono text-[10px]">
          <tbody>
            {p.directions.map((d) => (
              <tr key={`${d.conn}:${d.dir}`}>
                <td className="truncate border-b border-border/40 py-0.5 pr-1">#{d.conn} {d.dirLabel}</td>
                <td className="border-b border-border/40 py-0.5 pr-1 text-right">{d.frames} 帧</td>
                <td className="border-b border-border/40 py-0.5 pr-1 text-right">{fmtBytes(d.bytes)}</td>
                <td className="truncate border-b border-border/40 py-0.5 text-muted-foreground"
                  title={`索引状态：${d.index.state}${d.index.rebuilt ? '（事后重建：位置准，时刻当时没记）' : ''}`}>
                  {indexText(d.index.state, d.index.rebuilt)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {p.notes && p.notes.length > 0 && (
        <ul className="list-disc space-y-0.5 pl-4 text-muted-foreground">
          {p.notes.map((n, i) => <li key={i}>{n}</li>)}
        </ul>
      )}
      <div className="text-muted-foreground">在下面的表里点一行、或在条带上点一块,这里就换成它的细节。</div>
    </div>
  )
}

function indexText(state: string, rebuilt: boolean): string {
  switch (state) {
    case 'complete': return rebuilt ? '索引完整（重建·无时刻）' : '索引完整'
    case 'incomplete': return '索引不完整（可重建）'
    case 'stale': return '索引格式过时（可重建）'
    default: return '无索引'
  }
}

const K = ({ children, className }: { children: React.ReactNode; className?: string }) =>
  <span className={className || 'text-muted-foreground'}>{children}</span>
const V = ({ children, className }: { children: React.ReactNode; className?: string }) =>
  <span className={className ? `break-words ${className}` : 'break-words'}>{children}</span>

/** 概览页签:类型 / 目标 / 时序 / 流量。 */
function Summary({ p, frame }: { p: Props; frame: RecordingFrame }) {
  const a = p.action
  const t = a?.target
  const show = (v: number) => (p.precision === 'frame' && v > 0 ? `${fmtClock(v)}（${fmtRel(v, p.firstT)}）` : '—')
  return (
    <div className="space-y-2">
      <div className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-0.5">
        <K>类别</K><V><span className="font-mono">{frame.k}</span>{frame.k === 'cache' ? '（图片缓存：不是消息,不解析也不重放）' : ''}</V>
        <K>方向</K><V>{frame.d === 1 ? '运行期 → 前端（界面状态/回应）' : '前端 → 运行期（用户事件）'}</V>
        <K>连接</K><V>#{frame.c}</V>
        <K>序号</K><V><span className="font-mono">{frame.k === 'event' ? `_om ${frame.q}` : frame.q ? frame.q : '—'}</span></V>
        <K>时刻</K><V>{show(frame.t)}</V>
        <K>大小</K><V>{fmtBytes(frame.n)}（raw 偏移 {frame.o}）</V>
        {a && <><K>这一步</K><V>#{a.n} {a.kind === 'input' ? `输入 ${a.value || ''}` : '点击'}</V></>}
        {a && <><K>原始事件</K><V>{a.events.length} 个（其中 {a.dropped} 个是导航/上报，被语义层丢掉）</V></>}
      </div>

      {t && (
        <div>
          <div className="mb-0.5 text-muted-foreground">目标</div>
          <div className="space-y-0.5">
            <div>
              {t.name && t.confidence !== 'unresolved'
                ? <span>{t.name} <span className="font-mono text-muted-foreground">{t.type}#{t.id}</span></span>
                : <span className="font-mono">{t.type || '?'}#{t.id || a?.id}</span>}
              <sup className="ml-1 text-[10px] text-muted-foreground">{confidenceText(t.confidence)}</sup>
            </div>
            {t.window ? <div className="text-muted-foreground">窗口 <span className="font-mono">#{t.window}</span> {t.windowName || ''}</div> : null}
            {t.why ? <div className="text-muted-foreground">{t.why}</div> : null}
          </div>
        </div>
      )}

      {a && (
        <div>
          <div className="mb-0.5 text-muted-foreground">流量（运行期在这一步之后推回来的）</div>
          {a.responsesTotal ? (
            <div>
              {a.responsesTotal} 帧 / {fmtBytes((a.responses || []).reduce((s, r) => s + (r.bytes || 0), 0))}
              {a.responses && a.responses.length < a.responsesTotal ? `（下面只列了前 ${a.responses.length} 帧）` : ''}
            </div>
          ) : (
            <div className="text-muted-foreground">{a.responsesNote || '没有回包'}</div>
          )}
        </div>
      )}

      {!a && <div className="text-muted-foreground">这一帧不是语义动作，所以没有"目标/流量"两栏（看原始帧与它前后的回应）。</div>}
    </div>
  )
}

function confidenceText(c: string): string {
  switch (c) {
    case 'high': return '高置信'
    case 'low': return '低置信（仅排障）'
    default: return '名字未解析'
  }
}

/** 原始帧页签:那一段字节。允许取帧中间的一段(一帧实测可以 269 KB)。 */
function RawTab({ id, frame }: { id: string; frame: RecordingFrame }) {
  const [off, setOff] = useState(frame.o)
  const [len, setLen] = useState(Math.min(frame.n, RAW_CHUNK))
  const [data, setData] = useState<{ hex: string; printable: string; inFrame: boolean; note?: string } | null>(null)
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(false)
  // 换帧时把偏移与长度复位到这一帧的开头(不复位会拿上一帧的偏移去读这一帧,读出来是别人的字节)
  useEffect(() => {
    setOff(frame.o)
    setLen(Math.min(frame.n, RAW_CHUNK))
  }, [frame.o, frame.n])

  useEffect(() => {
    let alive = true
    setLoading(true)
    api.recordingRaw(id, { conn: frame.c, dir: frame.d, off, len })
      .then((r) => { if (alive) { setData(r); setErr('') } })
      .catch((e: any) => { if (alive) { setData(null); setErr(e?.message || String(e)) } })
      .finally(() => { if (alive) setLoading(false) })
    return () => { alive = false }
  }, [id, frame.c, frame.d, off, len])

  return (
    <div className="space-y-1">
      <div className="flex flex-wrap items-center gap-1 text-muted-foreground">
        <span>取</span>
        <input className="w-24 rounded-ui border border-border bg-background px-1 font-mono"
          value={off} onChange={(e) => setOff(Math.max(0, Number(e.target.value) || 0))} />
        <span>起</span>
        <input className="w-20 rounded-ui border border-border bg-background px-1 font-mono"
          value={len} onChange={(e) => setLen(Math.max(1, Math.min(8192, Number(e.target.value) || 1)))} />
        <span>字节（单次 ≤ 8192）</span>
      </div>
      {loading && <div className="text-muted-foreground">读取中…</div>}
      {err && <div className="text-red-600 dark:text-red-400">{err}</div>}
      {data && !err && (
        <>
          <div className="text-muted-foreground">
            {data.inFrame ? '这一段落在某一帧里' : (data.note || '这一段不在任何一帧里')}
          </div>
          <pre className="overflow-auto whitespace-pre rounded-ui border border-border bg-background p-1 font-mono text-[10px] leading-[1.35]">{data.hex}</pre>
          <div className="text-muted-foreground">可打印形态（控制字符转义；中文照留）</div>
          <pre className="max-h-40 overflow-auto whitespace-pre-wrap break-all rounded-ui border border-border bg-background p-1 font-mono text-[10px]">{data.printable}</pre>
        </>
      )}
    </div>
  )
}

/** 引起的回应页签:这一步之后、下一步之前运行期推回来的帧。 */
function Responses({ p }: { p: Props }) {
  const a = p.action
  if (!a) {
    return <Empty text="这一帧不是语义动作，所以没有「归属到它」的回应。回应只对语义动作归属（见服务端的时间窗规则）。" />
  }
  if (!a.responses || a.responses.length === 0) {
    return (
      <div className="space-y-1">
        <div className="text-muted-foreground">{a.responsesNote || '这一步之后运行期没有回包。'}</div>
        <div className="text-muted-foreground">「没有回包」不等于「0 毫秒」：它可能是本来就没应答，也可能是运行期没处理 —— 要看回放日志与作业状态。</div>
      </div>
    )
  }
  return (
    <div className="space-y-1">
      {a.responses.map((r) => (
        <div key={`${r.c}:${r.o}`} className="rounded-ui border border-border px-1 py-0.5">
          <div className="flex items-center gap-2 font-mono text-[10px] text-muted-foreground">
            <span>{fmtClock(r.t)}</span>
            <span>+{fmtDur(a.frame.t > 0 && r.t > 0 ? r.t - a.frame.t : 0)}</span>
            <span>{fmtBytes(r.n)}</span>
            <span className="ml-auto">{r.k} #{r.q}</span>
          </div>
          <div className="break-words font-mono">{r.preview || '（二进制）'}</div>
        </div>
      ))}
      {a.responsesTotal && a.responses.length < a.responsesTotal && (
        <div className="text-muted-foreground">共 {a.responsesTotal} 帧，上面只列了前 {a.responses.length} 帧（全文流里能看到全部）。</div>
      )}
    </div>
  )
}

/** 名字来源页签:这个名字是怎么来的。 */
function NameSource({ p }: { p: Props }) {
  const a = p.action
  const t = a?.target
  if (!a || !t) return <Empty text="这一帧不是语义动作，没有「名字」这回事。" />
  const rule = ruleText(t.source || '')
  return (
    <div className="space-y-2">
      <div className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-0.5">
        <K>结果</K>
        <V>{t.name && t.confidence !== 'unresolved'
          ? <span>{t.name}（{confidenceText(t.confidence)}）</span>
          : <span className="font-mono">{t.type}#{t.id}</span>}</V>
        <K>规则</K><V>{rule}</V>
        <K>来源</K><V><span className="font-mono">{t.source || '—'}</span></V>
        <K>置信度</K><V>{confidenceText(t.confidence)}{t.cycle ? '（父链有环：消息残缺/跨窗口合并的痕迹）' : ''}</V>
      </div>
      <div>
        <div className="mb-0.5 text-muted-foreground">父链（从父到根）</div>
        {t.parents && t.parents.length > 0 ? (
          <ul className="space-y-0.5 font-mono text-[10px]">
            {t.parents.map((x, i) => (
              <li key={`${x.id}:${i}`} className="truncate">
                {'→'.repeat(0)}{'　'.repeat(i)}#{x.id} {x.type}{x.text ? ` text="${x.text}"` : ''}
              </li>
            ))}
          </ul>
        ) : <div className="text-muted-foreground">没有父链（这个对象是根，或它的父对象没被录到）。</div>}
      </div>
      {t.why ? <div className="text-muted-foreground">{t.why}</div> : (
        <div className="text-muted-foreground">这个名字能引用：来源明确、父链无环、且落在某个窗口里。</div>
      )}
    </div>
  )
}

function ruleText(source: string): string {
  if (source.startsWith('inherit#')) {
    const via = source.slice('inherit#'.length)
    return `R4 继承：沿父链在 ${via ? `#${via}` : '父对象'} 上找到了 text/comment`
  }
  switch (source) {
    case 'comment-key': return 'R1 字典标签：这个对象带 comment，而同一个 comment 的 Label 有文字'
    case 'name-suffix': return 'R2 列名后缀：col/colName 对应到某个 name 以它结尾、且带文字的兄弟对象'
    case 'own-text': return 'R3 自带文字：这个对象自己的 text 就是它的名字'
    default: return '没有命中任何规则（R1–R4 都试过）'
  }
}
