// Package multiplexer implements a log multiplexer
package multiplexer

import (
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
)

// Callback is the callback function that is called when data is written to the multiplexer
type Callback func(client *Client, b []byte) (int, error)

// Multiplexer is the log multiplexer. It generates clients that send logs to the log multiplexer
type Multiplexer struct {
	callback     Callback      // Called when a log is written
	mutex        sync.Mutex    // Ensures only one log is written at a time
	nextClientId atomic.Uint64 // Next client ID. Auto-incremented
}

func New(callback Callback) *Multiplexer {
	m := Multiplexer{
		callback: callback,
	}
	m.nextClientId.Store(1)

	return &m
}

// NewClient generates a new client. The arguments of this functions are metadata that the client can access. The level has not a specific
// format, but you should avoid using special characters and spaces (good examples: 'stdout', 'stderr', 'info', 'error', etc.).
func (m *Multiplexer) NewClient(name string, level string, color *color.Color) *Client {
	w := Client{
		multi: m,
		name:  name,
		level: level,
		clr:   color,
		id:    uint(m.nextClientId.Add(1) - 1),
	}

	return &w
}
