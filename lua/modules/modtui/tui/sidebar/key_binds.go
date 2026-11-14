package sidebar

import (
	"fmt"
	"strings"

	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/keybinds"
	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/manager"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type nodeHandlers struct {
	OnManager func(man *manager.Manager) error
	OnService func(man *manager.Manager, svc *enhanced.EnhancedService) error
	OnElse    func() error
}

// executeFunctionOnCurrentNode executes a function in the current node. The current node can be a manager, a service, or neither. The
// callback is executed based on the current node type.
//
// If the current node is neither a manager nor a service, `OnElse` is executed. If it is not set, `OnManager` is executed on each manager
func (s *Sidebar) executeFunctionOnCurrentNode(handlers *nodeHandlers) error {
	node := s.getCurrentNode()
	nodeReference := node.GetReference()

	// Current node is a manager
	if man, ok := nodeReference.(*manager.Manager); ok {
		if handlers.OnManager == nil {
			return nil
		}

		return handlers.OnManager(man)
	}

	// Current node is a service
	if svc, ok := nodeReference.(*enhanced.EnhancedService); ok {
		if handlers.OnService == nil {
			return nil
		}

		// Gets the manager of the current service
		var man *manager.Manager
		s.root.Walk(func(child, parent *tview.TreeNode) bool {
			if child == node {
				man = parent.GetReference().(*manager.Manager)
				return false
			}
			return true
		})

		return handlers.OnService(man, svc)
	}

	// Current node is neither a manager nor a service
	if handlers.OnElse != nil {
		return handlers.OnElse()
	}

	if handlers.OnManager != nil {
		for _, child := range s.root.GetChildren() {
			if man, ok := child.GetReference().(*manager.Manager); ok {
				err := handlers.OnManager(man)
				if err != nil {
					return fmt.Errorf("error handling manager '%s': %w", man.GetName(), err)
				}
			}
		}
	}

	return nil
}

// serviceLnavFilter returns the Lnav filter expression that matches the logs of a service
func serviceLnavFilter(svc *enhanced.EnhancedService) string {
	return fmt.Sprintf(":log_hostname == '%s'", svc.GetName())
}

// managerLnavFilter returns the Lnav filter expression that matches the logs of all services of a manager
func managerLnavFilter(man *manager.Manager) string {
	services := man.GetServices()
	filters := make([]string, 0, len(services))
	for _, svc := range services {
		filters = append(filters, serviceLnavFilter(svc))
	}

	return strings.Join(filters, " or ")
}

// openCurrentNodeInLnav opens the log file in `Lnav` filtered by the current node
func (s *Sidebar) openCurrentNodeInLnav() {
	var err error
	s.app.Suspend(func() {
		err = s.executeFunctionOnCurrentNode(&nodeHandlers{
			OnManager: func(man *manager.Manager) error {
				return s.openLogFileInLnav(managerLnavFilter(man))
			},
			OnService: func(man *manager.Manager, svc *enhanced.EnhancedService) error {
				return s.openLogFileInLnav(serviceLnavFilter(svc))
			},
			OnElse: func() error {
				return s.openLogFileInLnav("")
			},
		})
	})

	if err != nil {
		s.OnError(fmt.Errorf("error opening log file in Lnav: %w", err))
	}
}

// openLogFileInLessSuspended opens the log file in `less`, suspending the text user interface while it runs
func (s *Sidebar) openLogFileInLessSuspended() {
	var err error
	s.app.Suspend(func() {
		err = s.openLogFileInLess()
	})

	if err != nil {
		s.OnError(fmt.Errorf("error opening log file in less: %w", err))
	}
}

// nodeActionKeyBind creates an asynchronous key bind that executes the handlers on the current node. Errors are reported with the
// `errPrefix` message
func (s *Sidebar) nodeActionKeyBind(key rune, desc, errPrefix string, handlers *nodeHandlers) *keybinds.KeyBind {
	return &keybinds.KeyBind{
		Rune:  key,
		Desc:  desc,
		Async: true,
		Bind: func() {
			err := s.executeFunctionOnCurrentNode(handlers)
			if err != nil {
				s.OnError(fmt.Errorf("%s: %w", errPrefix, err))
			}
		},
	}
}

