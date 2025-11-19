# Offline Agent Tooling

**Local-first, offline-capable tools for AI agents** - Blazing fast, zero-dependency tools optimized for LLM/agent orchestration.

## 🎯 Quick Start for LLMs

**All tools use consistent patterns:**
```bash
# Go tools (FASTEST - use these for performance)
./bin/kv.exe --op <operation> [args...]
./bin/cache.exe --op <operation> [args...]
./bin/fs.exe --op <operation> [args...]

# Python tools (complex logic)
python tools/py/<tool>.py --op <operation> [args...]

# PowerShell tools (Windows-specific)
pwsh -File tools/ps/<tool>.ps1 --op <operation> [args...]

# Test availability
--op ping  # Returns {"ok": true, "data": {"pong": true}}
```

## ⚡ Performance-Optimized Tools

### Go Tools (Performance-Optimized)
| Tool | Speed | Purpose | Example |
|------|-------|---------|---------|
| **kv.exe** | 12ms | Key-value store | `./bin/kv.exe --op get --key k --compact` |
| **cache.exe** | 13ms | FS cache + TTL | `./bin/cache.exe --op get --key k --compact` |
| **fs.exe** | 13ms | File operations | `./bin/fs.exe --op read --path file.txt --compact` |
| **config.exe** | 15ms | Config loader | `./bin/config.exe --op load --paths config.env --compact` |
| **template.exe** | 32ms | Templates | `./bin/template.exe --op render --template "Hi ${name}" --compact` |
| **archive.exe** | 180ms | Zip/unzip | `./bin/archive.exe --op zip --source dir --dest out.zip --compact` |
| **http_tool.exe** | 40ms | HTTP + offline | `./bin/http_tool.exe --op get --url https://example.com --compact` |

### PowerShell Tools (Windows Native)
| Tool | Speed | Purpose | Example |
|------|-------|---------|---------|
| **git.ps1** | 750ms | Git wrapper | `pwsh -File tools/ps/git.ps1 --op status` |
| **fs.ps1** | 650ms | File ops (legacy) | `pwsh -File tools/ps/fs.ps1 --op read --path file.txt` |
| **search.ps1** | 700ms | Grep | `pwsh -File tools/ps/search.ps1 --op grep --root . --pattern TODO` |
| **log.ps1** | 680ms | JSON logging | `pwsh -File tools/ps/log.ps1 --op log --msg "test"` |
| **process.ps1** | 690ms | Run commands | `pwsh -File tools/ps/process.ps1 --op run --cmd "echo test"` |

## 🚀 Token Optimization

All tools support `--compact` mode for **30-60% token reduction**:

**Normal output** (60 bytes):
```json
{"ok":true,"data":{"rendered":"Hello!","vars_used":["name"]}}
```

**Compact output** (35 bytes):
```json
{"ok":true,"data":"Hello!"}
```

## 📊 Performance Comparison

### Go vs Python (3x faster on average)
- **KV Store**: 32ms (Go) vs 84ms (Python) = **2.6x faster**
- **Cache**: 28ms (Go) vs 96ms (Python) = **3.4x faster**
- **FS Ops**: 25ms (Go) vs 75ms (Python) = **3.0x faster**

### Token Savings
For 100 operations/day:
- **Without --compact**: ~2000 tokens
- **With --compact**: ~1200 tokens
- **With Go + --compact**: ~800 tokens
- **Total savings**: **60% fewer tokens**

## 🎨 Common Patterns

### Check Tool Availability
```bash
./bin/kv.exe --op ping
# Output: {"ok":true,"data":{"pong":true,"tool":"kv.go"}}
```

### Use Compact Mode (for LLMs)
```bash
./bin/kv.exe --op get --key mykey --compact
# Output: {"ok":true,"data":"myvalue"}
```

### Error Handling
All tools return consistent JSON:
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

## 📖 Full Documentation

- **[USAGE.md](./USAGE.md)** - Detailed examples for each tool
- **[QUICKSTART.md](./QUICKSTART.md)** - 5-minute getting started guide
- **[TODO.md](./TODO.md)** - Tool specifications and roadmap
- **[tests/](./tests/)** - Usage examples in test files

## 🧪 Testing

Run all Python tests:
```powershell
python -m unittest discover -s tests/python -p "test_*.py"
```

Test Go tools:
```powershell
./bin/kv.exe --op ping
./bin/cache.exe --op ping
./bin/fs.exe --op ping
```

## 💡 Design Principles

1. **Local-first**: No network required (offline mode for network tools)
2. **Stdlib-only**: No external dependencies (Python tools)
3. **Compiled binaries**: Single-file Go executables (kv, cache, fs)
4. **JSON API**: Consistent input/output format
5. **Fast**: Target <30ms for Go tools, <100ms for Python
6. **Deterministic**: Same input = same output
7. **LLM-friendly**: Clear errors, consistent interface, --compact mode

## 🏗️ Hybrid Architecture

**Go**: Performance-critical (All core tools)
**PowerShell**: Windows-specific (git, legacy tools)

This gives us maximum performance with native Windows integration.

## 🔧 Tool Standards

Every tool implements:
- ✅ `--op ping` for health check
- ✅ `--compact` for minimal output
- ✅ JSON output: `{"ok": bool, "data": {...}}`
- ✅ Error codes with descriptive messages
- ✅ Proper exit codes (0=success, 1=usage, 2=args, 3=notfound, etc.)

## 📁 Project Structure

```
tools/
  go/           # Go tools (compiled to bin/)
    kv/         # Key-value store
    cache/      # Filesystem cache
    fs/         # File operations
    config/     # Config loader
    template/   # Template engine
    http_tool/  # HTTP client
    archive/    # Archive tool
  ps/           # PowerShell tools
    git.ps1     # Git wrapper
    fs.ps1      # File ops (legacy)
bin/            # Compiled Go binaries
  kv.exe
  cache.exe
  fs.exe
  config.exe
  template.exe
  http_tool.exe
  archive.exe
```

## 🤖 For LLM Agents

**Key points to avoid hallucinations:**

1. **Use Go tools for performance** - 3x faster, saves tokens
2. **Always use `--compact`** - 40-60% fewer tokens
3. **Always use `--op` flag** - Required for all tools
4. **Check output.ok** before reading data
5. **Use absolute paths** when possible
6. **Test with `--op ping`** before use

**Example workflow:**
```bash
# 1. Test tool (Go is fastest)
./bin/kv.exe --op ping

# 2. Use tool with compact mode
./bin/kv.exe --op get --key data --compact

# 3. Parse JSON output
# Check if output["ok"] == true before accessing output["data"]
```

## 📝 License

Apache 2.0 - See [LICENSE](./LICENSE)
