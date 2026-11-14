package theme

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Apply sets the global styles of 'tview'. Must be called before creating the widgets (they copy the global styles when created)
func Apply(t *Theme) {
	tview.Styles = tview.Theme{
		PrimitiveBackgroundColor:    t.Background,
		ContrastBackgroundColor:     t.SelectionBackground,
		MoreContrastBackgroundColor: t.DialogBackground,
		BorderColor:                 t.Border,
		TitleColor:                  t.Title,
		GraphicsColor:               t.Border,
		PrimaryTextColor:            t.Foreground,
		SecondaryTextColor:          t.Muted,
		TertiaryTextColor:           t.Muted,
		InverseTextColor:            t.SelectionForeground,
		ContrastSecondaryTextColor:  t.SelectionForeground,
	}
}

// TextStyle is the style of normal text
func (t *Theme) TextStyle() tcell.Style {
	return tcell.StyleDefault.Foreground(t.Foreground).Background(t.Background)
}

// SelectionStyle is the style of the selected item of lists, trees and tables
func (t *Theme) SelectionStyle() tcell.Style {
	return tcell.StyleDefault.Foreground(t.SelectionForeground).Background(t.SelectionBackground)
}

// FocusBorder highlights the border and the title of the box while it has the focus
func (t *Theme) FocusBorder(box *tview.Box) {
	box.SetBorderColor(t.Border).SetTitleColor(t.Title)
	box.SetFocusFunc(func() {
		box.SetBorderColor(t.BorderFocused).SetTitleColor(t.BorderFocused)
	})
	box.SetBlurFunc(func() {
		box.SetBorderColor(t.Border).SetTitleColor(t.Title)
	})
}

// StyleModal applies the dialog colors to a modal
func (t *Theme) StyleModal(modal *tview.Modal) {
	modal.SetBackgroundColor(t.DialogBackground).
		SetTextColor(t.DialogForeground).
		SetButtonStyle(tcell.StyleDefault.Foreground(t.ButtonForeground).Background(t.ButtonBackground)).
		SetButtonActivatedStyle(tcell.StyleDefault.Foreground(t.ButtonActiveForeground).Background(t.ButtonActiveBackground))
	modal.SetBorderColor(t.BorderFocused).SetTitleColor(t.BorderFocused)
}
