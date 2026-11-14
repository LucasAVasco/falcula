package git

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestGetRepositoryRoot(t *testing.T) {
	repo := t.TempDir()
	nested := filepath.Join(repo, "a", "b")
	file := filepath.Join(nested, "file.txt")
	for _, dir := range []string{filepath.Join(repo, ".git"), nested} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{repo, nested, file} {
		root, err := GetRepositoryRoot(path)
		if err != nil {
			t.Fatalf("GetRepositoryRoot(%q): %v", path, err)
		}
		if root != repo {
			t.Errorf("GetRepositoryRoot(%q) = %q, want %q", path, root, repo)
		}
	}
}

func TestGetRepositoryRoot_OutsideRepository_ReturnsErrNoRepository(t *testing.T) {
	dir := t.TempDir()
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), ".git")); err == nil {
		t.Skip("the temporary directory is inside a git repository")
	}

	_, err := GetRepositoryRoot(dir)
	if !errors.Is(err, ErrNoRepository) {
		t.Fatalf("expected ErrNoRepository, got %v", err)
	}
}

func TestGetRepositoryRoot_MissingPath_ReturnsError(t *testing.T) {
	if _, err := GetRepositoryRoot(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected an error")
	}
}
