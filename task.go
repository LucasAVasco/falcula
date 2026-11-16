package falcula

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/LucasAVasco/falcula/lua/luaruntime"
	"github.com/LucasAVasco/falcula/lua/modules/modtui"
	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/project"
	"github.com/LucasAVasco/falcula/taskmenu"
	lua "github.com/yuin/gopher-lua"
)

// GetAllTasks returns all the tasks in the current project and its children projects.
func (a *App) GetAllTasks() (map[string]*project.Task, error) {
	return a.project.GetAllTasks()
}

// SelectTask shows the interactive menu that selects the task to run. Returns the task identifier, or an empty string if the user canceled.
// Requires an interactive terminal
func (a *App) SelectTask() (string, error) {
	return taskmenu.Run(a.project, &taskmenu.Options{Theme: a.theme})
}

// RunTaskById runs a task of the current project with the given arguments. Gets the task by its ID from the current project. The ID follows
// this format: "taskName.subTaskName"
func (a *App) RunTaskById(taskId string, args ...string) error {
	taskName, subTask, _ := strings.Cut(taskId, ".")

	task, err := a.project.GetTaskByName(taskName)
	if err != nil {
		return fmt.Errorf("error getting task to run: %w", err)
	}

	return a.RunTask(task, subTask, args...)
}

// RunTask runs a task of the current project with the given arguments. The subtask is optional.
func (a *App) RunTask(task *project.Task, subTask string, args ...string) error {
	// Lua runtime
	runtime, err := luaruntime.New(task.Project, func(task *project.Task, subTask string, args []string) error {
		return a.RunTask(task, subTask, args...)
	})
	if err != nil {
		return fmt.Errorf("error creating runtime: %w", err)
	}
	defer func() {
		err := runtime.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error closing runtime: %v\n", err)
		}
	}()

	// modtui does not closes the TUI when it is closed. The TUI is persistent across runs. Need to close it manually
	defer modtui.ClosePersistentTui()

	// Changes to task directory
	err = os.Chdir(task.Cwd)
	if err != nil {
		return fmt.Errorf("error changing to task working directory: %w", err)
	}

	switch {
	case task.Shell.IsNotEmpty():
		return a.runTaskShell(task, subTask, args)

	case task.Lua != "":
		err = a.runLua(&runLuaConfig{
			Runtime:  runtime,
			Code:     task.Lua,
			Args:     args,
			AfterRun: runSubTaskAfterRun(subTask, args),
		})
		if err != nil {
			return fmt.Errorf("error running Lua code: %w", err)
		}

	case task.ShellFile != "":
		err = a.runTaskShellScript(task, subTask, "source "+task.ShellFile, args)
		if err != nil {
			return fmt.Errorf("error running task shell script: %w", err)
		}

	case task.LuaFile != "":
		err = a.runLua(&runLuaConfig{
			Runtime:  runtime,
			File:     task.LuaFile,
			Args:     args,
			AfterRun: runSubTaskAfterRun(subTask, args),
		})
		if err != nil {
			return fmt.Errorf("error running Lua file: %w", err)
		}

	default:
		return fmt.Errorf("there is no action defined for the task '%s'", task.Name)
	}

	return nil
}

// runTaskShell runs the `shell` field of a task. It can be a command (list) or a shell script (string)
func (a *App) runTaskShell(task *project.Task, subTask string, args []string) error {
	if task.Shell.List != nil {
		cmdArgs := slices.Concat(task.Shell.List[1:], args) // Copies the list, so the task is not modified
		cmd := process.CreateCmd(false, task.Shell.List[0], cmdArgs...)
		a.configureTaskCmd(cmd, task, subTask)

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error running shell command: %w", err)
		}
		return nil
	}

	if task.Shell.String != "" {
		err := a.runTaskShellScript(task, subTask, task.Shell.String, args)
		if err != nil {
			return fmt.Errorf("error running shell script: %w", err)
		}
		return nil
	}

	return fmt.Errorf("shell is empty")
}

// runSubTaskAfterRun returns a callback that runs the subtask returned by the Lua code or file. Does nothing if there is no subtask
func runSubTaskAfterRun(subTask string, args []string) func(runtime *luaruntime.Runtime) error {
	return func(runtime *luaruntime.Runtime) error {
		if subTask == "" {
			return nil
		}

		err := handleReturnedLuaTask(runtime.GetLuaState(), subTask, args)
		if err != nil {
			return fmt.Errorf("error handling returned task: %w", err)
		}
		return nil
	}
}

// configureTaskCmd configures a task execution command
func (a *App) configureTaskCmd(cmd *exec.Cmd, task *project.Task, subTask string) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "FALCULA_PROJECT_DIR="+task.Project.Folder)
	cmd.Env = append(cmd.Env, "FALCULA_INVOKE_DIR="+a.invokeDir)
	cmd.Env = append(cmd.Env, "FALCULA_TASK_NAME="+task.Name)
	cmd.Env = append(cmd.Env, "FALCULA_SUB_TASK_NAME="+subTask)
}

// runTaskShellScript runs a shell script code with the given arguments and executes a subtask of it
func (a *App) runTaskShellScript(task *project.Task, subTask, code string, args []string) error {
	// Executes the subtask after the provided code
	if subTask != "" {
		code += "\n" + subTask + " \"$@\""
	}

	// Arguments passed to the shell (the first one is the $0)
	cmdArgs := []string{"sh"}
	cmdArgs = append(cmdArgs, args...)

	// Runs the command
	cmd := process.CreateCmd(true, code, cmdArgs...)
	a.configureTaskCmd(cmd, task, subTask)

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
