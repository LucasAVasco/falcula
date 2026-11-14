package process

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/multiplexer"
)

func newTestProcess(t *testing.T, onExit OnExitCallback, command string, args ...string) *Process {
	t.Helper()
	mux := multiplexer.New(func(_ *multiplexer.Client, b []byte) (int, error) { return len(b), nil })
	p, err := New(&Options{ManualStart: true, Multiplexer: mux, Name: "test", OnExit: onExit}, command, args...)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestStopBeforeStartPreventsStartAndReleasesWait(t *testing.T) {
	var exitCalls int
	p := newTestProcess(t, func(info *ExitInfo) { exitCalls++ }, "sleep", "10")

	info, err := p.Stop(true)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Stopped || info.HasError() {
		t.Fatalf("unexpected exit information: %+v", info)
	}

	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	if p.cmd.Process != nil {
		t.Fatal("the process must not start after being stopped")
	}
	if exitCalls != 1 {
		t.Fatalf("expected the exit callback to be called once, got %d", exitCalls)
	}
}

func TestStartErrorReleasesWait(t *testing.T) {
	p := newTestProcess(t, nil, "/nonexistent/command")
	if err := p.Start(); err == nil {
		t.Fatal("expected a start error")
	}

	done := make(chan *ExitInfo)
	go func() { done <- p.Wait() }()

	select {
	case info := <-done:
		if info.Error == nil {
			t.Fatal("expected the start error in the exit information")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Wait blocked after a start error")
	}
}

func TestConcurrentStopAndQueries(t *testing.T) {
	p := newTestProcess(t, func(*ExitInfo) {}, "sleep", "10")
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			_, _ = p.Stop(false)
			_ = p.Started()
			_ = p.Stopped()
		})
	}
	wg.Wait()

	if info := p.Wait(); !info.Stopped || info.HasError() {
		t.Fatalf("unexpected exit information: %+v", info)
	}
}

func TestNew_FailingCommand_ReportsExitCode(t *testing.T) {
	testutil.RequireShell(t)
	var received *ExitInfo
	p := newTestProcess(t, func(info *ExitInfo) { received = info }, "sh", "-c", "exit 3")

	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	info := p.Wait()

	if info.Code != 3 || info.Error == nil || !info.HasError() || info.Stopped {
		t.Fatalf("unexpected exit information: %+v", info)
	}
	if received == nil || received.Code != 3 {
		t.Fatalf("the exit callback received %+v", received)
	}
	if err := info.WrapError(); err == nil || !strings.Contains(err.Error(), "exit code: 3") {
		t.Fatalf("WrapError = %v", err)
	}
}

func TestNew_Options_DirEnvShellAndOutput(t *testing.T) {
	testutil.RequireShell(t)
	dir := t.TempDir()

	var mutex sync.Mutex
	var output strings.Builder
	mux := multiplexer.New(func(c *multiplexer.Client, b []byte) (int, error) {
		mutex.Lock()
		defer mutex.Unlock()
		output.WriteString(c.GetName() + "/" + c.GetLevel() + ":" + string(b))
		return len(b), nil
	})

	_, err := New(&Options{
		Dir:         dir,
		Env:         []string{"TEST_VALUE=from-env"},
		Shell:       true,
		Wait:        true,
		Multiplexer: mux,
		Name:        "proc",
	}, `echo "$TEST_VALUE $(pwd)"; echo problem >&2`)
	if err != nil {
		t.Fatal(err)
	}

	mutex.Lock()
	defer mutex.Unlock()
	got := output.String()
	for _, want := range []string{"proc/stdout:from-env " + dir + "\n", "proc/stderr:problem\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q does not contain %q", got, want)
		}
	}
}

func TestNew_WaitOption_ReturnsErrorOfFailingCommand(t *testing.T) {
	testutil.RequireShell(t)
	mux := multiplexer.New(func(_ *multiplexer.Client, b []byte) (int, error) { return len(b), nil })

	if _, err := New(&Options{Wait: true, Multiplexer: mux}, "false"); err == nil {
		t.Fatal("expected the error of the failing command")
	}
}

func TestExitInfo_StoppedIsNotAnError(t *testing.T) {
	info := ExitInfo{Code: 143, Stopped: true}
	if info.HasError() || info.WrapError() != nil {
		t.Fatal("a stopped process must not report an error")
	}
	if (&ExitInfo{}).HasError() {
		t.Fatal("a successful process must not report an error")
	}
}
