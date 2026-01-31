// Package modcmd is a module that provides command line arguments to Falcula
package modcmd

import (
	"github.com/LucasAVasco/falcula/lua/luatable"
	"github.com/LucasAVasco/falcula/lua/modules/base"

	lua "github.com/yuin/gopher-lua"
)

type Callbacks struct {
	OnSetAvailableCmdArgs func(args [][]string) // Called when the available command arguments are set from Lua
}

// Module is a module that provides command line arguments to Falcula
type Module struct {
	base.BaseModule
	cmdArgs          lua.LTable // The command line arguments
	availableCmdArgs [][]string // The available command line arguments (possible choices)
	callbacks        *Callbacks
}

func New(args []string, callbacks *Callbacks) *Module {
	m := Module{
		availableCmdArgs: make([][]string, 0),
		callbacks:        callbacks,
	}

	if m.callbacks == nil {
		m.callbacks = &Callbacks{}
	}

	if m.callbacks.OnSetAvailableCmdArgs == nil {
		m.callbacks.OnSetAvailableCmdArgs = func(args [][]string) {}
	}

	m.setCurrentArgs(args)

	return &m
}

func (m *Module) setCurrentArgs(args []string) {
	m.cmdArgs = lua.LTable{}

	for _, arg := range args {
		m.cmdArgs.Append(lua.LString(arg))
	}
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	L.SetField(mod, "args", &m.cmdArgs)

	L.SetField(mod, "set_available_args", L.NewFunction(func(L *lua.LState) int {
		argsList := L.ToTable(1)

		for i := 0; i < argsList.Len(); i++ {
			args := argsList.RawGetInt(i + 1).(*lua.LTable)

			m.availableCmdArgs = append(m.availableCmdArgs, luatable.GetStringsFromLuaTable(args))
		}

		m.callbacks.OnSetAvailableCmdArgs(m.availableCmdArgs)

		return 0
	}))

	L.SetField(mod, "get_available_args", L.NewFunction(func(L *lua.LState) int {
		newTable := L.NewTable()

		for _, args := range m.availableCmdArgs {
			newTable.Append(luatable.GetLuaTableFromStrings(L, args))
		}

		L.Push(newTable)
		return 1
	}))

	return nil
}
