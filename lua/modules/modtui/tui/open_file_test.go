package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/internal/testutil/testservice"
	"github.com/LucasAVasco/falcula/service/manager"

	"github.com/gdamore/tcell/v2"
)

// serviceWithFile is a test service with an associated file
type serviceWithFile struct {
	*testservice.Service
	file string
}

func (s *serviceWithFile) GetFile() string { return s.file }

// addServices adds a manager with the services (file of each service) to the sidebar and selects the first service
func (tt *testTui) addServices(t *testing.T, files ...string) {
	t.Helper()

	man := manager.New("man")
	if err := tt.AddManagerToSidebar(man); err != nil {
		t.Fatal(err)
	}
	for i, file := range files {
		svc := man.AddService(&serviceWithFile{Service: &testservice.Service{Name: "svc" + string(rune('a'+i))}, file: file}, nil)
		if err := tt.AddServiceToSidebar(man, svc); err != nil {
			t.Fatal(err)
		}
	}

	// Root -> manager -> first service
	tt.screen.InjectKey(tcell.KeyDown, 0, tcell.ModNone)
	tt.screen.InjectKey(tcell.KeyDown, 0, tcell.ModNone)
	if !tt.waitForCell(t, "svca", func(style tcell.Style) bool { return style == tt.config.Theme.SelectionStyle() }) {
		t.Fatal("the first service was not selected")
	}
}

func TestOpenFileKey_OpensTheServiceFileInTheEditor(t *testing.T) {
	testutil.RequireShell(t)
	dir := t.TempDir()

	record := filepath.Join(dir, "opened")
	editor := testutil.WriteFile(t, filepath.Join(dir, "editor"), "#!/bin/sh\necho \"$@\" > \""+record+"\"\n")
	if err := os.Chmod(editor, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", editor)

	tt := newTestTui(t)
	tt.addServices(t, "/project/compose.yaml")

	tt.screen.InjectKey(tcell.KeyRune, 'o', tcell.ModNone)

	deadline := time.Now().Add(10 * time.Second)
	for {
		data, err := os.ReadFile(record)
		if err == nil && strings.TrimSpace(string(data)) == "/project/compose.yaml" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("the editor did not open the service file (recorded %q, %v)", data, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// The text user interface is back after the editor ends
	if !tt.waitForCell(t, "Service managers", func(tcell.Style) bool { return true }) {
		t.Error("the text user interface was not restored after the editor ended")
	}
}

func TestOpenFileKey_ServiceWithoutFile_ReportsTheError(t *testing.T) {
	tt := newTestTui(t)
	tt.addServices(t, "")

	tt.screen.InjectKey(tcell.KeyRune, 'o', tcell.ModNone)
	if !tt.waitForCell(t, "error opening the associated", func(tcell.Style) bool { return true }) {
		t.Error("the error must be shown in the debug logs")
	}
}
