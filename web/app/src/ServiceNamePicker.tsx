// 服务编号查询：服务名称字段的开窗（口径来自 T100 q_gzja005_1）。
//
// 三列：服务规格编号 / 服务名称 / API服务名称；多选，回传 API服务名称（gzja004），
// 以 `|` 串联 —— 与 T100 的 g_qryparam.return3 一致。双击某行 = 只选该行并确认。
import { useCallback, useEffect, useMemo, useState } from 'react'
import { QueryTableDialog, type ColumnSpec, type FilterSpec } from './QueryTableDialog'
import { api, type WSServiceItem } from './api'

const PAGE_SIZE = 200
const MAX_SELECT = 200
const EMPTY: Record<string, string> = {}

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

  // 打开时把当前值拆成已选项；第一个值预填「API服务名称」过滤框。
  const initial = useMemo(() => {
    if (!open) return EMPTY
    const cur = value.split('|').map((s) => s.trim()).filter(Boolean)
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
    const cur = value.split('|').map((s) => s.trim()).filter(Boolean)
    setSelected(new Set(cur))
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
    onPick([...selected].join('|'))
    onClose()
  }

  return (
    <QueryTableDialog
      open={open}
      title="服务编号查询"
      filters={FILTERS}
      columns={COLUMNS}
      rows={rows}
      rowKey={(r) => r.apiName || r.spec}
      loading={loading}
      error={error}
      total={total}
      page={page}
      pageSize={PAGE_SIZE}
      selection="multi"
      selected={selected}
      onToggle={toggle}
      onQuery={(values, p) => void load({ spec: values.spec, name: values.name, api: values.api }, p)}
      onOpenRow={(key) => { onPick(key); onClose() }}
      onConfirm={confirm}
      onClose={onClose}
      confirmText="选择"
      emptyHint="没有匹配的服务；换个条件，或先确认该环境有服务注册数据。"
      initialValues={initial}
    />
  )
}
