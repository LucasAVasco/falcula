package cmd

import (
	"fmt"

	"github.com/LucasAVasco/falcula/config"
	"github.com/spf13/cobra"
)

// configInitCmd represents the config init command
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Write the default configuration and theme files",
	Long: `Write the default configuration file (config.yaml) and the default theme (themes/soft.yaml) to the user configuration folder, as a
starting point to edit. Existing files are kept unless '--force' is used.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return fmt.Errorf("error getting value of 'force' flag: %w", err)
		}

		results, err := config.Init(force)
		for _, result := range results {
			if result.Written {
				fmt.Println("written:", result.Path)
			} else {
				fmt.Println("kept (already exists, use --force to overwrite):", result.Path)
			}
		}
		if err != nil {
			return fmt.Errorf("error writing configuration files: %w", err)
		}

		return nil
	},
}

func init() {
	configInitCmd.Flags().Bool("force", false, "Overwrite existing files")
	configCmd.AddCommand(configInitCmd)
}
