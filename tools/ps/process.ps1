[CmdletBinding()]
param(
  [string]$op = "help",
  [string]$cmd,
  [string]$cwd,
  [int]$timeout_ms = 0,
  [switch]$offline,
  [string]$trace_id
)

function Write-Json($obj) { $obj | ConvertTo-Json -Depth 6 -Compress }

function Run-Command([string]$command, [string]$workdir, [int]$timeout) {
  $psi = New-Object System.Diagnostics.ProcessStartInfo
  $psi.FileName = 'pwsh'
  $psi.ArgumentList.Add('-NoProfile') | Out-Null
  $psi.ArgumentList.Add('-Command') | Out-Null
  $psi.ArgumentList.Add($command) | Out-Null
  if ($workdir) { $psi.WorkingDirectory = $workdir }
  $psi.RedirectStandardOutput = $true
  $psi.RedirectStandardError = $true
  $psi.UseShellExecute = $false
  $p = New-Object System.Diagnostics.Process
  $p.StartInfo = $psi
  $null = $p.Start()
  if ($timeout -gt 0) { $exited = $p.WaitForExit($timeout) } else { $p.WaitForExit(); $exited = $true }
  if (-not $exited) {
    try { $p.Kill($true) } catch {}
    return @{ exit_code = -1; stdout = $p.StandardOutput.ReadToEnd(); stderr = 'TIMEOUT'; duration_ms = $timeout }
  }
  return @{ exit_code = $p.ExitCode; stdout = $p.StandardOutput.ReadToEnd(); stderr = $p.StandardError.ReadToEnd() }
}

try {
  switch ($op) {
    'ping' { Write-Json @{ ok=$true; data=@{ pong=$true; tool='process.ps1' } }; exit 0 }
    'run' {
      if (-not $cmd) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--cmd requerido' } }; exit 2 }
      $sw = [System.Diagnostics.Stopwatch]::StartNew()
      $res = Run-Command -command $cmd -workdir $cwd -timeout $timeout_ms
      $sw.Stop()
      $res['duration_ms'] = [int]$sw.Elapsed.TotalMilliseconds
      Write-Json @{ ok = ($res.exit_code -eq 0); data = $res }
      exit $res.exit_code
    }
    default {
      Write-Json @{ ok=$false; error=@{ code='USAGE'; message='Usa --op run --cmd "<pwsh comando>"' } }
      exit 1
    }
  }
}
catch {
  Write-Json @{ ok=$false; error=@{ code='EXCEPTION'; message=$_.Exception.Message } }
  exit 10
}
