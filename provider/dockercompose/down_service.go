package dockercompose

import (
	"fmt"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/provider/adapter"
	"github.com/LucasAVasco/falcula/provider/base"
	"github.com/LucasAVasco/falcula/provider/dockercompose/cmd"
	"github.com/LucasAVasco/falcula/service/iface"
)

// DownServiceOpts is the options for the DownService
type DownServiceOpts struct {
	base.ServiceOpts
	RemoveAnonymousVolumes bool // Also removes the anonymous volumes, default is false
}

// DownService is a service that runs the docker-compose down command
type DownService struct {
	*Service
	opts *DownServiceOpts
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
	args := []string{}
	if s.opts.RemoveAnonymousVolumes {
		args = append(args, "--volumes")
	}
	proc, err := cmd.Down(procOpts, s.Info.GetComposeFilePath(), args...)
	if err != nil {
		return nil, fmt.Errorf("error starting 'Down' command: %w", err)
	}

	return AbortWithKillDecorator(adapter.ProcessToStep(proc), s.Service), err
}
