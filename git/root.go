package git

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/LucasAVasco/falcula/filesystem"
)

// ErrNoRepository is returned when no git repository is found
var ErrNoRepository = errors.New("no git repository found")

// GetRepositoryRoot returns the root of the git repository that contains the given path.
func GetRepositoryRoot(path string) (string, error) {
	// Ensure path is an absolute path
	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %w", err)
	}

	// Ensure path is a directory
	isDir, err := filesystem.FileIsDir(path)
	if err != nil {
		return "", fmt.Errorf("error checking if path is a directory: %w", err)
	}
	if !isDir {
		return filepath.Dir(path), nil
	}

	// Traverse up the directory tree until we find a '.git' file/directory
	for {
		exists, err := filesystem.FileExists(filepath.Join(path, ".git"))
		if err != nil {
			return "", fmt.Errorf("error checking if '%s/.git' exists: %w", path, err)
		}
		if exists {
			return path, nil
		}
		path = filepath.Dir(path)
		if path == "/" {
			return "", ErrNoRepository
		}
	}
}
