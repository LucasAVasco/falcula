// Package modcodec implements the functions shared by the modules that encode and decode Lua values to and from a text format
package modcodec

import (
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/maplua"

	lua "github.com/yuin/gopher-lua"
)

// Codec is a text format that encodes and decodes Go values
type Codec struct {
	Name      string // Name of the format used in error messages (e.g. "JSON")
	Marshal   func(value any) ([]byte, error)
	Unmarshal func(data []byte, value any) error
}

// Functions returns the `encode` and `decode` Lua functions of the codec
func (c *Codec) Functions() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"encode": c.encode,
		"decode": c.decode,
	}
}

// encode encodes the Lua value at the first argument to a string
func (c *Codec) encode(L *lua.LState) int {
	// Lua to Go object
	var goValue any
	err := maplua.Unmarshal(L.Get(1), &goValue)
	if err != nil {
		return luaerror.Push(L, 1, fmt.Errorf("error mapping Lua object to Go object: %w", err))
	}

	// Go object to string
	encoded, err := c.Marshal(goValue)
	if err != nil {
		return luaerror.Push(L, 1, fmt.Errorf("error encoding Go object to %s: %w", c.Name, err))
	}

	L.Push(lua.LString(encoded))
	return 1
}

// decode decodes the string at the first argument to a Lua value
func (c *Codec) decode(L *lua.LState) int {
	// String to Go object
	var goValue any
	err := c.Unmarshal([]byte(L.ToString(1)), &goValue)
	if err != nil {
		return luaerror.Push(L, 1, fmt.Errorf("error decoding %s to Go object: %w", c.Name, err))
	}

	// Go object to Lua
	luaValue, err := maplua.Marshal(goValue)
	if err != nil {
		return luaerror.Push(L, 1, fmt.Errorf("error mapping Golang object to Lua object: %w", err))
	}

	L.Push(luaValue)
	return 1
}
