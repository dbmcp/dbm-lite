@echo off
title DBM-Lite Startup
setlocal enabledelayedexpansion
set "ROOT=%~dp0"
set "FRONTEND_DIR=%ROOT%frontend"
set "BACKEND_DIR=%ROOT%backend"

set "BACKEND_PORT=8080"
set "FRONTEND_PORT=3000"

for /f "usebackq tokens=1,2 delims==" %%a in ("%BACKEND_DIR%\.env") do (
    if /i "%%a"=="DBM_LITE_SERVER_PORT" set "BACKEND_PORT=%%b"
)

for /f "usebackq tokens=1,2 delims==" %%a in ("%FRONTEND_DIR%\.env") do (
    if /i "%%a"=="VITE_SERVER_PORT" set "FRONTEND_PORT=%%b"
)

echo ============================================================
echo   DBM-Lite Project Build and Run
echo   Project Path: %ROOT%
echo   Backend Port: %BACKEND_PORT%
echo   Frontend Port: %FRONTEND_PORT%
echo ============================================================

echo.
echo [1/6] Cleaning ports %BACKEND_PORT%, %FRONTEND_PORT% ...
for %%P in (%BACKEND_PORT% %FRONTEND_PORT%) do (
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%%P " ^| findstr LISTENING') do (
        echo   - Killing process on port %%P, PID: %%a
        taskkill /F /PID %%a >nul 2>&1
    )
)

timeout /t 1 /nobreak >nul

echo.
echo [2/6] Building backend (Go) ...
cd /d "%BACKEND_DIR%"
if exist dbm-lite.exe del /Q dbm-lite.exe
echo   Compiling dbm-lite.exe ...
go build -o dbm-lite.exe ./cmd/server
if errorlevel 1 (
    echo   ERROR: Backend build failed!
    pause
    exit /b 1
)
echo   Starting backend service on port %BACKEND_PORT% ...
start "DBM-Backend-%BACKEND_PORT%" cmd /k "cd /d %BACKEND_DIR% && dbm-lite.exe"

echo.
echo [3/6] Waiting for backend to be ready ...
set "retry_count=0"
set "max_retries=30"
:check_backend
timeout /t 2 /nobreak >nul
set /a retry_count+=1
curl -s http://localhost:%BACKEND_PORT%/api/health >nul 2>&1
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
echo   Starting frontend dev server on port %FRONTEND_PORT% ...
start "DBM-Frontend-%FRONTEND_PORT%" cmd /k "cd /d %FRONTEND_DIR% && yarn dev"

echo.
echo ============================================================
echo   Waiting for frontend to be ready ...
echo ============================================================
set "frontend_retry=0"
set "frontend_max_retries=30"
:check_frontend
timeout /t 2 /nobreak >nul
set /a frontend_retry+=1
curl -s http://localhost:%FRONTEND_PORT% >nul 2>&1
if errorlevel 1 (
    if !frontend_retry! lss %frontend_max_retries% (
        echo   Waiting for frontend... [!frontend_retry!/%frontend_max_retries%]
        goto check_frontend
    ) else (
        echo   WARNING: Frontend not responding, attempting rebuild...
        for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%FRONTEND_PORT% " ^| findstr LISTENING') do (
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
        start "DBM-Frontend-%FRONTEND_PORT%" cmd /k "cd /d %FRONTEND_DIR% && yarn dev"
        set "frontend_retry=0"
        goto check_frontend
    )
)
echo   Frontend service is ready!

echo.
echo   Opening browser: http://localhost:%FRONTEND_PORT%
start http://localhost:%FRONTEND_PORT%

echo.
echo ============================================================
echo   All services started successfully!
echo   - Frontend: http://localhost:%FRONTEND_PORT%
echo   - Backend API: http://localhost:%BACKEND_PORT%
echo   - Login: admin (password in .env config)
echo ============================================================
endlocal
pause