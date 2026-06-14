Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$BEDIR = 'G:\dbm\db\dbm-lite\backend'
$FEDIR = 'G:\dbm\db\dbm-lite\frontend'
$LOGDIR = 'G:\dbm\db\dbm-lite\logs'

if (-not (Test-Path $LOGDIR)) { New-Item -ItemType Directory -Path $LOGDIR -Force | Out-Null }

$TS = Get-Date -Format 'yyyyMMdd-HHmmss'
$BACKEND_LOG = Join-Path $LOGDIR "backend-$TS.log.txt"
$FRONTEND_LOG = Join-Path $LOGDIR "frontend-$TS.log.txt"
$START_LOG = Join-Path $LOGDIR "startup-$TS.log.txt"
$BACKEND_STDOUT = Join-Path $LOGDIR "backend-$TS.stdout.log.txt"
$BACKEND_STDERR = Join-Path $LOGDIR "backend-$TS.stderr.log.txt"
$FRONTEND_STDOUT = Join-Path $LOGDIR "frontend-$TS.stdout.log.txt"
$FRONTEND_STDERR = Join-Path $LOGDIR "frontend-$TS.stderr.log.txt"

function Write-StartLog {
    param([string]$Msg)
    $line = '[' + (Get-Date -Format 'yyyy-MM-dd HH:mm:ss') + '] ' + $msg
    Add-Content -Path $START_LOG -Value $line -Encoding UTF8
    Write-Host $line
}

Write-Host '========================================'
Write-Host '  DBM-Lite 启动脚本'
Write-Host '  轻量级数据库管控平台'
Write-Host '  双域架构：DB操作管控 + DB基础运维'
Write-Host '========================================'
Write-Host ''

'========================================' | Out-File -FilePath $BACKEND_LOG -Encoding UTF8
'  DBM-Lite 后端运行日志' | Out-File -FilePath $BACKEND_LOG -Encoding UTF8 -Append
('  启动时间   : ' + (Get-Date -Format 'yyyy-MM-dd HH:mm:ss')) | Out-File -FilePath $BACKEND_LOG -Encoding UTF8 -Append
('  日志文件名 : backend-' + $TS + '.log.txt') | Out-File -FilePath $BACKEND_LOG -Encoding UTF8 -Append
'  监听端口   : 8080' | Out-File -FilePath $BACKEND_LOG -Encoding UTF8 -Append
'========================================' | Out-File -FilePath $BACKEND_LOG -Encoding UTF8 -Append
'' | Out-File -FilePath $BACKEND_LOG -Encoding UTF8 -Append

'========================================' | Out-File -FilePath $FRONTEND_LOG -Encoding UTF8
'  DBM-Lite 前端运行日志' | Out-File -FilePath $FRONTEND_LOG -Encoding UTF8 -Append
('  启动时间   : ' + (Get-Date -Format 'yyyy-MM-dd HH:mm:ss')) | Out-File -FilePath $FRONTEND_LOG -Encoding UTF8 -Append
('  日志文件名 : frontend-' + $TS + '.log.txt') | Out-File -FilePath $FRONTEND_LOG -Encoding UTF8 -Append
'  监听端口   : 3000' | Out-File -FilePath $FRONTEND_LOG -Encoding UTF8 -Append
'========================================' | Out-File -FilePath $FRONTEND_LOG -Encoding UTF8 -Append
'' | Out-File -FilePath $FRONTEND_LOG -Encoding UTF8 -Append

'========================================' | Out-File -FilePath $START_LOG -Encoding UTF8
'  DBM-Lite 启动流程日志' | Out-File -FilePath $START_LOG -Encoding UTF8 -Append
('  启动时间   : ' + (Get-Date -Format 'yyyy-MM-dd HH:mm:ss')) | Out-File -FilePath $START_LOG -Encoding UTF8 -Append
('  后端日志   : ' + $BACKEND_LOG) | Out-File -FilePath $START_LOG -Encoding UTF8 -Append
('  前端日志   : ' + $FRONTEND_LOG) | Out-File -FilePath $START_LOG -Encoding UTF8 -Append
'========================================' | Out-File -FilePath $START_LOG -Encoding UTF8 -Append
'' | Out-File -FilePath $START_LOG -Encoding UTF8 -Append

Write-Host ('  后端日志 : ' + $BACKEND_LOG)
Write-Host ('  前端日志 : ' + $FRONTEND_LOG)
Write-Host ''

Write-StartLog '[1/4] 清理旧服务和端口占用...'

function Stop-ProcessOnPort {
    param([int]$Port)
    try {
        $conns = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
        foreach ($conn in $conns) {
            try {
                $p = Get-Process -Id $conn.OwningProcess -ErrorAction SilentlyContinue
                if ($p) {
                    Write-Host ('  终止占用 ' + $Port + ' 端口的进程 PID=' + $conn.OwningProcess + ' (' + $p.ProcessName + ')')
                    Write-StartLog ('  终止 PID=' + $conn.OwningProcess + ' (端口 ' + $Port + ')')
                    Stop-Process -Id $conn.OwningProcess -Force -ErrorAction SilentlyContinue
                } else {
                    Write-Host ('  终止占用 ' + $Port + ' 端口的进程 PID=' + $conn.OwningProcess)
                    Write-StartLog ('  终止 PID=' + $conn.OwningProcess + ' (端口 ' + $Port + ')')
                    Stop-Process -Id $conn.OwningProcess -Force -ErrorAction SilentlyContinue
                }
            } catch {}
        }
    } catch {}
}

