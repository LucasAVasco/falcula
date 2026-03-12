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

	"github.com/rivo/tview"
)

// Config is the configuration that the text user interface needs
type Config struct {
	Runtime      *luaruntime.Runtime
	OnSelectArgs func(args []string)
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
}

// New creates a new text user interface
func New(config *Config) (*Tui, error) {
	if config.OnSelectArgs == nil {
		config.OnSelectArgs = func(args []string) {}
	}

	t := Tui{
		config: config,
	}

	return &t, nil
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
		err := t.newApp()
		if err != nil {
			return fmt.Errorf("error creating 'tview' application: %w", err)
		}

		for _, manager := range t.config.Runtime.GetManagers() {
			err := t.AddManagerToSidebar(manager)
			if err != nil {
				return fmt.Errorf("error adding manager to sidebar: %w", err)
			}
		}

		t.SetAvailableScriptArgs(t.config.Runtime.GetScriptAvailableArgs())
		t.SetCurrentScriptArgs(t.config.Runtime.GetScriptCurrentArgs())
	}

	t.waitGroup.Go(func() {
		// Set loggers
		t.config.Runtime.Logger.SetOnServiceLog(t.mainPage.ServiceLogs.Write)
		t.config.Runtime.Logger.SetOnDebugLog(t.mainPage.DebugLogs.Write)
		t.config.Runtime.Logger.SetOnErrorLog(t.mainPage.DebugLogs.WriteError)

		// Run the app
		err := t.app.Run()
		if err != nil {
			t.config.Runtime.Logger.LogError(fmt.Errorf("error running 'tview' application: %w", err))
		}
	})

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
	return t.app.IsRunning()
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
}
