package modtemplate

import (
	"fmt"
	"text/template"

	"github.com/LucasAVasco/falcula/lua/luaclass"
	"github.com/LucasAVasco/falcula/lua/luaerror"

	lua "github.com/yuin/gopher-lua"
)

var templateClassFunctions = map[string]lua.LGFunction{
	"set_func": func(L *lua.LState) int {
		funcs := luaclass.GetAttribute(L, "_funcs").(template.FuncMap)
		functionName := L.ToString(2)
		function := L.ToFunction(3)

		funcs[functionName] = func(args ...any) string {
			// Function to call
			L.Push(function)

			// Pushes the arguments
			for _, arg := range args {
				L.Push(lua.LString(fmt.Sprintf("%v", arg)))
			}

			// Calls the function
			L.Call(len(args), 1)

			// Returns the result
			return L.ToString(-1)
		}
		return 0
	},

	"parse_string": func(L *lua.LState) int {
		str := L.ToString(2)
		data := getDataFromLuaTable(L, L.Get(3))
		funcs := luaclass.GetAttribute(L, "_funcs").(template.FuncMap)
		output, err := parseString(str, data, funcs)
		if err != nil {
			return luaerror.Push(L, 1, err)
		}

		L.Push(lua.LString(output))
		return 1
	},

	"parse_file": func(L *lua.LState) int {
		srcFile := L.ToString(2)
		data := getDataFromLuaTable(L, L.Get(3))
		funcs := luaclass.GetAttribute(L, "_funcs").(template.FuncMap)
		output, err := parseFile(srcFile, data, funcs)
		if err != nil {
			return luaerror.Push(L, 1, err)
		}

		L.Push(lua.LString(output))
		return 1
	},

	"parse_and_save_file": func(L *lua.LState) int {
		srcFile := L.ToString(2)
		destFile := L.ToString(3)
		data := getDataFromLuaTable(L, L.Get(4))
		funcs := luaclass.GetAttribute(L, "_funcs").(template.FuncMap)
		output, err := parseAndSaveFile(srcFile, destFile, data, funcs)
		if err != nil {
			return luaerror.Push(L, 1, err)
		}

		L.Push(lua.LString(output))
		return 1
	},
}
