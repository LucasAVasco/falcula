package modtui

import "github.com/LucasAVasco/falcula/lua/modules/modtui/tui"

// NOTE(LucasAVasco): executing a script will create a new instance of the text user interface module, but the text user interface should be
// preserved throughout the execution of the script. This is why we use a global variable to store the text user interface instance
var serviceTui *tui.Tui

// IsTuiVisible checks if the text user interface is visible
func IsTuiVisible() bool {
	if serviceTui == nil {
		return false
	}

	return serviceTui.IsVisible()
}

// HideTui hides the text user interface.
func HideTui() {
	if serviceTui != nil {
		serviceTui.Hide()
	}
}

// ClosePersistentTui closes (deletes) the persistent text user interface
func ClosePersistentTui() {
	if serviceTui != nil {
		serviceTui.Close()
		serviceTui = nil
	}
}

// WaitForTuiHide waits until the text user interface is hidden
func WaitForTuiHide() {
	if serviceTui == nil {
		return
	}

	serviceTui.WaitForHide()
}
