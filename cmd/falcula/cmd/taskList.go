package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// taskListCmd represents the taskList command
var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available tasks",
	Long:  `List available tasks of the current or specified project and its child projects.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := createFalculaApp(cmd)
		if err != nil {
			return fmt.Errorf("error creating falcula app: %w", err)
		}

		tasks, err := app.GetTaskList()
		if err != nil {
			return fmt.Errorf("error getting tasks list: %w", err)
		}
		for name := range tasks {
			fmt.Println(name)
		}

		return nil
	},
}

func init() {
	taskCmd.AddCommand(taskListCmd)
}
