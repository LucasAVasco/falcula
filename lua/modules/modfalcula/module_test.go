package modfalcula

import (
	"errors"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/lua/luaerror"

	lua "github.com/yuin/gopher-lua"
)

func TestConfigureAbortOnError(t *testing.T) {
	t.Cleanup(func() { luaerror.ConfigureAbortOnError(true) })

	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"falcula": New()})
	L.SetGlobal("fail", L.NewFunction(func(L *lua.LState) int {
		return luaerror.Push(L, 0, errors.New("boom"))
	}))

	err := L.DoString(`
		local falcula = require("falcula")

		falcula.configure_abort_on_error(false)
		assert(fail() == "boom")

		falcula.configure_abort_on_error(true)
		assert(not pcall(fail))

		assert(not pcall(falcula.configure_abort_on_error, "yes"), "the argument must be a boolean")
	`)
	if err != nil {
		t.Fatal(err)
	}
}
