// Package falcula implements the function to start Falcula CLI
package falcula

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LucasAVasco/falcula/lua/luaruntime"
	"github.com/LucasAVasco/falcula/lua/modules"
	"github.com/LucasAVasco/falcula/lua/modules/modtui"
)

// getMainScriptPath returns the path to the main script to run
func getMainScriptPath() (string, error) {
	possibleFiles := []string{"falcula.lua", "falcfg/init.lua", ".falcula.lua", ".falcfg/init.lua"}
	path := ""
	for _, possibleFile := range possibleFiles {
		if _, err := os.Stat(possibleFile); err == nil {
			path = possibleFile
			break
		}
	}
	if path == "" {
		return "", fmt.Errorf("can not find the main file, possible names: %s", strings.Join(possibleFiles, ", "))
	}

	return path, nil
}

// StartCli starts the Falcula CLI
func StartCli() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("must be called with 'run' or 'run-raw', but does not receive any argument")
	}

	if os.Args[1] != "run" && os.Args[1] != "run-raw" {
		return fmt.Errorf("must be called with 'run' or 'run-raw', but received '%s'", os.Args[1])
	}

	rawMode := os.Args[1] == "run-raw"

	runtime, err := luaruntime.New()
	if err != nil {
		return fmt.Errorf("error creating runtime: %w", err)
	}
	defer runtime.Close()

	// modtui does not closes the TUI when it is closed. The TUI is persistent across runs. Need to close it manually
	defer modtui.ClosePersistentTui()

	// Arguments required to run the main script
	scriptToRun, err := getMainScriptPath()
	if err != nil {
		return fmt.Errorf("error getting path of the script to run: %w", err)
	}

	args := os.Args[2:]

	// Runs the main script. Repeats the script if the user selects new arguments
	for {
		reRunScript := false
		var loader *modules.Loader

		// Loding modules
		loader, err = modules.LoadAllModules(runtime, &modules.AllModulesLoaderOptions{
			RawMode: rawMode,
			OnSelectArgs: func(newArgs []string) {
				reRunScript = true
				scriptToRun = runtime.GetLastExecutedFile()
				args = newArgs

				err := loader.Close()
				if err != nil {
					runtime.Logger.LogError(fmt.Errorf("error closing Lua modules loader: %w", err))
					runtime.CloseLuaState()
					return
				}

				runtime.ResetLuaState()
			},
		})
		if err != nil {
			runtime.Logger.LogError(fmt.Errorf("error loading modules: %w", err))
			continue // Must not run the script if the modules are not loaded
		}
		defer loader.Close()

		// Running the script
		err = runtime.RunScript(scriptToRun, args...)
		if err != nil {
			runtime.Logger.LogError(fmt.Errorf("error running script '%s': %w", scriptToRun, err))
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
