@echo off

REM Get machine architecture
set "ARCH_VALUE=unsupported"

wmic os get osarchitecture | findstr "64" >nul 2>nul
if %errorlevel% equ 0 (
  set "ARCH_VALUE=amd64"
)

wmic os get osarchitecture | findstr "ARM64" >nul 2>nul
if %errorlevel% equ 0 (
  set "ARCH_VALUE=arm64"
)

if "%ARCH_VALUE%"=="unsupported" (
  echo Unsupported architecture: %ARCH_VALUE%
  exit /b 1
)

REM Get machine operating system
set "OS=%OS%"

if "%OS%"=="Windows_NT" (
  set "OUTPUT_FOLDER=%USERPROFILE%\AppData\Roaming\daytona\server\binaries\v0.0.0-dev"
) else (
  echo Unsupported operating system: %OS%
  exit /b 1
)

REM Create output folder if it doesn't exist
mkdir "%OUTPUT_FOLDER%" 2>nul
if not exist "%OUTPUT_FOLDER%" (
  echo Failed to create output folder
  exit /b 1
)

REM Build the project container binary
set "GOOS=linux"
set "GOARCH=%ARCH_VALUE%"
go build -o "%OUTPUT_FOLDER%\daytona-linux-%ARCH_VALUE%" cmd\daytona\main.go
if %errorlevel% equ 1 (
  echo Build failed
  exit /b 1
)

echo Binary build successful