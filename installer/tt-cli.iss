; TT CLI 的 Inno Setup 安装包定义（用户级安装，不需要管理员）。
;
; 为什么是 Inno Setup（0.6.0 起放弃 MSI/WiX）：
;   · 一条工具链同时出 CLI 与桌面版，CLI 安装版和桌面版共用同一份中文文案；
;   · 编译期生成卸载器，打包机上不需要先"跑一次安装器"（NSIS 就是这么死的）。
;
; 为什么是**用户级**：
;   · 装到 %LOCALAPPDATA%\Programs\TT，不碰 Program Files —— 全程不需要管理员，
;     与工具一贯的取向一致（tt install path 也只动 HKCU\Environment，绝不碰系统 PATH）；
;   · PATH 由 tt 自己加/摘（[Run] / [UninstallRun] 调 `tt install path [--remove]`），
;     与设置页「加入 PATH」是同一份实现，不在这里复制一套注册表逻辑。
;
; 为什么装完不写 .portable：配置该落在 %APPDATA%\tt\config.json（用户可写），
; 而不是程序目录（升级会整目录替换，且普通用户写不进去）。build_cli_setup.bat 已经在
; 暂存时删掉了 .portable 与 config.json，这里再断言一次。
;
; 用法（仓库根目录）：
;   build_cli_setup.bat              -> dist\TT-<version>-x64-Setup.exe
; 或直接：
;   "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" /DAppVersion=0.6.0 installer\tt-cli.iss
;
; 前置条件：先跑过 build_portable.bat（构建脚本会替你做）。

#define AppName "TT"
#define AppExe "tt.exe"
#define AppPublisher "luyao"
#define AppId "{{5D2A8C41-7E63-4B09-8F1A-3C6E9D0B7A52}"

; 版本号唯一出处是仓库根的 VERSION：build_cli_setup.bat 会用 /DAppVersion=... 传进来。
; 这里留一个兜底值，单独用 ISCC 编时也不至于没版本号。
#ifndef AppVersion
  #define AppVersion "0.6.0"
#endif

; 载荷 = build_cli_setup.bat 暂存的安装目录（便携包去掉 .portable 与 config.json）
#define Payload "..\.tmp-build\tt-setup"

