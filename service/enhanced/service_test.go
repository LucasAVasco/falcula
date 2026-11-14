package enhanced

import (
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil/testservice"

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

// statusRecorder records the statuses reported by the callback, in order
type statusRecorder struct {
	mutex    sync.Mutex
	statuses []status.Status
}

func (r *statusRecorder) callbacks() *Callbacks {
	return &Callbacks{OnServiceStatusChanged: func(svc *EnhancedService) {
		r.mutex.Lock()
		defer r.mutex.Unlock()
		r.statuses = append(r.statuses, svc.GetStatus())
	}}
}

func (r *statusRecorder) get() []status.Status {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return slices.Clone(r.statuses)
}

func TestPrepareAndRun_Success_EndsWithEndedStatus(t *testing.T) {
	rec := &statusRecorder{}
	e := NewEnhancedService(&testservice.Service{Name: "svc"}, rec.callbacks())

	if _, err := e.Prepare(); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Run(); err != nil {
		t.Fatal(err)
	}

	want := []status.Status{status.Preparing, status.Ready, status.Running, status.Ended}
	if got := rec.get(); !slices.Equal(got, want) {
		t.Fatalf("statuses = %v, want %v", got, want)
	}
}

func TestRun_FailingStep_ReturnsErrorAndSetsErrorStatus(t *testing.T) {
	e := NewEnhancedService(&testservice.Service{Name: "svc", ExitCode: 2}, nil)
	if _, err := e.Prepare(); err != nil {
		t.Fatal(err)
	}

	info, err := e.Run()
	if err == nil || info.Code != 2 {
		t.Fatalf("Run = %+v, %v; want the exit code 2 and an error", info, err)
	}
	if e.GetStatus() != status.Error {
		t.Fatalf("status = %s, want Error", e.GetStatus())
	}

	// Reset from the Error status
	if _, err := e.Reset(false); err != nil || e.GetStatus() != status.None {
		t.Fatalf("Reset = %v, status %s", err, e.GetStatus())
	}
}

func TestStartBeforeReady_ReturnsInvalidStatus(t *testing.T) {
	e := NewEnhancedService(&testservice.Service{Name: "svc"}, nil)

	if err := e.Start(); !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("Start = %v, want ErrInvalidStatus", err)
	}
	if _, err := e.Wait(); !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("Wait = %v, want ErrInvalidStatus", err)
	}
	if _, err := e.WaitPrepare(); !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("WaitPrepare = %v, want ErrInvalidStatus", err)
	}
	if _, err := e.AbortPrepareOrStop(false); err != nil {
		t.Errorf("stopping a service that does nothing must not fail: %v", err)
	}
}

func TestStop_RunningService_SetsStoppedStatus(t *testing.T) {
	svc := &testservice.Service{Name: "svc", RunForever: true}
	e := NewEnhancedService(svc, nil)
	if _, err := e.Prepare(); err != nil {
		t.Fatal(err)
	}
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	if e.GetStatus() != status.Running {
		t.Fatalf("status = %s, want Running", e.GetStatus())
	}

	info, err := e.Stop(true)
	if err != nil || !info.Stopped || e.GetStatus() != status.Stopped {
		t.Fatalf("Stop = %+v, %v, status %s", info, err, e.GetStatus())
	}

	// Stopping again does nothing
	if info, err := e.Stop(true); info != nil || err != nil {
		t.Fatalf("second Stop = %+v, %v", info, err)
	}
}

func TestRestart_RunningService_RunsItAgain(t *testing.T) {
	svc := &testservice.Service{Name: "svc", RunForever: true}
	e := NewEnhancedService(svc, nil)
	if _, err := e.Prepare(); err != nil {
		t.Fatal(err)
	}
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}

	if _, err := e.Restart(true); err != nil {
		t.Fatal(err)
	}
	if svc.Starts() != 2 || e.GetStatus() != status.Running {
		t.Fatalf("starts = %d, status %s; want 2 starts and Running", svc.Starts(), e.GetStatus())
	}
	if _, err := e.Stop(true); err != nil {
		t.Fatal(err)
	}
}

func TestDisableAndEnable(t *testing.T) {
	svc := &testservice.Service{Name: "svc", RunForever: true}
	e := NewEnhancedService(svc, nil)

	if err := e.Enable(); !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("enabling an enabled service = %v, want ErrInvalidStatus", err)
	}
	if err := e.Disable(); err != nil || e.GetStatus() != status.Disabled {
		t.Fatalf("Disable = %v, status %s", err, e.GetStatus())
	}

	// A disabled service ignores every operation
	if _, err := e.Prepare(); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Run(); err != nil || svc.Starts() != 0 || e.GetStatus() != status.Disabled {
		t.Fatalf("Run = %v, starts %d, status %s", err, svc.Starts(), e.GetStatus())
	}

	if err := e.Enable(); err != nil || e.GetStatus() != status.None {
		t.Fatalf("Enable = %v, status %s", err, e.GetStatus())
	}

	// A running service can not be disabled
	if _, err := e.Prepare(); err != nil {
		t.Fatal(err)
	}
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	if err := e.Disable(); !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("disabling a running service = %v, want ErrInvalidStatus", err)
	}
	if _, err := e.Stop(true); err != nil {
		t.Fatal(err)
	}
}

func TestStartDisabledOption(t *testing.T) {
	e := NewEnhancedService(&testservice.Service{Name: "svc", Opts: iface.Opts{StartDisabled: true}}, nil)
	if e.GetStatus() != status.Disabled {
		t.Fatalf("status = %s, want Disabled", e.GetStatus())
	}
}
