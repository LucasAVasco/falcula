// Package modyaml is a module that provides functions to work with YAML
package modyaml

import (
	"github.com/LucasAVasco/falcula/lua/modules/base"
	"github.com/LucasAVasco/falcula/lua/modules/modcodec"
	"github.com/goccy/go-yaml"

	lua "github.com/yuin/gopher-lua"
)

var codec = modcodec.Codec{Name: "YAML", Marshal: yaml.Marshal, Unmarshal: yaml.Unmarshal}

// Module is a module that provides functions to work with YAML
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
