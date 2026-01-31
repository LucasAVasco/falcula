package adapter

import (
	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/service/iface"
)

// ProcessStep wraps a process in a Step interface
type ProcessStep struct {
	process *process.Process
}

// ProcessToStep converts a process to a Step interface
func ProcessToStep(process *process.Process) iface.Step {
	return &ProcessStep{
		process: process,
	}
}

func (p *ProcessStep) Wait() (*iface.ExitInfo, error) {
	return p.process.Wait(), nil
}

func (p *ProcessStep) Abort(force bool) (*iface.ExitInfo, error) {
	return p.process.Stop(force)
}
