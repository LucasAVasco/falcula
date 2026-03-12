package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "falcula",
	Short: "A programmable toolkit for services, containers, and image generation",
	Long: `Falcula is a tool that can be used to create and manage services, containers, and image generation.

It works by running a Lua script in the background and providing a TUI (optional) to control it.
Falcula exposes a set of Lua modules to create and manage services and containers.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("raw", false, "Run in raw mode (disables TUI)")
}
