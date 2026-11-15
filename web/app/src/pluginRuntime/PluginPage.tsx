// 同页插件页面（v2）：加载插件的浏览器模块，把它注册的 page 槽位渲染出来。
//
// 与 PluginFrame（legacy iframe）的分界：插件清单里有 contributions.ui 的 page
// 条目就走这里；否则仍走旧 iframe。两者共用外壳的标题栏与错误呈现。

import { useEffect } from 'react'
import { AlertTriangle, RotateCcw } from 'lucide-react'
import { IconBtn, Meta, ViewHeader } from '../Shell'
import { pluginPageEntry, pluginUIAssetURL, type PluginDescriptor } from '../plugins'
import { ensurePluginActivated, reloadPlugin, usePluginActivation } from './boot'
import { useSlot } from './registry'
import { SlotHost } from './SlotHost'

type Status = 'loading' | 'ready' | 'error'

export function PluginPage({ plugin, leading }: {
  plugin: PluginDescriptor
  leading?: React.ReactNode
}) {
  const page = pluginPageEntry(plugin)
  const url = page ? pluginUIAssetURL(plugin.id, page.entry) : null
  const act = usePluginActivation(plugin.id)
  const pageEntries = useSlot('page').filter(
    (e) => e.pluginId === plugin.id && (!e.route || e.route === plugin.route),
  )

  useEffect(() => {
    void ensurePluginActivated(plugin)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [plugin.id, plugin.version, url])

  const status: Status = act?.error ? 'error' : (act?.loading || !act ? 'loading' : 'ready')
  const error = act?.error ?? ''

  const reload = () => { void reloadPlugin(plugin) }

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
        actions={(
          <>
            <SlotHost slot="header.actions" filter={(e) => e.pluginId === plugin.id} />
            <IconBtn icon={RotateCcw} label="重新加载插件页面" onClick={reload} />
          </>
        )}
      />

      {status === 'error' && (
        <div role="alert" className="m-3 border border-red-500/30 bg-red-500/10 p-4 text-sm text-red-600 dark:text-red-400">
          <div className="flex items-center gap-2 font-medium">
            <AlertTriangle className="h-4 w-4 shrink-0" />
            <span>插件页面没有加载</span>
          </div>
          <p className="mt-2 break-all opacity-80">{error}</p>
          <button onClick={reload}
            className="mt-3 inline-flex items-center gap-1 rounded-ui border border-border px-2 py-1 text-foreground hover:bg-accent">
            <RotateCcw className="h-3.5 w-3.5" />
            重试
          </button>
        </div>
      )}

      {status !== 'error' && (
        <div className="relative min-h-0 flex-1 overflow-auto">
          {status === 'loading' && (
            <div className="flex h-full items-center justify-center p-8 text-xs text-muted-foreground">
              正在加载插件页面…
            </div>
          )}
          <SlotHost
            slot="page"
            filter={(e) => e.pluginId === plugin.id && (!e.route || e.route === plugin.route)}
          />
          {status === 'ready' && pageEntries.length === 0 && (
            <div className="flex h-full items-center justify-center p-8 text-xs text-muted-foreground">
              {plugin.title} 没有在「page」槽位注册内容
            </div>
          )}
        </div>
      )}
    </div>
  )
}
