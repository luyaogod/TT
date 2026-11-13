// 桌面版主窗口的桥。**只暴露窗口级动作**，不含任何业务数据。
//
// 为什么主窗口现在也要挂 preload 了：原生标题栏与菜单被收掉之后，标题栏由 SPA 自己画
// （见 web/app/src/DesktopTitleBar.tsx），它需要两个东西是网页做不到的：
//   1. 让 Windows 的原生窗口按钮跟着应用明暗色走（titleBarOverlay 的颜色要主进程改）；
//   2. 菜单里的动作（打开数据目录、检查更新、退出…）都落在主进程。
//
// 所以这里是**一份白名单式的窗口 API**：每个方法不带参数（目标由主进程决定），
// 不暴露任意 IPC、不暴露文件系统、不暴露 Electron 对象。要加能力先问一句
// "网页自己能不能做" —— 能做的就不该长在这里。

const { contextBridge, ipcRenderer } = require('electron')

const send = (type, payload) => ipcRenderer.send('desktop:shell', { type, payload })

contextBridge.exposeInMainWorld('ttDesktopShell', {
  isDesktop: true,
  // 明暗色：只影响窗口右上角那三颗原生按钮的底色与符号色
  setTheme: (dark) => send('set-theme', !!dark),
  openInBrowser: () => send('open-in-browser'),
  openDataDir: () => send('open-data-dir'),
  reload: () => send('reload'),
  resetZoom: () => send('reset-zoom'),
  toggleFullscreen: () => send('toggle-fullscreen'),
  checkUpdate: () => send('check-update'),
  openReleases: () => send('open-releases'),
  about: () => send('about'),
  quit: () => send('quit'),
})
