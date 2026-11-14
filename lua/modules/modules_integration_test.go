//go:build integration

package modules

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/internal/testutil/testruntime"
	"github.com/LucasAVasco/falcula/lua/luaruntime"
	"github.com/LucasAVasco/falcula/project"
	"github.com/LucasAVasco/falcula/service/status"
)

// newRuntime creates a runtime with all modules loaded (raw mode, without TUI) for the project in the directory
func newRuntime(t *testing.T, dir string) *luaruntime.Runtime {
	t.Helper()

	proj, err := project.LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}

	r := testruntime.New(t, proj, nil)
	loader := LoadAllModules(r, &AllModulesLoaderOptions{RawMode: true})
	t.Cleanup(func() {
		if err := loader.Close(); err != nil {
			t.Errorf("error closing modules: %v", err)
		}
	})
	return r
}

func TestIntegrationModulesWorkTogether(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "falcula.yaml"), "tasks: { hello: { shell: 'echo hi' } }\n")
	testutil.WriteFile(t, filepath.Join(dir, "templates", "config.tpl"), "name={{ .name }}\nports={{ range .ports }}{{ . }} {{ end }}")
	script := testutil.WriteFile(t, filepath.Join(dir, "main.lua"), `
		local json = require("falcula.json")
		local yaml = require("falcula.yaml")
		local tbl = require("falcula.tbl")
		local template = require("falcula.template")
		local path = require("falcula.path")
		local project = require("falcula.project")
		local cmd = require("falcula.cmd")

		-- Configuration merged from JSON and YAML
		local config = json.decode('{"name": "web", "ports": [80]}')
		tbl.deep_extend("keep", config, yaml.decode("name: ignored\nextra: { debug: true }\n"))
		assert(config.name == "web" and config.extra.debug == true)

		-- Template file relative to this script, saved next to it
		local out = path.abs("generated/config.txt")
		os.execute("mkdir -p " .. path.abs("generated"))
		local text = template.parse_and_save_file(path.abs("templates/config.tpl"), out, config)
		assert(text == "name=web\nports=80 ", text)

		-- Project information
		local p = project.get_current_project()
		assert(p:get_folder() == path.get_current_dir())
		assert(p:get_task_by_name("hello"):get_shell():get_string() == "echo hi")

		-- Script arguments
		assert(cmd.args[1] == "first" and arg[2] == "second")
	`)

	r := newRuntime(t, dir)
	if err := r.RunFile(script, "first", "second"); err != nil {
		t.Fatal(err)
	}

	if got := testutil.ReadFile(t, filepath.Join(dir, "generated", "config.txt")); got != "name=web\nports=80 " {
		t.Fatalf("generated file = %q", got)
	}
}

func TestIntegrationFsWatch_CallsTheCallbackOnChanges(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "falcula.yaml"), "tasks: {}\n")
	watched := testutil.WriteFile(t, filepath.Join(dir, "watched.txt"), "")
	script := testutil.WriteFile(t, filepath.Join(dir, "main.lua"), `
		local fs = require("falcula.fs")
		local seen = nil
		local ok, err = pcall(fs.watch, { "watched.txt" }, function(file, op)
			seen = fs.OP_CODES[op]
			error("stop")
		end)
		assert(not ok and seen ~= nil, tostring(err))
	`)

	// Keeps changing the file until the script returns (there is no signal of when the watcher is ready)
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-time.After(20 * time.Millisecond):
				_ = os.WriteFile(watched, []byte("x"), 0o644)
			}
		}
	}()

	r := newRuntime(t, dir)
	t.Chdir(dir) // The watched path is relative to the working directory
	if err := r.RunFile(script); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationInterrupt_StopsAScriptWaitingForRealProcesses(t *testing.T) {
	testutil.RequireShell(t)
	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "falcula.yaml"), "tasks: {}\n")
	script := testutil.WriteFile(t, filepath.Join(dir, "main.lua"), `
		local process = require("falcula.process")
		local man = require("falcula.manager").ServiceManager:new("long")
		man:add_service(process.Provider:new("sleeper", nil, "sleep 60"):new_service())
		man:run()
		while true do end
	`)

	r := newRuntime(t, dir)
	result := make(chan error, 1)
	go func() { result <- r.RunFile(script) }()

	// Waits until the service is running
	deadline := time.Now().Add(10 * time.Second)
	for !serviceIsRunning(r) {
		if time.Now().After(deadline) {
			t.Fatal("the service did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}

	r.Interrupt()
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("the interrupted script must return an error")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the script did not stop after the interruption")
	}

	// The runtime can run another script after resetting the state
	r.ResetLuaState()
	LoadAllModules(r, &AllModulesLoaderOptions{RawMode: true})
	if err := r.Run(`assert(require("falcula.json").encode({ 1 }) == "[1]")`); err != nil {
		t.Fatalf("running after the interruption: %v", err)
	}
}

func serviceIsRunning(r *luaruntime.Runtime) bool {
	for _, man := range r.GetManagers() {
		for _, svc := range man.GetServices() {
			if svc.GetStatus() == status.Running {
				return true
			}
		}
	}
	return false
}
