// Package argsview implements the arguments view
package argsview

import (
	"strings"

	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/app"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/help"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/keybinds"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ArgsView is a widget that shows the current command arguments and allows the user to select arguments
type ArgsView struct {
	app                *app.App
	mainFlex           *tview.Flex
	currentArgsPreview *tview.TextView
	availableArgsTable *tview.Table
	keyBindHandler     *keybinds.Handler
	help               *help.HelpWidget

	// Callbacks
	OnExit     func()
	OnSelected func(args []string)
}

func New(app *app.App, help *help.HelpWidget) *ArgsView {
	a := ArgsView{
		app:  app,
		help: help,

		// Callbacks
		OnExit:     func() {},
		OnSelected: func(arg []string) {},
	}

	// Main widget
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	a.mainFlex = mainFlex

	// Current arguments preview
	currentArgsPreview := tview.NewTextView().SetText("Current arguments: ")
	currentArgsPreview.SetBorder(true)
	mainFlex.AddItem(currentArgsPreview, 3, 0, false)
	a.currentArgsPreview = currentArgsPreview

	// Available arguments
	availableArgsTable := tview.NewTable().SetSelectable(true, false)
	availableArgsTable.SetBorder(true)
	availableArgsTable.SetSelectedFunc(func(row, column int) {
		cell := availableArgsTable.GetCell(row, column)

		// INFO(LucasAVasco): Can not call `app.Draw()` inside event handlers if they are invoked in response to a key event (see
		// https://github.com/rivo/tview/wiki/Concurrency). Doing so causes a deadlock
		go a.OnSelected(cell.GetReference().([]string))
	})
	mainFlex.AddItem(availableArgsTable, 0, 3, true)
	a.availableArgsTable = availableArgsTable

	// Key binds
	a.setKeyBinds()

	return &a
}

func (a *ArgsView) setKeyBinds() {
	a.keyBindHandler = keybinds.NewHandler("Arguments view")
	a.keyBindHandler.AddKeyBinds([]*keybinds.KeyBind{
		{
			Key:  tcell.KeyEscape,
			Desc: "Exit",
			Bind: func() { a.OnExit() },
		},
		{
			Rune: 'q',
			Desc: "Exit",
			Bind: func() { a.OnExit() },
		},
		{
			Key:  tcell.KeyEnter,
			Desc: "Run Lua script with the selected arguments",
		},
		{
			Rune: '?',
			Desc: "Show help",
			Bind: func() { a.help.Open(a.keyBindHandler, nil) },
		},
	})
	a.mainFlex.SetInputCapture(a.keyBindHandler.GetInputCaptureFunction())
}

// GetPrimitive returns the primitive of the widget (used to include it in another widget). This function must not be called if outputting
// to the standard output (raw stdout mode)
func (a *ArgsView) GetPrimitive() tview.Primitive {
	return a.mainFlex
}

// SetCurrentArgs sets the current command arguments in the view
func (a *ArgsView) SetCurrentArgs(args []string) {
	a.currentArgsPreview.SetText("Current arguments: " + strings.Join(args, " "))
	a.app.Draw()
}

// SetAvailableArgs sets the available command arguments in the list. The user can select one of them
func (a *ArgsView) SetAvailableArgs(argsList [][]string) {
	a.availableArgsTable.Clear()

	for i, arg := range argsList {
		cellText := strings.Join(arg, " ")
		cell := tview.NewTableCell(cellText).SetReference(arg)
		a.availableArgsTable.SetCell(i, 0, cell)
	}

	a.app.Draw()
}

// FocusAvailableArgs focuses the available arguments list
func (a *ArgsView) FocusAvailableArgs() {
	a.app.SetFocus(a.availableArgsTable)
	a.availableArgsTable.Select(0, 0)
}
