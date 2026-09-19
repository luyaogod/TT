@echo off
chcp 65001 >nul
rem Build the TT Windows portable package: frontend + exe + empty config + example
rem + README + skills, zipped.
rem
rem NOTE: this file must stay ASCII-only. cmd.exe parses .bat bytes in the OEM
rem       codepage, so UTF-8 Chinese in a .bat breaks the parser (even after
rem       chcp 65001) -- lines get split mid-character and cmd tries to run the
rem       fragments. Chinese text belongs in README.md / docs/, not here.
rem       The merged project's three source scripts carried the same warning.
rem
rem NOTE: never package this machine's config.json (it holds real SSH/DB
rem       credentials). The portable package ships config.empty.json instead;
rem       users configure it via `tt serve` or by editing config.json.
rem NOTE: never package erp_data.db either (it holds the customer's table
rem       dictionary / schema / enterprise codes); users run `tt dict db sync`
rem       after configuring an environment.
rem NOTE: skills/ ships as a plain directory (not embedded in the binary) so
rem       users can edit it; `tt install skills` copies it to the target dir.
rem       After the merge there are four skills in there, installed together.
setlocal
cd /d "%~dp0"

set STAGE=dist\tt-portable
set GOPROXY=https://goproxy.cn,direct
rem Release version, injected into the binary via -ldflags (shown by `tt version`).
set VERSION=0.1.0

if exist dist rmdir /s /q dist
mkdir "%STAGE%"

echo [1/5] Building frontend (web/dist) ...
pushd web
if not exist node_modules (
    call npm install
    if errorlevel 1 (
        echo NPM INSTALL FAILED
        popd
        exit /b 1
    )
)
call npm run build
if errorlevel 1 (
    echo FRONTEND BUILD FAILED
    popd
    exit /b 1
)
popd
rem main.go embeds web/dist, so the frontend must exist before go build.
if not exist "web\dist\index.html" (echo MISSING web\dist\index.html & exit /b 1)

echo [2/5] Building tt.exe (v%VERSION%) ...
call go build -trimpath -ldflags "-X tt/internal/cli.Version=%VERSION%" -o tt.exe .
if errorlevel 1 (
    echo BUILD FAILED
    exit /b 1
)

echo [3/5] Staging EMPTY config.json / config.example.json / README.md / skills ...
copy /y config.empty.json "%STAGE%\config.json" >nul || (echo COPY empty config FAILED & exit /b 1)
copy /y config.example.json "%STAGE%\" >nul || (echo COPY config.example.json FAILED & exit /b 1)
copy /y README.md "%STAGE%\" >nul || (echo COPY README.md FAILED & exit /b 1)
xcopy /e /i /y /q skills "%STAGE%\skills" >nul || (echo COPY skills FAILED & exit /b 1)
if exist "%STAGE%\tt.exe" del /q "%STAGE%\tt.exe"
copy /y tt.exe "%STAGE%\" >nul || (echo COPY tt.exe FAILED & exit /b 1)
rem Portable marker: the CLI keeps its config inside the package instead of
rem writing to the user directory (see internal/config/paths.go, IsPortable).
type nul > "%STAGE%\.portable"

rem [3b/5] Staging the .tzs engine (C#). It is built SEPARATELY -- see engine\BUILD.md.
rem   Exactly four files, listed by name on purpose: engine\out\ also holds a dozen probe
rem   programs (Probe / Edit / AddField / RoundTrip / Test* / E2E) that must NOT ship, so an
rem   xcopy of the whole directory would put them all in the portable package.
rem   Not built here: the engine needs csc + a bash script, and rebuilding it changes its MVID,
rem   which orphans every daemon started from the previous build (engine\BUILD.md explains).
set TZSENGINE=engine\out
for %%f in (tzs-server.exe tzs-cli.exe TzsCli.dll TzsCli.Designer.dll) do (
    if not exist "%TZSENGINE%\%%f" (
        echo MISSING %TZSENGINE%\%%f -- build the engine first, see engine\BUILD.md
        exit /b 1
    )
)
if not exist "%STAGE%\tzs" mkdir "%STAGE%\tzs"
for %%f in (tzs-server.exe tzs-cli.exe TzsCli.dll TzsCli.Designer.dll) do (
    copy /y "%TZSENGINE%\%%f" "%STAGE%\tzs\" >nul || (echo COPY engine %%f FAILED & exit /b 1)
)

echo [4/5] Packing zip ...
rem Recursive zip (includes the skills/ subtree); shutil.make_archive keeps the
rem tt-portable/ top-level directory inside the archive.
python tools\zip.py >nul 2>nul
if errorlevel 1 (
    python -c "import shutil; shutil.make_archive('dist/tt-portable','zip','dist','tt-portable')" >nul 2>nul
)
if errorlevel 1 (
    echo   python unavailable, falling back to PowerShell ...
    powershell -NoProfile -Command "Compress-Archive -Path '%STAGE%' -DestinationPath 'dist\tt-portable.zip' -Force" >nul 2>nul
    if errorlevel 1 (
        echo ZIP FAILED
        exit /b 1
    )
)

echo [5/5] Done: dist\tt-portable.zip
dir /b dist
endlocal
