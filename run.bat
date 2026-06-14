@echo off
title DBM-Lite - 启动脚本

echo ========================================
echo   DBM-Lite 启动脚本
echo   正在通过 PowerShell 启动服务...
echo ========================================
echo.

cd /d "%~dp0"
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0run.ps1"

if errorlevel 1 (
    echo.
    echo [ERROR] 启动失败，请查看上方错误信息
    echo.
    pause
    exit /b 1
)

endlocal
