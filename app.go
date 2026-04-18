// Package falcula is the main package of falcula. Contains the main application
package falcula

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LucasAVasco/falcula/project"
)

// App is the main application. Its is a facade to all falcula features
type App struct {
	rawMode   bool
	invokeDir string
	project   *project.Config
}

// NewApp creates a new app instance. The `rawMode` parameter is used to determine if the falcula should run in raw mode (disables the TUI)
func NewApp(rawMode bool) (*App, error) {
	a := &App{
		rawMode: rawMode,
	}

	// Invoke directory
	invokeDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current working directory (invoke directory): %w", err)
	}

	invokeDir, err = filepath.Abs(invokeDir)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path of current working directory (invoke directory): %w", err)
	}

	a.invokeDir = invokeDir

	// Project
	projectFile, err := getProjectFilePath()
	if err != nil {
		return nil, fmt.Errorf("error getting project file path: %w", err)
	}

	a.project, err = project.ReadConfigFile(projectFile)
	if err != nil {
		return nil, fmt.Errorf("error reading project file (falcula.yaml): %w", err)
	}

	return a, nil
}

// getProjectFilePath returns the path to the falcula configuration file that the application should use
func getProjectFilePath() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current working directory: %w", err)
	}

	currentDir, err = filepath.Abs(currentDir)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path of current working directory: %w", err)
	}

	// Iterate from the current directory to the root searching for a 'falcula.yaml' file
	var projectFile string
	for {
		projectFile = currentDir + "/falcula.yaml"
		_, err = os.Stat(projectFile)
		if err == nil {
			break
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("error checking if file '%s' exists: %w", projectFile, err)
		}

		if currentDir == "/" {
			return "", fmt.Errorf("project file 'falcula.yaml' not found")
		}

		// Next directory
		currentDir = filepath.Dir(currentDir)
	}

	return projectFile, nil
}
