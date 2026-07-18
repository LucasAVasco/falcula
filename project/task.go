package project

import (
	"fmt"
	"path/filepath"
)

// Task is a falcula task. It can be either a shell command, a shell file, Lua code or a Lua file (can not be more than one of them)
type Task struct {
	Project   *Project `yaml:"-"`
	Cwd       string   `yaml:"cwd"`
	Shell     Command  `yaml:"shell"`
	Lua       string   `yaml:"lua"`
	ShellFile string   `yaml:"shell_file"`
	LuaFile   string   `yaml:"lua_file"`
}

// ConvertToAbsPath converts the paths of the task to absolute paths
func (t *Task) ConvertToAbsPath(folder string) error {
	if t.Cwd == "" {
		t.Cwd = folder
	}

	var err error
	if !filepath.IsAbs(t.Cwd) {
		t.Cwd, err = filepath.Abs(filepath.Join(folder, t.Cwd))
		if err != nil {
			return fmt.Errorf("error getting absolute path of CWD: %w", err)
		}
	}

	if t.ShellFile != "" && !filepath.IsAbs(t.ShellFile) {
		t.ShellFile, err = filepath.Abs(filepath.Join(folder, t.ShellFile))
		if err != nil {
			return fmt.Errorf("error getting absolute path of shell file: %w", err)
		}
	}

	if t.LuaFile != "" && !filepath.IsAbs(t.LuaFile) {
		t.LuaFile, err = filepath.Abs(filepath.Join(folder, t.LuaFile))
		if err != nil {
			return fmt.Errorf("error getting absolute path of Lua file: %w", err)
		}
	}

	return nil
}

// Validate returns an error if the task is not valid
func (t *Task) Validate() error {
	numActions := 0

	if t.Shell.IsNotEmpty() {
		numActions++
	}

	if t.Lua != "" {
		numActions++
	}

	if t.ShellFile != "" {
		numActions++
	}

	if t.LuaFile != "" {
		numActions++
	}

	if numActions == 0 {
		return fmt.Errorf("no action defined")
	} else if numActions > 1 {
		return fmt.Errorf("multiple actions defined, only one is allowed")
	}

	return nil
}
