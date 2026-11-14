package modtemplate

import (
	"reflect"
	"text/template"

	"github.com/LucasAVasco/falcula/lua/luaerror"

	lua "github.com/yuin/gopher-lua"
)

// getDataFromLuaTable gets the data to use in a template from a Lua value
func getDataFromLuaTable(L *lua.LState, value lua.LValue) any {
	switch value.Type() {
	case lua.LTNil:
		return nil

	case lua.LTNumber:
		return float64(value.(lua.LNumber))

	case lua.LTBool:
		return bool(value.(lua.LBool))

	case lua.LTTable:
		mapData := make(map[any]any, 0)
		sliceData := make([]any, 0)

		L.ForEach(value.(*lua.LTable), func(key lua.LValue, value lua.LValue) {
			if key.Type() == lua.LTNumber {
				sliceData = append(sliceData, getDataFromLuaTable(L, value))
			} else {
				mapData[key.String()] = getDataFromLuaTable(L, value)
			}
		})

		// Returns the map if there is at least one key, otherwise returns the slice
		if len(mapData) > 0 {
			for index, value := range sliceData {
				// Maps and slices can not be used as map keys. Uses their position (1-based, like Lua) as key instead
				if reflect.TypeOf(value).Comparable() {
					mapData[value] = value
				} else {
					mapData[float64(index+1)] = value
				}
			}

			return mapData
		} else {
			return sliceData
		}

	default:
		return value.String()
	}
}

var moduleFunctions = parseFunctions(1, func(L *lua.LState) template.FuncMap { return nil })

// parseFunctions returns the functions that parse templates (`parse_string`, `parse_file` and `parse_and_save_file`). Their arguments
// start at the `firstArg` index (the index 1 is the `self` argument of methods). The `getFuncs` function returns the extra functions
// available in the templates
func parseFunctions(firstArg int, getFuncs func(L *lua.LState) template.FuncMap) map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"parse_string": func(L *lua.LState) int {
			str := L.ToString(firstArg)
			data := getDataFromLuaTable(L, L.Get(firstArg+1))
			output, err := parseString(str, data, getFuncs(L))
			return pushParseResult(L, output, err)
		},

		"parse_file": func(L *lua.LState) int {
			srcFile := L.ToString(firstArg)
			data := getDataFromLuaTable(L, L.Get(firstArg+1))
			output, err := parseFile(srcFile, data, getFuncs(L))
			return pushParseResult(L, output, err)
		},

		"parse_and_save_file": func(L *lua.LState) int {
			srcFile := L.ToString(firstArg)
			destFile := L.ToString(firstArg + 1)
			data := getDataFromLuaTable(L, L.Get(firstArg+2))
			output, err := parseAndSaveFile(srcFile, destFile, data, getFuncs(L))
			return pushParseResult(L, output, err)
		},
	}
}

// pushParseResult pushes the output of a template or the error
func pushParseResult(L *lua.LState, output string, err error) int {
	if err != nil {
		return luaerror.Push(L, 1, err)
	}

	L.Push(lua.LString(output))
	return 1
}
