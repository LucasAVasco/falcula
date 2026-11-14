package tui

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/LucasAVasco/falcula/internal/testutil/testservice"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/theme"
	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/manager"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// testTheme returns a theme with distinct colors (easy to find in the screen)
func testTheme(t *testing.T) *theme.Theme {
	t.Helper()

	files := map[string]string{"test": `
colors:
  border: "#101010"
  border_focused: "#202020"
  button_background: "#303030"
  button_active_background: "#404040"
status:
  none: "#505050"
`}
	th, err := theme.Resolve("test", func(name string) (*theme.File, error) {
		if data, ok := files[name]; ok {
			return theme.ParseFile([]byte(data))
		}
		return theme.BuiltinLookup(name)
	})
	if err != nil {
		t.Fatal(err)
	}
	return th
}

// waitForCell waits until a cell of the screen that starts the text matches the condition. Returns false on timeout
func (tt *testTui) waitForCell(t *testing.T, text string, matches func(style tcell.Style) bool) bool {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		tt.app.QueueUpdateDraw(func() {})

		cells, width, _ := tt.screen.GetContents()
		var row strings.Builder
		for i, cell := range cells {
			if len(cell.Runes) > 0 {
				row.WriteRune(cell.Runes[0])
			} else {
				row.WriteRune(' ')
			}

			if (i+1)%width == 0 {
				if index := strings.Index(row.String(), text); index >= 0 {
					col := utf8.RuneCountInString(row.String()[:index]) // Column (the borders are multi-byte runes)
					y := i / width
					_, _, style, _ := tt.screen.GetContent(col, y)
					if matches(style) {
						return true
					}
				}
				row.Reset()
			}
		}

		time.Sleep(10 * time.Millisecond)
	}

	return false
}

func foreground(color tcell.Color) func(tcell.Style) bool {
	return func(style tcell.Style) bool {
		fg, _, _ := style.Decompose()
		return fg == color
	}
}

func background(color tcell.Color) func(tcell.Style) bool {
	return func(style tcell.Style) bool {
		_, bg, _ := style.Decompose()
		return bg == color
	}
}

func TestTheme_Borders(t *testing.T) {
	saved := tview.Styles
	t.Cleanup(func() { tview.Styles = saved })

	th := testTheme(t)
	tt := newTestTuiWithTheme(t, th)

	// The sidebar has the focus, the debug logs do not
	if !tt.waitForCell(t, "Service managers", foreground(th.BorderFocused)) {
		t.Error("the title of the focused panel must use the focused border color")
	}
	if !tt.waitForCell(t, "Debug logs", foreground(th.Title)) {
		t.Error("the title of the other panels must use the title color")
	}
}

func TestTheme_ServiceStatus(t *testing.T) {
	saved := tview.Styles
	t.Cleanup(func() { tview.Styles = saved })

	th := testTheme(t)
	tt := newTestTuiWithTheme(t, th)

	man := manager.New("man")
	svc := enhanced.NewEnhancedService(&testservice.Service{Name: "svc[x]"}, nil)
	if err := tt.AddManagerToSidebar(man); err != nil {
		t.Fatal(err)
	}
	if err := tt.AddServiceToSidebar(man, svc); err != nil {
		t.Fatal(err)
	}

	if !tt.waitForCell(t, "None", foreground(th.StatusColor(svc.GetStatus()))) {
		t.Error("the service status must use the status color")
	}
	if !tt.waitForCell(t, "svc[x]", func(tcell.Style) bool { return true }) {
		t.Error("the service name must be escaped (square brackets are not color tags)")
	}
}

func TestTheme_ExitDialog(t *testing.T) {
	saved := tview.Styles
	t.Cleanup(func() { tview.Styles = saved })

	th := testTheme(t)
	tt := newTestTuiWithTheme(t, th)

	tt.screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	tt.waitForPage(t, exitPageName)

	if !tt.waitForCell(t, exitButtonRawOutput, background(th.ButtonActiveBackground)) {
		t.Error("the focused button must use the active button color")
	}
	if !tt.waitForCell(t, exitButtonQuit, background(th.ButtonBackground)) {
		t.Error("the other buttons must use the button color")
	}
}
