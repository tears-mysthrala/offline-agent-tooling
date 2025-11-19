package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
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

func main() {
	op := flag.String("op", "", "Operation: run|ping|version")
	cmdStr := flag.String("cmd", "", "Command to run")
	cwd := flag.String("cwd", "", "Working directory")
	timeoutMs := flag.Int("timeout_ms", 0, "Timeout in milliseconds")
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
			Data: map[string]interface{}{"pong": true, "tool": "process.go"},
		})
		return
	}

	if *op == "version" {
		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"version": "1.0.0", "tool": "process.go"},
		})
		return
	}

	if *op == "run" {
		if *cmdStr == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--cmd required"},
			})
			os.Exit(2)
		}

		// Use pwsh to run the command to maintain compatibility with ps1 version behavior
		cmdArgs := []string{"-NoProfile", "-Command", *cmdStr}
		cmd := exec.Command("pwsh", cmdArgs...)

		if *cwd != "" {
			cmd.Dir = *cwd
		}

		start := time.Now()
		var out []byte
		var err error

		if *timeoutMs > 0 {
			// Command with timeout
			done := make(chan error, 1)
			go func() {
				out, err = cmd.CombinedOutput()
				done <- err
			}()

			select {
			case <-time.After(time.Duration(*timeoutMs) * time.Millisecond):
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
				writeJSON(Response{
					OK: false,
					Error: &Error{
						Code:    "TIMEOUT",
						Message: fmt.Sprintf("Command timed out after %dms", *timeoutMs),
					},
				})
				os.Exit(1)
			case <-done:
				// Completed within timeout
			}
		} else {
			// Command without timeout
			out, err = cmd.CombinedOutput()
		}

		duration := time.Since(start).Milliseconds()
		exitCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				exitCode = 1
			}
		}

		outputStr := string(out)

		if *compact {
			// Truncate for compact mode
			if len(outputStr) > 200 {
				outputStr = outputStr[:200] + "..."
			}
			writeJSON(Response{OK: exitCode == 0, Data: outputStr})
		} else {
			writeJSON(Response{
				OK: exitCode == 0,
				Data: map[string]interface{}{
					"exit_code":   exitCode,
					"stdout":      outputStr,
					"stderr":      "", // CombinedOutput puts everything in stdout
					"duration_ms": duration,
				},
			})
		}

		if exitCode != 0 {
			os.Exit(exitCode)
		}
		return
	}

	writeJSON(Response{
		OK:    false,
		Error: &Error{Code: "INVALID_OP", Message: fmt.Sprintf("Unknown operation: %s", *op)},
	})
	os.Exit(1)
}
