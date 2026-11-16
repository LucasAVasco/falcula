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
	cmdArgs *lua.LTable // The command line arguments
}

func New() *Module {
	return &Module{
		cmdArgs: &lua.LTable{Metatable: lua.LNil}, // NOTE(LucasAVasco): a zero value table has a nil meta-table, which is invalid
	}
}

func (m *Module) SetCurrentScriptArgs(args []string) {
	// Removes the old arguments (collects the keys first, because the table must not be modified while iterating it)
	keys := []lua.LValue{}
	m.cmdArgs.ForEach(func(key, _ lua.LValue) {
		keys = append(keys, key)
	})
	for _, key := range keys {
		m.cmdArgs.RawSet(key, lua.LNil)
	}

	// Sets the new arguments
	for _, arg := range args {
		m.cmdArgs.Append(lua.LString(arg))
	}
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	L.SetField(mod, "args", m.cmdArgs)

	// NOTE(LucasAVasco): the available arguments are the `args` option of the task in the project configuration file
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
