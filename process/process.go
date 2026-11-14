// Package process implements a system process interface
package process

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/LucasAVasco/falcula/colorgen"
	"github.com/LucasAVasco/falcula/multiplexer"

	"github.com/fatih/color"
)

type OnExitCallback func(info *ExitInfo)

// Process represents a process. Extends the `exec.Cmd` interface with support to run shell commands and gracefully stop the process
//
// All methods are safe for concurrent use
type Process struct {
	cmd    *exec.Cmd
	onExit OnExitCallback

	mutex    sync.Mutex // Protects the fields below
	started  bool       // `Start` was called (or the process was stopped before starting)
	ended    bool       // The process ended and the exit information is final
	exitInfo ExitInfo   // Only read without the mutex after `waitGroup` is done

	waitGroup sync.WaitGroup // Done when the process ends (see `Wait`)
}

type Options struct {
	Dir         string         // Working directory. If empty, the current working directory will be used
	Shell       bool           // Run the command in a shell
	Wait        bool           // The `New` command will wait for the process to end
	ManualStart bool           // Does not start the process automatically
	OnExit      OnExitCallback // Callback called when the process ends
	Env         []string       // Environment variables to add (or override) in the process. Use the format `ENVIRONMENT_NAME=value`

	Multiplexer *multiplexer.Multiplexer // Multiplexer used for logging
	Name        string                   // Name of the process used for logging
	Color       *color.Color             // Color used for logging
}

// CreateCmd creates a new `exec.Cmd` with supported to run it in a shell (if `shell` is true)
func CreateCmd(shell bool, command string, args ...string) *exec.Cmd {
	if shell {
		shellBaseCommand := getShellBaseCommand()
		newCommand := shellBaseCommand[0]
		newArgs := []string{shellBaseCommand[1], command}
		newArgs = append(newArgs, args...)

		command = newCommand
		args = newArgs
	}

	return exec.Command(command, args...)
}

func New(opts *Options, command string, args ...string) (*Process, error) {
	p := Process{
		cmd:    CreateCmd(opts.Shell, command, args...),
		onExit: opts.OnExit,
	}

	// Working directory
	if opts.Dir != "" {
		p.cmd.Dir = opts.Dir
	}

	// Environment variables
	p.cmd.Env = append(os.Environ(), opts.Env...)

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

// Starts the process if it has not been started yet. If the process can not be started, the error is also stored in the exit information
// (the process is considered ended)
func (p *Process) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Only once
	if p.started {
		return nil
	}
	p.started = true

	// Starts the command
	err := p.cmd.Start()
	if err != nil {
		err = fmt.Errorf("error starting process: %w", err)
		p.exitInfo.Error = err
		p.endWithoutLock()
		return err
	}

	// Routine to wait for the process to end and get the exit code
	go func() {
		err := p.cmd.Wait()

		p.mutex.Lock()
		if err != nil && !p.exitInfo.Stopped { // If we stopped the process, it will return an error. We should not display it
			p.exitInfo.Error = err
			p.exitInfo.Code = GetExitCodeFromError(err)
		}
		p.endWithoutLock()
		p.mutex.Unlock()
	}()

	return nil
}

// endWithoutLock marks the process as ended, calls the exit callback and releases the `Wait` method. The caller must hold the mutex
func (p *Process) endWithoutLock() {
	p.ended = true

	// The callback receives a copy, so it does not race with the exit information
	if p.onExit != nil {
		exitInfo := p.exitInfo
		go func() {
			p.onExit(&exitInfo)
			p.waitGroup.Done()
		}()
		return
	}

	p.waitGroup.Done() // NOTE(LucasAVasco): the `waitGroup.Add` method is called in the `New` method
}

func (p *Process) Started() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.started
}

// Wait until for the process to end. Returns the exit information of the process
func (p *Process) Wait() (info *ExitInfo) {
	p.waitGroup.Wait()

	return &p.exitInfo
}

// Kill a process. Non-blocking operation. If the process has not been started yet, it will never start (`Wait` returns immediately)
//
// The `force` parameter will force the process to be killed (send SIGKILL instead of executing a graceful shutdown)
func (p *Process) Kill(force bool) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.ended {
		return nil
	}

	p.exitInfo.Stopped = true

	// Not started yet: prevents it from starting
	if !p.started {
		p.started = true
		p.endWithoutLock()
		return nil
	}

	// Kills the process
	var err error
	if force {
		err = p.cmd.Process.Kill()
	} else {
		err = gracefulStop(p.cmd)
	}
	if err != nil && !errors.Is(err, os.ErrProcessDone) {
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

// Stopped checks if the process has been stopped by this interface. Returns false if an external command stopped the process (e.g.: `kill`
// command in Linux). If the user calls the `Stop` method after the process has been ended (without using the `Stop` method), this method
// will return false
func (p *Process) Stopped() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.exitInfo.Stopped
}
