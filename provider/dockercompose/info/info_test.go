package info

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
)

const composeFile = `
services:
  web:
    image: example/web
    build: { context: . }
  worker:
    build: { context: ./worker }
  db:
    image: postgres
`

func newInfo(t *testing.T) (*DockerComposeInfo, string) {
	t.Helper()
	path := testutil.WriteFile(t, filepath.Join(t.TempDir(), "compose.yaml"), composeFile)
	return NewComposeInfo(path), path
}

func TestDefaultBuildServices_AreTheBuildableServices(t *testing.T) {
	info, path := newInfo(t)
	if info.GetComposeFilePath() != path {
		t.Errorf("GetComposeFilePath = %q", info.GetComposeFilePath())
	}

	services, err := info.GetDefaultBuildServices()
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(services)
	if !slices.Equal(services, []string{"web", "worker"}) {
		t.Fatalf("GetDefaultBuildServices = %v", services)
	}
}

func TestDefaultPushImages(t *testing.T) {
	info, _ := newInfo(t)
	images, err := info.GetDefaultPushImages()
	if err != nil || !slices.Equal(images, []string{"example/web"}) {
		t.Fatalf("buildable services with an image = %v, %v", images, err)
	}

	custom, _ := newInfo(t)
	custom.AddDefaultPushImage("a")
	custom.AddDefaultPushImages([]string{"b", "c"})
	images, err = custom.GetDefaultPushImages()
	if err != nil || !slices.Equal(images, []string{"a", "b", "c"}) {
		t.Fatalf("custom push images = %v, %v", images, err)
	}
}

func TestGetServiceImage(t *testing.T) {
	info, _ := newInfo(t)

	if image, err := info.GetServiceImage("db"); err != nil || image != "postgres" {
		t.Errorf("db image = %q, %v", image, err)
	}
	if image, err := info.GetServiceImage("worker"); err != nil || image != "" {
		t.Errorf("a service without image must return an empty image, got %q, %v", image, err)
	}
	if _, err := info.GetServiceImage("missing"); err == nil {
		t.Error("an unknown service must return an error")
	}
}

func TestGetComposeFile_IsCachedForAShortTime(t *testing.T) {
	info, path := newInfo(t)
	first, err := info.GetComposeFile()
	if err != nil {
		t.Fatal(err)
	}

	testutil.WriteFile(t, path, "services: {}\n")

	second, err := info.GetComposeFile()
	if err != nil {
		t.Fatal(err)
	}
	if second != first || len(second.Services) != 3 {
		t.Fatal("the compose file must not be parsed again right after being read")
	}
}

func TestGetComposeFile_InvalidFile_ReturnsError(t *testing.T) {
	info := NewComposeInfo(filepath.Join(t.TempDir(), "missing.yaml"))
	if _, err := info.GetDefaultBuildServices(); err == nil {
		t.Error("GetDefaultBuildServices must fail")
	}
	if _, err := info.GetServiceImage("x"); err == nil {
		t.Error("GetServiceImage must fail")
	}
}
