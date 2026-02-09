// Package manager implements a service manager
package manager

import (
	"errors"
	"fmt"
	"sync"

	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/iface"
	"github.com/LucasAVasco/falcula/waiter"
)

// Manager is a service manager. It enhances the provided services and manages its lifecycle
type Manager struct {
	name             string
	serviceListMutex sync.Mutex
	services         []*enhanced.EnhancedService
	OnError          func(man *Manager, err error)
}

func New(name string) *Manager {
	return &Manager{
		name:    name,
		OnError: func(man *Manager, err error) {},
	}
}

func (m *Manager) GetName() string {
	return m.name
}

func (m *Manager) Close(force bool, onServiceStop enhanced.ExitStepCallback) error {
	return m.Stop(force, onServiceStop).Wait()
}

// AddService adds a service to the manager and returns the enhanced service. The callback is used to create the enhanced service, but is
// optional
func (m *Manager) AddService(svc iface.Service, callback *enhanced.Callbacks) *enhanced.EnhancedService {
	m.serviceListMutex.Lock()
	defer m.serviceListMutex.Unlock()

	enhancedService := enhanced.NewEnhancedService(svc, callback)
	m.services = append(m.services, enhancedService)

	return enhancedService
}

// AddServices adds multiple services to the manager and returns the enhanced services. The callback is used to create the enhanced service,
// but is optional
func (m *Manager) AddServices(svc []iface.Service, callback *enhanced.Callbacks) []*enhanced.EnhancedService {
	services := make([]*enhanced.EnhancedService, len(svc))

	for i, svc := range svc {
		enhancedService := m.AddService(svc, callback)
		services[i] = enhancedService
	}

	return services
}

// RemoveService removes a service from the manager and returns the enhanced service. The onServiceStop callback is optional and will be
// called when the service is stopped. The force parameter is used to force the stop (instead of execute a graceful shutdown)
func (m *Manager) RemoveService(svc *enhanced.EnhancedService, force bool, onServiceStop enhanced.ExitStepCallback) error {
	onServiceStop = applyDefaultExitProcessCallback(onServiceStop)

	// Stops the service
	exitInfo, err := svc.AbortPrepareOrStop(force)
	if err != nil {
		err = fmt.Errorf("error aborting prepare or stopping enhanced service '%s': %w", svc.GetName(), err)
		m.OnError(m, err)
		return err
	}
	onServiceStop(svc, exitInfo, err)

	// Removes the service from the list
	m.serviceListMutex.Lock()
	defer m.serviceListMutex.Unlock()

	for i, s := range m.services {
		if s == svc {
			m.services = append(m.services[:i], m.services[i+1:]...)
			break
		}
	}

	return nil
}

// RemoveServices removes multiple services from the manager. The onServiceStop callback is optional and will be called when the service is
// stopped. The force parameter is used to force the stop (instead of execute a graceful shutdown)
func (m *Manager) RemoveServices(services []*enhanced.EnhancedService, force bool, onServiceStop enhanced.ExitStepCallback) error {
	for _, svc := range services {
		err := m.RemoveService(svc, force, onServiceStop)
		if err != nil {
			return fmt.Errorf("error removing enhanced service '%s': %w", svc.GetName(), err)
		}
	}

	return nil
}

// GetServices returns the managed enhanced services. You should not modify the returned slice, only use its services. If you want to add or
// remove services, use the AddService and RemoveService methods
func (m *Manager) GetServices() []*enhanced.EnhancedService {
	return m.services
}

// routineForEachService executes a callback for each service in a goroutine and returns a waiter with the results of the callbacks
func (m *Manager) routineForEachService(callback func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error)) *waiter.Waiter {
	waiter := waiter.Waiter{}

	for _, svc := range m.services {
		waiter.Go(func() {
			exitCode, err := callback(svc)

			// Extends the error message
			if err != nil {
				err = fmt.Errorf("error processing enhanced service '%s': %w", svc.GetName(), err)
			} else if exitCode != nil {
				if exitCode.HasError() {
					err = fmt.Errorf("exit code of enhanced service '%s' has an error: %w", svc.GetName(), exitCode.WrapError())
				}
			}

			// Add the error to the waiter
			if err != nil {
				m.OnError(m, err)
				waiter.AddError(err)
			}
		})
	}

	return &waiter
}

