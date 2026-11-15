@echo off
chcp 65001 >nul
rem Build TT Desktop (Electron, x64): Inno Setup installer + portable zip.
rem
rem Usage: build_desktop.bat [reuse]
rem   reuse   do not rebuild the portable payload stage (build_dist.bat uses it)
rem
rem The payload is NOT defined here. It is exactly the portable package
rem (tt.exe + tzs/ + designer/ + skills/ + config.example.json + README.md), so
rem this script calls build_portable.bat and copies that stage -- build_cli_setup.bat
rem does the same thing for the same reason: two file lists would drift, and the
rem one that drifts is always the one nobody rebuilds.
rem
rem Three deliberate differences from the portable copy:
rem   config.json  REMOVED -- an empty config next to the exe looks like the real
rem                one and confuses "where do I edit".
rem   .portable    REMOVED -- that marker tells the CLI to keep its config inside
rem                the package. A desktop install must keep it in
rem                %APPDATA%\tt (shared with the CLI), and resources\ gets
rem                replaced on every update.
rem   .desktop     ADDED   -- tells `tt update` this binary ships with the desktop
rem                app and must not be self-replaced (internal/update/kind.go).
rem
rem Products land in dist\:
rem   TT-Desktop-<version>-x64-Setup.exe   per-user installer (no admin)
rem   TT-Desktop-<version>-x64.zip         unzip-and-run
rem
rem Installer: Inno Setup only. NSIS is gone for good -- NSIS installers cannot
rem RUN on this build box (endpoint security blocks the NSIS runtime), and
rem electron-builder has to run the installer once while packaging to extract
rem the uninstaller, so the NSIS target died at its last step. Inno Setup does
rem everything at compile time and runs fine. See desktop\installer\tt-desktop.iss.
rem
rem NOTE: this file must stay ASCII-only. cmd.exe parses .bat bytes in the OEM
rem       codepage, so UTF-8 Chinese in a .bat breaks the parser (even after
rem       chcp 65001) -- lines get split mid-character and cmd tries to run the
rem       fragments. Chinese text belongs in README.md / docs/, not here.
setlocal
cd /d "%~dp0"

set REUSE=%~1
set STAGE=.tmp-build\tt-portable
set PAYLOAD=desktop\payload\tt
set RELDIR=desktop\release
set UNPACKED=%RELDIR%\win-unpacked

rem Mirrors: electron (~110MB) and electron-builder's helper binaries come from
rem GitHub by default, which is unusable on many corporate networks. Same
rem precedent as GOPROXY=https://goproxy.cn in build_portable.bat -- change these
rem two lines for a local mirror instead of editing every script.
set ELECTRON_MIRROR=https://npmmirror.com/mirrors/electron/
set ELECTRON_BUILDER_BINARIES_MIRROR=https://npmmirror.com/mirrors/electron-builder-binaries/

rem Release version: read from the repo-root VERSION file, the ONE place a version
rem number is written (the other build scripts read the same file; `make
rem version-check` also pins desktop/package.json to it).
set VERSION=
for /f "usebackq delims=" %%v in ("VERSION") do if not defined VERSION set VERSION=%%v
if not defined VERSION (echo MISSING VERSION file ^(repo root^) & exit /b 1)
powershell -NoProfile -Command "exit ([int]((Get-Content VERSION -Raw).Trim() -notmatch '^\d+\.\d+\.\d+$'))" || (echo BAD VERSION in VERSION file: %VERSION% -- expected MAJOR.MINOR.PATCH & exit /b 1)

echo [1/7] Building the payload ^(same as the portable package^) ...
rem Call by absolute path: with NoDefaultCurrentDirectoryInExePath set, a bare
rem file name is not resolved from the current directory.
call "%~dp0build_portable.bat" %REUSE%
if errorlevel 1 (echo PORTABLE BUILD FAILED & exit /b 1)
if not exist "%STAGE%\tt.exe" (echo MISSING staged tt.exe & exit /b 1)

