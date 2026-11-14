// Package enhanced enhances the service with status and functions to control it
package enhanced

import (
	"errors"
	"fmt"
	"sync"

	"github.com/LucasAVasco/falcula/service/empty"
	"github.com/LucasAVasco/falcula/service/iface"
	"github.com/LucasAVasco/falcula/service/status"
)

// The error returned when the user tries to execute a function but the service is in an status that does not allow it
var ErrInvalidStatus = errors.New("invalid status")

// The error returned when the user tries to wait or abort a step that is still being created
var ErrStepNotAvailable = errors.New("step not available")

// exitInfoHasError returns true if the exitInfo has an error. If it is nil, it returns false
func exitInfoHasError(exitInfo *iface.ExitInfo) bool {
	if exitInfo == nil {
		return false
	} else {
		return exitInfo.HasError()
	}
}

// Enhances the service with status and functions to control it
//
// The status and the current step are safe for concurrent access
type EnhancedService struct {
	svc       iface.Service
	callbacks *Callbacks

	mutex     sync.Mutex // Protects the fields below
	status    status.Status
	step      iface.Step
	stepPhase stepPhase // Phase of the current step
}

// stepPhase is the phase of the service that a step belongs to
type stepPhase int

const (
	prepareStep stepPhase = iota + 1 // Created by `Prepare` of the service
	mainStep                         // Created by `Start` of the service
)

// NewEnhancedService returns a new EnhancedService. The callbacks parameter is optional
func NewEnhancedService(svc iface.Service, callbacks *Callbacks) *EnhancedService {
	e := EnhancedService{
		svc:       svc,
		status:    status.None,
		callbacks: fillCallbacksWithDefaults(callbacks),
	}

	if e.svc.GetOpts().StartDisabled {
		e.setStatus(status.Disabled)
	}

	return &e
}

func (e *EnhancedService) GetName() string {
	return e.svc.GetName()
}

func (e *EnhancedService) GetStatus() status.Status {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.status
}

// Disable disables the service. It can not be prepared or started until it is enabled again
func (e *EnhancedService) Disable() error {
	if !e.GetStatus().IsDoingNothing() {
		return fmt.Errorf(
			"the service '%s' must not do anything when disabling it: %w",
			e.GetName(),
			ErrInvalidStatus,
		)
	}

	e.setStatus(status.Disabled)

	return nil
}

// Enable enables the service. It can be prepared or started again
func (e *EnhancedService) Enable() error {
	if e.GetStatus() != status.Disabled {
		return fmt.Errorf(
			"the service '%s' must be disabled when enabling it: %w",
			e.GetName(),
			ErrInvalidStatus,
		)
	}

	e.setStatus(status.None)
	return nil
}

func (e *EnhancedService) setStatus(status status.Status) {
	e.mutex.Lock()
	e.status = status
	e.mutex.Unlock()

	e.callbacks.OnServiceStatusChanged(e)
}

// getStep returns the current step if it belongs to the phase. Returns an error if the step is not available (the service is creating it
// in another goroutine, or the current step belongs to another phase)
func (e *EnhancedService) getStep(phase stepPhase) (iface.Step, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.step == nil || e.stepPhase != phase {
		return nil, fmt.Errorf("the step of service '%s' is not available: %w", e.svc.GetName(), ErrStepNotAvailable)
	}
	return e.step, nil
}

// clearStep removes the current step
func (e *EnhancedService) clearStep() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.step = nil
}

