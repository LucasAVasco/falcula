package filesystem

import (
	"errors"
	"fmt"
	"os"
)

// FileExists returns true if the file exists and false otherwise.
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("error getting file information: %w", err)
	}

	return true, nil
}

// FileIsDir returns true if the file is a directory and false otherwise.
func FileIsDir(path string) (bool, error) {
	status, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("error getting file information: %w", err)
	}
	return status.IsDir(), nil
}
