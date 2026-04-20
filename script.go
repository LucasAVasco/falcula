package falcula

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/LucasAVasco/falcula/lua/luaruntime"
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

	if script.Command.IsNotEmpty() {
		var cmd *exec.Cmd
		if script.Command.List != nil {
			cmd = process.CreateCmd(false, script.Command.List[0], script.Command.List[1:]...)
		} else {
			cmd = process.CreateCmd(true, script.Command.String)
		}
		a.configureCmd(cmd, script.Project.Folder)

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error running command: %w", err)
		}

	} else if script.Lua != "" {
		config := &runLuaConfig{
			Runtime: runtime,
			Code:    script.Lua,
			Args:    args,
		}

		err := a.runLuaCode(config)
		if err != nil {
			return fmt.Errorf("error running Lua code: %w", err)
		}

	} else if script.File != "" {
		cmd := process.CreateCmd(false, script.File)
		a.configureCmd(cmd, script.Project.Folder)

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error running command: %w", err)
		}

	} else if script.LuaFile != "" {
		config := &runLuaConfig{
			Runtime: runtime,
			File:    script.LuaFile,
			Args:    args,
		}

		err := a.runLuaFile(config)
		if err != nil {
			return fmt.Errorf("error running Lua file: %w", err)
		}

	} else {
		return fmt.Errorf("there is no action defined for the script '%s'", scriptName)
	}

	return nil
}