// setStep sets the current step and its phase. Uses an empty step if it is nil
func (e *EnhancedService) setStep(step iface.Step, phase stepPhase) {
	if step == nil {
		step = empty.New()
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.step = step
	e.stepPhase = phase
}

// onStepExit returns the callback called when a step ends. It updates the status of the service according to the exit information: Error
// if the step failed, `abortedStatus` if the user stopped it and `successStatus` otherwise
func (e *EnhancedService) onStepExit(successStatus, abortedStatus status.Status) iface.OnExitCallback {
	return func(exitInfo *iface.ExitInfo, err error) {
		if exitInfoHasError(exitInfo) || err != nil {
			e.setStatus(status.Error)
		} else if exitInfo.Stopped {
			e.setStatus(abortedStatus)
		} else {
			e.setStatus(successStatus)
		}

		e.callbacks.OnExitProcess(e, exitInfo, err)
	}
}

// checkStepResult returns an error if the step operation (described by `action`) returned an error or if its exit information has an error
func (e *EnhancedService) checkStepResult(action string, exitInfo *iface.ExitInfo, err error) error {
	if err != nil {
		return fmt.Errorf("error %s of service '%s': %w", action, e.GetName(), err)
	}

	if exitInfoHasError(exitInfo) {
		return fmt.Errorf("%s of service '%s' returned an error: %w", action, e.GetName(), exitInfo.WrapError())
	}

	return nil
}

// StartPrepare starts the preparing step
func (e *EnhancedService) StartPrepare() error {
	if e.GetStatus() == status.Disabled {
		return nil
	}

	if e.GetStatus() != status.None {
		return fmt.Errorf("service '%s' must not have a status associated with it when preparing it: %w", e.GetName(), ErrInvalidStatus)
	}

	// Starts the preparing step
	e.clearStep() // The previous step is finished. The new one is set after it is created
	e.setStatus(status.Preparing)
	step, err := e.svc.Prepare(e.onStepExit(status.Ready, status.PrepareAborted))
	if err != nil {
		e.setStatus(status.Error)
		return fmt.Errorf("error preparing service '%s': %w", e.GetName(), err)
	}
	e.setStep(step, prepareStep)

	return nil
}

// WaitPrepare waits the preparing step to finish and returns its exit information if any
func (e *EnhancedService) WaitPrepare() (*iface.ExitInfo, error) {
	if e.GetStatus() == status.Disabled {
		return &iface.ExitInfo{}, nil
	}

	// NOTE(LucasAVasco): the step may have already ended (and changed the status) before this method is called
	switch e.GetStatus() {
	case status.Preparing, status.Ready, status.AbortingPrepare, status.PrepareAborted, status.Error:
	default:
		return nil, fmt.Errorf(
			"service '%s' is not preparing or prepared (current status: %s): %w",
			e.GetName(),
			e.GetStatus(),
			ErrInvalidStatus,
		)
	}

	step, err := e.getStep(prepareStep)
	if err != nil {
		return nil, err
	}

	exitInfo, err := step.Wait()
	return exitInfo, e.checkStepResult("waiting prepare step", exitInfo, err)
}

// Prepare starts the preparing step and waits it to finish. Returns its exit information if any
func (e *EnhancedService) Prepare() (*iface.ExitInfo, error) {
	if e.GetStatus() == status.Disabled {
		return &iface.ExitInfo{}, nil
	}

	err := e.StartPrepare()
	if err != nil {
		return nil, fmt.Errorf("error starting preparing service '%s': %w", e.GetName(), err)
	}

	exitInfo, err := e.WaitPrepare()
	if err != nil {
		return exitInfo, fmt.Errorf("error waiting preparing service '%s': %w", e.GetName(), err)
	}

	return exitInfo, nil
}

// AbortPrepare aborts the preparing step and returns its exit information if any. The force parameter is used to force the abort (instead
// of execute a graceful shutdown)
//
// The abort operation may affect the returned exit information
func (e *EnhancedService) AbortPrepare(force bool) (*iface.ExitInfo, error) {
	if e.GetStatus() == status.Disabled {
		return &iface.ExitInfo{}, nil
	}

	if e.GetStatus().IsDoingNothing() {
		return nil, nil
	}

	if e.GetStatus() != status.Preparing && e.GetStatus() != status.AbortingPrepare {
		return nil, fmt.Errorf("service '%s' is not preparing: %w", e.GetName(), ErrInvalidStatus)
	}

	e.setStatus(status.AbortingPrepare)
	step, err := e.getStep(prepareStep)
	if err != nil {
		return nil, err
	}

	exitInfo, err := step.Abort(force)
	return exitInfo, e.checkStepResult("aborting prepare step", exitInfo, err)
}

// Start starts the main step. Must be called after the service be prepared
func (e *EnhancedService) Start() error {
	if e.GetStatus() == status.Disabled {
		return nil
	}

	if e.GetStatus() != status.Ready {
		return fmt.Errorf("service '%s' is not ready to start: %w", e.GetName(), ErrInvalidStatus)
	}

	// Starts the main step
	e.clearStep() // The previous step is finished. The new one is set after it is created
	e.setStatus(status.Running)
	step, err := e.svc.Start(e.onStepExit(status.Ended, status.Stopped))
	if err != nil {
		e.setStatus(status.Error)
		return fmt.Errorf("error starting service '%s': %w", e.GetName(), err)
	}
	e.setStep(step, mainStep)

	return nil
}

// Wait waits the main step to finish and returns its exit information if any
func (e *EnhancedService) Wait() (*iface.ExitInfo, error) {
	if e.GetStatus() == status.Disabled {
		return &iface.ExitInfo{}, nil
	}

	// NOTE(LucasAVasco): the step may have already ended (and changed the status) before this method is called
	switch e.GetStatus() {
	case status.Running, status.Stopping, status.Ended, status.Stopped, status.Error:
	default:
		return nil, fmt.Errorf(
			"service '%s' is not running or ended (current status: %s): %w",
			e.GetName(),
			e.GetStatus(),
			ErrInvalidStatus,
		)
	}

	step, err := e.getStep(mainStep)
	if err != nil {
		return nil, err
	}

	return step.Wait()
}

// Run starts the main step and waits it to finish. Returns its exit information if any
func (e *EnhancedService) Run() (*iface.ExitInfo, error) {
	if e.GetStatus() == status.Disabled {
		return &iface.ExitInfo{}, nil
	}

	err := e.Start()
	if err != nil {
		return nil, fmt.Errorf("error starting service '%s': %w", e.GetName(), err)
	}

	exitInfo, err := e.Wait()
	return exitInfo, e.checkStepResult("waiting running step", exitInfo, err)
}

// Stop stops the main step and returns its exit information if any. The force parameter is used to force the abort (instead of execute a
// graceful shutdown)
//
// The stop operation may affect the returned exit information
func (e *EnhancedService) Stop(force bool) (*iface.ExitInfo, error) {
	if e.GetStatus() == status.Disabled {
		return &iface.ExitInfo{}, nil
	}

	if e.GetStatus().IsDoingNothing() {
		return nil, nil
	}

	if e.GetStatus() != status.Running && e.GetStatus() != status.Stopping {
		return nil, fmt.Errorf("service '%s' is not running: %w", e.GetName(), ErrInvalidStatus)
	}

	// Stops the waiter
	e.setStatus(status.Stopping)
	step, err := e.getStep(mainStep)
	if err != nil {
		return nil, err
	}

	exitInfo, err := step.Abort(force)
	if err := e.checkStepResult("aborting running step", exitInfo, err); err != nil {
		return exitInfo, err
	}

	// Next status
	e.setStatus(status.Stopped)
	return exitInfo, nil
}

// AbortPrepareOrStop aborts the preparing step (if preparing) or stops the running step (if running).The force parameter is used to force
// the abort (instead of execute a graceful shutdown)
func (e *EnhancedService) AbortPrepareOrStop(force bool) (*iface.ExitInfo, error) {
	if e.GetStatus() == status.Disabled {
		return &iface.ExitInfo{}, nil
	}

	if e.GetStatus().IsDoingNothing() {
		return nil, nil
	}

	switch e.GetStatus() {
	case status.Preparing, status.AbortingPrepare:
		exitInfo, err := e.AbortPrepare(force)
		if err != nil {
			return exitInfo, fmt.Errorf("error aborting 'preparing' step '%s': %w", e.GetName(), err)
		}
		return exitInfo, nil

	case status.Running, status.Stopping:
		exitInfo, err := e.Stop(force)
		if err != nil {
			return exitInfo, fmt.Errorf("error stopping 'running' step '%s': %w", e.GetName(), err)
		}
		return exitInfo, nil

	default:
		return nil, fmt.Errorf("service '%s' is not preparing or running: %w", e.GetName(), ErrInvalidStatus)
	}
}

// Reset resets the service to the initial state (None) and returns a exit information related to the process if any. The force parameter is
// used to force the abort (instead of execute a graceful shutdown)
func (e *EnhancedService) Reset(force bool) (*iface.ExitInfo, error) {
	if e.GetStatus() == status.Disabled {
		return &iface.ExitInfo{}, nil
	}

	// Does not need to reset if the service never started
	if e.GetStatus() == status.None {
		return nil, nil
	}

	if e.GetStatus() == status.Error {
		e.setStatus(status.None)
		return nil, nil
	}

	// Aborts the service
	exitInfo, err := e.AbortPrepareOrStop(force)
	if err != nil {
		return exitInfo, fmt.Errorf("error aborting or stopping service '%s': %w", e.GetName(), err)
	}

	// Resets the status
	e.setStatus(status.None)

	return exitInfo, nil
}

// Restart restarts the service. The force parameter is used to force the abort (instead of execute a graceful shutdown)
func (e *EnhancedService) Restart(force bool) (*iface.ExitInfo, error) {
	if e.GetStatus() == status.Disabled {
		return &iface.ExitInfo{}, nil
	}
	// Resets the service
	exitInfo, err := e.Reset(force)
	if err != nil {
		return exitInfo, fmt.Errorf("error resetting service '%s': %w", e.GetName(), err)
	}

	// Re-prepares the service
	exitInfo, err = e.Prepare()
	if err != nil {
		return exitInfo, fmt.Errorf("error preparing service '%s': %w", e.GetName(), err)
	}

	// Restarts the service
	err = e.Start()
	if err != nil {
		return nil, fmt.Errorf("error starting service '%s': %w", e.GetName(), err)
	}

	return nil, nil
}
