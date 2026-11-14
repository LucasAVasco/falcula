package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWildcardAndRemoveEvents(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	events := make(chan WatcherOperation, 10)
	w, err := NewWatcher(func(f string, op WatcherOperation) error {
		if f == file {
			events <- op
		}
		return nil
	}, &WatcherOpts{Wildcard: true})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if err := w.AddFile(filepath.Join(dir, "*.txt")); err != nil {
		t.Fatalf("adding a wildcard path: %v", err)
	}

	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}

	select {
	case op := <-events:
		if !op.Has(WatcherOperationRemove) {
			t.Fatalf("expected a remove event, got %v", op)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no event received")
	}

	// The watcher must still be running after the remove event
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := w.Wait(); err != nil {
		t.Fatalf("watcher stopped with an error: %v", err)
	}
	if err := w.Wait(); err != nil { // Must not block when called again
		t.Fatal(err)
	}
}
