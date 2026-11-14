// Package modtbl is a module that provides table related functions
package modtbl

import (
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/luatable"
	"github.com/LucasAVasco/falcula/lua/modules/base"

	lua "github.com/yuin/gopher-lua"
)

// Module is a module that provides table related functions
type Module struct {
	base.BaseModule
}

func New() *Module {
	return &Module{}
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	L.SetFuncs(mod, map[string]lua.LGFunction{
		"extend":      extendFunction(luatable.Extend, "error extending table"),
		"deep_extend": extendFunction(luatable.DeepExtend, "error deep extending table"),
	})

	return nil
}

// extendFunction creates a Lua function with the `(behavior, dest_table, ...)` signature that extends the destination table with the
// `extend` function
func extendFunction(
	extend func(behavior luatable.ExtendBehavior, destTable *lua.LTable, tables ...*lua.LTable) error,
	errMessage string,
) lua.LGFunction {
	return func(L *lua.LState) int {
		// Arguments
		behavior := luatable.ExtendBehaviorFromString(L.CheckString(1))
		if behavior == luatable.ExtendBehaviorInvalid {
			return luaerror.Push(L, 0, fmt.Errorf("invalid behavior: %s", L.CheckString(1)))
		}

		destTable := L.CheckTable(2)

		tables := []*lua.LTable{}
		for i := 3; i <= L.GetTop(); i++ {
			tables = append(tables, L.CheckTable(i))
		}

		// Execution
		err := extend(behavior, destTable, tables...)
		if err != nil {
			return luaerror.Push(L, 0, fmt.Errorf("%s: %w", errMessage, err))
		}

		return 0
	}
}
