// 服务编号查询：服务名称字段的开窗（口径来自 T100 q_gzja005_1）。
//
// 三列：服务规格编号 / 服务名称 / API服务名称；多选，回传 API服务名称（gzja004），
// 以 `|` 串联 —— 与 T100 的 g_qryparam.return3 一致。双击某行 = 只选该行并确认。
//
// 勾选按**行**（服务规格编号）算，不按 API 名称：wssp00002 与 wssp00002_wf 的 API 名称
// 相同，但它们是两行；点哪行只勾哪行。回传时再把重复 API 名去重（字段最终只认服务名）。
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { QueryTableDialog, type ColumnSpec, type FilterSpec } from './QueryTableDialog'
import { api, type WSServiceItem } from './api'

const PAGE_SIZE = 200
const MAX_SELECT = 200
const EMPTY: Record<string, string> = {}
// 行 key 的分隔符：API 名里不会出现 U+0001，用它把「API名 + 规格编号」拼成唯一的行标识。
const ROW_SEP = '\u0001'
const rowKeyOf = (r: WSServiceItem) => `${r.apiName || r.spec}${ROW_SEP}${r.spec}`
const syntheticKeyOf = (api: string) => `${api}${ROW_SEP}`
const apiOfKey = (key: string) => key.split(ROW_SEP)[0]

const FILTERS: FilterSpec[] = [
  { key: 'spec', label: '服务规格编号', placeholder: '如 wssp*', width: 'w-32' },
  { key: 'name', label: '服务名称', placeholder: '如 查询', width: 'w-44' },
  { key: 'api', label: 'API服务名称', placeholder: '如 easy.query', width: 'w-52' },
]

const COLUMNS: ColumnSpec<WSServiceItem>[] = [
  {
    key: 'spec', head: '服务规格编号', width: 'w-32',
    render: (r) => <span className="block truncate font-mono text-[11px] text-muted-foreground" title={r.spec}>{r.spec}</span>,
  },
  {
    key: 'name', head: '服务名称', width: 'w-64',
    render: (r) => <span className="block truncate" title={r.name}>{r.name || '-'}</span>,
  },
  {
    key: 'apiName', head: 'API服务名称',
    render: (r) => <span className="block truncate font-mono text-[11px]" title={r.apiName}>{r.apiName || '-'}</span>,
  },
]

export function ServiceNamePicker({ open, env, value, onPick, onClose }: {
  open: boolean
  env: string
  value: string
  onPick: (v: string) => void
  onClose: () => void
}) {
  const [rows, setRows] = useState<WSServiceItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [selected, setSelected] = useState<Set<string>>(new Set())
  // 当前值里还没在当前页找到对应行的 API 名：用合成 key 先占位，确认时原样带回。
  const carriedRef = useRef<Set<string>>(new Set())

  // 打开时把当前值拆成已选项；第一个值预填「API服务名称」过滤框。
  const initial = useMemo(() => {
    if (!open) return EMPTY
    const cur = [...new Set(value.split('|').map((s) => s.trim()).filter(Boolean))]
    return { api: cur[0] || '' }
  }, [open, value])

  const load = useCallback(async (q: { spec?: string; name?: string; api?: string }, p: number) => {
    setLoading(true)
    setError('')
    try {
      const r = await api.wsServices({ env, ...q, page: p, pageSize: PAGE_SIZE })
      setRows(r.items)
      setTotal(r.total)
      setPage(r.page)
      // 当前值里每个 API 只映射到**第一行**（同名服务有 _wf 两行时，不要两行都勾上）。
      const mapped = new Map<string, string>()
      for (const it of r.items) {
        const api = it.apiName || it.spec
        if (carriedRef.current.has(api)) {
          mapped.set(api, rowKeyOf(it))
          carriedRef.current.delete(api)
        }
      }
      if (mapped.size) {
        setSelected((prev) => {
          const next = new Set(prev)
          for (const [api, key] of mapped) {
            next.delete(syntheticKeyOf(api))
            next.add(key)
          }
          return next
        })
      }
    } catch (e: any) {
      setError(e?.message || String(e))
      setRows([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [env])

  useEffect(() => {
    if (!open) return
    const cur = [...new Set(value.split('|').map((s) => s.trim()).filter(Boolean))]
    carriedRef.current = new Set(cur)
    setSelected(new Set(cur.map(syntheticKeyOf)))
    setPage(1)
    void load(initial, 1)
    // value 只在打开那一刻取一次；打开期间用户选择会直接关闭弹窗。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, env, initial])

  const toggle = (key: string) => setSelected((prev) => {
    const next = new Set(prev)
    if (next.has(key)) next.delete(key)
    else if (next.size < MAX_SELECT) next.add(key)
    return next
  })

  const confirm = () => {
    // 勾选是行级的；字段只要 API 名，所以按 API 去重后再拼接。
    const apis = new Set<string>()
    for (const key of selected) apis.add(apiOfKey(key))
    onPick([...apis].join('|'))
    onClose()
  }

  return (
    <QueryTableDialog
      open={open}
      title="服务编号查询"
      filters={FILTERS}
      columns={COLUMNS}
      rows={rows}
      rowKey={rowKeyOf}
      loading={loading}
      error={error}
      total={total}
      page={page}
      pageSize={PAGE_SIZE}
      selection="multi"
      selected={selected}
      onToggle={toggle}
      onQuery={(values, p) => void load({ spec: values.spec, name: values.name, api: values.api }, p)}
      onOpenRow={(key) => { onPick(apiOfKey(key)); onClose() }}
      onConfirm={confirm}
      onClose={onClose}
      confirmText="选择"
      emptyHint="没有匹配的服务；换个条件，或先确认该环境有服务注册数据。"
      initialValues={initial}
    />
  )
}
