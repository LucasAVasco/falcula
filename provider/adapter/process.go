package adapter

import (
	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/service/iface"
)

// ProcessStep wraps a process in a Step interface. It automatically starts the process if it is not already started
type ProcessStep struct {
	startErr error
	process  *process.Process
}

// ProcessToStep converts a process to a Step interface. It automatically starts the process if it is not already started
func ProcessToStep(process *process.Process) iface.Step {
	p := ProcessStep{
		process: process,
	}

	// Starts the process if it is not already started
	if !process.Started() {
		p.startErr = process.Start()
	}

	return &p
}

func (p *ProcessStep) Wait() (*iface.ExitInfo, error) {
	if p.startErr != nil {
		return &iface.ExitInfo{}, p.startErr
	}

	return p.process.Wait(), nil
}

func (p *ProcessStep) Abort(force bool) (*iface.ExitInfo, error) {
	return p.process.Stop(force)
}
