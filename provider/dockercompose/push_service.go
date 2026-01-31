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
	provider     *Provider
	name         string
	images       []string
	repositories []string
}

func (s *PushService) GetName() string {
	return s.name
}

func (s *PushService) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	procOpts := process.Options{
		Multiplexer: s.provider.Multiplexer,
		Name:        s.name,
		Wait:        true,
		Color:       s.provider.Color,
		OnExit:      func(info *process.ExitInfo) { callback(info, nil) },
	}

	// Does not need to build if there are no images
	if len(s.images) == 0 {
		return nil, nil
	}

	// Builds the images
	proc, err := cmd.Build(&procOpts, s.provider.composeFile)
	if err != nil {
		return nil, fmt.Errorf("error running 'Build' command: %w", err)
	}

	return adapter.ProcessToStep(proc), err
}

func (s *PushService) Start(callback iface.OnExitCallback) (iface.Step, error) {
	// Tags the images to push them to the repositories
	procOpts := process.Options{
		Multiplexer: s.provider.Multiplexer,
		Name:        s.name,
		Wait:        true, // Tag a image is a fast operation
		Color:       s.provider.Color,
	}

	// Tags the images
	for _, repository := range s.repositories {
		for _, image := range s.images {
			proc, err := cmd.Tag(&procOpts, image, repository+"/"+image)
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
	procOpts = process.Options{
		Multiplexer: s.provider.Multiplexer,
		Name:        s.name,
		Wait:        false,
		Color:       s.provider.Color,
	}

	procList := make([]*process.Process, 0)
	for _, repository := range s.repositories {
		for _, image := range s.images {
			proc, err := cmd.Push(&procOpts, image, repository)
			if err != nil {
				return nil, fmt.Errorf("error running 'Push' command: %w", err)
			}

			procList = append(procList, proc)
		}
	}

	return adapter.ParallelProcessesToStep(procList, callback), nil
}
