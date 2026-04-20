// Package project contains the project configuration file parser
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

const ProjectFileName = "falcula.yaml"

// Config is the project configuration
type Config struct {
	Folder         string             `yaml:"-"`
	Projects       map[string]string  `yaml:"projects"`
	Root           bool               `yaml:"root"`
	Scripts        map[string]*Script `yaml:"scripts"`
	FallbackScript string             `yaml:"fallback_script"`
	Tasks          map[string]*Task   `yaml:"tasks"`
	FallbackTask   string             `yaml:"fallback_task"`
}

// ReadConfigFile reads the project configuration file and parses it
func ReadConfigFile(path string) (*Config, error) {
	c := Config{
		Projects: make(map[string]string),
		Scripts:  make(map[string]*Script),
	}

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

	// Configuring tasks
	for _, task := range c.Tasks {
		task.Project = &c
		err := task.ConvertToAbsPath(c.Folder)
		if err != nil {
			return nil, fmt.Errorf("error converting task '%s' to absolute path: %w", task.File, err)
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

	return &c, nil
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

	project, err := ReadConfigFile(projectPath + "/" + ProjectFileName)
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
			File: name,
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
