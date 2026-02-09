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
	*Service
	platform string
}

func (s *UpService) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	procOpts := s.NewProcessOptions()
	procOpts.OnExit = func(info *process.ExitInfo) { callback(info, nil) }

	setDockerPlatformEnv(procOpts, s.platform)

	// Run UP process that only builds the containers without starting
	proc, err := cmd.Up(procOpts, s.Info.GetComposeFilePath(), "--build", "--no-start")
	if err != nil {
		return nil, fmt.Errorf("error running 'up --build --no-start' command: %w", err)
	}

	return adapter.ProcessToStep(proc), nil
}

func (s *UpService) Start(callback iface.OnExitCallback) (iface.Step, error) {
	procOpts := s.NewProcessOptions()
	procOpts.OnExit = func(info *process.ExitInfo) { callback(info, nil) }
	setDockerPlatformEnv(procOpts, s.platform)

	// Starts the process
	proc, err := cmd.Up(procOpts, s.Info.GetComposeFilePath())
	if err != nil {
		return nil, fmt.Errorf("error starting 'Up' command: %w", err)
	}

	return AbortWithKillDecorator(adapter.ProcessToStep(proc), s.Service), err
}
