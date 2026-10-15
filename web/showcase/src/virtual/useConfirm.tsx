// 最佳实践 ①:把同步的 window.confirm 变成 await 的确认器。
//
// 产品里有三处 window.confirm(全量重建镜像 / 字典同步 / 升级 tt),它们靠**同步布尔返回值**驱动控制流。
// 换成 Radix 之后如果就地改写,每个调用点都要拆成"先 setPending,再在回调里继续"——
// 那是最容易改错行为的一处。这里把异步收进一个 Provider,调用点只多一个 await:
//
//   const confirm = useConfirm()
//   if (!(await confirm({ title: '要现在升级 tt 吗?', description: '…' }))) return
//
// 行为等价、调用点形状不变,Dialog 保持官方实现。
import { createContext, useCallback, useContext, useRef, useState, type ReactNode } from 'react'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from '@/components/ui/alert-dialog'

export interface ConfirmAsk {
  title: string
  description?: ReactNode
  confirmText?: string
  cancelText?: string
}

const ConfirmCtx = createContext<(ask: ConfirmAsk) => Promise<boolean>>(async () => false)

export function ConfirmProvider({ children }: { children: ReactNode }) {
  const [ask, setAsk] = useState<ConfirmAsk | null>(null)
  const resolver = useRef<((ok: boolean) => void) | null>(null)

  const confirm = useCallback((a: ConfirmAsk) => {
    return new Promise<boolean>((resolve) => {
      resolver.current = resolve
      setAsk(a)
    })
  }, [])

  const settle = (ok: boolean) => {
    resolver.current?.(ok)
    resolver.current = null
    setAsk(null)
  }

  return (
    <ConfirmCtx.Provider value={confirm}>
      {children}
      <AlertDialog open={ask !== null} onOpenChange={(o) => { if (!o) settle(false) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{ask?.title}</AlertDialogTitle>
            {ask?.description && <AlertDialogDescription>{ask.description}</AlertDialogDescription>}
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => settle(false)}>{ask?.cancelText || '取消'}</AlertDialogCancel>
            <AlertDialogAction onClick={() => settle(true)}>{ask?.confirmText || '确定'}</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </ConfirmCtx.Provider>
  )
}

export const useConfirm = () => useContext(ConfirmCtx)
