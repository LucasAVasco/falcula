package manager

import (
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LucasAVasco/falcula/internal/testutil/testservice"
	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/iface"
	"github.com/LucasAVasco/falcula/service/status"
)

type fakeService struct{ name string }

func (f *fakeService) GetName() string                                  { return f.name }
func (f *fakeService) GetOpts() *iface.Opts                             { return &iface.Opts{} }
func (f *fakeService) Prepare(iface.OnExitCallback) (iface.Step, error) { return nil, nil }
func (f *fakeService) Start(iface.OnExitCallback) (iface.Step, error)   { return nil, nil }

func TestRemoveServicesRemovesAllWhenPassedGetServices(t *testing.T) {
	m := New("test")
	for _, name := range []string{"a", "b", "c", "d"} {
		m.AddService(&fakeService{name: name}, nil)
	}

	if err := m.RemoveServices(m.GetServices(), false, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(m.GetServices()); got != 0 {
		t.Fatalf("expected no services left, got %d", got)
	}
}

func TestConcurrentServiceListAccess(t *testing.T) {
	m := New("test")

	var wg sync.WaitGroup
	for i := range 8 {
		wg.Go(func() {
			svc := m.AddService(&fakeService{name: string(rune('a' + i))}, nil)
			_ = m.Prepare(nil).Wait()
			_ = m.GetServices()
			_ = m.RemoveService(svc, true, nil)
		})
	}
	wg.Wait()

	if got := len(m.GetServices()); got != 0 {
		t.Fatalf("expected no services left, got %d", got)
	}
}

// endedRecorder records the names of the services passed to an exit callback
type endedRecorder struct {
	mutex sync.Mutex
	names []string
	errs  int
}

func (r *endedRecorder) callback(svc *enhanced.EnhancedService, _ *iface.ExitInfo, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.names = append(r.names, svc.GetName())
	if err != nil {
		r.errs++
	}
}

func newManager(services ...*testservice.Service) *Manager {
	m := New("test")
	for _, svc := range services {
		m.AddService(svc, nil)
	}
	return m
}

func TestRun_AllServicesSucceed(t *testing.T) {
	a, b := &testservice.Service{Name: "a"}, &testservice.Service{Name: "b"}
	m := newManager(a, b)

	rec := &endedRecorder{}
	if err := m.Run(rec.callback); err != nil {
		t.Fatal(err)
	}

	slices.Sort(rec.names)
	if !slices.Equal(rec.names, []string{"a", "b"}) || rec.errs != 0 {
		t.Fatalf("callback received %v with %d errors", rec.names, rec.errs)
	}
	for _, svc := range m.GetServices() {
		if svc.GetStatus() != status.Ended {
			t.Errorf("service %s status = %s, want Ended", svc.GetName(), svc.GetStatus())
		}
	}
}

func TestRun_FailingService_ReturnsErrorAndReportsIt(t *testing.T) {
	var reported []error
	m := newManager(&testservice.Service{Name: "ok"}, &testservice.Service{Name: "bad", ExitCode: 1})
	m.OnError = func(_ *Manager, err error) { reported = append(reported, err) }

	err := m.Run(nil)
	if err == nil || !strings.Contains(err.Error(), "bad") {
		t.Fatalf("Run = %v, want an error about the failing service", err)
	}
	if len(reported) == 0 {
		t.Fatal("OnError must be called")
	}
}

func TestRun_PrepareFailure_DoesNotStartServices(t *testing.T) {
	svc := &testservice.Service{Name: "svc"}
	m := newManager(svc, &testservice.Service{Name: "broken", FailToPrepare: errors.New("boom")})

	if err := m.Run(nil); err == nil || !strings.Contains(err.Error(), "error preparing services") {
		t.Fatalf("Run = %v", err)
	}
	if svc.Starts() != 0 {
		t.Fatal("no service must start when the preparation fails")
	}
}

func TestRunSerial_StopsAtTheFirstFailure(t *testing.T) {
	first := &testservice.Service{Name: "first"}
	failing := &testservice.Service{Name: "failing", ExitCode: 1}
	last := &testservice.Service{Name: "last"}
	m := newManager(first, failing, last)

	rec := &endedRecorder{}
	if err := m.RunSerial(nil, rec.callback); err == nil {
		t.Fatal("expected an error")
	}

	if !slices.Equal(rec.names, []string{"first", "failing"}) || rec.errs != 1 {
		t.Fatalf("callback received %v with %d errors, want [first failing] with 1 error", rec.names, rec.errs)
	}
	if last.Starts() != 0 {
		t.Fatal("the service after the failure must not start")
	}
}

func TestStopAndRestart(t *testing.T) {
	running := &testservice.Service{Name: "running", RunForever: true}
	idle := &testservice.Service{Name: "idle"}
	m := newManager(running)
	m.AddService(idle, nil) // Never prepared

	runningSvc := m.GetServices()[0]
	if _, err := runningSvc.Prepare(); err != nil {
		t.Fatal(err)
	}
	if err := runningSvc.Start(); err != nil {
		t.Fatal(err)
	}

	// Services that are not running are ignored
	if err := m.Stop(true, nil).Wait(); err != nil {
		t.Fatal(err)
	}
	if runningSvc.GetStatus() != status.Stopped {
		t.Fatalf("status = %s, want Stopped", runningSvc.GetStatus())
	}

	// Restart resets, prepares and starts every service
	done := make(chan error, 1)
	go func() { done <- m.Restart(true, nil) }()
	waitFor(t, func() bool { return running.Starts() == 2 })
	if err := m.Close(true, nil); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if idle.Starts() != 1 {
		t.Fatalf("the idle service must also be started by the restart, starts = %d", idle.Starts())
	}
}

func TestDisableAndEnableAllServices(t *testing.T) {
	m := newManager(&testservice.Service{Name: "a"}, &testservice.Service{Name: "b"})

	if err := m.Disable(nil).Wait(); err != nil {
		t.Fatal(err)
	}
	for _, svc := range m.GetServices() {
		if svc.GetStatus() != status.Disabled {
			t.Fatalf("%s status = %s, want Disabled", svc.GetName(), svc.GetStatus())
		}
	}

	if err := m.Disable(nil).Wait(); err == nil {
		t.Fatal("disabling disabled services must fail")
	}
	if err := m.Enable(nil).Wait(); err != nil {
		t.Fatal(err)
	}
	if err := m.Run(nil); err != nil {
		t.Fatalf("enabled services must run: %v", err)
	}
}

// waitFor polls the condition until it is true (fails the test after a timeout)
func waitFor(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("condition not satisfied in time")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
