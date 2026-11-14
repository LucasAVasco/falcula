package dockercompose

import (
	"fmt"

	"github.com/LucasAVasco/falcula/provider/base"
	"github.com/LucasAVasco/falcula/provider/dockercompose/cmd"
	"github.com/LucasAVasco/falcula/service/iface"
)

// DownServiceOpts is the options for the DownService
type DownServiceOpts struct {
	base.ServiceOpts       `lua:",inline"`
	RemoveAnonymousVolumes bool `lua:"volumes"` // Also removes the anonymous volumes, default is false
}

// DownService is a service that runs the docker-compose down command
type DownService struct {
	*Service
	opts *DownServiceOpts
}

func (s *DownService) Start(callback iface.OnExitCallback) (iface.Step, error) {
	args := []string{}
	if s.opts.RemoveAnonymousVolumes {
		args = append(args, "--volumes")
	}

	step, err := s.runComposeCommand(callback, "", cmd.Down, args...)
	if err != nil {
		return nil, fmt.Errorf("error starting 'Down' command: %w", err)
	}

	return step, nil
}
