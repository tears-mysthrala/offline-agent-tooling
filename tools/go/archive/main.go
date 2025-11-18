package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

func zipDirectory(source, dest string) error {
	zipFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			return err
		}

		return nil
	})
}

func unzipArchive(source, dest string) error {
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	os.MkdirAll(dest, 0755)

	for _, file := range reader.File {
		path := filepath.Join(dest, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}

func listArchive(source string) ([]string, error) {
	reader, err := zip.OpenReader(source)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var files []string
	for _, file := range reader.File {
		files = append(files, file.Name)
	}

	return files, nil
}

func main() {
	op := flag.String("op", "", "Operation: zip|unzip|list|ping|version")
	source := flag.String("source", "", "Source file or directory")
	dest := flag.String("dest", "", "Destination file or directory")
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
			Data: map[string]interface{}{"pong": true, "tool": "archive.go"},
		})
		return
	}

	if *op == "version" {
		writeJSON(Response{
			OK:   true,
			Data: map[string]interface{}{"version": "1.0.0", "tool": "archive.go"},
		})
		return
	}

	switch *op {
	case "zip":
		if *source == "" || *dest == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--source and --dest required"},
			})
			os.Exit(2)
		}

		if err := zipDirectory(*source, *dest); err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ZIP_ERROR", Message: err.Error()},
			})
			os.Exit(5)
		}

		info, _ := os.Stat(*dest)
		if *compact {
			writeJSON(Response{OK: true, Data: *dest})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"source":  *source,
					"archive": *dest,
					"size":    info.Size(),
				},
			})
		}

	case "unzip":
		if *source == "" || *dest == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--source and --dest required"},
			})
			os.Exit(2)
		}

		if err := unzipArchive(*source, *dest); err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "UNZIP_ERROR", Message: err.Error()},
			})
			os.Exit(5)
		}

		if *compact {
			writeJSON(Response{OK: true, Data: *dest})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"archive": *source,
					"dest":    *dest,
				},
			})
		}

	case "list":
		if *source == "" {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "ARG_MISSING", Message: "--source required"},
			})
			os.Exit(2)
		}

		files, err := listArchive(*source)
		if err != nil {
			writeJSON(Response{
				OK:    false,
				Error: &Error{Code: "LIST_ERROR", Message: err.Error()},
			})
			os.Exit(5)
		}

		if *compact {
			writeJSON(Response{OK: true, Data: files})
		} else {
			writeJSON(Response{
				OK: true,
				Data: map[string]interface{}{
					"archive": *source,
					"files":   files,
					"count":   len(files),
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
