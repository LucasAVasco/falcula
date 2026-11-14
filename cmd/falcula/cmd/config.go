package cmd

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "User configuration commands",
	Long: `User configuration commands.

The user configuration folder is '$FALCULA_CONFIG_DIR' or the 'falcula' folder in the user configuration folder (e.g. '~/.config/falcula'
on Linux). It has a 'config.yaml' file and a 'themes' folder with the themes of the text user interface (YAML files).`,
}

func init() {
	rootCmd.AddCommand(configCmd)
}
