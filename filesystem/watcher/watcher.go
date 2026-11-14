package watcher

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"

	"github.com/LucasAVasco/falcula/filesystem"
	"github.com/LucasAVasco/falcula/git"
	"github.com/LucasAVasco/falcula/git/ignore"
	"github.com/fsnotify/fsnotify"
)

// WatcherOperation is an operation related to a file system event (file system change).
type WatcherOperation = fsnotify.Op

const (
	WatcherOperationCreate WatcherOperation = fsnotify.Create
	WatcherOperationWrite  WatcherOperation = fsnotify.Write
	WatcherOperationRemove WatcherOperation = fsnotify.Remove
	WatcherOperationRename WatcherOperation = fsnotify.Rename
	WatcherOperationChmod  WatcherOperation = fsnotify.Chmod
)

// WatcherCallback is a function that will be called when an event (file system change) is triggered.
type WatcherCallback = func(file string, operation WatcherOperation) error

// WatcherOpts is a struct that represents the options for the watcher.
type WatcherOpts struct {
	Recursive bool `lua:"recursive"` // Watch directories recursively (default: false).
	Wildcard  bool `lua:"wildcard"`  // Expand wildcards (default: false).
	Vcs       bool `lua:"vcs"`       // Watch version control system files (default: false).
}

// Watcher is a file system Watcher. It is used to watch for file system changes and call a callback function when an event is triggered.
type Watcher struct {
	watcher *fsnotify.Watcher
	opts    *WatcherOpts

	// Files to watch. Shared between the routine that listens for events and the methods that add files
	mutex        sync.Mutex
	vcsFilters   map[string]*ignore.Filter
	filesToWatch map[string]bool
	dirsToWatch  map[string]bool

	// Result of the routine that listens for events
	routineError chan error
	waitOnce     sync.Once
	waitErr      error

	closeOnce sync.Once
	closeErr  error
}

// NewWatcher creates a new file system Watcher.
func NewWatcher(callback WatcherCallback, opts *WatcherOpts) (*Watcher, error) {
	if opts == nil {
		opts = &WatcherOpts{}
	}

	w := Watcher{
		vcsFilters:   make(map[string]*ignore.Filter),
		filesToWatch: make(map[string]bool),
		dirsToWatch:  make(map[string]bool),
		routineError: make(chan error, 1),
		opts:         opts,
	}

	var err error
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creating fsnotify watcher: %w", err)
	}

	go func() {
		w.routineError <- w.listenEvents(callback)
	}()

	return &w, nil
}

// Close closes the watcher. Can be called multiple times
func (w *Watcher) Close() error {
	w.closeOnce.Do(func() {
		w.closeErr = w.watcher.Close()
	})
	return w.closeErr
}

// listenEvents listens for events and handles them accordingly.
func (w *Watcher) listenEvents(callback WatcherCallback) error {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}

			file := filepath.Clean(event.Name)
			shouldNotify, err := w.shouldNotify(file)
			if err != nil {
				return err
			}
			if !shouldNotify {
				continue
			}

			// The file is tracked and the event is not ignored
			err = callback(file, event.Op)
			if err != nil {
				return fmt.Errorf("callback returned error: %w", err)
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}

			return fmt.Errorf("watcher returned error: %w", err)
		}
	}
}

// shouldNotify returns true if the file is tracked by the watcher and is not ignored
func (w *Watcher) shouldNotify(file string) (bool, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Get the file and directory
	isDir, err := filesystem.FileIsDir(file)
	if errors.Is(err, fs.ErrNotExist) {
		// The file was removed or renamed, so it can not be checked anymore. Uses the tracked directories instead
		isDir = w.dirsToWatch[file]
	} else if err != nil {
		return false, fmt.Errorf("error checking if '%s' is a directory: %w", file, err)
	}

	dir := file
	if !isDir {
		dir = filepath.Dir(file)
	}

	// Only notify if the file or directory is tracked
	if !w.dirsToWatch[dir] && !w.filesToWatch[file] {
		return false, nil
	}

	// Filter the event
	isIgnored, err := w.fileIsIgnored(file, isDir)
	if err != nil {
		return false, fmt.Errorf("error checking if file is ignored: %w", err)
	}

	return !isIgnored, nil
}

// fileIsIgnored checks if a file is ignored by the version control system. Returns false if this feature is not enabled. The caller must
// hold the mutex
func (w *Watcher) fileIsIgnored(file string, isDir bool) (bool, error) {
	if !w.opts.Vcs {
		return false, nil
	}

	// Get the filter for the repository
	repositoryRoot, err := git.GetRepositoryRoot(file)
	if err != nil {
		return false, fmt.Errorf("error getting repository root: %w", err)
	}

	filter, ok := w.vcsFilters[repositoryRoot]
	if !ok {
		return false, nil
	}

	// Check if the file is ignored
	isIgnored, err := filter.IsPathIgnored(file, isDir)
	if err != nil {
		return false, fmt.Errorf("error checking if file is ignored: %w", err)
	}

	return isIgnored, nil
}