Stop-ProcessOnPort 8080
Stop-ProcessOnPort 3000
Start-Sleep -Seconds 1
Write-StartLog '[OK] 端口清理完成'
Write-Host ''

Write-StartLog '[2/4] 编译后端服务 (端口 8080)...'

Push-Location $BEDIR
try {
    if (Test-Path (Join-Path $BEDIR 'dbm-lite.exe')) {
        Write-Host '  发现旧版 dbm-lite.exe，准备清理...'
        Remove-Item (Join-Path $BEDIR 'dbm-lite.exe') -Force
    }

    Write-Host '  正在执行 go build...'
    $goOutput = & go build -o dbm-lite.exe ./cmd/server 2>&1
    if ($goOutput) {
        $goOutput | ForEach-Object {
            Add-Content -Path $BACKEND_LOG -Value $_.ToString() -Encoding UTF8
        }
    }

    if ($LASTEXITCODE -ne 0) {
        Write-Host ('[ERROR] 后端编译失败！请查看日志: ' + $BACKEND_LOG)
        Write-StartLog '[ERROR] 后端编译失败'
        Read-Host '按回车键退出'
        exit 1
    }

    Write-Host '[OK] 后端编译成功 (dbm-lite.exe)'
    Write-StartLog '[OK] 后端编译成功'

    Write-Host '  正在启动后端服务...'
    Write-StartLog ('  启动后端服务 - 路径: ' + (Join-Path $BEDIR 'dbm-lite.exe'))

    $backendProcess = Start-Process -FilePath (Join-Path $BEDIR 'dbm-lite.exe') `
        -WorkingDirectory $BEDIR `
        -RedirectStandardOutput $BACKEND_STDOUT `
        -RedirectStandardError $BACKEND_STDERR `
        -PassThru -NoNewWindow

    Write-StartLog ('  后端服务进程已启动 PID=' + $backendProcess.Id)

} finally {
    Pop-Location
}

Write-Host ''

Write-StartLog '[3/4] 启动前端开发服务器 (端口 3000)...'

Push-Location $FEDIR
try {
    if (-not (Test-Path (Join-Path $FEDIR 'node_modules'))) {
        Write-Host '  首次运行，正在安装前端依赖，这可能需要几分钟...'
        Write-StartLog '  首次运行: npm install 开始'

        $npmInstStdout = Join-Path $LOGDIR ('npm-install-' + $TS + '.stdout.log.txt')
        $npmInstStderr = Join-Path $LOGDIR ('npm-install-' + $TS + '.stderr.log.txt')

        $npmInstallProcess = Start-Process -FilePath 'npm.cmd' -ArgumentList @('install') `
            -WorkingDirectory $FEDIR `
            -RedirectStandardOutput $npmInstStdout `
            -RedirectStandardError $npmInstStderr `
            -PassThru -NoNewWindow -Wait

        if ($npmInstallProcess.ExitCode -ne 0) {
            Write-Host ('[ERROR] 前端依赖安装失败！请查看日志: ' + $FRONTEND_LOG)
            Write-StartLog '[ERROR] 前端依赖安装失败'
            Read-Host '按回车键退出'
            exit 1
        }

        Write-Host '[OK] 前端依赖安装成功'
        Write-StartLog '[OK] 前端依赖安装成功'
    }

    Write-Host '  正在启动前端开发服务...'
    Write-StartLog '  启动前端开发服务'

    $frontendProcess = Start-Process -FilePath 'npm.cmd' -ArgumentList @('run', 'dev') `
        -WorkingDirectory $FEDIR `
        -RedirectStandardOutput $FRONTEND_STDOUT `
        -RedirectStandardError $FRONTEND_STDERR `
        -PassThru -NoNewWindow

    Write-StartLog ('  前端服务进程已启动 PID=' + $frontendProcess.Id)

} finally {
    Pop-Location
}

Write-Host ''

Write-StartLog '[4/4] 等待服务就绪 (约 5 秒)...'
Start-Sleep -Seconds 5

Write-Host ''
Write-Host '========================================'
Write-Host '  [OK] 启动完成！'
Write-Host '  请在浏览器中访问: http://localhost:3000'
Write-Host ''
Write-Host '  默认账号: admin'
Write-Host '  默认密码: admin123'
Write-Host '========================================'
Write-Host ''
Write-Host '  运行日志：'
Write-Host ('    ' + $BACKEND_LOG)
Write-Host ('    ' + $FRONTEND_LOG)
Write-Host ''

Write-StartLog '[完成] 启动流程结束'
Write-StartLog '  打开浏览器登录页 http://localhost:3000'

Start-Process 'http://localhost:3000'

Write-Host ''
Read-Host '按回车键退出'