[Setup]
AppId={#AppId}
AppName={#AppName}
AppVersion={#AppVersion}
AppVerName={#AppName} {#AppVersion}
AppPublisher={#AppPublisher}
AppPublisherURL=https://github.com/luyaogod/TT
AppSupportURL=https://github.com/luyaogod/TT
AppUpdatesURL=https://github.com/luyaogod/TT/releases
DefaultDirName={localappdata}\Programs\{#AppName}
; 升级装到上次那个目录（用户改过安装位置时也认），而不是又铺一份
UsePreviousAppDir=yes
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
; 只做 per-user 安装：不加 PrivilegesRequiredOverridesAllowed=dialog，
; 否则双击第一屏会多出一步「为所有用户 / 仅为我」的英文选择框（实测多余且打断了向导）
PrivilegesRequired=lowest
OutputDir=..\dist
OutputBaseFilename=TT-{#AppVersion}-x64-Setup
SetupIconFile=icon.ico
UninstallDisplayIcon={app}\{#AppExe}
UninstallDisplayName={#AppName}
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
; Inno 6 默认跳过欢迎页；显式打开，安装流程才是"欢迎 → 选位置 → 安装 → 完成"
DisableWelcomePage=no
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
; 升级时台后服务（tt serve）可能正跑着安装目录里的 tt.exe：让安装器按常规流程
; 发现"文件在用"并让用户关掉；静默升级（tt update）由更新器自己先停服务。
CloseApplications=yes
RestartApplications=no
; PATH 是 HKCU 的：改完广播一次环境变量变更，新开的终端才看得到 tt
ChangesEnvironment=yes
VersionInfoCompany={#AppPublisher}
VersionInfoProductName={#AppName}
VersionInfoProductVersion={#AppVersion}
VersionInfoDescription={#AppName} 安装程序

[Files]
; 载荷原样铺到安装目录：tt.exe、skills\、tzs\（引擎 + 设计器）、README.md、
; config.example.json。不含 .portable / config.json（构建脚本已删，这里断言）。
Source: "{#Payload}\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
; CLI 装完没有图标可点，开始菜单里放 README 入口，至少让人找得到文档。
Name: "{autoprograms}\{#AppName} 使用说明"; Filename: "{app}\README.md"

[Run]
; 用户 PATH（HKCU）：与 `tt install path` / 设置页是同一份实现，失败不阻断安装
Filename: "{app}\{#AppExe}"; Parameters: "install path"; StatusMsg: "正在把 TT 加入用户 PATH…"; Flags: runhidden waituntilterminated
; 随包的 BDL 文档：目标取 config.json 的 bdldoc.dir，没有就落 %APPDATA%\ttBDL-Markdown
Filename: "{app}\{#AppExe}"; Parameters: "install bdl"; StatusMsg: "正在解压 BDL 文档…"; Flags: runhidden waituntilterminated

[UninstallRun]
; 卸载时先把安装目录从用户 PATH 里摘掉，再删文件（tt.exe 此刻还在）
Filename: "{app}\{#AppExe}"; Parameters: "install path --remove"; RunOnceId: "RemoveUserPath"; Flags: runhidden waituntilterminated skipifdoesntexist

[UninstallDelete]
; **不要递归删 {app}**（0.6.1 起）：程序文件由卸载日志逐条删除（升级是往旧日志追加，
; 旧版本文件也在里面）；用户放在安装目录里的东西（例如把 mirror.dir 指到 {app}\mirror 的
; 源码镜像）不在日志里，递归删 {app} 会把它一并删掉。删不删镜像由人在卸载时选（见 [Code]），
; 静默默认保留；%APPDATA%\tt（配置与业务数据）**永远不动**，与其它软件一致。

#include "zh-CN-messages.iss"

[Code]
const
  UninstallKey = 'Software\Microsoft\Windows\CurrentVersion\Uninstall';

// 从 0.5.x 的 MSI 升级过来时，旧产品在「应用和功能」里还留着一条登记，而且它的
// 卸载程序删的正是我们现在的安装目录 —— 不先摘掉，用户以后点它会把 0.6.0 删掉。
// 只认 DisplayName='TT' + Publisher='luyao' + WindowsInstaller=1 这一条，
// 静默卸载失败不阻断本次安装（剩下的只是多一条登记，安装本身照常）。
procedure RemoveLegacyMSI;
var
  Names: TArrayOfString;
  I: Integer;
  Key, Display, Publisher: String;
  WinInst: Cardinal;
  ResultCode: Integer;
begin
  if not RegGetSubkeyNames(HKCU, UninstallKey, Names) then
    Exit;
  for I := 0 to GetArrayLength(Names) - 1 do
  begin
    Key := UninstallKey + '\' + Names[I];
    if RegQueryDWordValue(HKCU, Key, 'WindowsInstaller', WinInst) and (WinInst = 1)
       and RegQueryStringValue(HKCU, Key, 'DisplayName', Display) and (CompareText(Display, 'TT') = 0)
       and RegQueryStringValue(HKCU, Key, 'Publisher', Publisher) and (CompareText(Publisher, 'luyao') = 0) then
    begin
      Log('检测到旧 MSI 安装（' + Names[I] + '），先静默卸载它');
      Exec(ExpandConstant('{sys}\msiexec.exe'), '/x ' + AddQuotes(Names[I]) + ' /qn /norestart', '',
        SW_HIDE, ewWaitUntilTerminated, ResultCode);
      Log('msiexec /x 退出码 ' + IntToStr(ResultCode));
    end;
  end;
end;

function PrepareToInstall(var NeedsRestart: Boolean): String;
begin
  RemoveLegacyMSI;
  Result := '';
end;

// 卸载时"删不删用户镜像"是三态（与桌面版同一套口径）：
//   /DELETEMIRROR      脚本/静默卸载显式要求删
//   /VERYSILENT|/SILENT 没给上面那个开关 → 保留（静默不能弹窗，默认最安全）
//   交互卸载           弹「是/否」，默认按钮是「否」
function HasUninstallParam(const Value: String): Boolean;
var
  I: Integer;
begin
  Result := False;
  for I := 1 to ParamCount do
    if CompareText(ParamStr(I), Value) = 0 then
    begin
      Result := True;
      Exit;
    end;
end;

function ShouldDeleteMirror(): Boolean;
begin
  Result := False;
  if HasUninstallParam('/DELETEMIRROR') then
  begin
    Result := True;
    Exit;
  end;
  if UninstallSilent then
    Exit;
  Result := MsgBox(ExpandConstant('{cm:UninstallMirrorPrompt}'), mbConfirmation,
    MB_YESNO or MB_DEFBUTTON2) = IDYES;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if (CurUninstallStep = usPostUninstall) and ShouldDeleteMirror() then
  begin
    DelTree(ExpandConstant('{app}\mirror'), True, True, True);
    RemoveDir(ExpandConstant('{app}'));
  end;
end;
