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
		"extend": func(L *lua.LState) int {
			// Arguments
			behavior := luatable.ExtendBehaviorFromString(L.CheckString(1))
			if behavior == luatable.ExtendBehaviorInvalid {
				luaerror.Push(L, 1, fmt.Errorf("invalid behavior: %s", L.CheckString(1)))
				return 1
			}

			destTable := L.CheckTable(2)

			tables := []*lua.LTable{}
			for i := 3; i <= L.GetTop(); i++ {
				tables = append(tables, L.CheckTable(i))
			}

			// Execution
			err := luatable.Extend(behavior, destTable, tables...)
			if err != nil {
				luaerror.Push(L, 1, fmt.Errorf("error extending table: %w", err))
			}

			return 0
		},

		"deep_extend": func(L *lua.LState) int {
			// Arguments
			behavior := luatable.ExtendBehaviorFromString(L.CheckString(1))
			if behavior == luatable.ExtendBehaviorInvalid {
				luaerror.Push(L, 1, fmt.Errorf("invalid behavior: %s", L.CheckString(1)))
				return 1
			}

			destTable := L.CheckTable(2)

			tables := []*lua.LTable{}
			for i := 3; i <= L.GetTop(); i++ {
				tables = append(tables, L.CheckTable(i))
			}

			// Execution
			err := luatable.DeepExtend(behavior, destTable, tables...)
			if err != nil {
				luaerror.Push(L, 1, fmt.Errorf("error deep extending table: %w", err))
			}

			return 0
		},
	})

	return nil
}
