// Package modcmd is a module that provides command line arguments to Falcula
package modcmd

import (
	"github.com/LucasAVasco/falcula/lua/luatable"
	"github.com/LucasAVasco/falcula/lua/modules/base"

	lua "github.com/yuin/gopher-lua"
)

// Module is a module that provides command line arguments to Falcula
type Module struct {
	base.BaseModule
	cmdArgs lua.LTable // The command line arguments
}

func New() *Module {
	m := Module{}

	m.cmdArgs = lua.LTable{}

	return &m
}

func (m *Module) SetCurrentScriptArgs(args []string) {
	// Removes the old arguments
	m.cmdArgs.ForEach(func(l1, l2 lua.LValue) {
		m.cmdArgs.RawSet(l1, lua.LNil)
	})

	// Sets the new arguments
	for _, arg := range args {
		m.cmdArgs.Append(lua.LString(arg))
	}
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	L.SetField(mod, "args", &m.cmdArgs)

	L.SetField(mod, "set_available_args", L.NewFunction(func(L *lua.LState) int {
		argsList := L.ToTable(1)
		availableCmdArgs := [][]string{}

		for i := 0; i < argsList.Len(); i++ {
			args := argsList.RawGetInt(i + 1).(*lua.LTable)

			availableCmdArgs = append(availableCmdArgs, luatable.GetStringsFromLuaTable(args))
		}

		m.Config.Runtime.SetScriptAvailableArgs(availableCmdArgs)

		return 0
	}))

	L.SetField(mod, "get_available_args", L.NewFunction(func(L *lua.LState) int {
		newTable := L.NewTable()

		for _, args := range m.Config.Runtime.GetScriptAvailableArgs() {
			newTable.Append(luatable.GetLuaTableFromStrings(L, args))
		}

		L.Push(newTable)
		return 1
	}))

	return nil
}
