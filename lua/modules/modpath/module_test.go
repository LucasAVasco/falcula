package modpath

import (
	"path/filepath"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/lua/luaerror"

	lua "github.com/yuin/gopher-lua"
)

func TestPathFunctions_AreRelativeToTheScript(t *testing.T) {
	luaerror.ConfigureAbortOnError(false)
	t.Cleanup(func() { luaerror.ConfigureAbortOnError(true) })

	dir := t.TempDir()
	scriptDir := filepath.Join(dir, "scripts")
	script := testutil.WriteFile(t, filepath.Join(scriptDir, "main.lua"), `
		local path = require("path")
		local dir = ...

		assert(path.get_current_file() == dir .. "/main.lua", path.get_current_file())
		assert(path.get_current_dir() == dir)
		assert(path.abs("a/b") == dir .. "/a/b")

		local list = path.abs_list({ "x", "/y" })
		assert(#list == 2 and list[1] == dir .. "/x" and list[2] == "/y")

		assert(path.rel(dir .. "/a/b") == "a/b")
		assert(path.rel(dir .. "/a/b", "a") == "b")

		local rel = path.rel_list({ dir .. "/p", dir .. "/q/r" }, dir .. "/q")
		assert(rel[1] == "../p" and rel[2] == "r")

		-- Cannot be relative to a different base kind
		local result, err = path.rel("relative", "")
		assert(result == "relative", tostring(err))
	`)

	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"path": New()})
	fn, err := L.LoadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	L.Push(fn)
	L.Push(lua.LString(scriptDir))
	if err := L.PCall(1, 0, nil); err != nil {
		t.Fatal(err)
	}
}

func TestPathFunctions_OutsideAScript_ReturnError(t *testing.T) {
	luaerror.ConfigureAbortOnError(false)
	t.Cleanup(func() { luaerror.ConfigureAbortOnError(true) })

	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"path": New()})
	err := L.DoString(`
		local path = require("path")
		local file, err = path.get_current_file()
		assert(file == nil and err ~= nil)

		local list, err = path.abs_list({ "relative" })
		assert(list == nil and err ~= nil, "abs_list must stop at the first error")
	`)
	if err != nil {
		t.Fatal(err)
	}
}
