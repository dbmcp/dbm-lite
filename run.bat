@echo off
title DBM-Lite Startup
setlocal enabledelayedexpansion
set "ROOT=%~dp0"
set "FRONTEND_DIR=%ROOT%frontend"

echo ============================================================
echo   DBM-Lite Project Build and Run
echo   Project Path: %ROOT%
echo ============================================================

echo.
echo [1/6] Cleaning ports 8080, 3000 ...
for %%P in (8080 3000) do (
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%%P " ^| findstr LISTENING') do (
        echo   - Killing process on port %%P, PID: %%a
        taskkill /F /PID %%a >nul 2>&1
    )
)

timeout /t 1 /nobreak >nul

echo.
echo [2/6] Building backend (Go) ...
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
echo [3/6] Waiting for backend to be ready ...
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
echo [4/6] Installing frontend dependencies ...
cd /d "%FRONTEND_DIR%"
echo   Installing npm dependencies ...
call yarn install --frozen-lockfile
if errorlevel 1 (
    echo   WARNING: yarn install failed, trying without --frozen-lockfile ...
    call yarn install
    if errorlevel 1 (
        echo   ERROR: Frontend dependencies installation failed!
        pause
        exit /b 1
    )
)
echo   Frontend dependencies installed successfully!

echo.
echo [5/6] Building frontend (Vue) ...
set "build_attempt=1"
set "max_build_attempts=3"
:build_frontend
echo   Attempt !build_attempt!/%max_build_attempts%: Building frontend ...
call yarn build
if errorlevel 1 (
    echo   WARNING: Frontend build failed on attempt !build_attempt!
    if !build_attempt! lss %max_build_attempts% (
        set /a build_attempt+=1
        echo   Retrying build...
        timeout /t 2 /nobreak >nul
        goto build_frontend
    ) else (
        echo   ERROR: Frontend build failed after %max_build_attempts% attempts!
        pause
        exit /b 1
    )
)
echo   Frontend build successful!

echo.
echo [6/6] Starting frontend dev server ...
echo   Starting frontend dev server on port 3000 ...
start "DBM-Frontend-3000" cmd /k "cd /d %FRONTEND_DIR% && yarn dev"

echo.
echo ============================================================
echo   Waiting for frontend to be ready ...
echo ============================================================
set "frontend_retry=0"
set "frontend_max_retries=30"
:check_frontend
timeout /t 2 /nobreak >nul
set /a frontend_retry+=1
curl -s http://localhost:3000 >nul 2>&1
if errorlevel 1 (
    if !frontend_retry! lss %frontend_max_retries% (
        echo   Waiting for frontend... [!frontend_retry!/%frontend_max_retries%]
        goto check_frontend
    ) else (
        echo   WARNING: Frontend not responding, attempting rebuild...
        for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3000 " ^| findstr LISTENING') do (
            taskkill /F /PID %%a >nul 2>&1
        )
        timeout /t 1 /nobreak >nul
        cd /d "%FRONTEND_DIR%"
        echo   Rebuilding frontend...
        call yarn build
        if errorlevel 1 (
            echo   ERROR: Frontend rebuild failed!
            pause
            exit /b 1
        )
        echo   Restarting frontend dev server...
        start "DBM-Frontend-3000" cmd /k "cd /d %FRONTEND_DIR% && yarn dev"
        set "frontend_retry=0"
        goto check_frontend
    )
)
echo   Frontend service is ready!

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