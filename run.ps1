﻿﻿﻿﻿﻿﻿﻿﻿﻿Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$BEDIR = Join-Path $PSScriptRoot 'backend'
$FEDIR = Join-Path $PSScriptRoot 'frontend'
$LOGDIR = Join-Path $PSScriptRoot 'logs'

if (-not (Test-Path $LOGDIR)) { New-Item -ItemType Directory -Path $LOGDIR -Force | Out-Null }

$TS = Get-Date -Format 'yyyyMMdd-HHmmss'
$START_LOG = Join-Path $LOGDIR "startup-$TS.log.txt"
$BACKEND_STDOUT = Join-Path $LOGDIR "backend-$TS.stdout.log.txt"
$BACKEND_STDERR = Join-Path $LOGDIR "backend-$TS.stderr.log.txt"
$FRONTEND_STDOUT = Join-Path $LOGDIR "frontend-$TS.stdout.log.txt"
$FRONTEND_STDERR = Join-Path $LOGDIR "frontend-$TS.stderr.log.txt"

function Write-StartLog {
    param([string]$Msg)
    $line = '[' + (Get-Date -Format 'yyyy-MM-dd HH:mm:ss') + '] ' + $Msg
    Add-Content -Path $START_LOG -Value $line -Encoding UTF8
    Write-Host $line
}

Write-Host '========================================'
Write-Host '  DBM-Lite Startup Script'
Write-Host '  Backend port: 8080'
Write-Host '  Frontend port: 3000'
Write-Host '========================================'
Write-Host ''

'========================================' | Out-File -FilePath $START_LOG -Encoding UTF8
'  DBM-Lite startup log' | Out-File -FilePath $START_LOG -Encoding UTF8 -Append
('  Time: ' + (Get-Date -Format 'yyyy-MM-dd HH:mm:ss')) | Out-File -FilePath $START_LOG -Encoding UTF8 -Append
'========================================' | Out-File -FilePath $START_LOG -Encoding UTF8 -Append
'' | Out-File -FilePath $START_LOG -Encoding UTF8 -Append

Write-StartLog '[1/3] Killing processes on ports 8080 and 3000...'

function Stop-ProcessOnPort {
    param([int]$Port)
    try {
        $conns = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
        foreach ($conn in $conns) {
            try {
                $p = Get-Process -Id $conn.OwningProcess -ErrorAction SilentlyContinue
                if ($p) {
                    Write-Host ('  Killing PID=' + $conn.OwningProcess + ' (' + $p.ProcessName + ') on port ' + $Port)
                    Write-StartLog ('  Killing PID=' + $conn.OwningProcess + ' on port ' + $Port)
                    Stop-Process -Id $conn.OwningProcess -Force -ErrorAction SilentlyContinue
                } else {
                    Write-Host ('  Killing PID=' + $conn.OwningProcess + ' on port ' + $Port)
                    Write-StartLog ('  Killing PID=' + $conn.OwningProcess + ' on port ' + $Port)
                    Stop-Process -Id $conn.OwningProcess -Force -ErrorAction SilentlyContinue
                }
            } catch {}
        }
    } catch {}
}

Stop-ProcessOnPort 8080
Stop-ProcessOnPort 3000
Start-Sleep -Seconds 1
Write-StartLog '[OK] Ports cleaned (8080, 3000)'
Write-Host ''

Write-StartLog '[2/3] Building backend (port 8080)...'

Push-Location $BEDIR
try {
    $exePath = Join-Path $BEDIR 'dbm-lite.exe'
    if (Test-Path $exePath) {
        Write-Host '  Found old dbm-lite.exe, removing...'
        Remove-Item $exePath -Force
    }

    Write-Host '  Running go build...'
    $goOutput = & go build -o dbm-lite.exe ./cmd/server 2>&1
    if ($goOutput) {
        $goOutput | ForEach-Object {
            Add-Content -Path $START_LOG -Value $_.ToString() -Encoding UTF8
        }
    }

    if ($LASTEXITCODE -ne 0) {
        Write-Host '[ERROR] Backend build failed!'
        Write-StartLog '[ERROR] Backend build failed'
        exit 1
    }

    Write-Host '[OK] Backend build succeeded (dbm-lite.exe)'
    Write-StartLog '[OK] Backend build succeeded'

    Write-Host '  Starting backend service...'
    Write-StartLog ('  Starting backend - path: ' + $exePath)

    $backendProcess = Start-Process -FilePath $exePath `
        -WorkingDirectory $BEDIR `
        -RedirectStandardOutput $BACKEND_STDOUT `
        -RedirectStandardError $BACKEND_STDERR `
        -PassThru -NoNewWindow

    Write-StartLog ('  Backend PID=' + $backendProcess.Id)

} finally {
    Pop-Location
}

Write-Host ''

Write-StartLog '[3/3] Starting frontend dev server (port 3000)...'

Push-Location $FEDIR
try {
    if (-not (Test-Path (Join-Path $FEDIR 'node_modules'))) {
        Write-Host '  First run, installing frontend deps...'
        Write-StartLog '  First run: npm install'

        $npmInstStdout = Join-Path $LOGDIR ('npm-install-' + $TS + '.stdout.log.txt')
        $npmInstStderr = Join-Path $LOGDIR ('npm-install-' + $TS + '.stderr.log.txt')

        $npmInstallProcess = Start-Process -FilePath 'npm.cmd' -ArgumentList @('install') `
            -WorkingDirectory $FEDIR `
            -RedirectStandardOutput $npmInstStdout `
            -RedirectStandardError $npmInstStderr `
            -PassThru -NoNewWindow -Wait

        if ($npmInstallProcess.ExitCode -ne 0) {
            Write-Host '[ERROR] Frontend deps install failed!'
            Write-StartLog '[ERROR] Frontend deps install failed'
            exit 1
        }

        Write-Host '[OK] Frontend deps installed'
        Write-StartLog '[OK] Frontend deps installed'
    }

    Write-Host '  Starting frontend dev server...'
    Write-StartLog '  Starting frontend dev server'

    $frontendProcess = Start-Process -FilePath 'npm.cmd' -ArgumentList @('run', 'dev') `
        -WorkingDirectory $FEDIR `
        -RedirectStandardOutput $FRONTEND_STDOUT `
        -RedirectStandardError $FRONTEND_STDERR `
        -PassThru -NoNewWindow

    Write-StartLog ('  Frontend PID=' + $frontendProcess.Id)

} finally {
    Pop-Location
}

Write-Host ''

Write-StartLog '[4/4] Waiting services ready (5 seconds)...'
Start-Sleep -Seconds 5

Write-Host ''
Write-Host '========================================'
Write-Host '  [OK] Startup complete!'
Write-Host '  Please visit: http://localhost:3000'
Write-Host ''
Write-Host '  Default user: admin'
Write-Host '  Default password: admin123'
Write-Host '========================================'

Write-StartLog '[Done] Startup finished'
Write-StartLog '  Opening browser to http://localhost:3000'

Start-Process 'http://localhost:3000'
