
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
   ./bin/config.exe --paths config.env
   
   # ✅ CORRECT
   ./bin/config.exe --op load --paths config.env
   ```

2. **Not checking `ok` field**
   ```json
   // ❌ WRONG
   // Assuming data exists without checking ok
   
   // ✅ CORRECT
   // Check if "ok": true before accessing "data"
   ```

3. **Wrong path separators**
   ```bash
   # Use forward slashes or double backslashes on Windows
   ./bin/fs.exe --op read --path "d:/tooling/file.txt"
   ```

4. **Forgetting `--confirm` for destructive operations**
   ```bash
   ./bin/fs.exe --op delete --path "file.txt" --confirm
   ```

5. **Not testing tool availability first**
   ```bash
   # Always ping first
   ./bin/config.exe --op ping
   # Then use
   ./bin/config.exe --op load --paths config.env
   ```
