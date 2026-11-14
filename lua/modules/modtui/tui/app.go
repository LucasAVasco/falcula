package tui

import (
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/app"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/argsview"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/help"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/mainpage"

	"github.com/rivo/tview"
)

// Names of the pages of the application
const (
	mainPageName      = "main"
	argumentsPageName = "arguments"
	helpPageName      = "help"
)

func (t *Tui) focusMainPage() {
	t.pages.SwitchToPage(mainPageName)
	t.app.SetFocus(t.mainPage.SideBar.GetPrimitive())
}

func (t *Tui) focusArgumentsPage() {
	t.pages.SwitchToPage(argumentsPageName)
	t.argsView.FocusAvailableArgs()
}

// newApp creates a new application if not in raw stdout mode
func (t *Tui) newApp() {
	// Main application
	t.app = app.Extend(tview.NewApplication())

	// Pages
	t.pages = tview.NewPages()
	t.app.SetRoot(t.pages, true).
		EnableMouse(true)

	// Help page
	t.help = help.New(t.app)
	t.help.OnOpen = func() {
		t.pages.ShowPage(helpPageName)
	}
	t.help.OnExit = func() {
		t.pages.HidePage(helpPageName)
	}

	// Main page
	t.mainPage = mainpage.New(t.app, t.config.Runtime.Logger.GetLogFilePath(), t.help)
	t.mainPage.OnFocusArgs = t.focusArgumentsPage
	t.mainPage.OnExit = t.Close
	t.pages.AddPage(mainPageName, t.mainPage.GetPrimitive(), true, false)

	// Arguments page
	t.argsView = argsview.New(t.app, t.help)
	t.argsView.OnExit = t.focusMainPage
	t.argsView.OnSelected = func(args []string) {
		t.focusMainPage()
		t.config.OnSelectArgs(args)
	}
	t.pages.AddPage(argumentsPageName, t.argsView.GetPrimitive(), true, false)

	// Adds the help page to the pages
	// NOTE(LucasAVasco): must be the last page in the list to be on top
	t.pages.AddPage(helpPageName, t.help.GetPrimitive(), true, false)

	// Initial focus
	t.focusMainPage()
}
