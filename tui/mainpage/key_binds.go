package mainpage

import "github.com/LucasAVasco/falcula/tui/keybinds"

func (m *MainPage) setMainPageKeyBinds() {
	m.keyBindsHandler = keybinds.NewHandler("Main page")
	m.keyBindsHandler.AddChildHandler(m.SideBar.GetKeyBindsHandler())

	m.keyBindsHandler.AddKeyBinds([]*keybinds.KeyBind{
		// Exit
		{
			Rune: 'q',
			Desc: "Exit",
			Bind: func() { m.app.Stop() },
		},
		// Reload
		{
			Rune:  'R',
			Desc:  "Restart Lua script",
			Async: true,
			Bind:  func() { m.OnResetLua() },
		},
		// Scroll the logs
		{
			Rune: 'g',
			Desc: "Scroll to the beginning of the logs",
			Bind: func() {
				m.ServiceLogs.ScrollToBeginning()
				m.DebugLogs.ScrollToBeginning()
			},
		},
		{
			Rune: 'G',
			Desc: "Scroll to the end of the logs",
			Bind: func() {
				m.ServiceLogs.ScrollToEnd()
				m.DebugLogs.ScrollToEnd()
			},
		},
		// Other pages
		{
			Rune: 'a',
			Desc: "Open arguments page",
			Bind: func() { m.OnFocusArgs() },
		},
		// Help menu
		{
			Rune: '?',
			Desc: "Open the help menu",
			Bind: func() {
				m.help.Open(m.keyBindsHandler, func() {
					m.SideBar.SetFocus()
				})
			},
		},
	})

	m.mainFlex.SetInputCapture(m.keyBindsHandler.GetInputCaptureFunction())
}
