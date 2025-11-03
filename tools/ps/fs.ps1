[CmdletBinding()]
param(
  [string]$op = "help",
  [string]$path,
  [string]$dest,
  [string]$content,
  [string]$encoding = "utf8",
  [switch]$offline,
  [switch]$dry_run,
  [switch]$confirm,
  [int]$timeout_ms = 0,
  [string]$trace_id
)

function Write-Json($obj) {
  $obj | ConvertTo-Json -Depth 8 -Compress
}

function Ensure-Dir($p) {
  $d = Split-Path -Parent $p
  if ($d -and -not (Test-Path -LiteralPath $d)) {
    New-Item -ItemType Directory -Path $d -Force | Out-Null
  }
}

function Get-PS1Encoding($enc) {
  switch ($enc.ToLower()) {
    'utf8' { 'utf8' }
    'utf-8' { 'utf8' }
    'utf16' { 'unicode' }
    'utf-16' { 'unicode' }
    default { 'utf8' }
  }
}

try {
  switch ($op) {
    'ping' {
      Write-Json @{ ok = $true; data = @{ pong = $true; tool = 'fs.ps1' } }
      exit 0
    }
    'write' {
      if (-not $path) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--path requerido' } }; exit 2 }
      if ($dry_run) {
        $len = if ($null -ne $content) { $content.Length } else { 0 }
        Write-Json @{ ok=$true; data=@{ dry_run = $true; path=$path; length=$len } }
        exit 0
      }
      Ensure-Dir $path
      $enc = Get-PS1Encoding $encoding
      Set-Content -LiteralPath $path -Value $content -Encoding $enc
      Write-Json @{ ok=$true; data=@{ path=$path; length=$content.Length } }
      exit 0
    }
    'read' {
      if (-not $path) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--path requerido' } }; exit 2 }
      if (-not (Test-Path -LiteralPath $path)) { Write-Json @{ ok=$false; error=@{ code='NOT_FOUND'; message="No existe: $path" } }; exit 3 }
      $txt = Get-Content -LiteralPath $path -Raw -ErrorAction Stop
      Write-Json @{ ok=$true; data=@{ path=$path; content=$txt } }
      exit 0
    }
    default {
      $help = @(
        'fs.ps1 --op ping',
        'fs.ps1 --op write --path <ruta> --content <texto> [--encoding utf8|utf16] [--dry-run]',
        'fs.ps1 --op read --path <ruta>'
      )
      Write-Json @{ ok=$false; error=@{ code='USAGE'; message='Operación no soportada'; details=$help } }
      exit 1
    }
  }
}
catch {
  Write-Json @{ ok=$false; error=@{ code='EXCEPTION'; message=$_.Exception.Message; details=$_.Exception.StackTrace } }
  exit 10
}
