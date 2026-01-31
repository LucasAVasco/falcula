// Package help implements the help widget (show available key binds)
package help

import (
	"github.com/LucasAVasco/falcula/tui/keybinds"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// HelpWidget is a widget that shows the available key binds
type HelpWidget struct {
	app   *tview.Application
	index int

	// Widgets

	frame *tview.Frame
	flex  *tview.Flex
	table *tview.Table

	// Callbacks

	OnOpen func()
	OnExit func()
	onExit func() // Added by the `Open` method
}

func New(app *tview.Application) *HelpWidget {
	h := HelpWidget{
		app:    app,
		OnOpen: func() {},
		OnExit: func() {},
	}

	// Flex
	h.flex = tview.NewFlex()
	h.flex.SetBorder(true)
	h.flex.SetTitle("Help")

	// Frame
	h.frame = tview.NewFrame(h.flex)
	h.frame.SetBorderPadding(5, 5, 5, 5)

	// Table with the key binds
	h.table = tview.NewTable()
	h.flex.AddItem(h.table, 0, 1, true)

	// Key binds (exit with `q` or `ESC`)
	h.flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' || event.Key() == tcell.KeyEscape {
			h.OnExit()
			if h.onExit != nil {
				h.onExit()
			}
			return nil
		}

		return event
	})

	return &h
}

// GetPrimitive returns the primitive of the widget (used to include it in another widget). This function must not be called if outputting
// to the standard output (raw stdout mode)
func (h *HelpWidget) GetPrimitive() tview.Primitive {
	return h.frame
}

// addHandlerKeyBinds adds the key binds description of the handler to the list of key binds (table)
func (h *HelpWidget) addHandlerKeyBinds(handler *keybinds.Handler) {
	// Key binds
	h.table.SetCellSimple(h.index, 0, handler.GetName())
	h.index++

	for _, keyBind := range handler.GetKeyBinds() {
		h.table.SetCell(h.index, 0, tview.NewTableCell(keyBind.GetName()))
		h.table.SetCell(h.index, 1, tview.NewTableCell(keyBind.Desc))
		h.index++
	}

	// Child handlers
	for _, childHandler := range handler.GetChildHandlers() {
		h.addHandlerKeyBinds(childHandler)
	}
}

// Open the help widget
func (h *HelpWidget) Open(handler *keybinds.Handler, onExit func()) {
	h.onExit = onExit
	h.OnOpen()

	// Removes the old key binds
	h.index = 0
	h.table.Clear()

	// New key binds
	h.addHandlerKeyBinds(handler)
}
