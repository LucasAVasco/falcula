package modprocess

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/internal/testutil/testruntime"
	"github.com/LucasAVasco/falcula/service/iface"

	lua "github.com/yuin/gopher-lua"
)

func TestGetProcessCommandFromLuaValue(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	if err := L.DoString(`list = { "echo", "a", dir = "/tmp" }; bad_dir = { "echo", dir = 1 }`); err != nil {
		t.Fatal(err)
	}

	shell, err := getProcessCommandFromLuaValue(lua.LString("echo hi"))
	if err != nil || !shell.Shell || len(shell.Command) != 1 || shell.Command[0] != "echo hi" {
		t.Errorf("string command = %+v, %v", shell, err)
	}

	list, err := getProcessCommandFromLuaValue(L.GetGlobal("list"))
	if err != nil || list.Shell || len(list.Command) != 2 || list.Dir != "/tmp" {
		t.Errorf("list command = %+v, %v", list, err)
	}

	if cmd, err := getProcessCommandFromLuaValue(lua.LNil); cmd != nil || err != nil {
		t.Errorf("nil command = %+v, %v; want no command", cmd, err)
	}

	for name, value := range map[string]lua.LValue{"number": lua.LNumber(1), "non string dir": L.GetGlobal("bad_dir")} {
		if _, err := getProcessCommandFromLuaValue(value); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestProvider_RunsItsCommandsAsServices(t *testing.T) {
	testutil.RequireShell(t)
	dir := t.TempDir()

	r := testruntime.New(t, nil, nil)
	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"process": testruntime.Configure(r, New())})
	L.SetGlobal("dir", lua.LString(dir))

	var svc any
	L.SetGlobal("capture", L.NewFunction(func(L *lua.LState) int {
		svc = L.CheckUserData(1).Value
		return 0
	}))

	err := L.DoString(`
		local process = require("process")
		local provider = process.Provider:new("web", "touch prepared", { "touch", "started", dir = dir })
		assert(provider:get_name() == "web")
		capture(provider:new_service({ start_disabled = false }))
	`)
	if err != nil {
		t.Fatal(err)
	}

	// The service runs the prepare command in the current directory and the main command in its own directory
	t.Chdir(dir)
	runService(t, svc)

	for _, file := range []string{"prepared", "started"} {
		if _, err := os.Stat(filepath.Join(dir, file)); err != nil {
			t.Errorf("the command that creates '%s' did not run: %v", file, err)
		}
	}
}

func TestProvider_InvalidArguments_RaiseErrors(t *testing.T) {
	r := testruntime.New(t, nil, nil)
	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"process": testruntime.Configure(r, New())})

	for _, code := range []string{
		`require("process").Provider:new("x", 1, nil)`,
		`require("process").Provider:new("x", nil, true)`,
		`require("process").Provider:new("x", nil, nil, { service_opts = "bad" })`,
		`local p = require("process").Provider:new("x"); p.get_name()`,
	} {
		if err := L.DoString(code); err == nil {
			t.Errorf("%s: expected an error", code)
		}
	}
}

func TestProvider_ServiceFile_IsRelativeToTheScript(t *testing.T) {
	dir := t.TempDir()
	r := testruntime.New(t, nil, nil)
	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"process": testruntime.Configure(r, New())})

	var services []any
	L.SetGlobal("capture", L.NewFunction(func(L *lua.LState) int {
		services = append(services, L.CheckUserData(1).Value)
		return 0
	}))

	script := testutil.WriteFile(t, filepath.Join(dir, "main.lua"), `
		local process = require("process")
		local provider = process.Provider:new("web", nil, "true")
		capture(provider:new_service({ file = "src/main.go" }))
		capture(provider:new_service({ file = "/abs/file" }))
		capture(provider:new_service())
	`)
	if err := L.DoFile(script); err != nil {
		t.Fatal(err)
	}

	want := []string{filepath.Join(dir, "src", "main.go"), "/abs/file", ""}
	for i, svc := range services {
		owner, ok := svc.(iface.FileOwner)
		if !ok {
			t.Fatalf("service %d (%T) does not implement iface.FileOwner", i, svc)
		}
		if got := owner.GetFile(); got != want[i] {
			t.Errorf("service %d file = %q, want %q", i, got, want[i])
		}
	}
}
