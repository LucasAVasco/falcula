package project

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
)

func TestSplitCommandLine(t *testing.T) {
	tests := []struct {
		line string
		want []string
	}{
		{"dev", []string{"dev"}},
		{"  prod   --verbose  ", []string{"prod", "--verbose"}},
		{`deploy "my server"`, []string{"deploy", "my server"}},
		{`'single $quoted \n' x`, []string{`single $quoted \n`, "x"}},
		{`"a \"b\" \\ \n"`, []string{`a "b" \ \n`}},
		{`with\ space`, []string{"with space"}},
		{`--tag="a b"c`, []string{"--tag=a bc"}},
		{`"" ''`, []string{"", ""}},
		{"", nil},
		{"ação 'ñ'", []string{"ação", "ñ"}},
	}
	for _, tt := range tests {
		got, err := SplitCommandLine(tt.line)
		if err != nil || !slices.Equal(got, tt.want) {
			t.Errorf("SplitCommandLine(%q) = %q, %v; want %q", tt.line, got, err, tt.want)
		}
	}

	for _, line := range []string{`"unclosed`, `'unclosed`, `trailing\`} {
		if _, err := SplitCommandLine(line); err == nil {
			t.Errorf("SplitCommandLine(%q) must fail", line)
		}
	}
}

func loadTaskArgs(t *testing.T, args string) (*Task, error) {
	t.Helper()
	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "falcula.yaml"), "tasks:\n  main:\n    lua: 'print(1)'\n    args:\n"+args)

	p, err := LoadProject(dir)
	if err != nil {
		return nil, err
	}
	return p.Tasks["main"], nil
}

func TestTaskArgs_Parse(t *testing.T) {
	task, err := loadTaskArgs(t, `
      - dev
      - prod --verbose
      - deploy "my server"
      - [build, "--tag=a b"]
`)
	if err != nil {
		t.Fatal(err)
	}

	want := [][]string{{"dev"}, {"prod", "--verbose"}, {"deploy", "my server"}, {"build", "--tag=a b"}}
	if got := task.AvailableArgs(); !slices.EqualFunc(got, want, slices.Equal) {
		t.Errorf("AvailableArgs() = %q, want %q", got, want)
	}
}

func TestTaskArgs_Errors(t *testing.T) {
	tests := map[string]string{
		"empty list":     "      - []\n",
		"empty string":   "      - ''\n",
		"unclosed quote": "      - 'deploy \"x'\n",
		"map":            "      - { a: b }\n",
		"nested list":    "      - [[a]]\n",
	}
	for name, args := range tests {
		_, err := loadTaskArgs(t, args)
		if err == nil {
			t.Errorf("%s: loading must fail", name)
		} else if !strings.Contains(err.Error(), "main") {
			t.Errorf("%s: error must name the task: %v", name, err)
		}
	}
}
