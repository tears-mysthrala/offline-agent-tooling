# Offline Agent Tooling

**Local-first, offline-capable tools for AI agents** - Fast, zero-dependency tools designed for LLM/agent orchestration.

## 🎯 Quick Start for LLMs

**All tools follow a consistent pattern:**
```bash
# PowerShell tools
pwsh -File tools/ps/<tool>.ps1 --op <operation> [args...]

# Python tools
python tools/py/<tool>.py --op <operation> [args...]

# Test availability
--op ping  # Returns {"ok": true, "data": {"pong": true}}
```

## 📋 Available Tools

### Core Tools (MVP)
| Tool | Language | Purpose | Example |
|------|----------|---------|---------|
| **fs.ps1** | PowerShell | File operations | `--op read --path file.txt` |
| **process.ps1** | PowerShell | Run commands | `--op run --cmd "echo test"` |
| **search.ps1** | PowerShell | Grep/search files | `--op grep --root . --pattern "TODO"` |
| **log.ps1** | PowerShell | JSON logging | `--op log --msg "test" --level info` |
| **http_tool.py** | Python | HTTP + offline | `--op get --offline --fixture-key test` |
| **config.py** | Python | Config loader | `--op load --paths config.env` |

### Plus Tools
| Tool | Language | Purpose | Example |
|------|----------|---------|---------|
| **cache.py** | Python | FS cache + TTL | `--op put --key k --value "v" --ttl-s 60` |
| **git.ps1** | PowerShell | Git wrapper | `--op status --repo .` |
| **template.py** | Python | Variable substitution | `--op render --template "Hi ${name}"` |
| **archive.py** | Python | Zip/unzip | `--op zip --source dir --dest out.zip` |
| **kv.py** | Python | SQLite KV store | `--op set --key k --value v` |

## 🚀 Common Patterns

### Check Tool Availability
```bash
python tools/py/config.py --op ping
# Output: {"ok":true,"data":{"pong":true,"tool":"config.py"}}
```

### Error Handling
All tools return JSON with this structure:
```json
{
  "ok": true|false,
  "data": {...},           // if ok=true
  "error": {               // if ok=false
    "code": "ERROR_CODE",
    "message": "details"
  }
}
```

### Offline Mode
Tools with network capabilities support `--offline`:
```bash
python tools/py/http_tool.py --op get --offline --fixture-key example
```

## 📖 Full Documentation

- **[USAGE.md](./USAGE.md)** - Detailed examples for each tool
- **[TODO.md](./TODO.md)** - Tool specifications and roadmap
- **[tests/](./tests/)** - Usage examples in test files

## 🧪 Testing

Run all Python tests:
```powershell
python -m unittest discover -s tests/python -p "test_*.py"
```

Run specific tool tests:
```powershell
python -m unittest tests/python/test_config.py -v
```

## 💡 Design Principles

1. **Local-first**: No network required (offline mode for network tools)
2. **Stdlib-only**: No external dependencies
3. **JSON API**: Consistent input/output format
4. **Fast**: Target <200ms for most operations
5. **Deterministic**: Same input = same output
6. **LLM-friendly**: Clear errors, consistent interface

## 🔧 Tool Standards

Every tool implements:
- ✅ `--op ping` for health check
- ✅ JSON output: `{"ok": bool, "data": {...}}`
- ✅ Error codes with descriptive messages
- ✅ `--trace-id` for request correlation
- ✅ Proper exit codes (0=success, 1=usage, 2=args, 3=notfound, 4+=other)

## 📁 Project Structure

```
tools/
  ps/          # PowerShell tools
  py/          # Python tools
tests/
  python/      # Python unit tests
fixtures/      # Test data
scripts/       # Utility scripts
```

## 🤖 For LLM Agents

**Key points to avoid hallucinations:**

1. **Always use `--op` flag** - Required for all tools except help
2. **Check output.ok** before reading data
3. **Use absolute paths** when possible
4. **Test with `--op ping`** before use
5. **Read error.message** when ok=false

**Example workflow:**
```bash
# 1. Test tool availability
python tools/py/config.py --op ping

# 2. Use tool
python tools/py/config.py --op load --paths config.env

# 3. Parse JSON output
# Check if output["ok"] == true before accessing output["data"]
```

## 📝 License

Apache 2.0 - See [LICENSE](./LICENSE)
