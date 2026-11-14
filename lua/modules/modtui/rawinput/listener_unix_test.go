//go:build linux || darwin || freebsd || netbsd || openbsd

package rawinput

import (
	"os"
	"sync/atomic"
	"testing"
	"time"
)

// startOnPipe starts a listener that reads a pipe instead of a terminal. Returns the listener, the writer of the pipe and a flag that is
// set when the listener restores the terminal
func startOnPipe(t *testing.T) (*Listener, *os.File, *atomic.Bool) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.Close(); w.Close() })

	var restored atomic.Bool
	return start(int(r.Fd()), func() { restored.Store(true) }), w, &restored
}

func waitWithTimeout(t *testing.T, l *Listener) bool {
	t.Helper()
	result := make(chan bool, 1)
	go func() { result <- l.Wait() }()
	select {
	case pressed := <-result:
		return pressed
	case <-time.After(10 * time.Second):
		t.Fatal("the listener did not end")
		return false
	}
}

func TestListener_AltT_EndsTheListener(t *testing.T) {
	l, w, restored := startOnPipe(t)

	if _, err := w.Write([]byte("abc\x1b")); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("t")); err != nil {
		t.Fatal(err)
	}

	if !waitWithTimeout(t, l) {
		t.Fatal("expected Alt+T to be detected")
	}
	if !restored.Load() {
		t.Fatal("the terminal must be restored before Wait returns")
	}
	l.Stop() // Stopping an ended listener must not block
}

func TestListener_Stop_EndsWithoutInput(t *testing.T) {
	l, _, restored := startOnPipe(t)

	l.Stop()
	if l.Wait() {
		t.Fatal("a stopped listener must not report a key press")
	}
	if !restored.Load() {
		t.Fatal("the terminal must be restored before Stop returns")
	}
	l.Stop()
}

func TestListener_ClosedInput_EndsTheListener(t *testing.T) {
	l, w, _ := startOnPipe(t)

	w.Close()
	if waitWithTimeout(t, l) {
		t.Fatal("a closed input must not report a key press")
	}
}

func TestListen_NotATerminal_ReturnsError(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	if _, err := Listen(r); err != ErrNotTerminal {
		t.Fatalf("Listen(pipe) = %v, want ErrNotTerminal", err)
	}
}
