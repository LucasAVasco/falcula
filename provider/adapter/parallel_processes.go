package adapter

import (
	"errors"
	"fmt"
	"sync"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/service/iface"
)

var ErrAtLeastOneProcessFailed = errors.New("at least one process failed")

// ParallelProcessesStep wraps a list of processes in a Step interface and executes them in parallel until all processes end
type ParallelProcessesStep struct {
	processes []*process.Process

	aborted   bool
	waitGroup sync.WaitGroup

	exitInfoErrors []error
	startErrors    []error
	startError     error
	exitInfo       iface.ExitInfo
}

// ParallelProcessesToStep converts a list of processes to a Step that runs them in parallel and ends when all processes end
func ParallelProcessesToStep(processes []*process.Process, onEnd iface.OnExitCallback) iface.Step {
	p := ParallelProcessesStep{
		processes: processes,
	}

	p.waitGroup.Go(func() {
		for _, process := range p.processes {
			// Starts the process
			if !process.Started() {
				err := process.Start()
				if err != nil {
					p.startErrors = append(p.startErrors, err)
					break
				}
			}

			// Waits for the process to end
			exitInfo := process.Wait()
			if exitInfo.Error != nil {
				p.exitInfoErrors = append(p.exitInfoErrors, exitInfo.Error)
			}

			if exitInfo.HasError() {
				p.exitInfo.Code = 255 // Exit code 255 means that at least one process failed
			}
		}

		if len(p.exitInfoErrors) > 0 {
			p.exitInfoErrors = append(p.exitInfoErrors, ErrAtLeastOneProcessFailed)
			p.exitInfo.Error = errors.Join(p.exitInfoErrors...)
		}

		if len(p.startErrors) > 0 {
			p.startError = errors.Join(p.startErrors...)
		}

		if onEnd != nil {
			onEnd(&p.exitInfo, p.startError)
		}
	})

	return &p
}

func (p *ParallelProcessesStep) Wait() (*iface.ExitInfo, error) {
	p.waitGroup.Wait()

	return &p.exitInfo, p.startError
}

func (p *ParallelProcessesStep) Abort(force bool) (*iface.ExitInfo, error) {
	p.aborted = true

	for _, process := range p.processes {
		_, err := process.Stop(force)
		if err != nil {
			return &iface.ExitInfo{}, fmt.Errorf("error stopping process: %w", err)
		}
	}

	return &iface.ExitInfo{}, nil
}
