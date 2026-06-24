@echo off
title DBM-Lite Startup
setlocal enabledelayedexpansion
set "ROOT=%~dp0"

echo ============================================================
echo   DBM-Lite Project Build and Run
echo   Project Path: %ROOT%
echo ============================================================

echo.
echo [1/3] Cleaning ports 8080, 3000 ...
for %%P in (8080 3000) do (
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%%P " ^| findstr LISTENING') do (
        echo   - Killing process on port %%P, PID: %%a
        taskkill /F /PID %%a >nul 2>&1
    )
)

timeout /t 1 /nobreak >nul

echo.
echo [2/3] Building backend (Go) ...
cd /d "%ROOT%backend"
if exist dbm-lite.exe del /Q dbm-lite.exe
echo   Compiling dbm-lite.exe ...
go build -o dbm-lite.exe ./cmd/server
if errorlevel 1 (
    echo   ERROR: Backend build failed!
    pause
    exit /b 1
)
echo   Starting backend service on port 8080 ...
start "DBM-Backend-8080" cmd /k "cd /d %ROOT%backend && dbm-lite.exe"

echo.
echo [3/3] Waiting for backend to be ready ...
set "retry_count=0"
set "max_retries=30"
:check_backend
timeout /t 2 /nobreak >nul
set /a retry_count+=1
curl -s http://localhost:8080/api/health >nul 2>&1
if errorlevel 1 (
    if !retry_count! lss %max_retries% (
        echo   Waiting for backend... [!retry_count!/%max_retries%]
        goto check_backend
    ) else (
        echo   ERROR: Backend service failed to start!
        pause
        exit /b 1
    )
)
echo   Backend service is ready!

echo.
echo [4/4] Starting frontend dev server ...
cd /d "%ROOT%frontend"
echo   Starting frontend dev server on port 3000 ...
start "DBM-Frontend-3000" cmd /k "cd /d %ROOT%frontend && yarn dev"

echo.
echo ============================================================
echo   Waiting for frontend to be ready (5 seconds) ...
echo ============================================================
timeout /t 5 /nobreak >nul

echo.
echo   Opening browser: http://localhost:3000
start http://localhost:3000

echo.
echo ============================================================
echo   All services started successfully!
echo   - Frontend: http://localhost:3000
echo   - Backend API: http://localhost:8080
echo   - Login: admin (password in .env config)
echo ============================================================
endlocal
pause