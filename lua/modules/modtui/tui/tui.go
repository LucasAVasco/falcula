// Package tui implements the text user interface
package tui

import (
	"fmt"
	"sync"

	"github.com/LucasAVasco/falcula/lua/luaruntime"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/app"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/argsview"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/help"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/mainpage"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/theme"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Config is the configuration that the text user interface needs
type Config struct {
	Runtime      *luaruntime.Runtime
	OnSelectArgs func(args []string)

	// Called when the user chooses to leave the text user interface. Run in their own goroutine. If not provided, they only hide it
	OnRawOutput func() // Switch to raw output mode (the services keep running)
	OnQuit      func() // Quit falcula

	// Colors of the text user interface. Optional: uses the default theme if nil
	Theme *theme.Theme

	// Screen used by the text user interface. Optional: uses the terminal if nil (tests use a simulation screen)
	Screen tcell.Screen
}

// Tui is the text user interface widget.
type Tui struct {
	config *Config

	mutex     sync.Mutex
	waitGroup sync.WaitGroup

	// User interface

	app      *app.App
	pages    *tview.Pages
	mainPage *mainpage.MainPage
	argsView *argsview.ArgsView
	help     *help.HelpWidget

	exitDialog            *tview.Modal
	focusBeforeExitDialog tview.Primitive // Focused widget when the exit dialog was opened (restored when it closes)
}

// New creates a new text user interface
func New(config *Config) *Tui {
	t := &Tui{config: config}

	if config.Theme == nil {
		config.Theme = theme.Default()
	}
	if config.OnSelectArgs == nil {
		config.OnSelectArgs = func(args []string) {}
	}
	if config.OnRawOutput == nil {
		config.OnRawOutput = t.Hide
	}
	if config.OnQuit == nil {
		config.OnQuit = t.Hide
	}

	return t
}

// Close the text user interface. Can be called multiple times.
func (t *Tui) Close() {
	t.Hide()
}

// Show the text user interface. This is a non-blocking call and is idempotent
func (t *Tui) Show() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.app == nil {
		t.newApp()

		for _, manager := range t.config.Runtime.GetManagers() {
			err := t.AddManagerToSidebar(manager)
			if err != nil {
				return fmt.Errorf("error adding manager to sidebar: %w", err)
			}
		}

		t.SetAvailableScriptArgs(t.config.Runtime.GetScriptAvailableArgs())
		t.SetCurrentScriptArgs(t.config.Runtime.GetScriptCurrentArgs())
	}

	// The logs are shown in the text user interface
	t.config.Runtime.Logger.SetOnServiceLog(t.mainPage.ServiceLogs.Write)
	t.config.Runtime.Logger.SetOnDebugLog(t.mainPage.DebugLogs.Write)
	t.config.Runtime.Logger.SetOnErrorLog(t.mainPage.DebugLogs.WriteError)

	// Runs the app. It is already marked as running when `Start` returns, so `Hide` always stops it
	t.waitGroup.Add(1)
	started := t.app.Start(func(err error) {
		defer t.waitGroup.Done()
		if err != nil {
			t.config.Runtime.Logger.LogError(fmt.Errorf("error running 'tview' application: %w", err))
			go t.config.OnQuit() // The user can not use the text user interface (e.g. there is no terminal)
		}
	})
	if !started {
		t.waitGroup.Done() // Already running
	}

	return nil
}

// Hide the text user interface. This is a non-blocking call and is idempotent
func (t *Tui) Hide() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.app == nil {
		return
	}

	t.config.Runtime.Logger.ResetLoggers()
	t.app.Stop()
}

// IsVisible checks if the text user interface is visible
func (t *Tui) IsVisible() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.app != nil && t.app.IsRunning()
}

// WaitForHide waits until the text user interface is hidden
func (t *Tui) WaitForHide() {
	t.waitGroup.Wait()
}

// UpdateConfig updates the configuration of the text user interface
func (t *Tui) UpdateConfig(config *Config) {
	if config.Runtime != nil {
		t.config.Runtime = config.Runtime
	}

	if config.OnSelectArgs != nil {
		t.config.OnSelectArgs = config.OnSelectArgs
	}

	if config.OnRawOutput != nil {
		t.config.OnRawOutput = config.OnRawOutput
	}

	if config.OnQuit != nil {
		t.config.OnQuit = config.OnQuit
	}

	if config.Theme != nil {
		t.config.Theme = config.Theme // Only affects the widgets created after this call
	}
}
