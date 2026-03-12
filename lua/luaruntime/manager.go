package luaruntime

import (
	"fmt"
	"slices"

	"github.com/LucasAVasco/falcula/service/manager"
)

// AddManager adds a service manager to the runtime list
func (r *Runtime) AddManager(manager *manager.Manager) {
	r.managers = append(r.managers, manager)
}

// GetManagers gets the list of service managers in the runtime list
func (r *Runtime) GetManagers() []*manager.Manager {
	return r.managers
}

// RemoveManager removes a service manager from the runtime list
func (r *Runtime) RemoveManager(man *manager.Manager) {
	index := slices.Index(r.managers, man)
	r.managers = slices.Delete(r.managers, index, index+1)
}

func (r *Runtime) closeAllManagersWithoutLock() {
	r.onSetScriptCurrentArgs([]string{})
	r.onSetScriptAvailableArgs([][]string{})

	for _, man := range r.managers {
		err := man.Close(true, nil)
		if err != nil {
			r.Logger.LogError(fmt.Errorf("error closing manager '%s': %v", man.GetName(), err.Error()))
		}
	}

	r.managers = make([]*manager.Manager, 0)
}

// CloseAllManagers closes all service managers. Errors are logged in the logger
func (r *Runtime) CloseAllManagers() {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	r.closeAllManagersWithoutLock()
}
