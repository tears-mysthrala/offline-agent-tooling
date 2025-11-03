[CmdletBinding()]
param(
  [string]$op = "help",
  [string]$root = ".",
  [string]$pattern,
  [switch]$is_regex,
  [string]$include,
  [string]$exclude,
  [switch]$recursive,
  [int]$limit = 2000,
  [switch]$ignore_git,
  [switch]$offline,
  [string]$trace_id
)

function Write-Json($obj) { $obj | ConvertTo-Json -Depth 6 -Compress }

try {
  switch ($op) {
    'ping' { Write-Json @{ ok=$true; data=@{ pong=$true; tool='search.ps1' } }; exit 0 }
    'grep' {
      if (-not $pattern) { Write-Json @{ ok=$false; error=@{ code='ARG_MISSING'; message='--pattern requerido' } }; exit 2 }
      if (-not (Test-Path -LiteralPath $root)) { Write-Json @{ ok=$false; error=@{ code='NOT_FOUND'; message="No existe root: $root" } }; exit 3 }
      $incs = @()
      if ($include) { $incs = $include -split ',' | ForEach-Object { $_.Trim() } }
      $exs = @()
      if ($exclude) { $exs = $exclude -split ',' | ForEach-Object { $_.Trim() } }
      if ($ignore_git) {
        $gitignore = Join-Path $root '.gitignore'
        if (Test-Path $gitignore) {
          $lines = Get-Content -LiteralPath $gitignore | ForEach-Object { $_.Trim() } | Where-Object { $_ -and -not ($_ -like '#*') }
          foreach ($l in $lines) { $exs += $l }
        }
      }

      $gciParams = @{ Path = $root; File = $true; ErrorAction = 'SilentlyContinue' }
      if ($recursive) { $gciParams['Recurse'] = $true }
      if ($incs.Count -gt 0) { $gciParams['Include'] = $incs }
      if ($exs.Count -gt 0) { $gciParams['Exclude'] = $exs }

      $files = Get-ChildItem @gciParams | Select-Object -ExpandProperty FullName
      $opts = @{ Pattern = $pattern; AllMatches = $true; Encoding = 'utf8' }
      if ($is_regex) { $opts['SimpleMatch'] = $false } else { $opts['SimpleMatch'] = $true }
      $res = @()
      foreach ($f in $files) {
        try {
          $matches = Select-String -Path $f @opts -ErrorAction Stop
          foreach ($m in $matches) {
            $res += @{ file=$m.Path; line=$m.LineNumber; col=$m.ColumnNumber; text=$m.Line.Trim() }
            if ($res.Count -ge $limit) { break }
          }
        } catch {}
        if ($res.Count -ge $limit) { break }
      }
      Write-Json @{ ok=$true; data=@{ matches=$res } }
      exit 0
    }
    default {
      Write-Json @{ ok=$false; error=@{ code='USAGE'; message='Usa --op grep --root <dir> --pattern <texto>' } }
      exit 1
    }
  }
}
catch {
  Write-Json @{ ok=$false; error=@{ code='EXCEPTION'; message=$_.Exception.Message } }
  exit 10
}
