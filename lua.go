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
	}
}

// runLuaOnce runs the Lua code or file of the configuration and waits for the user to close the TUI or select new arguments. Returns true
// if the user selected new arguments (the configuration is updated with them)
func (a *App) runLuaOnce(config *runLuaConfig) (bool, error) {
	var reRun atomic.Bool
	var loader *modules.Loader

	loader = modules.LoadAllModules(config.Runtime, &modules.AllModulesLoaderOptions{
		RawMode: a.rawMode,
		OnSelectArgs: func(newArgs []string) {
			// Only Lua files can be re-run with new arguments
			if config.File == "" {
				return
			}
			defer reRun.Store(true)

			config.File = config.Runtime.GetLastExecutedFile()
			config.Args = newArgs

			err := loader.Close()
			if err != nil {
				config.Runtime.Logger.LogError(fmt.Errorf("error closing Lua modules loader: %w", err))
				config.Runtime.CloseLuaState()
				return
			}

			config.Runtime.ResetLuaState()
		},
	})
	defer loader.Close()

	// Running the code or file
	if config.File != "" {
		err := config.Runtime.RunFile(config.File, config.Args...)
		if err != nil {
			config.Runtime.Logger.LogError(fmt.Errorf("error running Lua file '%s': %w", config.File, err))
		}
	} else {
		err := config.Runtime.Run(config.Code, config.Args...)
		if err != nil {
			config.Runtime.Logger.LogError(fmt.Errorf("error running Lua code: %w", err))
		}
	}

	// Calling the after run callback
	if config.AfterRun != nil {
		err := config.AfterRun(config.Runtime)
		if err != nil {
			return false, fmt.Errorf("error calling callback after running Lua: %w", err)
		}
	}

	// Wait until the TUI is closed or the user selects new arguments
	for !reRun.Load() && modtui.IsTuiVisible() {
		time.Sleep(100 * time.Millisecond)
	}

	return reRun.Load(), nil
}
