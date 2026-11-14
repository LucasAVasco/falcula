// Package modpath is a module that provides functions and classes for working with paths
package modpath

import (
	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/luapath"
	"github.com/LucasAVasco/falcula/lua/luatable"
	"github.com/LucasAVasco/falcula/lua/modules/base"

	lua "github.com/yuin/gopher-lua"
)

type Module struct {
	base.BaseModule
}

func New() *Module {
	return &Module{}
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	L.SetFuncs(mod, map[string]lua.LGFunction{
		"get_current_file": func(L *lua.LState) int {
			file, err := luapath.GetCurrentLuaFile(L)

			if err != nil {
				return luaerror.Push(L, 1, err)
			}

			L.Push(lua.LString(file))
			return 1
		},

		"get_current_dir": func(L *lua.LState) int {
			file, err := luapath.GetCurrentLuaDirectory(L)

			if err != nil {
				return luaerror.Push(L, 1, err)
			}

			L.Push(lua.LString(file))
			return 1
		},

		"abs": func(L *lua.LState) int {
			path := L.ToString(1)
			path, err := luapath.GetAbs(L, path)
			if err != nil {
				return luaerror.Push(L, 1, err)
			}

			L.Push(lua.LString(path))
			return 1
		},

		// Calls "abs" to each element of the table and returns a table with the absolute paths
		"abs_list": func(L *lua.LState) int {
			return mapPaths(L, L.ToTable(1), func(path string) (string, error) {
				return luapath.GetAbs(L, path)
			})
		},

		"rel": func(L *lua.LState) int {
			path := L.ToString(1)
			base := L.OptString(2, "")

			path, err := luapath.GetRel(L, path, base)
			if err != nil {
				return luaerror.Push(L, 1, err)
			}

			L.Push(lua.LString(path))
			return 1
		},

		// Calls "rel" to each element of the table and returns a table with the relative paths
		"rel_list": func(L *lua.LState) int {
			base := L.OptString(2, "")
			return mapPaths(L, L.ToTable(1), func(path string) (string, error) {
				return luapath.GetRel(L, path, base)
			})
		},
	})

	return nil
}

// mapPaths applies the conversion function to each path of the table and pushes a list with the converted paths. Stops at the first error
func mapPaths(L *lua.LState, paths *lua.LTable, convert func(path string) (string, error)) int {
	res := L.NewTable()
	err := luatable.ForEach(paths, func(_, value lua.LValue) error {
		path, err := convert(value.String())
		if err != nil {
			return err
		}

		res.Append(lua.LString(path))
		return nil
	})
	if err != nil {
		return luaerror.Push(L, 1, err)
	}

	L.Push(res)
	return 1
}
