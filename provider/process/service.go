package process

import (
	"fmt"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/provider/adapter"
	"github.com/LucasAVasco/falcula/provider/base"
	"github.com/LucasAVasco/falcula/service/iface"
)

// Service represents a process service
type Service struct {
	*base.Service
	prepareCmd *Command
	mainCmd    *Command
}

func (s *Service) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	if s.prepareCmd == nil {
		callback(&iface.ExitInfo{}, nil)
		return nil, nil
	}

	procOpts := s.NewProcessOptions()
	procOpts.Shell = s.prepareCmd.Shell
	procOpts.OnExit = func(info *process.ExitInfo) { callback(info, nil) }

	proc, err := process.New(procOpts, s.prepareCmd.Command[0], s.prepareCmd.Command[1:]...)
	if err != nil {
		return nil, fmt.Errorf("error running 'Prepare' command: %w", err)
	}

	return adapter.ProcessToStep(proc), nil
}

func (s *Service) Start(callback iface.OnExitCallback) (iface.Step, error) {
	if s.mainCmd == nil {
		callback(&iface.ExitInfo{}, nil)
		return nil, nil
	}

	procOpts := s.NewProcessOptions()
	procOpts.Shell = s.mainCmd.Shell
	procOpts.OnExit = func(info *process.ExitInfo) { callback(info, nil) }

	// Starts the process
	proc, err := process.New(procOpts, s.mainCmd.Command[0], s.mainCmd.Command[1:]...)
	if err != nil {
		return nil, fmt.Errorf("error starting 'Up' command: %w", err)
	}

	return adapter.ProcessToStep(proc), err
}
