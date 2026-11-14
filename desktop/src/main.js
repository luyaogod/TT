// TT Desktop 的主进程：**只有壳**，没有业务。
//
// 它做的四件事：
//   1. 把内置的 tt 服务拉起来（前台子进程），等它就绪拿到真实地址；
//   2. 窗口生命：启动页 → 主窗（加载服务提供的同一套 SPA）→ 失败页；
//   3. 收尾：关窗即退出，并且只收掉**自己拉起来的**那份服务；
//   4. 壳级能力：单实例、菜单、外链、日志、PATH 一键、更新提示。
//
// 界面不在这一层：所有页面与接口都由 Go 服务提供（见 internal/web）。这里一旦
// 开始写业务，浏览器版与桌面版就会各有一份真相。

// ESM 主进程里不能从 'electron' 具名 import：那个包是 CJS，导出是运行时挂上去的，
// 静态解析会报 "does not provide an export named 'BrowserWindow'"。默认导入再解构
// 才稳（Electron 的 ESM 支持就是这么用的）。
import electron from 'electron'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { appUrl, originOf, resolveExe, startServer, stopServer } from './serve.js'

const { app, BrowserWindow, Menu, clipboard, dialog, ipcMain, session, shell } = electron

const APP_ID = 'com.luyao.tt.desktop'
// 发布页从哪来：与 Go 侧 internal/update 的 RepoSlug 同一个仓库。
// 这里只在"打不开服务、读不到 source"时兜底用，正常以 /api/update 回的 source 为准。
const REPO_SLUG = 'luyaogod/TT'

// 自绘标题栏的高度与两套配色。**必须与 web/app/src/DesktopTitleBar.tsx 的 h-9 一致**：
// 那一条栏占掉 36px，原生窗口按钮也按这个高度画。
const TITLEBAR_HEIGHT = 36
// 颜色取自 SPA 的 card token（暗色 .dark .dsh-shell 的 --card、亮色默认的 --card），
// 否则右上角按钮的底色会和标题栏差一档，看起来像贴了块补丁。
const TITLEBAR_COLORS = {
  dark: { color: '#1b1b1c', symbolColor: '#cccccc' },
  light: { color: '#ffffff', symbolColor: '#3b3b3b' },
}

const here = path.dirname(fileURLToPath(import.meta.url))
// desktop/src → 仓库根（只在开发态用：打包后 exe 在 resources/tt/）
const repoRoot = path.resolve(here, '..', '..')

const S = {
  server: null,
  origin: '',
  configPath: '',
  exePath: '',
  splash: null,
  main: null,
  welcome: null,
  openingMain: false,
  starting: false,
  quitting: false,
  quittingDone: false,
  logFile: '',
  logStream: null,
  recent: [],
}

// ---------- 日志 ----------

function openLog() {
  try {
    const dir = app.getPath('logs')
    fs.mkdirSync(dir, { recursive: true })
    S.logFile = path.join(dir, 'desktop.log')
    try {
      // 单文件 5MB 轮转一份：桌面版可能被开很多天，日志不该无限长
      if (fs.statSync(S.logFile).size > 5 * 1024 * 1024) {
        const old = S.logFile + '.1'
        try {
          fs.rmSync(old, { force: true })
        } catch {
          /* 上一次的备份删不掉就覆盖，不挡启动 */
        }
        fs.renameSync(S.logFile, old)
      }
    } catch {
      /* 文件还不存在 —— 正常 */
    }
    S.logStream = fs.createWriteStream(S.logFile, { flags: 'a' })
  } catch {
    // 日志不可用不该挡住启动：退回内存环形缓冲，复制诊断还拿得到
    S.logFile = ''
  }
}

function log(msg) {
  const line = `[${new Date().toISOString()}] ${msg}`
  S.recent.push(line)
  if (S.recent.length > 160) S.recent.shift()
  try {
    S.logStream?.write(line + '\n')
  } catch {
    /* ignore */
  }
}

