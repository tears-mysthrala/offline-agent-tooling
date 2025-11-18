#!/usr/bin/env pwsh

# Example: Config Tool Usage
# Demonstrates loading configuration from multiple sources

Write-Host "====================" -ForegroundColor Cyan
Write-Host "Config Tool Examples" -ForegroundColor Cyan
Write-Host "====================" -ForegroundColor Cyan

# Example 1: Load .env file
Write-Host "`n1. Load .env file:" -ForegroundColor Yellow
python tools/py/config.py --op load --paths fixtures/config/test.env

# Example 2: Load JSON file
Write-Host "`n2. Load JSON file:" -ForegroundColor Yellow
python tools/py/config.py --op load --paths fixtures/config/test.json

# Example 3: Load with overrides
Write-Host "`n3. Load with overrides:" -ForegroundColor Yellow
python tools/py/config.py --op load --paths fixtures/config/test.env --overrides '{\"PORT\":\"9090\"}'

# Example 4: Help
Write-Host "`n4. Show help:" -ForegroundColor Yellow
python tools/py/config.py --op help
