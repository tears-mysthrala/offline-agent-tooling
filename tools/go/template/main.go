package main

import (
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

func renderTemplate(template string, vars map[string]string, safe bool) (string, error) {
	result := template
	
	for key, value := range vars {
		placeholder := "${" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	// Check for unsubstituted variables
	if !safe && strings.Contains(result, "${") {
		start := strings.Index(result, "${")
		end := strings.Index(result[start:], "}")
		if end != -1 {
			missingVar := result[start+2 : start+end]
			return "", fmt.Errorf("missing variable: %s", missingVar)
		}
	}
	
	return result, nil
}

func main() {
	op := flag.String("op", "", "Operation: render|ping|version")
	template := flag.String("template", "", "Template with ${variable} syntax")
	varsJSON := flag.String("vars", "", "JSON object with variable values")
	safe := flag.Bool("safe", false, "Safe mode (keep undefined vars)")
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
			Data: map[string]interface{}{"pong": true, "tool": "template.go"},
		})
		return
	}

	if *op == "version" {
		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"version": "1.0.0", "tool": "template.go"},
		})
		return
	}

	if *op == "render" {
		if *template == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--template required"},
			})
			os.Exit(2)
		}

		vars := make(map[string]string)
		if *varsJSON != "" {
			var rawVars map[string]interface{}
			if err := json.Unmarshal([]byte(*varsJSON), &rawVars); err != nil {
				writeJSON(Response{
					OK:    false,
					Error: &Error{Code: "INVALID_VARS", Message: err.Error()},
				})
				os.Exit(2)
			}
			
			// Convert all values to strings
			for k, v := range rawVars {
				vars[k] = fmt.Sprintf("%v", v)
			}
		}

		rendered, err := renderTemplate(*template, vars, *safe)
		if err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "RENDER_ERROR", Message: err.Error()},
			})
			os.Exit(4)
		}

		if *compact {
			writeJSON(Response{OK: true, Data: rendered})
		} else {
			varKeys := make([]string, 0, len(vars))
			for k := range vars {
				varKeys = append(varKeys, k)
			}
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"rendered":  rendered,
					"vars_used": varKeys,
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
