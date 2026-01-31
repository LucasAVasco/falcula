// Package modprocess is a module that provides functions and classes for working with system processes
package modprocess

import (
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luaclass"
	"github.com/LucasAVasco/falcula/lua/luadata"
	"github.com/LucasAVasco/falcula/lua/luatable"
	"github.com/LucasAVasco/falcula/lua/modules/base"
	"github.com/LucasAVasco/falcula/provider/process"

	lua "github.com/yuin/gopher-lua"
)

// Loader is a module that provides functions and classes for working with system processes
type Loader struct {
	base.BaseModule
}

func New() *Loader {
	return &Loader{}
}

func getProcessCommandFromLuaValue(value lua.LValue) (*process.Command, error) {
	command := process.Command{}

	switch value.Type() {
	case lua.LTNil:
		return nil, nil

	case lua.LTString:
		command.Shell = true
		command.Command = []string{value.(*lua.LString).String()}

	case lua.LTTable:
		command.Shell = false
		command.Command = luatable.GetStringsFromLuaTable(value.(*lua.LTable))

	default:
		return nil, fmt.Errorf("invalid command type: %T", value)
	}

	return &command, nil
}

func (l *Loader) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	info := luaclass.Info{
		Name: "Provider",
		Constructor: func(L *lua.LState, newObj *lua.LTable) error {
			name := L.ToString(2)
			prepareCmd, err := getProcessCommandFromLuaValue(L.Get(3))
			if err != nil {
				return fmt.Errorf("error getting prepare command: %w", err)
			}

			mainCmd, err := getProcessCommandFromLuaValue(L.Get(4))
			if err != nil {
				return fmt.Errorf("error getting main command: %w", err)
			}

			// Sets the provider in the instance
			provider := process.New(l.Opts.Multiplexer, name, prepareCmd, mainCmd)
			luaclass.SetAttribute(L, newObj, "_provider", provider)

			return nil
		},
		Methods: methods,
	}
	class, err := luaclass.New(L, &info, l.Opts.OnError)
	if err != nil {
		return fmt.Errorf("error creating class '%s' of '%s' module: %w", info.Name, name, err)
	}

	L.SetField(mod, info.Name, class)

	return nil
}

// getProvider gets the process provider when called inside a method. Must not be used outside a method
func getProvider(L *lua.LState) *process.Provider {
	return luaclass.GetAttribute(L, "_provider").(*process.Provider)
}

var methods = map[string]lua.LGFunction{
	"get_name": func(L *lua.LState) int {
		provider := getProvider(L)
		L.Push(lua.LString(provider.GetName()))
		return 1
	},

	"new_service": func(L *lua.LState) int {
		provider := getProvider(L)
		L.Push(luadata.NewUserData(L, provider.NewService()))
		return 1
	},
}
