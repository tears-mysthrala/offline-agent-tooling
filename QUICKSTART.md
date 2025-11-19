```

### Caching (cache.exe)
```powershell
# Store value
./bin/cache.exe --op put --key greeting --value '"Hello"'

# Retrieve value
./bin/cache.exe --op get --key greeting
```

## 3. Get Help

Every tool has built-in help:

```powershell
./bin/config.exe --op help
./bin/cache.exe --op help
./bin/template.exe --op help

# For detailed examples, see USAGE.md
```

## Next Steps

1. Read [README.md](./README.md) for complete tool overview
2. Check [USAGE.md](./USAGE.md) for detailed examples
3. Review [TODO.md](./TODO.md) for tool specifications

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
./bin/<tool>.exe --op help
```
