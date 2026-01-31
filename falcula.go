// Package falcula implements the function to start Falcula CLI
package falcula

import (
	"fmt"
	"os"

	"github.com/LucasAVasco/falcula/tui"
)

// StartCli starts the Falcula CLI
func StartCli() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("must be called with 'run' or 'run-raw', but does not receive any argument")
	}

	ui, err := tui.New(os.Args[1] == "run-raw")
	if err != nil {
		return fmt.Errorf("error creating UI: %w", err)
	}
	defer ui.Close()

	err = ui.Open()
	if err != nil {
		return fmt.Errorf("error opening UI: %w", err)
	}

	err = ui.Close()
	if err != nil {
		return fmt.Errorf("error closing UI: %w", err)
	}

	return nil
}
