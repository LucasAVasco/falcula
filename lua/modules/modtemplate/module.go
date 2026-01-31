// Package modtemplate is a module that provides functions and classes for working with Go templates
package modtemplate

import (
	"fmt"
	"text/template"

	"github.com/LucasAVasco/falcula/lua/luaclass"
	"github.com/LucasAVasco/falcula/lua/modules/base"

	lua "github.com/yuin/gopher-lua"
)

// Module is a module that provides functions and classes for working with Go templates
type Module struct {
	base.BaseModule
	tmplate *template.Template
}

func New() *Module {
	return &Module{}
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	// Register the module functions
	L.SetFuncs(mod, moduleFunctions)

	// Class
	info := luaclass.Info{
		Name: "Template",
		Constructor: func(L *lua.LState, newObj *lua.LTable) error {
			luaclass.SetAttribute(L, newObj, "_funcs", template.FuncMap{})
			return nil
		},
		Methods: templateClassFunctions,
	}

	// Register the 'Template' class
	class, err := luaclass.New(L, &info, m.Opts.OnError)
	if err != nil {
		return fmt.Errorf("error creating class '%s' of '%s' module: %w", info.Name, name, err)

	}

	L.SetField(mod, info.Name, class)

	return nil
}
