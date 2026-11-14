package cmd

import (
	"fmt"
	"os"

	"github.com/LucasAVasco/falcula"
	"github.com/LucasAVasco/falcula/internal/terminal"
	"github.com/LucasAVasco/falcula/version"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "falcula",
	Short: "A programmable toolkit for services, containers, and image generation",
	Long: `Falcula is a tool that can be used to create and manage services, containers, and image generation.

It works by running tasks written in shell or Lua in the background and providing a TUI (optional) to control it.
Falcula exposes a set of Lua modules to create and manage services and containers.`,

	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if version.CurrentVersionString != "" {
		rootCmd.Version = version.CurrentVersionString
	} else {
		rootCmd.Version = "unknown"
	}

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("raw", false,
		"Run in raw mode (disables TUI). Forced in CI environments, when the output is redirected or the input is not a terminal")
}

// createFalculaApp create a new falcula application
func createFalculaApp(cmd *cobra.Command) (*falcula.App, error) {
	rawMode, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return nil, fmt.Errorf("error getting value of 'raw' flag: %w", err)
	}

	// The TUI needs a terminal to draw on and to read the user input from
	if terminal.NonInteractiveReason() != terminal.ReasonNone {
		rawMode = true
	}

	app, err := falcula.NewApp(rawMode)
	if err != nil {
		return nil, fmt.Errorf("error creating app: %w", err)
	}

	return app, nil
}
