// Package config loads the user configuration of falcula. It is a folder (see `Dir`) with a `config.yaml` file and a `themes` folder
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LucasAVasco/falcula/filesystem"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/theme"

	"github.com/goccy/go-yaml"
)

// DirEnv is the environment variable that overrides the configuration folder
const DirEnv = "FALCULA_CONFIG_DIR"

// themesDirName is the name of the folder with the user themes (inside the configuration folder)
const themesDirName = "themes"

// Config is the user configuration
type Config struct {
	Theme string `yaml:"theme"` // Name of the theme of the text user interface (file name in the themes folder without extension)

	dir string // Configuration folder
}

// Dir returns the configuration folder: `$FALCULA_CONFIG_DIR` or the `falcula` folder in the user configuration folder (e.g.
// `~/.config/falcula` on Linux)
func Dir() (string, error) {
	if dir := os.Getenv(DirEnv); dir != "" {
		return filepath.Abs(dir)
	}

	userDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("error getting the user configuration folder (set %s to choose one): %w", DirEnv, err)
	}

	return filepath.Join(userDir, "falcula"), nil
}

// Default returns the default configuration (used when there is no configuration file)
func Default() *Config {
	return &Config{Theme: theme.DefaultName}
}

// Load loads the configuration file of the configuration folder. Returns the default configuration if there is no configuration file. On
// error, also returns a usable configuration (the default one)
func Load() (*Config, error) {
	dir, err := Dir()
	if err != nil {
		return Default(), err
	}

	config := Default()
	config.dir = dir

	file, err := firstExistingFile(filepath.Join(dir, "config.yaml"), filepath.Join(dir, "config.yml"))
	if err != nil || file == "" {
		return config, err
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return config, fmt.Errorf("error reading configuration file: %w", err)
	}

	if err := yaml.UnmarshalWithOptions(data, config, yaml.DisallowUnknownField()); err != nil {
		config.Theme = theme.DefaultName
		return config, fmt.Errorf("error parsing configuration file '%s': %w", file, err)
	}
	if config.Theme == "" {
		config.Theme = theme.DefaultName
	}

	return config, nil
}

// LoadTheme loads the theme of the configuration. The themes of the themes folder override the built-in ones
func (c *Config) LoadTheme() (*theme.Theme, error) {
	return theme.Resolve(c.Theme, c.lookupTheme)
}

// lookupTheme returns the theme file with the name. Searches the themes folder and then the built-in themes
func (c *Config) lookupTheme(name string) (*theme.File, error) {
	if c.dir != "" {
		themesDir := filepath.Join(c.dir, themesDirName)
		file, err := firstExistingFile(filepath.Join(themesDir, name+".yaml"), filepath.Join(themesDir, name+".yml"))
		if err != nil {
			return nil, err
		}

		if file != "" {
			data, err := os.ReadFile(file)
			if err != nil {
				return nil, fmt.Errorf("error reading theme file: %w", err)
			}

			parsed, err := theme.ParseFile(data)
			if err != nil {
				return nil, fmt.Errorf("error parsing theme file '%s': %w", file, err)
			}
			return parsed, nil
		}
	}

	parsed, err := theme.BuiltinLookup(name)
	if errors.Is(err, theme.ErrNotFound) {
		return nil, fmt.Errorf("%w (no '%s.yaml' file in '%s' and no built-in theme with this name)", err, name,
			filepath.Join(c.dir, themesDirName))
	}
	return parsed, err
}

// firstExistingFile returns the first file that exists in the list. Returns "" if none exists
func firstExistingFile(files ...string) (string, error) {
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
