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
	*Service
	provider *Provider
}

func (s *DownService) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	callback(&iface.ExitInfo{}, nil)
	return nil, nil
}

func (s *DownService) Start(callback iface.OnExitCallback) (iface.Step, error) {
	procOpts := s.NewProcessOptions()
	procOpts.OnExit = func(info *process.ExitInfo) {
		callback(info, nil)
	}

	// Starts the process
	proc, err := cmd.Down(procOpts, s.Info.GetComposeFilePath())
	if err != nil {
		return nil, fmt.Errorf("error starting 'Down' command: %w", err)
	}

	return AbortWithKillDecorator(adapter.ProcessToStep(proc), s.Service), err
}
