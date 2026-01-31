package dockercompose

import (
	"fmt"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/provider/adapter"
	"github.com/LucasAVasco/falcula/provider/dockercompose/cmd"
	"github.com/LucasAVasco/falcula/service/iface"
)

// DownService is a service that runs the docker-compose down command
type DownService struct {
	provider *Provider
	name     string
}

func (s *DownService) GetName() string {
	return s.name
}

func (s *DownService) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	callback(&iface.ExitInfo{}, nil)
	return nil, nil
}

func (s *DownService) Start(callback iface.OnExitCallback) (iface.Step, error) {
	procOpts := process.Options{
		Multiplexer: s.provider.Multiplexer,
		Name:        s.name,
		Wait:        false,
		Color:       s.provider.Color,
		OnExit: func(info *process.ExitInfo) {
			callback(info, nil)
		},
	}

	// Starts the process
	proc, err := cmd.Down(&procOpts, s.provider.composeFile)
	if err != nil {
		return nil, fmt.Errorf("error starting 'Down' command: %w", err)
	}

	return AbortWithKillDecorator(adapter.ProcessToStep(proc), s.provider), err
}
