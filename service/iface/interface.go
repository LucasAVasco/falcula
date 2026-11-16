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

// Opts is the service options
type Opts struct {
	StartDisabled bool `lua:"start_disabled"` // The service will be disabled by default (it must be enabled before it can be used)
}

// FileOwner is an optional interface of a service that has an associated file (e.g. the compose file of a docker-compose service). The
// user can open it from the text user interface
type FileOwner interface {
	GetFile() string // Absolute path of the file. Empty if the service has no associated file
}

// Service represents a service managed by this application. All services must implement this interface
// NOTE(LucasAVasco): this application considers that the `Prepare` and `Start` are fast operations, they must delegate all long running
// operations to the returned `Step`
type Service interface {
	GetName() string
	GetOpts() *Opts                                // Gets the service options. Can not be `nil`. Can not be changed
	Prepare(callback OnExitCallback) (Step, error) // Generates a Step that will prepare the service to run.
	Start(callback OnExitCallback) (Step, error)   // Generates a Step that will start the service
}
