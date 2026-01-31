// Package enhanced enhances the service with status and functions to control it
package enhanced

import (
	"errors"
	"fmt"

	"github.com/LucasAVasco/falcula/service/empty"
	"github.com/LucasAVasco/falcula/service/iface"
	"github.com/LucasAVasco/falcula/service/status"
)

// The error returned when the user tries to execute a function but the service is in an status that does not allow it
var ErrInvalidStatus = errors.New("invalid status")

// exitInfoHasError returns true if the exitInfo has an error. If it is nil, it returns false
func exitInfoHasError(exitInfo *iface.ExitInfo) bool {
	if exitInfo == nil {
		return false
	} else {
		return exitInfo.HasError()
	}
}

// Enhances the service with status and functions to control it
type EnhancedService struct {
	svc       iface.Service
	status    status.Status
	callbacks *Callbacks
	step      iface.Step
}

// NewEnhancedService returns a new EnhancedService. The callbacks parameter is optional
func NewEnhancedService(svc iface.Service, callbacks *Callbacks) *EnhancedService {
	e := EnhancedService{
		svc:       svc,
		status:    status.None,
		callbacks: fillCallbacksWithDefaults(callbacks),
	}

	return &e
}

func (e *EnhancedService) GetName() string {
	return e.svc.GetName()
}

// GetService returns the original service
func (e *EnhancedService) GetService() iface.Service {
	return e.svc
}

func (e *EnhancedService) GetStatus() status.Status {
	return e.status
}

func (e *EnhancedService) setStatus(status status.Status) {
	e.status = status
	e.callbacks.OnServiceStatusChanged(e)
}

func (e *EnhancedService) setErrorStatus() {
	e.setStatus(status.Error)
}

// StartPrepare starts the preparing step
func (e *EnhancedService) StartPrepare() error {
	if e.status != status.None {
		return fmt.Errorf("service '%s' must not have a status associated with it when preparing it: %w", e.GetName(), ErrInvalidStatus)
	}

	// Starts the preparing step
	e.setStatus(status.Preparing)
	var err error
	e.step, err = e.svc.Prepare(func(exitInfo *iface.ExitInfo, err error) {
		if exitInfoHasError(exitInfo) || err != nil {
			e.setErrorStatus()
		} else if exitInfo.Stopped {
			e.setStatus(status.PrepareAborted)
		} else {
			e.setStatus(status.Ready)
		}

		e.callbacks.OnExitProcess(e, exitInfo, err)
	})
	if err != nil {
		return fmt.Errorf("error preparing service '%s': %w", e.GetName(), err)
	}

	// Fallback to an empty step
	if e.step == nil {
		e.step = empty.New()
	}

	return nil
}

// WaitPrepare waits the preparing step to finish and returns its exit information if any
func (e *EnhancedService) WaitPrepare() (*iface.ExitInfo, error) {
	if e.status != status.Preparing && e.status != status.Ready {
		return nil, fmt.Errorf(
			"service '%s' is not preparing or ready (current status: %s): %w",
			e.GetName(),
			e.status.ToString(),
			ErrInvalidStatus,
		)
	}

	exitInfo, err := e.step.Wait()
	if err != nil {
		return exitInfo, fmt.Errorf("error waiting prepare step of service '%s': %w", e.GetName(), err)
	}
	if exitInfoHasError(exitInfo) {
		return exitInfo, fmt.Errorf(
			"waiting prepare step of service '%s' returned an error: %w",
			e.GetName(),
			exitInfo.WrapError(),
		)
	}

	return exitInfo, nil
}

// Prepare starts the preparing step and waits it to finish. Returns its exit information if any
func (e *EnhancedService) Prepare() (*iface.ExitInfo, error) {
	err := e.StartPrepare()
	if err != nil {
		return nil, fmt.Errorf("error starting preparing service '%s': %w", e.GetName(), err)
	}

	exitInfo, err := e.WaitPrepare()
	if err != nil {
		return exitInfo, fmt.Errorf("error waiting preparing service '%s': %w", e.GetName(), err)
	}
	if exitInfoHasError(exitInfo) {
		return exitInfo, fmt.Errorf(
			"waiting prepare service '%s' returned an error: %w",
			e.GetName(),
			exitInfo.WrapError(),
		)
	}

	return exitInfo, err
}

