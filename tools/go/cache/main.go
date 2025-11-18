package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type Response struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error *Error      `json:"error,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CacheEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	CreatedAt int64       `json:"created_at"`
	ExpiresAt *int64      `json:"expires_at,omitempty"`
}

func writeJSON(resp Response) {
	json.NewEncoder(os.Stdout).Encode(resp)
}

func getCacheDir() string {
	ex, _ := os.Executable()
	exPath := filepath.Dir(ex)
	if stat, err := os.Stat("../../.cache/cache"); err == nil && stat.IsDir() {
		return "../../.cache/cache"
	}
	return filepath.Join(exPath, "..", "..", ".cache", "cache")
}

func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func getCachePath(key string) string {
	return filepath.Join(getCacheDir(), hashKey(key)+".json")
}

func main() {
	op := flag.String("op", "", "Operation: put|get|del|stats|clear-expired|clear-all|ping")
	key := flag.String("key", "", "Cache key")
	value := flag.String("value", "", "Value (JSON string)")
	ttl := flag.Int("ttl-s", 0, "Time-to-live in seconds")
	compact := flag.Bool("compact", false, "Minimal output")
	flag.Parse()

	if *op == "" {
		writeJSON(Response{
			OK: false,
			Error: &Error{Code: "ARG_MISSING", Message: "--op required"},
		})
		os.Exit(2)
	}

	if *op == "ping" {
		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"pong": true, "tool": "cache.go"},
		})
		return
	}

	if *op == "version" {
		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"version": "1.0.0", "tool": "cache.go"},
		})
		return
	}

	os.MkdirAll(getCacheDir(), 0755)

	switch *op {
	case "put":
		if *key == "" || *value == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--key and --value required"},
			})
			os.Exit(2)
		}

		var val interface{}
		if err := json.Unmarshal([]byte(*value), &val); err != nil {
			val = *value
		}

		now := time.Now().Unix()
		var expiresAt *int64
		if *ttl > 0 {
			exp := now + int64(*ttl)
			expiresAt = &exp
		}

		entry := CacheEntry{
			Key:       *key,
			Value:     val,
			CreatedAt: now,
			ExpiresAt: expiresAt,
		}

		data, _ := json.Marshal(entry)
		if err := ioutil.WriteFile(getCachePath(*key), data, 0644); err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "CACHE_ERROR", Message: err.Error()},
			})
			os.Exit(5)
		}

		if *compact {
			writeJSON(Response{OK: true, Data: true})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"key":    *key,
					"cached": true,
					"ttl_s":  *ttl,
				},
			})
		}

	case "get":
		if *key == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--key required"},
			})
			os.Exit(2)
		}

		cachePath := getCachePath(*key)
		data, err := ioutil.ReadFile(cachePath)
		if os.IsNotExist(err) {
			if *compact {
				writeJSON(Response{OK: true, Data: nil})
			} else {
				writeJSON(Response{
					OK:   true,
					Data: map[string]interface{}{"key": *key, "found": false, "value": nil},
				})
			}
			return
		}

		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			os.Remove(cachePath)
			if *compact {
				writeJSON(Response{OK: true, Data: nil})
			} else {
				writeJSON(Response{
					OK:   true,
					Data: map[string]interface{}{"key": *key, "found": false, "value": nil},
				})
			}
			return
		}

		if entry.ExpiresAt != nil && *entry.ExpiresAt < time.Now().Unix() {
			os.Remove(cachePath)
			if *compact {
				writeJSON(Response{OK: true, Data: nil})
			} else {
				writeJSON(Response{
					OK:   true,
					Data: map[string]interface{}{"key": *key, "found": false, "value": nil},
				})
			}
			return
		}

		if *compact {
			writeJSON(Response{OK: true, Data: entry.Value})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"key":   *key,
					"found": true,
					"value": entry.Value,
				},
			})
		}

	case "del":
		if *key == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--key required"},
			})
			os.Exit(2)
		}

		os.Remove(getCachePath(*key))
		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"key": *key, "deleted": true},
		})

	case "stats":
		files, _ := filepath.Glob(filepath.Join(getCacheDir(), "*.json"))
		totalSize := int64(0)
		validCount := 0
		expiredCount := 0
		now := time.Now().Unix()

		for _, file := range files {
			info, _ := os.Stat(file)
			totalSize += info.Size()

			data, _ := ioutil.ReadFile(file)
			var entry CacheEntry
			if json.Unmarshal(data, &entry) == nil {
				if entry.ExpiresAt != nil && *entry.ExpiresAt < now {
					expiredCount++
				} else {
					validCount++
				}
			}
		}

		writeJSON(Response{
			OK: true,
			Data: map[string]interface{}{
				"total_entries":    len(files),
				"valid_entries":    validCount,
				"expired_entries":  expiredCount,
				"total_size_bytes": totalSize,
				"cache_dir":        getCacheDir(),
			},
		})

	case "clear-expired":
		files, _ := filepath.Glob(filepath.Join(getCacheDir(), "*.json"))
		removed := 0
		now := time.Now().Unix()

		for _, file := range files {
			data, _ := ioutil.ReadFile(file)
			var entry CacheEntry
			if json.Unmarshal(data, &entry) == nil {
				if entry.ExpiresAt != nil && *entry.ExpiresAt < now {
					os.Remove(file)
					removed++
				}
			}
		}

		writeJSON(Response{
			OK:   true,
			Data: map[string]int{"removed": removed},
		})

	case "clear-all":
		files, _ := filepath.Glob(filepath.Join(getCacheDir(), "*.json"))
		for _, file := range files {
			os.Remove(file)
		}

		writeJSON(Response{
			OK:   true,
			Data: map[string]int{"removed": len(files)},
		})

	default:
		writeJSON(Response{
			OK: false,
			Error: &Error{
				Code:    "INVALID_OP",
				Message: fmt.Sprintf("Unknown operation: %s", *op),
			},
		})
		os.Exit(1)
	}
}
