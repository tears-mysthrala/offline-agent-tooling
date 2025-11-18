# Performance Comparison Report

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "  Performance Comparison: Go vs Python" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan

# Warm-up
Write-Host "`nWarming up..." -ForegroundColor Yellow
1..3 | ForEach-Object {
    python tools/py/kv.py --op ping | Out-Null
    python tools/py/cache.py --op ping | Out-Null
    ./bin/kv.exe --op ping | Out-Null
    ./bin/cache.exe --op ping | Out-Null
    ./bin/fs.exe --op ping | Out-Null
}

# KV Benchmark
Write-Host "`n--- KV Store ---" -ForegroundColor Green
$pyKv = (1..10 | ForEach-Object { (Measure-Command { python tools/py/kv.py --op ping | Out-Null }).TotalMilliseconds } | Measure-Object -Average).Average
$goKv = (1..10 | ForEach-Object { (Measure-Command { ./bin/kv.exe --op ping | Out-Null }).TotalMilliseconds } | Measure-Object -Average).Average
Write-Host "Python:  $([math]::Round($pyKv, 1))ms"
Write-Host "Go:      $([math]::Round($goKv, 1))ms"
Write-Host "Speedup: $([math]::Round($pyKv/$goKv, 1))x faster" -ForegroundColor Cyan

# Cache Benchmark
Write-Host "`n--- Cache ---" -ForegroundColor Green
$pyCache = (1..10 | ForEach-Object { (Measure-Command { python tools/py/cache.py --op ping | Out-Null }).TotalMilliseconds } | Measure-Object -Average).Average
$goCache = (1..10 | ForEach-Object { (Measure-Command { ./bin/cache.exe --op ping | Out-Null }).TotalMilliseconds } | Measure-Object -Average).Average
Write-Host "Python:  $([math]::Round($pyCache, 1))ms"
Write-Host "Go:      $([math]::Round($goCache, 1))ms"
Write-Host "Speedup: $([math]::Round($pyCache/$goCache, 1))x faster" -ForegroundColor Cyan

# FS Benchmark
Write-Host "`n--- Filesystem ---" -ForegroundColor Green
$goFs = (1..10 | ForEach-Object { (Measure-Command { ./bin/fs.exe --op ping | Out-Null }).TotalMilliseconds } | Measure-Object -Average).Average
Write-Host "Go:      $([math]::Round($goFs, 1))ms"

# Summary
Write-Host "`n=====================================" -ForegroundColor Cyan
Write-Host "  Summary" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
$avgPython = ($pyKv + $pyCache) / 2
$avgGo = ($goKv + $goCache + $goFs) / 3
Write-Host "Average Python: $([math]::Round($avgPython, 1))ms"
Write-Host "Average Go:     $([math]::Round($avgGo, 1))ms"
Write-Host "Overall Speedup: $([math]::Round($avgPython/$avgGo, 1))x faster" -ForegroundColor Green

# Token savings
Write-Host "`n--- Token Savings Estimate ---" -ForegroundColor Yellow
$normalBytes = 60
$compactBytes = 35
$reduction = (($normalBytes - $compactBytes) / $normalBytes) * 100
Write-Host "Normal output: ~${normalBytes} bytes"
Write-Host "Compact output: ~${compactBytes} bytes"
Write-Host "Reduction: $([math]::Round($reduction, 0))%" -ForegroundColor Green

$ops = 100
$pythonTokens = $ops * 20
$goCompactTokens = $ops * 12
Write-Host "`nFor $ops operations/day:"
Write-Host "  Python normal: ~$pythonTokens tokens"
Write-Host "  Go + compact:  ~$goCompactTokens tokens"
Write-Host "  Saved:         ~$($pythonTokens - $goCompactTokens) tokens/day" -ForegroundColor Green
$savingsPercent = (($pythonTokens - $goCompactTokens) / $pythonTokens) * 100
Write-Host "  Savings:       $([math]::Round($savingsPercent, 0))%" -ForegroundColor Green
