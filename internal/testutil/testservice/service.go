// Package testservice implements a deterministic fake service for tests. It must only be imported by tests
package testservice

import (
	"sync"

	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/service/iface"
)

// Step is a fake step. Like a real process, its exit callback is called before `Wait` returns
type Step struct {
	done     chan struct{}
	once     sync.Once
	info     *iface.ExitInfo
	callback iface.OnExitCallback
}

func newStep(callback iface.OnExitCallback) *Step {
	return &Step{done: make(chan struct{}), callback: callback}
}

// End ends the step with the exit information (only the first call has effect)
func (s *Step) End(info *iface.ExitInfo) {
	s.once.Do(func() {
		s.info = info
		s.callback(info, nil)
		close(s.done)
	})
}

func (s *Step) Wait() (*iface.ExitInfo, error) {
	<-s.done
	return s.info, nil
}

func (s *Step) Abort(bool) (*iface.ExitInfo, error) {
	s.End(&iface.ExitInfo{Stopped: true})
	return s.info, nil
}

// Service is a fake service. Its preparing step succeeds immediately. Its main step ends immediately with `ExitCode`, or runs until it
// is aborted if `RunForever` is true
type Service struct {
	Name          string
	Opts          iface.Opts
	ExitCode      int  // Exit code of the main step
	RunForever    bool // The main step only ends when aborted
	FailToPrepare error

	mutex  sync.Mutex
	starts int
}

func (s *Service) GetName() string      { return s.Name }
func (s *Service) GetOpts() *iface.Opts { return &s.Opts }

func (s *Service) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	if s.FailToPrepare != nil {
		return nil, s.FailToPrepare
	}

	step := newStep(callback)
	step.End(&iface.ExitInfo{})
	return step, nil
}

func (s *Service) Start(callback iface.OnExitCallback) (iface.Step, error) {
	s.mutex.Lock()
	s.starts++
	s.mutex.Unlock()

	step := newStep(callback)
	if !s.RunForever {
		step.End(&iface.ExitInfo{Code: process.ExitCode(s.ExitCode)})
	}
	return step, nil
}

// Starts returns how many times the main step was started
func (s *Service) Starts() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.starts
}
