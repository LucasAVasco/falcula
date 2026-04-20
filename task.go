package falcula

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/LucasAVasco/falcula/lua/luaruntime"
	"github.com/LucasAVasco/falcula/lua/modules/modtui"
	"github.com/LucasAVasco/falcula/process"
	lua "github.com/yuin/gopher-lua"
)

// RunTask runs a task of the current project with the given arguments. The taskName can be the name of a named task or the path to
// a task
func (a *App) RunTask(taskId string, args ...string) error {
	taskName, subTask, _ := strings.Cut(taskId, ".")

	// Lua runtime
	runtime, err := luaruntime.New()
	if err != nil {
		return fmt.Errorf("error creating runtime: %w", err)
	}
	defer runtime.Close()

	// modtui does not closes the TUI when it is closed. The TUI is persistent across runs. Need to close it manually
	defer modtui.ClosePersistentTui()

	// Task to run
	task, err := a.project.GetTaskByName(taskName)
	if err != nil {
		return fmt.Errorf("error getting task to run: %w", err)
	}

	// Changes to task directory
	err = os.Chdir(task.Cwd)
	if err != nil {
		return fmt.Errorf("error changing to task working directory: %w", err)
	}

	if task.Command.IsNotEmpty() {
		if task.Command.List != nil {
			cmdArgs := task.Command.List[1:]
			cmdArgs = append(cmdArgs, args...)
			cmd := process.CreateCmd(false, task.Command.List[0], cmdArgs...)
			a.configureTaskCmd(cmd, task, taskId)

			err := cmd.Run()
			if err != nil {
				return fmt.Errorf("error running task command: %w", err)
			}

		} else if task.Command.String != "" {
			code := task.Command.String
			err := a.runTaskShellScript(task, taskId, code, args)
			if err != nil {
				return fmt.Errorf("error running task shell script: %w", err)
			}

		} else {
			return fmt.Errorf("command is empty")
		}

	} else if task.Lua != "" {
		config := &runLuaConfig{
			Runtime: runtime,
			Code:    task.Lua,
			Args:    args,
			AfterRun: func(runtime *luaruntime.Runtime) error {
				err = handleReturnedLuaTask(runtime.GetLuaState(), subTask, args)
				if err != nil {
					return fmt.Errorf("error handling returned task: %w", err)
				}
				return nil
			},
		}

		err = a.runLuaCode(config)
		if err != nil {
			return fmt.Errorf("error running Lua code: %w", err)
		}

	} else if task.File != "" {
		code := "source " + task.File

		err := a.runTaskShellScript(task, taskId, code, args)
		if err != nil {
			return fmt.Errorf("error running task shell script: %w", err)
		}

	} else if task.LuaFile != "" {
		config := &runLuaConfig{
			Runtime: runtime,
			Code:    task.LuaFile,
			Args:    args,
			AfterRun: func(runtime *luaruntime.Runtime) error {
				err = handleReturnedLuaTask(runtime.GetLuaState(), subTask, args)
				if err != nil {
					return fmt.Errorf("error handling returned task: %w", err)
				}
				return nil
			},
		}

		err = a.runLuaFile(config)
		if err != nil {
			return fmt.Errorf("error running Lua file: %w", err)
		}

	} else {
		return fmt.Errorf("there is no action defined for the task '%s'", taskName)
	}

	return nil
}

// configureTaskCmd configures a task execution command
func (a *App) configureTaskCmd(cmd *exec.Cmd, task *project.Task, taskId string) {
	a.configureCmd(cmd, task.Project.Folder)

	taskName, subTask, _ := strings.Cut(taskId, ".")
	cmd.Env = append(cmd.Env, "FALCULA_TASK_ID="+taskId)
	cmd.Env = append(cmd.Env, "FALCULA_TASK_NAME="+taskName)
	cmd.Env = append(cmd.Env, "FALCULA_SUB_TASK_NAME="+subTask)
}

// runTaskShellScript runs a shell script code with the given arguments and executes a subtask of it
func (a *App) runTaskShellScript(task *project.Task, taskId, code string, args []string) error {
	_, subTask, _ := strings.Cut(taskId, ".")

	// Executes the subtask after the provided code
	if subTask != "" {
		code += "\n" + subTask + " \"$@\""
	}

	// Arguments passed to the shell (the first one is the $0)
	cmdArgs := []string{"sh"}
	cmdArgs = append(cmdArgs, args...)

	// Runs the command
	cmd := process.CreateCmd(true, code, cmdArgs...)
	a.configureTaskCmd(cmd, task, taskId)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running command: %w", err)
	}

	return nil
}

// handleReturnedLuaTask handles the task returned after running a Lua code or file. It gets the task/sub-task function from the returned
// value and executes it
func handleReturnedLuaTask(L *lua.LState, subTask string, args []string) error {
	// Gets the returned table
	value := L.Get(-1)
	function, err := getSubTaskFunction(value, subTask)
	if err != nil {
		return fmt.Errorf("error getting subtask function for subtask '%s': %w", subTask, err)
	}

	// Formats the arguments to Lua values
	functionArgs := make([]lua.LValue, 0, len(args))
	for _, arg := range args {
		functionArgs = append(functionArgs, lua.LString(arg))
	}

	// Calls the task function
	err = L.CallByParam(lua.P{
		Fn:      function,
		NRet:    0,
		Protect: true,
	}, functionArgs...)
	if err != nil {
		return fmt.Errorf("error calling task function: %w", err)
	}

	return nil
}

// getSubTaskFunction gets the task/sub-task function from the returned value. It does not execute it
func getSubTaskFunction(value lua.LValue, subTask string) (lua.LValue, error) {
	if subTask == "" {
		if value.Type() != lua.LTFunction && value.Type() != lua.LTTable {
			return nil, fmt.Errorf(
				"the returned value must be a function or a table (with a metatable with '__call'), but got %s",
				value.Type().String(),
			)
		}
		return value, nil
	}

	key, subTask, _ := strings.Cut(subTask, ".")
	if value.Type() != lua.LTTable {
		return nil, fmt.Errorf("the returned value must be a table, but got %s", value.Type().String())
	}

	table := value.(*lua.LTable)
	value = table.RawGet(lua.LString(key))
	function, err := getSubTaskFunction(value, subTask)
	if err != nil {
		return nil, fmt.Errorf("error getting subtask function for subtask '%s': %w", subTask, err)
	}

	return function, nil
}
