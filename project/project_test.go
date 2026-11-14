package project

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/version"
	"github.com/goccy/go-yaml"
)

// newProjectTree creates a project with a child project, a grandchild project and an extended file. Returns the root folder
func newProjectTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	testutil.WriteFile(t, filepath.Join(root, "falcula.yaml"), `
cwd: work
extends: [extra.yaml]
projects:
  child: sub
tasks:
  build:
    shell: ["make", "build"]
  script:
    lua_file: scripts/main.lua
    cwd: /tmp
fallback_task: build
`)
	testutil.WriteFile(t, filepath.Join(root, "extra.yaml"), `
tasks:
  build:
    shell: "overridden"
  extra:
    lua: "print('extra')"
`)
	testutil.WriteFile(t, filepath.Join(root, "sub", "falcula.yaml"), `
projects:
  grand: inner
tasks:
  hello:
    shell: "echo hello"
`)
	testutil.WriteFile(t, filepath.Join(root, "sub", "inner", "falcula.yml"), `
tasks:
  deep:
    shell_file: deep.sh
`)

	return root
}

func loadProject(t *testing.T, folder string) *Project {
	t.Helper()
	p, err := LoadProject(folder)
	if err != nil {
		t.Fatalf("LoadProject(%q): %v", folder, err)
	}
	return p
}

func TestLoadProject_ResolvesPathsRelativeToTheProjectFolder(t *testing.T) {
	root := newProjectTree(t)
	p := loadProject(t, root)

	if p.File != filepath.Join(root, "falcula.yaml") || p.Folder != root {
		t.Errorf("File = %q, Folder = %q", p.File, p.Folder)
	}
	if want := filepath.Join(root, "work"); p.Cwd != want {
		t.Errorf("Cwd = %q, want %q", p.Cwd, want)
	}
	if want := filepath.Join(root, "sub"); p.Projects["child"] != want {
		t.Errorf("child project path = %q, want %q", p.Projects["child"], want)
	}

	script := p.Tasks["script"]
	if script.Name != "script" || script.Project != p {
		t.Errorf("task name/project not set: %+v", script)
	}
	if want := filepath.Join(root, "work", "scripts", "main.lua"); script.LuaFile != want {
		t.Errorf("LuaFile = %q, want %q (relative to the project cwd)", script.LuaFile, want)
	}
	if script.Cwd != "/tmp" {
		t.Errorf("an absolute task cwd must be kept, got %q", script.Cwd)
	}
	if want := filepath.Join(root, "work"); p.Tasks["build"].Cwd != want {
		t.Errorf("task cwd defaults to the project cwd: got %q, want %q", p.Tasks["build"].Cwd, want)
	}
}

func TestLoadProject_DefaultCwd_IsTheProjectFolder(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "falcula.yaml"), "tasks: {}\n")

	if p := loadProject(t, root); p.Cwd != root {
		t.Fatalf("Cwd = %q, want %q", p.Cwd, root)
	}
}

func TestLoadProject_Extends_AddsTasksWithoutOverriding(t *testing.T) {
	p := loadProject(t, newProjectTree(t))

	if got := p.Tasks["build"].Shell.List; !slices.Equal(got, []string{"make", "build"}) {
		t.Errorf("the extended file must not override existing tasks, build = %+v", p.Tasks["build"].Shell)
	}
	if p.Tasks["extra"] == nil || p.Tasks["extra"].Lua != "print('extra')" {
		t.Errorf("the task of the extended file was not added: %+v", p.Tasks["extra"])
	}
}

func TestLoadProject_LocalFile_HasPriority(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "falcula.yaml"), "tasks: {a: {shell: regular}}\n")
	testutil.WriteFile(t, filepath.Join(root, "falcula.local.yaml"), "tasks: {a: {shell: local}}\n")

	file, err := GetConfigFile(root)
	if err != nil || file != filepath.Join(root, "falcula.local.yaml") {
		t.Fatalf("GetConfigFile = %q, %v", file, err)
	}
	if got := loadProject(t, root).Tasks["a"].Shell.String; got != "local" {
		t.Fatalf("expected the local file to be loaded, got task shell %q", got)
	}
}

