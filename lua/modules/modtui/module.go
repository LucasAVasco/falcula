// Package modtui is a module to work with the text user interface
package modtui

import (
	"fmt"
	"sync"

	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/modules/base"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/theme"
	lua "github.com/yuin/gopher-lua"
)

// Config is the configuration that the text user interface module needs
type Config struct {
	RawMode      bool
	Theme        *theme.Theme // Optional: the default theme if nil
	ThemeError   error        // Error loading the configured theme (reported in the debug logs when the TUI is shown). Optional
	CurrentArgs  []string
	OnSelectArgs func(args []string)
}

// reportThemeError reports the theme error only once (the text user interface is shown again after each script re-run)
var reportThemeError sync.Once

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

		"show_args": func(L *lua.LState) int {
			err := m.ShowArgsPage()
			if err != nil {
				return luaerror.Push(L, 0, fmt.Errorf("error showing the arguments page: %w", err))
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

// ShowTui shows the text user interface (does nothing in raw mode)
func (m *Module) ShowTui() error {
	return m.showTui(false)
}

// ShowArgsPage shows the text user interface on the arguments page (does nothing in raw mode)
func (m *Module) ShowArgsPage() error {
	return m.showTui(true)
}

// showTui shows the text user interface. If `argsPage` is true, opens the arguments page
func (m *Module) showTui(argsPage bool) error {
	// Don't show the text user interface in raw mode
	if m.tuiConfig.RawMode {
		return nil
	}

	// Show the text user interface (creates it if it does not exist)
	err := showTui(func() *tui.Tui {
		return tui.New(&tui.Config{
			Runtime:      m.Config.Runtime,
			OnSelectArgs: m.tuiConfig.OnSelectArgs,
			Theme:        m.tuiConfig.Theme,
			OnRawOutput:  enterRawOutput,
			OnQuit:       quit,
		})
	})
	if err != nil {
		return fmt.Errorf("error showing text user interface: %w", err)
	}

	if argsPage {
		if t := getTui(); t != nil {
			t.OpenArgumentsPage()
		}
	}

	// Shown in the debug logs of the text user interface (it uses the default theme)
	if m.tuiConfig.ThemeError != nil {
		reportThemeError.Do(func() {
			m.Config.Runtime.Logger.LogError(fmt.Errorf("using the default theme: %w", m.tuiConfig.ThemeError))
		})
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
