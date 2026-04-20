package cmd

import (
	"github.com/spf13/cobra"
)

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Task commands",
	Long:  `Task related commands. Including running and listing tasks.`,
}

func init() {
	rootCmd.AddCommand(taskCmd)
}
