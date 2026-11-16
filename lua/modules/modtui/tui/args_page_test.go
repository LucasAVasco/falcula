package tui

import (
	"slices"
	"testing"
	"time"

	"github.com/LucasAVasco/falcula/internal/testutil/testruntime"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// argsTableHasFocus checks if the arguments table has the focus (read in the application goroutine)
func (tt *testTui) argsTableHasFocus(t *testing.T) bool {
	t.Helper()
	focused := make(chan bool, 1)
	go tt.app.QueueUpdate(func() { focused <- tt.argsView.GetPrimitive().HasFocus() })
	select {
	case ok := <-focused:
		return ok
	case <-time.After(10 * time.Second):
		t.Fatal("the application did not process the update")
		return false
	}
}

func TestOpenArgumentsPage_RightAfterShow(t *testing.T) {
	saved := tview.Styles
	t.Cleanup(func() { tview.Styles = saved })

	// Same as `tui.show_args()` in Lua: the page is opened before the event loop runs
	selected := make(chan []string, 1)
	tt := &testTui{screen: tcell.NewSimulationScreen("UTF-8"), calls: make(chan string, 10)}
	tt.Tui = New(&Config{
		Runtime:      testruntime.New(t, nil, nil),
		Screen:       tt.screen,
		OnSelectArgs: func(args []string) { selected <- args },
	})
	tt.config.Runtime.SetScriptAvailableArgs([][]string{{"prod", "--verbose"}, {"dev"}})

	if err := tt.Show(); err != nil {
		t.Fatal(err)
	}
	tt.OpenArgumentsPage()
	t.Cleanup(func() {
		tt.Hide()
		tt.WaitForHide()
	})

	tt.waitForPage(t, argumentsPageName)
	if !tt.argsTableHasFocus(t) {
		t.Error("the arguments table must have the focus")
	}

	// Enter selects the first argument set
	tt.screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	select {
	case args := <-selected:
		if !slices.Equal(args, []string{"prod", "--verbose"}) {
			t.Errorf("selected arguments = %q", args)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("no arguments were selected")
	}
	tt.waitForPage(t, mainPageName)
}

func TestOpenArgumentsPage_WhileVisible(t *testing.T) {
	tt := newTestTui(t)
	tt.waitForPage(t, mainPageName)

	tt.OpenArgumentsPage()
	tt.waitForPage(t, argumentsPageName)
	if !tt.argsTableHasFocus(t) {
		t.Error("the arguments table must have the focus")
	}
}

func TestOpenArgumentsPage_BeforeShow_DoesNothing(t *testing.T) {
	New(&Config{Runtime: testruntime.New(t, nil, nil)}).OpenArgumentsPage() // Must not panic
}
