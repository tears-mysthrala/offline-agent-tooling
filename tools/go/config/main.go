package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
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

// parseEnvFile parses a .env file into a map
func parseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	envMap := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		envMap[key] = value
	}
	return envMap, scanner.Err()
}

// redactSecrets hides sensitive values
func redactSecrets(config map[string]interface{}) map[string]interface{} {
	redacted := make(map[string]interface{})
	secretKeywords := []string{"key", "secret", "password", "token", "auth", "credential"}

	for k, v := range config {
		keyLower := strings.ToLower(k)
		isSecret := false
		for _, keyword := range secretKeywords {
			if strings.Contains(keyLower, keyword) {
				isSecret = true
				break
			}
		}

		if isSecret {
			redacted[k] = "REDACTED"
		} else {
			redacted[k] = v
		}
	}
	return redacted
}

func main() {
	op := flag.String("op", "", "Operation: load|get|ping|version")
	paths := flag.String("paths", "", "Comma-separated paths to .env or .json files")
	prefix := flag.String("prefix", "", "Filter env vars by prefix")
	key := flag.String("key", "", "Key to get")
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
			Data: map[string]interface{}{"pong": true, "tool": "config.go"},
		})
		return
	}

	if *op == "version" {
		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"version": "1.0.0", "tool": "config.go"},
		})
		return
	}

	// Load configuration
	config := make(map[string]interface{})

	// 1. Load from files
	if *paths != "" {
		filePaths := strings.Split(*paths, ",")
		for _, path := range filePaths {
			path = strings.TrimSpace(path)
			if strings.HasSuffix(path, ".json") {
				data, err := os.ReadFile(path)
				if err == nil {
					var jsonConfig map[string]interface{}
					if json.Unmarshal(data, &jsonConfig) == nil {
						for k, v := range jsonConfig {
							config[k] = v
						}
					}
				}
			} else {
				// Assume .env format
				envMap, err := parseEnvFile(path)
				if err == nil {
					for k, v := range envMap {
						config[k] = v
					}
				}
			}
		}
	}

	// 2. Load from environment variables (if prefix provided)
	if *prefix != "" {
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 && strings.HasPrefix(parts[0], *prefix) {
				config[parts[0]] = parts[1]
			}
		}
	}

	switch *op {
	case "load":
		if *compact {
			writeJSON(Response{OK: true, Data: config})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"config": redactSecrets(config),
					"count":  len(config),
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

		val, ok := config[*key]
		if !ok {
			// Try looking up in OS env directly if not found in loaded config
			if envVal, exists := os.LookupEnv(*key); exists {
				val = envVal
				ok = true
			}
		}

		if !ok {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "KEY_NOT_FOUND", Message: fmt.Sprintf("Key not found: %s", *key)},
			})
			os.Exit(1)
		}

		if *compact {
			writeJSON(Response{OK: true, Data: val})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"key":   *key,
					"value": val,
				},
			})
		}

	default:
		writeJSON(Response{
			OK:    false,
			Error: &Error{Code: "INVALID_OP", Message: fmt.Sprintf("Unknown operation: %s", *op)},
		})
		os.Exit(1)
	}
}
