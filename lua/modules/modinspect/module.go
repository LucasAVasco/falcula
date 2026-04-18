// Package modinspect is a module that provides functions and classes for inspecting values
package modinspect

import (
	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/luainspect"
	lua "github.com/yuin/gopher-lua"
)

// LoadFunction loads the inspect function
func LoadFunction(name string, L *lua.LState) int {
	function := L.NewFunction(func(l *lua.LState) int {
		value := L.Get(1)

		result, err := luainspect.Inspect(value)
		if err != nil {
			return luaerror.Push(L, 1, err)
		}

		L.Push(lua.LString(result))
		return 1
	})
	L.Push(function)
	return 1
}
