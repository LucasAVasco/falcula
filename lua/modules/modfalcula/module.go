// Package modfalcula is a module that provides general functions to configure Falcula
package modfalcula

import (
	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/modules/base"

	lua "github.com/yuin/gopher-lua"
)

// Module is a module that provides general functions to configure Falcula
type Module struct {
	base.BaseModule
}

func New() *Module {
	return &Module{}
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	L.SetFuncs(mod, map[string]lua.LGFunction{
		"configure_abort_on_error": func(L *lua.LState) int {
			abortOnError := bool(L.Get(1).(lua.LBool))
			luaerror.ConfigureAbortOnError(abortOnError)
			return 0
		},
	})

	return nil
}
