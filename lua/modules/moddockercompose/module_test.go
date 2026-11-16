package moddockercompose

import (
	"path/filepath"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/internal/testutil/testruntime"
	"github.com/LucasAVasco/falcula/service/iface"

	lua "github.com/yuin/gopher-lua"
)

func TestServiceFile_DefaultsToTheComposeFile(t *testing.T) {
	dir := t.TempDir()
	r := testruntime.New(t, nil, nil)
	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"falcula.compose": testruntime.Configure(r, New())})

	var services []any
	L.SetGlobal("capture", L.NewFunction(func(L *lua.LState) int {
		services = append(services, L.CheckUserData(1).Value)
		return 0
	}))

	compose := testutil.WriteFile(t, filepath.Join(dir, "compose.yaml"), "services: { web: { image: example/web } }\n")
	script := testutil.WriteFile(t, filepath.Join(dir, "main.lua"), `
		local compose = require("falcula.compose")
		local provider = compose.Provider:new("app", "compose.yaml")
		capture(provider:new_up_service())
		capture(provider:new_down_service())
		capture(provider:new_build_service())
		capture(provider:new_push_service())
		capture(provider:new_up_service(nil, { file = "docs/web.md" }))
		capture(provider:new_build_service({ file = "Dockerfile" }))
	`)
	if err := L.DoFile(script); err != nil {
		t.Fatal(err)
	}

	want := []string{compose, compose, compose, compose, filepath.Join(dir, "docs", "web.md"), filepath.Join(dir, "Dockerfile")}
	if len(services) != len(want) {
		t.Fatalf("got %d services, want %d", len(services), len(want))
	}
	for i, svc := range services {
		owner, ok := svc.(iface.FileOwner)
		if !ok {
			t.Fatalf("service %d (%T) does not implement iface.FileOwner", i, svc)
		}
		if got := owner.GetFile(); got != want[i] {
			t.Errorf("service %d (%T) file = %q, want %q", i, svc, got, want[i])
		}
	}
}
