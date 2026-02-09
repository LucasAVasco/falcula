package dockercompose

import (
	"fmt"
	"strings"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/provider/adapter"
	"github.com/LucasAVasco/falcula/provider/dockercompose/cmd"
	"github.com/LucasAVasco/falcula/service/iface"
)

// PushInfo is the information to push a docker-compose image
type PushInfo struct {
	Services   []string // Push images from this service
	Images     []string // Push these images
	Platforms  []string // Platforms of the images to push, Applied to all images
	Registries []string // Pushes the images to these registries
	Tag        string   // Tag of the image in the registry. Applied to all images
}

// PushServiceOpts is the options for the push service
type PushServiceOpts struct {
	Pushes []*PushInfo
}

// PushService is a service that push docker-compose images to a registry
type PushService struct {
	*Service
	opts *PushServiceOpts
}

// getPlatformSuffix gets the suffix to use in the tag of the image
func getPlatformSuffix(platform string) string {
	return strings.ReplaceAll(platform, "/", "-")
}

// generatePushImageProcesses generates the processes to push images to a registry. Supports single platform push and multiple platforms
// images (manifest). The tag is the tag of the image in the registry
func (s *PushService) generatePushImageProcesses(image, registry, tag string, platforms []string) ([]*process.Process, error) {
	// Validation
	if strings.Contains(image, ":") {
		return nil, fmt.Errorf("image name cannot contain ':', but it is '%s'", image)
	}

	if strings.Contains(tag, ":") {
		return nil, fmt.Errorf("tag cannot contain ':', but it is '%s'", tag)
	}

	// Process data
	procOpts := s.NewProcessOptions()
	procList := make([]*process.Process, 0, 1+len(platforms))

	// Default platform
	if platforms == nil {
		platforms = []string{}
	}

	// Manifest in the registry with all images
	registryManifest := registry + "/" + image
	if tag != "" {
		registryManifest += ":" + tag
	}

	// Images in the host and registry. Related by index. Ex: hostImages[0] -> registryImages[0]
	hostImages := make([]string, 0, len(platforms))
	registryImages := make([]string, 0, len(platforms))

	if len(platforms) == 0 {
		// No platform provided. Pushes the default image to the registry as a Manifest
		hostImages = append(hostImages, image)
		registryImages = append(registryImages, registryManifest)
	} else if len(platforms) == 1 {
		// One platform provided. Pushes the provided platform to the registry as a Manifest
		hostImages = append(hostImages, image+":"+getPlatformSuffix(platforms[0]))
		registryImages = append(registryImages, registryManifest)
	} else {
		// Multiple platforms provided. Pushes all images to the registry with the platforms suffixes
		for _, platform := range platforms {
			hostImages = append(hostImages, image+":"+getPlatformSuffix(platform))
			registryImages = append(registryImages, registry+"/"+image+":"+getPlatformSuffix(platform))
		}
	}

	// Pushes all images to the registry
	for i, hostImage := range hostImages {
		registryImage := registryImages[i]

		// Tags the image with the registry
		proc, err := cmd.Tag(procOpts, hostImage, registryImage)
		if err != nil {
			return nil, fmt.Errorf("error creating 'tag' command: %w", err)
		}
		procList = append(procList, proc)

		// Pushes the image to the registry
		proc, err = cmd.Push(procOpts, registryImage, "")
		if err != nil {
			return nil, fmt.Errorf("error creating 'push' command: %w", err)
		}
		procList = append(procList, proc)
	}

	// Creates the manifest
	if len(platforms) > 1 {
		proc, err := cmd.ManifestCreate(procOpts, registryManifest, "", registryImages...)
		if err != nil {
			return nil, fmt.Errorf("error creating 'manifest create' command: %w", err)
		}
		procList = append(procList, proc)
	}

	return procList, nil
}

// generatePushProcesses generates the processes to push images to a registry
func (s *PushService) generatePushProcesses(info *PushInfo) ([]*process.Process, error) {
	// Default images to push
	if info.Images == nil {
		info.Images = []string{}
	}
	if info.Services == nil {
		info.Services = []string{}
	}

	if len(info.Images) == 0 && len(info.Services) == 0 {
		images, err := s.Info.GetDefaultPushImages()
		if err != nil {
			return nil, fmt.Errorf("error getting default push images: %w", err)
		}
		info.Images = images
	}

	// Images to push
	images := append([]string{}, info.Images...)

	// Services images to push
	for _, service := range info.Services {
		serviceImage, err := s.Info.GetServiceImage(service)
		if err != nil {
			return nil, fmt.Errorf("error getting service image: %w", err)
		}
		images = append(images, serviceImage)
	}

	// Pushes the images to all registries
	procList := make([]*process.Process, 0, 1+len(images))
	for _, registry := range info.Registries {
		for _, image := range images {
			newProcList, err := s.generatePushImageProcesses(image, registry, info.Tag, info.Platforms)
			if err != nil {
				return nil, fmt.Errorf("error getting push processes for the image '%s': %w", image, err)
			}
			procList = append(procList, newProcList...)
		}
	}

	return procList, nil
}

// generatePushListProcesses generates the processes to push images of a list to a registry
func (s *PushService) generatePushListProcesses(infoList []*PushInfo) ([]*process.Process, error) {
	procList := make([]*process.Process, 0, len(infoList))
	for _, info := range infoList {
		newProcList, err := s.generatePushProcesses(info)
		if err != nil {
			return nil, fmt.Errorf("error generating push processes for information '%+v': %w", info, err)
		}
		procList = append(procList, newProcList...)
	}
	return procList, nil
}

func (s *PushService) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	callback(&iface.ExitInfo{}, nil)
	return nil, nil
}

func (s *PushService) Start(callback iface.OnExitCallback) (iface.Step, error) {
	if s.opts == nil {
		return nil, fmt.Errorf("no options provided")
	}

	if s.opts.Pushes == nil {
		return nil, fmt.Errorf("no push information provided")
	}

	// __AUTO_GENERATED_PRINT_VAR_START__
	procList, err := s.generatePushListProcesses(s.opts.Pushes)
	if err != nil {
		return nil, fmt.Errorf("error getting push processes: %w", err)
	}
	return adapter.SerialProcessesToStep(procList, callback), nil
}
