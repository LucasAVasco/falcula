package modtemplate

import (
	"fmt"
	"text/template"

	"github.com/LucasAVasco/falcula/lua/luaclass"

	lua "github.com/yuin/gopher-lua"
)

// getTemplateFuncs gets the functions of the template object when called inside a method
func getTemplateFuncs(L *lua.LState) template.FuncMap {
	return luaclass.GetAttribute(L, "_funcs").(template.FuncMap)
}

var templateClassFunctions = func() map[string]lua.LGFunction {
	functions := parseFunctions(2, getTemplateFuncs)

	functions["set_func"] = func(L *lua.LState) int {
		funcs := getTemplateFuncs(L)
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
	}

	return functions
}()
