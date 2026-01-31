package modules

import (
	"github.com/LucasAVasco/falcula/lua/modules/base"

	lua "github.com/yuin/gopher-lua"
)

// Module is the interface that all modules must implement
type Module interface {
	SetOpts(*base.Options)
	Loader(L *lua.LState, name string, mod *lua.LTable) error
	Close() error
}
