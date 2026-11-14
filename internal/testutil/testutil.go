// Package testutil contains helpers shared by the tests of this module. It must only be imported by tests
package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// ModuleLoader is a Lua module that can be preloaded in a Lua state (implemented by all modules of the 'lua/modules' package)
type ModuleLoader interface {
	Loader(L *lua.LState, name string, mod *lua.LTable) error
}

// NewLuaState returns a new Lua state with the modules preloaded (`require(name)` loads them). The state is closed when the test ends
func NewLuaState(t testing.TB, modules map[string]ModuleLoader) *lua.LState {
	t.Helper()

	L := lua.NewState()
	t.Cleanup(L.Close)

	for name, module := range modules {
		L.PreloadModule(name, func(L *lua.LState) int {
			mod := L.NewTable()
			if err := module.Loader(L, name, mod); err != nil {
				t.Errorf("error loading module '%s': %v", name, err)
			}
			L.Push(mod)
			return 1
		})
	}

	return L
}

// WriteFile writes a file (creating its parent directories) and returns its path
func WriteFile(t testing.TB, path, content string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

// ReadFile returns the content of a file
func ReadFile(t testing.TB, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

// RequireShell skips the test if the POSIX shell is not available (e.g. on Windows)
func RequireShell(t testing.TB) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("requires a POSIX shell")
	}
}

// FakeDocker places a fake `docker` executable first in the `PATH`. Each call appends one line to the returned log file with the value of
// the `DOCKER_DEFAULT_PLATFORM` environment variable and the arguments: "platform=<value> <arguments...>". The fake exits with the code in
// the `FAKE_DOCKER_EXIT` environment variable (0 if empty)
func FakeDocker(t testing.TB) (logPath string) {
	t.Helper()
	RequireShell(t)

	dir := t.TempDir()
	logPath = filepath.Join(dir, "docker.log")

	script := "#!/bin/sh\n" +
		"echo \"platform=${DOCKER_DEFAULT_PLATFORM:-} $*\" >> \"$FAKE_DOCKER_LOG\"\n" +
		"exit \"${FAKE_DOCKER_EXIT:-0}\"\n"
	if err := os.WriteFile(filepath.Join(dir, "docker"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_DOCKER_LOG", logPath)

	return logPath
}

// ReadLines returns the non-empty lines of a file. Returns nil if the file does not exist
func ReadLines(t testing.TB, path string) []string {
	t.Helper()

	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}

	lines := []string{}
	for line := range strings.SplitSeq(string(content), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}
