package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

func runGit(args []string, cwd string) (string, error) {
	cmd := exec.Command("git", args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

func main() {
	op := flag.String("op", "", "Operation: status|log|diff|branch|ping|version")
	repo := flag.String("repo", ".", "Repository path")
	file := flag.String("file", "", "Specific file for diff")
	limit := flag.Int("limit", 20, "Log limit")
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
			Data: map[string]interface{}{"pong": true, "tool": "git.go"},
		})
		return
	}

	if *op == "version" {
		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"version": "1.0.0", "tool": "git.go"},
		})
		return
	}

	// Check git availability
	if _, err := exec.LookPath("git"); err != nil {
		writeJSON(Response{
			OK:    false,
			Error: &Error{Code: "GIT_MISSING", Message: "git binary not found"},
		})
		os.Exit(3)
	}

	absRepo, err := filepath.Abs(*repo)
	if err != nil {
		writeJSON(Response{
			OK:    false,
			Error: &Error{Code: "PATH_ERROR", Message: err.Error()},
		})
		os.Exit(2)
	}

	switch *op {
	case "status":
		out, err := runGit([]string{"status", "--short"}, absRepo)
		if err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "GIT_ERROR", Message: out},
			})
			os.Exit(5)
		}

		modified := []string{}
		untracked := []string{}
		lines := strings.Split(out, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "M ") || strings.HasPrefix(line, "MM ") || strings.HasPrefix(line, " M") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					modified = append(modified, parts[len(parts)-1])
				}
			} else if strings.HasPrefix(line, "?? ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					untracked = append(untracked, parts[len(parts)-1])
				}
			}
		}

		if *compact {
			writeJSON(Response{OK: true, Data: map[string]int{"modified": len(modified), "untracked": len(untracked)}})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"modified":  modified,
					"untracked": untracked,
				},
			})
		}

	case "log":
		args := []string{"log", "--oneline", fmt.Sprintf("-n%d", *limit)}
		out, err := runGit(args, absRepo)
		if err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "GIT_ERROR", Message: out},
			})
			os.Exit(5)
		}

		type Commit struct {
			Hash    string `json:"hash"`
			Message string `json:"message"`
		}
		commits := []Commit{}
		lines := strings.Split(out, "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				commits = append(commits, Commit{Hash: parts[0], Message: parts[1]})
			}
		}

		if *compact {
			writeJSON(Response{OK: true, Data: commits})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"commits": commits,
					"count":   len(commits),
				},
			})
		}

	case "diff":
		args := []string{"diff"}
		if *file != "" {
			args = append(args, *file)
		}
		out, err := runGit(args, absRepo)
		if err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "GIT_ERROR", Message: out},
			})
			os.Exit(5)
		}

		if *compact {
			// Truncate diff for compact mode
			if len(out) > 500 {
				out = out[:500] + "... (truncated)"
			}
			writeJSON(Response{OK: true, Data: out})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"diff": out,
				},
			})
		}

	case "branch":
		// Get current branch
		current, err := runGit([]string{"branch", "--show-current"}, absRepo)
		if err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "GIT_ERROR", Message: current},
			})
			os.Exit(5)
		}
		current = strings.TrimSpace(current)

		// Get all branches
		out, err := runGit([]string{"branch", "--list"}, absRepo)
		if err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "GIT_ERROR", Message: out},
			})
			os.Exit(5)
		}

		branches := []string{}
		lines := strings.Split(out, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			line = strings.TrimPrefix(line, "* ")
			if line != "" {
				branches = append(branches, line)
			}
		}

		if *compact {
			writeJSON(Response{OK: true, Data: current})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"current": current,
					"all":     branches,
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
