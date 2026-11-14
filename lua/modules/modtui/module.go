// Package modtui is a module to work with the text user interface
package modtui

import (
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/modules/base"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui"
	lua "github.com/yuin/gopher-lua"
)

// Config is the configuration that the text user interface module needs
type Config struct {
	RawMode      bool
	CurrentArgs  []string
	OnSelectArgs func(args []string)
}

// Module is a module that provides functions and classes for working with the text user interface
type Module struct {
	base.BaseModule

	tuiConfig *Config
}

func New(config *Config) *Module {
	m := Module{
		tuiConfig: config,
	}

	if t := getTui(); t != nil {
		t.SetCurrentScriptArgs(config.CurrentArgs)
		t.UpdateConfig(&tui.Config{
			OnSelectArgs: config.OnSelectArgs,
		})
	}

	return &m
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	var functions = map[string]lua.LGFunction{
		"show": func(L *lua.LState) int {
			err := m.ShowTui()
			if err != nil {
				return luaerror.Push(L, 0, fmt.Errorf("error showing text user interface: %w", err))
			}
			return 0
		},

		"hide": func(L *lua.LState) int {
			HideTui()
			return 0
		},
	}

	L.SetFuncs(mod, functions)
	return nil
}

// GetTui returns the text user interface instance
func (m *Module) GetTui() *tui.Tui {
	return getTui()
}

// ShowTui shows the text user interface.
func (m *Module) ShowTui() error {
	// Don't show the text user interface in raw mode
	if m.tuiConfig.RawMode {
		return nil
	}

	// Show the text user interface (creates it if it does not exist)
	err := showTui(func() *tui.Tui {
		return tui.New(&tui.Config{
			Runtime:      m.Config.Runtime,
			OnSelectArgs: m.tuiConfig.OnSelectArgs,
			OnRawOutput:  enterRawOutput,
			OnQuit:       quit,
		})
	})
	if err != nil {
		return fmt.Errorf("error showing text user interface: %w", err)
	}

	return nil
}

// Close closes the module. Can be called multiple times
func (m *Module) Close() error {
	if t := getTui(); t != nil {
		t.RemoveAllManagersFromSidebar()
	}

	return nil
}
