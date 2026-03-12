package tui

// SetCurrentScriptArgs sets the current command arguments. This value is displayed in the arguments view
func (t *Tui) SetCurrentScriptArgs(args []string) {
	t.argsView.SetCurrentArgs(args)
}

// SetAvailableScriptArgs sets the available command arguments. This value is displayed in the arguments view
func (t *Tui) SetAvailableScriptArgs(args [][]string) {
	t.argsView.SetAvailableArgs(args)
}
