// Package multiplexer implements a log multiplexer
package multiplexer

import (
	"sync"

	"github.com/fatih/color"
)

// Callback is the callback function that is called when data is written to the multiplexer
type Callback func(writer *Client, b []byte) (int, error)

// Multiplexer is the log multiplexer. It generates clients that send logs to the log multiplexer
type Multiplexer struct {
	callback     Callback   // Called when a log is written
	mutex        sync.Mutex // Ensures only one log is written at a time
	nextClientId uint       // Client ID. Auto-incremented
}

func New(callback Callback) *Multiplexer {
	m := Multiplexer{
		mutex:        sync.Mutex{},
		callback:     callback,
		nextClientId: 1,
	}

	return &m
}

// Close the multiplexer. Does not close the log file. The user must close the log file manually.
func (m *Multiplexer) Close() error {
	return nil
}

// NewClient generates a new client. The arguments of this functions are metadata that the client can access. The level has not a specific
// format, but you should avoid using special characters and spaces (good examples: 'stdout', 'stderr', 'info', 'error', etc.).
func (m *Multiplexer) NewClient(name string, level string, color *color.Color) *Client {
	w := Client{
		multi: m,
		name:  name,
		level: level,
		clr:   color,
		id:    m.nextClientId,
	}

	// Updates the client ID
	m.nextClientId++

	return &w
}