// AbortPrepare aborts the preparing step and returns its exit information if any. The force parameter is used to force the abort (instead
// of execute a graceful shutdown)
//
// The abort operation may affect the returned exit information
func (e *EnhancedService) AbortPrepare(force bool) (*iface.ExitInfo, error) {
	if e.status.IsDoingNothing() {
		return nil, nil
	}

	if e.status != status.Preparing && e.status != status.AbortingPrepare {
		return nil, fmt.Errorf("service '%s' is not preparing: %w", e.GetName(), ErrInvalidStatus)
	}

	e.setStatus(status.AbortingPrepare)
	exitInfo, err := e.step.Abort(force)
	if err != nil {
		return exitInfo, fmt.Errorf("error aborting service '%s': %w", e.GetName(), err)
	}
	if exitInfoHasError(exitInfo) {
		return exitInfo, fmt.Errorf(
			"aborting prepare step of service '%s' returned an error: %w",
			e.GetName(),
			exitInfo.WrapError(),
		)
	}

	return exitInfo, nil
}

// Start starts the main step. Must be called after the service be prepared
func (e *EnhancedService) Start() error {
	if e.status != status.Ready {
		return fmt.Errorf("service '%s' is not ready to start: %w", e.GetName(), ErrInvalidStatus)
	}

	// Starts the main step
	e.setStatus(status.Running)
	var err error
	e.step, err = e.svc.Start(func(exitInfo *iface.ExitInfo, err error) {
		if exitInfoHasError(exitInfo) || err != nil {
			e.setErrorStatus()
		} else if exitInfo.Stopped {
			e.setStatus(status.Stopped)
		} else {
			e.setStatus(status.Ended)
		}

		e.callbacks.OnExitProcess(e, exitInfo, err)
	})
	if err != nil {
		return fmt.Errorf("error starting service '%s': %w", e.GetName(), err)
	}

	// Fallback to an empty step
	if e.step == nil {
		e.step = empty.New()
	}

	return nil
}

// Wait waits the main step to finish and returns its exit information if any
func (e *EnhancedService) Wait() (*iface.ExitInfo, error) {
	if e.status != status.Running && e.status != status.Ended {
		return nil, fmt.Errorf(
			"service '%s' is not running or ended (current status: %s): %w",
			e.GetName(),
			e.status.ToString(),
			ErrInvalidStatus,
		)
	}

	return e.step.Wait()
}

// Run starts the main step and waits it to finish. Returns its exit information if any
func (e *EnhancedService) Run() (*iface.ExitInfo, error) {
	err := e.Start()
	if err != nil {
		return nil, fmt.Errorf("error starting service '%s': %w", e.GetName(), err)
	}

	exitInfo, err := e.Wait()
	if err != nil {
		return exitInfo, fmt.Errorf("error running service '%s': %w", e.GetName(), err)
	}
	if exitInfoHasError(exitInfo) {
		return exitInfo, fmt.Errorf(
			"waiting service '%s' returned an error: %w",
			e.GetName(),
			exitInfo.WrapError(),
		)
	}

	return exitInfo, nil
}

// Stop stops the main step and returns its exit information if any. The force parameter is used to force the abort (instead of execute a
// graceful shutdown)
//
// The stop operation may affect the returned exit information
func (e *EnhancedService) Stop(force bool) (*iface.ExitInfo, error) {
	if e.status.IsDoingNothing() {
		return nil, nil
	}

	if e.status != status.Running && e.status != status.Stopping {
		return nil, fmt.Errorf("service '%s' is not running: %w", e.GetName(), ErrInvalidStatus)
	}

	// Stops the waiter
	e.setStatus(status.Stopping)
	exitInfo, err := e.step.Abort(force)
	if err != nil {
		return exitInfo, fmt.Errorf("error stopping service '%s': %w", e.GetName(), err)
	}
	if exitInfoHasError(exitInfo) {
		return exitInfo, fmt.Errorf(
			"aborting running step of service '%s' returned an error: %w",
			e.GetName(),
			exitInfo.WrapError(),
		)
	}

	// Next status
	e.setStatus(status.Stopped)
	return exitInfo, nil
}

