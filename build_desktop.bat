@echo off
chcp 65001 >nul
rem Build the TT Windows desktop packages (Electron shell + Go backend):
rem   web/dist -> tt.exe (go:embed) -> electron-builder (NSIS installer)
rem   -> green/portable: rename the unpacked dir, drop the .portable marker + readme, zip it.
rem NOTE: keep this file ASCII-only. cmd.exe parses .bat bytes in the OEM codepage,
rem       so UTF-8 Chinese in a .bat breaks the parser (Chinese text lives in
rem       desktop/build/*.txt and desktop/README.md instead).
rem
rem 合并说明：合并前只有 TDebug 有桌面外壳（TDev 与 TDictCli 都是纯 CLI）。
rem 现在外壳拉起的是统一服务 `tt.exe serve --desktop`，窗口落在 /debug/，
rem 界面里的导航可以跳到 /dict/ —— 一个外壳同时覆盖两套页面。
setlocal
cd /d "%~dp0"

set STAGE=dist\desktop
set GOPROXY=https://goproxy.cn,direct
rem Mirrors for China networks (electron binary + electron-builder helper binaries)
if "%ELECTRON_MIRROR%"=="" set ELECTRON_MIRROR=https://npmmirror.com/mirrors/electron/
if "%ELECTRON_BUILDER_BINARIES_MIRROR%"=="" set ELECTRON_BUILDER_BINARIES_MIRROR=https://npmmirror.com/mirrors/electron-builder-binaries/

rem Version comes from desktop/package.json (single source of truth).
rem NOTE: no pipe inside the for-command (cmd would need caret escaping there).
set VER=
for /f "delims=" %%v in ('powershell -NoProfile -Command "(ConvertFrom-Json (Get-Content -Raw desktop\package.json)).version"') do set VER=%%v
if "%VER%"=="" set VER=0.1.0
echo   version: %VER%
set GREEN=%STAGE%\TT-%VER%-portable

if exist "%STAGE%" rmdir /s /q "%STAGE%"

echo [1/6] Building web frontend ...
if not exist web\dist\debug\index.html (
    pushd web
    call npm install
    if errorlevel 1 (echo NPM INSTALL FAILED & popd & exit /b 1)
    call npm run build
    if errorlevel 1 (echo WEB BUILD FAILED & popd & exit /b 1)
    popd
) else (
    echo   web\dist exists, skip
)

echo [2/6] Building tt.exe (with embedded web/dist) ...
call go build -trimpath -o tt.exe .
if errorlevel 1 (
    echo BUILD FAILED
    exit /b 1
)

echo [3/6] Desktop dependencies ...
if not exist desktop\node_modules (
    echo   installing (ELECTRON_MIRROR=%ELECTRON_MIRROR%)
    pushd desktop
    call npm install
    if errorlevel 1 (echo NPM INSTALL FAILED & popd & exit /b 1)
    popd
) else (
    echo   desktop\node_modules exists, skip
)
rem Some npm policies block dependency install scripts, so the electron binary
rem may be missing: fetch it explicitly.
if not exist desktop\node_modules\electron\dist\electron.exe (
    echo   electron binary missing, downloading ...
    pushd desktop
    call node node_modules\electron\install.js
    if errorlevel 1 (echo ELECTRON DOWNLOAD FAILED & popd & exit /b 1)
    popd
)

echo [4/6] Generating icon ...
pushd desktop
call npm run icon
if errorlevel 1 (echo ICON FAILED & popd & exit /b 1)
popd

echo [5/6] Packaging installer (electron-builder: NSIS) ...
pushd desktop
call npm run dist
if errorlevel 1 (echo PACKAGING FAILED & popd & exit /b 1)
popd

echo [6/6] Packing green/portable zip ...
rem The unpacked dir becomes the root folder inside the zip; the .portable marker
rem tells the app to keep its data (config/logs/breakpoints) next to the exe.
if exist "%GREEN%" rmdir /s /q "%GREEN%"
move /y "%STAGE%\win-unpacked" "%GREEN%" >nul
if errorlevel 1 (echo RENAME FAILED & exit /b 1)
copy /y desktop\build\.portable "%GREEN%\.portable" >nul
copy /y desktop\build\README-portable.txt "%GREEN%\README-portable.txt" >nul
if errorlevel 1 (echo COPY MARKER FAILED & exit /b 1)
powershell -NoProfile -Command "Compress-Archive -Path '%GREEN%' -DestinationPath '%STAGE%\TT-%VER%-portable.zip' -Force"
if errorlevel 1 (echo ZIP FAILED & exit /b 1)

echo Done: %STAGE%
dir /b "%STAGE%"
endlocal
