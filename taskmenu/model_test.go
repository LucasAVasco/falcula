package taskmenu

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/project"
)

// newTree creates a project tree. The projects that must not be loaded ('project2' and 'project1:subproject3') are invalid
func newTree(t *testing.T) *project.Project {
	t.Helper()
	dir := t.TempDir()

	const invalid = "tasks: [this is: not valid\n"
	testutil.WriteFile(t, filepath.Join(dir, "falcula.yaml"), `
projects: { project1: p1, project2: p2 }
tasks: { build: { shell: 'true' }, Test: { shell: 'true' } }
`)
	testutil.WriteFile(t, filepath.Join(dir, "p1", "falcula.yaml"), `
projects: { subproject2: s2, subproject3: s3 }
tasks: { deploy: { shell: 'true' } }
`)
	testutil.WriteFile(t, filepath.Join(dir, "p1", "s2", "falcula.yaml"), "tasks: { my-task: { shell: 'true' }, lint: { shell: 'true' } }\n")
	testutil.WriteFile(t, filepath.Join(dir, "p1", "s3", "falcula.yaml"), invalid)
	testutil.WriteFile(t, filepath.Join(dir, "p2", "falcula.yaml"), invalid)

	p, err := project.LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestModel_Entries(t *testing.T) {
	m := NewModel(newTree(t))

	tests := []struct {
		filter string
		want   []Entry
	}{
		{"", []Entry{{"Test", false}, {"build", false}, {"project1", true}, {"project2", true}}},
		{"te", []Entry{{"Test", false}}},                          // Case insensitive
		{"ject", []Entry{{"project1", true}, {"project2", true}}}, // Substring
		{"none", nil},
	}
	for _, tt := range tests {
		if got := m.Entries(tt.filter); !slices.Equal(got, tt.want) {
			t.Errorf("Entries(%q) = %v, want %v", tt.filter, got, tt.want)
		}
	}
}

func TestModel_Navigation(t *testing.T) {
	m := NewModel(newTree(t))

	if err := m.Open("project1"); err != nil {
		t.Fatal(err)
	}
	if m.Path() != "project1" {
		t.Errorf("Path() = %q", m.Path())
	}
	want := []Entry{{"deploy", false}, {"subproject2", true}, {"subproject3", true}}
	if got := m.Entries(""); !slices.Equal(got, want) {
		t.Errorf("Entries() in project1 = %v, want %v", got, want)
	}

	if err := m.Open("subproject2"); err != nil {
		t.Fatal(err)
	}
	if got := m.TaskID("my-task"); got != "project1:subproject2:my-task" {
		t.Errorf("TaskID() = %q", got)
	}

	if !m.Back() || m.Path() != "project1" || !m.Back() || m.Path() != "" {
		t.Errorf("Back() must return to the parent projects (path %q)", m.Path())
	}
	if m.Back() {
		t.Error("Back() at the root must return false")
	}
	if got := m.TaskID("build"); got != "build" {
		t.Errorf("TaskID() at the root = %q", got)
	}
}

func TestModel_OpenInvalidProject_KeepsTheLevel(t *testing.T) {
	m := NewModel(newTree(t))
	if err := m.Open("project1"); err != nil {
		t.Fatal(err)
	}

	if err := m.Open("subproject3"); err == nil {
		t.Error("opening an invalid project must fail")
	}
	if m.Path() != "project1" {
		t.Errorf("Path() after an error = %q, want project1", m.Path())
	}
}
