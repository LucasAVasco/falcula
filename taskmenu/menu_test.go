package taskmenu

import (
	"strings"
	"testing"
	"time"

	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/theme"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// testMenu is a menu that runs in a simulation screen
type testMenu struct {
	*menu
	screen tcell.SimulationScreen
	done   chan error // Receives the result of the run when the menu ends
}

func startMenu(t *testing.T) *testMenu {
	t.Helper()
	saved := tview.Styles
	t.Cleanup(func() { tview.Styles = saved })

	tm := &testMenu{
		menu:   newMenu(NewModel(newTree(t)), theme.Default()),
		screen: tcell.NewSimulationScreen("UTF-8"),
		done:   make(chan error, 1),
	}
	tm.app.SetScreen(tm.screen)
	go func() { tm.done <- tm.app.Run() }()
	t.Cleanup(func() {
		tm.app.Stop()
		select {
		case <-tm.done:
		case <-time.After(10 * time.Second):
		}
	})

	// Waits until the event loop runs
	tm.read(t, func() {})
	return tm
}

// read runs the function in the application goroutine and waits for it
func (tm *testMenu) read(t *testing.T, fn func()) {
	t.Helper()
	executed := make(chan struct{})
	go tm.app.QueueUpdate(func() { fn(); close(executed) })
	select {
	case <-executed:
	case <-time.After(10 * time.Second):
		t.Fatal("the menu did not process the update")
	}
}

// keys injects the keys. Runes are typed, other keys are tcell keys
func (tm *testMenu) keys(keys ...any) {
	for _, key := range keys {
		switch k := key.(type) {
		case rune:
			tm.screen.InjectKey(tcell.KeyRune, k, tcell.ModNone)
		case string:
			for _, r := range k {
				tm.screen.InjectKey(tcell.KeyRune, r, tcell.ModNone)
			}
		case tcell.Key:
			tm.screen.InjectKey(k, 0, tcell.ModNone)
		}
	}
}

// result waits for the menu to end and returns the selected task
func (tm *testMenu) result(t *testing.T) string {
	t.Helper()
	select {
	case err := <-tm.done:
		if err != nil {
			t.Fatal(err)
		}
		tm.done <- nil // The cleanup waits for it
		return tm.menu.result
	case <-time.After(10 * time.Second):
		t.Fatal("the menu did not end")
		return ""
	}
}

// waitFor waits until the condition (evaluated in the application goroutine) is true
func (tm *testMenu) waitFor(t *testing.T, description string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		var ok bool
		tm.read(t, func() { ok = condition() })
		if ok {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for: %s", description)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestMenu_EnterRunsTheFirstTask(t *testing.T) {
	tm := startMenu(t)
	tm.keys(tcell.KeyEnter)
	if got := tm.result(t); got != "Test" {
		t.Errorf("result = %q, want the first task", got)
	}
}

func TestMenu_FilterThenEnter(t *testing.T) {
	tm := startMenu(t)
	tm.keys("bui", tcell.KeyEnter)
	if got := tm.result(t); got != "build" {
		t.Errorf("result = %q, want build", got)
	}
}

func TestMenu_OpensProjectsAndGoesBack(t *testing.T) {
	tm := startMenu(t)

	// Root: Test, build, project1, project2
	tm.keys(tcell.KeyDown, tcell.KeyDown, tcell.KeyEnter)
	tm.waitFor(t, "project1 opened", func() bool { return tm.model.Path() == "project1" })

	// project1: deploy, subproject2, subproject3. Backspace with an empty filter goes back
	tm.keys(tcell.KeyBackspace2)
	tm.waitFor(t, "back to the root", func() bool { return tm.model.Path() == "" })

	tm.keys("project1", tcell.KeyEnter, "sub", tcell.KeyEnter, "my", tcell.KeyEnter)
	if got := tm.result(t); got != "project1:subproject2:my-task" {
		t.Errorf("result = %q", got)
	}
}

func TestMenu_InvalidProject_ShowsTheErrorAndStays(t *testing.T) {
	tm := startMenu(t)

	tm.keys("project2", tcell.KeyEnter)
	tm.waitFor(t, "error in the status line", func() bool {
		return strings.Contains(tm.status.GetText(true), "error opening project 'project2'")
	})
	tm.read(t, func() {
		if tm.model.Path() != "" {
			t.Errorf("Path() = %q, want the root", tm.model.Path())
		}
	})

	// Still usable
	tm.keys(tcell.KeyCtrlU) // Clears the filter (tview input field)
	tm.keys("build", tcell.KeyEnter)
	if got := tm.result(t); got != "build" {
		t.Errorf("result = %q, want build", got)
	}
}

func TestMenu_Cancel(t *testing.T) {
	for name, key := range map[string]tcell.Key{"esc": tcell.KeyEscape, "ctrl+c": tcell.KeyCtrlC} {
		t.Run(name, func(t *testing.T) {
			tm := startMenu(t)
			tm.keys(key)
			if got := tm.result(t); got != "" {
				t.Errorf("result = %q, want empty (canceled)", got)
			}
		})
	}
}
