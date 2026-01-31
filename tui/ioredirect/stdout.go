// Package ioredirect is a package that redirects the output of a Lua state to a Writer
package ioredirect

import (
	"strings"

	"github.com/LucasAVasco/falcula/process"

	lua "github.com/yuin/gopher-lua"
)

type Writer interface {
	Append(message ...any)
	Write(message []byte) (n int, err error)
}

// Redirect redirects the output of a Lua state to a Writer
func Redirect(L *lua.LState, out Writer) {
	// Overrides the `print` to show logs in the debug logs preview
	L.SetGlobal("print", L.NewFunction(func(l *lua.LState) int {
		numArgs := l.GetTop()

		if numArgs == 1 {
			message := l.Get(1).String()
			out.Append(message, "\n")
		} else if numArgs > 1 {
			args := make([]string, numArgs)
			for i := 1; i <= numArgs; i++ {
				args[i-1] = l.Get(i).String()
			}

			out.Append(strings.Join(args, "\t"), "\n")
		}

		return 0
	}))

	// Overrides the `os.execute` to show logs in the debug logs preview
	globalOs := L.GetGlobal("os")
	L.SetField(globalOs, "execute", L.NewFunction(func(L *lua.LState) int {
		// If the command is not provided, `os.execute` returns the availability of the shell
		if L.Get(1) == lua.LNil {
			L.Push(lua.LBool(process.ShellIsAvailable()))
			return 1
		}

		// Gets the command to run
		shellCommand := L.ToString(1)

		// Runs the shell command using the Go 'exec' package
		cmd := process.CreateCmd(true, shellCommand)
		cmd.Stderr = out
		cmd.Stdout = out
		err := cmd.Run()

		// Pushes the exit code
		exitCode := process.GetExitCodeFromError(err)

		// First return value (success or not)
		if exitCode == 0 {
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}

		if cmd.ProcessState.Exited() {
			// Second return value (the process exited by without a signal)
			L.Push(lua.LString("exit"))

			// Third return value (exit code)
			L.Push(lua.LNumber(exitCode))
		} else {
			// Second return value (the process exited by a signal)
			L.Push(lua.LString("signal"))

			// Third return value (signal code)
			L.Push(lua.LNumber(process.GetExitSignalFromError(err)))
		}

		// Shows the error if any in the logs
		if err != nil {
			out.Append(err)
		}

		return 3
	}))

	// Overrides the `globalIo.write` to show logs in the debug logs preview
	globalIo := L.GetGlobal("io")
	L.SetField(globalIo, "write", L.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		for i := 1; i <= numArgs; i++ {
			message := L.Get(i).String()
			out.Append(message)
		}

		return 0
	}))
}
