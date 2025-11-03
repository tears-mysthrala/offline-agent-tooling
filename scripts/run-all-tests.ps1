<#
Run both Python tests and PowerShell Pester tests and produce test reports.
This is a convenience script for local development and CI debugging.
#>

param(
    [switch]$FailOnError = $true
)

Write-Host "Running Python tests (pytest -> test-reports/python-junit.xml)"
if (-not (Test-Path -Path test-reports)) { New-Item -ItemType Directory -Path test-reports | Out-Null }
try {
    python -m pip install -r requirements-dev.txt
    pytest tests/python --junitxml=test-reports/python-junit.xml --cov=.
} catch {
    Write-Error "Python tests failed: $_"
    if ($FailOnError) { exit 1 }
}

Write-Host "Running PowerShell Pester tests (-> test-reports/pester-nunit.xml)"
try {
    if (-not (Get-Module -ListAvailable -Name Pester)) { Install-Module -Name Pester -Scope CurrentUser -Force -AllowClobber }
    Invoke-Pester -Script tests/ps/pester -OutputFormat NUnitXml -OutputFile test-reports/pester-nunit.xml -PassThru | Out-Null
} catch {
    Write-Error "Pester tests failed: $_"
    if ($FailOnError) { exit 2 }
}

Write-Host "All test runs complete. Reports are in test-reports/*.xml"
