package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LucasAVasco/falcula/lua/luaruntime"
	"github.com/LucasAVasco/falcula/lua/modules"
	"github.com/LucasAVasco/falcula/lua/modules/modtui"
	"github.com/spf13/cobra"
)

// getScriptPath returns the path to the script to run
func getScriptPath() (string, error) {
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

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a falcula script",
	Long: `Runs a Lua script with all falcula modules available.

The provided arguments are passed to the script.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		rawMode, err := cmd.Flags().GetBool("raw")
		if err != nil {
			return fmt.Errorf("error getting value of 'raw' flag: %w", err)
		}

		runtime, err := luaruntime.New()
		if err != nil {
			return fmt.Errorf("error creating runtime: %w", err)
		}
		defer runtime.Close()

		// modtui does not closes the TUI when it is closed. The TUI is persistent across runs. Need to close it manually
		defer modtui.ClosePersistentTui()

		// Arguments required to run the main script
		scriptToRun, err := getScriptPath()
		if err != nil {
			return fmt.Errorf("error getting path of the script to run: %w", err)
		}

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
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
