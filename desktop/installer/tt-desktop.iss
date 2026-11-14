; TT Desktop 的 Inno Setup 安装包定义（VS Code 在 Windows 上用的也是 Inno Setup）。
;
; 为什么不是 electron-builder 的 NSIS：NSIS 生成的安装器在本机能编、跑不起来
; （深信服 EAIO 终端安全拦 NSIS 运行时：空白 NSIS 安装器一启动就 exit 2），
; 而 electron-builder 打包时必须自己跑一次安装器来生成卸载器，于是整条流水线卡死。
; Inno Setup 不依赖那套运行时：ISCC 直接编出 Setup.exe，卸载器 unins000.exe 也是编译期生成。
;
; 用法（在仓库根目录）：
;   build_desktop_setup.bat          —— 生成 dist\TT-Desktop-<version>-Setup.exe
; 或直接：
;   "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" desktop\installer\tt-desktop.iss
;
; 前置条件：先跑过 build_desktop.bat（或至少让 desktop\release\win-unpacked 存在）。

#define AppName "TT Desktop"
; 版本号唯一出处是仓库根的 VERSION：build_desktop.bat /inno 会用 /DAppVersion=... 传进来。
; 这里留一个兜底值，单独用 ISCC 编时也不至于没版本号。
#ifndef AppVersion
  #define AppVersion "0.5.3"
#endif
#define AppPublisher "luyao"
#define AppExe "TT Desktop.exe"
#define AppId "{{8F3C1E52-6C1B-4E3A-9C0D-2A7B5E4D1F23}"
; 载荷 = electron-builder 的 win-unpacked（与 nsis/zip 目标吃的是同一份）
#define Payload "..\release\win-unpacked"

[Setup]
AppId={#AppId}
AppName={#AppName}
AppVersion={#AppVersion}
AppVerName={#AppName} {#AppVersion}
AppPublisher={#AppPublisher}
AppPublisherURL=https://github.com/luyaogod/TT
AppSupportURL=https://github.com/luyaogod/TT
AppUpdatesURL=https://github.com/luyaogod/TT/releases
; 装到当前用户目录：不需要管理员，和之前 MSI 的 perUser 策略一致
DefaultDirName={localappdata}\Programs\{#AppName}
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
; 只做 per-user 安装：不加 PrivilegesRequiredOverridesAllowed=dialog，
; 否则双击第一屏会多出一步「为所有用户 / 仅为我」的英文选择框（实测多余且打断了向导）
PrivilegesRequired=lowest
OutputDir=..\..\dist
OutputBaseFilename=TT-Desktop-{#AppVersion}-x64-Setup
SetupIconFile=..\build\icon.ico
UninstallDisplayIcon={app}\{#AppExe}
UninstallDisplayName={#AppName}
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
; Inno 6 默认跳过欢迎页；显式打开，安装流程才是"欢迎 → 选位置 → 附加任务 → 安装 → 完成"
DisableWelcomePage=no
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
VersionInfoCompany={#AppPublisher}
VersionInfoProductName={#AppName}
VersionInfoProductVersion={#AppVersion}
VersionInfoDescription={#AppName} 安装程序

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"

[Files]
; 载荷原样铺到安装目录：tt.exe、resources\tt（含 .desktop 标记）、app.asar 等
Source: "{#Payload}\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{autoprograms}\{#AppName}"; Filename: "{app}\{#AppExe}"; WorkingDir: "{app}"
Name: "{autodesktop}\{#AppName}"; Filename: "{app}\{#AppExe}"; WorkingDir: "{app}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#AppExe}"; Description: "{cm:LaunchProgram,{#AppName}}"; Flags: nowait postinstall skipifsilent

[UninstallDelete]
; 装进去的整个目录都带走（升级留下的旧文件也一并清掉）
Type: filesandordirs; Name: "{app}"
; 壳自己的缓存/窗口状态/日志；%APPDATA%\tt（配置与业务数据）**不动**，与其它软件一致
Type: filesandordirs; Name: "{userappdata}\TT Desktop"

[Messages]
; Inno Setup 6 官方语言包不含简体中文（Languages 目录里没有 ChineseSimplified.isl），
; 这里按 Default.isl 的消息名把向导里用户会看到的文案覆盖成中文。未覆盖的只剩错误/日志路径。
SetupAppTitle=安装程序
SetupWindowTitle=安装 - %1
UninstallAppTitle=卸载
UninstallAppFullTitle=卸载 %1
ConfirmTitle=确认
ExitSetupTitle=退出安装
ExitSetupMessage=安装尚未完成。现在退出，程序不会被安装。%n%n可以稍后重新运行安装程序来完成安装。%n%n要退出吗？
UninstallAppRunningError=卸载程序检测到 %1 正在运行。%n%n请先关闭它，再点击「确定」继续。
SelectLanguageTitle=选择安装语言
SelectLanguageLabel=请选择安装过程中使用的语言。
ButtonBack=< 上一步(&B)
ButtonNext=下一步(&N) >
ButtonInstall=安装(&I)
ButtonOK=确定
ButtonCancel=取消
ButtonYes=是(&Y)
ButtonNo=否(&N)
ButtonFinish=完成(&F)
ButtonBrowse=浏览(&B)…
ButtonWizardBrowse=浏览(&R)…
ButtonNewFolder=新建文件夹(&M)
ClickNext=点击「下一步」继续，或点击「取消」退出安装。
WelcomeLabel1=欢迎使用 [name] 安装向导
WelcomeLabel2=现在将把 [name/ver] 安装到你的电脑上。%n%n继续之前，建议先关闭其它正在运行的程序。
WizardSelectDir=选择安装位置
SelectDirDesc=[name] 要安装到哪里？
SelectDirLabel3=安装程序会把 [name] 安装到下面这个文件夹。
SelectDirBrowseLabel=要继续，请点击「下一步」。想换一个文件夹，请点击「浏览」。
DiskSpaceMBLabel=至少需要 [mb] MB 的磁盘空间。
WizardSelectTasks=选择附加任务
SelectTasksDesc=安装时需要额外做哪些事？
SelectTasksLabel2=请选择安装 [name] 时要执行的附加任务，然后点击「下一步」。
WizardReady=准备安装
ReadyLabel1=安装程序已经准备好，可以开始把 [name] 安装到你的电脑上了。
ReadyLabel2a=点击「安装」开始安装；想检查或修改设置，请点击「上一步」。
ReadyLabel2b=点击「安装」开始安装。
ReadyMemoDir=安装位置：
ReadyMemoTasks=附加任务：
WizardInstalling=正在安装
InstallingLabel=正在安装 [name]，请稍候…
PreparingDesc=安装程序正在做准备工作…
FinishedHeadingLabel=[name] 安装完成
FinishedLabelNoIcons=[name] 已安装到你的电脑上。
FinishedLabel=[name] 已安装到你的电脑上，可以从快捷方式启动它。
StatusExtractFiles=正在解压文件…
StatusCreateDirs=正在创建文件夹…
StatusCreateIcons=正在创建快捷方式…
StatusCreateRegistryEntries=正在写入注册表…
StatusSavingUninstall=正在保存卸载信息…
StatusRunProgram=正在完成安装…
StatusRollback=正在回滚修改…
ConfirmUninstall=确定要完全移除 %1 及其所有组件吗？
UninstallUninstalling=正在卸载 %1，请稍候…
UninstallUninstalled=%1 已从你的电脑上卸载。
UninstallStatusLabel=正在移除 %1，请稍候…

[CustomMessages]
CreateDesktopIcon=创建桌面快捷方式(&D)
LaunchProgram=运行 %1
AdditionalIcons=附加图标：