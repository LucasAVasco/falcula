package theme

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/goccy/go-yaml"
)

// Color is a color of a theme file. Accepts "#rrggbb", a color name (e.g. "slategray") or "default" (the terminal's own color)
type Color tcell.Color

// ParseColor parses a color of a theme file
func ParseColor(text string) (Color, error) {
	name := strings.ToLower(strings.TrimSpace(text))
	if name == "default" {
		return Color(tcell.ColorDefault), nil
	}

	color := tcell.GetColor(name)
	if color == tcell.ColorDefault {
		return 0, fmt.Errorf("invalid color '%s' (expected \"#rrggbb\", a color name or \"default\")", text)
	}

	return Color(color), nil
}

func (c *Color) UnmarshalYAML(data []byte) error {
	var text string
	if err := yaml.Unmarshal(data, &text); err != nil {
		return fmt.Errorf("a color must be a string: %w", err)
	}

	color, err := ParseColor(text)
	if err != nil {
		return err
	}
	*c = color

	return nil
}

// tag returns the tview color tag that sets the foreground to the color (resets it if the color is the default one)
func tag(color tcell.Color) string {
	if color == tcell.ColorDefault {
		return "[-]"
	}
	return fmt.Sprintf("[#%06x]", color.Hex())
}

// ansi returns the ANSI escape sequence that sets the foreground to the color. Empty if the color is the default one
func ansi(color tcell.Color) string {
	if color == tcell.ColorDefault {
		return ""
	}
	r, g, b := color.RGB()
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}
