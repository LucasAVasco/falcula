package manager

import (
	"testing"

	"github.com/LucasAVasco/falcula/service/iface"
)

type fakeService struct{ name string }

func (f *fakeService) GetName() string                                  { return f.name }
func (f *fakeService) GetOpts() *iface.Opts                             { return &iface.Opts{} }
func (f *fakeService) Prepare(iface.OnExitCallback) (iface.Step, error) { return nil, nil }
func (f *fakeService) Start(iface.OnExitCallback) (iface.Step, error)   { return nil, nil }

func TestRemoveServicesRemovesAllWhenPassedGetServices(t *testing.T) {
	m := New("test")
	for _, name := range []string{"a", "b", "c", "d"} {
		m.AddService(&fakeService{name: name}, nil)
	}

	if err := m.RemoveServices(m.GetServices(), false, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(m.GetServices()); got != 0 {
		t.Fatalf("expected no services left, got %d", got)
	}
}