// StartPrepare starts the preparing step for each service
func (m *Manager) StartPrepare(onServicePrepared enhanced.ExitStepCallback) *waiter.Waiter {
	onServicePrepared = applyDefaultExitProcessCallback(onServicePrepared)

	return m.routineForEachService(func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
		err := svc.StartPrepare()
		if err != nil {
			err = fmt.Errorf("error preparing enhanced service '%s': %w", svc.GetName(), err)
		}

		onServicePrepared(svc, nil, err)
		return nil, err
	})
}

// WaitPrepare waits the preparing step to finish for each service
func (m *Manager) WaitPrepare(onServicePrepared enhanced.ExitStepCallback) error {
	onServicePrepared = applyDefaultExitProcessCallback(onServicePrepared)

	return m.routineForEachService(func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
		exitInfo, err := svc.WaitPrepare()
		if err != nil {
			err = fmt.Errorf("error waiting enhanced service '%s' to prepare: %w", svc.GetName(), err)
		}

		onServicePrepared(svc, exitInfo, err)
		return exitInfo, err
	}).Wait()
}

// Prepare starts the preparing step for each service
func (m *Manager) Prepare(onServicePrepared enhanced.ExitStepCallback) *waiter.Waiter {
	onServicePrepared = applyDefaultExitProcessCallback(onServicePrepared)

	return m.routineForEachService(func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
		exitInfo, err := svc.Prepare()
		if err != nil {
			err = fmt.Errorf("error preparing enhanced service '%s': %w", svc.GetName(), err)
		}

		onServicePrepared(svc, exitInfo, err)
		return exitInfo, err
	})
}

// AbortPrepare aborts the preparing step for each service
func (m *Manager) AbortPrepare(force bool, onServicePrepared enhanced.ExitStepCallback) *waiter.Waiter {
	onServicePrepared = applyDefaultExitProcessCallback(onServicePrepared)

	return m.routineForEachService(func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
		exitInfo, err := svc.AbortPrepare(force)
		if err != nil {
			err = fmt.Errorf("error aborting enhanced service '%s': %w", svc.GetName(), err)
		}

		onServicePrepared(svc, exitInfo, err)
		return exitInfo, err
	})
}

// Start starts the main step for each service
func (m *Manager) Start(callback enhanced.ExitStepCallback) *waiter.Waiter {
	callback = applyDefaultExitProcessCallback(callback)

	return m.routineForEachService(func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
		err := svc.Start()
		if err != nil {
			err = fmt.Errorf("error starting enhanced service '%s': %w", svc.GetName(), err)
			callback(svc, nil, err)
			return nil, err
		}

		exitInfo, err := svc.Wait()
		if err != nil {
			err = fmt.Errorf("error waiting enhanced service '%s': %w", svc.GetName(), err)
			callback(svc, exitInfo, err)
			return nil, err
		}

		return exitInfo, nil
	})
}

// Wait waits the main step to finish for each service
func (m *Manager) Wait(onServiceEnded enhanced.ExitStepCallback) error {
	onServiceEnded = applyDefaultExitProcessCallback(onServiceEnded)

	err := m.routineForEachService(func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
		exitInfo, err := svc.Wait()
		if err != nil {
			err = fmt.Errorf("error waiting enhanced service '%s' to end: %w", svc.GetName(), err)
		}

		onServiceEnded(svc, exitInfo, err)
		return exitInfo, err
	}).Wait()

	if err != nil {
		return fmt.Errorf("error waiting for services to end: %w", err)
	}

	return nil
}

// Run runs all services together. The services starts only after all of them are prepared
func (m *Manager) Run(onServiceEnded enhanced.ExitStepCallback) error {
	err := m.Prepare(nil).Wait()
	if err != nil {
		return fmt.Errorf("error preparing services: %w", err)
	}

	err = m.Start(onServiceEnded).Wait()
	if err != nil {
		return fmt.Errorf("error starting services: %w", err)
	}

	return nil
}

// RunEach runs each service separately. The service starts after its own preparing step is finished (does not wait for other services)
func (m *Manager) RunEach(onServiceEnded enhanced.ExitStepCallback) *waiter.Waiter {
	onServiceEnded = applyDefaultExitProcessCallback(onServiceEnded)

	return m.routineForEachService(func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
		exitInfo, err := svc.Run()
		if err != nil {
			err = fmt.Errorf("error running enhanced service '%s': %w", svc.GetName(), err)
		}

		onServiceEnded(svc, exitInfo, err)
		return exitInfo, err
	})
}

