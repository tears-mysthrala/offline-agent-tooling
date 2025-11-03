# PowerShell smoke tests for FS and Search tools (no external deps)
$ErrorActionPreference = 'Stop'

function Write-JsonAndExitIfError($obj) {
    $json = $obj | ConvertTo-Json -Depth 5 -Compress
    Write-Host $json
    if (-not $obj.ok) {
        throw "Test failed: $($json)"
    }
}

$repo = (Resolve-Path "${PSScriptRoot}\..\.." ).ProviderPath
$tmp = Join-Path $repo 'tmp_ps_test.txt'

Write-Host "[TEST] FS: write"
$out = & "$repo\tools\ps\fs.ps1" -op write -path $tmp -content 'ps-hello'
$obj = $out | ConvertFrom-Json
Write-JsonAndExitIfError $obj

Write-Host "[TEST] FS: read"
$out = & "$repo\tools\ps\fs.ps1" -op read -path $tmp
$obj = $out | ConvertFrom-Json
Write-JsonAndExitIfError $obj
if ($obj.data.content -notmatch 'ps-hello') { throw 'content mismatch' }

Write-Host "[TEST] FS: append"
& "$repo\tools\ps\fs.ps1" -op append -path $tmp -content 'ps-append' | Out-Null

Write-Host "[TEST] FS: read-lines"
$out = & "$repo\tools\ps\fs.ps1" -op read-lines -path $tmp
$obj = $out | ConvertFrom-Json
Write-JsonAndExitIfError $obj
if ($obj.data.lines.Count -lt 2) { throw 'read-lines count < 2' }

Write-Host "[TEST] FS: checksum"
$out = & "$repo\tools\ps\fs.ps1" -op checksum -path $tmp
$obj = $out | ConvertFrom-Json
Write-JsonAndExitIfError $obj
if (-not $obj.data.sha256) { throw 'no sha256' }

Write-Host "[TEST] Search: grep sample TODO"
$fixture = Join-Path $repo 'fixtures\search\sample.txt'
if (-not (Test-Path $fixture)) { throw 'fixture missing' }

$out = & "$repo\tools\ps\search.ps1" -op grep -root $repo -pattern 'TODO' -include '*.txt' -recursive
$obj = $out | ConvertFrom-Json
Write-JsonAndExitIfError $obj
if (-not ($obj.data.matches | Where-Object { $_.file -match 'sample.txt' })) {
    throw 'sample.txt not found in search results'
}

Write-Host "All PS tests passed. Cleaning up."
Remove-Item -Path $tmp -ErrorAction SilentlyContinue

Exit 0
