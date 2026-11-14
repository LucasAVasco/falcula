// Package theme implements the color themes of the text user interface. Themes are YAML files (see `soft.yaml`, the default theme)
package theme

import (
	_ "embed"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/LucasAVasco/falcula/service/status"

	"github.com/gdamore/tcell/v2"
	"github.com/goccy/go-yaml"
)

// DefaultName is the name of the default theme
const DefaultName = "soft"

//go:embed soft.yaml
var softFile []byte

// builtinFiles are the themes embedded in falcula (name -> YAML file)
var builtinFiles = map[string][]byte{
	DefaultName: softFile,
}

// ErrNotFound is returned by a lookup function when the theme does not exist
var ErrNotFound = errors.New("theme not found")

// Theme is a resolved theme: all colors are set
type Theme struct {
	Background             tcell.Color
	Foreground             tcell.Color
	Muted                  tcell.Color
	Border                 tcell.Color
	BorderFocused          tcell.Color
	Title                  tcell.Color
	SelectionForeground    tcell.Color
	SelectionBackground    tcell.Color
	DialogBackground       tcell.Color
	DialogForeground       tcell.Color
	ButtonForeground       tcell.Color
	ButtonBackground       tcell.Color
	ButtonActiveForeground tcell.Color
	ButtonActiveBackground tcell.Color
	Error                  tcell.Color

	statuses map[status.Status]tcell.Color
}

// colorKeys maps the keys of the `colors` section of a theme file to the fields of the theme
var colorKeys = map[string]func(t *Theme) *tcell.Color{
	"background":               func(t *Theme) *tcell.Color { return &t.Background },
	"foreground":               func(t *Theme) *tcell.Color { return &t.Foreground },
	"muted":                    func(t *Theme) *tcell.Color { return &t.Muted },
	"border":                   func(t *Theme) *tcell.Color { return &t.Border },
	"border_focused":           func(t *Theme) *tcell.Color { return &t.BorderFocused },
	"title":                    func(t *Theme) *tcell.Color { return &t.Title },
	"selection_foreground":     func(t *Theme) *tcell.Color { return &t.SelectionForeground },
	"selection_background":     func(t *Theme) *tcell.Color { return &t.SelectionBackground },
	"dialog_background":        func(t *Theme) *tcell.Color { return &t.DialogBackground },
	"dialog_foreground":        func(t *Theme) *tcell.Color { return &t.DialogForeground },
	"button_foreground":        func(t *Theme) *tcell.Color { return &t.ButtonForeground },
	"button_background":        func(t *Theme) *tcell.Color { return &t.ButtonBackground },
	"button_active_foreground": func(t *Theme) *tcell.Color { return &t.ButtonActiveForeground },
	"button_active_background": func(t *Theme) *tcell.Color { return &t.ButtonActiveBackground },
	"error":                    func(t *Theme) *tcell.Color { return &t.Error },
}

// statusKeys maps the keys of the `status` section of a theme file to the service statuses
var statusKeys = map[string]status.Status{
	"none":             status.None,
	"preparing":        status.Preparing,
	"ready":            status.Ready,
	"aborting_prepare": status.AbortingPrepare,
	"prepare_aborted":  status.PrepareAborted,
	"running":          status.Running,
	"ended":            status.Ended,
	"stopping":         status.Stopping,
	"stopped":          status.Stopped,
	"disabled":         status.Disabled,
	"error":            status.Error,
}

// File is the content of a theme file. All keys are optional: the missing ones are inherited from the `Extends` theme
type File struct {
	Extends string           `yaml:"extends"` // Default: the default theme. Ignored by the default theme itself
	Colors  map[string]Color `yaml:"colors"`
	Status  map[string]Color `yaml:"status"`
}

// ParseFile parses the content of a theme file. Unknown keys are errors
func ParseFile(data []byte) (*File, error) {
	var file File
	if err := yaml.UnmarshalWithOptions(data, &file, yaml.DisallowUnknownField()); err != nil {
		return nil, err
	}

	for key := range file.Colors {
		if _, ok := colorKeys[key]; !ok {
			return nil, fmt.Errorf("unknown key 'colors.%s' (valid keys: %s)", key, strings.Join(slices.Sorted(maps.Keys(colorKeys)), ", "))
		}
	}
	for key := range file.Status {
		if _, ok := statusKeys[key]; !ok {
			return nil, fmt.Errorf("unknown key 'status.%s' (valid keys: %s)", key, strings.Join(slices.Sorted(maps.Keys(statusKeys)), ", "))
		}
	}

	return &file, nil
}

// BuiltinFile returns the content of a theme embedded in falcula. Returns nil if there is no such theme
func BuiltinFile(name string) []byte {
	return builtinFiles[name]
}

// BuiltinLookup is a lookup function (see `Resolve`) for the themes embedded in falcula
func BuiltinLookup(name string) (*File, error) {
	data, ok := builtinFiles[name]
	if !ok {
		return nil, ErrNotFound
	}
	return ParseFile(data)
}

// Default returns the default theme
func Default() *Theme {
	t, err := Resolve(DefaultName, BuiltinLookup)
	if err != nil {
		panic(fmt.Sprintf("the built-in theme '%s' is invalid: %v", DefaultName, err)) // Never happens: covered by the tests
	}
	return t
}

// Resolve loads the theme with the name and the themes it extends. The `lookup` function returns the file of a theme (or an error that
// wraps `ErrNotFound`). The keys that no theme of the chain sets come from the built-in default theme
func Resolve(name string, lookup func(name string) (*File, error)) (*Theme, error) {
	// Chain of files, from the requested theme to the default one (it ends the chain, its `extends` is ignored)
	var chain []*File
	visited := map[string]bool{}
	for current := name; ; {
		if visited[current] {
			return nil, fmt.Errorf("theme '%s' has an 'extends' cycle (at theme '%s')", name, current)
		}
		visited[current] = true

		file, err := lookup(current)
		if err != nil {
			return nil, fmt.Errorf("error loading theme '%s': %w", current, err)
		}
		chain = append(chain, file)

		if current == DefaultName {
			break
		}
		current = file.Extends
		if current == "" {
			current = DefaultName
		}
	}

	// The built-in default theme is the base of all themes
	base, err := BuiltinLookup(DefaultName)
	if err != nil {
		return nil, fmt.Errorf("error loading the built-in theme '%s': %w", DefaultName, err)
	}
	chain = append(chain, base)

	// Applies the files from the base to the requested theme
	t := &Theme{statuses: map[status.Status]tcell.Color{}}
	for _, file := range slices.Backward(chain) {
		for key, color := range file.Colors {
			*colorKeys[key](t) = tcell.Color(color)
		}
		for key, color := range file.Status {
			t.statuses[statusKeys[key]] = tcell.Color(color)
		}
	}

	return t, nil
}

// StatusColor returns the color of a service status. Unknown statuses use the muted color
func (t *Theme) StatusColor(s status.Status) tcell.Color {
	if color, ok := t.statuses[s]; ok {
		return color
	}
	return t.Muted
}

// Colorize returns the text with tview color tags that show it with the color. The text must already be escaped
func Colorize(text string, color tcell.Color) string {
	return tag(color) + text + "[-]"
}

// ColorizeANSI returns the text with the ANSI escape sequences that show it with the color
func ColorizeANSI(text string, color tcell.Color) string {
	prefix := ansi(color)
	if prefix == "" {
		return text
	}
	return prefix + text + "\x1b[0m"
}
