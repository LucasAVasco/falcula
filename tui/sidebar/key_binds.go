package sidebar

import (
	"fmt"

	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/manager"
	"github.com/LucasAVasco/falcula/tui/keybinds"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type NodeHandlers struct {
	OnManager func(man *manager.Manager) error
	OnService func(man *manager.Manager, svc *enhanced.EnhancedService) error
	OnElse    func() error
}

// executeFunctionOnCurrentNode executes a function in the current node. The current node can be a manager, a service, or neither. The
// callback is executed based on the current node type
func (s *Sidebar) executeFunctionOnCurrentNode(handlers *NodeHandlers) error {
	if !s.HasApplication() {
		return nil
	}

	node := s.getCurrentNode()
	nodeReference := node.GetReference()

	// Current node is a manager
	if manager, ok := nodeReference.(*manager.Manager); ok {
		if handlers.OnManager == nil {
			return nil
		}

		return handlers.OnManager(manager)
	}

	// Current node is a service
	if service, ok := nodeReference.(*enhanced.EnhancedService); ok {
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

		// Executes the `OnService` callback
		return handlers.OnService(man, service)
	}

	// Current node is neither a manager nor a service
	if handlers.OnElse != nil { // Try to execute the `OnElse` callback
		return handlers.OnElse()
	} else if handlers.OnManager != nil { // Try to execute the `OnManager` callback on each manager
		for _, child := range s.root.GetChildren() {
			if manager, ok := child.GetReference().(*manager.Manager); ok {
				err := handlers.OnManager(manager)
				if err != nil {
					return fmt.Errorf("error handling manager '%s': %w", manager.GetName(), err)
				}
			}
		}
	}

	return nil
}

// setKeyBinds sets the key binds for the side bar
func (s *Sidebar) setKeyBinds() {
	if !s.HasApplication() {
		return
	}

	// Callback called when the user wants to open the log file in `Lnav`
	LnavKeyBind := func() {
		s.app.Suspend(func() {
			s.executeFunctionOnCurrentNode(&NodeHandlers{
				OnManager: func(man *manager.Manager) error {
					services := man.GetServices()
					filter := ""
					for i, svc := range services {
						name := svc.GetName()
						if i > 0 {
							filter += " or "
						}
						filter += fmt.Sprintf(":log_hostname == '%s'", name)
					}
					return s.openLogFileInLnav(filter)
				},
				OnService: func(man *manager.Manager, svc *enhanced.EnhancedService) error {
					filter := fmt.Sprintf(":log_hostname == '%s'", svc.GetName())
					return s.openLogFileInLnav(filter)
				},
				OnElse: func() error {
					return s.openLogFileInLnav("")
				},
			})
		})
	}

	// Key binds hanlder
	s.keyBindsHandler = keybinds.NewHandler("Sidebar")
	s.keyBindsHandler.AddKeyBinds([]*keybinds.KeyBind{
		// Open current node
		{
			Key:  tcell.KeyEnter,
			Desc: "Open current node in Lnav",
			Bind: LnavKeyBind,
		},
		{
			Rune: 'l',
			Desc: "Open current node in Less",
			Bind: func() {
				s.app.Suspend(func() {
					s.openLogFileInLess()
				})
			},
		},
		{
			Rune: 'L',
			Desc: "Open current node in Lnav",
			Bind: LnavKeyBind,
		},
		// Restart
		{
			Rune:  'r',
			Desc:  "Restart current node",
			Async: true,
			Bind: func() {
				err := s.executeFunctionOnCurrentNode(&NodeHandlers{
					OnManager: func(man *manager.Manager) error {
						return man.Restart(false, nil)
					},
					OnService: func(man *manager.Manager, svc *enhanced.EnhancedService) error {
						_, err := svc.Restart(false)
						return err
					},
				})

				if err != nil {
					s.OnError(fmt.Errorf("error restarting service: %w", err))
				}
			},
		},
		{
			Rune:  'R',
			Desc:  "Restart current node",
			Async: true,
			Bind: func() {
				err := s.executeFunctionOnCurrentNode(&NodeHandlers{
					OnManager: func(man *manager.Manager) error {
						return man.Restart(true, nil)
					},
					OnService: func(man *manager.Manager, svc *enhanced.EnhancedService) error {
						_, err := svc.Restart(true)
						return err
					},
				})

				if err != nil {
					s.OnError(fmt.Errorf("error restarting service: %w", err))
				}
			},
		},
		// Stop
		{
			Rune:  's',
			Desc:  "Stop or abort prepare current node",
			Async: true,
			Bind: func() {
				err := s.executeFunctionOnCurrentNode(&NodeHandlers{
					OnManager: func(man *manager.Manager) error {
						return man.AbortPrepareOrStop(false, nil).Wait()
					},
					OnService: func(man *manager.Manager, svc *enhanced.EnhancedService) error {
						_, err := svc.AbortPrepareOrStop(false)
						return err
					},
				})

				if err != nil {
					s.OnError(fmt.Errorf("error stopping service: %w", err))
				}
			},
		},
		{
			Rune:  'S',
			Desc:  "Stop or abort prepare current node (force)",
			Async: true,
			Bind: func() {
				err := s.executeFunctionOnCurrentNode(&NodeHandlers{
					OnManager: func(man *manager.Manager) error {
						return man.AbortPrepareOrStop(true, nil).Wait()
					},
					OnService: func(man *manager.Manager, svc *enhanced.EnhancedService) error {
						_, err := svc.AbortPrepareOrStop(true)
						return err
					},
				})

				if err != nil {
					s.OnError(fmt.Errorf("error stopping service: %w", err))
				}
			},
		},
		// Delete service
		{
			Rune:  'd',
			Desc:  "Delete service",
			Async: true,
			Bind: func() {
				err := s.executeFunctionOnCurrentNode(&NodeHandlers{
					OnManager: func(man *manager.Manager) error {
						err := man.RemoveServices(man.GetServices(), false, nil)
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
						err := man.RemoveService(svc, false, nil)
						if err != nil {
							return fmt.Errorf("error removing service from manager: %w", err)
						}

						err = s.RemoveService(man, svc)
						if err != nil {
							return fmt.Errorf("error removing service from sidebar: %w", err)
						}

						return nil
					},
				})

				if err != nil {
					s.OnError(fmt.Errorf("error stopping service: %w", err))
				}
			},
		},
		{
			Rune:  'D',
			Desc:  "Delete service (force deletion)",
			Async: true,
			Bind: func() {
				err := s.executeFunctionOnCurrentNode(&NodeHandlers{
					OnManager: func(man *manager.Manager) error {
						err := man.RemoveServices(man.GetServices(), true, nil)
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
						err := man.RemoveService(svc, true, nil)
						if err != nil {
							return fmt.Errorf("error removing service from manager: %w", err)
						}

						err = s.RemoveService(man, svc)
						if err != nil {
							return fmt.Errorf("error removing service from sidebar: %w", err)
						}

						return nil
					},
				})

				if err != nil {
					s.OnError(fmt.Errorf("error stopping service: %w", err))
				}
			},
		},
	})

	// Adds the key binds to the tree
	s.tree.SetInputCapture(s.keyBindsHandler.GetInputCaptureFunction())
}

func (s *Sidebar) GetKeyBindsHandler() *keybinds.Handler {
	if !s.HasApplication() {
		return nil
	}

	return s.keyBindsHandler
}
