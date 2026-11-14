package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/LucasAVasco/falcula/filesystem"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/theme"
)

// defaultConfigFile is the configuration file written by `Init`
const defaultConfigFile = `# Falcula configuration

# Theme of the text user interface. Name of a file in the 'themes' folder (without the extension) or a built-in theme (soft)
theme: soft
`

// InitResult is the result of writing a file of the configuration folder
type InitResult struct {
	Path    string
	Written bool // False if the file already existed and was kept
}

// Init writes the default configuration file and the default theme to the configuration folder. Existing files are only overwritten if
// `force` is true
func Init(force bool) ([]InitResult, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	files := []struct {
		path    string
		content []byte
	}{
		{filepath.Join(dir, "config.yaml"), []byte(defaultConfigFile)},
		{filepath.Join(dir, themesDirName, theme.DefaultName+".yaml"), theme.BuiltinFile(theme.DefaultName)},
	}

	results := make([]InitResult, 0, len(files))
	for _, file := range files {
		exists, err := filesystem.FileExists(file.path)
		if err != nil {
			return results, fmt.Errorf("error checking if file '%s' exists: %w", file.path, err)
		}
		if exists && !force {
			results = append(results, InitResult{Path: file.path})
			continue
		}

		if err := os.MkdirAll(filepath.Dir(file.path), 0o755); err != nil {
			return results, fmt.Errorf("error creating folder: %w", err)
		}
		if err := os.WriteFile(file.path, file.content, 0o644); err != nil {
			return results, fmt.Errorf("error writing file: %w", err)
		}
		results = append(results, InitResult{Path: file.path, Written: true})
	}

	return results, nil
}
