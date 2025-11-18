[CmdletBinding()]
param(
  [string]$op = "help",
  [string]$repo = ".",
  [string]$file,
  [int]$limit = 20,
  [switch]$offline,
  [string]$trace_id
)

function Write-Json($obj) { $obj | ConvertTo-Json -Depth 8 -Compress }

function Test-GitAvailable {
  try {
    $null = & git --version 2>$null
    return $true
  } catch {
    return $false
  }
}

try {
  switch ($op) {
    'ping' { 
      $hasGit = Test-GitAvailable
      Write-Json @{ ok=$true; data=@{ pong=$true; tool='git.ps1'; has_git=$hasGit } }
      exit 0
    }
    'status' {
      if (-not (Test-GitAvailable)) {
        Write-Json @{ ok=$false; error=@{ code='GIT_MISSING'; message='git binary not found' } }
        exit 3
      }
      Push-Location -Path $repo
      try {
        $output = & git status --short 2>&1
        $modified = @()
        $untracked = @()
        foreach ($line in $output) {
          if ($line -match '^\s*M\s+(.+)$') { $modified += $matches[1] }
          elseif ($line -match '^\?\?\s+(.+)$') { $untracked += $matches[1] }
        }
        Write-Json @{ ok=$true; data=@{ modified=$modified; untracked=$untracked } }
        exit 0
      } finally {
        Pop-Location
      }
    }
    'log' {
      if (-not (Test-GitAvailable)) {
        Write-Json @{ ok=$false; error=@{ code='GIT_MISSING'; message='git binary not found' } }
        exit 3
      }
      Push-Location -Path $repo
      try {
        $output = & git log --oneline -n $limit 2>&1
        $commits = @()
        foreach ($line in $output) {
          if ($line -match '^([a-f0-9]+)\s+(.+)$') {
            $commits += @{ hash=$matches[1]; message=$matches[2] }
          }
        }
        Write-Json @{ ok=$true; data=@{ commits=$commits } }
        exit 0
      } finally {
        Pop-Location
      }
    }
    'diff' {
      if (-not (Test-GitAvailable)) {
        Write-Json @{ ok=$false; error=@{ code='GIT_MISSING'; message='git binary not found' } }
        exit 3
      }
      Push-Location -Path $repo
      try {
        $args = @('diff')
        if ($file) { $args += $file }
        $output = & git @args 2>&1 | Out-String
        Write-Json @{ ok=$true; data=@{ diff=$output } }
        exit 0
      } finally {
        Pop-Location
      }
    }
    'branch' {
      if (-not (Test-GitAvailable)) {
        Write-Json @{ ok=$false; error=@{ code='GIT_MISSING'; message='git binary not found' } }
        exit 3
      }
      Push-Location -Path $repo
      try {
        $current = & git branch --show-current 2>&1
        $all = & git branch --list 2>&1 | ForEach-Object { $_.Trim('* ').Trim() }
        Write-Json @{ ok=$true; data=@{ current=$current; all=$all } }
        exit 0
      } finally {
        Pop-Location
      }
    }
    default {
      Write-Json @{ ok=$false; error=@{ code='USAGE'; message='Usa --op status|log|diff|branch [--repo <path>]' } }
      exit 1
    }
  }
}
catch {
  Write-Json @{ ok=$false; error=@{ code='EXCEPTION'; message=$_.Exception.Message } }
  exit 10
}
