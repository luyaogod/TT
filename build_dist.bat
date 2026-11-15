@echo off
chcp 65001 >nul
rem Build the full TT release matrix: 4 artifacts in dist\ (x64).
rem
rem                    install (Inno)                  portable (zip)
rem   CLI              TT-<v>-x64-Setup.exe            tt-<v>-x64-portable.zip
rem   Desktop          TT-Desktop-<v>-x64-Setup.exe    TT-Desktop-<v>-x64.zip
rem
rem The payload is built once: the portable stage is reused by the installer and
rem the desktop build (that is what the "reuse" argument means).
rem
rem Usage: build_dist.bat
rem
rem NOTE: this file must stay ASCII-only. cmd.exe parses .bat bytes in the OEM
rem       codepage, so UTF-8 Chinese in a .bat breaks the parser (even after
rem       chcp 65001) -- lines get split mid-character and cmd tries to run the
rem       fragments. Chinese text belongs in README.md / docs/, not here.
setlocal
cd /d "%~dp0"

set VERSION=
for /f "usebackq delims=" %%v in ("VERSION") do if not defined VERSION set VERSION=%%v
if not defined VERSION (echo MISSING VERSION file ^(repo root^) & exit /b 1)
powershell -NoProfile -Command "exit ([int]((Get-Content VERSION -Raw).Trim() -notmatch '^\d+\.\d+\.\d+$'))" || (echo BAD VERSION in VERSION file: %VERSION% -- expected MAJOR.MINOR.PATCH & exit /b 1)

call "%~dp0build_portable.bat"
if errorlevel 1 (echo PORTABLE FAILED & exit /b 1)
call "%~dp0build_cli_setup.bat" reuse
if errorlevel 1 (echo CLI SETUP FAILED & exit /b 1)
call "%~dp0build_desktop.bat" reuse
if errorlevel 1 (echo DESKTOP FAILED & exit /b 1)

echo.
echo ===================== verifying 4 artifacts =====================
set MISSING=0
for %%f in (
    "dist\TT-%VERSION%-x64-Setup.exe"
    "dist\tt-%VERSION%-x64-portable.zip"
    "dist\TT-Desktop-%VERSION%-x64-Setup.exe"
    "dist\TT-Desktop-%VERSION%-x64.zip"
) do (
    if exist %%f (
        for %%s in (%%f) do echo   [ok] %%~nxs  %%~zs bytes
    ) else (
        echo   [MISSING] %%f
        set MISSING=1
    )
)
if "%MISSING%"=="1" (echo SOME ARTIFACTS ARE MISSING & exit /b 1)
echo.
echo done: 4 artifacts in dist\
endlocal
exit /b 0
