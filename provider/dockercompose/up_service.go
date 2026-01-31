package dockercompose

import (
	"fmt"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/provider/adapter"
	"github.com/LucasAVasco/falcula/provider/dockercompose/cmd"
	"github.com/LucasAVasco/falcula/service/iface"
)

// UpService is a service that runs the docker-compose up command
type UpService struct {
	provider *Provider
	name     string
}

func (s *UpService) GetName() string {
	return s.name
}

func (s *UpService) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	procOpts := process.Options{
		Multiplexer: s.provider.Multiplexer,
		Name:        s.name,
		Wait:        false,
		ManualStart: true,
		Color:       s.provider.Color,
	}

	// Pull and build processes
	procs := make([]*process.Process, 0, 2)

	proc, err := cmd.Pull(&procOpts, s.provider.composeFile)
	if err != nil {
		return nil, fmt.Errorf("error running 'Pull' command: %w", err)
	}
	procs = append(procs, proc)

	proc, err = cmd.Build(&procOpts, s.provider.composeFile)
	if err != nil {
		return nil, fmt.Errorf("error running 'Build' command: %w", err)
	}
	procs = append(procs, proc)

	return adapter.SerialProcessesToStep(procs, callback), nil
}

func (s *UpService) Start(callback iface.OnExitCallback) (iface.Step, error) {
	procOpts := process.Options{
		Multiplexer: s.provider.Multiplexer,
		Name:        s.name,
		Wait:        false,
		Color:       s.provider.Color,
		OnExit:      func(info *process.ExitInfo) { callback(info, nil) },
	}

	// Starts the process
	proc, err := cmd.Up(&procOpts, s.provider.composeFile)
	if err != nil {
		return nil, fmt.Errorf("error starting 'Up' command: %w", err)
	}

	return AbortWithKillDecorator(adapter.ProcessToStep(proc), s.provider), err
}
