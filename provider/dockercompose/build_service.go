package dockercompose

import (
	"fmt"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/provider/adapter"
	"github.com/LucasAVasco/falcula/provider/dockercompose/cmd"
	"github.com/LucasAVasco/falcula/service/iface"
)

// BuildService is a service that pulls and builds the docker-compose images.
type BuildService struct {
	provider  *Provider
	name      string
	onlyBuild bool
}

func (s *BuildService) GetName() string {
	return s.name
}

func (s *BuildService) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	callback(&iface.ExitInfo{}, nil)
	return nil, nil
}

func (s *BuildService) Start(callback iface.OnExitCallback) (iface.Step, error) {
	procOpts := process.Options{
		Multiplexer: s.provider.Multiplexer,
		Name:        s.name,
		Wait:        false,
		ManualStart: true,
		Color:       s.provider.Color,
	}

	// Pull and build processes
	procs := make([]*process.Process, 0, 2)

	if !s.onlyBuild {
		proc, err := cmd.Pull(&procOpts, s.provider.composeFile)
		if err != nil {
			return nil, fmt.Errorf("error running 'Pull' command: %w", err)
		}
		procs = append(procs, proc)
	}

	proc, err := cmd.Build(&procOpts, s.provider.composeFile)
	if err != nil {
		return nil, fmt.Errorf("error running 'Build' command: %w", err)
	}
	procs = append(procs, proc)

	return adapter.SerialProcessesToStep(procs, callback), err
}
