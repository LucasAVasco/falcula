//go:build integration

package falcula

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
)

// newApp creates a project in a temporary directory, changes to it and creates the application (raw mode, without TUI). Tasks write their
// results to files in the project directory, which is returned
func newApp(t *testing.T) (*App, string) {
	t.Helper()
	testutil.RequireShell(t)

	dir, err := filepath.EvalSymlinks(t.TempDir()) // The tasks see the real path (e.g. macOS links /var to /private/var)
	if err != nil {
		t.Fatal(err)
	}

	testutil.WriteFile(t, filepath.Join(dir, "falcula.yaml"), `
projects:
  child: child
tasks:
  list:
    shell: ["sh", "-c", 'echo "$0 $*" > list.out', "first"]
  string:
    shell: 'echo "$FALCULA_TASK_NAME|$FALCULA_SUB_TASK_NAME|$FALCULA_PROJECT_DIR|$FALCULA_INVOKE_DIR|$*" > string.out'
  shell_file:
    shell_file: scripts/task.sh
  inline:
    lua: |
      local out = io.open("inline.out", "w")
      out:write("main:" .. table.concat(arg, ","))
      out:close()
      return {
        sub = {
          deep = function(...)
            local f = io.open("deep.out", "w")
            f:write("deep:" .. table.concat({ ... }, ","))
            f:close()
          end,
        },
      }
  services:
    lua_file: scripts/services.lua
  broken:
    lua: "error('broken script')"
`)
	testutil.WriteFile(t, filepath.Join(dir, "scripts", "task.sh"), `
greet() { echo "greet $*" > shell_file.out; }
`)
	testutil.WriteFile(t, filepath.Join(dir, "scripts", "services.lua"), `
local process = require("falcula.process")
local manager = require("falcula.manager")

local parallel = manager.ServiceManager:new("parallel")
parallel:add_services({
  process.Provider:new("a", "echo prepared > a.prepare", "echo ran > a.run"):new_service(),
  process.Provider:new("b", nil, { "sh", "-c", "echo ran > b.run" }):new_service(),
})
assert(parallel:run() == nil)

local serial = manager.ServiceManager:new("serial")
serial:add_service(process.Provider:new("first", nil, "echo first >> serial.out"):new_service())
serial:add_service(process.Provider:new("second", nil, "echo second >> serial.out"):new_service())
assert(serial:run_serial() == nil)

local out = io.open("args.out", "w")
out:write(table.concat(arg, " "))
out:close()
`)
	testutil.WriteFile(t, filepath.Join(dir, "child", "falcula.yaml"), `
tasks:
  hello:
    shell: 'echo "child $*" > child.out'
`)

	t.Chdir(dir)
	app, err := NewApp(true)
	if err != nil {
		t.Fatal(err)
	}
	return app, dir
}

func readOutput(t *testing.T, path string) string {
	t.Helper()
	return strings.TrimSpace(testutil.ReadFile(t, path))
}

func TestIntegrationShellListTask_ReceivesArguments(t *testing.T) {
	app, dir := newApp(t)

	if err := app.RunTaskById("list", "a", "b c"); err != nil {
		t.Fatal(err)
	}
	if got := readOutput(t, filepath.Join(dir, "list.out")); got != "first a b c" {
		t.Fatalf("output = %q", got)
	}
}

func TestIntegrationShellStringTask_ReceivesArgumentsAndEnvironment(t *testing.T) {
	app, dir := newApp(t)

	if err := app.RunTaskById("string", "x", "y"); err != nil {
		t.Fatal(err)
	}
	want := "string||" + dir + "|" + dir + "|x y"
	if got := readOutput(t, filepath.Join(dir, "string.out")); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestIntegrationShellFileTask_RunsTheSubtaskFunction(t *testing.T) {
	app, dir := newApp(t)

	if err := app.RunTaskById("shell_file.greet", "world"); err != nil {
		t.Fatal(err)
	}
	if got := readOutput(t, filepath.Join(dir, "shell_file.out")); got != "greet world" {
		t.Fatalf("output = %q", got)
	}
}

func TestIntegrationLuaTask_RunsNestedSubtask(t *testing.T) {
	app, dir := newApp(t)

	if err := app.RunTaskById("inline.sub.deep", "1", "2"); err != nil {
		t.Fatal(err)
	}
	if got := readOutput(t, filepath.Join(dir, "inline.out")); got != "main:1,2" {
		t.Errorf("main chunk output = %q", got)
	}
	if got := readOutput(t, filepath.Join(dir, "deep.out")); got != "deep:1,2" {
		t.Errorf("subtask output = %q", got)
	}
}

func TestIntegrationLuaTask_MissingSubtask_ReturnsError(t *testing.T) {
	app, _ := newApp(t)

	if err := app.RunTaskById("inline.sub.missing"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestIntegrationLuaFileTask_RunsProcessServices(t *testing.T) {
	app, dir := newApp(t)

	if err := app.RunTaskById("services", "arg1", "arg2"); err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"a.prepare":  "prepared",
		"a.run":      "ran",
		"b.run":      "ran",
		"serial.out": "first\nsecond",
		"args.out":   "arg1 arg2",
	}
	for file, want := range expected {
		if got := readOutput(t, filepath.Join(dir, file)); got != want {
			t.Errorf("%s = %q, want %q", file, got, want)
		}
	}
}

func TestIntegrationChildProjectTask(t *testing.T) {
	app, dir := newApp(t)

	if err := app.RunTaskById("child:hello", "from", "root"); err != nil {
		t.Fatal(err)
	}
	// The child task runs in the child project folder
	if got := readOutput(t, filepath.Join(dir, "child", "child.out")); got != "child from root" {
		t.Fatalf("output = %q", got)
	}
}

func TestIntegrationGetAllTasks_IncludesChildTasks(t *testing.T) {
	app, _ := newApp(t)

	tasks, err := app.GetAllTasks()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"list", "inline", "services", "child:hello"} {
		if tasks[name] == nil {
			t.Errorf("task %q is missing", name)
		}
	}
}

func TestIntegrationErrors(t *testing.T) {
	app, _ := newApp(t)

	if err := app.RunTaskById("unknown"); err == nil {
		t.Error("an unknown task must return an error")
	}

	// A Lua error is logged (the script is user code) and does not stop the application
	if err := app.RunTaskById("broken"); err != nil {
		t.Errorf("a Lua runtime error must be logged, not returned: %v", err)
	}
}

func TestIntegrationNewApp_OutsideAProject_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "falcula.yaml")); err == nil {
		t.Skip("the temporary directory is inside a falcula project")
	}

	t.Chdir(dir)
	if _, err := NewApp(true); err == nil {
		t.Fatal("expected an error")
	}
}
