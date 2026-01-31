package dockercompose

import (
	"fmt"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/provider/dockercompose/cmd"
	"github.com/LucasAVasco/falcula/service/iface"
)

// StepWithKill is a step that overrides the Abort method to use `docker-compose kill` to kill the process if the abort is forced
type StepWithKill struct {
	step     iface.Step
	provider *Provider
}

// AbortWithKillDecorator overrides the Abort method of the Step to use `docker-compose kill` to kill the process if the abort is forced
func AbortWithKillDecorator(step iface.Step, provider *Provider) iface.Step {
	return &StepWithKill{
		step:     step,
		provider: provider,
	}
}

func (s *StepWithKill) Wait() (*iface.ExitInfo, error) {
	return s.step.Wait()
}

func (s *StepWithKill) Abort(force bool) (*iface.ExitInfo, error) {
	if force {
		procOpts := process.Options{
			Multiplexer: s.provider.Multiplexer,
			Name:        s.provider.Name + ".kill",
			Color:       s.provider.Color,
		}

		proc, err := cmd.Kill(&procOpts, s.provider.composeFile)
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