function diagText() {
  const lines = [
    `TT Desktop ${app.getVersion()}`,
    `Electron ${process.versions.electron} · Chromium ${process.versions.chrome} · Node ${process.versions.node}`,
    `打包: ${app.isPackaged ? '是' : '否（开发态）'}`,
    `tt: ${S.exePath || '(未定位)'}`,
    `服务: ${S.server?.url || '(未启动)'}  pid=${S.server?.pid ?? '-'}  ${S.server?.attached ? '附着' : '自持'}`,
    `配置: ${S.configPath || '(未知)'}`,
    `日志: ${S.logFile || '(不可用)'}`,
    '',
    '—— 最近日志 ——',
    ...S.recent.slice(-40),
  ]
  return lines.join('\n')
}

// ---------- 窗口状态 ----------

function windowStateFile() {
  return path.join(app.getPath('userData'), 'window.json')
}

function loadWindowState() {
  try {
    const v = JSON.parse(fs.readFileSync(windowStateFile(), 'utf8'))
    if (v && Number.isFinite(v.width) && Number.isFinite(v.height)) return v
  } catch {
    /* 第一次运行或文件坏了：用默认尺寸 */
  }
  return null
}

function saveWindowState(win) {
  try {
    if (!win || win.isDestroyed() || win.isMinimized() || win.isFullScreen()) return
    fs.mkdirSync(path.dirname(windowStateFile()), { recursive: true })
    fs.writeFileSync(windowStateFile(), JSON.stringify(win.getNormalBounds()))
  } catch {
    /* 记不住尺寸不算故障 */
  }
}

function iconPath() {
  if (app.isPackaged) return undefined // 打包后窗口图标跟随 exe
  const p = path.join(here, '..', 'build', 'icon.ico')
  return fs.existsSync(p) ? p : undefined
}

function welcomeMarkFile() {
  return path.join(app.getPath('userData'), 'welcomed.json')
}

// ---------- 启动页 / 欢迎页 ----------

function asQuery(obj) {
  const out = {}
  for (const [k, v] of Object.entries(obj || {})) out[k] = v == null ? '' : String(v)
  return out
}

function showSplash(payload) {
  const loading = (payload.state || 'loading') === 'loading'
  const width = loading ? 460 : 560
  const height = loading ? 300 : 470
  if (S.splash && !S.splash.isDestroyed()) {
    S.splash.setSize(width, height)
    S.splash.loadFile(path.join(here, 'splash.html'), { query: asQuery(payload) })
    S.splash.show()
    return S.splash
  }
  const win = new BrowserWindow({
    width,
    height,
    // 启动页/失败页没有标题栏：它只活几秒，顶一条系统浅色标题栏会闪一下 ——
    // 而 tt 是暗色的。整页可拖（见 splash.html 的 app-region），按钮自己标 no-drag。
    frame: false,
    resizable: false,
    maximizable: false,
    fullscreenable: false,
    skipTaskbar: true,
    title: 'TT Desktop',
    backgroundColor: '#181818',
    show: false,
    icon: iconPath(),
    webPreferences: {
      preload: path.join(here, 'preload.cjs'),
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: true,
    },
  })
  win.once('ready-to-show', () => win.show())
  win.loadFile(path.join(here, 'splash.html'), { query: asQuery(payload) })
  win.on('closed', () => {
    if (S.splash === win) S.splash = null
  })
  S.splash = win
  return win
}

function closeSplash() {
  if (S.splash && !S.splash.isDestroyed()) S.splash.destroy()
  S.splash = null
}

function openWelcome() {
  const win = new BrowserWindow({
    width: 780,
    height: 620,
    minWidth: 640,
    minHeight: 520,
    title: '欢迎使用 TT',
    backgroundColor: '#181818',
    show: false,
    icon: iconPath(),
    webPreferences: {
      preload: path.join(here, 'preload.cjs'),
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: true,
    },
  })
  win.once('ready-to-show', () => {
    win.show()
    closeSplash()
  })
  win.loadFile(path.join(here, 'welcome.html'))
  win.on('closed', () => {
    if (S.welcome === win) S.welcome = null
    // 欢迎页是过渡屏：关掉它（或点了任一步）就该看到主界面。
    // 但 openMain 自己会销毁这一页 —— 那时不能再递归开一次窗口（会开成"没有哈希"的
    // 那一屏，用户点了"开始配置"却落在首页）；启动过程中（S.starting）也不该开。
    if (S.openingMain || S.starting) return
    if (!S.main && S.server && !S.quitting) openMain('')
  })
  S.welcome = win
  return win
}

