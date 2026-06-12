package ignore

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/format/gitignore"
)

// Filter is a gitignore filter. It checks if a path is ignored by the version control system.
type Filter struct {
	repositoryPath string
	matcher        gitignore.Matcher
}

// NewFilter creates a new gitignore filter for the given repository.
func NewFilter(repositoryPath string) (*Filter, error) {
	repositoryPath, err := filepath.Abs(repositoryPath)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path of repository: %w", err)
	}

	// Open repository and get worktree
	repo, err := git.PlainOpen(repositoryPath)
	if err != nil {
		return nil, fmt.Errorf("error opening repository: %w", err)
	}
	defer repo.Close()

	workTree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("error getting worktree: %w", err)
	}
	workTreeFilesystem := workTree.Filesystem()

	// Get ignore patterns
	gitPatterns := gitignore.ParsePattern(".git", nil)

	systemPatterns, err := gitignore.LoadSystemPatterns(workTreeFilesystem)
	if err != nil {
		return nil, fmt.Errorf("error compiling system ignore files: %w", err)
	}

	globalPatterns, err := gitignore.LoadGlobalPatterns(workTreeFilesystem)
	if err != nil {
		return nil, fmt.Errorf("error compiling global ignore files: %w", err)
	}

	repositoryPatterns, err := gitignore.ReadPatterns(workTreeFilesystem, nil)
	if err != nil {
		return nil, fmt.Errorf("error compiling repository ignore files: %w", err)
	}

	patterns := slices.Concat([]gitignore.Pattern{gitPatterns}, systemPatterns, globalPatterns, repositoryPatterns)

	// Create matcher
	matcher := gitignore.NewMatcher(patterns)

	return &Filter{
		repositoryPath: repositoryPath,
		matcher:        matcher,
	}, nil
}

// IsPathIgnored checks if an path is ignored by the version control system. Supports both absolute and relative paths. If it is a relative
// path, it is relative to the repository root.
func (f *Filter) IsPathIgnored(path string, isDir bool) (bool, error) {
	if filepath.IsAbs(path) {
		var err error
		path, err = filepath.Rel(f.repositoryPath, path)
		if err != nil {
			return false, fmt.Errorf("error getting relative path: %w", err)
		}
	}

	path = filepath.Clean(path)
	return f.matcher.Match(strings.Split(path, "/"), isDir), nil
}
