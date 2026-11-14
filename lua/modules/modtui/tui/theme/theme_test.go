package theme

import (
	"errors"
	"strings"
	"testing"

	"github.com/LucasAVasco/falcula/service/status"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestParseColor(t *testing.T) {
	tests := []struct {
		text string
		want tcell.Color
	}{
		{"#8fa6bf", tcell.NewHexColor(0x8fa6bf)},
		{"#8FA6BF", tcell.NewHexColor(0x8fa6bf)},
		{"slategray", tcell.ColorSlateGray},
		{" SlateGray ", tcell.ColorSlateGray},
		{"default", tcell.ColorDefault},
	}
	for _, tt := range tests {
		got, err := ParseColor(tt.text)
		if err != nil || tcell.Color(got) != tt.want {
			t.Errorf("ParseColor(%q) = %v, %v; want %v", tt.text, got, err, tt.want)
		}
	}

	for _, text := range []string{"", "#12345", "#gggggg", "not-a-color"} {
		if _, err := ParseColor(text); err == nil {
			t.Errorf("ParseColor(%q) must fail", text)
		}
	}
}

func TestParseFile_Errors(t *testing.T) {
	tests := map[string]string{
		"unknown top key":    "colours: {}",
		"unknown color key":  "colors: { backgrund: '#000000' }",
		"unknown status key": "status: { runing: '#000000' }",
		"invalid color":      "colors: { border: 'blurple' }",
		"color not a string": "colors: { border: [1, 2] }",
	}
	for name, data := range tests {
		if _, err := ParseFile([]byte(data)); err == nil {
			t.Errorf("%s: ParseFile(%q) must fail", name, data)
		}
	}
}

func TestBuiltinTheme_SetsAllKeys(t *testing.T) {
	file, err := BuiltinLookup(DefaultName)
	if err != nil {
		t.Fatal(err)
	}
	for key := range colorKeys {
		if _, ok := file.Colors[key]; !ok {
			t.Errorf("the default theme does not set 'colors.%s'", key)
		}
	}
	for key := range statusKeys {
		if _, ok := file.Status[key]; !ok {
			t.Errorf("the default theme does not set 'status.%s'", key)
		}
	}

	if BuiltinFile(DefaultName) == nil || BuiltinFile("missing") != nil {
		t.Error("BuiltinFile returns the embedded files only")
	}
}

// lookupFrom returns a lookup function that uses the files (name -> YAML) and then the built-in themes
func lookupFrom(t *testing.T, files map[string]string) func(string) (*File, error) {
	return func(name string) (*File, error) {
		if data, ok := files[name]; ok {
			file, err := ParseFile([]byte(data))
			if err != nil {
				t.Fatalf("theme %s: %v", name, err)
			}
			return file, nil
		}
		return BuiltinLookup(name)
	}
}

func TestResolve_InheritsMissingKeys(t *testing.T) {
	soft := Default()
	lookup := lookupFrom(t, map[string]string{
		"base":  "colors: { border: '#111111', title: '#222222' }\nstatus: { running: '#333333' }",
		"child": "extends: base\ncolors: { title: '#444444' }",
	})

	got, err := Resolve("child", lookup)
	if err != nil {
		t.Fatal(err)
	}

	if got.Title != tcell.NewHexColor(0x444444) {
		t.Errorf("title = %v, want the child color", got.Title)
	}
	if got.Border != tcell.NewHexColor(0x111111) {
		t.Errorf("border = %v, want the parent color", got.Border)
	}
	if got.StatusColor(status.Running) != tcell.NewHexColor(0x333333) {
		t.Errorf("running = %v, want the parent color", got.StatusColor(status.Running))
	}
	if got.Foreground != soft.Foreground || got.StatusColor(status.Error) != soft.StatusColor(status.Error) {
		t.Error("keys that no theme sets must come from the default theme")
	}
}

func TestResolve_PartialOverrideOfTheDefaultTheme(t *testing.T) {
	// A user theme named like the default one (e.g. edited after `falcula config init`) may be partial
	got, err := Resolve(DefaultName, lookupFrom(t, map[string]string{DefaultName: "colors: { border: '#111111' }"}))
	if err != nil {
		t.Fatal(err)
	}
	if got.Border != tcell.NewHexColor(0x111111) || got.Foreground != Default().Foreground {
		t.Errorf("partial default theme = border %v, foreground %v", got.Border, got.Foreground)
	}
}

func TestResolve_Errors(t *testing.T) {
	lookup := lookupFrom(t, map[string]string{
		"a":       "extends: b",
		"b":       "extends: a",
		"orphan":  "extends: missing",
		"selfish": "extends: selfish",
	})

	for _, name := range []string{"a", "selfish"} {
		if _, err := Resolve(name, lookup); err == nil || !strings.Contains(err.Error(), "cycle") {
			t.Errorf("Resolve(%q) = %v, want a cycle error", name, err)
		}
	}
	for _, name := range []string{"missing", "orphan"} {
		if _, err := Resolve(name, lookup); !errors.Is(err, ErrNotFound) {
			t.Errorf("Resolve(%q) = %v, want ErrNotFound", name, err)
		}
	}
}

func TestStatusColor_AllStatuses(t *testing.T) {
	theme := Default()
	for _, s := range statusKeys {
		if theme.StatusColor(s) == tcell.ColorDefault {
			t.Errorf("status %s has no color", s)
		}
	}
	if theme.StatusColor(status.Status(100)) != theme.Muted {
		t.Error("unknown statuses must use the muted color")
	}
}

func TestApply_SetsTviewStyles(t *testing.T) {
	saved := tview.Styles
	t.Cleanup(func() { tview.Styles = saved })

	theme := Default()
	Apply(theme)

	if tview.Styles.PrimitiveBackgroundColor != theme.Background || tview.Styles.BorderColor != theme.Border ||
		tview.Styles.PrimaryTextColor != theme.Foreground {
		t.Errorf("tview.Styles = %+v", tview.Styles)
	}
}

func TestColorize(t *testing.T) {
	color := tcell.NewHexColor(0x8fa6bf)
	if got := Colorize("x", color); got != "[#8fa6bf]x[-]" {
		t.Errorf("Colorize = %q", got)
	}
	if got := Colorize("x", tcell.ColorDefault); got != "[-]x[-]" {
		t.Errorf("Colorize with the default color = %q", got)
	}
	if got := ColorizeANSI("x", color); got != "\x1b[38;2;143;166;191mx\x1b[0m" {
		t.Errorf("ColorizeANSI = %q", got)
	}
	if got := ColorizeANSI("x", tcell.ColorDefault); got != "x" {
		t.Errorf("ColorizeANSI with the default color = %q", got)
	}
}
