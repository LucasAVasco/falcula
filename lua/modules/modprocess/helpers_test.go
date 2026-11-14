package modprocess

import (
	"testing"

	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/iface"
)

// runService prepares and runs a service created by the module, failing the test on errors
func runService(t *testing.T, value any) {
	t.Helper()

	svc, ok := value.(iface.Service)
	if !ok {
		t.Fatalf("expected a service, got %T", value)
	}

	e := enhanced.NewEnhancedService(svc, nil)
	if _, err := e.Prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if _, err := e.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
}
