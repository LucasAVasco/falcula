package cmd

import (
	"fmt"

	"github.com/LucasAVasco/falcula"
	"github.com/spf13/cobra"
)

// taskRunCmd represents the taskRun command
var taskRunCmd = &cobra.Command{
	Use:   "run <task>[.<sub-task>]... [arguments...]",
	Short: "Run a task",
	Long: `Run a task of the current project or specified project.

If the task has a sub-task, you must provide it after the task name separated by a dot. For example: "taskName.subTaskName". A subtask can
have its own subtasks. For example: "taskName.subTaskName.subSubTaskName".

The provided arguments are passed to the task.

You can access inner project tasks using the following syntax for the task: "innerProject1:innerProject2:taskName"
`,

	Example: `
falcula task run taskName

falcula task run taskName arg1 arg2

falcula task run taskName.subTask

falcula task run taskName.subTask.subSubTask

falcula task run innerProject1:innerProject2:taskName

falcula task run innerProject1:innerProject2:taskName.subTask

falcula task run innerProject1:innerProject2:taskName arg1 arg2`,

	Args: cobra.MinimumNArgs(1),

	ValidArgsFunction: completeTaskRunArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := createFalculaApp(cmd)
		if err != nil {
			return fmt.Errorf("error creating falcula app: %w", err)
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
// in the typed name are loaded. The other arguments are completed as files (they are passed to the task)
func completeTaskRunArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveDefault
	}

	proj, err := falcula.FindProject()
	if err != nil {
		cobra.CompDebugln(fmt.Sprintf("error loading project: %v", err), false)
		return nil, cobra.ShellCompDirectiveNoFileComp
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
