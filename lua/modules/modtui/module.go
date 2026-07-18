// Package modtui is a modules to work with the text user interface
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

	config *Config
}

func New(config *Config) *Module {
	m := Module{
		config: config,
	}

	if serviceTui != nil {
		serviceTui.SetCurrentScriptArgs(config.CurrentArgs)
		serviceTui.UpdateConfig(&tui.Config{
			OnSelectArgs: config.OnSelectArgs,
		})
	}

	return &m
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	var functions = map[string]lua.LGFunction{
		"show": func(l *lua.LState) int {
			err := m.ShowTui()
			if err != nil {
				luaerror.Push(L, 0, fmt.Errorf("error showing text user interface: %w", err))
			}
			return 0
		},

		"hide": func(l *lua.LState) int {
			HideTui()
			return 0
		},
	}

	L.SetFuncs(mod, functions)
	return nil
}

// GetTui returns the text user interface instance
func (m *Module) GetTui() *tui.Tui {
	return serviceTui
}

func (m *Module) IsTuiVisible() bool {
	return IsTuiVisible()
}

// ShowTui shows the text user interface.
func (m *Module) ShowTui() error {
	// Don't show the text user interface in raw mode
	if m.config.RawMode {
		return nil
	}

	// Ensure that the text user interface is created
	if serviceTui == nil {
		var err error
		serviceTui, err = tui.New(&tui.Config{
			Runtime:      m.Config.Runtime,
			OnSelectArgs: m.config.OnSelectArgs,
		})
		if err != nil {
			return fmt.Errorf("error creating text user interface: %w", err)
		}
	}

	// Show the text user interface
	err := serviceTui.Show()
	if err != nil {
		return fmt.Errorf("error showing text user interface: %w", err)
	}

	return nil
}

// HideTui hides the text user interface.
func (m *Module) HideTui() {
	HideTui()
}

// WaitForTuiHide waits until the text user interface is hidden.
func (m *Module) WaitForTuiHide() {
	WaitForTuiHide()
}

// Close closes the module. Can be called multiple times
func (m *Module) Close() error {
	if serviceTui != nil {
		serviceTui.RemoveAllManagersFromSidebar()
	}

	return nil
}
