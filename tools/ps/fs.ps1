[CmdletBinding()]
param(
  [string]$op = "help",
  [string]$path,
  [string]$dest,
  [string]$content,
  [string]$encoding = "utf8",
  [string]$pattern,
  [switch]$recursive,
  [int]$limit = 1000,
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
    'write-json' {
      if (-not $path) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--path requerido' } }; exit 2 }
      if (-not $content) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--content requerido' } }; exit 2 }
      Ensure-Dir $path
      $enc = Get-PS1Encoding $encoding
      $json = $null
      try { $json = $content | ConvertFrom-Json -ErrorAction Stop } catch { $json = $content }
      $json | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $path -Encoding $enc
      Write-Json @{ ok=$true; data=@{ path=$path } }
      exit 0
    }
    'read-json' {
      if (-not $path) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--path requerido' } }; exit 2 }
      if (-not (Test-Path -LiteralPath $path)) { Write-Json @{ ok=$false; error=@{ code='NOT_FOUND'; message="No existe: $path" } }; exit 3 }
      $txt = Get-Content -LiteralPath $path -Raw -ErrorAction Stop
      try { $obj = $txt | ConvertFrom-Json -ErrorAction Stop; Write-Json @{ ok=$true; data=@{ path=$path; json=$obj } }; exit 0 } catch { Write-Json @{ ok=$false; error=@{ code='PARSE_ERROR'; message='JSON inválido'; details=$_.Exception.Message } }; exit 6 }
    }
    'append' {
      if (-not $path) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--path requerido' } }; exit 2 }
      Ensure-Dir $path
      Add-Content -LiteralPath $path -Value $content
      Write-Json @{ ok=$true; data=@{ path=$path } }
      exit 0
    }
    'mkdir' {
      if (-not $path) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--path requerido' } }; exit 2 }
      if ($dry_run) { Write-Json @{ ok=$true; data=@{ dry_run=$true; path=$path } }; exit 0 }
      if (-not (Test-Path -LiteralPath $path)) { New-Item -ItemType Directory -Path $path -Force | Out-Null }
      Write-Json @{ ok=$true; data=@{ path=$path } }
      exit 0
    }
    'delete' {
      if (-not $path) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--path requerido' } }; exit 2 }
      if (-not (Test-Path -LiteralPath $path)) { Write-Json @{ ok=$false; error=@{ code='NOT_FOUND'; message="No existe: $path" } }; exit 3 }
      if (-not $confirm) { Write-Json @{ ok=$false; error=@{ code='CONFIRM_REQUIRED'; message='Use --confirm para borrar' } }; exit 4 }
      Remove-Item -LiteralPath $path -Recurse -Force
      Write-Json @{ ok=$true; data=@{ path=$path; deleted=$true } }
      exit 0
    }
    'stat' {
      if (-not $path) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--path requerido' } }; exit 2 }
      if (-not (Test-Path -LiteralPath $path)) { Write-Json @{ ok=$false; error=@{ code='NOT_FOUND'; message="No existe: $path" } }; exit 3 }
      $info = Get-Item -LiteralPath $path
      $meta = @{ FullName=$info.FullName; Length = if ($info.PSIsContainer) { 0 } else { $info.Length }; CreationTime = $info.CreationTime; LastWriteTime = $info.LastWriteTime; Mode = $info.Mode }
      Write-Json @{ ok=$true; data=$meta }
      exit 0
    }
    'checksum' {
      if (-not $path) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--path requerido' } }; exit 2 }
      if (-not (Test-Path -LiteralPath $path)) { Write-Json @{ ok=$false; error=@{ code='NOT_FOUND'; message="No existe: $path" } }; exit 3 }
      $sha = New-Object -TypeName System.Security.Cryptography.SHA256Managed
      $stream = [System.IO.File]::OpenRead($path)
      try { $hash = $sha.ComputeHash($stream) } finally { $stream.Close() }
      $hex = ($hash | ForEach-Object { $_.ToString('x2') }) -join ''
      Write-Json @{ ok=$true; data=@{ path=$path; sha256=$hex } }
      exit 0
    }
    'list' {
      $p = $path ? $path : '.'
      $searchOpt = if ($recursive) { '-Recurse' } else { '' }
      $items = Get-ChildItem -LiteralPath $p -File -Recurse:$recursive -ErrorAction SilentlyContinue | Select-Object -ExpandProperty FullName
      Write-Json @{ ok=$true; data=@{ items=$items } }
      exit 0
    }
    'glob' {
      if (-not $pattern) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--pattern requerido' } }; exit 2 }
      $root = if ($path) { $path } else { '.' }
      $matches = Get-ChildItem -Path $root -Recurse:$recursive -Include $pattern -File -ErrorAction SilentlyContinue | Select-Object -ExpandProperty FullName
      Write-Json @{ ok=$true; data=@{ matches=$matches } }
      exit 0
    }
    'copy' {
      if (-not $path -or -not $dest) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--path y --dest requeridos' } }; exit 2 }
      Ensure-Dir $dest
      Copy-Item -LiteralPath $path -Destination $dest -Recurse -Force
      Write-Json @{ ok=$true; data=@{ from=$path; to=$dest } }
      exit 0
    }
    'move' {
      if (-not $path -or -not $dest) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--path y --dest requeridos' } }; exit 2 }
      Move-Item -LiteralPath $path -Destination $dest -Force
      Write-Json @{ ok=$true; data=@{ from=$path; to=$dest } }
      exit 0
    }
    'read-lines' {
      if (-not $path) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--path requerido' } }; exit 2 }
      $lines = Get-Content -LiteralPath $path -ErrorAction Stop
      Write-Json @{ ok=$true; data=@{ path=$path; lines=$lines } }
      exit 0
    }
    'write-lines' {
      if (-not $path) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--path requerido' } }; exit 2 }
      Ensure-Dir $path
      Set-Content -LiteralPath $path -Value $content -Encoding (Get-PS1Encoding $encoding)
      Write-Json @{ ok=$true; data=@{ path=$path } }
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
