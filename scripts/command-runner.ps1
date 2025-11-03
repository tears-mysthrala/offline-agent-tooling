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

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$logFile = Join-Path -Path $logsPath -ChildPath "command-runner-$timestamp.log"

"Command runner started at $(Get-Date -Format o)" | Out-File -FilePath $logFile -Encoding UTF8
"Working directory: $workspaceRoot" | Out-File -FilePath $logFile -Append -Encoding UTF8
"Continue on error: $ContinueOnError" | Out-File -FilePath $logFile -Append -Encoding UTF8
"Commands file: $commandsPath" | Out-File -FilePath $logFile -Append -Encoding UTF8

$overallSuccess = $true

Push-Location -Path $workspaceRoot
try {
    foreach ($command in $commands) {
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

if ($overallSuccess) {
    Write-Host "All commands completed. Log saved to $logFile"
    "All commands completed successfully." | Out-File -FilePath $logFile -Append -Encoding UTF8
    exit 0
} else {
    Write-Warning "One or more commands failed. Review $logFile for details."
    "Completed with errors." | Out-File -FilePath $logFile -Append -Encoding UTF8
    exit 1
}
