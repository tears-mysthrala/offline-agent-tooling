param(
    [switch]$ContinueOnError
)

$workspaceRoot = Split-Path -Path $PSScriptRoot -Parent
$commandsPath = Join-Path -Path $PSScriptRoot -ChildPath "commands.txt"
$logsPath = Join-Path -Path $workspaceRoot -ChildPath "logs"

if (-not (Test-Path -Path $commandsPath)) {
    Write-Error "Command queue not found at $commandsPath"
    exit 1
}

$commands = Get-Content -Path $commandsPath | Where-Object { $_.Trim() -ne "" -and -not $_.TrimStart().StartsWith('#') }

if ($commands.Count -eq 0) {
    Write-Warning "No commands to run. Populate scripts/commands.txt first."
    exit 0
}

if (-not (Test-Path -Path $logsPath)) {
    New-Item -ItemType Directory -Path $logsPath | Out-Null
}

# Prepare archive directory and rotate old logs
$archivePath = Join-Path -Path $logsPath -ChildPath "archive"
if (-not (Test-Path -Path $archivePath)) {
    New-Item -ItemType Directory -Path $archivePath | Out-Null
}

# Rotate logs older than $retentionDays or when total logs exceed $maxBytes
$retentionDays = 7
$maxBytes = 1MB

# Move logs older than retention window
$cutoff = (Get-Date).AddDays(-$retentionDays)
$oldLogs = Get-ChildItem -Path $logsPath -File -Filter 'command-runner-*.log' | Where-Object { $_.LastWriteTime -lt $cutoff }

function Add-Files-To-DailyArchive {
    param(
        [Parameter(Mandatory=$true)] [System.IO.FileInfo[]] $FilesToAdd
    )

    if (-not $FilesToAdd -or $FilesToAdd.Count -eq 0) { return }

    $dateStr = (Get-Date).ToString('yyyy-MM-dd')
    $dailyZip = Join-Path -Path $archivePath -ChildPath ("$dateStr.zip")

    $sevenZip = Get-Command 7z -ErrorAction SilentlyContinue
    if ($sevenZip) {
        # Use a temporary list file to avoid command-line length limits
        $listFile = [System.IO.Path]::GetTempFileName()
        try {
            $FilesToAdd | ForEach-Object { $_.FullName } | Out-File -FilePath $listFile -Encoding ASCII
            $arg = "@${listFile}"
            & $sevenZip.Path a $dailyZip $arg -tzip -mx=9 | Out-Null
            if ($LASTEXITCODE -ne 0) {
                Write-Warning "7z failed with exit code $LASTEXITCODE; falling back to .NET compression."
                throw "7z-failed"
            }
        } catch {
            # fallback to .NET below
            if ($_) { Write-Verbose "7z error: $_" }
            goto DotNetFallback
        } finally {
            if (Test-Path $listFile) { Remove-Item $listFile -Force }
        }
    } else {
        # no 7z found; fallback
        DotNetFallback:
        Add-Type -AssemblyName System.IO.Compression.FileSystem -ErrorAction SilentlyContinue
        $mode = if (Test-Path $dailyZip) { [System.IO.Compression.ZipArchiveMode]::Update } else { [System.IO.Compression.ZipArchiveMode]::Create }
        $zip = [System.IO.Compression.ZipFile]::Open($dailyZip, $mode)
        foreach ($f in $FilesToAdd) {
            # remove existing entry if present
            $existing = $zip.Entries | Where-Object { $_.Name -eq $f.Name }
            foreach ($e in $existing) { $e.Delete() }
            [System.IO.Compression.ZipFileExtensions]::CreateEntryFromFile($zip, $f.FullName, $f.Name, [System.IO.Compression.CompressionLevel]::Optimal)
        }
        $zip.Dispose()
    }

    # remove originals after successful archive
    foreach ($f in $FilesToAdd) { Remove-Item -Path $f.FullName -Force }
}

# Add old logs to daily archive
if ($oldLogs.Count -gt 0) { Add-Files-To-DailyArchive -FilesToAdd $oldLogs }

# If total log size exceeds threshold, move oldest logs to archive until under limit
$allLogs = Get-ChildItem -Path $logsPath -File -Filter 'command-runner-*.log' | Sort-Object LastWriteTime
$totalSize = ($allLogs | Measure-Object -Property Length -Sum).Sum
if ($null -eq $totalSize) { $totalSize = 0 }
if ($totalSize -gt $maxBytes) {
    $toArchive = @()
    foreach ($f in $allLogs) {
        if ($totalSize -le $maxBytes) { break }
        $toArchive += $f
        $totalSize -= $f.Length
    }
    if ($toArchive.Count -gt 0) { Add-Files-To-DailyArchive -FilesToAdd $toArchive }
}

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$logFile = Join-Path -Path $logsPath -ChildPath "command-runner-$timestamp.log"

