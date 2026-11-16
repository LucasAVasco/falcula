// Package taskmenu implements the interactive menu that selects the task to run (`falcula run` without a task)
package taskmenu

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/LucasAVasco/falcula/project"
)

// Entry is an entry of the menu: a task or a child project of the current project
type Entry struct {
	Name      string
	IsProject bool
}

// level is a project opened in the menu
type level struct {
	path    string // Name of the project relative to the root project (e.g. "project1:subproject2"). Empty for the root project
	project *project.Project
}

// Model is the state of the menu (without user interface). The child projects are only loaded when opened
type Model struct {
	levels []level // Opened projects, from the root to the current one
}

// NewModel creates the model of the menu starting at the root project
func NewModel(root *project.Project) *Model {
	return &Model{levels: []level{{project: root}}}
}

func (m *Model) current() level {
	return m.levels[len(m.levels)-1]
}

// Path returns the name of the current project relative to the root project. Empty for the root project
func (m *Model) Path() string {
	return m.current().path
}

// Entries returns the tasks (sorted) and then the child projects (sorted) of the current project whose name contains the filter (case
// insensitive). The child projects are not loaded
func (m *Model) Entries(filter string) []Entry {
	filter = strings.ToLower(filter)
	matches := func(name string) bool {
		return strings.Contains(strings.ToLower(name), filter)
	}

	proj := m.current().project
	var entries []Entry
	for _, name := range slices.Sorted(maps.Keys(proj.Tasks)) {
		if matches(name) {
			entries = append(entries, Entry{Name: name})
		}
	}
	for _, name := range slices.Sorted(maps.Keys(proj.Projects)) {
		if matches(name) {
			entries = append(entries, Entry{Name: name, IsProject: true})
		}
	}

	return entries
}

// Open loads the child project with the name and makes it the current project. On error, the current project does not change
func (m *Model) Open(name string) error {
	current := m.current()
	child, err := current.project.GetChildProjectByName(name)
	if err != nil {
		return fmt.Errorf("error opening project '%s': %w", m.TaskID(name), err)
	}

	m.levels = append(m.levels, level{path: m.TaskID(name), project: child})
	return nil
}

// Back makes the parent project the current project. Returns false if the current project is the root project
func (m *Model) Back() bool {
	if len(m.levels) == 1 {
		return false
	}
	m.levels = m.levels[:len(m.levels)-1]
	return true
}

// TaskID returns the identifier (relative to the root project) of a task or child project of the current project
func (m *Model) TaskID(name string) string {
	if path := m.Path(); path != "" {
		return path + ":" + name
	}
	return name
}
