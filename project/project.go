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

// Project is the project configuration
type Project struct {
	File           string            `yaml:"-"`
	Folder         string            `yaml:"-"`
	MinimumVersion string            `yaml:"minimum_version"`
	Cwd            string            `yaml:"cwd"` // Current working directory used for tasks. Defaults to Folder
	Extends        []string          `yaml:"extends"`
	Projects       map[string]string `yaml:"projects"`
	Root           bool              `yaml:"root"`
	Tasks          map[string]*Task  `yaml:"tasks"`
	FallbackTask   string            `yaml:"fallback_task"`
}

// LoadProject reads the project configuration file and parses it
func LoadProjectFile(file string) (*Project, error) {
	c := Project{
		Extends:  make([]string, 0),
		Projects: make(map[string]string),
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

	// Configuring tasks
	for _, task := range c.Tasks {
		task.Project = &c
		err := task.ConvertToAbsPath(c.Cwd)
		if err != nil {
			return nil, fmt.Errorf("error converting task '%s' to absolute path: %w", task.ShellFile, err)
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
func LoadProject(folder string) (*Project, error) {
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
func (p *Project) GetChildProjectByName(name string) (*Project, error) {
	subProjectName, innerName, hasSubProjectName := strings.Cut(name, ":")

	// Gets first inner project
	projectPath, ok := p.Projects[subProjectName]
	if !ok {
		return nil, fmt.Errorf("project '%s' not found", subProjectName)
	}

	if !filepath.IsAbs(projectPath) {
		projectPath = filepath.Join(p.Folder, projectPath)
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

// extractProjectName extracts the project name from the given name. Returns the project name and the rest of the name (not including the
// project name). Example: "project1:project2:task" -> "project1:project2", "task".
//
// If there is no project name, returns: "", name
func (p *Project) extractProjectName(name string) (projectName string, rest string) {
	lastProjectSeparatorIndex := strings.LastIndex(name, ":")
	if lastProjectSeparatorIndex == -1 {
		return "", name
	}

	projectName = name[:lastProjectSeparatorIndex]
	rest = name[lastProjectSeparatorIndex+1:]

	return projectName, rest
}

// getTaskRelativeToProject returns the task with the given name relative to the project folder
func (p *Project) getTaskRelativeToProject(name string) *Task {
	task, ok := p.Tasks[name]
	if ok {
		return task
	}

	// Uses fallback task if can not find a task with the given name
	task, ok = p.Tasks[p.FallbackTask]
	if ok {
		return task
	}

	return nil
}

// GetTaskByName returns the task with the given name. If the task is not found, it returns the fallback task. If there is no
// fallback task, returns an error
func (p *Project) GetTaskByName(name string) (*Task, error) {
	projectName, taskName := p.extractProjectName(name)

	// Gets the task in the current project
	if projectName == "" {
		task := p.getTaskRelativeToProject(name)
		if task == nil {
			return nil, fmt.Errorf("task '%s' not found", name)
		}

		return task, nil
	}

	// Gets the task in a child project
	project, err := p.GetChildProjectByName(projectName)
	if err != nil {
		return nil, fmt.Errorf("error getting child project '%s': %w", projectName, err)
	}

	task, err := project.GetTaskByName(taskName)
	if err != nil {
		return nil, fmt.Errorf("error getting task '%s' from project '%s': %w", taskName, projectName, err)
	}

	return task, nil
}

// GetAllTasks returns all the tasks in the current project and its children projects as a map
func (p *Project) GetAllTasks() (map[string]*Task, error) {
	tasks := make(map[string]*Task)
	maps.Copy(tasks, p.Tasks)

	// Adds tasks from children projects
	for projectName, projectPath := range p.Projects {
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
