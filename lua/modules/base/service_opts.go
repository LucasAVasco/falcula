package base

import (
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luapath"
	provbase "github.com/LucasAVasco/falcula/provider/base"

	lua "github.com/yuin/gopher-lua"
)

// ResolveServiceOpts resolves the service options read from Lua: a relative `file` is relative to the directory of the current Lua file
func ResolveServiceOpts(L *lua.LState, opts *provbase.ServiceOpts) error {
	if opts.File == nil || *opts.File == "" {
		return nil
	}

	file, err := luapath.GetAbs(L, *opts.File)
	if err != nil {
		return fmt.Errorf("error getting absolute path of the service file '%s': %w", *opts.File, err)
	}
	opts.File = &file

	return nil
}
