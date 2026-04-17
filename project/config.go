// Package project contains the project configuration file parser
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// Config is the project configuration
type Config struct {
	Folder         string             `yaml:"-"`
	Root           bool               `yaml:"root"`
	Scripts        map[string]*Script `yaml:"scripts"`
	FallbackScript string             `yaml:"fallback_script"`
}

// ReadConfigFile reads the project configuration file and parses it
func ReadConfigFile(path string) (*Config, error) {
	c := Config{}

	fileContent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %w", err)
	}

	err = yaml.Unmarshal(fileContent, &c)
	if err != nil {
		return nil, fmt.Errorf("error parsing configuration file: %w", err)
	}

	c.Folder = filepath.Dir(path)

	// Configuring scripts
	for _, script := range c.Scripts {
		script.Project = &c
		err := script.ConvertToAbsPath(c.Folder)
		if err != nil {
			return nil, fmt.Errorf("error converting script '%s' to absolute path: %w", script.File, err)
		}
	}

	// Validates the scripts
	for name, script := range c.Scripts {
		err = script.Validate()
		if err != nil {
			return nil, fmt.Errorf("error validating script '%s': %w", name, err)
		}
	}

	return &c, nil
}

// getScriptRelativeToProject returns the script with the given name relative to the project folder
func (c *Config) getScriptRelativeToProject(name string) *Script {
	script, ok := c.Scripts[name]
	if ok {
		return script
	}

	// Uses fallback script if can not find a script with the given name
	script, ok = c.Scripts[c.FallbackScript]
	if ok {
		return script
	}

	// Treats the name as a script path
	if strings.HasSuffix(name, ".lua") {
		script = &Script{
			LuaFile: name,
		}
	} else {
		script = &Script{
			File: name,
		}
	}

	return script
}

// GetScriptByName returns the script with the given name. If the script is not found, it returns the fallback script. If there is no
// fallback script, it treats the name as a script path
func (c *Config) GetScriptByName(name string) (*Script, error) {
	script := c.getScriptRelativeToProject(name)

	return script, nil
}
