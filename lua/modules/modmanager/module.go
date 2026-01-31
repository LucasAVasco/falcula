// Package modmanager is a module that provides functions and classes for working with service managers
package modmanager

import (
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luaclass"
	"github.com/LucasAVasco/falcula/lua/modules/base"
	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/manager"

	lua "github.com/yuin/gopher-lua"
)

// Callbacks is a struct that contains callbacks for the module. All callbacks are optional
type Callbacks struct {
	OnNewManager           func(man *manager.Manager)
	OnDeleteManager        func(man *manager.Manager)
	OnAddService           func(man *manager.Manager, svc *enhanced.EnhancedService)
	OnServiceStatusChanged func(man *manager.Manager, svc *enhanced.EnhancedService)
	OnServiceLog           func(svc *enhanced.EnhancedService, message string)
	OnDebugLog             func(message string)
}

// Module is a module that provides functions and classes for working with service managers
type Module struct {
	base.BaseModule

	managers  []*manager.Manager
	callbacks *Callbacks
}

func New(callbacks *Callbacks) *Module {
	m := Module{
		managers: make([]*manager.Manager, 0),

		callbacks: callbacks,
	}

	return &m
}

func (m *Module) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	info := luaclass.Info{
		Name: "ServiceManager",
		Constructor: func(L *lua.LState, newObj *lua.LTable) error {
			name := L.ToString(2)

			man := manager.New(name)
			m.managers = append(m.managers, man)
			luaclass.SetAttribute(L, newObj, "_manager", man)

			m.callbacks.OnNewManager(man)

			return nil
		},
		Methods: m.GetMethods(),
	}

	class, err := luaclass.New(L, &info, nil)
	if err != nil {
		return fmt.Errorf("error creating class '%s' of '%s' module: %w", info.Name, name, err)
	}

	L.SetField(mod, info.Name, class)

	return nil
}

// Close closes the module and all created managers. Can be called multiple times.
func (m *Module) Close() error {
	for {
		if len(m.managers) == 0 {
			break
		}

		// Closes the manager
		man := m.managers[0]
		err := man.Close(true, nil)
		if err != nil {
			m.callbacks.OnDebugLog("error closing manager: " + err.Error() + "\n")
		}

		// Removes the manager from the slice
		m.managers = m.managers[1:]
		m.callbacks.OnDeleteManager(man)
	}

	return nil
}
