//go:build integration

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
)

// The test binary runs the CLI (the `main` function) when this environment variable is set. It avoids building the CLI separately
const runCLIEnv = "FALCULA_TEST_RUN_CLI"

func TestMain(m *testing.M) {
	if os.Getenv(runCLIEnv) == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// runCLI runs the CLI in the directory with the arguments. Returns its standard output and exit code
func runCLI(t *testing.T, dir string, args ...string) (stdout string, exitCode int) {
	t.Helper()

	cmd := exec.Command(os.Args[0], args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), runCLIEnv+"=1")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		return out.String(), exitErr.ExitCode()
	}
	if err != nil {
		t.Fatal(err)
	}
	return out.String(), 0
}

func newProject(t *testing.T) string {
	t.Helper()
	testutil.RequireShell(t)

	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "falcula.yaml"), `
projects: { child: child }
tasks:
  echo:
    shell: 'echo "args: $*"'
  lua:
    lua: 'print("lua args: " .. table.concat(arg, ","))'
`)
	testutil.WriteFile(t, filepath.Join(dir, "child", "falcula.yaml"), "tasks: { hello: { shell: 'echo hello from child' } }\n")
	return dir
}

func TestIntegrationCLI_TaskList_PrintsAllTasks(t *testing.T) {
	dir := newProject(t)

	out, code := runCLI(t, filepath.Join(dir, "child"), "task", "list") // Child projects are their own projects
	if code != 0 || strings.TrimSpace(out) != "hello" {
		t.Errorf("task list in the child project = %q (exit %d)", out, code)
	}

	out, code = runCLI(t, dir, "task", "list")
	names := strings.Fields(out)
	slices.Sort(names)
	if code != 0 || !slices.Equal(names, []string{"child:hello", "echo", "lua"}) {
		t.Fatalf("task list = %q (exit %d)", out, code)
	}
}

func TestIntegrationCLI_RunTask_PassesArguments(t *testing.T) {
	dir := newProject(t)

	for _, args := range [][]string{
		{"task", "run", "echo", "a", "b"},
		{"run", "echo", "a", "b"}, // Alias
	} {
		out, code := runCLI(t, dir, args...)
		if code != 0 || strings.TrimSpace(out) != "args: a b" {
			t.Errorf("%v = %q (exit %d)", args, out, code)
		}
	}

	out, code := runCLI(t, dir, "--raw", "run", "lua", "x", "y")
	if code != 0 || !strings.Contains(out, "lua args: x,y") {
		t.Errorf("Lua task in raw mode = %q (exit %d)", out, code)
	}

	out, code = runCLI(t, filepath.Join(dir, "child"), "run", "hello")
	if code != 0 || strings.TrimSpace(out) != "hello from child" {
		t.Errorf("task from a sub directory = %q (exit %d)", out, code)
	}
}

func TestIntegrationCLI_Errors_ExitWithNonZeroCode(t *testing.T) {
	dir := newProject(t)

	if out, code := runCLI(t, dir, "run", "unknown"); code == 0 || !strings.Contains(out, "task 'unknown' not found") {
		t.Errorf("unknown task = %q (exit %d)", out, code)
	}
	// Without a terminal, the task menu can not be shown
	if out, code := runCLI(t, dir, "run"); code == 0 || !strings.Contains(out, "a task name is required") {
		t.Errorf("running without a task name and a terminal = %q (exit %d), want an error", out, code)
	}
	if out, code := runCLI(t, dir, "--raw", "run"); code == 0 || !strings.Contains(out, "raw mode") {
		t.Errorf("running without a task name in raw mode = %q (exit %d), want an error", out, code)
	}

	outside := t.TempDir()
	if _, err := os.Stat(filepath.Join(filepath.Dir(outside), "falcula.yaml")); err == nil {
		t.Skip("the temporary directory is inside a falcula project")
	}
	if out, code := runCLI(t, outside, "task", "list"); code == 0 || !strings.Contains(out, "project file not found") {
		t.Errorf("outside a project = %q (exit %d)", out, code)
	}
}

