package cmd

import (
	"fmt"

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

	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := createFalculaApp(cmd)
		if err != nil {
			return fmt.Errorf("error creating falcula app: %w", err)
		}

		err = app.RunTask(args[0], args...)
		if err != nil {
			return fmt.Errorf("error running task: %w", err)
		}

		return nil
	},
}

func init() {
	taskCmd.AddCommand(taskRunCmd)
}
