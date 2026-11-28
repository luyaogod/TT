; TT 安装包共用中文文案（CLI 安装版 installer\tt-cli.iss 与桌面版
; desktop\installer\tt-desktop.iss 都 #include 它）。
;
; 为什么要覆盖：Inno Setup 6 官方语言包不含简体中文（Languages 目录里没有
; ChineseSimplified.isl），而没说中文的安装向导对这批用户就是"下一步在哪"都要猜。
; 这里按 Default.isl 的消息名把用户会看到的文案换成中文；错误/日志路径之类没覆盖的
; 仍是英文。
;
; 只放文案，不放产品名/版本/架构 —— 那些是各安装包自己的事。

[Messages]
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

[CustomMessages]
; 卸载时问"要不要连镜像一起删"（桌面版与 CLI 版共用；见各自 .iss 的 [Code]）
UninstallMirrorPrompt=是否同时删除源码镜像（安装目录下的 mirror 文件夹，可能数 GB）？选择「否」会保留镜像与安装目录内你的其它文件（下次安装可直接复用）；选择「是」会永久删除它们。TT 的配置数据（AppData 下的 tt 目录）不在删除范围内，任何选择都不会动它。
