package app

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func newTestApp() *App {
	a := Extend(tview.NewApplication())
	a.SetScreen(tcell.NewSimulationScreen("UTF-8"))
	a.SetRoot(tview.NewBox(), true)
	return a
}

func waitEnd(t *testing.T, ended <-chan error) {
	t.Helper()
	select {
	case <-ended:
	case <-time.After(10 * time.Second):
		t.Fatal("the application did not stop")
	}
}

// Stopping right after starting used to be ignored (the application kept running)
func TestStop_RightAfterStart_StopsTheApplication(t *testing.T) {
	a := newTestApp()

	ended := make(chan error, 1)
	if !a.Start(func(err error) { ended <- err }) {
		t.Fatal("Start must start a stopped application")
	}
	a.Stop()

	waitEnd(t, ended)
	if a.IsRunning() {
		t.Fatal("the application must not be running after it stops")
	}
	a.Stop() // Stopping a stopped application does nothing
}

func TestStart_AlreadyRunning_ReturnsFalse(t *testing.T) {
	a := newTestApp()

	ended := make(chan error, 1)
	a.Start(func(err error) { ended <- err })
	if a.Start(func(error) { t.Error("the second run must not call its callback") }) {
		t.Fatal("Start must not start a running application again")
	}

	deadline := time.Now().Add(10 * time.Second)
	for !a.IsRunning() {
		if time.Now().After(deadline) {
			t.Fatal("the application did not start")
		}
		time.Sleep(5 * time.Millisecond)
	}
	a.Draw() // Drawing a running application must not block

	a.Stop()
	waitEnd(t, ended)
}
