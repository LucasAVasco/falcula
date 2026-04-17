// Package falcula is the main package of falcula. Contains the main application
package falcula

import (
	"fmt"
	"github.com/LucasAVasco/falcula/project"
)

// App is the main application. Its is a facade to all falcula features
type App struct {
	rawMode bool
	project *project.Config
}

// NewApp creates a new app instance. The `rawMode` parameter is used to determine if the falcula should run in raw mode (disables the TUI)
func NewApp(rawMode bool) (*App, error) {
	a := &App{
		rawMode: rawMode,
	}

	var err error
	a.project, err = project.ReadConfigFile("falcula.yaml")
	if err != nil {
		return nil, fmt.Errorf("error reading project file (falcula.yaml): %w", err)
	}

	return a, nil
}
