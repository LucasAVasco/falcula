package process

import (
	"sync"
	"testing"
	"time"

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
