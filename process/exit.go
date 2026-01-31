package process

import (
	"fmt"
	"os/exec"
	"syscall"
)

// Gets the exit code from a process error if available.
func GetExitCodeFromError(err error) ExitCode {
	if err, ok := err.(*exec.ExitError); ok {
		return ExitCode(err.ExitCode())
	}

	return 0
}

func GetExitSignalFromError(err error) syscall.Signal {
	if err, ok := err.(*exec.ExitError); ok {
		return err.Sys().(syscall.WaitStatus).Signal()
	}

	return 0
}

// ExitCode is the exit code of a process
type ExitCode int

// ExitInfo is the exit information of a process
type ExitInfo struct {
	Code    ExitCode
	Error   error
	Stopped bool // Manually stopped by the user
}

// HasError checks if the process has exited with an error (checks the exit code and the process error)
func (e *ExitInfo) HasError() bool {
	if e.Stopped {
		return false
	}

	return e.Error != nil || e.Code != 0
}

// WrapError wraps the exit code and the error in an single error interface
func (e *ExitInfo) WrapError() error {
	if !e.HasError() {
		return nil
	}

	return fmt.Errorf("exit code: %d, error: %w", e.Code, e.Error)
}
