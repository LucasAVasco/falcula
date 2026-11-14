// Package modfs is a module that provides filesystem related functions
package modfs

import (
	"errors"
	"fmt"

	"github.com/LucasAVasco/falcula/filesystem/watcher"
	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/luatable"
	"github.com/LucasAVasco/falcula/lua/maplua"
	"github.com/LucasAVasco/falcula/lua/modules/base"
	lua "github.com/yuin/gopher-lua"
)

// Module is a module that provides filesystem related functions
type Module struct {
	base.BaseModule
}

func New() *Module {
	return &Module{}
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	L.SetFuncs(mod, functions)

	operationsTable := L.NewTable()
	operationsCodesTable := L.NewTable()
	for _, op := range []watcher.WatcherOperation{
		watcher.WatcherOperationCreate,
		watcher.WatcherOperationWrite,
		watcher.WatcherOperationRemove,
		watcher.WatcherOperationRename,
		watcher.WatcherOperationChmod,
	} {
		operationsTable.RawSetString(op.String(), lua.LNumber(int(op)))
		operationsCodesTable.RawSetInt(int(op), lua.LString(op.String()))
	}

	L.SetField(mod, "OPERATIONS", operationsTable)
	L.SetField(mod, "OP", operationsTable)
	L.SetField(mod, "OPERATIONS_CODES", operationsCodesTable)
	L.SetField(mod, "OP_CODES", operationsCodesTable)

	return nil
}

// WatchParallelEntry is a struct that represents a entry for the 'watch_parallel' function.
type WatchParallelEntry struct {
	Files    []string             `lua:"files"`    // files to watch.
	Callback *lua.LFunction       `lua:"callback"` // callback function. Will receive the file and the operation.
	Opts     *watcher.WatcherOpts `lua:"opts"`     // Watcher options.
}

var functions = map[string]lua.LGFunction{
	"watch": func(L *lua.LState) int {
		// Arguments
		entry := WatchParallelEntry{
			Files:    luatable.GetStringsFromLuaTableThatKeyIsNumber(L.CheckTable(1)),
			Callback: L.CheckFunction(2),
			Opts:     &watcher.WatcherOpts{},
		}
		err := maplua.Unmarshal(L.OptTable(3, L.NewTable()), entry.Opts)
		if err != nil {
			return luaerror.Push(L, 0, fmt.Errorf("error unmarshalling options: %w", err))
		}

		err = watchAll(L, []WatchParallelEntry{entry})
		if err != nil {
			return luaerror.Push(L, 0, err)
		}

		return 0
	},

	"watch_parallel": func(L *lua.LState) int {
		entries := []WatchParallelEntry{}
		err := maplua.Unmarshal(L.CheckTable(1), &entries)
		if err != nil {
			return luaerror.Push(L, 0, fmt.Errorf("error unmarshalling argument: %w", err))
		}

		err = watchAll(L, entries)
		if err != nil {
			return luaerror.Push(L, 0, err)
		}

		return 0
	},
}

// callbackCall is a request to call a Lua callback because of a file system event
type callbackCall struct {
	callback  *lua.LFunction
	file      string
	operation watcher.WatcherOperation
	result    chan error
}

// watchAll creates a watcher for each entry and waits for all of them to stop. Returns the first error.
//
// The watchers detect events in their own goroutines, but the Lua state can only be used by the goroutine that is running it. So the
// watchers send their callbacks to this function, which calls them one at a time
func watchAll(L *lua.LState, entries []WatchParallelEntry) error {
	calls := make(chan callbackCall)
	stop := make(chan struct{}) // Closed when this function returns. Unblocks the watchers that are sending callbacks
	defer close(stop)

	// Create watchers for each entry
	watchers := make([]*watcher.Watcher, 0, len(entries))
	for _, entry := range entries {
		fileWatcher, err := createWatcher(entry, calls, stop)
		if err != nil {
			return fmt.Errorf("error creating watcher: %w", err)
		}
		defer fileWatcher.Close()

		watchers = append(watchers, fileWatcher)
	}

	// Wait for all watchers to stop in parallel
	errorChannel := make(chan error, len(watchers))
	for _, fileWatcher := range watchers {
		go func() {
			err := fileWatcher.Wait()
			if err != nil {
				err = fmt.Errorf("error waiting for watcher: %w", err)
			}
			errorChannel <- err
		}()
	}

	// Stops watching if the script is interrupted (see `LState.SetContext`)
	var interrupted <-chan struct{} // Nil channel (blocks forever) if there is no context
	if ctx := L.Context(); ctx != nil {
		interrupted = ctx.Done()
	}

	// Calls the callbacks while waiting
	for remaining := len(watchers); remaining > 0; {
		select {
		case <-interrupted:
			return errors.New("the script was interrupted")

		case call := <-calls:
			L.Push(call.callback)
			L.Push(lua.LString(call.file))
			L.Push(lua.LNumber(call.operation))
			call.result <- L.PCall(2, 0, nil)

		case err := <-errorChannel:
			if err != nil {
				return err
			}
			remaining--
		}
	}

	// Close watchers
	for _, fileWatcher := range watchers {
		err := fileWatcher.Close()
		if err != nil {
			return fmt.Errorf("error closing watcher: %w", err)
		}
	}

	return nil
}

// createWatcher creates a new file system watcher for the entry. Its callback is sent to the `calls` channel (see `watchAll`). If the
// `stop` channel is closed, the callback is not called and returns an error
func createWatcher(entry WatchParallelEntry, calls chan<- callbackCall, stop <-chan struct{}) (*watcher.Watcher, error) {
	errStopped := errors.New("the function that watches the files has returned")

	fileWatcher, err := watcher.NewWatcher(func(file string, operation watcher.WatcherOperation) error {
		call := callbackCall{callback: entry.Callback, file: file, operation: operation, result: make(chan error, 1)}

		select {
		case calls <- call:
		case <-stop:
			return errStopped
		}

		select {
		case err := <-call.result:
			return err
		case <-stop:
			return errStopped
		}
	}, entry.Opts)
	if err != nil {
		return nil, fmt.Errorf("error creating new watcher: %w", err)
	}

	err = fileWatcher.AddFiles(entry.Files)
	if err != nil {
		return nil, fmt.Errorf("error adding files to watcher: %w", err)
	}

	return fileWatcher, nil
}
