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

	// Pull process
	procs := make([]*process.Process, 0, 2)

	proc, err := cmd.Pull(procOpts, s.Info.GetComposeFilePath())
	if err != nil {
		return nil, fmt.Errorf("error running 'Pull' command: %w", err)
	}
	procs = append(procs, proc)

	// Build process
	setDockerPlatformEnv(procOpts, s.platform)

	proc, err = cmd.Build(procOpts, s.Info.GetComposeFilePath())
	if err != nil {
		return nil, fmt.Errorf("error running 'Build' command: %w", err)
	}
	procs = append(procs, proc)

	return adapter.SerialProcessesToStep(procs, callback), nil
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
