// Package app extends the text user interface application with additional features
package app

import (
	"sync"
	"time"

	"github.com/rivo/tview"
)

// App is the extended text user interface application. It adds thread safety and support to check if it is running
type App struct {
	*tview.Application
	mutex     sync.Mutex
	isRunning bool
}

// Extend extends the text user interface application with thread safety and support to check if it is running
func Extend(app *tview.Application) *App {
	return &App{
		Application: app,
	}
}

// IsRunning checks if the application is running
func (a *App) IsRunning() bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.isRunning
}

// Run starts the application and blocks until it is stopped. Equivalent to `tview.Application.Run()`, but updates the status (whether the
// app is running or not) in a thread-safe way
func (a *App) Run() error {
	a.mutex.Lock()

	if a.isRunning {
		a.mutex.Unlock()
		return nil
	}

	// Start the app
	var err error
	wg := sync.WaitGroup{}
	wg.Go(func() {
		err = a.Application.Run()
	})

	// Mark the app as running after a delay (ensure the app is actually running)
	time.Sleep(100 * time.Millisecond)
	a.isRunning = true
	a.mutex.Unlock()

	// Wait for the app to stop and return its error
	wg.Wait()
	return err
}

// Stop stops the application. Equivalent to `tview.Application.Stop()`, but updates the status (whether the app is running or not) in a
// thread-safe way
func (a *App) Stop() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.isRunning {
		return
	}

	a.isRunning = false
	a.Application.Stop()
}

// Draw updates the display of the application. This method is thread-safe
func (a *App) Draw() {
	a.mutex.Lock()

	if !a.isRunning {
		a.mutex.Unlock()
		return
	}

	// NOTE(LucasAVasco): the Draw method may cause some child widgets to also call the Draw method, we must unlock before calling it
	a.mutex.Unlock()
	a.Application.Draw()
}
