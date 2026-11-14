package project

import (
	"fmt"
	"path/filepath"

	"github.com/LucasAVasco/falcula/filesystem"
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
		filepath.Join(folder, "falcula.local.yaml"),
		filepath.Join(folder, "falcula.local.yml"),
		filepath.Join(folder, "falcula.yaml"),
		filepath.Join(folder, "falcula.yml"),
	)
	if err != nil {
		return "", fmt.Errorf("error getting first project file: %w", err)
	}

	return projectFile, nil
}

// getFirstExistingFile returns the first file that exists in the given list. Follows the order in the list. Returns "" if not found
func getFirstExistingFile(files ...string) (string, error) {
	for _, file := range files {
		exists, err := filesystem.FileExists(file)
		if err != nil {
			return "", fmt.Errorf("error checking if file '%s' exists: %w", file, err)
		}

		if exists {
			return file, nil
		}
	}

	return "", nil
}
