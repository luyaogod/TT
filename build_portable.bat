@echo off
chcp 65001 >nul
rem 构建 TT 的 Windows 便携包：前端 + exe + 空配置 + 示例配置 + README + skills，打包成 zip。
rem
rem NOTE: 绝不打包本机 config.json（含真实 SSH/数据库凭据）；便携版落地的是 config.empty.json，
rem       用户首次用 `tt serve` 或手改自行配置。
rem NOTE: 也不打包 erp_data.db（含客户表字典/schema/企业码等数据）；用户配好环境后自行
rem       `tt dict db sync` 拉取本地字典库。
rem NOTE: skills/ 以普通目录随包分发（不内嵌二进制），用户可直接编辑；
rem       `tt install skills` 把它复制到当前目录。合并后这里是四套技能一起装。
rem
rem fmt: 合并前三个项目各有一份同形脚本（注释里写着"与另外两个一致，改动请三边同步"），
rem      合并后只有这一份；唯一新增的是第 1 步的前端构建 —— main.go 的
rem      //go:embed all:web/dist 要求产物先存在，否则 go build 会把空目录嵌进去。
setlocal
cd /d "%~dp0"

set STAGE=dist\tt-portable
set GOPROXY=https://goproxy.cn,direct
rem 发布版本号：由 -ldflags 注入二进制（`tt version` 显示）；发新版改这一行
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
if not exist "web\dist\debug\index.html" (echo MISSING web\dist\debug\index.html & exit /b 1)
if not exist "web\dist\dict\index.html"  (echo MISSING web\dist\dict\index.html  & exit /b 1)

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
rem 便携标记：CLI 据此把配置留在包内而不是写用户目录（见 internal/config/paths.go 的 IsPortable）
type nul > "%STAGE%\.portable"

echo [4/5] Packing zip ...
rem 递归打包（含 skills/ 子目录）；shutil.make_archive 会保留 tt-portable/ 顶层目录
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