echo [2/7] Staging the payload for Electron ...
if exist "%PAYLOAD%" rmdir /s /q "%PAYLOAD%"
mkdir "%PAYLOAD%" || exit /b 1
xcopy /e /i /y /q "%STAGE%" "%PAYLOAD%" >nul || (echo PAYLOAD COPY FAILED & exit /b 1)
if exist "%PAYLOAD%\config.json" del /q "%PAYLOAD%\config.json"
if exist "%PAYLOAD%\.portable" del /q "%PAYLOAD%\.portable"
type nul > "%PAYLOAD%\.desktop"
if not exist "%PAYLOAD%\tt.exe" (echo MISSING payload tt.exe & exit /b 1)
if not exist "%PAYLOAD%\config.example.json" (echo MISSING payload config.example.json & exit /b 1)
if exist "%PAYLOAD%\config.json" (echo payload MUST NOT carry config.json & exit /b 1)
if exist "%PAYLOAD%\.portable" (echo payload MUST NOT carry .portable & exit /b 1)

echo [3/7] Installing desktop dependencies ...
pushd desktop
if not exist package-lock.json (
    call npm install --no-audit --no-fund
    if errorlevel 1 (echo NPM INSTALL FAILED & popd & exit /b 1)
) else (
    call npm ci --no-audit --no-fund
    if errorlevel 1 (echo NPM CI FAILED & popd & exit /b 1)
)
rem electron ships no postinstall script any more (45+): the runtime is fetched
rem lazily, so pull it explicitly here. A package run that silently raced a lazy
rem download would produce an installer without a runtime.
node node_modules\electron\install.js
if errorlevel 1 (echo ELECTRON RUNTIME DOWNLOAD FAILED & popd & exit /b 1)
if not exist node_modules\electron\dist\electron.exe (echo MISSING electron.exe after install & popd & exit /b 1)

echo [4/7] Running desktop checks ...
call npm test
if errorlevel 1 (echo DESKTOP CHECKS FAILED & popd & exit /b 1)

echo [5/7] Packaging the app directory ^(electron-builder --dir^) ...
if exist release rmdir /s /q release
rem --dir only: the app directory is all Inno Setup needs (it never touches NSIS).
call npx electron-builder --win --x64 --dir
if errorlevel 1 (echo ELECTRON-BUILDER ^(--dir^) FAILED & popd & exit /b 1)
popd
if not exist "%UNPACKED%\TT Desktop.exe" (echo MISSING %UNPACKED%\TT Desktop.exe & exit /b 1)

echo [6/7] Compiling the installer ^(Inno Setup^) ...
rem NOTE: no parenthesised blocks in here. The default path contains "(x86)";
rem inside an if-block that ")" closes the block early and cmd dies with
rem "\Inno was unexpected at this time" (hit once, keeping it flat).
set "ISCC=%ProgramFiles(x86)%\Inno Setup 6\ISCC.exe"
if not "%INNO%"=="" set "ISCC=%INNO%"
if exist "%ISCC%" goto :compile
echo MISSING Inno Setup 6 ^(ISCC.exe^) under Program Files ^(x86^).
echo   install it from https://jrsoftware.org/isdl.php, or set INNO to ISCC.exe
exit /b 1

:compile
if not exist "dist" mkdir "dist"
"%ISCC%" /Qp /DAppVersion=%VERSION% "desktop\installer\tt-desktop.iss"
if errorlevel 1 (echo INNO SETUP COMPILE FAILED & exit /b 1)

echo [7/7] Packing the portable zip ...
rem Flat zip: the archive IS the app directory (no extra wrapper folder), which
rem is what the old electron-builder zip target produced -- users unzip and run
rem "TT Desktop.exe" directly.
go run ./tools/zip -flat "%UNPACKED%" "dist\TT-Desktop-%VERSION%-x64.zip"
if errorlevel 1 (echo DESKTOP ZIP FAILED & exit /b 1)

if not exist "dist\TT-Desktop-%VERSION%-x64-Setup.exe" (echo MISSING desktop Setup.exe & exit /b 1)
if not exist "dist\TT-Desktop-%VERSION%-x64.zip" (echo MISSING desktop zip & exit /b 1)
echo [done] dist\TT-Desktop-%VERSION%-x64-Setup.exe / dist\TT-Desktop-%VERSION%-x64.zip
endlocal
exit /b 0
