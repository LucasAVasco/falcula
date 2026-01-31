// Package empty exports an empty Step (does nothing)
package empty

import (
	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/service/iface"
)

// EmptyStep implements the Step interface, but does nothing. Returns no error
type EmptyStep struct {
}

func New() iface.Step {
	return &EmptyStep{}
}

func (s *EmptyStep) Wait() (*iface.ExitInfo, error) {
	return &iface.ExitInfo{}, nil
}

func (s *EmptyStep) Abort(force bool) (*process.ExitInfo, error) {
	return &process.ExitInfo{}, nil
}
