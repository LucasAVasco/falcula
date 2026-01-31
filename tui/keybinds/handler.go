// Package keybinds implements the key bind interface and the key bind handler
package keybinds

import (
	"github.com/gdamore/tcell/v2"
)

// Handler represents a key bind handler. It can hold key binds and generate a InputCaptureFunction that can be used to bind its key binds
// to a widget
type Handler struct {
	name          string // The name of the handler, should be compatible with the widget name
	keyBinds      []*KeyBind
	runeKeyBinds  map[rune]*KeyBind
	tcellKeyBinds map[tcell.Key]*KeyBind

	// Other key bind handlers that can be considered children of this. This is used by the help widget to show all key binds, including of
	// the children widgets. The `GetInputCaptureFunction()` will not use the key binds of the children handlers
	children []*Handler
}

// NewHandler returns a new handler. The name should be compatible with the widget name
func NewHandler(name string) *Handler {
	return &Handler{
		name:          name,
		keyBinds:      make([]*KeyBind, 0),
		runeKeyBinds:  make(map[rune]*KeyBind),
		tcellKeyBinds: make(map[tcell.Key]*KeyBind),
		children:      make([]*Handler, 0),
	}
}

func (h *Handler) GetName() string {
	return h.name
}

func (h *Handler) GetKeyBinds() []*KeyBind {
	return h.keyBinds
}

// AddKeyBinds adds key binds to the handler
func (h *Handler) AddKeyBinds(keyBinds []*KeyBind) {
	h.keyBinds = append(h.keyBinds, keyBinds...)

	for _, keyBind := range keyBinds {
		if keyBind.Rune != 0 {
			h.runeKeyBinds[keyBind.Rune] = keyBind
		} else if keyBind.Key != tcell.Key(0) {
			h.tcellKeyBinds[keyBind.Key] = keyBind
		} else {
			panic("The key bind must implement either a `rune` or a `tcell.Key`")
		}
	}
}

// GetInputCaptureFunction returns a function that can be used to bind the key binds of the handler to a widget
func (h *Handler) GetInputCaptureFunction() func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		var matchedKeyBind *KeyBind

		// Gets the key bind that matches the event
		if keyBind, ok := h.runeKeyBinds[event.Rune()]; ok {
			matchedKeyBind = keyBind
		} else if keyBind, ok := h.tcellKeyBinds[event.Key()]; ok {
			matchedKeyBind = keyBind
		} else {
			return event
		}

		// If the `Bind()` method is not provided, it is ignored. The user may want to bind a key only to add a description to it
		if matchedKeyBind.Bind == nil {
			return event
		}

		// Executes the key bind
		if matchedKeyBind.Async {
			go matchedKeyBind.Bind()
		} else {
			matchedKeyBind.Bind()
		}

		return nil
	}
}

// AddChildHandler adds a child handler. Used only by the help widget, does not affect the generated InputCaptureFunction
func (h *Handler) AddChildHandler(handler *Handler) {
	h.children = append(h.children, handler)
}

// GetChildHandlers returns the child handlers
func (h *Handler) GetChildHandlers() []*Handler {
	return h.children
}