// AbortPrepareOrStop aborts the preparing step (if preparing) or stops the running step (if running).The force parameter is used to force
// the abort (instead of execute a graceful shutdown)
func (e *EnhancedService) AbortPrepareOrStop(force bool) (*iface.ExitInfo, error) {
	if e.status.IsDoingNothing() {
		return nil, nil
	}

	var exitInfo *iface.ExitInfo
	var err error

	switch e.status {
	case status.Preparing, status.AbortingPrepare:
		exitInfo, err = e.AbortPrepare(force)
		if err != nil {
			return exitInfo, fmt.Errorf("error aborting 'preparing' step '%s': %w", e.GetName(), err)
		}
		if exitInfoHasError(exitInfo) {
			return exitInfo, fmt.Errorf(
				"aborting preparing service '%s' returned an error: %w",
				e.GetName(),
				exitInfo.WrapError(),
			)
		}

	case status.Running, status.Stopping:
		exitInfo, err = e.Stop(force)
		if err != nil {
			return exitInfo, fmt.Errorf("error stopping 'running' step '%s': %w", e.GetName(), err)
		}
		if exitInfoHasError(exitInfo) {
			return exitInfo, fmt.Errorf(
				"stopping running service '%s' returned an error: %w",
				e.GetName(),
				exitInfo.WrapError(),
			)
		}

	default:
		return nil, fmt.Errorf("service '%s' is not preparing or running: %w", e.GetName(), ErrInvalidStatus)
	}

	return exitInfo, nil
}

// Reset resets the service to the initial state (None) and returns a exit information related to the process if any. The force parameter is
// used to force the abort (instead of execute a graceful shutdown)
func (e *EnhancedService) Reset(force bool) (*iface.ExitInfo, error) {
	// Does not need to reset if the service never started
	if e.status == status.None {
		return nil, nil
	}

	if e.status == status.Error {
		e.setStatus(status.None)
		return nil, nil
	}

	// Aborts the service
	exitInfo, err := e.AbortPrepareOrStop(force)
	if err != nil {
		return exitInfo, fmt.Errorf("error aborting 'preparing' step '%s': %w", e.GetName(), err)
	}
	if exitInfoHasError(exitInfo) {
		return exitInfo, fmt.Errorf(
			"aborting preparing service '%s' returned an error: %w",
			e.GetName(),
			exitInfo.WrapError(),
		)
	}

	// Resets the status
	e.setStatus(status.None)

	return exitInfo, nil
}

// Restart restarts the service. The force parameter is used to force the abort (instead of execute a graceful shutdown)
func (e *EnhancedService) Restart(force bool) (*iface.ExitInfo, error) {
	// Resets the service
	exitInfo, err := e.Reset(force)
	if err != nil {
		return exitInfo, fmt.Errorf("error resetting service '%s': %w", e.GetName(), err)
	}
	if exitInfoHasError(exitInfo) {
		return exitInfo, fmt.Errorf("exit information of resetting service '%s' has error: %w", e.GetName(), exitInfo.Error)
	}

	// Re-prepares the service
	exitInfo, err = e.Prepare()
	if err != nil {
		return exitInfo, fmt.Errorf("error preparing service '%s': %w", e.GetName(), err)
	}

	if exitInfoHasError(exitInfo) {
		return exitInfo, fmt.Errorf("exit information of preparing service '%s' has error: %w", e.GetName(), exitInfo.Error)
	}

	// Restarts the service
	err = e.Start()
	if err != nil {
		return nil, fmt.Errorf("error starting service '%s': %w", e.GetName(), err)
	}

	return nil, err
}
