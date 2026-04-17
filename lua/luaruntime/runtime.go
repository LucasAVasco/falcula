// Package luaruntime implements the environment for running Lua scripts
package luaruntime

import (
	"fmt"
	"sync"

	"github.com/LucasAVasco/falcula/lua/luaruntime/ioredirect"
	"github.com/LucasAVasco/falcula/lua/luaruntime/logger"
	"github.com/LucasAVasco/falcula/service/manager"
	lua "github.com/yuin/gopher-lua"
)

// Runtime represents the environment for running Lua scripts. It has a Lua state and data common to all modules, includeing, the created
// service managers the current state of the script (current arguments and available arguments)
type Runtime struct {
	// Current Lua state
	luaState         *lua.LState
	stateMutex       sync.Mutex
	lastExecutedFile string

	// All logs are sent to this logger
	Logger *logger.Logger

	// Service managers
	managers []*manager.Manager

	// Script arguments
	scriptCurrentArgs        []string
	scriptAvailableArgs      [][]string
	onSetScriptCurrentArgs   func(args []string)
	onSetScriptAvailableArgs func(args [][]string)
}

func New() (*Runtime, error) {
	r := Runtime{
		managers: make([]*manager.Manager, 0),
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

// Close closes the runtime. Can be called multiple times
func (r *Runtime) Close() error {
	r.CloseLuaState()
	return nil
}

func (r *Runtime) closeLuaStateWithoutLock() {
	if r.luaState == nil {
		return
	}

	r.closeAllManagersWithoutLock()
	r.luaState.Close()
	r.luaState = nil
}

func (r *Runtime) resetLusStateWithoutLock() {
	r.closeLuaStateWithoutLock()

	r.luaState = lua.NewState()
	ioredirect.Redirect(r.luaState, r.Logger)
}

// GetLuaState returns the current Lua state. Creates a new one if it doesn't exist
func (r *Runtime) GetLuaState() *lua.LState {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	if r.luaState == nil {
		r.resetLusStateWithoutLock()
	}
	return r.luaState
}

// CloseLuaState closes the current Lua state and all service managers
func (r *Runtime) CloseLuaState() {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	r.closeLuaStateWithoutLock()
}

// ResetLuaState closes the current Lua state and creates a new one
func (r *Runtime) ResetLuaState() *lua.LState {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	r.resetLusStateWithoutLock()
	return r.luaState
}

// GetLastExecutedFile returns the path of the last file executed with `RunScript()`. The `RunScript()` method updates the value returned by
// this method. If a file was executed without calling `RunScript()`, it will not update this value
func (r *Runtime) GetLastExecutedFile() string {
	return r.lastExecutedFile
}

// Run runs a Lua code with the given arguments. It updates the current arguments, but does not update the last executed file because there
// is no file
func (r *Runtime) Run(luaCode string, args ...string) error {
	r.lastExecutedFile = ""

	// Lua state used to execute the file
	state := r.GetLuaState()
	if state == nil {
		return fmt.Errorf("lua state is nil")
	}

	// Sets the command line arguments
	r.scriptCurrentArgs = args
	argTable := state.NewTable()
	for i, arg := range args {
		argTable.RawSetInt(i+1, lua.LString(arg))
	}
	state.SetGlobal("arg", argTable)
	r.onSetScriptCurrentArgs(args)

	// Executes the file
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

	// Lua state used to execute the file
	state := r.GetLuaState()
	if state == nil {
		return fmt.Errorf("lua state is nil")
	}

	// Sets the command line arguments
	r.scriptCurrentArgs = args
	argTable := state.NewTable()
	for i, arg := range args {
		argTable.RawSetInt(i+1, lua.LString(arg))
	}
	state.SetGlobal("arg", argTable)
	r.onSetScriptCurrentArgs(args)

	// Executes the file
	err := state.DoFile(file)
	if err != nil {
		return fmt.Errorf("error executing file '%s': %w", file, err)
	}
	return nil
}