// AddFiles adds several files (or directories) to the watcher (listens for changes on them).
func (w *Watcher) AddFiles(files []string) error {
	for _, file := range files {
		err := w.AddFile(file)
		if err != nil {
			return fmt.Errorf("error adding file '%s' to watcher: %w", file, err)
		}
	}

	return nil
}

// AddFile adds a file (or directory) to the watcher (listens for changes on it).
func (w *Watcher) AddFile(path string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	path, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("error getting absolute path of '%s': %w", path, err)
	}

	// Filter to apply to this file
	var filter *ignore.Filter
	if w.opts.Vcs {
		repositoryRoot, err := git.GetRepositoryRoot(path)
		if err != nil {
			return fmt.Errorf("error getting repository root: %w", err)
		}

		var ok bool
		filter, ok = w.vcsFilters[repositoryRoot]
		if !ok {
			filter, err = ignore.NewFilter(repositoryRoot)
			if err != nil {
				return fmt.Errorf("error creating ignore filter for repository '%s': %w", repositoryRoot, err)
			}
			w.vcsFilters[repositoryRoot] = filter
		}
	}

	err = w.addAbs(path, filter, w.opts)
	if err != nil {
		return fmt.Errorf("error adding file '%s' to watcher: %w", path, err)
	}

	return nil
}

// addAbs adds a file (or directory) to the watcher (listens for changes on it). Must be called with an absolute path. The caller must hold
// the mutex
func (w *Watcher) addAbs(file string, filter *ignore.Filter, opts *WatcherOpts) error {
	// Expand wildcards
	if opts.Wildcard {
		files, err := filepath.Glob(file)
		if err != nil {
			return fmt.Errorf("error expanding wildcard '%s': %w", file, err)
		}

		opts := *opts
		opts.Wildcard = false
		for _, file := range files {
			err = w.addAbs(file, filter, &opts)
			if err != nil {
				return fmt.Errorf("error adding file '%s' to watcher: %w", file, err)
			}
		}

		return nil
	}

	// Watch directories recursively
	if opts.Recursive {
		opts := *opts
		opts.Recursive = false

		// Adds non-directories without recursion
		isDir, err := filesystem.FileIsDir(file)
		if err != nil {
			return fmt.Errorf("error checking if '%s' is a directory: %w", file, err)
		}
		if !isDir {
			return w.addAbs(file, filter, &opts)
		}

		// Walk directory recursively and add sub-directories
		err = filepath.WalkDir(file, func(file string, d fs.DirEntry, err error) error {
			if err != nil {
				if isUnrecoverableError(err) {
					return fmt.Errorf("error walking directory: %w", err)
				}
				return nil
			}

			// Filter
			if filter != nil {
				isIgnored, err := filter.IsPathIgnored(file, d.IsDir())
				if err != nil {
					return fmt.Errorf("error checking if file is ignored: %w", err)
				}
				if isIgnored {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}

			if d.IsDir() {
				err = w.addAbs(file, filter, &opts)
				if err != nil {
					return fmt.Errorf("error adding file '%s' to watcher: %w", file, err)
				}
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("error walking directory '%s': %w", file, err)
		}

		return nil
	}

	isDir, err := filesystem.FileIsDir(file)
	if err != nil {
		return fmt.Errorf("error checking if file is a directory: %w", err)
	}

	// Filter
	if filter != nil {
		isIgnored, err := filter.IsPathIgnored(file, isDir)
		if err != nil {
			return fmt.Errorf("error checking if file is ignored: %w", err)
		}
		if isIgnored {
			return nil
		}
	}

	// Add file to watch list
	if isDir {
		w.dirsToWatch[file] = true
	} else {
		w.filesToWatch[file] = true

		// Watch the directory instead of the file. The watcher routine will filter the file from the event
		file = filepath.Dir(file)
	}

	// Watch the directory
	err = w.watcher.Add(file)
	if err != nil {
		if isUnrecoverableError(err) {
			return fmt.Errorf("error adding file to watcher: %w", err)
		}
		return nil
	}

	return nil
}

// Wait waits for the watcher to stop or return an error
func (w *Watcher) Wait() error {
	w.waitOnce.Do(func() {
		w.waitErr = <-w.routineError
	})

	if w.waitErr != nil {
		return fmt.Errorf("error watching files: %w", w.waitErr)
	}

	return nil
}

// isUnrecoverableError returns true if the error is unrecoverable (should abort watching files).
func isUnrecoverableError(err error) bool {
	return !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, fs.ErrPermission)
}
