// Package luatable is a collection of utilities functions related to Lua tables
package luatable

import "github.com/yuin/gopher-lua"

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
	strs := make([]string, 0, table.Len())
	table.ForEach(func(key lua.LValue, value lua.LValue) {
		strs = append(strs, value.(lua.LString).String())
	})

	return strs
}

// GetLuaTableFromValues returns a Lua table from a slice of Lua values
func GetLuaTableFromValues(L *lua.LState, values []lua.LValue) *lua.LTable {
	table := L.NewTable()

	for _, value := range values {
		table.Append(value)
	}

	return table
}

// GetLuaTableFromStrings returns a Lua table from a slice of strings
func GetLuaTableFromStrings(L *lua.LState, stringList []string) *lua.LTable {
	table := L.NewTable()

	for _, value := range stringList {
		table.Append(lua.LString(value))
	}

	return table
}
