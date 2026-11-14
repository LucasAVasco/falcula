// Package app extends the text user interface application with additional features
package app

import (
	"sync"

	"github.com/rivo/tview"
)

// App is the extended text user interface application. It adds thread safety, support to check if it is running and a reliable `Stop`
type App struct {
	*tview.Application

	mutex   sync.Mutex
	running bool          // Between `Start` and the end of the run
	started chan struct{} // Closed when the event loop of the current run starts handling events
	ended   chan struct{} // Closed when the current run ends
}

// Extend extends the text user interface application with thread safety and support to check if it is running
func Extend(app *tview.Application) *App {
	return &App{
		Application: app,
	}
}

// Start runs the application in a new goroutine (non-blocking). The `onEnd` function is called with the error of the run when it stops.
// Returns false (and does not call `onEnd`) if the application is already running
func (a *App) Start(onEnd func(err error)) bool {
	a.mutex.Lock()
	if a.running {
		a.mutex.Unlock()
		return false
	}

	a.running = true
	started, ended := make(chan struct{}), make(chan struct{})
	a.started, a.ended = started, ended
	a.mutex.Unlock()

	// The first update is executed when the event loop starts. NOTE(LucasAVasco): `QueueUpdate` returns only after the update is executed,
	// so it runs in its own goroutine (it never returns if the application fails to start)
	go a.Application.QueueUpdate(func() { close(started) })

	go func() {
		err := a.Application.Run()

		a.mutex.Lock()
		a.running = false
		a.mutex.Unlock()
		close(ended)

		onEnd(err)
	}()

	return true
}

// IsRunning checks if the application is running (its event loop is handling events)
func (a *App) IsRunning() bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.running {
		return false
	}

	select {
	case <-a.started:
		return true
	default:
		return false
	}
}

// Stop stops the application. Equivalent to `tview.Application.Stop()`, but also works if called while the application is starting
func (a *App) Stop() {
	a.mutex.Lock()
	if !a.running {
		a.mutex.Unlock()
		return
	}
	started, ended := a.started, a.ended
	a.mutex.Unlock()

	// NOTE(LucasAVasco): stopping before the event loop starts does nothing, the application would keep running
	select {
	case <-started:
	case <-ended:
		return
	}

	a.Application.Stop()
}

// Draw updates the display of the application. This method is thread-safe
func (a *App) Draw() {
	// NOTE(LucasAVasco): the Draw method may cause some child widgets to also call the Draw method, it must not hold the mutex
	if a.IsRunning() {
		a.Application.Draw()
	}
}
