@echo off
chcp 65001 >nul
rem Build the TT CLI installer (Inno Setup, x64).
rem
rem Usage: build_cli_setup.bat [reuse]
rem   reuse   do not rebuild the portable stage (build_dist.bat uses it)
rem
rem The payload is the portable package minus two things, for the same reasons
rem the old MSI had:
rem   config.json  REMOVED -- an empty config next to the exe looks like the real
rem                one and confuses "where do I edit". Installed TT keeps its
rem                config in %APPDATA%\tt (see internal/config/paths.go).
rem   .portable    REMOVED -- that marker tells the CLI to keep its config inside
rem                the package. An installed copy must not be portable: the
rem                install directory is replaced on every upgrade and is not
rem                where user data belongs.
rem
rem Requires Inno Setup 6 (ISCC.exe). Point INNO at it to use a non-default
rem install, e.g. set INNO=D:\tools\Inno\ISCC.exe
rem
rem NOTE: this file must stay ASCII-only. cmd.exe parses .bat bytes in the OEM
rem       codepage, so UTF-8 Chinese in a .bat breaks the parser (even after
rem       chcp 65001) -- lines get split mid-character and cmd tries to run the
rem       fragments. Chinese text belongs in README.md / docs/, not here.
setlocal
cd /d "%~dp0"

rem Release version: read from the repo-root VERSION file, the ONE place a version
rem number is written (the other build scripts read the same file; `make
rem version-check` also pins it against web/ and desktop/package.json).
set VERSION=
for /f "usebackq delims=" %%v in ("VERSION") do if not defined VERSION set VERSION=%%v
if not defined VERSION (echo MISSING VERSION file ^(repo root^) & exit /b 1)
powershell -NoProfile -Command "exit ([int]((Get-Content VERSION -Raw).Trim() -notmatch '^\d+\.\d+\.\d+$'))" || (echo BAD VERSION in VERSION file: %VERSION% -- expected MAJOR.MINOR.PATCH & exit /b 1)

set REUSE=%~1
set STAGE=.tmp-build\tt-portable
set SETUPSTAGE=.tmp-build\tt-setup
set OUT=dist\TT-%VERSION%-x64-Setup.exe

echo [1/4] Building the payload ^(same as the portable package^) ...
rem Call by absolute path: with NoDefaultCurrentDirectoryInExePath set, a bare
rem file name is not resolved from the current directory.
call "%~dp0build_portable.bat" %REUSE%
if errorlevel 1 (echo PORTABLE BUILD FAILED & exit /b 1)

echo [2/4] Staging the install payload ...
if exist "%SETUPSTAGE%" rmdir /s /q "%SETUPSTAGE%"
mkdir "%SETUPSTAGE%" || exit /b 1
xcopy /e /i /y /q "%STAGE%" "%SETUPSTAGE%" >nul || (echo STAGE COPY FAILED & exit /b 1)
if exist "%SETUPSTAGE%\.portable" del /q "%SETUPSTAGE%\.portable"
if exist "%SETUPSTAGE%\config.json" del /q "%SETUPSTAGE%\config.json"
if not exist "%SETUPSTAGE%\tt.exe" (echo MISSING staged tt.exe & exit /b 1)
if not exist "%SETUPSTAGE%\config.example.json" (echo MISSING staged config.example.json & exit /b 1)
if exist "%SETUPSTAGE%\.portable" (echo install payload MUST NOT be portable & exit /b 1)
if exist "%SETUPSTAGE%\config.json" (echo install payload MUST NOT carry config.json & exit /b 1)

echo [3/4] Compiling the installer ^(Inno Setup^) ...
rem NOTE: no parenthesised blocks around the ISCC call. The default path
rem contains "(x86)"; inside an if-block that ")" closes the block early and cmd
rem dies with "\Inno was unexpected at this time" (hit once, keeping it flat).
set "ISCC=%ProgramFiles(x86)%\Inno Setup 6\ISCC.exe"
if not "%INNO%"=="" set "ISCC=%INNO%"
if exist "%ISCC%" goto :compile
echo MISSING Inno Setup 6 ^(ISCC.exe^) under Program Files ^(x86^).
echo   install it from https://jrsoftware.org/isdl.php, or set INNO to ISCC.exe
exit /b 1

:compile
if not exist "dist" mkdir "dist"
"%ISCC%" /Qp /DAppVersion=%VERSION% "installer\tt-cli.iss"
if errorlevel 1 (echo INNO SETUP COMPILE FAILED & exit /b 1)

echo [4/4] Done: %OUT%
if not exist "%OUT%" (echo MISSING %OUT% & exit /b 1)
endlocal
exit /b 0
