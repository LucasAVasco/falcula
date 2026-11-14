package process

import (
	"fmt"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/provider/adapter"
	"github.com/LucasAVasco/falcula/provider/base"
	"github.com/LucasAVasco/falcula/service/iface"
)

// ServiceOpts is the base options for a process service
type ServiceOpts = base.ServiceOpts

// Service represents a process service
type Service struct {
	*base.Service
	prepareCmd *Command
	mainCmd    *Command
}

func (s *Service) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	step, err := s.runCommand(s.prepareCmd, callback)
	if err != nil {
		return nil, fmt.Errorf("error running 'Prepare' command: %w", err)
	}
	return step, nil
}

func (s *Service) Start(callback iface.OnExitCallback) (iface.Step, error) {
	step, err := s.runCommand(s.mainCmd, callback)
	if err != nil {
		return nil, fmt.Errorf("error starting main command: %w", err)
	}
	return step, nil
}

// runCommand starts the command and returns a step that waits for it. If the command is nil, calls the callback immediately and returns
// no step
func (s *Service) runCommand(command *Command, callback iface.OnExitCallback) (iface.Step, error) {
	if command == nil {
		callback(&iface.ExitInfo{}, nil)
		return nil, nil
	}

	procOpts := s.NewProcessOptions()
	procOpts.Dir = command.Dir
	procOpts.Shell = command.Shell
	procOpts.OnExit = func(info *process.ExitInfo) { callback(info, nil) }

	proc, err := process.New(procOpts, command.Command[0], command.Command[1:]...)
	if err != nil {
		return nil, err
	}

	return adapter.ProcessToStep(proc), nil
}