function closeWelcome() {
  if (S.welcome && !S.welcome.isDestroyed()) S.welcome.destroy()
  S.welcome = null
}

// ---------- 主窗口 ----------

function hardenMain(win) {
  const wc = win.webContents
  // 新窗口：一律不在应用里开，交给系统浏览器
  wc.setWindowOpenHandler(({ url }) => {
    if (/^https?:/i.test(url)) shell.openExternal(url)
    return { action: 'deny' }
  })
  // 同源之外一律拦下（SPA 是哈希路由，正常不会离开 /debug/）
  wc.on('will-navigate', (e, target) => {
    if (S.origin && target.startsWith(S.origin + '/')) return
    e.preventDefault()
    if (/^https?:/i.test(target)) shell.openExternal(target)
  })
  wc.on('will-attach-webview', (e) => e.preventDefault())
  // 菜单收掉之后，这些快捷键要自己接（原来是原生菜单的 role 提供的）。
  // 只接"窗口级"的三项：编辑类快捷键由 Chromium 自己处理，不需要再管。
  //
  // **F11 不接**：它是调试台的「步入」，页面自己要收这个键（web/app/src/App.tsx 的
  // 调试快捷键）。主进程抢先 preventDefault 的话，页面永远收不到 —— 之前接成全屏时，
  // F11 在调试台里一直是"切全屏"而不是"步入"，且没有别的入口能用键盘切全屏。
  wc.on('before-input-event', (event, input) => {
    if (input.type !== 'keyDown') return
    const key = String(input.key || '')
    const ctrl = input.control || input.meta
    if (key === 'F5' || (ctrl && (key === 'r' || key === 'R'))) {
      event.preventDefault()
      wc.reload()
      return
    }
    if (ctrl && key === '0') {
      event.preventDefault()
      wc.setZoomLevel(0)
      return
    }
    if (!app.isPackaged && ctrl && input.shift && (key === 'i' || key === 'I')) {
      event.preventDefault()
      wc.toggleDevTools()
    }
  })
  wc.on('render-process-gone', (_e, details) => {
    log(`渲染进程退出: ${details?.reason || '未知'}`)
    dropMainAndShowError({
      title: '界面进程退出了',
      detail: `原因：${details?.reason || '未知'}。点「重试」会重新加载界面。`,
    })
  })
}

function createMain(hash) {
  const saved = loadWindowState()
  const win = new BrowserWindow({
    width: saved?.width || 1440,
    height: saved?.height || 900,
    x: saved?.x,
    y: saved?.y,
    minWidth: 1024,
    minHeight: 640,
    title: 'TT — T100 人机协同调试',
    backgroundColor: '#181818',
    show: false,
    icon: iconPath(),
    // 标题栏与菜单都由应用自己画（见 web/app/src/DesktopTitleBar.tsx）：
    // 系统标题栏只跟 Windows 的明暗走，而 tt 的主题是应用内设置，两者会对不上。
    // 右上角三颗按钮仍是原生的（titleBarOverlay），配色由 SPA 通过 IPC 告诉我们。
    titleBarStyle: 'hidden',
    titleBarOverlay: { ...TITLEBAR_COLORS.dark, height: TITLEBAR_HEIGHT },
    webPreferences: {
      // 这份 preload 只给**标题栏窗口动作**（换色、打开数据目录、检查更新…），
      // 是一份白名单；业务数据一条都不经过它。
      preload: path.join(here, 'shell-preload.cjs'),
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: true,
      webviewTag: false,
      spellcheck: false,
    },
  })
  hardenMain(win)
  win.once('ready-to-show', () => {
    win.show()
    closeSplash()
  })
  win.on('close', () => saveWindowState(win))
  win.on('closed', () => {
    if (S.main === win) S.main = null
  })
  win.webContents.on('did-fail-load', (_e, code, desc, url, isMainFrame) => {
    // -3 是 ABORTED：我们主动换页时会有，不算失败
    if (!isMainFrame || code === -3) return
    log(`页面加载失败 ${code} ${desc} ${url}`)
    dropMainAndShowError({
      title: '界面没能加载',
      detail: `${desc || '加载失败'}（${code}）`,
    })
  })
  win.loadURL(appUrl(S.server.url, hash))
  return win
}

