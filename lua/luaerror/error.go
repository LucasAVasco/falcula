// Package luaerror contains functions for working with errors in Lua
package luaerror

import (
	"sync/atomic"

	lua "github.com/yuin/gopher-lua"
)

var abortOnError atomic.Bool

// Push pushes an error message to the Lua stack. The user can access the error message as the return value at index
// `numReturnWithoutError + 1` where `numReturnWithoutError` is the number of returns that would be returned if no error occurred
func Push(L *lua.LState, numReturnWithoutError int, err error) int {
	if abortOnError.Load() {
		L.RaiseError("error running Lua: %v", err)
	}

	for range numReturnWithoutError {
		L.Push(lua.LNil)
	}

	// Returns the error
	L.Push(lua.LString(err.Error()))

	// Returns the number of returns with error
	return numReturnWithoutError + 1
}

// ConfigureAbortOnError configures whether the script should abort on error (abort = true) or just return the error appended to the end of
// the return values (abort = false)
func ConfigureAbortOnError(abort bool) {
	abortOnError.Store(abort)
}

func init() {
	abortOnError.Store(true)
}
