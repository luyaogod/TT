// 启动页/欢迎页的桥。
//
// **只有这两个本地窗口挂它**，主窗口（加载 SPA 的那个）刻意不挂：那个页面的内容
// 来自 Go 服务，没有任何理由让它能调主进程的 IPC。
//
// sandbox:true 下 preload 只能 require 受限的几个模块，`electron` 的
// contextBridge / ipcRenderer 就在其中 —— 所以这里不需要打包任何依赖。

const { contextBridge, ipcRenderer } = require('electron')

contextBridge.exposeInMainWorld('ttDesktop', {
  // 版本与日志路径（启动页与欢迎页的标题栏用）
  boot: () => ipcRenderer.invoke('desktop:info'),
  // 失败页
  retry: () => ipcRenderer.send('desktop:retry'),
  openLog: () => ipcRenderer.send('desktop:open-log'),
  copyDiag: () => ipcRenderer.invoke('desktop:copy-diag').then(() => true, () => false),
  quit: () => ipcRenderer.send('desktop:quit'),
  // 欢迎页：两个深链 + 一个动作
  welcome: (where, dontShow) => ipcRenderer.send('desktop:welcome', { where, dontShow: !!dontShow }),
  installPath: () => ipcRenderer.invoke('desktop:install-path'),
})
