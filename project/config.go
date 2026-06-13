// Package project contains the project configuration file parser
package project

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/LucasAVasco/falcula/version"
	"github.com/goccy/go-yaml"
)

// Config is the project configuration
type Config struct {
	File           string             `yaml:"-"`
	Folder         string             `yaml:"-"`
	MinimumVersion string             `yaml:"minimum_version"`
	Cwd            string             `yaml:"cwd"` // Current working directory used for scripts and tasks. Defaults to Folder
	Extends        []string           `yaml:"extends"`
	Projects       map[string]string  `yaml:"projects"`
	Root           bool               `yaml:"root"`
	Scripts        map[string]*Script `yaml:"scripts"`
	FallbackScript string             `yaml:"fallback_script"`
	Tasks          map[string]*Task   `yaml:"tasks"`
	FallbackTask   string             `yaml:"fallback_task"`
}

// LoadProject reads the project configuration file and parses it
func LoadProjectFile(file string) (*Config, error) {
	c := Config{
		Extends:  make([]string, 0),
		Projects: make(map[string]string),
		Scripts:  make(map[string]*Script),
		Tasks:    make(map[string]*Task),
	}

	// Getting absolute path of file and folder
	file, err := filepath.Abs(file)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path of file '%s': %w", file, err)
	}
	folder := filepath.Dir(file)

	// Reading configuration file
	fileContent, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %w", err)
	}

	// Parsing configuration file
	err = yaml.Unmarshal(fileContent, &c)
	if err != nil {
		return nil, fmt.Errorf("error parsing configuration file: %w", err)
	}

	// Validating minimum version
	if c.MinimumVersion != "" {
		minVersion, err := version.FromString(c.MinimumVersion)
		if err != nil {
			return nil, fmt.Errorf("error parsing minimum version: %w", err)
		}

		if version.CurrentVersion != nil && minVersion.Compare(version.CurrentVersion) > 0 {
			return nil, fmt.Errorf(
				"minimum version (%s) is greater than current version (%s), you must update falcula to run this project",
				c.MinimumVersion,
				version.CurrentVersionString,
			)
		}
	}

	// Configuring paths after parsing
	c.File = file
	c.Folder = folder
	if c.Cwd == "" {
		c.Cwd = folder
	}
	if !filepath.IsAbs(c.Cwd) {
		cwd := filepath.Join(folder, c.Cwd)
		cwd, err = filepath.Abs(cwd)
		if err != nil {
			return nil, fmt.Errorf("error getting absolute path of CWD: %w", err)
		}
		c.Cwd = cwd
	}

	// Configuring projects
	for name, path := range c.Projects {
		if !filepath.IsAbs(path) {
			path = filepath.Join(folder, path)
		}
		c.Projects[name] = path
	}

	// Configuring scripts
	for _, script := range c.Scripts {
		script.Project = &c
		err := script.ConvertToAbsPath(c.Cwd)
		if err != nil {
			return nil, fmt.Errorf("error converting script '%s' to absolute path: %w", script.ShellFile, err)
		}
	}

	// Configuring tasks
	for _, task := range c.Tasks {
		task.Project = &c
		err := task.ConvertToAbsPath(c.Cwd)
		if err != nil {
			return nil, fmt.Errorf("error converting task '%s' to absolute path: %w", task.ShellFile, err)
		}
	}

	// Validates the scripts
	for name, script := range c.Scripts {
		err = script.Validate()
		if err != nil {
			return nil, fmt.Errorf("error validating script '%s': %w", name, err)
		}
	}

	// Validates the tasks
	for name, task := range c.Tasks {
		err = task.Validate()
		if err != nil {
			return nil, fmt.Errorf("error validating task '%s': %w", name, err)
		}
	}

	// Extending configuration
	for _, extendFile := range c.Extends {
		if !filepath.IsAbs(extendFile) {
			extendFile = filepath.Join(folder, extendFile)
			extendFile, err = filepath.Abs(extendFile)
			if err != nil {
				return nil, fmt.Errorf("error getting absolute path of file '%s': %w", extendFile, err)
			}
		}
		extend, err := LoadProjectFile(extendFile)
		if err != nil {
			return nil, fmt.Errorf("error loading project configuration file '%s': %w", extendFile, err)
		}

		mergeWithoutOverride(c.Projects, extend.Projects)
		mergeWithoutOverride(c.Scripts, extend.Scripts)
		mergeWithoutOverride(c.Tasks, extend.Tasks)
	}

	return &c, nil
}

// mergeWithoutOverride merges the given maps without overriding existing keys
func mergeWithoutOverride[T any](dst, src map[string]T) {
	for k, v := range src {
		if _, ok := dst[k]; ok {
			continue
		}

		dst[k] = v
	}
}

// LoadProject reads the first project configuration file in the given folder and loads it
func LoadProject(folder string) (*Config, error) {
	path, err := GetConfigFile(folder)
	if err != nil {
		return nil, fmt.Errorf("error getting project configuration file: %w", err)
	}
	if path == "" {
		return nil, fmt.Errorf("project configuration file not found")
	}

	project, err := LoadProjectFile(path)
	if err != nil {
		return nil, fmt.Errorf("error loading project configuration file: %w", err)
	}

	return project, nil
}

