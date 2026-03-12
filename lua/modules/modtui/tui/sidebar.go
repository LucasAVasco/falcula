package tui

import (
	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/manager"
)

// AddManagerToSidebar adds a manager to the sidebar
func (t *Tui) AddManagerToSidebar(man *manager.Manager) error {
	return t.mainPage.SideBar.AddManager(man)
}

// RemoveManagerFromSidebar removes a manager from the sidebar. Does not delete the manager, just removes the visual representation from the
// sidebar
func (t *Tui) RemoveManagerFromSidebar(man *manager.Manager) error {
	return t.mainPage.SideBar.RemoveManager(man)
}

// AddServiceToSidebar adds a service to a manager in the sidebar
func (t *Tui) AddServiceToSidebar(man *manager.Manager, svc *enhanced.EnhancedService) error {
	return t.mainPage.SideBar.AddService(man, svc)
}

// RemoveServiceFromSidebar removes a service from a manager in the sidebar. Does not delete the service or remove it from the manager, just
// removes the visual representation from the sidebar
func (t *Tui) RemoveServiceFromSidebar(man *manager.Manager, svc *enhanced.EnhancedService) error {
	return t.mainPage.SideBar.RemoveService(man, svc)
}

// UpdateServiceStatusInSidebar updates the status of a service in the sidebar
func (t *Tui) UpdateServiceStatusInSidebar(man *manager.Manager, svc *enhanced.EnhancedService) error {
	return t.mainPage.SideBar.UpdateServiceStatus(man, svc)
}

// RemoveAllManagersFromSidebar removes all managers from the sidebar. Does not delete the managers or its services, just removes the visual
// representation from the sidebar
func (t *Tui) RemoveAllManagersFromSidebar() error {
	return t.mainPage.SideBar.RemoveAllManagers()
}
