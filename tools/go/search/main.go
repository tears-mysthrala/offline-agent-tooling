package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

type Match struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

func loadGitIgnore(root string) ([]string, error) {
	path := filepath.Join(root, ".gitignore")
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}
	return patterns, nil
}

func shouldIgnore(path string, patterns []string) bool {
	for _, p := range patterns {
		// Simple glob matching
		matched, _ := filepath.Match(p, filepath.Base(path))
		if matched {
			return true
		}
		// Check directory matches (very basic)
		if strings.Contains(path, string(os.PathSeparator)+p+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

func main() {
	op := flag.String("op", "", "Operation: grep|ping|version")
	root := flag.String("root", ".", "Root directory")
	pattern := flag.String("pattern", "", "Search pattern")
	isRegex := flag.Bool("is_regex", false, "Treat pattern as regex")
	include := flag.String("include", "", "Include glob pattern")
	exclude := flag.String("exclude", "", "Exclude glob pattern")
	recursive := flag.Bool("recursive", false, "Recursive search")
	ignoreGit := flag.Bool("ignore_git", false, "Respect .gitignore")
	limit := flag.Int("limit", 2000, "Max results")
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
			Data: map[string]interface{}{"pong": true, "tool": "search.go"},
		})
		return
	}

	if *op == "version" {
		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"version": "1.0.0", "tool": "search.go"},
		})
		return
	}

	if *op == "grep" {
		if *pattern == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--pattern required"},
			})
			os.Exit(2)
		}

		var re *regexp.Regexp
		var err error
		if *isRegex {
			re, err = regexp.Compile(*pattern)
			if err != nil {
				writeJSON(Response{
					OK:    false,
					Error: &Error{Code: "REGEX_ERROR", Message: err.Error()},
				})
				os.Exit(2)
			}
		}

		ignorePatterns := []string{".git", ".svn", ".hg", ".DS_Store"}
		if *exclude != "" {
			ignorePatterns = append(ignorePatterns, strings.Split(*exclude, ",")...)
		}
		if *ignoreGit {
			gitPatterns, _ := loadGitIgnore(*root)
			ignorePatterns = append(ignorePatterns, gitPatterns...)
		}

		matches := []Match{}

		err = filepath.WalkDir(*root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if len(matches) >= *limit {
				return filepath.SkipDir
			}

			if d.IsDir() {
				if path != *root && !*recursive {
					return filepath.SkipDir
				}
				if shouldIgnore(path, ignorePatterns) {
					return filepath.SkipDir
				}
				return nil
			}

			if shouldIgnore(path, ignorePatterns) {
				return nil
			}

			if *include != "" {
				matched, _ := filepath.Match(*include, d.Name())
				if !matched {
					return nil
				}
			}

			// Read file line by line using scanner
			file, err := os.Open(path)
			if err != nil {
				return nil
			}
			// Manual close for performance in hot loop
			// defer file.Close()

			scanner := bufio.NewScanner(file)
			lineNum := 0
			for scanner.Scan() {
				lineNum++
				line := strings.TrimSpace(scanner.Text())
				found := false
				if *isRegex {
					found = re.MatchString(line)
				} else {
					found = strings.Contains(line, *pattern)
				}

				if found {
					matches = append(matches, Match{
						File: path,
						Line: lineNum,
						Text: line,
					})
					if len(matches) >= *limit {
						file.Close()
						return filepath.SkipDir
					}
				}
			}
			file.Close()

			return nil
		})

		if *compact {
			writeJSON(Response{OK: true, Data: matches})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"matches": matches,
					"count":   len(matches),
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
