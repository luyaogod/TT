// 开窗字段：T100 的 ButtonEdit（可输入 + 右侧一个开窗按钮）。
//
// 两件东西分开导出：
//   - OpenWindowButton：只有图标按钮，已有 Input 的字段可以直接在右侧挂它；
//   - LookupField：label + Input + 按钮的现成组合，字段区用这个最省事。
// 组件不碰查询/取数：点按钮只是 onOpen，弹窗与数据都由调用方决定。
import { Search } from 'lucide-react'
import { Input } from '../../shared/ui'
import { cn } from '../../shared/utils'

/** 开窗按钮（Search 图标）。视觉与 Input 同高、同边框，不新增颜色档。 */
export function OpenWindowButton({ onClick, disabled, title = '选择', className }: {
  onClick: () => void
  disabled?: boolean
  title?: string
  className?: string
}) {
  return (
    <button
      type="button"
      title={title}
      aria-label={title}
      onClick={onClick}
      disabled={disabled}
      className={cn(
        'flex h-8 w-8 shrink-0 items-center justify-center rounded-ui border border-input text-muted-foreground transition-colors',
        'hover:bg-accent hover:text-foreground disabled:pointer-events-none disabled:opacity-40',
        className,
      )}
    >
      <Search className="h-4 w-4" />
    </button>
  )
}

/** label + 可编辑 Input + 开窗按钮（图标在字段内部，与输入框同一边框）。 */
export function LookupField({ label, value, onChange, onOpen, onKeyDown, placeholder, title, disabled, className, buttonTitle }: {
  label: string
  value: string
  onChange: (v: string) => void
  onOpen: () => void
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void
  placeholder?: string
  title?: string
  disabled?: boolean
  className?: string
  buttonTitle?: string
}) {
  return (
    <div className={cn('flex min-w-0 items-center gap-1.5', className)} title={title}>
      <span className="w-20 shrink-0 text-xs text-muted-foreground">{label}</span>
      <div className="relative min-w-0 flex-1">
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onKeyDown={onKeyDown}
          placeholder={placeholder}
          disabled={disabled}
          className="h-8 pr-9 text-[13px]"
        />
        <button
          type="button"
          title={buttonTitle || `选择${label}`}
          aria-label={buttonTitle || `选择${label}`}
          onClick={onOpen}
          disabled={disabled}
          className="absolute inset-y-0 right-0 flex w-8 items-center justify-center text-muted-foreground transition-colors hover:text-foreground disabled:pointer-events-none disabled:opacity-40"
        >
          <Search className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}
