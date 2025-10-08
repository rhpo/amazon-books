package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

func CacheValid(path string, maxAge time.Duration) bool {
	// make the parent directories recursively if they do not exist
	if err := os.MkdirAll(CACHE_DIRECTORY, 0755); err != nil {
		Report("Failed to create parent cache directory: " + err.Error())
		return false
	}

	if err := os.MkdirAll(CACHE_DIRECTORY+"/books", 0755); err != nil {
		Report("Failed to create childbook cache directory: " + err.Error())
		return false
	}

	// handle wildcard in path, e.g., books-cache/2-*.json
	matches, err := filepath.Glob(path)
	if err != nil || len(matches) == 0 {
		return false
	}

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		age := time.Since(info.ModTime())
		if age < maxAge {
			return true
		}
	}
	return false
}

func ReadFile(path string) (string, string, error) {
	matches, err := filepath.Glob(path)
	if err != nil || len(matches) == 0 {
		return "", "", err
	}
	// Read the first matching file
	content, err := os.ReadFile(matches[0])
	if err != nil {
		return "", "", err
	}
	return string(content), matches[0], nil
}

func WriteFile(path string, content string) error {
	err := os.WriteFile(path, []byte(content), 0644) // 0644 is the permission for the file: // read/write for owner, read for group and others
	if err != nil {
		return err
	}
	return nil
}

func ParseJson(content string, v any) error {
	err := json.Unmarshal([]byte(content), v)
	if err != nil {
		return err
	}
	return nil
}

func ToJson(v any) (string, error) {
	content, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}
