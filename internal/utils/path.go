package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func ExpandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

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

	expanded := os.ExpandEnv(path)

	return expanded, nil
}

func NormalizePath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

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
