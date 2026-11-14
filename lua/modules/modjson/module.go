// Package modjson is a module that provides functions to work with JSON
package modjson

import (
	"encoding/json"

	"github.com/LucasAVasco/falcula/lua/modules/base"
	"github.com/LucasAVasco/falcula/lua/modules/modcodec"

	lua "github.com/yuin/gopher-lua"
)

var codec = modcodec.Codec{Name: "JSON", Marshal: json.Marshal, Unmarshal: json.Unmarshal}

// Module is a module that provides functions to work with JSON
type Module struct {
	base.BaseModule
}

func New() *Module {
	return &Module{}
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	L.SetFuncs(mod, codec.Functions())
	return nil
}
