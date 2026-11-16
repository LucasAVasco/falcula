package luaruntime

import (
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil/testservice"
	"github.com/LucasAVasco/falcula/service/manager"
)

func TestRemoveUnknownManagerDoesNothing(t *testing.T) {
	r := &Runtime{}
	known := manager.New("known")
	r.AddManager(known)

	r.RemoveManager(manager.New("unknown"))

	if got := r.GetManagers(); len(got) != 1 || got[0] != known {
		t.Fatalf("expected only the known manager to remain, got %v", got)
	}
}

func TestHasActiveServices(t *testing.T) {
	r := &Runtime{}
	if r.HasActiveServices() {
		t.Fatal("a runtime without managers has no active services")
	}

	man := manager.New("m")
	ended := man.AddService(&testservice.Service{Name: "ended"}, nil)
	running := man.AddService(&testservice.Service{Name: "running", RunForever: true}, nil)
	r.AddManager(man)

	if _, err := ended.Prepare(); err != nil {
		t.Fatal(err)
	}
	if _, err := ended.Run(); err != nil {
		t.Fatal(err)
	}
	if r.HasActiveServices() {
		t.Fatal("ended services are not active")
	}

	if _, err := running.Prepare(); err != nil {
		t.Fatal(err)
	}
	if err := running.Start(); err != nil {
		t.Fatal(err)
	}
	if !r.HasActiveServices() {
		t.Fatal("a running service is active")
	}

	if _, err := running.Stop(true); err != nil {
		t.Fatal(err)
	}
	if r.HasActiveServices() {
		t.Fatal("a stopped service is not active")
	}
}

func TestCloseAllManagers_KeepsTheAvailableArgs(t *testing.T) {
	r := &Runtime{}
	r.SetOnCurrentScriptArgsChange(func([]string) {})
	var notified [][]string
	r.SetOnScriptAvailableArgsChange(func(args [][]string) { notified = args })

	want := [][]string{{"dev"}}
	r.SetScriptAvailableArgs(want) // The `args` option of the task
	r.AddManager(manager.New("m"))
	r.closeAllManagers() // Before a re-run with other arguments

	if got := r.GetScriptAvailableArgs(); len(got) != 1 || got[0][0] != "dev" || len(notified) != 1 {
		t.Errorf("available args after closing the managers = %v (notified %v), want %v", got, notified, want)
	}
}
