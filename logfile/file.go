// Package logfile generates a log file in the temporary directory of the current user
package logfile

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

// New creates a new log file in the temporary directory of the current user. You need to close and delete the log file manually.
func New() (*os.File, error) {
	// Get current user
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("error getting current user: %w", err)
	}

	// Temporary directory
	tmpDir := os.TempDir() + "/falcula-" + currentUser.Username

	err = os.MkdirAll(tmpDir, 0700)
	if err != nil {
		return nil, fmt.Errorf("error creating log file directory: %w", err)
	}

	// Get the current working directory (base name)
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current working directory: %w", err)
	}

	currentDir = filepath.Base(currentDir)

	// Creates the log file
	logFilePath := currentDir + "-" + time.Now().Format(time.RFC3339)
	logFile, err := os.CreateTemp(tmpDir, logFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening log file: %w", err)
	}

	return logFile, nil
}
