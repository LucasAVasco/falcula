// Package modtui is a modules to work with the text user interface
package modtui

import (
	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/modules/base"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui"
	lua "github.com/yuin/gopher-lua"
)

// NOTE(LucasAVasco): executing a script will create a new instance of the text user interface module, but the text user interface should be
// preserved throughout the execution of the script. This is why we use a global variable to store the text user interface instance
var serviceTui *tui.Tui

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

// GetTui returns the text user interface instance
func (m *Module) GetTui() *tui.Tui {
	return serviceTui
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	var functions = map[string]lua.LGFunction{
		"show": func(l *lua.LState) int {
			// Don't show the text user interface in raw mode
			if m.config.RawMode {
				return 0
			}

			// Ensure that the text user interface is created
			if serviceTui == nil {
				var err error
				serviceTui, err = tui.New(&tui.Config{
					Runtime:      m.Config.Runtime,
					OnSelectArgs: m.config.OnSelectArgs,
				})
				if err != nil {
					return luaerror.Push(l, 0, err)
				}
			}

			// Show the text user interface
			serviceTui.Show()
			return 0
		},

		"hide": func(l *lua.LState) int {
			if serviceTui != nil {
				serviceTui.Hide()
			}
			return 0
		},
	}

	L.SetFuncs(mod, functions)
	return nil
}

// Close closes the module. Can be called multiple times
func (m *Module) Close() error {
	if serviceTui != nil {
		serviceTui.RemoveAllManagersFromSidebar()
	}

	return nil
}

// TuiIsVisible checks if the text user interface is visible
func TuiIsVisible() bool {
	if serviceTui == nil {
		return false
	}

	return serviceTui.IsVisible()
}

// ClosePersistentTui closes (deletes) the persistent text user interface
func ClosePersistentTui() {
	if serviceTui != nil {
		serviceTui.Close()
		serviceTui = nil
	}
}

// WaitForTuiHide waits until the text user interface is hidden
func WaitForTuiHide() {
	if serviceTui == nil {
		return
	}

	serviceTui.WaitForHide()
}
