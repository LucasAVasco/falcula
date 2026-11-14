package luapath

import (
	"path/filepath"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"

	lua "github.com/yuin/gopher-lua"
)

// runFile runs a Lua file that calls `capture()`. Returns the results of the function passed to it, called inside the file
func runFile(t *testing.T, file string, fn func(L *lua.LState) []string) []string {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(L.Close)

	var results []string
	L.SetGlobal("capture", L.NewFunction(func(L *lua.LState) int {
		results = fn(L)
		return 0
	}))

	if err := L.DoFile(file); err != nil {
		t.Fatal(err)
	}
	return results
}

func TestPaths_AreRelativeToTheCurrentLuaFile(t *testing.T) {
	dir := t.TempDir()
	file := testutil.WriteFile(t, filepath.Join(dir, "scripts", "main.lua"), "capture()\n")
	scripts := filepath.Join(dir, "scripts")

	results := runFile(t, file, func(L *lua.LState) []string {
		mustGet := func(value string, err error) string {
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		return []string{
			mustGet(GetCurrentLuaFile(L)),
			mustGet(GetCurrentLuaDirectory(L)),
			mustGet(GetAbs(L, "data/x.txt")),
			mustGet(GetAbs(L, "/abs/../path")),
			mustGet(GetRel(L, filepath.Join(dir, "other", "y"), "")),
			mustGet(GetRel(L, "a/b", "a")),
		}
	})

	want := []string{
		file,
		scripts,
		filepath.Join(scripts, "data", "x.txt"),
		"/path",
		filepath.Join("..", "other", "y"),
		"b",
	}
	for i := range want {
		if results[i] != want[i] {
			t.Errorf("result %d = %q, want %q", i, results[i], want[i])
		}
	}
}

func TestGetCurrentLuaFile_OutsideAFile_ReturnsError(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	var err error
	L.SetGlobal("capture", L.NewFunction(func(L *lua.LState) int {
		_, err = GetCurrentLuaFile(L)
		return 0
	}))
	if runErr := L.DoString("capture()"); runErr != nil {
		t.Fatal(runErr)
	}
	if err == nil {
		t.Fatal("expected an error when no Lua file is running")
	}
}
