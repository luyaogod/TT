// 渲染检查:用 Vite 的 SSR 通道把每一屏真实渲染一遍,再断言关键内容确实出现在结果里。
//
// 为什么需要它:演示台没有测试,而"能构建"不等于"能打开" ——
// Radix 的浮层组件缺 Provider、组件被误用、导入名写错,这些类型检查都拦不住,只有在渲染时才炸。
//
//   cd web/showcase && npm run check:render
//
// 它不覆盖:交互行为(点击、键盘、拖拽)、样式是否正确、以及任何与后端有关的东西。
import { createServer } from 'vite'

const EXPECTED = {
  overview: ['决策', 'neutral', '迁移台账'],
  foundation: ['保存', '服务名称', '断点持久化'],
  workbench: ['调用栈', '变量监视', '原始协议流'],
  interfaces: ['服务名称', 'icd.erp.wo.get', '基本信息'],
  settings: ['站点管理', '环境清单', '工作区目录'],
  patterns: ['Monaco', '变量悬浮卡片', '协作模式'],
  // 目标态虚拟应用:四个视图都渲染一遍,断言各自的关键内容在位
  'virtual-debug': ['目标态预览', '调用栈', '原始协议流', 'bsft001_wf.4gl'],
  'virtual-logs': ['服务名称', 'icd.erp.wo.get', '仅失败', '复制标识'],
  'virtual-test': ['接口方式', 'awsp920', '请求报文(JSON / XML)', '执行'],
  'virtual-settings': ['站点管理', '环境清单', '工作区目录', '数据库账号(schema)'],
  'virtual-settings-dict': ['查询数据源', '源码镜像', '数据同步', '开始同步'],
  'virtual-settings-app': ['外观', '跟随系统', '运行信息', '清除缓存'],
}

const server = await createServer({
  server: { middlewareMode: true },
  appType: 'custom',
  logLevel: 'error',
})

let failed = 0
let done = 0
try {
  const { renderAll, BLOCK_IDS } = await server.ssrLoadModule('/scripts/smoke-entry.tsx')
  const out = renderAll()
  for (const [name, needles] of Object.entries(EXPECTED)) {
    const html = out[name]
    if (typeof html !== 'string' || html.length < 500) {
      console.error(`FAIL ${name}:渲染结果过短(${html?.length ?? 0} 字符)`)
      failed++
      continue
    }
    const missing = needles.filter((n) => !html.includes(n))
    if (missing.length) {
      console.error(`FAIL ${name}:缺少关键内容 ${JSON.stringify(missing)}`)
      failed++
      continue
    }
    // 结构不变量:台账里登记的每一个功能点都必须在这一屏里渲染出锚点,
    // 否则左导航与总览台账的"去看"会跳到空处(静默失效,没有任何报错)。
    const ids = BLOCK_IDS[name] || []
    const noAnchor = ids.filter((id) => !html.includes(`id="${id}"`))
    if (noAnchor.length) {
      console.error(`FAIL ${name}:台账登记了但没有渲染锚点的功能点 ${JSON.stringify(noAnchor)}`)
      failed++
      continue
    }
    console.log(`ok   ${name}  ${html.length} 字符 · ${needles.length} 项关键内容 · ${ids.length} 个锚点`)
    done++
  }
} catch (e) {
  console.error('FAIL 渲染抛错:', e?.stack || e?.message || e)
  failed++
} finally {
  await server.close()
}

if (failed) {
  console.error(`\n渲染检查失败:${failed} 屏`)
  process.exit(1)
}
console.log(`\n渲染检查通过:${done} 屏全部渲染成功`)
