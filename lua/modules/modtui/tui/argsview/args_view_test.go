package argsview

import "testing"

func TestFormatArgs(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{nil, "(no arguments)"},
		{[]string{"dev", "--verbose"}, "dev --verbose"},
		{[]string{"deploy", "my server"}, "deploy 'my server'"},
		{[]string{"deploy", ""}, "deploy ''"},
		{[]string{"it's"}, `'it'\''s'`},
		{[]string{"[red]"}, "[red]"}, // Escaped for 'tview' by the caller
	}
	for _, tt := range tests {
		if got := FormatArgs(tt.args); got != tt.want {
			t.Errorf("FormatArgs(%q) = %q, want %q", tt.args, got, tt.want)
		}
	}
}
