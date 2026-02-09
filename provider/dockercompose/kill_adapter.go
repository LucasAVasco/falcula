package dockercompose

import (
	"fmt"

	"github.com/LucasAVasco/falcula/provider/dockercompose/cmd"
	"github.com/LucasAVasco/falcula/service/iface"
)

// StepWithKill is a step that overrides the Abort method to use `docker-compose kill` to kill the process if the abort is forced
type StepWithKill struct {
	step    iface.Step
	service *Service
}

// AbortWithKillDecorator overrides the Abort method of the Step to use `docker-compose kill` to kill the process if the abort is forced
func AbortWithKillDecorator(step iface.Step, service *Service) iface.Step {
	return &StepWithKill{
		step:    step,
		service: service,
	}
}

func (s *StepWithKill) Wait() (*iface.ExitInfo, error) {
	return s.step.Wait()
}

func (s *StepWithKill) Abort(force bool) (*iface.ExitInfo, error) {
	if force {
		procOpts := s.service.NewProcessOptions()
		procOpts.ManualStart = false

		proc, err := cmd.Kill(procOpts, s.service.Info.GetComposeFilePath())
		if err != nil {
			return nil, fmt.Errorf("error running 'Kill' command: %w", err)
		}

		exitInfo := proc.Wait()
		if exitInfo.HasError() {
			return exitInfo, fmt.Errorf("error from exit information of 'Kill' command: %w", exitInfo.WrapError())
		}
	}

	return s.step.Abort(force)
}