// restartHandlers returns the handlers that restart the current node
func restartHandlers(force bool) *nodeHandlers {
	return &nodeHandlers{
		OnManager: func(man *manager.Manager) error {
			return man.Restart(force, nil)
		},
		OnService: func(man *manager.Manager, svc *enhanced.EnhancedService) error {
			_, err := svc.Restart(force)
			return err
		},
	}
}

// stopHandlers returns the handlers that stop (or abort the preparation of) the current node
func stopHandlers(force bool) *nodeHandlers {
	return &nodeHandlers{
		OnManager: func(man *manager.Manager) error {
			return man.AbortPrepareOrStop(force, nil).Wait()
		},
		OnService: func(man *manager.Manager, svc *enhanced.EnhancedService) error {
			_, err := svc.AbortPrepareOrStop(force)
			return err
		},
	}
}

// deleteHandlers returns the handlers that delete the current node from its manager and from the side bar
func (s *Sidebar) deleteHandlers(force bool) *nodeHandlers {
	return &nodeHandlers{
		OnManager: func(man *manager.Manager) error {
			err := man.RemoveServices(man.GetServices(), force, nil)
			if err != nil {
				return fmt.Errorf("error removing services from manager: %w", err)
			}

			err = s.RemoveManager(man)
			if err != nil {
				return fmt.Errorf("error removing manager from sidebar: %w", err)
			}

			return nil
		},
		OnService: func(man *manager.Manager, svc *enhanced.EnhancedService) error {
			err := man.RemoveService(svc, force, nil)
			if err != nil {
				return fmt.Errorf("error removing service from manager: %w", err)
			}

			err = s.RemoveService(man, svc)
			if err != nil {
				return fmt.Errorf("error removing service from sidebar: %w", err)
			}

			return nil
		},
	}
}

// setKeyBinds sets the key binds for the side bar
func (s *Sidebar) setKeyBinds() {
	s.keyBindsHandler = keybinds.NewHandler("Sidebar")
	s.keyBindsHandler.AddKeyBinds([]*keybinds.KeyBind{
		// Open current node
		{Key: tcell.KeyEnter, Desc: "Open current node in Lnav", Bind: s.openCurrentNodeInLnav},
		{Rune: 'l', Desc: "Open current node in Less", Bind: s.openLogFileInLessSuspended},
		{Rune: 'L', Desc: "Open current node in Lnav", Bind: s.openCurrentNodeInLnav},

		// Restart
		s.nodeActionKeyBind('r', "Restart current node", "error restarting service", restartHandlers(false)),
		s.nodeActionKeyBind('R', "Restart current node (force)", "error restarting service", restartHandlers(true)),

		// Stop
		s.nodeActionKeyBind('s', "Stop or abort prepare current node", "error stopping service", stopHandlers(false)),
		s.nodeActionKeyBind('S', "Stop or abort prepare current node (force)", "error stopping service", stopHandlers(true)),

		// Delete
		s.nodeActionKeyBind('x', "Delete service", "error deleting service", s.deleteHandlers(false)),
		s.nodeActionKeyBind('X', "Delete service (force deletion)", "error deleting service", s.deleteHandlers(true)),

		// Disable and enable
		s.nodeActionKeyBind('d', "Disable service (must be enabled to be started)", "error disabling service", &nodeHandlers{
			OnManager: func(man *manager.Manager) error {
				return man.Disable(nil).Wait()
			},
			OnService: func(man *manager.Manager, svc *enhanced.EnhancedService) error {
				return svc.Disable()
			},
		}),
		s.nodeActionKeyBind('e', "Enable service", "error enabling service", &nodeHandlers{
			OnManager: func(man *manager.Manager) error {
				return man.Enable(nil).Wait()
			},
			OnService: func(man *manager.Manager, svc *enhanced.EnhancedService) error {
				return svc.Enable()
			},
		}),
	})

	// Adds the key binds to the tree
	s.tree.SetInputCapture(s.keyBindsHandler.GetInputCaptureFunction())
}

func (s *Sidebar) GetKeyBindsHandler() *keybinds.Handler {
	return s.keyBindsHandler
}
