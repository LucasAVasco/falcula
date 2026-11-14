package dockercompose

import (
	"fmt"
	"strings"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/provider/adapter"
	"github.com/LucasAVasco/falcula/provider/base"
	"github.com/LucasAVasco/falcula/provider/dockercompose/info"
	"github.com/LucasAVasco/falcula/service/iface"
)

// ServiceOpts is the base options for a docker-compose service
type ServiceOpts = base.ServiceOpts

// Service represents a generic docker-compose service
type Service struct {
	*base.Service
	Info *info.DockerComposeInfo
}

func NewService(service *base.Service, info *info.DockerComposeInfo) *Service {
	return &Service{
		Service: service,
		Info:    info,
	}
}

// Prepare does nothing (most docker-compose services do not have a preparation step). Calls the callback immediately
func (s *Service) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	callback(&iface.ExitInfo{}, nil)
	return nil, nil
}

// getPlatformSuffix gets the suffix to use in the tag of an image built for a platform (e.g. 'linux/amd64' -> 'linux-amd64')
func getPlatformSuffix(platform string) string {
	return strings.ReplaceAll(platform, "/", "-")
}

// setDockerPlatformEnv sets the DOCKER_DEFAULT_PLATFORM environment variable if the platform is not empty
func setDockerPlatformEnv(opts *process.Options, platform string) {
	if platform != "" {
		opts.Env = append(opts.Env, fmt.Sprintf("DOCKER_DEFAULT_PLATFORM=%s", platform))
	}
}

// composeCommand is a function that creates a docker-compose command process (see the `cmd` package)
type composeCommand = func(opts *process.Options, composeFile string, args ...string) (*process.Process, error)

// runComposeCommand runs a single docker-compose command for the given platform (optional) and returns a step that waits for it. The step
// uses `docker-compose kill` when its abort is forced
func (s *Service) runComposeCommand(
	callback iface.OnExitCallback,
	platform string,
	command composeCommand,
	args ...string,
) (iface.Step, error) {
	procOpts := s.NewProcessOptions()
	procOpts.OnExit = func(info *process.ExitInfo) { callback(info, nil) }
	setDockerPlatformEnv(procOpts, platform)

	proc, err := command(procOpts, s.Info.GetComposeFilePath(), args...)
	if err != nil {
		return nil, err
	}

	return AbortWithKillDecorator(adapter.ProcessToStep(proc), s), nil
}
