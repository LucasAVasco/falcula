package process

import (
	"fmt"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/provider/adapter"
	"github.com/LucasAVasco/falcula/service/iface"
)

// Service represents a process service
type Service struct {
	provider   *Provider
	name       string
	prepareCmd *Command
	mainCmd    *Command
}

func (s *Service) GetName() string {
	return s.name
}

func (s *Service) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	if s.prepareCmd == nil {
		return nil, nil
	}

	procOpts := process.Options{
		Shell:       s.prepareCmd.Shell,
		Multiplexer: s.provider.Multiplexer,
		Name:        s.name,
		Wait:        false,
		Color:       s.provider.Color,
		OnExit:      func(info *process.ExitInfo) { callback(info, nil) },
	}

	proc, err := process.New(&procOpts, s.prepareCmd.Command[0], s.prepareCmd.Command[1:]...)
	if err != nil {
		return nil, fmt.Errorf("error running 'Prepare' command: %w", err)
	}

	return adapter.ProcessToStep(proc), nil
}

func (s *Service) Start(callback iface.OnExitCallback) (iface.Step, error) {
	if s.mainCmd == nil {
		return nil, nil
	}

	procOpts := process.Options{
		Shell:       s.mainCmd.Shell,
		Multiplexer: s.provider.Multiplexer,
		Name:        s.name,
		Wait:        false,
		Color:       s.provider.Color,
		OnExit:      func(info *process.ExitInfo) { callback(info, nil) },
	}

	// Starts the process
	proc, err := process.New(&procOpts, s.mainCmd.Command[0], s.mainCmd.Command[1:]...)
	if err != nil {
		return nil, fmt.Errorf("error starting 'Up' command: %w", err)
	}

	return adapter.ProcessToStep(proc), err
}
