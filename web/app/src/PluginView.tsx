// 第三方插件页面容器：插件代码只在 iframe 里加载，主 SPA 不静态 import 它。
//
// 这一层只做四件事：
//   1. 把 #plugin/<route> 对到后端注册表里的真实插件（未知/停用 → 显式错误页）；
//   2. 用 sandbox="allow-scripts" 的 iframe 承载插件页面（不加任何同源/导航/弹窗权限）；
//   3. 每个 iframe 实例一套随机 session + nonce 的消息桥（见 pluginBridge.ts）；
//   4. 把被拒绝的消息、桥错误和资源不可用摆在界面上 —— 不静默失败。
//
// 插件页不参与宿主页面的 keep-alive：切走就卸载 iframe、销毁桥并取消在途请求，
// 第三方代码不该在后台继续跑。

import { useEffect, useRef, useState, type ReactNode } from 'react'
import { AlertTriangle, RotateCcw } from 'lucide-react'
import { IconBtn, Meta, ViewHeader } from './Shell'
import { pluginPageEntry, pluginUIPath, resolvePluginPage, type PluginDescriptor, type PluginPageErrorReason } from './plugins'
import { PluginPage } from './pluginRuntime/PluginPage'
import { PluginBridge, createFetchTransport } from './pluginBridge'
import type { BridgeErrorBody } from './pluginProtocol'

export interface PluginViewProps {
  /** 片段里的插件路由（#plugin/<route>）。 */
  route: string
  /** 片段语法就不合法（parseRoute 给的是 plugin-error），不必等注册表。 */
  malformed?: boolean
  plugins: PluginDescriptor[]
  loaded: boolean
  /** 读插件清单失败的原因；空串 = 没失败。 */
  loadError: string
  onRetry: () => void
  /** 侧栏收起后借标题行保留展开入口。 */
  leading?: ReactNode
}

export function PluginView({
  route, malformed, plugins, loaded, loadError, onRetry, leading,
}: PluginViewProps) {
  if (malformed) {
    return <PluginErrorPage leading={leading} route={route} reason="malformed" onRetry={onRetry} />
  }
  if (loadError) {
    return (
      <PluginErrorPage
        leading={leading}
        route={route}
        reason="unknown"
        detail={`读取插件清单失败：${loadError}`}
        onRetry={onRetry}
      />
    )
  }
  if (!loaded) {
    return (
      <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
        <ViewHeader title="插件页面" leading={leading} />
        <div className="flex flex-1 items-center justify-center p-8 text-xs text-muted-foreground">
          正在读取插件清单…
        </div>
      </div>
    )
  }

  const resolution = resolvePluginPage(route, plugins)
  if (resolution.kind === 'error') {
    return (
      <PluginErrorPage
        leading={leading}
        route={route}
        reason={resolution.reason}
        plugin={resolution.plugin}
        onRetry={onRetry}
      />
    )
  }
  // v2：清单里有 page 槽位贡献 → 同页渲染；否则走 legacy iframe。
  if (pluginPageEntry(resolution.plugin)) {
    return <PluginPage plugin={resolution.plugin} leading={leading} />
  }
  return <PluginFrame plugin={resolution.plugin} route={route} leading={leading} />
}

