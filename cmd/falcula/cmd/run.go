package cmd

import (
	"fmt"

	"github.com/LucasAVasco/falcula"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a falcula script",
	Long: `Runs a Lua script with all falcula modules available.

The provided arguments are passed to the script.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		rawMode, err := cmd.Flags().GetBool("raw")
		if err != nil {
			return fmt.Errorf("error getting value of 'raw' flag: %w", err)
		}

		app, err := falcula.NewApp(rawMode)
		if err != nil {
			return fmt.Errorf("error creating falcula: %w", err)
		}

		err = app.RunScript(args[0], args...)
		if err != nil {
			return fmt.Errorf("error running script: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
