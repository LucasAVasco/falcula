// Package process implements a system process interface
package process

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/LucasAVasco/falcula/colorgen"
	"github.com/LucasAVasco/falcula/multiplexer"

	"github.com/fatih/color"
)

type OnExitCallback func(info *ExitInfo)

// Process represents a process. Extends the `exec.Cmd` interface with support to run shell commands and gracefully stop the process
type Process struct {
	cmd *exec.Cmd

	started bool
	onExit  OnExitCallback

	waitGroup sync.WaitGroup // `Wait` method
	exitInfo  ExitInfo       // `Wait` method
}

type Options struct {
	Shell       bool           // Run the command in a shell
	Wait        bool           // The `New` command will wait for the process to end
	ManualStart bool           // Does not start the process automatically
	OnExit      OnExitCallback // Callback called when the process ends

	Multiplexer *multiplexer.Multiplexer // Multiplexer used for logging
	Name        string                   // Name of the process used for logging
	Color       *color.Color             // Color used for logging
}

// CreateCmd creates a new `exec.Cmd` with supported to run it in a shell (if `shell` is true)
func CreateCmd(shell bool, command string, args ...string) *exec.Cmd {
	if shell {
		if len(args) > 0 {
			command = command + " " + strings.Join(args, " ")
		}
		shellBaseCommand := getShellBaseCommand()
		cmdName := shellBaseCommand[0]
		cmdArgs := append(shellBaseCommand[1:], command)
		return exec.Command(cmdName, cmdArgs...)
	} else {
		return exec.Command(command, args...)
	}
}

func New(opts *Options, command string, args ...string) (*Process, error) {
	p := Process{
		cmd:    CreateCmd(opts.Shell, command, args...),
		onExit: opts.OnExit,
	}

	// Command
	if opts.Shell {
		if len(args) > 0 {
			command = command + " " + strings.Join(args, " ")
		}
		shellBaseCommand := getShellBaseCommand()
		cmdName := shellBaseCommand[0]
		cmdArgs := append(shellBaseCommand[1:], command)
		p.cmd = exec.Command(cmdName, cmdArgs...)
	} else {
		p.cmd = exec.Command(command, args...)
	}

	// Ensures a color exists
	color := opts.Color
	if color == nil {
		color = colorgen.Default
	}

	// Standard output client
	p.cmd.Stdout = opts.Multiplexer.NewClient(opts.Name, "stdout", color)

	// Standard error client
	p.cmd.Stderr = opts.Multiplexer.NewClient(opts.Name, "stderr", color)

	// Starts the process
	p.waitGroup.Add(1) // NOTE(LucasAVasco): Will be done when the process ends (see the routine in `Start`)
	if !opts.ManualStart {
		err := p.Start()
		if err != nil {
			return nil, fmt.Errorf("error starting process: %w", err)
		}
	}

	// Waits until the process ends
	if opts.Wait {
		exitInfo := p.Wait()
		if exitInfo.Error != nil {
			return nil, fmt.Errorf("error waiting for process: %w", exitInfo.Error)
		}
	}

	return &p, nil
}

// Starts the process if it has not been started yet
func (p *Process) Start() error {
	// Only once
	if p.started {
		return nil
	}
	p.started = true

	// Starts the command
	err := p.cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting process: %w", err)
	}

	// Routine to wait for the process to end and get the exit code
	// NOTE(LucasAVasco): the `waitGroup.Add` method is called in the `New` method, can not call it again
	go func() {
		defer p.waitGroup.Done()

		err := p.cmd.Wait()
		if err != nil && !p.Stopped() { // If we stopped the process, it will return an error. We should not display it
			p.exitInfo.Error = err
			p.exitInfo.Code = GetExitCodeFromError(err)
		}

		if p.onExit != nil {
			p.onExit(&p.exitInfo)
		}
	}()

	return nil
}

func (p *Process) Started() bool {
	return p.started
}

// Wait until for the process to end.
//
// Return the exit code of the process and an error.
func (p *Process) Wait() (info *ExitInfo) {
	p.waitGroup.Wait()

	return &p.exitInfo
}

// GetExitCode returns the exit code of the process. Blocks until the process ends
func (p *Process) GetExitCode() ExitCode {
	exitInfo := p.Wait()
	return exitInfo.Code
}

// GetExitError returns the exit error of the process. Blocks until the process ends
func (p *Process) GetExitError() error {
	exitInfo := p.Wait()
	return exitInfo.Error
}

// Kill a process. Non-blocking operation
//
// The `force` parameter will force the process to be killed (send SIGKILL instead of executing a graceful shutdown)
func (p *Process) Kill(force bool) error {
	if p.cmd.ProcessState != nil {
		if p.cmd.ProcessState.Exited() {
			return nil
		}
	}

	p.exitInfo.Stopped = true

	// Kills the process
	var err error
	if force {
		err = p.cmd.Process.Kill()
	} else {
		err = gracefulStop(p.cmd)
	}
	if err != nil {
		return fmt.Errorf("error killing command: %w", err)
	}

	return nil
}

// Stop the process. Blocking operation
//
// The `force` parameter will force the process to be killed (send SIGKILL instead of executing a graceful shutdown)
//
// Return the exit code of the process and an error.
func (p *Process) Stop(force bool) (*ExitInfo, error) {
	// Kills the process
	err := p.Kill(force)
	if err != nil {
		return nil, fmt.Errorf("error killing process: %w", err)
	}

	// Waits for the process to end
	return p.Wait(), nil
}

// Exited checks if the process has exited. If the process has not been started yet, this method will return false
func (p *Process) Exited() bool {
	if p.cmd.ProcessState == nil {
		return false
	}

	return p.cmd.ProcessState.Exited()
}

// Stopped checks if the process has been stopped by this interface. Returns false if an external command stopped the process (e.g.: `kill`
// command in Linux). If the user calls the `Stop` method after the process has been ended (without using the `Stop` method), this method
// will return false
func (p *Process) Stopped() bool {
	return p.exitInfo.Stopped
}
