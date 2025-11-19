package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func writeJSON(resp Response) {
	json.NewEncoder(os.Stdout).Encode(resp)
}

func getCacheDir() string {
	ex, _ := os.Executable()
	exPath := filepath.Dir(ex)
	// Assuming binary is in bin/, go up one level to root, then .cache/http
	return filepath.Join(exPath, "..", ".cache", "http")
}

func getCachePath(key string) string {
	hash := sha256.Sum256([]byte(key))
	return filepath.Join(getCacheDir(), hex.EncodeToString(hash[:])+".json")
}

type CachedResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	URL        string            `json:"url"`
	Timestamp  int64             `json:"timestamp"`
}

func main() {
	op := flag.String("op", "", "Operation: get|post|put|delete|head|ping|version")
	urlStr := flag.String("url", "", "Target URL")
	body := flag.String("body", "", "Request body")
	headers := flag.String("headers", "", "JSON headers")
	offline := flag.Bool("offline", false, "Offline mode (use fixtures only)")
	fixtureKey := flag.String("fixture-key", "", "Key for offline fixture")
	cache := flag.Bool("cache", false, "Enable caching")
	compact := flag.Bool("compact", false, "Minimal output")
	flag.Parse()

	if *op == "" {
		writeJSON(Response{
			OK:    false,
			Error: &Error{Code: "ARG_MISSING", Message: "--op required"},
		})
		os.Exit(2)
	}

	if *op == "ping" {
		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"pong": true, "tool": "http_tool.go"},
		})
		return
	}

	if *op == "version" {
		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"version": "1.0.0", "tool": "http_tool.go"},
		})
		return
	}

	// Method mapping
	method := strings.ToUpper(*op)
	if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" && method != "HEAD" {
		writeJSON(Response{
			OK:    false,
			Error: &Error{Code: "INVALID_OP", Message: fmt.Sprintf("Unknown operation: %s", *op)},
		})
		os.Exit(1)
	}

	if *urlStr == "" && !*offline {
		writeJSON(Response{
			OK:    false,
			Error: &Error{Code: "ARG_MISSING", Message: "--url required"},
		})
		os.Exit(2)
	}

	// Offline Mode / Fixtures
	if *offline {
		if *fixtureKey == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--fixture-key required in offline mode"},
			})
			os.Exit(2)
		}
		
		// Try to load from cache/fixtures
		// For simplicity, we'll use the cache directory as the fixture source for now
		// In a real scenario, fixtures might be in a separate dir
		cachePath := getCachePath(*fixtureKey)
		data, err := os.ReadFile(cachePath)
		if err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "FIXTURE_NOT_FOUND", Message: fmt.Sprintf("Fixture not found for key: %s", *fixtureKey)},
			})
			os.Exit(1)
		}
		
		var cachedResp CachedResponse
		json.Unmarshal(data, &cachedResp)
		
		if *compact {
			writeJSON(Response{OK: true, Data: cachedResp.Body})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"status":  cachedResp.StatusCode,
					"headers": cachedResp.Headers,
					"body":    cachedResp.Body,
					"cached":  true,
				},
			})
		}
		return
	}

	// Caching check (Read)
	cacheKey := ""
	if *cache && method == "GET" {
		cacheKey = *urlStr
		if *fixtureKey != "" {
			cacheKey = *fixtureKey
		}
		
		cachePath := getCachePath(cacheKey)
		data, err := os.ReadFile(cachePath)
		if err == nil {
			var cachedResp CachedResponse
			if json.Unmarshal(data, &cachedResp) == nil {
				// Cache hit
				if *compact {
					writeJSON(Response{OK: true, Data: cachedResp.Body})
				} else {
					writeJSON(Response{
						OK: true,
						Data: map[string]interface{}{
							"status":  cachedResp.StatusCode,
							"headers": cachedResp.Headers,
							"body":    cachedResp.Body,
							"cached":  true,
						},
					})
				}
				return
			}
		}
	}

	// Make Request
	req, err := http.NewRequest(method, *urlStr, strings.NewReader(*body))
	if err != nil {
		writeJSON(Response{
			OK:    false,
			Error: &Error{Code: "REQ_ERROR", Message: err.Error()},
		})
		os.Exit(5)
	}

	// Add headers
	if *headers != "" {
		var headerMap map[string]string
		if json.Unmarshal([]byte(*headers), &headerMap) == nil {
			for k, v := range headerMap {
				req.Header.Set(k, v)
			}
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		writeJSON(Response{
			OK:    false,
			Error: &Error{Code: "NET_ERROR", Message: err.Error()},
		})
		os.Exit(5)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		writeJSON(Response{
			OK:    false,
			Error: &Error{Code: "READ_ERROR", Message: err.Error()},
		})
		os.Exit(5)
	}

	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		respHeaders[k] = v[0]
	}

	respStr := string(respBody)

	// Cache Write
	if *cache && method == "GET" {
		os.MkdirAll(getCacheDir(), 0755)
		cachedResp := CachedResponse{
			StatusCode: resp.StatusCode,
			Headers:    respHeaders,
			Body:       respStr,
			URL:        *urlStr,
			Timestamp:  time.Now().Unix(),
		}
		
		data, _ := json.Marshal(cachedResp)
		os.WriteFile(getCachePath(cacheKey), data, 0644)
	}

	if *compact {
		writeJSON(Response{OK: true, Data: respStr})
	} else {
		writeJSON(Response{
			OK: true,
			Data: map[string]interface{}{
				"status":  resp.StatusCode,
				"headers": respHeaders,
				"body":    respStr,
			},
		})
	}
}
