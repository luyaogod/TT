import React from 'react'
import ReactDOM from 'react-dom/client'
import { App } from './App'
import { applyDark, resolveDark, readStoredTheme } from '../../shared/theme'
import './index.css'

// 启动即同步主题类(index.html 的内联脚本负责更早的一帧,避免首屏闪色)。
// 少了这一步,共享 tokens.css 的 .dark 变体就没人打开 —— 页面会一直是亮色。
applyDark(resolveDark(readStoredTheme()))

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
