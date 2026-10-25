@echo off
rem Argus launcher for Windows.
rem Usage: start.bat or start.bat -addr :9000

setlocal enabledelayedexpansion
cd /d "%~dp0"

set "BIN=argus.exe"
if not exist "%BIN%" (
	echo [Argus] executable not found: %BIN% 1>&2
	exit /b 1
)

set "RESTART_CODE=75"
set "MAX_DELAY=60"
set /a delay=1

:loop
"%BIN%" %*
set "code=!ERRORLEVEL!"

if "!code!"=="0" (
	echo [Argus] exited normally
	exit /b 0
)

if "!code!"=="%RESTART_CODE%" (
	echo [Argus] restart requested; applying the new version
	set /a delay=1
	goto loop
)

echo [Argus] Startup failed (code=!code!). 1>&2
echo See: %LOCALAPPDATA%\Argus\startup.log
pause
exit /b !code!