func TestLoadProject_Errors(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{"invalid yaml", "tasks: [", "error parsing configuration file"},
		{"task without action", "tasks: {a: {cwd: x}}", "no action defined"},
		{"task with several actions", "tasks: {a: {shell: x, lua: y}}", "multiple actions"},
		{"invalid minimum version", "minimum_version: abc", "error parsing minimum version"},
		{"missing extended file", "extends: [missing.yaml]", "missing.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			testutil.WriteFile(t, filepath.Join(root, "falcula.yaml"), tt.content)

			_, err := LoadProject(root)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected an error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestLoadProject_NoConfigFile_ReturnsError(t *testing.T) {
	if _, err := LoadProject(t.TempDir()); err == nil {
		t.Fatal("expected an error")
	}
	if ok, err := HasConfigFile(t.TempDir()); ok || err != nil {
		t.Fatalf("HasConfigFile = %v, %v", ok, err)
	}
}

func TestLoadProject_MinimumVersionAboveCurrent_ReturnsError(t *testing.T) {
	current, err := version.FromString("v1.2.0")
	if err != nil {
		t.Fatal(err)
	}
	oldVersion, oldString := version.CurrentVersion, version.CurrentVersionString
	version.CurrentVersion, version.CurrentVersionString = current, "v1.2.0"
	t.Cleanup(func() { version.CurrentVersion, version.CurrentVersionString = oldVersion, oldString })

	for _, tt := range []struct {
		minimum string
		ok      bool
	}{{"v1.1.0", true}, {"v1.2.0", true}, {"v1.3.0", false}} {
		root := t.TempDir()
		testutil.WriteFile(t, filepath.Join(root, "falcula.yaml"), "minimum_version: "+tt.minimum+"\n")

		_, err := LoadProject(root)
		if (err == nil) != tt.ok {
			t.Errorf("minimum_version %s with current v1.2.0: err = %v, want ok=%v", tt.minimum, err, tt.ok)
		}
	}
}

func TestGetTaskByName(t *testing.T) {
	p := loadProject(t, newProjectTree(t))

	tests := []struct {
		name     string
		wantTask string
	}{
		{"script", "script"},
		{"unknown", "build"}, // Fallback task
		{"child:hello", "hello"},
		{"child:grand:deep", "deep"},
	}
	for _, tt := range tests {
		task, err := p.GetTaskByName(tt.name)
		if err != nil {
			t.Fatalf("GetTaskByName(%q): %v", tt.name, err)
		}
		if task.Name != tt.wantTask {
			t.Errorf("GetTaskByName(%q) = %q, want %q", tt.name, task.Name, tt.wantTask)
		}
	}

	for _, name := range []string{"missing:hello", "child:unknown", "child:grand:unknown"} {
		if _, err := p.GetTaskByName(name); err == nil {
			t.Errorf("GetTaskByName(%q) must return an error (children have no fallback task)", name)
		}
	}
}

func TestGetAllTasksNames_IncludesChildTasksWithPrefix(t *testing.T) {
	p := loadProject(t, newProjectTree(t))

	names, err := p.GetAllTasksNames()
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(names)

	want := []string{"build", "child:grand:deep", "child:hello", "extra", "script"}
	if !slices.Equal(names, want) {
		t.Fatalf("GetAllTasksNames = %v, want %v", names, want)
	}
}

func TestChildProjects(t *testing.T) {
	root := newProjectTree(t)
	p := loadProject(t, root)

	if names := p.GetChildProjectsNames(); !slices.Equal(names, []string{"child"}) {
		t.Errorf("GetChildProjectsNames = %v", names)
	}

	all, err := p.GetAllChildProjectsNames()
	slices.Sort(all)
	if err != nil || !slices.Equal(all, []string{"child", "child:grand"}) {
		t.Errorf("GetAllChildProjectsNames = %v, %v", all, err)
	}

	grand, err := p.GetChildProjectByName("child:grand")
	if err != nil || grand.Folder != filepath.Join(root, "sub", "inner") {
		t.Errorf("GetChildProjectByName(child:grand) = %+v, %v", grand, err)
	}

	if _, err := p.GetChildProjectByName("child:missing"); err == nil {
		t.Error("expected an error for a missing nested project")
	}
}

func TestChildProjects_MissingFolder_ReturnsError(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "falcula.yaml"), "projects: {broken: missing}\n")
	p := loadProject(t, root)

	if _, err := p.GetChildProjects(); err == nil {
		t.Error("GetChildProjects must fail")
	}
	if _, err := p.GetAllTasks(); err == nil {
		t.Error("GetAllTasks must fail")
	}
}

func TestCommand_UnmarshalYAML(t *testing.T) {
	var single, list Command
	if err := yaml.Unmarshal([]byte(`echo "hi"`), &single); err != nil || single.String != "echo \"hi\"" || single.List != nil {
		t.Errorf("string command = %+v, %v", single, err)
	}
	if err := yaml.Unmarshal([]byte(`[echo, hi]`), &list); err != nil || !slices.Equal(list.List, []string{"echo", "hi"}) {
		t.Errorf("list command = %+v, %v", list, err)
	}

	var invalid Command
	if err := yaml.Unmarshal([]byte(`{a: 1}`), &invalid); err == nil {
		t.Error("a map must not be a valid command")
	}

	if (&Command{}).IsNotEmpty() || !single.IsNotEmpty() || !list.IsNotEmpty() {
		t.Error("IsNotEmpty returned the wrong value")
	}
}

func TestTask_Validate(t *testing.T) {
	tests := []struct {
		name  string
		task  Task
		valid bool
	}{
		{"no action", Task{}, false},
		{"shell string", Task{Shell: Command{String: "x"}}, true},
		{"shell list", Task{Shell: Command{List: []string{"x"}}}, true},
		{"lua", Task{Lua: "x"}, true},
		{"shell file", Task{ShellFile: "x"}, true},
		{"lua file", Task{LuaFile: "x"}, true},
		{"two actions", Task{Lua: "x", LuaFile: "y"}, false},
	}
	for _, tt := range tests {
		if err := tt.task.Validate(); (err == nil) != tt.valid {
			t.Errorf("%s: Validate() = %v, want valid=%v", tt.name, err, tt.valid)
		}
	}
}
