package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// HasConfigFile returns true if a project configuration file exists at the given folder. Returns false otherwise
func HasConfigFile(folder string) (bool, error) {
	projectFile, err := GetConfigFile(folder)
	if err != nil {
		return false, fmt.Errorf("error getting first project file: %w", err)
	}

	return projectFile != "", nil
}

// GetConfigFile returns the project configuration file at the given folder. Returns "" if not found
func GetConfigFile(folder string) (string, error) {
	folder, err := filepath.Abs(folder)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path of folder '%s': %w", folder, err)
	}

	projectFile, err := getFirstExistingFile(
		// NOTE(LucasAVasco): the order matters here. Loads the local file first
		folder+"/falcula.local.yaml",
		folder+"/falcula.local.yml",
		folder+"/falcula.yaml",
		folder+"/falcula.yml",
	)
	if err != nil {
		return "", fmt.Errorf("error getting first project file: %w", err)
	}

	return projectFile, nil
}

// getFirstExistingFile returns the first file that exists in the given list. Follows the order in the list. Returns "" if not found
func getFirstExistingFile(files ...string) (string, error) {
	for _, file := range files {
		exists, err := fileExists(file)
		if err != nil {
			return "", fmt.Errorf("error checking if file '%s' exists: %w", file, err)
		}

		if exists {
			return file, nil
		}
	}

	return "", nil
}

// fileExists checks if a file exists
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, fmt.Errorf("error getting file file status: %w", err)
}
