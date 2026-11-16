package luaruntime

import (
	"fmt"
	"slices"

	"github.com/LucasAVasco/falcula/service/manager"
)

// AddManager adds a service manager to the runtime list. Safe for concurrent use
func (r *Runtime) AddManager(manager *manager.Manager) {
	r.managersMutex.Lock()
	defer r.managersMutex.Unlock()

	r.managers = append(r.managers, manager)
}

// GetManagers gets a copy of the list of service managers in the runtime list. Safe for concurrent use
func (r *Runtime) GetManagers() []*manager.Manager {
	r.managersMutex.Lock()
	defer r.managersMutex.Unlock()

	return slices.Clone(r.managers)
}

// RemoveManager removes a service manager from the runtime list. Safe for concurrent use
func (r *Runtime) RemoveManager(man *manager.Manager) {
	r.managersMutex.Lock()
	defer r.managersMutex.Unlock()

	r.managers = slices.DeleteFunc(r.managers, func(m *manager.Manager) bool {
		return m == man
	})
}

// HasActiveServices returns true if a service of any manager is active (preparing, running or stopping). Safe for concurrent use
func (r *Runtime) HasActiveServices() bool {
	for _, man := range r.GetManagers() {
		for _, svc := range man.GetServices() {
			if svc.GetStatus().IsActive() {
				return true
			}
		}
	}

	return false
}

// closeAllManagers closes all service managers and removes them from the runtime list. Safe for concurrent use
func (r *Runtime) closeAllManagers() {
	// NOTE(LucasAVasco): the available arguments are kept, they are the configuration of the task (not of the script run)
	r.onSetScriptCurrentArgs([]string{})

	// Takes the managers out of the list before closing them (closing may take a while and must not block the list)
	r.managersMutex.Lock()
	managers := r.managers
	r.managers = nil
	r.managersMutex.Unlock()

	for _, man := range managers {
		err := man.Close(true, nil)
		if err != nil {
			r.Logger.LogError(fmt.Errorf("error closing manager '%s': %w", man.GetName(), err))
		}
	}
}
