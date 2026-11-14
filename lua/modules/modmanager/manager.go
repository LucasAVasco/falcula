package modmanager

import (
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luaclass"
	"github.com/LucasAVasco/falcula/lua/luadata"
	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/iface"
	"github.com/LucasAVasco/falcula/service/manager"

	lua "github.com/yuin/gopher-lua"
)

// getManager gets the manager when called inside a method. Must not be used outside a method
func getManager(L *lua.LState) *manager.Manager {
	return luaclass.GetAttributeOf[*manager.Manager](L, "_manager")
}

// serviceFromValue gets the service of a Lua value. Returns false if the value is not a service
func serviceFromValue(value lua.LValue) (iface.Service, bool) {
	userData, ok := value.(*lua.LUserData)
	if !ok {
		return nil, false
	}

	svc, ok := userData.Value.(iface.Service)
	return svc, ok
}

// returnErrorMessage returns an error message if err is not nil. Must not be used outside a method
func (m *Module) returnErrorMessage(L *lua.LState, err error) int {
	if err == nil {
		return 0
	}

	m.Config.Runtime.Logger.LogError(err)
	return luaerror.Push(L, 0, err)
}

func (m *Module) GetMethods() map[string]lua.LGFunction {
	createServiceCallbacks := func(man *manager.Manager) *enhanced.Callbacks {
		return &enhanced.Callbacks{
			OnServiceStatusChanged: func(svc *enhanced.EnhancedService) {
				m.callbacks.OnServiceStatusChanged(man, svc)
			},
		}
	}

	return map[string]lua.LGFunction{
		"add_service": func(L *lua.LState) int {
			man := getManager(L)
			svc, ok := luadata.GetValueFromArgs(L, 2).(iface.Service)
			if !ok {
				L.ArgError(2, "service expected")
			}
			enhancedService := man.AddService(svc, createServiceCallbacks(man))
			m.callbacks.OnAddService(man, enhancedService)
			return 0
		},

		"add_services": func(L *lua.LState) int {
			man := getManager(L)
			services := L.CheckTable(2)

			for i := 1; i <= services.Len(); i++ {
				svc, ok := serviceFromValue(services.RawGetInt(i))
				if !ok {
					L.ArgError(2, fmt.Sprintf("list of services expected (element %d is not a service)", i))
				}
				enhancedService := man.AddService(svc, createServiceCallbacks(man))
				m.callbacks.OnAddService(man, enhancedService)
			}

			return 0
		},

		"start_prepare": func(L *lua.LState) int {
			man := getManager(L)
			man.StartPrepare(nil)
			return m.returnErrorMessage(L, nil)
		},

		"wait_prepare": func(L *lua.LState) int {
			man := getManager(L)
			return m.returnErrorMessage(L, man.WaitPrepare(nil))
		},

		"prepare": func(L *lua.LState) int {
			man := getManager(L)
			return m.returnErrorMessage(L, man.Prepare(nil).Wait())
		},

		"abort_prepare": func(L *lua.LState) int {
			man := getManager(L)
			force := L.OptBool(2, false)
			return m.returnErrorMessage(L, man.AbortPrepare(force, nil).Wait())
		},

		"start": func(L *lua.LState) int {
			man := getManager(L)
			man.Start(nil)
			return m.returnErrorMessage(L, nil)
		},

		"wait": func(L *lua.LState) int {
			man := getManager(L)
			return m.returnErrorMessage(L, man.Wait(nil))
		},

		"run": func(L *lua.LState) int {
			man := getManager(L)
			return m.returnErrorMessage(L, man.Run(nil))
		},

		"run_serial": func(L *lua.LState) int {
			man := getManager(L)
			return m.returnErrorMessage(L, man.RunSerial(nil, nil))
		},

		"stop": func(L *lua.LState) int {
			man := getManager(L)
			force := L.OptBool(2, false)
			return m.returnErrorMessage(L, man.Stop(force, nil).Wait())
		},

		"close": func(L *lua.LState) int {
			man := getManager(L)
			force := L.OptBool(2, false)
			m.callbacks.OnDeleteManager(man)
			return m.returnErrorMessage(L, man.Close(force, nil))
		},
	}
}
