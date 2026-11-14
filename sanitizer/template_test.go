package sanitizer

import "testing"

func TestSanitizeTemplate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		indent string
		want   string
	}{
		{"no markers is unchanged", "key: {{ .value }}", "", "key: {{ .value }}"},
		{"prefix marker replaced by spaces", "# !{{ .x }}", "", "   {{ .x }}"},
		{"prefix marker uses tabs if the line starts with a tab", "\t# !{{ .x }}", "", "\t\t\t\t{{ .x }}"},
		{"indent override", "ab!{{ .x }}", "-", "---{{ .x }}"},
		{"suffix marker blanks the rest of the line", "{{ .x }}! # comment", "", "{{ .x }}           "},
		{"marker without template is unchanged", "a ! b", "", "a ! b"},
		{
			"each line is sanitized",
			"# !{{ if .a }}\nvalue\n# !{{ end }}!",
			"",
			"   {{ if .a }}\nvalue\n   {{ end }} ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeTemplate(tt.input, tt.indent); got != tt.want {
				t.Fatalf("SanitizeTemplate(%q, %q) = %q, want %q", tt.input, tt.indent, got, tt.want)
			}
		})
	}
}