// GetChildProjectByName returns the child project with the given name or nil if not found
func (c *Config) GetChildProjectByName(name string) (*Config, error) {
	subProjectName, innerName, hasSubProjectName := strings.Cut(name, ":")

	// Gets first inner project
	projectPath, ok := c.Projects[subProjectName]
	if !ok {
		return nil, fmt.Errorf("project '%s' not found", subProjectName)
	}

	if !filepath.IsAbs(projectPath) {
		projectPath = filepath.Join(c.Folder, projectPath)
	}

	project, err := LoadProject(projectPath)
	if err != nil {
		return nil, fmt.Errorf("error reading child project configuration file '%s': %w", innerName, err)
	}
	if !hasSubProjectName {
		return project, nil
	}

	// Gets inner project recursively until there is no more sub projects
	project, err = project.GetChildProjectByName(innerName)
	if err != nil {
		return nil, fmt.Errorf("error getting child project '%s': %w", innerName, err)
	}

	return project, nil
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
			ShellFile: name,
		}
	}

	return script
}

// extractProjectName extracts the project name from the given name. Returns the project name and the rest of the name (not including the
// project name). Example: "project1:project2:script" -> "project1:project2", "script".
//
// If there is no project name, returns: "", name
func (c *Config) extractProjectName(name string) (projectName string, rest string) {
	lastProjectSeparatorIndex := strings.LastIndex(name, ":")
	if lastProjectSeparatorIndex == -1 {
		return "", name
	}

	projectName = name[:lastProjectSeparatorIndex]
	rest = name[lastProjectSeparatorIndex+1:]

	return projectName, rest
}

// GetScriptByName returns the script with the given name. If the script is not found, it returns the fallback script. If there is no
// fallback script, it treats the name as a script path
func (c *Config) GetScriptByName(name string) (*Script, error) {
	projectName, scriptName := c.extractProjectName(name)

	// Gets the script in the current project
	if projectName == "" {
		script := c.getScriptRelativeToProject(name)
		if script == nil {
			return nil, fmt.Errorf("script '%s' not found", name)
		}

		return script, nil
	}

	// Gets the script in a child project
	project, err := c.GetChildProjectByName(projectName)
	if err != nil {
		return nil, fmt.Errorf("error getting child project '%s': %w", projectName, err)
	}

	script, err := project.GetScriptByName(scriptName)
	if err != nil {
		return nil, fmt.Errorf("error getting script '%s' from project '%s': %w", scriptName, projectName, err)
	}

	return script, nil
}

// GetAllScripts returns all the scripts in the current project and its children projects as a map
func (c *Config) GetAllScripts() (map[string]*Script, error) {
	scripts := make(map[string]*Script)
	maps.Copy(scripts, c.Scripts)

	// Adds scripts from children projects
	for projectName, projectPath := range c.Projects {
		project, err := LoadProject(projectPath)
		if err != nil {
			return nil, fmt.Errorf("error reading configuration file of project '%s' at path '%s': %v", projectName, projectPath, err)
		}
		subScripts, err := project.GetAllScripts()
		if err != nil {
			return nil, fmt.Errorf("error getting scripts from child project '%s': %w", projectName, err)
		}

		for name, script := range subScripts {
			scripts[projectName+":"+name] = script
		}
	}

	return scripts, nil
}

// getTaskRelativeToProject returns the task with the given name relative to the project folder
func (c *Config) getTaskRelativeToProject(name string) *Task {
	task, ok := c.Tasks[name]
	if ok {
		return task
	}

	// Uses fallback task if can not find a task with the given name
	task, ok = c.Tasks[c.FallbackTask]
	if ok {
		return task
	}

	return nil
}

// GetTaskByName returns the task with the given name. If the task is not found, it returns the fallback task. If there is no
// fallback task, returns an error
func (c *Config) GetTaskByName(name string) (*Task, error) {
	projectName, scriptName := c.extractProjectName(name)

	// Gets the task in the current project
	if projectName == "" {
		task := c.getTaskRelativeToProject(name)
		if task == nil {
			return nil, fmt.Errorf("task '%s' not found", name)
		}

		return task, nil
	}

	// Gets the task in a child project
	project, err := c.GetChildProjectByName(projectName)
	if err != nil {
		return nil, fmt.Errorf("error getting child project '%s': %w", projectName, err)
	}

	task, err := project.GetTaskByName(scriptName)
	if err != nil {
		return nil, fmt.Errorf("error getting task '%s' from project '%s': %w", scriptName, projectName, err)
	}

	return task, nil
}

// GetAllTasks returns all the tasks in the current project and its children projects as a map
func (c *Config) GetAllTasks() (map[string]*Task, error) {
	tasks := make(map[string]*Task)
	maps.Copy(tasks, c.Tasks)

	// Adds tasks from children projects
	for projectName, projectPath := range c.Projects {
		project, err := LoadProject(projectPath)
		if err != nil {
			return nil, fmt.Errorf("error reading configuration file of project '%s' at path '%s': %v", projectName, projectPath, err)
		}

		subTasks, err := project.GetAllTasks()
		if err != nil {
			return nil, fmt.Errorf("error getting tasks from child project '%s': %w", projectName, err)
		}

		for name, task := range subTasks {
			tasks[projectName+":"+name] = task
		}
	}

	return tasks, nil
}
