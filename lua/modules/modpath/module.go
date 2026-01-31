// Package modpath is a module that provides functions and classes for working with paths
package modpath

import (
	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/luapath"
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
			table := L.ToTable(1)

			res := L.NewTable() // Response table
			index := 1
			var retError error

			table.ForEach(func(k, v lua.LValue) {
				path, err := luapath.GetAbs(L, v.String())
				if err != nil {
					retError = err
					return
				}

				// Sets the absolute path in the response table
				L.RawSetInt(res, index, lua.LString(path))
				index++
			})

			if retError != nil {
				return luaerror.Push(L, 1, retError)
			}

			L.Push(res)
			return 1
		},

		"rel": func(L *lua.LState) int {
			path := L.ToString(1)
			base := L.OptString(2, "")

			path, err := luapath.GetRel(L, path, base)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(lua.LString(path))
			return 1
		},

		// Calls "rel" to each element of the table and returns a table with the relative paths
		"rel_list": func(L *lua.LState) int {
			table := L.ToTable(1)
			base := L.OptString(2, "")

			res := L.NewTable() // Response table
			index := 1
			var retError error

			table.ForEach(func(k, v lua.LValue) {
				path, err := luapath.GetRel(L, v.String(), base)
				if err != nil {
					retError = err
					return
				}

				// Sets the absolute path in the response table
				L.RawSetInt(res, index, lua.LString(path))
				index++
			})

			if retError != nil {
				return luaerror.Push(L, 1, retError)
			}

			L.Push(res)
			return 1
		},
	})

	return nil
}
