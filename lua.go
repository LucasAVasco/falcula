package falcula

import (
	"fmt"
	"time"

	"github.com/LucasAVasco/falcula/lua/luaruntime"
	"github.com/LucasAVasco/falcula/lua/modules"
	"github.com/LucasAVasco/falcula/lua/modules/modtui"
)

// runLuaConfig is the configuration required to run a Lua code or file
type runLuaConfig struct {
	Runtime *luaruntime.Runtime
	Code    string
	File    string   // Path to the Lua file
	Args    []string // Arguments provided to the code or file

	// Called after the code or file is executed and before waiting for the user to close the TUI. Optional (can be nil)
	AfterRun func(runtime *luaruntime.Runtime) error
}

// runLuaCode runs a Lua code. Waits for the user to close the TUI if it is visible
func (a *App) runLuaCode(config *runLuaConfig) error {
	// Runs the main script. Repeats the script if the user selects new arguments

	// Loding modules
	loader, err := modules.LoadAllModules(config.Runtime, &modules.AllModulesLoaderOptions{
		RawMode:      a.rawMode,
		OnSelectArgs: func(newArgs []string) {},
	})
	if err != nil {
		config.Runtime.Logger.LogError(fmt.Errorf("error loading modules: %w", err))
		return nil
	}
	defer loader.Close()

	// Running the script
	err = config.Runtime.Run(config.Code, config.Args...)
	if err != nil {
		config.Runtime.Logger.LogError(fmt.Errorf("error running script Lua code: %w", err))
	}

	// Calling the after run callback
	if config.AfterRun != nil {
		err = config.AfterRun(config.Runtime)
		if err != nil {
			return fmt.Errorf("error calling callback after running script: %w", err)
		}
	}

	// Wait until the TUI is closed or the user selects new arguments
	for {
		if modtui.TuiIsVisible() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		break
	}

	return nil
}

// runLuaFile runs a Lua file. Waits for the user to close the TUI if it is visible
func (a *App) runLuaFile(config *runLuaConfig) error {
	// Runs the main script. Repeats the script if the user selects new arguments
	for {
		reRunScript := false
		var loader *modules.Loader

		// Loding modules
		var err error
		loader, err = modules.LoadAllModules(config.Runtime, &modules.AllModulesLoaderOptions{
			RawMode: a.rawMode,
			OnSelectArgs: func(newArgs []string) {
				reRunScript = true
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
		if err != nil {
			config.Runtime.Logger.LogError(fmt.Errorf("error loading modules: %w", err))
			continue // Must not run the script if the modules are not loaded
		}
		defer loader.Close()

		// Running the script
		err = config.Runtime.RunFile(config.File, config.Args...)
		if err != nil {
			config.Runtime.Logger.LogError(fmt.Errorf("error running script '%s': %w", config.File, err))
		}

		// Calling the after run callback
		if config.AfterRun != nil {
			err = config.AfterRun(config.Runtime)
			if err != nil {
				return fmt.Errorf("error calling callback after running script: %w", err)
			}
		}

		// Wait until the TUI is closed or the user selects new arguments
		for {
			if !reRunScript && modtui.TuiIsVisible() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			break
		}

		// Exit if the user closed the TUI without selecting new arguments
		if !reRunScript {
			break
		}
	}

	return nil
}
