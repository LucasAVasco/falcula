package tui

import (
	"testing"
	"time"

	"github.com/LucasAVasco/falcula/internal/testutil/testruntime"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/theme"

	"github.com/gdamore/tcell/v2"
)

// testTui is a text user interface that draws in a simulation screen
type testTui struct {
	*Tui
	screen tcell.SimulationScreen
	calls  chan string // Records the exit callbacks
}

func newTestTui(t *testing.T) *testTui {
	t.Helper()
	return newTestTuiWithTheme(t, nil)
}

// newTestTuiWithTheme creates a test text user interface with the theme (the default one if nil)
func newTestTuiWithTheme(t *testing.T, th *theme.Theme) *testTui {
	t.Helper()

	tt := &testTui{screen: tcell.NewSimulationScreen("UTF-8"), calls: make(chan string, 10)}
	tt.Tui = New(&Config{
		Runtime:     testruntime.New(t, nil, nil),
		Theme:       th,
		Screen:      tt.screen,
		OnRawOutput: func() { tt.calls <- "raw output" },
		OnQuit:      func() { tt.calls <- "quit" },
	})

	if err := tt.Show(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		tt.Hide()
		tt.WaitForHide()
	})

	// Waits until the application is running (it handles the injected keys)
	deadline := time.Now().Add(10 * time.Second)
	for !tt.app.IsRunning() {
		if time.Now().After(deadline) {
			t.Fatal("the application did not start")
		}
		time.Sleep(5 * time.Millisecond)
	}

	return tt
}

// frontPage returns the name of the page in front (read in the application goroutine)
func (tt *testTui) frontPage(t *testing.T) string {
	t.Helper()
	name := make(chan string, 1)
	tt.app.QueueUpdate(func() {
		front, _ := tt.pages.GetFrontPage()
		name <- front
	})
	select {
	case n := <-name:
		return n
	case <-time.After(10 * time.Second):
		t.Fatal("the application did not process the update")
		return ""
	}
}

// waitForPage waits until the page is in front. NOTE: the injected keys and the updates use different queues, so it polls
func (tt *testTui) waitForPage(t *testing.T, want string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		page := tt.frontPage(t)
		if page == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("front page = %q, want %q", page, want)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func (tt *testTui) press(key tcell.Key, r rune) {
	tt.screen.InjectKey(key, r, tcell.ModNone)
}

// expectCall waits for an exit callback and checks it
func (tt *testTui) expectCall(t *testing.T, want string) {
	t.Helper()
	select {
	case got := <-tt.calls:
		if got != want {
			t.Fatalf("callback = %q, want %q", got, want)
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("the %q callback was not called", want)
	}
}

func (tt *testTui) expectNoCall(t *testing.T) {
	t.Helper()
	select {
	case got := <-tt.calls:
		t.Fatalf("unexpected callback %q", got)
	default:
	}
}

func TestExitDialog_QKey_OpensTheDialog(t *testing.T) {
	tt := newTestTui(t)
	tt.waitForPage(t, mainPageName)

	tt.press(tcell.KeyRune, 'q')
	tt.waitForPage(t, exitPageName)
	tt.expectNoCall(t)
}

func TestExitDialog_CtrlC_OpensTheDialogInsteadOfClosing(t *testing.T) {
	tt := newTestTui(t)

	tt.press(tcell.KeyCtrlC, 0)
	tt.waitForPage(t, exitPageName)
	if !tt.app.IsRunning() {
		t.Fatal("Ctrl+C must not close the application")
	}
}

func TestExitDialog_RawOutputButton(t *testing.T) {
	tt := newTestTui(t)

	tt.press(tcell.KeyRune, 'q')
	tt.waitForPage(t, exitPageName)
	tt.press(tcell.KeyEnter, 0) // The first button is focused

	tt.expectCall(t, "raw output")
	tt.waitForPage(t, mainPageName)
}

func TestExitDialog_QuitButton(t *testing.T) {
	tt := newTestTui(t)

	tt.press(tcell.KeyRune, 'q')
	tt.waitForPage(t, exitPageName)
	tt.press(tcell.KeyTab, 0)
	tt.press(tcell.KeyEnter, 0)

	tt.expectCall(t, "quit")
}

func TestExitDialog_CancelKeepsTheTui(t *testing.T) {
	for name, keys := range map[string][]tcell.Key{
		"cancel button": {tcell.KeyTab, tcell.KeyTab, tcell.KeyEnter},
		"escape":        {tcell.KeyEscape},
	} {
		t.Run(name, func(t *testing.T) {
			tt := newTestTui(t)

			tt.press(tcell.KeyRune, 'q')
			tt.waitForPage(t, exitPageName)
			for _, key := range keys {
				tt.press(key, 0)
			}

			tt.waitForPage(t, mainPageName)
			tt.expectNoCall(t)

			// The focus returns to the main page: 'q' opens the dialog again
			tt.press(tcell.KeyRune, 'q')
			tt.waitForPage(t, exitPageName)
		})
	}
}
