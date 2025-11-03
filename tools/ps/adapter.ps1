[CmdletBinding()]
param(
  [string]$op = "help",
  [string]$trace_id
)

function Write-Json($obj) { $obj | ConvertTo-Json -Depth 6 -Compress }

try {
  switch ($op) {
    'ping' { Write-Json @{ ok=$true; data=@{ pong=$true; tool='adapter.ps1'; version='0.1.0' } }; exit 0 }
    'list' {
      $toolsDir = Join-Path (Split-Path -Parent $PSCommandPath) '.'
      $psTools = Get-ChildItem -LiteralPath $toolsDir -Filter '*.ps1' -File | Where-Object { $_.Name -ne 'adapter.ps1' } | ForEach-Object {
        @{ name = $_.BaseName; path = $_.FullName; lang = 'ps' }
      }
      Write-Json @{ ok=$true; data=@{ tools = $psTools } }
      exit 0
    }
    default {
      Write-Json @{ ok=$false; error=@{ code='USAGE'; message='Usa --op ping | --op list' } }
      exit 1
    }
  }
}
catch {
  Write-Json @{ ok=$false; error=@{ code='EXCEPTION'; message=$_.Exception.Message } }
  exit 10
}
