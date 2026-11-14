//go:build linux || darwin || freebsd || netbsd || openbsd

package modtui

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/LucasAVasco/falcula/internal/testutil/testruntime"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui"

	"github.com/gdamore/tcell/v2"
)

// showTestTui shows a text user interface that draws in a simulation screen. The standard input is a pipe (not a terminal), so the raw
// output mode does not listen for Alt+T
func showTestTui(t *testing.T) *tui.Tui {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		ClosePersistentTui()
		os.Stdin = oldStdin
		r.Close()
		w.Close()
	})

	runtime := testruntime.New(t, nil, nil)
	err = showTui(func() *tui.Tui {
		return tui.New(&tui.Config{
			Runtime:     runtime,
			Screen:      tcell.NewSimulationScreen("UTF-8"),
			OnRawOutput: enterRawOutput,
			OnQuit:      quit,
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	if CurrentMode() != ModeTui {
		t.Fatalf("mode = %v, want ModeTui", CurrentMode())
	}

	return getTui()
}

func waitForMode(t *testing.T, want Mode) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for CurrentMode() != want {
		if time.Now().After(deadline) {
			t.Fatalf("mode = %v, want %v", CurrentMode(), want)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestRawOutputMode_HidesTheTuiAndCtrlCQuits(t *testing.T) {
	showTestTui(t)

	enterRawOutput()
	if CurrentMode() != ModeRawOutput {
		t.Fatalf("mode = %v, want ModeRawOutput", CurrentMode())
	}

	// Ctrl+C in raw output mode quits instead of killing falcula
	if err := syscall.Kill(os.Getpid(), syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	waitForMode(t, ModeNone)

	current.mutex.Lock()
	defer current.mutex.Unlock()
	if current.stopSignals != nil || current.listener != nil {
		t.Fatal("quitting must stop handling Ctrl+C and listening for keys")
	}
}

func TestQuit_FromTheTui_EndsTheSession(t *testing.T) {
	ui := showTestTui(t)

	quit()
	if CurrentMode() != ModeNone {
		t.Fatalf("mode = %v, want ModeNone", CurrentMode())
	}

	done := make(chan struct{})
	go func() {
		ui.WaitForHide()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the text user interface was not hidden")
	}
}

func TestClosePersistentTui_FromRawOutputMode(t *testing.T) {
	showTestTui(t)
	enterRawOutput()

	ClosePersistentTui()
	if CurrentMode() != ModeNone || getTui() != nil {
		t.Fatalf("mode = %v, tui = %v; want no session", CurrentMode(), getTui())
	}
}
