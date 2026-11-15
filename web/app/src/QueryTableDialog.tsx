// 顶部带过滤字段区域的 Table 控件 + 装它的查询弹窗。
//
// 为什么要拆两层：
//   - FilterTable 是"过滤区 + 表格 + 分页 + 三态"的纯展示件，可以嵌在页面里；
//   - QueryTableDialog 只是给 FilterTable 套一个 Dialog 壳与底部确认/取消。
// 组件不取数：过滤值、分页、行都由调用方通过 props 给，onQuery 只把"要查什么"带出去。
import { useMemo, useState, type ReactNode } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { Button, Input, Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../shared/ui'
import { Checkbox, Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '../../shared/ui-radix'
import { cn } from '../../shared/utils'

export type FilterSpec = {
  key: string
  label: string
  placeholder?: string
  width?: string
}

export type ColumnSpec<T> = {
  key: string
  head: string
  width?: string
  render?: (row: T) => ReactNode
}

export type FilterTableProps<T> = {
  filters: FilterSpec[]
  columns: ColumnSpec<T>[]
  rows: T[]
  rowKey: (row: T) => string
  loading?: boolean
  error?: string
  total: number
  page: number
  pageSize: number
  selection: 'single' | 'multi'
  selected: Set<string>
  /** 单击行：多选=切换勾选，单选=只选这一行。 */
  onToggle: (key: string) => void
  /** 过滤区提交；page 从 1 起。 */
  onQuery: (values: Record<string, string>, page: number) => void
  /** 双击行：给了就直接按这一行执行（开窗常用）；没给就等同"确定"。 */
  onOpenRow?: (key: string) => void
  onConfirm?: () => void
  confirmText?: string
  emptyHint?: string
  initialValues?: Record<string, string>
}

/** 过滤区 + 表格 + 分页 + 三态（不含 Dialog）。 */
export function FilterTable<T>({
  filters, columns, rows, rowKey, loading, error, total, page, pageSize,
  selection, selected, onToggle, onQuery, onOpenRow, onConfirm,
  emptyHint = '没有匹配的数据', initialValues,
}: FilterTableProps<T>) {
  const [values, setValues] = useState<Record<string, string>>(() => ({ ...(initialValues || {}) }))
  const totalPages = Math.max(1, Math.ceil(total / Math.max(1, pageSize)))
  const hasMore = page < totalPages

  const submit = (p = 1) => onQuery(values, p)

  const body = useMemo(() => rows.map((row) => {
    const key = rowKey(row)
    const on = selected.has(key)
    return (
      <TableRow
        key={key}
        data-state={on ? 'selected' : undefined}
        onClick={() => onToggle(key)}
        onDoubleClick={() => (onOpenRow ? onOpenRow(key) : onConfirm?.())}
        className="h-8 cursor-pointer hover:bg-accent/40 data-[state=selected]:bg-sky-500/10"
      >
        {selection === 'multi' && (
          <TableCell className="w-8 border-t-0 border-l-0 px-2 py-0">
            <Checkbox checked={on} onCheckedChange={() => onToggle(key)}
              onClick={(e) => e.stopPropagation()} aria-label="选择这一行" />
          </TableCell>
        )}
        {columns.map((c) => (
          <TableCell key={c.key} className={cn('border-t-0 border-l-0 px-2 py-1', c.width)}>
            {c.render ? c.render(row) : String((row as Record<string, unknown>)[c.key] ?? '')}
          </TableCell>
        ))}
      </TableRow>
    )
  }), [rows, selected, selection, columns, rowKey, onToggle, onOpenRow, onConfirm])

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {/* 顶部过滤区 */}
      <div className="flex shrink-0 flex-wrap items-end gap-2 border-b border-border pb-2">
        {filters.map((f) => (
          <label key={f.key} className="block">
            <span className="mb-1 block text-[11px] text-muted-foreground">{f.label}</span>
            <Input
              value={values[f.key] || ''}
              onChange={(e) => setValues((v) => ({ ...v, [f.key]: e.target.value }))}
              onKeyDown={(e) => { if (e.key === 'Enter') submit(1) }}
              placeholder={f.placeholder}
              className={cn('h-7 text-xs', f.width || 'w-40')}
            />
          </label>
        ))}
        <Button size="sm" onClick={() => submit(1)} disabled={loading}>{loading ? '查询中…' : '查询'}</Button>
      </div>

      {error && (
        <div className="shrink-0 border-b border-red-500/20 bg-red-500/10 px-2 py-1 text-[11px] text-red-600 dark:text-red-400">
          查询失败：{error}
        </div>
      )}

      {/* 表格 */}
      <div className="min-h-0 flex-1 overflow-auto">
        <Table container={false} className="table-fixed border-separate border-spacing-0 text-xs">
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              {selection === 'multi' && <TableHead className="w-8 border-t-0 border-l-0 bg-background" />}
              {columns.map((c) => (
                <TableHead key={c.key} className={cn('sticky top-0 z-10 border-t-0 border-l-0 bg-background', c.width)}>
                  {c.head}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>{body}</TableBody>
        </Table>
        {!loading && !error && rows.length === 0 && (
          <div className="p-6 text-center text-xs text-muted-foreground">{emptyHint}</div>
        )}
        {loading && rows.length === 0 && (
          <div className="p-6 text-center text-xs text-muted-foreground">正在读取…</div>
        )}
      </div>

      {/* 分页 */}
      <div className="flex shrink-0 items-center justify-between border-t border-border pt-2 text-[11px] text-muted-foreground">
        <span>共 {total} 条</span>
        <div className="flex items-center gap-1">
          <button onClick={() => submit(page - 1)} disabled={loading || page <= 1} title="上一页"
            className="rounded-ui p-1 hover:bg-accent hover:text-foreground disabled:opacity-30">
            <ChevronLeft className="h-4 w-4" />
          </button>
          <span className="whitespace-nowrap font-mono">第 {page} / {totalPages} 页</span>
          <button onClick={() => submit(page + 1)} disabled={loading || !hasMore} title="下一页"
            className="rounded-ui p-1 hover:bg-accent hover:text-foreground disabled:opacity-30">
            <ChevronRight className="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  )
}

/** Dialog 壳 + FilterTable + Footer。 */
export function QueryTableDialog<T>({ open, title, onClose, onConfirm, confirmText = '确定', ...ft }: FilterTableProps<T> & {
  open: boolean
  title: string
  onClose: () => void
  onConfirm: () => void
  confirmText?: string
}) {
  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent className="flex max-h-[min(760px,86vh)] w-[min(920px,94vw)] max-w-none flex-col">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <FilterTable {...ft} onConfirm={onConfirm} />
        <DialogFooter>
          <Button variant="outline" size="sm" onClick={onClose}>取消</Button>
          <Button size="sm" onClick={onConfirm} disabled={ft.loading || ft.selected.size === 0}>
            {confirmText}{ft.selection === 'multi' && ft.selected.size > 0 ? `（${ft.selected.size}）` : ''}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
