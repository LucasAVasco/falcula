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
	if _, code := runCLI(t, dir, "run"); code == 0 {
		t.Error("running without a task name must fail")
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
