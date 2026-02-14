@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM VSCode Plugin Build Script
REM Build TypeScript and package to .vsix file

set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%.."
set "RELEASE_DIR=%PROJECT_ROOT%\release"
set "VSIX_NAME=aiflow-vscode.vsix"

echo [INFO] === Start VSCode Plugin Build ===

REM Check node_modules
if not exist "%PROJECT_ROOT%\node_modules" (
    echo [INFO] Installing dependencies...
    cd /d "%PROJECT_ROOT%"
    call yarn install
    if errorlevel 1 (
        echo [ERROR] Failed to install dependencies
        exit /b 1
    )
    echo [INFO] Dependencies installed
)

REM Compile TypeScript
echo [INFO] Compiling TypeScript...
cd /d "%PROJECT_ROOT%"
call yarn compile
if errorlevel 1 (
    echo [ERROR] TypeScript compilation failed
    exit /b 1
)
echo [INFO] TypeScript compiled

REM Prepare release directory
echo [INFO] Preparing release directory: %RELEASE_DIR%
if not exist "%RELEASE_DIR%" (
    mkdir "%RELEASE_DIR%"
    echo [INFO] Created release directory
)

REM Delete existing package
set "OUTPUT_PATH=%RELEASE_DIR%\%VSIX_NAME%"
if exist "%OUTPUT_PATH%" (
    echo [INFO] Deleting existing package
    del /f "%OUTPUT_PATH%"
)

REM Package VSCode extension
echo [INFO] Packaging VSCode plugin...
cd /d "%PROJECT_ROOT%"
call npx vsce package --no-yarn --out "%OUTPUT_PATH%"
if errorlevel 1 (
    echo [ERROR] Failed to package VSCode plugin
    exit /b 1
)

echo [INFO] VSCode plugin packaged successfully
echo [INFO] Output: %OUTPUT_PATH%
echo [INFO] === Build Complete ===

endlocal
