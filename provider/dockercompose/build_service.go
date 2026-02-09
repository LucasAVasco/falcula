package dockercompose

import (
	"fmt"
	"strings"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/provider/adapter"
	"github.com/LucasAVasco/falcula/provider/dockercompose/cmd"
	"github.com/LucasAVasco/falcula/service/iface"
)

// BuildInfo is the information to build the docker-compose images. Used to build all provided services to all provided platforms
type BuildInfo struct {
	ServicesNames []string
	Platforms     []string
}

type BuildServiceOpts struct {
	NoPull bool         // Does not pull images, only builds
	Builds []*BuildInfo // Builds
}

// BuildService is a service that pulls and builds the docker-compose images.
type BuildService struct {
	*Service
	opts *BuildServiceOpts
}

// setDockerPlatformEnv sets the DOCKER_DEFAULT_PLATFORM environment variable if the platform is not empty
func setDockerPlatformEnv(opts *process.Options, platform string) {
	if platform != "" {
		opts.Env = append(opts.Env, fmt.Sprintf("DOCKER_DEFAULT_PLATFORM=%s", platform))
	}
}

// generateBuildProcessesForPlatform generates build processes for a specific platform.
//
// If building to a different platform, the images will be tagged with the platform suffix. Example: a image will be tagged as
// 'my-image-name:linux-amd64' if building to the 'linux/amd64' platform
func (s *BuildService) generateBuildProcessesForPlatform(services []string, platform string) ([]*process.Process, error) {
	// Default services
	if services == nil {
		services = []string{}
	}
	if len(services) == 0 {
		var err error
		services, err = s.Info.GetDefaultBuildServices()
		if err != nil {
			return nil, fmt.Errorf("error getting default build services: %w", err)
		}
	}

	// Build service
	procOpts := s.NewProcessOptions()
	setDockerPlatformEnv(procOpts, platform)

	procList := make([]*process.Process, 0, 1+len(services))
	proc, err := cmd.Build(procOpts, s.Info.GetComposeFilePath(), services...)
	if err != nil {
		return nil, fmt.Errorf("error creating 'Build' command: %w", err)
	}
	procList = append(procList, proc)

	// Tag the images with a platform suffix
	if platform != "" {
		procOpts.Env = []string{} // No need to set the default platform

		platform = strings.ReplaceAll(platform, "/", "-")
		for _, service := range services {
			// Source and destination srcImage
			srcImage, err := s.Info.GetServiceImage(service)
			if err != nil {
				return nil, fmt.Errorf("error getting service image: %w", err)
			}

			destImage := srcImage
			if platform != "" {
				destImage = srcImage + ":" + platform
			}

			// Process
			proc, err = cmd.Tag(procOpts, srcImage, destImage)
			if err != nil {
				return nil, fmt.Errorf("error creating 'Tag' command: %w", err)
			}
			procList = append(procList, proc)
		}
	}

	return procList, nil
}

// generateBuildProcesses generates build processes for a build information
//
// Tags the images with a platform suffix if the platform is provided in the build info
func (s *BuildService) generateBuildProcesses(info *BuildInfo) ([]*process.Process, error) {
	procList := make([]*process.Process, 0, 1+len(info.Platforms))

	// Default platforms
	if info.Platforms == nil {
		info.Platforms = []string{}
	}

	// No platform provided (use default platform)
	if len(info.Platforms) == 0 {
		newProcList, err := s.generateBuildProcessesForPlatform(info.ServicesNames, "")
		if err != nil {
			return nil, fmt.Errorf("error getting build processes of '%v' services for default platform: %w", info.ServicesNames, err)
		}
		procList = append(procList, newProcList...)
	}

	// Platforms provided (use provided platforms)
	for _, platform := range info.Platforms {
		newProcList, err := s.generateBuildProcessesForPlatform(info.ServicesNames, platform)
		if err != nil {
			return nil, fmt.Errorf(
				"error getting build processes of '%v' services for platform '%s': %w",
				info.ServicesNames,
				platform,
				err,
			)
		}
		procList = append(procList, newProcList...)
	}

	return procList, nil
}

// generateBuildListProcesses generates build processes for a list of build information
//
// Tags the images with a platform suffix if the platform is provided in the build info
func (s *BuildService) generateBuildListProcesses(infoList []*BuildInfo) ([]*process.Process, error) {
	procList := make([]*process.Process, 0, 1+len(infoList))

	for _, info := range infoList {
		newProcList, err := s.generateBuildProcesses(info)
		if err != nil {
			return nil, fmt.Errorf("error getting build processes for the '%+v' build information: %w", info, err)
		}
		procList = append(procList, newProcList...)
	}

	return procList, nil
}

func (s *BuildService) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	callback(&iface.ExitInfo{}, nil)
	return nil, nil
}

func (s *BuildService) Start(callback iface.OnExitCallback) (iface.Step, error) {
	procOpts := s.NewProcessOptions()

	// Pull and build processes
	procList := make([]*process.Process, 0, 2)

	if !s.opts.NoPull {
		proc, err := cmd.Pull(procOpts, s.Info.GetComposeFilePath())
		if err != nil {
			return nil, fmt.Errorf("error running 'Pull' command: %w", err)
		}
		procList = append(procList, proc)
	}

	newProcList, err := s.generateBuildListProcesses(s.opts.Builds)
	if err != nil {
		return nil, fmt.Errorf("error running 'Build' command: %w", err)
	}
	procList = append(procList, newProcList...)

	return adapter.SerialProcessesToStep(procList, callback), err
}
