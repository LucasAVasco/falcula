package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// scriptRunCmd represents the scriptRun command
var scriptRunCmd = &cobra.Command{
	Use:   "run <script> [arguments...]",
	Short: "Run a script",
	Long: `Run a script of the current or specified project.

The provided arguments are passed to the script.

You can access inner project scripts using the following syntax for the script: "innerProject1:innerProject2:script"
`,

	Example: `
falcula script run scriptName

falcula script run scriptName arg1 arg2

falcula script run innerProject1:innerProject2:scriptName

falcula script run innerProject1:innerProject2:scriptName arg1 arg2`,

	Args: cobra.MinimumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := createFalculaApp(cmd)
		if err != nil {
			return fmt.Errorf("error creating falcula app: %w", err)
		}

		err = app.RunScript(args[0], args...)
		if err != nil {
			return fmt.Errorf("error running script: %w", err)
		}

		return nil
	},
}

func init() {
	scriptCmd.AddCommand(scriptRunCmd)
	rootCmd.AddCommand(scriptRunCmd) // Alias to run scripts with `falcula run <script>`
}
