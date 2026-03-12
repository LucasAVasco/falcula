// Package sidebar contains the side bar of the application.
package sidebar

import (
	"fmt"
	"slices"

	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/app"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/keybinds"
	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/manager"

	"github.com/rivo/tview"
)

// Sidebar is the side bar widget of the application.
type Sidebar struct {
	app             *app.App
	keyBindsHandler *keybinds.Handler
	logFilePath     string

	// Widgets

	tree *tview.TreeView
	root *tview.TreeNode

	// Callbacks
	OnError func(err error)
}

// Create a new side bar widget.
func New(app *app.App, logFilePath string) *Sidebar {
	s := Sidebar{
		app:         app,
		logFilePath: logFilePath,

		// Callbacks
		OnError: func(err error) {},
	}

	s.tree = tview.NewTreeView()
	s.root = tview.NewTreeNode("Service managers")
	s.tree.SetRoot(s.root)
	s.tree.SetCurrentNode(s.root)
	s.tree.SetBorder(true).SetTitle("Service managers")

	s.setKeyBinds()

	return &s
}

// GetPrimitive returns the primitive of the widget (used to include it in another widget). This function must not be called if outputting
// to the standard output (raw stdout mode)
func (s *Sidebar) GetPrimitive() tview.Primitive {
	return s.tree
}

// SetFocus sets the focus on the side bar
func (s *Sidebar) SetFocus() {
	s.app.SetFocus(s.tree)
}

func (s *Sidebar) getCurrentNode() *tview.TreeNode {
	return s.tree.GetCurrentNode()
}

// manager functions {{{

// getManagerNode gets the node of a service manager
func (s *Sidebar) getManagerNode(man *manager.Manager) *tview.TreeNode {
	children := s.root.GetChildren()

	// Index of the node
	index := slices.IndexFunc(children, func(child *tview.TreeNode) bool {
		return child.GetReference() == man
	})

	if index == -1 {
		return nil
	}

	return children[index]
}

// HasManager returns `true` if the side bar has a manager
func (s *Sidebar) HasManager(man *manager.Manager) bool {
	return s.getManagerNode(man) != nil
}

// AddManager adds a manager to the side bar
func (s *Sidebar) AddManager(man *manager.Manager) error {
	if s.HasManager(man) {
		return fmt.Errorf("manager '%s' already exists", man.GetName())
	}

	// Creates a new manager node and adds it to the root
	newNode := tview.NewTreeNode(man.GetName()).SetReference(man).SetSelectable(true)
	s.root.AddChild(newNode)

	// Updates the UI
	s.app.Draw()

	return nil
}

// RemoveManager removes a manager from the side bar
func (s *Sidebar) RemoveManager(man *manager.Manager) error {
	// Gets the manager node
	node := s.getManagerNode(man)
	if node == nil {
		return fmt.Errorf("manager '%s' not found", man.GetName())
	}

	// Removes it
	s.root.RemoveChild(node)

	// Updates the UI
	s.app.Draw()

	return nil
}

// RemoveAllManagers removes all managers from the side bar
func (s *Sidebar) RemoveAllManagers() error {
	for _, child := range s.root.GetChildren() {
		s.root.RemoveChild(child)
	}

	// Updates the UI
	s.app.Draw()
	return nil
}

// }}}

// service functions {{{

// getServiceNode gets the node of a service
func (s *Sidebar) getServiceNode(man *manager.Manager, svc *enhanced.EnhancedService) (*tview.TreeNode, error) {
	managerNode := s.getManagerNode(man)
	if managerNode == nil {
		return nil, fmt.Errorf("manager '%s' not found", man.GetName())
	}

	// Gets the service from the manager node
	service := managerNode.GetChildren()
	index := slices.IndexFunc(service, func(child *tview.TreeNode) bool {
		return child.GetReference() == svc
	})

	if index == -1 {
		return nil, nil
	}

	return service[index], nil
}

// generateServiceText gets the text to show in the service node
func (s *Sidebar) generateServiceText(svc *enhanced.EnhancedService) string {
	return svc.GetName() + " (" + svc.GetStatus().ToString() + ")"
}

// HasService checks if the side bar contains a service
func (s *Sidebar) HasService(man *manager.Manager, svc *enhanced.EnhancedService) (bool, error) {
	node, err := s.getServiceNode(man, svc)
	if err != nil {
		return false, fmt.Errorf("error getting node of service '%s' of manager '%s': %w", svc.GetName(), man.GetName(), err)
	}

	return node != nil, nil
}

// AddService adds a service to the side bar
func (s *Sidebar) AddService(man *manager.Manager, svc *enhanced.EnhancedService) error {
	// Checks if the service already exists
	serviceNode, err := s.HasService(man, svc)
	if err != nil {
		return fmt.Errorf("error checking if service exists: %w", err)
	}

	if serviceNode {
		return fmt.Errorf("service '%s' already exists in manager '%s'", svc.GetName(), man.GetName())
	}

	// Gets the manager node
	managerNode := s.getManagerNode(man)
	if managerNode == nil {
		return fmt.Errorf("manager '%s' not found", man.GetName())
	}

	// Creates the new service node and adds it to the manager node as a child
	text := s.generateServiceText(svc)
	newNode := tview.NewTreeNode(text).SetReference(svc).SetSelectable(true)
	managerNode.AddChild(newNode)

	// Updates the UI
	s.app.Draw()

	return nil
}

// RemoveService removes a service from the side bar
func (s *Sidebar) RemoveService(man *manager.Manager, svc *enhanced.EnhancedService) error {
	// Gets the service node
	node, err := s.getServiceNode(man, svc)
	if err != nil {
		return fmt.Errorf("error getting node of service '%s' of manager '%s': %w", svc.GetName(), man.GetName(), err)
	}
	if node == nil {
		return fmt.Errorf("service '%s' not found in manager '%s'", svc.GetName(), man.GetName())
	}

	// Removes the service node from the manager node children
	managerNode := s.getManagerNode(man)
	managerNode.RemoveChild(node)

	// Updates the UI
	s.app.Draw()

	return nil
}

// UpdateServiceStatus updates the status of a service in the side bar
func (s *Sidebar) UpdateServiceStatus(man *manager.Manager, svc *enhanced.EnhancedService) error {
	// Gets the service node
	node, err := s.getServiceNode(man, svc)
	if err != nil {
		return fmt.Errorf("error getting node of service '%s' of manager '%s': %w", svc.GetName(), man.GetName(), err)
	}
	if node == nil {
		return fmt.Errorf("service '%s' not found in manager '%s", svc.GetName(), man.GetName())
	}

	// Updates the service text
	node.SetText(s.generateServiceText(svc))

	// Updates the UI
	s.app.Draw()

	return nil
}

// }}}
