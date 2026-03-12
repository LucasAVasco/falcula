// Package logpreview implements the log preview widget
package logpreview

import (
	"fmt"
	"io"
	"sync"

	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/app"
	"github.com/rivo/tview"
)

// Preview is a widget that shows the logs. Supports both TUI and raw stdout mode
type Preview struct {
	mutex         sync.Mutex
	textView      *tview.TextView
	ansiConverter io.Writer // Convert ANSI escape sequences to the color tags supported by `tview` and write to the text view
}

// Creates a new log preview.
//
// The widget is a text view with the provided title.
//
// You can set the maximum number of lines of the text view with `maxLines`.
func New(app *app.App, title string, maxLines int) *Preview {
	p := Preview{}

	// Text view creation and configuration
	p.textView = tview.NewTextView().SetDynamicColors(true).SetWrap(true).SetScrollable(true).SetMaxLines(maxLines).
		SetChangedFunc(func() {
			app.Draw()
		})
	p.textView.SetBorder(true).SetTitle(title)

	// Starts with the cursor at the end
	p.textView.ScrollToEnd()

	// Writer that automatically convert ANSI escape sequences to the color tags supported by `tview`
	p.ansiConverter = tview.ANSIWriter(p.textView)

	return &p
}

// GetPrimitive returns the primitive of the widget (used to include it in another widget). This function must not be called if outputting
// to the standard output (raw stdout mode)
func (p *Preview) GetPrimitive() tview.Primitive {
	return p.textView
}

// Appends logs to the preview.
//
// This function is thread-safe.
func (p *Preview) Append(logs ...any) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Outputs the logs to the text view
	fmt.Fprint(p.ansiConverter, logs...)
}

// Implements the `io.Writer` interface
func (p *Preview) Write(b []byte) (n int, err error) {
	p.Append(string(b))
	return len(b), nil
}

func (p *Preview) WriteError(errToWrite error) (n int, err error) {
	message := errToWrite.Error()
	p.Append(message, "\n")
	return len(message), nil
}

func (p *Preview) ScrollToBeginning() {
	p.textView.ScrollToBeginning()
}

func (p *Preview) ScrollToEnd() {
	p.textView.ScrollToEnd()
}
