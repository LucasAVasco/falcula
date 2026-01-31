// Package logpreview implements the log preview widget
package logpreview

import (
	"fmt"
	"io"
	"sync"

	"github.com/rivo/tview"
)

// Preview is a widget that shows the logs. Supports both TUI and raw stdout mode
type Preview struct {
	toStdout bool

	mutex         sync.Mutex
	textView      *tview.TextView
	ansiConverter io.Writer // Convert ANSI escape sequences to the color tags supported by `tview` and write to the text view
}

// Creates a new log preview.
//
// The widget is a text view with the provided title. If you do not provide a `app`, this widget will output all the logs to the standard
// output (raw stdout mode). Otherwise, it will output the logs to the text view widget of the provided `app`.
//
// You can set the maximum number of lines of the text view with `maxLines`.
func New(app *tview.Application, title string, maxLines int) *Preview {
	p := Preview{
		toStdout: app == nil,
	}

	// No need to create the text view if outputting to the standard output
	if app == nil {
		return &p
	}

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

	// Outputs the logs to the standard output
	if p.toStdout {
		fmt.Print(logs...)
		return
	}

	// Outputs the logs to the text view
	fmt.Fprint(p.ansiConverter, logs...)
}

// Implements the `io.Writer` interface
func (p *Preview) Write(b []byte) (n int, err error) {
	p.Append(string(b))
	return len(b), nil
}

func (p *Preview) ScrollToBeginning() {
	p.textView.ScrollToBeginning()
}

func (p *Preview) ScrollToEnd() {
	p.textView.ScrollToEnd()
}
