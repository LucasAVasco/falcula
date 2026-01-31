package process

import (
	"os/exec"
	"runtime"
)

// getShellBaseCommand returns the arguments required to run a command in a new shell
func getShellBaseCommand() []string {
	if runtime.GOOS == "windows" {
		return []string{"cmd", "/c"}
	} else {
		return []string{"sh", "-c"}
	}
}

// ShellIsAvailable checks if the shell is available
func ShellIsAvailable() bool {
	shell := getShellBaseCommand()[0]
	_, err := exec.LookPath(shell)
	return err == nil
}
