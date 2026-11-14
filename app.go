// Package falcula is the main package of falcula. Contains the main application
package falcula

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/LucasAVasco/falcula/config"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/theme"
	"github.com/LucasAVasco/falcula/project"
)

// App is the main application. It is a facade to all falcula features
type App struct {
	rawMode   bool
	invokeDir string
	project   *project.Project

	theme      *theme.Theme // Theme of the TUI (from the user configuration)
	themeError error        // Error loading the theme. Falcula uses the default theme and reports it in the TUI
}

// NewApp creates a new app instance. The `rawMode` parameter is used to determine if the falcula should run in raw mode (disables the TUI)
func NewApp(rawMode bool) (*App, error) {
	a := &App{
		rawMode: rawMode,
	}

	// Invoke directory
	invokeDir, err := getWorkingDir()
	if err != nil {
		return nil, fmt.Errorf("error getting invoke directory: %w", err)
	}
	a.invokeDir = invokeDir

	// Project
	projectDir, err := getProjectDir(invokeDir)
	if err != nil {
		return nil, fmt.Errorf("error getting project file path: %w", err)
	}

	a.project, err = project.LoadProject(projectDir)
	if err != nil {
		return nil, fmt.Errorf("error reading project file (falcula.yaml): %w", err)
	}

	// User configuration. Errors do not prevent falcula from running
	if !rawMode {
		a.theme, a.themeError = loadTheme()
	}

	return a, nil
}

// getWorkingDir returns the absolute path of the current working directory
func getWorkingDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current working directory: %w", err)
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path of current working directory: %w", err)
	}

	return dir, nil
}

// getProjectDir returns the path to the falcula configuration directory that the application should use. Searches from the start
// directory (absolute path) to the file system root
func getProjectDir(currentDir string) (string, error) {
	// Iterate from the current directory to the root searching for a 'falcula.yaml' file
	for {
		hasConfigFile, err := project.HasConfigFile(currentDir)
		if err != nil {
			return "", fmt.Errorf("error checking if '%s' has a project configuration file: %w", currentDir, err)
		}

		// Found
		if hasConfigFile {
			return currentDir, nil
		}

		// At root (no project file found)
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", fmt.Errorf("project file not found")
		}

		currentDir = parentDir
	}
}

// loadTheme loads the theme of the user configuration. Returns the default theme and the error if it can not be loaded
func loadTheme() (*theme.Theme, error) {
	userConfig, err := config.Load()
	if err != nil {
		return theme.Default(), err
	}

	t, err := userConfig.LoadTheme()
	if err != nil {
		return theme.Default(), err
	}

	return t, nil
}
