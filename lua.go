package falcula

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/LucasAVasco/falcula/lua/luaruntime"
	"github.com/LucasAVasco/falcula/lua/modules"
	"github.com/LucasAVasco/falcula/lua/modules/modtui"
)

// runLuaConfig is the configuration required to run a Lua code or file
type runLuaConfig struct {
	Runtime *luaruntime.Runtime
	Code    string   // Lua code to run. Ignored if `File` is set
	File    string   // Path to the Lua file
	Args    []string // Arguments provided to the code or file

	// Called after the code or file is executed and before waiting for the user to close the TUI. Optional (can be nil)
	AfterRun func(runtime *luaruntime.Runtime) error
}

// runLua runs the Lua code or file of the configuration. Waits for the user to close the TUI if it is visible. A Lua file runs again
// (with the new arguments) each time the user selects new arguments in the TUI
func (a *App) runLua(config *runLuaConfig) error {
	for {
		reRun, err := a.runLuaOnce(config)
		if err != nil || !reRun {
			return err
		}

		// The previous script may have left anything in the Lua state. Runs the next one in a clean state
		config.Runtime.ResetLuaState()
	}
}

// runLuaOnce runs the Lua code or file of the configuration and waits for the user to close the TUI or select new arguments. Returns true
// if the user selected new arguments (the configuration is updated with them).
//
// NOTE(LucasAVasco): the TUI selects the arguments in its own goroutine. It only interrupts the script (stops its services and aborts it),
// because the Lua state can only be closed or reset by this goroutine, after the script returns
func (a *App) runLuaOnce(config *runLuaConfig) (bool, error) {
	var selectedArgs atomic.Pointer[[]string] // Arguments selected by the user to re-run the script. Nil if not selected
	canReRun := config.File != ""             // Only Lua files can be re-run with new arguments

	loader := modules.LoadAllModules(config.Runtime, &modules.AllModulesLoaderOptions{
		RawMode: a.rawMode,
		OnSelectArgs: func(newArgs []string) {
			if !canReRun {
				return
			}

			// NOTE(LucasAVasco): the arguments are only published after the interruption ends, so the next run can not be affected by it
			config.Runtime.Interrupt()
			selectedArgs.Store(&newArgs)
		},
	})
	defer func() {
		err := loader.Close()
		if err != nil {
			config.Runtime.Logger.LogError(fmt.Errorf("error closing Lua modules loader: %w", err))
		}
	}()

	// Running the code or file
	var err error
	if config.File != "" {
		err = config.Runtime.RunFile(config.File, config.Args...)
		if err != nil {
			err = fmt.Errorf("error running Lua file '%s': %w", config.File, err)
		}
	} else {
		err = config.Runtime.Run(config.Code, config.Args...)
		if err != nil {
			err = fmt.Errorf("error running Lua code: %w", err)
		}
	}

	// An interrupted script fails on purpose and must not run its subtask
	interrupted := config.Runtime.Interrupted()
	if err != nil && !interrupted {
		config.Runtime.Logger.LogError(err)
	}

	// Calling the after run callback
	if config.AfterRun != nil && !interrupted {
		err := config.AfterRun(config.Runtime)
		if err != nil {
			return false, fmt.Errorf("error calling callback after running Lua: %w", err)
		}
	}

	// Wait until the user leaves the TUI (or the raw output mode) or selects new arguments
	for selectedArgs.Load() == nil && waitForUser(config.Runtime) {
		time.Sleep(100 * time.Millisecond)
	}

	newArgs := selectedArgs.Load()
	if newArgs == nil {
		return false, nil
	}

	config.File = config.Runtime.GetLastExecutedFile()
	config.Args = *newArgs
	return true, nil
}

// waitForUser returns true while falcula must wait for the user: while the TUI is visible, and in raw output mode while a service is active
// (the user can still open the TUI with Alt+T)
func waitForUser(runtime *luaruntime.Runtime) bool {
	switch modtui.CurrentMode() {
	case modtui.ModeTui:
		return true
	case modtui.ModeRawOutput:
		return runtime.HasActiveServices()
	default:
		return false
	}
}
