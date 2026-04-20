package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// scriptListCmd represents the scriptList command
var scriptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available scripts",
	Long:  `List available scripts of the current or specified project and its child projects.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := createFalculaApp(cmd)
		if err != nil {
			return fmt.Errorf("error creating falcula app: %w", err)
		}

		scripts, err := app.GetScriptList()
		if err != nil {
			return fmt.Errorf("error getting scripts list: %w", err)
		}
		for name := range scripts {
			fmt.Println(name)
		}

		return nil
	},
}

func init() {
	scriptCmd.AddCommand(scriptListCmd)
}
