#!/usr/bin/env pwsh

# Example: Cache Tool Usage
# Demonstrates caching with TTL

Write-Host "==================" -ForegroundColor Cyan
Write-Host "Cache Tool Examples" -ForegroundColor Cyan
Write-Host "==================" -ForegroundColor Cyan

# Example 1: Put value
Write-Host "`n1. Put value in cache:" -ForegroundColor Yellow
python tools/py/cache.py --op put --key mykey --value '"test data"'

# Example 2: Get value
Write-Host "`n2. Get value from cache:" -ForegroundColor Yellow
python tools/py/cache.py --op get --key mykey

# Example 3: Put with TTL
Write-Host "`n3. Put with 10 second TTL:" -ForegroundColor Yellow
python tools/py/cache.py --op put --key tempkey --value '"expires soon"' --ttl-s 10

# Example 4: Cache stats
Write-Host "`n4. Get cache statistics:" -ForegroundColor Yellow
python tools/py/cache.py --op stats

# Example 5: Help
Write-Host "`n5. Show help:" -ForegroundColor Yellow
python tools/py/cache.py --op help
