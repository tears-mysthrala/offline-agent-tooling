# Quick Start Guide

This guide will get you started with the offline agent tooling in 5 minutes.

## Prerequisites
- Windows with PowerShell
- Python 3.x installed

## 1. Test Tool Availability

All tools support the `ping` operation for health checks:

```powershell
# Test Python tools
python tools/py/config.py --op ping
python tools/py/cache.py --op ping

# Test PowerShell tools
pwsh -File tools/ps/fs.ps1 --op ping
pwsh -File tools/ps/git.ps1 --op ping
```

## 2. Try Basic Operations

### File Operations (fs.ps1)
```powershell
# Create a file
pwsh -File tools/ps/fs.ps1 --op write --path test.txt --content "Hello World"

# Read the file
pwsh -File tools/ps/fs.ps1 --op read --path test.txt

# Clean up
pwsh -File tools/ps/fs.ps1 --op delete --path test.txt --confirm
```

### Configuration Loading (config.py)
```powershell
# Load example config
python tools/py/config.py --op load --paths fixtures/config/test.env
```

### Caching (cache.py)
```powershell
# Store value
python tools/py/cache.py --op put --key greeting --value '"Hello"'

# Retrieve value
python tools/py/cache.py --op get --key greeting
```

## 3. Run Test Suite

```powershell
# Run all Python tests
python -m unittest discover -s tests/python -p "test_*.py"
```

## 4. Get Help

Every tool has built-in help:

```powershell
# Python tools
python tools/py/config.py --op help
python tools/py/cache.py --op help
python tools/py/template.py --op help

# For detailed examples, see USAGE.md
```

## 5. Run Example Scripts

```powershell
# Config examples
pwsh -File scripts/examples/config-examples.ps1

# Cache examples
pwsh -File scripts/examples/cache-examples.ps1
```

## Next Steps

1. Read [README.md](./README.md) for complete tool overview
2. Check [USAGE.md](./USAGE.md) for detailed examples
3. Review [TODO.md](./TODO.md) for tool specifications
4. Explore test files in `tests/python/` for usage patterns

## Common Patterns for LLMs

**Always check tool health first:**
```python
result = run_tool("--op ping")
if result["ok"]:
    # Tool is available
    actual_result = run_tool("--op operation --args ...")
```

**Always check the `ok` field:**
```python
result = run_tool(...)
if result["ok"]:
    data = result["data"]
else:
    error_msg = result["error"]["message"]
```

**Use --help for examples:**
```bash
python tools/py/<tool>.py --op help
```
