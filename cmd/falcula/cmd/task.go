package cmd

import (
	"fmt"

	"github.com/LucasAVasco/falcula"
	"github.com/spf13/cobra"
)

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Run a task",
	Long: `Run a task of the current or specified project.

The provided arguments are passed to the task.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		rawMode, err := cmd.Flags().GetBool("raw")
		if err != nil {
			return fmt.Errorf("error getting value of 'raw' flag: %w", err)
		}

		app, err := falcula.NewApp(rawMode)
		if err != nil {
			return fmt.Errorf("error creating app: %w", err)
		}

		err = app.RunTask(args[0], args...)
		if err != nil {
			return fmt.Errorf("error running task: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
}
