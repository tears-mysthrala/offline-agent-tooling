# Tool Usage Guide

**Detailed examples for every tool.** Copy-paste ready commands for LLM agents.

## Table of Contents
- [File System (fs.ps1)](#file-system-fsps1)
- [Process Runner (process.ps1)](#process-runner-processps1)
- [Search (search.ps1)](#search-searchps1)
- [Logging (log.ps1)](#logging-logps1)
- [HTTP/Mock (http_tool.py)](#httpmock-http_toolpy)
- [Config (config.py)](#config-configpy)
- [Cache (cache.py)](#cache-cachepy)
- [Git (git.ps1)](#git-gitps1)
- [Template (template.py)](#template-templatepy)
- [Archive (archive.py)](#archive-archivepy)
- [KV Store (kv.py)](#kv-store-kvpy)

---

## File System (fs.ps1)

**Location**: `tools/ps/fs.ps1`

### Read File
```powershell
pwsh -File tools/ps/fs.ps1 --op read --path "myfile.txt"
```

### Write File
```powershell
pwsh -File tools/ps/fs.ps1 --op write --path "output.txt" --content "Hello World"
```

### Append to File
```powershell
pwsh -File tools/ps/fs.ps1 --op append --path "log.txt" --content "New line"
```

### Create Directory
```powershell
pwsh -File tools/ps/fs.ps1 --op mkdir --path "newdir"
```

### Delete (requires --confirm)
```powershell
pwsh -File tools/ps/fs.ps1 --op delete --path "file.txt" --confirm
```

### Get File Stats
```powershell
pwsh -File tools/ps/fs.ps1 --op stat --path "file.txt"
```

### Calculate Checksum (SHA-256)
```powershell
pwsh -File tools/ps/fs.ps1 --op checksum --path "file.txt"
```

### List Files
```powershell
pwsh -File tools/ps/fs.ps1 --op list --path "." --recursive
```

### Glob Pattern Match
```powershell
pwsh -File tools/ps/fs.ps1 --op glob --path "." --pattern "*.txt" --recursive
```

### Read/Write JSON
```powershell
# Write JSON
pwsh -File tools/ps/fs.ps1 --op write-json --path "data.json" --content '{"key":"value"}'

# Read JSON
pwsh -File tools/ps/fs.ps1 --op read-json --path "data.json"
```

---

## Process Runner (process.ps1)

**Location**: `tools/ps/process.ps1`

### Run Command
```powershell
pwsh -File tools/ps/process.ps1 --op run --cmd "echo Hello"
```

### Run with Timeout
```powershell
pwsh -File tools/ps/process.ps1 --op run --cmd "long-running-task" --timeout_ms 5000
```

### Run in Specific Directory
```powershell
pwsh -File tools/ps/process.ps1 --op run --cmd "git status" --cwd "d:\myrepo"
```

---

## Search (search.ps1)

**Location**: `tools/ps/search.ps1`

### Grep for Text
```powershell
pwsh -File tools/ps/search.ps1 --op grep --root "." --pattern "TODO"
```

### Recursive Search
```powershell
pwsh -File tools/ps/search.ps1 --op grep --root "." --pattern "FIXME" --recursive
```

### Regex Search
```powershell
pwsh -File tools/ps/search.ps1 --op grep --root "." --pattern "function\s+\w+" --is_regex
```

### Include/Exclude Patterns
```powershell
pwsh -File tools/ps/search.ps1 --op grep --root "." --pattern "test" --include "*.py" --exclude "*.pyc"
```

---

## Logging (log.ps1)

**Location**: `tools/ps/log.ps1`

### Log Message
```powershell
pwsh -File tools/ps/log.ps1 --op log --msg "Operation completed" --level info
```

### Log with Fields
```powershell
pwsh -File tools/ps/log.ps1 --op log --msg "User login" --level info --fieldsJson '{"user":"alice"}'
```

### Log Levels
```powershell
# info, warn, error
pwsh -File tools/ps/log.ps1 --op log --msg "Warning message" --level warn
```

---

## HTTP/Mock (http_tool.py)

**Location**: `tools/py/http_tool.py`

### Online GET Request
```bash
python tools/py/http_tool.py --op get --url "https://api.example.com/data"
```

### Offline Mode (uses fixtures)
```bash
python tools/py/http_tool.py --op get --offline --fixture-key example-users
```

### With Cache
```bash
python tools/py/http_tool.py --op get --url "https://api.example.com/data" --cache
```

### With Timeout
```bash
python tools/py/http_tool.py --op get --url "https://api.example.com/data" --timeout-ms 3000
```

---

## Config (config.py)

**Location**: `tools/py/config.py`

### Load .env File
```bash
python tools/py/config.py --op load --paths "config.env"
```

### Load Multiple Files
```bash
python tools/py/config.py --op load --paths "base.env,prod.json"
```

### Load with Environment Variables
```bash
python tools/py/config.py --op load --paths "config.env" --env-prefix "APP_"
```

### Load with Overrides
```bash
python tools/py/config.py --op load --paths "config.env" --overrides '{"PORT":"8080"}'
```

**Note**: Secret values (containing 'password', 'secret', 'token', 'api_key') are auto-redacted in output.

---

## Cache (cache.py)

**Location**: `tools/py/cache.py`

### Put Value in Cache
```bash
python tools/py/cache.py --op put --key "mykey" --value '"myvalue"'
```

### Put with TTL (60 seconds)
```bash
python tools/py/cache.py --op put --key "session" --value '"data"' --ttl-s 60
```

### Get Value from Cache
```bash
python tools/py/cache.py --op get --key "mykey"
```

### Delete Value
```bash
python tools/py/cache.py --op del --key "mykey"
```

### Get Cache Statistics
```bash
python tools/py/cache.py --op stats
```

### Clear Expired Entries
```bash
python tools/py/cache.py --op clear-expired
```

### Clear All Cache
```bash
python tools/py/cache.py --op clear-all
```

---

## Git (git.ps1)

**Location**: `tools/ps/git.ps1`

### Get Status
```powershell
pwsh -File tools/ps/git.ps1 --op status --repo "."
```

### Get Log (last 10 commits)
```powershell
pwsh -File tools/ps/git.ps1 --op log --repo "." --limit 10
```

### Get Diff
```powershell
pwsh -File tools/ps/git.ps1 --op diff --repo "."
```

### Get Diff for Specific File
```powershell
pwsh -File tools/ps/git.ps1 --op diff --repo "." --file "README.md"
```

### Get Branch Info
```powershell
pwsh -File tools/ps/git.ps1 --op branch --repo "."
```

---

## Template (template.py)

**Location**: `tools/py/template.py`

### Simple Substitution
```bash
python tools/py/template.py --op render --template "Hello ${name}!" --vars '{"name":"World"}'
```

### Multiple Variables
```bash
python tools/py/template.py --op render --template "User: ${user}, Age: ${age}" --vars '{"user":"Alice","age":"30"}'
```

### Safe Substitution (leaves missing vars)
```bash
python tools/py/template.py --op render --template "Hello ${name}, ${missing}" --vars '{"name":"Bob"}' --safe
# Output: "Hello Bob, ${missing}"
```

---

## Archive (archive.py)

**Location**: `tools/py/archive.py`

### Create Zip Archive
```bash
python tools/py/archive.py --op zip --source "mydir" --dest "archive.zip"
```

### Extract Zip Archive
```bash
python tools/py/archive.py --op unzip --source "archive.zip" --dest "output/"
```

### List Archive Contents
```bash
python tools/py/archive.py --op list --source "archive.zip"
```

---

## KV Store (kv.py)

**Location**: `tools/py/kv.py`

### Set Value
```bash
python tools/py/kv.py --op set --key "username" --value "alice"
```

### Set with Namespace
```bash
python tools/py/kv.py --op set --ns "session" --key "token" --value "abc123"
```

### Set with TTL (expires in 60 seconds)
```bash
python tools/py/kv.py --op set --key "temp" --value "data" --ttl 60
```

### Get Value
```bash
python tools/py/kv.py --op get --key "username"
```

### Delete Value
```bash
python tools/py/kv.py --op del --key "username"
```

### List All Keys in Namespace
```bash
python tools/py/kv.py --op keys --ns "default"
```

### Purge Expired Entries
```bash
python tools/py/kv.py --op purge-expired
```

---

## 🚨 Common Mistakes to Avoid (for LLMs)

1. **Missing `--op` flag** - Always required!
   ```bash
   # ❌ WRONG
   python tools/py/config.py --paths config.env
   
   # ✅ CORRECT
   python tools/py/config.py --op load --paths config.env
   ```

2. **Not checking `ok` field**
   ```python
   # ❌ WRONG
   result = run_tool()
   data = result["data"]  # Will crash if ok=false
   
   # ✅ CORRECT
   result = run_tool()
   if result["ok"]:
       data = result["data"]
   else:
       print(result["error"]["message"])
   ```

3. **Wrong path separators**
   ```bash
   # Use forward slashes or double backslashes on Windows
   pwsh -File tools/ps/fs.ps1 --op read --path "d:/tooling/file.txt"
   ```

4. **Forgetting `--confirm` for destructive operations**
   ```bash
   pwsh -File tools/ps/fs.ps1 --op delete --path "file.txt" --confirm
   ```

5. **Not testing tool availability first**
   ```bash
   # Always ping first
   python tools/py/config.py --op ping
   # Then use
   python tools/py/config.py --op load --paths config.env
   ```
