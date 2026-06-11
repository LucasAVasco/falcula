// Package modfs is a module that provides filesystem related functions
package modfs

import (
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
	Opts     *watcher.WatcherOpts `lua:"opts"`     // watcher options.
}

var functions = map[string]lua.LGFunction{
	"watch": func(L *lua.LState) int {
		// Arguments
		files := luatable.GetStringsFromLuaTableThatKeyIsNumber(L.CheckTable(1))
		callback := L.CheckFunction(2)
		opts := watcher.WatcherOpts{}
		err := maplua.Unmarshal(L.OptTable(3, L.NewTable()), &opts)
		if err != nil {
			return luaerror.Push(L, 0, fmt.Errorf("error unmarshalling options: %w", err))
		}

		// Create watcher
		watcher, err := createWatcher(L, files, callback, &opts)
		if err != nil {
			return luaerror.Push(L, 0, fmt.Errorf("error creating watcher: %w", err))
		}
		defer watcher.Close()

		// Wait for watcher to stop
		err = watcher.Wait()
		if err != nil {
			return luaerror.Push(L, 0, fmt.Errorf("error waiting for watcher: %w", err))
		}

		// Close watcher
		err = watcher.Close()
		if err != nil {
			return luaerror.Push(L, 0, fmt.Errorf("error closing watcher: %w", err))
		}

		return 0
	},

	"watch_parallel": func(L *lua.LState) int {
		entries := []WatchParallelEntry{}
		err := maplua.Unmarshal(L.CheckTable(1), &entries)
		if err != nil {
			return luaerror.Push(L, 0, fmt.Errorf("error unmarshalling argument: %w", err))
		}

		// Create watchers for each entry
		watchers := make([]*watcher.Watcher, 0, len(entries))
		for _, entry := range entries {
			watcher, err := createWatcher(L, entry.Files, entry.Callback, entry.Opts)
			if err != nil {
				return luaerror.Push(L, 0, fmt.Errorf("error creating watcher: %w", err))
			}
			defer watcher.Close()

			watchers = append(watchers, watcher)
		}

		// Wait for all watchers to stop in parallel (using goroutines). Send their errors to a channel and waits for all of them to stop
		errorChannel := make(chan error, len(watchers))
		for _, watcher := range watchers {
			func() {
				err := watcher.Wait()
				if err != nil {
					errorChannel <- fmt.Errorf("error waiting for watcher: %w", err)
				}
				errorChannel <- nil
			}()
		}

		for i := 0; i < len(watchers); i++ {
			err := <-errorChannel
			if err != nil {
				return luaerror.Push(L, 0, err)
			}
		}

		// Close watchers
		for _, watcher := range watchers {
			err := watcher.Close()
			if err != nil {
				return luaerror.Push(L, 0, fmt.Errorf("error closing watcher: %w", err))
			}
		}

		return 0
	},
}

// createWatcher creates a new file system Watcher with the given files, callback function and options.
func createWatcher(L *lua.LState, files []string, callback *lua.LFunction, opts *watcher.WatcherOpts) (*watcher.Watcher, error) {
	watcher, err := watcher.NewWatcher(func(file string, operation watcher.WatcherOperation) error {
		L.Push(callback)
		L.Push(lua.LString(file))
		L.Push(lua.LNumber(operation))
		return L.PCall(2, 0, nil)
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("error creating new watcher: %w", err)
	}

	err = watcher.AddFiles(files)
	if err != nil {
		return nil, fmt.Errorf("error adding files to watcher: %w", err)
	}

	return watcher, nil
}