/** 插件页面本体的唯一入口：iframe + 消息桥生命周期。 */
function PluginFrame({ plugin, route, leading }: {
  plugin: PluginDescriptor
  route: string
  leading?: ReactNode
}) {
  const frameRef = useRef<HTMLIFrameElement | null>(null)
  const bridgeRef = useRef<PluginBridge | null>(null)
  const [bridgeError, setBridgeError] = useState<BridgeErrorBody | null>(null)
  // 重新加载 = 换一个 iframe 实例（新 session/nonce），不是只改 src 字符串。
  const [reloadKey, setReloadKey] = useState(0)
  const src = pluginUIPath(plugin.id)

  useEffect(() => {
    const bridge = new PluginBridge({
      pluginId: plugin.id,
      route,
      transport: createFetchTransport(),
      onError: (e) => setBridgeError(e),
    })
    bridgeRef.current = bridge
    const onMessage = (event: MessageEvent) => {
      // 授权与分发都在桥里；这里只把事件交给它，不自己解析 data。
      bridge.handleMessage(event)
    }
    window.addEventListener('message', onMessage)
    return () => {
      window.removeEventListener('message', onMessage)
      bridge.destroy()
      if (bridgeRef.current === bridge) bridgeRef.current = null
    }
    // reloadKey 进依赖：重新加载必须换掉 nonce，而不是复用旧桥。
  }, [plugin.id, plugin.version, route, reloadKey])

  const reload = () => {
    setBridgeError(null)
    setReloadKey((n) => n + 1)
  }

  if (!src) {
    return <PluginErrorPage leading={leading} route={route} reason="invalid" plugin={plugin} />
  }

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
      <ViewHeader
        title={plugin.title}
        leading={leading}
        meta={(
          <>
            <Meta title={plugin.id}>{plugin.id}</Meta>
            <Meta>v{plugin.version || '未知'}</Meta>
            {plugin.healthy === false && (
              <span className="text-red-600 dark:text-red-400" title={plugin.lastError || '插件服务不健康'}>
                服务不健康
              </span>
            )}
          </>
        )}
        actions={<IconBtn icon={RotateCcw} label="重新加载插件页面" onClick={reload} />}
      />

      {bridgeError && (
        <div role="alert"
          className="flex shrink-0 items-center gap-2 border-b border-red-500/20 bg-red-500/10 px-3 py-1.5 text-xs text-red-600 dark:text-red-400">
          <AlertTriangle className="h-3.5 w-3.5 shrink-0" />
          <span className="min-w-0 flex-1 break-all">
            消息桥拒绝了页面的一条消息（{bridgeError.code}）：{bridgeError.message}
          </span>
          <button onClick={() => setBridgeError(null)} title="关闭这条提示"
            className="shrink-0 border border-red-500/30 px-2 py-0.5 hover:bg-red-500/10">
            关闭
          </button>
        </div>
      )}

      {/* sandbox 只给 allow-scripts。不加 allow-same-origin（否则同源后能摸 DOM/store）、
          不加 allow-top-navigation（不能把宿主带走）、不加 allow-popups（不能开窗写父页面）。
          src 固定从活跃版本目录出（/plugins/<id>/ui/），不拼任何查询串。 */}
      <iframe
        key={`${plugin.id}@${plugin.version}:${reloadKey}`}
        ref={frameRef}
        title={plugin.title}
        src={src}
        sandbox="allow-scripts"
        referrerPolicy="no-referrer"
        onLoad={() => bridgeRef.current?.attach(frameRef.current)}
        className="min-h-0 w-full flex-1 border-0 bg-background"
      />
    </div>
  )
}

/** 插件页面的显式失败态：说清楚是哪一种失败，并给一个可重试的动作。 */
function PluginErrorPage({
  leading, route, reason, plugin, detail, onRetry,
}: {
  leading?: ReactNode
  route: string
  reason: PluginPageErrorReason
  plugin?: PluginDescriptor
  detail?: string
  onRetry?: () => void
}) {
  const title = plugin ? plugin.title : route || '插件页面'
  const message = errorMessageOf(reason, plugin, detail)
  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
      <ViewHeader title={title} leading={leading} meta={<Meta>#plugin/{route || '（空）'}</Meta>} />
      <div className="flex flex-1 items-center justify-center p-8">
        <div className="w-full max-w-xl border border-border bg-card p-5 text-sm">
          <div className="flex items-center gap-2 text-foreground">
            <AlertTriangle className="h-4 w-4 shrink-0 text-amber-500" />
            <span className="font-medium">插件页面没有加载</span>
          </div>
          <p className="mt-3 text-muted-foreground">{message}</p>
          {plugin && (
            <p className="mt-2 break-all text-[11px] text-muted-foreground">
              {plugin.id} · v{plugin.version || '未知'}
              {plugin.lastError ? ` · ${plugin.lastError}` : ''}
            </p>
          )}
          <div className="mt-4 flex items-center gap-2">
            {onRetry && (
              <button onClick={onRetry}
                className="inline-flex items-center gap-1 rounded-ui border border-border px-2 py-1 text-foreground hover:bg-accent">
                <RotateCcw className="h-3.5 w-3.5" />
                重新读取插件清单
              </button>
            )}
            {reason === 'disabled' && plugin && (
              <span className="text-[11px] text-muted-foreground">
                启用插件：tt plugin enable {plugin.id}
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function errorMessageOf(
  reason: PluginPageErrorReason, plugin?: PluginDescriptor, detail?: string,
): string {
  if (detail) return detail
  switch (reason) {
    case 'malformed':
      return '地址里的插件路由不合法。插件路由只能是小写字母、数字与连字符（最长 32 个字符），形如 #plugin/<route>。'
    case 'disabled':
      return plugin
        ? `插件「${plugin.title}」当前未启用，它的页面不会加载。`
        : '这个插件页面所属的插件当前未启用。'
    case 'invalid':
      return '插件清单里的页面标识不合法，为避免拼出错误地址，宿主没有加载它。'
    default:
      return '没有已安装的插件提供这个页面。它可能已被卸载，或路由写错了。'
  }
}