function navigateMain(hash) {
  const win = S.main
  if (!win || win.isDestroyed()) return
  const target = '#' + String(hash || '').replace(/^#/, '')
  const current = win.webContents.getURL() || ''
  if (S.origin && current.startsWith(S.origin + '/')) {
    // 已在这一页：只改哈希，让 SPA 自己切屏（不重载，调试面板不闪）
    win.webContents.executeJavaScript(`location.hash = ${JSON.stringify(target)}`).catch(() => {})
  } else {
    win.loadURL(appUrl(S.server.url, hash))
  }
}

function openMain(hash = '') {
  // 销毁欢迎页会触发它的 'closed' 回调（那里会"兜底开主窗"）。开主窗这件事由下面
  // 唯一负责，所以用一个标志把它短路掉 —— 否则用户点「开始配置」会落到首页。
  S.openingMain = true
  try {
    closeWelcome()
    if (S.main && !S.main.isDestroyed()) {
      navigateMain(hash)
      S.main.show()
      S.main.focus()
      return S.main
    }
    S.main = createMain(hash)
    return S.main
  } finally {
    S.openingMain = false
  }
}

/** 主窗口没救了：换成失败页（先建失败页再毁主窗，避免"全部窗口关闭"把应用带走）。 */
function dropMainAndShowError({ title, detail }) {
  if (S.quitting) return
  const payload = { state: 'crashed', title, detail, log: S.recent.slice(-20).join('\n') }
  showSplash(payload)
  if (S.main && !S.main.isDestroyed()) S.main.destroy()
  S.main = null
}

// ---------- 服务 ----------

async function apiJson(pathname, init = {}) {
  if (!S.server) throw new Error('本地服务未启动')
  const base = S.server.url.replace(/\/+$/, '')
  const res = await fetch(base + pathname, { ...init, signal: AbortSignal.timeout(20000) })
  const text = await res.text()
  try {
    return JSON.parse(text)
  } catch {
    return { ok: res.ok, raw: text }
  }
}

function onServerExit({ code, signal }) {
  if (S.quitting) return
  log(`服务退出 code=${code} signal=${signal}`)
  S.server = null
  dropMainAndShowError({
    title: '本地服务意外退出',
    detail: `退出码 ${code ?? signal ?? '未知'}。点「重试」会重新启动它。`,
  })
}

async function isFirstRun() {
  try {
    if (fs.existsSync(welcomeMarkFile())) return false
    const hosts = await apiJson('/api/hosts')
    const list = Array.isArray(hosts?.sshs) ? hosts.sshs : []
    return list.length === 0
  } catch {
    return false
  }
}

async function boot(retry = false) {
  if (S.starting) return
  S.starting = true
  try {
    // 先立启动页，再拆旧的：中间不留"一个窗口都没有"的瞬间（window-all-closed 会把应用带走）
    showSplash({ state: 'loading', title: '正在启动…', detail: '正在准备本地服务，第一次启动会稍慢一点。' })

    if (retry) {
      if (S.main && !S.main.isDestroyed()) S.main.destroy()
      S.main = null
      if (S.welcome && !S.welcome.isDestroyed()) S.welcome.destroy()
      S.welcome = null
      if (S.server) {
        await stopServer(S.server).catch(() => {})
        S.server = null
      }
    }

    S.exePath = resolveExe({
      isPackaged: app.isPackaged,
      resourcesPath: process.resourcesPath,
      repoRoot,
      env: process.env,
    })
    if (!fs.existsSync(S.exePath)) {
      throw new Error(
        `找不到 tt 程序：${S.exePath}\n` +
          (app.isPackaged ? '这份安装可能不完整，建议重新安装。' : '开发态请先在仓库根运行 make build。'),
      )
    }

    const srv = await startServer({
      exePath: S.exePath,
      env: process.env,
      onLog: (line) => log(`[tt] ${line}`),
      onExit: onServerExit,
    })
    S.server = srv
    S.origin = originOf(srv.url)
    S.configPath = srv.config || ''
    log(`服务就绪 ${srv.url} pid=${srv.pid} attached=${srv.attached} config=${srv.config}`)

    if (await isFirstRun()) openWelcome()
    else openMain('')
  } catch (err) {
    const msg = err?.message || String(err)
    const tail = err?.tail || ''
    log(`启动失败: ${msg}${tail ? '\n' + tail : ''}`)
    showSplash({ state: 'error', title: '本地服务没能起来', detail: msg, log: tail })
  } finally {
    S.starting = false
  }
}

// ---------- 菜单 ----------
//
// 原生菜单不设：Windows 的菜单栏只跟系统明暗走，和"应用内主题"对不上，
// 所以标题栏与菜单都交给 SPA 画（web/app/src/DesktopTitleBar.tsx），
// 菜单里的动作从 desktop:shell 那条 IPC 回来。这里只把它们接住。

/** 把窗口右上角原生按钮的配色换成与标题栏一致的那一套。 */
function applyOverlayTheme(win, dark) {
  if (!win || win.isDestroyed()) return
  try {
    win.setTitleBarOverlay({ ...(dark ? TITLEBAR_COLORS.dark : TITLEBAR_COLORS.light), height: TITLEBAR_HEIGHT })
  } catch {
    /* 老系统上 titleBarOverlay 可能不支持：按钮配色退回系统默认，不影响功能 */
  }
}

/** 标题栏发来的窗口级动作。**白名单**：认不出的类型直接丢。 */
function handleShellAction(win, type, payload) {
  switch (type) {
    case 'set-theme':
      applyOverlayTheme(win, payload === true)
      return
    case 'open-in-browser':
      if (S.server) shell.openExternal(appUrl(S.server.url))
      return
    case 'open-data-dir':
      if (S.configPath) shell.openPath(path.dirname(S.configPath))
      return
    case 'reload':
      win?.webContents.reload()
      return
    case 'reset-zoom':
      win?.webContents.setZoomLevel(0)
      return
    case 'check-update':
      void checkUpdate()
      return
    case 'open-releases':
      shell.openExternal(releaseUrl())
      return
    case 'about':
      void showAbout()
      return
    case 'quit':
      app.quit()
      return
    default:
      return
  }
}

function releaseUrl(source) {
  const slug = typeof source === 'string' && source.includes('/') ? source : REPO_SLUG
  return `https://github.com/${slug}/releases`
}

async function checkUpdate() {
  try {
    const res = await apiJson('/api/update/check', { method: 'POST' })
    if (!res?.ok) {
      await dialog.showMessageBox({
        type: 'warning',
        message: '检查更新失败',
        detail: typeof res?.error === 'string' ? res.error : '网络或代理问题，稍后再试。',
        buttons: ['好'],
        noLink: true,
      })
      return
    }
    const check = res.check || {}
    if (check.newer) {
      const { response } = await dialog.showMessageBox({
        type: 'info',
        message: `有新版本 ${check.latest}`,
        detail: `当前是 ${app.getVersion()}。桌面版请下载新的安装包覆盖安装。`,
        buttons: ['打开发布页', '以后再说'],
        defaultId: 0,
        cancelId: 1,
        noLink: true,
      })
      if (response === 0) shell.openExternal(releaseUrl(res.source))
      return
    }
    await dialog.showMessageBox({
      type: 'info',
      message: '已是最新版本',
      detail: `当前是 ${app.getVersion()}。`,
      buttons: ['好'],
      noLink: true,
    })
  } catch (err) {
    await dialog.showMessageBox({
      type: 'warning',
      message: '检查更新失败',
      detail: String(err?.message || err),
      buttons: ['好'],
      noLink: true,
    })
  }
}

async function showAbout() {
  const detail = [
    `版本 ${app.getVersion()}`,
    `服务 ${S.server?.url || '未启动'}`,
    `配置 ${S.configPath || '未知'}`,
    '',
    'tt 的命令行入口在「设置 → 应用设置」里可以加入 PATH，脚本与 AI 用它。',
  ].join('\n')
  await dialog.showMessageBox({
    type: 'info',
    message: 'TT Desktop',
    detail,
    buttons: ['好'],
    noLink: true,
  })
}

// ---------- IPC（只服务启动页/欢迎页两个本地窗口） ----------

ipcMain.handle('desktop:info', () => ({ version: app.getVersion(), logPath: S.logFile }))
ipcMain.on('desktop:retry', () => {
  boot(true)
})
ipcMain.on('desktop:open-log', () => {
  if (S.logFile) shell.openPath(S.logFile)
})
ipcMain.handle('desktop:copy-diag', () => {
  const text = diagText()
  clipboard.writeText(text)
  return text
})
ipcMain.on('desktop:quit', () => app.quit())
ipcMain.on('desktop:welcome', (_e, arg = {}) => {
  if (arg.dontShow) markWelcomed()
  openMain(arg.where === 'skills' ? 'settings/skills' : 'settings/sites')
})
ipcMain.handle('desktop:install-path', async () => {
  try {
    const st = await apiJson('/api/install', { method: 'POST' })
    if (st?.error) {
      return { ok: false, error: typeof st.error === 'string' ? st.error : st.error?.message || '写入失败' }
    }
    if (st?.inUserPath === true) return { ok: true, exeDir: st.exeDir || '' }
    return { ok: false, error: st?.manual || `没有把 ${st?.exeDir || 'tt'} 写入 PATH` }
  } catch (err) {
    return { ok: false, error: String(err?.message || err) }
  }
})

function markWelcomed() {
  try {
    fs.mkdirSync(path.dirname(welcomeMarkFile()), { recursive: true })
    fs.writeFileSync(welcomeMarkFile(), JSON.stringify({ at: new Date().toISOString() }))
  } catch {
    /* 记不住就下次再问一遍，不算故障 */
  }
}

// ---------- 应用生命周期 ----------

app.setAppUserModelId(APP_ID)

if (!app.requestSingleInstanceLock()) {
  // 第二次启动：把已有窗口叫到前面来，不另起一个服务
  app.quit()
} else {
  app.on('second-instance', () => {
    const win = S.main || S.welcome || S.splash
    if (!win || win.isDestroyed()) return
    if (win.isMinimized()) win.restore()
    win.show()
    win.focus()
  })

  app.on('window-all-closed', () => {
    if (S.quitting) return
    // 窗口切换（启动页→主窗、欢迎页→主窗）会有极短的"零窗口"窗口期，
    // 那时引用还在 —— 用户没关应用，不该把它带走。
    if (S.main || S.splash || S.welcome) return
    app.quit()
  })

  app.on('before-quit', (e) => {
    if (S.quitting) return
    e.preventDefault()
    S.quitting = true
    log('准备退出：收尾服务')
    const finish = () => {
      if (S.quittingDone) return
      S.quittingDone = true
      try {
        S.logStream?.end()
      } catch {
        /* ignore */
      }
      app.exit(0)
    }
    // 兜底：收尾最多等 8 秒，不能因为一个卡住的服务让窗口关不掉
    const timer = setTimeout(finish, 8000)
    const stopping = S.server ? stopServer(S.server) : Promise.resolve()
    stopping
      .catch((err) => log(`收尾失败: ${err?.message || err}`))
      .finally(() => {
        clearTimeout(timer)
        log('退出完成')
        finish()
      })
  })

  app.whenReady().then(() => {
    openLog()
    log(`启动 TT Desktop ${app.getVersion()} (Electron ${process.versions.electron})`)
    // 远程页面用得到剪贴板；其余权限一律不放行
    session.defaultSession.setPermissionRequestHandler((_wc, permission, callback) => {
      callback(permission === 'clipboard-sanitized-write')
    })
    session.defaultSession.on('will-download', (_e, item) => {
      log(`下载 ${item.getFilename()}`)
    })
    // 原生菜单与原生标题栏都不用：Windows 的菜单栏只跟系统明暗走，
    // 而 tt 的主题是应用内设置（默认暗色）。两者会打起来，见 applyOverlayTheme。
    Menu.setApplicationMenu(null)
    ipcMain.on('desktop:shell', (event, msg) => {
      if (!msg || typeof msg.type !== 'string') return
      const win = BrowserWindow.fromWebContents(event.sender)
      handleShellAction(win, msg.type, msg.payload)
    })
    boot(false)
  })
}
