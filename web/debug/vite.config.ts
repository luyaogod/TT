import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// 调试工作台 SPA。由 `tt serve` 挂在 /debug/ 下 —— 统一服务的顶层 mux 同时挂着
// 字典页 /dict/ 与其 API，所以 base 必须是 /debug/，否则两套页面的资源路径会搅在一起。
//
// API 分两类：
//   /debug/api/…  调试专属（会话、断点、WebSocket）—— 对外是 /debug/api/，后端挂载后是 /api/
//   /api/…        两套页面共用的端点（hosts 配置、dbprobe、conntest…）
//
// 开发时 vite 在 /debug/ 上服务 SPA，并把上面两类都代理到后端。
// 后端地址默认 127.0.0.1:28670（tt serve 的默认监听）；端口被占用顺延时用 TT_PROXY 指定。
export default defineConfig(() => {
  const target = process.env.TT_PROXY || 'http://127.0.0.1:28670'
  return {
    base: '/debug/',
    plugins: [react(), tailwindcss()],
    build: {
      outDir: '../dist/debug',
      emptyOutDir: true,
      chunkSizeWarningLimit: 9000,
    },
    server: {
      proxy: {
        // 调试专属接口：去掉 /debug 前缀后转给后端（后端内部按 /api/… 注册）
        '/debug/api': { target, ws: true, rewrite: (p) => p.replace(/^\/debug/, '') },
        // 共用接口：原样转发
        '/api': { target, ws: true },
      },
    },
  }
})
