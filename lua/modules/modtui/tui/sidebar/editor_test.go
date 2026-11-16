package sidebar

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/internal/testutil/testservice"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/app"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/theme"
	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/manager"

	"github.com/rivo/tview"
)

func TestEditorCommand_ChoosesTheEditor(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"VISUAL first", map[string]string{"VISUAL": "code --wait", "EDITOR": "nano"}, "code --wait"},
		{"EDITOR", map[string]string{"EDITOR": "nano"}, "nano"},
		{"default", nil, defaultEditor},
	}
	for _, tt := range tests {
		name, args := editorCommand(func(key string) string { return tt.env[key] }, []string{"/a"})
		want := []string{"-c", tt.want + ` "$@"`, "sh", "/a"}
		if name != "sh" || !slices.Equal(args, want) {
			t.Errorf("%s: editorCommand = %s %q, want sh %q", tt.name, name, args, want)
		}
	}
}

func TestEditorCommand_PassesTheFilesUnchanged(t *testing.T) {
	testutil.RequireShell(t)
	dir := t.TempDir()

	// Stub editor (with an argument) that records the arguments it receives, one per line
	record := filepath.Join(dir, "args")
	editor := testutil.WriteFile(t, filepath.Join(dir, "editor"), "#!/bin/sh\nfor arg; do echo \"$arg\"; done > \""+record+"\"\n")
	if err := os.Chmod(editor, 0o755); err != nil {
		t.Fatal(err)
	}

	files := []string{filepath.Join(dir, "with space.yaml"), filepath.Join(dir, "$(touch injected)"), filepath.Join(dir, "a;b")}
	name, args := editorCommand(func(key string) string {
		if key == "EDITOR" {
			return editor + " --flag"
		}
		return ""
	}, files)

	if out, err := exec.Command(name, args...).CombinedOutput(); err != nil {
		t.Fatalf("editor: %v: %s", err, out)
	}

	data, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if want := append([]string{"--flag"}, files...); !slices.Equal(got, want) {
		t.Errorf("editor arguments = %q, want %q", got, want)
	}
	if _, err := os.Stat(filepath.Join(dir, "injected")); err == nil {
		t.Error("a file name was run by the shell")
	}
}

// serviceWithFile is a test service with an associated file
type serviceWithFile struct {
	*testservice.Service
	file string
}

func (s *serviceWithFile) GetFile() string { return s.file }

func TestFilesOfCurrentNode(t *testing.T) {
	s := New(app.Extend(tview.NewApplication()), "", theme.Default()) // The application does not run (`Draw` does nothing)

	newService := func(man *manager.Manager, name, file string) *enhanced.EnhancedService {
		svc := man.AddService(&serviceWithFile{Service: &testservice.Service{Name: name}, file: file}, nil)
		if err := s.AddService(man, svc); err != nil {
			t.Fatal(err)
		}
		return svc
	}

	compose := manager.New("compose")
	process := manager.New("process")
	empty := manager.New("empty")
	for _, man := range []*manager.Manager{compose, process, empty} {
		if err := s.AddManager(man); err != nil {
			t.Fatal(err)
		}
	}
	web := newService(compose, "web", "/compose.yaml")
	newService(compose, "db", "/compose.yaml")
	api := newService(process, "api", "/api/main.go")
	noFile := newService(process, "worker", "")

	selectNode := func(reference any) {
		t.Helper()
		var found *tview.TreeNode
		s.root.Walk(func(node, _ *tview.TreeNode) bool {
			if node.GetReference() == reference {
				found = node
			}
			return found == nil
		})
		if found == nil {
			t.Fatalf("node of %v not found", reference)
		}
		s.tree.SetCurrentNode(found)
	}

	tests := []struct {
		name    string
		node    any // Reference of the node to select. Nil selects the root
		want    []string
		wantErr string
	}{
		{"service", web, []string{"/compose.yaml"}, ""},
		{"other service", api, []string{"/api/main.go"}, ""},
		{"service without file", noFile, nil, "has no associated file"},
		{"manager without duplicates", compose, []string{"/compose.yaml"}, ""},
		{"manager skips services without file", process, []string{"/api/main.go"}, ""},
		{"manager without files", empty, nil, "no service has an associated file"},
		{"root", nil, []string{"/compose.yaml", "/api/main.go"}, ""},
	}
	for _, tt := range tests {
		if tt.node == nil {
			s.tree.SetCurrentNode(s.root)
		} else {
			selectNode(tt.node)
		}

		got, err := s.filesOfCurrentNode()
		if tt.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("%s: error = %v, want %q", tt.name, err, tt.wantErr)
			}
			continue
		}
		if err != nil || !slices.Equal(got, tt.want) {
			t.Errorf("%s: files = %q, %v; want %q", tt.name, got, err, tt.want)
		}
	}
}
