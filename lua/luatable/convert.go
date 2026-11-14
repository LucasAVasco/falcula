// Package luatable is a collection of utilities functions related to Lua tables
package luatable

import (
	"github.com/yuin/gopher-lua"
)

// GetValuesFromLuaTable returns all values from the (key, value) pairs of a Lua table
func GetValuesFromLuaTable(table *lua.LTable) []lua.LValue {
	slice := make([]lua.LValue, 0, table.Len())
	table.ForEach(func(key lua.LValue, value lua.LValue) {
		slice = append(slice, value)
	})

	return slice
}

// GetStringsFromLuaTable returns all values from the (key, value) pairs of a Lua table as strings
func GetStringsFromLuaTable(table *lua.LTable) []string {
	result := make([]string, 0, table.Len())
	table.ForEach(func(key lua.LValue, value lua.LValue) {
		result = append(result, value.String()) // Same as `tostring`. Supports numbers
	})

	return result
}

// GetStringsFromLuaTableThatKeyIsNumber returns all values from the (key, value) pairs of a Lua table as strings where the key is a number.
// If the key is not a number, then it is ignored
func GetStringsFromLuaTableThatKeyIsNumber(table *lua.LTable) []string {
	result := make([]string, 0, table.Len())
	table.ForEach(func(key lua.LValue, value lua.LValue) {
		if key.Type() != lua.LTNumber {
			return
		}

		result = append(result, value.String()) // Same as `tostring`. Supports numbers
	})

	return result
}

// GetLuaTableFromStrings returns a Lua table from a slice of strings
func GetLuaTableFromStrings(L *lua.LState, stringList []string) *lua.LTable {
	table := L.NewTable()

	for _, value := range stringList {
		table.Append(lua.LString(value))
	}

	return table
}
