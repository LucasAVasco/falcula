// Package mainpage contains the main page widget of the application
package mainpage

import (
	"github.com/LucasAVasco/falcula/tui/help"
	"github.com/LucasAVasco/falcula/tui/keybinds"
	"github.com/LucasAVasco/falcula/tui/logpreview"
	"github.com/LucasAVasco/falcula/tui/sidebar"

	"github.com/rivo/tview"
)

// MainPage is the main page widget of the application
type MainPage struct {
	app             *tview.Application
	keyBindsHandler *keybinds.Handler
	logFilePath     string
	help            *help.HelpWidget

	// User interface

	mainFlex    *tview.Flex
	SideBar     *sidebar.Sidebar
	ServiceLogs *logpreview.Preview
	DebugLogs   *logpreview.Preview

	// Callbacks

	OnResetLua  func()
	OnFocusArgs func()
}

func New(app *tview.Application, logFilePath string, help *help.HelpWidget) *MainPage {
	m := MainPage{
		app:         app,
		logFilePath: logFilePath,
		help:        help,

		// Callbacks
		OnResetLua:  func() {},
		OnFocusArgs: func() {},
	}

	// Raw stdout mode
	if app == nil {
		m.SideBar = sidebar.New(nil, logFilePath)
		m.ServiceLogs = logpreview.New(nil, "", 0)
		m.DebugLogs = logpreview.New(nil, "", 0)
		return &m
	}

	// Main layout
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn)
	m.mainFlex = mainFlex

	// Side bar
	m.SideBar = sidebar.New(app, logFilePath)
	m.SideBar.OnError = func(err error) {
		m.DebugLogs.Append(err)
	}
	mainFlex.AddItem(m.SideBar.GetPrimitive(), 0, 1, false)

	// Flex with the logs preview (service logs and debug logs)
	logsFlex := tview.NewFlex().
		SetDirection(tview.FlexRow)
	mainFlex.AddItem(logsFlex, 0, 3, false)

	// Service logs
	m.ServiceLogs = logpreview.New(app, "Services logs", 100)
	logsFlex.AddItem(m.ServiceLogs.GetPrimitive(), 0, 2, false)

	// Debug logs
	m.DebugLogs = logpreview.New(app, "Debug logs", 100)
	logsFlex.AddItem(m.DebugLogs.GetPrimitive(), 0, 1, false)

	// Key binds
	m.setMainPageKeyBinds()

	return &m
}

// GetPrimitive returns the primitive of the widget (used to include it in another widget). This function must not be called if outputting
// to the standard output (raw stdout mode)
func (m *MainPage) GetPrimitive() tview.Primitive {
	return m.mainFlex
}
