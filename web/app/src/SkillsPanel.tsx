// 设置页「技能」分区。
//
// 两套来源在这里第一次同时可见：
//   官方   随 tt 发布包自带（<exe 目录>\skills），走专用路径，只读；
//   第三方 用户在下面登记的任意路径，一行一个技能目录。
//
// 这一页的每个按钮都是**立即生效**的动作（增删登记 / 安装），不参与设置页那套
// 「改了再保存」的脏标记 —— 所以它不 markDirty，页头的保存按钮也不该因它亮起。
//
// 为什么安装按钮必须是人点：它会往磁盘写一整棵技能树。和「检查更新」同一条纪律 ——
// 服务启动不装、没有定时器，页面上这一次点击就是那次显式命令。
import { useCallback, useEffect, useState } from 'react'
import { AlertCircle, FolderOpen, Plus, RefreshCw, Trash2 } from 'lucide-react'
import { api, type SkillReport, type SkillRow, type SkillsResp } from './api'
import { Button, Input, Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../shared/ui'
import { Card, InfoRow, SettingRow } from '../../shared/settings'
import { Checkbox } from '../../shared/ui-radix'
import { DirPicker } from './DirPicker'
import { useToast } from './toast'
import { cn } from '../../shared/utils'

const input = 'h-7 text-xs'

/** 状态 → 中文与颜色。四个状态里只有 ok 是"这次会装上"。 */
const STATE_UI: Record<string, { text: string; cls: string }> = {
  ok: { text: '正常', cls: 'text-emerald-600 dark:text-emerald-400' },
  missing: { text: '找不到', cls: 'text-amber-600 dark:text-amber-400' },
  invalid: { text: '不合规', cls: 'text-red-600 dark:text-red-400' },
  conflict: { text: '撞名', cls: 'text-red-600 dark:text-red-400' },
}

export function SkillsPanel() {
  const toast = useToast()
  const [data, setData] = useState<SkillsResp | null>(null)
  const [loadErr, setLoadErr] = useState('')
  const [busy, setBusy] = useState('')

  // 安装：目标目录（agent 读技能的位置）+ 覆盖开关 + 上一次的结果
  const [to, setTo] = useState('')
  const [force, setForce] = useState(false)
  const [report, setReport] = useState<SkillReport | null>(null)
  const [warns, setWarns] = useState<string[]>([])
  // 一个选择器，两种意图：'add' 挑要登记的技能目录，'to' 挑安装目标。
  // 不靠“哪个输入框是空的”去猜 —— 猜错就会把工程目录当成技能登记进去。
  const [pickMode, setPickMode] = useState<'' | 'add' | 'to'>('')

  const load = useCallback(async () => {
    try {
      setLoadErr('')
      setData(await api.skills())
    } catch (e: any) {
      setLoadErr(e?.message || String(e))
    }
  }, [])

  useEffect(() => { void load() }, [load])

  const addPath = async (path: string) => {
    setBusy('add')
    try {
      const r = await api.skillsAdd(path)
      // 失联允许存下来，但要当场说一声 —— 否则用户会以为它已经装上了。
      if (r.warning) toast.info(r.warning)
      else toast.success(r.unchanged ? '这个目录已经登记过了' : `已登记 ${r.entry?.name || path}`)
      await load()
    } catch (e: any) {
      toast.error(e?.message || String(e))
    } finally {
      setBusy('')
    }
  }

  const removePath = async (path: string) => {
    setBusy('remove:' + path)
    try {
      await api.skillsRemove(path)
      toast.success('已从登记表移除（磁盘上的文件没动）')
      await load()
    } catch (e: any) {
      toast.error(e?.message || String(e))
    } finally {
      setBusy('')
    }
  }

  const install = async (dryRun: boolean) => {
    if (!to.trim()) {
      toast.error('先给出目标目录 —— agent 会去读技能的那个位置')
      return
    }
    setBusy(dryRun ? 'dry' : 'install')
    try {
      const r = await api.skillsInstall(to.trim(), force, dryRun)
      setReport(r.report)
      setWarns(r.warningTexts || [])
      if (!dryRun) await load()
      if ((r.warningTexts || []).length > 0) {
        toast.error(`装完了，但有 ${r.warningTexts.length} 条告警（见下方列表）`)
      } else {
        toast.success(dryRun ? '试算完成（没写任何文件）' : `已装到 ${r.report.target}`)
      }
    } catch (e: any) {
      toast.error(e?.message || String(e))
    } finally {
      setBusy('')
    }
  }

  if (loadErr) {
    return (
      <Card title="技能" description="读不到技能登记。">
        <div className="flex items-start gap-2 text-[11px] text-red-600 dark:text-red-400">
          <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          <span className="break-all">{loadErr}</span>
        </div>
        <div className="pt-2">
          <Button size="sm" variant="outline" onClick={() => void load()}>
            <RefreshCw className="mr-1 h-3.5 w-3.5" />重试
          </Button>
        </div>
      </Card>
    )
  }

  return (
    <>
      <Card id="card-skills-official" title="官方技能"
        description={`随 tt 发布包自带（${data?.officialDir || '<exe 目录>\\skills'}）。走专用路径，不受本机配置影响，这里只读。`}>
        {data?.officialError && (
          <div className="flex items-start gap-2 pb-2 text-[11px] text-red-600 dark:text-red-400">
            <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
            <span className="break-all">{data.officialError}</span>
          </div>
        )}
        {(data?.official ?? []).map((s) => (
          <div key={s.name} className="border-t border-border py-1.5 first:border-t-0">
            <div className="font-mono text-[11px] text-foreground">{s.name}</div>
            <div className="mt-0.5 text-[11px] leading-4 text-muted-foreground">{s.description || '（无 description）'}</div>
          </div>
        ))}
        {(data?.official ?? []).length === 0 && !data?.officialError && (
          <p className="text-[11px] text-muted-foreground">没有官方技能目录。</p>
        )}
      </Card>

      <Card id="card-skills-external" title="第三方技能"
        description="登记硬盘上的技能目录（一行一个，目录里必须有 SKILL.md）。安装时整目录复制过去，但根级 assets/ 不复制 —— 目标里的 SKILL.md 首行会写一行指回这里的绝对路径，避免知识库把目标目录撑大。"
        right={
          <Button size="sm" variant="outline" disabled={busy !== ''} onClick={() => setPickMode('add')}>
            <Plus className="mr-1 h-3.5 w-3.5" />添加目录
          </Button>
        }>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>目录</TableHead>
              <TableHead className="w-28">技能名</TableHead>
              <TableHead className="w-24">状态</TableHead>
              <TableHead className="w-10" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {(data?.externals ?? []).map((s: SkillRow) => {
              const ui = STATE_UI[s.state] || { text: s.state, cls: '' }
              const key = s.path || s.dir
              return (
                <TableRow key={key}>
                  <TableCell className="max-w-0">
                    <div className="truncate font-mono text-[11px]" title={s.dir}>{s.path || s.dir}</div>
                    {s.why && <div className="mt-0.5 break-all text-[11px] text-muted-foreground">{s.why}</div>}
                  </TableCell>
                  <TableCell className="font-mono text-[11px]">{s.name || s.label || '—'}</TableCell>
                  <TableCell className={cn('text-[11px]', ui.cls)}>{ui.text}</TableCell>
                  <TableCell>
                    <Button size="sm" variant="ghost" disabled={busy !== ''}
                      title="只从登记表移除，不动磁盘上的文件"
                      onClick={() => void removePath(s.path || s.dir)}>
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </TableCell>
                </TableRow>
              )
            })}
            {(data?.externals ?? []).length === 0 && (
              <TableRow>
                <TableCell colSpan={4} className="text-[11px] text-muted-foreground">
                  还没有登记任何第三方技能。点右上角「添加目录」挑一个带 SKILL.md 的目录。
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
        <p className="pt-2 text-[11px] leading-4 text-muted-foreground">
          「找不到」「不合规」「撞名」都不影响其它技能安装：跑安装时它们只告警并跳过。
          撞名时官方优先，两条第三方同名时先登记的优先。
        </p>
      </Card>

      <Card id="card-skills-install" title="安装到 agent 目录"
        description="把官方技能与上面登记的一起复制过去。目标目录通常是 <工程目录>\.agents\skills（pi、Codex、Cursor、Gemini CLI 等都读它）。">
        <SettingRow id="skills.install.to" label="目标目录"
          description="已经是 agent 技能目录就追加：同名条目默认跳过，不覆盖。"
          control={
            <div className="flex items-center gap-1.5">
              <Input className={input} value={to} placeholder="D:\项目\.agents\skills"
                onChange={(e) => setTo(e.target.value)} />
              <Button size="sm" variant="outline" onClick={() => setPickMode('to')} title="选择文件夹">
                <FolderOpen className="h-3.5 w-3.5" />
              </Button>
            </div>
          } />
        <SettingRow id="skills.install.force" label="覆盖同名条目"
          description="刷新目标里已有的同名技能（升级后想跟当前版本对齐时打开）。"
          control={
            <label className="flex cursor-pointer items-center gap-2 text-[11px] text-muted-foreground">
              <Checkbox checked={force} onCheckedChange={(v) => setForce(v === true)} />
              覆盖
            </label>
          } />
        <div className="flex flex-wrap items-center gap-1.5 pt-2">
          <Button size="sm" disabled={busy !== ''} onClick={() => void install(false)}>
            <RefreshCw className={cn('mr-1 h-3.5 w-3.5', busy === 'install' && 'animate-spin')} />
            {busy === 'install' ? '安装中…' : '安装'}
          </Button>
          <Button size="sm" variant="outline" disabled={busy !== ''} onClick={() => void install(true)}>
            {busy === 'dry' ? '试算中…' : '只算不装'}
          </Button>
          <span className="text-[11px] text-muted-foreground">只算不装会列出会复制哪些文件，一个字节都不写。</span>
        </div>

        {report && (
          <div className="mt-3 space-y-1 border-t border-border pt-2">
            <InfoRow label={report.dryRun ? '试算目标' : '已装到'} value={report.target} />
            <InfoRow label="文件" value={`${report.files.length} 个`} />
            {(report.skipped ?? []).length > 0 && (
              <InfoRow label="跳过" value={(report.skipped ?? []).join(', ')} />
            )}
            {(report.rewritten ?? []).length > 0 && (
              <InfoRow label="改写指针" value={(report.rewritten ?? []).join(', ')} />
            )}
            {warns.length > 0 && (
              <div className="space-y-1 pt-1">
                <div className="flex items-center gap-1 text-[11px] text-amber-600 dark:text-amber-400">
                  <AlertCircle className="h-3.5 w-3.5" />{warns.length} 条告警（这些没装，其余照装）
                </div>
                {warns.map((w, i) => (
                  <div key={i} className="break-all text-[11px] leading-4 text-muted-foreground">{w}</div>
                ))}
              </div>
            )}
            {report.files.length > 0 && (
              <details className="pt-1">
                <summary className="cursor-pointer text-[11px] text-muted-foreground">
                  看文件清单（{report.files.length}）
                </summary>
                <pre className="mt-1 max-h-56 overflow-auto border border-border bg-muted/30 p-2 text-[11px]">
                  {report.files.join('\n')}
                </pre>
              </details>
            )}
          </div>
        )}
      </Card>

      <Card id="card-skills-last" title="上次装到过哪些地方">
        {(data?.targets ?? []).length > 0 ? (
          <>
            <InfoRow label="装的版本" value={data?.stampVersion || '(未知)'} />
            {(data?.targets ?? []).map((t) => <InfoRow key={t} label="目录" value={t} />)}
          </>
        ) : (
          <p className="text-[11px] text-muted-foreground">
            还没装过（或这份构建没有版本号，不记这笔账）。升级 tt 之后会把这几处刷成同版。
          </p>
        )}
        {data?.drift && (
          <div className="flex items-start gap-2 pt-2 text-[11px] text-amber-600 dark:text-amber-400">
            <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
            <span className="break-all">{data.drift}</span>
          </div>
        )}
      </Card>

      <DirPicker
        open={pickMode !== ''}
        initial={pickMode === 'to' ? to : ''}
        title={pickMode === 'add' ? '选择技能目录（里面要有 SKILL.md）' : '选择安装目标目录'}
        onPick={(p) => {
          const mode = pickMode
          setPickMode('')
          if (mode === 'add') void addPath(p)
          else if (mode === 'to') setTo(p)
        }}
        onClose={() => setPickMode('')}
      />
    </>
  )
}
