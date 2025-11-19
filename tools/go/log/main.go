package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

type LogEntry struct {
	Timestamp string      `json:"ts"`
	Level     string      `json:"level"`
	Message   string      `json:"msg"`
	Fields    interface{} `json:"fields,omitempty"`
	TraceID   string      `json:"trace_id,omitempty"`
}

func writeJSON(resp Response) {
	json.NewEncoder(os.Stdout).Encode(resp)
}

func getLogPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	// bin/log.exe -> root -> logs/
	rootDir := filepath.Dir(filepath.Dir(ex))
	logsDir := filepath.Join(rootDir, "logs")

	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return "", err
	}

	dateStr := time.Now().Format("20060102")
	return filepath.Join(logsDir, fmt.Sprintf("agent-%s.log", dateStr)), nil
}

func main() {
	op := flag.String("op", "", "Operation: log|ping|version")
	level := flag.String("level", "info", "Log level")
	msg := flag.String("msg", "", "Log message")
	fieldsJson := flag.String("fieldsJson", "", "JSON fields")
	traceId := flag.String("trace_id", "", "Trace ID")
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
			Data: map[string]interface{}{"pong": true, "tool": "log.go"},
		})
		return
	}

	if *op == "version" {
		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"version": "1.0.0", "tool": "log.go"},
		})
		return
	}

	if *op == "log" {
		if *msg == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--msg required"},
			})
			os.Exit(2)
		}

		var fields interface{}
		if *fieldsJson != "" {
			json.Unmarshal([]byte(*fieldsJson), &fields)
		}

		entry := LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     *level,
			Message:   *msg,
			Fields:    fields,
			TraceID:   *traceId,
		}

		logPath, err := getLogPath()
		if err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "FS_ERROR", Message: err.Error()},
			})
			os.Exit(5)
		}

		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "FS_ERROR", Message: err.Error()},
			})
			os.Exit(5)
		}
		defer f.Close()

		line, _ := json.Marshal(entry)
		f.Write(line)
		f.WriteString("\n")

		if *compact {
			writeJSON(Response{OK: true, Data: logPath})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"path":  logPath,
					"entry": entry,
				},
			})
		}
		return
	}

	writeJSON(Response{
		OK:    false,
		Error: &Error{Code: "INVALID_OP", Message: fmt.Sprintf("Unknown operation: %s", *op)},
	})
	os.Exit(1)
}
