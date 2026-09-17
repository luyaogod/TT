import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// 字典配置页 SPA。由 `tt serve` 挂在 /dict/ 下（base 必须是 /dict/ —— 统一服务的
// 顶层 mux 同时挂着调试工作台 /debug/ 与其 API）。
//
// API 分两类：
//   /dict/api/…  字典专属（镜像、同步、PATH 安装、BDL 文档）
//   /api/…       两套页面共用的端点（hosts 配置、dbprobe、conntest…）
//
// 合并前这个配置里写着「调试服务的 /api、/mcp 代理已随调试功能移除；后续新增
// 后端接口时在此补 server.proxy」—— 现在后端就是同一个统一服务，代理补回来了。
export default defineConfig(() => {
  const target = process.env.TT_PROXY || 'http://127.0.0.1:28670'
  return {
    base: '/dict/',
    plugins: [react(), tailwindcss()],
    build: {
      outDir: '../dist/dict',
      emptyOutDir: true,
    },
    server: {
      proxy: {
        '/dict/api': { target, rewrite: (p) => p.replace(/^\/dict/, '') },
        '/api': { target },
      },
    },
  }
})
