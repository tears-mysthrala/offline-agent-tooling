[CmdletBinding()]
param(
  [string]$op = "help",
  [string]$level = "info",
  [string]$msg,
  [string]$fieldsJson,
  [string]$trace_id
)

function Write-Json($obj) { $obj | ConvertTo-Json -Depth 6 -Compress }

function Get-LogPath() {
  $date = Get-Date -Format 'yyyyMMdd'
  $dir = Join-Path (Split-Path -Parent $PSCommandPath) '..' | Resolve-Path | Select-Object -ExpandProperty Path
  # Preferir carpeta "logs" del repo actual
  $repoLogs = Join-Path (Get-Location) 'logs'
  if (-not (Test-Path $repoLogs)) { New-Item -ItemType Directory -Path $repoLogs -Force | Out-Null }
  return (Join-Path $repoLogs "agent-$date.log")
}

try {
  switch ($op) {
    'ping' { Write-Json @{ ok=$true; data=@{ pong=$true; tool='log.ps1' } }; exit 0 }
    'log' {
      if (-not $msg) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--msg requerido' } }; exit 2 }
      $ts = (Get-Date).ToString('o')
      $fields = $null
      if ($fieldsJson) { try { $fields = $fieldsJson | ConvertFrom-Json } catch { $fields = @{ parse_error = $fieldsJson } } }
      $entry = @{ ts=$ts; level=$level; msg=$msg; fields=$fields; trace_id=$trace_id }
      $line = ($entry | ConvertTo-Json -Depth 8 -Compress)
      $logPath = Get-LogPath
      Add-Content -LiteralPath $logPath -Value $line
      Write-Json @{ ok=$true; data=@{ path=$logPath } }
      exit 0
    }
    default {
      Write-Json @{ ok=$false; error=@{ code='USAGE'; message='Usa --op log --level info|warn|error --msg "texto"' } }
      exit 1
    }
  }
}
catch {
  Write-Json @{ ok=$false; error=@{ code='EXCEPTION'; message=$_.Exception.Message } }
  exit 10
}
