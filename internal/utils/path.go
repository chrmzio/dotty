// internal/utils/path.go
package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands ~ and environment variables in a path
func ExpandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") || path == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		if path == "~" {
			return homeDir, nil
		}

		return filepath.Join(homeDir, path[2:]), nil
	}

	// Expand environment variables
	expanded := os.ExpandEnv(path)

	return expanded, nil
}

// NormalizePath converts an absolute path back to use ~ for the home directory
func NormalizePath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	// Clean the path first
	path = filepath.Clean(path)

	// If the path starts with the home directory, replace it with ~
	if after, ok := strings.CutPrefix(path, homeDir); ok {
		relativePath := after
		if relativePath == "" {
			return "~"
		}
		if strings.HasPrefix(relativePath, string(filepath.Separator)) {
			return "~" + relativePath
		}
		return "~" + string(filepath.Separator) + relativePath
	}

	return path
}
