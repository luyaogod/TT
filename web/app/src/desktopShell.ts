// 桌面版窗口 API 的**前端侧入口**。
//
// 浏览器里 ttDesktopShell 不存在，于是这里全部是 no-op —— 同一个前端产物在浏览器与
// 桌面版里跑，靠这一个判断分开，不需要额外构建变量。
//
// 类型只写到"用得到的那几个方法"：主进程那份见 desktop/src/shell-preload.cjs，
// 两边改名要一起改（没有共享类型仓库，别指望编译器提醒）。

export interface DesktopShellApi {
  isDesktop?: boolean
  /** 改窗口右上角原生按钮的配色（标题栏由 SPA 自己画，只有按钮归系统）。 */
  setTheme: (dark: boolean) => void
  openInBrowser: () => void
  openDataDir: () => void
  reload: () => void
  resetZoom: () => void
  checkUpdate: () => void
  openReleases: () => void
  about: () => void
  quit: () => void
}

export function desktopShell(): DesktopShellApi | undefined {
  if (typeof window === 'undefined') return undefined
  return (window as unknown as { ttDesktopShell?: DesktopShellApi }).ttDesktopShell
}

/** 把当前明暗告诉桌面壳（浏览器里是空操作）。 */
export function syncShellTheme(dark: boolean): void {
  desktopShell()?.setTheme(dark)
}
