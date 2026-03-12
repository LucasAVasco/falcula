package tui

import (
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/app"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/argsview"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/help"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/mainpage"

	"github.com/rivo/tview"
)

func (t *Tui) focusMainPage() {
	t.pages.SwitchToPage("main")
	t.app.SetFocus(t.mainPage.SideBar.GetPrimitive())
}

func (t *Tui) focusArgumentsPage() {
	t.pages.SwitchToPage("arguments")
	t.argsView.FocusAvailableArgs()
}

// newApp creates a new application if not in raw stdout mode
func (t *Tui) newApp() error {
	// Main application
	t.app = app.Extend(tview.NewApplication())

	// Pages
	t.pages = tview.NewPages()
	t.app.SetRoot(t.pages, true).
		EnableMouse(true)

	// Help page
	t.help = help.New(t.app)
	t.help.OnOpen = func() {
		t.pages.ShowPage("help")
	}
	t.help.OnExit = func() {
		t.pages.HidePage("help")
	}

	// Main page
	t.mainPage = mainpage.New(t.app, t.config.Runtime.Logger.GetLogFilePath(), t.help)
	t.mainPage.OnFocusArgs = t.focusArgumentsPage
	t.mainPage.OnExit = t.Close
	t.pages.AddPage("main", t.mainPage.GetPrimitive(), true, false)

	// Arguments page
	t.argsView = argsview.New(t.app, t.help)
	t.argsView.OnExit = t.focusMainPage
	t.argsView.OnSelected = func(args []string) {
		t.focusMainPage()
		t.config.OnSelectArgs(args)
	}
	t.pages.AddPage("arguments", t.argsView.GetPrimitive(), true, false)

	// Adds the help page to the pages
	// NOTE(LucasAVasco): must be the last page in the list to be on top
	t.pages.AddPage("help", t.help.GetPrimitive(), true, false)

	// Initial focus
	t.focusMainPage()

	return nil
}
