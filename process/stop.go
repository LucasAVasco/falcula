package process

import (
	"fmt"
	"os/exec"
	"runtime"
	"syscall"
)

// gracefulStop gracefully stops the process. Sends a SIGTERM signal in case of Linux and uses 'taskkill' in case of Windows
func gracefulStop(cmd *exec.Cmd) error {
	if runtime.GOOS == "windows" {
		pid := cmd.Process.Pid
		return exec.Command("taskkill", "/PID", fmt.Sprint(pid)).Run()
	} else {
		return cmd.Process.Signal(syscall.SIGTERM)
	}
}
