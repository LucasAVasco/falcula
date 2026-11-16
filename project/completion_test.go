package project

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
)

// newCompletionTree creates a project tree. The projects that must not be loaded have invalid configuration files (loading them fails)
func newCompletionTree(t *testing.T) *Project {
	t.Helper()
	dir := t.TempDir()

	const invalid = "tasks: [this is: not valid\n"
	testutil.WriteFile(t, filepath.Join(dir, "falcula.yaml"), `
projects: { project1: p1, project2: p2 }
tasks: { build: { shell: 'true' }, test: { shell: 'true' } }
`)
	testutil.WriteFile(t, filepath.Join(dir, "p1", "falcula.yaml"), `
projects: { subproject2: s2, subproject3: s3 }
tasks: { deploy: { shell: 'true' } }
`)
	testutil.WriteFile(t, filepath.Join(dir, "p1", "s2", "falcula.yaml"), `
tasks:
  my-task: { shell: 'true', args: [dev, prod --verbose, prod --quiet, [deploy, my server], '', [deploy, ''], [empty, '', x]] }
  my-other: { shell: 'true' }
  lint: { shell: 'true' }
`)
	testutil.WriteFile(t, filepath.Join(dir, "p1", "s3", "falcula.yaml"), invalid)
	testutil.WriteFile(t, filepath.Join(dir, "p2", "falcula.yaml"), invalid)

	p, err := LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCompleteTaskName(t *testing.T) {
	p := newCompletionTree(t)

	tests := []struct {
		toComplete  string
		want        []string
		hasProjects bool
	}{
		{"", []string{"build", "project1:", "project2:", "test"}, true},
		{"b", []string{"build"}, false},
		{"pro", []string{"project1:", "project2:"}, true},
		{"project1:", []string{"project1:deploy", "project1:subproject2:", "project1:subproject3:"}, true},
		{"project1:subproject2:my-t", []string{"project1:subproject2:my-task"}, false},
		{"project1:subproject2:my-tas", []string{"project1:subproject2:my-task"}, false},
		{"project1:subproject2:", []string{"project1:subproject2:lint", "project1:subproject2:my-other", "project1:subproject2:my-task"}, false},
		{"none", nil, false},
	}

	// NOTE: project2 and subproject3 are invalid, so these cases also check that they are not loaded
	for _, tt := range tests {
		got, hasProjects, err := p.CompleteTaskName(tt.toComplete)
		if err != nil || !slices.Equal(got, tt.want) || hasProjects != tt.hasProjects {
			t.Errorf("CompleteTaskName(%q) = %q, %v, %v; want %q, %v", tt.toComplete, got, hasProjects, err, tt.want, tt.hasProjects)
		}
	}
}

func TestCompleteTaskName_Errors(t *testing.T) {
	p := newCompletionTree(t)

	for _, toComplete := range []string{"unknown:", "project1:unknown:ta", "project2:"} {
		if _, _, err := p.CompleteTaskName(toComplete); err == nil {
			t.Errorf("CompleteTaskName(%q) must fail", toComplete)
		}
	}
}

func TestCompleteTaskArgs(t *testing.T) {
	p := newCompletionTree(t)
	const task = "project1:subproject2:my-task"

	tests := []struct {
		taskID     string
		typed      []string
		toComplete string
		want       []string
	}{
		{task, nil, "", []string{"deploy", "dev", "empty", "prod"}}, // "prod" only once
		{task, nil, "de", []string{"deploy", "dev"}},
		{task, []string{"prod"}, "", []string{"--quiet", "--verbose"}},
		{task, []string{"deploy"}, "my", []string{"my server"}},
		{task, []string{"deploy"}, "", []string{"my server"}}, // The empty argument is not offered
		{task, []string{"empty", ""}, "", []string{"x"}},      // An empty argument typed with '' or ""
		{task, []string{"dev"}, "", nil},                      // No more arguments
		{task, []string{"other"}, "", nil},
		{task + ".sub-task", nil, "d", []string{"deploy", "dev"}},
		{task, nil, "e", []string{"empty"}},
	}
	for _, tt := range tests {
		got, ok, err := p.CompleteTaskArgs(tt.taskID, tt.typed, tt.toComplete)
		if err != nil || !ok || !slices.Equal(got, tt.want) {
			t.Errorf("CompleteTaskArgs(%q, %q, %q) = %q, %v, %v; want %q", tt.taskID, tt.typed, tt.toComplete, got, ok, err, tt.want)
		}
	}

	// Task without argument sets
	if got, ok, err := p.CompleteTaskArgs("build", nil, ""); ok || err != nil || got != nil {
		t.Errorf("task without args = %q, %v, %v; want not ok", got, ok, err)
	}
	if _, _, err := p.CompleteTaskArgs("unknown", nil, ""); err == nil {
		t.Error("an unknown task must fail")
	}
}
