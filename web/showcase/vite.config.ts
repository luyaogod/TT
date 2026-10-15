import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath } from 'node:url'

// UI 决策演示台:独立子应用,只服务"看与决策",不是产品的一部分。
// 因此它 base 为根、产物落在自己目录下,绝不写入 ../dist(生产 SPA 的产物,会被 Go 内嵌)。
export default defineConfig({
  base: '/',
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  server: {
    port: 5174,
    strictPort: false,
  },
})
