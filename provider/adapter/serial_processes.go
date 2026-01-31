package adapter

import (
	"fmt"
	"sync"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/service/iface"
)

// SerialProcessesStep wraps a list of processes in a Step interface and executes them serially until all processes end
type SerialProcessesStep struct {
	processes []*process.Process

	aborted    bool
	waitGroup  sync.WaitGroup
	exitInfo   *iface.ExitInfo
	startError error
}

// SerialProcessesToStep converts a list of processes to a Step that runs them serially and ends when all processes end
func SerialProcessesToStep(processes []*process.Process, onEnd iface.OnExitCallback) iface.Step {
	s := SerialProcessesStep{
		processes: processes,
		exitInfo:  &iface.ExitInfo{},
	}

	s.waitGroup.Go(func() {
		for _, process := range s.processes {
			if s.aborted {
				break
			}

			err := process.Start()
			if err != nil {
				s.startError = err
				break
			}

			s.exitInfo = process.Wait()
			if s.exitInfo.HasError() {
				break
			}
		}

		if onEnd != nil {
			onEnd(s.exitInfo, s.startError)
		}
	})

	return &s
}

func (s *SerialProcessesStep) Wait() (*iface.ExitInfo, error) {
	s.waitGroup.Wait()
	return s.exitInfo, nil
}

func (s *SerialProcessesStep) Abort(force bool) (*iface.ExitInfo, error) {
	s.aborted = true

	for _, process := range s.processes {
		if !process.Started() {
			continue
		}

		exitInfo, err := process.Stop(force)
		if err != nil {
			return exitInfo, fmt.Errorf("error stopping process: %w", err)
		}

		if exitInfo.HasError() {
			return exitInfo, fmt.Errorf("error from exit information of stop process: %w", exitInfo.Error)
		}
	}

	return &iface.ExitInfo{}, nil
}
