package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/lua/modules/modtui/tui/theme"

	"github.com/gdamore/tcell/v2"
)

// useConfigDir sets a new empty configuration folder for the test
func useConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv(DirEnv, dir)
	return dir
}

func TestDir(t *testing.T) {
	dir := useConfigDir(t)
	if got, err := Dir(); err != nil || got != dir {
		t.Errorf("Dir() = %q, %v; want %q", got, err, dir)
	}

	t.Setenv(DirEnv, "")
	t.Setenv("XDG_CONFIG_HOME", dir) // Used by `os.UserConfigDir` on Linux and other Unix systems
	if got, err := Dir(); err != nil || !strings.HasSuffix(got, "falcula") {
		t.Errorf("Dir() without %s = %q, %v", DirEnv, got, err)
	}
}

func TestLoad_WithoutFiles_UsesDefaults(t *testing.T) {
	useConfigDir(t)

	config, err := Load()
	if err != nil || config.Theme != theme.DefaultName {
		t.Fatalf("Load() = %+v, %v", config, err)
	}

	got, err := config.LoadTheme()
	if err != nil || got.Border != theme.Default().Border {
		t.Errorf("LoadTheme() = %+v, %v; want the default theme", got, err)
	}
}

func TestLoad_UserTheme(t *testing.T) {
	for _, ext := range []string{".yaml", ".yml"} {
		t.Run(ext, func(t *testing.T) {
			dir := useConfigDir(t)
			testutil.WriteFile(t, filepath.Join(dir, "config"+ext), "theme: mine\n")
			testutil.WriteFile(t, filepath.Join(dir, "themes", "mine"+ext), "colors: { border: '#111111' }\n")

			config, err := Load()
			if err != nil || config.Theme != "mine" {
				t.Fatalf("Load() = %+v, %v", config, err)
			}

			got, err := config.LoadTheme()
			if err != nil {
				t.Fatal(err)
			}
			if got.Border != tcell.NewHexColor(0x111111) || got.Foreground != theme.Default().Foreground {
				t.Errorf("theme = border %v, foreground %v", got.Border, got.Foreground)
			}
		})
	}
}

func TestLoad_UserThemeOverridesBuiltin(t *testing.T) {
	dir := useConfigDir(t)
	testutil.WriteFile(t, filepath.Join(dir, "themes", "soft.yaml"), "colors: { border: '#111111' }\n")

	config, _ := Load()
	got, err := config.LoadTheme()
	if err != nil || got.Border != tcell.NewHexColor(0x111111) {
		t.Errorf("LoadTheme() = %+v, %v; want the user soft theme", got, err)
	}
}

func TestLoad_Errors(t *testing.T) {
	dir := useConfigDir(t)

	testutil.WriteFile(t, filepath.Join(dir, "config.yaml"), "them: soft\n")
	if config, err := Load(); err == nil || config == nil || config.Theme != theme.DefaultName {
		t.Errorf("unknown key: Load() = %+v, %v; want an error and the default configuration", config, err)
	}

	testutil.WriteFile(t, filepath.Join(dir, "config.yaml"), "theme: missing\n")
	config, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := config.LoadTheme(); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Errorf("unknown theme: LoadTheme() error = %v", err)
	}

	testutil.WriteFile(t, filepath.Join(dir, "config.yaml"), "theme: broken\n")
	testutil.WriteFile(t, filepath.Join(dir, "themes", "broken.yaml"), "colors: { border: blurple }\n")
	config, _ = Load()
	if _, err := config.LoadTheme(); err == nil || !strings.Contains(err.Error(), "broken.yaml") {
		t.Errorf("invalid theme: LoadTheme() error = %v, want the file path", err)
	}
}

func TestInit(t *testing.T) {
	dir := useConfigDir(t)
	configFile := filepath.Join(dir, "config.yaml")
	themeFile := filepath.Join(dir, "themes", "soft.yaml")

	results, err := Init(false)
	if err != nil || len(results) != 2 || !results[0].Written || !results[1].Written {
		t.Fatalf("Init() = %+v, %v", results, err)
	}

	// The written files are a valid configuration
	config, err := Load()
	if err != nil || config.Theme != theme.DefaultName {
		t.Fatalf("Load() after Init = %+v, %v", config, err)
	}
	if _, err := config.LoadTheme(); err != nil {
		t.Fatal(err)
	}

	// Existing files are kept
	testutil.WriteFile(t, configFile, "theme: mine\n")
	results, err = Init(false)
	if err != nil || results[0].Written || results[1].Written {
		t.Errorf("second Init() = %+v, %v; want nothing written", results, err)
	}
	if data, _ := os.ReadFile(configFile); string(data) != "theme: mine\n" {
		t.Errorf("Init() overwrote the configuration file: %q", data)
	}

	// Unless forced
	if results, err = Init(true); err != nil || !results[0].Written {
		t.Errorf("forced Init() = %+v, %v", results, err)
	}
	if data, _ := os.ReadFile(themeFile); string(data) != string(theme.BuiltinFile(theme.DefaultName)) {
		t.Error("the theme file must be the built-in theme")
	}
}
