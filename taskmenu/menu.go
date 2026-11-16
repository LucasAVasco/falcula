package taskmenu

import (
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/theme"
	"github.com/LucasAVasco/falcula/project"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// helpText is shown in the status line
const helpText = "type to filter · enter: run task / open project · backspace: back · esc: quit"

// listPageSize is the number of entries skipped by Page Up and Page Down
const listPageSize = 10

// Options are the options of the menu. All are optional
type Options struct {
	Theme  *theme.Theme // Default: the default theme
	Screen tcell.Screen // Default: the terminal (tests use a simulation screen)
}

// menu is the user interface of the task menu
type menu struct {
	model *Model
	theme *theme.Theme

	app    *tview.Application
	frame  *tview.Flex
	filter *tview.InputField
	list   *tview.List
	status *tview.TextView

	entries []Entry // Entries shown in the list
	result  string  // Selected task identifier. Empty if canceled
}

// Run shows the menu and waits for the user to select a task. Returns the identifier of the task relative to the root project (e.g.
// "project1:subproject2:task"), or an empty string if the user canceled. Only the projects opened by the user are loaded
func Run(root *project.Project, opts *Options) (string, error) {
	if opts == nil {
		opts = &Options{}
	}
	th := opts.Theme
	if th == nil {
		th = theme.Default()
	}

	m := newMenu(NewModel(root), th)
	if opts.Screen != nil {
		m.app.SetScreen(opts.Screen)
	}

	if err := m.app.Run(); err != nil {
		return "", err
	}

	return m.result, nil
}

func newMenu(model *Model, th *theme.Theme) *menu {
	// NOTE(LucasAVasco): the widgets copy the global styles when created, so the theme must be applied first
	theme.Apply(th)

	m := &menu{model: model, theme: th, app: tview.NewApplication()}

	m.filter = tview.NewInputField().
		SetLabel("Filter: ").
		SetLabelColor(th.Title).
		SetFieldBackgroundColor(th.Background).
		SetFieldTextColor(th.Foreground).
		SetChangedFunc(func(string) { m.refresh() })
	m.filter.SetInputCapture(m.onKey)

	m.list = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetMainTextColor(th.Foreground).
		SetSelectedStyle(th.SelectionStyle())

	m.status = tview.NewTextView().SetDynamicColors(true)

	m.frame = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(m.filter, 1, 0, true).
		AddItem(m.list, 0, 1, false).
		AddItem(m.status, 1, 0, false)
	m.frame.SetBorder(true).SetBorderPadding(0, 0, 1, 1)
	m.frame.SetBorderColor(th.BorderFocused).SetTitleColor(th.BorderFocused)

	m.app.SetRoot(m.frame, true).SetFocus(m.filter)
	m.setStatus("")
	m.refresh()

	return m
}

// refresh updates the title and the list with the entries of the current project that match the filter. Selects the first entry
func (m *menu) refresh() {
	title := " Run a task "
	if path := m.model.Path(); path != "" {
		title = " Run a task — " + tview.Escape(path) + " "
	}
	m.frame.SetTitle(title)

	m.entries = m.model.Entries(m.filter.GetText())
	m.list.Clear()
	for _, entry := range m.entries {
		text := tview.Escape(entry.Name)
		if entry.IsProject {
			text += ":  (project)"
		}
		m.list.AddItem(text, "", 0, nil)
	}
	if len(m.entries) == 0 {
		m.list.AddItem("(no tasks)", "", 0, nil)
	}
	m.list.SetCurrentItem(0)
}

// setStatus shows the error in the status line. Shows the help if the message is empty
func (m *menu) setStatus(errorMessage string) {
	if errorMessage == "" {
		m.status.SetText(theme.Colorize(helpText, m.theme.Muted))
		return
	}
	m.status.SetText(theme.Colorize(tview.Escape(errorMessage), m.theme.Error))
}

// moveSelection moves the selected entry of the list (bounded to the list)
func (m *menu) moveSelection(delta int) {
	index := m.list.GetCurrentItem() + delta
	index = max(0, min(index, m.list.GetItemCount()-1))
	m.list.SetCurrentItem(index)
}

// onKey handles the keys of the filter input (it always has the focus)
func (m *menu) onKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyUp:
		m.moveSelection(-1)
	case tcell.KeyDown:
		m.moveSelection(1)
	case tcell.KeyPgUp:
		m.moveSelection(-listPageSize)
	case tcell.KeyPgDn:
		m.moveSelection(listPageSize)
	case tcell.KeyEnter:
		m.selectCurrent()
	case tcell.KeyEscape:
		m.app.Stop()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if m.filter.GetText() != "" {
			return event // Edits the filter
		}
		if m.model.Back() {
			m.setStatus("")
			m.refresh()
		}
	default:
		return event
	}

	return nil
}

// selectCurrent runs the selected task (stops the menu) or opens the selected project
func (m *menu) selectCurrent() {
	index := m.list.GetCurrentItem()
	if index < 0 || index >= len(m.entries) {
		return
	}
	entry := m.entries[index]

	if !entry.IsProject {
		m.result = m.model.TaskID(entry.Name)
		m.app.Stop()
		return
	}

	if err := m.model.Open(entry.Name); err != nil {
		m.setStatus(err.Error())
		return
	}
	m.setStatus("")
	m.filter.SetText("") // Also refreshes the list (changed function)
	m.refresh()
}
