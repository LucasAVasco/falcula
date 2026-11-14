// Package sidebar contains the side bar of the application.
package sidebar

import (
	"fmt"
	"slices"
	"sync/atomic"

	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/app"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/keybinds"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/theme"
	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/manager"

	"github.com/rivo/tview"
)

// Sidebar is the side bar widget of the application.
type Sidebar struct {
	app             *app.App
	theme           *theme.Theme
	keyBindsHandler *keybinds.Handler
	logFilePath     string

	// Widgets

	tree     *tview.TreeView
	root     *tview.TreeNode
	selected atomic.Pointer[tview.TreeNode] // Selected node (its text has no colors, they would be hard to read on the selection background)

	// Callbacks
	OnError func(err error)
}

// Create a new side bar widget.
func New(app *app.App, logFilePath string, th *theme.Theme) *Sidebar {
	s := Sidebar{
		app:         app,
		theme:       th,
		logFilePath: logFilePath,

		// Callbacks
		OnError: func(err error) {},
	}

	s.tree = tview.NewTreeView()
	s.tree.SetGraphicsColor(th.Border)
	s.root = s.newNode("Service managers")
	s.tree.SetRoot(s.root)
	s.tree.SetCurrentNode(s.root)
	s.selected.Store(s.root)
	s.tree.SetChangedFunc(s.onSelectionChanged) // Called in the application goroutine before drawing the nodes
	s.tree.SetBorder(true).SetTitle("Service managers")
	th.FocusBorder(s.tree.Box)

	s.setKeyBinds()

	return &s
}

// newNode creates a tree node with the theme styles. The text may have color tags
func (s *Sidebar) newNode(text string) *tview.TreeNode {
	return tview.NewTreeNode(text).SetTextStyle(s.theme.TextStyle()).SetSelectedTextStyle(s.theme.SelectionStyle())
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
	newNode := s.newNode(tview.Escape(man.GetName())).SetReference(man).SetSelectable(true)
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
func (s *Sidebar) RemoveAllManagers() {
	s.root.ClearChildren()

	// Updates the UI
	s.app.Draw()
}

// }}}

// service functions {{{

// getServiceNode gets the node of a service and the node of its manager. Returns an error if the manager is not in the side bar. The
// service node is nil if the manager node does not contain the service
func (s *Sidebar) getServiceNode(
	man *manager.Manager,
	svc *enhanced.EnhancedService,
) (managerNode *tview.TreeNode, serviceNode *tview.TreeNode, err error) {
	managerNode = s.getManagerNode(man)
	if managerNode == nil {
		return nil, nil, fmt.Errorf("manager '%s' not found", man.GetName())
	}

	// Gets the service from the manager node
	services := managerNode.GetChildren()
	index := slices.IndexFunc(services, func(child *tview.TreeNode) bool {
		return child.GetReference() == svc
	})

	if index == -1 {
		return managerNode, nil, nil
	}

	return managerNode, services[index], nil
}

// getExistingServiceNode is similar to `getServiceNode`, but also returns an error if the service is not in the side bar
func (s *Sidebar) getExistingServiceNode(
	man *manager.Manager,
	svc *enhanced.EnhancedService,
) (managerNode *tview.TreeNode, serviceNode *tview.TreeNode, err error) {
	managerNode, serviceNode, err = s.getServiceNode(man, svc)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting node of service '%s' of manager '%s': %w", svc.GetName(), man.GetName(), err)
	}
	if serviceNode == nil {
		return nil, nil, fmt.Errorf("service '%s' not found in manager '%s'", svc.GetName(), man.GetName())
	}

	return managerNode, serviceNode, nil
}

// generateServiceText gets the text to show in the service node. The status of a selected node is not colored (the selection style is
// used, the status color would be hard to read on the selection background)
func (s *Sidebar) generateServiceText(svc *enhanced.EnhancedService, selected bool) string {
	svcStatus := svc.GetStatus()
	statusText := svcStatus.String()
	if !selected {
		statusText = theme.Colorize(statusText, s.theme.StatusColor(svcStatus))
	}
	return tview.Escape(svc.GetName()) + " (" + statusText + ")"
}

// updateServiceNodeText updates the text of a service node. Does nothing if the node is not a service node
func (s *Sidebar) updateServiceNodeText(node *tview.TreeNode) {
	if node == nil {
		return
	}
	if svc, ok := node.GetReference().(*enhanced.EnhancedService); ok {
		node.SetText(s.generateServiceText(svc, node == s.selected.Load()))
	}
}

// onSelectionChanged updates the texts of the previous and the new selected nodes
func (s *Sidebar) onSelectionChanged(node *tview.TreeNode) {
	previous := s.selected.Swap(node)
	s.updateServiceNodeText(previous)
	s.updateServiceNodeText(node)
}

// AddService adds a service to the side bar
func (s *Sidebar) AddService(man *manager.Manager, svc *enhanced.EnhancedService) error {
	// Checks if the service already exists
	managerNode, serviceNode, err := s.getServiceNode(man, svc)
	if err != nil {
		return fmt.Errorf("error checking if service exists: %w", err)
	}

	if serviceNode != nil {
		return fmt.Errorf("service '%s' already exists in manager '%s'", svc.GetName(), man.GetName())
	}

	// Creates the new service node and adds it to the manager node as a child
	newNode := s.newNode("").SetReference(svc).SetSelectable(true)
	s.updateServiceNodeText(newNode)
	managerNode.AddChild(newNode)

	// Updates the UI
	s.app.Draw()

	return nil
}

// RemoveService removes a service from the side bar
func (s *Sidebar) RemoveService(man *manager.Manager, svc *enhanced.EnhancedService) error {
	managerNode, node, err := s.getExistingServiceNode(man, svc)
	if err != nil {
		return err
	}

	// Removes the service node from the manager node children
	managerNode.RemoveChild(node)

	// Updates the UI
	s.app.Draw()

	return nil
}

// UpdateServiceStatus updates the status of a service in the side bar
func (s *Sidebar) UpdateServiceStatus(man *manager.Manager, svc *enhanced.EnhancedService) error {
	_, node, err := s.getExistingServiceNode(man, svc)
	if err != nil {
		return err
	}

	// Updates the service text
	s.updateServiceNodeText(node)

	// Updates the UI
	s.app.Draw()

	return nil
}

// }}}
