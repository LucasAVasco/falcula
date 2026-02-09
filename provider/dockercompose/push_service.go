package dockercompose

import (
	"fmt"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/provider/adapter"
	"github.com/LucasAVasco/falcula/provider/dockercompose/cmd"
	"github.com/LucasAVasco/falcula/service/iface"
)

// PushService is a service that push docker-compose images to a registry
type PushService struct {
	*Service
	repositories []string
}

func (s *PushService) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	procOpts := s.NewProcessOptions()
	procOpts.OnExit = func(info *process.ExitInfo) { callback(info, nil) }

	// Images to push
	images := s.Info.GetDefaultPushImages()

	// Does not need to build if there are no images
	if len(images) == 0 {
		return nil, nil
	}

	// Builds the images
	proc, err := cmd.Build(procOpts, s.Info.GetComposeFilePath())
	if err != nil {
		return nil, fmt.Errorf("error running 'Build' command: %w", err)
	}

	return adapter.ProcessToStep(proc), err
}

func (s *PushService) Start(callback iface.OnExitCallback) (iface.Step, error) {
	// Tags the images to push them to the repositories
	procOpts := s.NewProcessOptions()
	procOpts.Wait = true // Tag a image is a fast operation

	// Images to push
	images := s.Info.GetDefaultPushImages()

	// Tags the images
	for _, repository := range s.repositories {
		for _, image := range images {
			proc, err := cmd.Tag(procOpts, image, repository+"/"+image)
			if err != nil {
				return nil, fmt.Errorf("error running 'tag' command: %w", err)
			}

			exitInfo := proc.Wait()
			if exitInfo.Error != nil {
				return nil, fmt.Errorf("non zero exit code returned from 'tag' command (exit code: %d):  %w", exitInfo.Code, exitInfo.Error)
			}
		}
	}

	// Pushes the images to the repositories
	procOpts.Wait = false // Push a image is a slow operation

	procList := make([]*process.Process, 0)
	for _, repository := range s.repositories {
		for _, image := range images {
			proc, err := cmd.Push(procOpts, image, repository)
			if err != nil {
				return nil, fmt.Errorf("error running 'Push' command: %w", err)
			}

			procList = append(procList, proc)
		}
	}

	return adapter.ParallelProcessesToStep(procList, callback), nil
}
