package luaclass

import (
	"errors"
	"strings"
	"testing"

	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/maplua"

	lua "github.com/yuin/gopher-lua"
)

func TestClassErrorsAreLuaErrors(t *testing.T) {
	luaerror.ConfigureAbortOnError(false)
	t.Cleanup(func() { luaerror.ConfigureAbortOnError(true) })

	L := lua.NewState()
	defer L.Close()

	mod := L.NewTable()
	_, err := Register(L, mod, &Info{
		Name: "Thing",
		Constructor: func(L *lua.LState, newObj *lua.LTable) error {
			if L.Get(2) == lua.LNil {
				return errors.New("missing name")
			}
			SetInstance(L, newObj, L.ToString(2))
			return nil
		},
		Methods: map[string]Method{
			"name": func(L *lua.LState) int {
				L.Push(lua.LString(GetInstance[string](L)))
				return 1
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	L.SetGlobal("Thing", L.GetField(mod, "Thing"))

	// The decoded table must work with the standard library functions
	decoded, err := maplua.Marshal(map[string]any{"a": 1.0})
	if err != nil {
		t.Fatal(err)
	}
	L.SetGlobal("decoded", decoded)

	err = L.DoString(`
		local obj, err = Thing:new()
		assert(obj == nil and err:find("missing name"), tostring(err))

		local list, err = Thing:new_list({ { "a" }, 3 })
		assert(list == nil and err:find("must be a table"), tostring(err))

		local thing = Thing:new("x")
		assert(thing:name() == "x")

		assert(getmetatable(decoded) == nil)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Calling a method with '.' raises a Lua error instead of crashing
	err = L.DoString(`Thing:new("x").name()`)
	if err == nil || !strings.Contains(err.Error(), "call methods with ':'") {
		t.Fatalf("unexpected error: %v", err)
	}
}
