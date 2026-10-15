import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { ThemeProvider } from 'next-themes'
import { TooltipProvider } from '@/components/ui/tooltip'
import { App } from './App'
import './index.css'

// 主题用官方推荐的那一套:next-themes(attribute="class")。
// 键名与取值刻意对齐产品现有的 theme.ts:storageKey='tt.theme'、值域 light/dark/system,
// 所以"是否把产品也换成 next-themes"这个决策不会造成用户偏好丢失。
// index.html 里的首屏脚本仍保留:Vite 下 next-themes 是挂载后才生效,没有它就闪一下亮色。
// 分享/取证用:?theme=light|dark 直接把这份偏好写进 localStorage,再把参数从地址栏摘掉。
// 走的仍是同一份偏好(tt.theme),不是第二套主题机制 —— 摘掉参数之后右上角的切换照常可用。
const forced = new URLSearchParams(window.location.search).get('theme')
if (forced === 'light' || forced === 'dark') {
  try {
    localStorage.setItem('tt.theme', forced)
    history.replaceState(null, '', window.location.pathname + window.location.hash)
  } catch { /* 无痕模式等写不进去:保持默认 */ }
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ThemeProvider attribute="class" defaultTheme="dark" enableSystem storageKey="tt.theme" disableTransitionOnChange>
      <TooltipProvider>
        <App />
      </TooltipProvider>
    </ThemeProvider>
  </StrictMode>,
)
