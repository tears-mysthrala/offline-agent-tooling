$py = Get-Command python -ErrorAction SilentlyContinue
if (-not $py) { $py = Get-Command py -ErrorAction SilentlyContinue }
if ($py) {
  & $py.Path -c "import runpy,sys; sys.argv=['http_tool.py','--op','get','--offline','--fixture-key','example-users']; runpy.run_path('tools/py/http_tool.py', run_name='__main__')"
  exit $LASTEXITCODE
} else {
  Write-Output 'PYTHON_MISSING'
  exit 0
}
