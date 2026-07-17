package project

import (
	"fmt"
	"path/filepath"
)

// Script is a falcula script. It can be either a shell command, a shell file, Lua code or a Lua file (can not be more than one of them)
type Script struct {
	Project   *Project `yaml:"-"`
	Cwd       string   `yaml:"cwd"`
	Shell     Command  `yaml:"shell"`
	Lua       string   `yaml:"lua"`
	ShellFile string   `yaml:"shell_file"`
	LuaFile   string   `yaml:"lua_file"`
}

// ConvertToAbsPath converts the paths of the script to absolute paths
func (s *Script) ConvertToAbsPath(folder string) error {
	if s.Cwd == "" {
		s.Cwd = folder
	}

	var err error
	if !filepath.IsAbs(s.Cwd) {
		s.Cwd, err = filepath.Abs(filepath.Join(folder, s.Cwd))
		if err != nil {
			return fmt.Errorf("error getting absolute path of CWD: %w", err)
		}
	}

	if s.ShellFile != "" && !filepath.IsAbs(s.ShellFile) {
		s.ShellFile, err = filepath.Abs(filepath.Join(folder, s.ShellFile))
		if err != nil {
			return fmt.Errorf("error getting absolute path of shell file: %w", err)
		}
	}

	if s.LuaFile != "" && !filepath.IsAbs(s.LuaFile) {
		s.LuaFile, err = filepath.Abs(filepath.Join(folder, s.LuaFile))
		if err != nil {
			return fmt.Errorf("error getting absolute path of Lua file: %w", err)
		}
	}

	return nil
}

// Validate returns an error if the script is not valid
func (s *Script) Validate() error {
	numActions := 0

	if s.Shell.IsNotEmpty() {
		numActions++
	}

	if s.Lua != "" {
		numActions++
	}

	if s.ShellFile != "" {
		numActions++
	}

	if s.LuaFile != "" {
		numActions++
	}

	if numActions == 0 {
		return fmt.Errorf("no action defined")
	} else if numActions > 1 {
		return fmt.Errorf("multiple actions defined, only one is allowed")
	}

	return nil
}
