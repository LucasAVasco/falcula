package modtbl

import (
	"testing"

	"github.com/LucasAVasco/falcula/lua/luaerror"

	lua "github.com/yuin/gopher-lua"
)

func newState(t *testing.T) *lua.LState {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(L.Close)
	L.PreloadModule("tbl", func(L *lua.LState) int {
		mod := L.NewTable()
		if err := New().Loader(L, "tbl", mod); err != nil {
			t.Fatal(err)
		}
		L.Push(mod)
		return 1
	})
	return L
}

func TestExtendReturnsOnlyTheErrorOnFailure(t *testing.T) {
	luaerror.ConfigureAbortOnError(false)
	t.Cleanup(func() { luaerror.ConfigureAbortOnError(true) })

	L := newState(t)
	err := L.DoString(`
		local tbl = require("tbl")
		local n1, e1 = select("#", tbl.extend("invalid", {})), tbl.extend("invalid", {})
		local n2, e2 = select("#", tbl.deep_extend("error", {a = 1}, {a = 2})), tbl.deep_extend("error", {a = 1}, {a = 2})
		assert(n1 == 1 and type(e1) == "string", "extend: " .. tostring(n1) .. " " .. tostring(e1))
		assert(n2 == 1 and type(e2) == "string", "deep_extend: " .. tostring(n2) .. " " .. tostring(e2))
		assert(select("#", tbl.extend("keep", {}, {b = 1})) == 0)
	`)
	if err != nil {
		t.Fatal(err)
	}
}
