package ignore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v6"
)

func TestFilter_IsPathIgnored(t *testing.T) {
	repo := t.TempDir()
	if _, err := git.PlainInit(repo, false); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("*.log\nbuild/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	filter, err := NewFilter(repo)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path    string
		isDir   bool
		ignored bool
	}{
		{"app.log", false, true},
		{"src/app.log", false, true},
		{filepath.Join(repo, "debug.log"), false, true},
		{"build", true, true},
		{".git", true, true},
		{"main.go", false, false},
		{filepath.Join(repo, "src", "main.go"), false, false},
		{"build", false, false}, // Only directories named 'build' are ignored
	}

	for _, tt := range tests {
		got, err := filter.IsPathIgnored(tt.path, tt.isDir)
		if err != nil {
			t.Fatalf("IsPathIgnored(%q): %v", tt.path, err)
		}
		if got != tt.ignored {
			t.Errorf("IsPathIgnored(%q, isDir=%v) = %v, want %v", tt.path, tt.isDir, got, tt.ignored)
		}
	}
}

func TestNewFilter_NotARepository_ReturnsError(t *testing.T) {
	if _, err := NewFilter(t.TempDir()); err == nil {
		t.Fatal("expected an error")
	}
}
