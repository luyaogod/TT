// 设置页「技能」分区：卡片墙 + 一条安装栏。
//
// 只放三件事：有哪些技能（卡片）、加/删（卡片上的动作）、装到哪（底部那一条）。
// 刻意不摆说明书：来源路径、assets 指针规则、撞名优先级这些都在
// internal/skill/README.md 与 `tt install skills --help` 里，页面上重复一遍只会把
// 真正要操作的东西挤下去。
//
// 一次安装的结果走 toast（成功/告警），只有「有技能这次没装成」才在页面上留一条
// 可读的清单 —— 那是唯一需要人动手的信息，不能靠 8 秒后消失的提示。
//
// 每个按钮都是立即生效的动作（增删登记 / 安装），不参与设置页那套「改了再保存」的
// 脏标记 —— 所以它不 markDirty，页头的保存按钮也不该因它亮起。
import { useCallback, useEffect, useState } from 'react'
import { AlertCircle, FolderOpen, Plus, RefreshCw, Trash2 } from 'lucide-react'
import { api, type SkillInfo, type SkillRow, type SkillsResp } from './api'
import { Badge, Button, Input } from '../../shared/ui'
import { SectionTitle } from '../../shared/settings'
import { Checkbox } from '../../shared/ui-radix'
import { DirPicker } from './DirPicker'
import { useToast } from './toast'
import { cn } from '../../shared/utils'

/** 安装目标记在本地：换页签再回来不用重填（它不属于 config.json）。 */
const TARGET_KEY = 'tt.skills.target'

// 隐私模式/禁用存储时 localStorage 会直接抛，页面不该因此崩掉。
const readTarget = () => {
  try { return localStorage.getItem(TARGET_KEY) || '' } catch { return '' }
}
const saveTarget = (v: string) => {
  try { localStorage.setItem(TARGET_KEY, v) } catch { /* 记不住就算了，不影响功能 */ }
}

/** 非 ok 的三个状态各给一个徽标；ok 什么都不显示（正常状态不需要提示）。 */
const BADGE: Record<string, { text: string; tone: 'yellow' | 'red' }> = {
  missing: { text: '找不到', tone: 'yellow' },
  invalid: { text: '不合规', tone: 'red' },
  conflict: { text: '撞名', tone: 'red' },
}

