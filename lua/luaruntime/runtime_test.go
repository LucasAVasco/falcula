package luaruntime

import (
	"sync"
	"testing"
	"time"

	"github.com/LucasAVasco/falcula/service/iface"
	"github.com/LucasAVasco/falcula/service/manager"

	lua "github.com/yuin/gopher-lua"
)

// blockingStep is a step that runs until it is aborted
type blockingStep struct {
	running  chan struct{} // Closed when someone waits for the step
	stop     chan struct{}
	stopOnce sync.Once
	runOnce  sync.Once
	callback iface.OnExitCallback
}

func (s *blockingStep) Wait() (*iface.ExitInfo, error) {
	s.runOnce.Do(func() { close(s.running) })
	<-s.stop
	return &iface.ExitInfo{Stopped: true}, nil
}

func (s *blockingStep) Abort(bool) (*iface.ExitInfo, error) {
	s.stopOnce.Do(func() {
		close(s.stop)
		s.callback(&iface.ExitInfo{Stopped: true}, nil)
	})
	return &iface.ExitInfo{Stopped: true}, nil
}

// blockingService is a service whose main step runs until it is stopped
type blockingService struct {
	running chan struct{}
}

func (s *blockingService) GetName() string      { return "blocking" }
func (s *blockingService) GetOpts() *iface.Opts { return &iface.Opts{} }
func (s *blockingService) Prepare(callback iface.OnExitCallback) (iface.Step, error) {
	callback(&iface.ExitInfo{}, nil)
	return nil, nil
}
func (s *blockingService) Start(callback iface.OnExitCallback) (iface.Step, error) {
	return &blockingStep{running: s.running, stop: make(chan struct{}), callback: callback}, nil
}

func newTestRuntime(t *testing.T) *Runtime {
	t.Helper()
	r, err := New(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r
}

// runInBackground runs the Lua code in another goroutine and returns a channel with its result
func runInBackground(r *Runtime, code string) <-chan error {
	result := make(chan error, 1)
	go func() { result <- r.Run(code) }()
	return result
}

func waitResult(t *testing.T, result <-chan error) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(10 * time.Second):
		t.Fatal("the script did not return after being interrupted")
		return nil
	}
}

// Reproduces the crash when the user selects new arguments while the script waits for its services
func TestInterruptWhileWaitingForServices(t *testing.T) {
	r := newTestRuntime(t)

	running := make(chan struct{})
	man := manager.New("test")
	man.AddService(&blockingService{running: running}, nil)
	r.AddManager(man)

	r.GetLuaState().SetGlobal("run_manager", r.GetLuaState().NewFunction(func(L *lua.LState) int {
		_ = man.Run(nil)
		return 0
	}))

	result := runInBackground(r, `
		run_manager()
		while true do end -- Keeps running after the services stop. Only the interruption ends it
	`)

	<-running
	r.Interrupt()

	if err := waitResult(t, result); err == nil {
		t.Fatal("expected the interrupted script to return an error")
	}
	if !r.Interrupted() {
		t.Fatal("expected the runtime to report the interruption")
	}
	if got := len(r.GetManagers()); got != 0 {
		t.Fatalf("expected the managers to be closed, got %d", got)
	}

	// The state can be reset and reused by the goroutine that runs the scripts
	r.ResetLuaState()
	if err := r.Run(`x = 1`); err != nil {
		t.Fatalf("running after the interruption: %v", err)
	}
	if r.Interrupted() {
		t.Fatal("a new script must not be reported as interrupted")
	}
}

func TestInterruptBusyScript(t *testing.T) {
	r := newTestRuntime(t)

	result := runInBackground(r, `while true do end`)

	time.Sleep(100 * time.Millisecond) // Lets the script start
	r.Interrupt()

	if err := waitResult(t, result); err == nil {
		t.Fatal("expected the interrupted script to return an error")
	}
}

func TestConcurrentManagersAccess(t *testing.T) {
	r := newTestRuntime(t)

	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 50 {
				r.AddManager(manager.New("m"))
				_ = r.GetManagers()
			}
		})
	}
	wg.Go(func() {
		for range 20 {
			r.Interrupt()
		}
	})
	wg.Wait()
}
