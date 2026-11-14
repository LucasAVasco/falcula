// Package testruntime creates Lua runtimes for tests. It must only be imported by tests
package testruntime

import (
	"testing"

	"github.com/LucasAVasco/falcula/lua/luaruntime"
	"github.com/LucasAVasco/falcula/lua/modules/base"
	"github.com/LucasAVasco/falcula/project"
)

// New creates a Lua runtime that is closed when the test ends (its log file is also removed). The project and the callback are optional
func New(t testing.TB, proj *project.Project, onRunTask luaruntime.OnRunTask) *luaruntime.Runtime {
	t.Helper()

	r, err := luaruntime.New(proj, onRunTask)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := r.Close(); err != nil {
			t.Errorf("error closing runtime: %v", err)
		}
	})

	return r
}

// Configurable is a module that requires the base module configuration (implemented by all modules of the 'lua/modules' package)
type Configurable interface {
	SetBaseModuleConfig(config *base.Config)
}

// Configure sets the runtime in the base configuration of the module and returns it
func Configure[T Configurable](r *luaruntime.Runtime, module T) T {
	module.SetBaseModuleConfig(&base.Config{Runtime: r})
	return module
}