"Command runner started at $(Get-Date -Format o)" | Out-File -FilePath $logFile -Encoding UTF8
"Working directory: $workspaceRoot" | Out-File -FilePath $logFile -Append -Encoding UTF8
"Continue on error: $ContinueOnError" | Out-File -FilePath $logFile -Append -Encoding UTF8
"Commands file: $commandsPath" | Out-File -FilePath $logFile -Append -Encoding UTF8

$overallSuccess = $true
$failedCommands = @()
$commandIndex = 0

Push-Location -Path $workspaceRoot
try {
    foreach ($command in $commands) {
        $commandIndex++
        $header = "`n>>> $command"
        Write-Host $header
        $header | Out-File -FilePath $logFile -Append -Encoding UTF8

        $commandSucceeded = $true
        $nativeExit = $null
        $oldPreference = $ErrorActionPreference
        $ErrorActionPreference = 'Stop'

        try {
            Invoke-Expression $command 2>&1 |
                Tee-Object -FilePath $logFile -Append |
                ForEach-Object {
                    $text = $_.ToString()
                    if ($text.Length -gt 0) {
                        Write-Host $text
                    }
                }
        } catch {
            $commandSucceeded = $false
            $message = $_.Exception.Message
            Write-Host $message -ForegroundColor Red
            $message | Out-File -FilePath $logFile -Append -Encoding UTF8
            if ($_.InvocationInfo -and $_.InvocationInfo.PositionMessage) {
                $_.InvocationInfo.PositionMessage | Out-File -FilePath $logFile -Append -Encoding UTF8
            }
        } finally {
            $ErrorActionPreference = $oldPreference
            $nativeExit = if ($LASTEXITCODE -ne $null) { $LASTEXITCODE } else { 0 }
        }

        $exitCode = if ($nativeExit -ne 0) { $nativeExit } elseif ($commandSucceeded) { 0 } else { 1 }

        $resultLine = "Exit code: $exitCode"
        Write-Host $resultLine
        $resultLine | Out-File -FilePath $logFile -Append -Encoding UTF8

        if ($exitCode -ne 0) {
            $overallSuccess = $false
            # Store a simple delimited string to avoid serialization/display issues
            $failedCommands += "${commandIndex}|${exitCode}|${command}"
            if (-not $ContinueOnError) {
                Write-Error "Command failed and ContinueOnError is not set. Halting queue."
                "Command failed. Aborting remaining commands." | Out-File -FilePath $logFile -Append -Encoding UTF8
                break
            }
        }
    }
} finally {
    Pop-Location
}

# Write summary to log and console
"`n--- Summary ---`n" | Out-File -FilePath $logFile -Append -Encoding UTF8
"Run finished at $(Get-Date -Format o)" | Out-File -FilePath $logFile -Append -Encoding UTF8
"Commands run: $commandIndex" | Out-File -FilePath $logFile -Append -Encoding UTF8
"Failures: $($failedCommands.Count)" | Out-File -FilePath $logFile -Append -Encoding UTF8

if ($failedCommands.Count -gt 0) {
    Write-Host "Failures summary:" -ForegroundColor Yellow
    foreach ($f in $failedCommands) {
        $parts = $f -split '\|',3
        $idx = $parts[0]
        $code = $parts[1]
        $cmd = $parts[2]
        $line = "[#${idx}] Exit ${code} - ${cmd}"
        Write-Host $line -ForegroundColor Yellow
        $line | Out-File -FilePath $logFile -Append -Encoding UTF8
    }
} else {
    Write-Host "All commands completed successfully." -ForegroundColor Green
}

"Log file: $logFile" | Out-File -FilePath $logFile -Append -Encoding UTF8

if ($overallSuccess) {
    Write-Host "All commands completed. Log saved to $logFile"
    "All commands completed successfully." | Out-File -FilePath $logFile -Append -Encoding UTF8
    exit 0
} else {
    Write-Warning "One or more commands failed. Review $logFile for details."
    "Completed with errors." | Out-File -FilePath $logFile -Append -Encoding UTF8
    exit 1
}
