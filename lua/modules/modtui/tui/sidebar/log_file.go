package sidebar

import (
	"fmt"
	"os"
	"os/exec"
)

// runInteractive runs a command synchronously connected to the terminal (standard input, output and error)
func runInteractive(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// openLogFileInLess opens the log file in `less`
func (s *Sidebar) openLogFileInLess() error {
	err := runInteractive("less", s.logFilePath)
	if err != nil {
		return fmt.Errorf("error running 'less': %w", err)
	}

	return nil
}

// openLogFileInLnav opens the log file in `Lnav`. The filter is a Lnav expression used in the `:filter-expr` command. It is optional.
func (s *Sidebar) openLogFileInLnav(filter string) error {
	commandWhenOpen := ":reset-session" // Opens all services (ensure the filter is cleared)
	if filter != "" {
		commandWhenOpen = ":filter-expr " + filter // Opens specific services
	}

	err := runInteractive("lnav", s.logFilePath, "-c", commandWhenOpen)
	if err != nil {
		return fmt.Errorf("error running 'Lnav': %w", err)
	}

	return nil
}
