$ErrorActionPreference = "Stop"

function Measure-Tool {
    param($Name, $Command)
    Write-Host "Measuring $Name..." -NoNewline
    $time = Measure-Command { Invoke-Expression $Command }
    Write-Host " $($time.TotalMilliseconds)ms"
}

# Setup
$testDir = "d:\tooling\perf_test_data"
if (Test-Path $testDir) { Remove-Item $testDir -Recurse -Force }
New-Item -ItemType Directory -Path $testDir | Out-Null

Write-Host "Generating 5000 test files..."
1..5000 | ForEach-Object {
    $id = $_
    $content = "This is file number $id`nSome random content for searching.`nKey: value$id"
    Set-Content -Path "$testDir\file_$id.txt" -Value $content
}

# Test fs.exe list
Measure-Tool "fs list (cold)" ".\bin\fs.exe --op list --path $testDir --compact"
Measure-Tool "fs list (warm)" ".\bin\fs.exe --op list --path $testDir --compact"

# Test search.exe grep
Measure-Tool "search grep (cold)" ".\bin\search.exe --op grep --root $testDir --pattern 'value4000' --compact"
Measure-Tool "search grep (warm)" ".\bin\search.exe --op grep --root $testDir --pattern 'value4000' --compact"

# Cleanup
Remove-Item $testDir -Recurse -Force
