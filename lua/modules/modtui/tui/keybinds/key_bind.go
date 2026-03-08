package keybinds

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// KeyBind represents a key bind manager by the handler
type KeyBind struct {
	Key   tcell.Key // Key that will trigger the key bind
	Rune  rune      // Rune that will trigger the key bind
	Desc  string    // Description of the key bind
	Async bool      // If true, the key bind will be executed asynchronously (in a goroutine)
	Bind  func()    // Function that will be executed when the key bind is triggered
}

// GetName returns a pretty name for the key bind
func (k *KeyBind) GetName() string {
	if k.Rune != 0 {
		return string(k.Rune)
	} else if k.Key != tcell.Key(0) {
		return "<" + tcell.KeyNames[k.Key] + ">"
	}

	panic(fmt.Sprintf("The key bind must implement either a `rune` or a `tcell.Key`, got: %#v", k))
}
