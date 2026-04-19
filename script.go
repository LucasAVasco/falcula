package falcula

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/LucasAVasco/falcula/lua/luaruntime"
	"github.com/LucasAVasco/falcula/lua/modules"
	"github.com/LucasAVasco/falcula/lua/modules/modtui"
	"github.com/LucasAVasco/falcula/process"
)

// RunScript runs a script of the current project with the given arguments. The scriptName can be the name of a named script or the path to
// a script
func (a *App) RunScript(scriptName string, args ...string) error {
	// Lua runtime
	runtime, err := luaruntime.New()
	if err != nil {
		return fmt.Errorf("error creating runtime: %w", err)
	}
	defer runtime.Close()

	// modtui does not closes the TUI when it is closed. The TUI is persistent across runs. Need to close it manually
	defer modtui.ClosePersistentTui()

	// Script to run
	script, err := a.project.GetScriptByName(scriptName)
	if err != nil {
		return fmt.Errorf("error getting script to run: %w", err)
	}

	// Changes to script directory
	err = os.Chdir(script.Cwd)
	if err != nil {
		return fmt.Errorf("error changing to script working directory: %w", err)
	}

	configureCmd := func(cmd *exec.Cmd) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "FALCULA_PROJECT_DIR="+a.project.Folder)
		cmd.Env = append(cmd.Env, "FALCULA_INVOKE_DIR="+a.invokeDir)
	}

	if script.Command.IsNotEmpty() {
		var cmd *exec.Cmd
		if script.Command.List != nil {
			cmd = process.CreateCmd(false, script.Command.List[0], script.Command.List[1:]...)
		} else {
			cmd = process.CreateCmd(true, script.Command.String)
		}
		configureCmd(cmd)

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error running command: %w", err)
		}

	} else if script.Lua != "" {
		// Runs the main script. Repeats the script if the user selects new arguments
		var loader *modules.Loader

		// Loding modules
		loader, err = modules.LoadAllModules(runtime, &modules.AllModulesLoaderOptions{
			RawMode:      a.rawMode,
			OnSelectArgs: func(newArgs []string) {},
		})
		if err != nil {
			runtime.Logger.LogError(fmt.Errorf("error loading modules: %w", err))
			return nil
		}
		defer loader.Close()

		// Running the script
		err = runtime.Run(script.Lua, args...)
		if err != nil {
			runtime.Logger.LogError(fmt.Errorf("error running script Lua code: %w", err))
		}

		// Wait until the TUI is closed or the user selects new arguments
		for {
			if modtui.TuiIsVisible() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			break
		}

	} else if script.File != "" {
		cmd := process.CreateCmd(false, script.File)
		configureCmd(cmd)

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error running command: %w", err)
		}

	} else if script.LuaFile != "" {
		scriptToRun := script.LuaFile

		// Runs the main script. Repeats the script if the user selects new arguments
		for {
			reRunScript := false
			var loader *modules.Loader

			// Loding modules
			loader, err = modules.LoadAllModules(runtime, &modules.AllModulesLoaderOptions{
				RawMode: a.rawMode,
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
			err = runtime.RunFile(scriptToRun, args...)
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

	} else {
		return fmt.Errorf("there is no action defined for the script '%s'", scriptName)
	}

	return nil
}
