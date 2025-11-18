package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
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

func getDBPath() string {
	ex, _ := os.Executable()
	exPath := filepath.Dir(ex)
	// For development, use relative path
	if stat, err := os.Stat("../../.data"); err == nil && stat.IsDir() {
		return "../../.data/kv.sqlite"
	}
	return filepath.Join(exPath, "..", "..", ".data", "kv.sqlite")
}

func initDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS kv (
		ns TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT,
		expires_at INTEGER,
		PRIMARY KEY (ns, key)
	);`
	
	_, err := db.Exec(schema)
	if err != nil {
		return err
	}
	
	// Enable WAL mode for better concurrency
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout=5000")
	
	return nil
}

func main() {
	op := flag.String("op", "", "Operation: set|get|del|keys|purge-expired|ping")
	ns := flag.String("ns", "default", "Namespace")
	key := flag.String("key", "", "Key")
	value := flag.String("value", "", "Value")
	ttl := flag.Int("ttl", 0, "Time-to-live in seconds")
	compact := flag.Bool("compact", false, "Minimal output")
	flag.Parse()

	if *op == "" {
		writeJSON(Response{
			OK: false,
			Error: &Error{
				Code:    "ARG_MISSING",
				Message: "--op required",
			},
		})
		os.Exit(2)
	}

	if *op == "ping" {
		writeJSON(Response{
			OK: true,
			Data: map[string]interface{}{
				"pong": true,
				"tool": "kv.go",
			},
		})
		return
	}

	if *op == "version" {
		writeJSON(Response{
			OK: true,
			Data: map[string]interface{}{
				"version": "1.0.0",
				"tool":    "kv.go",
			},
		})
		return
	}

	// Create .data directory if not exists
	dataDir := filepath.Dir(getDBPath())
	os.MkdirAll(dataDir, 0755)

	db, err := sql.Open("sqlite3", getDBPath())
	if err != nil {
		writeJSON(Response{
			OK: false,
			Error: &Error{
				Code:    "DB_ERROR",
				Message: err.Error(),
			},
		})
		os.Exit(5)
	}
	defer db.Close()

	if err := initDB(db); err != nil {
		writeJSON(Response{
			OK: false,
			Error: &Error{
				Code:    "DB_INIT_ERROR",
				Message: err.Error(),
			},
		})
		os.Exit(5)
	}

	now := time.Now().Unix()

	switch *op {
	case "set":
		if *key == "" || *value == "" {
			writeJSON(Response{
				OK: false,
				Error: &Error{
					Code:    "ARG_MISSING",
					Message: "--key and --value required",
				},
			})
			os.Exit(2)
		}

		var expiresAt *int64
		if *ttl > 0 {
			exp := now + int64(*ttl)
			expiresAt = &exp
		}

		_, err := db.Exec(
			"REPLACE INTO kv (ns, key, value, expires_at) VALUES (?, ?, ?, ?)",
			*ns, *key, *value, expiresAt,
		)
		if err != nil {
			writeJSON(Response{
				OK: false,
				Error: &Error{
					Code:    "DB_ERROR",
					Message: err.Error(),
				},
			})
			os.Exit(5)
		}

		if *compact {
			writeJSON(Response{OK: true, Data: map[string]string{"key": *key}})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]string{
					"ns":  *ns,
					"key": *key,
				},
			})
		}

	case "get":
		if *key == "" {
			writeJSON(Response{
				OK: false,
				Error: &Error{
					Code:    "ARG_MISSING",
					Message: "--key required",
				},
			})
			os.Exit(2)
		}

		var val string
		var expiresAt *int64
		err := db.QueryRow(
			"SELECT value, expires_at FROM kv WHERE ns=? AND key=?",
			*ns, *key,
		).Scan(&val, &expiresAt)

		if err == sql.ErrNoRows {
			if *compact {
				writeJSON(Response{OK: true, Data: nil})
			} else {
				writeJSON(Response{
					OK: true,
					Data: map[string]interface{}{
						"found": false,
					},
				})
			}
			return
		}

		if err != nil {
			writeJSON(Response{
				OK: false,
				Error: &Error{
					Code:    "DB_ERROR",
					Message: err.Error(),
				},
			})
			os.Exit(5)
		}

		// Check expiration
		if expiresAt != nil && *expiresAt < now {
			db.Exec("DELETE FROM kv WHERE ns=? AND key=?", *ns, *key)
			if *compact {
				writeJSON(Response{OK: true, Data: nil})
			} else {
				writeJSON(Response{
					OK: true,
					Data: map[string]interface{}{
						"found":   false,
						"expired": true,
					},
				})
			}
			return
		}

		if *compact {
			writeJSON(Response{OK: true, Data: val})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"found": true,
					"value": val,
				},
			})
		}

	case "del":
		if *key == "" {
			writeJSON(Response{
				OK: false,
				Error: &Error{
					Code:    "ARG_MISSING",
					Message: "--key required",
				},
			})
			os.Exit(2)
		}

		_, err := db.Exec("DELETE FROM kv WHERE ns=? AND key=?", *ns, *key)
		if err != nil {
			writeJSON(Response{
				OK: false,
				Error: &Error{
					Code:    "DB_ERROR",
					Message: err.Error(),
				},
			})
			os.Exit(5)
		}

		writeJSON(Response{
			OK: true,
			Data: map[string]bool{
				"deleted": true,
			},
		})

	case "keys":
		rows, err := db.Query("SELECT key, expires_at FROM kv WHERE ns=?", *ns)
		if err != nil {
			writeJSON(Response{
				OK: false,
				Error: &Error{
					Code:    "DB_ERROR",
					Message: err.Error(),
				},
			})
			os.Exit(5)
		}
		defer rows.Close()

		var keys []string
		for rows.Next() {
			var k string
			var exp *int64
			rows.Scan(&k, &exp)
			
			if exp != nil && *exp < now {
				db.Exec("DELETE FROM kv WHERE ns=? AND key=?", *ns, k)
				continue
			}
			keys = append(keys, k)
		}

		if *compact {
			writeJSON(Response{OK: true, Data: keys})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"keys": keys,
				},
			})
		}

	case "purge-expired":
		result, err := db.Exec(
			"DELETE FROM kv WHERE expires_at IS NOT NULL AND expires_at<?",
			now,
		)
		if err != nil {
			writeJSON(Response{
				OK: false,
				Error: &Error{
					Code:    "DB_ERROR",
					Message: err.Error(),
				},
			})
			os.Exit(5)
		}

		purged, _ := result.RowsAffected()
		writeJSON(Response{
			OK: true,
			Data: map[string]int64{
				"purged": purged,
			},
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
