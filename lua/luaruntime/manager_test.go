package luaruntime

import (
	"testing"

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
