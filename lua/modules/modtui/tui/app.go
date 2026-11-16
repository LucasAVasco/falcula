package tui

import (
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/app"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/argsview"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/help"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/mainpage"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/theme"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Names of the pages of the application
const (
	mainPageName      = "main"
	argumentsPageName = "arguments"
	helpPageName      = "help"
	exitPageName      = "exit"
)

// Buttons of the exit dialog
const (
	exitButtonRawOutput = "Raw output mode"
	exitButtonQuit      = "Quit falcula"
	exitButtonCancel    = "Cancel"
)

func (t *Tui) focusMainPage() {
	t.pages.SwitchToPage(mainPageName)
	t.app.SetFocus(t.mainPage.SideBar.GetPrimitive())
}

func (t *Tui) focusArgumentsPage() {
	t.pages.SwitchToPage(argumentsPageName)
	t.argsView.FocusAvailableArgs()
}

// openExitDialog shows the dialog that asks how to leave the text user interface. Must be called in the application goroutine
func (t *Tui) openExitDialog() {
	if name, _ := t.pages.GetFrontPage(); name == exitPageName {
		return
	}

	t.focusBeforeExitDialog = t.app.GetFocus()
	t.pages.ShowPage(exitPageName)
	t.app.SetFocus(t.exitDialog)
}

// closeExitDialog hides the exit dialog and restores the focus. Must be called in the application goroutine
func (t *Tui) closeExitDialog() {
	t.pages.HidePage(exitPageName)
	t.app.SetFocus(t.focusBeforeExitDialog)
}

// newExitDialog creates the dialog that asks how to leave the text user interface
func (t *Tui) newExitDialog() *tview.Modal {
	modal := tview.NewModal()
	t.config.Theme.StyleModal(modal)

	return modal.
		SetText("Leave the text user interface?\n\n" +
			"Raw output mode keeps the services running and prints their logs in the terminal (press Alt+T to come back).").
		AddButtons([]string{exitButtonRawOutput, exitButtonQuit, exitButtonCancel}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			t.closeExitDialog()

			// NOTE(LucasAVasco): the callbacks stop the application, so they can not run in the application goroutine (this function)
			switch buttonLabel {
			case exitButtonRawOutput:
				go t.config.OnRawOutput()
			case exitButtonQuit:
				go t.config.OnQuit()
			}
		})
}

// newApp creates a new application if not in raw stdout mode
func (t *Tui) newApp() {
	// NOTE(LucasAVasco): the widgets copy the global styles when created, so the theme must be applied first
	theme.Apply(t.config.Theme)

	// Main application
	t.app = app.Extend(tview.NewApplication())
	if t.config.Screen != nil {
		t.app.SetScreen(t.config.Screen)
	}

	// Ctrl+C asks how to leave the text user interface instead of closing it (default behavior of 'tview')
	t.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			t.openExitDialog()
			return nil
		}
		return event
	})

	// Pages
	t.pages = tview.NewPages()
	t.app.SetRoot(t.pages, true).
		EnableMouse(true)

	// Help page
	t.help = help.New(t.app, t.config.Theme)
	t.help.OnOpen = func() {
		t.pages.ShowPage(helpPageName)
	}
	t.help.OnExit = func() {
		t.pages.HidePage(helpPageName)
	}

	// Main page
	t.mainPage = mainpage.New(t.app, t.config.Runtime.Logger.GetLogFilePath(), t.help, t.config.Theme)
	t.mainPage.OnFocusArgs = t.focusArgumentsPage
	t.mainPage.OnExit = t.openExitDialog
	t.pages.AddPage(mainPageName, t.mainPage.GetPrimitive(), true, false)

	// Arguments page
	t.argsView = argsview.New(t.app, t.help, t.config.Theme)
	t.argsView.OnExit = t.focusMainPage
	t.argsView.OnSelected = func(args []string) {
		// NOTE(LucasAVasco): called in its own goroutine (not the application one), so the widgets are changed with an update. It does not
		// wait for the update (it never runs if the application stops at the same time)
		go t.app.QueueUpdateDraw(t.focusMainPage)
		t.config.OnSelectArgs(args)
	}
	t.pages.AddPage(argumentsPageName, t.argsView.GetPrimitive(), true, false)

	// Adds the help page to the pages
	// NOTE(LucasAVasco): must be the last page in the list to be on top
	t.pages.AddPage(helpPageName, t.help.GetPrimitive(), true, false)

	// Exit dialog (on top of all pages)
	t.exitDialog = t.newExitDialog()
	t.pages.AddPage(exitPageName, t.exitDialog, true, false)

	// Initial focus
	t.focusMainPage()
}
