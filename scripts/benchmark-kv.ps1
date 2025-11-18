# Performance Benchmark Script
# Compares Python vs Go implementations

Write-Host "====================" -ForegroundColor Cyan
Write-Host "KV Store Benchmark" -ForegroundColor Cyan  
Write-Host "====================" -ForegroundColor Cyan

# Warm-up runs
Write-Host "`nWarming up..." -ForegroundColor Yellow
1..3 | ForEach-Object {
    python tools/py/kv.py --op ping | Out-Null
    ./bin/kv.exe --op ping | Out-Null
}

# Python benchmark
Write-Host "`nPython (kv.py):" -ForegroundColor Yellow
$pythonTimes = 1..10 | ForEach-Object {
    (Measure-Command { python tools/py/kv.py --op ping | Out-Null }).TotalMilliseconds
}
$pythonAvg = ($pythonTimes | Measure-Object -Average).Average
Write-Host "  Average: $([math]::Round($pythonAvg, 2))ms"

# Go benchmark
Write-Host "`nGo (kv.exe):" -ForegroundColor Yellow
$goTimes = 1..10 | ForEach-Object {
    (Measure-Command { ./bin/kv.exe --op ping | Out-Null }).TotalMilliseconds
}
$goAvg = ($goTimes | Measure-Object -Average).Average
Write-Host "  Average: $([math]::Round($goAvg, 2))ms"

# Results
Write-Host "`nResults:" -ForegroundColor Green
Write-Host "  Python: $([math]::Round($pythonAvg, 2))ms"
Write-Host "  Go:     $([math]::Round($goAvg, 2))ms"
$speedup = $pythonAvg / $goAvg
Write-Host "  Speedup: $([math]::Round($speedup, 1))x faster" -ForegroundColor Cyan

# Token savings estimate
$calls = 100
$pythonTokens = $calls * 20
$goTokens = $calls * 15
Write-Host "`nToken savings (100 calls/day):" -ForegroundColor Yellow
Write-Host "  Python: ~$pythonTokens tokens"
Write-Host "  Go:     ~$goTokens tokens"  
Write-Host "  Saved:  ~$($pythonTokens - $goTokens) tokens/day" -ForegroundColor Green
