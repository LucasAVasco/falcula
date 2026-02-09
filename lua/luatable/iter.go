package luatable

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// ForEach iterates over a Lua table. Returns an error if the function returns an error. Equivalent to the `ForEach` method of a Lua table
// object, but with support to return an error
func ForEach(table *lua.LTable, fn func(key, value lua.LValue) error) error {
	var err error
	table.ForEach(func(key, value lua.LValue) {
		if err != nil {
			return
		}

		err = fn(key, value)
		if err != nil {
			err = fmt.Errorf("error iterating table (key: %v, value: %v): %w", key, value, err)
		}
	})

	return err
}
