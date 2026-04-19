// Package modjson is a module that provides functions to work with JSON
package modjson

import (
	"encoding/json"
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/maplua"
	"github.com/LucasAVasco/falcula/lua/modules/base"

	lua "github.com/yuin/gopher-lua"
)

// Module is a module that provides functions to work with JSON
type Module struct {
	base.BaseModule
}

func New() *Module {
	return &Module{}
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	L.SetFuncs(mod, map[string]lua.LGFunction{
		"encode": func(L *lua.LState) int {
			luaValue := L.Get(1)

			// Lua to Go object
			var goValue any
			err := maplua.Unmarshal(luaValue, &goValue)
			if err != nil {
				return luaerror.Push(L, 1, fmt.Errorf("error mapping Lua object to Go object: %w", err))
			}

			// Go object to JSON string
			jsonValue, err := json.Marshal(goValue)
			if err != nil {
				return luaerror.Push(L, 1, fmt.Errorf("error encoding Go object to JSON: %w", err))
			}

			L.Push(lua.LString(jsonValue))
			return 1
		},

		"decode": func(L *lua.LState) int {
			jsonValue := L.ToString(1)

			// JSON to Go object
			var goValue any
			err := json.Unmarshal([]byte(jsonValue), &goValue)
			if err != nil {
				return luaerror.Push(L, 1, fmt.Errorf("error decoding JSON to Go object: %w", err))
			}

			// Go object to Lua
			luaValue, err := maplua.Marshal(goValue)
			if err != nil {
				return luaerror.Push(L, 1, fmt.Errorf("error mapping Golang object to Lua object: %w", err))
			}

			L.Push(luaValue)
			return 1
		},
	})

	return nil
}
