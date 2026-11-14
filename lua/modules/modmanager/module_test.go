package modmanager

import (
	"sync"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/internal/testutil/testruntime"
	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/modules/modprocess"
	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/manager"
	"github.com/LucasAVasco/falcula/service/status"

	lua "github.com/yuin/gopher-lua"
)

// recorder records the module callbacks. Status callbacks are called concurrently by the services
type recorder struct {
	mutex    sync.Mutex
	events   []string
	statuses map[string][]status.Status
}

func (r *recorder) event(e string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.events = append(r.events, e)
}

func newManagerState(t *testing.T) (*lua.LState, *recorder) {
	t.Helper()
	luaerror.ConfigureAbortOnError(false)
	t.Cleanup(func() { luaerror.ConfigureAbortOnError(true) })

	rec := &recorder{statuses: map[string][]status.Status{}}
	callbacks := &Callbacks{
		OnNewManager:    func(man *manager.Manager) { rec.event("new:" + man.GetName()) },
		OnDeleteManager: func(man *manager.Manager) { rec.event("delete:" + man.GetName()) },
		OnAddService: func(man *manager.Manager, svc *enhanced.EnhancedService) {
			rec.event("add:" + man.GetName() + ":" + svc.GetName())
		},
		OnServiceStatusChanged: func(man *manager.Manager, svc *enhanced.EnhancedService) {
			rec.mutex.Lock()
			defer rec.mutex.Unlock()
			rec.statuses[svc.GetName()] = append(rec.statuses[svc.GetName()], svc.GetStatus())
		},
	}

	r := testruntime.New(t, nil, nil)
	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{
		"manager": testruntime.Configure(r, New(callbacks)),
		"process": testruntime.Configure(r, modprocess.New()),
	})
	return L, rec
}

func TestManager_RunsServicesAndReportsEvents(t *testing.T) {
	testutil.RequireShell(t)
	L, rec := newManagerState(t)

	err := L.DoString(`
		local process = require("process")
		local man = require("manager").ServiceManager:new("main")

		man:add_service(process.Provider:new("ok", "true", "true"):new_service())
		man:add_services({ process.Provider:new("other", nil, { "true" }):new_service() })

		assert(man:run() == nil)

		-- Ended services must be reset (e.g. restarted) before running again
		local err = man:run_serial()
		assert(err ~= nil and err:find("invalid status"), tostring(err))

		man:close()
	`)
	if err != nil {
		t.Fatal(err)
	}

	wantEvents := []string{"new:main", "add:main:ok", "add:main:other", "delete:main"}
	if len(rec.events) != len(wantEvents) {
		t.Fatalf("events = %v, want %v", rec.events, wantEvents)
	}
	for i := range wantEvents {
		if rec.events[i] != wantEvents[i] {
			t.Fatalf("events = %v, want %v", rec.events, wantEvents)
		}
	}

	if got := rec.statuses["ok"]; len(got) == 0 || got[len(got)-1] != status.Ended {
		t.Fatalf("the service must end with the Ended status, got %v", got)
	}
}

func TestManager_FailingService_ReturnsError(t *testing.T) {
	testutil.RequireShell(t)
	L, _ := newManagerState(t)

	err := L.DoString(`
		local man = require("manager").ServiceManager:new("main")
		man:add_service(require("process").Provider:new("bad", nil, "exit 3"):new_service())

		local err = man:run()
		assert(err ~= nil and err:find("exit code: 3"), tostring(err))
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestManager_InvalidArguments_RaiseErrors(t *testing.T) {
	L, _ := newManagerState(t)
	luaerror.ConfigureAbortOnError(true)

	for _, code := range []string{
		`require("manager").ServiceManager:new("m"):add_service("not a service")`,
		`require("manager").ServiceManager:new("m"):add_services({ 1 })`,
		`require("manager").ServiceManager:new("m"):add_services("x")`,
		`local m = require("manager").ServiceManager:new("m"); m.run()`,
	} {
		if err := L.DoString(code); err == nil {
			t.Errorf("%s: expected an error", code)
		}
	}
}
