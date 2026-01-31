package sidebar

import (
	"fmt"
	"os"
	"os/exec"
)

// openLogFileInLess opens the log file in `less`
func (s *Sidebar) openLogFileInLess() error {
	cmd := exec.Command("less", s.logFilePath)

	// Redirection
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run synchronously
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running 'less': %w", err)
	}

	return nil
}

// openLogFileInLnav opens the log file in `Lnav`. The filter is a Lnav expression used in the `:filter-expr` command. It is optional.
func (s *Sidebar) openLogFileInLnav(filter string) error {
	args := []string{s.logFilePath}

	if filter != "" {
		// Opens a specific service
		commandWhenOpen := ":filter-expr " + filter
		args = append(args, "-c", commandWhenOpen)
	} else {
		// Opens all services (ensure the filter is cleared)
		args = append(args, "-c", ":reset-session")
	}

	cmd := exec.Command("lnav", args...)

	// Redirection
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run synchronously
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running 'Lnav': %w", err)
	}

	return nil
}
