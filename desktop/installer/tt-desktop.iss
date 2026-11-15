; TT Desktop 的 Inno Setup 安装包定义（VS Code 在 Windows 上用的也是 Inno Setup）。
;
; 为什么不是 electron-builder 的 NSIS：NSIS 生成的安装器在本机能编、跑不起来
; （深信服 EAIO 终端安全拦 NSIS 运行时：空白 NSIS 安装器一启动就 exit 2），
; 而 electron-builder 打包时必须自己跑一次安装器来生成卸载器，于是整条流水线卡死。
; Inno Setup 不依赖那套运行时：ISCC 直接编出 Setup.exe，卸载器 unins000.exe 也是编译期生成。
; 0.6.0 起整个发布统一走 Inno（CLI 安装版见 ../../installer/tt-cli.iss），NSIS 不再使用。
;
; 用法（在仓库根目录）：
;   build_desktop.bat              —— 生成 dist\TT-Desktop-<version>-x64-Setup.exe
; 或直接：
;   "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" /DAppVersion=0.6.0 desktop\installer\tt-desktop.iss
;
; 前置条件：先跑过 build_desktop.bat（或至少让 desktop\release\win-unpacked 存在）。

#define AppName "TT Desktop"
#define AppExe "TT Desktop.exe"
#define AppPublisher "luyao"
#define AppId "{{8F3C1E52-6C1B-4E3A-9C0D-2A7B5E4D1F23}"

; 版本号唯一出处是仓库根的 VERSION：build_desktop.bat 会用 /DAppVersion=... 传进来。
; 这里留一个兜底值，单独用 ISCC 编时也不至于没版本号。
#ifndef AppVersion
  #define AppVersion "0.6.0"
#endif

; 载荷 = electron-builder 的 win-unpacked（与免安装 zip 吃的是同一份）
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
UsePreviousAppDir=yes
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
; 载荷原样铺到安装目录：TT Desktop.exe、resources\tt（含 .desktop 标记）、app.asar 等
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

; 两个安装包共用同一份中文文案（错误/日志路径没覆盖的仍是英文）
#include "..\..\installer\zh-CN-messages.iss"