// RunSerial runs each service separately (e.g.: service 2 is only started after service 1 has ended)
func (m *Manager) RunSerial(onServicePrepared enhanced.ExitStepCallback, onServiceEnded enhanced.ExitStepCallback) error {
	onServiceEnded = applyDefaultExitProcessCallback(onServiceEnded)

	err := m.Prepare(onServicePrepared).Wait()
	if err != nil {
		return fmt.Errorf("error preparing services: %w", err)
	}

	// Running the services
	for _, svc := range m.services {
		exitInfo, err := svc.Run()
		if err != nil {
			err = fmt.Errorf("error running enhanced service '%s': %w", svc.GetName(), err)
			onServiceEnded(svc, exitInfo, err)
			return err
		}
	}

	return nil
}

// Stop stops the main step for each service
func (m *Manager) Stop(force bool, onServiceEnded enhanced.ExitStepCallback) *waiter.Waiter {
	onServiceEnded = applyDefaultExitProcessCallback(onServiceEnded)

	return m.routineForEachService(func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
		exitInfo, err := svc.Stop(force)
		if err != nil {
			if errors.Is(err, enhanced.ErrInvalidStatus) {
				err = nil
			} else {
				err = fmt.Errorf("error stopping enhanced service '%s': %w", svc.GetName(), err)
			}
		}

		onServiceEnded(svc, exitInfo, err)
		return exitInfo, err
	})
}

// AbortPrepareOrStop aborts the preparing step (if preparing) or stops the running step (if running). The force parameter is used to force
// the abort (instead of execute a graceful shutdown)
func (m *Manager) AbortPrepareOrStop(force bool, onServiceEnded enhanced.ExitStepCallback) *waiter.Waiter {
	onServiceEnded = applyDefaultExitProcessCallback(onServiceEnded)

	return m.routineForEachService(func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
		exitInfo, err := svc.AbortPrepareOrStop(force)
		if err != nil {
			if errors.Is(err, enhanced.ErrInvalidStatus) {
				err = nil
			} else {
				err = fmt.Errorf("error aborting or stopping enhanced service '%s': %w", svc.GetName(), err)
			}
		}

		onServiceEnded(svc, exitInfo, err)
		return exitInfo, err
	})
}

// Reset resets the services to the initial state (None)
func (m *Manager) Reset(force bool, onServiceEnded enhanced.ExitStepCallback) *waiter.Waiter {
	onServiceEnded = applyDefaultExitProcessCallback(onServiceEnded)

	return m.routineForEachService(func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
		exitInfo, err := svc.Reset(force)
		if err != nil {
			err = fmt.Errorf("error resetting enhanced service '%s': %w", svc.GetName(), err)
		}

		onServiceEnded(svc, exitInfo, err)
		return exitInfo, err
	})
}

// Restart restarts all services together (e.g.: service 1 waits for service 2 to prepare before restarting)
func (m *Manager) Restart(force bool, onServiceEnded enhanced.ExitStepCallback) error {
	onServiceEnded = applyDefaultExitProcessCallback(onServiceEnded)

	err := m.Reset(force, onServiceEnded).Wait()
	if err != nil {
		return fmt.Errorf("error resetting services: %w", err)
	}

	err = m.Prepare(onServiceEnded).Wait()
	if err != nil {
		return fmt.Errorf("error preparing services: %w", err)
	}

	err = m.Start(onServiceEnded).Wait()
	if err != nil {
		return fmt.Errorf("error starting services: %w", err)
	}

	return nil
}

// RestartEach restarts each service separately. The service starts after its own preparing step is finished (does not wait for other
// services)
func (m *Manager) RestartEach(force bool, onServiceEnded enhanced.ExitStepCallback) *waiter.Waiter {
	onServiceEnded = applyDefaultExitProcessCallback(onServiceEnded)

	return m.routineForEachService(func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
		exitInfo, err := svc.Restart(force)
		if err != nil {
			err = fmt.Errorf("error resetting enhanced service '%s': %w", svc.GetName(), err)
		}

		onServiceEnded(svc, exitInfo, err)
		return exitInfo, err
	})
}
