package modtui

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/LucasAVasco/falcula/lua/modules/modtui/rawinput"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui"
)

// Mode is how falcula shows the services to the user
type Mode int

const (
	ModeNone      Mode = iota // The text user interface is not used (e.g. it was never shown or the user quit)
	ModeTui                   // The text user interface is visible
	ModeRawOutput             // The text user interface is hidden and the logs are printed in the terminal. Alt+T shows it again
)

// session holds the text user interface and the current mode.
//
// NOTE(LucasAVasco): executing a script will create a new instance of the text user interface module, but the text user interface should be
// preserved throughout the execution of the script. This is why the session is a global variable
type session struct {
	mutex       sync.Mutex
	tui         *tui.Tui
	mode        Mode
	listener    *rawinput.Listener // Waits for Alt+T in raw output mode (nil if not listening)
	stopSignals func()             // Stops handling Ctrl+C in raw output mode (nil if not handling)
}

var current session

// CurrentMode returns how falcula is showing the services to the user
func CurrentMode() Mode {
	current.mutex.Lock()
	defer current.mutex.Unlock()

	return current.mode
}

// getTui returns the text user interface instance (nil if it was not created)
func getTui() *tui.Tui {
	current.mutex.Lock()
	defer current.mutex.Unlock()

	return current.tui
}

// stopRawOutputWithoutLock stops listening for Alt+T and handling Ctrl+C. Restores the terminal. The caller must hold the mutex
func (s *session) stopRawOutputWithoutLock() {
	if s.listener != nil {
		s.listener.Stop()
		s.listener = nil
	}

	if s.stopSignals != nil {
		s.stopSignals()
		s.stopSignals = nil
	}
}

// showTui shows the text user interface (creating it with `newTui` if it does not exist)
func showTui(newTui func() *tui.Tui) error {
	current.mutex.Lock()
	current.stopRawOutputWithoutLock() // The text user interface needs the terminal
	if current.tui == nil {
		current.tui = newTui()
	}
	current.mode = ModeTui
	t := current.tui
	current.mutex.Unlock()

	return t.Show()
}

// enterRawOutput hides the text user interface and prints the logs in the terminal. The services keep running. Alt+T shows the text user
// interface again and Ctrl+C quits
func enterRawOutput() {
	t := getTui()
	if t == nil {
		return
	}

	// The terminal must be restored by the text user interface before listening for keys
	t.Hide()
	t.WaitForHide()

	current.mutex.Lock()
	defer current.mutex.Unlock()

	current.mode = ModeRawOutput
	current.stopSignals = handleInterrupt(func() { go quit() })

	listener, err := rawinput.Listen(os.Stdin)
	if err != nil {
		fmt.Println("falcula: raw output mode - press Ctrl+C to quit")
		return
	}

	current.listener = listener
	fmt.Println("falcula: raw output mode - press Alt+T to open the text user interface, Ctrl+C to quit")
	go showTuiOnAltT(listener)
}

// showTuiOnAltT shows the text user interface when the listener detects Alt+T
func showTuiOnAltT(listener *rawinput.Listener) {
	if !listener.Wait() {
		return // Stopped
	}

	current.mutex.Lock()
	if current.listener != listener { // The user quit (or the text user interface was shown) at the same time
		current.mutex.Unlock()
		return
	}
	current.listener = nil // Already ended
	t := current.tui
	current.mutex.Unlock()

	err := showTui(func() *tui.Tui { return t })
	if err != nil {
		fmt.Fprintf(os.Stderr, "falcula: error showing the text user interface: %v\n", err)
	}
}

// quit leaves the text user interface and the raw output mode (falcula stops waiting for the user)
func quit() {
	current.mutex.Lock()
	current.stopRawOutputWithoutLock()
	current.mode = ModeNone
	t := current.tui
	current.mutex.Unlock()

	if t != nil {
		t.Hide()
	}
}

// handleInterrupt calls the function when the user presses Ctrl+C (SIGINT), instead of killing falcula. Returns a function that restores
// the default behavior
func handleInterrupt(onInterrupt func()) (stop func()) {
	signals := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(signals, os.Interrupt)

	go func() {
		select {
		case <-signals:
			onInterrupt()
		case <-done:
		}
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			signal.Stop(signals)
			close(done)
		})
	}
}

// HideTui hides the text user interface (falcula stops waiting for the user)
func HideTui() {
	quit()
}

// ClosePersistentTui closes (deletes) the persistent text user interface and restores the terminal
func ClosePersistentTui() {
	quit()

	current.mutex.Lock()
	t := current.tui
	current.tui = nil
	current.mutex.Unlock()

	if t != nil {
		t.Close()
	}
}
