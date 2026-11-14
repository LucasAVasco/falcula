// Package luadata is a package that contains functions for working with Lua user data
package luadata

import (
	"github.com/yuin/gopher-lua"
)

func NewUserData(L *lua.LState, value any) *lua.LUserData {
	data := L.NewUserData()
	data.Value = value

	return data
}

// GetValueFromArgs returns the value of the user data argument at the given index. Raises a Lua error if the argument is not a user data
func GetValueFromArgs(L *lua.LState, index int) any {
	return L.CheckUserData(index).Value
}
