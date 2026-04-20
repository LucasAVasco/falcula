package cmd

import (
	"github.com/spf13/cobra"
)

// scriptCmd represents the script command
var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "Script commands",
	Long:  `Script related commands. Including running and listing scripts.`,
}

func init() {
	rootCmd.AddCommand(scriptCmd)
}
