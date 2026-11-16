package cmd

import (
	"fmt"

	"github.com/LucasAVasco/falcula"
	"github.com/LucasAVasco/falcula/internal/terminal"
	"github.com/spf13/cobra"
)

// taskRunCmd represents the taskRun command
var taskRunCmd = &cobra.Command{
	Use:   "run [<task>[.<sub-task>]... [arguments...]]",
	Short: "Run a task",
	Long: `Run a task of the current project or specified project.

Without a task, opens an interactive menu to select the task (requires an interactive terminal). Type to filter, Enter runs a task or opens
a sub-project, Backspace goes back to the parent project and Esc quits.

If the task has a sub-task, you must provide it after the task name separated by a dot. For example: "taskName.subTaskName". A subtask can
have its own subtasks. For example: "taskName.subTaskName.subSubTaskName".

The provided arguments are passed to the task.

You can access inner project tasks using the following syntax for the task: "innerProject1:innerProject2:taskName"
`,

	Example: `
falcula run

falcula task run taskName

falcula task run taskName arg1 arg2

falcula task run taskName.subTask

falcula task run taskName.subTask.subSubTask

falcula task run innerProject1:innerProject2:taskName

falcula task run innerProject1:innerProject2:taskName.subTask

falcula task run innerProject1:innerProject2:taskName arg1 arg2`,

	Args: cobra.ArbitraryArgs,

	ValidArgsFunction: completeTaskRunArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := createFalculaApp(cmd)
		if err != nil {
			return fmt.Errorf("error creating falcula app: %w", err)
		}

		// Without a task, the user selects it in a menu
		if len(args) == 0 {
			taskId, err := selectTask(cmd, app)
			if err != nil || taskId == "" {
				return err // Canceled if empty
			}
			args = []string{taskId}
		}

		err = app.RunTaskById(args[0], args[1:]...)
		if err != nil {
			return fmt.Errorf("error running task: %w", err)
		}

		return nil
	},
}

func init() {
	taskCmd.AddCommand(taskRunCmd)
	rootCmd.AddCommand(taskRunCmd) // Alias to run tasks with `falcula run <task-name>`
}

// completeTaskRunArgs completes the task name (first argument) with the tasks and child projects of the current project. Only the projects
// in the typed name are loaded. The other arguments are completed with the argument sets of the task (`args` option), or as files
func completeTaskRunArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	proj, err := falcula.FindProject()
	if err != nil {
		cobra.CompDebugln(fmt.Sprintf("error loading project: %v", err), false)
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveDefault
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Task arguments: the argument sets of the task (`args` option), or files
	if len(args) > 0 {
		candidates, ok, err := proj.CompleteTaskArgs(args[0], args[1:], toComplete)
		if err != nil {
			cobra.CompDebugln(fmt.Sprintf("error completing task arguments: %v", err), false)
		}
		if !ok || len(candidates) == 0 {
			return nil, cobra.ShellCompDirectiveDefault
		}
		return candidates, cobra.ShellCompDirectiveNoFileComp
	}

	candidates, hasProjects, err := proj.CompleteTaskName(toComplete)
	if err != nil {
		cobra.CompDebugln(fmt.Sprintf("error completing task name: %v", err), false)
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	directive := cobra.ShellCompDirectiveNoFileComp
	if hasProjects {
		directive |= cobra.ShellCompDirectiveNoSpace // The user continues typing the task of the project
	}

	return candidates, directive
}

// selectTask shows the menu that selects the task to run. Returns an error if there is no interactive terminal (or in raw mode)
func selectTask(cmd *cobra.Command, app *falcula.App) (string, error) {
	rawMode, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return "", fmt.Errorf("error getting value of 'raw' flag: %w", err)
	}

	reason := terminal.NonInteractiveReason()
	if rawMode {
		reason = "raw mode"
	}
	if reason != terminal.ReasonNone {
		return "", fmt.Errorf("a task name is required (the task menu needs an interactive terminal: %s)", reason)
	}

	taskId, err := app.SelectTask()
	if err != nil {
		return "", fmt.Errorf("error selecting the task: %w", err)
	}

	return taskId, nil
}
