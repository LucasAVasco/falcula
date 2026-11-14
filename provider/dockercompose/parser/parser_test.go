package parser

import (
	"path/filepath"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
)

func TestParseFile(t *testing.T) {
	file := testutil.WriteFile(t, filepath.Join(t.TempDir(), "compose.yaml"), `
services:
  web:
    image: example/web
    build:
      context: ./web
      dockerfile: Dockerfile.dev
  db:
    image: postgres
`)

	parsed, err := ParseFile(file)
	if err != nil {
		t.Fatal(err)
	}

	web, db := parsed.Services["web"], parsed.Services["db"]
	if *web.Image != "example/web" || *web.Build.Context != "./web" || *web.Build.Dockerfile != "Dockerfile.dev" {
		t.Errorf("web = %+v", web)
	}
	if *db.Image != "postgres" || db.Build != nil {
		t.Errorf("db = %+v", db)
	}
}

func TestParseFile_Errors(t *testing.T) {
	dir := t.TempDir()
	invalid := testutil.WriteFile(t, filepath.Join(dir, "invalid.yaml"), "services: [")

	for _, file := range []string{invalid, filepath.Join(dir, "missing.yaml")} {
		if _, err := ParseFile(file); err == nil {
			t.Errorf("ParseFile(%q) must fail", file)
		}
	}
}
