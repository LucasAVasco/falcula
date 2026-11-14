package enhanced

import (
	"errors"
	"sync"
	"testing"

	"github.com/LucasAVasco/falcula/service/iface"
	"github.com/LucasAVasco/falcula/service/status"
)

type fakeService struct{ prepareErr error }

func (f *fakeService) GetName() string      { return "fake" }
func (f *fakeService) GetOpts() *iface.Opts { return &iface.Opts{} }
func (f *fakeService) Prepare(cb iface.OnExitCallback) (iface.Step, error) {
	if f.prepareErr != nil {
		return nil, f.prepareErr
	}
	go cb(&iface.ExitInfo{}, nil)
	return nil, nil
}
func (f *fakeService) Start(cb iface.OnExitCallback) (iface.Step, error) {
	go cb(&iface.ExitInfo{}, nil)
	return nil, nil
}

func TestPrepareErrorSetsErrorStatus(t *testing.T) {
	e := NewEnhancedService(&fakeService{prepareErr: errors.New("boom")}, nil)
	if err := e.StartPrepare(); err == nil {
		t.Fatal("expected an error")
	}
	if got := e.GetStatus(); got != status.Error {
		t.Fatalf("expected the Error status, got %s", got)
	}
}

func TestStatusIsSafeForConcurrentReads(t *testing.T) {
	e := NewEnhancedService(&fakeService{}, nil)

	var wg sync.WaitGroup
	wg.Go(func() {
		_, _ = e.Prepare()
		_ = e.Start()
	})
	for range 4 {
		wg.Go(func() {
			for range 100 {
				_ = e.GetStatus()
			}
		})
	}
	wg.Wait()
}
