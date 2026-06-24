@echo off
chcp 65001 >nul
echo === STEP 1: CLEANUP ===
echo Killing dbm-lite.exe...
taskkill /F /IM dbm-lite.exe 2>nul
echo Killing node.exe...
taskkill /F /IM node.exe 2>nul
echo Waiting 2 seconds...
timeout /T 2 /NOBREAK >nul
echo STEP 1 DONE
echo.

echo === STEP 2: BUILD & START BACKEND ===
cd /d d:\dbm\dbm-lite\backend
echo Building...
go build -o dbm-lite.exe ./cmd/server
echo Build exit code: %ERRORLEVEL%
echo Starting backend...
start /B "" dbm-lite.exe
echo Waiting 5 seconds...
timeout /T 5 /NOBREAK >nul

echo Verifying backend on 8088...
powershell -Command "$body='{\"username\":\"admin\",\"password\":\"admin123\"}'; try { $r=Invoke-RestMethod -Uri 'http://localhost:8088/api/auth/login' -Method Post -Body $body -ContentType 'application/json'; Write-Host 'BACKEND_OK: success=' $r.success } catch { Write-Host 'BACKEND_FAIL1: retry in 3s...'; Start-Sleep -Seconds 3; try { $r=Invoke-RestMethod -Uri 'http://localhost:8088/api/auth/login' -Method Post -Body $body -ContentType 'application/json'; Write-Host 'BACKEND_OK_RETRY: success=' $r.success } catch { Write-Host 'BACKEND_FAIL2' } }"

echo STEP 2 DONE
echo.

echo === STEP 3: START FRONTEND ===
cd /d d:\dbm\dbm-lite\frontend
echo Starting frontend dev server...
start /B "" yarn.cmd dev
echo Waiting 10 seconds...
timeout /T 10 /NOBREAK >nul

echo Verifying frontend on 3033...
powershell -Command "try { $r=Invoke-WebRequest -Uri 'http://localhost:3033/' -UseBasicParsing; Write-Host 'FRONTEND_OK: StatusCode=' $r.StatusCode } catch { Write-Host 'FRONTEND_FAIL1: retry in 5s...'; Start-Sleep -Seconds 5; try { $r=Invoke-WebRequest -Uri 'http://localhost:3033/' -UseBasicParsing; Write-Host 'FRONTEND_OK_RETRY: StatusCode=' $r.StatusCode } catch { Write-Host 'FRONTEND_FAIL2' } }"

echo.
echo === ALL STEPS COMPLETE ===
