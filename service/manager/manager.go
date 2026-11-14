// Package manager implements a service manager
package manager

import (
	"errors"
	"fmt"
	"slices"
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

	m.services = slices.DeleteFunc(m.services, func(s *enhanced.EnhancedService) bool {
		return s == svc
	})

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

// GetServices returns a copy of the list of managed enhanced services. If you want to add or remove services, use the AddService and
// RemoveService methods
func (m *Manager) GetServices() []*enhanced.EnhancedService {
	m.serviceListMutex.Lock()
	defer m.serviceListMutex.Unlock()

	return slices.Clone(m.services)
}

// routineForEachService executes a callback for each service in a goroutine and returns a waiter with the results of the callbacks
func (m *Manager) routineForEachService(callback func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error)) *waiter.Waiter {
	waiter := waiter.Waiter{}

	for _, svc := range m.GetServices() {
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

// forEachService executes an operation on each service in its own goroutine and returns a waiter with the results. Errors are wrapped
// with `errFormat` (receives the service name and the error). The callback (optional) receives the result of each operation. If
// `ignoreInvalidStatus` is true, `enhanced.ErrInvalidStatus` errors are discarded (the service was not in a state that requires the
// operation)
func (m *Manager) forEachService(
	errFormat string,
	callback enhanced.ExitStepCallback,
	ignoreInvalidStatus bool,
	operation func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error),
) *waiter.Waiter {
	callback = applyDefaultExitProcessCallback(callback)

	return m.routineForEachService(func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
		exitInfo, err := operation(svc)
		if ignoreInvalidStatus && errors.Is(err, enhanced.ErrInvalidStatus) {
			err = nil
		} else if err != nil {
			err = fmt.Errorf(errFormat, svc.GetName(), err)
		}

		callback(svc, exitInfo, err)
		return exitInfo, err
	})
}

// StartPrepare starts the preparing step for each service
func (m *Manager) StartPrepare(onServicePrepared enhanced.ExitStepCallback) *waiter.Waiter {
	return m.forEachService("error preparing enhanced service '%s': %w", onServicePrepared, false,
		func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
			return nil, svc.StartPrepare()
		})
}

// WaitPrepare waits the preparing step to finish for each service
func (m *Manager) WaitPrepare(onServicePrepared enhanced.ExitStepCallback) error {
	return m.forEachService("error waiting enhanced service '%s' to prepare: %w", onServicePrepared, false,
		(*enhanced.EnhancedService).WaitPrepare).Wait()
}

// Prepare starts the preparing step for each service
func (m *Manager) Prepare(onServicePrepared enhanced.ExitStepCallback) *waiter.Waiter {
	return m.forEachService("error preparing enhanced service '%s': %w", onServicePrepared, false,
		(*enhanced.EnhancedService).Prepare)
}

// AbortPrepare aborts the preparing step for each service
func (m *Manager) AbortPrepare(force bool, onServicePrepared enhanced.ExitStepCallback) *waiter.Waiter {
	return m.forEachService("error aborting enhanced service '%s': %w", onServicePrepared, false,
		func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
			return svc.AbortPrepare(force)
		})
}

// Start starts the main step for each service and waits for it to end. The callback is called when each service ends
func (m *Manager) Start(onServiceEnded enhanced.ExitStepCallback) *waiter.Waiter {
	return m.forEachService("error running enhanced service '%s': %w", onServiceEnded, false,
		func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
			err := svc.Start()
			if err != nil {
				return nil, fmt.Errorf("error starting: %w", err)
			}

			exitInfo, err := svc.Wait()
			if err != nil {
				return exitInfo, fmt.Errorf("error waiting: %w", err)
			}

			return exitInfo, nil
		})
}

// Wait waits the main step to finish for each service
func (m *Manager) Wait(onServiceEnded enhanced.ExitStepCallback) error {
	err := m.forEachService("error waiting enhanced service '%s' to end: %w", onServiceEnded, false,
		(*enhanced.EnhancedService).Wait).Wait()
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

// RunSerial runs each service separately (e.g.: service 2 is only started after service 1 has ended). The `onServiceEnded` callback is
// called when each service ends. Stops at the first service that fails
func (m *Manager) RunSerial(onServicePrepared enhanced.ExitStepCallback, onServiceEnded enhanced.ExitStepCallback) error {
	onServiceEnded = applyDefaultExitProcessCallback(onServiceEnded)

	err := m.Prepare(onServicePrepared).Wait()
	if err != nil {
		return fmt.Errorf("error preparing services: %w", err)
	}

	// Running the services
	for _, svc := range m.GetServices() {
		exitInfo, err := svc.Run()
		if err != nil {
			err = fmt.Errorf("error running enhanced service '%s': %w", svc.GetName(), err)
		}

		onServiceEnded(svc, exitInfo, err)
		if err != nil {
			return err
		}
	}

	return nil
}

// Stop stops the main step for each service. Services that are not running are ignored
func (m *Manager) Stop(force bool, onServiceEnded enhanced.ExitStepCallback) *waiter.Waiter {
	return m.forEachService("error stopping enhanced service '%s': %w", onServiceEnded, true,
		func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
			return svc.Stop(force)
		})
}

// AbortPrepareOrStop aborts the preparing step (if preparing) or stops the running step (if running). The force parameter is used to force
// the abort (instead of execute a graceful shutdown). Services that are neither preparing nor running are ignored
func (m *Manager) AbortPrepareOrStop(force bool, onServiceEnded enhanced.ExitStepCallback) *waiter.Waiter {
	return m.forEachService("error aborting or stopping enhanced service '%s': %w", onServiceEnded, true,
		func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
			return svc.AbortPrepareOrStop(force)
		})
}

// Reset resets the services to the initial state (None)
func (m *Manager) Reset(force bool, onServiceEnded enhanced.ExitStepCallback) *waiter.Waiter {
	return m.forEachService("error resetting enhanced service '%s': %w", onServiceEnded, false,
		func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
			return svc.Reset(force)
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

// Disable disables all services
func (m *Manager) Disable(onServiceEnded enhanced.ExitStepCallback) *waiter.Waiter {
	return m.forEachService("error disabling enhanced service '%s': %w", onServiceEnded, false,
		func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
			return nil, svc.Disable()
		})
}

// Enable enables all services
func (m *Manager) Enable(onServiceEnded enhanced.ExitStepCallback) *waiter.Waiter {
	return m.forEachService("error enabling enhanced service '%s': %w", onServiceEnded, false,
		func(svc *enhanced.EnhancedService) (*iface.ExitInfo, error) {
			return nil, svc.Enable()
		})
}
