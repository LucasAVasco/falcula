package falcula

import (
	"os"
	"os/exec"
)

// configureCmd configures a script or task execution command
func (a *App) configureCmd(cmd *exec.Cmd, projectFolder string) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "FALCULA_PROJECT_DIR="+projectFolder)
	cmd.Env = append(cmd.Env, "FALCULA_INVOKE_DIR="+a.invokeDir)
}
