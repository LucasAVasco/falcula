// Package rawinput listens for key presses in the terminal while the text user interface is hidden (raw output mode)
package rawinput

import (
	"errors"
	"os"
	"sync"

	"golang.org/x/term"
)

var (
	// ErrNotTerminal is returned when the input is not a terminal (e.g. it is redirected from a file)
	ErrNotTerminal = errors.New("the input is not a terminal")

	// ErrUnsupported is returned when the platform does not support listening for key presses
	ErrUnsupported = errors.New("listening for key presses is not supported on this platform")
)

// Listener waits for the Alt+T key in the terminal. While it is listening, the terminal does not echo the typed keys and sends them without
// waiting for Enter. The output and the signals (e.g. Ctrl+C) work as usual
type Listener struct {
	stop     chan struct{}
	stopOnce sync.Once
	done     chan struct{}
	pressed  bool
}

// Listen starts listening for the Alt+T key in the terminal of the input file (usually `os.Stdin`)
func Listen(input *os.File) (*Listener, error) {
	fd := int(input.Fd())
	if !term.IsTerminal(fd) {
		return nil, ErrNotTerminal
	}

	restore, err := makeCbreak(fd)
	if err != nil {
		return nil, err
	}

	return start(fd, restore), nil
}

// start starts reading the file descriptor in a goroutine. The `restore` function is called after reading ends (before `Wait` and `Stop`
// return)
func start(fd int, restore func()) *Listener {
	l := &Listener{
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}

	go func() {
		defer close(l.done)
		defer restore()

		l.pressed = waitForAltT(fd, l.stop)
	}()

	return l
}

// Wait waits until the listener ends. Returns true if Alt+T was pressed, false if the listener was stopped or the input was closed. The
// terminal is restored before it returns
func (l *Listener) Wait() bool {
	<-l.done
	return l.pressed
}

// Stop stops listening and waits for the terminal to be restored. Can be called multiple times and from any goroutine
func (l *Listener) Stop() {
	l.stopOnce.Do(func() { close(l.stop) })
	<-l.done
}
