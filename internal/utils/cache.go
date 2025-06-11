package utils

import (
	"encoding/json"
	"os"
	"time"
)

func CacheValid(path string, maxAge time.Duration) bool {

	// make the parent directories recursively if they do not exist
	if err := os.MkdirAll(CACHE_DIRECTORY, 0755); err != nil {
		Report("Failed to create cache directory: " + err.Error())

		return false
	}

	// get file info
	info, err := os.Stat(path)

	if err != nil {
		return false
	}

	age := time.Since(info.ModTime())
	return age < maxAge
}

func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
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
