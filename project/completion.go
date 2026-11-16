package project

import (
	"fmt"
	"slices"
	"strings"
)

// CompleteTaskName returns the task names that complete the partial task name (shell completion). The candidates are the tasks and the
// child projects (with a trailing ":") of the project in the partial name that start with its last part. Example: "a:b:ta" completes to
// the tasks of the project "a:b" that start with "ta".
//
// Only the projects in the partial name are loaded (e.g. "a" and "a:b"), not the other child projects. `hasProjects` is true if a
// candidate is a project (the shell must not add a space after it, the user continues typing its task)
func (p *Project) CompleteTaskName(toComplete string) (candidates []string, hasProjects bool, err error) {
	projectName, partial := p.extractProjectName(toComplete)

	target := p
	prefix := ""
	if projectName != "" {
		target, err = p.GetChildProjectByName(projectName)
		if err != nil {
			return nil, false, fmt.Errorf("error getting project '%s': %w", projectName, err)
		}
		prefix = projectName + ":"
	}

	for name := range target.Tasks {
		if strings.HasPrefix(name, partial) {
			candidates = append(candidates, prefix+name)
		}
	}

	// NOTE(LucasAVasco): only the names of the child projects are used, they are not loaded
	for name := range target.Projects {
		if strings.HasPrefix(name, partial) {
			candidates = append(candidates, prefix+name+":")
			hasProjects = true
		}
	}

	slices.Sort(candidates)
	return candidates, hasProjects, nil
}

// CompleteTaskArgs returns the next argument of the argument sets (`args` option) of the task that start with the typed arguments and
// `toComplete`. The task identifier may have a sub-task (".sub-task"). Only the projects in the task identifier are loaded. `ok` is false
// if the task has no argument sets (the caller can complete something else, e.g. files)
func (p *Project) CompleteTaskArgs(taskID string, typed []string, toComplete string) (candidates []string, ok bool, err error) {
	taskName, _, _ := strings.Cut(taskID, ".")
	task, err := p.GetTaskByName(taskName)
	if err != nil {
		return nil, false, err
	}
	if len(task.Args) == 0 {
		return nil, false, nil
	}

	for _, args := range task.Args {
		if len(args) <= len(typed) || !slices.Equal([]string(args[:len(typed)]), typed) {
			continue
		}
		if next := args[len(typed)]; strings.HasPrefix(next, toComplete) && !slices.Contains(candidates, next) {
			candidates = append(candidates, next)
		}
	}

	slices.Sort(candidates)
	return candidates, true, nil
}
