package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

func main() {
	op := flag.String("op", "", "Operation: read|write|delete|mkdir|checksum|stat|list|ping")
	path := flag.String("path", "", "File or directory path")
	content := flag.String("content", "", "Content to write")
	recursive := flag.Bool("recursive", false, "Recursive operation")
	confirm := flag.Bool("confirm", false, "Confirm destructive operations")
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
			Data: map[string]interface{}{"pong": true, "tool": "fs.go"},
		})
		return
	}

	if *op == "version" {
		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"version": "1.0.0", "tool": "fs.go"},
		})
		return
	}

	switch *op {
	case "read":
		if *path == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--path required"},
			})
			os.Exit(2)
		}

		data, err := ioutil.ReadFile(*path)
		if err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "READ_ERROR", Message: err.Error()},
			})
			os.Exit(3)
		}

		if *compact {
			writeJSON(Response{OK: true, Data: string(data)})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"path":    *path,
					"content": string(data),
					"size":    len(data),
				},
			})
		}

	case "write":
		if *path == "" || *content == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--path and --content required"},
			})
			os.Exit(2)
		}

		if err := ioutil.WriteFile(*path, []byte(*content), 0644); err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "WRITE_ERROR", Message: err.Error()},
			})
			os.Exit(5)
		}

		if *compact {
			writeJSON(Response{OK: true, Data: true})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"path":  *path,
					"bytes": len(*content),
				},
			})
		}

	case "delete":
		if *path == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--path required"},
			})
			os.Exit(2)
		}

		if !*confirm {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "CONFIRM_REQUIRED", Message: "Use --confirm for delete"},
			})
			os.Exit(2)
		}

		if err := os.RemoveAll(*path); err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "DELETE_ERROR", Message: err.Error()},
			})
			os.Exit(5)
		}

		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"deleted": true, "path": *path},
		})

	case "mkdir":
		if *path == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--path required"},
			})
			os.Exit(2)
		}

		if err := os.MkdirAll(*path, 0755); err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "MKDIR_ERROR", Message: err.Error()},
			})
			os.Exit(5)
		}

		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"created": true, "path": *path},
		})

	case "checksum":
		if *path == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--path required"},
			})
			os.Exit(2)
		}

		file, err := os.Open(*path)
		if err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "READ_ERROR", Message: err.Error()},
			})
			os.Exit(3)
		}
		defer file.Close()

		hash := sha256.New()
		if _, err := io.Copy(hash, file); err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "HASH_ERROR", Message: err.Error()},
			})
			os.Exit(5)
		}

		checksum := hex.EncodeToString(hash.Sum(nil))

		if *compact {
			writeJSON(Response{OK: true, Data: checksum})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"path":     *path,
					"checksum": checksum,
					"algorithm": "sha256",
				},
			})
		}

	case "stat":
		if *path == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--path required"},
			})
			os.Exit(2)
		}

		info, err := os.Stat(*path)
		if err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "STAT_ERROR", Message: err.Error()},
			})
			os.Exit(3)
		}

		writeJSON(Response{
			OK: true,
			Data: map[string]interface{}{
				"path":    *path,
				"size":    info.Size(),
				"is_dir":  info.IsDir(),
				"mode":    info.Mode().String(),
				"modtime": info.ModTime().Unix(),
			},
		})

	case "list":
		if *path == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--path required"},
			})
			os.Exit(2)
		}

		var files []string
		if *recursive {
			filepath.Walk(*path, func(p string, info os.FileInfo, err error) error {
				if err == nil {
					files = append(files, p)
				}
				return nil
			})
		} else {
			entries, _ := ioutil.ReadDir(*path)
			for _, entry := range entries {
				files = append(files, filepath.Join(*path, entry.Name()))
			}
		}

		if *compact {
			writeJSON(Response{OK: true, Data: files})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"path":  *path,
					"files": files,
					"count": len(files),
				},
			})
		}

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
