package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExistsAndFileIsDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "missing")

	tests := []struct {
		path          string
		exists, isDir bool
	}{
		{dir, true, true},
		{file, true, false},
	}
	for _, tt := range tests {
		exists, err := FileExists(tt.path)
		if err != nil || exists != tt.exists {
			t.Errorf("FileExists(%q) = %v, %v; want %v", tt.path, exists, err, tt.exists)
		}
		isDir, err := FileIsDir(tt.path)
		if err != nil || isDir != tt.isDir {
			t.Errorf("FileIsDir(%q) = %v, %v; want %v", tt.path, isDir, err, tt.isDir)
		}
	}

	if exists, err := FileExists(missing); err != nil || exists {
		t.Errorf("FileExists(missing) = %v, %v; want false, nil", exists, err)
	}
	if _, err := FileIsDir(missing); err == nil {
		t.Error("FileIsDir(missing) must return an error")
	}
}