export function SkillsPanel() {
  const toast = useToast()
  const [data, setData] = useState<SkillsResp | null>(null)
  const [loadErr, setLoadErr] = useState('')
  const [busy, setBusy] = useState('')

  // 安装栏：目标目录 + 覆盖开关
  const [to, setTo] = useState(readTarget)
  const [force, setForce] = useState(false)
  // 一个选择器，两种意图：'add' 挑要登记的技能目录，'to' 挑安装目标。
  const [pickMode, setPickMode] = useState<'' | 'add' | 'to'>('')
  // 这次安装里没装成的那些（唯一需要在页面上留下来的信息）
  const [warns, setWarns] = useState<string[]>([])

  const load = useCallback(async () => {
    try {
      setLoadErr('')
      setData(await api.skills())
    } catch (e: any) {
      setLoadErr(e?.message || String(e))
    }
  }, [])

  useEffect(() => { void load() }, [load])

  const setTarget = (v: string) => {
    setTo(v)
    saveTarget(v)
  }

  const addPath = async (path: string) => {
    setBusy('add')
    try {
      const r = await api.skillsAdd(path)
      // 失联允许存下来，但要当场说一声 —— 否则用户会以为它已经装上了。
      if (r.warning) toast.info(r.warning)
      else toast.success(r.unchanged ? '这个目录已经登记过了' : `已添加 ${r.entry?.name || path}`)
      await load()
    } catch (e: any) {
      toast.error(e?.message || String(e))
    } finally {
      setBusy('')
    }
  }

  const removePath = async (path: string) => {
    setBusy('remove')
    try {
      await api.skillsRemove(path)
      toast.success('已移除登记（磁盘上的目录没动）')
      await load()
    } catch (e: any) {
      toast.error(e?.message || String(e))
    } finally {
      setBusy('')
    }
  }

  const install = async () => {
    if (!to.trim()) {
      toast.error('先填安装目标目录')
      return
    }
    setBusy('install')
    try {
      const r = await api.skillsInstall(to.trim(), force)
      const texts = r.warningTexts || []
      const n = r.report.files.length
      const skipped = (r.report.skipped ?? []).length
      setWarns(texts)
      if (texts.length > 0) toast.error(`装了 ${n} 个文件，${texts.length} 个技能没装成`)
      else if (n === 0 && skipped > 0) toast.info(`${skipped} 个技能目标里已经是最新的（要刷新就勾「覆盖同名」）`)
      else toast.success(`已装 ${n} 个文件到 ${r.report.target}`)
      await load()
    } catch (e: any) {
      toast.error(e?.message || String(e))
    } finally {
      setBusy('')
    }
  }

  if (loadErr) {
    return (
      <div className="space-y-2 rounded-ui border border-border p-3">
        <div className="flex items-start gap-2 text-[11px] text-red-600 dark:text-red-400">
          <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          <span className="break-all">{loadErr}</span>
        </div>
        <Button size="sm" variant="outline" onClick={() => void load()}>
          <RefreshCw className="h-3.5 w-3.5" />重试
        </Button>
      </div>
    )
  }

  const official = data?.official ?? []
  const externals = data?.externals ?? []

  return (
    <div className="space-y-4">
      {/* 顶部：一眼看清有几张卡，右边是唯一的新增入口 */}
      <div className="flex items-center justify-between gap-2">
        <span className="text-[11px] text-muted-foreground">
          官方 {official.length} · 第三方 {externals.length}
        </span>
        <Button size="sm" variant="outline" disabled={busy !== ''} onClick={() => setPickMode('add')}>
          <Plus className="h-3.5 w-3.5" />添加技能
        </Button>
      </div>

      {data?.officialError && (
        <div className="flex items-start gap-2 rounded-ui border border-red-500/40 p-2 text-[11px] text-red-600 dark:text-red-400">
          <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          <span className="break-all">{data.officialError}</span>
        </div>
      )}

      <section>
        <SectionTitle>官方</SectionTitle>
        <div className="grid gap-2 sm:grid-cols-2">
          {official.map((s) => <SkillCard key={s.name} s={s} />)}
        </div>
      </section>

      <section>
        <SectionTitle>第三方</SectionTitle>
        <div className="grid gap-2 sm:grid-cols-2">
          {externals.map((s) => (
            <SkillCard key={s.path || s.dir} s={s}
              onRemove={busy === '' ? () => void removePath(s.path || s.dir) : undefined} />
          ))}
          <button type="button" disabled={busy !== ''} onClick={() => setPickMode('add')}
            className={cn(
              'flex min-h-[4.75rem] items-center justify-center gap-1.5 rounded-ui border border-dashed border-border text-[11px] text-muted-foreground transition-colors hover:border-foreground/30 hover:text-foreground disabled:opacity-40',
              externals.length === 0 && 'sm:col-span-2',
            )}>
            <Plus className="h-3.5 w-3.5" />添加技能目录
          </button>
        </div>
      </section>

      {/* 只在有技能没装成时出现：那是唯一需要人动手的信息 */}
      {warns.length > 0 && (
        <div className="space-y-1 rounded-ui border border-amber-500/40 p-2">
          <div className="flex items-center gap-1.5 text-[11px] text-amber-600 dark:text-amber-400">
            <AlertCircle className="h-3.5 w-3.5 shrink-0" />这次有 {warns.length} 个技能没装成
          </div>
          {warns.map((w, i) => (
            <div key={i} className="break-all text-[11px] leading-4 text-muted-foreground">{w}</div>
          ))}
        </div>
      )}

      {/* 安装栏：目标 + 覆盖 + 装 */}
      <div className="flex flex-wrap items-center gap-1.5 border-t border-border pt-3">
        <Input className="h-7 min-w-0 flex-1 text-xs" value={to}
          placeholder="安装目标目录，如 D:\项目\.agents\skills"
          onChange={(e) => setTarget(e.target.value)} />
        <Button size="icon" variant="outline" title="选择文件夹" onClick={() => setPickMode('to')}>
          <FolderOpen className="h-3.5 w-3.5" />
        </Button>
        <label className="flex cursor-pointer items-center gap-1.5 px-1 text-[11px] text-muted-foreground">
          <Checkbox checked={force} onCheckedChange={(v) => setForce(v === true)} />覆盖同名
        </label>
        <Button size="sm" disabled={busy !== ''} onClick={() => void install()}>
          <RefreshCw className={cn('h-3.5 w-3.5', busy === 'install' && 'animate-spin')} />
          {busy === 'install' ? '安装中…' : '安装'}
        </Button>
      </div>

      <DirPicker
        open={pickMode !== ''}
        initial={pickMode === 'to' ? to : ''}
        title={pickMode === 'add' ? '选择技能目录（里面要有 SKILL.md）' : '选择安装目标目录'}
        onPick={(p) => {
          const mode = pickMode
          setPickMode('')
          if (mode === 'add') void addPath(p)
          else if (mode === 'to') setTarget(p)
        }}
        onClose={() => setPickMode('')}
      />
    </div>
  )
}

/** 一张技能卡：名字 + 一句话说明（不 ok 时换成原因）+ 第三方才有目录与删除。 */
function SkillCard({ s, onRemove }: { s: SkillInfo | SkillRow; onRemove?: () => void }) {
  const badge = BADGE[s.state]
  // 失联的第三方读不到 name，退回登记时留的 label；都没有就明确显示“未命名”。
  const label = 'label' in s && s.label ? s.label : ''
  const title = s.name || label || '未命名'
  return (
    <div className={cn(
      'group relative flex min-h-[4.75rem] flex-col gap-1 rounded-ui border bg-card p-2.5',
      badge ? (badge.tone === 'red' ? 'border-red-500/40' : 'border-amber-500/40') : 'border-border',
      onRemove && 'pr-8',
    )}>
      <div className="flex min-w-0 items-center gap-1.5">
        <span className="min-w-0 flex-1 truncate font-mono text-xs text-foreground" title={title}>{title}</span>
        {badge && <Badge tone={badge.tone}>{badge.text}</Badge>}
      </div>
      <p className="line-clamp-2 text-[11px] leading-4 text-muted-foreground">
        {(s.state === 'ok' ? s.description : s.why) || '（无说明）'}
      </p>
      {/* 目录只对第三方有意义（官方的路径是固定的，写出来只是噪声） */}
      {onRemove && (
        <p className="mt-auto truncate pt-1 font-mono text-[10px] text-muted-foreground/70" title={s.dir}>{s.dir}</p>
      )}
      {onRemove && (
        <button type="button" title="移除登记（不动磁盘文件）" onClick={onRemove}
          className="absolute right-1.5 top-1.5 rounded-ui p-1 text-muted-foreground opacity-50 transition-opacity hover:bg-accent hover:text-foreground group-hover:opacity-100 focus-visible:opacity-100">
          <Trash2 className="h-3.5 w-3.5" />
        </button>
      )}
    </div>
  )
}
