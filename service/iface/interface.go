// Package iface implements the interface of a service
package iface

import (
	"github.com/LucasAVasco/falcula/process"
)

type ExitInfo = process.ExitInfo

// Called when the Step is done
type OnExitCallback = func(info *ExitInfo, err error)

// Step represents a step of a service execution
type Step interface {
	Wait() (*ExitInfo, error) // Wait for the step to end and return the exit information

	// Abort the step. Blocking operation. The force parameter will force the step to be aborted instead of execute a graceful shutdown. The
	// exit information may be affected by the stop operation.
	Abort(force bool) (*ExitInfo, error)
}

// Service represents a service managed by this application. All services must implement this interface
// NOTE(LucasAVasco): this application considers that the `Prepare` and `Start` are fast operations, they must delegate all long running
// operations to the returned `Step`
type Service interface {
	GetName() string
	Prepare(callback OnExitCallback) (Step, error) // Generates a Step that will prepare the service to run.
	Start(callback OnExitCallback) (Step, error)   // Generates a Step that will start the service
}
