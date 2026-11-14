@echo off
chcp 65001 >nul
rem Build the TT Desktop (Electron) installer + portable zip.
rem
rem NOTE: this file must stay ASCII-only. cmd.exe parses .bat bytes in the OEM
rem       codepage, so UTF-8 Chinese in a .bat breaks the parser (even after
rem       chcp 65001) -- lines get split mid-character and cmd tries to run the
rem       fragments. Chinese text belongs in README.md / docs/, not here.
rem
rem The payload is NOT defined here. It is exactly the portable package
rem (tt.exe + tzs/ + designer/ + skills/ + config.example.json + README.md), so
rem this script calls build_portable.bat and copies that stage -- build_msi.bat
rem does the same thing for the same reason: two file lists would drift, and the
rem one that drifts is always the one nobody rebuilds.
rem
rem Three deliberate differences from the portable copy:
rem   config.json  REMOVED -- same as the MSI: an empty config next to the exe
rem                looks like the real one and confuses "where do I edit".
rem   .portable    REMOVED -- that marker tells the CLI to keep its config inside
rem                the package. A desktop install must keep it in
rem                %APPDATA%\tt (shared with the CLI), and resources\ gets
rem                replaced on every update.
rem   .desktop     ADDED   -- tells `tt update` this binary ships with the desktop
rem                app and must not be self-replaced (internal/update/kind.go).
rem
rem Products land in dist\ next to the MSI and the portable zip:
rem   TT-Desktop-<version>-x64-Setup.exe   per-user installer (no admin)
rem   TT-Desktop-<version>-x64.zip         unzip-and-run
rem
rem Usage: build_desktop.bat [inno]
rem   inno  -> build the installer with Inno Setup instead of electron-builder's NSIS.
rem            Why this exists: NSIS installers cannot RUN on this build box (endpoint
rem            security blocks the NSIS runtime -- a 38KB blank NSIS installer exits 2
rem            without running a single line), and electron-builder must run the installer
rem            once while packaging to extract the uninstaller, so the NSIS target dies at
rem            its last step. Inno Setup does everything at compile time and runs fine.
rem            See desktop\installer\tt-desktop.iss for the full story.
setlocal
cd /d "%~dp0"

rem `inno` mode: skip the NSIS/zip targets, produce dist\TT-Desktop-<version>-x64-Setup.exe
set "INNO="
if /i "%~1"=="inno" set "INNO=1"

set STAGE=dist\tt-portable
set PAYLOAD=desktop\payload\tt
set RELDIR=desktop\release

rem Mirrors: electron (~110MB) and the NSIS/builder binaries come from GitHub by
rem default, which is unusable on many corporate networks. Same precedent as
rem GOPROXY=https://goproxy.cn in build_portable.bat -- change these two lines for
rem a local mirror instead of editing every script.
set ELECTRON_MIRROR=https://npmmirror.com/mirrors/electron/
set ELECTRON_BUILDER_BINARIES_MIRROR=https://npmmirror.com/mirrors/electron-builder-binaries/

rem Release version: read from the repo-root VERSION file, the ONE place a version
rem number is written (the other two .bat files read the same file; `make
rem version-check` also pins desktop/package.json to it).
set VERSION=
for /f "usebackq delims=" %%v in ("VERSION") do if not defined VERSION set VERSION=%%v
if not defined VERSION (echo MISSING VERSION file ^(repo root^) & exit /b 1)
powershell -NoProfile -Command "exit ([int]((Get-Content VERSION -Raw).Trim() -notmatch '^\d+\.\d+\.\d+$'))" || (echo BAD VERSION in VERSION file: %VERSION% -- expected MAJOR.MINOR.PATCH & exit /b 1)

echo [1/6] Building the payload ^(same as the portable package^) ...
rem Call by absolute path: with NoDefaultCurrentDirectoryInExePath set, a bare
rem file name is not resolved from the current directory.
call "%~dp0build_portable.bat"
if errorlevel 1 (echo PORTABLE BUILD FAILED & exit /b 1)
if not exist "%STAGE%\tt.exe" (echo MISSING staged tt.exe & exit /b 1)

echo [2/6] Staging the payload for Electron ...
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

echo [3/6] Installing desktop dependencies ...
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

echo [4/6] Running desktop checks ...
call npm test
if errorlevel 1 (echo DESKTOP CHECKS FAILED & popd & exit /b 1)

echo [5/6] Packaging ^(electron-builder^) ...
if exist release rmdir /s /q release
if defined INNO (
    rem --dir only: the app directory is all Inno Setup needs (it never touches NSIS)
    call npx electron-builder --win --x64 --dir
    if errorlevel 1 (echo ELECTRON-BUILDER ^(--dir^) FAILED & popd & exit /b 1)
    call :run_inno
    if errorlevel 1 (popd & exit /b 1)
) else (
    call npm run dist
    if errorlevel 1 (echo ELECTRON-BUILDER FAILED & popd & exit /b 1)
)
popd

echo [6/6] Copying artifacts to dist\ ...
if not exist "dist" mkdir "dist" || exit /b 1
if defined INNO (
    if not exist "dist\TT-Desktop-%VERSION%-x64-Setup.exe" (echo MISSING Inno Setup output in dist\ & exit /b 1)
    echo [done] dist\TT-Desktop-%VERSION%-x64-Setup.exe
) else (
    copy /y "%RELDIR%\TT-Desktop-%VERSION%-x64-Setup.exe" "dist\" >nul || (echo MISSING Setup.exe in %RELDIR% & exit /b 1)
    copy /y "%RELDIR%\TT-Desktop-%VERSION%-x64.zip" "dist\" >nul || (echo MISSING portable zip in %RELDIR% & exit /b 1)
    echo [done] dist\TT-Desktop-%VERSION%-x64-Setup.exe / .zip
)
dir /b dist
endlocal
exit /b 0

rem ---- Inno Setup step (runs inside desktop\) ----
rem NOTE: no parenthesised blocks in here. `%ISCC%` expands to a path containing
rem "(x86)"; inside an if-block that `)` would close the block early and cmd dies with
rem "\Inno was unexpected at this time" (hit once, keeping it flat).
:run_inno
set "ISCC=%ProgramFiles(x86)%\Inno Setup 6\ISCC.exe"
if exist "%ISCC%" goto :inno_compile
echo MISSING Inno Setup 6 ^(ISCC.exe^) under Program Files ^(x86^).
echo   install it from https://jrsoftware.org/isdl.php, then re-run: build_desktop.bat inno
exit /b 1

:inno_compile
echo Compiling installer with Inno Setup ...
"%ISCC%" /Qp /DAppVersion=%VERSION% "installer\tt-desktop.iss"
if errorlevel 1 goto :inno_failed
exit /b 0

:inno_failed
echo INNO SETUP COMPILE FAILED
exit /b 1
