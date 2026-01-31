package modtemplate

import (
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
			if len(sliceData) > 0 {
				for _, value := range sliceData {
					mapData[value] = value
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

var moduleFunctions = map[string]lua.LGFunction{
	"parse_string": func(L *lua.LState) int {
		str := L.ToString(1)
		data := getDataFromLuaTable(L, L.Get(2))
		output, err := parseString(str, data, nil)
		if err != nil {
			return luaerror.Push(L, 1, err)
		}

		L.Push(lua.LString(output))
		return 1
	},

	"parse_file": func(L *lua.LState) int {
		srcFile := L.ToString(1)
		data := getDataFromLuaTable(L, L.Get(2))
		output, err := parseFile(srcFile, data, nil)
		if err != nil {
			return luaerror.Push(L, 1, err)
		}

		L.Push(lua.LString(output))
		return 1
	},

	"parse_and_save_file": func(L *lua.LState) int {
		srcFile := L.ToString(1)
		destFile := L.ToString(2)
		data := getDataFromLuaTable(L, L.Get(3))
		output, err := parseAndSaveFile(srcFile, destFile, data, nil)
		if err != nil {
			return luaerror.Push(L, 1, err)
		}

		L.Push(lua.LString(output))
		return 1
	},
}