func TestIntegrationCLI_NoTerminal_ForcesRawMode(t *testing.T) {
	dir := newProject(t)
	testutil.WriteFile(t, filepath.Join(dir, "tui.lua"), `
local tui = require('falcula.tui')
tui.show()
print('after show')
`)
	testutil.WriteFile(t, filepath.Join(dir, "falcula.yaml"), "tasks: { tui: { lua: 'dofile(\"tui.lua\")' } }\n")

	// The output is a pipe, so the TUI must not be shown even without `--raw`
	out, code := runCLI(t, dir, "run", "tui")
	if code != 0 || !strings.Contains(out, "after show") || strings.Contains(out, "\x1b[") {
		t.Errorf("TUI task without a terminal = %q (exit %d)", out, code)
	}
}

func TestIntegrationCLI_ConfigInit_WritesFilesOnce(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("FALCULA_CONFIG_DIR", configDir) // Inherited by the CLI process
	outside := t.TempDir()                    // Does not need a project

	out, code := runCLI(t, outside, "config", "init")
	if code != 0 || strings.Count(out, "written:") != 2 {
		t.Fatalf("config init = %q (exit %d)", out, code)
	}
	for _, file := range []string{"config.yaml", filepath.Join("themes", "soft.yaml")} {
		if _, err := os.Stat(filepath.Join(configDir, file)); err != nil {
			t.Errorf("config init did not write %s: %v", file, err)
		}
	}

	out, code = runCLI(t, outside, "config", "init")
	if code != 0 || strings.Count(out, "kept") != 2 {
		t.Errorf("second config init = %q (exit %d), want the files kept", out, code)
	}

	out, code = runCLI(t, outside, "config", "init", "--force")
	if code != 0 || strings.Count(out, "written:") != 2 {
		t.Errorf("config init --force = %q (exit %d)", out, code)
	}
}

func TestIntegrationCLI_BrokenConfig_StillRuns(t *testing.T) {
	dir := newProject(t)
	configDir := t.TempDir()
	t.Setenv("FALCULA_CONFIG_DIR", configDir)
	testutil.WriteFile(t, filepath.Join(configDir, "config.yaml"), "theme: [not, a, name]\n")

	out, code := runCLI(t, dir, "run", "lua", "x")
	if code != 0 || !strings.Contains(out, "lua args: x") {
		t.Errorf("task with a broken configuration = %q (exit %d)", out, code)
	}
}

func TestIntegrationCLI_Completion_TaskNames(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "falcula.yaml"), "projects: { project1: p1 }\ntasks: { build: { shell: 'true' } }\n")
	testutil.WriteFile(t, filepath.Join(dir, "p1", "falcula.yaml"),
		"projects: { subproject2: s2, subproject3: s3 }\ntasks: { deploy: { shell: 'true' } }\n")
	testutil.WriteFile(t, filepath.Join(dir, "p1", "s2", "falcula.yaml"), "tasks: { my-task: { shell: 'true' } }\n")
	testutil.WriteFile(t, filepath.Join(dir, "p1", "s3", "falcula.yaml"), "tasks: [broken\n") // Must not be loaded

	tests := []struct {
		args []string
		want string // Candidates and directive, one per line
	}{
		{[]string{"run", ""}, "build\nproject1:\n:6"},
		{[]string{"task", "run", "project1:"}, "project1:deploy\nproject1:subproject2:\nproject1:subproject3:\n:6"},
		{[]string{"run", "project1:subproject2:my-tas"}, "project1:subproject2:my-task\n:4"},
		{[]string{"run", "unknown:"}, ":4"},
		{[]string{"run", "build", ""}, ":0"}, // Task arguments: files
	}
	for _, tt := range tests {
		out, code := runCLI(t, dir, append([]string{"__complete"}, tt.args...)...)
		// The last line is a message for the user (in the standard error)
		lines := strings.Split(strings.TrimSpace(out), "\n")
		got := strings.Join(lines[:len(lines)-1], "\n")
		if code != 0 || got != tt.want {
			t.Errorf("completion of %q = %q (exit %d), want %q", tt.args, got, code, tt.want)
		}
	}
}
