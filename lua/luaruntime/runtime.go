// Package luaruntime implements the environment for running Lua scripts
package luaruntime

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/LucasAVasco/falcula/lua/luaruntime/ioredirect"
	"github.com/LucasAVasco/falcula/lua/luaruntime/logger"
	"github.com/LucasAVasco/falcula/project"
	"github.com/LucasAVasco/falcula/service/manager"
	lua "github.com/yuin/gopher-lua"
)

type OnRunTask = func(task *project.Task, subTask string, args []string) error

// Runtime represents the environment for running Lua scripts. It has a Lua state and data common to all modules, including the created
// service managers the current state of the script (current arguments and available arguments)
type Runtime struct {
	project   *project.Project
	onRunTask OnRunTask

	// Current Lua state. It must only be closed or reset by the goroutine that runs scripts on it (other goroutines use `Interrupt`)
	luaState         *lua.LState
	stateMutex       sync.Mutex
	lastExecutedFile string

	// Cancels the script that is running (nil if no script is running)
	cancelRun      context.CancelFunc
	cancelRunMutex sync.Mutex
	interrupted    atomic.Bool // `Interrupt` was called after the last script started

	// All logs are sent to this logger
	Logger *logger.Logger

	// Service managers
	managersMutex sync.Mutex
	managers      []*manager.Manager

	// Script arguments
	scriptCurrentArgs        []string
	scriptAvailableArgs      [][]string
	onSetScriptCurrentArgs   func(args []string)
	onSetScriptAvailableArgs func(args [][]string)
}

func New(proj *project.Project, onRunTask OnRunTask) (*Runtime, error) {
	r := Runtime{
		project:   proj,
		onRunTask: onRunTask,
		managers:  make([]*manager.Manager, 0),
	}

	// Default callbacks
	r.SetOnCurrentScriptArgsChange(func(args []string) {})
	r.SetOnScriptAvailableArgsChange(func(args [][]string) {})

	// Logger
	var err error
	r.Logger, err = logger.New()
	if err != nil {
		return nil, fmt.Errorf("error creating logger: %w", err)
	}

	return &r, nil
}

// Close closes the runtime (the Lua state, its service managers and the logger). Can be called multiple times
func (r *Runtime) Close() error {
	r.CloseLuaState()

	err := r.Logger.Close()
	if err != nil {
		return fmt.Errorf("error closing logger: %w", err)
	}

	return nil
}

func (r *Runtime) closeLuaStateWithoutLock() {
	if r.luaState == nil {
		return
	}

	r.closeAllManagers()
	r.luaState.Close()
	r.luaState = nil
}

func (r *Runtime) resetLuaStateWithoutLock() {
	r.closeLuaStateWithoutLock()

	r.luaState = lua.NewState()
	ioredirect.Redirect(r.luaState, r.Logger)
}

// GetProject returns the current project if it exists. If it doesn't exist, returns `nil`.
func (r *Runtime) GetProject() *project.Project {
	return r.project
}

// GetOnRunTask returns the function that is called when a task is selected to run.
func (r *Runtime) GetOnRunTask() OnRunTask {
	return r.onRunTask
}

// GetLuaState returns the current Lua state. Creates a new one if it doesn't exist
func (r *Runtime) GetLuaState() *lua.LState {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	if r.luaState == nil {
		r.resetLuaStateWithoutLock()
	}
	return r.luaState
}

// Interrupt stops the script that is running (if any) and closes all service managers. Unlike `CloseLuaState`, it does not close the Lua
// state, so it is safe to call from any goroutine while a script is running: blocking calls that wait for the services return and the
// script is aborted (`Run` and `RunFile` return an error). The goroutine that runs the script must reset the Lua state before reusing it
func (r *Runtime) Interrupt() {
	r.cancelRunMutex.Lock()
	r.interrupted.Store(true)
	if r.cancelRun != nil {
		r.cancelRun()
	}
	r.cancelRunMutex.Unlock()

	r.closeAllManagers()
}

// Interrupted returns true if `Interrupt` was called after the last script started
func (r *Runtime) Interrupted() bool {
	return r.interrupted.Load()
}

// CloseLuaState closes the current Lua state and all service managers. Must not be called while a script is running (see `Interrupt`)
func (r *Runtime) CloseLuaState() {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	r.closeLuaStateWithoutLock()
}

// ResetLuaState closes the current Lua state and creates a new one
func (r *Runtime) ResetLuaState() *lua.LState {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	r.resetLuaStateWithoutLock()
	return r.luaState
}

// GetLastExecutedFile returns the path of the last file executed with `RunFile()`. The `RunFile()` method updates the value returned by
// this method. If a file was executed without calling `RunFile()`, it will not update this value
func (r *Runtime) GetLastExecutedFile() string {
	return r.lastExecutedFile
}

// Run runs a Lua code with the given arguments. It updates the current arguments, but does not update the last executed file because there
// is no file
func (r *Runtime) Run(luaCode string, args ...string) error {
	r.lastExecutedFile = ""

	state, done := r.prepareRun(args)
	defer done()

	err := state.DoString(luaCode)
	if err != nil {
		return fmt.Errorf("error executing Lua code: %w", err)
	}
	return nil
}

// RunFile executes a Lua script with the given arguments. It updates the current arguments and the last executed file returned by
// `GetLastExecutedFile()`
func (r *Runtime) RunFile(file string, args ...string) error {
	r.lastExecutedFile = file

	state, done := r.prepareRun(args)
	defer done()

	err := state.DoFile(file)
	if err != nil {
		return fmt.Errorf("error executing file '%s': %w", file, err)
	}
	return nil
}

// prepareRun sets the command line arguments of the script (the global `arg` table) and returns the Lua state used to run it. The script
// can be cancelled with `Interrupt` until the returned `done` function is called
func (r *Runtime) prepareRun(args []string) (state *lua.LState, done func()) {
	state = r.GetLuaState()

	// The Lua virtual machine aborts the script when the context is cancelled
	ctx, cancel := context.WithCancel(context.Background())
	state.SetContext(ctx)

	r.cancelRunMutex.Lock()
	r.cancelRun = cancel
	r.interrupted.Store(false)
	r.cancelRunMutex.Unlock()

	done = func() {
		r.cancelRunMutex.Lock()
		r.cancelRun = nil
		r.cancelRunMutex.Unlock()

		state.RemoveContext()
		cancel()
	}

	r.scriptCurrentArgs = args
	argTable := state.NewTable()
	for i, arg := range args {
		argTable.RawSetInt(i+1, lua.LString(arg))
	}
	state.SetGlobal("arg", argTable)
	r.onSetScriptCurrentArgs(args)

	return state, done
}
